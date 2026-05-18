package repository

import (
	"context"
	"database/sql"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
)

// TenantRepository provides access to the tenants storage.
type TenantRepository struct {
	db *sql.DB
}

// NewTenantRepository creates a new TenantRepository.
func NewTenantRepository(db *sql.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// Create inserts a new tenant row into the database.
func (r *TenantRepository) Create(ctx context.Context, tenant *model.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, plan, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Plan,
		tenant.Settings,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	return nil
}

// GetByID retrieves a tenant by its primary key.
func (r *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error) {
	query := `
		SELECT id, name, plan, settings, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var t model.Tenant
	err := row.Scan(
		&t.ID,
		&t.Name,
		&t.Plan,
		&t.Settings,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &t, nil
}
