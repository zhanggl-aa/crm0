package repository

import (
	"context"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AlgorithmRepository struct {
	db *gorm.DB
}

func NewAlgorithmRepository(db *gorm.DB) *AlgorithmRepository {
	return &AlgorithmRepository{db: db}
}

// ─── Churn Predictions ────────────────────────────────────────────────────────

func (r *AlgorithmRepository) SaveChurnPrediction(ctx context.Context, prediction *model.ChurnPrediction) error {
	if prediction.ID == uuid.Nil {
		prediction.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(prediction).Error
}

func (r *AlgorithmRepository) GetChurnPredictions(ctx context.Context, tenantID uuid.UUID) ([]*model.ChurnPrediction, error) {
	var predictions []*model.ChurnPrediction
	err := r.db.WithContext(ctx).
		Joins("JOIN customers ON customers.id = churn_predictions.customer_id").
		Where("customers.tenant_id = ?", tenantID).
		Order("churn_predictions.risk_score DESC").
		Find(&predictions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get churn predictions: %w", err)
	}
	return predictions, nil
}

func (r *AlgorithmRepository) GetChurnPredictionByCustomer(ctx context.Context, customerID uuid.UUID) (*model.ChurnPrediction, error) {
	var p model.ChurnPrediction
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("predicted_at DESC").First(&p).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("churn prediction not found for customer")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get churn prediction by customer: %w", err)
	}
	return &p, nil
}

// ─── Customer Segments ────────────────────────────────────────────────────────

func (r *AlgorithmRepository) SaveCustomerSegment(ctx context.Context, segment *model.CustomerSegment) error {
	if segment.ID == uuid.Nil {
		segment.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(segment).Error
}

func (r *AlgorithmRepository) GetCustomerSegments(ctx context.Context, tenantID uuid.UUID) ([]*model.CustomerSegment, error) {
	var segments []*model.CustomerSegment
	err := r.db.WithContext(ctx).
		Joins("JOIN customers ON customers.id = customer_segments.customer_id").
		Where("customers.tenant_id = ?", tenantID).
		Order("customer_segments.updated_at DESC").
		Find(&segments).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get customer segments: %w", err)
	}
	return segments, nil
}

// ─── LTV Predictions ──────────────────────────────────────────────────────────

func (r *AlgorithmRepository) SaveLTVPrediction(ctx context.Context, prediction *model.LTVPrediction) error {
	if prediction.ID == uuid.Nil {
		prediction.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(prediction).Error
}

func (r *AlgorithmRepository) GetLTVPredictions(ctx context.Context, tenantID uuid.UUID) ([]*model.LTVPrediction, error) {
	var predictions []*model.LTVPrediction
	err := r.db.WithContext(ctx).
		Joins("JOIN customers ON customers.id = ltv_predictions.customer_id").
		Where("customers.tenant_id = ?", tenantID).
		Order("ltv_predictions.predicted_ltv DESC").
		Find(&predictions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get LTV predictions: %w", err)
	}
	return predictions, nil
}

// ─── NBA Recommendations ──────────────────────────────────────────────────────

func (r *AlgorithmRepository) SaveNBARecommendation(ctx context.Context, rec *model.NBARecommendation) error {
	if rec.ID == uuid.Nil {
		rec.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *AlgorithmRepository) GetNBARecommendations(ctx context.Context, tenantID uuid.UUID) ([]*model.NBARecommendation, error) {
	var recommendations []*model.NBARecommendation
	err := r.db.WithContext(ctx).
		Joins("JOIN customers ON customers.id = nba_recommendations.customer_id").
		Where("customers.tenant_id = ?", tenantID).
		Order("nba_recommendations.priority DESC, nba_recommendations.created_at DESC").
		Find(&recommendations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get NBA recommendations: %w", err)
	}
	return recommendations, nil
}

// ─── Dashboard Overview ───────────────────────────────────────────────────────

func (r *AlgorithmRepository) GetDashboardOverview(ctx context.Context, tenantID uuid.UUID) (*model.DashboardOverview, error) {
	var o model.DashboardOverview

	// Customer counts
	customerQuery := `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'active')
		FROM customers WHERE tenant_id = ?
	`
	if err := r.db.WithContext(ctx).Raw(customerQuery, tenantID).Row().Scan(&o.TotalCustomers, &o.ActiveCustomers); err != nil {
		return nil, fmt.Errorf("failed to compute customer counts: %w", err)
	}

	// Total MRR
	mrrQuery := `
		SELECT COALESCE(SUM(s.mrr), 0)
		FROM subscriptions s
		JOIN customers c ON c.id = s.customer_id
		WHERE c.tenant_id = ? AND s.status = 'active'
	`
	if err := r.db.WithContext(ctx).Raw(mrrQuery, tenantID).Row().Scan(&o.MRR); err != nil {
		return nil, fmt.Errorf("failed to compute total MRR: %w", err)
	}

	// Churn rate
	churnQuery := `
		SELECT
			CASE
				WHEN COUNT(*) = 0 THEN 0
				ELSE CAST(COUNT(*) FILTER (WHERE s.status = 'canceled') AS FLOAT) / COUNT(*)
			END
		FROM subscriptions s
		JOIN customers c ON c.id = s.customer_id
		WHERE c.tenant_id = ? AND s.status IN ('active', 'canceled')
	`
	if err := r.db.WithContext(ctx).Raw(churnQuery, tenantID).Row().Scan(&o.ChurnRate); err != nil {
		return nil, fmt.Errorf("failed to compute churn rate: %w", err)
	}

	// Average LTV
	ltvQuery := `
		SELECT COALESCE(AVG(lp.predicted_ltv), 0)
		FROM ltv_predictions lp
		JOIN customers c ON c.id = lp.customer_id
		WHERE c.tenant_id = ?
	`
	if err := r.db.WithContext(ctx).Raw(ltvQuery, tenantID).Row().Scan(&o.AvgLTV); err != nil {
		return nil, fmt.Errorf("failed to compute average LTV: %w", err)
	}

	// High risk count
	highRiskQuery := `
		SELECT COUNT(DISTINCT cp.customer_id)
		FROM churn_predictions cp
		JOIN customers c ON c.id = cp.customer_id
		WHERE c.tenant_id = ? AND cp.risk_level = 'high'
	`
	if err := r.db.WithContext(ctx).Raw(highRiskQuery, tenantID).Row().Scan(&o.HighRiskCount); err != nil {
		return nil, fmt.Errorf("failed to compute high risk count: %w", err)
	}

	// Pending NBA actions
	pendingQuery := `
		SELECT COUNT(*)
		FROM nba_recommendations nr
		JOIN customers c ON c.id = nr.customer_id
		WHERE c.tenant_id = ? AND nr.status = 'pending'
	`
	if err := r.db.WithContext(ctx).Raw(pendingQuery, tenantID).Row().Scan(&o.PendingActions); err != nil {
		return nil, fmt.Errorf("failed to compute pending actions: %w", err)
	}

	return &o, nil
}
