import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { getCurrentUser } from "@/lib/auth";
import ApproveButton from "./ApproveButton";

export const metadata: Metadata = {
  title: "Authorize CLI — NullApt",
  robots: { index: false },
};

export default async function CLIAuthPage({
  searchParams,
}: {
  searchParams: Promise<{ id?: string }>;
}) {
  const { id } = await searchParams;

  if (!id) {
    redirect("/");
  }

  const user = await getCurrentUser();
  if (!user) {
    redirect(`/login?next=${encodeURIComponent(`/cli-auth?id=${id}`)}`);
  }

  return (
    <div className="max-w-md mx-auto px-6 py-24">
      <div className="text-xs mb-4" style={{ color: "var(--muted)" }}>
        <a href="/" className="hover:text-white transition-colors">~/nullapt</a>
        <span className="mx-2">/</span>
        <span style={{ color: "var(--accent)" }}>cli-auth</span>
      </div>

      <h1 className="text-2xl font-bold mb-2">Authorize CLI</h1>
      <p style={{ color: "var(--muted)" }} className="text-sm mb-8">
        A NullApt CLI session is requesting access to your account.
      </p>

      <div
        style={{ background: "var(--surface)", border: "1px solid var(--border)" }}
        className="rounded-lg p-5 mb-6"
      >
        <div className="flex items-center gap-3 mb-4">
          {user.avatar_url && (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={user.avatar_url}
              alt={user.username}
              width={36}
              height={36}
              style={{ borderRadius: "50%" }}
            />
          )}
          <div>
            <div className="text-sm font-semibold" style={{ color: "var(--accent)" }}>
              @{user.username}
            </div>
            <div className="text-xs" style={{ color: "var(--muted)" }}>
              {user.email}
            </div>
          </div>
        </div>

        <div className="text-xs space-y-1" style={{ color: "var(--muted)" }}>
          <p>Approving will create a new API token named <strong style={{ color: "var(--fg)" }}>&ldquo;CLI (device login)&rdquo;</strong> and send it to your terminal.</p>
          <p className="pt-1">You can revoke it anytime from <a href="/settings/tokens" className="hover:text-white transition-colors underline">Settings → Tokens</a>.</p>
        </div>
      </div>

      <ApproveButton requestId={id} />

      <p className="text-xs mt-6" style={{ color: "var(--muted)" }}>
        If you did not run <code style={{ color: "var(--accent)" }}>nullapt login</code>, close this page and do not approve.
      </p>
    </div>
  );
}
