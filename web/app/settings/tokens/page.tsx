import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { authedFetch, getCurrentUser, type APITokenMeta } from "@/lib/auth";
import { CreateTokenForm } from "./CreateTokenForm";
import { RevokeTokenButton } from "./RevokeTokenButton";

export const metadata: Metadata = {
  title: "API tokens — NullApt",
  description: "Manage API tokens for the NullApt CLI.",
  robots: { index: false },
};

export default async function TokensPage() {
  const user = await getCurrentUser();
  if (!user) redirect("/login?next=/settings/tokens");

  let tokens: APITokenMeta[] = [];
  try {
    tokens = (await authedFetch<APITokenMeta[] | null>("/v1/tokens")) ?? [];
  } catch {
    // ignore — render empty state
  }

  return (
    <div className="max-w-3xl mx-auto px-6 py-12">
      <div className="text-xs mb-3" style={{ color: "var(--muted)" }}>
        <a href="/" className="hover:text-white transition-colors">~/nullapt</a>
        <span className="mx-2">/</span>
        <a href="/settings/tokens" style={{ color: "var(--accent)" }} className="hover:underline">
          settings
        </a>
      </div>

      <div className="flex items-center gap-3 mb-2">
        {user.avatar_url && (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={user.avatar_url}
            alt={user.username}
            width={28}
            height={28}
            style={{ borderRadius: "50%", border: "1px solid var(--border)" }}
          />
        )}
        <h1 className="text-2xl font-bold">API tokens</h1>
      </div>
      <p style={{ color: "var(--muted)" }} className="text-sm mb-8">
        Signed in as <span style={{ color: "var(--accent)" }}>@{user.username}</span>. Tokens
        authenticate the CLI when you publish skills.
      </p>

      {/* Create */}
      <section className="mb-10">
        <h2 className="text-xs font-semibold mb-3" style={{ color: "var(--muted)" }}>
          GENERATE NEW TOKEN
        </h2>
        <CreateTokenForm />
      </section>

      {/* List */}
      <section>
        <h2 className="text-xs font-semibold mb-3" style={{ color: "var(--muted)" }}>
          ACTIVE TOKENS ({tokens.length})
        </h2>
        {tokens.length === 0 ? (
          <p className="text-sm" style={{ color: "var(--muted)" }}>
            No tokens yet. Generate one above to publish from the CLI.
          </p>
        ) : (
          <div
            style={{ border: "1px solid var(--border)", background: "var(--surface)" }}
            className="rounded overflow-hidden"
          >
            {tokens.map((t, i) => (
              <div
                key={t.id}
                className="px-4 py-3 flex items-center justify-between gap-4"
                style={{ borderTop: i > 0 ? "1px solid var(--border)" : undefined }}
              >
                <div className="min-w-0">
                  <div className="text-sm font-semibold truncate">{t.name}</div>
                  <div className="text-xs mt-0.5" style={{ color: "var(--muted)" }}>
                    Created {new Date(t.created_at).toISOString().slice(0, 10)}
                    {t.last_used && ` · last used ${new Date(t.last_used).toISOString().slice(0, 10)}`}
                    {!t.last_used && " · never used"}
                  </div>
                </div>
                <RevokeTokenButton id={t.id} name={t.name} />
              </div>
            ))}
          </div>
        )}
      </section>

      {/* CLI usage */}
      <section className="mt-10">
        <h2 className="text-xs font-semibold mb-3" style={{ color: "var(--muted)" }}>
          USE FROM THE CLI
        </h2>
        <div
          style={{ background: "var(--surface)", border: "1px solid var(--border)" }}
          className="rounded p-4 text-sm font-mono"
        >
          <div style={{ color: "var(--muted)" }}>$ nullapt login</div>
          <div style={{ color: "var(--muted)" }} className="mt-1">
            Enter your NullApt API token: <span style={{ color: "var(--accent)" }}>nlpt_…</span>
          </div>
        </div>
        <p style={{ color: "var(--muted)" }} className="text-xs mt-3">
          Paste the token shown when you generated it. Tokens are stored in{" "}
          <code style={{ color: "var(--accent)" }}>~/.nullapt_token</code> on your machine.
        </p>
      </section>
    </div>
  );
}
