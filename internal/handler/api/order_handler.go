package api

import (
	"encoding/json"
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
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) listOrders(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Simplification: only list recent. Status filter can be applied similarly to products.
	orders, err := h.svc.ListOrders(r.Context(), limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
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
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing ID")
		return
	}

	order, err := h.svc.GetOrder(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Order not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.Status == "confirmed" {
		if err := h.svc.ConfirmOrder(r.Context(), id); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else if req.Status == "cancelled" {
		if err := h.svc.CancelOrder(r.Context(), id); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		// Just a simple status update for others like preparing/completed
		if err := h.svc.OrderRepo.UpdateStatus(r.Context(), id, req.Status); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
