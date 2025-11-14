package mcp

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter implements per-client rate limiting for HTTP requests.
//
// It uses the token bucket algorithm via golang.org/x/time/rate to enforce
// request rate limits on a per-client basis. Each client (identified by IP)
// gets their own rate limiter instance, ensuring fair resource allocation
// and preventing DoS attacks.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	rl, err := NewRateLimiter(10) // 10 requests per second per client
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	mux := http.NewServeMux()
//	mux.Handle("/api", rl.Middleware(apiHandler))
type RateLimiter struct {
	// limiters maps client IDs to their rate limiters
	limiters map[string]*rate.Limiter

	// limit is the number of requests per second allowed per client
	limit int

	// mu protects the limiters map
	mu sync.RWMutex
}

// NewRateLimiter creates a new rate limiter with the specified requests per second.
//
// The rate limiter uses a token bucket algorithm where each client can make
// up to `requestsPerSecond` requests per second, with a burst capacity of
// 2x the rate limit to allow for bursty traffic patterns.
//
// Parameters:
//   - requestsPerSecond: Maximum requests per second per client (must be > 0)
//
// Returns:
//   - *RateLimiter: Configured rate limiter instance
//   - error: Validation error if requestsPerSecond is invalid
//
// Example:
//
//	rl, err := NewRateLimiter(100) // 100 req/s per client
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewRateLimiter(requestsPerSecond int) (*RateLimiter, error) {
	if requestsPerSecond <= 0 {
		return nil, fmt.Errorf("rate limit must be positive, got %d", requestsPerSecond)
	}

	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		limit:    requestsPerSecond,
	}, nil
}

// Middleware wraps an HTTP handler with rate limiting.
//
// Requests exceeding the rate limit receive a 429 Too Many Requests response.
// The client IP is extracted from the request (supporting X-Forwarded-For and
// X-Real-IP headers for proxy scenarios).
//
// Parameters:
//   - next: The HTTP handler to wrap
//
// Returns:
//   - http.Handler: Wrapped handler with rate limiting
//
// Example:
//
//	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	    w.WriteHeader(http.StatusOK)
//	})
//	limited := rl.Middleware(handler)
//	http.ListenAndServe(":8080", limited)
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract client IP
		clientID := getClientIP(r)

		// Get or create limiter for this client
		limiter := rl.getLimiter(clientID)

		// Check if request is allowed
		if !limiter.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Request allowed - pass to next handler
		next.ServeHTTP(w, r)
	})
}

// getLimiter returns the rate limiter for a client, creating one if needed.
//
// This method is thread-safe and uses double-checked locking to minimize
// lock contention. The limiter is configured with:
//   - Rate: limit requests per second
//   - Burst: 2x limit (allows short bursts)
//
// Parameters:
//   - clientID: Unique identifier for the client (typically IP address)
//
// Returns:
//   - *rate.Limiter: Rate limiter instance for this client
func (rl *RateLimiter) getLimiter(clientID string) *rate.Limiter {
	// Fast path: read lock for existing limiter
	rl.mu.RLock()
	limiter, exists := rl.limiters[clientID]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	// Slow path: write lock to create new limiter
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check: another goroutine might have created it
	limiter, exists = rl.limiters[clientID]
	if exists {
		return limiter
	}

	// Create new limiter with burst capacity equal to rate
	// This enforces strict rate limiting without allowing bursts
	limiter = rate.NewLimiter(rate.Limit(rl.limit), rl.limit)
	rl.limiters[clientID] = limiter

	return limiter
}

// getClientIP extracts the client IP address from an HTTP request.
//
// It checks headers in this order:
//  1. X-Forwarded-For (first IP in comma-separated list)
//  2. X-Real-IP
//  3. RemoteAddr (direct connection)
//
// This supports both direct connections and proxied requests (load balancers,
// reverse proxies, etc.). For X-Forwarded-For, only the first (leftmost) IP
// is used, as that's the original client IP.
//
// Parameters:
//   - r: HTTP request
//
// Returns:
//   - string: Client IP address (without port)
//
// Example IPs:
//   - "192.168.1.1" (IPv4)
//   - "2001:db8::1" (IPv6)
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (proxy/load balancer)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// We want the first (leftmost) IP, which is the original client
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (alternative proxy header)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr (direct connection)
	// RemoteAddr format: "IP:port" or "[IPv6]:port"
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, return RemoteAddr as-is
		// This handles edge cases like invalid formats
		return r.RemoteAddr
	}

	return ip
}
