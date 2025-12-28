# Service Template Refactoring - learn-im

## Implementation Checklist

### Phase 1: Package Structure Migration

- [ ] Move files from `internal/cmd/learn/im` to layered architecture directories
- [ ] Update package imports (api/business/repository/util)
- [ ] Verify build succeeds after package restructure
- [ ] Run tests to detect broken imports

### Phase 2: Shared Infrastructure Integration

- [ ] Integrate `internal/shared/barrier` for barrier layer encryption
- [ ] Integrate `internal/shared/crypto/jose` for JWK generation
- [ ] Remove version flag from learn-im CLI parameters
- [ ] Update `internal/learn/server/server.go` to initialize shared services

### Phase 3: Database Schema

- [ ] Create migration files (same pattern as KMS)
- [ ] Define `users_jwks` table with GORM models
- [ ] Define `users_messages_jwks` table with GORM models
- [ ] Define `messages_jwks` table with GORM models
- [ ] Define `messages` table with JWE compact serialization column
- [ ] Embed migrations with `//go:embed migrations/*.sql`

### Phase 4: Remove Hardcoded Secrets

- [ ] Remove hardcoded JWTSecret from `internal/cmd/learn/im/im.go`
- [ ] Implement barrier encryption for JWK storage
- [ ] Update user authentication to use encrypted JWKs from database
- [ ] Verify NO cleartext secrets in code or config files

### Phase 5: JWE Message Encryption

- [ ] Generate per-message JWKs using `internal/shared/crypto/jose`
- [ ] Encrypt messages with JWE Compact Serialization format
- [ ] Store encrypted messages in `messages` table
- [ ] Implement decryption using key ID lookup

### Phase 6: Manual Key Rotation Support

- [ ] Create admin API endpoint for manual key rotation
- [ ] Update active key ID on rotation
- [ ] Maintain historical keys for decryption
- [ ] Document rotation procedures

### Phase 7: Testing & Validation

- [ ] Unit tests for barrier encryption integration
- [ ] Unit tests for JWK generation
- [ ] Integration tests for message encryption/decryption
- [ ] E2E tests with Docker Compose
- [ ] Verify coverage ≥95% (production) / ≥98% (infrastructure)

---

## Architecture Decisions

ssd

### Shared Infrastructure Usage

**Barrier Layer Encryption**: Use `internal/shared/barrier` (already extracted for template/server).

**JWK Generation**: Use `internal/shared/crypto/jose` (shared infrastructure, no duplication).

**Federation Pattern**: Direct Go package imports for performance and simplicity (avoid network calls).

**Import Cycles**: One-way dependency only - learn-im imports shared infrastructure, NEVER the reverse.

### Database Schema

**Migration Pattern**: Same design as KMS - embedded SQL migrations with golang-migrate.

**Storage Format**: JWE Compact Serialization (`eyJ...`) in TEXT columns.

**Versioning**: Docker image tags provide versioning (no CLI version flag needed).

### Cryptographic Algorithms

| Table | Purpose | JWK Headers |
|-------|---------|-------------|
| `users_jwks` | Per-user encryption keys | `alg: ECDH-ES`, `enc: A256GCM` |
| `users_messages_jwks` | Per-user/message encryption keys | `alg: dir`, `enc: A256GCM` |
| `messages_jwks` | Per-message encryption keys | `alg: dir`, `enc: A256GCM` |
| `messages` | Encrypted message content | JWE Compact Serialization |

**Key Rotation**: Manual rotation via Admin API (on-demand, not time-based).

**ECDH P-256**: Key agreement derives shared secret used as AES-256 key for content encryption.

**Direct Encryption**: Use `alg: dir` (direct key agreement) with `enc: A256GCM` for all tables (simplified from AESGCMKW).

### Secret Management Rules - MANDATORY

**NO hardcoded secrets** in code or config files.

**NO cleartext secrets** in database.

**ONLY encrypted secrets** in database (barrier layer encryption).

**See**: `03-06.security.instructions.md` for Docker secrets pattern, `02-07.cryptography.instructions.md` for key hierarchy.
