import type { Metadata, Viewport } from "next";
import { Suspense } from "react";
import Link from "next/link";
import "./globals.css";
import { safeJsonLd } from "@/lib/jsonld";
import { ThemeToggle } from "@/components/ThemeToggle";
import { UserNav } from "@/components/UserNav";
import { Analytics } from "@vercel/analytics/next";

const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL || "https://nullapt.dev";
const SITE_NAME = "NullApt";
const SITE_TITLE = "AI Skills Package Manager for MCP — NullApt";
const SITE_DESCRIPTION =
  "NullApt is the package manager for MCP and AI skills. Every skill is cryptographically signed, WASM sandboxed, and offline-first. Works with Claude, Cursor, Ollama, and LM Studio.";

export const metadata: Metadata = {
  metadataBase: new URL(SITE_URL),
  title: {
    default: SITE_TITLE,
    template: `%s · ${SITE_NAME}`,
  },
  description: SITE_DESCRIPTION,
  keywords: [
    "MCP package manager",
    "MCP server manager",
    "AI skills package manager",
    "secure MCP server installation",
    "MCP supply chain security",
    "signed MCP tools",
    "Claude Desktop tools",
    "Cursor MCP skills",
    "LM Studio skills",
    "Ollama tools",
    "AI tool installation",
    "MCP security",
    "WASM AI sandbox",
    "trusted MCP servers",
    "private AI skill registry",
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
    title: SITE_TITLE,
    description: SITE_DESCRIPTION,
    // /opengraph-image is served by app/opengraph-image.tsx automatically
  },
  twitter: {
    card: "summary_large_image",
    title: SITE_TITLE,
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
            <Link href="/" style={{ color: "var(--accent)" }} className="font-bold text-base tracking-tight">
              <span style={{ color: "var(--muted)" }}>~/</span>nullapt
            </Link>
            <nav className="flex items-center gap-6 text-sm" style={{ color: "var(--muted)" }}>
              <Link href="/skills" className="hover:text-white transition-colors">browse</Link>
              <Link href="/releases" className="hover:text-white transition-colors">releases</Link>
              <Link href="/docs" className="hover:text-white transition-colors">docs</Link>
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
          <div className="max-w-6xl mx-auto px-6 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <span>
              <span style={{ color: "var(--accent)" }}>~/</span>nullapt — secure package manager for MCP skills
            </span>
            <nav className="flex flex-wrap items-center gap-x-4 gap-y-2">
              <Link href="/skills" className="hover:text-white transition-colors">
                browse skills
              </Link>
              <Link href="/docs" className="hover:text-white transition-colors">
                docs
              </Link>
              <Link href="/releases" className="hover:text-white transition-colors">
                releases
              </Link>
              <a
                href="https://github.com/nullapt/nullapt"
                className="hover:text-white transition-colors"
                target="_blank"
                rel="noreferrer"
              >
                github
              </a>
              <span aria-hidden="true">·</span>
              <span>offline first · Ed25519 signed</span>
            </nav>
          </div>
        </footer>
        <Analytics />
      </body>
    </html>
  );
}
