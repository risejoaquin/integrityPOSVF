package model

import (
	"encoding/json"
	"time"
)

type Product struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	PriceCents  int             `json:"price_cents"`
	Category    string          `json:"category,omitempty"`
	Stock       int             `json:"stock"`
	IsAvailable bool            `json:"is_available"`
	Attributes  json.RawMessage `json:"attributes,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
