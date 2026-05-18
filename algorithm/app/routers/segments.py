"""
Customer segmentation router.
"""

import logging
from fastapi import APIRouter, HTTPException

from app.models.schemas import SegmentationRequest, SegmentationResponse
from app.services import segment_service

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/segments", tags=["Customer Segmentation"])


@router.post("/run", response_model=list[SegmentationResponse])
def run_segmentation(request: SegmentationRequest):
    """Run customer segmentation for a tenant."""
    try:
        results = segment_service.segment_customers(request.tenant_id, request.segment_type)
        return results
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        logger.error("Segmentation failed: %s", e, exc_info=True)
        raise HTTPException(status_code=500, detail="Segmentation failed")
