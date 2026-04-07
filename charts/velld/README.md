# Velld Helm Chart

A Helm chart for deploying [Velld](https://github.com/dendianugerah/velld) — a self-hosted database backup management tool.

## Prerequisites

- Kubernetes 1.23+
- Helm 3.x

## Installation

### Quick start (development / testing)

```bash
helm install velld ./charts/velld \
  --set secrets.create=true \
  --set secrets.data.JWT_SECRET="$(openssl rand -hex 32)" \
  --set secrets.data.ENCRYPTION_KEY="$(openssl rand -hex 32)"
```

### Using an existing Secret

Create the secret manually:

```bash
kubectl create secret generic velld-secrets \
  --from-literal=JWT_SECRET="$(openssl rand -hex 32)" \
  --from-literal=ENCRYPTION_KEY="$(openssl rand -hex 32)" \
  --from-literal=ADMIN_USERNAME_CREDENTIAL=admin \
  --from-literal=ADMIN_PASSWORD_CREDENTIAL=changeme \
  --from-literal=NEXT_PUBLIC_API_URL=http://velld-api:8080
```

Then install the chart:

```bash
helm install velld ./charts/velld \
  --set secrets.existingSecret=velld-secrets
```

### With ingress

```bash
helm install velld ./charts/velld \
  --set secrets.existingSecret=velld-secrets \
  --set api.ingress.enabled=true \
  --set api.ingress.host=velld-api.example.com \
  --set web.ingress.enabled=true \
  --set web.ingress.host=velld.example.com
```

## Uninstall

```bash
helm uninstall velld
```

> **Note:** PersistentVolumeClaims are **not** deleted automatically. Remove them
> manually if you no longer need the data.

## Configuration

### Global

| Parameter | Description | Default |
|---|---|---|
| `nameOverride` | Override the chart name | `""` |
| `fullnameOverride` | Override the full release name | `""` |

### API

| Parameter | Description | Default |
|---|---|---|
| `api.image.repository` | API image repository | `ghcr.io/dendianugerah/velld/api` |
| `api.image.tag` | API image tag | `latest` |
| `api.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `api.service.port` | API service port | `8080` |
| `api.ingress.enabled` | Enable API ingress | `false` |
| `api.ingress.host` | API ingress hostname | `velld-api.example.com` |
| `api.ingress.annotations` | API ingress annotations | `{}` |
| `api.ingress.tls` | Enable TLS on API ingress | `false` |
| `api.ingress.secretName` | TLS secret name for API ingress | `""` |
| `api.ingress.ingressClassName` | Ingress class name for API | `""` |
| `api.resources` | CPU/Memory resource requests and limits | `{}` |
| `api.persistence.data.size` | Size of the data PVC | `1Gi` |
| `api.persistence.data.storageClass` | StorageClass for the data PVC | `""` |
| `api.persistence.backups.size` | Size of the backups PVC | `5Gi` |
| `api.persistence.backups.storageClass` | StorageClass for the backups PVC | `""` |

### Web

| Parameter | Description | Default |
|---|---|---|
| `web.image.repository` | Web image repository | `ghcr.io/dendianugerah/velld/web` |
| `web.image.tag` | Web image tag | `latest` |
| `web.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `web.service.port` | Web service port | `3000` |
| `web.ingress.enabled` | Enable Web ingress | `false` |
| `web.ingress.host` | Web ingress hostname | `velld.example.com` |
| `web.ingress.annotations` | Web ingress annotations | `{}` |
| `web.ingress.tls` | Enable TLS on Web ingress | `false` |
| `web.ingress.secretName` | TLS secret name for Web ingress | `""` |
| `web.ingress.ingressClassName` | Ingress class name for Web | `""` |
| `web.env.ALLOW_REGISTER` | Allow new user registration | `"true"` |

### Secrets

| Parameter | Description | Default |
|---|---|---|
| `secrets.create` | Create a Secret from `secrets.data` | `false` |
| `secrets.existingSecret` | Name of a pre-existing Secret | `""` |
| `secrets.data.JWT_SECRET` | JWT signing key (64-char hex) | `changeme-...` |
| `secrets.data.ENCRYPTION_KEY` | Data encryption key (64-char hex) | `changeme-...` |
| `secrets.data.ADMIN_USERNAME_CREDENTIAL` | Initial admin username | `admin` |
| `secrets.data.ADMIN_PASSWORD_CREDENTIAL` | Initial admin password | `changeme` |
| `secrets.data.NEXT_PUBLIC_API_URL` | Internal URL the web pod uses to reach the API | `http://velld-api:8080` |
