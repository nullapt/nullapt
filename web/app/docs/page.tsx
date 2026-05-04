import type { Metadata } from "next";
import Integrations from "./Integrations";

export const metadata: Metadata = {
  title: "Docs — NullApt",
  description: "NullApt documentation — install the CLI, build and publish skills, command reference.",
  alternates: { canonical: "https://nullapt.dev/docs" },
  openGraph: {
    title: "Docs — NullApt",
    description: "Install, build, and publish AI skills with NullApt.",
    url: "https://nullapt.dev/docs",
  },
};

function Section({ id, title, children }: { id: string; title: string; children: React.ReactNode }) {
  return (
    <section id={id} className="mb-12">
      <h2
        className="text-sm font-semibold mb-4 pb-2"
        style={{ color: "var(--muted)", borderBottom: "1px solid var(--border)" }}
      >
        {title}
      </h2>
      {children}
    </section>
  );
}

function Code({ children }: { children: string }) {
  return (
    <div
      style={{ background: "var(--surface)", border: "1px solid var(--border)" }}
      className="rounded p-4 text-sm font-mono overflow-x-auto mb-4"
    >
      {children.split("\n").map((line, i) => (
        <div key={i}>
          {line.startsWith("#") ? (
            <span style={{ color: "var(--muted)" }}>{line}</span>
          ) : line.startsWith("$") ? (
            <>
              <span style={{ color: "var(--muted)" }}>$ </span>
              <span style={{ color: "var(--accent)" }}>{line.slice(2)}</span>
            </>
          ) : (
            <span>{line}</span>
          )}
        </div>
      ))}
    </div>
  );
}

function Cmd({ name, desc }: { name: string; desc: string }) {
  return (
    <div
      className="flex items-start gap-4 px-4 py-3 text-sm"
      style={{ borderBottom: "1px solid var(--border)" }}
    >
      <code style={{ color: "var(--accent)", minWidth: "220px" }} className="shrink-0">
        {name}
      </code>
      <span style={{ color: "var(--muted)" }}>{desc}</span>
    </div>
  );
}

const TOC = [
  ["installation", "Installation"],
  ["quick-start", "Quick Start"],
  ["integrations", "Use in your client"],
  ["commands", "Command Reference"],
  ["building", "Building a Skill"],
  ["signing", "Signing & Publishing"],
  ["self-hosting", "Self-Hosting"],
];

