"use client";

import { useState } from "react";
import { approveCLIAuth } from "./actions";

export default function ApproveButton({ requestId }: { requestId: string }) {
  const [state, setState] = useState<"idle" | "loading" | "done" | "error">("idle");
  const [error, setError] = useState("");

  async function handleApprove() {
    setState("loading");
    const result = await approveCLIAuth(requestId);
    if (result?.error) {
      setError(result.error);
      setState("error");
    } else {
      setState("done");
    }
  }

  if (state === "done") {
    return (
      <div
        style={{ background: "var(--surface)", border: "1px solid var(--accent)", color: "var(--accent)" }}
        className="rounded-lg p-4 text-sm text-center"
      >
        ✓ Approved — you can close this tab. Your terminal should continue automatically.
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <button
        onClick={handleApprove}
        disabled={state === "loading"}
        style={{ background: "var(--accent)", color: "#000" }}
        className="w-full rounded px-4 py-3 text-sm font-semibold hover:opacity-90 transition-opacity disabled:opacity-50"
      >
        {state === "loading" ? "Authorizing…" : "Authorize CLI"}
      </button>
      <a
        href="/"
        className="block text-center text-xs hover:text-white transition-colors"
        style={{ color: "var(--muted)" }}
      >
        Cancel
      </a>
      {state === "error" && (
        <p className="text-xs text-center" style={{ color: "var(--warning)" }}>
          {error}
        </p>
      )}
    </div>
  );
}
