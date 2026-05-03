import { revalidateTag } from "next/cache";
import { SKILLS_TAG, skillTag } from "@/lib/api";
import { GITHUB_TAG } from "@/lib/github";

const SECRET = process.env.REVALIDATE_SECRET;

type Target = "skills" | "github";

export async function POST(request: Request) {
  if (!SECRET) {
    return Response.json({ error: "REVALIDATE_SECRET not configured" }, { status: 500 });
  }

  const provided = request.headers.get("x-revalidate-secret");
  if (provided !== SECRET) {
    return Response.json({ error: "unauthorized" }, { status: 401 });
  }

  let body: { target?: Target; skill?: string } = {};
  try {
    body = await request.json();
  } catch {
    // Empty body is allowed — defaults to invalidating the skills list tag.
  }

  const target: Target = body.target ?? "skills";
  const tags: string[] = [];

  if (target === "github") {
    tags.push(GITHUB_TAG);
  } else {
    tags.push(SKILLS_TAG);
    if (body.skill) tags.push(skillTag(body.skill));
  }

  for (const tag of tags) {
    revalidateTag(tag, "max");
  }

  return Response.json({ revalidated: tags, now: Date.now() });
}
