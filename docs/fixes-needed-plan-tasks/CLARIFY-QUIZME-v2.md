# CLARIFY-QUIZME-v2: Implementation Details

**Purpose**: Identify remaining tactical implementation details for jose-ja service-template refactoring.

**Context**: QUIZME-v1 answered all strategic architecture questions (100% complete per Q10.2). This v2 focuses on tactical implementation details discovered during plan/tasks updates.

**Instructions**:

- Answer with A/B/C/D or write-in for E
- Write-in answers provide detailed rationale
- Mark unanswered questions with ❓ for discussion
- Reference QUIZME-v1 answers where relevant

---

## Section 1: pending_users Schema and Constraints

### Q1.1: pending_users Username Uniqueness

**Question**: Should `username` be unique across BOTH `pending_users` and `users` tables?

**Context**: If same username can exist in both tables, user could register while their approval is pending (creating duplicate pending entries). If enforced unique across tables, prevents duplicate registrations.

**Options**:

A. **Unique within pending_users only** - Same username can exist in both tables (pending + users)
B. **Unique across pending_users + users** - Database constraint prevents duplicates
C. **Unique per tenant within pending_users** - Same username allowed for different tenants
D. **No uniqueness constraint** - Application logic handles duplicates
E. **Write-in**: Unique per tenant across users and pending_users

**Answer**: E

---

### Q1.2: pending_users Indexes

**Question**: What indexes are needed on `pending_users` table?

**Context**: Queries: (1) Admin dashboard lists pending users, (2) Login checks if user exists in pending_users before users table, (3) Approval workflow updates by id, (4) Cleanup job deletes rejected/expired entries.

**Options**:

A. **PRIMARY KEY (id) only** - Minimal indexes
B. **PRIMARY KEY (id), INDEX (username)** - Fast login checks
C. **PRIMARY KEY (id), INDEX (username), INDEX (tenant_id), INDEX (status)** - All query patterns
D. **PRIMARY KEY (id), UNIQUE INDEX (username, tenant_id), INDEX (status, requested_at)** - Composite indexes
E. **Write-in**: you figure it out!!!

**Answer**: E

---

### Q1.3: pending_users Email Validation

**Question**: Is email validation required for username field during registration?

**Context**: If username must be valid email (user@example.com), simplifies password reset workflow. If username can be non-email (johndoe), requires separate email field.

**Options**:

A. **Username must be valid email** - Validate email format, send verification email
B. **Username can be non-email** - Add separate `email` field to pending_users
C. **Username email validation optional** - Accept both email and non-email usernames
D. **Username email validation per realm** - Email realm requires email, password realm allows non-email
E. **Write-in**: Email/password is a different realm, not implemented yet

**Answer**:  E

---

### Q1.4: pending_users Expiration

**Question**: Should pending_users entries expire after N days without approval/rejection?

**Context**: Old pending entries could accumulate indefinitely. Expiration prevents database bloat, but requires notification before expiration.

**Options**:

A. **No expiration** - Entries remain until explicitly approved/rejected
B. **30-day expiration** - Auto-delete after 30 days, send email reminder at 7 days
C. **90-day expiration** - Auto-delete after 90 days, no reminders
D. **Configurable expiration** - YAML config: `pending_users_expiration_days: 30`
E. **Write-in**: D, but specify in hours not days

**Answer**:  E

---

## Section 2: Approval Workflow UX

### Q2.1: Admin Dashboard for Join Requests

**Question**: Should admin dashboard API be part of template infrastructure or domain-specific?

**Context**: All 9 services need admin approval workflow. Template could provide `/admin/api/v1/pending-users` GET/POST endpoints, or each service implements custom dashboard.

**Options**:

A. **Template infrastructure** - All 9 services get same admin dashboard APIs
B. **Domain-specific** - Each service implements custom admin dashboard (jose-ja, cipher-im, etc.)
C. **Hybrid** - Template provides base GET/POST, domain adds custom fields/validation
D. **External service** - Separate admin-dashboard service manages all 9 services
E. **Write-in**:

**Answer**: A

---

### Q2.2: Email Notifications

**Question**: Should email notifications be sent on approval/rejection?

**Context**: User experience improved with email notifications ("Your registration was approved!"). Requires SMTP configuration and email templates.

**Options**:

