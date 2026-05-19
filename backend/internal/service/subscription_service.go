package service

import (
	"context"

	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/google/uuid"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

func NewSubscriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) List(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*model.Subscription, int, error) {
	return s.repo.ListByTenant(ctx, tenantID, page, pageSize)
}

func (s *SubscriptionService) Get(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) Create(ctx context.Context, req *model.CreateSubscriptionRequest) (*model.Subscription, error) {
	sub := &model.Subscription{ID: uuid.New(), CustomerID: req.CustomerID, PlanID: req.PlanID, Status: "trial"}
	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *SubscriptionService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateSubscriptionRequest) (*model.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Status != nil {
		sub.Status = *req.Status
	}
	if req.PlanID != nil {
		sub.PlanID = *req.PlanID
	}
	if req.CanceledAt != nil {
		sub.CanceledAt = req.CanceledAt
	}
	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	sub.Status = "canceled"
	return s.repo.Update(ctx, sub)
}

func (s *SubscriptionService) GetMetrics(ctx context.Context, tenantID uuid.UUID) (*model.SubscriptionMetrics, error) {
	totalMRR, activeCount, _, err := s.repo.GetMetrics(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return &model.SubscriptionMetrics{TotalMRR: totalMRR, ActiveCount: activeCount}, nil
}
