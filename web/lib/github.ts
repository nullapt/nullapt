import { cacheLife, cacheTag } from "next/cache";

const REPO = "nullapt/nullapt";

export const GITHUB_TAG = "github";

export interface GitHubRepo {
  stargazers_count: number;
  forks_count: number;
}

export interface GitHubRelease {
  id: number;
  tag_name: string;
  name: string;
  body: string;
  draft: boolean;
  prerelease: boolean;
  published_at: string;
  html_url: string;
  assets: {
    id: number;
    name: string;
    download_count: number;
    size: number;
    browser_download_url: string;
  }[];
}

async function ghFetch<T>(path: string): Promise<T> {
  const headers: HeadersInit = { Accept: "application/vnd.github+json" };
  if (process.env.GITHUB_TOKEN) {
    headers["Authorization"] = `Bearer ${process.env.GITHUB_TOKEN}`;
  }
  const res = await fetch(`https://api.github.com/repos/${REPO}${path}`, { headers });
  if (!res.ok) throw new Error(`GitHub API ${res.status} for ${path}`);
  return res.json() as Promise<T>;
}

export async function getRepo(): Promise<GitHubRepo> {
  "use cache";
  cacheTag(GITHUB_TAG);
  cacheLife("hours");
  return ghFetch<GitHubRepo>("");
}

export async function getReleases(): Promise<GitHubRelease[]> {
  "use cache";
  cacheTag(GITHUB_TAG);
  cacheLife("hours");
  return ghFetch<GitHubRelease[]>("/releases?per_page=20");
}
