"""
LTV (Lifetime Value) Prediction Service.

Uses Cox Proportional Hazards model from lifelines for survival analysis.
Falls back to simple MRR * industry_average_months for cold start.
"""

import os
import logging
from datetime import datetime, timedelta

import numpy as np
import pandas as pd
import joblib
from lifelines import CoxPHFitter

from app.database import get_cursor, get_connection
from app.models.schemas import LTVResponse, ChannelROI

logger = logging.getLogger(__name__)

MODELS_DIR = os.path.join(os.path.dirname(os.path.dirname(__file__)), "ml", "models")
LTV_MODEL_PATH = os.path.join(MODELS_DIR, "ltv_cox_model.pkl")

INDUSTRY_AVERAGE_MONTHS = 12.0


# ─── Data Extraction ────────────────────────────────────────────────────────

def _get_customer_data(customer_id: str, tenant_id: str) -> dict | None:
    """Get customer subscription and behavioral data for LTV prediction."""
    now = datetime.utcnow()
    thirty_days_ago = now - timedelta(days=30)

    with get_cursor(dict_cursor=True) as cur:
        # Customer and subscription info
        cur.execute(
            """
            SELECT c.id, c.created_at, c.acquisition_channel,
                   s.id AS sub_id, s.status AS sub_status,
                   s.mrr, s.current_period_start, s.current_period_end
            FROM customers c
            LEFT JOIN subscriptions s ON s.customer_id = c.id AND s.tenant_id = %s
            WHERE c.id = %s AND c.tenant_id = %s
            ORDER BY s.created_at DESC LIMIT 1
            """,
            (tenant_id, customer_id, tenant_id),
        )
        customer = cur.fetchone()
        if not customer:
            return None

        # Behavioral features
        cur.execute(
            """
            SELECT COUNT(*) FILTER (WHERE event_type = 'login') AS login_count,
                   COUNT(DISTINCT event_type) AS feature_breadth,
                   COUNT(*) FILTER (WHERE event_type = 'payment_failed') AS payment_failures,
                   MAX(created_at) AS last_event_at
            FROM user_events
            WHERE customer_id = %s AND tenant_id = %s AND created_at >= %s
            """,
            (customer_id, tenant_id, thirty_days_ago),
        )
        behavior = cur.fetchone()

        # Subscription duration in months
        created_at = customer.get("created_at")
        if created_at:
            duration_days = (now - created_at).days
            duration_months = duration_days / 30.0
        else:
            duration_months = 0.0

        # Check if customer has churned (subscription cancelled)
        cur.execute(
            """
            SELECT 1 FROM user_events
            WHERE customer_id = %s AND tenant_id = %s
              AND event_type = 'subscription_cancelled'
            LIMIT 1
            """,
            (customer_id, tenant_id),
        )
        churned = cur.fetchone() is not None

    return {
        "customer_id": customer_id,
        "tenant_id": tenant_id,
        "created_at": created_at,
        "acquisition_channel": customer.get("acquisition_channel", "unknown"),
        "mrr": float(customer.get("mrr") or 0),
        "sub_status": customer.get("sub_status"),
        "duration_months": duration_months,
        "churned": churned,
        "login_count": float(behavior["login_count"] or 0) if behavior else 0.0,
        "feature_breadth": float(behavior["feature_breadth"] or 0) if behavior else 0.0,
        "payment_failures": float(behavior["payment_failures"] or 0) if behavior else 0.0,
        "last_event_at": behavior["last_event_at"] if behavior else None,
    }


def _get_tenant_customer_data(tenant_id: str) -> list[dict]:
    """Get data for all customers of a tenant."""
    with get_cursor(dict_cursor=True) as cur:
        cur.execute(
            "SELECT id FROM customers WHERE tenant_id = %s",
            (tenant_id,),
        )
        customers = cur.fetchall()

    results = []
    for row in customers:
        data = _get_customer_data(row["id"], tenant_id)
        if data is not None:
            results.append(data)
    return results


# ─── Model Training ─────────────────────────────────────────────────────────

def _has_trained_model() -> bool:
    """Check if a trained Cox model exists on disk."""
    return os.path.exists(LTV_MODEL_PATH)


def _load_model() -> CoxPHFitter:
    """Load the trained Cox model from disk."""
    return joblib.load(LTV_MODEL_PATH)


def _save_model(model: CoxPHFitter) -> None:
    """Persist the trained model to disk."""
    os.makedirs(MODELS_DIR, exist_ok=True)
    joblib.dump(model, LTV_MODEL_PATH)
    logger.info("LTV model saved to %s", LTV_MODEL_PATH)


