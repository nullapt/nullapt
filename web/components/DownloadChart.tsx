import type { GitHubRelease } from "@/lib/github";

interface Props {
  releases: GitHubRelease[];
}

function isBinary(name: string) {
  return name.endsWith(".tar.gz") || name.endsWith(".zip");
}

function totalDownloads(release: GitHubRelease) {
  return release.assets.reduce((s, a) => (isBinary(a.name) ? s + a.download_count : s), 0);
}

// Platform label from asset filename: nullapt_0.1.0_darwin_arm64.tar.gz → darwin arm64
function platformLabel(name: string): string {
  const m = name.match(/_(\w+)_(\w+)\.(tar\.gz|zip)$/);
  if (!m) return name;
  return `${m[1]} ${m[2]}`;
}

export function DownloadChart({ releases }: Props) {
  const published = releases.filter((r) => !r.draft).slice(0, 8);

  // ── Per-release total downloads bar chart ─────────────────────────────────
  const releaseData = published.map((r) => ({
    tag: r.tag_name,
    count: totalDownloads(r),
    prerelease: r.prerelease,
  }));

  const maxCount = Math.max(...releaseData.map((d) => d.count), 1);

  // ── Platform breakdown for latest release ─────────────────────────────────
  const latest = published[0];
  const platformData = latest
    ? latest.assets
        .filter((a) => isBinary(a.name))
        .map((a) => ({ label: platformLabel(a.name), count: a.download_count }))
        .sort((a, b) => b.count - a.count)
    : [];
  const maxPlatform = Math.max(...platformData.map((d) => d.count), 1);

  const totalAll = releaseData.reduce((s, d) => s + d.count, 0);

  return (
    <div className="flex flex-col gap-8">
      {/* Summary strip */}
      <div className="grid grid-cols-3 gap-4 text-center text-sm">
        {[
          ["Total downloads", totalAll.toLocaleString()],
          ["Releases", published.length.toString()],
          ["Latest", published[0]?.tag_name ?? "—"],
        ].map(([label, value]) => (
          <div
            key={label}
            style={{ border: "1px solid var(--border)", background: "var(--surface)" }}
            className="rounded p-4"
          >
            <div style={{ color: "var(--accent)" }} className="font-bold text-lg">
              {value}
            </div>
            <div style={{ color: "var(--muted)" }} className="text-xs mt-1">
              {label}
            </div>
          </div>
        ))}
      </div>

      {/* Downloads per release */}
      <div>
        <div className="text-xs font-semibold mb-4" style={{ color: "var(--muted)" }}>
          DOWNLOADS PER RELEASE
        </div>
        {releaseData.every((d) => d.count === 0) ? (
          <p className="text-xs" style={{ color: "var(--muted)" }}>
            No download data yet — check back after the first install.
          </p>
        ) : (
          <div className="flex flex-col gap-3">
            {releaseData.map((d) => {
              const pct = Math.max((d.count / maxCount) * 100, d.count > 0 ? 2 : 0);
              return (
                <div key={d.tag} className="flex items-center gap-3 text-xs">
                  <span
                    style={{ color: "var(--muted)", minWidth: "72px" }}
                    className="font-mono text-right"
                  >
                    {d.tag}
                  </span>
                  <div
                    style={{ background: "var(--border)", flex: 1 }}
                    className="rounded-full h-2 overflow-hidden"
                  >
                    <div
                      style={{
                        width: `${pct}%`,
                        background: d.prerelease ? "#fbbf24" : "#4ade80",
                        height: "100%",
                        borderRadius: "9999px",
                        transition: "width 0.3s ease",
                      }}
                    />
                  </div>
                  <span style={{ color: "var(--foreground)", minWidth: "48px" }} className="text-right">
                    {d.count.toLocaleString()}
                  </span>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Platform breakdown for latest */}
      {platformData.length > 0 && (
        <div>
          <div className="text-xs font-semibold mb-4" style={{ color: "var(--muted)" }}>
            PLATFORM BREAKDOWN · {latest?.tag_name}
          </div>
          {platformData.every((d) => d.count === 0) ? (
            <p className="text-xs" style={{ color: "var(--muted)" }}>
              No platform data yet.
            </p>
          ) : (
            <div className="flex flex-col gap-3">
              {platformData.map((d) => {
                const pct = Math.max((d.count / maxPlatform) * 100, d.count > 0 ? 2 : 0);
                return (
                  <div key={d.label} className="flex items-center gap-3 text-xs">
                    <span
                      style={{ color: "var(--muted)", minWidth: "140px" }}
                      className="font-mono"
                    >
                      {d.label}
                    </span>
                    <div
                      style={{ background: "var(--border)", flex: 1 }}
                      className="rounded-full h-2 overflow-hidden"
                    >
                      <div
                        style={{
                          width: `${pct}%`,
                          background: "#4ade80",
                          height: "100%",
                          borderRadius: "9999px",
                        }}
                      />
                    </div>
                    <span style={{ color: "var(--foreground)", minWidth: "48px" }} className="text-right">
                      {d.count.toLocaleString()}
                    </span>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
