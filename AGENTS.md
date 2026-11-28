# Velld — Agent Instructions

## Repo Structure

- `apps/api/` — Go REST API (gorilla/mux, SQLite via goose migrations)
- `apps/web/` — Next.js 15 dashboard (App Router, Cloudflare Kumo, Tailwind)
- `docs/web/` — Fumadocs documentation site
- `docker-compose.yml` — Full stack (builds both apps)
- `docker-compose.prebuilt.yml` — Pre-built images from GHCR

## Dev Commands

**API** (`apps/api/`):
```bash
go run cmd/api-server/main.go
```
Runs on `:8080`. Loads `.env` from repo root AND `apps/api/.env` (via `godotenv.Load`).

**Web and API** (repo root):
```bash
bun install --frozen-lockfile
bun run dev    # use dev:web or dev:api for one app
bun run build
bun run check    # API vet/tests; web lint/types/tests/dead code
```
Web runs on `:3000` with Turbopack and reads `NEXT_PUBLIC_API_URL` from root `.env` via `dotenv.config`.

**Docs** (`docs/web/`):
```bash
bun install --frozen-lockfile
bun run dev --port 3001
bun run build
```

**Both with Docker**:
```bash
cp .env.example .env
docker compose up --build
```

## Environment / Secrets

- `ENCRYPTION_KEY` must be exactly 64 hex characters (32 bytes). Generate with: `openssl rand -hex 32`
- `JWT_SECRET` and `ENCRYPTION_KEY` are **required** — app exits on startup if missing
- Shell command syntax like `$(openssl rand -hex 32)` in `.env` files **does not execute** — the app has explicit checks and error messages for this

## Architecture Notes

- API uses SQLite at `data/velld.db` (default path, configurable via `DB_PATH`)
- Migrations live in `internal/database/migrations/` (goose, auto-runs on startup)
- Secrets are loaded once via `sync.Once` — no hot-reload
- API unit tests: `go -C apps/api test ./...`; Docker integration tests are documented in `apps/api/DOCKER.md`.
- Web checks: `bun run check:web`; typecheck alone: `bun run --cwd apps/web typecheck`.
- `.github/workflows/web-check.yml` runs dashboard checks and production build; separate workflows build images.
- Use Bun for JavaScript commands and lockfiles, not npm.

## Docker / Build Quirks

- API Dockerfile: multi-stage with CGO enabled (`CGO_ENABLED=1`) for `go-sqlite3`
- API Dockerfile copies `./internal/database` separately for migrations loaded at startup.
- Web Dockerfile expects `output: "standalone"` in `next.config.ts`
- Root `docker-compose.yml` builds the API from `./apps/api` and the web from the repository root
- Client bundles: `Dockerfile` (all engines), `Dockerfile.postgres`, `Dockerfile.mysql`, `Dockerfile.mongo`; see `apps/api/DOCKER.md`.

## Web Stack

- Next.js 15 App Router, React 19, TypeScript (strict mode)
- Cloudflare Kumo components plus local Radix/shadcn wrappers; icons from Lucide and Phosphor.
- Path alias: `@/*` maps to root of `apps/web/`
- Radix UI primitives under `components/ui/`

## Key Files

- API entrypoint: `apps/api/cmd/api-server/main.go`
- DB init + migrations: `apps/api/internal/database/database.go`
- Secrets loading: `apps/api/internal/common/secrets.go`
- Web next config: `apps/web/next.config.ts` (loads root `.env` path)