def _train_model(customer_data_list: list[dict]) -> None:
    """
    Train a Cox Proportional Hazards model on tenant data.
    Requires customers with both churned and active status.
    """
    if len(customer_data_list) < 20:
        logger.warning("Not enough data to train LTV model (%d customers).", len(customer_data_list))
        return

    df = pd.DataFrame(customer_data_list)

    # Prepare survival data
    df["T"] = df["duration_months"].clip(lower=0.1)  # duration
    df["E"] = df["churned"].astype(int)  # event observed

    covariates = ["login_count", "feature_breadth", "payment_failures", "mrr"]
    train_df = df[["T", "E"] + covariates].copy()

    # Remove zero-variance columns
    for col in covariates:
        if train_df[col].std() == 0:
            train_df[col] = train_df[col] + np.random.normal(0, 0.01, len(train_df))

    try:
        model = CoxPHFitter()
        model.fit(train_df, duration_col="T", event_col="E")
        _save_model(model)
        logger.info("LTV Cox model trained successfully on %d samples.", len(train_df))
    except Exception as e:
        logger.warning("Cox model training failed: %s", e)


# ─── ML-Based Prediction ────────────────────────────────────────────────────

def _ml_predict_ltv(data: dict) -> tuple[float, float, float]:
    """
    ML-based LTV prediction using Cox Proportional Hazards.
    Returns (predicted_ltv, confidence, expected_lifetime_months).
    """
    model = _load_model()

    covariates = ["login_count", "feature_breadth", "payment_failures", "mrr"]
    individual_df = pd.DataFrame([{col: data[col] for col in covariates}])

    # Predict survival function
    survival = model.predict_survival_function(individual_df)
    sf = survival.iloc[:, 0]

    # Expected lifetime: area under survival curve (median method)
    # Find the time where survival probability drops below 0.5
    median_survival = model.predict_median(individual_df).iloc[0]

    if pd.isna(median_survival) or median_survival <= 0:
        # Fallback to mean survival time
        expected_lifetime = data["duration_months"] + INDUSTRY_AVERAGE_MONTHS
    else:
        expected_lifetime = float(median_survival)

    # If already active for some months, remaining lifetime
    remaining_months = max(expected_lifetime - data["duration_months"], 1.0)

    # LTV = MRR * remaining months
    predicted_ltv = data["mrr"] * remaining_months

    # Confidence based on number of observations and model concordance
    try:
        concordance = model.concordance_index_
        confidence = round(min(max(concordance, 0.5), 0.95), 4)
    except Exception:
        confidence = 0.6

    return round(predicted_ltv, 2), confidence, round(remaining_months, 2)


# ─── Cold Start Fallback ────────────────────────────────────────────────────

def _cold_start_ltv(data: dict) -> tuple[float, float, float]:
    """
    Rule-based LTV for cold start: LTV = MRR * industry_average_months.
    Returns (predicted_ltv, confidence, expected_lifetime_months).
    """
    expected_lifetime = INDUSTRY_AVERAGE_MONTHS

    # Adjust based on subscription status
    sub_status = data.get("sub_status")
    if sub_status == "trialing":
        expected_lifetime = INDUSTRY_AVERAGE_MONTHS * 0.5  # trial less likely to convert
    elif sub_status == "active":
        # Already active for some time, give credit
        expected_lifetime = max(INDUSTRY_AVERAGE_MONTHS, data["duration_months"] + 6)

    predicted_ltv = data["mrr"] * expected_lifetime
    confidence = 0.4  # Low confidence for rule-based

    return round(predicted_ltv, 2), confidence, round(expected_lifetime, 2)


# ─── Save Prediction ────────────────────────────────────────────────────────

