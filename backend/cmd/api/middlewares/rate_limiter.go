package middlewares

import (
	"fmt"
	"net/http"
	"strings"
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
			// Get real client IP (handles proxies, load balancers)
			ip := getClientIP(r)

			// Get rate limiter for this client
			clientLimiter := limiter.GetLimiter(ip)

			// Check if this request is allowed
			if !clientLimiter.Allow() {
				// Calculate remaining tokens and retry after time
				remaining := int(clientLimiter.Tokens())
				if remaining < 0 {
					remaining = 0
				}

				// Calculate retry after based on rate limit
				retryAfter := max(int(1.0 / float64(limiter.rps)), 1)

				// Add headers to inform client about rate limiting using actual values
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.rps))
				w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				http.Error(w, "Rate limit exceeded. Try again later.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the real client IP from request headers
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (most common)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Check CF-Connecting-IP (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		return strings.TrimSpace(cfip)
	}

	// Fallback to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}
