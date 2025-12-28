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

---

## Phase 8: Code Quality & Cleanup Tasks

### 8.1 Remove Deprecated Code and TODOs

- [ ] Remove `jwtSecret` field from `internal/learn/server/server.go` Config struct
- [ ] Remove `jwtSecret` parameter from `NewPublicServer()` function
- [ ] Remove `JWTSecret` constant from `internal/learn/server/middleware.go`
- [ ] Remove hardcoded `testJWTSecret` from `internal/learn/server/public_test.go`
- [ ] Remove hardcoded `jwtSecret` from test helper functions in `public_test.go`
- [ ] Remove comment "In-memory key cache for Phase 5a (no barrier service yet)."
- [ ] Remove all `TODO(Phase 5)` comments from learn-im codebase
- [ ] Remove `messageKeysCache sync.Map` in-memory cache from `public.go`

### 8.2 Move Magic Constants to Magic Package

- [ ] Define `MinUsernameLength` in `internal/learn/magic/magic.go`
- [ ] Define `MaxUsernameLength` in `internal/learn/magic/magic.go`
- [ ] Define `MinPasswordLength` in `internal/learn/magic/magic.go`
- [ ] Define `JWTIssuer` in `internal/learn/magic/magic.go`
- [ ] Define `JWTExpiration` in `internal/learn/magic/magic.go`
- [ ] Update references to use magic constants instead of literals

### 8.3 Use Shared Crypto Infrastructure

- [ ] Replace custom password hashing with `internal/shared/crypto/hash/hash_high_random_provider.go`
- [ ] Use `HashPasswordWithContext()` for user registration
- [ ] Use `VerifyPasswordWithContext()` for user login
- [ ] Remove or simplify `internal/learn/crypto/password.go` if duplicating shared infrastructure

### 8.4 File Size Limit Violations (300/400/500 lines)

**public.go (688 lines - CRITICAL VIOLATION)**:

- [ ] Split `internal/learn/server/public.go` into smaller files (target <400 lines per file)
  - [ ] Create `internal/learn/server/handlers_auth.go` (register, login handlers)
  - [ ] Create `internal/learn/server/handlers_messages.go` (send, receive handlers)
  - [ ] Keep shared server setup in `public.go`

**public_test.go (2401 lines - CRITICAL VIOLATION)**:

- [ ] Split `internal/learn/server/public_test.go` into smaller test files (target <500 lines per file)
  - [ ] Create `internal/learn/server/auth_test.go` (register, login tests)
  - [ ] Create `internal/learn/server/messages_test.go` (send, receive tests)
  - [ ] Create `internal/learn/server/test_helpers.go` (shared test utilities)
  - [ ] Remove all hardcoded passwords - generate random passwords in tests

**learn_im_e2e_test.go (782 lines - VIOLATION)**:

- [ ] Split `internal/learn/e2e/learn_im_e2e_test.go` into smaller E2E test files (target <500 lines)
  - [ ] Create `internal/learn/e2e/auth_e2e_test.go` (authentication E2E tests)
  - [ ] Create `internal/learn/e2e/messages_e2e_test.go` (messaging E2E tests)
  - [ ] Create `internal/learn/e2e/helpers_e2e_test.go` (shared E2E helpers)

### 8.5 Convert All Tests to Table-Driven

- [ ] Refactor `internal/learn/server/public_test.go` - ensure ALL tests are table-driven
- [ ] Refactor `internal/learn/crypto/encrypt_test.go` - ensure table-driven
- [ ] Refactor `internal/learn/crypto/password_test.go` - ensure table-driven
- [ ] Verify each test uses `t.Run(tt.name, ...)` pattern with test cases slice

### 8.6 Use ServerSettings Struct

- [ ] Update `internal/learn/server/server.go` to use `internal/shared/config/config.go` ServerSettings
- [ ] Remove custom learn-im config struct - reuse shared ServerSettings
- [ ] Update CLI flags in `cmd/learn-im/main.go` to map to ServerSettings fields
- [ ] Ensure backward compatibility with existing config patterns

### 8.7 Simplify Crypto Package

- [ ] Remove `internal/learn/crypto/keygen.go` (ECDH generation no longer needed)
- [ ] Remove `internal/learn/crypto/encrypt.go` (use shared JWE utilities instead)
- [ ] Use `internal/shared/crypto/jose/jwe_message_util.go`:
  - [ ] Use `EncryptBytesWithContext()` for message encryption
  - [ ] Use `DecryptBytesWithContext()` for message decryption
