package repository

import (
	"context"
	"fmt"

	"crm0/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, o *model.Order) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *OrderRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.Order, error) {
	var o model.Order
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&o).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return &o, nil
}

func (r *OrderRepository) GetByPlatformOrderID(ctx context.Context, integrationID uuid.UUID, platformOrderID string) (*model.Order, error) {
	var o model.Order
	err := r.db.WithContext(ctx).Where("integration_id = ? AND platform_order_id = ?", integrationID, platformOrderID).First(&o).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order by platform id: %w", err)
	}
	return &o, nil
}

func (r *OrderRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, platform, status string, page, pageSize int) ([]*model.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).Model(&model.Order{}).Where("tenant_id = ?", tenantID)
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	var result []*model.Order
	err := query.Order("ordered_at DESC NULLS LAST").Limit(pageSize).Offset(offset).Find(&result).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}

	return result, int(total), nil
}

func (r *OrderRepository) GetMetrics(ctx context.Context, tenantID uuid.UUID) (*model.OrderMetrics, error) {
	var m model.OrderMetrics
	query := `SELECT COUNT(*), COALESCE(SUM(total), 0), COALESCE(AVG(total), 0) FROM orders WHERE tenant_id = ?`
	if err := r.db.WithContext(ctx).Raw(query, tenantID).Row().Scan(&m.TotalOrders, &m.TotalRevenue, &m.AvgOrderValue); err != nil {
		return nil, fmt.Errorf("failed to get order metrics: %w", err)
	}

	platformQuery := `SELECT platform, COUNT(*), COALESCE(SUM(total), 0) FROM orders WHERE tenant_id = ? GROUP BY platform`
	rows, err := r.db.WithContext(ctx).Raw(platformQuery, tenantID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get platform metrics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pm model.PlatformOrderMetric
		if err := rows.Scan(&pm.Platform, &pm.Count, &pm.Revenue); err != nil {
			return nil, fmt.Errorf("failed to scan platform metric: %w", err)
		}
		m.ByPlatform = append(m.ByPlatform, pm)
	}
	return &m, nil
}
