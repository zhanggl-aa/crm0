"""
Pydantic schemas for the algorithm service API.
"""

from typing import Optional
from pydantic import BaseModel, Field


# ─── Churn Prediction ───────────────────────────────────────────────────────

class ChurnFactor(BaseModel):
    factor: str = Field(..., description="Feature name contributing to churn risk")
    weight: float = Field(..., description="Importance weight of this factor")


class ChurnPredictionRequest(BaseModel):
    customer_id: str = Field(..., description="Customer ID to predict churn for")
    tenant_id: str = Field(..., description="Tenant ID for multi-tenancy")


class ChurnPredictionResponse(BaseModel):
    customer_id: str
    risk_score: float = Field(..., ge=0.0, le=1.0, description="Churn risk score 0-1")
    risk_level: str = Field(..., description="Risk level: high, medium, or low")
    factors: list[ChurnFactor] = Field(..., description="Top contributing factors")


class BatchChurnRequest(BaseModel):
    tenant_id: str = Field(..., description="Tenant ID for batch prediction")


# ─── Customer Segmentation ──────────────────────────────────────────────────

class SegmentationRequest(BaseModel):
    tenant_id: str = Field(..., description="Tenant ID")
    segment_type: str = Field(
        ..., description="Segmentation type: rfm, behavioral, or value"
    )


class SegmentationResponse(BaseModel):
    customer_id: str
    segment_name: str = Field(..., description="Assigned segment name")
    score: float = Field(..., description="Segment score or cluster assignment")


# ─── LTV Prediction ─────────────────────────────────────────────────────────

class LTVRequest(BaseModel):
    customer_id: str = Field(..., description="Customer ID to predict LTV for")
    tenant_id: str = Field(..., description="Tenant ID")


class LTVResponse(BaseModel):
    customer_id: str
    predicted_ltv: float = Field(..., description="Predicted lifetime value in dollars")
    confidence: float = Field(..., ge=0.0, le=1.0, description="Model confidence")
    expected_lifetime_months: float = Field(..., description="Expected remaining lifetime in months")


class ChannelROI(BaseModel):
    channel: str = Field(..., description="Acquisition channel name")
    cac: float = Field(..., description="Customer acquisition cost")
    predicted_ltv: float = Field(..., description="Average predicted LTV for this channel")
    ltv_cac_ratio: float = Field(..., description="LTV to CAC ratio")


# ─── Next Best Action ───────────────────────────────────────────────────────

class NBAAction(BaseModel):
    action_type: str = Field(..., description="Action type: call, discount, email, feature_guide")
    action_detail: str = Field(..., description="Detailed description of the action")
    expected_impact: float = Field(..., description="Expected impact score 0-1")
    priority: int = Field(..., ge=1, description="Priority ranking (1 = highest)")


class NBARequest(BaseModel):
    customer_id: str = Field(..., description="Customer ID")
    tenant_id: str = Field(..., description="Tenant ID")


class NBAResponse(BaseModel):
    customer_id: str
    actions: list[NBAAction] = Field(..., description="Recommended actions sorted by priority")


# ─── Task / Async ────────────────────────────────────────────────────────────

class TaskResponse(BaseModel):
    task_id: str = Field(..., description="Background task ID")
    status: str = Field(..., description="Task status: pending, running, completed, failed")
    result: Optional[dict] = Field(None, description="Task result when completed")
