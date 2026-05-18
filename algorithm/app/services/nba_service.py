"""
Next Best Action (NBA) Recommendation Service.

Rule-based action selection based on churn risk, segment,
subscription status, and usage patterns.
"""

import logging
from datetime import datetime, timedelta

from app.database import get_cursor, get_connection
from app.models.schemas import NBAResponse, NBAAction

logger = logging.getLogger(__name__)


# ─── Data Gathering ─────────────────────────────────────────────────────────

def _get_customer_context(customer_id: str, tenant_id: str) -> dict | None:
    """Gather all context needed for NBA: churn risk, segment, subscription, usage."""
    with get_cursor(dict_cursor=True) as cur:
        # Customer basics
        cur.execute(
            """
            SELECT c.id, c.created_at,
                   s.status AS sub_status, s.mrr,
                   s.current_period_end, s.trial_end
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

        # Churn risk
        cur.execute(
            """
            SELECT risk_score, risk_level
            FROM churn_predictions
            WHERE customer_id = %s AND tenant_id = %s
            """,
            (customer_id, tenant_id),
        )
        churn = cur.fetchone()
        risk_score = float(churn["risk_score"]) if churn else 0.0
        risk_level = churn["risk_level"] if churn else "low"

        # Segment
        cur.execute(
            """
            SELECT segment_name, segment_type
            FROM customer_segments
            WHERE customer_id = %s AND tenant_id = %s
            ORDER BY segmented_at DESC LIMIT 1
            """,
            (customer_id, tenant_id),
        )
        segment_row = cur.fetchone()
        segment_name = segment_row["segment_name"] if segment_row else None

        # Usage in last 30 days
        thirty_days_ago = datetime.utcnow() - timedelta(days=30)
        cur.execute(
            """
            SELECT COUNT(*) AS event_count,
                   COUNT(DISTINCT event_type) AS feature_breadth
            FROM user_events
            WHERE customer_id = %s AND tenant_id = %s AND created_at >= %s
            """,
            (customer_id, tenant_id, thirty_days_ago),
        )
        usage = cur.fetchone()
        event_count = float(usage["event_count"] or 0) if usage else 0.0
        feature_breadth = float(usage["feature_breadth"] or 0) if usage else 0.0

        # Customer age in days
        created_at = customer.get("created_at")
        age_days = float((datetime.utcnow() - created_at).days) if created_at else 0.0

    return {
        "customer_id": customer_id,
        "tenant_id": tenant_id,
        "risk_score": risk_score,
        "risk_level": risk_level,
        "segment_name": segment_name,
        "sub_status": customer.get("sub_status", "unknown"),
        "mrr": float(customer.get("mrr") or 0),
        "current_period_end": customer.get("current_period_end"),
        "trial_end": customer.get("trial_end"),
        "event_count": event_count,
        "feature_breadth": feature_breadth,
        "age_days": age_days,
    }


# ─── Action Rules ───────────────────────────────────────────────────────────

def _is_trial_expiring_soon(ctx: dict) -> bool:
    """Check if the customer's trial is expiring within 7 days."""
    trial_end = ctx.get("trial_end")
    if not trial_end:
        return False
    days_left = (trial_end - datetime.utcnow()).days
    return 0 <= days_left <= 7


def _is_subscription_past_due(ctx: dict) -> bool:
    """Check if the customer's subscription is past due or unpaid."""
    return ctx.get("sub_status") in ("past_due", "unpaid", "incomplete")


def _is_new_customer(ctx: dict) -> bool:
    """Check if the customer is new (< 30 days old)."""
    return ctx["age_days"] < 30


def _is_low_usage(ctx: dict) -> bool:
    """Check if the customer has low usage (< 10 events in 30 days)."""
    return ctx["event_count"] < 10


def _generate_actions(ctx: dict) -> list[NBAAction]:
    """
    Generate recommended actions based on customer context using rule engine.
    Returns up to 3 prioritized actions.
    """
    actions: list[NBAAction] = []
    risk_level = ctx["risk_level"]
    sub_status = ctx.get("sub_status", "unknown")

    # Rule 1: High churn risk + active subscription -> personal outreach (call)
    if risk_level == "high" and sub_status in ("active", "trialing"):
        actions.append(NBAAction(
            action_type="call",
            action_detail="Schedule personal outreach call to discuss concerns and prevent churn. "
                          "Customer shows high churn risk signals.",
            expected_impact=0.85,
            priority=1,
        ))

    # Rule 2: High churn risk + past due -> win-back discount
    if risk_level == "high" and _is_subscription_past_due(ctx):
        actions.append(NBAAction(
            action_type="discount",
            action_detail="Offer a win-back discount (20-30% off) to recover the past-due "
                          "subscription and prevent permanent churn.",
            expected_impact=0.75,
            priority=1,
        ))

    # Rule 3: Trial about to expire -> conversion discount
    if _is_trial_expiring_soon(ctx):
        actions.append(NBAAction(
            action_type="discount",
            action_detail="Send a time-limited conversion offer before the trial expires. "
                          "Emphasize value already realized during the trial period.",
            expected_impact=0.70,
            priority=1,
        ))

    # Rule 4: Low usage + new customer -> feature guide (onboarding)
    if _is_low_usage(ctx) and _is_new_customer(ctx):
        actions.append(NBAAction(
            action_type="feature_guide",
            action_detail="Send personalized onboarding guide highlighting key features. "
                          "Schedule a product walkthrough to drive early engagement.",
            expected_impact=0.65,
            priority=2,
        ))

    # Rule 5: Low usage + established customer -> re-engagement email
    if _is_low_usage(ctx) and not _is_new_customer(ctx):
        actions.append(NBAAction(
            action_type="email",
            action_detail="Send a re-engagement email with tips on underutilized features "
                          "and recent product updates relevant to their usage history.",
            expected_impact=0.55,
            priority=2,
        ))

    # Rule 6: Medium churn risk -> proactive check-in email
    if risk_level == "medium" and sub_status == "active":
        actions.append(NBAAction(
            action_type="email",
            action_detail="Send a proactive check-in email to understand satisfaction level. "
                          "Include a quick NPS survey and offer to schedule a review call.",
            expected_impact=0.50,
            priority=2,
        ))

    # Rule 7: High-value segment -> upsell opportunity
    segment = ctx.get("segment_name") or ""
    if "platinum" in segment.lower() or "champion" in segment.lower():
        actions.append(NBAAction(
            action_type="email",
            action_detail="Present premium add-on or upgrade options tailored to their "
                          "high-value usage patterns.",
            expected_impact=0.60,
            priority=3,
        ))

    # Rule 8: Low feature breadth -> feature education
    if ctx["feature_breadth"] < 3 and ctx["event_count"] >= 10:
        actions.append(NBAAction(
            action_type="feature_guide",
            action_detail="Recommend features the customer has not yet tried based on "
                          "their usage pattern and similar customers' behavior.",
            expected_impact=0.45,
            priority=3,
        ))

    # Default: if no specific rules match, add a general engagement action
    if not actions:
        actions.append(NBAAction(
            action_type="email",
            action_detail="Send monthly product update and best practices newsletter "
                          "to maintain engagement.",
            expected_impact=0.30,
            priority=1,
        ))

    # Sort by priority, then by expected_impact descending
    actions.sort(key=lambda a: (a.priority, -a.expected_impact))

    # Return top 3
    return actions[:3]


# ─── Save Results ───────────────────────────────────────────────────────────

def _save_recommendations(customer_id: str, tenant_id: str,
                          actions: list[NBAAction]) -> None:
    """Persist NBA recommendations to the nba_recommendations table."""
    import json
    actions_json = json.dumps([a.model_dump() for a in actions])

    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO nba_recommendations
                    (customer_id, tenant_id, actions, recommended_at)
                VALUES (%s, %s, %s, NOW())
                ON CONFLICT (customer_id, tenant_id)
                DO UPDATE SET
                    actions = EXCLUDED.actions,
                    recommended_at = NOW()
                """,
                (customer_id, tenant_id, actions_json),
            )


# ─── Public API ─────────────────────────────────────────────────────────────

def recommend(customer_id: str, tenant_id: str) -> NBAResponse:
    """
    Generate next best action recommendations for a customer.

    Uses a rule engine that considers:
    - Churn risk level
    - Subscription status (active, past_due, trialing)
    - Usage patterns (event count, feature breadth)
    - Customer age (new vs established)
    - Customer segment (value tier)
    - Trial expiration

    Returns the top 3 recommended actions with expected impact and priority.
    """
    ctx = _get_customer_context(customer_id, tenant_id)
    if ctx is None:
        raise ValueError(f"Customer {customer_id} not found for tenant {tenant_id}")

    actions = _generate_actions(ctx)
    _save_recommendations(customer_id, tenant_id, actions)

    logger.info(
        "NBA recommendations generated: customer=%s, actions=%d, top=%s",
        customer_id, len(actions), actions[0].action_type if actions else "none",
    )

    return NBAResponse(
        customer_id=customer_id,
        actions=actions,
    )
