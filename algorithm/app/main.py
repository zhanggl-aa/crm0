"""
CRM Algorithm Service - FastAPI Application.

Provides ML-powered algorithms for SaaS CRM:
- Churn Prediction (XGBoost + rule-based fallback)
- Customer Segmentation (RFM, Behavioral, Value)
- LTV Prediction (Cox Proportional Hazards)
- Next Best Action (Rule-based recommendation engine)
"""

import logging

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.routers import churn, segments, ltv, nba

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger(__name__)

# Create FastAPI application
app = FastAPI(
    title="CRM Algorithm Service",
    description="ML-powered algorithms for SaaS CRM: churn prediction, segmentation, LTV, and NBA",
    version="1.0.0",
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(churn.router)
app.include_router(segments.router)
app.include_router(ltv.router)
app.include_router(nba.router)


@app.on_event("startup")
async def startup_event():
    """Log service startup and initialize resources."""
    logger.info("CRM Algorithm Service starting up...")
    # Pre-initialize the database connection pool
    from app.database import init_connection_pool
    try:
        init_connection_pool()
        logger.info("Database connection pool initialized.")
    except Exception as e:
        logger.warning("Database connection pool initialization failed: %s. "
                       "Connections will be established on first request.", e)
    logger.info("CRM Algorithm Service is ready.")


@app.on_event("shutdown")
async def shutdown_event():
    """Clean up resources on shutdown."""
    logger.info("CRM Algorithm Service shutting down...")
    from app.database import close_connection_pool
    close_connection_pool()
    logger.info("CRM Algorithm Service stopped.")


@app.get("/health")
def health_check():
    """Health check endpoint."""
    return {"status": "healthy", "service": "crm-algorithm"}
