package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Table struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	Status   string `json:"status"` // available, occupied, reserved
	Metadata string `json:"metadata"`
}

type TablesRepository struct {
	Pool *pgxpool.Pool
}

func (r *TablesRepository) List(ctx context.Context) ([]Table, error) {
	rows, err := r.Pool.Query(ctx, "SELECT id, name, capacity, status, metadata FROM tables ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []Table
	for rows.Next() {
		var t Table
		var meta *string
		if err := rows.Scan(&t.ID, &t.Name, &t.Capacity, &t.Status, &meta); err != nil {
			return nil, err
		}
		if meta != nil {
			t.Metadata = *meta
		}
		tables = append(tables, t)
	}
	return tables, nil
}

func (r *TablesRepository) Create(ctx context.Context, t *Table) error {
	return r.Pool.QueryRow(ctx, "INSERT INTO tables (name, capacity, status) VALUES ($1, $2, $3) RETURNING id", t.Name, t.Capacity, t.Status).Scan(&t.ID)
}

func (r *TablesRepository) Update(ctx context.Context, t *Table) error {
	_, err := r.Pool.Exec(ctx, "UPDATE tables SET name = $1, capacity = $2, status = $3 WHERE id = $4", t.Name, t.Capacity, t.Status, t.ID)
	return err
}

func (r *TablesRepository) AssignOrder(ctx context.Context, tableID, orderID string) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	
	_, err = tx.Exec(ctx, "UPDATE orders SET table_id = $1 WHERE id = $2", tableID, orderID)
	if err != nil { return err }
	
	_, err = tx.Exec(ctx, "UPDATE tables SET status = 'occupied' WHERE id = $1", tableID)
	if err != nil { return err }
	
	return tx.Commit(ctx)
}

func (r *TablesRepository) FreeTable(ctx context.Context, tableID string) error {
	_, err := r.Pool.Exec(ctx, "UPDATE tables SET status = 'available' WHERE id = $1", tableID)
	return err
}
