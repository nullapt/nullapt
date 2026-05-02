import { ImageResponse } from "next/og";

export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default function OGImage() {
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
        {/* Grid lines — subtle texture */}
        <div
          style={{
            position: "absolute",
            inset: 0,
            backgroundImage:
              "linear-gradient(rgba(74,222,128,0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(74,222,128,0.03) 1px, transparent 1px)",
            backgroundSize: "60px 60px",
          }}
        />

        {/* Top: registry label */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "10px",
            marginBottom: "48px",
          }}
        >
          <div
            style={{
              width: "10px",
              height: "10px",
              borderRadius: "50%",
              background: "#4ade80",
            }}
          />
          <span style={{ color: "#4ade80", fontSize: "18px", letterSpacing: "0.15em" }}>
            NULLAPT REGISTRY · v1 · LIVE
          </span>
        </div>

        {/* Main logo */}
        <div
          style={{
            display: "flex",
            alignItems: "flex-end",
            gap: "0px",
            marginBottom: "24px",
          }}
        >
          <span style={{ color: "#52525b", fontSize: "96px", fontWeight: 700, lineHeight: 1 }}>
            ~/
          </span>
          <span style={{ color: "#4ade80", fontSize: "96px", fontWeight: 700, lineHeight: 1 }}>
            nullapt
          </span>
        </div>

        {/* Tagline */}
        <div
          style={{
            color: "#a1a1aa",
            fontSize: "28px",
            marginBottom: "56px",
            lineHeight: 1.4,
          }}
        >
          The Private-First Package Manager for AI Skills
        </div>

        {/* Feature pills */}
        <div style={{ display: "flex", gap: "16px" }}>
          {[
            "Ed25519 Signed",
            "WASM Sandboxed",
            "Offline First",
            "Open Source",
          ].map((label) => (
            <div
              key={label}
              style={{
                display: "flex",
                alignItems: "center",
                padding: "8px 18px",
                border: "1px solid #1e1e22",
                borderRadius: "6px",
                color: "#71717a",
                fontSize: "16px",
                background: "#111113",
              }}
            >
              {label}
            </div>
          ))}
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
          <span style={{ color: "#4ade80", fontSize: "20px" }}>nullapt get web-search</span>
        </div>
      </div>
    ),
    { ...size }
  );
}
