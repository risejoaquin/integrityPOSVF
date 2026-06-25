package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/model"
)

type ProductRepository struct {
	Pool *pgxpool.Pool
}

type ProductFilter struct {
	Category      string
	AvailableOnly bool
	Search        string
}

func (r *ProductRepository) BeginTx(ctx context.Context) (Tx, error) {
	return r.Pool.Begin(ctx)
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (model.Product, error) {
	var p model.Product
	err := r.Pool.QueryRow(ctx,
		`SELECT id, name, price_cents, category, stock, is_available, attributes, created_at, updated_at
		 FROM products WHERE id = $1`, id).Scan(
		&p.ID, &p.Name, &p.PriceCents, &p.Category, &p.Stock, &p.IsAvailable, &p.Attributes, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return p, fmt.Errorf("error fetching product by id: %w", err)
	}
	return p, nil
}

func (r *ProductRepository) List(ctx context.Context, filter ProductFilter) ([]model.Product, error) {
	query := `SELECT id, name, price_cents, category, stock, is_available, attributes, created_at, updated_at FROM products WHERE 1=1`
	var args []interface{}
	argID := 1

	if filter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argID)
		args = append(args, filter.Category)
		argID++
	}
	if filter.AvailableOnly {
		query += fmt.Sprintf(" AND is_available = $%d", argID)
		args = append(args, true)
		argID++
	}
	if filter.Search != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argID)
		args = append(args, "%"+filter.Search+"%")
		argID++
	}

	query += " ORDER BY name ASC"

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error listing products: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.PriceCents, &p.Category, &p.Stock, &p.IsAvailable, &p.Attributes, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning product: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}

func (r *ProductRepository) Create(ctx context.Context, p *model.Product) error {
	err := r.Pool.QueryRow(ctx,
		`INSERT INTO products (name, price_cents, category, stock, is_available, attributes)
		 VALUES ($1, $2, $3, $4, $5, COALESCE($6, '{}'::jsonb))
		 RETURNING id, created_at, updated_at`,
		p.Name, p.PriceCents, p.Category, p.Stock, p.IsAvailable, p.Attributes).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("error creating product: %w", err)
	}
	return nil
}

func (r *ProductRepository) Update(ctx context.Context, p *model.Product) error {
	cmd, err := r.Pool.Exec(ctx,
		`UPDATE products SET name=$1, price_cents=$2, category=$3, is_available=$4, attributes=COALESCE($5, '{}'::jsonb), updated_at=now()
		 WHERE id=$6`,
		p.Name, p.PriceCents, p.Category, p.IsAvailable, p.Attributes, p.ID)
	if err != nil {
		return fmt.Errorf("error updating product: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("product with id %s not found", p.ID)
	}
	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	cmd, err := r.Pool.Exec(ctx, `DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("product with id %s not found", id)
	}
	return nil
}

func (r *ProductRepository) GetStock(ctx context.Context, id string) (int, error) {
	var stock int
	err := r.Pool.QueryRow(ctx, `SELECT stock FROM products WHERE id=$1`, id).Scan(&stock)
	if err != nil {
		return 0, fmt.Errorf("error getting stock: %w", err)
	}
	return stock, nil
}

func (r *ProductRepository) UpdateStock(ctx context.Context, tx Tx, productID string, delta int) error {
	query := `UPDATE products SET stock = stock + $1, updated_at=now() WHERE id = $2`
	
	var err error
	var cmd pgx.CommandTag
	if tx != nil {
		cmd, err = tx.Exec(ctx, query, delta, productID)
	} else {
		cmd, err = r.Pool.Exec(ctx, query, delta, productID)
	}
	
	if err != nil {
		return fmt.Errorf("error updating stock: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("product with id %s not found", productID)
	}
	return nil
}
