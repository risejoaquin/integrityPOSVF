package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/repository"
)

// Dummy Tx
type mockTx struct {
	pgx.Tx
	commitErr   error
	rollbackErr error
}

func (m *mockTx) Commit(ctx context.Context) error {
	return m.commitErr
}

func (m *mockTx) Rollback(ctx context.Context) error {
	return m.rollbackErr
}

func (m *mockTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (m *mockTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}
func (m *mockTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return nil
}

type mockDBBeginner struct {
	tx  pgx.Tx
	err error
}

func (m *mockDBBeginner) Begin(ctx context.Context) (pgx.Tx, error) {
	return m.tx, m.err
}

type mockOrderRepo struct {
	createFn         func(ctx context.Context, db repository.DBTX, order *model.Order) error
	getByIDFn        func(ctx context.Context, db repository.DBTX, id string) (model.Order, error)
	listFn           func(ctx context.Context, db repository.DBTX, status string, limit, offset int) ([]model.Order, error)
	updateStatusTxFn func(ctx context.Context, db repository.DBTX, id string, status string) error
}

func (m *mockOrderRepo) Create(ctx context.Context, db repository.DBTX, order *model.Order) error {
	if m.createFn != nil {
		return m.createFn(ctx, db, order)
	}
	return nil
}

func (m *mockOrderRepo) GetByID(ctx context.Context, db repository.DBTX, id string) (model.Order, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, db, id)
	}
	return model.Order{}, nil
}

func (m *mockOrderRepo) List(ctx context.Context, db repository.DBTX, status string, limit, offset int) ([]model.Order, error) {
	if m.listFn != nil {
		return m.listFn(ctx, db, status, limit, offset)
	}
	return []model.Order{}, nil
}

func (m *mockOrderRepo) UpdateStatusTx(ctx context.Context, db repository.DBTX, id string, status string) error {
	if m.updateStatusTxFn != nil {
		return m.updateStatusTxFn(ctx, db, id, status)
	}
	return nil
}

type mockProductRepo struct {
	getByIDsFn             func(ctx context.Context, db repository.DBTX, ids []string) ([]model.Product, error)
	decrementStockAtomicFn func(ctx context.Context, db repository.DBTX, productID string, quantity int) error
	listFn                 func(ctx context.Context, filter repository.ProductFilter) ([]model.Product, error)
	getByIDFn              func(ctx context.Context, id string) (model.Product, error)
	getStockFn             func(ctx context.Context, db repository.DBTX, id string) (int, error)
	createFn               func(ctx context.Context, db repository.DBTX, p *model.Product) error
	updateFn               func(ctx context.Context, p *model.Product) error
	deleteFn               func(ctx context.Context, id string) error
}

func (m *mockProductRepo) GetByIDs(ctx context.Context, db repository.DBTX, ids []string) ([]model.Product, error) {
	if m.getByIDsFn != nil {
		return m.getByIDsFn(ctx, db, ids)
	}
	return []model.Product{}, nil
}

func (m *mockProductRepo) DecrementStockAtomic(ctx context.Context, db repository.DBTX, productID string, quantity int) error {
	if m.decrementStockAtomicFn != nil {
		return m.decrementStockAtomicFn(ctx, db, productID, quantity)
	}
	return nil
}

func (m *mockProductRepo) List(ctx context.Context, filter repository.ProductFilter) ([]model.Product, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return []model.Product{}, nil
}

func (m *mockProductRepo) GetByID(ctx context.Context, id string) (model.Product, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return model.Product{}, nil
}

func (m *mockProductRepo) GetStock(ctx context.Context, db repository.DBTX, id string) (int, error) {
	if m.getStockFn != nil {
		return m.getStockFn(ctx, db, id)
	}
	return 0, nil
}

func (m *mockProductRepo) Create(ctx context.Context, db repository.DBTX, p *model.Product) error {
	if m.createFn != nil {
		return m.createFn(ctx, db, p)
	}
	return nil
}

func (m *mockProductRepo) Update(ctx context.Context, p *model.Product) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, p)
	}
	return nil
}

