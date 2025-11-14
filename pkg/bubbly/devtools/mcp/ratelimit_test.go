package mcp

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRateLimiter tests rate limiter creation
func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name              string
		requestsPerSecond int
		wantErr           bool
		errMsg            string
	}{
		{
			name:              "valid rate limit",
			requestsPerSecond: 10,
			wantErr:           false,
		},
		{
			name:              "zero rate limit",
			requestsPerSecond: 0,
			wantErr:           true,
			errMsg:            "rate limit must be positive",
		},
		{
			name:              "negative rate limit",
			requestsPerSecond: -5,
			wantErr:           true,
			errMsg:            "rate limit must be positive",
		},
		{
			name:              "high rate limit",
			requestsPerSecond: 1000,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl, err := NewRateLimiter(tt.requestsPerSecond)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, rl)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, rl)
				assert.Equal(t, tt.requestsPerSecond, rl.limit)
				assert.NotNil(t, rl.limiters)
			}
		})
	}
}

// TestRateLimiter_Middleware_EnforcesLimit tests rate limiting per client
func TestRateLimiter_Middleware_EnforcesLimit(t *testing.T) {
	tests := []struct {
		name           string
		rateLimit      int
		requests       int
		expectAllowed  int
		expectRejected int
	}{
		{
			name:           "under limit",
			rateLimit:      10,
			requests:       5,
			expectAllowed:  5,
			expectRejected: 0,
		},
		{
			name:           "at limit",
			rateLimit:      10,
			requests:       10,
			expectAllowed:  10,
			expectRejected: 0,
		},
		{
			name:           "over limit",
			rateLimit:      5,
			requests:       10,
			expectAllowed:  5,
			expectRejected: 5,
		},
		{
			name:           "single request allowed",
			rateLimit:      1,
			requests:       1,
			expectAllowed:  1,
			expectRejected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl, err := NewRateLimiter(tt.rateLimit)
			require.NoError(t, err)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			wrapped := rl.Middleware(handler)

			allowed := 0
			rejected := 0

			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "192.168.1.1:12345" // Same client
				w := httptest.NewRecorder()

				wrapped.ServeHTTP(w, req)

				if w.Code == http.StatusOK {
					allowed++
				} else if w.Code == http.StatusTooManyRequests {
					rejected++
				}
			}

			assert.Equal(t, tt.expectAllowed, allowed, "allowed requests mismatch")
			assert.Equal(t, tt.expectRejected, rejected, "rejected requests mismatch")
		})
	}
}

// TestRateLimiter_Middleware_ResetOverTime tests that rate limit resets
func TestRateLimiter_Middleware_ResetOverTime(t *testing.T) {
	rl, err := NewRateLimiter(2) // 2 requests per second
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware(handler)

	// Make 2 requests (should succeed)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed", i+1)
	}

	// Third request should fail
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code, "third request should be rate limited")

	// Wait for rate limit to reset (1 second + buffer)
	time.Sleep(1100 * time.Millisecond)

	// Request should succeed again
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w = httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "request after reset should succeed")
}

// TestRateLimiter_Middleware_DifferentClientsIndependent tests per-client isolation
func TestRateLimiter_Middleware_DifferentClientsIndependent(t *testing.T) {
	rl, err := NewRateLimiter(5) // 5 requests per second per client
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware(handler)

	clients := []string{
		"192.168.1.1:12345",
		"192.168.1.2:12346",
		"192.168.1.3:12347",
	}

	// Each client makes 5 requests (at limit)
	for _, clientAddr := range clients {
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = clientAddr
			w := httptest.NewRecorder()
			wrapped.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code, "client %s request %d should succeed", clientAddr, i+1)
		}
	}

	// Each client's 6th request should fail
	for _, clientAddr := range clients {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = clientAddr
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code, "client %s 6th request should be rate limited", clientAddr)
	}
}

// TestRateLimiter_Middleware_ThreadSafe tests concurrent access
func TestRateLimiter_Middleware_ThreadSafe(t *testing.T) {
	rl, err := NewRateLimiter(100) // High limit for concurrency test
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware(handler)

	const numGoroutines = 10
	const requestsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch concurrent requests from different clients
	for i := 0; i < numGoroutines; i++ {
		clientAddr := "192.168.1." + string(rune('1'+i)) + ":12345"
		go func(addr string) {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = addr
				w := httptest.NewRecorder()
				wrapped.ServeHTTP(w, req)
				// Should succeed since we're under the limit
				assert.Equal(t, http.StatusOK, w.Code)
			}
		}(clientAddr)
	}

	wg.Wait()

	// Verify no panics or race conditions occurred
	// (race detector will catch issues)
}

// TestRateLimiter_Middleware_NoMemoryLeaks tests cleanup
func TestRateLimiter_Middleware_NoMemoryLeaks(t *testing.T) {
	rl, err := NewRateLimiter(10)
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware(handler)

	// Simulate many different clients
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1." + string(rune('1'+i%10)) + ":12345"
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
	}

	// Check that limiters map doesn't grow unbounded
	rl.mu.RLock()
	limiterCount := len(rl.limiters)
	rl.mu.RUnlock()

	// Should have at most 10 unique clients (due to i%10)
	assert.LessOrEqual(t, limiterCount, 10, "limiter map should not grow unbounded")
}

// TestRateLimiter_GetClientIP tests client IP extraction
func TestRateLimiter_GetClientIP(t *testing.T) {
	tests := []struct {
		name          string
		remoteAddr    string
		xForwardedFor string
		xRealIP       string
		expected      string
	}{
		{
			name:       "direct connection",
			remoteAddr: "192.168.1.1:12345",
			expected:   "192.168.1.1",
		},
		{
			name:          "x-forwarded-for single",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1",
			expected:      "203.0.113.1",
		},
		{
			name:          "x-forwarded-for multiple",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1, 198.51.100.1, 192.0.2.1",
			expected:      "203.0.113.1",
		},
		{
			name:       "x-real-ip",
			remoteAddr: "10.0.0.1:12345",
			xRealIP:    "203.0.113.1",
			expected:   "203.0.113.1",
		},
		{
			name:       "ipv6",
			remoteAddr: "[2001:db8::1]:12345",
			expected:   "2001:db8::1",
		},
		{
			name:       "invalid format fallback",
			remoteAddr: "invalid",
			expected:   "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			clientIP := getClientIP(req)
			assert.Equal(t, tt.expected, clientIP)
		})
	}
}
