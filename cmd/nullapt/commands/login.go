package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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

func registryURL() string {
	if u := os.Getenv("NULLAPT_REGISTRY_URL"); u != "" {
		return u
	}
	return "https://registry.nullapt.dev"
}

func NewLoginCmd() *cobra.Command {
	var browserless bool
	var pasteToken bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the NullApt registry",
		Long: `Authenticate your CLI with the NullApt registry.

By default, a browser window opens and you approve access with one click.
Use --browserless to get a URL you can open on any device.
Use --token to paste an existing API token directly.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if pasteToken {
				return loginWithPaste()
			}
			return loginWithBrowser(browserless)
		},
	}

	cmd.Flags().BoolVar(&browserless, "browserless", false, "print the auth URL instead of opening a browser")
	cmd.Flags().BoolVar(&pasteToken, "token", false, "paste an existing API token instead of using the browser flow")
	return cmd
}

// loginWithBrowser uses the device-flow: init a request, open/print the URL, poll until approved.
func loginWithBrowser(browserless bool) error {
	base := registryURL()

	// 1. Init the auth request
	req, err := registryRequest(http.MethodPost, base+"/v1/auth/cli/init", nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := registryHTTP.Do(req)
	if err != nil {
		return fmt.Errorf("connecting to registry: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("registry returned HTTP %d at /v1/auth/cli/init", resp.StatusCode)
	}

	var init struct {
		ID              string `json:"id"`
		VerificationURL string `json:"verification_url"`
		ExpiresIn       int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&init); err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	// 2. Open or print the URL
	fmt.Println()
	if browserless {
		fmt.Println("  Open this URL in your browser to authorize the CLI:")
		fmt.Println()
		fmt.Println(" ", init.VerificationURL)
		fmt.Println()
	} else {
		fmt.Println("  Opening your browser to authorize the CLI…")
		fmt.Println("  If it doesn't open, visit:")
		fmt.Println()
		fmt.Println(" ", init.VerificationURL)
		fmt.Println()
		_ = openBrowser(init.VerificationURL)
	}

	fmt.Print("  Waiting for approval")

	// 3. Poll until approved or expired
	deadline := time.Now().Add(time.Duration(init.ExpiresIn) * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)
		fmt.Print(".")

		token, done, err := pollCLIAuth(base, init.ID)
		if err != nil {
			fmt.Println()
			return err
		}
		if done && token != "" {
			fmt.Println()
			return saveToken(token)
		}
		if done {
			fmt.Println()
			return fmt.Errorf("auth request expired — run 'nullapt login' again")
		}
	}

	fmt.Println()
	return fmt.Errorf("timed out waiting for approval")
}

func pollCLIAuth(base, id string) (token string, done bool, err error) {
	req, err := registryRequest(http.MethodGet, base+"/v1/auth/cli/poll?id="+id, nil)
	if err != nil {
		return "", false, fmt.Errorf("building request: %w", err)
	}
	resp, err := registryHTTP.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("polling: %w", err)
	}
	defer resp.Body.Close()

	var body struct {
		Status string `json:"status"`
		Token  string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", false, nil
	}

	switch body.Status {
	case "approved":
		return body.Token, true, nil
	case "expired":
		return "", true, nil
	default:
		return "", false, nil
	}
}

// loginWithPaste is the original manual flow for users who already have a token.
func loginWithPaste() error {
	fmt.Println("Generate a token at: https://nullapt.dev/settings/tokens")
	fmt.Print("Enter your NullApt API token: ")
	var token string
	if _, err := fmt.Scan(&token); err != nil {
		return fmt.Errorf("reading token: %w", err)
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	if !strings.HasPrefix(token, "nlpt_") {
		fmt.Println("Warning: token does not start with 'nlpt_'. Make sure you copied the full token.")
	}
	return saveToken(token)
}

func saveToken(token string) error {
	p, err := tokenPath()
	if err != nil {
		return err
	}
	if err := os.WriteFile(p, []byte(token+"\n"), 0o600); err != nil {
		return fmt.Errorf("saving token: %w", err)
	}
	fmt.Println("  Logged in. Token saved to", p)
	return nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	return exec.Command(cmd, args...).Start()
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
