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
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Plan      string          `json:"plan"`
	Settings  json.RawMessage `json:"settings"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// User represents an authenticated user belonging to a tenant.
type User struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Customer represents a customer record within a tenant.
type Customer struct {
	ID              uuid.UUID       `json:"id"`
	TenantID        uuid.UUID       `json:"tenant_id"`
	Name            string          `json:"name"`
	Email           string          `json:"email"`
	Company         string          `json:"company"`
	Phone           string          `json:"phone"`
	Status          string          `json:"status"`
	Tags            json.RawMessage `json:"tags"`
	CustomFields    json.RawMessage `json:"custom_fields"`
	AcquiredChannel string          `json:"acquired_channel"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// Plan represents a billing plan available to customers.
type Plan struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	Name          string    `json:"name"`
	Price         float64   `json:"price"`
	BillingCycle  string    `json:"billing_cycle"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Subscription links a customer to a plan and tracks subscription state.
type Subscription struct {
	ID         uuid.UUID  `json:"id"`
	CustomerID uuid.UUID  `json:"customer_id"`
	PlanID     uuid.UUID  `json:"plan_id"`
	Status     string     `json:"status"`
	MRR        float64    `json:"mrr"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	CanceledAt *time.Time `json:"canceled_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// UserEvent represents a tracked event performed by or for a customer.
type UserEvent struct {
	ID          uuid.UUID       `json:"id"`
	CustomerID  uuid.UUID       `json:"customer_id"`
	EventType   string          `json:"event_type"`
	Properties  json.RawMessage `json:"properties"`
	OccurredAt  time.Time       `json:"occurred_at"`
}

// ChurnPrediction stores the output of a churn-risk prediction for a customer.
type ChurnPrediction struct {
	ID          uuid.UUID       `json:"id"`
	CustomerID  uuid.UUID       `json:"customer_id"`
	RiskScore   float64         `json:"risk_score"`
	RiskLevel   string          `json:"risk_level"`
	Factors     json.RawMessage `json:"factors"`
	PredictedAt time.Time       `json:"predicted_at"`
}

// CustomerSegment records a customer's membership in an analytics segment.
type CustomerSegment struct {
	ID          uuid.UUID `json:"id"`
	CustomerID  uuid.UUID `json:"customer_id"`
	SegmentType string    `json:"segment_type"`
	SegmentName string    `json:"segment_name"`
	Score       float64   `json:"score"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LTVPrediction stores the predicted lifetime value for a customer.
type LTVPrediction struct {
	ID                     uuid.UUID `json:"id"`
	CustomerID             uuid.UUID `json:"customer_id"`
	PredictedLTV           float64   `json:"predicted_ltv"`
	Confidence             float64   `json:"confidence"`
	ExpectedLifetimeMonths *int      `json:"expected_lifetime_months,omitempty"`
	ModelVersion           string    `json:"model_version"`
	PredictedAt            time.Time `json:"predicted_at"`
}

// NBARecommendation stores a next-best-action recommendation for a customer.
type NBARecommendation struct {
	ID             uuid.UUID       `json:"id"`
	CustomerID     uuid.UUID       `json:"customer_id"`
	ActionType     string          `json:"action_type"`
	ActionDetail   json.RawMessage `json:"action_detail"`
	ExpectedImpact float64         `json:"expected_impact"`
	Priority       int             `json:"priority"`
	Status         string          `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
}

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
