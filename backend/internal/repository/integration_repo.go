package repository

import (
	"context"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IntegrationRepository struct {
	db *gorm.DB
}

func NewIntegrationRepository(db *gorm.DB) *IntegrationRepository {
	return &IntegrationRepository{db: db}
}

func (r *IntegrationRepository) GetByTenantAndPlatform(ctx context.Context, tenantID uuid.UUID, platform string) (*model.PlatformIntegration, error) {
	var i model.PlatformIntegration
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND platform = ?", tenantID, platform).First(&i).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get integration: %w", err)
	}
	return &i, nil
}

func (r *IntegrationRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.PlatformIntegration, error) {
	var result []*model.PlatformIntegration
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at").Find(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list integrations: %w", err)
	}
	return result, nil
}

func (r *IntegrationRepository) Create(ctx context.Context, i *model.PlatformIntegration) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(i).Error
}

func (r *IntegrationRepository) Update(ctx context.Context, i *model.PlatformIntegration) error {
	result := r.db.WithContext(ctx).Model(i).
		Select("access_token", "refresh_token", "sync_status", "last_synced_at", "sync_cursor").
		Updates(i)
	if result.Error != nil {
		return fmt.Errorf("failed to update integration: %w", result.Error)
	}
	return nil
}

func (r *IntegrationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.PlatformIntegration{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete integration: %w", result.Error)
	}
	return nil
}
