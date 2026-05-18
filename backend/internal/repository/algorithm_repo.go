package repository

import (
	"context"
	"database/sql"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
)

// AlgorithmRepository provides access to the algorithm output tables
// (churn_predictions, customer_segments, ltv_predictions, nba_recommendations)
// and the dashboard overview aggregate query.
type AlgorithmRepository struct {
	db *sql.DB
}

// NewAlgorithmRepository creates a new AlgorithmRepository.
func NewAlgorithmRepository(db *sql.DB) *AlgorithmRepository {
	return &AlgorithmRepository{db: db}
}

// ──────────────────────────────────────────────
// Churn Predictions
// ──────────────────────────────────────────────

// SaveChurnPrediction inserts a new churn prediction row.
func (r *AlgorithmRepository) SaveChurnPrediction(ctx context.Context, prediction *model.ChurnPrediction) error {
	query := `
		INSERT INTO churn_predictions (id, customer_id, risk_score, risk_level, factors, predicted_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if prediction.ID == uuid.Nil {
		prediction.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query,
		prediction.ID,
		prediction.CustomerID,
		prediction.RiskScore,
		prediction.RiskLevel,
		prediction.Factors,
		prediction.PredictedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save churn prediction: %w", err)
	}

	return nil
}

// GetChurnPredictions retrieves all churn predictions for customers belonging to a tenant.
func (r *AlgorithmRepository) GetChurnPredictions(ctx context.Context, tenantID uuid.UUID) ([]*model.ChurnPrediction, error) {
	query := `
		SELECT cp.id, cp.customer_id, cp.risk_score, cp.risk_level, cp.factors, cp.predicted_at
		FROM churn_predictions cp
		JOIN customers c ON c.id = cp.customer_id
		WHERE c.tenant_id = $1
		ORDER BY cp.risk_score DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get churn predictions: %w", err)
	}
	defer rows.Close()

	predictions := make([]*model.ChurnPrediction, 0)
	for rows.Next() {
		var p model.ChurnPrediction
		if err := rows.Scan(
			&p.ID,
			&p.CustomerID,
			&p.RiskScore,
			&p.RiskLevel,
			&p.Factors,
			&p.PredictedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan churn prediction row: %w", err)
		}
		predictions = append(predictions, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating churn prediction rows: %w", err)
	}

	return predictions, nil
}

