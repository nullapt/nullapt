import { ImageResponse } from "next/og";

export const size = { width: 32, height: 32 };
export const contentType = "image/png";

// Renders the ~/nullapt icon at favicon size.
// Next.js also uses this for 16×16 by downscaling.
export default function Icon() {
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
          borderRadius: "6px",
          fontFamily: "monospace",
        }}
      >
        {/* Tilde only at small size — readable at 16px */}
        <span style={{ color: "#4ade80", fontSize: "22px", fontWeight: 700, lineHeight: 1 }}>
          ~/
        </span>
      </div>
    ),
    { ...size }
  );
}
