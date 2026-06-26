package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/solidbit/integritypos/internal/config"
)

type visitor struct {
	count    int
	lastSeen time.Time
}

func RateLimit(cfg config.RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if !cfg.Enabled || cfg.RequestsPerMinute <= 0 {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		}

		visitors := make(map[string]*visitor)
		var mu sync.Mutex

		cleanupInterval := time.Duration(cfg.CleanupIntervalMin) * time.Minute
		if cleanupInterval == 0 {
			cleanupInterval = time.Minute
		}

		go func() {
			for {
				time.Sleep(cleanupInterval)
				mu.Lock()
				now := time.Now()
				for ip, v := range visitors {
					if now.Sub(v.lastSeen) > time.Minute {
						delete(visitors, ip)
					}
				}
				mu.Unlock()
			}
		}()

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)

			mu.Lock()
			v, exists := visitors[ip]
			if !exists {
				visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			if time.Since(v.lastSeen) > time.Minute {
				v.count = 1
				v.lastSeen = time.Now()
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			v.count++
			v.lastSeen = time.Now()
			
			if v.count > cfg.RequestsPerMinute {
				mu.Unlock()
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
