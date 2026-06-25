package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WhatsAppMessageRepo struct {
	Pool *pgxpool.Pool
}

func (r *WhatsAppMessageRepo) SaveMessage(ctx context.Context, direction, from, to, body string) error {
	_, err := r.Pool.Exec(ctx, `INSERT INTO whatsapp_messages (direction, from_number, to_number, body) VALUES ($1, $2, $3, $4)`, direction, from, to, body)
	if err != nil {
		return fmt.Errorf("error saving whatsapp message: %w", err)
	}
	return nil
}
