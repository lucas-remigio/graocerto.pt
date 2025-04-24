package middlewares

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// limiterWithLastAccess pairs a rate limiter with its last access time
type limiterWithLastAccess struct {
	limiter   *rate.Limiter
	lastUsage time.Time
}

// ClientRateLimiter stores rate limiters for different clients
type ClientRateLimiter struct {
	limiters        map[string]*limiterWithLastAccess
	mu              sync.Mutex
	rps             int
	maxBurst        int
	cleanupInterval time.Duration
}

// NewClientRateLimiter creates a new client rate limiter
func NewClientRateLimiter(rps, maxBurst int, cleanupInterval time.Duration) *ClientRateLimiter {
	limiter := &ClientRateLimiter{
		limiters:        make(map[string]*limiterWithLastAccess),
		rps:             rps,
		maxBurst:        maxBurst,
		cleanupInterval: cleanupInterval,
	}

	// Start cleanup routine
	go limiter.cleanup()

	return limiter
}

// Cleanup removes old limiters
func (c *ClientRateLimiter) cleanup() {
	for {
		time.Sleep(c.cleanupInterval)
		c.mu.Lock()
		for ip, limiterData := range c.limiters {
			if time.Since(limiterData.lastUsage) > c.cleanupInterval {
				delete(c.limiters, ip)
			}
		}
		c.mu.Unlock()
	}
}

// GetLimiter returns the rate limiter for the provided IP
func (c *ClientRateLimiter) GetLimiter(ip string) *rate.Limiter {
	c.mu.Lock()
	defer c.mu.Unlock()

	limiterData, exists := c.limiters[ip]
	if !exists {
		limiterData = &limiterWithLastAccess{
			limiter:   rate.NewLimiter(rate.Limit(c.rps), c.maxBurst),
			lastUsage: time.Now(),
		}
		c.limiters[ip] = limiterData
	} else {
		// Update the last usage time
		limiterData.lastUsage = time.Now()
	}

	return limiterData.limiter
}

// RateLimitMiddleware is a middleware that limits request rates
func RateLimitMiddleware(limiter *ClientRateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			ip := r.RemoteAddr

			// Get rate limiter for this client
			clientLimiter := limiter.GetLimiter(ip)

			// Check if this request is allowed
			if !clientLimiter.Allow() {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
