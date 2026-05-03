package commands

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/nullapt/nullapt/internal/manifest"
	"github.com/nullapt/nullapt/internal/store"
	"github.com/spf13/cobra"
)

func NewDescribeCmd() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:     "describe <skill>",
		Aliases: []string{"info", "show"},
		Short:   "Show an installed skill's tools, schemas, and permissions",
		Long: `Print the manifest of an installed skill in a form that's useful for
calling 'nullapt run' or for plugging the skill into an MCP host.

Includes each tool's input schema and the skill's declared sandbox
permissions, so you can shape a payload without opening SKILL.json.`,
		Example: `  nullapt describe nullapt/prompt-optimizer
  nullapt describe my-skill --json | jq .tools`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			skill, err := s.Get(args[0])
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(skill)
			}
			return printDescribe(out, skill)
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Emit the full manifest as JSON")
	return cmd
}

func printDescribe(out interface {
	Write(p []byte) (n int, err error)
}, skill *manifest.Skill) error {
	w := func(format string, a ...any) {
		fmt.Fprintf(out, format, a...)
	}

	w("%s@%s\n", skill.Name, skill.Version)
	if skill.Description != "" {
		w("  %s\n", skill.Description)
	}
	w("  author: %s", skill.Author)
	if skill.License != "" {
		w("    license: %s", skill.License)
	}
	w("\n")
	if skill.Homepage != "" {
		w("  homepage: %s\n", skill.Homepage)
	}

	w("\npermissions\n")
	if skill.Permissions.Network.Allowed {
		domains := skill.Permissions.Network.Domains
		if len(domains) == 0 {
			w("  network: allowed (no domains declared)\n")
		} else {
			w("  network: %s\n", strings.Join(domains, ", "))
		}
	} else {
		w("  network: none (offline-only)\n")
	}
	if r := skill.Permissions.Filesystem.Read; len(r) > 0 {
		w("  fs read:  %s\n", strings.Join(r, ", "))
	}
	if wp := skill.Permissions.Filesystem.Write; len(wp) > 0 {
		w("  fs write: %s\n", strings.Join(wp, ", "))
	}
	if env := skill.Permissions.Env; len(env) > 0 {
		w("  env:      %s\n", strings.Join(env, ", "))
	}

	w("\ntools\n")
	if len(skill.Interface.Tools) == 0 {
		w("  (none declared)\n")
		return nil
	}
	for _, t := range skill.Interface.Tools {
		w("  %s\n", t.Name)
		if t.Description != "" {
			w("    %s\n", t.Description)
		}
		if len(t.InputSchema) > 0 {
			w("    input:\n")
			printSchemaProps(out, t.InputSchema, "      ")
		}
		w("    example:\n")
		w("      nullapt run %s %s --input '%s'\n", skill.Name, t.Name, exampleInput(t.InputSchema))
	}
	return nil
}

// printSchemaProps formats a JSON Schema 'object' as one line per property:
// "name (type, required) — description". Anything more elaborate, the user
// can see with --json.
func printSchemaProps(out interface {
	Write(p []byte) (n int, err error)
}, schema map[string]any, indent string) {
	props, _ := schema["properties"].(map[string]any)
	required := map[string]bool{}
	if reqList, ok := schema["required"].([]any); ok {
		for _, r := range reqList {
			if s, ok := r.(string); ok {
				required[s] = true
			}
		}
	}
	if len(props) == 0 {
		fmt.Fprintf(out, "%s(no properties)\n", indent)
		return
	}
	names := make([]string, 0, len(props))
	for n := range props {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		p, _ := props[name].(map[string]any)
		typ := "any"
		if t, ok := p["type"].(string); ok {
			typ = t
		}
		marker := ""
		if required[name] {
			marker = ", required"
		}
		desc, _ := p["description"].(string)
		if desc == "" {
			fmt.Fprintf(out, "%s%s (%s%s)\n", indent, name, typ, marker)
		} else {
			fmt.Fprintf(out, "%s%s (%s%s) — %s\n", indent, name, typ, marker, desc)
		}
	}
}

// exampleInput builds a minimal JSON object containing only the required
// properties of the schema, with placeholder values per type. It is best-effort
// — the goal is a copy-pasteable starting point, not a valid call.
func exampleInput(schema map[string]any) string {
	if schema == nil {
		return "{}"
	}
	props, _ := schema["properties"].(map[string]any)
	required, _ := schema["required"].([]any)
	if len(required) == 0 || len(props) == 0 {
		return "{}"
	}
	example := map[string]any{}
	for _, r := range required {
		name, ok := r.(string)
		if !ok {
			continue
		}
		p, _ := props[name].(map[string]any)
		example[name] = placeholderForType(p)
	}
	data, err := json.Marshal(example)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func placeholderForType(p map[string]any) any {
	if p == nil {
		return ""
	}
	switch p["type"] {
	case "string":
		return "..."
	case "integer", "number":
		return 0
	case "boolean":
		return false
	case "array":
		return []any{}
	case "object":
		return map[string]any{}
	}
	return ""
}
