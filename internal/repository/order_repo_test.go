package repository

import (
	"context"
	"testing"

	"github.com/solidbit/integritypos/internal/model"
)

func TestOrderRepositoryIntegration(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := &OrderRepository{Pool: pool}

	t.Run("CreateAndReadOrder", func(t *testing.T) {
		ctx := context.Background()

		order := &model.Order{
			Status:        "pending",
			Source:        "pos",
			CustomerName:  "Test Customer",
			CustomerPhone: "1234567890",
			TotalCents:    2000,
		}

		tx, err := repo.BeginTx(ctx)
		if err != nil {
			t.Fatalf("BeginTx failed: %v", err)
		}

		err = repo.Create(ctx, tx, order)
		if err != nil {
			tx.Rollback(ctx)
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
		fetched, err := repo.GetByID(ctx, order.ID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if fetched.CustomerName != "Test Customer" {
			t.Errorf("Expected name 'Test Customer', got %s", fetched.CustomerName)
		}

		// Update Status
		err = repo.UpdateStatus(ctx, order.ID, "completed")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		fetched, _ = repo.GetByID(ctx, order.ID)
		if fetched.Status != "completed" {
			t.Errorf("Expected status 'completed', got %s", fetched.Status)
		}

		// Cleanup
		pool.Exec(ctx, "DELETE FROM orders WHERE id = $1", order.ID)
	})
}
