package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"kabadiconnect/backend/internal/httpx"
)

type visitor struct {
	count       int
	windowStart time.Time
}

type FixedWindowLimiter struct {
	mu       sync.Mutex
	visitors map[string]visitor
	limit    int
	window   time.Duration
}

func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{visitors: make(map[string]visitor), limit: limit, window: window}
}

func (l *FixedWindowLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r)
		now := time.Now()
		l.mu.Lock()
		for ip, old := range l.visitors {
			if now.Sub(old.windowStart) >= l.window {
				delete(l.visitors, ip)
			}
		}
		v := l.visitors[key]
		if v.windowStart.IsZero() || now.Sub(v.windowStart) >= l.window {
			v = visitor{windowStart: now}
		}
		if v.count >= l.limit {
			l.mu.Unlock()
			w.Header().Set("Retry-After", "60")
			httpx.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many authentication attempts. Try again shortly.")
			return
		}
		v.count++
		l.visitors[key] = v
		l.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
