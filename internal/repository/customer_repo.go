package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/model"
)

type CustomerRepository struct {
	Pool *pgxpool.Pool
}

func (r *CustomerRepository) GetByID(ctx context.Context, id string) (model.Customer, error) {
	var c model.Customer
	var phone, email *string
	err := r.Pool.QueryRow(ctx, `SELECT id, name, phone, email, metadata, created_at FROM customers WHERE id = $1`, id).Scan(
		&c.ID, &c.Name, &phone, &email, &c.Metadata, &c.CreatedAt)
	if err != nil {
		return c, fmt.Errorf("error fetching customer by id: %w", err)
	}
	if phone != nil {
		c.Phone = *phone
	}
	if email != nil {
		c.Email = *email
	}
	return c, nil
}

func (r *CustomerRepository) GetByPhone(ctx context.Context, phone string) (model.Customer, error) {
	var c model.Customer
	var email *string
	err := r.Pool.QueryRow(ctx, `SELECT id, name, phone, email, metadata, created_at FROM customers WHERE phone = $1`, phone).Scan(
		&c.ID, &c.Name, &c.Phone, &email, &c.Metadata, &c.CreatedAt)
	if err != nil {
		return c, fmt.Errorf("error fetching customer by phone: %w", err)
	}
	if email != nil {
		c.Email = *email
	}
	return c, nil
}

func (r *CustomerRepository) Create(ctx context.Context, c *model.Customer) error {
	var phone, email *string
	if c.Phone != "" {
		phone = &c.Phone
	}
	if c.Email != "" {
		email = &c.Email
	}
	err := r.Pool.QueryRow(ctx,
		`INSERT INTO customers (name, phone, email, metadata) VALUES ($1, $2, $3, COALESCE($4, '{}'::jsonb)) RETURNING id, created_at`,
		c.Name, phone, email, c.Metadata).Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creating customer: %w", err)
	}
	return nil
}

func (r *CustomerRepository) Update(ctx context.Context, c *model.Customer) error {
	var phone, email *string
	if c.Phone != "" {
		phone = &c.Phone
	}
	if c.Email != "" {
		email = &c.Email
	}
	cmd, err := r.Pool.Exec(ctx,
		`UPDATE customers SET name=$1, phone=$2, email=$3, metadata=COALESCE($4, '{}'::jsonb) WHERE id=$5`,
		c.Name, phone, email, c.Metadata, c.ID)
	if err != nil {
		return fmt.Errorf("error updating customer: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("customer with id %s not found", c.ID)
	}
	return nil
}
