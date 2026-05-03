// safeJsonLd serialises a value for embedding inside an inline
// <script type="application/ld+json"> tag. JSON.stringify alone is unsafe
// because attacker-controlled strings may contain "</script>" sequences
// (or U+2028/U+2029 in legacy parsers) that break out of the script block.
const LS = String.fromCharCode(0x2028);
const PS = String.fromCharCode(0x2029);
const UNSAFE = new RegExp(`[<>&${LS}${PS}]`, "g");
const ESCAPES: Record<string, string> = {
  "<": "\\u003c",
  ">": "\\u003e",
  "&": "\\u0026",
  [LS]: "\\u2028",
  [PS]: "\\u2029",
};

export function safeJsonLd(value: unknown): string {
  return JSON.stringify(value).replace(UNSAFE, (c) => ESCAPES[c]);
}
