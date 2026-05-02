package main

import (
	"fmt"
	"os"

	"github.com/nullapt/nullapt/cmd/nullapt/commands"
	"github.com/spf13/cobra"
)

var version = "0.1.0-dev"

func main() {
	root := &cobra.Command{
		Use:   "nullapt",
		Short: "The private-first package manager for AI skills",
		Long: `nullapt — Zero-Knowledge AI Skill Manager

Install, verify, and run cryptographically-signed AI tools (skills) that
execute in a sandboxed WASM environment. No data leaves your machine unless
a skill's manifest explicitly declares network permissions.

Get started:
  nullapt get web-search        Install a skill from the registry
  nullapt list                  List installed skills
  nullapt verify SKILL.json     Verify a manifest's signature
  nullapt publish SKILL.json    Publish a skill to the registry
`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		commands.NewGetCmd(),
		commands.NewListCmd(),
		commands.NewRemoveCmd(),
		commands.NewVerifyCmd(),
		commands.NewPublishCmd(),
		commands.NewLoginCmd(),
		commands.NewLogoutCmd(),
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
