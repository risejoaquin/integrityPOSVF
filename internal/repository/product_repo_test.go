package repository

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/model"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set")
	}

	dsn := "postgres://postgres:postgres@localhost:5432/integritypos_test?sslmode=disable"
	if envDsn := os.Getenv("TEST_DB_DSN"); envDsn != "" {
		dsn = envDsn
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}

	// Assuming migrations have already been run against the test db.
	// Otherwise we would execute schemas here.
	return pool
}

func TestProductRepositoryIntegration(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := &ProductRepository{Pool: pool}

	t.Run("CreateReadUpdateDelete", func(t *testing.T) {
		ctx := context.Background()

		// Cleanup before
		pool.Exec(ctx, "DELETE FROM products WHERE name = 'IntTestProduct'")

		p := &model.Product{
			Name:        "IntTestProduct",
			PriceCents:  1000,
			Category:    "Test",
			Stock:       50,
			IsAvailable: true,
		}

		err := repo.Create(ctx, p)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if p.ID == "" {
			t.Fatalf("Expected ID to be set")
		}

		// Read
		fetched, err := repo.GetByID(ctx, p.ID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if fetched.Name != "IntTestProduct" {
			t.Errorf("Expected Name 'IntTestProduct', got %s", fetched.Name)
		}

		// Update
		p.PriceCents = 1500
		err = repo.Update(ctx, p)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		fetched, _ = repo.GetByID(ctx, p.ID)
		if fetched.PriceCents != 1500 {
			t.Errorf("Expected Price 1500, got %d", fetched.PriceCents)
		}

		// Delete
		err = repo.Delete(ctx, p.ID)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = repo.GetByID(ctx, p.ID)
		if err == nil {
			t.Fatalf("Expected error after deletion")
		}
	})
}
