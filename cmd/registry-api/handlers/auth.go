package handlers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"github.com/nullapt/nullapt/internal/db"
)

type contextKey string

const contextKeyUserID contextKey = "user_id"

// AuthMiddleware validates the Bearer token and injects the user ID into context.
func AuthMiddleware(pool *db.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "missing or invalid Authorization header")
				return
			}
			rawToken := strings.TrimPrefix(header, "Bearer ")

			userID, err := resolveToken(r.Context(), pool, rawToken)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth injects the user ID if a valid token is present, but does not
// reject requests without one. Used on routes that have both public and
// authenticated views.
func OptionalAuth(pool *db.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if rawToken, ok := strings.CutPrefix(header, "Bearer "); ok {
				if userID, err := resolveToken(r.Context(), pool, rawToken); err == nil {
					ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func resolveToken(ctx context.Context, pool *db.Pool, rawToken string) (string, error) {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := fmt.Sprintf("%x", hash)

	var userID string
	err := pool.QueryRow(ctx, `
		UPDATE api_tokens SET last_used = NOW()
		WHERE token_hash = $1
		RETURNING user_id
	`, tokenHash).Scan(&userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}
