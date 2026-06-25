package api

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/handler/middleware"
)

type AdminDashboardHandler struct {
	db *pgxpool.Pool
}

func NewAdminDashboardHandler(mux *http.ServeMux, db *pgxpool.Pool, auth middleware.AuthFunc) {
	h := &AdminDashboardHandler{db: db}
	mux.Handle("GET /api/v1/admin/dashboard", auth(http.HandlerFunc(h.handleDashboard)))
}

func (h *AdminDashboardHandler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var pendingOrdersCount int
	err := h.db.QueryRow(ctx, "SELECT COUNT(*) FROM orders WHERE status = 'pending'").Scan(&pendingOrdersCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var lowStockCount int
	err = h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT product_id FROM inventory_transactions 
			GROUP BY product_id 
			HAVING SUM(quantity) <= 5
		) as ls
	`).Scan(&lowStockCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var todaySales int
	err = h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_cents), 0) 
		FROM orders 
		WHERE status = 'confirmed' AND DATE(created_at) = CURRENT_DATE
	`).Scan(&todaySales)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"pending_orders_count": pendingOrdersCount,
		"low_stock_count":      lowStockCount,
		"today_sales":          todaySales,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
