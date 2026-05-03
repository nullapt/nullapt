package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

// nameRE restricts skill names to a safe alphabet that cannot escape a
// directory: lowercase letters, digits, hyphen, underscore, and a single
// optional namespace segment separated by `/` (e.g. "acme/web-search").
var nameRE = regexp.MustCompile(`^[a-z0-9][a-z0-9_\-]*(/[a-z0-9][a-z0-9_\-]*)?$`)

const SchemaVersion = "1.0"

// Skill is the parsed representation of a SKILL.json manifest.
type Skill struct {
	SchemaVersion string      `json:"schema_version"`
	Name          string      `json:"name"`
	Version       string      `json:"version"`
	Description   string      `json:"description"`
	Author        string      `json:"author"`
	Homepage      string      `json:"homepage,omitempty"`
	License       string      `json:"license"`
	Permissions   Permissions `json:"permissions"`
	Entry         string      `json:"entry"`
	Interface     Interface   `json:"interface"`
	Signature     Signature   `json:"signature"`
}

type Permissions struct {
	Network    NetworkPerm `json:"network"`
	Filesystem FsPerm      `json:"filesystem"`
	Env        []string    `json:"env"`
}

type NetworkPerm struct {
	Allowed bool     `json:"allowed"`
	Domains []string `json:"domains,omitempty"`
}

type FsPerm struct {
	Read  []string `json:"read"`
	Write []string `json:"write"`
}

type Interface struct {
	Tools []Tool `json:"tools"`
}

type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

type Signature struct {
	Algorithm string `json:"algorithm"`
	PublicKey string `json:"public_key"`
	Value     string `json:"value"`
}

// Parse reads and decodes a SKILL.json file from path.
func Parse(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}
	return ParseBytes(data)
}

// ParseBytes decodes SKILL.json from raw bytes.
func ParseBytes(data []byte) (*Skill, error) {
	var s Skill
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	if err := s.validate(); err != nil {
		return nil, err
	}
	return &s, nil
}

// ParseUnsigned reads and decodes a SKILL.json without requiring signature
// fields to be present. Use this before signing, when the manifest may not
// yet have a signature.
func ParseUnsigned(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}
	var s Skill
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	if err := s.validateStructure(); err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *Skill) validate() error {
	if err := s.validateStructure(); err != nil {
		return err
	}
	if s.Signature.Algorithm == "" || s.Signature.PublicKey == "" || s.Signature.Value == "" {
		return fmt.Errorf("manifest missing required signature fields")
	}
	return nil
}

// validateStructure checks all manifest fields except the signature block,
// used when parsing a manifest that hasn't been signed yet.
func (s *Skill) validateStructure() error {
	if s.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q (want %q)", s.SchemaVersion, SchemaVersion)
	}
	if s.Name == "" {
		return fmt.Errorf("manifest missing required field: name")
	}
	if !nameRE.MatchString(s.Name) {
		return fmt.Errorf("invalid skill name %q: must match %s", s.Name, nameRE)
	}
	if s.Version == "" {
		return fmt.Errorf("manifest missing required field: version")
	}
	if s.Entry == "" {
		return fmt.Errorf("manifest missing required field: entry")
	}
	if err := validateRelPath(s.Entry, "entry"); err != nil {
		return err
	}
	for _, d := range s.Permissions.Network.Domains {
		if strings.Contains(d, "*") {
			return fmt.Errorf("wildcard network domain %q is not allowed; list domains explicitly", d)
		}
	}
	return nil
}

// validateRelPath ensures p is a safe, contained relative path with no
// traversal, no absolute prefix, and no NUL bytes. It is used for fields
// (such as Entry) that are joined onto a trusted root directory.
func validateRelPath(p, field string) error {
	if p == "" {
		return fmt.Errorf("manifest %s is empty", field)
	}
	if strings.ContainsRune(p, 0) {
		return fmt.Errorf("manifest %s contains NUL byte", field)
	}
	if strings.ContainsRune(p, '\\') {
		return fmt.Errorf("manifest %s must use forward slashes", field)
	}
	if path.IsAbs(p) || strings.HasPrefix(p, "/") {
		return fmt.Errorf("manifest %s must be a relative path", field)
	}
	cleaned := path.Clean(p)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") || cleaned == "." {
		return fmt.Errorf("manifest %s %q escapes skill directory", field, p)
	}
	return nil
}

// PayloadBytes returns the canonical byte representation used for signing.
// The signature field itself is excluded before hashing.
func (s *Skill) PayloadBytes() ([]byte, error) {
	type payloadShape struct {
		SchemaVersion string      `json:"schema_version"`
		Name          string      `json:"name"`
		Version       string      `json:"version"`
		Description   string      `json:"description"`
		Author        string      `json:"author"`
		Homepage      string      `json:"homepage,omitempty"`
		License       string      `json:"license"`
		Permissions   Permissions `json:"permissions"`
		Entry         string      `json:"entry"`
		Interface     Interface   `json:"interface"`
	}
	p := payloadShape{
		SchemaVersion: s.SchemaVersion,
		Name:          s.Name,
		Version:       s.Version,
		Description:   s.Description,
		Author:        s.Author,
		Homepage:      s.Homepage,
		License:       s.License,
		Permissions:   s.Permissions,
		Entry:         s.Entry,
		Interface:     s.Interface,
	}
	return json.Marshal(p)
}
