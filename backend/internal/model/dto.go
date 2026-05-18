package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ──────────────────────────────────────────────
// Customer DTOs
// ──────────────────────────────────────────────

// CreateCustomerRequest is the payload for creating a new customer.
type CreateCustomerRequest struct {
	Name            string          `json:"name"`
	Email           string          `json:"email"`
	Company         string          `json:"company,omitempty"`
	Phone           string          `json:"phone,omitempty"`
	Status          string          `json:"status,omitempty"`
	Tags            json.RawMessage `json:"tags,omitempty"`
	CustomFields    json.RawMessage `json:"custom_fields,omitempty"`
	AcquiredChannel string          `json:"acquired_channel,omitempty"`
}

// UpdateCustomerRequest is the payload for patching an existing customer.
type UpdateCustomerRequest struct {
	Name            *string         `json:"name,omitempty"`
	Email           *string         `json:"email,omitempty"`
	Company         *string         `json:"company,omitempty"`
	Phone           *string         `json:"phone,omitempty"`
	Status          *string         `json:"status,omitempty"`
	Tags            json.RawMessage `json:"tags,omitempty"`
	CustomFields    json.RawMessage `json:"custom_fields,omitempty"`
	AcquiredChannel *string         `json:"acquired_channel,omitempty"`
}

// CustomerInsightResponse aggregates a customer with their analytics data.
type CustomerInsightResponse struct {
	Customer       Customer               `json:"customer"`
	ChurnPrediction *ChurnPrediction       `json:"churn_prediction,omitempty"`
	LTV            *LTVPrediction         `json:"ltv,omitempty"`
	Segments       []CustomerSegment      `json:"segments,omitempty"`
	Recommendations []NBARecommendation   `json:"recommendations,omitempty"`
}

// ──────────────────────────────────────────────
// Subscription DTOs
// ──────────────────────────────────────────────

// CreateSubscriptionRequest is the payload for creating a new subscription.
type CreateSubscriptionRequest struct {
	CustomerID uuid.UUID `json:"customer_id"`
	PlanID     uuid.UUID `json:"plan_id"`
}

// UpdateSubscriptionRequest is the payload for updating a subscription.
type UpdateSubscriptionRequest struct {
	Status     *string    `json:"status,omitempty"`
	PlanID     *uuid.UUID `json:"plan_id,omitempty"`
	CanceledAt *time.Time `json:"canceled_at,omitempty"`
}

// SubscriptionMetrics contains aggregate metrics for subscriptions.
type SubscriptionMetrics struct {
	TotalMRR          float64 `json:"total_mrr"`
	ActiveCount       int     `json:"active_count"`
	CanceledCount     int     `json:"canceled_count"`
	TrialCount        int     `json:"trial_count"`
	AvgMRRPerCustomer float64 `json:"avg_mrr_per_customer"`
}

// ──────────────────────────────────────────────
// Plan DTOs
// ──────────────────────────────────────────────

// CreatePlanRequest is the payload for creating a new billing plan.
type CreatePlanRequest struct {
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	BillingCycle string  `json:"billing_cycle"`
}

// UpdatePlanRequest is the payload for updating a billing plan.
type UpdatePlanRequest struct {
	Name         *string  `json:"name,omitempty"`
	Price        *float64 `json:"price,omitempty"`
	BillingCycle *string  `json:"billing_cycle,omitempty"`
}

// ──────────────────────────────────────────────
// Event DTOs
// ──────────────────────────────────────────────

// TrackEventRequest is the payload for recording a customer event.
type TrackEventRequest struct {
	CustomerID uuid.UUID       `json:"customer_id"`
	EventType  string          `json:"event_type"`
	Properties json.RawMessage `json:"properties,omitempty"`
	OccurredAt *time.Time      `json:"occurred_at,omitempty"`
}

// ──────────────────────────────────────────────
// Analytics response DTOs
// ──────────────────────────────────────────────

// ChurnPredictionResponse is the API response for a churn prediction.
type ChurnPredictionResponse struct {
	ID          uuid.UUID       `json:"id"`
	CustomerID  uuid.UUID       `json:"customer_id"`
	RiskScore   float64         `json:"risk_score"`
	RiskLevel   string          `json:"risk_level"`
	Factors     json.RawMessage `json:"factors"`
	PredictedAt time.Time       `json:"predicted_at"`
}

// SegmentResponse is the API response for a customer segment.
type SegmentResponse struct {
	ID          uuid.UUID `json:"id"`
	CustomerID  uuid.UUID `json:"customer_id"`
	SegmentType string    `json:"segment_type"`
	SegmentName string    `json:"segment_name"`
	Score       float64   `json:"score"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LTVResponse is the API response for an LTV prediction.
type LTVResponse struct {
	ID                     uuid.UUID `json:"id"`
	CustomerID             uuid.UUID `json:"customer_id"`
	PredictedLTV           float64   `json:"predicted_ltv"`
	Confidence             float64   `json:"confidence"`
	ExpectedLifetimeMonths int       `json:"expected_lifetime_months"`
	ModelVersion           string    `json:"model_version"`
	PredictedAt            time.Time `json:"predicted_at"`
}

// NBARecommendationResponse is the API response for a next-best-action.
type NBARecommendationResponse struct {
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
// Dashboard DTOs
// ──────────────────────────────────────────────

// DashboardOverview provides high-level metrics for the tenant dashboard.
type DashboardOverview struct {
	TotalCustomers  int              `json:"total_customers"`
	ActiveCustomers int              `json:"active_customers"`
	MRR             float64          `json:"mrr"`
	ChurnRate       float64          `json:"churn_rate"`
	AvgLTV          float64          `json:"avg_ltv"`
	HighRiskCount   int              `json:"high_risk_count"`
	PendingActions  int              `json:"pending_actions"`
	RecentAlerts    []DashboardAlert `json:"recent_alerts,omitempty"`
}

// DashboardAlert represents a single alert item on the dashboard.
type DashboardAlert struct {
	ID      uuid.UUID `json:"id"`
	Type    string    `json:"type"`
	Message string    `json:"message"`
	Severity string  `json:"severity"`
	Time    time.Time `json:"time"`
}

// ──────────────────────────────────────────────
// Algorithm task DTOs
// ──────────────────────────────────────────────

// AlgorithmTaskResponse represents the status/result of an async algorithm task.
type AlgorithmTaskResponse struct {
	TaskID string          `json:"task_id"`
	Status string          `json:"status"`
	Result json.RawMessage `json:"result,omitempty"`
}
