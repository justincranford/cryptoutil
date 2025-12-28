# Service Template QUIZME - learn-im Cleanup

**Purpose**: Clarify implementation details before starting Phase 8-11 cleanup tasks.

**Instructions**: Please answer each question below. For multiple choice, select A-D or write your custom answer in option E.

---

## Section 1: JWT Authentication Replacement

### Q1: How should JWT authentication be replaced in learn-im?

**Context**: Currently using hardcoded `jwtSecret` for JWT signing. Need to migrate to proper authentication.

**A)** Use session-based authentication with encrypted session cookies (JWE) stored in database
**B)** Use OAuth 2.1 client credentials flow with identity-authz federation
**C)** Use both: sessions for browser clients (`/browser/**`), OAuth tokens for service clients (`/service/**`)
**D)** Keep JWT but derive secret from barrier-encrypted JWK in database
**E)** Other (please specify):

**Your Answer**: D, you need config support for using JWTs as JWE or JWS or opaque; JWE and JWS are stateless options, opaque tokens require session storage in DB.

---

## Section 2: Password Hashing Migration

### Q2: Which hash provider should learn-im use for password hashing?

**Context**: Currently using custom PBKDF2 in `internal/learn/crypto/password.go`. Instructions say to use `hash_high_random_provider.go`.

**A)** `hash_high_random_provider.go` (high-entropy random hash with HKDF)
**B)** `hash_low_random_provider.go` (low-entropy random hash with PBKDF2)
**C)** `hash_high_deterministic_provider.go` (high-entropy deterministic hash)
**D)** `hash_low_deterministic_provider.go` (low-entropy deterministic hash)
**E)** Other (please specify):

**Your Answer**: B

**Follow-up Q2a**: Should `internal/learn/crypto/password.go` be removed entirely or kept for additional logic?

**Your Answer**: No, remove entirely and use shared hash provider directly in user service.

---

## Section 3: ServerSettings Integration

### Q3: How should learn-im integrate with `internal/shared/config/config.go` ServerSettings?

**Context**: learn-im has custom Config struct. Need to migrate to shared ServerSettings.

**A)** Replace entire learn-im Config struct with ServerSettings (remove all custom fields)
**B)** Embed ServerSettings in learn-im Config struct, add learn-im-specific fields as needed
**C)** Use ServerSettings for network/TLS only, keep separate AppConfig for business logic
**D)** Create adapter layer to map ServerSettings to learn-im's existing Config
**E)** Other (please specify):

**Your Answer**: C

**Follow-up Q3a**: Are there learn-im-specific config fields that don't belong in ServerSettings?

**Examples**: Message encryption settings, user validation rules, etc.

**Your Answer**:
AppConfig fields like JWE algorithm settings, message min/max settings, recipients min/max count settings

Service Template needs to offer username/password realm, in file and DB
Add Realms setting in ServerSettings; list of realm configuration files: 01-username-password-file.yml, 02-username-password-db.yml

- username and password min/max length settings
Add BrowserSessionCookie setting in ServerSettings; e.g. config file: browser-session-cookie.yml
- options include choosing cookie type: JWTs as JWE or JWS or opaque; default to JWS

---

## Section 4: Message Encryption Simplification

### Q4: Should `internal/learn/crypto/keygen.go` be removed or refactored?

**Context**: Instructions say to remove if ECDH is removed. Need clarification on ECDH usage.

**A)** Remove entirely - no longer needed (use shared JWK generation)
**B)** Keep but remove ECDH generation - keep other key generation logic
**C)** Refactor to wrapper around `internal/shared/crypto/jose` for learn-im-specific needs
**D)** Keep as-is for now - will be needed for future features
**E)** Other (please specify):

**Your Answer**: A

### Q5: Should `internal/learn/crypto/encrypt.go` be removed or refactored?

**Context**: Instructions say to use `jwe_message_util.go` instead of custom encryption.

**A)** Remove entirely - use `EncryptBytesWithContext` and `DecryptBytesWithContext` directly
**B)** Keep as thin wrapper around `jwe_message_util.go` for learn-im-specific error handling
**C)** Keep but significantly simplify - remove hybrid ECDH logic, keep utility functions
**D)** Keep as-is - provides valuable abstraction layer
**E)** Other (please specify):

**Your Answer**: A

---

## Section 5: UpdatedAt Field Usage

### Q6: Is the `UpdatedAt` field in domain models actually used?

