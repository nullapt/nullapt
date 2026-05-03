"use server";

import { authedFetch } from "@/lib/auth";

export async function approveCLIAuth(requestId: string): Promise<{ error?: string } | void> {
  try {
    await authedFetch("/v1/auth/cli/approve", {
      method: "POST",
      body: JSON.stringify({ id: requestId }),
    });
  } catch (e) {
    return { error: e instanceof Error ? e.message : "Failed to approve. The request may have expired." };
  }
}
