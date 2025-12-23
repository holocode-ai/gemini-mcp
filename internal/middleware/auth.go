package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
)

// Context keys for storing request headers
type contextKey string

const (
	// UploadMediaPathKey is the context key for X-Upload-Media-Path header
	UploadMediaPathKey contextKey = "uploadMediaPath"
	// AuthTokenKey is the context key for Authorization token
	AuthTokenKey contextKey = "authToken"
	// ServerURLKey is the context key for server base URL
	ServerURLKey contextKey = "serverURL"
)

// GetUploadMediaPath extracts the upload media path from context
func GetUploadMediaPath(ctx context.Context) string {
	if v := ctx.Value(UploadMediaPathKey); v != nil {
		return v.(string)
	}
	return ""
}

// GetAuthToken extracts the authorization token from context
func GetAuthToken(ctx context.Context) string {
	if v := ctx.Value(AuthTokenKey); v != nil {
		return v.(string)
	}
	return ""
}

// GetServerURL extracts the server URL from context
func GetServerURL(ctx context.Context) string {
	if v := ctx.Value(ServerURLKey); v != nil {
		return v.(string)
	}
	return ""
}

// HeadersMiddleware injects custom headers into the request context
func HeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Inject X-Upload-Media-Path header into context
		if path := r.Header.Get("X-Upload-Media-Path"); path != "" {
			ctx = context.WithValue(ctx, UploadMediaPathKey, path)
		}

		// Inject Authorization token into context (strip "Bearer " prefix)
		if auth := r.Header.Get("Authorization"); auth != "" {
			if strings.HasPrefix(auth, "Bearer ") {
				ctx = context.WithValue(ctx, AuthTokenKey, strings.TrimPrefix(auth, "Bearer "))
			} else {
				ctx = context.WithValue(ctx, AuthTokenKey, auth)
			}
		}

		// Build server URL from request
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		// Check X-Forwarded-Proto header for proxied requests
		if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		}
		serverURL := scheme + "://" + r.Host
		ctx = context.WithValue(ctx, ServerURLKey, serverURL)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

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
