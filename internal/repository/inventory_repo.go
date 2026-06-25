package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/model"
)

type InventoryRepository struct {
	Pool *pgxpool.Pool
}

func (r *InventoryRepository) RecordMovement(ctx context.Context, tx Tx, productID string, delta int, reason string, orderID string) error {
	query := `INSERT INTO inventory_movements (product_id, delta, reason, order_id) VALUES ($1, $2, $3, $4)`
	
	var err error
	var oid *string
	if orderID != "" {
		oid = &orderID
	}

	if tx != nil {
		_, err = tx.Exec(ctx, query, productID, delta, reason, oid)
	} else {
		_, err = r.Pool.Exec(ctx, query, productID, delta, reason, oid)
	}

	if err != nil {
		return fmt.Errorf("error recording inventory movement: %w", err)
	}
	return nil
}

func (r *InventoryRepository) GetMovements(ctx context.Context, productID string, limit int) ([]model.InventoryMovement, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, product_id, delta, reason, order_id, created_at FROM inventory_movements WHERE product_id = $1 ORDER BY created_at DESC LIMIT $2`, productID, limit)
	if err != nil {
		return nil, fmt.Errorf("error listing inventory movements: %w", err)
	}
	defer rows.Close()

	var movements []model.InventoryMovement
	for rows.Next() {
		var m model.InventoryMovement
		var oid *string
		var reason *string
		if err := rows.Scan(&m.ID, &m.ProductID, &m.Delta, &reason, &oid, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning inventory movement: %w", err)
		}
		if oid != nil {
			m.OrderID = *oid
		}
		if reason != nil {
			m.Reason = *reason
		}
		movements = append(movements, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return movements, nil
}
