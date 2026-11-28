# Velld API images

The root Compose file builds `apps/api/Dockerfile`. All API images compile Go with CGO for SQLite; D1 exports and SQLite restore checks are built into the API.

## Client bundles

| Dockerfile | Database tools |
| --- | --- |
| `Dockerfile` | PostgreSQL, MariaDB/MySQL, MongoDB, and MySQL binlog reader |
| `Dockerfile.postgres` | PostgreSQL client |
| `Dockerfile.mysql` | MariaDB/MySQL client and MySQL binlog reader |
| `Dockerfile.mongo` | MongoDB Database Tools |

Build from this directory:

```bash
docker build -t velld-api .
docker build -f Dockerfile.postgres -t velld-api-postgres .
```

To change the source-build Compose image, set `services.api.build.dockerfile` in the root `docker-compose.yml`. The prebuilt Compose file pulls GHCR images instead.

Database-specific images only bundle the listed clients. Use the full image for mixed engines. Velld's MySQL commands use MariaDB client flags; Oracle MySQL clients are not a drop-in replacement.

## Runtime

Configure required secrets and admin credentials using the [root setup guide](../../README.md#run-with-docker). Keep SQLite (`/app/data`) and backups (`/app/backups`) in persistent volumes.

The optional root `docker-compose.discovery.yml` mounts `/var/run/docker.sock` and adds the host gateway. It enables discovery, native-engine restore checks, and disposable recovery containers. The API talks to the Docker Engine API, not a Docker CLI. Socket access grants control of the host's Docker daemon.

## Tests

From this directory, with the Go version in `go.mod` and a C compiler:

```bash
go vet ./...
go test ./...
```

Integration tests also require Docker, PostgreSQL 14–16 client tools, MariaDB's `mysqldump`/`mysql`, and `mongodump`/`mongorestore`:

```bash
go test -v -count=1 -tags integration -run '^TestDocker' ./internal/backup
```

The suite starts and removes disposable containers on Docker-assigned loopback ports. It fails if Docker or required clients are unavailable.
