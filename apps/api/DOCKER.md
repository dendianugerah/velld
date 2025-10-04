# Velld API Docker Images

This directory contains multiple Dockerfile configurations for different database setups.

## Available Dockerfiles

### `Dockerfile` (Default - All Databases)
Includes PostgreSQL, MySQL, and MongoDB clients. Use this if you need to backup multiple database types.

**Size:** ~120MB  
**Includes:** postgresql-client, mysql-client, mongodb-tools

```bash
docker build -t velld-api .
```

### `Dockerfile.postgres` (PostgreSQL Only)
Lightweight image with only PostgreSQL client.

**Size:** ~70MB  
**Includes:** postgresql-client only

```bash
docker build -f Dockerfile.postgres -t velld-api .
```

### `Dockerfile.mysql` (MySQL Only)
Lightweight image with only MySQL client.

**Size:** ~65MB  
**Includes:** mysql-client only

```bash
docker build -f Dockerfile.mysql -t velld-api .
```

### `Dockerfile.mongo` (MongoDB Only)
Lightweight image with only MongoDB tools.

**Size:** ~80MB  
**Includes:** mongodb-tools only

```bash
docker build -f Dockerfile.mongo -t velld-api .
```

## Usage with Docker Compose

To use a specific Dockerfile, update your `docker-compose.yml`:

```yaml
services:
  api:
    build:
      context: ./apps/api
      dockerfile: Dockerfile.postgres  # Change this to your preferred Dockerfile
    ports:
      - "8080:8080"
    # ...rest of configuration
```

## Choosing the Right Dockerfile

| Database(s) You Need | Dockerfile to Use | Image Size |
|---------------------|-------------------|------------|
| PostgreSQL only | `Dockerfile.postgres` | ~70MB |
| MySQL only | `Dockerfile.mysql` | ~65MB |
| MongoDB only | `Dockerfile.mongo` | ~80MB |
| Multiple types | `Dockerfile` (default) | ~120MB |

## Security Benefits

Using database-specific Dockerfiles:
- ✅ Smaller attack surface (fewer tools installed)
- ✅ Faster builds and deployments
- ✅ Reduced image size
- ✅ Only install what you actually need

## Development

All Dockerfiles use multi-stage builds:
1. **Builder stage:** Compiles the Go application with CGO enabled (for SQLite support)
2. **Runtime stage:** Minimal Alpine Linux with only necessary database clients

This keeps the final image as small and secure as possible.
