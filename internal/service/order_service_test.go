package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/repository"
)

type mockOrderRepo struct {
	orders     map[string]model.Order
	status     string
	createFunc func(ctx context.Context, tx repository.Tx, order *model.Order) error
	txFunc     func(ctx context.Context) (repository.Tx, error)
}

type mockTx struct {
	repository.Tx
}

func (m *mockTx) Rollback(ctx context.Context) error { return nil }
func (m *mockTx) Commit(ctx context.Context) error   { return nil }

func (m *mockOrderRepo) BeginTx(ctx context.Context) (repository.Tx, error) {
	if m.txFunc != nil {
		return m.txFunc(ctx)
	}
	return &mockTx{}, nil
}

func (m *mockOrderRepo) Create(ctx context.Context, tx repository.Tx, order *model.Order) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, tx, order)
	}
	order.ID = "test-order-123"
	m.orders[order.ID] = *order
	return nil
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id string) (model.Order, error) {
	if order, ok := m.orders[id]; ok {
		return order, nil
	}
	return model.Order{}, fmt.Errorf("not found")
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id, status string) error {
	if order, ok := m.orders[id]; ok {
		order.Status = status
		m.orders[id] = order
		m.status = status
		return nil
	}
	return fmt.Errorf("not found")
}

func (m *mockOrderRepo) List(ctx context.Context) ([]model.Order, error) {
	var list []model.Order
	for _, o := range m.orders {
		list = append(list, o)
	}
	return list, nil
}

type mockProductRepo struct {
	products map[string]model.Product
}

func (m *mockProductRepo) GetByID(ctx context.Context, id string) (model.Product, error) {
	if p, ok := m.products[id]; ok {
		return p, nil
	}
	return model.Product{}, fmt.Errorf("not found")
}

func (m *mockProductRepo) UpdateStock(ctx context.Context, tx repository.Tx, id string, delta int) error {
	if p, ok := m.products[id]; ok {
		p.Stock += delta
		m.products[id] = p
		return nil
	}
	return fmt.Errorf("not found")
}

type mockInventoryRepo struct{}

func (m *mockInventoryRepo) AdjustStock(ctx context.Context, tx repository.Tx, req repository.StockAdjustmentReq) error {
	return nil
}

func TestCreateOrder(t *testing.T) {
	orderRepo := &mockOrderRepo{orders: make(map[string]model.Order)}
	productRepo := &mockProductRepo{
		products: map[string]model.Product{
			"p1": {ID: "p1", Name: "Burger", PriceCents: 500, Stock: 10, IsAvailable: true},
			"p2": {ID: "p2", Name: "Fries", PriceCents: 250, Stock: 5, IsAvailable: true},
			"p3": {ID: "p3", Name: "Soda", PriceCents: 150, Stock: 0, IsAvailable: true},
		},
	}
	invRepo := &mockInventoryRepo{}

	svc := &OrderService{
		OrderRepo:     orderRepo,
		ProductRepo:   productRepo,
		InventoryRepo: invRepo,
	}

	t.Run("successful order", func(t *testing.T) {
		cart := Cart{
			Items: []CartItem{
				{ProductID: "p1", Quantity: 2},
				{ProductID: "p2", Quantity: 1},
			},
		}
		
		order, err := svc.CreateOrder(context.Background(), cart, "test", CustomerInfo{Name: "John"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if order.ID == "" {
			t.Errorf("expected order ID to be set")
		}
		if order.TotalCents != 1250 {
			t.Errorf("expected total 1250, got %d", order.TotalCents)
		}
		if productRepo.products["p1"].Stock != 8 {
			t.Errorf("expected p1 stock to be 8, got %d", productRepo.products["p1"].Stock)
		}
	})

	t.Run("empty cart", func(t *testing.T) {
		_, err := svc.CreateOrder(context.Background(), Cart{}, "test", CustomerInfo{})
		if err == nil {
			t.Errorf("expected error for empty cart")
		}
	})

	t.Run("product not found", func(t *testing.T) {
		cart := Cart{Items: []CartItem{{ProductID: "invalid", Quantity: 1}}}
		_, err := svc.CreateOrder(context.Background(), cart, "test", CustomerInfo{})
		if err == nil {
			t.Errorf("expected error for invalid product")
		}
	})

	t.Run("insufficient stock", func(t *testing.T) {
		cart := Cart{Items: []CartItem{{ProductID: "p3", Quantity: 1}}}
		_, err := svc.CreateOrder(context.Background(), cart, "test", CustomerInfo{})
		if err == nil {
			t.Errorf("expected error for insufficient stock")
		}
	})
}

func TestConfirmOrder(t *testing.T) {
	orderRepo := &mockOrderRepo{
		orders: map[string]model.Order{
			"o1": {ID: "o1", Status: "pending"},
		},
	}

	svc := &OrderService{
		OrderRepo: orderRepo,
	}

	err := svc.ConfirmOrder(context.Background(), "o1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if orderRepo.status != "confirmed" {
		t.Errorf("expected status 'confirmed', got %s", orderRepo.status)
	}
}
