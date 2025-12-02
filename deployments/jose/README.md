# JOSE Deployment Configuration

This directory contains deployment configurations for the JOSE product.

## Current Status

**Placeholder**: JOSE currently runs as an embedded library within other products.
This directory is prepared for future standalone JOSE service deployments.

## Planned Files

```text
deployments/jose/
├── compose.yml              # Docker Compose for JOSE service
├── compose.demo.yml         # Demo overlay
├── Dockerfile               # JOSE service container
├── config/
│   ├── jose-sqlite.yml      # SQLite backend config
│   └── jose-postgres.yml    # PostgreSQL backend config
└── secrets/
    └── .gitkeep             # Secrets placeholder
```

## Service Ports

| Service | Public Port | Admin Port | Backend |
|---------|-------------|------------|---------|
| jose-sqlite | 8080 | 9090 | SQLite in-memory |
| jose-postgres-1 | 8081 | 9090 | PostgreSQL |
| jose-postgres-2 | 8082 | 9090 | PostgreSQL |
