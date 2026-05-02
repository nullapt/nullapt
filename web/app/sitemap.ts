import type { MetadataRoute } from "next";
import { listSkills } from "@/lib/api";

const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL ?? "https://nullapt.dev";

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const staticRoutes: MetadataRoute.Sitemap = [
    {
      url: SITE_URL,
      lastModified: new Date(),
      changeFrequency: "daily",
      priority: 1,
    },
    {
      url: `${SITE_URL}/skills`,
      lastModified: new Date(),
      changeFrequency: "hourly",
      priority: 0.9,
    },
  ];

  let skillRoutes: MetadataRoute.Sitemap = [];
  try {
    const skills = await listSkills();
    skillRoutes = skills.map((skill) => ({
      url: `${SITE_URL}/skills/${skill.name}`,
      lastModified: new Date(skill.published_at),
      changeFrequency: "weekly" as const,
      priority: 0.7,
    }));
  } catch {
    // Registry unreachable at build time — static routes still generated
  }

  return [...staticRoutes, ...skillRoutes];
}