export default function DocsPage() {
  return (
    <div className="max-w-5xl mx-auto px-6 py-12 flex gap-12">
      {/* Sidebar TOC */}
      <aside className="hidden lg:block w-48 shrink-0">
        <div className="sticky top-20">
          <div className="text-xs mb-3" style={{ color: "var(--muted)" }}>ON THIS PAGE</div>
          <nav className="flex flex-col gap-2">
            {TOC.map(([id, label]) => (
              <a
                key={id}
                href={`#${id}`}
                className="text-xs hover:text-white transition-colors"
                style={{ color: "var(--muted)" }}
              >
                {label}
              </a>
            ))}
          </nav>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 min-w-0">
        <div className="text-xs mb-3" style={{ color: "var(--muted)" }}>
          <a href="/" className="hover:text-white transition-colors">~/nullapt</a>
          <span className="mx-2">/</span>
          <span style={{ color: "var(--accent)" }}>docs</span>
        </div>
        <h1 className="text-2xl font-bold mb-8">Documentation</h1>

        {/* Installation */}
        <Section id="installation" title="INSTALLATION">
          <p style={{ color: "var(--muted)" }} className="text-sm mb-4">
            The fastest way to install NullApt on macOS or Linux:
          </p>
          <Code>$ curl -fsSL https://nullapt.dev/install.sh | sh</Code>

          <p style={{ color: "var(--muted)" }} className="text-sm mb-2 mt-6">Go (requires Go 1.23+)</p>
          <Code>$ go install github.com/nullapt/nullapt/cmd/nullapt@latest</Code>

          <p style={{ color: "var(--muted)" }} className="text-sm mt-6 mb-2">Windows</p>
          <p style={{ color: "var(--muted)" }} className="text-sm">
            Download the latest <code style={{ color: "var(--accent)" }}>.exe</code> from the{" "}
            <a
              href="https://github.com/nullapt/nullapt/releases"
              target="_blank"
              rel="noreferrer"
              style={{ color: "var(--accent)" }}
              className="hover:underline"
            >
              releases page
            </a>
            .
          </p>
        </Section>

        {/* Quick start */}
        <Section id="quick-start" title="QUICK START">
          <Code>{`# Install a skill from the registry (org/name format)
$ nullapt get nullapt/prompt-optimizer

# Short form (same as get)
$ nullapt i nullapt/prompt-optimizer

# Install a specific version
$ nullapt i nullapt/prompt-optimizer@1.2.0

# List installed skills
$ nullapt list

# Remove a skill
$ nullapt remove nullapt/prompt-optimizer`}</Code>
          <p style={{ color: "var(--muted)" }} className="text-sm">
            Skills install to <code style={{ color: "var(--accent)" }}>~/.nullapt/skills/</code>.
            Once you register{" "}
            <code style={{ color: "var(--accent)" }}>nullapt mcp</code> as an MCP server in your
            client (see <a href="#integrations" style={{ color: "var(--accent)" }} className="hover:underline">Use in your client</a>),
            every installed skill is exposed as a tool — newly installed skills appear without
            restarting the host.
          </p>

          <p style={{ color: "var(--muted)" }} className="text-sm mt-4">
            Or invoke a tool directly from the shell:
          </p>
          <Code>{`$ nullapt run nullapt/web-search web_search --input '{"query":"linux kernel"}'`}</Code>
        </Section>

        {/* Integrations */}
        <Section id="integrations" title="USE IN YOUR CLIENT">
          <Integrations />
        </Section>

        {/* Commands */}
        <Section id="commands" title="COMMAND REFERENCE">
          <div
            style={{ border: "1px solid var(--border)", background: "var(--surface)" }}
            className="rounded overflow-hidden"
          >
            <Cmd name="nullapt get <skill[@ver]>" desc="Install a skill from the registry. Alias: i, install" />
            <Cmd name="nullapt remove <skill>" desc="Uninstall a skill" />
            <Cmd name="nullapt list" desc="List installed skills" />
            <Cmd name="nullapt run <skill> <tool>" desc="Invoke a tool from the shell. Reads JSON from --input/--input-file/stdin" />
            <Cmd name="nullapt mcp" desc="Stdio MCP server exposing every installed skill — register in your client config" />
            <Cmd name="nullapt verify <SKILL.json>" desc="Verify a manifest's Ed25519 signature" />
            <Cmd name="nullapt keygen" desc="Generate an Ed25519 signing keypair" />
            <Cmd name="nullapt sign <SKILL.json>" desc="Sign a manifest in place with your private key" />
            <Cmd name="nullapt publish <SKILL.json>" desc="Publish a skill (requires login)" />
            <Cmd name="nullapt login" desc="Authenticate with the registry" />
            <Cmd name="nullapt logout" desc="Remove stored credentials" />
          </div>

          <p style={{ color: "var(--muted)" }} className="text-sm mt-4 mb-2">Flags</p>
          <div
            style={{ border: "1px solid var(--border)", background: "var(--surface)" }}
            className="rounded overflow-hidden"
          >
            <Cmd name="--registry <url>" desc="Use a custom or self-hosted registry" />
            <Cmd name="--skip-verify" desc="Skip signature check (development only)" />
          </div>
        </Section>

        {/* Building */}
        <Section id="building" title="BUILDING A SKILL">
          <p style={{ color: "var(--muted)" }} className="text-sm mb-4">
            Skills are WASM-WASI binaries compiled from any language that targets{" "}
            <code style={{ color: "var(--accent)" }}>wasm32-wasip1</code>. Rust is recommended;
            TinyGo, AssemblyScript, Zig, and C/C++ also work.
          </p>
          <Code>{`# 0. Add the WASI target (one-time)
$ rustup target add wasm32-wasip1

# 1. Create a new Rust library
$ cargo new --lib my-skill && cd my-skill
$ cargo add extism-pdk

# 2. Build for WASM
$ cargo build --target wasm32-wasip1 --release
$ cp target/wasm32-wasip1/release/my_skill.wasm skill.wasm`}</Code>

          <p style={{ color: "var(--muted)" }} className="text-sm mb-4">
            Your skill exports named functions that match the tool names in{" "}
            <code style={{ color: "var(--accent)" }}>interface.tools</code>:
          </p>
          <Code>{`// src/lib.rs
use extism_pdk::*;

#[plugin_fn]
pub fn web_search(input: Json<SearchInput>) -> FnResult<Json<SearchOutput>> {
    // your logic here
}`}</Code>

          <p style={{ color: "var(--muted)" }} className="text-sm mb-4">
            Write a <code style={{ color: "var(--accent)" }}>SKILL.json</code> manifest that declares
            exactly which permissions your skill needs:
          </p>
          <div
            style={{ background: "var(--surface)", border: "1px solid var(--border)" }}
            className="rounded p-4 text-sm font-mono overflow-x-auto mb-4"
          >
            <pre style={{ color: "var(--foreground)" }}>{JSON.stringify({
              schema_version: "1.0",
              name: "your-username/my-skill",
              version: "0.1.0",
              description: "What your skill does",
              author: "your-username",
              license: "MIT",
              permissions: {
                network: { allowed: false },
                filesystem: { read: [], write: [] },
                env: [],
              },
              entry: "skill.wasm",
              interface: {
                tools: [{
                  name: "my_tool",
                  description: "Tool description",
                  input_schema: { type: "object", properties: { query: { type: "string" } }, required: ["query"] },
                }],
              },
            }, null, 2)}</pre>
          </div>
        </Section>

        {/* Signing & Publishing */}
        <Section id="signing" title="SIGNING & PUBLISHING">
          <p style={{ color: "var(--muted)" }} className="text-sm mb-4">
            Every skill must be signed before publishing. NullApt uses Ed25519 — your private key
            never leaves your machine.
          </p>
          <Code>{`# Generate a key pair (one-time setup)
$ nullapt keygen --output ~/.nullapt/keys

# Sign your manifest (embeds signature in SKILL.json)
$ nullapt sign ./SKILL.json --key ~/.nullapt/keys/private.pem

# Verify the signature locally
$ nullapt verify ./SKILL.json

# Publish — uploads manifest + WASM, records in transparency log
$ nullapt login
$ nullapt publish ./SKILL.json`}</Code>
          <p style={{ color: "var(--muted)" }} className="text-sm">
            Your public key is permanently recorded in the{" "}
            <strong style={{ color: "var(--foreground)" }}>transparency log</strong> so users can
            audit key history and detect compromise.
          </p>
        </Section>

        {/* Self-hosting */}
        <Section id="self-hosting" title="SELF-HOSTING">
          <p style={{ color: "var(--muted)" }} className="text-sm mb-4">
            The registry API is a single Go binary. Run your own private registry in minutes.
          </p>
          <Code>{`# Clone and configure
$ git clone https://github.com/nullapt/nullapt
$ cd nullapt
$ cp .env.example .env
# Edit DATABASE_URL and BLOB_READ_WRITE_TOKEN

# Apply schema
$ psql $DATABASE_URL -f internal/db/schema.sql

# Run
$ go run ./cmd/registry-api`}</Code>
          <p style={{ color: "var(--muted)" }} className="text-sm mb-4">
            Point the CLI at your instance:
          </p>
          <Code>{`$ nullapt i my-org/my-skill --registry https://my-registry.internal
$ nullapt publish SKILL.json --registry https://my-registry.internal`}</Code>
        </Section>
      </div>
    </div>
  );
}
