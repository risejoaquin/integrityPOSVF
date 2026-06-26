package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/solidbit/integritypos/internal/service"
)

type ReportHandler struct {
	reportSvc *service.ReportService
}

func NewReportHandler(mux *http.ServeMux, reportSvc *service.ReportService, auth func(http.Handler) http.Handler) {
	h := &ReportHandler{reportSvc: reportSvc}

	mux.Handle("GET /api/v1/reports/sales-summary", auth(http.HandlerFunc(h.handleSalesSummary)))
	mux.Handle("GET /api/v1/reports/top-products", auth(http.HandlerFunc(h.handleTopProducts)))
	mux.Handle("GET /api/v1/reports/low-stock", auth(http.HandlerFunc(h.handleLowStock)))
	mux.Handle("GET /api/v1/reports/daily-sales", auth(http.HandlerFunc(h.handleDailySales)))
}

func (h *ReportHandler) handleSalesSummary(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		from = time.Now().AddDate(0, -1, 0) // Default to last month
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		to = time.Now().AddDate(0, 0, 1) // Default to tomorrow to include today
	}

	summary, err := h.reportSvc.SalesSummary(r.Context(), from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *ReportHandler) handleTopProducts(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	limitStr := r.URL.Query().Get("limit")

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		from = time.Now().AddDate(0, -1, 0)
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		to = time.Now().AddDate(0, 0, 1)
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	products, err := h.reportSvc.TopProducts(r.Context(), from, to, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *ReportHandler) handleLowStock(w http.ResponseWriter, r *http.Request) {
	thresholdStr := r.URL.Query().Get("threshold")
	threshold, err := strconv.Atoi(thresholdStr)
	if err != nil {
		threshold = 5
	}

	products, err := h.reportSvc.LowStock(r.Context(), threshold)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *ReportHandler) handleDailySales(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 7
	}

	sales, err := h.reportSvc.DailySales(r.Context(), days)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sales)
}
