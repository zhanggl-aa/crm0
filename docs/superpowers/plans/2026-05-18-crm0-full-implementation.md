# CRM0 Full Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement all missing backend (Go) layers, frontend (Vue 3) views, and wire them together to produce a working SaaS CRM.

**Architecture:** Go 1.24 stdlib HTTP server with layered architecture (handler → service → repository). Vue 3 SPA with Pinia stores, Element Plus UI, ECharts. FastAPI algorithm service already complete.

**Tech Stack:** Go 1.24, PostgreSQL, JWT (golang-jwt), bcrypt; Vue 3, TypeScript, Vite, Pinia, Element Plus, ECharts, Axios

---

## File Structure

### Backend — New Files
```
backend/
├── cmd/server/main.go                         # Entry point
├── internal/
│   ├── middleware/
│   │   ├── auth.go                            # JWT auth middleware
│   │   └── cors.go                            # CORS + logging middleware
│   ├── service/
│   │   ├── auth_service.go                    # Login, register, refresh
│   │   ├── customer_service.go                # Customer CRUD + insights
│   │   ├── subscription_service.go            # Subscription CRUD + metrics
│   │   ├── plan_service.go                    # Plan CRUD
│   │   ├── event_service.go                   # Event tracking
│   │   └── analytics_service.go               # Analytics proxy to algorithm svc
│   ├── handler/
│   │   ├── auth_handler.go                    # Auth HTTP handlers
│   │   ├── customer_handler.go                # Customer HTTP handlers
│   │   ├── subscription_handler.go            # Subscription HTTP handlers
│   │   ├── plan_handler.go                    # Plan HTTP handlers
│   │   ├── event_handler.go                   # Event HTTP handlers
│   │   ├── analytics_handler.go               # Analytics HTTP handlers
│   │   └── response.go                        # JSON response helpers
│   └── algorithm/
│       └── client.go                          # HTTP client for algorithm service
├── go.mod                                     # Updated: add golang-jwt + bcrypt
└── go.sum
```

### Backend — Modified Files
```
backend/internal/repository/tenant_repo.go     # Add Update, Delete, List
```

### Frontend — New Files
```
frontend/src/
├── views/
│   ├── Customers.vue                          # Customer list
│   ├── CustomerDetail.vue                     # Customer detail + insights
│   ├── Subscriptions.vue                      # Subscription management
│   ├── Settings.vue                           # Tenant + user settings
│   └── analytics/
│       ├── ChurnAnalysis.vue                  # Churn predictions
│       ├── SegmentAnalysis.vue                # Customer segments
│       ├── LTVAnalysis.vue                    # LTV + channel ROI
│       └── NBARecommendations.vue             # Next best action
```

---

## Task 1: Add JWT and Bcrypt Dependencies

**Files:**
- Modify: `backend/go.mod`

- [ ] **Step 1: Add dependencies**

```bash
cd backend
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
```

- [ ] **Step 2: Verify dependencies resolve**

```bash
go mod tidy
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "feat: add jwt and bcrypt dependencies"
```

---

## Task 2: JSON Response Helpers

**Files:**
- Create: `backend/internal/handler/response.go`

- [ ] **Step 1: Create response helpers**

```go
package handler

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && go build ./internal/handler/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/handler/response.go
git commit -m "feat: add JSON response helpers"
```

---

## Task 3: Auth Middleware

**Files:**
- Create: `backend/internal/middleware/auth.go`

- [ ] **Step 1: Create auth middleware**

```go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	TenantIDKey contextKey = "tenant_id"
	UserIDKey   contextKey = "user_id"
)

func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == authHeader {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"invalid claims"}`, http.StatusUnauthorized)
				return
			}

			tenantID, err := uuid.Parse(claims["tenant_id"].(string))
			if err != nil {
				http.Error(w, `{"error":"invalid tenant_id in token"}`, http.StatusUnauthorized)
				return
			}
			userID, err := uuid.Parse(claims["user_id"].(string))
			if err != nil {
				http.Error(w, `{"error":"invalid user_id in token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
			ctx = context.WithValue(ctx, UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetTenantID(ctx context.Context) uuid.UUID {
	return ctx.Value(TenantIDKey).(uuid.UUID)
}

func GetUserID(ctx context.Context) uuid.UUID {
	return ctx.Value(UserIDKey).(uuid.UUID)
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && go build ./internal/middleware/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/middleware/auth.go
git commit -m "feat: add JWT auth middleware"
```

---

## Task 4: CORS and Logging Middleware

**Files:**
- Create: `backend/internal/middleware/cors.go`

- [ ] **Step 1: Create CORS and logging middleware**

```go
package middleware

import (
	"log"
	"net/http"
	"time"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && go build ./internal/middleware/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/middleware/cors.go
git commit -m "feat: add CORS and logging middleware"
```

---

## Task 5: Algorithm Service Client

**Files:**
- Create: `backend/internal/algorithm/client.go`

- [ ] **Step 1: Create algorithm HTTP client**

```go
package algorithm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) post(path string, body any) (map[string]any, error) {
	data, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("algorithm service unreachable: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("algorithm service error: %s", string(raw))
	}

	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) PredictChurn(tenantID, customerID uuid.UUID) (map[string]any, error) {
	return c.post("/churn/predict", map[string]string{
		"tenant_id":   tenantID.String(),
		"customer_id": customerID.String(),
	})
}

func (c *Client) PredictChurnBatch(tenantID uuid.UUID) (map[string]any, error) {
	return c.post("/churn/batch", map[string]string{
		"tenant_id": tenantID.String(),
	})
}

func (c *Client) PredictLTV(tenantID, customerID uuid.UUID) (map[string]any, error) {
	return c.post("/ltv/predict", map[string]string{
		"tenant_id":   tenantID.String(),
		"customer_id": customerID.String(),
	})
}

func (c *Client) GetChannelROI(tenantID uuid.UUID) (map[string]any, error) {
	return c.post("/ltv/channel-roi", map[string]string{
		"tenant_id": tenantID.String(),
	})
}

func (c *Client) RunSegmentation(tenantID uuid.UUID, method string) (map[string]any, error) {
	return c.post("/segments/run", map[string]string{
		"tenant_id": tenantID.String(),
		"method":    method,
	})
}

func (c *Client) GetNBA(tenantID, customerID uuid.UUID) (map[string]any, error) {
	return c.post("/nba/recommend", map[string]string{
		"tenant_id":   tenantID.String(),
		"customer_id": customerID.String(),
	})
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && go build ./internal/algorithm/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/algorithm/client.go
git commit -m "feat: add algorithm service HTTP client"
```

---

## Task 6: Complete Tenant Repository

**Files:**
- Modify: `backend/internal/repository/tenant_repo.go`

- [ ] **Step 1: Add Update, Delete, List methods to TenantRepository**

Add these methods to the end of `tenant_repo.go`:

```go
// Update modifies an existing tenant row.
func (r *TenantRepository) Update(ctx context.Context, tenant *model.Tenant) error {
	query := `UPDATE tenants SET name = $1, plan = $2, settings = $3, updated_at = NOW() WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, tenant.Name, tenant.Plan, tenant.Settings, tenant.ID)
	return err
}

// Delete removes a tenant by its primary key.
func (r *TenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tenants WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// List returns all tenants with pagination.
func (r *TenantRepository) List(ctx context.Context, page, pageSize int) ([]*model.Tenant, error) {
	offset := (page - 1) * pageSize
	query := `SELECT id, name, plan, settings, created_at, updated_at FROM tenants ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*model.Tenant
	for rows.Next() {
		t := &model.Tenant{}
		if err := rows.Scan(&t.ID, &t.Name, &t.Plan, &t.Settings, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && go build ./internal/repository/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/repository/tenant_repo.go
git commit -m "feat: add Update, Delete, List to TenantRepository"
```

---

## Task 7: Auth Service

**Files:**
- Create: `backend/internal/service/auth_service.go`

- [ ] **Step 1: Add GetByEmailGlobal to UserRepository**

Add to `backend/internal/repository/user_repo.go`:

```go
// GetByEmailGlobal retrieves a user by email across all tenants.
func (r *UserRepository) GetByEmailGlobal(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, tenant_id, email, password_hash, name, role, created_at, updated_at FROM users WHERE email = $1`
	u := &model.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}
```

- [ ] **Step 2: Create auth_service.go**

```go
package service

import (
	"context"
	"errors"
	"time"

	"crm0/backend/internal/config"
	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	tenantRepo *repository.TenantRepository
	cfg        *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, tenantRepo *repository.TenantRepository, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, tenantRepo: tenantRepo, cfg: cfg}
}

func (s *AuthService) Login(email, password string) (*model.AuthResponse, error) {
	ctx := context.Background()

	user, err := s.userRepo.GetByEmailGlobal(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *user,
	}, nil
}

