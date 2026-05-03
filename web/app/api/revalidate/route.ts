import { revalidateTag } from "next/cache";
import { SKILLS_TAG, skillTag } from "@/lib/api";

const SECRET = process.env.REVALIDATE_SECRET;

export async function POST(request: Request) {
  if (!SECRET) {
    return Response.json({ error: "REVALIDATE_SECRET not configured" }, { status: 500 });
  }

  const provided = request.headers.get("x-revalidate-secret");
  if (provided !== SECRET) {
    return Response.json({ error: "unauthorized" }, { status: 401 });
  }

  let body: { skill?: string } = {};
  try {
    body = await request.json();
  } catch {
    // Empty body is allowed — invalidate the list tag only.
  }

  const tags = [SKILLS_TAG];
  if (body.skill) tags.push(skillTag(body.skill));

  for (const tag of tags) {
    revalidateTag(tag, "max");
  }

  return Response.json({ revalidated: tags, now: Date.now() });
}
