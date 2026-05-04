import type { Metadata } from "next";
import { Suspense } from "react";
import Link from "next/link";
import { SkillsBrowser, SkillsSkeleton } from "@/components/SkillsBrowser";

export const metadata: Metadata = {
  title: "Browse MCP Skills — Signed & WASM-Sandboxed",
  description:
    "Browse the NullApt registry — every MCP skill is Ed25519 signed, WASM-sandboxed, and offline-first. Filter by network access, filesystem permissions, license, and author.",
  alternates: { canonical: "https://nullapt.dev/skills" },
  openGraph: {
    title: "Browse MCP Skills — NullApt",
    description:
      "Verified, signed, sandboxed MCP skills for Claude Desktop, Cursor, LM Studio, and Ollama.",
    url: "https://nullapt.dev/skills",
  },
};

export default function SkillsBrowsePage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string }>;
}) {
  return (
    <div className="max-w-6xl mx-auto px-6 py-12">
      <div className="text-xs mb-3" style={{ color: "var(--muted)" }}>
        <Link href="/" className="hover:text-white transition-colors">~/nullapt</Link>
        <span className="mx-2">/</span>
        <span style={{ color: "var(--accent)" }}>skills</span>
      </div>
      <h1 className="text-3xl font-bold mb-3 leading-tight">
        Browse MCP Skills
      </h1>
      <p style={{ color: "var(--muted)" }} className="mb-10 max-w-2xl text-sm leading-relaxed">
        Every skill in the NullApt registry is Ed25519 signed, runs inside a WASM-WASI
        sandbox, and declares exactly which network domains, files, and env vars it can
        access. Install any skill into Claude Desktop, Cursor, LM Studio, or Ollama with{" "}
        <code style={{ color: "var(--accent)" }}>nullapt get &lt;skill&gt;</code>.
      </p>
      <Suspense fallback={<SkillsSkeleton />}>
        <SkillsBrowser searchParams={searchParams} showTrending={false} />
      </Suspense>
    </div>
  );
}
