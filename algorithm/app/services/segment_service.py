"""
Customer Segmentation Service.

Supports RFM, behavioral, and value-based segmentation.
Uses K-Means, DBSCAN, and quantile-based methods respectively.
Falls back to simple MRR-tier rules for cold start.
"""

import os
import logging
from datetime import datetime, timedelta

import numpy as np
import pandas as pd
import joblib
from sklearn.cluster import KMeans, DBSCAN
from sklearn.preprocessing import StandardScaler

from app.database import get_cursor, get_connection
from app.models.schemas import SegmentationResponse

logger = logging.getLogger(__name__)

MODELS_DIR = os.path.join(os.path.dirname(os.path.dirname(__file__)), "ml", "models")


# ─── Data Extraction ────────────────────────────────────────────────────────

def _get_tenant_customers(tenant_id: str) -> list[dict]:
    """Get all customers with their basic info for a tenant."""
    with get_cursor(dict_cursor=True) as cur:
        cur.execute(
            """
            SELECT c.id AS customer_id, c.created_at,
                   COALESCE(s.mrr, 0) AS mrr, s.status AS sub_status
            FROM customers c
            LEFT JOIN subscriptions s ON s.customer_id = c.id
            WHERE c.tenant_id = %s
            """,
            (tenant_id,),
        )
        return cur.fetchall()


def _get_event_stats(tenant_id: str, days: int = 30) -> dict[str, dict]:
    """Get per-customer event statistics for the given time window."""
    since = datetime.utcnow() - timedelta(days=days)
    with get_cursor(dict_cursor=True) as cur:
        cur.execute(
            """
            SELECT customer_id,
                   COUNT(*) AS event_count,
                   MAX(created_at) AS last_event_at,
                   COUNT(DISTINCT event_type) AS distinct_event_types
            FROM user_events
            WHERE tenant_id = %s AND created_at >= %s
            GROUP BY customer_id
            """,
            (tenant_id, since),
        )
        rows = cur.fetchall()

    stats = {}
    for r in rows:
        # Build feature usage vector: count per event_type
        stats[r["customer_id"]] = {
            "event_count": float(r["event_count"] or 0),
            "last_event_at": r["last_event_at"],
            "distinct_event_types": float(r["distinct_event_types"] or 0),
        }
    return stats


def _get_feature_usage_vectors(tenant_id: str, days: int = 30) -> dict[str, dict]:
    """Get per-customer feature usage vectors for behavioral clustering."""
    since = datetime.utcnow() - timedelta(days=days)
    with get_cursor(dict_cursor=True) as cur:
        cur.execute(
            """
            SELECT customer_id, event_type, COUNT(*) AS cnt
            FROM user_events
            WHERE tenant_id = %s AND created_at >= %s AND event_type != 'login'
            GROUP BY customer_id, event_type
            """,
            (tenant_id, since),
        )
        rows = cur.fetchall()

    vectors: dict[str, dict[str, float]] = {}
    for r in rows:
        cid = r["customer_id"]
        if cid not in vectors:
            vectors[cid] = {}
        vectors[cid][r["event_type"]] = float(r["cnt"])
    return vectors


# ─── RFM Segmentation ───────────────────────────────────────────────────────

RFM_SEGMENT_NAMES = {
    0: "Champions",
    1: "Loyal Customers",
    2: "At Risk",
    3: "Hibernating",
}


def _segment_rfm(tenant_id: str, customers: list[dict]) -> list[SegmentationResponse]:
    """
    RFM segmentation: Recency, Frequency, Monetary.
    K-Means with k=4 clusters.
    """
    event_stats = _get_event_stats(tenant_id, days=30)
    now = datetime.utcnow()

    rows = []
    for c in customers:
        cid = c["customer_id"]
        stats = event_stats.get(cid, {})
        last_event = stats.get("last_event_at")
        recency = float((now - last_event).days) if last_event else 999.0
        frequency = stats.get("event_count", 0.0)
        monetary = float(c.get("mrr") or 0)
        rows.append({
            "customer_id": cid,
            "recency": recency,
            "frequency": frequency,
            "monetary": monetary,
        })

    if len(rows) < 4:
        return _cold_start_segment(tenant_id, customers, "rfm")

    df = pd.DataFrame(rows)
    features = df[["recency", "frequency", "monetary"]].values.astype(np.float64)

    # Handle zero variance
    std = features.std(axis=0)
    if np.any(std == 0):
        return _cold_start_segment(tenant_id, customers, "rfm")

    scaler = StandardScaler()
    X_scaled = scaler.fit_transform(features)

    k = min(4, len(rows))
    model = KMeans(n_clusters=k, random_state=42, n_init=10)
    labels = model.fit_predict(X_scaled)

    # Sort clusters by RFM score (lower recency + higher frequency + higher monetary = better)
    cluster_centers = model.cluster_centers_
    # RFM score: -recency + frequency + monetary (after scaling)
    rfm_scores = -cluster_centers[:, 0] + cluster_centers[:, 1] + cluster_centers[:, 2]
    sorted_indices = np.argsort(rfm_scores)[::-1]  # best first

    # Map cluster labels to segment names
    label_to_segment = {}
    for rank, idx in enumerate(sorted_indices):
        if rank < len(RFM_SEGMENT_NAMES):
            label_to_segment[idx] = RFM_SEGMENT_NAMES[rank]
        else:
            label_to_segment[idx] = f"Segment_{rank}"

    results = []
    for i, row in enumerate(rows):
        label = int(labels[i])
        segment_name = label_to_segment.get(label, f"Segment_{label}")
        # Score: distance to cluster center (normalized)
        dist = float(np.linalg.norm(X_scaled[i] - cluster_centers[label]))
        score = round(max(0.0, 1.0 - dist), 4)
        results.append(SegmentationResponse(
            customer_id=row["customer_id"],
            segment_name=segment_name,
            score=score,
        ))

    # Save model
    _save_segmentation_model(tenant_id, "rfm", model, scaler)

    return results


