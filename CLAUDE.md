# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CRM0 is a SaaS intelligent CRM system with three services: a Go backend API, a Vue 3 frontend, and a Python algorithm/ML microservice. All services share a PostgreSQL database (`crm0_db`).

## Architecture

```
frontend (Vue 3, :5173)  →  backend (Go, :3000)  →  algorithm (FastAPI, :8001)
         Vite proxy /api        REST + JWT              ML predictions
```

- **Backend** is the primary API gateway. It handles auth, CRUD, and calls the algorithm service for ML predictions.
- **Algorithm** service is stateless — it loads trained models from disk and queries PostgreSQL directly for feature data.
- **Frontend** proxies all `/api` requests to the backend via Vite dev server.

### Multi-tenancy

All domain entities are tenant-scoped. The `tenant_id` column is present on users, customers, plans, and all analytics tables. Repository queries always filter by `tenant_id`. JWT tokens carry the tenant ID.

### Backend (Go) — Layered Architecture

```
cmd/server/          →  Entry point
internal/config/     →  Env-based config (.env support via godotenv)
internal/model/      →  Domain entities + request/response DTOs
internal/repository/ →  Data access (PostgreSQL via lib/pq, connection pool max=25)
internal/service/    →  Business logic (handler ↔ repository orchestration)
internal/handler/    →  HTTP handlers
internal/middleware/  →  Auth, logging middleware
internal/algorithm/  →  Client for algorithm service
migrations/          →  SQL migrations (run via repository.RunMigrations)
```

Key models: Tenant, User, Customer, Plan, Subscription, UserEvent, ChurnPrediction, CustomerSegment, LTVPrediction, NBARecommendation.

Config defaults: DB on localhost:5432, JWT 72h expiry, algorithm service at localhost:8001, app port 3000.

### Frontend (Vue 3) — Composition API + Pinia

```
src/api/          →  Axios instances per domain (auth, customer, analytics, subscription)
src/stores/       →  Pinia stores (auth, customer, analytics)
src/router/       →  Vue Router with auth guards
src/views/        →  Page components (lazy-loaded)
src/components/   →  AppLayout (sidebar + header shell)
```

- UI framework: Element Plus with Chinese locale (zhCn)
- Charts: ECharts via vue-echarts
- API base: `/api/v1` — auth token injected via Axios request interceptor
- 401 responses trigger redirect to `/login`
- Auth state persisted in localStorage

### Algorithm (Python FastAPI) — ML Microservice

```
app/main.py       →  FastAPI app with CORS, health check, startup/shutdown DB pool
app/database.py   →  psycopg2 ThreadedConnectionPool (min=2, max=10)
app/routers/      →  /churn, /ltv, /segments, /nba
app/services/     →  Model loading + prediction logic
app/ml/models/    →  Trained model files (.pkl)
```

ML models:
- **Churn**: XGBoost + StandardScaler (features: login frequency, feature usage breadth, payment failures, support tickets, subscription days, MRR change rate)
- **LTV**: Cox Proportional Hazards (lifelines)
- **Segmentation**: K-Means/DBSCAN (behavioral), RFM scoring, value-based quantiles
- **NBA**: Rule-based engine combining churn risk + segment + subscription status

All services have cold-start fallbacks (rule-based) when trained models aren't available.

## Development Commands

### Backend (Go)
```bash
cd backend
go run cmd/server/main.go          # Start API server on :3000
go test ./...                       # Run all tests
go test ./internal/repository/...   # Run specific package tests
```

### Frontend (Vue)
```bash
cd frontend
npm install                         # Install dependencies
npm run dev                         # Dev server on :5173 (proxies /api → :3000)
npm run build                       # Type-check (vue-tsc) + production build
npm run preview                     # Preview production build
```

### Algorithm (Python)
```bash
cd algorithm
pip install -r requirements.txt     # Install dependencies
uvicorn app.main:app --port 8001 --reload  # Start on :8001
```

## Database

PostgreSQL with a single migration file at `backend/migrations/001_init.sql`. All tables use UUID primary keys (`gen_random_uuid()`). JSONB columns for flexible data (tags, custom_fields, settings, properties, factors). Migrations run automatically via `repository.RunMigrations()`.

## Environment Variables

Backend reads from `.env` file or environment (see `internal/config/config.go`). Key variables:

| Variable | Default | Purpose |
|---|---|---|
| DB_HOST | localhost | PostgreSQL host |
| DB_PORT | 5432 | PostgreSQL port |
| DB_USER | postgres | DB user |
| DB_PASSWORD | 123456 | DB password |
| DB_NAME | crm0_db | Database name |
| JWT_SECRET | crm0-secret-key | JWT signing key |
| JWT_EXPIRY_HOURS | 72 | Token lifetime |
| ALGORITHM_SERVICE_URL | http://localhost:8001 | ML service URL |
| REDIS_URL | localhost:6379 | Redis (for caching) |
| APP_PORT | 3000 | Backend listen port |
