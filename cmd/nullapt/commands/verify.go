package commands

import (
	"fmt"

	"github.com/nullapt/nullapt/internal/manifest"
	"github.com/spf13/cobra"
)

func NewVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify <SKILL.json>",
		Short: "Verify the Ed25519 signature of a local SKILL.json",
		Long: `Parse and cryptographically verify a SKILL.json manifest without installing it.

Useful for auditing skills before running nullapt get, or for CI pipelines
that validate publisher artifacts before distribution.`,
		Args:    cobra.ExactArgs(1),
		Example: "  nullapt verify ./my-skill/SKILL.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			skill, err := manifest.Parse(path)
			if err != nil {
				return fmt.Errorf("parsing manifest: %w", err)
			}

			fmt.Printf("Manifest:  %s@%s\n", skill.Name, skill.Version)
			fmt.Printf("Author:    %s\n", skill.Author)
			fmt.Printf("Algorithm: %s\n", skill.Signature.Algorithm)
			fmt.Print("Signature: ")

			if err := skill.Verify(); err != nil {
				fmt.Println("INVALID")
				return err
			}

			fmt.Println("VALID")
			return nil
		},
	}
}
