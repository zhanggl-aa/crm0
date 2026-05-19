"""
Churn Prediction Service.

Uses XGBoost classifier with behavioral features from user_events.
Falls back to rule-based prediction for cold-start scenarios.
"""

import os
import logging
from datetime import datetime, timedelta

import numpy as np
import pandas as pd
import joblib
from xgboost import XGBClassifier
from sklearn.preprocessing import StandardScaler

from app.database import get_cursor, get_connection
from app.models.schemas import ChurnPredictionResponse, ChurnFactor

logger = logging.getLogger(__name__)

MODELS_DIR = os.path.join(os.path.dirname(os.path.dirname(__file__)), "ml", "models")
CHURN_MODEL_PATH = os.path.join(MODELS_DIR, "churn_model.pkl")
CHURN_SCALER_PATH = os.path.join(MODELS_DIR, "churn_scaler.pkl")

FEATURE_COLUMNS = [
    "login_freq_7d",
    "login_freq_14d",
    "login_freq_30d",
    "feature_usage_breadth",
    "payment_failure_count",
    "support_ticket_count",
    "subscription_days",
    "mrr_change_rate",
]

# ─── Data Extraction ────────────────────────────────────────────────────────

def _extract_customer_features(customer_id: str, tenant_id: str) -> dict | None:
    """Extract engineered features from the database for a single customer."""
    now = datetime.utcnow()
    seven_days_ago = now - timedelta(days=7)
    fourteen_days_ago = now - timedelta(days=14)
    thirty_days_ago = now - timedelta(days=30)

    with get_cursor(dict_cursor=True) as cur:
        # Verify customer exists and get subscription info
        cur.execute(
            """
            SELECT c.id, c.created_at,
                   s.id AS subscription_id, s.status AS sub_status,
                   s.current_period_start, s.current_period_end, s.mrr
            FROM customers c
            LEFT JOIN subscriptions s ON s.customer_id = c.id
            WHERE c.id = %s AND c.tenant_id = %s
            ORDER BY s.created_at DESC LIMIT 1
            """,
            (customer_id, tenant_id),
        )
        customer = cur.fetchone()
        if not customer:
            return None

        # Login frequency in different windows
        cur.execute(
            """
            SELECT COUNT(*) FILTER (WHERE created_at >= %s) AS login_freq_7d,
                   COUNT(*) FILTER (WHERE created_at >= %s) AS login_freq_14d,
                   COUNT(*) AS login_freq_30d
            FROM user_events
            WHERE customer_id = %s AND tenant_id = %s
              AND event_type = 'login' AND created_at >= %s
            """,
            (seven_days_ago, fourteen_days_ago, customer_id, tenant_id, thirty_days_ago),
        )
        login_row = cur.fetchone()
        login_freq_7d = float(login_row["login_freq_7d"] or 0)
        login_freq_14d = float(login_row["login_freq_14d"] or 0)
        login_freq_30d = float(login_row["login_freq_30d"] or 0)

        # Feature usage breadth: distinct feature types used in last 30 days
        cur.execute(
            """
            SELECT COUNT(DISTINCT event_type) AS breadth
            FROM user_events
            WHERE customer_id = %s AND tenant_id = %s
              AND event_type != 'login' AND created_at >= %s
            """,
            (customer_id, tenant_id, thirty_days_ago),
        )
        breadth_row = cur.fetchone()
        feature_usage_breadth = float(breadth_row["breadth"] or 0)

        # Payment failure count in last 30 days
        cur.execute(
            """
            SELECT COUNT(*) AS fail_count
            FROM user_events
            WHERE customer_id = %s AND tenant_id = %s
              AND event_type = 'payment_failed' AND created_at >= %s
            """,
            (customer_id, tenant_id, thirty_days_ago),
        )
        fail_row = cur.fetchone()
        payment_failure_count = float(fail_row["fail_count"] or 0)

        # Support ticket count in last 30 days
        cur.execute(
            """
            SELECT COUNT(*) AS ticket_count
            FROM user_events
            WHERE customer_id = %s AND tenant_id = %s
              AND event_type = 'support_ticket' AND created_at >= %s
            """,
            (customer_id, tenant_id, thirty_days_ago),
        )
        ticket_row = cur.fetchone()
        support_ticket_count = float(ticket_row["ticket_count"] or 0)

        # Subscription duration in days
        created_at = customer.get("created_at")
        subscription_days = float((now - created_at).days) if created_at else 0.0

        # MRR change rate: compare current MRR to MRR 30 days ago
        current_mrr = float(customer.get("mrr") or 0)
        cur.execute(
            """
            SELECT mrr FROM subscriptions
            WHERE customer_id = %s AND tenant_id = %s AND created_at < %s
            ORDER BY created_at DESC LIMIT 1
            """,
            (customer_id, tenant_id, thirty_days_ago),
        )
        prev_mrr_row = cur.fetchone()
        prev_mrr = float(prev_mrr_row["mrr"] or 0) if prev_mrr_row else current_mrr
        if prev_mrr > 0:
            mrr_change_rate = (current_mrr - prev_mrr) / prev_mrr
        else:
            mrr_change_rate = 0.0

    return {
        "customer_id": customer_id,
        "tenant_id": tenant_id,
        "login_freq_7d": login_freq_7d,
        "login_freq_14d": login_freq_14d,
        "login_freq_30d": login_freq_30d,
        "feature_usage_breadth": feature_usage_breadth,
        "payment_failure_count": payment_failure_count,
        "support_ticket_count": support_ticket_count,
        "subscription_days": subscription_days,
        "mrr_change_rate": mrr_change_rate,
        "current_mrr": current_mrr,
        "sub_status": customer.get("sub_status"),
        "current_period_end": customer.get("current_period_end"),
    }


