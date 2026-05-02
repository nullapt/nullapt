import { ImageResponse } from "next/og";

export const size = { width: 512, height: 512 };
export const contentType = "image/png";

export default function Icon512() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          background: "#0c0c0d",
          fontFamily: "monospace",
          position: "relative",
        }}
      >
        {/* Grid texture */}
        <div
          style={{
            position: "absolute",
            inset: 0,
            backgroundImage:
              "linear-gradient(rgba(74,222,128,0.04) 1px, transparent 1px), linear-gradient(90deg, rgba(74,222,128,0.04) 1px, transparent 1px)",
            backgroundSize: "48px 48px",
          }}
        />

        {/* Wordmark */}
        <div style={{ display: "flex", alignItems: "baseline" }}>
          <span style={{ color: "#52525b", fontSize: "80px", fontWeight: 700, lineHeight: 1 }}>
            ~/
          </span>
          <span style={{ color: "#4ade80", fontSize: "80px", fontWeight: 700, lineHeight: 1 }}>
            nullapt
          </span>
        </div>

        {/* Tagline */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "8px",
            marginTop: "20px",
          }}
        >
          <div
            style={{
              width: "7px",
              height: "7px",
              borderRadius: "50%",
              background: "#4ade80",
            }}
          />
          <span style={{ color: "#4ade80", fontSize: "16px", letterSpacing: "0.15em" }}>
            PRIVATE-FIRST AI SKILL REGISTRY
          </span>
        </div>
      </div>
    ),
    { ...size }
  );
}