A. **No email notifications** - Users poll status via login attempts (HTTP 403/401 responses)
B. **Email notifications with SMTP config** - Template infrastructure with configurable SMTP
C. **Email notifications via external service** - Queue email jobs to external email service
D. **Email notifications optional per deployment** - YAML config: `email_notifications_enabled: true/false`
E. **Write-in**:

**Answer**: A

---

### Q2.3: Webhook Callbacks

**Question**: Should registration approval/rejection trigger webhooks for external integrations?

**Context**: Enterprise deployments may want to trigger workflows (Slack notifications, SIEM logging, provisioning systems) on user registration events.

**Options**:

A. **No webhooks** - Keep template simple
B. **Webhooks with configurable URLs** - YAML config: `webhook_url: https://example.com/hooks/user-approved`
C. **Webhooks with retry logic** - Template provides reliable delivery with exponential backoff
D. **Webhooks via event bus** - Template publishes events to NATS/Kafka/Redis Streams
E. **Write-in**:

**Answer**: A

---

### Q2.4: Status Polling API

**Question**: Should unauthenticated users be able to poll registration status?

**Context**: User registers, receives HTTP 403 on login. Without status API, user must keep trying login (wasteful). Status API allows polling without full authentication.

**Options**:

A. **No status API** - User polls via login attempts (HTTP 403 = pending, HTTP 401 = rejected, HTTP 200 = approved)
B. **Status API with registration ID** - `/browser/api/v1/registration/:id/status` returns pending/approved/rejected
C. **Status API with username** - `/browser/api/v1/registration/status?username=user@example.com` (security risk?)
D. **Status API with magic link** - Email contains unique status check URL (most secure)
E. **Write-in**:

**Answer**: A

---

## Section 3: Rate Limiting Strategy

### Q3.1: Rate Limiting Scope

**Question**: What should registration rate limiting measure?

**Context**: Prevent abuse (spam registrations, brute force username enumeration). Rate limit per IP address, per username, or both?

**Options**:

A. **Per IP address only** - 10 registrations per IP per hour
B. **Per username only** - 3 registrations per username per hour (prevent username squatting)
C. **Both IP and username** - 10/hour per IP AND 3/hour per username (most restrictive)
D. **Per IP with CAPTCHA after threshold** - 10/hour per IP, require CAPTCHA after 3 attempts
E. **Write-in**:

**Answer**: A

---

### Q3.2: Rate Limiting Storage

**Question**: Where should rate limiting state be stored?

**Context**: In-memory (simple, not distributed), Redis (distributed, requires dependency), PostgreSQL (leverages existing DB, slower).

**Options**:

A. **In-memory (sync.Map)** - Simple, single-node only, lost on restart
B. **Redis** - Distributed, requires Redis dependency (adds complexity)
C. **PostgreSQL** - Leverages existing DB, slower than Redis but consistent
D. **SQLite** - For SQLite deployments use in-memory, PostgreSQL deployments use PostgreSQL
E. **Write-in**:

**Answer**: A

---

### Q3.3: Rate Limiting Thresholds

**Question**: What are acceptable registration rate limits?

**Context**: Too strict prevents legitimate registrations (large organizations), too lenient allows abuse.

**Options**:

A. **10 per hour per IP** - Strict, suitable for small deployments
B. **100 per hour per IP** - Lenient, suitable for large organizations
C. **10 per minute per IP, 100 per hour per IP** - Burst allows quick legitimate use, hour limit prevents sustained abuse
D. **Configurable** - YAML: `registration_rate_limit_per_hour: 100`
E. **Write-in**:

**Answer**: D; low default

---

## Section 4: Hash Service Configuration

### Q4.1: PBKDF2 Iterations

**Question**: What PBKDF2 iteration count should be used for password hashing?

**Context**: OWASP recommends 600,000 iterations for PBKDF2-HMAC-SHA256 (2023). Higher iterations = more secure but slower registration/login.

**Options**:

A. **100,000 iterations** - Fast, minimum acceptable for 2025
B. **200,000 iterations** - Balanced, NIST SP 800-132 recommendation
C. **600,000 iterations** - Secure, OWASP 2023 recommendation
D. **Configurable** - YAML: `pbkdf2_iterations: 600000`
E. **Write-in**: It's 610,000 and its already implemented in hash service. Look it up there.

**Answer**: E

---

### Q4.2: Pepper Rotation Strategy

**Question**: How should pepper rotation be handled when upgrading hash policies?

