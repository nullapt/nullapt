"use client";

import { useState, useTransition } from "react";
import { createToken } from "./actions";

export function CreateTokenForm() {
  const [pending, startTransition] = useTransition();
  const [result, setResult] = useState<{ token: string; name: string } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    const formData = new FormData(e.currentTarget);
    const form = e.currentTarget;
    startTransition(async () => {
      const res = await createToken(formData);
      if ("error" in res) setError(res.error);
      else {
        setResult(res);
        form.reset();
      }
    });
  }

  function copyToken() {
    if (!result) return;
    navigator.clipboard.writeText(result.token);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div className="flex flex-col gap-3">
      <form onSubmit={onSubmit} className="flex gap-2">
        <input
          type="text"
          name="name"
          required
          placeholder='token name (e.g. "laptop")'
          style={{
            background: "var(--surface)",
            border: "1px solid var(--border)",
            color: "var(--foreground)",
          }}
          className="flex-1 rounded px-3 py-2 text-sm outline-none focus:border-green-500 placeholder:text-zinc-600"
        />
        <button
          type="submit"
          disabled={pending}
          style={{ background: "var(--accent)", color: "#000" }}
          className="rounded px-4 py-2 text-sm font-semibold hover:opacity-90 transition-opacity disabled:opacity-50"
        >
          {pending ? "creating…" : "generate"}
        </button>
      </form>

      {error && (
        <div
          style={{ border: "1px solid var(--danger)", color: "var(--danger)" }}
          className="rounded px-3 py-2 text-xs"
        >
          {error}
        </div>
      )}

      {result && (
        <div
          style={{ border: "1px solid var(--accent)", background: "var(--surface)" }}
          className="rounded p-4"
        >
          <div className="text-xs font-semibold mb-2" style={{ color: "var(--accent)" }}>
            ✓ TOKEN CREATED · COPY IT NOW — IT WILL NEVER BE SHOWN AGAIN
          </div>
          <div className="flex items-center gap-2">
            <code
              className="flex-1 px-3 py-2 text-xs break-all rounded"
              style={{ background: "#0c0c0d", color: "var(--accent)" }}
            >
              {result.token}
            </code>
            <button
              type="button"
              onClick={copyToken}
              style={{ border: "1px solid var(--border)", color: "var(--foreground)" }}
              className="rounded px-3 py-2 text-xs hover:border-green-500 transition-colors"
            >
              {copied ? "copied!" : "copy"}
            </button>
          </div>
          <p className="text-xs mt-3" style={{ color: "var(--muted)" }}>
            Run <code style={{ color: "var(--accent)" }}>nullapt login</code> on your machine and
            paste this when prompted.
          </p>
        </div>
      )}
    </div>
  );
}
