package loyalty

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/config"
	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/module"
	"github.com/solidbit/integritypos/internal/repository"
)

type Module struct {}

func init() {
	module.Register(&Module{})
}

func (m *Module) Name() string {
	return "loyalty"
}

func (m *Module) RequiredTables() []string {
	return []string{"loyalty_accounts"}
}

func (m *Module) Init(mux *http.ServeMux, db *pgxpool.Pool, cfg *config.Config, bus *module.EventBus) error {
	repo := &repository.LoyaltyRepository{Pool: db}
	custRepo := &repository.CustomerRepository{Pool: db}
	
	mux.HandleFunc("GET /api/v1/loyalty/points/{customer_id}", func(w http.ResponseWriter, r *http.Request) {
		cid := r.PathValue("customer_id")
		acc, err := repo.GetByCustomer(r.Context(), cid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(acc)
	})

	mux.HandleFunc("POST /api/v1/loyalty/earn", func(w http.ResponseWriter, r *http.Request) {
		var req struct { CustomerID string `json:"customer_id"`; Points int `json:"points"` }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if err := repo.AddPoints(r.Context(), req.CustomerID, req.Points); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /api/v1/loyalty/redeem", func(w http.ResponseWriter, r *http.Request) {
		var req struct { CustomerID string `json:"customer_id"`; Points int `json:"points"` }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if err := repo.RedeemPoints(r.Context(), req.CustomerID, req.Points); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Subscribe to order confirmed
	bus.Subscribe("order.confirmed", func(data interface{}) {
		order, ok := data.(model.Order)
		if ok && order.CustomerPhone != "" {
			cust, err := custRepo.GetByPhone(context.Background(), order.CustomerPhone)
			if err == nil {
				points := order.TotalCents / 100 // 1 point per dollar
				if err := repo.AddPoints(context.Background(), cust.ID, points); err == nil {
					log.Printf("Loyalty: Added %d points to customer %s", points, cust.ID)
				}
			}
		}
	})
	
	return nil
}
