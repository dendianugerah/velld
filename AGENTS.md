# Velld — Agent Instructions

## Repo Structure

- `apps/api/` — Go REST API (gorilla/mux, SQLite via goose migrations)
- `apps/web/` — Next.js 15 frontend (App Router, shadcn/ui, Tailwind)
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
bun install
bun run dev    # use dev:web or dev:api for one app
bun run build
bun run --cwd apps/web lint
```
Web runs on `:3000` with Turbopack and reads `NEXT_PUBLIC_API_URL` from root `.env` via `dotenv.config`.

**Docs** (`docs/web/`):
```bash
bun install
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
- API has **no tests** (no `_test.go` files found)
- Web has **no tests** and **no typecheck script** (only `bun run lint` in `apps/web/`)
- CI only runs Docker build/push — no lint or test checks

## Docker / Build Quirks

- API Dockerfile: multi-stage with CGO enabled (`CGO_ENABLED=1`) for `go-sqlite3`
- API Dockerfile copies `./internal/database` separately (for embedded migrations)
- Web Dockerfile expects `output: "standalone"` in `next.config.ts`
- Root `docker-compose.yml` builds the API from `./apps/api` and the web from the repository root
- Multiple API Dockerfiles exist: `Dockerfile` (all DBs ~120MB), `Dockerfile.postgres` (~70MB), `Dockerfile.mysql` (~65MB), `Dockerfile.mongo` (~80MB)

## Web Stack

- Next.js 15 App Router, React 19, TypeScript (strict mode)
- UI components from shadcn/ui (style: `new-york`, icons: lucide)
- Path alias: `@/*` maps to root of `apps/web/`
- Radix UI primitives under `components/ui/`

## Key Files

- API entrypoint: `apps/api/cmd/api-server/main.go`
- DB init + migrations: `apps/api/internal/database/database.go`
- Secrets loading: `apps/api/internal/common/secrets.go`
- Web next config: `apps/web/next.config.ts` (loads root `.env` path)
