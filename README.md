# Velld

A self-hosted database backup management tool. Schedule automated backups, monitor status, and manage multiple databases from one place.

**[Documentation](https://velld.vercel.app)** · **[Quick Start](https://velld.vercel.app/docs/quick-start)**

## Features

- Multiple database support (PostgreSQL, MySQL, MongoDB)
- Automated scheduling with cron syntax
- S3-compatible storage integration
- Database restore functionality
- Email and webhook notifications for failed backups

## Getting Started

```bash
curl -o docker-compose.yml https://raw.githubusercontent.com/dendianugerah/velld/main/docker-compose.prebuilt.yml
```

Create a `.env` file:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080

# Generate secure keys (IMPORTANT!)
JWT_SECRET=$(openssl rand -hex 32)
ENCRYPTION_KEY=$(openssl rand -hex 32)

# Admin credentials for initial setup
ADMIN_USERNAME_CREDENTIAL=admin
ADMIN_PASSWORD_CREDENTIAL=changeme
```

> **Important:** The `ENCRYPTION_KEY` must be a **64-character hex string**. Don't use shell commands like `$(openssl rand -hex 32)` directly in `.env` files - they won't execute. Instead, run the command in your terminal and paste the output.

**Generate keys properly:**

```bash
# Generate and display keys
echo "JWT_SECRET=$(openssl rand -hex 32)"
echo "ENCRYPTION_KEY=$(openssl rand -hex 32)"

# Copy the output to your .env file
```

Start the containers:

```bash
docker compose up -d
```

On startup, the API creates the configured admin account if that username does not already exist. Visit [localhost:3000](http://localhost:3000) and log in with those credentials. Existing accounts continue to work; public registration is unavailable.

For detailed setup instructions, see the [Installation Guide](https://velld.vercel.app/docs/installation).

### Discover databases on the Docker host

Enable Docker discovery by adding the optional Compose file:

```bash
docker compose -f docker-compose.yml -f docker-compose.discovery.yml up -d
```

Use the first filename that matches your installation (for example, `docker-compose.prebuilt.yml` when using the published images). Then log in as the configured admin and open **Connections → Discover Docker**. Velld finds running PostgreSQL, MySQL/MariaDB, and MongoDB containers on a shared Docker network or a reachable published port. **Add & schedule** adds a connection and enables daily backups at midnight (API server time), with 30 days of retention; you can change that schedule from Connections.

### Restore checks

With the same Docker access, Velld restores every completed backup into a throwaway container of the same database (PostgreSQL, MySQL/MariaDB, or MongoDB, matching the source server version when the dump records it), confirms the restored tables, rows, or documents match the backup, then removes the container. Backup history shows **Restore verified** or **Restore check failed**, and failures send the same webhook and email alerts as failed backups. Without Docker access the check is recorded as **Restore not checked**. Set `RESTORE_CHECK_ENABLED=false` to turn it off.

Velld reads the database image's standard initialization environment variables. If credentials were changed later, or supplied via Docker secrets instead of environment variables, add the connection manually. Only enable this on a trusted Velld installation: giving the API access to the Docker socket grants privileged access to the host even with a read-only bind mount.

### Branch from a backup

With Docker access enabled, choose **Branch from backup…** on any completed PostgreSQL, MySQL/MariaDB, or MongoDB backup in History. Velld restores the backup into a separate container and gives you a database-scoped connection string only after masking succeeds. The port binds to the Docker host's loopback address, so connect on that host or forward the port over SSH. Velld's API removes the container after 24 hours.

Masking uses column and field names to find common email, name, and phone fields, including nested MongoDB fields. It does not inspect unrelated fields or free text, so review the branch before sharing its data. If recognized fields cannot be safely updated, Velld removes the failed branch instead of issuing credentials.

## Screenshots

<table>
  <tr>
    <td><img src="docs/images/dashboard.png" alt="Dashboard" /></td>
    <td><img src="docs/images/connections.png" alt="Connections" /></td>
  </tr>
  <tr>
    <td align="center"><b>Dashboard</b></td>
    <td align="center"><b>Connections</b></td>
  </tr>
  <tr>
    <td><img src="docs/images/history.png" alt="History" /></td>
    <td></td>
  </tr>
  <tr>
    <td align="center"><b>Backup History</b></td>
    <td></td>
  </tr>
</table>

## Documentation

Complete documentation is available at [velld.vercel.app](https://velld.vercel.app):

- [Installation](https://velld.vercel.app/docs/installation)
- [Database Configuration](https://velld.vercel.app/docs/databases)
- [Troubleshooting](https://velld.vercel.app/docs/troubleshooting)

## Development

Clone the repository:

```bash
git clone https://github.com/dendianugerah/velld.git
cd velld
```

With Docker:

```bash
cp .env.example .env
docker compose up -d
```

Without Docker, install [Bun](https://bun.com/), Node.js 20+, Go 1.24+, and a C compiler (needed by SQLite). Copy the environment template:

```bash
cp .env.example .env
```

Replace the `JWT_SECRET` and `ENCRYPTION_KEY` placeholders in `.env` with separate outputs from `openssl rand -hex 32`, and set the admin username and password you'll use to sign in. Then run both apps from the repo root:

```bash
bun install && bun run dev
```

The web app is at [localhost:3000](http://localhost:3000) and the API is at [localhost:8080](http://localhost:8080/health). Press Ctrl+C to stop both. To run just one app, use `bun run dev:web` or `bun run dev:api`. Use `bun run build` to build both without starting them.

For the docs, run `bun install --cwd docs/web` once, then `bun run dev:docs` to serve them at [localhost:3001](http://localhost:3001) or `bun run build:docs` to build them. Use `bun run dev:all` to start the web app, API, and docs together.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

MIT License - see [LICENSE](LICENSE) for details.