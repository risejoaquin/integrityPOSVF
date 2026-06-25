package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/model"
)

type OrderRepository struct {
	Pool *pgxpool.Pool
}

func (r *OrderRepository) BeginTx(ctx context.Context) (Tx, error) {
	return r.Pool.Begin(ctx)
}

func (r *OrderRepository) Create(ctx context.Context, tx Tx, order *model.Order) error {
	query := `INSERT INTO orders (status, source, customer_name, customer_phone, notes, total_cents, metadata)
			  VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, '{}'::jsonb))
			  RETURNING id, created_at, updated_at`
			  
	var err error
	if tx != nil {
		err = tx.QueryRow(ctx, query, order.Status, order.Source, order.CustomerName, order.CustomerPhone, order.Notes, order.TotalCents, order.Metadata).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	} else {
		err = r.Pool.QueryRow(ctx, query, order.Status, order.Source, order.CustomerName, order.CustomerPhone, order.Notes, order.TotalCents, order.Metadata).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	}

	if err != nil {
		return fmt.Errorf("error creating order: %w", err)
	}

	itemQuery := `INSERT INTO order_items (order_id, product_id, product_name, quantity, unit_price_cents, total_cents, customizations)
				  VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, '{}'::jsonb))
				  RETURNING id`
				  
	for i := range order.Items {
		order.Items[i].OrderID = order.ID
		var itemErr error
		if tx != nil {
			itemErr = tx.QueryRow(ctx, itemQuery, order.ID, order.Items[i].ProductID, order.Items[i].ProductName, order.Items[i].Quantity, order.Items[i].UnitPriceCents, order.Items[i].TotalCents, order.Items[i].Customizations).Scan(&order.Items[i].ID)
		} else {
			itemErr = r.Pool.QueryRow(ctx, itemQuery, order.ID, order.Items[i].ProductID, order.Items[i].ProductName, order.Items[i].Quantity, order.Items[i].UnitPriceCents, order.Items[i].TotalCents, order.Items[i].Customizations).Scan(&order.Items[i].ID)
		}
		if itemErr != nil {
			return fmt.Errorf("error creating order item: %w", itemErr)
		}
	}

	return nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (model.Order, error) {
	var o model.Order
	err := r.Pool.QueryRow(ctx, `SELECT id, status, source, customer_name, customer_phone, notes, total_cents, metadata, created_at, updated_at FROM orders WHERE id = $1`, id).Scan(
		&o.ID, &o.Status, &o.Source, &o.CustomerName, &o.CustomerPhone, &o.Notes, &o.TotalCents, &o.Metadata, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return o, fmt.Errorf("error fetching order by id: %w", err)
	}

	items, err := r.GetItems(ctx, o.ID)
	if err != nil {
		return o, fmt.Errorf("error fetching order items: %w", err)
	}
	o.Items = items

	return o, nil
}

func (r *OrderRepository) ListByStatus(ctx context.Context, status string) ([]model.Order, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, status, source, customer_name, customer_phone, notes, total_cents, metadata, created_at, updated_at FROM orders WHERE status = $1 ORDER BY created_at DESC`, status)
	if err != nil {
		return nil, fmt.Errorf("error listing orders by status: %w", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.Status, &o.Source, &o.CustomerName, &o.CustomerPhone, &o.Notes, &o.TotalCents, &o.Metadata, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return orders, nil
}

func (r *OrderRepository) ListRecent(ctx context.Context, limit int) ([]model.Order, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, status, source, customer_name, customer_phone, notes, total_cents, metadata, created_at, updated_at FROM orders ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("error listing recent orders: %w", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.Status, &o.Source, &o.CustomerName, &o.CustomerPhone, &o.Notes, &o.TotalCents, &o.Metadata, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	cmd, err := r.Pool.Exec(ctx, `UPDATE orders SET status=$1, updated_at=now() WHERE id=$2`, status, id)
	if err != nil {
		return fmt.Errorf("error updating order status: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("order with id %s not found", id)
	}
	return nil
}

func (r *OrderRepository) GetItems(ctx context.Context, orderID string) ([]model.OrderItem, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, order_id, product_id, product_name, quantity, unit_price_cents, total_cents, customizations FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, fmt.Errorf("error listing order items: %w", err)
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		var item model.OrderItem
		var pid *string
		if err := rows.Scan(&item.ID, &item.OrderID, &pid, &item.ProductName, &item.Quantity, &item.UnitPriceCents, &item.TotalCents, &item.Customizations); err != nil {
			return nil, fmt.Errorf("error scanning order item: %w", err)
		}
		if pid != nil {
			item.ProductID = *pid
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return items, nil
}
