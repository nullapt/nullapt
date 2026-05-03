package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nullapt/nullapt/internal/manifest"
)

const defaultBaseURL = "https://registry.nullapt.dev"

// Version is set by main at startup so registry requests advertise the CLI version.
// Cloudflare bot heuristics block the default Go-http-client UA, so this must be
// kept in sync with the binary's version.
var Version = "dev"

// Client talks to the NullApt registry API.
type Client struct {
	base   string
	http   *http.Client
	token  string // bearer token set after `nullapt login`
}

// New returns a Client pointed at the production registry.
func New(token string) *Client {
	return &Client{
		base:  defaultBaseURL,
		http:  &http.Client{Timeout: 30 * time.Second},
		token: token,
	}
}

// NewWithBase returns a Client pointed at a custom registry URL (dev/self-hosted).
func NewWithBase(base, token string) *Client {
	c := New(token)
	c.base = base
	return c
}

// SkillMeta is the registry's response for a skill lookup.
type SkillMeta struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Downloads   int64  `json:"downloads"`
	ManifestURL string `json:"manifest_url"`
	WASMUrl     string `json:"wasm_url"`
}

// Lookup fetches metadata for a skill by name (and optional @version).
func (c *Client) Lookup(ctx context.Context, name string) (*SkillMeta, error) {
	url := fmt.Sprintf("%s/v1/skills/%s", c.base, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("registry lookup: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("skill %q not found in registry", name)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned HTTP %d", resp.StatusCode)
	}

	var meta SkillMeta
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, fmt.Errorf("decoding registry response: %w", err)
	}
	return &meta, nil
}

// DownloadManifest fetches and parses the SKILL.json for a skill.
// manifest_url from the registry is a relative path; WASM URLs are absolute.
func (c *Client) DownloadManifest(ctx context.Context, meta *SkillMeta) (*manifest.Skill, error) {
	url := meta.ManifestURL
	if strings.HasPrefix(url, "/") {
		url = c.base + url
	}
	data, err := c.get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("downloading manifest: %w", err)
	}
	return manifest.ParseBytes(data)
}

// DownloadWASM fetches the raw WASM binary for a skill.
func (c *Client) DownloadWASM(ctx context.Context, meta *SkillMeta) ([]byte, error) {
	data, err := c.get(ctx, meta.WASMUrl)
	if err != nil {
		return nil, fmt.Errorf("downloading wasm: %w", err)
	}
	return data, nil
}

// Publish uploads a signed SKILL.json and its WASM binary to the registry.
// The manifest's Entry field is resolved relative to manifestDir.
func (c *Client) Publish(ctx context.Context, skill *manifest.Skill, manifestDir string) error {
	if c.token == "" {
		return fmt.Errorf("not authenticated — run `nullapt login` first")
	}

	wasmPath := filepath.Join(manifestDir, skill.Entry)
	wasmData, err := os.ReadFile(wasmPath)
	if err != nil {
		return fmt.Errorf("reading wasm binary at %s: %w", wasmPath, err)
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	// Part 1: SKILL.json
	manifestJSON, err := json.Marshal(skill)
	if err != nil {
		return fmt.Errorf("serialising manifest: %w", err)
	}
	mf, err := mw.CreateFormFile("manifest", "SKILL.json")
	if err != nil {
		return err
	}
	if _, err := mf.Write(manifestJSON); err != nil {
		return err
	}

	// Part 2: WASM binary
	wf, err := mw.CreateFormFile("wasm", skill.Entry)
	if err != nil {
		return err
	}
	if _, err := wf.Write(wasmData); err != nil {
		return err
	}

	if err := mw.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/v1/skills", &body)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("publish request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registry returned HTTP %d: %s", resp.StatusCode, raw)
	}
	return nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", fmt.Sprintf("nullapt-cli/%s (%s; %s)", Version, runtime.GOOS, runtime.GOARCH))
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}
