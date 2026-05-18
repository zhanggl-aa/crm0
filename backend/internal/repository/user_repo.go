package repository

import (
	"context"
	"database/sql"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
)

// UserRepository provides access to the users storage.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user row into the database.
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, tenant_id, email, password_hash, name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.TenantID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByEmail retrieves a user by tenant ID and email address.
func (r *UserRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*model.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, name, role, created_at, updated_at
		FROM users
		WHERE tenant_id = $1 AND email = $2
	`
	row := r.db.QueryRowContext(ctx, query, tenantID, email)

	var u model.User
	err := row.Scan(
		&u.ID,
		&u.TenantID,
		&u.Email,
		&u.PasswordHash,
		&u.Name,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &u, nil
}

// GetByID retrieves a user by its primary key.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, name, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var u model.User
	err := row.Scan(
		&u.ID,
		&u.TenantID,
		&u.Email,
		&u.PasswordHash,
		&u.Name,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &u, nil
}
