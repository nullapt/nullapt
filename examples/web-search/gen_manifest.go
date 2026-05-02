//go:build ignore

// Run with: go run gen_manifest.go
// Generates a valid signed SKILL.json for development and testing.

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/nullapt/nullapt/internal/manifest"
)

func main() {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	skill := &manifest.Skill{
		SchemaVersion: "1.0",
		Name:          "web-search",
		Version:       "1.0.0",
		Description:   "Search the web using DuckDuckGo Instant Answers API",
		Author:        "nullapt-team",
		Homepage:      "https://nullapt.dev/skills/web-search",
		License:       "MIT",
		Permissions: manifest.Permissions{
			Network: manifest.NetworkPerm{
				Allowed: true,
				Domains: []string{"api.duckduckgo.com"},
			},
			Filesystem: manifest.FsPerm{Read: []string{}, Write: []string{}},
			Env:        []string{},
		},
		Entry: "skill.wasm",
		Interface: manifest.Interface{
			Tools: []manifest.Tool{
				{
					Name:        "web_search",
					Description: "Search the web and return top results",
					InputSchema: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"query": map[string]any{
								"type":        "string",
								"description": "The search query",
							},
						},
						"required": []string{"query"},
					},
				},
			},
		},
		Signature: manifest.Signature{
			Algorithm: "ed25519",
			PublicKey: base64.StdEncoding.EncodeToString(pub),
			Value:     "", // filled below
		},
	}

	payload, err := skill.PayloadBytes()
	if err != nil {
		panic(err)
	}
	sig := ed25519.Sign(priv, payload)
	skill.Signature.Value = base64.StdEncoding.EncodeToString(sig)

	out, err := json.MarshalIndent(skill, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("SKILL.json", out, 0o644); err != nil {
		panic(err)
	}
	fmt.Println("Wrote SKILL.json")
	fmt.Printf("Public key: %s\n", base64.StdEncoding.EncodeToString(pub))
}
