package repository

import (
	"context"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OnboardingRepository struct {
	db *gorm.DB
}

func NewOnboardingRepository(db *gorm.DB) *OnboardingRepository {
	return &OnboardingRepository{db: db}
}

func (r *OnboardingRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*model.OnboardingStep, error) {
	var o model.OnboardingStep
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&o).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get onboarding: %w", err)
	}
	return &o, nil
}

func (r *OnboardingRepository) Create(ctx context.Context, o *model.OnboardingStep) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *OnboardingRepository) Update(ctx context.Context, o *model.OnboardingStep) error {
	result := r.db.WithContext(ctx).Model(o).Where("tenant_id = ?", o.TenantID).
		Select("step", "completed_steps", "demo_data_seeded").Updates(o)
	if result.Error != nil {
		return fmt.Errorf("failed to update onboarding: %w", result.Error)
	}
	return nil
}
