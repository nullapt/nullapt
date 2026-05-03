import { NextRequest, NextResponse } from "next/server";
import { SESSION_COOKIE } from "@/lib/auth";

const REGISTRY_URL = process.env.NEXT_PUBLIC_REGISTRY_URL || "https://registry.nullapt.dev";

export async function GET(req: NextRequest) {
  const url = new URL(req.url);
  const code = url.searchParams.get("code");
  const state = url.searchParams.get("state") ?? "/settings/tokens";
  const error = url.searchParams.get("error");

  if (error || !code) {
    return NextResponse.redirect(new URL(`/login?error=${encodeURIComponent(error ?? "missing_code")}`, url));
  }

  // Exchange the code via the registry API
  const res = await fetch(`${REGISTRY_URL}/v1/auth/github/callback`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ code }),
    cache: "no-store",
  });

  if (!res.ok) {
    const body = await res.text();
    console.error("OAuth callback failed", res.status, body);
    return NextResponse.redirect(new URL(`/login?error=oauth_exchange_failed`, url));
  }

  const data = (await res.json()) as { session_token: string };

  if (!data.session_token) {
    console.error("OAuth callback: backend returned no session_token");
    return NextResponse.redirect(new URL(`/login?error=oauth_exchange_failed`, url));
  }

  // Validate the redirect target — must be a relative path on this site
  const safeNext = state.startsWith("/") && !state.startsWith("//") ? state : "/settings/tokens";
  const redirectURL = new URL(safeNext, url);

  const response = NextResponse.redirect(redirectURL);
  response.cookies.set(SESSION_COOKIE, data.session_token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: 30 * 24 * 60 * 60, // 30 days, matches server-side session TTL
  });
  return response;
}
