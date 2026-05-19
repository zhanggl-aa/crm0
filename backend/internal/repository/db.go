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
	// Rename existing unique constraints to GORM-expected names so AutoMigrate
	// doesn't fail trying to drop constraints that exist under different names.
	alignConstraints(db)

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

// alignConstraints renames existing unique constraints to match GORM's expected naming
// convention (uni_<table>_<column>) so AutoMigrate doesn't fail.
func alignConstraints(db *gorm.DB) {
	type rename struct {
		table  string
		column string
	}
	renames := []rename{
		{"onboarding_steps", "tenant_id"},
		{"tenant_billing", "tenant_id"},
	}
	for _, r := range renames {
		gormName := fmt.Sprintf("uni_%s_%s", r.table, r.column)
		var currentName string
		db.Raw(`
			SELECT conname FROM pg_constraint c
			JOIN pg_class t ON c.conrelid = t.oid
			JOIN pg_namespace n ON t.relnamespace = n.oid
			WHERE t.relname = ? AND c.contype = 'u' AND conname != ?
			AND EXISTS (
				SELECT 1 FROM pg_attribute a WHERE a.attrelid = t.oid AND a.attnum = ANY(c.conkey) AND a.attname = ?
			)
			LIMIT 1`, r.table, gormName, r.column).Scan(&currentName)
		if currentName != "" {
			db.Exec(fmt.Sprintf("ALTER TABLE %s RENAME CONSTRAINT %s TO %s", r.table, currentName, gormName))
		}
	}
}
