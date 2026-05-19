package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/google/uuid"
)

type OnboardingService struct {
	repo         *repository.OnboardingRepository
	tenantRepo   *repository.TenantRepository
	customerRepo *repository.CustomerRepository
}

func NewOnboardingService(repo *repository.OnboardingRepository, tenantRepo *repository.TenantRepository, customerRepo *repository.CustomerRepository) *OnboardingService {
	return &OnboardingService{repo: repo, tenantRepo: tenantRepo, customerRepo: customerRepo}
}

func (s *OnboardingService) GetStatus(ctx context.Context, tenantID uuid.UUID) (*model.OnboardingStatusResponse, error) {
	o, err := s.repo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		now := time.Now()
		o = &model.OnboardingStep{
			ID:             uuid.New(),
			TenantID:       tenantID,
			Step:           0,
			CompletedSteps: json.RawMessage(`[]`),
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := s.repo.Create(ctx, o); err != nil {
			return nil, err
		}
	}
	return &model.OnboardingStatusResponse{
		TenantID:       o.TenantID,
		Step:           o.Step,
		CompletedSteps: o.CompletedSteps,
		DemoDataSeeded: o.DemoDataSeeded,
		Completed:      o.Step >= 3,
	}, nil
}

func (s *OnboardingService) CompleteStep(ctx context.Context, tenantID uuid.UUID, step int) error {
	o, err := s.repo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}
	if o == nil {
		now := time.Now()
		o = &model.OnboardingStep{
			ID:             uuid.New(),
			TenantID:       tenantID,
			Step:           step,
			CompletedSteps: json.RawMessage(`[]`),
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		return s.repo.Create(ctx, o)
	}

	if step > o.Step {
		o.Step = step
	}
	o.UpdatedAt = time.Now()
	return s.repo.Update(ctx, o)
}

func (s *OnboardingService) SeedDemoData(ctx context.Context, tenantID uuid.UUID) error {
	o, err := s.repo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}
	if o != nil && o.DemoDataSeeded {
		return nil
	}

	channels := []string{"tiktok_shop", "shopify", "amazon", "meta"}
	for i := 0; i < 20; i++ {
		c := &model.Customer{
			ID:              uuid.New(),
			TenantID:        tenantID,
			Name:            fmt.Sprintf("Demo Customer %d", i+1),
			Email:           fmt.Sprintf("demo%d@example.com", i+1),
			Company:         model.PtrString("Demo Corp"),
			Status:          "active",
			Tags:            json.RawMessage(`["demo"]`),
			CustomFields:    json.RawMessage(`{}`),
			AcquiredChannel: model.PtrString(channels[i%len(channels)]),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		if err := s.customerRepo.Create(ctx, c); err != nil {
			return err
		}
	}

	if o == nil {
		now := time.Now()
		o = &model.OnboardingStep{
			ID:             uuid.New(),
			TenantID:       tenantID,
			Step:           2,
			CompletedSteps: json.RawMessage(`[0,1,2]`),
			DemoDataSeeded: true,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		return s.repo.Create(ctx, o)
	}

	o.DemoDataSeeded = true
	o.UpdatedAt = time.Now()
	return s.repo.Update(ctx, o)
}

func (s *OnboardingService) SkipOnboarding(ctx context.Context, tenantID uuid.UUID) error {
	o, err := s.repo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}
	if o == nil {
		now := time.Now()
		o = &model.OnboardingStep{
			ID:             uuid.New(),
			TenantID:       tenantID,
			Step:           3,
			CompletedSteps: json.RawMessage(`[0,1,2,3]`),
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := s.repo.Create(ctx, o); err != nil {
			return err
		}
	} else {
		o.Step = 3
		o.UpdatedAt = time.Now()
		if err := s.repo.Update(ctx, o); err != nil {
			return err
		}
	}

	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}
	tenant.Plan = "free"
	return s.tenantRepo.Update(ctx, tenant)
}
