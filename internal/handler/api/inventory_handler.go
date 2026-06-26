package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/service"
)

type InventoryHandler struct {
	svc *service.InventoryService
}

func NewInventoryHandler(mux *http.ServeMux, svc *service.InventoryService) {
	h := &InventoryHandler{svc: svc}
	mux.HandleFunc("GET /api/v1/inventory/movements", h.listMovements)
}

func (h *InventoryHandler) listMovements(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("product_id")
	if !IsValidUUID(productID) {
		writeJSONError(w, http.StatusBadRequest, "Invalid product_id format")
		return
	}

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

	movements, err := h.svc.GetMovements(r.Context(), productID, limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if movements == nil {
		movements = []model.InventoryMovement{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movements)
}
