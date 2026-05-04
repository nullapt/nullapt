"use client";

import Image from "next/image";
import { useState } from "react";

type Logo = { src: string; alt: string; w: number; h: number };

type Integration = {
  id: string;
  name: string;
  logo: Logo;
  /** Preferred / fastest setup. */
  primary: { kind: "cli"; command: string } | { kind: "json"; path: string; snippet: string };
  /** Where the config ends up after the primary action runs. */
  configPath?: string;
  notes: string;
};

const LOGOS: Record<string, Logo> = {
  claude: { src: "/logos/claude.svg", alt: "Claude", w: 32, h: 32 },
  cursor: { src: "/logos/cursor.svg", alt: "Cursor", w: 32, h: 32 },
  lmstudio: { src: "/logos/lmstudio.png", alt: "LM Studio", w: 32, h: 32 },
  anythingllm: { src: "/logos/anythingllm.svg", alt: "AnythingLLM", w: 32, h: 32 },
  zed: { src: "/logos/zed.png", alt: "Zed", w: 32, h: 32 },
  vscode: { src: "/logos/vscode.svg", alt: "VS Code", w: 32, h: 32 },
  ollama: { src: "/logos/ollama.svg", alt: "Ollama", w: 32, h: 32 },
};

const SHARED_JSON = `{
  "mcpServers": {
    "nullapt": {
      "command": "nullapt",
      "args": ["mcp"]
    }
  }
}`;

const INTEGRATIONS: Integration[] = [
  {
    id: "claude-code",
    name: "Claude Code",
    logo: LOGOS.claude,
    primary: { kind: "cli", command: "claude mcp add nullapt --scope user -- nullapt mcp" },
    configPath: "~/.claude.json (user scope)  •  .mcp.json (project scope)",
    notes:
      "User scope makes installed skills available across every project. For a team-shared setup, use --scope project — Claude Code writes to .mcp.json at the project root, which you can commit.",
  },
  {
    id: "claude-desktop",
    name: "Claude Desktop",
    logo: LOGOS.claude,
    primary: {
      kind: "json",
      path: "~/Library/Application Support/Claude/claude_desktop_config.json (macOS)\n%APPDATA%\\Claude\\claude_desktop_config.json (Windows)",
      snippet: SHARED_JSON,
    },
    notes:
      "Settings → Developer → Edit Config opens the file. Save, fully quit, and relaunch Claude Desktop — the MCP indicator appears in the message composer when nullapt is loaded.",
  },
  {
    id: "cursor",
    name: "Cursor",
    logo: LOGOS.cursor,
    primary: {
      kind: "json",
      path: "~/.cursor/mcp.json (global)  •  .cursor/mcp.json (per project)",
      snippet: SHARED_JSON,
    },
    notes:
      "Global config makes nullapt available in every Cursor workspace. Settings → MCP shows a green dot once the server is registered and tools are discovered.",
  },
  {
    id: "lm-studio",
    name: "LM Studio",
    logo: LOGOS.lmstudio,
    primary: {
      kind: "json",
      path: "Settings → Program → Install → Edit mcp.json",
      snippet: SHARED_JSON,
    },
    notes:
      "LM Studio follows the same mcp.json shape as Cursor. Toggle the nullapt server on in the chat-window tools panel to make installed skills callable in conversations.",
  },
  {
    id: "anythingllm",
    name: "AnythingLLM",
    logo: LOGOS.anythingllm,
    primary: {
      kind: "json",
      path: "anythingllm_mcp_servers.json (in your AnythingLLM storage plugins directory — auto-generated)",
      snippet: SHARED_JSON,
    },
    notes:
      "Settings → Agent Skills → MCP Servers → Add. Each tool becomes available to the @agent persona; enable the ones you want exposed.",
  },
  {
    id: "zed",
    name: "Zed",
    logo: LOGOS.zed,
    primary: {
      kind: "json",
      path: "~/.config/zed/settings.json — under the context_servers key",
      snippet: `{
  "context_servers": {
    "nullapt": {
      "source": "custom",
      "command": "nullapt",
      "args": ["mcp"]
    }
  }
}`,
    },
    notes:
      "Zed calls them context servers but the protocol is MCP. Reload the assistant after saving.",
  },
  {
    id: "vscode",
    name: "VS Code",
    logo: LOGOS.vscode,
    primary: {
      kind: "json",
      path: ".vscode/mcp.json (per workspace)  •  Settings → MCP: User Configuration (per user)",
      snippet: `{
  "servers": {
    "nullapt": {
      "type": "stdio",
      "command": "nullapt",
      "args": ["mcp"]
    }
  }
}`,
    },
    notes:
      "GitHub Copilot's agent mode and the official MCP extension both read this file. The Continue extension uses its own ~/.continue/config.yaml — see continue.dev/docs.",
  },
];

