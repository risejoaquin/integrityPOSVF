package repository

import (
	"context"
	"testing"

	"github.com/solidbit/integritypos/internal/model"
)

func TestProductRepositoryIntegration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewProductRepository(pool)
	ctx := context.Background()

	t.Run("CreateReadUpdateDelete", func(t *testing.T) {
		p := &model.Product{
			Name:       "Integration Test Product",
			PriceCents: 1500,
			Category:   "Test",
			Stock:      10,
		}

		err := repo.Create(ctx, p)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if p.ID == "" {
			t.Fatalf("Expected ID")
		}

		fetched, err := repo.GetByID(ctx, p.ID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if fetched.Name != "Integration Test Product" {
			t.Errorf("Expected name 'Integration Test Product', got %s", fetched.Name)
		}

		p.Name = "Updated Product"
		err = repo.Update(ctx, p)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		err = repo.DecrementStockAtomic(ctx, pool, p.ID, -5)
		if err != nil {
			t.Fatalf("DecrementStockAtomic failed: %v", err)
		}
		fetched2, _ := repo.GetByID(ctx, p.ID)
		if fetched2.Stock != 5 {
			t.Errorf("Expected stock 5, got %d", fetched2.Stock)
		}

		err = repo.Delete(ctx, p.ID)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
	})
}
