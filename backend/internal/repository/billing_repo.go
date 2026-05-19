package repository

import (
	"context"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BillingRepository struct {
	db *gorm.DB
}

func NewBillingRepository(db *gorm.DB) *BillingRepository {
	return &BillingRepository{db: db}
}

func (r *BillingRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*model.TenantBilling, error) {
	var b model.TenantBilling
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&b).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get billing: %w", err)
	}
	return &b, nil
}

func (r *BillingRepository) GetByStripeSubscriptionID(ctx context.Context, subID string) (*model.TenantBilling, error) {
	var b model.TenantBilling
	err := r.db.WithContext(ctx).Where("stripe_subscription_id = ?", subID).First(&b).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get billing by stripe sub: %w", err)
	}
	return &b, nil
}

func (r *BillingRepository) Create(ctx context.Context, b *model.TenantBilling) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *BillingRepository) Update(ctx context.Context, b *model.TenantBilling) error {
	result := r.db.WithContext(ctx).Model(b).Where("tenant_id = ?", b.TenantID).
		Select("stripe_customer_id", "stripe_subscription_id", "plan", "status", "current_period_end").
		Updates(b)
	if result.Error != nil {
		return fmt.Errorf("failed to update billing: %w", result.Error)
	}
	return nil
}
