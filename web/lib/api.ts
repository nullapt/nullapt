const REGISTRY_URL = process.env.NEXT_PUBLIC_REGISTRY_URL ?? "https://registry.nullapt.dev";

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
  const res = await fetch(`${REGISTRY_URL}${path}`, { next: { revalidate: 60 } });
  if (!res.ok) throw new Error(`Registry returned ${res.status} for ${path}`);
  return res.json() as Promise<T>;
}

export const listSkills = (q = "") =>
  get<SkillMeta[]>(`/v1/skills${q ? `?q=${encodeURIComponent(q)}` : ""}`);

export const getSkill = (name: string, version?: string) =>
  get<SkillMeta>(`/v1/skills/${name}${version ? `@${version}` : ""}`);

export const getTransparencyLog = (name: string) =>
  get<TransparencyEntry[]>(`/v1/skills/${name}/log`);