func (s *AuthService) Register(tenantName, email, password, name string) (*model.AuthResponse, error) {
	ctx := context.Background()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	tenant := &model.Tenant{
		ID:   uuid.New(),
		Name: tenantName,
		Plan: "free",
	}
	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	user := &model.User{
		ID:           uuid.New(),
		TenantID:     tenant.ID,
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         "admin",
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *user,
	}, nil
}

func (s *AuthService) RefreshToken(userID uuid.UUID) (*model.AuthResponse, error) {
	ctx := context.Background()
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *user,
	}, nil
}

func (s *AuthService) generateToken(user *model.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.cfg.JWTExpiryHours) * time.Hour)
	claims := jwt.MapClaims{
		"user_id":   user.ID.String(),
		"tenant_id": user.TenantID.String(),
		"email":     user.Email,
		"role":      user.Role,
		"exp":       expiresAt.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenStr, expiresAt, nil
}
```

- [ ] **Step 3: Verify it compiles**

```bash
cd backend && go build ./internal/service/... ./internal/repository/...
```

- [ ] **Step 4: Commit**

```bash
git add internal/service/auth_service.go internal/repository/user_repo.go
git commit -m "feat: add auth service with login, register, refresh"
```

---

## Task 8: Customer, Plan, Event, Subscription, Analytics Services

**Files:**
- Create: `backend/internal/service/customer_service.go`
- Create: `backend/internal/service/plan_service.go`
- Create: `backend/internal/service/event_service.go`
- Create: `backend/internal/service/subscription_service.go`
- Create: `backend/internal/service/analytics_service.go`

- [ ] **Step 1: Create customer_service.go**

```go
package service

import (
	"context"
	"encoding/json"

	"crm0/backend/internal/algorithm"
	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/google/uuid"
)

type CustomerService struct {
	repo      *repository.CustomerRepository
	eventRepo *repository.EventRepository
	algoRepo  *repository.AlgorithmRepository
	algo      *algorithm.Client
}

func NewCustomerService(repo *repository.CustomerRepository, eventRepo *repository.EventRepository, algoRepo *repository.AlgorithmRepository, algo *algorithm.Client) *CustomerService {
	return &CustomerService{repo: repo, eventRepo: eventRepo, algoRepo: algoRepo, algo: algo}
}

func (s *CustomerService) List(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*model.Customer, int, error) {
	return s.repo.List(ctx, tenantID, page, pageSize)
}

