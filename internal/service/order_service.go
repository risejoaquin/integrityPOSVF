package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/module"
	"github.com/solidbit/integritypos/internal/repository"
)

var numericRegex = regexp.MustCompile(`[^0-9]`)
var uuidRegex = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

type OrderRepo interface {
	Create(ctx context.Context, db repository.DBTX, order *model.Order) error
	GetByID(ctx context.Context, db repository.DBTX, id string) (model.Order, error)
	List(ctx context.Context, db repository.DBTX, status string, limit, offset int) ([]model.Order, error)
	UpdateStatusTx(ctx context.Context, db repository.DBTX, id string, status string) error
}

type ProductRepo interface {
	GetByIDs(ctx context.Context, db repository.DBTX, ids []string) ([]model.Product, error)
	DecrementStockAtomic(ctx context.Context, db repository.DBTX, productID string, quantity int) error
	List(ctx context.Context, filter repository.ProductFilter) ([]model.Product, error)
	GetByID(ctx context.Context, id string) (model.Product, error)
	GetStock(ctx context.Context, db repository.DBTX, id string) (int, error)
	Create(ctx context.Context, db repository.DBTX, p *model.Product) error
	Update(ctx context.Context, p *model.Product) error
	Delete(ctx context.Context, id string) error
}

type InventoryRepo interface {
	RecordMovement(ctx context.Context, db repository.DBTX, productID string, delta int, reason string, orderID string) error
}

