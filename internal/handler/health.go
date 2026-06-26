package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/integration/printer"
)

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Printer  string `json:"printer,omitempty"`
}

type HealthHandler struct {
	DB      *pgxpool.Pool
	Printer printer.TicketPrinter
}

func NewHealthHandler(db *pgxpool.Pool, p printer.TicketPrinter) *HealthHandler {
	return &HealthHandler{
		DB:      db,
		Printer: p,
	}
}

func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:   "ok",
		Database: "ok",
		Printer:  "ok",
	}
	statusCode := http.StatusOK

	if err := h.DB.Ping(context.Background()); err != nil {
		resp.Database = "degraded"
		resp.Status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	if h.Printer != nil {
		if err := h.Printer.Ping(); err != nil {
			resp.Printer = "degraded"
			if resp.Status == "ok" {
				resp.Status = "degraded"
			}
		}
	} else {
		resp.Printer = "not_configured"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}
