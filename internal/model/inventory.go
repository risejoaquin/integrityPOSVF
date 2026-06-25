package model

import "time"

type InventoryMovement struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	Delta     int       `json:"delta"`
	Reason    string    `json:"reason,omitempty"`
	OrderID   string    `json:"order_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
