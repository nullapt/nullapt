import Link from "next/link";
import { listSkills, type SkillMeta } from "@/lib/api";

async function fetchSkills(q?: string): Promise<SkillMeta[]> {
  try {
    return await listSkills(q);
  } catch {
    return [];
  }
}

export default async function HomePage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string }>;
}) {
  const { q } = await searchParams;
  const skills = await fetchSkills(q);

  return (
    <div className="max-w-6xl mx-auto px-6 py-12">
      {/* Hero */}
      <div className="mb-12">
        <div className="text-xs mb-4" style={{ color: "var(--muted)" }}>
          <span style={{ color: "var(--accent)" }}>●</span> LIVE · nullapt registry v1
        </div>
        <h1 className="text-3xl font-bold mb-3 leading-tight">
          The Private-First Registry
          <br />
          <span style={{ color: "var(--accent)" }}>for AI Skills</span>
        </h1>
        <p style={{ color: "var(--muted)" }} className="mb-8 max-w-xl">
          Every skill is cryptographically signed, statically analyzed, and sandboxed in WASM-WASI.
          No data leaves your machine unless explicitly declared in the manifest.
        </p>
        <div
          style={{ background: "var(--surface)", border: "1px solid var(--border)" }}
          className="rounded p-4 inline-block text-sm"
        >
          <span style={{ color: "var(--muted)" }}>$ </span>
          <span style={{ color: "var(--accent)" }}>nullapt get web-search</span>
        </div>
      </div>

      {/* Search */}
      <form method="GET" className="mb-8">
        <div className="flex gap-3 max-w-lg">
          <input
            type="text"
            name="q"
            defaultValue={q ?? ""}
            placeholder="search skills..."
            style={{
              background: "var(--surface)",
              border: "1px solid var(--border)",
              color: "var(--foreground)",
            }}
            className="flex-1 rounded px-4 py-2 text-sm outline-none focus:border-green-500 placeholder:text-zinc-600"
          />
          <button
            type="submit"
            style={{ background: "var(--accent)", color: "#000" }}
            className="rounded px-4 py-2 text-sm font-semibold hover:opacity-90 transition-opacity"
          >
            search
          </button>
        </div>
      </form>

      {/* Skill table */}
      {skills.length === 0 ? (
        <div style={{ color: "var(--muted)" }} className="py-12 text-center text-sm">
          {q ? `No skills matched "${q}"` : "Registry is empty or unreachable."}
        </div>
      ) : (
        <div style={{ border: "1px solid var(--border)" }} className="rounded overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr
                style={{
                  background: "var(--surface)",
                  color: "var(--muted)",
                  borderBottom: "1px solid var(--border)",
                }}
              >
                <th className="text-left px-4 py-3 font-medium">skill</th>
                <th className="text-left px-4 py-3 font-medium">version</th>
                <th className="text-left px-4 py-3 font-medium hidden md:table-cell">author</th>
                <th className="text-left px-4 py-3 font-medium hidden lg:table-cell">description</th>
                <th className="text-right px-4 py-3 font-medium">downloads</th>
              </tr>
            </thead>
            <tbody>
              {skills.map((skill: SkillMeta, i: number) => (
                <tr
                  key={skill.name}
                  style={{ borderTop: i > 0 ? "1px solid var(--border)" : undefined }}
                  className="hover:bg-white/2 transition-colors"
                >
                  <td className="px-4 py-3">
                    <Link
                      href={`/skills/${skill.name}`}
                      style={{ color: "var(--accent)" }}
                      className="hover:underline font-semibold"
                    >
                      {skill.name}
                    </Link>
                  </td>
                  <td className="px-4 py-3" style={{ color: "var(--muted)" }}>
                    {skill.version}
                  </td>
                  <td className="px-4 py-3 hidden md:table-cell" style={{ color: "var(--muted)" }}>
                    {skill.author}
                  </td>
                  <td className="px-4 py-3 hidden lg:table-cell" style={{ color: "var(--muted)" }}>
                    <span className="line-clamp-1">{skill.description}</span>
                  </td>
                  <td className="px-4 py-3 text-right" style={{ color: "var(--muted)" }}>
                    {skill.downloads.toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Stats strip */}
      <div className="mt-12 grid grid-cols-3 gap-4 text-center text-sm" style={{ color: "var(--muted)" }}>
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
      </div>
    </div>
  );
}
