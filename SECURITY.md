# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| latest  | ✅ |
| < 1.0   | ❌ |

## Reporting a vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

### Option A — GitHub private advisory (preferred)

Use GitHub's [private vulnerability reporting](https://github.com/nullapt/nullapt/security/advisories/new). This keeps the report confidential until a fix is released.

### Option B — Email

Send details to **security@nullapt.dev**.  
PGP key: [nullapt.dev/security.asc](https://nullapt.dev/security.asc)

### What to include

- A clear description of the vulnerability and its impact
- Steps to reproduce
- Affected component (CLI, registry API, web, SKILL.json parser, sandbox)
- Any proof-of-concept code or manifest you used

## Response timeline

| Step | Target |
|------|--------|
| Initial acknowledgement | 48 hours |
| Severity assessment | 5 days |
| Fix and coordinated disclosure | 7–14 days |

## Threat model

NullApt's core security promise is **capability containment**: a skill cannot exceed the permissions declared in its `SKILL.json`. The critical enforcement layers are:

1. **Ed25519 signature verification** — CLI verifies before install; registry verifies before accepting a publish.
2. **WASM-WASI sandbox** — the Extism/wazero host enforces `AllowedHosts` and `AllowedPaths` at the syscall boundary.
3. **Transparency log** — every publish records the author's public key, enabling key-rotation audits.

Vulnerabilities that break any of these layers are **critical severity**.

## Scope

| In scope | Out of scope |
|----------|--------------|
| Signature bypass | Intentionally malicious skills published by their authors |
| Sandbox escape (network, filesystem, env) | Skills that declare and use their permissions honestly |
| Registry API auth bypass | Brute-force of weak user passwords |
| Transparency log tampering | Social engineering of maintainers |
| Manifest parser memory safety | DoS via extremely large manifests (best-effort) |

## Disclosure policy

We follow coordinated disclosure. We will credit reporters in the release notes unless you prefer to remain anonymous.
