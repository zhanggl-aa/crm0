package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ──────────────────────────────────────────────
// Domain models
// ──────────────────────────────────────────────

// Tenant represents an organization using the CRM platform.
type Tenant struct {
	ID                  uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name                string          `json:"name" gorm:"type:varchar(255);not null"`
	Plan                string          `json:"plan" gorm:"type:varchar(50);not null;default:'free'"`
	Settings            json.RawMessage `json:"settings" gorm:"type:jsonb;default:'{}'"`
	OnboardingCompleted bool            `json:"onboarding_completed" gorm:"default:false"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

func (Tenant) TableName() string { return "tenants" }

// User represents an authenticated user belonging to a tenant.
type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;uniqueIndex:idx_tenant_email"`
	Email        string    `json:"email" gorm:"type:varchar(255);not null;uniqueIndex:idx_tenant_email"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;type:varchar(255);not null"`
	Name         string    `json:"name" gorm:"type:varchar(255);not null"`
	Role         string    `json:"role" gorm:"type:varchar(20);not null;default:'member'"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string { return "users" }

// Customer represents a customer record within a tenant.
type Customer struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID        uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null"`
	Name            string          `json:"name" gorm:"type:varchar(255);not null"`
	Email           string          `json:"email" gorm:"type:varchar(255);not null"`
	Company         *string         `json:"company" gorm:"type:varchar(255)"`
	Phone           *string         `json:"phone" gorm:"type:varchar(50)"`
	Status          string          `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	Tags            json.RawMessage `json:"tags" gorm:"type:jsonb;default:'[]'"`
	CustomFields    json.RawMessage `json:"custom_fields" gorm:"type:jsonb;default:'{}'"`
	AcquiredChannel *string         `json:"acquired_channel" gorm:"type:varchar(100)"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func (Customer) TableName() string { return "customers" }

// Plan represents a billing plan available to customers.
type Plan struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
	Name         string    `json:"name" gorm:"type:varchar(255);not null"`
	Price        float64   `json:"price" gorm:"type:decimal(10,2);not null"`
	BillingCycle string    `json:"billing_cycle" gorm:"type:varchar(20);not null;default:'monthly'"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Plan) TableName() string { return "plans" }

// Subscription links a customer to a plan and tracks subscription state.
type Subscription struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CustomerID uuid.UUID  `json:"customer_id" gorm:"type:uuid;not null"`
	PlanID     uuid.UUID  `json:"plan_id" gorm:"type:uuid;not null"`
	Status     string     `json:"status" gorm:"type:varchar(20);not null;default:'trial'"`
	MRR        float64    `json:"mrr" gorm:"type:decimal(10,2);not null;default:0"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	CanceledAt *time.Time `json:"canceled_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (Subscription) TableName() string { return "subscriptions" }

// UserEvent represents a tracked event performed by or for a customer.
type UserEvent struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CustomerID  uuid.UUID       `json:"customer_id" gorm:"type:uuid;not null"`
	TenantID    uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null"`
	EventType   string          `json:"event_type" gorm:"type:varchar(100);not null"`
	Properties  json.RawMessage `json:"properties" gorm:"type:jsonb;default:'{}'"`
	OccurredAt  time.Time       `json:"occurred_at"`
	CreatedAt   time.Time       `json:"created_at" gorm:"autoCreateTime"`
}

func (UserEvent) TableName() string { return "user_events" }

// ChurnPrediction stores the output of a churn-risk prediction for a customer.
type ChurnPrediction struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CustomerID  uuid.UUID       `json:"customer_id" gorm:"type:uuid;not null"`
	TenantID    uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null"`
	RiskScore   float64         `json:"risk_score" gorm:"not null"`
	RiskLevel   string          `json:"risk_level" gorm:"type:varchar(10);not null"`
	Factors     json.RawMessage `json:"factors" gorm:"type:jsonb;default:'[]'"`
	PredictedAt time.Time       `json:"predicted_at"`
}

func (ChurnPrediction) TableName() string { return "churn_predictions" }

