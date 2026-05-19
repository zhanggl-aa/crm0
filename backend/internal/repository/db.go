package repository

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"crm0/backend/internal/model"
)

// NewDB opens a GORM connection to PostgreSQL using the provided DSN
// and verifies connectivity.
func NewDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	return db, nil
}

// RunMigrations uses GORM AutoMigrate to create/update tables for all domain models.
func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Tenant{},
		&model.User{},
		&model.Customer{},
		&model.Plan{},
		&model.Subscription{},
		&model.UserEvent{},
		&model.ChurnPrediction{},
		&model.CustomerSegment{},
		&model.LTVPrediction{},
		&model.NBARecommendation{},
		&model.PlatformIntegration{},
		&model.Order{},
		&model.OnboardingStep{},
		&model.TenantBilling{},
	)
}
