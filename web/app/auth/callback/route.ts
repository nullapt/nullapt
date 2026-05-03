import { NextRequest, NextResponse } from "next/server";
import { SESSION_COOKIE } from "@/lib/auth";

const REGISTRY_URL = process.env.NEXT_PUBLIC_REGISTRY_URL || "https://registry.nullapt.dev";
const GITHUB_CLIENT_ID = process.env.GITHUB_CLIENT_ID || process.env.NEXT_PUBLIC_GITHUB_CLIENT_ID || "";
const GITHUB_CLIENT_SECRET = process.env.GITHUB_CLIENT_SECRET || "";

export async function GET(req: NextRequest) {
  const url = new URL(req.url);
  const code = url.searchParams.get("code");
  const state = url.searchParams.get("state") ?? "/settings/tokens";
  const error = url.searchParams.get("error");

  if (error || !code) {
    return NextResponse.redirect(new URL(`/login?error=${encodeURIComponent(error ?? "missing_code")}`, url));
  }

  // If we have the OAuth secret here, exchange the code ourselves.
  // This avoids the Cloudflare bot challenge that blocks server→registry calls.
  let sessionToken: string | null = null;

  if (GITHUB_CLIENT_ID && GITHUB_CLIENT_SECRET) {
    sessionToken = await exchangeViaWebApp(code, url);
  } else {
    sessionToken = await exchangeViaRegistry(code);
  }

  if (!sessionToken) {
    return NextResponse.redirect(new URL("/login?error=oauth_exchange_failed", url));
  }

  const safeNext = state.startsWith("/") && !state.startsWith("//") ? state : "/settings/tokens";
  const response = NextResponse.redirect(new URL(safeNext, url));
  response.cookies.set(SESSION_COOKIE, sessionToken, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: 30 * 24 * 60 * 60,
  });
  return response;
}

// Exchange code with GitHub directly, then create a session via the registry.
async function exchangeViaWebApp(code: string, reqUrl: URL): Promise<string | null> {
  try {
    // 1. Exchange with GitHub
    const ghRes = await fetch("https://github.com/login/oauth/access_token", {
      method: "POST",
      headers: { "Accept": "application/json", "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({ client_id: GITHUB_CLIENT_ID, client_secret: GITHUB_CLIENT_SECRET, code }),
    });
    const ghData = await ghRes.json() as { access_token?: string };
    if (!ghData.access_token) return null;

    // 2. Forward to registry with the GitHub access token for user upsert + session creation
    const regRes = await fetch(`${REGISTRY_URL}/v1/auth/github/callback`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ access_token: ghData.access_token }),
      cache: "no-store",
    });
    if (!regRes.ok) {
      console.error("registry callback failed", regRes.status, await regRes.text());
      return null;
    }
    const data = await regRes.json() as { session_token?: string };
    return data.session_token ?? null;
  } catch (e) {
    console.error("exchangeViaWebApp error", e);
    return null;
  }
}

// Fallback: send the raw code to the registry (requires Cloudflare to allow it).
async function exchangeViaRegistry(code: string): Promise<string | null> {
  try {
    const res = await fetch(`${REGISTRY_URL}/v1/auth/github/callback`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ code }),
      cache: "no-store",
    });
    if (!res.ok) return null;
    const data = await res.json() as { session_token?: string };
    return data.session_token ?? null;
  } catch {
    return null;
  }
}
