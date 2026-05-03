package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nullapt/nullapt/internal/db"
)

// TokensHandler manages user-created API tokens (used by the CLI).
type TokensHandler struct {
	db *db.Pool
}

func NewTokensHandler(pool *db.Pool) *TokensHandler {
	return &TokensHandler{db: pool}
}

// generateRawToken returns a 32-byte URL-safe random token prefixed with "nlpt_".
// Total length ≈ 48 chars. The prefix lets us distinguish API tokens from session ids.
func generateRawToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "nlpt_" + base64.RawURLEncoding.EncodeToString(b), nil
}

// Create POST /v1/tokens  body: { "name": "..." }
// Returns the raw token exactly once; only its hash is stored.
func (h *TokensHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(contextKeyUserID).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		writeError(w, http.StatusBadRequest, "missing 'name'")
		return
	}

	rawToken, err := generateRawToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "generating token: "+err.Error())
		return
	}

	id, err := h.db.CreateAPIToken(r.Context(), userID, body.Name, rawToken)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "saving token: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":    id,
		"name":  body.Name,
		"token": rawToken, // shown once, never again
	})
}

// List GET /v1/tokens
func (h *TokensHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(contextKeyUserID).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	tokens, err := h.db.ListAPITokens(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "listing tokens: "+err.Error())
		return
	}

	type item struct {
		ID        string  `json:"id"`
		Name      string  `json:"name"`
		CreatedAt string  `json:"created_at"`
		LastUsed  *string `json:"last_used"`
	}

	out := make([]item, len(tokens))
	for i, t := range tokens {
		var last *string
		if t.LastUsed != nil {
			s := t.LastUsed.Format("2006-01-02T15:04:05Z")
			last = &s
		}
		out[i] = item{
			ID:        t.ID,
			Name:      t.Name,
			CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
			LastUsed:  last,
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// Revoke DELETE /v1/tokens/{id}
func (h *TokensHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(contextKeyUserID).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	tokenID := chi.URLParam(r, "id")
	if err := h.db.RevokeAPIToken(r.Context(), userID, tokenID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Me GET /v1/me — returns the authenticated user.
func Me(pool *db.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextKeyUserID).(string)
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		user, err := pool.GetUser(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"name":       user.DisplayName,
			"avatar_url": user.AvatarURL,
		})
	}
}

// MyOrgs GET /v1/me/orgs — returns the orgs the current user can publish under.
func MyOrgs(pool *db.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextKeyUserID).(string)
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		orgs, err := pool.GetUserOrgs(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "fetching orgs: "+err.Error())
			return
		}
		out := make([]map[string]any, 0, len(orgs))
		for _, o := range orgs {
			out = append(out, map[string]any{
				"id":         o.ID,
				"login":      o.Login,
				"name":       o.DisplayName,
				"avatar_url": o.AvatarURL,
			})
		}
		writeJSON(w, http.StatusOK, out)
	}
}

// Logout DELETE /v1/auth/logout — invalidates the current session.
// Looks for a session id in the Authorization header.
func Logout(pool *db.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if len(header) > 7 && header[:7] == "Bearer " {
			sessionID := header[7:]
			// Best-effort delete — also works if it's an api token (no-op).
			pool.DeleteSession(r.Context(), sessionID) //nolint:errcheck
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
