package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter provides per-client rate limiting for incoming HTTP requests.
type RateLimiter struct {
	limiters sync.Map
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a rate limiter with the specified requests per second and burst.
func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		rate:  rate.Limit(requestsPerSecond),
		burst: burst,
	}
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Middleware returns an HTTP middleware that rate limits requests per client IP.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	// Start cleanup goroutine
	go rl.cleanup()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		entry, _ := rl.limiters.LoadOrStore(ip, &limiterEntry{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: time.Now(),
		})

		e := entry.(*limiterEntry)
		e.lastSeen = time.Now()

		if !e.limiter.Allow() {
			slog.Warn("Rate limit exceeded", "remote", ip, "path", r.URL.Path)
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		rl.limiters.Range(func(key, value interface{}) bool {
			entry := value.(*limiterEntry)
			if now.Sub(entry.lastSeen) > 10*time.Minute {
				rl.limiters.Delete(key)
			}
			return true
		})
	}
}
