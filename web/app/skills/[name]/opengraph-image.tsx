import { ImageResponse } from "next/og";
import { getSkill } from "@/lib/api";

export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default async function SkillOGImage({
  params,
}: {
  params: Promise<{ name: string }>;
}) {
  const { name } = await params;

  let skillName = name;
  let description = "";
  let author = "";
  let version = "";
  let networkAllowed = false;
  let toolCount = 0;

  try {
    const skill = await getSkill(name);
    skillName = skill.name;
    description = skill.description;
    author = skill.author;
    version = skill.version;
    // @ts-expect-error - extended fields present in full manifest
    networkAllowed = skill.permissions?.network?.allowed ?? false;
    // @ts-expect-error - extended fields present in full manifest
    toolCount = skill.interface?.tools?.length ?? 0;
  } catch {
    // Fallback to name-only if registry unreachable
  }

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          background: "#0c0c0d",
          fontFamily: "monospace",
          padding: "64px",
          position: "relative",
        }}
      >
        {/* Grid texture */}
        <div
          style={{
            position: "absolute",
            inset: 0,
            backgroundImage:
              "linear-gradient(rgba(74,222,128,0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(74,222,128,0.03) 1px, transparent 1px)",
            backgroundSize: "60px 60px",
          }}
        />

        {/* Breadcrumb */}
        <div style={{ display: "flex", alignItems: "center", gap: "10px", marginBottom: "40px" }}>
          <span style={{ color: "#52525b", fontSize: "20px" }}>~/nullapt</span>
          <span style={{ color: "#1e1e22", fontSize: "20px" }}>/</span>
          <span style={{ color: "#4ade80", fontSize: "20px" }}>{skillName}</span>
        </div>

        {/* Skill name */}
        <div
          style={{
            color: "#e4e4e7",
            fontSize: "80px",
            fontWeight: 700,
            lineHeight: 1,
            marginBottom: "20px",
          }}
        >
          {skillName}
        </div>

        {/* Description */}
        {description && (
          <div
            style={{
              color: "#71717a",
              fontSize: "26px",
              lineHeight: 1.4,
              marginBottom: "48px",
              maxWidth: "800px",
            }}
          >
            {description.length > 90 ? description.slice(0, 90) + "…" : description}
          </div>
        )}

        {/* Meta row */}
        <div style={{ display: "flex", gap: "20px" }}>
          {author && (
            <div
              style={{
                display: "flex",
                padding: "8px 18px",
                border: "1px solid #1e1e22",
                borderRadius: "6px",
                color: "#71717a",
                fontSize: "18px",
                background: "#111113",
              }}
            >
              by {author}
            </div>
          )}
          {version && (
            <div
              style={{
                display: "flex",
                padding: "8px 18px",
                border: "1px solid #1e1e22",
                borderRadius: "6px",
                color: "#71717a",
                fontSize: "18px",
                background: "#111113",
              }}
            >
              v{version}
            </div>
          )}
          {toolCount > 0 && (
            <div
              style={{
                display: "flex",
                padding: "8px 18px",
                border: "1px solid #1e1e22",
                borderRadius: "6px",
                color: "#71717a",
                fontSize: "18px",
                background: "#111113",
              }}
            >
              {toolCount} {toolCount === 1 ? "tool" : "tools"}
            </div>
          )}
          <div
            style={{
              display: "flex",
              padding: "8px 18px",
              border: `1px solid ${networkAllowed ? "#451a03" : "#1e1e22"}`,
              borderRadius: "6px",
              color: networkAllowed ? "#fbbf24" : "#4ade80",
              fontSize: "18px",
              background: networkAllowed ? "#1c1008" : "#111113",
            }}
          >
            {networkAllowed ? "network access" : "offline-only"}
          </div>
        </div>

        {/* Bottom-right: install command */}
        <div
          style={{
            position: "absolute",
            bottom: "64px",
            right: "64px",
            display: "flex",
            alignItems: "center",
            gap: "14px",
            padding: "16px 28px",
            background: "#111113",
            border: "1px solid #1e1e22",
            borderRadius: "8px",
          }}
        >
          <span style={{ color: "#52525b", fontSize: "20px" }}>$</span>
          <span style={{ color: "#4ade80", fontSize: "20px" }}>
            nullapt get {skillName}
          </span>
        </div>
      </div>
    ),
    { ...size }
  );
}
