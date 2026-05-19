package repository

import (
	"context"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *model.Subscription) error {
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	var s model.Subscription
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("subscription not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return &s, nil
}

func (r *SubscriptionRepository) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*model.Subscription, error) {
	var subs []*model.Subscription
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("created_at DESC").Find(&subs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions by customer: %w", err)
	}
	return subs, nil
}

func (r *SubscriptionRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*model.Subscription, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := r.db.WithContext(ctx).
		Model(&model.Subscription{}).
		Joins("JOIN customers ON customers.id = subscriptions.customer_id").
		Where("customers.tenant_id = ?", tenantID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	var subs []*model.Subscription
	err := r.db.WithContext(ctx).
		Joins("JOIN customers ON customers.id = subscriptions.customer_id").
		Where("customers.tenant_id = ?", tenantID).
		Order("subscriptions.created_at DESC").
		Limit(pageSize).Offset(offset).
		Find(&subs).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list subscriptions by tenant: %w", err)
	}

	return subs, int(total), nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *model.Subscription) error {
	result := r.db.WithContext(ctx).Model(sub).Select("status", "mrr", "started_at", "canceled_at").Updates(sub)
	if result.Error != nil {
		return fmt.Errorf("failed to update subscription: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription not found for update")
	}
	return nil
}

func (r *SubscriptionRepository) GetMetrics(ctx context.Context, tenantID uuid.UUID) (totalMRR float64, activeCount int, churnRate float64, err error) {
	// Total MRR and active count
	mrrQuery := `
		SELECT COALESCE(SUM(s.mrr), 0), COUNT(*)
		FROM subscriptions s
		JOIN customers c ON c.id = s.customer_id
		WHERE c.tenant_id = ? AND s.status = 'active'
	`
	if err := r.db.WithContext(ctx).Raw(mrrQuery, tenantID).Row().Scan(&totalMRR, &activeCount); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to compute MRR metrics: %w", err)
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
	if err := r.db.WithContext(ctx).Raw(churnQuery, tenantID).Row().Scan(&churnRate); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to compute churn rate: %w", err)
	}

	return totalMRR, activeCount, churnRate, nil
}
