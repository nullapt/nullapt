package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/nullapt/nullapt/internal/manifest"
	"github.com/nullapt/nullapt/internal/registry"
	"github.com/spf13/cobra"
)

func NewPublishCmd() *cobra.Command {
	var registryURL string

	cmd := &cobra.Command{
		Use:   "publish <SKILL.json>",
		Short: "Publish a skill to the registry",
		Long: `Sign and publish a skill manifest + WASM bundle to the NullApt registry.

You must be logged in (nullapt login) and the manifest must carry a valid
Ed25519 signature produced with your registered key pair.`,
		Args:    cobra.ExactArgs(1),
		Example: "  nullapt publish ./my-skill/SKILL.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			ctx := context.Background()

			skill, err := manifest.Parse(path)
			if err != nil {
				return fmt.Errorf("parsing manifest: %w", err)
			}

			fmt.Print("Verifying signature before publish... ")
			if err := skill.Verify(); err != nil {
				fmt.Println("FAILED")
				return err
			}
			fmt.Println("OK")

			token, err := loadToken()
			if err != nil {
				return fmt.Errorf("not logged in: run `nullapt login` first")
			}

			client := registry.NewWithBase(registryURL, token)
			if err := client.Publish(ctx, skill, filepath.Dir(path)); err != nil {
				return err
			}

			fmt.Printf("Published %s@%s\n", skill.Name, skill.Version)
			return nil
		},
	}

	cmd.Flags().StringVar(&registryURL, "registry", "https://registry.nullapt.dev", "Registry base URL")
	return cmd
}
