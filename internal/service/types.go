package service

import "github.com/solidbit/integritypos/internal/model"

type Cart struct {
	Items []CartItem `json:"items"`
}

type CartItem struct {
	ProductID      string                 `json:"product_id"`
	Quantity       int                    `json:"quantity"`
	Customizations map[string]interface{} `json:"customizations,omitempty"`
}

type CustomerInfo struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type TicketPrinter interface {
	PrintTicket(order model.Order) error
}