- [ ] Keep `internal/learn/crypto/password.go` only if it adds value over shared hash provider

### 8.8 Remove UpdatedAt if Unused

- [ ] Verify if `UpdatedAt` field in `internal/learn/domain/user.go` is actually used
- [ ] Verify if `UpdatedAt` field in `internal/learn/domain/jwk.go` is actually used
- [ ] If unused, remove `UpdatedAt` from domain models and database schema
- [ ] If used, document the use case and keep it

---

## Phase 9: Feature Enhancements

### 9.1 Message Listing APIs

- [ ] Implement `/service/api/v1/messages/inbox` - list received messages for authenticated user
  - [ ] Support query parameter `?limit=N` (default 50, max 1000)
  - [ ] Support query parameter `?offset=N` (pagination)
  - [ ] Support query parameter `?read=true|false` (filter by read status)
  - [ ] Support query parameter `?from=<username>` (filter by sender)
  - [ ] Support query parameter `?sort=created_at:desc|asc` (sort order)
- [ ] Implement `/service/api/v1/messages/sent` - list sent messages for authenticated user
  - [ ] Support query parameter `?limit=N` (default 50, max 1000)
  - [ ] Support query parameter `?offset=N` (pagination)
  - [ ] Support query parameter `?to=<username>` (filter by recipient)
  - [ ] Support query parameter `?sort=created_at:desc|asc` (sort order)
- [ ] Add database indexes for efficient querying (user_id, created_at, read status)
- [ ] Unit tests for inbox/sent listing with filters and pagination
- [ ] Integration tests for inbox/sent APIs

### 9.2 Long Poll API ("You've Got Mail")

- [ ] Implement `/service/api/v1/messages/poll` - long poll endpoint for new messages
  - [ ] Accept query parameter `?timeout=<seconds>` (default 30s, max 60s)
  - [ ] Return immediately if unread messages exist
  - [ ] Block up to timeout seconds waiting for new messages
  - [ ] Return HTTP 200 with message count when new messages arrive
  - [ ] Return HTTP 204 (No Content) on timeout with no new messages
- [ ] Implement in-memory notification channel (per user ID)
- [ ] Trigger notification on `/service/api/v1/messages/tx` (send message)
- [ ] Unit tests for long poll endpoint (immediate return, timeout, notification)
- [ ] Integration tests for concurrent long poll clients

### 9.3 Concurrency Integration Tests

- [ ] Create `internal/learn/integration/concurrent_test.go` for robustness testing
  - [ ] Test with N=5 users, M=4 concurrent sends (1 recipient each)
  - [ ] Test with N=5 users, P=3 concurrent sends (2 recipients each)
  - [ ] Test with N=5 users, Q=2 concurrent sends (all recipients broadcast)
  - [ ] Verify all messages correctly encrypted/decrypted
  - [ ] Verify no race conditions or data corruption
  - [ ] Target runtime ~4 seconds (adjust N/M/P/Q if needed)
- [ ] Add table-driven test structure for different concurrency scenarios
- [ ] Verify proper transaction isolation and locking

---

## Phase 10: Testing & Validation Commands

### 10.1 Unit Tests

**Run Command**:

```bash
go test ./internal/learn/... -short -coverprofile=./test-output/coverage_learn_unit.out
go tool cover -html=./test-output/coverage_learn_unit.out -o ./test-output/coverage_learn_unit.html
```

**Expected Results**:

- [ ] All unit tests pass (0 failures)
- [ ] Coverage ≥95% for production code (handlers, domain logic)
- [ ] Coverage ≥98% for infrastructure code (repositories, crypto utilities)
- [ ] No test timeouts or flakiness

### 10.2 Integration Tests

**Run Command**:

```bash
go test ./internal/learn/integration/... -coverprofile=./test-output/coverage_learn_integration.out
go tool cover -html=./test-output/coverage_learn_integration.out -o ./test-output/coverage_learn_integration.html
```

**Expected Results**:

- [ ] All integration tests pass (0 failures)
- [ ] Concurrent message tests complete in ~4 seconds
- [ ] No race conditions detected with `-race` flag
- [ ] Coverage includes repository and database interactions