type DBBeginner interface {
	repository.DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

type OrderService struct {
	DB            DBBeginner
	OrderRepo     OrderRepo
	ProductRepo   ProductRepo
	InventoryRepo InventoryRepo
	Printer       TicketPrinter
	EventBus      *module.EventBus
}

func (s *OrderService) consolidateCart(cart Cart) (Cart, error) {
	consolidated := make(map[string]*CartItem)
	var genericItems []CartItem

	for _, item := range cart.Items {
		if item.Quantity <= 0 {
			return Cart{}, fmt.Errorf("%w: invalid quantity %d", model.ErrInvalidInput, item.Quantity)
		}
		if item.ProductID == "" {
			if item.ProductName == "" || len(item.ProductName) > 100 {
				return Cart{}, fmt.Errorf("%w: invalid generic product name", model.ErrInvalidInput)
			}
			if item.UnitPriceCents < 0 {
				return Cart{}, fmt.Errorf("%w: invalid generic product price", model.ErrInvalidInput)
			}
			genericItems = append(genericItems, item)
			continue
		}
		if !uuidRegex.MatchString(item.ProductID) {
			return Cart{}, fmt.Errorf("%w: invalid product ID format", model.ErrInvalidInput)
		}
		if existing, ok := consolidated[item.ProductID]; ok {
			existing.Quantity += item.Quantity
		} else {
			copyItem := item
			consolidated[item.ProductID] = &copyItem
		}
	}

	var newCart Cart
	for _, item := range consolidated {
		newCart.Items = append(newCart.Items, *item)
	}
	newCart.Items = append(newCart.Items, genericItems...)
	return newCart, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, cart Cart, source string, customerInfo CustomerInfo) (model.Order, error) {
	consolidatedCart, err := s.consolidateCart(cart)
	if err != nil {
		return model.Order{}, err
	}

	if len(consolidatedCart.Items) == 0 {
		return model.Order{}, model.ErrCartEmpty
	}

	if len(customerInfo.Name) > 100 {
		return model.Order{}, fmt.Errorf("%w: customer name too long", model.ErrInvalidInput)
	}

	if customerInfo.Phone != "" {
		digits := numericRegex.ReplaceAllString(customerInfo.Phone, "")
		if len(digits) < 7 || len(digits) > 15 {
			return model.Order{}, fmt.Errorf("%w: customer phone must have between 7 and 15 digits", model.ErrInvalidInput)
		}
	}

	var productIDs []string
	for _, item := range consolidatedCart.Items {
		if item.ProductID != "" {
			productIDs = append(productIDs, item.ProductID)
		}
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	var commitErr error
	defer func() {
		if commitErr != nil {
			if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
				log.Printf("error rolling back transaction: %v", rollbackErr)
			}
		}
	}()

	products, err := s.ProductRepo.GetByIDs(ctx, tx, productIDs)
	if err != nil {
		commitErr = fmt.Errorf("OrderService.CreateOrder: error fetching products: %w", err)
		return model.Order{}, commitErr
	}

	productMap := make(map[string]model.Product)
	for _, p := range products {
		productMap[p.ID] = p
	}

	order := model.Order{
		Status:        string(model.StatusPending),
		Source:        source,
		CustomerName:  customerInfo.Name,
		CustomerPhone: customerInfo.Phone,
	}

	var totalCents int
	var orderItems []model.OrderItem

	for _, item := range consolidatedCart.Items {
		var priceCents int
		var productName string
		
		if item.ProductID != "" {
			product, ok := productMap[item.ProductID]
			if !ok {
				commitErr = fmt.Errorf("%w: %s", model.ErrProductNotFound, item.ProductID)
				return model.Order{}, commitErr
			}
			if !product.IsAvailable {
				commitErr = fmt.Errorf("%w: product %s is not available", model.ErrInvalidInput, product.Name)
				return model.Order{}, commitErr
			}
			if product.PriceCents <= 0 {
				commitErr = fmt.Errorf("%w: product %s has zero price", model.ErrInvalidInput, product.Name)
				return model.Order{}, commitErr
			}
			priceCents = product.PriceCents
			productName = product.Name
		} else {
			priceCents = item.UnitPriceCents
			productName = item.ProductName
		}

		itemTotal := priceCents * item.Quantity
		totalCents += itemTotal
		customizationsJSON, err := json.Marshal(item.Customizations)
		if err != nil {
			log.Printf("Warning: failed to marshal customizations for item %s: %v", productName, err)
			customizationsJSON = []byte("{}")
		}

		orderItems = append(orderItems, model.OrderItem{
			ProductID:      item.ProductID,
			ProductName:    productName,
			Quantity:       item.Quantity,
			UnitPriceCents: priceCents,
			TotalCents:     itemTotal,
			Customizations: customizationsJSON,
		})

		if item.ProductID != "" {
			// Deduct stock atomically
			if err := s.ProductRepo.DecrementStockAtomic(ctx, tx, item.ProductID, item.Quantity); err != nil {
				commitErr = fmt.Errorf("%w for %s: %w", model.ErrStockInsufficient, productName, err)
				return model.Order{}, commitErr
			}
		}
	}

	order.TotalCents = totalCents
	order.Items = orderItems

	if err := s.OrderRepo.Create(ctx, tx, &order); err != nil {
		commitErr = fmt.Errorf("OrderService.CreateOrder: error creating order: %w", err)
		return model.Order{}, commitErr
	}

	for _, item := range order.Items {
		if item.ProductID != "" {
			if err := s.InventoryRepo.RecordMovement(ctx, tx, item.ProductID, -item.Quantity, "venta", order.ID); err != nil {
				commitErr = fmt.Errorf("OrderService.CreateOrder: error recording movement for %s: %w", item.ProductName, err)
				return model.Order{}, commitErr
			}
		}
	}

	commitErr = tx.Commit(ctx)
	if commitErr != nil {
		return model.Order{}, fmt.Errorf("OrderService.CreateOrder: failed to commit transaction: %w", commitErr)
	}

	return order, nil
}

func (s *OrderService) ConfirmOrder(ctx context.Context, orderID string) error {
	return s.UpdateOrderStatus(ctx, orderID, model.StatusConfirmed)
}

func (s *OrderService) CancelOrder(ctx context.Context, id string) error {
	return s.UpdateOrderStatus(ctx, id, model.StatusCancelled)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, newStatus model.OrderStatus) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	var commitErr error
	defer func() {
		if commitErr != nil {
			if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
				log.Printf("error rolling back transaction: %v", rollbackErr)
			}
		}
	}()

	order, err := s.OrderRepo.GetByID(ctx, tx, orderID)
	if err != nil {
		commitErr = fmt.Errorf("OrderService.UpdateOrderStatus: %w", err)
		return commitErr
	}

	if !model.IsValidTransition(model.OrderStatus(order.Status), newStatus) {
		commitErr = fmt.Errorf("%w: from %s to %s", model.ErrInvalidTransition, order.Status, newStatus)
		return commitErr
	}

	if newStatus == model.StatusCancelled {
		// Revert stock
		for _, item := range order.Items {
			if item.ProductID != "" {
				if err := s.ProductRepo.DecrementStockAtomic(ctx, tx, item.ProductID, -item.Quantity); err != nil {
					commitErr = fmt.Errorf("OrderService.UpdateOrderStatus: error reverting stock for %s: %w", item.ProductName, err)
					return commitErr
				}
				if err := s.InventoryRepo.RecordMovement(ctx, tx, item.ProductID, item.Quantity, "cancelación", order.ID); err != nil {
					commitErr = fmt.Errorf("OrderService.UpdateOrderStatus: error recording reversion movement for %s: %w", item.ProductName, err)
					return commitErr
				}
			}
		}
	}

	if err := s.OrderRepo.UpdateStatusTx(ctx, tx, orderID, string(newStatus)); err != nil {
		commitErr = fmt.Errorf("OrderService.UpdateOrderStatus: error updating order status: %w", err)
		return commitErr
	}

	commitErr = tx.Commit(ctx)
	if commitErr != nil {
		return fmt.Errorf("OrderService.UpdateOrderStatus: failed to commit transaction: %w", commitErr)
	}

	if newStatus == model.StatusConfirmed {
		order.Status = string(newStatus)
		if s.EventBus != nil {
			s.EventBus.Publish("order.confirmed", order)
		}
		if s.Printer != nil {
			go func() {
				if err := s.Printer.PrintTicket(order); err != nil {
					log.Printf("error printing ticket for order %s: %v", order.ID, err)
				}
			}()
		}
	} else if newStatus == model.StatusCancelled {
		order.Status = string(newStatus)
		if s.EventBus != nil {
			s.EventBus.Publish("order.cancelled", order)
		}
	}

	return nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (model.Order, error) {
	o, err := s.OrderRepo.GetByID(ctx, s.DB, id)
	if err != nil {
		return o, fmt.Errorf("OrderService.GetOrder: %w", err)
	}
	return o, nil
}

func (s *OrderService) ListOrders(ctx context.Context, status string, limit, offset int) ([]model.Order, error) {
	orders, err := s.OrderRepo.List(ctx, s.DB, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("OrderService.ListOrders: %w", err)
	}
	return orders, nil
}
