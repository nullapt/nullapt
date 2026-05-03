import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { getCurrentUser } from "@/lib/auth";

export const metadata: Metadata = {
  title: "Sign in — NullApt",
  description: "Sign in to NullApt with GitHub to publish skills.",
  robots: { index: false },
};

const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL || "https://nullapt.dev";
const GITHUB_CLIENT_ID = process.env.NEXT_PUBLIC_GITHUB_CLIENT_ID || "";

const ERROR_MESSAGES: Record<string, string> = {
  oauth_exchange_failed: "GitHub sign-in failed. Please try again.",
  missing_code: "GitHub did not return an authorization code. Please try again.",
  redirect_uri_mismatch: "OAuth redirect URI mismatch — check your GitHub OAuth app settings.",
};

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ next?: string; error?: string }>;
}) {
  const user = await getCurrentUser();
  if (user) redirect("/settings/tokens");

  const { next, error } = await searchParams;
  const state = encodeURIComponent(next ?? "/settings/tokens");
  const redirectURI = encodeURIComponent(`${SITE_URL}/auth/callback`);
  const ghAuthURL = GITHUB_CLIENT_ID
    ? `https://github.com/login/oauth/authorize?client_id=${GITHUB_CLIENT_ID}&redirect_uri=${redirectURI}&scope=read:user%20user:email&state=${state}`
    : null;

  return (
    <div className="max-w-md mx-auto px-6 py-24">
      <div className="text-xs mb-4" style={{ color: "var(--muted)" }}>
        <a href="/" className="hover:text-white transition-colors">~/nullapt</a>
        <span className="mx-2">/</span>
        <span style={{ color: "var(--accent)" }}>login</span>
      </div>

      <h1 className="text-2xl font-bold mb-3">Sign in</h1>
      <p style={{ color: "var(--muted)" }} className="text-sm mb-8">
        NullApt uses GitHub to verify your identity. Your GitHub username becomes your publishing
        handle and is recorded in the transparency log.
      </p>

      {error && (
        <div
          style={{ background: "var(--surface)", border: "1px solid var(--border)", color: "var(--warning)" }}
          className="rounded p-4 text-sm mb-6"
        >
          {ERROR_MESSAGES[error] ?? `Sign-in error: ${error}`}
        </div>
      )}

      {ghAuthURL ? (
        <a
          href={ghAuthURL}
          style={{ background: "var(--accent)", color: "#000" }}
          className="rounded px-4 py-3 text-sm font-semibold hover:opacity-90 transition-opacity flex items-center justify-center gap-3"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden>
            <path d="M12 .5C5.65.5.5 5.65.5 12c0 5.08 3.29 9.39 7.86 10.91.58.11.79-.25.79-.56v-2c-3.2.7-3.87-1.37-3.87-1.37-.52-1.32-1.27-1.67-1.27-1.67-1.04-.71.08-.7.08-.7 1.15.08 1.75 1.18 1.75 1.18 1.02 1.74 2.67 1.24 3.32.95.1-.74.4-1.24.72-1.53-2.55-.29-5.24-1.27-5.24-5.66 0-1.25.45-2.27 1.18-3.07-.12-.29-.51-1.46.11-3.04 0 0 .96-.31 3.15 1.17a10.95 10.95 0 0 1 5.74 0c2.19-1.48 3.15-1.17 3.15-1.17.62 1.58.23 2.75.11 3.04.73.8 1.18 1.82 1.18 3.07 0 4.4-2.69 5.36-5.25 5.65.41.36.78 1.05.78 2.12v3.14c0 .31.21.68.8.56C20.21 21.39 23.5 17.08 23.5 12 23.5 5.65 18.35.5 12 .5z" />
          </svg>
          Continue with GitHub
        </a>
      ) : (
        <div
          style={{ background: "var(--surface)", border: "1px solid var(--border)", color: "var(--warning)" }}
          className="rounded p-4 text-sm"
        >
          GitHub OAuth is not configured. The site administrator needs to set{" "}
          <code>NEXT_PUBLIC_GITHUB_CLIENT_ID</code>.
        </div>
      )}

      <p style={{ color: "var(--muted)" }} className="text-xs mt-8 leading-relaxed">
        We request <code style={{ color: "var(--accent)" }}>read:user</code> and{" "}
        <code style={{ color: "var(--accent)" }}>user:email</code> only — no write access to your
        repos. If you have 2FA enabled on GitHub, you&apos;ll be prompted as part of the standard
        sign-in flow.
      </p>
    </div>
  );
}
