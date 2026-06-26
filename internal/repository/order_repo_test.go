package repository

import (
	"context"
	"testing"

	"github.com/solidbit/integritypos/internal/model"
)

func TestOrderRepositoryIntegration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewOrderRepository(pool)

	t.Run("CreateAndReadOrder", func(t *testing.T) {
		ctx := context.Background()

		order := &model.Order{
			Status:        "pending",
			Source:        "pos",
			CustomerName:  "Test Customer",
			CustomerPhone: "1234567890",
			TotalCents:    2000,
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("BeginTx failed: %v", err)
		}

		err = repo.Create(ctx, tx, order)
		if err != nil {
			tx.Rollback(context.Background())
			t.Fatalf("Create failed: %v", err)
		}
		
		err = tx.Commit(ctx)
		if err != nil {
			t.Fatalf("Commit failed: %v", err)
		}

		if order.ID == "" {
			t.Fatalf("Expected ID to be set")
		}

		// Read
		fetched, err := repo.GetByID(ctx, pool, order.ID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if fetched.CustomerName != "Test Customer" {
			t.Errorf("Expected name 'Test Customer', got %s", fetched.CustomerName)
		}

		// Update Status
		err = repo.UpdateStatusTx(ctx, pool, order.ID, "completed")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		fetched, _ = repo.GetByID(ctx, pool, order.ID)
		if fetched.Status != "completed" {
			t.Errorf("Expected status 'completed', got %s", fetched.Status)
		}
	})
}
