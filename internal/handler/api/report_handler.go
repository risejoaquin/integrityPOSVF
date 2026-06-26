package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/solidbit/integritypos/internal/service"
)

type ReportHandler struct {
	svc *service.ReportService
}

func NewReportHandler(mux *http.ServeMux, svc *service.ReportService, authMiddleware func(http.Handler) http.Handler) *ReportHandler {
	h := &ReportHandler{svc: svc}
	mux.Handle("GET /api/v1/reports/sales-summary", authMiddleware(http.HandlerFunc(h.salesSummary)))
	mux.Handle("GET /api/v1/reports/top-products", authMiddleware(http.HandlerFunc(h.topProducts)))
	mux.Handle("GET /api/v1/reports/low-stock", authMiddleware(http.HandlerFunc(h.lowStock)))
	mux.Handle("GET /api/v1/reports/daily-sales", authMiddleware(http.HandlerFunc(h.dailySales)))
	return h
}

func parseDateOrDefault(dateStr string, def time.Time) time.Time {
	if dateStr == "" {
		return def
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return def
	}
	return t
}

func (h *ReportHandler) salesSummary(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	from := parseDateOrDefault(r.URL.Query().Get("from"), now.AddDate(0, -1, 0)) // default 1 month ago
	to := parseDateOrDefault(r.URL.Query().Get("to"), now).Add(24 * time.Hour) // include the whole day

	summary, err := h.svc.SalesSummary(r.Context(), from, to)
	if err != nil {
		log.Printf("Internal error fetching sales summary: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *ReportHandler) topProducts(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	from := parseDateOrDefault(r.URL.Query().Get("from"), now.AddDate(0, -1, 0))
	to := parseDateOrDefault(r.URL.Query().Get("to"), now).Add(24 * time.Hour)

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	products, err := h.svc.TopProducts(r.Context(), from, to, limit)
	if err != nil {
		log.Printf("Internal error fetching top products: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *ReportHandler) lowStock(w http.ResponseWriter, r *http.Request) {
	thresholdStr := r.URL.Query().Get("threshold")
	threshold := 5
	if thresholdStr != "" {
		if t, err := strconv.Atoi(thresholdStr); err == nil && t > 0 {
			threshold = t
		}
	}

	products, err := h.svc.LowStock(r.Context(), threshold)
	if err != nil {
		log.Printf("Internal error fetching low stock: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *ReportHandler) dailySales(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 7
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	sales, err := h.svc.DailySales(r.Context(), days)
	if err != nil {
		log.Printf("Internal error fetching daily sales: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sales)
}
