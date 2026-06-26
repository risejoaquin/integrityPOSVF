package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(mux *http.ServeMux, svc *service.OrderService) {
	h := &OrderHandler{svc: svc}
	mux.HandleFunc("POST /api/v1/orders", h.createOrder)
	mux.HandleFunc("GET /api/v1/orders", h.listOrders)
	mux.HandleFunc("GET /api/v1/orders/{id}", h.getOrder)
	mux.HandleFunc("PUT /api/v1/orders/{id}/status", h.updateOrderStatus)
}

type createOrderRequest struct {
	service.Cart
	Source        string `json:"source"`
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
}

func (h *OrderHandler) createOrder(w http.ResponseWriter, r *http.Request) {
	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	source := req.Source
	if source == "" {
		source = "pos"
	}

	custInfo := service.CustomerInfo{
		Name:  req.CustomerName,
		Phone: req.CustomerPhone,
	}

	order, err := h.svc.CreateOrder(r.Context(), req.Cart, source, custInfo)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidInput), errors.Is(err, model.ErrCartEmpty):
			writeJSONError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, model.ErrProductNotFound):
			writeJSONError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, model.ErrStockInsufficient):
			writeJSONError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			log.Printf("Internal error creating order: %v", err)
			writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) listOrders(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > 100 {
				limit = 100
			} else {
				limit = l
			}
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o > 0 {
			offset = o
		}
	}

	orders, err := h.svc.ListOrders(r.Context(), status, limit, offset)
	if err != nil {
		log.Printf("Internal error listing orders: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if orders == nil {
		orders = []model.Order{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) getOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !IsValidUUID(id) {
		writeJSONError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	order, err := h.svc.GetOrder(r.Context(), id)
	if err != nil {
		log.Printf("Error fetching order: %v", err)
		writeJSONError(w, http.StatusNotFound, "Order not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !IsValidUUID(id) {
		writeJSONError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	
	if req.Status == "" {
		writeJSONError(w, http.StatusBadRequest, "Status is required")
		return
	}

	err := h.svc.UpdateOrderStatus(r.Context(), id, model.OrderStatus(req.Status))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidTransition):
			writeJSONError(w, http.StatusConflict, err.Error())
		case errors.Is(err, model.ErrStockInsufficient):
			writeJSONError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			log.Printf("Internal error updating order status: %v", err)
			writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

