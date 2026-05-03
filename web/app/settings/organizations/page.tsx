import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { authedFetch, getCurrentUser } from "@/lib/auth";

export const metadata: Metadata = {
  title: "Organizations — NullApt",
  description: "Connect GitHub organizations to publish skills under their namespace.",
  robots: { index: false },
};

const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL || "https://nullapt.dev";
const GITHUB_CLIENT_ID = process.env.NEXT_PUBLIC_GITHUB_CLIENT_ID || "";

type Org = {
  id: string;
  login: string;
  name: string | null;
  avatar_url: string | null;
};

export default async function OrganizationsPage() {
  const user = await getCurrentUser();
  if (!user) redirect("/login?next=/settings/organizations");

  let orgs: Org[] = [];
  try {
    orgs = (await authedFetch<Org[] | null>("/v1/me/orgs")) ?? [];
  } catch {
    // ignore — render empty state with the connect button
  }

  const redirectURI = encodeURIComponent(`${SITE_URL}/auth/callback`);
  const state = encodeURIComponent("/settings/organizations");
  const connectURL = GITHUB_CLIENT_ID
    ? `https://github.com/login/oauth/authorize?client_id=${GITHUB_CLIENT_ID}&redirect_uri=${redirectURI}&scope=read:user%20user:email%20read:org&state=${state}`
    : null;

  return (
    <div className="max-w-3xl mx-auto px-6 py-12">
      <div className="text-xs mb-3" style={{ color: "var(--muted)" }}>
        <a href="/" className="hover:text-white transition-colors">~/nullapt</a>
        <span className="mx-2">/</span>
        <a href="/settings/tokens" style={{ color: "var(--accent)" }} className="hover:underline">
          settings
        </a>
        <span className="mx-2">/</span>
        <span style={{ color: "var(--accent)" }}>organizations</span>
      </div>

      <h1 className="text-2xl font-bold mb-2">Organizations</h1>
      <p style={{ color: "var(--muted)" }} className="text-sm mb-8">
        Connect a GitHub organization to publish skills under its namespace
        (e.g. <code>nullapt/web-search</code>). You only need this if you want
        to publish under a shared brand — your personal namespace
        (<code>{user.username}/...</code>) works without any extra setup.
      </p>

      {orgs.length === 0 ? (
        <section className="border border-dashed rounded p-6 mb-6" style={{ borderColor: "var(--border)" }}>
          <h2 className="font-semibold mb-2">No organizations connected</h2>
          <p style={{ color: "var(--muted)" }} className="text-sm mb-4">
            Click below to grant NullApt read-only access to your GitHub org
            memberships. We use this only to verify you can publish under an
            org&apos;s namespace.
          </p>
          {connectURL ? (
            <a
              href={connectURL}
              className="inline-block px-4 py-2 rounded text-sm font-medium"
              style={{ background: "var(--accent)", color: "var(--bg)" }}
            >
              Connect GitHub organizations
            </a>
          ) : (
            <p style={{ color: "var(--muted)" }} className="text-sm">
              GitHub OAuth is not configured.
            </p>
          )}
        </section>
      ) : (
        <>
          <h2 className="text-xs font-semibold mb-3" style={{ color: "var(--muted)" }}>
            CONNECTED
          </h2>
          <ul className="space-y-2 mb-8">
            {orgs.map((o) => (
              <li
                key={o.id}
                className="flex items-center gap-3 border rounded px-4 py-3"
                style={{ borderColor: "var(--border)" }}
              >
                {o.avatar_url && (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={o.avatar_url}
                    alt={o.login}
                    width={32}
                    height={32}
                    style={{ borderRadius: "4px" }}
                  />
                )}
                <div className="flex-1 min-w-0">
                  <div className="font-medium">{o.login}</div>
                  {o.name && (
                    <div className="text-xs truncate" style={{ color: "var(--muted)" }}>
                      {o.name}
                    </div>
                  )}
                </div>
                <code className="text-xs" style={{ color: "var(--muted)" }}>
                  {o.login}/&lt;skill&gt;
                </code>
              </li>
            ))}
          </ul>
          {connectURL && (
            <a
              href={connectURL}
              className="inline-block px-3 py-2 rounded text-xs"
              style={{ border: "1px solid var(--border)", color: "var(--muted)" }}
            >
              Re-sync organizations
            </a>
          )}
        </>
      )}

      <section
        className="mt-12 border-t pt-6 text-sm"
        style={{ borderColor: "var(--border)", color: "var(--muted)" }}
      >
        <h3 className="font-semibold mb-2" style={{ color: "var(--fg)" }}>
          Why don&apos;t I see my org?
        </h3>
        <p className="mb-2">
          GitHub only exposes orgs where your membership is set to{" "}
          <strong>public</strong>. To make yours public, visit your org&apos;s
          People page on GitHub (e.g.{" "}
          <code>github.com/orgs/&lt;org&gt;/people</code>) and switch your row
          from <em>Private</em> to <em>Public</em>, then click Re-sync above.
        </p>
        <p>
          If your org has restricted third-party OAuth apps, an org owner needs
          to approve the NullApt OAuth app at the org level first.
        </p>
      </section>
    </div>
  );
}
