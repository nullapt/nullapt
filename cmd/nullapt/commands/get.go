package commands

import (
	"context"
	"fmt"

	"github.com/nullapt/nullapt/internal/registry"
	"github.com/nullapt/nullapt/internal/store"
	"github.com/spf13/cobra"
)

func NewGetCmd() *cobra.Command {
	var registryURL string
	var skipVerify bool

	cmd := &cobra.Command{
		Use:     "get <skill[@version]>",
		Aliases: []string{"i", "install"},
		Short:   "Install a skill from the registry",
		Long: `Download, verify, and install a skill from the NullApt registry.

Every skill is verified against its Ed25519 signature before installation.
Use --skip-verify only for local development with unsigned manifests.`,
		Args:    cobra.ExactArgs(1),
		Example: "  nullapt get web-search\n  nullapt get web-search@1.2.0",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			ctx := context.Background()

			s, err := store.New()
			if err != nil {
				return err
			}

			token, _ := loadToken()
			client := registry.NewWithBase(registryURL, token)

			fmt.Printf("Resolving %q from registry...\n", name)
			meta, err := client.Lookup(ctx, name)
			if err != nil {
				return err
			}

			fmt.Printf("Found %s@%s by %s\n", meta.Name, meta.Version, meta.Author)
			fmt.Println("Downloading manifest...")

			skill, err := client.DownloadManifest(ctx, meta)
			if err != nil {
				return err
			}

			if !skipVerify {
				fmt.Print("Verifying Ed25519 signature... ")
				if err := skill.Verify(); err != nil {
					fmt.Println("FAILED")
					return fmt.Errorf("signature verification: %w", err)
				}
				fmt.Println("OK")
			} else {
				fmt.Println("Warning: skipping signature verification (--skip-verify)")
			}

			fmt.Println("Downloading WASM module...")
			wasmData, err := client.DownloadWASM(ctx, meta)
			if err != nil {
				return err
			}

			if err := s.Install(skill, wasmData); err != nil {
				return err
			}

			fmt.Printf("\nInstalled %s@%s\n", skill.Name, skill.Version)
			fmt.Printf("  Tools: ")
			for i, t := range skill.Interface.Tools {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(t.Name)
			}
			fmt.Println()
			if skill.Permissions.Network.Allowed {
				fmt.Printf("  Network access: %v\n", skill.Permissions.Network.Domains)
			} else {
				fmt.Println("  Network access: none (offline-only)")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&registryURL, "registry", "https://registry.nullapt.dev", "Registry base URL")
	cmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "Skip signature verification (dev only)")
	return cmd
}
