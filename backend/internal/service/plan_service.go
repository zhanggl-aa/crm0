package service

import (
	"context"

	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/google/uuid"
)

type PlanService struct {
	repo *repository.PlanRepository
}

func NewPlanService(repo *repository.PlanRepository) *PlanService {
	return &PlanService{repo: repo}
}

func (s *PlanService) List(ctx context.Context, tenantID uuid.UUID) ([]*model.Plan, error) {
	return s.repo.ListByTenant(ctx, tenantID)
}

func (s *PlanService) Get(ctx context.Context, id uuid.UUID) (*model.Plan, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PlanService) Create(ctx context.Context, tenantID uuid.UUID, req *model.CreatePlanRequest) (*model.Plan, error) {
	p := &model.Plan{ID: uuid.New(), TenantID: tenantID, Name: req.Name, Price: req.Price, BillingCycle: req.BillingCycle}
	if p.BillingCycle == "" {
		p.BillingCycle = "monthly"
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PlanService) Update(ctx context.Context, id uuid.UUID, req *model.UpdatePlanRequest) (*model.Plan, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.BillingCycle != nil {
		p.BillingCycle = *req.BillingCycle
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PlanService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
