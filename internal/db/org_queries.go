package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Organization represents a GitHub org mirrored into our DB.
type Organization struct {
	ID          string
	GitHubOrgID int64
	Login       string
	DisplayName *string
	AvatarURL   *string
	CreatedAt   time.Time
}

// UpsertOrganization inserts or updates an org row keyed by github_org_id and
// returns the internal UUID. Login changes (org renames) are picked up here.
func (p *Pool) UpsertOrganization(ctx context.Context, githubOrgID int64, login, displayName, avatarURL string) (string, error) {
	var id string
	err := p.QueryRow(ctx, `
		INSERT INTO organizations (github_org_id, login, display_name, avatar_url)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''))
		ON CONFLICT (github_org_id) DO UPDATE SET
			login = EXCLUDED.login,
			display_name = COALESCE(EXCLUDED.display_name, organizations.display_name),
			avatar_url = COALESCE(EXCLUDED.avatar_url, organizations.avatar_url)
		RETURNING id
	`, githubOrgID, login, displayName, avatarURL).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("upserting organization %q: %w", login, err)
	}
	return id, nil
}

// ReplaceUserOrgMemberships sets the user's org membership rows to exactly
// the orgIDs given, removing any membership not in the list. Run this after
// every login so revocations on GitHub propagate.
func (p *Pool) ReplaceUserOrgMemberships(ctx context.Context, userID string, orgIDs []string) error {
	tx, err := p.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `DELETE FROM organization_members WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("clearing memberships: %w", err)
	}
	for _, orgID := range orgIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO organization_members (org_id, user_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, orgID, userID); err != nil {
			return fmt.Errorf("inserting membership: %w", err)
		}
	}
	return tx.Commit(ctx)
}

// GetUserOrgs returns every org the user is currently a member of.
func (p *Pool) GetUserOrgs(ctx context.Context, userID string) ([]Organization, error) {
	rows, err := p.Query(ctx, `
		SELECT o.id, o.github_org_id, o.login, o.display_name, o.avatar_url, o.created_at
		FROM organizations o
		JOIN organization_members m ON m.org_id = o.id
		WHERE m.user_id = $1
		ORDER BY o.login
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Organization
	for rows.Next() {
		var o Organization
		if err := rows.Scan(&o.ID, &o.GitHubOrgID, &o.Login, &o.DisplayName, &o.AvatarURL, &o.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// GetOrgByLogin looks up an org by its GitHub login (e.g. "nullapt").
// Returns ("", nil) if the org doesn't exist in our mirror yet.
func (p *Pool) GetOrgByLogin(ctx context.Context, login string) (string, error) {
	var id string
	err := p.QueryRow(ctx, `SELECT id FROM organizations WHERE login = $1`, login).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

// IsOrgMember reports whether the given user is a member of the given org.
func (p *Pool) IsOrgMember(ctx context.Context, userID, orgID string) (bool, error) {
	var exists bool
	err := p.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM organization_members WHERE user_id = $1 AND org_id = $2)
	`, userID, orgID).Scan(&exists)
	return exists, err
}

// SkillOwnership describes who can publish new versions of a skill.
// Either OwnerUserID or OwnerOrgID is set; the other is the empty string.
type SkillOwnership struct {
	SkillID     string
	OwnerUserID string
	OwnerOrgID  string
}

// GetSkillOwnership returns ownership info for an existing skill, or
// (nil, nil) if the skill doesn't exist (i.e. this would be a first publish).
func (p *Pool) GetSkillOwnership(ctx context.Context, name string) (*SkillOwnership, error) {
	var o SkillOwnership
	var ownerOrgID *string
	err := p.QueryRow(ctx, `
		SELECT id, owner_id, owner_org_id FROM skills WHERE name = $1
	`, name).Scan(&o.SkillID, &o.OwnerUserID, &ownerOrgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if ownerOrgID != nil {
		o.OwnerOrgID = *ownerOrgID
	}
	return &o, nil
}
