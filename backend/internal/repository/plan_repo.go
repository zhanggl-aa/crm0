package repository

import (
	"context"
	"database/sql"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
)

// PlanRepository provides access to the plans storage.
type PlanRepository struct {
	db *sql.DB
}

// NewPlanRepository creates a new PlanRepository.
func NewPlanRepository(db *sql.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

// Create inserts a new plan row into the database.
func (r *PlanRepository) Create(ctx context.Context, plan *model.Plan) error {
	query := `
		INSERT INTO plans (id, tenant_id, name, price, billing_cycle, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	if plan.ID == uuid.Nil {
		plan.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query,
		plan.ID,
		plan.TenantID,
		plan.Name,
		plan.Price,
		plan.BillingCycle,
		plan.CreatedAt,
		plan.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	return nil
}

// GetByID retrieves a plan by its primary key.
func (r *PlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Plan, error) {
	query := `
		SELECT id, tenant_id, name, price, billing_cycle, created_at, updated_at
		FROM plans
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var p model.Plan
	err := row.Scan(
		&p.ID,
		&p.TenantID,
		&p.Name,
		&p.Price,
		&p.BillingCycle,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("plan not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	return &p, nil
}

// ListByTenant retrieves all plans for a given tenant.
func (r *PlanRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.Plan, error) {
	query := `
		SELECT id, tenant_id, name, price, billing_cycle, created_at, updated_at
		FROM plans
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list plans by tenant: %w", err)
	}
	defer rows.Close()

	plans := make([]*model.Plan, 0)
	for rows.Next() {
		var p model.Plan
		if err := rows.Scan(
			&p.ID,
			&p.TenantID,
			&p.Name,
			&p.Price,
			&p.BillingCycle,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plan row: %w", err)
		}
		plans = append(plans, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plan rows: %w", err)
	}

	return plans, nil
}

// Update modifies an existing plan row.
func (r *PlanRepository) Update(ctx context.Context, plan *model.Plan) error {
	query := `
		UPDATE plans
		SET name = $1, price = $2, billing_cycle = $3, updated_at = $4
		WHERE id = $5
	`
	result, err := r.db.ExecContext(ctx, query,
		plan.Name,
		plan.Price,
		plan.BillingCycle,
		plan.UpdatedAt,
		plan.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("plan not found for update")
	}

	return nil
}

// Delete removes a plan by its primary key.
func (r *PlanRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM plans WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("plan not found for delete")
	}

	return nil
}
