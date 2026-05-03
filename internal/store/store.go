package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nullapt/nullapt/internal/manifest"
)

// Store manages installed skills on disk under ~/.nullapt/skills/.
type Store struct {
	root string
}

// New returns a Store rooted at the user's default nullapt directory.
func New() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolving home directory: %w", err)
	}
	root := filepath.Join(home, ".nullapt", "skills")
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, fmt.Errorf("creating skill store at %s: %w", root, err)
	}
	return &Store{root: root}, nil
}

// safeJoin joins parts onto root and verifies the result stays inside root.
// This is a defence-in-depth check: callers should already validate inputs
// (see manifest.validateRelPath), but we re-check here so a future bug or
// unsanitised caller cannot escape the skill store.
func safeJoin(root string, parts ...string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolving store root: %w", err)
	}
	joined := filepath.Join(append([]string{rootAbs}, parts...)...)
	cleaned := filepath.Clean(joined)
	rel, err := filepath.Rel(rootAbs, cleaned)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes skill store", filepath.Join(parts...))
	}
	return cleaned, nil
}

func (s *Store) skillDir(name string) (string, error) {
	return safeJoin(s.root, name)
}

func (s *Store) manifestPath(name string) (string, error) {
	dir, err := s.skillDir(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "SKILL.json"), nil
}

// Install writes the manifest and WASM blob for a skill.
// wasmData may be nil when running without a real registry (dev mode).
func (s *Store) Install(skill *manifest.Skill, wasmData []byte) error {
	dir, err := s.skillDir(skill.Name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("creating skill directory: %w", err)
	}

	manifestData, err := json.MarshalIndent(skill, "", "  ")
	if err != nil {
		return fmt.Errorf("serialising manifest: %w", err)
	}
	mp, err := s.manifestPath(skill.Name)
	if err != nil {
		return err
	}
	if err := os.WriteFile(mp, manifestData, 0o640); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}

	if len(wasmData) > 0 {
		wasmPath, err := safeJoin(dir, skill.Entry)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(wasmPath), 0o750); err != nil {
			return fmt.Errorf("creating wasm directory: %w", err)
		}
		if err := os.WriteFile(wasmPath, wasmData, 0o640); err != nil {
			return fmt.Errorf("writing wasm blob: %w", err)
		}
	}
	return nil
}

// Remove deletes a skill and all its associated files.
func (s *Store) Remove(name string) error {
	dir, err := s.skillDir(name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("skill %q is not installed", name)
	}
	return os.RemoveAll(dir)
}

// Get returns the parsed manifest for an installed skill.
func (s *Store) Get(name string) (*manifest.Skill, error) {
	p, err := s.manifestPath(name)
	if err != nil {
		return nil, fmt.Errorf("skill %q: %w", name, err)
	}
	skill, err := manifest.Parse(p)
	if err != nil {
		return nil, fmt.Errorf("skill %q: %w", name, err)
	}
	return skill, nil
}

// List returns all installed skill manifests.
func (s *Store) List() ([]*manifest.Skill, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		return nil, fmt.Errorf("reading skill store: %w", err)
	}
	var skills []*manifest.Skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skill, err := s.Get(e.Name())
		if err != nil {
			continue // tolerate corrupt entries; don't fail the whole list
		}
		skills = append(skills, skill)
	}
	return skills, nil
}

// WASMPath returns the absolute path to the installed WASM blob.
func (s *Store) WASMPath(name string) (string, error) {
	skill, err := s.Get(name)
	if err != nil {
		return "", err
	}
	dir, err := s.skillDir(name)
	if err != nil {
		return "", err
	}
	p, err := safeJoin(dir, skill.Entry)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("wasm blob for skill %q not found at %s", name, p)
	}
	return p, nil
}
