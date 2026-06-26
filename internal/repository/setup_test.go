package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	if os.Getenv("INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration test, INTEGRATION_TESTS!=1")
	}

	dsn := "postgres://localhost:5432/integritypos_test?sslmode=disable"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// Read and execute migrations
	migrationBytes, err := os.ReadFile("../../migrations/001_core.sql")
	if err != nil {
		t.Fatalf("failed to read migration: %v", err)
	}

	_, err = pool.Exec(context.Background(), string(migrationBytes))
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	cleanup := func() {
		pool.Exec(context.Background(), "DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
		pool.Close()
	}

	return pool, cleanup
}
