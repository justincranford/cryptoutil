# KMS (Key Management Service) Product

This directory contains the Key Management Service (KMS) product implementation.

## Current Status

**Note**: KMS code is currently in `internal/server/`. This directory is prepared for the
migration to the 4-products architecture.

## Target Structure

```text
internal/kms/
├── application/     # Server application lifecycle
├── barrier/         # Key hierarchy barrier
│   ├── contentkeysservice/
│   ├── intermediatekeysservice/
│   ├── rootkeysservice/
│   └── unsealkeysservice/
├── client/          # KMS HTTP client
├── config/          # KMS-specific configuration
├── demo/            # KMS demo utilities
├── handler/         # HTTP handlers
├── middleware/      # HTTP middleware
├── repository/      # Data repository
│   ├── orm/
│   └── sqlrepository/
└── service/         # Business logic (was businesslogic/)
```

## Key Concepts

### Key Hierarchy

```text
Unseal Secrets (file:///run/secrets/* or Yubikey)
    ↓
Root Keys (derived from unseal secrets)
    ↓
Intermediate Keys (per-tenant isolation)
    ↓
ElasticKey (policy container)
    ↓
MaterialKey (versioned key material)
```

### API Endpoints

- `POST /elastickey` - Create ElasticKey
- `GET /elastickey/{id}` - Get ElasticKey
- `GET /elastickeys` - List ElasticKeys
- `POST /elastickey/{id}/materialkey` - Create MaterialKey
- `POST /elastickey/{id}/encrypt` - Encrypt data
- `POST /elastickey/{id}/decrypt` - Decrypt data
- `POST /elastickey/{id}/sign` - Sign data
- `POST /elastickey/{id}/verify` - Verify signature

## Migration

Files will be migrated from `internal/server/` to `internal/kms/`:

| Source | Target |
|--------|--------|
| `server/application/` | `kms/application/` |
| `server/barrier/` | `kms/barrier/` |
| `server/businesslogic/` | `kms/service/` |
| `server/demo/` | `kms/demo/` |
| `server/handler/` | `kms/handler/` |
| `server/middleware/` | `kms/middleware/` |
| `server/repository/` | `kms/repository/` |
