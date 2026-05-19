package service

import (
	"context"
	"encoding/json"
	"time"

	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/google/uuid"
)

type EventService struct {
	repo *repository.EventRepository
}

func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) Create(ctx context.Context, tenantID uuid.UUID, req *model.TrackEventRequest) (*model.UserEvent, error) {
	e := &model.UserEvent{ID: uuid.New(), CustomerID: req.CustomerID, TenantID: tenantID, EventType: req.EventType, Properties: req.Properties}
	if req.OccurredAt != nil {
		e.OccurredAt = *req.OccurredAt
	} else {
		e.OccurredAt = time.Now()
	}
	if e.Properties == nil {
		e.Properties = json.RawMessage(`{}`)
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *EventService) ListByCustomer(ctx context.Context, customerID uuid.UUID, page, pageSize int) ([]*model.UserEvent, int, error) {
	return s.repo.ListByCustomer(ctx, customerID, page, pageSize)
}

func (s *EventService) GetRecentByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]*model.UserEvent, error) {
	return s.repo.GetRecentByTenant(ctx, tenantID, limit)
}
