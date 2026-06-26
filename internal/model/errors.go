package model

import "errors"

var (
	ErrInvalidTransition = errors.New("invalid order status transition")
	ErrCartEmpty         = errors.New("cart cannot be empty")
	ErrProductNotFound   = errors.New("product not found")
	ErrStockInsufficient = errors.New("insufficient stock")
	ErrInvalidInput      = errors.New("invalid input data")
	ErrOrderNotFound     = errors.New("order not found")
)
