package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

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

func (s *Skill) validate() error {
	if s.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q (want %q)", s.SchemaVersion, SchemaVersion)
	}
	if s.Name == "" {
		return fmt.Errorf("manifest missing required field: name")
	}
	if s.Version == "" {
		return fmt.Errorf("manifest missing required field: version")
	}
	if s.Entry == "" {
		return fmt.Errorf("manifest missing required field: entry")
	}
	if s.Signature.Algorithm == "" || s.Signature.PublicKey == "" || s.Signature.Value == "" {
		return fmt.Errorf("manifest missing required signature fields")
	}
	// Reject wildcard network domains — every domain must be explicit.
	for _, d := range s.Permissions.Network.Domains {
		if strings.Contains(d, "*") {
			return fmt.Errorf("wildcard network domain %q is not allowed; list domains explicitly", d)
		}
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