// CustomerSegment records a customer's membership in an analytics segment.
type CustomerSegment struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CustomerID  uuid.UUID `json:"customer_id" gorm:"type:uuid;not null"`
	TenantID    uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
	SegmentType string    `json:"segment_type" gorm:"type:varchar(20);not null"`
	SegmentName string    `json:"segment_name" gorm:"type:varchar(100);not null"`
	Score       float64   `json:"score" gorm:"not null;default:0"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (CustomerSegment) TableName() string { return "customer_segments" }

// LTVPrediction stores the predicted lifetime value for a customer.
type LTVPrediction struct {
	ID                     uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CustomerID             uuid.UUID `json:"customer_id" gorm:"type:uuid;not null"`
	TenantID               uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
	PredictedLTV           float64   `json:"predicted_ltv" gorm:"type:decimal(12,2);not null"`
	Confidence             float64   `json:"confidence" gorm:"not null"`
	ExpectedLifetimeMonths *int      `json:"expected_lifetime_months,omitempty"`
	ModelVersion           *string   `json:"model_version" gorm:"type:varchar(50)"`
	PredictedAt            time.Time `json:"predicted_at"`
}

func (LTVPrediction) TableName() string { return "ltv_predictions" }

// NBARecommendation stores a next-best-action recommendation for a customer.
type NBARecommendation struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CustomerID     uuid.UUID       `json:"customer_id" gorm:"type:uuid;not null"`
	TenantID       uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null"`
	ActionType     string          `json:"action_type" gorm:"type:varchar(50);not null"`
	ActionDetail   json.RawMessage `json:"action_detail" gorm:"type:jsonb;default:'{}'"`
	ExpectedImpact float64         `json:"expected_impact" gorm:"not null;default:0"`
	Priority       int             `json:"priority" gorm:"not null;default:0"`
	Status         string          `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	CreatedAt      time.Time       `json:"created_at"`
}

func (NBARecommendation) TableName() string { return "nba_recommendations" }

// PlatformIntegration represents an OAuth connection to an e-commerce platform.
type PlatformIntegration struct {
	ID           uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null;uniqueIndex:idx_tenant_platform"`
	Platform     string          `json:"platform" gorm:"type:varchar(30);not null;uniqueIndex:idx_tenant_platform"`
	AccessToken  string          `json:"access_token,omitempty" gorm:"type:text;not null;default:''"`
	RefreshToken string          `json:"refresh_token,omitempty" gorm:"type:text;not null;default:''"`
	SyncStatus   string          `json:"sync_status" gorm:"type:varchar(20);not null;default:'disconnected'"`
	LastSyncedAt *time.Time      `json:"last_synced_at,omitempty"`
	SyncCursor   json.RawMessage `json:"sync_cursor" gorm:"type:jsonb;default:'{}'"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

func (PlatformIntegration) TableName() string { return "platform_integrations" }

// Order represents an imported e-commerce order.
type Order struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CustomerID      *uuid.UUID      `json:"customer_id,omitempty" gorm:"type:uuid"`
	IntegrationID   uuid.UUID       `json:"integration_id" gorm:"type:uuid;not null;uniqueIndex:idx_integration_platform_order"`
	TenantID        uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null"`
	Platform        string          `json:"platform" gorm:"type:varchar(30);not null"`
	PlatformOrderID string          `json:"platform_order_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_integration_platform_order"`
	Status          string          `json:"status" gorm:"type:varchar(30);not null;default:'pending'"`
	Currency        string          `json:"currency" gorm:"type:varchar(10);not null;default:'USD'"`
	Total           float64         `json:"total" gorm:"type:decimal(12,2);not null;default:0"`
	Items           json.RawMessage `json:"items" gorm:"type:jsonb;default:'[]'"`
	OrderedAt       *time.Time      `json:"ordered_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func (Order) TableName() string { return "orders" }

// OnboardingStep tracks a tenant's onboarding progress.
type OnboardingStep struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID       uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null;uniqueIndex"`
	Step           int             `json:"step" gorm:"not null;default:0"`
	CompletedSteps json.RawMessage `json:"completed_steps" gorm:"type:jsonb;default:'[]'"`
	DemoDataSeeded bool            `json:"demo_data_seeded" gorm:"default:false"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func (OnboardingStep) TableName() string { return "onboarding_steps" }

// TenantBilling holds Stripe billing info for the CRM SaaS plan.
type TenantBilling struct {
	ID                   uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID             uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;uniqueIndex"`
	StripeCustomerID     string     `json:"stripe_customer_id" gorm:"type:varchar(255);not null;default:''"`
	StripeSubscriptionID string     `json:"stripe_subscription_id" gorm:"type:varchar(255);not null;default:''"`
	Plan                 string     `json:"plan" gorm:"type:varchar(50);not null;default:'free'"`
	Status               string     `json:"status" gorm:"type:varchar(30);not null;default:'active'"`
	CurrentPeriodEnd     *time.Time `json:"current_period_end,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func (TenantBilling) TableName() string { return "tenant_billing" }

// ──────────────────────────────────────────────
// Auth DTOs
// ──────────────────────────────────────────────

// LoginRequest is the payload for user authentication.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest is the payload for creating a new user account.
type RegisterRequest struct {
	TenantName string `json:"tenant_name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Name       string `json:"name"`
}

// AuthResponse is returned after successful login or registration.
type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}

// ErrorResponse is a generic error payload.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