**Context**: Found in `user.go` and `jwk.go` but only set in test code, not read anywhere.

**A)** Remove from all domain models and database schema (unused field)
**B)** Keep in User model (might be useful for audit), remove from JWK model
**C)** Keep in all models - GORM auto-updates it, useful for debugging
**D)** Add actual usage - display in user profile, track message edit history
**E)** Other (please specify):

**Your Answer**: D

---

## Section 6: File Splitting Strategy

### Q7: How should `public.go` (688 lines) be split?

**Context**: Violates 500-line hard limit. Need to split into multiple files.

**A)** 3 files: `handlers_auth.go` (register/login), `handlers_messages.go` (send/receive), `public.go` (server setup)
**B)** 4 files: Add `handlers_inbox.go` (inbox/sent/poll endpoints when implemented)
**C)** By layer: `routes.go` (route registration), `handlers.go` (HTTP handlers), `public.go` (server lifecycle)
**D)** By responsibility: `auth_handlers.go`, `message_handlers.go`, `middleware.go` (move from separate file), `server.go`
**E)** Other (please specify):

**Your Answer**: D

### Q8: How should `public_test.go` (2401 lines) be split?

**Context**: Violates 500-line hard limit by 4.8×. Needs aggressive splitting.

**A)** 5 files: `auth_test.go`, `messages_test.go`, `inbox_test.go`, `poll_test.go`, `test_helpers.go`
**B)** By test type: `unit_test.go`, `integration_test.go`, `helpers_test.go`
**C)** By feature: `register_test.go`, `login_test.go`, `send_test.go`, `receive_test.go`, `helpers_test.go`
**D)** Keep single file but convert to table-driven tests (should reduce size significantly)
**E)** Other (please specify):

**Your Answer**: B and C

---

## Section 7: Inbox/Sent/Poll API Design

### Q9: Should inbox/sent/poll APIs be part of Phase 8 or separate Phase 9?

**Context**: Instructions list these as Phase 9 feature enhancements. Could be implemented earlier.

**A)** Implement in Phase 9 (after all cleanup tasks complete)
**B)** Implement in Phase 8.9 (interleave with cleanup - implement while refactoring message handlers)
**C)** Implement as Phase 8.4.5 (part of file splitting - design APIs while splitting handlers)
**D)** Defer to Phase 12 (post-refactoring enhancements)
**E)** Other (please specify):

**Your Answer**: D

### Q10: How should the long poll API be implemented?

**Context**: Need real-time "you've got mail" notification without WebSockets.

**A)** In-memory channel per user (lost on restart, simple)
**B)** PostgreSQL LISTEN/NOTIFY (persistent, complex)
**C)** Database polling every 1-5 seconds (simple, higher load)
**D)** Redis pub/sub (requires Redis dependency, fast)
**E)** Other (please specify):

**Your Answer**: C

---

## Section 8: Integration Test Concurrency

### Q11: What are acceptable integration test runtime targets?

**Context**: Instructions suggest N=5, M=4, P=3, Q=2 for ~4 seconds. May need adjustment.

**A)** Target ~4 seconds (N=5, M=4, P=3, Q=2) - original suggestion
**B)** Target ~10 seconds (increase N/M/P/Q for more thorough testing)
**C)** Target ~2 seconds (reduce N/M/P/Q for faster CI/CD)
**D)** No hard target - optimize for maximum coverage within 15-second package limit
**E)** Other (please specify):

**Your Answer**: A

### Q12: Should integration tests use SQLite or PostgreSQL test-containers?

**Context**: Both are supported. SQLite faster, PostgreSQL more realistic.

**A)** SQLite only (faster, good enough for integration tests)
**B)** PostgreSQL test-containers only (slower, production-like)
**C)** Both - run tests twice with different databases
**D)** SQLite by default, PostgreSQL with `-tags=postgres` build tag
**E)** Other (please specify):

**Your Answer**: B

---

## Section 9: CLI Flag Testing Strategy

### Q13: How should CLI flag combinations be tested?

**Context**: Need to test `-d`, `-D <dsn>`, `-c learn.yml` modes.

**A)** Manual testing only (document commands, no automated tests)
**B)** Unit tests that verify flag parsing and config construction
**C)** Integration tests that start service with each flag combination
**D)** E2E tests that validate full service lifecycle with each mode
**E)** Other (please specify):

**Your Answer**: D

---

## Section 10: Test Coverage Targets

### Q14: What coverage targets should learn-im achieve?