**Context**: QUIZME-v1 Q5.2 answered "global pepper for all tenants". When rotating pepper (e.g., security incident), need migration strategy.

**Options**:

A. **No rotation** - Pepper is permanent, change only on security incident (manual intervention)
B. **Lazy migration** - New hashes use new pepper, old hashes remain on old pepper (version prefix identifies pepper)
C. **Forced re-hash** - All users must reset password on pepper rotation (disruptive)
D. **Dual-pepper support** - Support old + new pepper simultaneously, re-hash on next login
E. **Write-in**: Look up current implementation in hash service for pepper rotation strategy.

**Answer**:  E

---

### Q4.3: Hash Algorithm Versioning

**Question**: Should hash service support multiple hash algorithm versions simultaneously?

**Context**: Future-proofing for algorithm upgrades (PBKDF2 → Argon2 when FIPS-approved, SHA-256 → SHA-512). Version prefix in hash output enables gradual migration.

**Options**:

A. **Single version only** - All hashes use same algorithm, upgrade requires forced re-hash
B. **Multiple versions supported** - Hash output format: `{version}:{algorithm}:{iterations}:salt:hash`
C. **Automatic migration** - Old hashes upgraded to new version on next login (lazy migration)
D. **Configurable current version** - YAML: `hash_service.current_version: 2`
E. **Write-in**: C, Look up current implementation in hash service for hash algorithm versioning.

**Answer**: E

---

### Q4.4: Per-Tenant Iteration Count

**Question**: Should iteration count be configurable per tenant?

**Context**: Some tenants may want higher security (1M iterations), others prioritize performance (100k iterations).

**Options**:

A. **Global iteration count** - All tenants use same value (simplest)
B. **Per-tenant iteration count** - Tenants table has `pbkdf2_iterations` column
C. **Per-tenant hash policy** - Tenants table has `hash_policy_version` column referencing policy configurations
D. **No per-tenant configuration** - Security policy should be consistent across all tenants
E. **Write-in**: D, security policy is part of the hash service configuration, not per-tenant

**Answer**: E

---

## Section 5: Migration Sequencing

### Q5.1: pending_users and tenant_join_requests Transaction

**Question**: Can pending_users (1005) and tenant_join_requests (1006) migrations be applied in same transaction?

**Context**: golang-migrate applies migrations sequentially. If 1006 depends on 1005, transaction failure could leave inconsistent state.

**Options**:

A. **Separate transactions** - Each migration in own transaction (golang-migrate default)
B. **Single transaction** - Both migrations in one transaction (all-or-nothing)
C. **1005 then 1006 sequentially** - 1005 commits first, then 1006 starts (dependency enforcement)
D. **No dependencies** - 1006 does NOT depend on 1005 (independent tables)
E. **Write-in**: WTF is tenant_join_requests (1006)? Only pending_users (1005) needed?

**Answer**: E

---

### Q5.2: Migration Rollback Strategy

**Question**: Should DOWN migrations be implemented for pending_users and tenant_join_requests?

**Context**: QUIZME-v1 Q9.1 answered "No rollback, template design final after cipher-im validation". But DOWN migrations provide safety net.

**Options**:

A. **No DOWN migrations** - Align with QUIZME-v1 Q9.1 answer D (forward-only)
B. **DOWN migrations implemented** - Provide rollback capability for testing/development
C. **DOWN migrations for dev, not prod** - Development uses rollback, production forward-only
D. **Partial DOWN** - Drop tables but don't restore old schema (data loss acceptable for rollback)
E. **Write-in**:

**Answer**: B

---

## Section 6: Copilot Instructions applyTo Patterns

### Q6.1: applyTo Glob Patterns

**Question**: What glob patterns should be used for conditional instruction application?

**Context**: COPILOT-STRATEGY.md identified 0/28 instruction files have conditional applyTo patterns. Examples: testing instructions should apply to `**/*_test.go`, server-builder instructions to `internal/apps/*/server/**/*.go`.

**Options**:

A. **No conditional patterns** - All instructions apply to `**` (all files)
B. **Test-specific patterns** - Testing instructions use `**/*_test.go`, `**/*_bench_test.go`, etc.
C. **Layer-specific patterns** - Server instructions `internal/apps/*/server/**/*.go`, repository instructions `internal/apps/*/repository/**/*.go`
D. **Comprehensive patterns** - All 28 instruction files get specific glob patterns
E. **Write-in**:

