package middleware

import (
	"net/http"
	"strings"
)

func Auth(adminToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if adminToken == "" {
				next.ServeHTTP(w, r)
				return
			}
			
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token != adminToken {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}
