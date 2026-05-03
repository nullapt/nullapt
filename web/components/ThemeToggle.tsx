"use client";

import { useEffect, useState } from "react";

type Theme = "dark" | "light";

export function ThemeToggle() {
  const [theme, setTheme] = useState<Theme>("dark");

  useEffect(() => {
    const current = (document.documentElement.getAttribute("data-theme") as Theme) || "dark";
    setTheme(current);
  }, []);

  function toggle() {
    const next: Theme = theme === "dark" ? "light" : "dark";
    document.documentElement.setAttribute("data-theme", next);
    try {
      localStorage.setItem("nullapt-theme", next);
    } catch {
      // ignore (private browsing)
    }
    setTheme(next);
  }

  return (
    <button
      type="button"
      onClick={toggle}
      aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} theme`}
      title={`Switch to ${theme === "dark" ? "light" : "dark"} theme`}
      className="text-xs hover:text-white transition-colors"
      style={{ color: "var(--muted)" }}
    >
      {theme === "dark" ? "☀ light" : "☾ dark"}
    </button>
  );
}
