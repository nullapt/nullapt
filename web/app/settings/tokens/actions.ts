"use server";

import { revalidatePath } from "next/cache";
import { authedFetch } from "@/lib/auth";

export async function createToken(formData: FormData): Promise<{ token: string; name: string } | { error: string }> {
  const name = String(formData.get("name") ?? "").trim();
  if (!name) return { error: "Name is required" };

  try {
    const result = await authedFetch<{ id: string; name: string; token: string }>("/v1/tokens", {
      method: "POST",
      body: JSON.stringify({ name }),
    });
    revalidatePath("/settings/tokens");
    return { token: result.token, name: result.name };
  } catch (e) {
    return { error: (e as Error).message };
  }
}

export async function revokeToken(formData: FormData): Promise<{ ok: boolean; error?: string }> {
  const id = String(formData.get("id") ?? "");
  if (!id) return { ok: false, error: "Missing token id" };

  try {
    await authedFetch(`/v1/tokens/${encodeURIComponent(id)}`, { method: "DELETE" });
    revalidatePath("/settings/tokens");
    return { ok: true };
  } catch (e) {
    return { ok: false, error: (e as Error).message };
  }
}
