package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// mcpHost is a known MCP host whose config we know how to write.
type mcpHost struct {
	id          string                 // CLI flag value
	displayName string                 // human-readable
	configPath  func() (string, error) // resolves config path for this OS
	restartHint string                 // what to tell the user after a successful install
}

var knownMCPHosts = []mcpHost{
	{
		id:          "claude-desktop",
		displayName: "Claude Desktop",
		configPath:  claudeDesktopConfigPath,
		restartHint: "Quit Claude Desktop fully (Cmd+Q on macOS) and reopen it.",
	},
	{
		id:          "cursor",
		displayName: "Cursor",
		configPath:  cursorConfigPath,
		restartHint: "Reload the Cursor window or restart Cursor.",
	},
}

func newMCPInstallCmd() *cobra.Command {
	var clientFlag string
	var configPathFlag string
	var serverName string
	var force bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Wire `nullapt mcp` into an MCP host's config in one step",
		Long: `Add an entry for nullapt to a known MCP host's config file.

Without flags, every supported host that is installed on this machine is
configured. Use --client to target a specific host, or --config to write
to an arbitrary mcpServers-style JSON file (useful for clients we don't
yet recognise).

The skill list is read live on every tools/list, so future 'nullapt get'
calls do not require re-running this command — only the host's restart.`,
		Example: `  nullapt mcp install
  nullapt mcp install --client claude-desktop
  nullapt mcp install --config ~/.cursor/mcp.json
  nullapt mcp install --dry-run`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			binary, err := nullaptBinaryPath()
			if err != nil {
				return err
			}

			targets, err := resolveInstallTargets(clientFlag, configPathFlag)
			if err != nil {
				return err
			}
			if len(targets) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No supported MCP hosts found on this machine.")
				fmt.Fprintln(cmd.OutOrStdout(), "Run with --config <path> to install into an arbitrary config file.")
				return nil
			}

			out := cmd.OutOrStdout()
			anyChanged := false
			for _, t := range targets {
				changed, err := installInto(t, binary, serverName, force, dryRun)
				if err != nil {
					fmt.Fprintf(out, "  %s  %s — %v\n", failMark, t.label, err)
					continue
				}
				if changed {
					anyChanged = true
					if dryRun {
						fmt.Fprintf(out, "  %s  %s — would update %s\n", okMark, t.label, t.path)
					} else {
						fmt.Fprintf(out, "  %s  %s — updated %s\n", okMark, t.label, t.path)
					}
				} else {
					fmt.Fprintf(out, "  %s  %s — already configured (%s)\n", skipMark, t.label, t.path)
				}
			}

			if anyChanged && !dryRun {
				fmt.Fprintln(out)
				for _, t := range targets {
					if t.restartHint != "" {
						fmt.Fprintf(out, "  %s\n", t.restartHint)
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&clientFlag, "client", "", "Target a specific host: claude-desktop, cursor")
	cmd.Flags().StringVar(&configPathFlag, "config", "", "Write to an arbitrary mcpServers-style JSON file")
	cmd.Flags().StringVar(&serverName, "name", "nullapt", "Key to use under mcpServers")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite an existing entry without prompting")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would change without writing")
	return cmd
}

const (
	okMark   = "✓"
	skipMark = "·"
	failMark = "✗"
)

// installTarget pairs a resolved config path with display metadata.
type installTarget struct {
	label       string
	path        string
	restartHint string
}

func resolveInstallTargets(client, configPath string) ([]installTarget, error) {
	if client != "" && configPath != "" {
		return nil, fmt.Errorf("--client and --config are mutually exclusive")
	}

	if configPath != "" {
		expanded, err := expandHome(configPath)
		if err != nil {
			return nil, err
		}
		return []installTarget{{
			label:       configPath,
			path:        expanded,
			restartHint: "Restart your MCP client.",
		}}, nil
	}

	if client != "" {
		for _, h := range knownMCPHosts {
			if h.id == client {
				p, err := h.configPath()
				if err != nil {
					return nil, fmt.Errorf("%s: %w", h.displayName, err)
				}
				return []installTarget{{label: h.displayName, path: p, restartHint: h.restartHint}}, nil
			}
		}
		ids := make([]string, len(knownMCPHosts))
		for i, h := range knownMCPHosts {
			ids[i] = h.id
		}
		sort.Strings(ids)
		return nil, fmt.Errorf("unknown --client %q (known: %s)", client, strings.Join(ids, ", "))
	}

	// Auto-detect: include any known host whose parent config dir exists.
	var targets []installTarget
	for _, h := range knownMCPHosts {
		p, err := h.configPath()
		if err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Dir(p)); err != nil {
			continue
		}
		targets = append(targets, installTarget{label: h.displayName, path: p, restartHint: h.restartHint})
	}
	return targets, nil
}

// installInto reads the config at path (creating an empty one if missing),
// adds an mcpServers.<name> entry pointing at the nullapt binary, and writes
// it back. Returns (changed, err): changed=false means the entry already
// matched and the file was not touched.
func installInto(t installTarget, binary, name string, force, dryRun bool) (bool, error) {
	cfg, err := readJSONObject(t.path)
	if err != nil {
		return false, err
	}

	servers, _ := cfg["mcpServers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}

	desired := map[string]any{
		"command": binary,
		"args":    []any{"mcp"},
	}

	if existing, ok := servers[name]; ok {
		if jsonEqual(existing, desired) {
			return false, nil
		}
		if !force {
			return false, fmt.Errorf("%q already exists with a different value (use --force to overwrite)", name)
		}
	}

	if dryRun {
		return true, nil
	}

	servers[name] = desired
	cfg["mcpServers"] = servers

	if err := os.MkdirAll(filepath.Dir(t.path), 0o750); err != nil {
		return false, fmt.Errorf("creating config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return false, fmt.Errorf("encoding config: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(t.path, data, 0o640); err != nil {
		return false, fmt.Errorf("writing config: %w", err)
	}
	return true, nil
}

// readJSONObject reads a JSON object from path. Missing or empty files yield
// an empty object — that's the natural starting point when a host hasn't
// stored any MCP config yet.
func readJSONObject(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return map[string]any{}, nil
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, fmt.Errorf("parsing %s: %w (refusing to overwrite a non-JSON-object config)", path, err)
	}
	return obj, nil
}

func jsonEqual(a, b any) bool {
	aj, err1 := json.Marshal(a)
	bj, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return string(aj) == string(bj)
}

func nullaptBinaryPath() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolving nullapt binary path: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		return p, nil
	}
	return resolved, nil
}

func expandHome(p string) (string, error) {
	if !strings.HasPrefix(p, "~") {
		return filepath.Abs(p)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if p == "~" {
		return home, nil
	}
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(home, p[2:]), nil
	}
	// "~user/..." not supported — keep parity with most shells' lenient behavior.
	return filepath.Abs(p)
}

func claudeDesktopConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil
	case "linux":
		return filepath.Join(home, ".config", "Claude", "claude_desktop_config.json"), nil
	}
	return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func cursorConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cursor", "mcp.json"), nil
}
