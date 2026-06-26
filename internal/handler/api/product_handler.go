package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/repository"
	"github.com/solidbit/integritypos/internal/service"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(mux *http.ServeMux, svc *service.ProductService) {
	h := &ProductHandler{svc: svc}
	mux.HandleFunc("GET /api/v1/products", h.listProducts)
	mux.HandleFunc("POST /api/v1/products", h.createProduct)
	mux.HandleFunc("GET /api/v1/products/{id}", h.getProduct)
	mux.HandleFunc("PUT /api/v1/products/{id}", h.updateProduct)
	mux.HandleFunc("DELETE /api/v1/products/{id}", h.deleteProduct)
	mux.HandleFunc("POST /api/v1/products/{id}/stock", h.adjustStock)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (h *ProductHandler) listProducts(w http.ResponseWriter, r *http.Request) {
	filter := repository.ProductFilter{
		Category:      r.URL.Query().Get("category"),
		AvailableOnly: r.URL.Query().Get("available") == "true",
		Search:        r.URL.Query().Get("search"),
	}

	limitStr := r.URL.Query().Get("limit")
	filter.Limit = 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > 100 {
				filter.Limit = 100
			} else {
				filter.Limit = l
			}
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o > 0 {
			filter.Offset = o
		}
	}

	products, err := h.svc.List(r.Context(), filter)
	if err != nil {
		log.Printf("Internal error listing products: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if products == nil {
		products = []model.Product{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) createProduct(w http.ResponseWriter, r *http.Request) {
	var p model.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if err := h.svc.Create(r.Context(), &p); err != nil {
		if errors.Is(err, model.ErrInvalidInput) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("Internal error creating product: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler) getProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !IsValidUUID(id) {
		writeJSONError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	p, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Product not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler) updateProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !IsValidUUID(id) {
		writeJSONError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var p model.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	p.ID = id

	if err := h.svc.Update(r.Context(), &p); err != nil {
		if errors.Is(err, model.ErrInvalidInput) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("Internal error updating product: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !IsValidUUID(id) {
		writeJSONError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		log.Printf("Internal error deleting product: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) adjustStock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !IsValidUUID(id) {
		writeJSONError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req struct {
		Delta  int    `json:"delta"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if err := h.svc.AdjustStock(r.Context(), id, req.Delta, req.Reason); err != nil {
		if errors.Is(err, model.ErrStockInsufficient) {
			writeJSONError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		log.Printf("Internal error adjusting stock: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}


