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

### QUIZME Answers Integration

**JWT Authentication (Q1)**: Config support for JWT format - JWE (encrypted), JWS (signed), or opaque tokens. JWE/JWS are stateless, opaque requires session storage in database.

**Password Hashing (Q2)**: Use `internal/shared/crypto/hash/hash_low_random_provider.go` (PBKDF2-HMAC-SHA256 for low-entropy passwords).

**ServerSettings Integration (Q3)**:

- Use ServerSettings for network/TLS configuration
- Create separate AppConfig for learn-im business logic
- Add Realms setting in ServerSettings for username/password auth (file-based and DB-based)
- Add BrowserSessionCookie setting in ServerSettings for cookie configuration (JWE/JWS/opaque)

**Crypto Cleanup (Q4-Q5)**: Remove `keygen.go` and `encrypt.go` entirely, use shared JWK generation and JWE utilities.

**UpdatedAt Field (Q6)**: Keep UpdatedAt and add actual usage (display in user profile, track message edit history).

**File Splitting (Q7-Q8)**:

- public.go: Split by responsibility into `auth_handlers.go`, `message_handlers.go`, `middleware.go`, `server.go`
- public_test.go: Split by test type (unit/integration) AND by feature (register/login/send/receive/helpers)

**Inbox/Poll APIs (Q9-Q10)**: Defer to Phase 12, use database polling every 1-5 seconds.

**Test Configuration (Q11-Q12)**: Target ~4 seconds runtime (N=5, M=4, P=3, Q=2), use PostgreSQL test-containers only.

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

### 8.2 Move Magic Constants to Magic Package - ✅ COMPLETE

- [x] Define `MinUsernameLength` in `internal/learn/magic/magic.go`
- [x] Define `MaxUsernameLength` in `internal/learn/magic/magic.go`
- [x] Define `MinPasswordLength` in `internal/learn/magic/magic.go`
- [x] Define `JWTIssuer` in `internal/learn/magic/magic.go`
- [x] Define `JWTExpiration` in `internal/learn/magic/magic.go`
- [x] Update references to use magic constants instead of literals

### 8.3 Use Shared Crypto Infrastructure

- [ ] Replace custom password hashing with `internal/shared/crypto/hash/hash_high_random_provider.go`
- [ ] Use `HashPasswordWithContext()` for user registration
- [ ] Use `VerifyPasswordWithContext()` for user login
- [ ] Remove or simplify `internal/learn/crypto/password.go` if duplicating shared infrastructure

### 8.4 File Size Limit Violations (300/400/500 lines)

**public.go (688 lines - CRITICAL VIOLATION)** - ✅ COMPLETE:

- [x] Split `internal/learn/server/public.go` into smaller files by responsibility:
  - [x] Create `internal/learn/server/auth_handlers.go` (200 lines - register, login handlers)
  - [x] Create `internal/learn/server/message_handlers.go` (295 lines - send, receive handlers)
  - [x] Create `internal/learn/server/public_server.go` (250 lines - server lifecycle and setup)

**public_test.go (2162 lines - CRITICAL VIOLATION)** - ✅ COMPLETE:

- [x] Extract test helpers to `internal/learn/server/helpers_test.go` (shared utilities, mock setup)
- [x] Split remaining tests by feature category:
  - [x] Create `internal/learn/server/register_test.go` (registration feature tests - 8 tests)
  - [x] Create `internal/learn/server/login_test.go` (login feature tests - 8 tests)
  - [x] Create `internal/learn/server/send_test.go` (message send tests)
  - [x] Create `internal/learn/server/receive_delete_test.go` (receive and delete message tests)

**learn_im_e2e_test.go (783 lines - CRITICAL VIOLATION)** - ✅ COMPLETE:

