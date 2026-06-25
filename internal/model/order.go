package model

import (
	"encoding/json"
	"time"
)

type Order struct {
	ID            string          `json:"id"`
	Status        string          `json:"status"`
	Source        string          `json:"source"`
	CustomerName  string          `json:"customer_name,omitempty"`
	CustomerPhone string          `json:"customer_phone,omitempty"`
	Notes         string          `json:"notes,omitempty"`
	TotalCents    int             `json:"total_cents"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
	Items         []OrderItem     `json:"items,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type OrderItem struct {
	ID             string          `json:"id"`
	OrderID        string          `json:"order_id"`
	ProductID      string          `json:"product_id,omitempty"`
	ProductName    string          `json:"product_name"`
	Quantity       int             `json:"quantity"`
	UnitPriceCents int             `json:"unit_price_cents"`
	TotalCents     int             `json:"total_cents"`
	Customizations json.RawMessage `json:"customizations,omitempty"`
}
