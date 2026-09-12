package middleware

import (
	"kabadiconnect/backend/internal/httpx"
	"net/http"
)

// Browser cookie mutations must originate at the configured recycler origin.
// Native clients with no Origin still use the normal authentication checks.
func CheckOrigin(allowed string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
			if origin := r.Header.Get("Origin"); origin != "" && origin != allowed {
				httpx.Error(w, http.StatusForbidden, "ORIGIN_NOT_ALLOWED", "Request origin is not allowed")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
