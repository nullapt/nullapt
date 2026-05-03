import { cookies } from "next/headers";

const REGISTRY_URL = process.env.NEXT_PUBLIC_REGISTRY_URL || "https://registry.nullapt.dev";
export const SESSION_COOKIE = "nullapt_session";

export interface AuthUser {
  id: string;
  username: string;
  email: string;
  name?: string | null;
  avatar_url?: string | null;
}

export interface APITokenMeta {
  id: string;
  name: string;
  created_at: string;
  last_used: string | null;
}

// Server-only: read the session cookie and call /v1/me to validate it.
export async function getCurrentUser(): Promise<AuthUser | null> {
  const cookieStore = await cookies();
  const session = cookieStore.get(SESSION_COOKIE)?.value;
  if (!session) return null;

  try {
    const res = await fetch(`${REGISTRY_URL}/v1/me`, {
      headers: { Authorization: `Bearer ${session}` },
      cache: "no-store",
    });
    if (!res.ok) return null;
    return (await res.json()) as AuthUser;
  } catch {
    return null;
  }
}

export async function getSessionToken(): Promise<string | null> {
  const cookieStore = await cookies();
  return cookieStore.get(SESSION_COOKIE)?.value ?? null;
}

// Server-side fetch with auth. Returns parsed JSON or throws.
export async function authedFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const session = await getSessionToken();
  if (!session) throw new Error("not authenticated");

  const res = await fetch(`${REGISTRY_URL}${path}`, {
    ...init,
    headers: {
      ...(init?.headers ?? {}),
      Authorization: `Bearer ${session}`,
      "Content-Type": "application/json",
    },
    cache: "no-store",
  });
  if (!res.ok) throw new Error(`${res.status}: ${await res.text()}`);
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}