- [x] Split `internal/learn/e2e/learn_im_e2e_test.go` into 3 category files:
  - [x] Create `internal/learn/e2e/helpers_e2e_test.go` (shared E2E utilities - 480 lines)
  - [x] Create `internal/learn/e2e/browser_e2e_test.go` (browser path E2E tests - 4 tests, ~230 lines)
  - [x] Create `internal/learn/e2e/service_e2e_test.go` (service path E2E tests - 3 tests, ~175 lines)
- [x] Verify all E2E tests compile with no errors
- [x] Commit with descriptive message referencing Phase 8.4
  - [x] Create `internal/learn/server/middleware_test.go` (JWT middleware tests - 3 tests)
  - [x] Create `internal/learn/server/server_lifecycle_test.go` (server lifecycle and unit tests - 7 tests)
- [x] Delete original `public_test.go` to eliminate duplicate declarations
- [x] Run `go test ./internal/learn/server` to verify compilation
- [x] Fix linting errors (replace nil context with context.TODO())
- [x] Commit and push with descriptive message referencing Phase 8.4

**learn_im_e2e_test.go (782 lines - VIOLATION)** - ❌ TODO:

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

- [ ] Remove `internal/learn/crypto/keygen.go` entirely (use shared JWK generation)
- [ ] Remove `internal/learn/crypto/encrypt.go` entirely (use shared JWE utilities)
- [ ] Use `internal/shared/crypto/jose/jwe_message_util.go`:
  - [ ] Use `EncryptBytesWithContext()` for message encryption
  - [ ] Use `DecryptBytesWithContext()` for message decryption
- [ ] Migrate password hashing to use `internal/shared/crypto/hash/hash_low_random_provider.go`
- [ ] Remove `internal/learn/crypto/password.go` after migration complete

### 8.8 Implement UpdatedAt Field Usage

- [ ] Keep `UpdatedAt` field in `internal/learn/domain/user.go` and add actual usage
- [ ] Display UpdatedAt in user profile endpoint (when implemented)
- [ ] Use UpdatedAt for tracking last login time
- [ ] Keep `UpdatedAt` field in `internal/learn/domain/message.go` for message edit history
- [ ] Document UpdatedAt usage in domain model comments

---

## Phase 9: Concurrency Integration Tests

### 9.1 Concurrency Integration Tests

- [ ] Create `internal/learn/integration/concurrent_test.go` for robustness testing
  - [ ] Test with N=5 users, M=4 concurrent sends (1 recipient each) - target ~4s
  - [ ] Test with N=5 users, P=3 concurrent sends (2 recipients each)
  - [ ] Test with N=5 users, Q=2 concurrent sends (all recipients broadcast)
  - [ ] Verify all messages correctly encrypted/decrypted
  - [ ] Verify no race conditions or data corruption
  - [ ] Use PostgreSQL test-containers for all integration tests
- [ ] Add table-driven test structure for different concurrency scenarios
- [ ] Verify proper transaction isolation and locking
- [ ] Run with `-race` flag to detect race conditions

---

## Phase 10: ServerSettings Integration

### 10.1 Add ServerSettings Extensions

- [ ] Add Realms setting in `internal/shared/config/config.go` ServerSettings:
  - [ ] Support username/password realm configuration files
  - [ ] Example: `01-username-password-file.yml` for file-based auth
  - [ ] Example: `02-username-password-db.yml` for database-based auth
  - [ ] Add username/password min/max length settings per realm
- [ ] Add BrowserSessionCookie setting in ServerSettings:
  - [ ] Support cookie type configuration: JWE (encrypted), JWS (signed), opaque (database)
  - [ ] Example config file: `browser-session-cookie.yml`
  - [ ] Default to JWS (signed stateless tokens)
  - [ ] JWE/JWS are stateless, opaque requires session storage in DB

### 10.2 Update learn-im Config

- [ ] Create `internal/learn/server/config.go` with AppConfig struct
- [ ] Keep learn-im-specific settings in AppConfig:
  - [ ] JWE algorithm settings
  - [ ] Message min/max length settings
  - [ ] Recipients min/max count settings
