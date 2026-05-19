# CRM0 Full Implementation Design

## Goal

Complete all missing layers in the CRM0 SaaS intelligent CRM: Go backend (handlers, services, middleware, algorithm client, entry point), and Vue 3 frontend (all analytics views, subscriptions, customer detail, settings). The Python algorithm service is already complete.

## Architecture

Three services sharing a PostgreSQL database:
- **Backend** (Go, :3000) — primary API gateway, handles auth, CRUD, proxies to algorithm service
- **Frontend** (Vue 3, :5173) — SPA with Vite proxy to backend
- **Algorithm** (FastAPI, :8001) — ML microservice (already done)

## Backend Design

### Routing: Go 1.24 stdlib `http.ServeMux`

Method+path routing with no external dependencies. Routes:
```
POST   /api/v1/auth/login
POST   /api/v1/auth/register
POST   /api/v1/auth/refresh

GET    /api/v1/customers
POST   /api/v1/customers
GET    /api/v1/customers/{id}
PUT    /api/v1/customers/{id}
DELETE /api/v1/customers/{id}
GET    /api/v1/customers/{id}/insights
GET    /api/v1/customers/{id}/events

GET    /api/v1/subscriptions
POST   /api/v1/subscriptions
GET    /api/v1/subscriptions/{id}
PUT    /api/v1/subscriptions/{id}
DELETE /api/v1/subscriptions/{id}
GET    /api/v1/subscriptions/metrics

GET    /api/v1/plans
POST   /api/v1/plans
GET    /api/v1/plans/{id}
PUT    /api/v1/plans/{id}
DELETE /api/v1/plans/{id}

GET    /api/v1/events

GET    /api/v1/analytics/churn/predictions
POST   /api/v1/analytics/churn/trigger-prediction
GET    /api/v1/analytics/segments
POST   /api/v1/analytics/segments/trigger-segmentation
GET    /api/v1/analytics/ltv/predictions
GET    /api/v1/analytics/ltv/channel-roi
GET    /api/v1/analytics/nba/recommendations
POST   /api/v1/analytics/nba/trigger-nba

GET    /api/v1/dashboard
```

### Service Layer

One service per domain, thin orchestration between handlers and repositories:

- **AuthService**: login (verify bcrypt hash, issue JWT), register (hash password, create tenant+user in transaction), refresh token
- **CustomerService**: CRUD, fetch insights by calling algorithm service, fetch events
- **SubscriptionService**: CRUD, compute metrics (MRR totals, churn rate, active count)
- **PlanService**: CRUD
- **AnalyticsService**: Proxy to algorithm service endpoints, store results back via algorithm_repo, fetch cached predictions
- **EventService**: Create events, list by customer, recent by tenant

### Handler Layer

One handler struct per domain with service injected via constructor. Each handler method:
1. Parse request (JSON body, path params, query params)
2. Call service method
3. Return JSON response with appropriate status code

### Middleware

- **AuthMiddleware**: Extract Bearer token from Authorization header, validate JWT, inject `tenant_id` and `user_id` into request context. Skip for `/api/v1/auth/*` routes.
- **CORSMiddleware**: Allow all origins (dev), standard methods and headers.
- **LoggingMiddleware**: Log method, path, status code, duration.

### Algorithm Client (`internal/algorithm/client.go`)

HTTP client calling the FastAPI service:
- `PredictChurn(tenantID, customerID)` → POST /churn/predict
- `PredictChurnBatch(tenantID)` → POST /churn/batch
- `PredictLTV(tenantID, customerID)` → POST /ltv/predict
- `GetChannelROI(tenantID)` → POST /ltv/channel-roi
- `RunSegmentation(tenantID, method)` → POST /segments/run
- `GetNBA(tenantID, customerID)` → POST /nba/recommend

### Entry Point (`cmd/server/main.go`)

1. Load config
2. Connect to PostgreSQL, run migrations
3. Instantiate repositories
4. Instantiate algorithm client
5. Instantiate services (inject repos + algorithm client)
6. Instantiate handlers (inject services)
7. Register routes on mux with middleware chain
8. Start server on `:APP_PORT`

### Tenant Repo Completion

Add Update, Delete, List methods to `tenant_repo.go`.

## Frontend Design

### Missing Views

All views use `<script setup>`, Element Plus components, ECharts via vue-echarts, and Pinia stores.

**ChurnView** (`views/ChurnView.vue`):
- Table: customer name, email, risk score (progress bar), risk level (el-tag color), top factors
- "Trigger Prediction" button (calls analytics store triggerChurn)
- Bar chart: risk level distribution

**SegmentsView** (`views/SegmentsView.vue`):
- Pie chart: segment distribution
- Customer table grouped by segment
- "Run Segmentation" button with type selector (RFM / behavioral / value)
- Segment type tabs

**LTVView** (`views/LTVView.vue`):
- Table: customer, predicted LTV, confidence, expected lifetime
- Bar chart: top customers by LTV
- Channel ROI section: table + bar chart (CAC vs LTV by channel)

**NBAView** (`views/NBAView.vue`):
- Recommendation cards with action type icons, priority, expected impact
- "Trigger NBA" button
- Filter by action type, status

**SubscriptionsView** (`views/SubscriptionsView.vue`):
- Subscription table with status badges, plan, MRR
- CRUD via el-dialog forms
- Metrics cards (total MRR, active count, trial count)

**CustomerDetailView** (`views/CustomerDetailView.vue`):
- Profile card (name, email, company, phone, status, tags)
- Insights panel: churn risk gauge, LTV value, segment badge, NBA recommendations
- Event timeline (el-timeline)
- Edit/delete actions

**SettingsView** (`views/SettingsView.vue`):
- Tenant info form (name)
- User profile form (name, email, password change)
- Read-only plan display

### No New Dependencies

All views use existing packages: element-plus, echarts, vue-echarts, pinia stores, axios api modules.

## Error Handling

- Backend: Service layer returns application errors, handlers map to HTTP status codes (400, 401, 403, 404, 500). Algorithm service errors return 502.
- Frontend: Axios interceptor already handles 401 redirect. Views use el-message for error feedback.

## Multi-Tenancy

All backend handlers extract tenant_id from JWT context. All repository queries filter by tenant_id. Algorithm service receives tenant_id as parameter.

## Data Flow — Analytics Example

1. Frontend: "Trigger Churn Prediction" button → `analyticsStore.triggerChurn()`
2. Backend: `POST /api/v1/analytics/churn/trigger-prediction` → `AnalyticsService.TriggerChurnPrediction()`
3. Service calls `algorithmClient.PredictChurnBatch(tenantID)`
4. Algorithm service queries DB for customer features, runs XGBoost model, returns predictions
5. Service stores predictions via `algorithm_repo.SaveChurnPredictions()`
6. Frontend: `analyticsStore.fetchChurn()` → `GET /api/v1/analytics/churn/predictions` → displays table + chart
