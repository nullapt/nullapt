package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nullapt/nullapt/internal/blob"
	"github.com/nullapt/nullapt/internal/db"
	"github.com/nullapt/nullapt/internal/manifest"
)

// SkillsHandler groups all skill-related HTTP handlers.
type SkillsHandler struct {
	db   *db.Pool
	blob *blob.Client
}

func NewSkillsHandler(pool *db.Pool, blobClient *blob.Client) *SkillsHandler {
	return &SkillsHandler{db: pool, blob: blobClient}
}

// List handles GET /v1/skills?q=...
func (h *SkillsHandler) List(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")
	skills, err := h.db.ListSkills(r.Context(), search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "listing skills: "+err.Error())
		return
	}

	type item struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
		Author      string `json:"author"`
		Downloads   int64  `json:"downloads"`
		ManifestURL string `json:"manifest_url"`
		WASMUrl     string `json:"wasm_url"`
	}

	out := make([]item, len(skills))
	for i, s := range skills {
		out[i] = item{
			Name:        s.Name,
			Version:     s.Version,
			Description: s.Description,
			Author:      s.Author,
			Downloads:   s.Downloads,
			ManifestURL: fmt.Sprintf("/v1/skills/%s/manifest", s.Name),
			WASMUrl:     s.WASMUrl,
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// Get handles GET /v1/skills/{name} and GET /v1/skills/{name}/{version}
func (h *SkillsHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version") // empty string = latest

	skill, err := h.db.GetSkill(r.Context(), name, version)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	go h.db.IncrementDownloads(r.Context(), name, skill.Version) //nolint:errcheck

	writeJSON(w, http.StatusOK, map[string]any{
		"name":         skill.Name,
		"version":      skill.Version,
		"description":  skill.Description,
		"author":       skill.Author,
		"license":      skill.License,
		"downloads":    skill.Downloads,
		"manifest_url": fmt.Sprintf("/v1/skills/%s/manifest", skill.Name),
		"wasm_url":     skill.WASMUrl,
		"published_at": skill.PublishedAt,
	})
}

// Publish handles POST /v1/skills — multipart: "manifest" (SKILL.json) + "wasm" (binary).
func (h *SkillsHandler) Publish(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(contextKeyUserID).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// 64 MiB max request (4 MiB manifest + up to ~60 MiB WASM).
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "parsing multipart form: "+err.Error())
		return
	}

	// ── 1. Read and verify the manifest ──────────────────────────────────────

	manifestFile, _, err := r.FormFile("manifest")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'manifest' form field")
		return
	}
	defer manifestFile.Close()

	manifestData, err := io.ReadAll(io.LimitReader(manifestFile, 4<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "reading manifest: "+err.Error())
		return
	}

	skill, err := manifest.ParseBytes(manifestData)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid manifest: "+err.Error())
		return
	}
	if err := skill.Verify(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "signature verification failed: "+err.Error())
		return
	}

	// ── 2. Upload WASM to Vercel Blob ─────────────────────────────────────────

	wasmFile, wasmHeader, err := r.FormFile("wasm")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'wasm' form field")
		return
	}
	defer wasmFile.Close()

	blobPath := fmt.Sprintf("skills/%s/%s/%s", skill.Name, skill.Version, skill.Entry)

	// Buffer the WASM so we can report size — most WASM binaries are small.
	wasmData, err := io.ReadAll(io.LimitReader(wasmFile, 60<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "reading wasm: "+err.Error())
		return
	}
	_ = wasmHeader // available if we need filename validation later

	blobResult, err := h.blob.UploadStream(r.Context(), blobPath, bytes.NewReader(wasmData), int64(len(wasmData)))
	if err != nil {
		writeError(w, http.StatusBadGateway, "uploading wasm to blob storage: "+err.Error())
		return
	}

	// ── 3. Record in Postgres ─────────────────────────────────────────────────

	hash := sha256.Sum256(manifestData)
	manifestHash := fmt.Sprintf("%x", hash)

	if err := h.db.PublishSkill(r.Context(), userID, skill, blobResult.URL, manifestHash); err != nil {
		writeError(w, http.StatusConflict, "publishing skill: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"name":     skill.Name,
		"version":  skill.Version,
		"wasm_url": blobResult.URL,
		"status":   "published",
	})
}

// TransparencyLog handles GET /v1/skills/{name}/log
func (h *SkillsHandler) TransparencyLog(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	entries, err := h.db.GetTransparencyLog(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
