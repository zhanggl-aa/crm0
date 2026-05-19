package service

import (
	"context"
	"encoding/json"

	"crm0/backend/internal/algorithm"
	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/google/uuid"
)

type CustomerService struct {
	repo      *repository.CustomerRepository
	eventRepo *repository.EventRepository
	algoRepo  *repository.AlgorithmRepository
	algo      *algorithm.Client
}

func NewCustomerService(repo *repository.CustomerRepository, eventRepo *repository.EventRepository, algoRepo *repository.AlgorithmRepository, algo *algorithm.Client) *CustomerService {
	return &CustomerService{repo: repo, eventRepo: eventRepo, algoRepo: algoRepo, algo: algo}
}

func (s *CustomerService) List(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*model.Customer, int, error) {
	return s.repo.List(ctx, tenantID, page, pageSize)
}

func (s *CustomerService) Get(ctx context.Context, tenantID, id uuid.UUID) (*model.Customer, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *CustomerService) Create(ctx context.Context, tenantID uuid.UUID, req *model.CreateCustomerRequest) (*model.Customer, error) {
	c := &model.Customer{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Name:         req.Name,
		Email:        req.Email,
		Company:      strPtr(req.Company),
		Phone:        strPtr(req.Phone),
		Status:       req.Status,
		Tags:         req.Tags,
		CustomFields: req.CustomFields,
		AcquiredChannel: strPtr(req.AcquiredChannel),
	}
	if c.Status == "" {
		c.Status = "active"
	}
	if c.Tags == nil {
		c.Tags = json.RawMessage(`[]`)
	}
	if c.CustomFields == nil {
		c.CustomFields = json.RawMessage(`{}`)
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CustomerService) Update(ctx context.Context, tenantID, id uuid.UUID, req *model.UpdateCustomerRequest) (*model.Customer, error) {
	c, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Email != nil {
		c.Email = *req.Email
	}
	if req.Company != nil {
		c.Company = req.Company
	}
	if req.Phone != nil {
		c.Phone = req.Phone
	}
	if req.Status != nil {
		c.Status = *req.Status
	}
	if req.Tags != nil {
		c.Tags = req.Tags
	}
	if req.CustomFields != nil {
		c.CustomFields = req.CustomFields
	}
	if req.AcquiredChannel != nil {
		c.AcquiredChannel = req.AcquiredChannel
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CustomerService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.Delete(ctx, tenantID, id)
}

func (s *CustomerService) GetInsights(ctx context.Context, tenantID, customerID uuid.UUID) (*model.CustomerInsightResponse, error) {
	customer, err := s.repo.GetByID(ctx, tenantID, customerID)
	if err != nil {
		return nil, err
	}
	insight := &model.CustomerInsightResponse{Customer: *customer}

	if churn, _ := s.algoRepo.GetChurnPredictionByCustomer(ctx, customerID); churn != nil {
		insight.ChurnPrediction = churn
	}
	if ltvs, err := s.algoRepo.GetLTVPredictions(ctx, tenantID); err == nil {
		for _, ltv := range ltvs {
			if ltv.CustomerID == customerID {
				insight.LTV = ltv
				break
			}
		}
	}
	if segments, err := s.algoRepo.GetCustomerSegments(ctx, tenantID); err == nil {
		for _, seg := range segments {
			if seg.CustomerID == customerID {
				insight.Segments = append(insight.Segments, *seg)
			}
		}
	}
	if recs, err := s.algoRepo.GetNBARecommendations(ctx, tenantID); err == nil {
		for _, rec := range recs {
			if rec.CustomerID == customerID {
				insight.Recommendations = append(insight.Recommendations, *rec)
			}
		}
	}
	return insight, nil
}

func (s *CustomerService) GetEvents(ctx context.Context, customerID uuid.UUID, page, pageSize int) ([]*model.UserEvent, int, error) {
	return s.eventRepo.ListByCustomer(ctx, customerID, page, pageSize)
}

// strPtr returns a *string for non-empty strings, nil for empty.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
