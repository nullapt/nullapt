package commands

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nullapt/nullapt/internal/manifest"
	"github.com/spf13/cobra"
)

func NewSignCmd() *cobra.Command {
	var keyPath string

	cmd := &cobra.Command{
		Use:   "sign <SKILL.json>",
		Short: "Sign a SKILL.json manifest with your Ed25519 private key",
		Long: `Sign a skill manifest in-place with your Ed25519 private key.

The signature fields (algorithm, public_key, value) are written directly into
the manifest file. Run this before nullapt publish.`,
		Args:    cobra.ExactArgs(1),
		Example: "  nullapt sign ./my-skill/SKILL.json\n  nullapt sign ./my-skill/SKILL.json --key ./keys/private.pem",
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath := args[0]

			if keyPath == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				keyPath = filepath.Join(home, ".nullapt", "keys", "private.pem")
			}

			priv, pub, err := loadPrivateKey(keyPath)
			if err != nil {
				return fmt.Errorf("loading private key from %s: %w\n\nRun `nullapt keygen` to generate a keypair", keyPath, err)
			}

			skill, err := manifest.ParseUnsigned(manifestPath)
			if err != nil {
				return fmt.Errorf("parsing manifest: %w", err)
			}

			skill.Signature.Algorithm = "ed25519"
			skill.Signature.PublicKey = base64.StdEncoding.EncodeToString(pub)
			skill.Signature.Value = ""

			payload, err := skill.PayloadBytes()
			if err != nil {
				return fmt.Errorf("computing payload: %w", err)
			}

			skill.Signature.Value = base64.StdEncoding.EncodeToString(ed25519.Sign(priv, payload))

			out, err := json.MarshalIndent(skill, "", "  ")
			if err != nil {
				return fmt.Errorf("serialising manifest: %w", err)
			}
			if err := os.WriteFile(manifestPath, append(out, '\n'), 0644); err != nil {
				return fmt.Errorf("writing manifest: %w", err)
			}

			fmt.Printf("Signed %s@%s\n", skill.Name, skill.Version)
			fmt.Printf("  Algorithm:  ed25519\n")
			fmt.Printf("  Public key: %s\n", skill.Signature.PublicKey)
			return nil
		},
	}

	cmd.Flags().StringVar(&keyPath, "key", "", "Path to private key PEM (default: ~/.nullapt/keys/private.pem)")
	return cmd
}

func loadPrivateKey(path string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, nil, fmt.Errorf("no PEM block found in %s", path)
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing key: %w", err)
	}
	edKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("key is not Ed25519 (got %T)", key)
	}
	return edKey, edKey.Public().(ed25519.PublicKey), nil
}
