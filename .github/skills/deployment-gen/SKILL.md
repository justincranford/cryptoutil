---
name: deployment-gen
description: "Generate complete deployment structure for a new cryptoutil service including compose.yml, Dockerfile, secrets directory, and config overlay files. Use when adding a deployment for a new or existing service to ensure correct directory layout, port bindings, and Docker secrets integration."
argument-hint: "[PS-ID]"
disable-model-invocation: true
---

Generate complete deployment structure for a new cryptoutil service.

## Purpose

Use when creating deployment artifacts for a new service. Generates the full
directory structure: compose.yml, Dockerfile, secrets/, and config overlay files.

## Required Directory Structure

```
deployments/PS-ID/
├── compose.yml
├── Dockerfile
├── config/
│   ├── PS-ID-app-common.yml
│   ├── PS-ID-app-sqlite-1.yml
│   ├── PS-ID-app-sqlite-2.yml
│   ├── PS-ID-app-postgresql-1.yml
│   └── PS-ID-app-postgresql-2.yml
└── secrets/
    ├── unseal-1of5.secret
    ├── unseal-2of5.secret
    ├── unseal-3of5.secret
    ├── unseal-4of5.secret
    ├── unseal-5of5.secret
    ├── hash-pepper-v3.secret
    ├── postgres-database.secret
    ├── postgres-username.secret
    ├── postgres-password.secret
    ├── postgres-url.secret
    ├── browser-username.secret
    ├── browser-password.secret
    ├── service-username.secret
    └── service-password.secret
```

## Required Config Counterpart

```
configs/PS-ID/
└── PS-ID.yml               # Standalone dev domain config
```

## Key Rules

- ALWAYS relative paths in compose.yml (NEVER absolute)
- ALWAYS `127.0.0.1` in containers (NOT `localhost` — Alpine resolves to IPv6)
- Use `docker compose` (NOT `docker-compose`)
- Admin bind: `127.0.0.1:9090` inside containers (never exposed outside)
- Public bind: `0.0.0.0:8080` inside containers
- Host ports from service catalog (8000-8999 range, unique per service)
- Exactly 5 config overlay files per service (common, sqlite-1/2, postgresql-1/2)
- Use `wget` for healthchecks in Alpine containers (not `curl`)
- Healthcheck fields use hyphens: `start-period` (NOT `start_period`)
- ALL credentials MUST use Docker secrets, NEVER inline env vars

## Port Assignment

| Deployment Type | Port Offset | Example (sm-im base 8700) |
|-----------------|-------------|---------------------------|
| Service | 8XXX | 8700-8799 |
| Product | 8XXX + 10000 | 18700-18799 |
| Suite | 8XXX + 20000 | 28700-28799 |

## Template: compose.yml

