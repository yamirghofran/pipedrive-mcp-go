package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth provides JWT authentication middleware for the HTTP transport.
type JWTAuth struct {
	secret    jwt.Keyfunc
	algorithm string
	audience  string
	issuer    string
}

// NewJWTAuth creates a new JWT authentication middleware.
func NewJWTAuth(secret, algorithm, audience, issuer string) (*JWTAuth, error) {
	var keyFunc jwt.Keyfunc

	// Determine if the secret is a PEM-encoded public key or an HMAC secret
	if strings.HasPrefix(strings.TrimSpace(secret), "-----BEGIN") {
		// PEM-encoded key (for RSA algorithms)
		pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(secret))
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
		}
		keyFunc = func(token *jwt.Token) (interface{}, error) {
			return pubKey, nil
		}
	} else {
		// HMAC secret
		keyFunc = func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		}
	}

	return &JWTAuth{
		secret:    keyFunc,
		algorithm: algorithm,
		audience:  audience,
		issuer:    issuer,
	}, nil
}

// Middleware returns an HTTP middleware that validates JWT tokens.
func (j *JWTAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			slog.Warn("Missing Authorization header", "remote", r.RemoteAddr)
			http.Error(w, "Unauthorized: missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Parse "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			slog.Warn("Invalid Authorization header format", "remote", r.RemoteAddr)
			http.Error(w, "Unauthorized: invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Parse and validate the token
		parserOptions := []jwt.ParserOption{
			jwt.WithValidMethods([]string{j.algorithm}),
		}
		if j.audience != "" {
			parserOptions = append(parserOptions, jwt.WithAudience(j.audience))
		}
		if j.issuer != "" {
			parserOptions = append(parserOptions, jwt.WithIssuer(j.issuer))
		}

		token, err := jwt.Parse(tokenString, j.secret, parserOptions...)
		if err != nil {
			slog.Warn("JWT validation failed", "error", err, "remote", r.RemoteAddr)
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			slog.Warn("Invalid JWT token", "remote", r.RemoteAddr)
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			ctx := context.WithValue(r.Context(), "jwt_claims", claims)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}
