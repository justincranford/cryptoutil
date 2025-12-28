# Service Template Refactoring - learn-im

## Implementation Checklist

### Phase 1: Package Structure Migration ✅

- [x] Move files from `internal/cmd/learn/im` to layered architecture directories
- [x] Update package imports (api/business/repository/util)
- [x] Verify build succeeds after package restructure
- [x] Run tests to detect broken imports

### Phase 2: Shared Infrastructure Integration ✅

- [x] Integrate `internal/shared/barrier` for barrier layer encryption
- [x] Integrate `internal/shared/crypto/jose` for JWK generation
- [x] Remove version flag from learn-im CLI parameters
- [x] Update `internal/learn/server/server.go` to initialize shared services

### Phase 3: Database Schema ⚠️ (4 tables → needs refactor to 3 tables)

- [x] Create migration files (same pattern as KMS)
- [x] Define `users` table with GORM models (PBKDF2 password hash)
- [x] Define `messages` table with JWE column
- [x] Define `messages_jwks` table with GORM models (OLD - to be removed)
- [x] Define `users_jwks` table with GORM models (OLD - to be removed)
- [x] Define `users_messages_jwks` table with GORM models (OLD - to be removed)
- [ ] **REFACTOR**: Define `messages_recipient_jwks` table (NEW - replaces messages_jwks)
- [ ] **REFACTOR**: Remove obsolete tables (users_jwks, users_messages_jwks, messages_jwks)
- [x] Embed migrations with `//go:embed migrations/*.sql`

### Phase 4: Remove Hardcoded Secrets ⚠️ (in progress)

- [x] Remove hardcoded JWTSecret from `internal/cmd/learn/im/im.go` (moved to Config)
- [ ] **TODO**: Implement barrier encryption for JWK storage (Phase 5b marker exists)
- [ ] **TODO**: Update user authentication to use encrypted JWKs from database
- [ ] **TODO**: Verify NO cleartext secrets in code or config files

### Phase 5: JWE Message Encryption ⚠️ (partial implementation)

- [x] Generate per-message JWKs using `internal/shared/crypto/jose`
- [x] Basic message encryption implementation exists (hybrid ECDH+AES-GCM)
- [ ] **REFACTOR**: Update to use new 3-table schema (messages, messages_recipient_jwks, users)
- [ ] **REFACTOR**: Use `EncryptBytesWithContext` and `DecryptBytesWithContext` from `internal/shared/crypto/jose/jwe_message_util.go`
- [ ] **REFACTOR**: Store encrypted JWKs in `messages_recipient_jwks` table
- [ ] **REFACTOR**: Implement multi-recipient encryption (N recipient AES256 JWKs)

### Phase 6: Manual Key Rotation Support ❌

- [ ] Create admin API endpoint for manual key rotation
- [ ] Update active key ID on rotation
- [ ] Maintain historical keys for decryption
- [ ] Document rotation procedures

### Phase 7: Testing & Validation ⚠️ (tests exist but failing)

- [ ] Unit tests for barrier encryption integration
- [x] Unit tests for JWK generation (exists via shared infrastructure)
- [x] Integration tests for message encryption/decryption (exists but timing issues)
- [x] E2E tests with Docker Compose (exists)
- [ ] **FIX**: Resolve test timeout issues (5/10 tests failing with "context deadline exceeded")
- [ ] Verify coverage ≥95% (production) / ≥98% (infrastructure)

---

## Architecture Decisions

### Shared Infrastructure Usage ✅ COMPLETED

**Barrier Layer Encryption**: ✅ Imported `internal/shared/barrier` (initialized but not yet used).

**JWK Generation**: ✅ Integrated `internal/shared/crypto/jose` for JWK generation.

**Federation Pattern**: ✅ Direct Go package imports (no network calls).

**Import Cycles**: ✅ One-way dependency confirmed - learn-im imports shared infrastructure.

### Database Schema - UPDATED TO 3 TABLES

**Migration Pattern**: ✅ Same design as KMS - embedded SQL migrations with golang-migrate.

**Storage Format**: JWE JSON format (NOT Compact Serialization) in TEXT columns.

