package repository

import (
	"context"
	"database/sql"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
)

// CustomerRepository provides access to the customers storage.
type CustomerRepository struct {
	db *sql.DB
}

// NewCustomerRepository creates a new CustomerRepository.
func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// Create inserts a new customer row into the database.
func (r *CustomerRepository) Create(ctx context.Context, customer *model.Customer) error {
	query := `
		INSERT INTO customers (id, tenant_id, name, email, company, phone, status, tags, custom_fields, acquired_channel, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	if customer.ID == uuid.Nil {
		customer.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query,
		customer.ID,
		customer.TenantID,
		customer.Name,
		customer.Email,
		toNullString(customer.Company),
		toNullString(customer.Phone),
		customer.Status,
		customer.Tags,
		customer.CustomFields,
		toNullString(customer.AcquiredChannel),
		customer.CreatedAt,
		customer.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

// GetByID retrieves a customer by tenant ID and primary key.
// The tenantID parameter enforces tenant isolation.
func (r *CustomerRepository) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*model.Customer, error) {
	query := `
		SELECT id, tenant_id, name, email, company, phone, status, tags, custom_fields, acquired_channel, created_at, updated_at
		FROM customers
		WHERE tenant_id = $1 AND id = $2
	`
	row := r.db.QueryRowContext(ctx, query, tenantID, id)

	var c model.Customer
	var company, phone, acquiredChannel sql.NullString

	err := row.Scan(
		&c.ID,
		&c.TenantID,
		&c.Name,
		&c.Email,
		&company,
		&phone,
		&c.Status,
		&c.Tags,
		&c.CustomFields,
		&acquiredChannel,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("customer not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	c.Company = company.String
	c.Phone = phone.String
	c.AcquiredChannel = acquiredChannel.String

	return &c, nil
}

// List returns a paginated list of customers for a tenant along with the total count.
func (r *CustomerRepository) List(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*model.Customer, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Count total rows for the tenant.
	var total int
	countQuery := `SELECT COUNT(*) FROM customers WHERE tenant_id = $1`
	if err := r.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	// Fetch the page of results.
	query := `
		SELECT id, tenant_id, name, email, company, phone, status, tags, custom_fields, acquired_channel, created_at, updated_at
		FROM customers
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers: %w", err)
	}
	defer rows.Close()

	customers := make([]*model.Customer, 0)
	for rows.Next() {
		var c model.Customer
		var company, phone, acquiredChannel sql.NullString

		if err := rows.Scan(
			&c.ID,
			&c.TenantID,
			&c.Name,
			&c.Email,
			&company,
			&phone,
			&c.Status,
			&c.Tags,
			&c.CustomFields,
			&acquiredChannel,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan customer row: %w", err)
		}

		c.Company = company.String
		c.Phone = phone.String
		c.AcquiredChannel = acquiredChannel.String

		customers = append(customers, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating customer rows: %w", err)
	}

	return customers, total, nil
}

// Update modifies an existing customer row.
func (r *CustomerRepository) Update(ctx context.Context, customer *model.Customer) error {
	query := `
		UPDATE customers
		SET name = $1, email = $2, company = $3, phone = $4, status = $5,
		    tags = $6, custom_fields = $7, acquired_channel = $8, updated_at = $9
		WHERE id = $10 AND tenant_id = $11
	`
	result, err := r.db.ExecContext(ctx, query,
		customer.Name,
		customer.Email,
		toNullString(customer.Company),
		toNullString(customer.Phone),
		customer.Status,
		customer.Tags,
		customer.CustomFields,
		toNullString(customer.AcquiredChannel),
		customer.UpdatedAt,
		customer.ID,
		customer.TenantID,
	)
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("customer not found for update")
	}

	return nil
}

// Delete removes a customer by tenant ID and primary key.
func (r *CustomerRepository) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM customers WHERE tenant_id = $1 AND id = $2`
	result, err := r.db.ExecContext(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("customer not found for delete")
	}

	return nil
}

// toNullString converts a string to sql.NullString, treating empty strings as NULL.
func toNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}
