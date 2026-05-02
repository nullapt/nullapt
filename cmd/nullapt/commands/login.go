package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const tokenFileName = ".nullapt_token"

func tokenPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, tokenFileName), nil
}

// loadToken reads the stored API token from ~/.nullapt_token.
func loadToken() (string, error) {
	p, err := tokenPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func NewLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the NullApt registry",
		Long: `Authenticate your CLI with the NullApt registry by providing an API token.

Generate a token at https://nullapt.dev/settings/tokens`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print("Enter your NullApt API token: ")
			var token string
			if _, err := fmt.Scan(&token); err != nil {
				return fmt.Errorf("reading token: %w", err)
			}
			token = strings.TrimSpace(token)
			if token == "" {
				return fmt.Errorf("token cannot be empty")
			}

			p, err := tokenPath()
			if err != nil {
				return err
			}
			if err := os.WriteFile(p, []byte(token+"\n"), 0o600); err != nil {
				return fmt.Errorf("saving token: %w", err)
			}

			fmt.Println("Logged in. Token saved to", p)
			return nil
		},
	}
}

func NewLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored registry credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := tokenPath()
			if err != nil {
				return err
			}
			if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
				return err
			}
			fmt.Println("Logged out.")
			return nil
		},
	}
}
