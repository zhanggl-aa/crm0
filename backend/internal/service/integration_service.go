package service

import (
	"context"
	"encoding/json"
	"time"

	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/google/uuid"
)

type IntegrationService struct {
	repo        *repository.IntegrationRepository
	customerRepo *repository.CustomerRepository
	orderRepo   *repository.OrderRepository
}

func NewIntegrationService(repo *repository.IntegrationRepository, customerRepo *repository.CustomerRepository, orderRepo *repository.OrderRepository) *IntegrationService {
	return &IntegrationService{repo: repo, customerRepo: customerRepo, orderRepo: orderRepo}
}

func (s *IntegrationService) List(ctx context.Context, tenantID uuid.UUID) ([]*model.PlatformIntegrationResponse, error) {
	integrations, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]*model.PlatformIntegrationResponse, len(integrations))
	for i, ig := range integrations {
		result[i] = &model.PlatformIntegrationResponse{
			ID:           ig.ID,
			TenantID:     ig.TenantID,
			Platform:     ig.Platform,
			SyncStatus:   ig.SyncStatus,
			LastSyncedAt: ig.LastSyncedAt,
			CreatedAt:    ig.CreatedAt,
			UpdatedAt:    ig.UpdatedAt,
		}
	}
	return result, nil
}

func (s *IntegrationService) Connect(ctx context.Context, tenantID uuid.UUID, platform, code string) (*model.PlatformIntegrationResponse, error) {
	existing, err := s.repo.GetByTenantAndPlatform(ctx, tenantID, platform)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if existing != nil {
		existing.AccessToken = code
		existing.SyncStatus = "connected"
		existing.UpdatedAt = now
		if err := s.repo.Update(ctx, existing); err != nil {
			return nil, err
		}
		return &model.PlatformIntegrationResponse{
			ID: existing.ID, TenantID: existing.TenantID, Platform: existing.Platform,
			SyncStatus: existing.SyncStatus, LastSyncedAt: existing.LastSyncedAt,
			CreatedAt: existing.CreatedAt, UpdatedAt: existing.UpdatedAt,
		}, nil
	}

	ig := &model.PlatformIntegration{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Platform:     platform,
		AccessToken:  code,
		SyncStatus:   "connected",
		SyncCursor:   json.RawMessage(`{}`),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.Create(ctx, ig); err != nil {
		return nil, err
	}
	return &model.PlatformIntegrationResponse{
		ID: ig.ID, TenantID: ig.TenantID, Platform: ig.Platform,
		SyncStatus: ig.SyncStatus, CreatedAt: ig.CreatedAt, UpdatedAt: ig.UpdatedAt,
	}, nil
}

func (s *IntegrationService) Disconnect(ctx context.Context, tenantID uuid.UUID, platform string) error {
	ig, err := s.repo.GetByTenantAndPlatform(ctx, tenantID, platform)
	if err != nil {
		return err
	}
	if ig == nil {
		return nil
	}
	return s.repo.Delete(ctx, ig.ID)
}

func (s *IntegrationService) TriggerSync(ctx context.Context, tenantID uuid.UUID, platform string) error {
	ig, err := s.repo.GetByTenantAndPlatform(ctx, tenantID, platform)
	if err != nil {
		return err
	}
	if ig == nil {
		return nil
	}

	now := time.Now()
	ig.SyncStatus = "syncing"
	ig.UpdatedAt = now
	if err := s.repo.Update(ctx, ig); err != nil {
		return err
	}

	ig.LastSyncedAt = &now
	ig.SyncStatus = "connected"
	ig.UpdatedAt = now
	return s.repo.Update(ctx, ig)
}
