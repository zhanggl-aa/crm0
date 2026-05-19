package repository

import (
	"context"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Create(ctx context.Context, customer *model.Customer) error {
	if customer.ID == uuid.Nil {
		customer.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(customer).Error
}

func (r *CustomerRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.Customer, error) {
	var c model.Customer
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&c).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("customer not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	return &c, nil
}

func (r *CustomerRepository) List(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*model.Customer, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Customer{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	var customers []*model.Customer
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(pageSize).Offset(offset).
		Find(&customers).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers: %w", err)
	}

	return customers, int(total), nil
}

func (r *CustomerRepository) Update(ctx context.Context, customer *model.Customer) error {
	result := r.db.WithContext(ctx).Model(customer).
		Select("name", "email", "company", "phone", "status", "tags", "custom_fields", "acquired_channel").
		Updates(customer)
	if result.Error != nil {
		return fmt.Errorf("failed to update customer: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("customer not found for update")
	}
	return nil
}

func (r *CustomerRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&model.Customer{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete customer: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("customer not found for delete")
	}
	return nil
}