**Answer**: A

---

### Q6.2: applyTo Pattern Granularity

**Question**: How specific should applyTo patterns be?

**Context**: Broad patterns (`internal/apps/**`) apply to many files (simple, risk of incorrect application). Narrow patterns (`internal/apps/template/server/application.go`) precise but verbose.

**Options**:

A. **Broad patterns** - `internal/apps/**/*.go` (simple, applies to all services)
B. **Medium patterns** - `internal/apps/*/server/**/*.go` (layer-specific)
C. **Narrow patterns** - `internal/apps/template/server/application.go` (file-specific)
D. **Mixed granularity** - Critical files get narrow patterns, general guidance gets broad patterns
E. **Write-in**: NO GLOB PATTERNS

**Answer**: E

---

## Section 7: Prompt Implementation Priority

### Q7.1: Daily-Use Prompts Priority

**Question**: Which prompts should be implemented first?

**Context**: COPILOT-STRATEGY.md lists 6 prompts (code-review, test-generate, fix-bug, refactor-extract, optimize-performance, generate-docs). Daily-use prompts (code-review, test-generate) vs specialized (optimize-performance).

**Options**:

A. **Daily-use first** - code-review, test-generate (Week 1), rest later
B. **By complexity** - Simple prompts (generate-docs) first, complex (fix-bug) later
C. **By impact** - High-value prompts (test-generate ≥95% coverage) first
D. **User preference** - User decides priority based on immediate needs
E. **Write-in**: I don't want any of those prompts

**Answer**: E

---

### Q7.2: Prompt YAML Frontmatter Requirements

**Question**: What metadata should be in prompt file YAML frontmatter?

**Context**: COPILOT-STRATEGY.md shows example with `name`, `description`, `applyTo`. Additional metadata: `author`, `version`, `tags`, `prerequisites`.

**Options**:

A. **Minimal** - `name`, `description`, `applyTo` only
B. **Standard** - `name`, `description`, `applyTo`, `tags`, `prerequisites`
C. **Comprehensive** - Add `author`, `version`, `last_updated`, `examples`
D. **Match instructions** - Same frontmatter structure as `.github/instructions/*.instructions.md` files
E. **Write-in**: I have no idea what is the context for this, or what this is for

**Answer**: E

---

## Section 8: Agent Handoff Patterns

### Q8.1: Security-to-Database Handoff Trigger

**Question**: When should expert.security.agent.md handoff to expert.database.agent.md?

**Context**: Security agent finds SQL injection vulnerability. Should it fix directly or handoff to database agent (query optimization expertise)?

**Options**:

A. **Always handoff** - Security identifies issue, database agent fixes (clean separation)
B. **Never handoff** - Security agent fixes SQL injection directly (faster)
C. **Conditional handoff** - Simple fixes (parameterized queries) done by security, complex (query rewrite) handed off
D. **User decides** - Security agent asks user "Handoff to database agent or fix now?"
E. **Write-in**: I have no idea what is the context for this, or what this is for

**Answer**: E

---

### Q8.2: Performance-to-Testing Handoff Trigger

**Question**: When should expert.performance.agent.md handoff to expert.testing.agent.md?

**Context**: Performance agent optimizes algorithm, needs to verify no regressions. Should it write tests or handoff?

**Options**:

A. **Always handoff** - Performance optimizes, testing writes regression tests
B. **Never handoff** - Performance writes benchmarks and regression tests directly
C. **Conditional** - Simple benchmarks by performance, comprehensive test suite by testing agent
D. **User decides** - Performance agent asks "Write tests or handoff?"
E. **Write-in**: I have no idea what is the context for this, or what this is for

**Answer**: E

---

## Section 9: E2E Test Execution Pattern

### Q9.1: E2E Test Docker Compose vs Direct

**Question**: Should E2E tests run services in Docker Compose or direct Go test process?

**Context**: Docker Compose = realistic (containers, networking, secrets) but slower (startup overhead). Direct Go = faster but less realistic.

**Options**:

A. **Docker Compose** - E2E tests always use `docker compose up` (most realistic)
B. **Direct Go** - E2E tests start servers in TestMain (faster, simpler)
C. **Hybrid** - Local dev uses direct Go, CI uses Docker Compose (balance speed and realism)
D. **Configurable** - Flag: `go test -e2e-mode=docker` or `go test -e2e-mode=direct`
E. **Write-in**:

