package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nullapt/nullapt/internal/manifest"
)

// SkillRow is a single row from skill_versions joined with skills/users.
type SkillRow struct {
	Name        string
	Version     string
	Description string
	Author      string
	License     string
	WASMUrl     string
	ManifestURL string
	Downloads   int64
	PublishedAt time.Time
}

// ListSkills returns all latest-version skills, optionally filtered by a search term.
func (p *Pool) ListSkills(ctx context.Context, search string) ([]SkillRow, error) {
	query := `
		SELECT DISTINCT ON (s.name)
			s.name, sv.version, sv.description, COALESCE(o.login, u.username) AS author,
			sv.license, sv.wasm_url, sv.downloads, sv.published_at
		FROM skills s
		JOIN skill_versions sv ON sv.skill_id = s.id
		JOIN users u ON u.id = s.owner_id
			LEFT JOIN organizations o ON o.id = s.owner_org_id
		WHERE ($1 = '' OR s.name ILIKE '%' || $1 || '%' OR sv.description ILIKE '%' || $1 || '%')
		ORDER BY s.name, sv.published_at DESC
	`
	rows, err := p.Query(ctx, query, search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SkillRow
	for rows.Next() {
		var r SkillRow
		if err := rows.Scan(&r.Name, &r.Version, &r.Description, &r.Author,
			&r.License, &r.WASMUrl, &r.Downloads, &r.PublishedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// GetSkill fetches a specific skill version (empty version = latest).
func (p *Pool) GetSkill(ctx context.Context, name, version string) (*SkillRow, error) {
	var query string
	var args []any

	if version == "" {
		query = `
			SELECT s.name, sv.version, sv.description, COALESCE(o.login, u.username) AS author,
			       sv.license, sv.wasm_url, sv.downloads, sv.published_at
			FROM skills s
			JOIN skill_versions sv ON sv.skill_id = s.id
			JOIN users u ON u.id = s.owner_id
			LEFT JOIN organizations o ON o.id = s.owner_org_id
			WHERE s.name = $1
			ORDER BY sv.published_at DESC
			LIMIT 1
		`
		args = []any{name}
	} else {
		query = `
			SELECT s.name, sv.version, sv.description, COALESCE(o.login, u.username) AS author,
			       sv.license, sv.wasm_url, sv.downloads, sv.published_at
			FROM skills s
			JOIN skill_versions sv ON sv.skill_id = s.id
			JOIN users u ON u.id = s.owner_id
			LEFT JOIN organizations o ON o.id = s.owner_org_id
			WHERE s.name = $1 AND sv.version = $2
		`
		args = []any{name, version}
	}

	var r SkillRow
	err := p.QueryRow(ctx, query, args...).Scan(
		&r.Name, &r.Version, &r.Description, &r.Author,
		&r.License, &r.WASMUrl, &r.Downloads, &r.PublishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("skill %q not found: %w", name, err)
	}
	return &r, nil
}

// GetSkillManifest returns the raw manifest_json for a skill version
// (empty version = latest published).
func (p *Pool) GetSkillManifest(ctx context.Context, name, version string) ([]byte, error) {
	var manifestJSON []byte
	var err error
	if version == "" {
		err = p.QueryRow(ctx, `
			SELECT sv.manifest_json
			FROM skills s
			JOIN skill_versions sv ON sv.skill_id = s.id
			WHERE s.name = $1
			ORDER BY sv.published_at DESC
			LIMIT 1
		`, name).Scan(&manifestJSON)
	} else {
		err = p.QueryRow(ctx, `
			SELECT sv.manifest_json
			FROM skills s
			JOIN skill_versions sv ON sv.skill_id = s.id
			WHERE s.name = $1 AND sv.version = $2
		`, name, version).Scan(&manifestJSON)
	}
	if err != nil {
		return nil, fmt.Errorf("manifest for %q not found: %w", name, err)
	}
	return manifestJSON, nil
}

// PublishSkill inserts or updates a skill version and appends to the transparency log.
func (p *Pool) PublishSkill(ctx context.Context, ownerID, ownerOrgID string, skill *manifest.Skill, wasmURL, manifestHash string) error {
	manifestJSON, err := json.Marshal(skill)
	if err != nil {
		return err
	}

	tx, err := p.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Upsert skill record. Ownership is fixed at first publish; the handler
	// has already verified the caller is allowed to publish new versions.
	var orgArg interface{}
	if ownerOrgID != "" {
		orgArg = ownerOrgID
	}
	var skillID string
	err = tx.QueryRow(ctx, `
		INSERT INTO skills (name, owner_id, owner_org_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE SET owner_id = skills.owner_id
		RETURNING id
	`, skill.Name, ownerID, orgArg).Scan(&skillID)
	if err != nil {
		return fmt.Errorf("upserting skill: %w", err)
	}

	// Insert version (error on duplicate).
	_, err = tx.Exec(ctx, `
		INSERT INTO skill_versions (skill_id, version, description, license, manifest_json, wasm_url)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, skillID, skill.Version, skill.Description, skill.License, manifestJSON, wasmURL)
	if err != nil {
		return fmt.Errorf("inserting skill version: %w", err)
	}

	// Append transparency log entry.
	_, err = tx.Exec(ctx, `
		INSERT INTO transparency_log (skill_id, version, author_username, public_key_b64, manifest_hash)
		VALUES ($1, $2, $3, $4, $5)
	`, skillID, skill.Version, skill.Author, skill.Signature.PublicKey, manifestHash)
	if err != nil {
		return fmt.Errorf("appending transparency log: %w", err)
	}

	return tx.Commit(ctx)
}

// TransparencyLogEntry is one row from the transparency log.
type TransparencyLogEntry struct {
	Version      string
	Author       string
	PublicKeyB64 string
	ManifestHash string
	RecordedAt   time.Time
}

// GetTransparencyLog returns all transparency log entries for a skill.
func (p *Pool) GetTransparencyLog(ctx context.Context, skillName string) ([]TransparencyLogEntry, error) {
	rows, err := p.Query(ctx, `
		SELECT tl.version, tl.author_username, tl.public_key_b64, tl.manifest_hash, tl.recorded_at
		FROM transparency_log tl
		JOIN skills s ON s.id = tl.skill_id
		WHERE s.name = $1
		ORDER BY tl.id ASC
	`, skillName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []TransparencyLogEntry
	for rows.Next() {
		var e TransparencyLogEntry
		if err := rows.Scan(&e.Version, &e.Author, &e.PublicKeyB64, &e.ManifestHash, &e.RecordedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// IncrementDownloads bumps the download counter for a specific skill version.
func (p *Pool) IncrementDownloads(ctx context.Context, skillName, version string) {
	p.Exec(ctx, `
		UPDATE skill_versions sv
		SET downloads = downloads + 1
		FROM skills s
		WHERE sv.skill_id = s.id AND s.name = $1 AND sv.version = $2
	`, skillName, version) //nolint:errcheck
}
