"""
Churn prediction router.
"""

import logging
from fastapi import APIRouter, HTTPException

from app.models.schemas import (
    ChurnPredictionRequest,
    ChurnPredictionResponse,
    BatchChurnRequest,
)
from app.services import churn_service

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/churn", tags=["Churn Prediction"])


@router.post("/predict", response_model=ChurnPredictionResponse)
def predict_churn(request: ChurnPredictionRequest):
    """Predict churn risk for a single customer."""
    try:
        result = churn_service.predict_churn(request.customer_id, request.tenant_id)
        return result
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        logger.error("Churn prediction failed: %s", e, exc_info=True)
        raise HTTPException(status_code=500, detail="Churn prediction failed")


@router.post("/batch", response_model=list[ChurnPredictionResponse])
def batch_predict(request: BatchChurnRequest):
    """Run churn prediction for all customers of a tenant."""
    try:
        results = churn_service.batch_predict(request.tenant_id)
        return results
    except Exception as e:
        logger.error("Batch churn prediction failed: %s", e, exc_info=True)
        raise HTTPException(status_code=500, detail="Batch churn prediction failed")
