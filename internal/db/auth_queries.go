package db

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"
)

// User represents a registered user.
type User struct {
	ID          string
	Username    string
	Email       string
	GitHubID    *int64
	AvatarURL   *string
	DisplayName *string
	CreatedAt   time.Time
}

// UpsertGitHubUser inserts or updates a user identified by their GitHub ID.
// Returns the user's id (UUID).
func (p *Pool) UpsertGitHubUser(
	ctx context.Context,
	githubID int64,
	username, email, displayName, avatarURL string,
) (string, error) {
	var id string
	err := p.QueryRow(ctx, `
		INSERT INTO users (github_id, username, email, display_name, avatar_url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (github_id) DO UPDATE SET
			username = EXCLUDED.username,
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			avatar_url = EXCLUDED.avatar_url
		RETURNING id
	`, githubID, username, email, displayName, avatarURL).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("upserting github user: %w", err)
	}
	return id, nil
}

// GetUser fetches a user by id.
func (p *Pool) GetUser(ctx context.Context, userID string) (*User, error) {
	var u User
	err := p.QueryRow(ctx, `
		SELECT id, username, email, github_id, avatar_url, display_name, created_at
		FROM users WHERE id = $1
	`, userID).Scan(&u.ID, &u.Username, &u.Email, &u.GitHubID, &u.AvatarURL, &u.DisplayName, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ── Sessions ────────────────────────────────────────────────────────────────

// CreateSession creates a new browser session. Returns the session id (UUID).
func (p *Pool) CreateSession(ctx context.Context, userID string, ttl time.Duration) (string, error) {
	expiresAt := time.Now().Add(ttl)
	var id string
	err := p.QueryRow(ctx, `
		INSERT INTO sessions (user_id, expires_at) VALUES ($1, $2) RETURNING id
	`, userID, expiresAt).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("creating session: %w", err)
	}
	return id, nil
}

// SessionUser returns the user id for an active session, or empty string if invalid.
func (p *Pool) SessionUser(ctx context.Context, sessionID string) (string, error) {
	var userID string
	err := p.QueryRow(ctx, `
		SELECT user_id FROM sessions
		WHERE id = $1 AND expires_at > NOW()
	`, sessionID).Scan(&userID)
	return userID, err
}

// DeleteSession revokes a session.
func (p *Pool) DeleteSession(ctx context.Context, sessionID string) error {
	_, err := p.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, sessionID)
	return err
}

// ── API Tokens ──────────────────────────────────────────────────────────────

// APIToken is the metadata for a user's API token (without the raw value).
type APIToken struct {
	ID        string
	Name      string
	CreatedAt time.Time
	LastUsed  *time.Time
}

// CreateAPIToken stores the SHA-256 hash of a raw token. The caller is responsible
// for generating the raw token and showing it to the user exactly once.
func (p *Pool) CreateAPIToken(ctx context.Context, userID, name, rawToken string) (string, error) {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := fmt.Sprintf("%x", hash)

	var id string
	err := p.QueryRow(ctx, `
		INSERT INTO api_tokens (user_id, name, token_hash) VALUES ($1, $2, $3) RETURNING id
	`, userID, name, tokenHash).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("creating api token: %w", err)
	}
	return id, nil
}

// ListAPITokens returns all tokens for a user (metadata only, never the raw value).
func (p *Pool) ListAPITokens(ctx context.Context, userID string) ([]APIToken, error) {
	rows, err := p.Query(ctx, `
		SELECT id, name, created_at, last_used FROM api_tokens
		WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []APIToken
	for rows.Next() {
		var t APIToken
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt, &t.LastUsed); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

// RevokeAPIToken deletes a token if it belongs to the given user.
func (p *Pool) RevokeAPIToken(ctx context.Context, userID, tokenID string) error {
	tag, err := p.Exec(ctx, `
		DELETE FROM api_tokens WHERE id = $1 AND user_id = $2
	`, tokenID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("token not found")
	}
	return nil
}
