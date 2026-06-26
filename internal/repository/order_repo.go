package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/solidbit/integritypos/internal/model"
)

type OrderRepository struct {
	DB DBTX
}

func NewOrderRepository(db DBTX) *OrderRepository {
	return &OrderRepository{DB: db}
}

func (r *OrderRepository) Create(ctx context.Context, db DBTX, order *model.Order) error {
	query := `INSERT INTO orders (status, source, customer_name, customer_phone, notes, total_cents, metadata)
			  VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, '{}'::jsonb))
			  RETURNING id, created_at, updated_at`

	err := db.QueryRow(ctx, query, order.Status, order.Source, order.CustomerName, order.CustomerPhone, order.Notes, order.TotalCents, order.Metadata).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("error creating order: %w", err)
	}

	if len(order.Items) > 0 {
		batch := &pgx.Batch{}
		itemQuery := `INSERT INTO order_items (order_id, product_id, product_name, quantity, unit_price_cents, total_cents, customizations)
					  VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, '{}'::jsonb)) RETURNING id`
		for _, item := range order.Items {
			var pID *string
			if item.ProductID != "" {
				pID = &item.ProductID
			}
			batch.Queue(itemQuery, order.ID, pID, item.ProductName, item.Quantity, item.UnitPriceCents, item.TotalCents, item.Customizations)
		}
		
		br := db.SendBatch(ctx, batch)
		defer br.Close()

		for i := range order.Items {
			order.Items[i].OrderID = order.ID
			err := br.QueryRow().Scan(&order.Items[i].ID)
			if err != nil {
				return fmt.Errorf("error creating order item %d: %w", i, err)
			}
		}
	}

	return nil
}

func (r *OrderRepository) GetByID(ctx context.Context, db DBTX, id string) (model.Order, error) {
	var o model.Order
	err := db.QueryRow(ctx, `SELECT id, status, source, customer_name, customer_phone, notes, total_cents, metadata, created_at, updated_at FROM orders WHERE id = $1`, id).Scan(
		&o.ID, &o.Status, &o.Source, &o.CustomerName, &o.CustomerPhone, &o.Notes, &o.TotalCents, &o.Metadata, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return o, fmt.Errorf("OrderRepo.GetByID: error fetching order by id: %w", err)
	}

	items, err := r.GetItems(ctx, db, o.ID)
	if err != nil {
		return o, fmt.Errorf("OrderRepo.GetByID: error fetching order items: %w", err)
	}
	o.Items = items

	return o, nil
}

func (r *OrderRepository) List(ctx context.Context, db DBTX, status string, limit, offset int) ([]model.Order, error) {
	if limit <= 0 {
		limit = 50
	} else if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT id, status, source, customer_name, customer_phone, notes, total_cents, metadata, created_at, updated_at FROM orders WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, status, limit, offset)
	} else {
		query = `SELECT id, status, source, customer_name, customer_phone, notes, total_cents, metadata, created_at, updated_at FROM orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = append(args, limit, offset)
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error listing orders: %w", err)
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
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
	return r.UpdateStatusTx(ctx, r.DB, id, status)
}

func (r *OrderRepository) UpdateStatusTx(ctx context.Context, db DBTX, id string, status string) error {
	cmd, err := db.Exec(ctx, `UPDATE orders SET status=$1, updated_at=now() WHERE id=$2`, status, id)
	if err != nil {
		return fmt.Errorf("error updating order status: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("order with id %s not found", id)
	}
	return nil
}

func (r *OrderRepository) GetItems(ctx context.Context, db DBTX, orderID string) ([]model.OrderItem, error) {
	rows, err := db.Query(ctx, `SELECT id, order_id, product_id, product_name, quantity, unit_price_cents, total_cents, customizations FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, fmt.Errorf("error listing order items: %w", err)
	}
	defer rows.Close()

	items := make([]model.OrderItem, 0)
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