func (m *mockProductRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type mockInventoryRepo struct {
	recordMovementFn func(ctx context.Context, db repository.DBTX, productID string, delta int, reason string, orderID string) error
}

func (m *mockInventoryRepo) RecordMovement(ctx context.Context, db repository.DBTX, productID string, delta int, reason string, orderID string) error {
	if m.recordMovementFn != nil {
		return m.recordMovementFn(ctx, db, productID, delta, reason, orderID)
	}
	return nil
}

func TestCreateOrder_Success(t *testing.T) {
	db := &mockDBBeginner{tx: &mockTx{}}
	productRepo := &mockProductRepo{
		getByIDsFn: func(ctx context.Context, db repository.DBTX, ids []string) ([]model.Product, error) {
			return []model.Product{
				{ID: "00000000-0000-0000-0000-000000000001", Name: "Burger", PriceCents: 1000},
			}, nil
		},
	}
	orderRepo := &mockOrderRepo{
		createFn: func(ctx context.Context, db repository.DBTX, order *model.Order) error {
			order.ID = "order-123"
			return nil
		},
	}
	invRepo := &mockInventoryRepo{}

	svc := &OrderService{
		DB:            db,
		OrderRepo:     orderRepo,
		ProductRepo:   productRepo,
		InventoryRepo: invRepo,
	}

	cart := Cart{
		Source: "web",
		Items: []CartItem{
			{ProductID: "00000000-0000-0000-0000-000000000001", Quantity: 2},
		},
	}

	order, err := svc.CreateOrder(context.Background(), cart)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if order.Status != string(model.StatusPending) {
		t.Errorf("expected status pending, got: %v", order.Status)
	}
	if order.TotalCents != 2000 {
		t.Errorf("expected total 2000, got: %v", order.TotalCents)
	}
}

func TestCreateOrder_EmptyCart(t *testing.T) {
	svc := &OrderService{}
	_, err := svc.CreateOrder(context.Background(), Cart{})
	if !errors.Is(err, model.ErrCartEmpty) {
		t.Errorf("expected ErrCartEmpty, got: %v", err)
	}
}

func TestCreateOrder_InsufficientStock(t *testing.T) {
	db := &mockDBBeginner{tx: &mockTx{}}
	productRepo := &mockProductRepo{
		getByIDsFn: func(ctx context.Context, db repository.DBTX, ids []string) ([]model.Product, error) {
			return []model.Product{
				{ID: "00000000-0000-0000-0000-000000000001", Name: "Burger", PriceCents: 1000},
			}, nil
		},
		decrementStockAtomicFn: func(ctx context.Context, db repository.DBTX, productID string, quantity int) error {
			return fmt.Errorf("%w: for product %s", model.ErrStockInsufficient, productID)
		},
	}
	orderRepo := &mockOrderRepo{}
	svc := &OrderService{
		DB:            db,
		ProductRepo:   productRepo,
		OrderRepo:     orderRepo,
	}

	cart := Cart{
		Items: []CartItem{
			{ProductID: "00000000-0000-0000-0000-000000000001", Quantity: 100},
		},
	}
	_, err := svc.CreateOrder(context.Background(), cart)
	if !errors.Is(err, model.ErrStockInsufficient) {
		t.Errorf("expected ErrStockInsufficient, got: %v", err)
	}
}

func TestCreateOrder_InvalidInput(t *testing.T) {
	svc := &OrderService{}
	longName := string(make([]byte, 101))
	cart := Cart{
		CustomerName: longName,
		Items: []CartItem{
			{ProductName: "Generic", UnitPriceCents: 100, Quantity: 1},
		},
	}
	_, err := svc.CreateOrder(context.Background(), cart)
	if !errors.Is(err, model.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for long name, got: %v", err)
	}
}

func TestUpdateOrderStatus_InvalidTransition(t *testing.T) {
	db := &mockDBBeginner{tx: &mockTx{}}
	orderRepo := &mockOrderRepo{
		getByIDFn: func(ctx context.Context, db repository.DBTX, id string) (model.Order, error) {
			return model.Order{ID: "order-1", Status: string(model.StatusCompleted)}, nil
		},
	}
	svc := &OrderService{
		DB:        db,
		OrderRepo: orderRepo,
	}

	err := svc.UpdateOrderStatus(context.Background(), "order-1", model.StatusCancelled)
	if !errors.Is(err, model.ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got: %v", err)
	}
}

func TestUpdateOrderStatus_Cancel_RestocksInventory(t *testing.T) {
	db := &mockDBBeginner{tx: &mockTx{}}
	orderRepo := &mockOrderRepo{
		getByIDFn: func(ctx context.Context, db repository.DBTX, id string) (model.Order, error) {
			return model.Order{
				ID:     "order-1",
				Status: string(model.StatusConfirmed),
				Items: []model.OrderItem{
					{ProductID: "00000000-0000-0000-0000-000000000001", Quantity: 2},
				},
			}, nil
		},
	}
	
	restockCalled := false
	productRepo := &mockProductRepo{
		decrementStockAtomicFn: func(ctx context.Context, db repository.DBTX, productID string, quantity int) error {
			if productID == "00000000-0000-0000-0000-000000000001" && quantity == -2 {
				restockCalled = true
			}
			return nil
		},
	}
	invRepo := &mockInventoryRepo{}

	svc := &OrderService{
		DB:            db,
		OrderRepo:     orderRepo,
		ProductRepo:   productRepo,
		InventoryRepo: invRepo,
	}

	err := svc.UpdateOrderStatus(context.Background(), "order-1", model.StatusCancelled)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !restockCalled {
		t.Errorf("expected restock to be called")
	}
}
