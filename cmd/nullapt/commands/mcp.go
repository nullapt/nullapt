package commands

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nullapt/nullapt/internal/manifest"
	"github.com/nullapt/nullapt/internal/runtime"
	"github.com/nullapt/nullapt/internal/store"
	"github.com/spf13/cobra"
)

// mcpProtocolVersion is the MCP version this server speaks. Clients that
// negotiate a different version will receive this in the initialize response;
// most hosts (Claude Desktop, LM Studio, mcp-cli) accept any recent date.
const mcpProtocolVersion = "2024-11-05"

// mcpToolSep separates skill name from tool name in MCP-exposed tool ids.
// Skill names match `[a-z0-9_\-]+(/[a-z0-9_\-]+)?` so they never contain "__",
// which makes the boundary unambiguous in either direction.
const mcpToolSep = "__"

func NewMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Expose installed skills to any MCP-compliant client over stdio",
		Long: `Run a Model Context Protocol server that bridges installed nullapt skills
to any MCP host (Claude Desktop, LM Studio, mcp-cli, ...).

Tools are exposed as <skill>__<tool>, with '/' in skill names replaced by '_'
so the resulting id matches MCP's allowed character set. Each call rescans the
local skill store, so newly installed skills appear without a restart.

Configure your MCP host to launch:  nullapt mcp`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			srv := &mcpServer{store: s, in: cmd.InOrStdin(), out: cmd.OutOrStdout()}
			return srv.serve(cmd.Context())
		},
	}
}

type mcpServer struct {
	store *store.Store
	in    io.Reader
	out   io.Writer
	mu    sync.Mutex // guards writes to out
}

// jsonrpcMessage covers requests, responses, and notifications. A request has
// id+method, a response has id+(result|error), a notification has method only.
type jsonrpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeInternalError  = -32603
)

func (s *mcpServer) serve(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	// MCP framing is newline-delimited JSON. bufio.Scanner has a 64 KiB cap,
	// which is too small for tool payloads, so use Reader.ReadBytes('\n').
	r := bufio.NewReader(s.in)
	for {
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			trimmed := strings.TrimRight(string(line), "\r\n")
			if trimmed != "" {
				s.handleLine(ctx, trimmed)
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (s *mcpServer) handleLine(ctx context.Context, line string) {
	var msg jsonrpcMessage
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		s.writeError(nil, codeParseError, "invalid JSON: "+err.Error())
		return
	}

	// Notifications carry no id and expect no reply.
	isNotification := len(msg.ID) == 0

	switch msg.Method {
	case "initialize":
		s.writeResult(msg.ID, map[string]any{
			"protocolVersion": mcpProtocolVersion,
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "nullapt",
				"version": Version,
			},
		})
	case "notifications/initialized", "initialized":
		// no-op
	case "ping":
		s.writeResult(msg.ID, map[string]any{})
	case "tools/list":
		s.handleToolsList(msg.ID)
	case "tools/call":
		s.handleToolsCall(ctx, msg.ID, msg.Params)
	case "shutdown":
		if !isNotification {
			s.writeResult(msg.ID, nil)
		}
	default:
		if !isNotification {
			s.writeError(msg.ID, codeMethodNotFound, "method not found: "+msg.Method)
		}
	}
}

type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema"`
}

func (s *mcpServer) handleToolsList(id json.RawMessage) {
	skills, err := s.store.List()
	if err != nil {
		s.writeError(id, codeInternalError, "listing skills: "+err.Error())
		return
	}
	tools := make([]mcpTool, 0)
	for _, sk := range skills {
		for _, t := range sk.Interface.Tools {
			schema := t.InputSchema
			if schema == nil {
				schema = map[string]any{"type": "object"}
			}
			tools = append(tools, mcpTool{
				Name:        mcpToolName(sk, t.Name),
				Description: toolDescription(sk, t),
				InputSchema: schema,
			})
		}
	}
	s.writeResult(id, map[string]any{"tools": tools})
}

type toolsCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (s *mcpServer) handleToolsCall(ctx context.Context, id json.RawMessage, params json.RawMessage) {
	var p toolsCallParams
	if err := json.Unmarshal(params, &p); err != nil {
		s.writeError(id, codeInvalidParams, "invalid tools/call params: "+err.Error())
		return
	}
	if p.Name == "" {
		s.writeError(id, codeInvalidParams, "tools/call requires 'name'")
		return
	}

	skill, toolName, err := s.resolveTool(p.Name)
	if err != nil {
		s.writeError(id, codeInvalidParams, err.Error())
		return
	}
	wasmPath, err := s.store.WASMPath(skill.Name)
	if err != nil {
		s.writeError(id, codeInternalError, err.Error())
		return
	}

	input := []byte(p.Arguments)
	if len(input) == 0 || string(input) == "null" {
		input = []byte("{}")
	} else if !json.Valid(input) {
		s.writeError(id, codeInvalidParams, "tools/call 'arguments' is not valid JSON")
		return
	}

	callCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	runner := runtime.New(skill, wasmPath)
	result, err := runner.Call(callCtx, runtime.CallRequest{Tool: toolName, Input: input})
	if err != nil {
		// MCP returns a successful response with isError=true so the host can
		// surface tool failures to the model rather than treating them as a
		// transport-level error.
		s.writeResult(id, map[string]any{
			"isError": true,
			"content": []map[string]any{
				{"type": "text", "text": err.Error()},
			},
		})
		return
	}

	s.writeResult(id, map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": string(result.Output)},
		},
	})
}

// resolveTool walks the installed skill list to find the (skill, tool) pair
// whose mcpToolName matches the given id. We don't parse the id directly —
// tool names in manifests can contain "__", so a name→pair lookup is the
// only unambiguous resolution.
func (s *mcpServer) resolveTool(mcpName string) (*manifest.Skill, string, error) {
	skills, err := s.store.List()
	if err != nil {
		return nil, "", fmt.Errorf("listing skills: %w", err)
	}
	for _, sk := range skills {
		for _, t := range sk.Interface.Tools {
			if mcpToolName(sk, t.Name) == mcpName {
				return sk, t.Name, nil
			}
		}
	}
	return nil, "", fmt.Errorf("tool %q is not provided by any installed skill", mcpName)
}

func mcpToolName(sk *manifest.Skill, tool string) string {
	return strings.ReplaceAll(sk.Name, "/", "_") + mcpToolSep + tool
}

func toolDescription(sk *manifest.Skill, t manifest.Tool) string {
	if t.Description == "" {
		return fmt.Sprintf("(%s@%s)", sk.Name, sk.Version)
	}
	return fmt.Sprintf("%s (%s@%s)", t.Description, sk.Name, sk.Version)
}

func (s *mcpServer) writeResult(id json.RawMessage, result any) {
	s.writeMessage(jsonrpcMessage{JSONRPC: "2.0", ID: id, Result: result})
}

func (s *mcpServer) writeError(id json.RawMessage, code int, message string) {
	s.writeMessage(jsonrpcMessage{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &jsonrpcError{Code: code, Message: message},
	})
}

func (s *mcpServer) writeMessage(msg jsonrpcMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nullapt mcp: marshal error: %v\n", err)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	data = append(data, '\n')
	if _, err := s.out.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "nullapt mcp: write error: %v\n", err)
	}
}