// GetChurnPredictionByCustomer retrieves the latest churn prediction for a specific customer.
func (r *AlgorithmRepository) GetChurnPredictionByCustomer(ctx context.Context, customerID uuid.UUID) (*model.ChurnPrediction, error) {
	query := `
		SELECT id, customer_id, risk_score, risk_level, factors, predicted_at
		FROM churn_predictions
		WHERE customer_id = $1
		ORDER BY predicted_at DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, customerID)

	var p model.ChurnPrediction
	err := row.Scan(
		&p.ID,
		&p.CustomerID,
		&p.RiskScore,
		&p.RiskLevel,
		&p.Factors,
		&p.PredictedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("churn prediction not found for customer: %w", err)
		}
		return nil, fmt.Errorf("failed to get churn prediction by customer: %w", err)
	}

	return &p, nil
}

// ──────────────────────────────────────────────
// Customer Segments
// ──────────────────────────────────────────────

// SaveCustomerSegment inserts a new customer segment row.
func (r *AlgorithmRepository) SaveCustomerSegment(ctx context.Context, segment *model.CustomerSegment) error {
	query := `
		INSERT INTO customer_segments (id, customer_id, segment_type, segment_name, score, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if segment.ID == uuid.Nil {
		segment.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query,
		segment.ID,
		segment.CustomerID,
		segment.SegmentType,
		segment.SegmentName,
		segment.Score,
		segment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save customer segment: %w", err)
	}

	return nil
}

// GetCustomerSegments retrieves all customer segments for a tenant.
func (r *AlgorithmRepository) GetCustomerSegments(ctx context.Context, tenantID uuid.UUID) ([]*model.CustomerSegment, error) {
	query := `
		SELECT cs.id, cs.customer_id, cs.segment_type, cs.segment_name, cs.score, cs.updated_at
		FROM customer_segments cs
		JOIN customers c ON c.id = cs.customer_id
		WHERE c.tenant_id = $1
		ORDER BY cs.updated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer segments: %w", err)
	}
	defer rows.Close()

	segments := make([]*model.CustomerSegment, 0)
	for rows.Next() {
		var s model.CustomerSegment
		if err := rows.Scan(
			&s.ID,
			&s.CustomerID,
			&s.SegmentType,
			&s.SegmentName,
			&s.Score,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan customer segment row: %w", err)
		}
		segments = append(segments, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customer segment rows: %w", err)
	}

	return segments, nil
}

// ──────────────────────────────────────────────
// LTV Predictions
// ──────────────────────────────────────────────

// SaveLTVPrediction inserts a new LTV prediction row.
func (r *AlgorithmRepository) SaveLTVPrediction(ctx context.Context, prediction *model.LTVPrediction) error {
	query := `
		INSERT INTO ltv_predictions (id, customer_id, predicted_ltv, confidence, expected_lifetime_months, model_version, predicted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	if prediction.ID == uuid.Nil {
		prediction.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query,
		prediction.ID,
		prediction.CustomerID,
		prediction.PredictedLTV,
		prediction.Confidence,
		toNullIntPtr(prediction.ExpectedLifetimeMonths),
		prediction.ModelVersion,
		prediction.PredictedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save LTV prediction: %w", err)
	}

	return nil
}

// GetLTVPredictions retrieves all LTV predictions for a tenant.
func (r *AlgorithmRepository) GetLTVPredictions(ctx context.Context, tenantID uuid.UUID) ([]*model.LTVPrediction, error) {
	query := `
		SELECT lp.id, lp.customer_id, lp.predicted_ltv, lp.confidence, lp.expected_lifetime_months, lp.model_version, lp.predicted_at
		FROM ltv_predictions lp
		JOIN customers c ON c.id = lp.customer_id
		WHERE c.tenant_id = $1
		ORDER BY lp.predicted_ltv DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get LTV predictions: %w", err)
	}
	defer rows.Close()

	predictions := make([]*model.LTVPrediction, 0)
	for rows.Next() {
		var p model.LTVPrediction
		var expectedLifetimeMonths sql.NullInt64
		var modelVersion sql.NullString

		if err := rows.Scan(
			&p.ID,
			&p.CustomerID,
			&p.PredictedLTV,
			&p.Confidence,
			&expectedLifetimeMonths,
			&modelVersion,
			&p.PredictedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan LTV prediction row: %w", err)
		}

		if expectedLifetimeMonths.Valid {
			v := int(expectedLifetimeMonths.Int64)
			p.ExpectedLifetimeMonths = &v
		}
		p.ModelVersion = modelVersion.String

		predictions = append(predictions, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating LTV prediction rows: %w", err)
	}

	return predictions, nil
}

// ──────────────────────────────────────────────
// NBA Recommendations
// ──────────────────────────────────────────────

// SaveNBARecommendation inserts a new NBA recommendation row.
func (r *AlgorithmRepository) SaveNBARecommendation(ctx context.Context, rec *model.NBARecommendation) error {
	query := `
		INSERT INTO nba_recommendations (id, customer_id, action_type, action_detail, expected_impact, priority, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if rec.ID == uuid.Nil {
		rec.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query,
		rec.ID,
		rec.CustomerID,
		rec.ActionType,
		rec.ActionDetail,
		rec.ExpectedImpact,
		rec.Priority,
		rec.Status,
		rec.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save NBA recommendation: %w", err)
	}

	return nil
}

// GetNBARecommendations retrieves all NBA recommendations for a tenant.
func (r *AlgorithmRepository) GetNBARecommendations(ctx context.Context, tenantID uuid.UUID) ([]*model.NBARecommendation, error) {
	query := `
		SELECT nr.id, nr.customer_id, nr.action_type, nr.action_detail, nr.expected_impact, nr.priority, nr.status, nr.created_at
		FROM nba_recommendations nr
		JOIN customers c ON c.id = nr.customer_id
		WHERE c.tenant_id = $1
		ORDER BY nr.priority DESC, nr.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get NBA recommendations: %w", err)
	}
	defer rows.Close()

	recommendations := make([]*model.NBARecommendation, 0)
	for rows.Next() {
		var rec model.NBARecommendation
		if err := rows.Scan(
			&rec.ID,
			&rec.CustomerID,
			&rec.ActionType,
			&rec.ActionDetail,
			&rec.ExpectedImpact,
			&rec.Priority,
			&rec.Status,
			&rec.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan NBA recommendation row: %w", err)
		}
		recommendations = append(recommendations, &rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating NBA recommendation rows: %w", err)
	}

	return recommendations, nil
}

// ──────────────────────────────────────────────
// Dashboard Overview
// ──────────────────────────────────────────────

// GetDashboardOverview computes an aggregate dashboard overview for a tenant.
// It queries multiple tables to build a unified snapshot of key metrics.
func (r *AlgorithmRepository) GetDashboardOverview(ctx context.Context, tenantID uuid.UUID) (*model.DashboardOverview, error) {
	var o model.DashboardOverview

	// Customer counts.
	customerQuery := `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'active')
		FROM customers
		WHERE tenant_id = $1
	`
	if err := r.db.QueryRowContext(ctx, customerQuery, tenantID).Scan(&o.TotalCustomers, &o.ActiveCustomers); err != nil {
		return nil, fmt.Errorf("failed to compute customer counts: %w", err)
	}

	// Total MRR from active subscriptions.
	mrrQuery := `
		SELECT COALESCE(SUM(s.mrr), 0)
		FROM subscriptions s
		JOIN customers c ON c.id = s.customer_id
		WHERE c.tenant_id = $1 AND s.status = 'active'
	`
	if err := r.db.QueryRowContext(ctx, mrrQuery, tenantID).Scan(&o.MRR); err != nil {
		return nil, fmt.Errorf("failed to compute total MRR: %w", err)
	}

	// Churn rate.
	churnQuery := `
		SELECT
			CASE
				WHEN COUNT(*) = 0 THEN 0
				ELSE CAST(COUNT(*) FILTER (WHERE s.status = 'canceled') AS FLOAT) / COUNT(*)
			END
		FROM subscriptions s
		JOIN customers c ON c.id = s.customer_id
		WHERE c.tenant_id = $1 AND s.status IN ('active', 'canceled')
	`
	if err := r.db.QueryRowContext(ctx, churnQuery, tenantID).Scan(&o.ChurnRate); err != nil {
		return nil, fmt.Errorf("failed to compute churn rate: %w", err)
	}

	// Average LTV.
	ltvQuery := `
		SELECT COALESCE(AVG(lp.predicted_ltv), 0)
		FROM ltv_predictions lp
		JOIN customers c ON c.id = lp.customer_id
		WHERE c.tenant_id = $1
	`
	if err := r.db.QueryRowContext(ctx, ltvQuery, tenantID).Scan(&o.AvgLTV); err != nil {
		return nil, fmt.Errorf("failed to compute average LTV: %w", err)
	}

	// High risk customer count (risk_level = 'high').
	highRiskQuery := `
		SELECT COUNT(DISTINCT cp.customer_id)
		FROM churn_predictions cp
		JOIN customers c ON c.id = cp.customer_id
		WHERE c.tenant_id = $1 AND cp.risk_level = 'high'
	`
	if err := r.db.QueryRowContext(ctx, highRiskQuery, tenantID).Scan(&o.HighRiskCount); err != nil {
		return nil, fmt.Errorf("failed to compute high risk count: %w", err)
	}

	// Pending NBA actions.
	pendingQuery := `
		SELECT COUNT(*)
		FROM nba_recommendations nr
		JOIN customers c ON c.id = nr.customer_id
		WHERE c.tenant_id = $1 AND nr.status = 'pending'
	`
	if err := r.db.QueryRowContext(ctx, pendingQuery, tenantID).Scan(&o.PendingActions); err != nil {
		return nil, fmt.Errorf("failed to compute pending actions: %w", err)
	}

	return &o, nil
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

// toNullIntPtr converts a *int to sql.NullInt64, treating nil as NULL.
func toNullIntPtr(i *int) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{
		Int64: int64(*i),
		Valid: true,
	}
}
