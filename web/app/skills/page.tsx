import type { Metadata } from "next";
import { redirect } from "next/navigation";

export const metadata: Metadata = {
  title: "Browse AI Skills — NullApt",
  description:
    "Browse the NullApt registry — cryptographically signed, WASM-sandboxed AI skills for local LLMs. Filter by network access, license, and author.",
  alternates: { canonical: "https://nullapt.dev/skills" },
  openGraph: {
    title: "Browse AI Skills — NullApt",
    description: "The private-first registry for audited, signed AI skills.",
    url: "https://nullapt.dev/skills",
  },
};

// /skills is the canonical browse URL — the home page handles rendering.
export default function SkillsBrowsePage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string }>;
}) {
  // Redirect to home with query intact so the home page renders the results.
  // This gives /skills a canonical URL that search engines index separately
  // while keeping a single implementation of the browse UI.
  void searchParams;
  redirect("/");
}