**Context**: Instructions specify ≥95% production, ≥98% infrastructure. Is this realistic for learn-im?

**A)** Same as instructions: ≥95% production, ≥98% infrastructure (mandatory targets)
**B)** Relaxed: ≥90% production, ≥95% infrastructure (learn-im is demo service)
**C)** Strict: ≥98% production, ≥98% infrastructure (learn-im is template for other services)
**D)** Per-package targets: 95% domain/server, 90% crypto/util, 85% e2e
**E)** Other (please specify):

**Your Answer**: A

---

## Section 11: Magic Constants Organization

### Q15: Where should learn-im magic constants be defined?

**Context**: Need to move `MinUsernameLength`, `JWTIssuer`, etc. to magic package.

**A)** `internal/learn/magic/magic.go` (single file for all learn-im constants)
**B)** `internal/learn/magic/magic_server.go`, `magic_auth.go`, `magic_messages.go` (multiple files by category)
**C)** `internal/shared/magic/magic_learn.go` (shared magic package, learn-im section)
**D)** Keep in respective packages but export as constants (no separate magic package)
**E)** Other (please specify):

**Your Answer**: C

---

## Section 12: Dependency on Barrier Service

### Q16: What should learn-im do when barrier service integration is needed?

**Context**: Instructions mention barrier encryption for JWK storage but barrier service doesn't exist yet.

**A)** Add placeholder barrier interface, implement with simple AES256-GCM for now
**B)** Skip barrier encryption entirely - defer to Phase 12 (post-service-template work)
**C)** Implement barrier service as part of learn-im work (significant scope creep)
**D)** Use shared crypto utilities directly, migrate to barrier service when available
**E)** Other (please specify):

**Your Answer**: A for now; add phase at the end to extract barrier service from KMS to service template

---

## Section 13: Table-Driven Test Conversion

### Q17: How aggressive should table-driven test conversion be?

**Context**: Instructions say ALL tests must be table-driven. Current tests mix patterns.

**A)** Convert ALL tests to table-driven (no exceptions)
**B)** Convert only tests with multiple similar cases (skip one-off tests)
**C)** Convert happy path to table-driven, keep error cases as individual tests
**D)** Prioritize readability - use table-driven where it improves clarity, individual tests otherwise
**E)** Other (please specify):

**Your Answer**: D

---

## Section 14: Hardcoded Password Removal

### Q18: How should test passwords be generated?

**Context**: Tests currently use hardcoded passwords like "SecurePass123!". Need randomization.

**A)** Generate random passwords with `googleUuid.NewV7().String()` (simple, may not meet complexity rules)
**B)** Use test helper function `GenerateValidPassword()` (ensures complexity requirements met)
**C)** Use magic constant `TestPassword = "Test123!@#"` (still hardcoded but centralized)
**D)** Generate once per test file in `TestMain`, reuse across tests
**E)** Other (please specify): Use GenerateString in internal\shared\util\random\random.go

**Your Answer**:  E

---

## Section 15: E2E Test Scope

### Q19: What should learn-im E2E tests cover?

**Context**: Currently has basic E2E tests. Need to define comprehensive scope.

**A)** Full user journey: register → login → send message → receive message → logout
**B)** Multi-user scenarios: User A sends to User B, User B sends to User C, User C replies to User A
**C)** Edge cases: Invalid auth, message to non-existent user, concurrent sends, encryption failures
**D)** All of the above (comprehensive E2E coverage)
**E)** Other (please specify):

**Your Answer**: D

---

## Section 16: Docker Compose Configuration

### Q20: Should learn-im Docker Compose use shared telemetry services?

**Context**: KMS uses shared `telemetry/compose-telemetry.yml` for OTLP/Grafana.

**A)** Yes - reuse shared telemetry compose file (consistent with KMS pattern)
**B)** No - learn-im has its own telemetry setup (independence)
**C)** Optional - provide both standalone and shared telemetry variants
**D)** Defer - implement telemetry in Phase 12 (post-cleanup)
**E)** Other (please specify):

**Your Answer**: A

---

## Additional Notes

**Please add any additional clarifications, concerns, or questions below:**

---

**Submission Instructions**:

1. Fill in "Your Answer" for each question (A/B/C/D or custom E response)
2. Add any additional notes in the section above
3. Save this file and notify me when ready for review
4. I will use your answers to prioritize and implement Phase 8-11 tasks

**Thank you!**
