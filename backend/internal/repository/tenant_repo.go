package repository

import (
	"context"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) Create(ctx context.Context, tenant *model.Tenant) error {
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error) {
	var t model.Tenant
	err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return &t, nil
}

func (r *TenantRepository) Update(ctx context.Context, tenant *model.Tenant) error {
	result := r.db.WithContext(ctx).Model(tenant).Select("name", "plan", "settings").Updates(tenant)
	if result.Error != nil {
		return fmt.Errorf("failed to update tenant: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("tenant not found for update")
	}
	return nil
}

func (r *TenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Tenant{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete tenant: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("tenant not found for delete")
	}
	return nil
}

func (r *TenantRepository) List(ctx context.Context, page, pageSize int) ([]*model.Tenant, error) {
	offset := (page - 1) * pageSize
	var tenants []*model.Tenant
	err := r.db.WithContext(ctx).Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&tenants).Error
	return tenants, err
}
