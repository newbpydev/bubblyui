package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewAuthHandler tests the creation of AuthHandler instances
func TestNewAuthHandler(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		enabled     bool
		expectError bool
	}{
		{
			name:        "valid auth handler with token",
			token:       "secret-token-123",
			enabled:     true,
			expectError: false,
		},
		{
			name:        "valid auth handler disabled",
			token:       "",
			enabled:     false,
			expectError: false,
		},
		{
			name:        "empty token with auth enabled",
			token:       "",
			enabled:     true,
			expectError: true,
		},
		{
			name:        "whitespace token with auth enabled",
			token:       "   ",
			enabled:     true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewAuthHandler(tt.token, tt.enabled)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, handler)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, handler)
			}
		})
	}
}

// TestAuthHandler_Middleware_ValidToken tests successful authentication
func TestAuthHandler_Middleware_ValidToken(t *testing.T) {
	const validToken = "my-secret-token"

	handler, err := NewAuthHandler(validToken, true)
	require.NoError(t, err)

	// Create a test handler that will be called if auth succeeds
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	// Wrap with auth middleware
	authMiddleware := handler.Middleware(nextHandler)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid bearer token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "valid token with extra spaces",
			authHeader:     "Bearer  " + validToken,
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.authHeader)
			rec := httptest.NewRecorder()

			authMiddleware.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

// TestAuthHandler_Middleware_InvalidToken tests authentication failures
func TestAuthHandler_Middleware_InvalidToken(t *testing.T) {
	const validToken = "my-secret-token"

	handler, err := NewAuthHandler(validToken, true)
	require.NoError(t, err)

	// Create a test handler that should NOT be called
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for invalid auth")
	})

	// Wrap with auth middleware
	authMiddleware := handler.Middleware(nextHandler)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer wrong-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "malformed authorization header - no bearer",
			authHeader:     "my-secret-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "malformed authorization header - wrong scheme",
			authHeader:     "Basic " + validToken,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "empty bearer token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "case sensitive token",
			authHeader:     "Bearer MY-SECRET-TOKEN",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			authMiddleware.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestAuthHandler_Middleware_DisabledAuth tests that disabled auth allows all requests
func TestAuthHandler_Middleware_DisabledAuth(t *testing.T) {
	handler, err := NewAuthHandler("", false)
	require.NoError(t, err)

	// Create a test handler that should ALWAYS be called
	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with auth middleware
	authMiddleware := handler.Middleware(nextHandler)

	tests := []struct {
		name       string
		authHeader string
	}{
		{
			name:       "no auth header",
			authHeader: "",
		},
		{
			name:       "invalid auth header",
			authHeader: "Bearer invalid-token",
		},
		{
			name:       "malformed auth header",
			authHeader: "not-even-bearer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled = false
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			authMiddleware.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.True(t, nextCalled, "next handler should be called when auth is disabled")
		})
	}
}

// TestAuthHandler_Middleware_TokenNotInErrors tests that tokens are not leaked in error messages
func TestAuthHandler_Middleware_TokenNotInErrors(t *testing.T) {
	const secretToken = "super-secret-token-12345"

	handler, err := NewAuthHandler(secretToken, true)
	require.NoError(t, err)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	authMiddleware := handler.Middleware(nextHandler)

	tests := []struct {
		name       string
		authHeader string
	}{
		{
			name:       "invalid token",
			authHeader: "Bearer wrong-token",
		},
		{
			name:       "valid token format but wrong value",
			authHeader: "Bearer " + secretToken + "-wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.authHeader)
			rec := httptest.NewRecorder()

			authMiddleware.ServeHTTP(rec, req)

			// Check that response doesn't contain the secret token
			responseBody := rec.Body.String()
			assert.NotContains(t, responseBody, secretToken,
				"secret token should not appear in error response")

			// Check that response doesn't contain the provided token either
			providedToken := strings.TrimPrefix(tt.authHeader, "Bearer ")
			assert.NotContains(t, responseBody, providedToken,
				"provided token should not appear in error response")
		})
	}
}

// TestAuthHandler_Middleware_TimingAttackResistance tests timing attack resistance
func TestAuthHandler_Middleware_TimingAttackResistance(t *testing.T) {
	// This test verifies that we use constant-time comparison
	// We can't directly test timing, but we can verify the behavior is consistent

	const validToken = "my-secret-token-123456"

	handler, err := NewAuthHandler(validToken, true)
	require.NoError(t, err)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authMiddleware := handler.Middleware(nextHandler)

	// Test with tokens of same length but different values
	// If using constant-time comparison, all should fail with same behavior
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "wrong token - first char different",
			token: "xy-secret-token-123456",
		},
		{
			name:  "wrong token - last char different",
			token: "my-secret-token-123457",
		},
		{
			name:  "wrong token - middle different",
			token: "my-WRONG-token-123456",
		},
		{
			name:  "completely different token - same length",
			token: "xxxxxxxxxxxxxxxxxxxxxx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			rec := httptest.NewRecorder()

			authMiddleware.ServeHTTP(rec, req)

			// All should return 401 with same behavior
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	}
}

// TestAuthHandler_Middleware_ConcurrentAccess tests thread safety
func TestAuthHandler_Middleware_ConcurrentAccess(t *testing.T) {
	const validToken = "concurrent-test-token"

	handler, err := NewAuthHandler(validToken, true)
	require.NoError(t, err)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authMiddleware := handler.Middleware(nextHandler)

	// Run 100 concurrent requests
	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer func() { done <- true }()

			// Alternate between valid and invalid tokens
			token := validToken
			expectedStatus := http.StatusOK
			if idx%2 == 0 {
				token = "invalid-token"
				expectedStatus = http.StatusUnauthorized
			}

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			authMiddleware.ServeHTTP(rec, req)

			assert.Equal(t, expectedStatus, rec.Code)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
