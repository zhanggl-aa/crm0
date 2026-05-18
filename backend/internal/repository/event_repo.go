package repository

import (
	"context"
	"database/sql"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
)

// EventRepository provides access to the user_events storage.
type EventRepository struct {
	db *sql.DB
}

// NewEventRepository creates a new EventRepository.
func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

// Create inserts a new user event row into the database.
func (r *EventRepository) Create(ctx context.Context, event *model.UserEvent) error {
	query := `
		INSERT INTO user_events (id, customer_id, event_type, properties, occurred_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query,
		event.ID,
		event.CustomerID,
		event.EventType,
		event.Properties,
		event.OccurredAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// ListByCustomer returns a paginated list of events for a given customer along with the total count.
func (r *EventRepository) ListByCustomer(ctx context.Context, customerID uuid.UUID, page, pageSize int) ([]*model.UserEvent, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Count total for the customer.
	var total int
	countQuery := `SELECT COUNT(*) FROM user_events WHERE customer_id = $1`
	if err := r.db.QueryRowContext(ctx, countQuery, customerID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// Fetch the page of results.
	query := `
		SELECT id, customer_id, event_type, properties, occurred_at
		FROM user_events
		WHERE customer_id = $1
		ORDER BY occurred_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, customerID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list events by customer: %w", err)
	}
	defer rows.Close()

	events := make([]*model.UserEvent, 0)
	for rows.Next() {
		var e model.UserEvent
		if err := rows.Scan(
			&e.ID,
			&e.CustomerID,
			&e.EventType,
			&e.Properties,
			&e.OccurredAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan event row: %w", err)
		}
		events = append(events, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating event rows: %w", err)
	}

	return events, total, nil
}

// GetRecentByTenant retrieves the most recent events for a tenant, joining through
// the customers table to enforce tenant isolation.
func (r *EventRepository) GetRecentByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]*model.UserEvent, error) {
	if limit < 1 {
		limit = 10
	}

	query := `
		SELECT e.id, e.customer_id, e.event_type, e.properties, e.occurred_at
		FROM user_events e
		JOIN customers c ON c.id = e.customer_id
		WHERE c.tenant_id = $1
		ORDER BY e.occurred_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent events by tenant: %w", err)
	}
	defer rows.Close()

	events := make([]*model.UserEvent, 0)
	for rows.Next() {
		var e model.UserEvent
		if err := rows.Scan(
			&e.ID,
			&e.CustomerID,
			&e.EventType,
			&e.Properties,
			&e.OccurredAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}
		events = append(events, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event rows: %w", err)
	}

	return events, nil
}
