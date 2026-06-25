package module

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/config"
)

type Module interface {
	Name() string
	RequiredTables() []string
	Init(mux *http.ServeMux, db *pgxpool.Pool, cfg *config.Config, bus *EventBus) error
}

var (
	registry = make(map[string]Module)
	active   = make(map[string]bool)
	mu       sync.RWMutex
)

func Register(m Module) {
	mu.Lock()
	defer mu.Unlock()
	registry[m.Name()] = m
}

func InitAll(ctx context.Context, mux *http.ServeMux, db *pgxpool.Pool, cfg *config.Config, bus *EventBus) error {
	mu.Lock()
	defer mu.Unlock()

	for name, mod := range registry {
		if checkTables(ctx, db, mod.RequiredTables()) {
			if err := mod.Init(mux, db, cfg, bus); err != nil {
				return fmt.Errorf("error initializing module %s: %w", name, err)
			}
			active[name] = true
		}
	}
	return nil
}

func IsActive(name string) bool {
	mu.RLock()
	defer mu.RUnlock()
	return active[name]
}

func checkTables(ctx context.Context, db *pgxpool.Pool, tables []string) bool {
	for _, t := range tables {
		var exists bool
		err := db.QueryRow(ctx, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)", t).Scan(&exists)
		if err != nil || !exists {
			return false
		}
	}
	return true
}
