package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/repository"
)

func TestCreateProduct_InvalidName(t *testing.T) {
	svc := &ProductService{}
	err := svc.Create(context.Background(), &model.Product{Name: "   ", PriceCents: 100})
	if !errors.Is(err, model.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestCreateProduct_InvalidPrice(t *testing.T) {
	svc := &ProductService{}
	err := svc.Create(context.Background(), &model.Product{Name: "Burger", PriceCents: -100})
	if !errors.Is(err, model.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestAdjustStock_Decrement_Success(t *testing.T) {
	db := &mockDBBeginner{tx: &mockTx{}}
	productRepo := &mockProductRepo{
		getStockFn: func(ctx context.Context, db repository.DBTX, id string) (int, error) {
			return 10, nil
		},
		decrementStockAtomicFn: func(ctx context.Context, db repository.DBTX, productID string, quantity int) error {
			return nil
		},
	}
	invRepo := &mockInventoryRepo{}

	svc := &ProductService{
		DB:            db,
		Repo:          productRepo,
		InventoryRepo: invRepo,
	}

	err := svc.AdjustStock(context.Background(), db, "p1", -5, "venta")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestAdjustStock_Decrement_Insufficient(t *testing.T) {
	db := &mockDBBeginner{tx: &mockTx{}}
	productRepo := &mockProductRepo{
		getStockFn: func(ctx context.Context, db repository.DBTX, id string) (int, error) {
			return 10, nil
		},
		decrementStockAtomicFn: func(ctx context.Context, db repository.DBTX, productID string, quantity int) error {
			return fmt.Errorf("%w: for product %s", model.ErrStockInsufficient, productID)
		},
	}
	invRepo := &mockInventoryRepo{}

	svc := &ProductService{
		DB:            db,
		Repo:          productRepo,
		InventoryRepo: invRepo,
	}

	err := svc.AdjustStock(context.Background(), db, "p1", -15, "venta")
	if !errors.Is(err, model.ErrStockInsufficient) {
		t.Errorf("expected ErrStockInsufficient, got: %v", err)
	}
}