def _extract_tenant_features(tenant_id: str) -> list[dict]:
    """Extract features for all customers of a tenant for batch prediction."""
    with get_cursor(dict_cursor=True) as cur:
        cur.execute(
            """
            SELECT id FROM customers WHERE tenant_id = %s
            """,
            (tenant_id,),
        )
        customers = cur.fetchall()

    results = []
    for row in customers:
        features = _extract_customer_features(row["id"], tenant_id)
        if features is not None:
            results.append(features)
    return results


# ─── Model Training ─────────────────────────────────────────────────────────

def _has_trained_model() -> bool:
    """Check if a trained XGBoost model exists on disk."""
    return os.path.exists(CHURN_MODEL_PATH) and os.path.exists(CHURN_SCALER_PATH)


def _load_model() -> tuple[XGBClassifier, StandardScaler]:
    """Load the trained model and scaler from disk."""
    model = joblib.load(CHURN_MODEL_PATH)
    scaler = joblib.load(CHURN_SCALER_PATH)
    return model, scaler


def _save_model(model: XGBClassifier, scaler: StandardScaler) -> None:
    """Persist the trained model and scaler to disk."""
    os.makedirs(MODELS_DIR, exist_ok=True)
    joblib.dump(model, CHURN_MODEL_PATH)
    joblib.dump(scaler, CHURN_SCALER_PATH)
    logger.info("Churn model saved to %s", CHURN_MODEL_PATH)


def _train_model(features_list: list[dict]) -> None:
    """
    Train XGBoost churn model using historical churn labels.
    This is called when enough labeled data exists (>=30 days of history with churn outcomes).
    """
    if not features_list:
        logger.warning("No data available for training churn model.")
        return

    df = pd.DataFrame(features_list)
    with get_cursor(dict_cursor=True) as cur:
        # Get churn labels: customers who have a churn event
        customer_ids = tuple(df["customer_id"].tolist())
        if not customer_ids:
            return
        placeholders = ",".join(["%s"] * len(customer_ids))
        cur.execute(
            f"""
            SELECT customer_id, 1 AS churned
            FROM user_events
            WHERE tenant_id = %s
              AND event_type = 'subscription_cancelled'
              AND customer_id IN ({placeholders})
            GROUP BY customer_id
            """,
            (features_list[0]["tenant_id"], *customer_ids),
        )
        churn_rows = cur.fetchall()

    churn_map = {r["customer_id"]: 1 for r in churn_rows}
    df["churned"] = df["customer_id"].map(churn_map).fillna(0).astype(int)

    X = df[FEATURE_COLUMNS].values.astype(np.float32)
    y = df["churned"].values

    if len(np.unique(y)) < 2:
        logger.warning("Only one class present; cannot train churn model.")
        return

    scaler = StandardScaler()
    X_scaled = scaler.fit_transform(X)

    model = XGBClassifier(
        n_estimators=100,
        max_depth=4,
        learning_rate=0.1,
        subsample=0.8,
        colsample_bytree=0.8,
        random_state=42,
        use_label_encoder=False,
        eval_metric="logloss",
    )
    model.fit(X_scaled, y)
    _save_model(model, scaler)
    logger.info("Churn model trained successfully on %d samples.", len(X))


