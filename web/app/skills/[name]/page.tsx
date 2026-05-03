import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getSkill, getTransparencyLog, type TransparencyEntry } from "@/lib/api";

const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL || "https://nullapt.dev";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ name: string }>;
}): Promise<Metadata> {
  const { name } = await params;
  try {
    const skill = await getSkill(name);
    const title = `${skill.name} — NullApt`;
    const description = `${skill.description} · by ${skill.author} · v${skill.version} · ${skill.downloads.toLocaleString()} downloads`;
    return {
      title,
      description,
      openGraph: {
        title,
        description,
        url: `${SITE_URL}/skills/${name}`,
        type: "website",
      },
      twitter: { card: "summary", title, description },
      alternates: { canonical: `${SITE_URL}/skills/${name}` },
    };
  } catch {
    return { title: "Skill not found · NullApt" };
  }
}

export default async function SkillPage({ params }: { params: Promise<{ name: string }> }) {
  const { name } = await params;

  let skill;
  try {
    skill = await getSkill(name);
  } catch {
    notFound();
  }

  let log: TransparencyEntry[] = [];
  try {
    log = await getTransparencyLog(name);
  } catch {
    // transparency log is best-effort
  }

  const manifest = skill as typeof skill & {
    permissions?: {
      network: { allowed: boolean; domains?: string[] };
      filesystem: { read: string[]; write: string[] };
      env: string[];
    };
    interface?: { tools: { name: string; description: string }[] };
  };

  const perms = manifest.permissions;
  const tools = manifest.interface?.tools ?? [];

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "SoftwareSourceCode",
    name: skill.name,
    description: skill.description,
    version: skill.version,
    author: { "@type": "Person", name: skill.author },
    license: `https://spdx.org/licenses/${skill.license}.html`,
    url: `${SITE_URL}/skills/${skill.name}`,
    programmingLanguage: "WebAssembly",
    downloadUrl: skill.wasm_url,
  };

  return (
    <div className="max-w-4xl mx-auto px-6 py-12">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      {/* Breadcrumb */}
      <div className="text-xs mb-6" style={{ color: "var(--muted)" }}>
        <a href="/" className="hover:text-white transition-colors">~/nullapt</a>
        <span className="mx-2">/</span>
        <span style={{ color: "var(--accent)" }}>{skill.name}</span>
      </div>

      {/* Header */}
      <div className="flex items-start justify-between mb-8 gap-4">
        <div>
          <h1 className="text-2xl font-bold mb-1">{skill.name}</h1>
          <p style={{ color: "var(--muted)" }} className="text-sm">{skill.description}</p>
        </div>
        <div
          style={{ background: "var(--surface)", border: "1px solid var(--border)" }}
          className="rounded p-4 text-sm shrink-0"
        >
          <div style={{ color: "var(--muted)" }} className="text-xs mb-2">install</div>
          <code style={{ color: "var(--accent)" }}>nullapt get {skill.name}</code>
        </div>
      </div>

      {/* Meta row */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-10">
        {[
          ["version", skill.version],
          ["author", skill.author],
          ["license", skill.license],
          ["downloads", skill.downloads.toLocaleString()],
        ].map(([label, value]) => (
          <div
            key={label}
            style={{ background: "var(--surface)", border: "1px solid var(--border)" }}
            className="rounded p-3"
          >
            <div style={{ color: "var(--muted)" }} className="text-xs mb-1">{label}</div>
            <div className="text-sm font-semibold">{value}</div>
          </div>
        ))}
      </div>

      {/* Permissions */}
      {perms && (
        <section className="mb-10">
          <h2 className="text-sm font-semibold mb-3" style={{ color: "var(--muted)" }}>
            PERMISSIONS
          </h2>
          <div
            style={{ border: "1px solid var(--border)", background: "var(--surface)" }}
            className="rounded overflow-hidden"
          >
            <PermRow
              label="Network"
              granted={perms.network.allowed}
              detail={perms.network.domains?.join(", ") ?? "none"}
            />
            <PermRow
              label="Filesystem Read"
              granted={perms.filesystem.read.length > 0}
              detail={perms.filesystem.read.join(", ") || "none"}
            />
            <PermRow
              label="Filesystem Write"
              granted={perms.filesystem.write.length > 0}
              detail={perms.filesystem.write.join(", ") || "none"}
            />
            <PermRow
              label="Env Vars"
              granted={perms.env.length > 0}
              detail={perms.env.join(", ") || "none"}
              last
            />
          </div>
        </section>
      )}

      {/* Tools */}
      {tools.length > 0 && (
        <section className="mb-10">
          <h2 className="text-sm font-semibold mb-3" style={{ color: "var(--muted)" }}>
            TOOLS ({tools.length})
          </h2>
          <div className="flex flex-col gap-3">
            {tools.map((t) => (
              <div
                key={t.name}
                style={{ border: "1px solid var(--border)", background: "var(--surface)" }}
                className="rounded p-4"
              >
                <div style={{ color: "var(--accent)" }} className="font-semibold text-sm mb-1">
                  {t.name}
                </div>
                <div style={{ color: "var(--muted)" }} className="text-xs">{t.description}</div>
              </div>
            ))}
          </div>
        </section>
      )}

      {/* Transparency Log */}
      <section>
        <h2 className="text-sm font-semibold mb-3" style={{ color: "var(--muted)" }}>
          TRANSPARENCY LOG
        </h2>
        {log.length === 0 ? (
          <p className="text-xs" style={{ color: "var(--muted)" }}>No log entries found.</p>
        ) : (
          <div style={{ border: "1px solid var(--border)" }} className="rounded overflow-hidden">
            {log.map((entry, i) => (
              <div
                key={i}
                style={{ borderTop: i > 0 ? "1px solid var(--border)" : undefined }}
                className="px-4 py-3 text-xs"
              >
                <div className="flex items-center justify-between mb-1">
                  <span style={{ color: "var(--accent)" }}>v{entry.Version}</span>
                  <span style={{ color: "var(--muted)" }}>
                    {new Date(entry.RecordedAt).toISOString().slice(0, 10)}
                  </span>
                </div>
                <div style={{ color: "var(--muted)" }} className="font-mono truncate">
                  key: {entry.PublicKeyB64.slice(0, 32)}…
                </div>
                <div style={{ color: "var(--muted)" }} className="font-mono truncate">
                  sha: {entry.ManifestHash.slice(0, 32)}…
                </div>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

function PermRow({
  label,
  granted,
  detail,
  last,
}: {
  label: string;
  granted: boolean;
  detail: string;
  last?: boolean;
}) {
  return (
    <div
      className="flex items-start justify-between px-4 py-3 text-sm"
      style={{ borderBottom: last ? undefined : "1px solid var(--border)" }}
    >
      <div className="flex items-center gap-3">
        <span
          style={{ color: granted ? "var(--danger)" : "var(--accent)" }}
          className="text-xs font-bold"
        >
          {granted ? "GRANTED" : "DENIED"}
        </span>
        <span>{label}</span>
      </div>
      <span style={{ color: "var(--muted)" }} className="text-xs font-mono max-w-xs truncate text-right">
        {detail}
      </span>
    </div>
  );
}
