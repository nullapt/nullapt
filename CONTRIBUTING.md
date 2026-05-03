# Contributing to NullApt

Thank you for considering a contribution. NullApt is fully open source and community-driven.

## Ways to contribute

- **File a bug** — [Bug report template](.github/ISSUE_TEMPLATE/bug_report.yml)
- **Propose a feature** — [Feature request template](.github/ISSUE_TEMPLATE/feature_request.yml)
- **Submit a skill** — [Skill submission template](.github/ISSUE_TEMPLATE/skill_submission.yml)
- **Write code** — see below
- **Improve docs** — any `.md` file in the repo, or the web app in `web/`

## Development setup

### Prerequisites

- Go 1.25+
- Node.js 20+
- Docker (for local Postgres) or a Postgres 15+ instance
- Rust + `wasm32-wasip1` target (only if you're building example skills): `rustup target add wasm32-wasip1`

### Clone and build

```bash
git clone https://github.com/nullapt/nullapt
cd nullapt

# CLI
go build ./cmd/nullapt

# Registry API
go build ./cmd/registry-api

# Web marketplace
cd web && npm install
```

### Run tests

```bash
# Go tests
go test ./...

# Web type-check
cd web && npx tsc --noEmit
```

### Local registry

```bash
# Start Postgres
docker run -d --name nullapt-pg \
  -e POSTGRES_USER=nullapt \
  -e POSTGRES_PASSWORD=nullapt \
  -e POSTGRES_DB=nullapt \
  -p 5432:5432 postgres:16

# Apply schema
psql postgres://nullapt:nullapt@localhost:5432/nullapt -f internal/db/schema.sql

# Start registry API
DATABASE_URL=postgres://nullapt:nullapt@localhost:5432/nullapt \
  BLOB_READ_WRITE_TOKEN=dev-token-not-used \
  go run ./cmd/registry-api

# Start web
cd web
NEXT_PUBLIC_REGISTRY_URL=http://localhost:8080 npm run dev
```

## Code guidelines

- **Go**: standard `gofmt` formatting, `go vet` clean, no unused imports.
- **TypeScript**: strict mode, `npx tsc --noEmit` must pass.
- **Comments**: only when the *why* is non-obvious. No docstrings that just restate the function name.
- **Error messages**: lowercase, no trailing period, wrap with `%w` for Go errors.
- **Tests**: new public functions need at least a happy-path test.

## Commit messages

```
<type>: <short summary>

Optional body — explain the why, not the what.
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `ci`, `chore`

Examples:
```
feat: add transparency log endpoint
fix: reject wildcard network domains in manifest validation
docs: add self-hosting guide to README
```

## Pull request process

1. Fork and create a branch: `git checkout -b feat/my-thing`
2. Make your changes and ensure `go test ./...` and `npx tsc --noEmit` pass
3. Open a PR against `main` using the PR template
4. A maintainer will review within 3 business days

## Publishing a skill

Skills are published independently via `nullapt publish` from anywhere on your machine — same as `npm publish`. You don't need to fork or PR this repo to add a skill to the registry.

There are two contributor paths, with very different setup:

### A. Publish under your own name *(no extra setup)*

If your GitHub username is `alice`, you can publish `alice/cool-thing` (or just `cool-thing`) immediately after the standard `nullapt login`. The default OAuth grant only asks for `read:user user:email` — same scopes as npm — so casual contributors never see the org consent screen.

```bash
nullapt keygen                    # one-time keypair
nullapt login                     # one-time GitHub OAuth
nullapt sign  ./SKILL.json
nullapt publish ./SKILL.json
```

### B. Publish under an organization namespace

If you want to publish under an org (e.g. `acme/foo` or `nullapt/foo`), our backend needs to verify you're a member of that GitHub org. Two extra one-time steps per org:

1. **Make your org membership public on GitHub.** Visit `github.com/orgs/<org>/people`, find yourself in the list, set membership to **Public**. GitHub OAuth only exposes public memberships.
2. **Connect the org on nullapt.dev.** Sign in to [nullapt.dev](https://nullapt.dev), visit **Settings → Organizations**, click **Connect GitHub organizations**. This re-authorizes the app with the `read:org` scope and syncs your memberships.

After that, set `name` in your `SKILL.json` to `<org>/<skill>` and publish as usual. New versions of an org-namespaced skill require you to still be a member; we re-check on every publish.

If your org has restricted third-party OAuth apps, an org owner has to approve the NullApt app at the org level first.

### Getting your skill into the curated index

The list at [nullapt.dev](https://nullapt.dev) draws from the [`nullapt/skills`](https://github.com/nullapt/skills) repo — a curated index, not the source of truth. Open a PR there to add an entry pointing at your published skill.

For an **officially audited** badge, open a [Skill Submission issue](.github/ISSUE_TEMPLATE/skill_submission.yml) on this repo. A maintainer will review the manifest and WASM binary.

## Code of Conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md). Be kind.
