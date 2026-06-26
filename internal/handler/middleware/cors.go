package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/solidbit/integritypos/internal/config"
)

func CORS(cfg config.CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if !cfg.Enabled {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		}

		allowedOriginsMap := make(map[string]bool)
		allowAll := false
		for _, o := range cfg.AllowedOrigins {
			if o == "*" {
				allowAll = true
				break
			}
			allowedOriginsMap[o] = true
		}

		allowedMethods := strings.Join(cfg.AllowedMethods, ", ")
		allowedHeaders := strings.Join(cfg.AllowedHeaders, ", ")
		maxAge := strconv.Itoa(cfg.MaxAge)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" {
				if allowAll {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else if allowedOriginsMap[origin] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}

				if allowedMethods != "" {
					w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
				}
				if allowedHeaders != "" {
					w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
				}
				if cfg.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				if cfg.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", maxAge)
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
