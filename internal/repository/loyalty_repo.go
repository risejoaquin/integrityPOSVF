package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LoyaltyAccount struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	Points     int    `json:"points"`
	Tier       string `json:"tier"`
}

type LoyaltyRepository struct {
	Pool *pgxpool.Pool
}

func (r *LoyaltyRepository) GetByCustomer(ctx context.Context, customerID string) (LoyaltyAccount, error) {
	var a LoyaltyAccount
	err := r.Pool.QueryRow(ctx, "SELECT id, customer_id, points, tier FROM loyalty_accounts WHERE customer_id = $1", customerID).Scan(&a.ID, &a.CustomerID, &a.Points, &a.Tier)
	return a, err
}

func (r *LoyaltyRepository) AddPoints(ctx context.Context, customerID string, points int) error {
	cmd, err := r.Pool.Exec(ctx, "UPDATE loyalty_accounts SET points = points + $2, updated_at = now() WHERE customer_id = $1", customerID, points)
	if err != nil { return err }
	if cmd.RowsAffected() == 0 {
		_, err = r.Pool.Exec(ctx, "INSERT INTO loyalty_accounts (customer_id, points) VALUES ($1, $2)", customerID, points)
	}
	return err
}

func (r *LoyaltyRepository) RedeemPoints(ctx context.Context, customerID string, points int) error {
	_, err := r.Pool.Exec(ctx, "UPDATE loyalty_accounts SET points = points - $2, updated_at = now() WHERE customer_id = $1 AND points >= $2", customerID, points)
	return err
}
