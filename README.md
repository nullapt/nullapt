<div align="center">

```
~/nullapt
```



**The Private-First Package Manager for AI Skills**

*Cryptographically signed. WASM sandboxed. Offline first.*

[![CI](https://github.com/nullapt/nullapt/actions/workflows/ci.yml/badge.svg)](https://github.com/nullapt/nullapt/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/nullapt/nullapt?color=4ade80)](https://github.com/nullapt/nullapt/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-4ade80.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8.svg)](https://go.dev)
[![Registry](https://img.shields.io/badge/registry-nullapt.dev-4ade80)](https://nullapt.dev)

</div>

---

NullApt is an open-source package manager for **AI skills** — WASM-compiled tools that give local LLMs (LM Studio, AnythingLLM, Ollama) new capabilities via the Model Context Protocol. Every skill is:

- **Ed25519 signed** — you verify the author before anything runs
- **WASM sandboxed** — the runtime physically cannot access resources outside the manifest
- **Offline first** — skills with no `network.allowed: true` work completely air-gapped

```bash
# Install a skill
nullapt get nullapt/web-search

# The skill is immediately available to any MCP-compliant LLM client
# on your machine — no restart required.
```

---

## Contents

- [Contents](#contents)
- [Why NullApt](#why-nullapt)
- [Installation](#installation)
  - [macOS / Linux](#macos--linux)
  - [Homebrew](#homebrew)
  - [Go](#go)
  - [Windows](#windows)
- [Quick Start](#quick-start)
- [SKILL.json Reference](#skilljson-reference)
  - [Permission fields](#permission-fields)
- [Command Reference](#command-reference)
  - [Flags](#flags)
- [Building a Skill](#building-a-skill)
  - [Rust (default)](#rust-default)
  - [Signing your manifest](#signing-your-manifest)
  - [Testing locally](#testing-locally)
- [Publishing under an org namespace](#publishing-under-an-org-namespace)
- [Architecture](#architecture)
- [Publishing](#publishing)
- [Self-Hosting the Registry](#self-hosting-the-registry)
- [Contributing](#contributing)
- [Security](#security)
- [License](#license)

---

## Why NullApt

By 2026, every local AI assistant can run community-built tools via MCP. The problem is trust: a malicious skill can silently exfiltrate files, call home, or read environment variables. Existing registries offer no guarantees.

NullApt solves this with three layers:

| Layer | What it does |
|---|---|
| **Manifest** (`SKILL.json`) | Declares exact permissions — specific domains, specific paths, specific env vars |
| **Signature** (Ed25519) | Proves the binary matches what the author signed |
| **Sandbox** (WASM-WASI) | Enforces a hard capability limit at the runtime level |

A skill that doesn't declare `network.allowed: true` **cannot make HTTP requests** — not because we ask it not to, but because the WASM host will kill the call at the syscall boundary.

---

## Installation

### macOS / Linux

```bash
curl -fsSL https://nullapt.dev/install.sh | sh
```

### Homebrew

```bash
brew install nullapt/tap/nullapt
```

### Go

```bash
go install github.com/nullapt/nullapt/cmd/nullapt@latest
```

### Windows

Download the latest `.zip` from the [releases page](https://github.com/nullapt/nullapt/releases) and extract `nullapt.exe` somewhere on your PATH.

---

## Quick Start

```bash
# Install a published skill (namespaced or personal)
nullapt get nullapt/web-search
nullapt get nullapt/prompt-optimizer

# List what's installed locally
nullapt list

# Run a tool directly inside its WASM sandbox
nullapt run nullapt/prompt-optimizer optimize --input '{"prompt":"Could you please help"}'

# Verify a local manifest's signature (before installing)
nullapt verify ./my-skill/SKILL.json

# Remove a skill
nullapt remove nullapt/web-search
```

Browse the full catalog at [nullapt.dev](https://nullapt.dev).

### Using installed skills from an MCP client

`nullapt mcp` is a stdio Model Context Protocol server that exposes every installed skill's tools to any MCP-compatible host (Claude Desktop, LM Studio, AnythingLLM, mcp-cli, …). Tool ids are `<skill>__<tool>` with `/` in skill names rewritten to `_`.

The fastest path is one command — it auto-detects every supported host on the machine and writes the entry idempotently:

```bash
nullapt mcp install              # configure every host found
nullapt mcp install --client claude-desktop
nullapt mcp install --config ~/path/to/mcp.json   # any other client
```

After install, restart the host once. New skills appear automatically — the server rescans `~/.nullapt/skills/` on every `tools/list` and `tools/call`, no further restarts needed.

If you'd rather edit the config by hand, the snippet is:

```json
{
  "mcpServers": {
    "nullapt": { "command": "nullapt", "args": ["mcp"] }
  }
}
```

To inspect a skill's tools and input schemas before calling them:

```bash
nullapt describe nullapt/prompt-optimizer
```

---

## SKILL.json Reference

Every skill ships with a `SKILL.json` manifest. This is the single source of truth for permissions, interface, and cryptographic provenance.

```json
{
  "schema_version": "1.0",
  "name": "alice/web-search",
  "version": "1.2.0",
  "description": "Search the web using DuckDuckGo Instant Answers",
  "author": "alice",
  "homepage": "https://github.com/alice/nullapt-web-search",
  "license": "MIT",

  "permissions": {
    "network": {
      "allowed": true,
      "domains": ["api.duckduckgo.com"]
    },
    "filesystem": {
      "read": [],
      "write": []
    },
    "env": []
  },

  "entry": "skill.wasm",

  "interface": {
    "tools": [
      {
        "name": "web_search",
        "description": "Search the web and return top results",
        "input_schema": {
          "type": "object",
          "properties": {
            "query": { "type": "string", "description": "The search query" }
          },
          "required": ["query"]
        }
      }
    ]
  },

  "signature": {
    "algorithm": "ed25519",
    "public_key": "<base64-encoded author public key>",
    "value": "<base64-encoded signature over the manifest payload>"
  }
}
```

### Permission fields

| Field | Type | Description |
|---|---|---|
| `network.allowed` | `bool` | Whether any outbound HTTP is permitted |
| `network.domains` | `[]string` | Explicit allowlist of domains. Wildcards are rejected. |
| `filesystem.read` | `[]string` | Absolute paths the skill may open for reading |
| `filesystem.write` | `[]string` | Absolute paths the skill may write to |
| `env` | `[]string` | Environment variable names the skill may read |

Any permission not declared is denied at the WASM host level — not by policy, but by capability absence.

---

## Command Reference

```
nullapt get <skill[@version]>      Install a skill from the registry
nullapt remove <skill>             Uninstall a skill
nullapt list                       List installed skills
nullapt run <skill> <tool>         Invoke a tool inside its WASM sandbox
nullapt describe <skill>           Show a skill's tools, schemas, and permissions
nullapt mcp                        Serve installed skills over stdio MCP
nullapt mcp install                Wire `nullapt mcp` into a known MCP host's config
nullapt verify <SKILL.json>        Verify a manifest's Ed25519 signature
nullapt keygen                     Generate an Ed25519 signing keypair
nullapt sign <SKILL.json>          Sign a manifest in-place with your private key
nullapt publish <SKILL.json>       Publish a signed skill to the registry
nullapt login                      Authenticate with the registry (GitHub OAuth)
nullapt logout                     Remove stored credentials
```

Skill names follow `<scope>/<name>` (e.g. `nullapt/web-search`) for org-published skills, or just `<name>` (e.g. `my-thing`) for personal-namespace skills.

### Flags

```
nullapt get --registry <url>       Use a custom or self-hosted registry
nullapt get --skip-verify          Skip signature check (dev only)
nullapt publish --registry <url>   Publish to a custom registry
```

---

## Building a Skill

Skills are WASM-WASI binaries that export named functions matching the tool names in `interface.tools`. Any language that compiles to `wasm32-wasip1` works — **Rust** (best DX, our default), **TinyGo**, **AssemblyScript**, **Zig**, **C/C++**.

Working starter projects live in [github.com/nullapt/examples](https://github.com/nullapt/examples).

### Rust (default)

```bash
cargo new --lib my-skill
cd my-skill
cargo add extism-pdk
rustup target add wasm32-wasip1   # one-time
```

```rust
// src/lib.rs
use extism_pdk::*;

#[plugin_fn]
pub fn my_tool(input: Json<MyInput>) -> FnResult<Json<MyOutput>> {
    // ... your logic
}
```

```bash
cargo build --target wasm32-wasip1 --release
cp target/wasm32-wasip1/release/my_skill.wasm skill.wasm
```

> Note: `wasm32-wasip1` is the modern target name. The older `wasm32-wasi` target was renamed in Rust 1.78. If older docs say `wasm32-wasi`, use `wasm32-wasip1` instead.

### Signing your manifest

```bash
# Generate an Ed25519 keypair (writes to ~/.nullapt/keys/ by default).
nullapt keygen

# Sign SKILL.json in-place with your private key.
nullapt sign ./SKILL.json
```

Both commands accept `--output` / `--key` to override paths.

### Testing locally

```bash
nullapt verify ./SKILL.json        # check the signature
nullapt get ./SKILL.json           # install from local path
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        User's Machine                        │
│                                                              │
│  ┌──────────┐    ┌─────────────────────────────────────┐    │
│  │  LM      │    │           nullapt CLI                │    │
│  │  Studio  │◄──►│  get · list · verify · publish       │    │
│  │  / any   │    └──────────────┬──────────────────────┘    │
│  │  MCP     │                   │                            │
│  │  client  │    ┌──────────────▼──────────────────────┐    │
│  └──────────┘    │     ~/.nullapt/skills/               │    │
│        ▲         │     web-search/                      │    │
│        │ MCP     │       SKILL.json                     │    │
│        └─────────│       skill.wasm  ◄── WASM-WASI      │    │
│                  │                       sandbox         │    │
│                  └─────────────────────────────────────-┘    │
└─────────────────────────────────────────────────────────────┘
                              │ HTTPS
              ┌───────────────▼────────────────┐
              │       registry.nullapt.dev      │
              │  (Go API + Postgres + Blob)     │
              │                                 │
              │  GET  /v1/skills/:name          │
              │  POST /v1/skills  (publish)     │
              │  GET  /v1/skills/:name/log      │
              └─────────────────────────────────┘
```

**Data flow for `nullapt get web-search`:**

1. CLI resolves `web-search` → latest version metadata from registry
2. Downloads `SKILL.json` and verifies Ed25519 signature against the author's public key
3. Checks the transparency log: has this key published this skill before?
4. Downloads `skill.wasm` from Vercel Blob
5. Installs both to `~/.nullapt/skills/web-search/`
6. MCP host discovers the new tool directory — skill is immediately usable

---

## Publishing

```bash
# 1. Generate your signing keypair (one-time)
nullapt keygen

# 2. Authenticate (one-time, GitHub OAuth)
nullapt login

# 3. Build your WASM binary (see above)

# 4. Sign your manifest
nullapt sign ./SKILL.json

# 5. Publish — uploads manifest + WASM, appends to the transparency log
nullapt publish ./SKILL.json
```

The registry verifies your signature server-side before accepting the upload. Your public key is permanently recorded in the transparency log so users can audit key history.

You don't need to clone any other repo to publish — same as `npm publish`. Work in your own repo, point the CLI at your `SKILL.json`, done.

---

## Publishing under an org namespace

Skill names look like `<scope>/<name>` (e.g. `nullapt/web-search`). The scope is either:

- **Your GitHub username** — works automatically for any logged-in user, no extra setup. Example: `enochthedev/cool-thing`.
- **A GitHub organization** — requires you to be a public member of that GitHub org and to grant NullApt the `read:org` scope.

To publish under an org namespace:

1. Make your org membership public — visit `github.com/orgs/<org>/people`, find yourself, click the gear/dropdown, set membership to **Public**.
2. Sign in to [nullapt.dev](https://nullapt.dev) and visit **/settings/organizations** → click **Connect GitHub organizations**. This re-runs OAuth with the `read:org` scope so we can verify your membership.
3. Set `name` in your `SKILL.json` to `<org>/<skill>` and publish as usual.

Casual contributors who just want to publish under their own name **never see the org screen** — the default sign-in only asks for `read:user user:email`, same as npm.

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full contributor flow.

---

## Self-Hosting the Registry

The registry API is a single Go binary.

```bash
# Clone
git clone https://github.com/nullapt/nullapt
cd nullapt

# Configure
cp .env.example .env
# Edit DATABASE_URL and BLOB_READ_WRITE_TOKEN

# Apply schema
psql $DATABASE_URL -f internal/db/schema.sql

# Run
go run ./cmd/registry-api
```

Point the CLI at your instance:

```bash
nullapt get my-private-skill --registry https://my-registry.internal
nullapt publish SKILL.json   --registry https://my-registry.internal
```

The web marketplace can be self-hosted too:

```bash
cd web
cp .env.example .env.local
# Set NEXT_PUBLIC_REGISTRY_URL=https://my-registry.internal
npm run build && npm start
```

---

## Contributing

We welcome contributions of all kinds. See [CONTRIBUTING.md](CONTRIBUTING.md) for full details.

**Quick start:**

```bash
git clone https://github.com/nullapt/nullapt
cd nullapt

# CLI
go test ./...
go build ./cmd/nullapt

# Web
cd web && npm install && npm run dev
```

**Good first issues** are labelled [`good first issue`](https://github.com/nullapt/nullapt/labels/good%20first%20issue) on GitHub.

---

## Security

NullApt takes security seriously. If you discover a vulnerability:

- **Do not open a public issue.**
- Email **security@nullapt.dev** or use [GitHub private vulnerability reporting](https://github.com/nullapt/nullapt/security/advisories/new).
- We aim to respond within 48 hours and publish a fix within 7 days.

See [SECURITY.md](SECURITY.md) for the full disclosure policy.

---

## License

[MIT](LICENSE) — Copyright © 2026 NullApt Contributors
