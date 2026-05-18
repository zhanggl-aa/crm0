package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
)

// SubscriptionRepository provides access to the subscriptions storage.
type SubscriptionRepository struct {
	db *sql.DB
}

// NewSubscriptionRepository creates a new SubscriptionRepository.
func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create inserts a new subscription row into the database.
func (r *SubscriptionRepository) Create(ctx context.Context, sub *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (id, customer_id, plan_id, status, mrr, started_at, canceled_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query,
		sub.ID,
		sub.CustomerID,
		sub.PlanID,
		sub.Status,
		sub.MRR,
		toNullTimePtr(sub.StartedAt),
		toNullTimePtr(sub.CanceledAt),
		sub.CreatedAt,
		sub.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

// GetByID retrieves a subscription by its primary key.
func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	query := `
		SELECT id, customer_id, plan_id, status, mrr, started_at, canceled_at, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var s model.Subscription
	var startedAt, canceledAt sql.NullTime

	err := row.Scan(
		&s.ID,
		&s.CustomerID,
		&s.PlanID,
		&s.Status,
		&s.MRR,
		&startedAt,
		&canceledAt,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("subscription not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if startedAt.Valid {
		s.StartedAt = &startedAt.Time
	}
	if canceledAt.Valid {
		s.CanceledAt = &canceledAt.Time
	}

	return &s, nil
}

// ListByCustomer retrieves all subscriptions for a given customer.
func (r *SubscriptionRepository) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*model.Subscription, error) {
	query := `
		SELECT id, customer_id, plan_id, status, mrr, started_at, canceled_at, created_at, updated_at
		FROM subscriptions
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions by customer: %w", err)
	}
	defer rows.Close()

	subs := make([]*model.Subscription, 0)
	for rows.Next() {
		var s model.Subscription
		var startedAt, canceledAt sql.NullTime

		if err := rows.Scan(
			&s.ID,
			&s.CustomerID,
			&s.PlanID,
			&s.Status,
			&s.MRR,
			&startedAt,
			&canceledAt,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan subscription row: %w", err)
		}

		if startedAt.Valid {
			s.StartedAt = &startedAt.Time
		}
		if canceledAt.Valid {
			s.CanceledAt = &canceledAt.Time
		}

		subs = append(subs, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscription rows: %w", err)
	}

	return subs, nil
}

// ListByTenant returns a paginated list of subscriptions for a tenant along with the total count.
// It joins through the customers table to enforce tenant isolation.
func (r *SubscriptionRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*model.Subscription, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Count total for the tenant.
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM subscriptions s
		JOIN customers c ON c.id = s.customer_id
		WHERE c.tenant_id = $1
	`
	if err := r.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	// Fetch the page of results.
	query := `
		SELECT s.id, s.customer_id, s.plan_id, s.status, s.mrr, s.started_at, s.canceled_at, s.created_at, s.updated_at
		FROM subscriptions s
		JOIN customers c ON c.id = s.customer_id
		WHERE c.tenant_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list subscriptions by tenant: %w", err)
	}
	defer rows.Close()

	subs := make([]*model.Subscription, 0)
	for rows.Next() {
		var s model.Subscription
		var startedAt, canceledAt sql.NullTime

		if err := rows.Scan(
			&s.ID,
			&s.CustomerID,
			&s.PlanID,
			&s.Status,
			&s.MRR,
			&startedAt,
			&canceledAt,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan subscription row: %w", err)
		}

		if startedAt.Valid {
			s.StartedAt = &startedAt.Time
		}
		if canceledAt.Valid {
			s.CanceledAt = &canceledAt.Time
		}

		subs = append(subs, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating subscription rows: %w", err)
	}

	return subs, total, nil
}

// Update modifies an existing subscription row.
func (r *SubscriptionRepository) Update(ctx context.Context, sub *model.Subscription) error {
	query := `
		UPDATE subscriptions
		SET status = $1, mrr = $2, started_at = $3, canceled_at = $4, updated_at = $5
		WHERE id = $6
	`
	result, err := r.db.ExecContext(ctx, query,
		sub.Status,
		sub.MRR,
		toNullTimePtr(sub.StartedAt),
		toNullTimePtr(sub.CanceledAt),
		sub.UpdatedAt,
		sub.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found for update")
	}

	return nil
}

// GetMetrics computes aggregate subscription metrics for a tenant.
// It returns totalMRR, active subscription count, and churn rate.
func (r *SubscriptionRepository) GetMetrics(ctx context.Context, tenantID uuid.UUID) (totalMRR float64, activeCount int, churnRate float64, err error) {
	// Compute total MRR and active subscription count.
	mrrQuery := `
		SELECT COALESCE(SUM(s.mrr), 0), COUNT(*)
		FROM subscriptions s
		JOIN customers c ON c.id = s.customer_id
		WHERE c.tenant_id = $1 AND s.status = 'active'
	`
	if err := r.db.QueryRowContext(ctx, mrrQuery, tenantID).Scan(&totalMRR, &activeCount); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to compute MRR metrics: %w", err)
	}

	// Compute churn rate: ratio of canceled subscriptions to total (non-trial) subscriptions.
	churnQuery := `
		SELECT
			CASE
				WHEN COUNT(*) = 0 THEN 0
				ELSE CAST(COUNT(*) FILTER (WHERE s.status = 'canceled') AS FLOAT) / COUNT(*)
			END
		FROM subscriptions s
		JOIN customers c ON c.id = s.customer_id
		WHERE c.tenant_id = $1 AND s.status IN ('active', 'canceled')
	`
	if err := r.db.QueryRowContext(ctx, churnQuery, tenantID).Scan(&churnRate); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to compute churn rate: %w", err)
	}

	return totalMRR, activeCount, churnRate, nil
}

// toNullTimePtr converts a *time.Time to sql.NullTime, treating nil as NULL.
func toNullTimePtr(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{
		Time:  *t,
		Valid: true,
	}
}