func (s *CustomerService) Get(ctx context.Context, tenantID, id uuid.UUID) (*model.Customer, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *CustomerService) Create(ctx context.Context, tenantID uuid.UUID, req *model.CreateCustomerRequest) (*model.Customer, error) {
	c := &model.Customer{
		ID:              uuid.New(),
		TenantID:        tenantID,
		Name:            req.Name,
		Email:           req.Email,
		Company:         req.Company,
		Phone:           req.Phone,
		Status:          req.Status,
		Tags:            req.Tags,
		CustomFields:    req.CustomFields,
		AcquiredChannel: req.AcquiredChannel,
	}
	if c.Status == "" {
		c.Status = "active"
	}
	if c.Tags == nil {
		c.Tags = json.RawMessage(`[]`)
	}
	if c.CustomFields == nil {
		c.CustomFields = json.RawMessage(`{}`)
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CustomerService) Update(ctx context.Context, tenantID, id uuid.UUID, req *model.UpdateCustomerRequest) (*model.Customer, error) {
	c, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Email != nil {
		c.Email = *req.Email
	}
	if req.Company != nil {
		c.Company = *req.Company
	}
	if req.Phone != nil {
		c.Phone = *req.Phone
	}
	if req.Status != nil {
		c.Status = *req.Status
	}
	if req.Tags != nil {
		c.Tags = req.Tags
	}
	if req.CustomFields != nil {
		c.CustomFields = req.CustomFields
	}
	if req.AcquiredChannel != nil {
		c.AcquiredChannel = *req.AcquiredChannel
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CustomerService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.Delete(ctx, tenantID, id)
}

func (s *CustomerService) GetInsights(ctx context.Context, tenantID, customerID uuid.UUID) (*model.CustomerInsightResponse, error) {
	customer, err := s.repo.GetByID(ctx, tenantID, customerID)
	if err != nil {
		return nil, err
	}

	insight := &model.CustomerInsightResponse{Customer: *customer}

	churn, _ := s.algoRepo.GetChurnPredictionByCustomer(ctx, customerID)
	insight.ChurnPrediction = churn

	ltvs, _ := s.algoRepo.GetLTVPredictions(ctx, tenantID)
	for _, ltv := range ltvs {
		if ltv.CustomerID == customerID {
			insight.LTV = ltv
			break
		}
	}

	segments, _ := s.algoRepo.GetCustomerSegments(ctx, tenantID)
	for _, seg := range segments {
		if seg.CustomerID == customerID {
			insight.Segments = append(insight.Segments, *seg)
		}
	}

	recs, _ := s.algoRepo.GetNBARecommendations(ctx, tenantID)
	for _, rec := range recs {
		if rec.CustomerID == customerID {
			insight.Recommendations = append(insight.Recommendations, *rec)
		}
	}

	return insight, nil
}

func (s *CustomerService) GetEvents(ctx context.Context, customerID uuid.UUID, page, pageSize int) ([]*model.UserEvent, int, error) {
	return s.eventRepo.ListByCustomer(ctx, customerID, page, pageSize)
}
```

- [ ] **Step 2: Create plan_service.go**

```go
package service

import (
	"context"

	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/google/uuid"
)

type PlanService struct {
	repo *repository.PlanRepository
}

func NewPlanService(repo *repository.PlanRepository) *PlanService {
	return &PlanService{repo: repo}
}

func (s *PlanService) List(ctx context.Context, tenantID uuid.UUID) ([]*model.Plan, error) {
	return s.repo.ListByTenant(ctx, tenantID)
}

func (s *PlanService) Get(ctx context.Context, id uuid.UUID) (*model.Plan, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PlanService) Create(ctx context.Context, tenantID uuid.UUID, req *model.CreatePlanRequest) (*model.Plan, error) {
	p := &model.Plan{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Name:         req.Name,
		Price:        req.Price,
		BillingCycle: req.BillingCycle,
	}
	if p.BillingCycle == "" {
		p.BillingCycle = "monthly"
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PlanService) Update(ctx context.Context, id uuid.UUID, req *model.UpdatePlanRequest) (*model.Plan, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.BillingCycle != nil {
		p.BillingCycle = *req.BillingCycle
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PlanService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
```

- [ ] **Step 3: Create event_service.go**

```go
package service

import (
	"context"
	"encoding/json"
	"time"

	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/google/uuid"
)

type EventService struct {
	repo *repository.EventRepository
}

func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) Create(ctx context.Context, req *model.TrackEventRequest) (*model.UserEvent, error) {
	e := &model.UserEvent{
		ID:         uuid.New(),
		CustomerID: req.CustomerID,
		EventType:  req.EventType,
		Properties: req.Properties,
	}
	if req.OccurredAt != nil {
		e.OccurredAt = *req.OccurredAt
	} else {
		e.OccurredAt = time.Now()
	}
	if e.Properties == nil {
		e.Properties = json.RawMessage(`{}`)
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *EventService) ListByCustomer(ctx context.Context, customerID uuid.UUID, page, pageSize int) ([]*model.UserEvent, int, error) {
	return s.repo.ListByCustomer(ctx, customerID, page, pageSize)
}

func (s *EventService) GetRecentByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]*model.UserEvent, error) {
	return s.repo.GetRecentByTenant(ctx, tenantID, limit)
}
```

- [ ] **Step 4: Create subscription_service.go**

```go
package service

import (
	"context"

	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/google/uuid"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

func NewSubscriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) List(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*model.Subscription, int, error) {
	return s.repo.ListByTenant(ctx, tenantID, page, pageSize)
}

func (s *SubscriptionService) Get(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) Create(ctx context.Context, req *model.CreateSubscriptionRequest) (*model.Subscription, error) {
	sub := &model.Subscription{
		ID:         uuid.New(),
		CustomerID: req.CustomerID,
		PlanID:     req.PlanID,
		Status:     "trial",
	}
	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *SubscriptionService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateSubscriptionRequest) (*model.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Status != nil {
		sub.Status = *req.Status
	}
	if req.PlanID != nil {
		sub.PlanID = *req.PlanID
	}
	if req.CanceledAt != nil {
		sub.CanceledAt = req.CanceledAt
	}
	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	sub.Status = "canceled"
	return s.repo.Update(ctx, sub)
}

func (s *SubscriptionService) GetMetrics(ctx context.Context, tenantID uuid.UUID) (*model.SubscriptionMetrics, error) {
	totalMRR, activeCount, churnRate, err := s.repo.GetMetrics(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return &model.SubscriptionMetrics{
		TotalMRR:    totalMRR,
		ActiveCount: activeCount,
		ChurnRate:   churnRate,
	}, nil
}
```

- [ ] **Step 5: Create analytics_service.go**

```go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"crm0/backend/internal/algorithm"
	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/google/uuid"
)

type AnalyticsService struct {
	algoRepo *repository.AlgorithmRepository
	algo     *algorithm.Client
}

func NewAnalyticsService(algoRepo *repository.AlgorithmRepository, algo *algorithm.Client) *AnalyticsService {
	return &AnalyticsService{algoRepo: algoRepo, algo: algo}
}

func (s *AnalyticsService) GetChurnPredictions(ctx context.Context, tenantID uuid.UUID) ([]*model.ChurnPrediction, error) {
	return s.algoRepo.GetChurnPredictions(ctx, tenantID)
}

func (s *AnalyticsService) TriggerChurnPrediction(ctx context.Context, tenantID uuid.UUID) error {
	result, err := s.algo.PredictChurnBatch(tenantID)
	if err != nil {
		return fmt.Errorf("churn prediction failed: %w", err)
	}
	predictions, ok := result["predictions"].([]any)
	if !ok {
		return nil
	}
	for _, p := range predictions {
		pm := p.(map[string]any)
		customerID, _ := uuid.Parse(pm["customer_id"].(string))
		factors, _ := json.Marshal(pm["factors"])
		pred := &model.ChurnPrediction{
			ID:          uuid.New(),
			CustomerID:  customerID,
			RiskScore:   pm["risk_score"].(float64),
			RiskLevel:   pm["risk_level"].(string),
			Factors:     factors,
			PredictedAt: time.Now(),
		}
		s.algoRepo.SaveChurnPrediction(ctx, pred)
	}
	return nil
}

func (s *AnalyticsService) GetSegments(ctx context.Context, tenantID uuid.UUID) ([]*model.CustomerSegment, error) {
	return s.algoRepo.GetCustomerSegments(ctx, tenantID)
}

func (s *AnalyticsService) TriggerSegmentation(ctx context.Context, tenantID uuid.UUID, method string) error {
	result, err := s.algo.RunSegmentation(tenantID, method)
	if err != nil {
		return fmt.Errorf("segmentation failed: %w", err)
	}
	segments, ok := result["segments"].([]any)
	if !ok {
		return nil
	}
	for _, seg := range segments {
		sm := seg.(map[string]any)
		customerID, _ := uuid.Parse(sm["customer_id"].(string))
		s := &model.CustomerSegment{
			ID:          uuid.New(),
			CustomerID:  customerID,
			SegmentType: sm["segment_type"].(string),
			SegmentName: sm["segment_name"].(string),
			Score:       sm["score"].(float64),
			UpdatedAt:   time.Now(),
		}
		s.algoRepo.SaveCustomerSegment(ctx, s)
	}
	return nil
}

func (s *AnalyticsService) GetLTVPredictions(ctx context.Context, tenantID uuid.UUID) ([]*model.LTVPrediction, error) {
	return s.algoRepo.GetLTVPredictions(ctx, tenantID)
}

func (s *AnalyticsService) GetChannelROI(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	return s.algo.GetChannelROI(tenantID)
}

func (s *AnalyticsService) GetNBARecommendations(ctx context.Context, tenantID uuid.UUID) ([]*model.NBARecommendation, error) {
	return s.algoRepo.GetNBARecommendations(ctx, tenantID)
}

func (s *AnalyticsService) TriggerNBA(ctx context.Context, tenantID, customerID uuid.UUID) error {
	result, err := s.algo.GetNBA(tenantID, customerID)
	if err != nil {
		return fmt.Errorf("NBA failed: %w", err)
	}
	actions, ok := result["actions"].([]any)
	if !ok {
		return nil
	}
	for _, a := range actions {
		am := a.(map[string]any)
		custID, _ := uuid.Parse(am["customer_id"].(string))
		detail, _ := json.Marshal(am["action_detail"])
		rec := &model.NBARecommendation{
			ID:             uuid.New(),
			CustomerID:     custID,
			ActionType:     am["action_type"].(string),
			ActionDetail:   detail,
			ExpectedImpact: am["expected_impact"].(float64),
			Priority:       int(am["priority"].(float64)),
			Status:         "pending",
			CreatedAt:      time.Now(),
		}
		s.algoRepo.SaveNBARecommendation(ctx, rec)
	}
	return nil
}

func (s *AnalyticsService) GetDashboard(ctx context.Context, tenantID uuid.UUID) (*model.DashboardOverview, error) {
	return s.algoRepo.GetDashboardOverview(ctx, tenantID)
}
```

- [ ] **Step 6: Verify all services compile**

```bash
cd backend && go build ./internal/service/...
```

- [ ] **Step 7: Commit**

```bash
git add internal/service/
git commit -m "feat: add customer, plan, event, subscription, analytics services"
```

---

## Task 9: HTTP Handlers

**Files:**
- Create: `backend/internal/handler/auth_handler.go`
- Create: `backend/internal/handler/customer_handler.go`
- Create: `backend/internal/handler/subscription_handler.go`
- Create: `backend/internal/handler/plan_handler.go`
- Create: `backend/internal/handler/event_handler.go`
- Create: `backend/internal/handler/analytics_handler.go`

- [ ] **Step 1: Create auth_handler.go**

```go
package handler

import (
	"encoding/json"
	"net/http"

	"crm0/backend/internal/middleware"
	"crm0/backend/internal/model"
	"crm0/backend/internal/service"

	"github.com/google/uuid"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.svc.Login(req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.svc.Register(req.TenantName, req.Email, req.Password, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	resp, err := h.svc.RefreshToken(userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
```

- [ ] **Step 2: Create customer_handler.go**

```go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"crm0/backend/internal/middleware"
	"crm0/backend/internal/model"
	"crm0/backend/internal/service"

	"github.com/google/uuid"
)

type CustomerHandler struct {
	svc *service.CustomerService
}

func NewCustomerHandler(svc *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	customers, total, err := h.svc.List(r.Context(), tenantID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":     customers,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *CustomerHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer id")
		return
	}
	customer, err := h.svc.Get(r.Context(), tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}
	writeJSON(w, http.StatusOK, customer)
}

func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var req model.CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	customer, err := h.svc.Create(r.Context(), tenantID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, customer)
}

func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer id")
		return
	}
	var req model.UpdateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	customer, err := h.svc.Update(r.Context(), tenantID, id, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, customer)
}

func (h *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer id")
		return
	}
	if err := h.svc.Delete(r.Context(), tenantID, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (h *CustomerHandler) GetInsights(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer id")
		return
	}
	insight, err := h.svc.GetInsights(r.Context(), tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}
	writeJSON(w, http.StatusOK, insight)
}

func (h *CustomerHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer id")
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}
	events, total, err := h.svc.GetEvents(r.Context(), id, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":     events,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
```

- [ ] **Step 3: Create subscription_handler.go**

```go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"crm0/backend/internal/middleware"
	"crm0/backend/internal/model"
	"crm0/backend/internal/service"

	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	svc *service.SubscriptionService
}

func NewSubscriptionHandler(svc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}
	subs, total, err := h.svc.List(r.Context(), tenantID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":     subs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *SubscriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}
	sub, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "subscription not found")
		return
	}
	writeJSON(w, http.StatusOK, sub)
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sub, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, sub)
}

func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}
	var req model.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sub, err := h.svc.Update(r.Context(), id, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sub)
}

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (h *SubscriptionHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	metrics, err := h.svc.GetMetrics(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, metrics)
}
```

- [ ] **Step 4: Create plan_handler.go**

```go
package handler

import (
	"encoding/json"
	"net/http"

	"crm0/backend/internal/middleware"
	"crm0/backend/internal/model"
	"crm0/backend/internal/service"

	"github.com/google/uuid"
)

type PlanHandler struct {
	svc *service.PlanService
}

func NewPlanHandler(svc *service.PlanService) *PlanHandler {
	return &PlanHandler{svc: svc}
}

func (h *PlanHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	plans, err := h.svc.List(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, plans)
}

func (h *PlanHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid plan id")
		return
	}
	plan, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "plan not found")
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

func (h *PlanHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var req model.CreatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	plan, err := h.svc.Create(r.Context(), tenantID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, plan)
}

func (h *PlanHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid plan id")
		return
	}
	var req model.UpdatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	plan, err := h.svc.Update(r.Context(), id, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

func (h *PlanHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid plan id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}
```

- [ ] **Step 5: Create event_handler.go**

```go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"crm0/backend/internal/middleware"
	"crm0/backend/internal/model"
	"crm0/backend/internal/service"
)

type EventHandler struct {
	svc *service.EventService
}

func NewEventHandler(svc *service.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.TrackEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	event, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, event)
}

func (h *EventHandler) GetRecent(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}
	events, err := h.svc.GetRecentByTenant(r.Context(), tenantID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, events)
}
```

- [ ] **Step 6: Create analytics_handler.go**

```go
package handler

