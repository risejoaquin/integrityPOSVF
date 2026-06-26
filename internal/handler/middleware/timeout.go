package middleware

import (
	"net/http"
	"time"
)

func Timeout(writeTimeout int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if writeTimeout <= 0 {
			return next
		}
		return http.TimeoutHandler(next, time.Duration(writeTimeout)*time.Second, "Request Timeout")
	}
}
