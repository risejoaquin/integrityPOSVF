package model

import (
	"encoding/json"
	"time"
)

type Customer struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Phone     string          `json:"phone,omitempty"`
	Email     string          `json:"email,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}
