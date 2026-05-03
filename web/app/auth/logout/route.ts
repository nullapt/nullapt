import { NextRequest, NextResponse } from "next/server";
import { SESSION_COOKIE } from "@/lib/auth";

const REGISTRY_URL = process.env.NEXT_PUBLIC_REGISTRY_URL || "https://registry.nullapt.dev";

export async function POST(req: NextRequest) {
  const session = req.cookies.get(SESSION_COOKIE)?.value;

  if (session) {
    // Best-effort server-side revocation
    try {
      await fetch(`${REGISTRY_URL}/v1/auth/logout`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${session}` },
      });
    } catch {
      // ignore network errors — we still clear the cookie below
    }
  }

  const response = NextResponse.redirect(new URL("/", req.url));
  response.cookies.delete(SESSION_COOKIE);
  return response;
}

export const GET = POST;