# ─── Behavioral Segmentation ────────────────────────────────────────────────

BEHAVIORAL_SEGMENT_NAMES = {
    -1: "Noise",
    0: "Power Users",
    1: "Moderate Users",
    2: "Light Users",
    3: "Niche Users",
}


def _segment_behavioral(tenant_id: str, customers: list[dict]) -> list[SegmentationResponse]:
    """
    Behavioral segmentation using feature usage vectors.
    DBSCAN clustering on the usage vector space.
    """
    vectors = _get_feature_usage_vectors(tenant_id, days=30)

    if not vectors:
        return _cold_start_segment(tenant_id, customers, "behavioral")

    # Build a consistent feature list
    all_features = sorted(set(f for vec in vectors.values() for f in vec))
    if not all_features:
        return _cold_start_segment(tenant_id, customers, "behavioral")

    customer_ids = list(vectors.keys())
    X = np.zeros((len(customer_ids), len(all_features)), dtype=np.float64)
    for i, cid in enumerate(customer_ids):
        for j, feat in enumerate(all_features):
            X[i, j] = vectors[cid].get(feat, 0.0)

    if len(X) < 4:
        return _cold_start_segment(tenant_id, customers, "behavioral")

    # Standardize
    std = X.std(axis=0)
    std[std == 0] = 1.0  # avoid division by zero
    X_scaled = StandardScaler().fit_transform(X)

    # DBSCAN
    model = DBSCAN(eps=1.5, min_samples=3)
    labels = model.fit_predict(X_scaled)

    # Map labels to names
    unique_labels = sorted(set(labels) - {-1})
    label_to_name = {-1: "Noise"}
    for rank, lbl in enumerate(unique_labels):
        if rank < len(BEHAVIORAL_SEGMENT_NAMES) - 1:
            name = list(BEHAVIORAL_SEGMENT_NAMES.values())[rank + 1]
        else:
            name = f"Cluster_{lbl}"
        label_to_name[lbl] = name

    # Build results for customers with event data
    results_map = {}
    for i, cid in enumerate(customer_ids):
        label = int(labels[i])
        segment_name = label_to_name.get(label, f"Cluster_{label}")
        # Score: total usage count as a proxy for engagement
        score = round(float(X[i].sum()), 4)
        results_map[cid] = SegmentationResponse(
            customer_id=cid, segment_name=segment_name, score=score,
        )

    # Customers without events get "Inactive"
    results = []
    for c in customers:
        cid = c["customer_id"]
        if cid in results_map:
            results.append(results_map[cid])
        else:
            results.append(SegmentationResponse(
                customer_id=cid, segment_name="Inactive", score=0.0,
            ))

    return results


# ─── Value Segmentation ─────────────────────────────────────────────────────

VALUE_SEGMENT_NAMES = ["Platinum", "Gold", "Silver", "Bronze"]


