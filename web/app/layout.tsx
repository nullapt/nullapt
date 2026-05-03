import type { Metadata, Viewport } from "next";
import { Suspense } from "react";
import "./globals.css";
import { safeJsonLd } from "@/lib/jsonld";
import { ThemeToggle } from "@/components/ThemeToggle";
import { UserNav } from "@/components/UserNav";
import { Analytics } from "@vercel/analytics/next";

const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL || "https://nullapt.dev";
const SITE_NAME = "NullApt";
const SITE_DESCRIPTION =
  "The private-first package manager for AI skills. Cryptographically signed, WASM sandboxed, and offline first.";

export const metadata: Metadata = {
  metadataBase: new URL(SITE_URL),
  title: {
    default: `${SITE_NAME} — Private-First AI Skill Registry`,
    template: `%s · ${SITE_NAME}`,
  },
  description: SITE_DESCRIPTION,
  keywords: [
    "AI skills",
    "MCP",
    "Model Context Protocol",
    "local LLM",
    "package manager",
    "WASM sandbox",
    "Ed25519",
    "offline AI",
    "LM Studio",
    "AnythingLLM",
    "Ollama tools",
    "privacy AI",
    "nullapt",
  ],
  authors: [{ name: "NullApt Contributors", url: "https://github.com/nullapt/nullapt" }],
  creator: "NullApt",
  publisher: "NullApt",
  robots: {
    index: true,
    follow: true,
    googleBot: { index: true, follow: true, "max-image-preview": "large" },
  },
  openGraph: {
    type: "website",
    locale: "en_US",
    url: SITE_URL,
    siteName: SITE_NAME,
    title: `${SITE_NAME} — Private-First AI Skill Registry`,
    description: SITE_DESCRIPTION,
    // /opengraph-image is served by app/opengraph-image.tsx automatically
  },
  twitter: {
    card: "summary_large_image",
    title: `${SITE_NAME} — Private-First AI Skill Registry`,
    description: SITE_DESCRIPTION,
    // twitter image auto-resolved from /opengraph-image
    creator: "@nullapt",
    site: "@nullapt",
  },
  alternates: {
    canonical: SITE_URL,
  },
  manifest: "/site.webmanifest",
  // icon and apple-icon are served by app/icon.tsx and app/apple-icon.tsx
};

export const viewport: Viewport = {
  themeColor: [
    { media: "(prefers-color-scheme: dark)", color: "#4ade80" },
    { media: "(prefers-color-scheme: light)", color: "#16a34a" },
  ],
};

// Runs synchronously before paint to set data-theme from localStorage or
// system preference. Prevents the flash of dark when a light user loads.
const themeInitScript = `
(function(){try{
  var s=localStorage.getItem('nullapt-theme');
  var t=s||(window.matchMedia('(prefers-color-scheme: light)').matches?'light':'dark');
  document.documentElement.setAttribute('data-theme',t);
}catch(e){document.documentElement.setAttribute('data-theme','dark');}})();
`;

export default async function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className="h-full" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeInitScript }} />
        {/* JSON-LD: SoftwareApplication */}
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: safeJsonLd({
              "@context": "https://schema.org",
              "@type": "SoftwareApplication",
              name: "NullApt",
              applicationCategory: "DeveloperApplication",
              operatingSystem: "macOS, Linux, Windows",
              url: SITE_URL,
              description: SITE_DESCRIPTION,
              author: {
                "@type": "Organization",
                name: "NullApt",
                url: SITE_URL,
              },
              license: "https://opensource.org/licenses/MIT",
              codeRepository: "https://github.com/nullapt/nullapt",
            }),
          }}
        />
      </head>
      <body className="min-h-full flex flex-col">
        <header
          style={{ background: "var(--surface)", borderBottom: "1px solid var(--border)" }}
          className="sticky top-0 z-50"
        >
          <div className="max-w-6xl mx-auto px-6 h-12 flex items-center justify-between">
            <a href="/" style={{ color: "var(--accent)" }} className="font-bold text-base tracking-tight">
              <span style={{ color: "var(--muted)" }}>~/</span>nullapt
            </a>
            <nav className="flex items-center gap-6 text-sm" style={{ color: "var(--muted)" }}>
              <a href="/skills" className="hover:text-white transition-colors">browse</a>
              <a href="/releases" className="hover:text-white transition-colors">releases</a>
              <a href="/docs" className="hover:text-white transition-colors">docs</a>
              <a
                href="https://github.com/nullapt/nullapt"
                target="_blank"
                rel="noreferrer"
                className="hover:text-white transition-colors"
              >
                github
              </a>
              <ThemeToggle />
              <Suspense fallback={null}>
                <UserNav />
              </Suspense>
            </nav>
          </div>
        </header>

        <main className="flex-1">{children}</main>

        <footer
          style={{ borderTop: "1px solid var(--border)", color: "var(--muted)" }}
          className="text-xs py-6"
        >
          <div className="max-w-6xl mx-auto px-6 flex justify-between">
            <span>
              <span style={{ color: "var(--accent)" }}>~/</span>nullapt — zero-knowledge AI skill registry
            </span>
            <span>
              <a
                href="https://github.com/nullapt/nullapt"
                className="hover:text-white transition-colors"
                target="_blank"
                rel="noreferrer"
              >
                open source
              </a>
              {" · "}offline first · Ed25519 signed
            </span>
          </div>
        </footer>
        <Analytics />
      </body>
    </html>
  );
}