**Versioning**: ✅ Docker image tags provide versioning (no CLI version flag).

### Cryptographic Algorithms - UPDATED DESIGN

| Table | Purpose | JWK Format | Encryption |
|-------|---------|------------|------------|
| `users` | User accounts | N/A | Password: PBKDF2-HMAC-SHA256 |
| `messages` | Encrypted messages | JWE JSON (multi-recipient) | `enc: A256GCM`, `alg: A256GCMKW` per recipient |
| `messages_recipient_jwks` | Per-recipient decryption keys | JWK JSON (encrypted) | `enc: A256GCM`, `alg: dir` |

**Multi-Recipient Pattern**:

- Each message encrypted with N recipient AES256 JWKs (one per RecipientUserID)
- Use `EncryptBytesWithContext(plaintext, []RecipientJWK)` → generates JWE with N encrypted keys
- Use `DecryptBytesWithContext(jwe, recipientJWK)` → decrypts using recipient's specific key

**Key Storage**:

- `messages.JWE`: JWE JSON format with N encrypted CEK copies (one per recipient)
- `messages_recipient_jwks.JWK`: Each recipient's decryption JWK (encrypted with `alg: dir`, `enc: A256GCM`)

**No Key Rotation**: Messages are encrypted once, immutable. No rotation needed (ephemeral per-message keys).

### Secret Management Rules - MANDATORY

**NO hardcoded secrets** in code or config files.

**NO cleartext secrets** in database.

**ONLY encrypted secrets** in database (barrier layer encryption).

**See**: `03-06.security.instructions.md` for Docker secrets pattern, `02-07.cryptography.instructions.md` for key hierarchy.

---

## Refactoring Tasks for 3-Table Design

### 1. Database Schema Updates

- [ ] Create new migration `0002_refactor_to_3_tables.up.sql`
- [ ] Drop tables: `users_jwks`, `users_messages_jwks`, `messages_jwks`
- [ ] Create table: `messages_recipient_jwks` (ID, RecipientUserID, MessageID, JWK encrypted)
- [ ] Update `messages` table: Remove `key_id` column, update JWE format to JSON (not Compact)
- [ ] Remove BLOB columns from `users` table: `public_key`, `private_key`

### 2. Domain Model Updates

- [ ] Delete `internal/learn/domain/jwk.go` (UserJWK, UserMessageJWK, MessageJWK structs)
- [ ] Create `internal/learn/domain/message_recipient_jwk.go` with new MessageRecipientJWK struct
- [ ] Update `internal/learn/domain/user.go`: Remove PublicKey, PrivateKey fields
- [ ] Update `internal/learn/domain/message.go`: Change JWECompact to JWE (JSON format), remove KeyID field

### 3. Repository Updates

- [ ] Create `internal/learn/repository/message_recipient_jwk_repository.go`
- [ ] Update `message_repository.go`: Add multi-recipient message creation
- [ ] Update `user_repository.go`: Remove key-related methods

### 4. Encryption Implementation

- [ ] Update `internal/learn/server/public.go`: Replace hybrid ECDH encryption with JWE multi-recipient
- [ ] Use `EncryptBytesWithContext(plaintext, []RecipientJWK)` for sending messages
- [ ] Use `DecryptBytesWithContext(jwe, recipientJWK)` for receiving messages
- [ ] Store encrypted recipient JWKs in `messages_recipient_jwks` table
- [ ] Remove in-memory key cache (`messageKeysCache sync.Map`)

### 5. Crypto Package Cleanup

- [ ] Update or remove `internal/learn/crypto/encrypt.go` (hybrid ECDH no longer used)
- [ ] Update `internal/learn/crypto/keygen.go` if needed for JWK generation
- [ ] Keep `internal/learn/crypto/password.go` (PBKDF2 still used)

### 6. Testing Updates

- [ ] Fix test timeout issues (increase HTTP client timeouts to 10s+)
- [ ] Update tests to use new 3-table schema
- [ ] Verify multi-recipient encryption/decryption works
- [ ] Add tests for `messages_recipient_jwks` table operations
