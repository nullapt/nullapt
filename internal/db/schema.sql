-- NullApt Registry — Postgres Schema
-- Run with: psql $DATABASE_URL -f schema.sql

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─── Users ───────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username    TEXT UNIQUE NOT NULL,
    email       TEXT UNIQUE NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- API tokens (hashed — we never store the plaintext token)
CREATE TABLE IF NOT EXISTS api_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,                     -- human label e.g. "CI token"
    token_hash  TEXT UNIQUE NOT NULL,              -- SHA-256 of the raw token
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used   TIMESTAMPTZ
);

-- ─── Skills ──────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS skills (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT UNIQUE NOT NULL,
    owner_id    UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS skill_versions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id        UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    version         TEXT NOT NULL,
    description     TEXT NOT NULL,
    license         TEXT NOT NULL,
    manifest_json   JSONB NOT NULL,                -- full SKILL.json content
    wasm_size_bytes BIGINT NOT NULL DEFAULT 0,
    wasm_url        TEXT NOT NULL,                 -- Blob storage URL
    downloads       BIGINT NOT NULL DEFAULT 0,
    published_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (skill_id, version)
);

-- ─── Transparency Log ────────────────────────────────────────────────────────
-- Append-only. Every publish event is recorded with the signing public key
-- so users can audit key rotation and detect compromised keys.

CREATE TABLE IF NOT EXISTS transparency_log (
    id              BIGSERIAL PRIMARY KEY,
    skill_id        UUID NOT NULL REFERENCES skills(id),
    version         TEXT NOT NULL,
    author_username TEXT NOT NULL,
    public_key_b64  TEXT NOT NULL,                 -- Ed25519 public key (base64)
    manifest_hash   TEXT NOT NULL,                 -- SHA-256 of canonical manifest JSON
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Indexes ─────────────────────────────────────────────────────────────────

CREATE INDEX IF NOT EXISTS idx_skill_versions_skill_id ON skill_versions(skill_id);
CREATE INDEX IF NOT EXISTS idx_transparency_log_skill_id ON transparency_log(skill_id);
CREATE INDEX IF NOT EXISTS idx_transparency_log_public_key ON transparency_log(public_key_b64);
CREATE INDEX IF NOT EXISTS idx_skill_versions_manifest ON skill_versions USING gin(manifest_json);
