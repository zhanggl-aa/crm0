package repository

import (
	"context"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlanRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

func (r *PlanRepository) Create(ctx context.Context, plan *model.Plan) error {
	if plan.ID == uuid.Nil {
		plan.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *PlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Plan, error) {
	var p model.Plan
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("plan not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}
	return &p, nil
}

func (r *PlanRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.Plan, error) {
	var plans []*model.Plan
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&plans).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list plans by tenant: %w", err)
	}
	return plans, nil
}

func (r *PlanRepository) Update(ctx context.Context, plan *model.Plan) error {
	result := r.db.WithContext(ctx).Model(plan).Select("name", "price", "billing_cycle").Updates(plan)
	if result.Error != nil {
		return fmt.Errorf("failed to update plan: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("plan not found for update")
	}
	return nil
}

func (r *PlanRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Plan{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete plan: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("plan not found for delete")
	}
	return nil
}