# ─── Rule-Based Fallback ────────────────────────────────────────────────────

def _rule_based_predict(features: dict) -> tuple[float, str, list[ChurnFactor]]:
    """
    Rule-based churn prediction for cold start (no trained model or <30 days data).
    Returns (risk_score, risk_level, top_factors).
    """
    factors: list[ChurnFactor] = []
    login_weekly = features["login_freq_7d"]
    payment_failures = features["payment_failure_count"]
    login_freq_30d = features["login_freq_30d"]
    sub_days = features["subscription_days"]
    mrr_change = features["mrr_change_rate"]
    support_tickets = features["support_ticket_count"]
    feature_breadth = features["feature_usage_breadth"]

    # Calculate weekly login frequency
    weekly_login_freq = login_freq_30d / 4.3 if login_freq_30d > 0 else login_weekly

    # Check subscription ending soon (within 7 days)
    sub_ending_soon = False
    period_end = features.get("current_period_end")
    if period_end:
        days_until_end = (period_end - datetime.utcnow()).days
        if 0 <= days_until_end <= 7:
            sub_ending_soon = True

    risk_score = 0.0

    # High risk conditions
    if weekly_login_freq < 1.0:
        factors.append(ChurnFactor(factor="low_login_frequency", weight=0.35))
        risk_score += 0.35
    if payment_failures > 0:
        factors.append(ChurnFactor(factor="payment_failure", weight=0.30))
        risk_score += 0.30
    if sub_ending_soon:
        factors.append(ChurnFactor(factor="subscription_ending_soon", weight=0.20))
        risk_score += 0.20

    # Medium risk conditions
    if 1.0 <= weekly_login_freq < 3.0 and payment_failures == 0:
        factors.append(ChurnFactor(factor="moderate_login_frequency", weight=0.20))
        risk_score += 0.20
    if mrr_change < -0.1:
        factors.append(ChurnFactor(factor="mrr_declining", weight=0.15))
        risk_score += 0.15
    if support_tickets > 2:
        factors.append(ChurnFactor(factor="high_support_tickets", weight=0.10))
        risk_score += 0.10

    # Low usage breadth
    if feature_breadth < 2:
        factors.append(ChurnFactor(factor="low_feature_usage", weight=0.10))
        risk_score += 0.10

    # Cap risk score
    risk_score = min(risk_score, 1.0)

    # If no factors, customer is healthy
    if not factors:
        factors.append(ChurnFactor(factor="healthy_engagement", weight=0.05))
        risk_score = 0.05

    # Determine risk level
    if risk_score >= 0.5 or weekly_login_freq < 1.0 or payment_failures > 0:
        risk_level = "high"
    elif risk_score >= 0.25 or weekly_login_freq < 3.0 or sub_ending_soon:
        risk_level = "medium"
    else:
        risk_level = "low"

    # Sort factors by weight descending, keep top 3
    factors.sort(key=lambda f: f.weight, reverse=True)
    top_factors = factors[:3]

    return risk_score, risk_level, top_factors


# ─── ML-Based Prediction ────────────────────────────────────────────────────

def _ml_predict(features: dict) -> tuple[float, str, list[ChurnFactor]]:
    """
    ML-based churn prediction using the trained XGBoost model.
    Returns (risk_score, risk_level, top_factors).
    """
    model, scaler = _load_model()

    X = np.array([[features[col] for col in FEATURE_COLUMNS]], dtype=np.float32)
    X_scaled = scaler.transform(X)

    risk_score = float(model.predict_proba(X_scaled)[0][1])

    # Get feature importances
    importances = model.feature_importances_
    factor_weights = list(zip(FEATURE_COLUMNS, importances))
    factor_weights.sort(key=lambda x: x[1], reverse=True)

    top_factors = [
        ChurnFactor(factor=name, weight=round(float(weight), 4))
        for name, weight in factor_weights[:3]
    ]

    if risk_score >= 0.6:
        risk_level = "high"
    elif risk_score >= 0.3:
        risk_level = "medium"
    else:
        risk_level = "low"

    return risk_score, risk_level, top_factors


