"""
LTV prediction router.
"""

import logging
from fastapi import APIRouter, HTTPException

from pydantic import BaseModel, Field

from app.models.schemas import LTVRequest, LTVResponse, ChannelROI


class ChannelROIRequest(BaseModel):
    tenant_id: str = Field(..., description="Tenant ID for channel ROI analysis")


from app.services import ltv_service

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/ltv", tags=["LTV Prediction"])


@router.post("/predict", response_model=LTVResponse)
def predict_ltv(request: LTVRequest):
    """Predict lifetime value for a single customer."""
    try:
        result = ltv_service.predict_ltv(request.customer_id, request.tenant_id)
        return result
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        logger.error("LTV prediction failed: %s", e, exc_info=True)
        raise HTTPException(status_code=500, detail="LTV prediction failed")


@router.post("/channel-roi", response_model=list[ChannelROI])
def channel_roi(request: ChannelROIRequest):
    """Calculate ROI per acquisition channel for a tenant."""
    try:
        results = ltv_service.get_channel_roi(request.tenant_id)
        return results
    except Exception as e:
        logger.error("Channel ROI calculation failed: %s", e, exc_info=True)
        raise HTTPException(status_code=500, detail="Channel ROI calculation failed")
