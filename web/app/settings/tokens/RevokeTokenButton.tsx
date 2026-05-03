"use client";

import { useTransition } from "react";
import { revokeToken } from "./actions";

export function RevokeTokenButton({ id, name }: { id: string; name: string }) {
  const [pending, startTransition] = useTransition();

  function onClick() {
    if (!confirm(`Revoke token "${name}"? This cannot be undone.`)) return;
    const formData = new FormData();
    formData.set("id", id);
    startTransition(async () => {
      await revokeToken(formData);
    });
  }

  return (
    <button
      type="button"
      onClick={onClick}
      disabled={pending}
      style={{ border: "1px solid var(--border)", color: "var(--danger)" }}
      className="rounded px-3 py-1.5 text-xs hover:border-red-500 transition-colors disabled:opacity-50 shrink-0"
    >
      {pending ? "revoking…" : "revoke"}
    </button>
  );
}
