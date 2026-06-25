package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/module"
	"github.com/solidbit/integritypos/internal/repository"
)

type OrderRepository interface {
	BeginTx(ctx context.Context) (repository.Tx, error)
	Create(ctx context.Context, tx repository.Tx, order *model.Order) error
	GetByID(ctx context.Context, id string) (model.Order, error)
	UpdateStatus(ctx context.Context, id, status string) error
	ListRecent(ctx context.Context, limit int) ([]model.Order, error)
}

type ProductRepository interface {
	GetByID(ctx context.Context, id string) (model.Product, error)
	UpdateStock(ctx context.Context, tx repository.Tx, id string, delta int) error
}

type InventoryRepository interface {
	RecordMovement(ctx context.Context, tx repository.Tx, productID string, delta int, reason string, orderID string) error
}

type OrderService struct {
	OrderRepo     OrderRepository
	ProductRepo   ProductRepository
	InventoryRepo InventoryRepository
	Printer       TicketPrinter
	EventBus      *module.EventBus
}

func (s *OrderService) CreateOrder(ctx context.Context, cart Cart, source string, customerInfo CustomerInfo) (model.Order, error) {
	if len(cart.Items) == 0 {
		return model.Order{}, fmt.Errorf("cart is empty")
	}

	tx, err := s.OrderRepo.BeginTx(ctx)
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	order := model.Order{
		Status:        "pending",
		Source:        source,
		CustomerName:  customerInfo.Name,
		CustomerPhone: customerInfo.Phone,
	}

	var totalCents int
	var orderItems []model.OrderItem

	for _, item := range cart.Items {
		if item.Quantity <= 0 {
			return model.Order{}, fmt.Errorf("invalid quantity %d for product %s", item.Quantity, item.ProductID)
		}

		product, err := s.ProductRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return model.Order{}, fmt.Errorf("error fetching product %s: %w", item.ProductID, err)
		}

		if !product.IsAvailable {
			return model.Order{}, fmt.Errorf("product %s is not available", product.Name)
		}

		if product.Stock-item.Quantity < 0 {
			return model.Order{}, fmt.Errorf("insufficient stock for product %s", product.Name)
		}

		itemTotal := product.PriceCents * item.Quantity
		totalCents += itemTotal

		customizationsJSON, _ := json.Marshal(item.Customizations)

		orderItems = append(orderItems, model.OrderItem{
			ProductID:      product.ID,
			ProductName:    product.Name,
			Quantity:       item.Quantity,
			UnitPriceCents: product.PriceCents,
			TotalCents:     itemTotal,
			Customizations: customizationsJSON,
		})

		// Deduct stock
		if err := s.ProductRepo.UpdateStock(ctx, tx, product.ID, -item.Quantity); err != nil {
			return model.Order{}, fmt.Errorf("error updating stock for %s: %w", product.Name, err)
		}
	}

	order.TotalCents = totalCents
	order.Items = orderItems

	if err := s.OrderRepo.Create(ctx, tx, &order); err != nil {
		return model.Order{}, fmt.Errorf("error creating order: %w", err)
	}

	for _, item := range order.Items {
		if err := s.InventoryRepo.RecordMovement(ctx, tx, item.ProductID, -item.Quantity, "venta", order.ID); err != nil {
			return model.Order{}, fmt.Errorf("error recording movement for %s: %w", item.ProductName, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Order{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return order, nil
}

func (s *OrderService) ConfirmOrder(ctx context.Context, orderID string) error {
	if err := s.OrderRepo.UpdateStatus(ctx, orderID, "confirmed"); err != nil {
		return err
	}

	order, err := s.OrderRepo.GetByID(ctx, orderID)
	if err == nil {
		if s.Printer != nil {
			s.Printer.PrintTicket(order)
		}
		if s.EventBus != nil {
			s.EventBus.Publish("order.confirmed", order)
		}
	}

	return nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (model.Order, error) {
	return s.OrderRepo.GetByID(ctx, id)
}

func (s *OrderService) ListOrders(ctx context.Context, limit int) ([]model.Order, error) {
	return s.OrderRepo.ListRecent(ctx, limit)
}

func (s *OrderService) CancelOrder(ctx context.Context, id string) error {
	tx, err := s.OrderRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	order, err := s.OrderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if order.Status == "cancelled" {
		return fmt.Errorf("order already cancelled")
	}

	cmd, err := tx.Exec(ctx, `UPDATE orders SET status='cancelled', updated_at=now() WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("error updating order status: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("order with id %s not found", id)
	}

	// Revert stock
	for _, item := range order.Items {
		if item.ProductID != "" {
			if err := s.ProductRepo.UpdateStock(ctx, tx, item.ProductID, item.Quantity); err != nil {
				return fmt.Errorf("error reverting stock for %s: %w", item.ProductName, err)
			}
			if err := s.InventoryRepo.RecordMovement(ctx, tx, item.ProductID, item.Quantity, "cancelación", order.ID); err != nil {
				return fmt.Errorf("error recording reversion movement for %s: %w", item.ProductName, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