```yaml
services:
  PS-ID-app-sqlite-1:
    build:
      context: ../../
      dockerfile: deployments/PS-ID/Dockerfile
    image: cryptoutil:local
    container_name: PS-ID-app-sqlite-1
    ports:
      - "127.0.0.1:8XX0:8080"  # Public HTTPS
    configs:
      - source: PS-ID-app-common
        target: /app/config/PS-ID-app-common.yml
      - source: PS-ID-app-sqlite-1
        target: /app/config/PS-ID-app-sqlite-1.yml
    secrets:
      - unseal-1of5.secret
      - unseal-2of5.secret
      - unseal-3of5.secret
      - unseal-4of5.secret
      - unseal-5of5.secret
      - hash-pepper-v3.secret
      - browser-username.secret
      - browser-password.secret
      - service-username.secret
      - service-password.secret
    healthcheck:
      test: ["CMD", "wget", "--spider", "--no-check-certificate",
             "https://127.0.0.1:8080/browser/api/v1/health"]
      interval: 10s
      timeout: 5s
      retries: 3
      start-period: 15s
    command: ["server",
              "--config", "/app/config/PS-ID-app-common.yml",
              "--config", "/app/config/PS-ID-app-sqlite-1.yml"]

  PS-ID-postgres:
    image: postgres:18
    container_name: PS-ID-postgres
    ports:
      - "127.0.0.1:5432X:5432"
    secrets:
      - postgres-database.secret
      - postgres-username.secret
      - postgres-password.secret
    environment:
      POSTGRES_DB_FILE: /run/secrets/postgres-database.secret
      POSTGRES_USER_FILE: /run/secrets/postgres-username.secret
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres-password.secret
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $(cat /run/secrets/postgres-username.secret)"]
      interval: 5s
      timeout: 5s
      retries: 5

configs:
  PS-ID-app-common:
    file: ./config/PS-ID-app-common.yml
  PS-ID-app-sqlite-1:
    file: ./config/PS-ID-app-sqlite-1.yml
  PS-ID-app-postgresql-1:
    file: ./config/PS-ID-app-postgresql-1.yml

secrets:
  unseal-1of5.secret:
    file: ./secrets/unseal-1of5.secret
  unseal-2of5.secret:
    file: ./secrets/unseal-2of5.secret
  unseal-3of5.secret:
    file: ./secrets/unseal-3of5.secret
  unseal-4of5.secret:
    file: ./secrets/unseal-4of5.secret
  unseal-5of5.secret:
    file: ./secrets/unseal-5of5.secret
  hash-pepper-v3.secret:
    file: ./secrets/hash-pepper-v3.secret
  postgres-database.secret:
    file: ./secrets/postgres-database.secret
  postgres-username.secret:
    file: ./secrets/postgres-username.secret
  postgres-password.secret:
    file: ./secrets/postgres-password.secret
  postgres-url.secret:
    file: ./secrets/postgres-url.secret
  browser-username.secret:
    file: ./secrets/browser-username.secret
  browser-password.secret:
    file: ./secrets/browser-password.secret
  service-username.secret:
    file: ./secrets/service-username.secret
  service-password.secret:
    file: ./secrets/service-password.secret
```

## Template: Dockerfile

```dockerfile
ARG GO_VERSION=1.26.1
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/cryptoutil ./cmd/PS-ID

FROM alpine:3.19 AS validator
WORKDIR /app
COPY --from=builder /app/cryptoutil /app/cryptoutil
RUN echo "Secrets validation stage"

FROM alpine:3.19 AS runtime
WORKDIR /app
COPY --from=validator /app/cryptoutil /app/cryptoutil
ENTRYPOINT ["/app/cryptoutil"]
```

## Template: Config Overlay (common)

```yaml
# PS-ID-app-common.yml
bind-public-address: "0.0.0.0"
bind-public-port: 8080
bind-private-address: "127.0.0.1"
bind-private-port: 9090
```

## Validation Checklist

- [ ] Directory structure matches required layout exactly
- [ ] `compose.yml` uses relative paths only
- [ ] Admin port `127.0.0.1:9090` never exposed to host
- [ ] All credentials use Docker secrets (no inline env vars)
- [ ] Exactly 5 config overlay files: common, sqlite-1/2, postgresql-1/2
- [ ] `configs/PS-ID/PS-ID.yml` standalone config exists
- [ ] Port range matches service catalog assignment
- [ ] Healthcheck uses `wget` (Alpine container)
- [ ] Healthcheck fields use hyphens (`start-period`)
- [ ] `go run ./cmd/cicd-lint lint-deployments` passes
- [ ] Secrets generated with `/secret-gen` skill

## References

Read [ARCHITECTURE.md Section 12. Deployment Architecture](../../../docs/ARCHITECTURE.md#12-deployment-architecture) for Docker Compose patterns — follow all conventions for relative paths, Alpine containers, and healthcheck configuration.

Read [ARCHITECTURE.md Section 3.4 Port Assignments & Networking](../../../docs/ARCHITECTURE.md#34-port-assignments--networking) for port catalog — select the correct port range for the service being deployed.

Read [ARCHITECTURE.md Section 13.1 Deployment Tooling](../../../docs/ARCHITECTURE.md#131-deployment-tooling) for deployment validation — ensure all 8 validators pass after creating deployment artifacts.

Read [ARCHITECTURE.md Section 13.2 Config File Architecture](../../../docs/ARCHITECTURE.md#132-config-file-architecture) for config file types — follow flat kebab-case YAML key conventions for service framework configs.
