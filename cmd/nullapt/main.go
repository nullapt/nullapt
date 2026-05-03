package main

import (
	"fmt"
	"os"

	"github.com/nullapt/nullapt/cmd/nullapt/commands"
	"github.com/nullapt/nullapt/internal/registry"
	"github.com/spf13/cobra"
)

var version = "0.1.0-dev"

func main() {
	commands.Version = version
	registry.Version = version

	// Kick off the update check in background before running the command.
	updateCh := commands.CheckForUpdate(version)

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

	err := root.Execute()

	// Print update notice after the command, if a newer version was found.
	if latest, ok := <-updateCh; ok && latest != "" {
		fmt.Fprintf(os.Stderr, "\n   update available  %s → %s\n", version, latest)
		fmt.Fprintf(os.Stderr, "   brew upgrade nullapt  or  curl -fsSL https://nullapt.dev/install.sh | sh\n\n")
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
