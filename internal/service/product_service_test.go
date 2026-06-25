package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/repository"
)

type mockProductSvcRepo struct {
	products map[string]model.Product
	txFunc   func(ctx context.Context) (repository.Tx, error)
}

func (m *mockProductSvcRepo) List(ctx context.Context, filter repository.ProductFilter) ([]model.Product, error) {
	return nil, nil
}

func (m *mockProductSvcRepo) GetByID(ctx context.Context, id string) (model.Product, error) {
	if p, ok := m.products[id]; ok {
		return p, nil
	}
	return model.Product{}, fmt.Errorf("not found")
}

func (m *mockProductSvcRepo) Create(ctx context.Context, p *model.Product) error { return nil }
func (m *mockProductSvcRepo) Update(ctx context.Context, p *model.Product) error { return nil }
func (m *mockProductSvcRepo) Delete(ctx context.Context, id string) error        { return nil }

func (m *mockProductSvcRepo) GetStock(ctx context.Context, id string) (int, error) {
	if p, ok := m.products[id]; ok {
		return p.Stock, nil
	}
	return 0, fmt.Errorf("not found")
}

func (m *mockProductSvcRepo) UpdateStock(ctx context.Context, tx repository.Tx, id string, delta int) error {
	if p, ok := m.products[id]; ok {
		p.Stock += delta
		m.products[id] = p
		return nil
	}
	return fmt.Errorf("not found")
}

func (m *mockProductSvcRepo) BeginTx(ctx context.Context) (repository.Tx, error) {
	if m.txFunc != nil {
		return m.txFunc(ctx)
	}
	return &mockTx{}, nil
}

type mockInvSvcRepo struct{}

func (m *mockInvSvcRepo) RecordMovement(ctx context.Context, tx repository.Tx, productID string, delta int, reason string, orderID string) error {
	return nil
}
func (m *mockInvSvcRepo) GetMovements(ctx context.Context, productID string, limit int) ([]model.InventoryMovement, error) {
	return nil, nil
}

func TestAdjustStock(t *testing.T) {
	productRepo := &mockProductSvcRepo{
		products: map[string]model.Product{
			"p1": {ID: "p1", Name: "Burger", PriceCents: 500, Stock: 10},
		},
	}
	invRepo := &mockInvSvcRepo{}

	svc := &ProductService{
		Repo:          productRepo,
		InventoryRepo: invRepo,
	}

	t.Run("positive adjustment", func(t *testing.T) {
		err := svc.AdjustStock(context.Background(), "p1", 5, "restock")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if productRepo.products["p1"].Stock != 15 {
			t.Errorf("expected stock 15, got %d", productRepo.products["p1"].Stock)
		}
	})

	t.Run("negative adjustment sufficient stock", func(t *testing.T) {
		err := svc.AdjustStock(context.Background(), "p1", -5, "sale")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if productRepo.products["p1"].Stock != 10 {
			t.Errorf("expected stock 10, got %d", productRepo.products["p1"].Stock)
		}
	})

	t.Run("negative adjustment insufficient stock", func(t *testing.T) {
		err := svc.AdjustStock(context.Background(), "p1", -20, "sale")
		if err == nil {
			t.Errorf("expected error for insufficient stock")
		}
	})

	t.Run("negative adjustment reason initial", func(t *testing.T) {
		err := svc.AdjustStock(context.Background(), "p1", -20, "inventario inicial")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if productRepo.products["p1"].Stock != -10 {
			t.Errorf("expected stock -10, got %d", productRepo.products["p1"].Stock)
		}
	})
}
