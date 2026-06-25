package tables

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/config"
	"github.com/solidbit/integritypos/internal/module"
	"github.com/solidbit/integritypos/internal/repository"
)

type Module struct {}

func init() {
	module.Register(&Module{})
}

func (m *Module) Name() string {
	return "tables"
}

func (m *Module) RequiredTables() []string {
	return []string{"tables"}
}

func (m *Module) Init(mux *http.ServeMux, db *pgxpool.Pool, cfg *config.Config, bus *module.EventBus) error {
	repo := &repository.TablesRepository{Pool: db}
	
	mux.HandleFunc("GET /api/v1/tables", func(w http.ResponseWriter, r *http.Request) {
		tables, err := repo.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tables)
	})

	mux.HandleFunc("POST /api/v1/tables", func(w http.ResponseWriter, r *http.Request) {
		var t repository.Table
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if t.Status == "" { t.Status = "available" }
		if err := repo.Create(r.Context(), &t); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)
	})

	mux.HandleFunc("POST /api/v1/tables/{id}/assign", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var payload struct { OrderID string `json:"order_id"` }
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if err := repo.AssignOrder(r.Context(), id, payload.OrderID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /api/v1/tables/{id}/free", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := repo.FreeTable(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	
	return nil
}
