FROM oven/bun:1.4.5-slim AS builder

WORKDIR /app

COPY package.json bun.lock bunfig.toml ./
COPY apps/web/package.json ./apps/web/package.json
RUN bun install --frozen-lockfile

COPY apps/web ./apps/web
RUN bun run --bun build:web

FROM oven/bun:1.4.5-slim AS runner

WORKDIR /app

ENV NODE_ENV production
ENV HOSTNAME=0.0.0.0

COPY --from=builder /app/apps/web/.next/standalone ./
COPY --from=builder /app/apps/web/.next/static ./apps/web/.next/static
COPY --from=builder /app/apps/web/public ./apps/web/public

WORKDIR /app/apps/web

EXPOSE 3000

CMD ["bun", "server.js"]
