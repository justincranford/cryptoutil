# CA Deployment Configuration

This directory contains deployment configurations for the Certificate Authority (CA) product.

## Current Status

**Placeholder**: CA functionality is planned but not yet implemented.
This directory is prepared for future CA service deployments.

## Planned Files

```text
deployments/ca/
├── compose.yml              # Docker Compose for CA service
├── compose.demo.yml         # Demo overlay
├── Dockerfile               # CA service container
├── config/
│   ├── ca-sqlite.yml        # SQLite backend config
│   └── ca-postgres.yml      # PostgreSQL backend config
└── secrets/
    ├── root-ca.key          # Root CA private key (encrypted)
    ├── intermediate-ca.key  # Intermediate CA private key (encrypted)
    └── .gitkeep             # Secrets placeholder
```

## Service Ports

| Service | Public Port | Admin Port | Backend |
|---------|-------------|------------|---------|
| ca-sqlite | 8080 | 9090 | SQLite in-memory |
| ca-postgres-1 | 8081 | 9090 | PostgreSQL |
| ca-postgres-2 | 8082 | 9090 | PostgreSQL |

## Compliance Notes

- Root CA should be deployed offline or with air-gapped network
- HSM integration planned for production key storage
- Audit logging required for all certificate operations
