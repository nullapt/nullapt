package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
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
		PublishedAt string `json:"published_at"`
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
			PublishedAt: s.PublishedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// Get handles all GET /v1/skills/* lookups. The wildcard path is parsed as:
//
//	foo                       → name=foo
//	foo@1.0.0                 → name=foo, version=1.0.0
//	ns/foo                    → name=ns/foo
//	ns/foo@1.0.0              → name=ns/foo, version=1.0.0
//	(any of the above)/log    → transparency log for that name
func (h *SkillsHandler) Get(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "*")
	if strings.HasSuffix(path, "/log") {
		name := strings.TrimSuffix(path, "/log")
		h.transparencyLogFor(w, r, name)
		return
	}
	if strings.HasSuffix(path, "/manifest") {
		name := strings.TrimSuffix(path, "/manifest")
		h.manifestFor(w, r, name)
		return
	}

	name := path
	version := ""
	if i := strings.LastIndexByte(path, '@'); i > 0 {
		name = path[:i]
		version = path[i+1:]
	}
	if name == "" {
		writeError(w, http.StatusBadRequest, "missing skill name")
		return
	}

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

	// ── 1b. Resolve namespace and ownership ──────────────────────────────────
	// Names of the form "<org>/<skill>" are org-namespaced; the user must be
	// a member of that org. Single-segment names belong to the publisher.
	var ownerOrgID string
	if orgLogin, _, ok := strings.Cut(skill.Name, "/"); ok {
		orgID, err := h.db.GetOrgByLogin(r.Context(), orgLogin)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeError(w, http.StatusForbidden, fmt.Sprintf("org %q is not connected — visit https://nullapt.dev/settings/organizations to connect it, then publish again", orgLogin))
			} else {
				writeError(w, http.StatusInternalServerError, "looking up org: "+err.Error())
			}
			return
		}
		isMember, err := h.db.IsOrgMember(r.Context(), userID, orgID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "checking org membership: "+err.Error())
			return
		}
		if !isMember {
			writeError(w, http.StatusForbidden, fmt.Sprintf("you are not a public member of %q — set your membership to Public on GitHub, then re-sync at https://nullapt.dev/settings/organizations", orgLogin))
			return
		}
		ownerOrgID = orgID
	}

	// 1c. If the skill already exists, enforce that the publisher still owns it.
	existing, err := h.db.GetSkillOwnership(r.Context(), skill.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "checking ownership: "+err.Error())
		return
	}
	if existing != nil {
		if existing.OwnerOrgID != "" {
			isMember, err := h.db.IsOrgMember(r.Context(), userID, existing.OwnerOrgID)
			if err != nil || !isMember {
				writeError(w, http.StatusForbidden, "you are not a member of the org that owns this skill")
				return
			}
		} else if existing.OwnerUserID != userID {
			writeError(w, http.StatusForbidden, "this skill is owned by another user")
			return
		}
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

	if err := h.db.PublishSkill(r.Context(), userID, ownerOrgID, skill, blobResult.URL, manifestHash); err != nil {
		writeError(w, http.StatusConflict, "publishing skill: "+err.Error())
		return
	}

	notifyRevalidate(skill.Name)

	writeJSON(w, http.StatusCreated, map[string]any{
		"name":     skill.Name,
		"version":  skill.Version,
		"wasm_url": blobResult.URL,
		"status":   "published",
	})
}

// notifyRevalidate pings the web app's revalidate webhook so the homepage
// and skill page caches are marked stale. Fire-and-forget: a webhook failure
// must not surface to the publishing client.
func notifyRevalidate(skillName string) {
	url := os.Getenv("WEB_REVALIDATE_URL")
	secret := os.Getenv("WEB_REVALIDATE_SECRET")
	if url == "" || secret == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		body := fmt.Sprintf(`{"skill":%q}`, skillName)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Revalidate-Secret", secret)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return
		}
		_ = resp.Body.Close()
	}()
}

// manifestFor returns the raw signed SKILL.json for the latest version of a
// given skill. The CLI fetches this to verify signatures locally before install.
func (h *SkillsHandler) manifestFor(w http.ResponseWriter, r *http.Request, name string) {
	manifestJSON, err := h.db.GetSkillManifest(r.Context(), name, "")
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(manifestJSON) //nolint:errcheck
}

// transparencyLogFor renders the transparency log for a given skill name.
// Called from Get() when the URL ends in /log.
func (h *SkillsHandler) transparencyLogFor(w http.ResponseWriter, r *http.Request, name string) {
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