const TAB_BTN_BASE =
  "flex items-center gap-2 px-3 py-2 text-xs whitespace-nowrap transition-colors border-b-2";

function Snippet({ children }: { children: string }) {
  return (
    <pre
      className="rounded p-4 text-xs font-mono overflow-x-auto"
      style={{ background: "var(--surface)", border: "1px solid var(--border)" }}
    >
      <code>{children}</code>
    </pre>
  );
}

export default function Integrations() {
  const [activeId, setActiveId] = useState(INTEGRATIONS[0].id);
  const active = INTEGRATIONS.find((i) => i.id === activeId) ?? INTEGRATIONS[0];

  return (
    <div>
      <p style={{ color: "var(--muted)" }} className="text-sm mb-4">
        Once <code style={{ color: "var(--accent)" }}>nullapt mcp</code> is registered as an MCP
        server, every skill installed via <code style={{ color: "var(--accent)" }}>nullapt get</code>{" "}
        becomes available to the host — no per-skill setup. Pick your client:
      </p>

      <div
        className="flex overflow-x-auto"
        style={{ borderBottom: "1px solid var(--border)" }}
        role="tablist"
        aria-label="MCP client integrations"
      >
        {INTEGRATIONS.map((i) => {
          const isActive = i.id === activeId;
          return (
            <button
              key={i.id}
              role="tab"
              aria-selected={isActive}
              onClick={() => setActiveId(i.id)}
              className={TAB_BTN_BASE}
              style={{
                color: isActive ? "var(--foreground)" : "var(--muted)",
                borderBottomColor: isActive ? "var(--accent)" : "transparent",
              }}
            >
              <Image
                src={i.logo.src}
                alt=""
                width={i.logo.w}
                height={i.logo.h}
                style={{ width: 16, height: 16, objectFit: "contain" }}
                unoptimized
              />
              <span>{i.name}</span>
            </button>
          );
        })}
      </div>

      <div className="pt-6">
        <div className="flex items-center gap-3 mb-4">
          <Image
            src={active.logo.src}
            alt={active.logo.alt}
            width={active.logo.w}
            height={active.logo.h}
            style={{ width: 24, height: 24, objectFit: "contain" }}
            unoptimized
          />
          <h3 className="text-base font-semibold" style={{ color: "var(--foreground)" }}>
            {active.name}
          </h3>
        </div>

        {active.primary.kind === "cli" ? (
          <>
            <p style={{ color: "var(--muted)" }} className="text-sm mb-2">
              Run this once and the server is registered:
            </p>
            <Snippet>{`$ ${active.primary.command}`}</Snippet>
            {active.configPath && (
              <p style={{ color: "var(--muted)" }} className="text-xs mt-2 mb-4">
                Writes to: <code style={{ color: "var(--accent)" }}>{active.configPath}</code>
              </p>
            )}
          </>
        ) : (
          <>
            <p style={{ color: "var(--muted)" }} className="text-xs mb-2">
              Edit:{" "}
              <code style={{ color: "var(--accent)" }} className="whitespace-pre-wrap">
                {active.primary.path}
              </code>
            </p>
            <Snippet>{active.primary.snippet}</Snippet>
          </>
        )}

        <p style={{ color: "var(--muted)" }} className="text-xs mt-4">
          {active.notes}
        </p>

        <div
          className="mt-6 rounded p-4 text-xs"
          style={{ background: "var(--surface)", border: "1px solid var(--border)", color: "var(--muted)" }}
        >
          <strong style={{ color: "var(--foreground)" }}>Verify it loaded:</strong> ask the host to
          list its tools — every installed skill appears as{" "}
          <code style={{ color: "var(--accent)" }}>{`<scope>_<name>__<tool>`}</code> (e.g.{" "}
          <code style={{ color: "var(--accent)" }}>nullapt_web-search__web_search</code>). The
          server rescans <code style={{ color: "var(--accent)" }}>~/.nullapt/skills/</code> on
          every request, so newly installed skills appear without restarting the host.
        </div>
      </div>
    </div>
  );
}
