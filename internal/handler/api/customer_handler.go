package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/service"
)

type CustomerHandler struct {
	svc *service.CustomerService
}

func NewCustomerHandler(mux *http.ServeMux, svc *service.CustomerService) {
	h := &CustomerHandler{svc: svc}
	mux.HandleFunc("GET /api/v1/customers", h.listCustomers)
	mux.HandleFunc("POST /api/v1/customers", h.createCustomer)
	mux.HandleFunc("GET /api/v1/customers/{id}", h.getCustomer)
	mux.HandleFunc("PUT /api/v1/customers/{id}", h.updateCustomer)
}

func (h *CustomerHandler) listCustomers(w http.ResponseWriter, r *http.Request) {
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

	customers, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}

func (h *CustomerHandler) createCustomer(w http.ResponseWriter, r *http.Request) {
	var c model.Customer
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if err := h.svc.Create(r.Context(), &c); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func (h *CustomerHandler) getCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !IsValidUUID(id) {
		writeJSONError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	c, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Customer not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func (h *CustomerHandler) updateCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !IsValidUUID(id) {
		writeJSONError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var c model.Customer
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	c.ID = id

	if err := h.svc.Update(r.Context(), &c); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}
