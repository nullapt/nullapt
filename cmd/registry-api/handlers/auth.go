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

// AuthMiddleware validates the Bearer token (either an API token or a session id)
// and injects the user id into context. Rejects requests without a valid token.
func AuthMiddleware(pool *db.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := tryAuth(r, pool)
			if !ok {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth injects the user id if a valid token is present, but does not
// reject requests without one.
func OptionalAuth(pool *db.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if userID, ok := tryAuth(r, pool); ok {
				ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// tryAuth returns the user id and true if the Authorization header carries
// a valid API token (prefix "nlpt_") or a valid session id.
func tryAuth(r *http.Request, pool *db.Pool) (string, bool) {
	header := r.Header.Get("Authorization")
	rawToken, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || rawToken == "" {
		return "", false
	}

	if strings.HasPrefix(rawToken, "nlpt_") {
		userID, err := resolveAPIToken(r.Context(), pool, rawToken)
		if err == nil {
			return userID, true
		}
		return "", false
	}

	// Otherwise, treat as a session id (UUID).
	userID, err := pool.SessionUser(r.Context(), rawToken)
	if err == nil {
		return userID, true
	}
	return "", false
}

func resolveAPIToken(ctx context.Context, pool *db.Pool, rawToken string) (string, error) {
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
