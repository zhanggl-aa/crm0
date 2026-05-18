package repository

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"crm0/backend/internal/config"
)

// NewDB opens a connection to PostgreSQL using the provided configuration
// and verifies connectivity with a Ping.
func NewDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// RunMigrations reads and executes the initial migration SQL file.
func RunMigrations(db *sql.DB) error {
	content, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	return nil
}
