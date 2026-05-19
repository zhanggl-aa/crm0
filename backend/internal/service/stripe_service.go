package service

import (
	"context"
	"time"

	"crm0/backend/internal/config"
	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/google/uuid"
)

type StripeService struct {
	repo      *repository.BillingRepository
	tenantRepo *repository.TenantRepository
	cfg       *config.Config
}

func NewStripeService(repo *repository.BillingRepository, tenantRepo *repository.TenantRepository, cfg *config.Config) *StripeService {
	return &StripeService{repo: repo, tenantRepo: tenantRepo, cfg: cfg}
}

func (s *StripeService) GetBillingInfo(ctx context.Context, tenantID uuid.UUID) (*model.BillingInfoResponse, error) {
	b, err := s.repo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return &model.BillingInfoResponse{
			TenantID: tenantID,
			Plan:     "free",
			Status:   "active",
		}, nil
	}
	return &model.BillingInfoResponse{
		TenantID:         b.TenantID,
		Plan:             b.Plan,
		Status:           b.Status,
		StripeCustomerID: b.StripeCustomerID,
		CurrentPeriodEnd: b.CurrentPeriodEnd,
	}, nil
}

func (s *StripeService) CreateCheckoutSession(ctx context.Context, tenantID uuid.UUID, priceID string) (string, error) {
	if s.cfg.Stripe.SecretKey == "" {
		return "", nil
	}

	b, err := s.repo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return "", err
	}

	now := time.Now()
	if b == nil {
		b = &model.TenantBilling{
			ID:                   uuid.New(),
			TenantID:             tenantID,
			StripeCustomerID:     "",
			StripeSubscriptionID: "",
			Plan:                 "free",
			Status:               "active",
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if err := s.repo.Create(ctx, b); err != nil {
			return "", err
		}
	}

	return b.StripeCustomerID, nil
}

func (s *StripeService) CreatePortalSession(ctx context.Context, tenantID uuid.UUID) (string, error) {
	if s.cfg.Stripe.SecretKey == "" {
		return "", nil
	}
	b, err := s.repo.GetByTenantID(ctx, tenantID)
	if err != nil || b == nil {
		return "", err
	}
	return b.StripeCustomerID, nil
}

func (s *StripeService) HandleWebhook(ctx context.Context, eventType string, data map[string]any) error {
	switch eventType {
	case "checkout.session.completed":
		customerID, _ := data["customer"].(string)
		_, _ = data["subscription"].(string)
		if customerID == "" {
			return nil
		}
		return nil

	case "customer.subscription.updated", "customer.subscription.created":
		subID, _ := data["id"].(string)
		if subID == "" {
			return nil
		}
		b, err := s.repo.GetByStripeSubscriptionID(ctx, subID)
		if err != nil || b == nil {
			return err
		}
		b.Status = "active"
		b.UpdatedAt = time.Now()
		return s.repo.Update(ctx, b)

	case "customer.subscription.deleted":
		subID, _ := data["id"].(string)
		if subID == "" {
			return nil
		}
		b, err := s.repo.GetByStripeSubscriptionID(ctx, subID)
		if err != nil || b == nil {
			return err
		}
		b.Status = "canceled"
		b.Plan = "free"
		b.UpdatedAt = time.Now()
		if err := s.repo.Update(ctx, b); err != nil {
			return err
		}
		tenant, err := s.tenantRepo.GetByID(ctx, b.TenantID)
		if err != nil {
			return err
		}
		tenant.Plan = "free"
		return s.tenantRepo.Update(ctx, tenant)
	}
	return nil
}
