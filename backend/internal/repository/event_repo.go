package repository

import (
	"context"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, event *model.UserEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *EventRepository) ListByCustomer(ctx context.Context, customerID uuid.UUID, page, pageSize int) ([]*model.UserEvent, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := r.db.WithContext(ctx).Model(&model.UserEvent{}).Where("customer_id = ?", customerID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	var events []*model.UserEvent
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).
		Order("occurred_at DESC").
		Limit(pageSize).Offset(offset).
		Find(&events).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list events by customer: %w", err)
	}

	return events, int(total), nil
}

func (r *EventRepository) GetRecentByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]*model.UserEvent, error) {
	if limit < 1 {
		limit = 10
	}

	var events []*model.UserEvent
	err := r.db.WithContext(ctx).
		Joins("JOIN customers ON customers.id = user_events.customer_id").
		Where("customers.tenant_id = ?", tenantID).
		Order("user_events.occurred_at DESC").
		Limit(limit).
		Find(&events).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get recent events by tenant: %w", err)
	}

	return events, nil
}