**Answer**: A; e2e must be realistic customer experience

---

### Q9.2: E2E Test Database

**Question**: Should E2E tests use test-containers for PostgreSQL or mock database?

**Context**: Test-containers = real PostgreSQL (realistic migrations, constraints) but requires Docker. Mock database = fast but doesn't test real SQL.

**Options**:

A. **Always test-containers** - E2E tests use real PostgreSQL via test-containers
B. **Always mock** - E2E tests use mock database (fast, no Docker dependency)
C. **SQLite in-memory** - E2E tests use SQLite in-memory (real SQL, no Docker)
D. **Hybrid** - Local dev uses SQLite, CI uses PostgreSQL test-containers
E. **Write-in**: E2E uses docker compose which starts PostgreSQL container

**Answer**: E

---

### Q9.3: E2E Test Package Organization

**Question**: Should E2E tests be separate package or inline with unit tests?

**Context**: Separate package = clearer separation (e2e/, test/e2e/) but more directory clutter. Inline = same package but requires build tags to skip.

**Options**:

A. **Separate e2e/ directory** - `test/e2e/jose_ja_e2e_test.go`
B. **Inline with build tags** - `internal/apps/jose/ja/*_e2e_test.go` with `//go:build e2e`
C. **Inline without tags** - `*_e2e_test.go` runs with `go test ./...` (always executed)
D. **Per-package e2e/ subdirectory** - `internal/apps/jose/ja/e2e/registration_test.go`
E. **Write-in**: Per product-service e2e/ subdirectory

**Answer**: E

---

## Section 10: Documentation Standards

### Q10.1: ARCHITECTURE.md Detail Level

**Question**: How much implementation detail should ARCHITECTURE.md contain?

**Context**: High-level = concise, easy to review, no code examples. Detailed = comprehensive, includes code examples, harder to maintain.

**Options**:

A. **High-level only** - Patterns, principles, NO code examples (<1000 lines)
B. **Medium detail** - Patterns + minimal code examples (1000-2000 lines)
C. **Comprehensive** - Patterns + extensive code examples + troubleshooting (2000+ lines)
D. **Hybrid** - ARCHITECTURE.md high-level, separate docs/arch/*.md for detailed guides
E. **Write-in**:

**Answer**: A

---

### Q10.2: ARCHITECTURE.md Update Triggers

**Question**: What changes require updating ARCHITECTURE.md?

**Context**: Too broad = constant updates (every pattern tweak). Too narrow = doc becomes stale.

**Options**:

A. **Any pattern change** - Product added, service template modified, testing strategy updated
B. **Major changes only** - New product/service, core pattern overhaul, quality gate thresholds
C. **Breaking changes only** - Incompatible API changes, migration required changes
D. **Discretionary** - Developer judgment on significance
E. **Write-in**: When I decide

**Answer**: E

---

### Q10.3: ARCHITECTURE.md Versioning

**Question**: Should ARCHITECTURE.md be versioned?

**Context**: Versioning (v1.0.0, v1.1.0) tracks major changes. No versioning = simpler but harder to reference specific state.

**Options**:

A. **Semantic versioning** - v1.0.0 (major.minor.patch), increment on changes
B. **Date-based versioning** - v2025-01-18, increment daily/weekly
C. **Git-based versioning** - Git commit hash is version (no explicit version in file)
D. **No versioning** - Always latest, git log provides history
E. **Write-in**:

**Answer**: D

---

### Q10.4: Code Examples in Documentation

**Question**: Should instruction files (`.github/instructions/*.instructions.md`) include inline code examples?

**Context**: COPILOT-STRATEGY.md identified 10-15 instruction files missing code examples. Examples improve clarity but increase file size and maintenance burden.

**Options**:

A. **No code examples** - Descriptions only, keep files concise
B. **Minimal examples** - 1-2 code snippets per instruction file
C. **Comprehensive examples** - Multiple examples per pattern (good/bad, before/after)
D. **External examples** - Link to example files in testdata/ or docs/examples/
E. **Write-in**:

**Answer**: B

---

## Completion Status

**How many questions are answered?**: _____ / 40

**Percentage complete**: _____ %

**Estimated time to complete remaining**: _____ hours

**Follow-up needed?**: YES / NO

---

**End of CLARIFY-QUIZME-v2**
