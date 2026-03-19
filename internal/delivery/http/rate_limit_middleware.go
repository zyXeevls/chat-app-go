package http

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type clientRate struct {
	count   int
	resetAt time.Time
}

type IPRateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientRate
	limit   int
	window  time.Duration
}

func NewIPRateLimiter(limit int, window time.Duration) *IPRateLimiter {
	if limit <= 0 {
		limit = 60
	}

	if window <= 0 {
		window = time.Minute
	}

	return &IPRateLimiter{
		clients: make(map[string]*clientRate),
		limit:   limit,
		window:  window,
	}
}

func (l *IPRateLimiter) Allow(ip string) (bool, int, time.Duration) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	state, ok := l.clients[ip]
	if !ok || now.After(state.resetAt) {
		l.clients[ip] = &clientRate{
			count:   1,
			resetAt: now.Add(l.window),
		}
		return true, l.limit - 1, 0
	}

	if state.count >= l.limit {
		retryAfter := time.Until(state.resetAt)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, 0, retryAfter
	}

	state.count++
	return true, l.limit - state.count, 0
}

func (l *IPRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		allowed, remaining, retryAfter := l.Allow(ip)

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(l.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if !allowed {
			retryAfterSeconds := int(retryAfter.Seconds())
			if retryAfterSeconds < 1 {
				retryAfterSeconds = 1
			}

			w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "rate limit exceeded",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	realIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}

	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}

	return "unknown"
}
