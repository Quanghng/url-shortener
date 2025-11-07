package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter applique une fenêtre glissante par clé (ex: IP) en mémoire.
type RateLimiter struct {
	requests int
	window   time.Duration

	mu      sync.Mutex
	clients map[string][]time.Time
}

// NewRateLimiter crée une nouvelle instance configurée.
func NewRateLimiter(requests int, window time.Duration) *RateLimiter {
	if requests <= 0 {
		requests = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{
		requests: requests,
		window:   window,
		clients:  make(map[string][]time.Time),
	}
}

// Middleware retourne un gin.Handler qui applique la limitation.
func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if key == "" {
			key = "unknown"
		}
		if !r.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

func (r *RateLimiter) allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-r.window)

	r.mu.Lock()
	defer r.mu.Unlock()

	times := r.clients[key]
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	if len(filtered) >= r.requests {
		r.clients[key] = filtered
		return false
	}

	filtered = append(filtered, now)
	if len(filtered) == 0 {
		delete(r.clients, key)
	} else {
		r.clients[key] = filtered
	}
	return true
}
