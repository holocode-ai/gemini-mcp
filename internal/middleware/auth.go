package middleware

import (
	"log"
	"net/http"
	"strings"
)

// AuthMiddleware creates an HTTP middleware that validates Bearer tokens
func AuthMiddleware(validTokens []string, next http.Handler) http.Handler {
	// Build a set for O(1) token lookup
	tokenSet := make(map[string]struct{}, len(validTokens))
	for _, token := range validTokens {
		tokenSet[token] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication if no tokens are configured
		if len(tokenSet) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// Extract Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("Auth failed: missing Authorization header from %s", r.RemoteAddr)
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"jsonrpc":"2.0","error":{"code":-32001,"message":"Authorization header required"}}`, http.StatusUnauthorized)
			return
		}

		// Validate Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Printf("Auth failed: invalid Authorization format from %s", r.RemoteAddr)
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"jsonrpc":"2.0","error":{"code":-32001,"message":"Bearer token required"}}`, http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		token = strings.TrimSpace(token)

		// Validate token
		if _, valid := tokenSet[token]; !valid {
			log.Printf("Auth failed: invalid token from %s", r.RemoteAddr)
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"jsonrpc":"2.0","error":{"code":-32001,"message":"Invalid token"}}`, http.StatusUnauthorized)
			return
		}

		// Token is valid, proceed to handler
		next.ServeHTTP(w, r)
	})
}