- [ ] Embed ServerSettings in AppConfig for network/TLS configuration
- [ ] Update CLI flags to support both ServerSettings and AppConfig fields

---

## Phase 11: Testing & Validation Commands

### 11.1 Unit Tests

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

### 11.2 Integration Tests

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

### 11.3 Docker Compose (Development Environment)

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

### 11.4 Demo Application

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

### 11.5 E2E Tests

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

## Phase 12: CLI Flag Testing

### 12.1 Test with `-d` (SQLite Dev Mode)

**Command**:

```bash
go run ./cmd/learn-im -d
```

**Verification**:

- [ ] Uses SQLite in-memory database
- [ ] All default settings applied (bind addresses, ports, TLS)
- [ ] No external dependencies required (PostgreSQL, Docker)
- [ ] Service starts and handles requests correctly

### 12.2 Test with `-D <dsn>` (PostgreSQL Test-Container)

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

### 12.3 Test with `-c learn.yml` (Production-like Config)

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

## Phase 13: Future Enhancements (Deferred)

### 13.1 Message Listing APIs

**Note**: Deferred from Phase 9 per QUIZME Q9 answer.

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

### 13.2 Long Poll API ("You've Got Mail")

**Note**: Deferred from Phase 9 per QUIZME Q9-Q10 answers. Implementation uses database polling.

- [ ] Implement `/service/api/v1/messages/poll` - long poll endpoint for new messages
  - [ ] Accept query parameter `?timeout=<seconds>` (default 30s, max 60s)
  - [ ] Return immediately if unread messages exist
  - [ ] Block up to timeout seconds waiting for new messages (database polling every 1-5 seconds)
  - [ ] Return HTTP 200 with message count when new messages arrive
  - [ ] Return HTTP 204 (No Content) on timeout with no new messages
- [ ] Implement database polling mechanism (check messages table every 1-5 seconds)
- [ ] Trigger notification on `/service/api/v1/messages/tx` (send message)
- [ ] Unit tests for long poll endpoint (immediate return, timeout, notification)
- [ ] Integration tests for concurrent long poll clients

---

## Progress Tracking

**Last Updated**: 2025-12-28

**Overall Status**: 🟡 In Progress - QUIZME Answered, Ready for Implementation

- ✅ **Phase 1-2 Complete**: Package structure migration, shared infrastructure integration
- ✅ **QUIZME Complete**: All 12 questions answered, architecture decisions documented
- ⚠️ **Phase 3-7 Partial**: Database schema needs 3-table refactor, tests have timeout issues
- ❌ **Phase 8-12 Not Started**: Code quality cleanup, concurrency tests, ServerSettings integration, CLI testing
- ❌ **Phase 13 Deferred**: Inbox/sent listing and long poll APIs (future enhancements)

**Critical Blockers**:

1. File size violations (public.go: 688 lines, public_test.go: 2401 lines, e2e: 782 lines)
2. Test timeout issues (5/10 tests failing with "context deadline exceeded")
3. jwtSecret still hardcoded in multiple places (violates security requirements)
4. Custom crypto instead of shared infrastructure (violates DRY principle)

**Next Steps** (Prioritized):

1. **Phase 8.4**: Split large files (public.go → 4 files, public_test.go → 7+ files, e2e → 3 files)
2. **Phase 8.1**: Remove jwtSecret and deprecated code
3. **Phase 8.7**: Replace custom crypto with shared hash/JWE utilities
4. **Phase 8.3**: Migrate password hashing to shared `hash_low_random_provider.go`
5. **Phase 10**: Add ServerSettings extensions (Realms, BrowserSessionCookie)
6. **Phase 9**: Implement concurrency integration tests (N=5, M=4, P=3, Q=2, target ~4s)
7. **Phase 11-12**: Testing and CLI flag validation
8. **Phase 13**: Future - inbox/sent listing, long poll API (database polling)
