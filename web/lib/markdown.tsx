import type { ReactNode } from "react";

// Minimal markdown renderer for GitHub release bodies.
// Handles: ## headings, * lists, ```code blocks```, `inline code`, **bold**, plain text.
export function ReleaseBody({ body }: { body: string }) {
  const nodes: ReactNode[] = [];
  const lines = body.split("\n");
  let i = 0;
  let key = 0;

  while (i < lines.length) {
    const line = lines[i];

    // ── fenced code block ──────────────────────────────────────────
    if (line.trimStart().startsWith("```")) {
      const fence: string[] = [];
      i++;
      while (i < lines.length && !lines[i].trimStart().startsWith("```")) {
        fence.push(lines[i]);
        i++;
      }
      nodes.push(
        <pre
          key={key++}
          style={{ background: "#0c0c0d", border: "1px solid #1e1e22", color: "#4ade80" }}
          className="rounded p-3 text-xs overflow-x-auto my-2"
        >
          {fence.join("\n")}
        </pre>
      );
      i++;
      continue;
    }

    // ── h3 ─────────────────────────────────────────────────���──────
    if (line.startsWith("### ")) {
      nodes.push(
        <p key={key++} className="text-xs font-semibold mt-4 mb-1" style={{ color: "#e4e4e7" }}>
          {line.slice(4)}
        </p>
      );
      i++;
      continue;
    }

    // ── h2 ────────────────────────────────────────────────────────
    if (line.startsWith("## ")) {
      nodes.push(
        <p key={key++} className="text-sm font-bold mt-4 mb-2" style={{ color: "#e4e4e7" }}>
          {line.slice(3)}
        </p>
      );
      i++;
      continue;
    }

    // ── list item ─────────────────────────────────────────────────
    if (/^[*\-] /.test(line)) {
      // Collect contiguous list items
      const items: string[] = [];
      while (i < lines.length && /^[*\-] /.test(lines[i])) {
        items.push(lines[i].slice(2));
        i++;
      }
      nodes.push(
        <ul key={key++} className="my-2 flex flex-col gap-1">
          {items.map((item, idx) => (
            <li key={idx} className="flex gap-2 text-xs" style={{ color: "#71717a" }}>
              <span style={{ color: "#4ade80" }}>·</span>
              <span>{inlineFormat(item)}</span>
            </li>
          ))}
        </ul>
      );
      continue;
    }

    // ── blank line ────────────────────────────────────────────────
    if (line.trim() === "") {
      i++;
      continue;
    }

    // ── paragraph ─────────────────────────────────────────────────
    nodes.push(
      <p key={key++} className="text-xs my-1" style={{ color: "#71717a" }}>
        {inlineFormat(line)}
      </p>
    );
    i++;
  }

  return <div>{nodes}</div>;
}

// Handles `inline code` and **bold** within a string.
function inlineFormat(text: string): ReactNode {
  const parts: ReactNode[] = [];
  const re = /(`[^`]+`|\*\*[^*]+\*\*)/g;
  let last = 0;
  let match;
  let k = 0;

  while ((match = re.exec(text)) !== null) {
    if (match.index > last) parts.push(text.slice(last, match.index));
    const token = match[0];
    if (token.startsWith("`")) {
      parts.push(
        <code key={k++} style={{ color: "#4ade80", background: "#111113" }} className="px-1 rounded">
          {token.slice(1, -1)}
        </code>
      );
    } else {
      parts.push(
        <strong key={k++} style={{ color: "#e4e4e7" }}>
          {token.slice(2, -2)}
        </strong>
      );
    }
    last = match.index + token.length;
  }

  if (last < text.length) parts.push(text.slice(last));
  return parts.length === 1 && typeof parts[0] === "string" ? parts[0] : <>{parts}</>;
}
