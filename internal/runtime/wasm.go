package runtime

import (
	"context"
	"fmt"

	extism "github.com/extism/go-sdk"
	"github.com/nullapt/nullapt/internal/manifest"
)

// CallRequest is a single tool invocation sent to a skill's WASM module.
type CallRequest struct {
	Tool  string
	Input []byte // JSON-encoded input matching the tool's input_schema
}

// CallResult is the response from a WASM tool invocation.
type CallResult struct {
	Output []byte // JSON-encoded output
}

// Runner executes skill tools inside an isolated Extism/WASM sandbox.
type Runner struct {
	wasmPath string
	skill    *manifest.Skill
}

// New creates a Runner for the given skill.
func New(skill *manifest.Skill, wasmPath string) *Runner {
	return &Runner{skill: skill, wasmPath: wasmPath}
}

// Call invokes a named tool with JSON-encoded input and returns its output.
// The permission boundaries declared in SKILL.json (network domains, fs paths)
// are enforced at the wazero runtime level — the sandbox physically cannot
// open connections or paths outside the declared allow-lists.
func (r *Runner) Call(ctx context.Context, req CallRequest) (*CallResult, error) {
	if err := r.validateTool(req.Tool); err != nil {
		return nil, err
	}

	em := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmFile{Path: r.wasmPath},
		},
		// Hard Stop: only the domains listed in SKILL.json can be dialed.
		AllowedHosts: r.skill.Permissions.Network.Domains,
		// Filesystem access limited to declared read/write paths.
		AllowedPaths: buildFSAllowList(r.skill),
	}

	plugin, err := extism.NewPlugin(ctx, em, extism.PluginConfig{
		EnableWasi: true,
	}, []extism.HostFunction{})
	if err != nil {
		return nil, fmt.Errorf("loading wasm plugin %q: %w", r.skill.Name, err)
	}
	defer plugin.Close()

	_, output, err := plugin.Call(req.Tool, req.Input)
	if err != nil {
		return nil, fmt.Errorf("calling tool %q in %q: %w", req.Tool, r.skill.Name, err)
	}

	return &CallResult{Output: output}, nil
}

func (r *Runner) validateTool(name string) error {
	for _, t := range r.skill.Interface.Tools {
		if t.Name == name {
			return nil
		}
	}
	return fmt.Errorf("tool %q is not declared in skill %q manifest", name, r.skill.Name)
}

func buildFSAllowList(skill *manifest.Skill) map[string]string {
	paths := make(map[string]string)
	for _, p := range skill.Permissions.Filesystem.Read {
		paths[p] = p
	}
	for _, p := range skill.Permissions.Filesystem.Write {
		paths[p] = p
	}
	return paths
}
