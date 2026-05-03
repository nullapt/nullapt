import { cacheTag } from "next/cache";

const REGISTRY_URL = process.env.NEXT_PUBLIC_REGISTRY_URL ?? "https://registry.nullapt.dev";

export const SKILLS_TAG = "skills";
export const skillTag = (name: string) => `skill:${name}`;

export interface SkillMeta {
  name: string;
  version: string;
  description: string;
  author: string;
  license: string;
  downloads: number;
  manifest_url: string;
  wasm_url: string;
  published_at: string;
}

export interface SkillManifest {
  schema_version: string;
  name: string;
  version: string;
  description: string;
  author: string;
  homepage?: string;
  license: string;
  permissions: {
    network: { allowed: boolean; domains?: string[] };
    filesystem: { read: string[]; write: string[] };
    env: string[];
  };
  entry: string;
  interface: {
    tools: { name: string; description: string; input_schema: Record<string, unknown> }[];
  };
  signature: { algorithm: string; public_key: string; value: string };
}

export interface TransparencyEntry {
  Version: string;
  Author: string;
  PublicKeyB64: string;
  ManifestHash: string;
  RecordedAt: string;
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${REGISTRY_URL}${path}`, { cache: "no-store" });
  if (!res.ok) throw new Error(`Registry returned ${res.status} for ${path}`);
  return res.json() as Promise<T>;
}

export async function listSkills(q = ""): Promise<SkillMeta[]> {
  "use cache";
  cacheTag(SKILLS_TAG);
  return get<SkillMeta[]>(`/v1/skills${q ? `?q=${encodeURIComponent(q)}` : ""}`);
}

export async function getSkill(name: string, version?: string): Promise<SkillMeta> {
  "use cache";
  cacheTag(SKILLS_TAG, skillTag(name));
  return get<SkillMeta>(`/v1/skills/${name}${version ? `@${version}` : ""}`);
}

export async function getTransparencyLog(name: string): Promise<TransparencyEntry[]> {
  "use cache";
  cacheTag(SKILLS_TAG, skillTag(name));
  return get<TransparencyEntry[]>(`/v1/skills/${name}/log`);
}