def _save_prediction(customer_id: str, tenant_id: str, predicted_ltv: float,
                     confidence: float, expected_lifetime_months: float) -> None:
    """Persist LTV prediction to the ltv_predictions table."""
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO ltv_predictions
                    (customer_id, tenant_id, predicted_ltv, confidence,
                     expected_lifetime_months, predicted_at)
                VALUES (%s, %s, %s, %s, %s, NOW())
                ON CONFLICT (customer_id, tenant_id)
                DO UPDATE SET
                    predicted_ltv = EXCLUDED.predicted_ltv,
                    confidence = EXCLUDED.confidence,
                    expected_lifetime_months = EXCLUDED.expected_lifetime_months,
                    predicted_at = NOW()
                """,
                (customer_id, tenant_id, predicted_ltv, confidence, expected_lifetime_months),
            )


# ─── Channel ROI ────────────────────────────────────────────────────────────

def _get_channel_roi(tenant_id: str) -> list[ChannelROI]:
    """
    Calculate ROI per acquisition channel.
    CAC = total acquisition cost / customers acquired (estimated from MRR).
    LTV = average predicted LTV per channel.
    """
    with get_cursor(dict_cursor=True) as cur:
        # Get LTV predictions with channel info
        cur.execute(
            """
            SELECT c.acquisition_channel, l.predicted_ltv
            FROM ltv_predictions l
            JOIN customers c ON c.id = l.customer_id AND c.tenant_id = %s
            WHERE l.tenant_id = %s AND c.acquisition_channel IS NOT NULL
            """,
            (tenant_id, tenant_id),
        )
        rows = cur.fetchall()

    if not rows:
        # Fallback: use raw MRR-based estimates
        with get_cursor(dict_cursor=True) as cur:
            cur.execute(
                """
                SELECT acquisition_channel, AVG(mrr) AS avg_mrr, COUNT(*) AS cnt
                FROM customers c
                LEFT JOIN subscriptions s ON s.customer_id = c.id AND s.tenant_id = %s
                WHERE c.tenant_id = %s AND c.acquisition_channel IS NOT NULL
                GROUP BY acquisition_channel
                """,
                (tenant_id, tenant_id),
            )
            fallback_rows = cur.fetchall()

        results = []
        for r in fallback_rows:
            channel = r["acquisition_channel"] or "unknown"
            avg_mrr = float(r["avg_mrr"] or 0)
            # Estimate CAC as 3x MRR (industry heuristic)
            cac = avg_mrr * 3
            predicted_ltv = avg_mrr * INDUSTRY_AVERAGE_MONTHS
            ltv_cac_ratio = round(predicted_ltv / cac, 2) if cac > 0 else 0.0
            results.append(ChannelROI(
                channel=channel,
                cac=round(cac, 2),
                predicted_ltv=round(predicted_ltv, 2),
                ltv_cac_ratio=ltv_cac_ratio,
            ))
        return results

    # Group by channel
    channel_data: dict[str, list[float]] = {}
    for r in rows:
        ch = r["acquisition_channel"] or "unknown"
        if ch not in channel_data:
            channel_data[ch] = []
        channel_data[ch].append(float(r["predicted_ltv"] or 0))

    results = []
    for channel, ltv_values in channel_data.items():
        avg_ltv = np.mean(ltv_values)
        # Estimate CAC: assume 20% of first-year revenue
        cac = (avg_ltv / INDUSTRY_AVERAGE_MONTHS) * 12 * 0.2
        ltv_cac_ratio = round(avg_ltv / cac, 2) if cac > 0 else 0.0
        results.append(ChannelROI(
            channel=channel,
            cac=round(cac, 2),
            predicted_ltv=round(avg_ltv, 2),
            ltv_cac_ratio=ltv_cac_ratio,
        ))

    return results


# ─── Public API ─────────────────────────────────────────────────────────────

def predict_ltv(customer_id: str, tenant_id: str) -> LTVResponse:
    """
    Predict lifetime value for a single customer.

    Uses Cox Proportional Hazards model if trained and sufficient data exists.
    Otherwise falls back to MRR * industry_average_months.
    """
    data = _get_customer_data(customer_id, tenant_id)
    if data is None:
        raise ValueError(f"Customer {customer_id} not found for tenant {tenant_id}")

    use_ml = _has_trained_model() and data["duration_months"] >= 1.0

    if use_ml:
        try:
            predicted_ltv, confidence, expected_lifetime = _ml_predict_ltv(data)
        except Exception as e:
            logger.warning("ML LTV prediction failed, falling back to cold start: %s", e)
            predicted_ltv, confidence, expected_lifetime = _cold_start_ltv(data)
    else:
        predicted_ltv, confidence, expected_lifetime = _cold_start_ltv(data)

    _save_prediction(customer_id, tenant_id, predicted_ltv, confidence, expected_lifetime)

    return LTVResponse(
        customer_id=customer_id,
        predicted_ltv=predicted_ltv,
        confidence=confidence,
        expected_lifetime_months=expected_lifetime,
    )


def get_channel_roi(tenant_id: str) -> list[ChannelROI]:
    """
    Calculate ROI per acquisition channel for a tenant.
    """
    return _get_channel_roi(tenant_id)
