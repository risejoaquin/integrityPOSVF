package repository

import (
	"github.com/jackc/pgx/v5"
)

// Tx defines the interface for database transactions.
type Tx interface {
	pgx.Tx
}
