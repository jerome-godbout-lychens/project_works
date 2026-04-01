package postgres

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

// NewConnection opens a PostgreSQL connection pool with the given configuration.
// It returns a *sql.DB that manages the connection pool.
func NewConnection(dsn string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) (*sql.DB, error) {
	database, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	database.SetMaxOpenConns(maxOpenConns)
	database.SetMaxIdleConns(maxIdleConns)
	database.SetConnMaxLifetime(connMaxLifetime)

	// Test the connection
	err = database.Ping()
	if err != nil {
		return nil, err
	}

	return database, nil
}

// RunMigrations applies pending migrations from the given directory.
// It uses goose to manage schema migrations.
func RunMigrations(db *sql.DB, migrationsDir string) error {
	return goose.Up(db, migrationsDir)
}