# ─── Save Prediction ────────────────────────────────────────────────────────

def _save_prediction(customer_id: str, tenant_id: str, risk_score: float,
                     risk_level: str, factors: list[ChurnFactor]) -> None:
    """Persist churn prediction to the churn_predictions table."""
    import json
    factors_json = json.dumps([f.model_dump() for f in factors])

    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO churn_predictions
                    (customer_id, tenant_id, risk_score, risk_level, factors, predicted_at)
                VALUES (%s, %s, %s, %s, %s, NOW())
                ON CONFLICT (customer_id, tenant_id)
                DO UPDATE SET
                    risk_score = EXCLUDED.risk_score,
                    risk_level = EXCLUDED.risk_level,
                    factors = EXCLUDED.factors,
                    predicted_at = NOW()
                """,
                (customer_id, tenant_id, risk_score, risk_level, factors_json),
            )


# ─── Public API ─────────────────────────────────────────────────────────────

def predict_churn(customer_id: str, tenant_id: str) -> ChurnPredictionResponse:
    """
    Predict churn for a single customer.

    Uses XGBoost if a trained model exists and the customer has sufficient data.
    Otherwise falls back to rule-based prediction.
    """
    features = _extract_customer_features(customer_id, tenant_id)
    if features is None:
        raise ValueError(f"Customer {customer_id} not found for tenant {tenant_id}")

    # Decide whether to use ML or rule-based
    use_ml = _has_trained_model() and features["subscription_days"] >= 30

    if use_ml:
        try:
            risk_score, risk_level, top_factors = _ml_predict(features)
        except Exception as e:
            logger.warning("ML prediction failed, falling back to rules: %s", e)
            risk_score, risk_level, top_factors = _rule_based_predict(features)
    else:
        risk_score, risk_level, top_factors = _rule_based_predict(features)

    # Save to database
    _save_prediction(customer_id, tenant_id, risk_score, risk_level, top_factors)

    return ChurnPredictionResponse(
        customer_id=customer_id,
        risk_score=round(risk_score, 4),
        risk_level=risk_level,
        factors=top_factors,
    )


def batch_predict(tenant_id: str) -> list[ChurnPredictionResponse]:
    """
    Run churn prediction for all customers of a tenant.
    Attempts to train/retrain the model if sufficient labeled data exists.
    """
    features_list = _extract_tenant_features(tenant_id)
    if not features_list:
        return []

    # Attempt to train model if we have enough data
    try:
        with get_cursor(dict_cursor=True) as cur:
            cur.execute(
                """
                SELECT COUNT(DISTINCT customer_id) AS cnt
                FROM user_events
                WHERE tenant_id = %s
                  AND event_type = 'subscription_cancelled'
                """,
                (tenant_id,),
            )
            row = cur.fetchone()
            churn_count = row["cnt"] if row else 0

        if churn_count >= 10 and len(features_list) >= 30:
            _train_model(features_list)
    except Exception as e:
        logger.warning("Model training failed: %s", e)

    # Run predictions
    results = []
    for features in features_list:
        use_ml = _has_trained_model() and features["subscription_days"] >= 30

        if use_ml:
            try:
                risk_score, risk_level, top_factors = _ml_predict(features)
            except Exception:
                risk_score, risk_level, top_factors = _rule_based_predict(features)
        else:
            risk_score, risk_level, top_factors = _rule_based_predict(features)

        _save_prediction(features["customer_id"], tenant_id, risk_score, risk_level, top_factors)

        results.append(
            ChurnPredictionResponse(
                customer_id=features["customer_id"],
                risk_score=round(risk_score, 4),
                risk_level=risk_level,
                factors=top_factors,
            )
        )

    logger.info("Batch churn prediction completed for tenant %s: %d customers", tenant_id, len(results))
    return results
