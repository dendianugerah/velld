# Velld

A self-hosted database backup management tool. Schedule automated backups, monitor status, and manage multiple databases from one place.

**[Documentation](https://velld.vercel.app)** · **[Quick Start](https://velld.vercel.app/docs/quick-start)**

## Features

- Multiple database support (PostgreSQL, MySQL, MongoDB)
- Automated scheduling with cron syntax
- S3-compatible storage integration
- Built-in backup comparison and diff viewer
- Email notifications for failed backups
- Responsive web interface

## Getting Started

```bash
curl -o docker-compose.yml https://raw.githubusercontent.com/dendianugerah/velld/main/docker-compose.prebuilt.yml
```

Create a `.env` file:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
JWT_SECRET=$(openssl rand -hex 32)
ENCRYPTION_KEY=$(openssl rand -hex 32)
ADMIN_USERNAME_CREDENTIAL=admin
ADMIN_PASSWORD_CREDENTIAL=changeme
```

```bash
docker compose up -d
```

Visit [localhost:3000](http://localhost:3000) and log in with your admin credentials.

For detailed setup instructions, see the [Installation Guide](https://velld.vercel.app/docs/installation).

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

Without Docker - API:

```bash
cd apps/api
go mod download
go run cmd/api-server/main.go
```

Without Docker - Web:

```bash
cd apps/web
npm install
npm run dev
```

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

MIT License - see [LICENSE](LICENSE) for details.