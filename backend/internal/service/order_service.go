package service

import (
	"context"

	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/google/uuid"
)

type OrderService struct {
	repo *repository.OrderRepository
}

func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) List(ctx context.Context, tenantID uuid.UUID, platform, status string, page, pageSize int) ([]*model.Order, int, error) {
	return s.repo.ListByTenant(ctx, tenantID, platform, status, page, pageSize)
}

func (s *OrderService) Get(ctx context.Context, tenantID, id uuid.UUID) (*model.Order, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *OrderService) GetMetrics(ctx context.Context, tenantID uuid.UUID) (*model.OrderMetrics, error) {
	return s.repo.GetMetrics(ctx, tenantID)
}
