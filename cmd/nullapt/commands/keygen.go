package commands

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewKeygenCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "Generate an Ed25519 signing keypair",
		Long: `Generate a new Ed25519 keypair for signing skill manifests.

The private key is saved as private.pem (mode 0600) and the public key as
public.pem in the output directory. Keep your private key safe — anyone who
holds it can publish skills under your identity.`,
		Example: "  nullapt keygen\n  nullapt keygen --output ./my-keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			if outputDir == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				outputDir = filepath.Join(home, ".nullapt", "keys")
			}

			if err := os.MkdirAll(outputDir, 0700); err != nil {
				return fmt.Errorf("creating key directory: %w", err)
			}

			pub, priv, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				return fmt.Errorf("generating keypair: %w", err)
			}

			privDER, err := x509.MarshalPKCS8PrivateKey(priv)
			if err != nil {
				return fmt.Errorf("encoding private key: %w", err)
			}
			pubDER, err := x509.MarshalPKIXPublicKey(pub)
			if err != nil {
				return fmt.Errorf("encoding public key: %w", err)
			}

			privPath := filepath.Join(outputDir, "private.pem")
			pubPath := filepath.Join(outputDir, "public.pem")

			if err := os.WriteFile(privPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER}), 0600); err != nil {
				return fmt.Errorf("writing private key: %w", err)
			}
			if err := os.WriteFile(pubPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}), 0644); err != nil {
				return fmt.Errorf("writing public key: %w", err)
			}

			fmt.Println("Generated Ed25519 keypair:")
			fmt.Printf("  Private key: %s\n", privPath)
			fmt.Printf("  Public key:  %s\n", pubPath)
			fmt.Println()
			fmt.Println("Keep your private key secret. Sign your manifest with:")
			fmt.Println("  nullapt sign SKILL.json")
			return nil
		},
	}

	cmd.Flags().StringVar(&outputDir, "output", "", "Directory to write keys (default: ~/.nullapt/keys)")
	return cmd
}
