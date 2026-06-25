package api

import (
	"encoding/json"
	"net/http"

	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/service"
)

type CustomerHandler struct {
	svc *service.CustomerService
}

func NewCustomerHandler(mux *http.ServeMux, svc *service.CustomerService) {
	h := &CustomerHandler{svc: svc}
	mux.HandleFunc("POST /api/v1/customers", h.createCustomer)
	mux.HandleFunc("GET /api/v1/customers/{id}", h.getCustomer)
	mux.HandleFunc("PUT /api/v1/customers/{id}", h.updateCustomer)
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
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing ID")
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
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing ID")
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
