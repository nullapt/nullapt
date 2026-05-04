import { Suspense } from "react";
import { getRepo } from "@/lib/github";
import { SkillsBrowser, SkillsSkeleton } from "@/components/SkillsBrowser";

async function fetchStars(): Promise<number | null> {
  try {
    const repo = await getRepo();
    return repo.stargazers_count;
  } catch {
    return null;
  }
}

export default function HomePage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string }>;
}) {
  return (
    <div className="max-w-6xl mx-auto px-6 py-12">
      <Hero />
      <Suspense fallback={<SkillsSkeleton />}>
        <SkillsBrowser searchParams={searchParams} />
      </Suspense>
      <Suspense fallback={null}>
        <Stats />
      </Suspense>
      <KeywordSections />
    </div>
  );
}

function Hero() {
  return (
    <div className="mb-12">
      <div className="text-xs mb-4" style={{ color: "var(--muted)" }}>
        <span style={{ color: "var(--accent)" }}>●</span> LIVE · nullapt registry v1
      </div>
      <h1 className="text-3xl font-bold mb-3 leading-tight">
        The Secure Package Manager
        <br />
        <span style={{ color: "var(--accent)" }}>for MCP Skills</span>
      </h1>
      <p style={{ color: "var(--muted)" }} className="mb-8 max-w-xl">
        Install MCP servers and AI skills you can actually trust. Every skill is Ed25519
        signed, WASM-WASI sandboxed, and offline-first — no data leaves your machine
        unless the manifest explicitly declares it. Works with Claude Desktop, Cursor,
        LM Studio, and Ollama.
      </p>
      <div
        style={{ background: "var(--surface)", border: "1px solid var(--border)" }}
        className="rounded p-4 inline-block text-sm"
      >
        <span style={{ color: "var(--muted)" }}>$ </span>
        <span style={{ color: "var(--accent)" }}>nullapt get web-search</span>
      </div>
    </div>
  );
}

function KeywordSections() {
  return (
    <section className="mt-16 grid grid-cols-1 md:grid-cols-3 gap-4">
      <div
        style={{ border: "1px solid var(--border)", background: "var(--surface)" }}
        className="rounded p-5"
      >
        <h2 className="text-sm font-semibold mb-2" style={{ color: "var(--accent)" }}>
          Works with Claude, Cursor, LM Studio &amp; Ollama
        </h2>
        <p className="text-xs leading-relaxed" style={{ color: "var(--muted)" }}>
          Register <code style={{ color: "var(--accent)" }}>nullapt mcp</code> once in your
          MCP host and every installed skill is exposed as a tool. One CLI manages MCP
          servers across every major AI client.
        </p>
      </div>
      <div
        style={{ border: "1px solid var(--border)", background: "var(--surface)" }}
        className="rounded p-5"
      >
        <h2 className="text-sm font-semibold mb-2" style={{ color: "var(--accent)" }}>
          How NullApt Secures Every Skill
        </h2>
        <p className="text-xs leading-relaxed" style={{ color: "var(--muted)" }}>
          Every published skill carries an Ed25519 signature, a public transparency-log
          record, and a manifest declaring exactly which network domains, files, and env
          vars it can touch. Compromised keys are detectable, not silent.
        </p>
      </div>
      <div
        style={{ border: "1px solid var(--border)", background: "var(--surface)" }}
        className="rounded p-5"
      >
        <h2 className="text-sm font-semibold mb-2" style={{ color: "var(--accent)" }}>
          Why WASM Sandboxing Matters
        </h2>
        <p className="text-xs leading-relaxed" style={{ color: "var(--muted)" }}>
          Skills run in a WASM-WASI runtime that physically cannot reach outside their
          declared permissions. Other MCP registries trust the manifest; NullApt enforces
          it at runtime.
        </p>
      </div>
    </section>
  );
}

async function Stats() {
  const stars = await fetchStars();

  return (
    <div className="mt-12 grid grid-cols-2 md:grid-cols-4 gap-4 text-center text-sm" style={{ color: "var(--muted)" }}>
      {(
        [
          ["Ed25519 Signed", "every manifest verified"],
          ["WASM Sandboxed", "hard capability limits"],
          ["Offline First", "no cloud dependency"],
        ] as const
      ).map(([title, sub]) => (
        <div
          key={title}
          style={{ border: "1px solid var(--border)", background: "var(--surface)" }}
          className="rounded p-4"
        >
          <div style={{ color: "var(--accent)" }} className="font-semibold mb-1">
            {title}
          </div>
          <div className="text-xs">{sub}</div>
        </div>
      ))}
      {stars !== null && (
        <a
          href="https://github.com/nullapt/nullapt"
          target="_blank"
          rel="noreferrer"
          style={{ border: "1px solid var(--border)", background: "var(--surface)" }}
          className="rounded p-4 hover:border-green-500 transition-colors"
        >
          <div style={{ color: "var(--accent)" }} className="font-semibold mb-1">
            {stars.toLocaleString()} stars
          </div>
          <div className="text-xs">on GitHub</div>
        </a>
      )}
    </div>
  );
}
