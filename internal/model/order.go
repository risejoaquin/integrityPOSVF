package model

import (
	"encoding/json"
	"time"
)

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusPreparing OrderStatus = "preparing"
	StatusCompleted OrderStatus = "completed"
	StatusCancelled OrderStatus = "cancelled"
)

var AllowedTransitions = map[OrderStatus][]OrderStatus{
	StatusPending:   {StatusConfirmed, StatusCancelled},
	StatusConfirmed: {StatusPreparing, StatusCancelled},
	StatusPreparing: {StatusCompleted, StatusCancelled},
	StatusCompleted: {},
	StatusCancelled: {},
}

func IsValidTransition(from, to OrderStatus) bool {
	if from == to {
		return true // Allow same state transition (idempotent)
	}
	allowed, ok := AllowedTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

type Order struct {
	ID            string          `json:"id"`
	Status        string          `json:"status"` // using string for db compatibility but maps to OrderStatus
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