### 10.3 Docker Compose (Development Environment)

**Start Command**:

```bash
docker compose -f deployments/learn/compose.yml up -d
```

**Use Commands**:

```bash
# Check service health
docker compose -f deployments/learn/compose.yml ps

# View logs
docker compose -f deployments/learn/compose.yml logs -f learn-im

# Test API endpoints
curl -k https://localhost:8888/service/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"SecurePass123!"}'
```

**Stop Command**:

```bash
docker compose -f deployments/learn/compose.yml down
```

**Verification**:

- [ ] Service starts without errors
- [ ] Health check endpoint returns HTTP 200
- [ ] Registration and login APIs functional
- [ ] Message send/receive APIs functional

### 10.4 Demo Application

**Start Command**:

```bash
go run ./cmd/learn-im -d
```

**Use Commands**:

```bash
# In separate terminal, test with curl or Postman
curl -k https://localhost:8888/service/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"username":"bob","password":"SecurePass456!"}'
```

**Stop Command**:

```bash
# Ctrl+C or send SIGTERM
kill <PID>
```

**Verification**:

- [ ] SQLite in-memory mode works with `-d` flag
- [ ] All APIs functional in dev mode
- [ ] Graceful shutdown works

### 10.5 E2E Tests

**Run Command**:

```bash
go test ./internal/learn/e2e/... -coverprofile=./test-output/coverage_learn_e2e.out
go tool cover -html=./test-output/coverage_learn_e2e.out -o ./test-output/coverage_learn_e2e.html
```

**Expected Results**:

- [ ] All E2E tests pass (0 failures)
- [ ] Docker containers start/stop correctly
- [ ] Full message encryption/decryption workflow validated
- [ ] Multi-user scenarios work correctly

---

## Phase 11: CLI Flag Testing

### 11.1 Test with `-d` (SQLite Dev Mode)

**Command**:

```bash
go run ./cmd/learn-im -d
```

**Verification**:

- [ ] Uses SQLite in-memory database
- [ ] All default settings applied (bind addresses, ports, TLS)
- [ ] No external dependencies required (PostgreSQL, Docker)
- [ ] Service starts and handles requests correctly

### 11.2 Test with `-D <dsn>` (PostgreSQL Test-Container)

**Command**:

```bash
# Starts test-container automatically
go test ./internal/learn/integration/... -v
```

**Verification**:

- [ ] Test-container PostgreSQL instance starts automatically
- [ ] Unique database name generated (UUID-based)
- [ ] All tests use PostgreSQL instead of SQLite
- [ ] Test-container cleaned up after tests complete

### 11.3 Test with `-c learn.yml` (Production-like Config)

**Command**:

```bash
go run ./cmd/learn-im -c configs/learn/learn.yml
```

**Config File Requirements**:

- [ ] OTLP service name configured (e.g., `learn-im-1`)
- [ ] OTLP endpoint configured (e.g., `opentelemetry-collector:4317`)
- [ ] Database URL configured (PostgreSQL DSN or SQLite file path)
- [ ] TLS certificates configured
- [ ] Bind addresses and ports configured

**Verification**:

- [ ] Service starts with production-like settings
- [ ] OTLP telemetry exported correctly
- [ ] TLS certificates loaded and validated
- [ ] All configuration values override defaults

---

## Progress Tracking

**Last Updated**: 2025-12-28

**Overall Status**: 🟡 In Progress

- ✅ **Phase 1-2 Complete**: Package structure migration, shared infrastructure integration
- ⚠️ **Phase 3-7 Partial**: Database schema needs 3-table refactor, tests have timeout issues
- ❌ **Phase 8-11 Not Started**: Code quality cleanup, feature enhancements, comprehensive testing

**Critical Blockers**:

1. File size violations (public.go: 688 lines, public_test.go: 2401 lines, e2e: 782 lines)
2. Test timeout issues (5/10 tests failing with "context deadline exceeded")
3. jwtSecret still hardcoded in multiple places (violates security requirements)
4. Custom crypto instead of shared infrastructure (violates DRY principle)

**Next Steps**:

1. Address file size violations by splitting large files
2. Remove jwtSecret and use proper authentication infrastructure
3. Replace custom crypto with shared JWE utilities
4. Implement inbox/sent listing APIs
5. Add long poll API for real-time notifications
6. Add comprehensive concurrency integration tests