import (
	"net/http"

	"crm0/backend/internal/middleware"
	"crm0/backend/internal/service"

	"github.com/google/uuid"
)

type AnalyticsHandler struct {
	svc *service.AnalyticsService
}

func NewAnalyticsHandler(svc *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

func (h *AnalyticsHandler) GetChurnPredictions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	predictions, err := h.svc.GetChurnPredictions(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, predictions)
}

func (h *AnalyticsHandler) TriggerChurnPrediction(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if err := h.svc.TriggerChurnPrediction(r.Context(), tenantID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (h *AnalyticsHandler) GetSegments(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	segments, err := h.svc.GetSegments(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, segments)
}

func (h *AnalyticsHandler) TriggerSegmentation(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	method := r.URL.Query().Get("method")
	if method == "" {
		method = "rfm"
	}
	if err := h.svc.TriggerSegmentation(r.Context(), tenantID, method); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (h *AnalyticsHandler) GetLTVPredictions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	predictions, err := h.svc.GetLTVPredictions(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, predictions)
}

func (h *AnalyticsHandler) GetChannelROI(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	roi, err := h.svc.GetChannelROI(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, roi)
}

func (h *AnalyticsHandler) GetNBARecommendations(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	recs, err := h.svc.GetNBARecommendations(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, recs)
}

func (h *AnalyticsHandler) TriggerNBA(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	customerIDStr := r.URL.Query().Get("customer_id")
	var customerID uuid.UUID
	if customerIDStr != "" {
		var err error
		customerID, err = uuid.Parse(customerIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid customer_id")
			return
		}
	}
	if err := h.svc.TriggerNBA(r.Context(), tenantID, customerID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (h *AnalyticsHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	dashboard, err := h.svc.GetDashboard(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, dashboard)
}
```

- [ ] **Step 7: Verify all handlers compile**

```bash
cd backend && go build ./internal/handler/...
```

- [ ] **Step 8: Commit**

```bash
git add internal/handler/
git commit -m "feat: add all HTTP handlers"
```

---

## Task 10: Main Entry Point

**Files:**
- Create: `backend/cmd/server/main.go`

- [ ] **Step 1: Create main.go**

```go
package main

import (
	"log"
	"net/http"

	"crm0/backend/internal/algorithm"
	"crm0/backend/internal/config"
	"crm0/backend/internal/handler"
	"crm0/backend/internal/middleware"
	"crm0/backend/internal/repository"
	"crm0/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := repository.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := repository.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Repositories
	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	planRepo := repository.NewPlanRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	eventRepo := repository.NewEventRepository(db)
	algoRepo := repository.NewAlgorithmRepository(db)

	// Algorithm client
	algoClient := algorithm.NewClient(cfg.AlgorithmServiceURL)

	// Services
	authSvc := service.NewAuthService(userRepo, tenantRepo, cfg)
	customerSvc := service.NewCustomerService(customerRepo, eventRepo, algoRepo, algoClient)
	planSvc := service.NewPlanService(planRepo)
	subscriptionSvc := service.NewSubscriptionService(subscriptionRepo)
	eventSvc := service.NewEventService(eventRepo)
	analyticsSvc := service.NewAnalyticsService(algoRepo, algoClient)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	customerH := handler.NewCustomerHandler(customerSvc)
	planH := handler.NewPlanHandler(planSvc)
	subscriptionH := handler.NewSubscriptionHandler(subscriptionSvc)
	eventH := handler.NewEventHandler(eventSvc)
	analyticsH := handler.NewAnalyticsHandler(analyticsSvc)

	// All routes on one mux
	mux := http.NewServeMux()

	// Auth (public — JWT middleware skips these paths)
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)
	mux.HandleFunc("POST /api/v1/auth/refresh", authH.Refresh)

	// Customers
	mux.HandleFunc("GET /api/v1/customers", customerH.List)
	mux.HandleFunc("POST /api/v1/customers", customerH.Create)
	mux.HandleFunc("GET /api/v1/customers/{id}", customerH.Get)
	mux.HandleFunc("PUT /api/v1/customers/{id}", customerH.Update)
	mux.HandleFunc("DELETE /api/v1/customers/{id}", customerH.Delete)
	mux.HandleFunc("GET /api/v1/customers/{id}/insights", customerH.GetInsights)
	mux.HandleFunc("GET /api/v1/customers/{id}/events", customerH.GetEvents)

	// Subscriptions
	mux.HandleFunc("GET /api/v1/subscriptions", subscriptionH.List)
	mux.HandleFunc("POST /api/v1/subscriptions", subscriptionH.Create)
	mux.HandleFunc("GET /api/v1/subscriptions/{id}", subscriptionH.Get)
	mux.HandleFunc("PUT /api/v1/subscriptions/{id}", subscriptionH.Update)
	mux.HandleFunc("DELETE /api/v1/subscriptions/{id}", subscriptionH.Delete)
	mux.HandleFunc("GET /api/v1/subscriptions/metrics", subscriptionH.GetMetrics)

	// Plans
	mux.HandleFunc("GET /api/v1/plans", planH.List)
	mux.HandleFunc("POST /api/v1/plans", planH.Create)
	mux.HandleFunc("GET /api/v1/plans/{id}", planH.Get)
	mux.HandleFunc("PUT /api/v1/plans/{id}", planH.Update)
	mux.HandleFunc("DELETE /api/v1/plans/{id}", planH.Delete)

	// Events
	mux.HandleFunc("POST /api/v1/events", eventH.Create)
	mux.HandleFunc("GET /api/v1/events", eventH.GetRecent)

	// Analytics
	mux.HandleFunc("GET /api/v1/analytics/churn/predictions", analyticsH.GetChurnPredictions)
	mux.HandleFunc("POST /api/v1/analytics/churn/trigger-prediction", analyticsH.TriggerChurnPrediction)
	mux.HandleFunc("GET /api/v1/analytics/segments", analyticsH.GetSegments)
	mux.HandleFunc("POST /api/v1/analytics/segments/trigger-segmentation", analyticsH.TriggerSegmentation)
	mux.HandleFunc("GET /api/v1/analytics/ltv/predictions", analyticsH.GetLTVPredictions)
	mux.HandleFunc("GET /api/v1/analytics/ltv/channel-roi", analyticsH.GetChannelROI)
	mux.HandleFunc("GET /api/v1/analytics/nba/recommendations", analyticsH.GetNBARecommendations)
	mux.HandleFunc("POST /api/v1/analytics/nba/trigger-nba", analyticsH.TriggerNBA)
	mux.HandleFunc("GET /api/v1/dashboard", analyticsH.GetDashboard)

	// Middleware chain: CORS → Logging → Auth (skips login/register) → handler
	var h http.Handler = mux
	h = middleware.JWTAuth(cfg.JWTSecret)(h)
	h = middleware.Logging(h)
	h = middleware.CORS(h)

	log.Printf("Server starting on %s", cfg.AppAddr())
	if err := http.ListenAndServe(cfg.AppAddr(), h); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
```

- [ ] **Step 2: Verify the full backend compiles**

```bash
cd backend && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add cmd/server/main.go
git commit -m "feat: add main entry point with all routes wired up"
```

---

## Task 11: Frontend — Customers View

**Files:**
- Create: `frontend/src/views/Customers.vue`

- [ ] **Step 1: Create Customers.vue**

```vue
<template>
  <div class="customers-page">
    <div class="page-header">
      <h2>客户管理</h2>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon> 新增客户
      </el-button>
    </div>

    <div class="toolbar">
      <el-input v-model="search" placeholder="搜索客户名称或邮箱" clearable style="width: 300px" @clear="loadCustomers" @keyup.enter="loadCustomers">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-select v-model="statusFilter" placeholder="状态筛选" clearable style="width: 150px" @change="loadCustomers">
        <el-option label="活跃" value="active" />
        <el-option label="非活跃" value="inactive" />
        <el-option label="已流失" value="churned" />
      </el-select>
    </div>

    <el-table :data="customerStore.customers" v-loading="customerStore.loading" stripe>
      <el-table-column prop="name" label="姓名" min-width="120" />
      <el-table-column prop="email" label="邮箱" min-width="180" />
      <el-table-column prop="company" label="公司" min-width="150" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="acquired_channel" label="获客渠道" width="120" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="goDetail(row.id)">详情</el-button>
          <el-button link type="warning" @click="editCustomer(row)">编辑</el-button>
          <el-popconfirm title="确定删除该客户?" @confirm="handleDelete(row.id)">
            <template #reference>
              <el-button link type="danger">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="customerStore.total"
      layout="total, prev, pager, next"
      style="margin-top: 16px; justify-content: flex-end"
      @current-change="loadCustomers"
    />

    <!-- Create / Edit Dialog -->
    <el-dialog v-model="showCreateDialog" :title="editingCustomer ? '编辑客户' : '新增客户'" width="500">
      <el-form :model="form" label-width="80px">
        <el-form-item label="姓名" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="邮箱" required><el-input v-model="form.email" /></el-form-item>
        <el-form-item label="公司"><el-input v-model="form.company" /></el-form-item>
        <el-form-item label="电话"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option label="活跃" value="active" />
            <el-option label="非活跃" value="inactive" />
            <el-option label="已流失" value="churned" />
          </el-select>
        </el-form-item>
        <el-form-item label="获客渠道"><el-input v-model="form.acquired_channel" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { useCustomerStore } from '../stores/customer'
import type { Customer } from '../api/customer'

const router = useRouter()
const customerStore = useCustomerStore()

const page = ref(1)
const pageSize = ref(20)
const search = ref('')
const statusFilter = ref('')
const showCreateDialog = ref(false)
const editingCustomer = ref<Customer | null>(null)
const saving = ref(false)

const form = ref({
  name: '',
  email: '',
  company: '',
  phone: '',
  status: 'active',
  acquired_channel: ''
})

const statusType = (s: string) => s === 'active' ? 'success' : s === 'inactive' ? 'warning' : 'danger'
const statusLabel = (s: string) => s === 'active' ? '活跃' : s === 'inactive' ? '非活跃' : '已流失'

const loadCustomers = () => {
  customerStore.fetchCustomers({
    page: page.value,
    page_size: pageSize.value,
    search: search.value || undefined,
    status: statusFilter.value || undefined
  })
}

const goDetail = (id: string) => router.push(`/customers/${id}`)

const editCustomer = (c: Customer) => {
  editingCustomer.value = c
  form.value = { name: c.name, email: c.email, company: c.company || '', phone: c.phone || '', status: c.status, acquired_channel: c.acquired_channel || '' }
  showCreateDialog.value = true
}

const handleSave = async () => {
  saving.value = true
  try {
    if (editingCustomer.value) {
      await customerStore.updateCustomer(editingCustomer.value.id, form.value)
      ElMessage.success('客户已更新')
    } else {
      await customerStore.createCustomer(form.value)
      ElMessage.success('客户已创建')
    }
    showCreateDialog.value = false
    editingCustomer.value = null
    form.value = { name: '', email: '', company: '', phone: '', status: 'active', acquired_channel: '' }
    loadCustomers()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

const handleDelete = async (id: string) => {
  await customerStore.deleteCustomer(id)
  ElMessage.success('客户已删除')
  loadCustomers()
}

onMounted(loadCustomers)
</script>

<style scoped>
.customers-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.toolbar { display: flex; gap: 12px; margin-bottom: 16px; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/views/Customers.vue
git commit -m "feat: add Customers list view"
```

---

## Task 12: Frontend — Customer Detail View

**Files:**
- Create: `frontend/src/views/CustomerDetail.vue`

- [ ] **Step 1: Create CustomerDetail.vue**

```vue
<template>
  <div class="customer-detail" v-loading="customerStore.loading">
    <el-page-header @back="router.push('/customers')" :title="customer?.name || ''" content="客户详情" />

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="16">
        <el-card header="基本信息">
          <el-descriptions :column="2" border v-if="customer">
            <el-descriptions-item label="姓名">{{ customer.name }}</el-descriptions-item>
            <el-descriptions-item label="邮箱">{{ customer.email }}</el-descriptions-item>
            <el-descriptions-item label="公司">{{ customer.company || '-' }}</el-descriptions-item>
            <el-descriptions-item label="电话">{{ customer.phone || '-' }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="customer.status === 'active' ? 'success' : customer.status === 'inactive' ? 'warning' : 'danger'" size="small">{{ customer.status }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="获客渠道">{{ customer.acquired_channel || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card header="事件记录" style="margin-top: 16px">
          <el-timeline v-if="events.length">
            <el-timeline-item v-for="e in events" :key="e.id" :timestamp="e.occurred_at" placement="top">
              <el-card shadow="never">
                <h4>{{ e.event_type }}</h4>
                <p v-if="e.properties && Object.keys(e.properties).length" style="color: #909399; font-size: 13px">{{ JSON.stringify(e.properties) }}</p>
              </el-card>
            </el-timeline-item>
          </el-timeline>
          <el-empty v-else description="暂无事件记录" />
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card header="智能洞察">
          <div v-if="insight">
            <div class="insight-item">
              <span class="insight-label">流失风险</span>
              <el-progress v-if="insight.churn_prediction" :percentage="Math.round(insight.churn_prediction.risk_score * 100)" :color="churnColor(insight.churn_prediction.risk_score)" :stroke-width="18" :text-inside="true" />
              <span v-else style="color: #909399">未预测</span>
            </div>
            <div class="insight-item" v-if="insight.ltv">
              <span class="insight-label">预测LTV</span>
              <span class="insight-value">¥{{ insight.ltv.predicted_ltv.toLocaleString() }}</span>
              <span style="color: #909399; font-size: 12px; margin-left: 8px">置信度 {{ (insight.ltv.confidence * 100).toFixed(0) }}%</span>
            </div>
            <div class="insight-item" v-if="insight.segments?.length">
              <span class="insight-label">客户分群</span>
              <el-tag v-for="s in insight.segments" :key="s.id" size="small" style="margin-right: 4px">{{ s.segment_name }}</el-tag>
            </div>
            <div class="insight-item" v-if="insight.recommendations?.length">
              <span class="insight-label">推荐行动</span>
              <div v-for="r in insight.recommendations" :key="r.id" class="nba-item">
                <el-tag :type="r.action_type === 'call' ? 'danger' : r.action_type === 'discount' ? 'warning' : r.action_type === 'email' ? '' : 'success'" size="small">{{ r.action_type }}</el-tag>
                <span style="font-size: 13px; margin-left: 8px">影响度: {{ (r.expected_impact * 100).toFixed(0) }}%</span>
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无洞察数据" :image-size="60" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCustomerStore } from '../stores/customer'
import { getEvents } from '../api/customer'
import type { CustomerInsight } from '../api/customer'

const route = useRoute()
const router = useRouter()
const customerStore = useCustomerStore()
const customerId = route.params.id as string

const events = ref<any[]>([])
const insight = computed(() => customerStore.currentInsight)
const customer = computed(() => customerStore.currentCustomer)

const churnColor = (score: number) => score > 0.7 ? '#F56C6C' : score > 0.4 ? '#E6A23C' : '#67C23A'

onMounted(async () => {
  await customerStore.fetchCustomer(customerId)
  await customerStore.fetchInsights(customerId)
  try {
    const res = await getEvents(customerId, { page: 1, page_size: 20 })
    events.value = res.items
  } catch { /* ignore */ }
})
</script>

<style scoped>
.customer-detail { padding: 20px; }
.insight-item { margin-bottom: 16px; }
.insight-label { display: block; font-size: 13px; color: #909399; margin-bottom: 6px; }
.insight-value { font-size: 20px; font-weight: 600; color: #409EFF; }
.nba-item { display: flex; align-items: center; margin-bottom: 8px; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/views/CustomerDetail.vue
git commit -m "feat: add CustomerDetail view with insights panel"
```

---

## Task 13: Frontend — Subscriptions View

**Files:**
- Create: `frontend/src/views/Subscriptions.vue`

- [ ] **Step 1: Create Subscriptions.vue**

```vue
<template>
  <div class="subscriptions-page">
    <div class="page-header">
      <h2>订阅管理</h2>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon> 新增订阅
      </el-button>
    </div>

    <el-row :gutter="16" style="margin-bottom: 20px">
      <el-col :span="6"><el-card shadow="never"><el-statistic title="总MRR" :value="metrics.total_mrr" prefix="¥" /></el-card></el-col>
      <el-col :span="6"><el-card shadow="never"><el-statistic title="活跃订阅" :value="metrics.active_count" /></el-card></el-col>
      <el-col :span="6"><el-card shadow="never"><el-statistic title="试用中" :value="metrics.trial_count" /></el-card></el-col>
      <el-col :span="6"><el-card shadow="never"><el-statistic title="流失率" :value="metrics.churn_rate" suffix="%" :precision="1" /></el-card></el-col>
    </el-row>

    <el-table :data="subscriptions" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="100" show-overflow-tooltip />
      <el-table-column label="客户" min-width="150">
        <template #default="{ row }">{{ row.customer?.name || row.customer_id }}</template>
      </el-table-column>
      <el-table-column label="计划" min-width="150">
        <template #default="{ row }">{{ row.plan?.name || row.plan_id }}</template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : row.status === 'trial' ? 'warning' : 'danger'" size="small">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="mrr" label="MRR" width="120">
        <template #default="{ row }">¥{{ row.mrr }}</template>
      </el-table-column>
      <el-table-column prop="started_at" label="开始日期" width="180" />
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="warning" @click="editSub(row)">编辑</el-button>
          <el-popconfirm title="确定取消该订阅?" @confirm="handleCancel(row.id)">
            <template #reference>
              <el-button link type="danger">取消</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination v-model:current-page="page" :total="total" layout="total, prev, pager, next" style="margin-top: 16px; justify-content: flex-end" @current-change="loadData" />

    <el-dialog v-model="showCreateDialog" :title="editingSub ? '编辑订阅' : '新增订阅'" width="450">
      <el-form :model="form" label-width="80px">
        <el-form-item label="客户ID" v-if="!editingSub"><el-input v-model="form.customer_id" /></el-form-item>
        <el-form-item label="计划ID" v-if="!editingSub"><el-input v-model="form.plan_id" /></el-form-item>
        <el-form-item label="状态" v-if="editingSub">
          <el-select v-model="form.status">
            <el-option label="试用" value="trial" />
            <el-option label="活跃" value="active" />
            <el-option label="已取消" value="canceled" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { list, create, update, remove, getMetrics } from '../api/subscription'
import type { Subscription, SubscriptionMetrics } from '../api/subscription'

const subscriptions = ref<Subscription[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const metrics = ref<SubscriptionMetrics>({ total_mrr: 0, active_count: 0, canceled_count: 0, trial_count: 0, avg_mrr_per_customer: 0 })
const showCreateDialog = ref(false)
const editingSub = ref<Subscription | null>(null)
const saving = ref(false)
const form = ref({ customer_id: '', plan_id: '', status: 'trial' })

const loadData = async () => {
  loading.value = true
  try {
    const [listRes, metricsRes] = await Promise.all([
      list({ page: page.value, page_size: 20 }),
      getMetrics()
    ])
    subscriptions.value = listRes.items
    total.value = listRes.total
    metrics.value = metricsRes
  } finally {
    loading.value = false
  }
}

const editSub = (s: Subscription) => {
  editingSub.value = s
  form.value.status = s.status
  showCreateDialog.value = true
}

const handleSave = async () => {
  saving.value = true
  try {
    if (editingSub.value) {
      await update(editingSub.value.id, { status: form.value.status })
      ElMessage.success('订阅已更新')
    } else {
      await create({ customer_id: form.value.customer_id, plan_id: form.value.plan_id })
      ElMessage.success('订阅已创建')
    }
    showCreateDialog.value = false
    editingSub.value = null
    loadData()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

const handleCancel = async (id: string) => {
  await remove(id)
  ElMessage.success('订阅已取消')
  loadData()
}

onMounted(loadData)
</script>

<style scoped>
.subscriptions-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/views/Subscriptions.vue
git commit -m "feat: add Subscriptions management view"
```

---

## Task 14: Frontend — Churn Analysis View

**Files:**
- Create: `frontend/src/views/analytics/ChurnAnalysis.vue`

- [ ] **Step 1: Create ChurnAnalysis.vue**

```vue
<template>
  <div class="churn-page">
    <div class="page-header">
      <h2>流失预测分析</h2>
      <el-button type="primary" @click="handleTrigger" :loading="analyticsStore.loading">触发预测</el-button>
    </div>

    <el-row :gutter="20">
      <el-col :span="10">
        <el-card header="风险分布">
          <v-chart :option="chartOption" style="height: 300px" autoresize />
        </el-card>
      </el-col>
      <el-col :span="14">
        <el-card header="客户流失风险">
          <el-table :data="analyticsStore.churnPredictions" v-loading="analyticsStore.loading" stripe max-height="400">
            <el-table-column label="客户" min-width="140">
              <template #default="{ row }">{{ row.customer?.name || row.customer_id }}</template>
            </el-table-column>
            <el-table-column label="风险分" width="160">
              <template #default="{ row }">
                <el-progress :percentage="Math.round(row.risk_score * 100)" :color="riskColor(row.risk_score)" :stroke-width="16" :text-inside="true" />
              </template>
            </el-table-column>
            <el-table-column label="风险等级" width="100">
              <template #default="{ row }">
                <el-tag :type="row.risk_level === 'high' ? 'danger' : row.risk_level === 'medium' ? 'warning' : 'success'" size="small">{{ row.risk_level }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="关键因素" min-width="200">
              <template #default="{ row }">
                <el-tag v-for="(val, key) in (row.factors || {})" :key="key" size="small" style="margin: 2px">{{ key }}: {{ (Number(val) * 100).toFixed(0) }}%</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { BarChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, GridComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { useAnalyticsStore } from '../../stores/analytics'

use([BarChart, TitleComponent, TooltipComponent, GridComponent, CanvasRenderer])

const analyticsStore = useAnalyticsStore()

const riskColor = (score: number) => score > 0.7 ? '#F56C6C' : score > 0.4 ? '#E6A23C' : '#67C23A'

const chartOption = computed(() => {
  const high = analyticsStore.churnPredictions.filter(p => p.risk_level === 'high').length
  const medium = analyticsStore.churnPredictions.filter(p => p.risk_level === 'medium').length
  const low = analyticsStore.churnPredictions.filter(p => p.risk_level === 'low').length
  return {
    tooltip: {},
    xAxis: { type: 'category', data: ['高风险', '中风险', '低风险'] },
    yAxis: { type: 'value' },
    series: [{
      type: 'bar',
      data: [
        { value: high, itemStyle: { color: '#F56C6C' } },
        { value: medium, itemStyle: { color: '#E6A23C' } },
        { value: low, itemStyle: { color: '#67C23A' } }
      ]
    }]
  }
})

const handleTrigger = async () => {
  try {
    await analyticsStore.triggerChurn()
    ElMessage.success('流失预测已完成')
    analyticsStore.fetchChurn()
  } catch {
    ElMessage.error('预测失败，请检查算法服务')
  }
}

onMounted(() => analyticsStore.fetchChurn())
</script>

<style scoped>
.churn-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
</style>
```

- [ ] **Step 2: Create analytics directory and commit**

```bash
mkdir -p frontend/src/views/analytics
git add frontend/src/views/analytics/ChurnAnalysis.vue
git commit -m "feat: add Churn analysis view"
```

---

## Task 15: Frontend — Segment Analysis View

**Files:**
- Create: `frontend/src/views/analytics/SegmentAnalysis.vue`

- [ ] **Step 1: Create SegmentAnalysis.vue**

```vue
<template>
  <div class="segment-page">
    <div class="page-header">
      <h2>客户分群分析</h2>
      <div>
        <el-select v-model="segmentType" style="width: 140px; margin-right: 12px">
          <el-option label="RFM分群" value="rfm" />
          <el-option label="行为分群" value="behavioral" />
          <el-option label="价值分群" value="value" />
        </el-select>
        <el-button type="primary" @click="handleTrigger" :loading="analyticsStore.loading">执行分群</el-button>
      </div>
    </div>

    <el-row :gutter="20">
      <el-col :span="10">
        <el-card header="分群分布">
          <v-chart :option="chartOption" style="height: 350px" autoresize />
        </el-card>
      </el-col>
      <el-col :span="14">
        <el-card header="分群结果">
          <el-table :data="analyticsStore.segments" v-loading="analyticsStore.loading" stripe max-height="400">
            <el-table-column label="客户" min-width="140">
              <template #default="{ row }">{{ row.customer?.name || row.customer_id }}</template>
            </el-table-column>
            <el-table-column prop="segment_type" label="分群类型" width="100" />
            <el-table-column prop="segment_name" label="分群名称" width="120" />
            <el-table-column prop="score" label="分数" width="120">
              <template #default="{ row }">{{ row.score.toFixed(2) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { PieChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { useAnalyticsStore } from '../../stores/analytics'

use([PieChart, TitleComponent, TooltipComponent, LegendComponent, CanvasRenderer])

const analyticsStore = useAnalyticsStore()
const segmentType = ref('rfm')

const chartOption = computed(() => {
  const counts: Record<string, number> = {}
  for (const s of analyticsStore.segments) {
    counts[s.segment_name] = (counts[s.segment_name] || 0) + 1
  }
  return {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0 },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      data: Object.entries(counts).map(([name, value]) => ({ name, value }))
    }]
  }
})

const handleTrigger = async () => {
  try {
    await analyticsStore.triggerSegmentation(segmentType.value)
    ElMessage.success('分群完成')
    analyticsStore.fetchSegments(segmentType.value)
  } catch {
    ElMessage.error('分群失败，请检查算法服务')
  }
}

onMounted(() => analyticsStore.fetchSegments(segmentType.value))
</script>

<style scoped>
.segment-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/views/analytics/SegmentAnalysis.vue
git commit -m "feat: add Segment analysis view"
```

---

## Task 16: Frontend — LTV Analysis View

**Files:**
- Create: `frontend/src/views/analytics/LTVAnalysis.vue`

- [ ] **Step 1: Create LTVAnalysis.vue**

```vue
<template>
  <div class="ltv-page">
    <div class="page-header">
      <h2>LTV 分析</h2>
    </div>

    <el-row :gutter="20">
      <el-col :span="14">
        <el-card header="客户LTV排名">
          <v-chart :option="ltvChartOption" style="height: 350px" autoresize />
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card header="渠道ROI">
          <v-chart :option="roiChartOption" style="height: 350px" autoresize />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 16px">
      <el-col :span="14">
        <el-card header="LTV详情">
          <el-table :data="analyticsStore.ltvPredictions" v-loading="analyticsStore.loading" stripe max-height="350">
            <el-table-column label="客户" min-width="140">
              <template #default="{ row }">{{ row.customer?.name || row.customer_id }}</template>
            </el-table-column>
            <el-table-column label="预测LTV" width="130">
              <template #default="{ row }">¥{{ row.predicted_ltv.toLocaleString() }}</template>
            </el-table-column>
            <el-table-column label="置信度" width="100">
              <template #default="{ row }">{{ (row.confidence * 100).toFixed(0) }}%</template>
            </el-table-column>
            <el-table-column label="预期生命周期" width="130">
              <template #default="{ row }">{{ row.expected_lifetime_months }}月</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card header="渠道ROI详情">
          <el-table :data="analyticsStore.channelROI" stripe max-height="350">
            <el-table-column prop="channel" label="渠道" min-width="100" />
            <el-table-column label="CAC" width="90">
              <template #default="{ row }">¥{{ row.cac }}</template>
            </el-table-column>
            <el-table-column label="LTV" width="90">
              <template #default="{ row }">¥{{ row.ltv }}</template>
            </el-table-column>
            <el-table-column label="LTV/CAC" width="90">
              <template #default="{ row }">{{ row.ltv_cac_ratio.toFixed(1) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { BarChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, GridComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { useAnalyticsStore } from '../../stores/analytics'

use([BarChart, TitleComponent, TooltipComponent, GridComponent, LegendComponent, CanvasRenderer])

const analyticsStore = useAnalyticsStore()

const ltvChartOption = computed(() => {
  const top10 = [...analyticsStore.ltvPredictions].sort((a, b) => b.predicted_ltv - a.predicted_ltv).slice(0, 10)
  return {
    tooltip: {},
    xAxis: { type: 'category', data: top10.map(p => p.customer?.name || p.customer_id.slice(0, 8)), axisLabel: { rotate: 30 } },
    yAxis: { type: 'value', name: 'LTV (¥)' },
    series: [{ type: 'bar', data: top10.map(p => p.predicted_ltv), itemStyle: { color: '#409EFF' } }]
  }
})

const roiChartOption = computed(() => {
  const channels = analyticsStore.channelROI
  return {
    tooltip: {},
    legend: {},
    xAxis: { type: 'category', data: channels.map(c => c.channel) },
    yAxis: { type: 'value', name: '¥' },
    series: [
      { name: 'CAC', type: 'bar', data: channels.map(c => c.cac), itemStyle: { color: '#F56C6C' } },
      { name: 'LTV', type: 'bar', data: channels.map(c => c.ltv), itemStyle: { color: '#67C23A' } }
    ]
  }
})

onMounted(() => {
  analyticsStore.fetchLTV()
  analyticsStore.fetchChannelROI()
})
</script>

<style scoped>
.ltv-page { padding: 20px; }
.page-header { margin-bottom: 16px; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/views/analytics/LTVAnalysis.vue
git commit -m "feat: add LTV analysis view"
```

---

## Task 17: Frontend — NBA Recommendations View

**Files:**
- Create: `frontend/src/views/analytics/NBARecommendations.vue`

- [ ] **Step 1: Create NBARecommendations.vue**

```vue
<template>
  <div class="nba-page">
    <div class="page-header">
      <h2>智能推荐行动</h2>
      <el-button type="primary" @click="handleTrigger" :loading="analyticsStore.loading">生成推荐</el-button>
    </div>

    <el-row :gutter="16">
      <el-col :span="6" v-for="rec in analyticsStore.nbaRecommendations" :key="rec.id">
        <el-card shadow="hover" style="margin-bottom: 16px">
          <div class="nba-card">
            <div class="nba-header">
              <el-tag :type="actionType(rec.action_type)" size="large">{{ actionLabel(rec.action_type) }}</el-tag>
              <el-tag type="info" size="small">优先级 {{ rec.priority }}</el-tag>
            </div>
            <h4 style="margin: 12px 0 8px">{{ rec.customer?.name || '客户' }}</h4>
            <p style="color: #909399; font-size: 13px; margin-bottom: 8px">{{ recActionDetail(rec.action_detail) }}</p>
            <div class="nba-impact">
              <span>预期影响</span>
              <el-progress :percentage="Math.round(rec.expected_impact * 100)" :stroke-width="12" style="flex: 1; margin-left: 8px" />
            </div>
            <div style="margin-top: 8px; display: flex; justify-content: space-between; align-items: center">
              <el-tag :type="rec.status === 'pending' ? 'warning' : rec.status === 'completed' ? 'success' : 'info'" size="small">{{ rec.status }}</el-tag>
              <span style="color: #C0C4CC; font-size: 12px">{{ rec.created_at?.slice(0, 10) }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty v-if="!analyticsStore.nbaRecommendations.length && !analyticsStore.loading" description="暂无推荐行动" />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useAnalyticsStore } from '../../stores/analytics'

const analyticsStore = useAnalyticsStore()

const actionType = (t: string) => t === 'call' ? 'danger' : t === 'discount' ? 'warning' : t === 'email' ? '' : 'success'
const actionLabel = (t: string) => t === 'call' ? '电话' : t === 'discount' ? '优惠' : t === 'email' ? '邮件' : t === 'feature_guide' ? '功能引导' : t
const recActionDetail = (d: Record<string, unknown>) => {
  if (!d || !Object.keys(d).length) return ''
  return Object.entries(d).map(([k, v]) => `${k}: ${v}`).join(', ')
}

const handleTrigger = async () => {
  try {
    await analyticsStore.triggerNBAAction()
    ElMessage.success('推荐已生成')
    analyticsStore.fetchNBA()
  } catch {
    ElMessage.error('生成失败，请检查算法服务')
  }
}

onMounted(() => analyticsStore.fetchNBA())
</script>

<style scoped>
.nba-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.nba-card { text-align: left; }
.nba-header { display: flex; justify-content: space-between; align-items: center; }
.nba-impact { display: flex; align-items: center; font-size: 13px; color: #606266; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/views/analytics/NBARecommendations.vue
git commit -m "feat: add NBA recommendations view"
```

---

## Task 18: Frontend — Settings View

**Files:**
- Create: `frontend/src/views/Settings.vue`

- [ ] **Step 1: Create Settings.vue**

```vue
<template>
  <div class="settings-page">
    <h2>系统设置</h2>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card header="组织信息">
          <el-form :model="tenantForm" label-width="100px">
            <el-form-item label="组织名称">
              <el-input v-model="tenantForm.name" disabled />
            </el-form-item>
            <el-form-item label="当前套餐">
              <el-tag>{{ tenantForm.plan }}</el-tag>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card header="个人信息">
          <el-form :model="userForm" label-width="100px">
            <el-form-item label="姓名">
              <el-input v-model="userForm.name" />
            </el-form-item>
            <el-form-item label="邮箱">
              <el-input v-model="userForm.email" disabled />
            </el-form-item>
            <el-form-item label="角色">
              <el-tag>{{ userForm.role }}</el-tag>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleSaveProfile" :loading="saving">保存修改</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const saving = ref(false)

const tenantForm = ref({ name: '', plan: '' })
const userForm = ref({ name: '', email: '', role: '' })

onMounted(() => {
  const user = authStore.user
  if (user) {
    userForm.value = { name: user.name, email: user.email, role: user.role }
  }
  tenantForm.value = { name: '我的组织', plan: 'free' }
})

const handleSaveProfile = async () => {
  saving.value = true
  try {
    // TODO: call user update API when available
    ElMessage.success('个人信息已更新')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.settings-page { padding: 20px; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/views/Settings.vue
git commit -m "feat: add Settings view"
```

---

## Task 19: Fix Backend Router — Auth Route Handling

The main.go has a routing issue: public auth routes are wrapped in the auth middleware via the catch-all `/api/v1/` handler. Fix this by serving auth routes directly and only applying JWT to protected routes.

**Files:**
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Fix the route registration in main.go**

Replace the route registration section in main.go with:

```go
	// Public routes (no auth)
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)

	// Protected routes
	protected := http.NewServeMux()
	protected.HandleFunc("POST /api/v1/auth/refresh", authH.Refresh)

	protected.HandleFunc("GET /api/v1/customers", customerH.List)
	protected.HandleFunc("POST /api/v1/customers", customerH.Create)
	protected.HandleFunc("GET /api/v1/customers/{id}", customerH.Get)
	protected.HandleFunc("PUT /api/v1/customers/{id}", customerH.Update)
	protected.HandleFunc("DELETE /api/v1/customers/{id}", customerH.Delete)
	protected.HandleFunc("GET /api/v1/customers/{id}/insights", customerH.GetInsights)
	protected.HandleFunc("GET /api/v1/customers/{id}/events", customerH.GetEvents)

	protected.HandleFunc("GET /api/v1/subscriptions", subscriptionH.List)
	protected.HandleFunc("POST /api/v1/subscriptions", subscriptionH.Create)
	protected.HandleFunc("GET /api/v1/subscriptions/{id}", subscriptionH.Get)
	protected.HandleFunc("PUT /api/v1/subscriptions/{id}", subscriptionH.Update)
	protected.HandleFunc("DELETE /api/v1/subscriptions/{id}", subscriptionH.Delete)
	protected.HandleFunc("GET /api/v1/subscriptions/metrics", subscriptionH.GetMetrics)

	protected.HandleFunc("GET /api/v1/plans", planH.List)
	protected.HandleFunc("POST /api/v1/plans", planH.Create)
	protected.HandleFunc("GET /api/v1/plans/{id}", planH.Get)
	protected.HandleFunc("PUT /api/v1/plans/{id}", planH.Update)
	protected.HandleFunc("DELETE /api/v1/plans/{id}", planH.Delete)

	protected.HandleFunc("POST /api/v1/events", eventH.Create)
	protected.HandleFunc("GET /api/v1/events", eventH.GetRecent)

	protected.HandleFunc("GET /api/v1/analytics/churn/predictions", analyticsH.GetChurnPredictions)
	protected.HandleFunc("POST /api/v1/analytics/churn/trigger-prediction", analyticsH.TriggerChurnPrediction)
	protected.HandleFunc("GET /api/v1/analytics/segments", analyticsH.GetSegments)
	protected.HandleFunc("POST /api/v1/analytics/segments/trigger-segmentation", analyticsH.TriggerSegmentation)
	protected.HandleFunc("GET /api/v1/analytics/ltv/predictions", analyticsH.GetLTVPredictions)
	protected.HandleFunc("GET /api/v1/analytics/ltv/channel-roi", analyticsH.GetChannelROI)
	protected.HandleFunc("GET /api/v1/analytics/nba/recommendations", analyticsH.GetNBARecommendations)
	protected.HandleFunc("POST /api/v1/analytics/nba/trigger-nba", analyticsH.TriggerNBA)
	protected.HandleFunc("GET /api/v1/dashboard", analyticsH.GetDashboard)

	// Chain middleware: CORS → Logging → [Auth for protected routes]
	authMiddleware := middleware.JWTAuth(cfg.JWTSecret)

	// Route: public auth endpoints bypass JWT, everything else goes through auth
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Public routes
		if r.URL.Path == "/api/v1/auth/login" || r.URL.Path == "/api/v1/auth/register" {
			middleware.CORS(middleware.Logging(http.DefaultServeMux)).ServeHTTP(w, r)
			return
		}
		middleware.CORS(middleware.Logging(authMiddleware(protected))).ServeHTTP(w, r)
	})

	// Register public routes on default mux
	http.HandleFunc("POST /api/v1/auth/login", authH.Login)
	http.HandleFunc("POST /api/v1/auth/register", authH.Register)

	log.Printf("Server starting on %s", cfg.AppAddr())
	if err := http.ListenAndServe(cfg.AppAddr(), handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
```

Actually, this is getting complex with the default mux. Let's use a simpler approach — register everything on one mux and handle auth in the middleware by checking the path:

Replace the entire route section with:

```go
	// All routes on one mux
	mux := http.NewServeMux()

	// Auth (public)
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)

	// Auth (protected - needs token)
	mux.HandleFunc("POST /api/v1/auth/refresh", authH.Refresh)

	// Customers
	mux.HandleFunc("GET /api/v1/customers", customerH.List)
	mux.HandleFunc("POST /api/v1/customers", customerH.Create)
	mux.HandleFunc("GET /api/v1/customers/{id}", customerH.Get)
	mux.HandleFunc("PUT /api/v1/customers/{id}", customerH.Update)
	mux.HandleFunc("DELETE /api/v1/customers/{id}", customerH.Delete)
	mux.HandleFunc("GET /api/v1/customers/{id}/insights", customerH.GetInsights)
	mux.HandleFunc("GET /api/v1/customers/{id}/events", customerH.GetEvents)

	// Subscriptions
	mux.HandleFunc("GET /api/v1/subscriptions", subscriptionH.List)
	mux.HandleFunc("POST /api/v1/subscriptions", subscriptionH.Create)
	mux.HandleFunc("GET /api/v1/subscriptions/{id}", subscriptionH.Get)
	mux.HandleFunc("PUT /api/v1/subscriptions/{id}", subscriptionH.Update)
	mux.HandleFunc("DELETE /api/v1/subscriptions/{id}", subscriptionH.Delete)
	mux.HandleFunc("GET /api/v1/subscriptions/metrics", subscriptionH.GetMetrics)

	// Plans
	mux.HandleFunc("GET /api/v1/plans", planH.List)
	mux.HandleFunc("POST /api/v1/plans", planH.Create)
	mux.HandleFunc("GET /api/v1/plans/{id}", planH.Get)
	mux.HandleFunc("PUT /api/v1/plans/{id}", planH.Update)
	mux.HandleFunc("DELETE /api/v1/plans/{id}", planH.Delete)

	// Events
	mux.HandleFunc("POST /api/v1/events", eventH.Create)
	mux.HandleFunc("GET /api/v1/events", eventH.GetRecent)

	// Analytics
	mux.HandleFunc("GET /api/v1/analytics/churn/predictions", analyticsH.GetChurnPredictions)
	mux.HandleFunc("POST /api/v1/analytics/churn/trigger-prediction", analyticsH.TriggerChurnPrediction)
	mux.HandleFunc("GET /api/v1/analytics/segments", analyticsH.GetSegments)
	mux.HandleFunc("POST /api/v1/analytics/segments/trigger-segmentation", analyticsH.TriggerSegmentation)
	mux.HandleFunc("GET /api/v1/analytics/ltv/predictions", analyticsH.GetLTVPredictions)
	mux.HandleFunc("GET /api/v1/analytics/ltv/channel-roi", analyticsH.GetChannelROI)
	mux.HandleFunc("GET /api/v1/analytics/nba/recommendations", analyticsH.GetNBARecommendations)
	mux.HandleFunc("POST /api/v1/analytics/nba/trigger-nba", analyticsH.TriggerNBA)
	mux.HandleFunc("GET /api/v1/dashboard", analyticsH.GetDashboard)

	// Middleware chain: CORS → Logging → Auth (skip for login/register) → handler
	var handler http.Handler = mux
	handler = middleware.JWTAuth(cfg.JWTSecret)(handler)
	handler = middleware.Logging(handler)
	handler = middleware.CORS(handler)

	log.Printf("Server starting on %s", cfg.AppAddr())
	if err := http.ListenAndServe(cfg.AppAddr(), handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
```

- [ ] **Step 2: Update JWTAuth middleware to skip public routes**

Modify `backend/internal/middleware/auth.go` — update the JWTAuth function to skip auth for login and register paths:

```go
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for public endpoints
			if r.URL.Path == "/api/v1/auth/login" || r.URL.Path == "/api/v1/auth/register" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == authHeader {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"invalid claims"}`, http.StatusUnauthorized)
				return
			}

			tenantID, err := uuid.Parse(claims["tenant_id"].(string))
			if err != nil {
				http.Error(w, `{"error":"invalid tenant_id in token"}`, http.StatusUnauthorized)
				return
			}
			userID, err := uuid.Parse(claims["user_id"].(string))
			if err != nil {
				http.Error(w, `{"error":"invalid user_id in token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
			ctx = context.WithValue(ctx, UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
```

- [ ] **Step 3: Verify full backend compiles**

```bash
cd backend && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add cmd/server/main.go internal/middleware/auth.go
git commit -m "fix: properly handle public vs protected routes"
```

---

## Task 20: Verify Frontend Builds

- [ ] **Step 1: Run frontend type-check and build**

```bash
cd frontend && npm run build
```

Expected: successful build with no TypeScript errors

- [ ] **Step 2: Fix any compilation errors**

If the build fails due to missing view files or import issues, fix them and re-run.

- [ ] **Step 3: Commit any fixes**

```bash
git add -A && git commit -m "fix: resolve frontend build errors"
```

---

## Task 21: Verify Backend Builds

- [ ] **Step 1: Run full backend build**

```bash
cd backend && go build ./...
```

Expected: no errors

- [ ] **Step 2: Run go vet**

```bash
cd backend && go vet ./...
```

Expected: no warnings

- [ ] **Step 3: Fix any issues and commit**

```bash
git add -A && git commit -m "fix: resolve backend build warnings"
```
