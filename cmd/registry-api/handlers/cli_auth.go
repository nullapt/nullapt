package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/nullapt/nullapt/internal/db"
)

// CLIAuthHandler manages the device-flow CLI login.
type CLIAuthHandler struct {
	db      *db.Pool
	siteURL string
}

func NewCLIAuthHandler(pool *db.Pool) *CLIAuthHandler {
	site := os.Getenv("SITE_URL")
	if site == "" {
		site = "https://nullapt.dev"
	}
	return &CLIAuthHandler{db: pool, siteURL: site}
}

// Init POST /v1/auth/cli/init
// Creates a pending auth request and returns the URL the user should visit.
func (h *CLIAuthHandler) Init(w http.ResponseWriter, r *http.Request) {
	id, err := h.db.CreateCLIAuthRequest(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "creating auth request: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":               id,
		"verification_url": h.siteURL + "/cli-auth?id=" + id,
		"expires_in":       900, // 15 minutes
	})
}

// Poll GET /v1/auth/cli/poll?id=...
// Returns 202 while pending, 200 with token when approved, 410 when expired.
func (h *CLIAuthHandler) Poll(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing 'id'")
		return
	}

	status, token, err := h.db.PollCLIAuthRequest(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "auth request not found")
		return
	}

	switch status {
	case db.CLIAuthApproved:
		writeJSON(w, http.StatusOK, map[string]any{"status": "approved", "token": token})
	case db.CLIAuthExpired:
		writeJSON(w, http.StatusGone, map[string]any{"status": "expired"})
	default:
		w.WriteHeader(http.StatusAccepted)
		writeJSON(w, http.StatusAccepted, map[string]any{"status": "pending"})
	}
}

// Approve POST /v1/auth/cli/approve  (requires auth)
// Called by the web app after the user signs in and confirms.
// Body: { "id": "<request_uuid>" }
func (h *CLIAuthHandler) Approve(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(contextKeyUserID).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
		writeError(w, http.StatusBadRequest, "missing 'id'")
		return
	}

	rawToken, err := generateRawToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "generating token")
		return
	}

	if err := h.db.ApproveCLIAuthRequest(r.Context(), body.ID, userID, rawToken); err != nil {
		writeError(w, http.StatusBadRequest, "approving request: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
