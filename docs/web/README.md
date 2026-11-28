# Velld homepage and docs

Next.js and Fumadocs. No backend or credentials are needed.

## Run

From the repository root:

```bash
bun install --cwd docs/web --frozen-lockfile
bun run dev:docs
```

Open [localhost:3001](http://localhost:3001). `bun run dev:all` also starts the dashboard and API.

## Build and typecheck

```bash
bun run build:docs
bun run --cwd docs/web tsc --noEmit
```

## Content

| Path | Purpose |
| --- | --- |
| `app/(home)` | Homepage |
| `components/backup-story.tsx`, `components/backup-figure.tsx` | Interactive backup diagram |
| `content/docs/*.mdx` | Documentation pages |
| `content/docs/meta.json` | Guide order |
| `app/docs` | Documentation layout and install-prompt preview |
| `app/api/search/route.ts` | Search endpoint |
| `app/llms.txt`, `app/llms-full.txt` | Agent index and full-text docs |
| `lib/llms.ts` | Plain-text documentation generator |

Documented settings and navigation must match `apps/web`; setup commands must match the root Compose files and `.env.example`.
