package api

import (
	"encoding/json"
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

	products, err := h.svc.List(r.Context(), filter)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
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
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler) getProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing ID")
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
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing ID")
		return
	}

	var p model.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	p.ID = id

	if err := h.svc.Update(r.Context(), &p); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing ID")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) adjustStock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing ID")
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
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
