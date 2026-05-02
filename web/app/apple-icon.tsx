import { ImageResponse } from "next/og";

export const size = { width: 180, height: 180 };
export const contentType = "image/png";

export default function AppleIcon() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: "#0c0c0d",
          fontFamily: "monospace",
        }}
      >
        <div style={{ display: "flex", alignItems: "baseline" }}>
          <span style={{ color: "#52525b", fontSize: "28px", fontWeight: 700, lineHeight: 1 }}>
            ~/
          </span>
          <span style={{ color: "#4ade80", fontSize: "28px", fontWeight: 700, lineHeight: 1 }}>
            nullapt
          </span>
        </div>
      </div>
    ),
    { ...size }
  );
}
