package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/nullapt/nullapt/internal/db"
)

// GitHubOAuthHandler handles the GitHub OAuth callback flow.
type GitHubOAuthHandler struct {
	db           *db.Pool
	clientID     string
	clientSecret string
}

func NewGitHubOAuthHandler(pool *db.Pool) *GitHubOAuthHandler {
	return &GitHubOAuthHandler{
		db:           pool,
		clientID:     os.Getenv("GITHUB_CLIENT_ID"),
		clientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	}
}

type ghTokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type ghUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type ghEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// Callback exchanges a GitHub authorization code for a session.
// POST /v1/auth/github/callback  body: { "code": "..." }
// Response: { "session_token": "...", "user": { ... } }
func (h *GitHubOAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	if h.clientID == "" || h.clientSecret == "" {
		writeError(w, http.StatusServiceUnavailable, "GitHub OAuth is not configured on this server")
		return
	}

	var body struct {
		Code        string `json:"code"`
		AccessToken string `json:"access_token"` // pre-exchanged by the web app
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || (body.Code == "" && body.AccessToken == "") {
		writeError(w, http.StatusBadRequest, "missing 'code' or 'access_token'")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// 1. Get GitHub access token (either exchange code or use pre-exchanged token)
	var accessToken string
	if body.AccessToken != "" {
		accessToken = body.AccessToken
	} else {
		var err error
		accessToken, err = h.exchangeCode(ctx, body.Code)
		if err != nil {
			writeError(w, http.StatusBadGateway, "exchanging github code: "+err.Error())
			return
		}
	}

	// 2. Fetch user info
	user, err := h.fetchUser(ctx, accessToken)
	if err != nil {
		writeError(w, http.StatusBadGateway, "fetching github user: "+err.Error())
		return
	}

	// 3. Fetch primary email if not on profile
	if user.Email == "" {
		email, err := h.fetchPrimaryEmail(ctx, accessToken)
		if err != nil {
			writeError(w, http.StatusBadGateway, "fetching github email: "+err.Error())
			return
		}
		user.Email = email
	}

	// 4. Upsert user
	userID, err := h.db.UpsertGitHubUser(ctx, user.ID, user.Login, user.Email, user.Name, user.AvatarURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "saving user: "+err.Error())
		return
	}

	// 5. Create a 30-day session
	sessionID, err := h.db.CreateSession(ctx, userID, 30*24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "creating session: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"session_token": sessionID,
		"user": map[string]any{
			"id":         userID,
			"username":   user.Login,
			"email":      user.Email,
			"name":       user.Name,
			"avatar_url": user.AvatarURL,
		},
	})
}

func (h *GitHubOAuthHandler) exchangeCode(ctx context.Context, code string) (string, error) {
	form := url.Values{}
	form.Set("client_id", h.clientID)
	form.Set("client_secret", h.clientSecret)
	form.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github returned %d: %s", resp.StatusCode, string(body))
	}

	var tok ghTokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", err
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("github did not return an access token: %s", string(body))
	}
	return tok.AccessToken, nil
}

func (h *GitHubOAuthHandler) fetchUser(ctx context.Context, accessToken string) (*ghUser, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github user endpoint %d: %s", resp.StatusCode, string(body))
	}

	var u ghUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (h *GitHubOAuthHandler) fetchPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github emails endpoint %d", resp.StatusCode)
	}

	var emails []ghEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	if len(emails) > 0 {
		return emails[0].Email, nil
	}
	return "", fmt.Errorf("no email found on github account")
}
