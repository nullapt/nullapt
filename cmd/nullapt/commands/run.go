package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/nullapt/nullapt/internal/runtime"
	"github.com/nullapt/nullapt/internal/store"
	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	var inputJSON string
	var inputFile string
	var timeoutSec int

	cmd := &cobra.Command{
		Use:   "run <skill> <tool>",
		Short: "Invoke a tool exposed by an installed skill",
		Long: `Run a single tool from an installed skill inside its declared WASM sandbox.

Input is read from --input, --input-file, or stdin (in that order). If none
is supplied, an empty object '{}' is sent. The tool's raw JSON output is
written to stdout; diagnostics go to stderr.`,
		Example: `  nullapt run nullapt/web-search web_search --input '{"query":"wasi"}'
  echo '{"query":"wasi"}' | nullapt run nullapt/web-search web_search
  nullapt run my-skill my_tool --input-file ./payload.json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName, toolName := args[0], args[1]

			input, err := readToolInput(inputJSON, inputFile, cmd.InOrStdin())
			if err != nil {
				return err
			}

			s, err := store.New()
			if err != nil {
				return err
			}
			skill, err := s.Get(skillName)
			if err != nil {
				return err
			}
			wasmPath, err := s.WASMPath(skillName)
			if err != nil {
				return err
			}

			ctx := context.Background()
			if timeoutSec > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
				defer cancel()
			}

			runner := runtime.New(skill, wasmPath)
			result, err := runner.Call(ctx, runtime.CallRequest{Tool: toolName, Input: input})
			if err != nil {
				return err
			}
			if _, err := os.Stdout.Write(result.Output); err != nil {
				return err
			}
			if len(result.Output) > 0 && result.Output[len(result.Output)-1] != '\n' {
				fmt.Fprintln(os.Stdout)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&inputJSON, "input", "", "JSON-encoded tool input")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "Path to a file containing JSON tool input")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 30, "Per-call timeout in seconds (0 disables)")
	return cmd
}

// readToolInput resolves the JSON payload for a run/MCP tool call, in priority
// order: --input flag, --input-file, stdin (if not a TTY), else "{}".
func readToolInput(inline, file string, stdin io.Reader) ([]byte, error) {
	switch {
	case inline != "" && file != "":
		return nil, fmt.Errorf("--input and --input-file are mutually exclusive")
	case inline != "":
		if !json.Valid([]byte(inline)) {
			return nil, fmt.Errorf("--input is not valid JSON")
		}
		return []byte(inline), nil
	case file != "":
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("reading --input-file: %w", err)
		}
		if !json.Valid(data) {
			return nil, fmt.Errorf("%s does not contain valid JSON", file)
		}
		return data, nil
	}

	// Try stdin only when it's actually piped — avoid hanging on an interactive TTY.
	if f, ok := stdin.(*os.File); ok {
		info, err := f.Stat()
		if err == nil && (info.Mode()&os.ModeCharDevice) != 0 {
			return []byte("{}"), nil
		}
	}
	data, err := io.ReadAll(stdin)
	if err != nil {
		return nil, fmt.Errorf("reading stdin: %w", err)
	}
	if len(data) == 0 {
		return []byte("{}"), nil
	}
	if !json.Valid(data) {
		return nil, fmt.Errorf("stdin payload is not valid JSON")
	}
	return data, nil
}
