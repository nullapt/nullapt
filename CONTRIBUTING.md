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

- Go 1.23+
- Node.js 20+
- Docker (for local Postgres) or a Postgres 15+ instance

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

## Releasing a new skill to the registry

Skills are published independently via `nullapt publish`. You do not need a PR to add a skill to the public registry — just sign and publish it directly.

If you want your skill listed as **officially audited**, open a [Skill Submission issue](.github/ISSUE_TEMPLATE/skill_submission.yml) and a maintainer will review the manifest and WASM binary.

## Code of Conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md). Be kind.
