import Link from "next/link";
import { listSkills, type SkillMeta } from "@/lib/api";

async function fetchSkills(q?: string): Promise<SkillMeta[]> {
  try {
    return await listSkills(q);
  } catch {
    return [];
  }
}

export async function SkillsBrowser({
  searchParams,
  showTrending = true,
}: {
  searchParams: Promise<{ q?: string }>;
  showTrending?: boolean;
}) {
  const { q } = await searchParams;
  const skills = await fetchSkills(q);

  const trending = showTrending && !q
    ? [...skills].sort((a, b) => b.downloads - a.downloads).slice(0, 5)
    : [];

  return (
    <>
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

      {trending.length > 0 && !q && (
        <div className="mt-14">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xs font-semibold" style={{ color: "var(--muted)" }}>
              TRENDING SKILLS
            </h2>
            <Link href="/skills" style={{ color: "var(--accent)" }} className="text-xs hover:underline">
              browse all →
            </Link>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {trending.map((skill, rank) => {
              const maxDownloads = trending[0].downloads || 1;
              const pct = Math.max((skill.downloads / maxDownloads) * 100, skill.downloads > 0 ? 4 : 0);
              return (
                <Link
                  key={skill.name}
                  href={`/skills/${skill.name}`}
                  style={{ border: "1px solid var(--border)", background: "var(--surface)" }}
                  className="rounded p-4 hover:border-green-500 transition-colors flex flex-col gap-2"
                >
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <span style={{ color: "var(--accent)" }} className="font-semibold text-sm">
                        {skill.name}
                      </span>
                      <span style={{ color: "var(--muted)" }} className="text-xs ml-2">
                        v{skill.version}
                      </span>
                    </div>
                    <span
                      style={{ color: "var(--muted)", background: "var(--border)", fontSize: "10px" }}
                      className="px-1.5 py-0.5 rounded shrink-0"
                    >
                      #{rank + 1}
                    </span>
                  </div>
                  <p className="text-xs line-clamp-1" style={{ color: "var(--muted)" }}>
                    {skill.description}
                  </p>
                  <div className="flex items-center gap-2 mt-1">
                    <div
                      style={{ background: "var(--border)", flex: 1 }}
                      className="rounded-full h-1 overflow-hidden"
                    >
                      <div
                        style={{ width: `${pct}%`, background: "#4ade80", height: "100%", borderRadius: "9999px" }}
                      />
                    </div>
                    <span className="text-xs shrink-0" style={{ color: "var(--muted)" }}>
                      {skill.downloads.toLocaleString()} dl
                    </span>
                  </div>
                </Link>
              );
            })}
          </div>
        </div>
      )}
    </>
  );
}

export function SkillsSkeleton() {
  return (
    <div style={{ color: "var(--muted)" }} className="py-12 text-center text-sm">
      Loading skills…
    </div>
  );
}