def _segment_value(tenant_id: str, customers: list[dict]) -> list[SegmentationResponse]:
    """
    Value-based segmentation using LTV / MRR quantiles.
    """
    # Try to use LTV predictions if available, otherwise MRR
    with get_cursor(dict_cursor=True) as cur:
        cur.execute(
            """
            SELECT customer_id, predicted_ltv FROM ltv_predictions WHERE tenant_id = %s
            """,
            (tenant_id,),
        )
        ltv_rows = cur.fetchall()

    ltv_map = {r["customer_id"]: float(r["predicted_ltv"] or 0) for r in ltv_rows}

    rows = []
    for c in customers:
        cid = c["customer_id"]
        value = ltv_map.get(cid, float(c.get("mrr") or 0))
        rows.append({"customer_id": cid, "value": value})

    if len(rows) < 4:
        return _cold_start_segment(tenant_id, customers, "value")

    df = pd.DataFrame(rows)
    values = df["value"].values.astype(np.float64)

    if np.all(values == values[0]):
        # All same value
        return _cold_start_segment(tenant_id, customers, "value")

    # Quantile-based segmentation
    quantiles = np.percentile(values, [25, 50, 75])

    results = []
    for _, row in df.iterrows():
        v = row["value"]
        if v >= quantiles[2]:
            segment_name = VALUE_SEGMENT_NAMES[0]  # Platinum
        elif v >= quantiles[1]:
            segment_name = VALUE_SEGMENT_NAMES[1]  # Gold
        elif v >= quantiles[0]:
            segment_name = VALUE_SEGMENT_NAMES[2]  # Silver
        else:
            segment_name = VALUE_SEGMENT_NAMES[3]  # Bronze

        score = round(v, 4)
        results.append(SegmentationResponse(
            customer_id=row["customer_id"], segment_name=segment_name, score=score,
        ))

    return results


# ─── Cold Start Fallback ────────────────────────────────────────────────────

MRR_TIERS = [
    (0, "Free"),
    (50, "Starter"),
    (200, "Growth"),
    (1000, "Enterprise"),
]


def _cold_start_segment(tenant_id: str, customers: list[dict],
                        segment_type: str) -> list[SegmentationResponse]:
    """
    Rule-based segmentation for cold start scenarios.
    Uses MRR tiers to assign segments.
    """
    results = []
    for c in customers:
        mrr = float(c.get("mrr") or 0)
        segment_name = MRR_TIERS[0][1]
        for threshold, name in MRR_TIERS:
            if mrr >= threshold:
                segment_name = name
        results.append(SegmentationResponse(
            customer_id=c["customer_id"],
            segment_name=f"{segment_type}_{segment_name}",
            score=round(mrr, 4),
        ))
    return results


# ─── Model Persistence ──────────────────────────────────────────────────────

def _save_segmentation_model(tenant_id: str, segment_type: str,
                             model, scaler) -> None:
    """Save a segmentation model to disk."""
    os.makedirs(MODELS_DIR, exist_ok=True)
    model_path = os.path.join(MODELS_DIR, f"segment_{segment_type}_{tenant_id}.pkl")
    scaler_path = os.path.join(MODELS_DIR, f"segment_{segment_type}_{tenant_id}_scaler.pkl")
    joblib.dump(model, model_path)
    joblib.dump(scaler, scaler_path)
    logger.info("Segmentation model saved: %s", model_path)


# ─── Save Results ───────────────────────────────────────────────────────────

def _save_results(tenant_id: str, segment_type: str,
                  results: list[SegmentationResponse]) -> None:
    """Persist segmentation results to customer_segments table."""
    with get_connection() as conn:
        with conn.cursor() as cur:
            for r in results:
                cur.execute(
                    """
                    INSERT INTO customer_segments
                        (customer_id, tenant_id, segment_type, segment_name, score, segmented_at)
                    VALUES (%s, %s, %s, %s, %s, NOW())
                    ON CONFLICT (customer_id, tenant_id, segment_type)
                    DO UPDATE SET
                        segment_name = EXCLUDED.segment_name,
                        score = EXCLUDED.score,
                        segmented_at = NOW()
                    """,
                    (r.customer_id, tenant_id, segment_type, r.segment_name, r.score),
                )


# ─── Public API ─────────────────────────────────────────────────────────────

def segment_customers(tenant_id: str, segment_type: str) -> list[SegmentationResponse]:
    """
    Run customer segmentation for a tenant.

    Args:
        tenant_id: The tenant to segment
        segment_type: One of 'rfm', 'behavioral', or 'value'

    Returns:
        List of segmentation results
    """
    customers = _get_tenant_customers(tenant_id)
    if not customers:
        logger.warning("No customers found for tenant %s", tenant_id)
        return []

    if segment_type == "rfm":
        results = _segment_rfm(tenant_id, customers)
    elif segment_type == "behavioral":
        results = _segment_behavioral(tenant_id, customers)
    elif segment_type == "value":
        results = _segment_value(tenant_id, customers)
    else:
        raise ValueError(f"Unknown segment_type: {segment_type}. Must be rfm, behavioral, or value.")

    _save_results(tenant_id, segment_type, results)
    logger.info(
        "Segmentation completed: tenant=%s, type=%s, customers=%d",
        tenant_id, segment_type, len(results),
    )
    return results
