import type { Metadata } from "next";
import { getReleases, type GitHubRelease } from "@/lib/github";

export const metadata: Metadata = {
  title: "Releases — NullApt",
  description: "NullApt CLI release history — changelogs, binaries, and upgrade instructions.",
  alternates: { canonical: "https://nullapt.dev/releases" },
  openGraph: {
    title: "Releases — NullApt",
    description: "NullApt CLI release history.",
    url: "https://nullapt.dev/releases",
  },
};

function totalDownloads(release: GitHubRelease): number {
  return release.assets.reduce((sum, a) => sum + a.download_count, 0);
}

function platformAssets(release: GitHubRelease) {
  return release.assets.filter((a) => a.name.endsWith(".tar.gz") || a.name.endsWith(".zip"));
}

export default async function ReleasesPage() {
  let releases: GitHubRelease[] = [];
  try {
    releases = await getReleases();
  } catch {
    // GitHub API unavailable — show empty state
  }

  const published = releases.filter((r) => !r.draft);

  return (
    <div className="max-w-4xl mx-auto px-6 py-12">
      {/* Header */}
      <div className="mb-10">
        <div className="text-xs mb-3" style={{ color: "var(--muted)" }}>
          <a href="/" className="hover:text-white transition-colors">~/nullapt</a>
          <span className="mx-2">/</span>
          <span style={{ color: "var(--accent)" }}>releases</span>
        </div>
        <h1 className="text-2xl font-bold mb-2">Releases</h1>
        <p style={{ color: "var(--muted)" }} className="text-sm">
          NullApt CLI release history. Install the latest version:
        </p>
        <div
          style={{ background: "var(--surface)", border: "1px solid var(--border)" }}
          className="rounded p-4 mt-4 inline-block text-sm"
        >
          <span style={{ color: "var(--muted)" }}>$ </span>
          <span style={{ color: "var(--accent)" }}>
            curl -fsSL https://nullapt.dev/install.sh | sh
          </span>
        </div>
      </div>

      {published.length === 0 ? (
        <div style={{ color: "var(--muted)" }} className="py-12 text-center text-sm">
          No releases yet.{" "}
          <a
            href="https://github.com/nullapt/nullapt/releases"
            target="_blank"
            rel="noreferrer"
            style={{ color: "var(--accent)" }}
            className="hover:underline"
          >
            Check GitHub
          </a>
        </div>
      ) : (
        <div className="flex flex-col gap-6">
          {published.map((release, idx) => {
            const platforms = platformAssets(release);
            const downloads = totalDownloads(release);
            const date = new Date(release.published_at).toISOString().slice(0, 10);

            return (
              <div
                key={release.id}
                style={{ border: "1px solid var(--border)", background: "var(--surface)" }}
                className="rounded overflow-hidden"
              >
                {/* Release header */}
                <div
                  className="px-6 py-4 flex items-center justify-between gap-4"
                  style={{ borderBottom: "1px solid var(--border)" }}
                >
                  <div className="flex items-center gap-3">
                    <a
                      href={release.html_url}
                      target="_blank"
                      rel="noreferrer"
                      style={{ color: "var(--accent)" }}
                      className="font-bold hover:underline"
                    >
                      {release.tag_name}
                    </a>
                    {idx === 0 && !release.prerelease && (
                      <span
                        style={{
                          background: "var(--accent-dim)",
                          color: "var(--accent)",
                          fontSize: "10px",
                        }}
                        className="px-2 py-0.5 rounded font-semibold"
                      >
                        latest
                      </span>
                    )}
                    {release.prerelease && (
                      <span
                        style={{
                          background: "#451a03",
                          color: "var(--warning)",
                          fontSize: "10px",
                        }}
                        className="px-2 py-0.5 rounded font-semibold"
                      >
                        pre-release
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-4 text-xs" style={{ color: "var(--muted)" }}>
                    <span>{downloads.toLocaleString()} downloads</span>
                    <span>{date}</span>
                  </div>
                </div>

                {/* Changelog */}
                {release.body && (
                  <div
                    className="px-6 py-4 text-sm"
                    style={{ color: "var(--muted)", borderBottom: "1px solid var(--border)", whiteSpace: "pre-wrap" }}
                  >
                    {release.body.slice(0, 800)}
                    {release.body.length > 800 && (
                      <>
                        {"…  "}
                        <a
                          href={release.html_url}
                          target="_blank"
                          rel="noreferrer"
                          style={{ color: "var(--accent)" }}
                          className="hover:underline"
                        >
                          read more
                        </a>
                      </>
                    )}
                  </div>
                )}

                {/* Download links */}
                {platforms.length > 0 && (
                  <div className="px-6 py-4">
                    <div className="text-xs mb-3" style={{ color: "var(--muted)" }}>
                      BINARIES
                    </div>
                    <div className="flex flex-wrap gap-2">
                      {platforms.map((asset) => (
                        <a
                          key={asset.id}
                          href={asset.browser_download_url}
                          style={{ border: "1px solid var(--border)", color: "var(--foreground)" }}
                          className="rounded px-3 py-1.5 text-xs hover:border-green-500 transition-colors flex items-center gap-2"
                        >
                          <span style={{ color: "var(--accent)" }}>↓</span>
                          {asset.name}
                          <span style={{ color: "var(--muted)" }}>
                            {(asset.size / 1024 / 1024).toFixed(1)}MB
                          </span>
                        </a>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
