# Velld dashboard

Next.js App Router, React, TypeScript, Tailwind CSS, Cloudflare Kumo, and TanStack Query.

## Run

From the repository root, configure `.env` using the [setup guide](../../README.md#run-with-docker). `NEXT_PUBLIC_API_URL` must be reachable from the browser.

```bash
bun install --frozen-lockfile
bun run dev
```

This starts the dashboard on port 3000 and the Go API on 8080. For just the frontend, run `bun run dev:web` with an API already running.

## Check and build

From the repository root:

```bash
bun run check:web
bun run build:web
```

`check:web` runs ESLint, design lint, TypeScript, Bun tests, and dead-code checks. Run `bun run check` to include API checks.

Set `NEXT_PUBLIC_API_URL` before building: Next.js includes this value in the production build.

## Routes

| Route | Page |
| --- | --- |
| `/` | Overview |
| `/connections` | Connections |
| `/connections/new` | Details, URL, or Docker connection setup |
| `/connections/[id]` | Overview, schedule, recovery, configuration |
| `/connections/[id]/edit` | Connection editor |
| `/history` | Backup history and filters |
| `/history/[id]` | Backup details and supported recovery actions |
| `/settings` | Account, storage, notifications |
| `/login` | Sign in |

Views live in `components/views`; API clients in `lib/api`; data hooks in `hooks`. The homepage and documentation are a separate app in `docs/web`.
