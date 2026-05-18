"""
Next Best Action router.
"""

import logging
from fastapi import APIRouter, HTTPException

from app.models.schemas import NBARequest, NBAResponse
from app.services import nba_service

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/nba", tags=["Next Best Action"])


@router.post("/recommend", response_model=NBAResponse)
def recommend(request: NBARequest):
    """Generate next best action recommendations for a customer."""
    try:
        result = nba_service.recommend(request.customer_id, request.tenant_id)
        return result
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        logger.error("NBA recommendation failed: %s", e, exc_info=True)
        raise HTTPException(status_code=500, detail="NBA recommendation failed")
