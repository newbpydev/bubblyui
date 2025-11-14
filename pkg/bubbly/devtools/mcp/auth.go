package mcp

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
)

// AuthHandler provides bearer token authentication for HTTP transport.
//
// The handler implements HTTP middleware that validates bearer tokens in the
// Authorization header. It uses constant-time comparison to prevent timing
// attacks and never logs or exposes tokens in error messages.
//
// Security Features:
//   - Constant-time token comparison (timing attack resistant)
//   - Token sanitization in error messages
//   - Configurable enable/disable
//   - Thread-safe operation
//
// Example:
//
//	auth, err := NewAuthHandler("secret-token-123", true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	mux := http.NewServeMux()
//	mux.Handle("/api", auth.Middleware(apiHandler))
//
//	http.ListenAndServe(":8080", mux)
type AuthHandler struct {
	token   string
	enabled bool
}

// NewAuthHandler creates a new authentication handler.
//
// The handler validates bearer tokens in the Authorization header using
// constant-time comparison to prevent timing attacks. When enabled is false,
// all requests are allowed through without authentication.
//
// Parameters:
//   - token: The bearer token to validate against (required if enabled is true)
//   - enabled: Whether authentication is enabled
//
// Returns:
//   - *AuthHandler: The configured authentication handler
//   - error: Validation error if token is empty when enabled is true
//
// Example:
//
//	// Enable authentication
//	auth, err := NewAuthHandler("my-secret-token", true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Disable authentication (for development)
//	auth, err := NewAuthHandler("", false)
func NewAuthHandler(token string, enabled bool) (*AuthHandler, error) {
	// Validate that token is provided when auth is enabled
	if enabled && strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("token cannot be empty when authentication is enabled")
	}

	return &AuthHandler{
		token:   token,
		enabled: enabled,
	}, nil
}

// Middleware wraps an HTTP handler with bearer token authentication.
//
// The middleware validates the Authorization header and only calls the next
// handler if authentication succeeds. When authentication is disabled, all
// requests pass through without validation.
//
// Authentication Flow:
//  1. If auth is disabled, call next handler immediately
//  2. Extract Authorization header
//  3. Validate "Bearer <token>" format
//  4. Compare token using constant-time comparison
//  5. Return 401 Unauthorized on failure
//  6. Call next handler on success
//
// Security:
//   - Uses subtle.ConstantTimeCompare to prevent timing attacks
//   - Never includes tokens in error messages
//   - Returns generic error messages to prevent information leakage
//
// Parameters:
//   - next: The HTTP handler to call if authentication succeeds
//
// Returns:
//   - http.Handler: The wrapped handler with authentication
//
// Example:
//
//	auth, _ := NewAuthHandler("secret", true)
//	protectedHandler := auth.Middleware(myHandler)
//	http.Handle("/api", protectedHandler)
func (a *AuthHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If auth is disabled, allow all requests
		if !a.enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Extract Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		// Parse "Bearer <token>" format
		parts := strings.Fields(authHeader) // Fields handles multiple spaces
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization format", http.StatusUnauthorized)
			return
		}

		providedToken := parts[1]

		// Validate token is not empty
		if providedToken == "" {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Use constant-time comparison to prevent timing attacks
		// This ensures that token comparison takes the same time regardless
		// of where the mismatch occurs, preventing attackers from using
		// timing information to guess the token character by character
		if !constantTimeCompare(providedToken, a.token) {
			// Generic error message - don't leak token information
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Authentication successful - call next handler
		next.ServeHTTP(w, r)
	})
}

// constantTimeCompare performs constant-time string comparison.
//
// This function uses subtle.ConstantTimeCompare to prevent timing attacks.
// It ensures that the comparison takes the same amount of time regardless
// of where the strings differ, making it impossible for attackers to use
// timing information to guess the token.
//
// The function first checks if the lengths match (this is safe to do in
// non-constant time as length is not secret information), then performs
// constant-time byte-by-byte comparison.
//
// Parameters:
//   - a: First string to compare
//   - b: Second string to compare
//
// Returns:
//   - bool: true if strings are equal, false otherwise
func constantTimeCompare(a, b string) bool {
	// Convert strings to byte slices for comparison
	aBytes := []byte(a)
	bBytes := []byte(b)

	// subtle.ConstantTimeCompare returns 1 if equal, 0 if not
	// It also returns 0 if lengths don't match
	return subtle.ConstantTimeCompare(aBytes, bBytes) == 1
}
