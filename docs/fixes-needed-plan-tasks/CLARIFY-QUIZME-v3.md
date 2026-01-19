# CLARIFY-QUIZME-v3: Remaining Implementation Details

**Purpose**: Identify remaining unknowns NOT covered in QUIZME-v1 (strategic) or QUIZME-v2 (tactical).

**Context**: QUIZME-v1 answered strategic architecture (100% complete). QUIZME-v2 answered tactical implementation details (100% complete per commit da6212f8). This v3 focuses on remaining gaps discovered during plan review.

**Instructions**:

- Answer with A/B/C/D or write-in for E
- Write-in answers provide detailed rationale
- Mark unanswered questions with blank for discussion
- Reference QUIZME-v1 and QUIZME-v2 answers where relevant

**CRITICAL FORMATTING RULES** (corrected from v1/v2):
- Write-in option: `E. **Write-in**:` (NO trailing underscores)
- Answer label: `**Answer**:` (NO trailing question mark ❓)

---

## Section 1: Session Endpoints vs Authentication Endpoints

### Q1.1: Session Endpoints Definition

**Question**: How is "Session Endpoints" different from "Authentication Endpoints"? Are both sets needed?

**Context**: User asked: "How is Session Endpoints different from authentication endpoints? Are both sets needed?" Currently unclear if these are:
- Same endpoints with different names (documentation inconsistency)
- Different endpoints serving different purposes (clear separation needed)
- Overlapping endpoints with partial redundancy (needs consolidation)

**Options**:

A. **Session endpoints = Authentication endpoints** - Same thing, different names in docs (consolidate terminology)
B. **Session endpoints handle session lifecycle** - Create/validate/refresh/revoke sessions (after authentication succeeds)
C. **Authentication endpoints handle credential verification** - Login/logout/register (before session creation)
D. **Both needed with clear separation** - Authentication verifies credentials → Session manages lifecycle
E. **Write-in**:

**Answer**:

---

### Q1.2: Session Endpoint Responsibilities

**Question**: If session endpoints and authentication endpoints are different, what are the specific responsibilities of each?

**Context**: Need clear delineation to avoid overlap and ensure complete coverage.

**Options**:

A. **Authentication endpoints**: `/auth/login`, `/auth/logout`, `/auth/register`, `/auth/verify`
B. **Session endpoints**: `/session/create`, `/session/validate`, `/session/refresh`, `/session/revoke`
C. **Combined approach**: Single set of endpoints handles both (e.g., `/auth/login` creates session internally)
D. **Authentication = browser/service registration flow**, **Session = template infrastructure session management**
E. **Write-in**:

**Answer**:

---

## Section 2: Service Federation Configuration

### Q2.1: Federation Service Discovery Updates

**Question**: How should service federation configuration handle dynamic service discovery in Docker Compose and Kubernetes?

**Context**: Plan mentions service discovery via config file → Docker Compose → Kubernetes DNS. Need clarification on when to use each method and how to handle updates.

**Options**:

A. **Static config file only** - Hard-code URLs in YAML, restart service on changes
B. **Docker Compose DNS** - Use service names (e.g., `identity-authz:8180`), no restarts needed
C. **Kubernetes DNS** - Use FQDN (e.g., `identity-authz.cryptoutil-ns.svc.cluster.local:8180`)
D. **Hybrid**: Config file for dev, Docker DNS for Docker Compose, K8s DNS for K8s
E. **Write-in**:

**Answer**:

---

### Q2.2: Federation Fallback Mode Activation

**Question**: When should federation fallback modes activate? What triggers circuit breaker state changes?

**Context**: Plan mentions circuit breaker + fallback modes but doesn't specify activation thresholds.

**Options**:

A. **Immediate failover** - First error activates fallback mode (fast but sensitive to transient errors)
B. **N consecutive failures** - 3-5 consecutive errors before fallback (balanced)
C. **Percentage-based** - 50% error rate over 1 minute activates fallback (statistical)
D. **Configurable threshold** - YAML config: `circuit_breaker_threshold: 5`
E. **Write-in**:

**Answer**:

---

## Section 3: Message Encryption Key Rotation

### Q3.1: Cipher-IM Message Key Rotation Frequency

**Question**: How frequently should message encryption keys be rotated in cipher-im?

**Context**: cipher-im demonstrates service template usage. Message encryption keys need rotation strategy.

**Options**:

A. **Never rotate** - Use same key for all messages (simple, not recommended for production)
B. **Rotate per message** - New key for each message (most secure, highest overhead)
C. **Rotate hourly** - New key every hour (balanced)
D. **Rotate daily** - New key every 24 hours (reduces overhead)
E. **Write-in**:

**Answer**:

---

### Q3.2: Cipher-IM Key Storage Pattern

**Question**: Should cipher-im message encryption keys be stored in Barrier service or separate table?

**Context**: Template provides Barrier service for encryption-at-rest. Need to determine if cipher-im should use Barrier or implement separate key storage.

**Options**:

A. **Barrier service** - Use template infrastructure (consistent with other services)
B. **Separate message_keys table** - Domain-specific storage (more control)
C. **Hybrid** - Barrier for encryption keys, separate table for metadata
D. **No separate storage** - Derive keys from master key + message ID (deterministic)
E. **Write-in**:

**Answer**:

---

## Section 4: Identity Product OAuth 2.1 Flows

### Q4.1: OAuth 2.1 Flow Priority

**Question**: Which OAuth 2.1 flows should identity-authz implement first?

**Context**: OAuth 2.1 supports multiple flows. Need prioritization for implementation.

**Options**:

A. **Authorization Code + PKCE only** - Modern, secure, browser + native apps
B. **Client Credentials only** - Service-to-service (simplest, most common in cryptoutil)
C. **Both Authorization Code + PKCE and Client Credentials** - Cover all use cases
D. **All flows** - Authorization Code, Client Credentials, Device Code, Refresh Token
E. **Write-in**:

**Answer**:

---

### Q4.2: Identity Product Token Storage

**Question**: Should identity product tokens be stored in database or issued as stateless JWTs?

**Context**: Trade-off between revocation capabilities (database) and performance (stateless).

**Options**:

A. **Database storage** - Tokens in table, allows instant revocation (slower validation)
B. **Stateless JWTs** - No storage, fast validation, delayed revocation (requires expiry)
C. **Hybrid** - Access tokens stateless, refresh tokens in database
D. **Configurable** - YAML config: `token_storage_mode: database` or `stateless`
E. **Write-in**:

**Answer**:

---

## Section 5: Multi-Tenant Data Isolation Verification

### Q5.1: Multi-Tenant Isolation Testing Strategy

**Question**: How should multi-tenant data isolation be verified in E2E tests?

**Context**: Critical security requirement - tenants must NOT access each other's data. Need comprehensive test strategy.

**Options**:

A. **Unit tests only** - Test repository WHERE clauses include tenant_id (insufficient for E2E)
B. **E2E tests with 2 tenants** - Register 2 users/tenants, verify no cross-tenant access
C. **E2E tests with N tenants** - Parameterized tests with multiple tenants (comprehensive)
D. **Audit log verification** - Check audit logs confirm NO cross-tenant queries
E. **Write-in**:

**Answer**:

---

### Q5.2: Tenant Isolation Test Data

**Question**: Should E2E tenant isolation tests use shared test data or generate unique data per test?

**Context**: Shared data = faster setup but risk of contamination. Unique data = slower but isolated.

**Options**:

A. **Shared test data** - TestMain creates 2 tenants, all tests share (fast, risk of contamination)
B. **Unique per test** - Each test registers new tenants (slow, fully isolated)
C. **Unique per test suite** - Each test file gets unique tenants (balanced)
D. **Hybrid** - Shared for read-only tests, unique for write tests
E. **Write-in**:

**Answer**:

---

## Section 6: E2E Test Data Cleanup

### Q6.1: E2E Test Database Cleanup Strategy

**Question**: How should E2E test databases be cleaned up after test runs?

**Context**: Docker Compose E2E tests create PostgreSQL containers with test data. Need cleanup strategy.

**Options**:

A. **No cleanup** - Each test run creates new database (fast, uses more disk space)
B. **Delete after test run** - `docker compose down -v` removes volumes (clean, but loses debug data)
C. **Keep on failure, delete on success** - Preserves data for debugging failures
D. **Configurable retention** - Keep last N test databases (e.g., 5), delete older
E. **Write-in**:

**Answer**:

---

### Q6.2: E2E Test Tenant Cleanup

**Question**: Should E2E tests clean up tenants after each test or rely on database cleanup?

**Context**: Tests register tenants/users. Need to determine cleanup level.

**Options**:

A. **No tenant cleanup** - Rely on database cleanup (simple)
B. **DELETE FROM users/tenants** - Explicit cleanup in test teardown (more control)
C. **TRUNCATE tables** - Fast cleanup, resets sequences
D. **Test-level isolation** - Each test gets fresh database container (slowest, most isolated)
E. **Write-in**:

**Answer**:

---

## Section 7: Performance Optimization Priorities

### Q7.1: Query Optimization Focus Areas

**Question**: Which database queries should be optimized first in jose-ja implementation?

**Context**: Limited time for optimization. Need prioritization based on expected usage patterns.

**Options**:

A. **JWK lookup by KID** - Most frequent query (sign/verify operations)
B. **Material key rotation queries** - Less frequent but complex
C. **Audit log queries** - High volume but async (low priority for critical path)
D. **All equally** - Optimize as discovered during implementation
E. **Write-in**:

**Answer**:

---

### Q7.2: Caching Strategy for JWKs

**Question**: Should jose-ja cache JWKs in memory or always query database?

**Context**: Trade-off between stale data risk (cache) and query latency (no cache).

**Options**:

A. **No caching** - Always query database (simplest, fresh data, higher latency)
B. **In-memory LRU cache** - Cache 100 most recent JWKs (faster, stale data risk)
C. **TTL-based cache** - Cache with 5-minute expiry (balanced)
D. **Configurable caching** - YAML config: `jwk_cache_enabled: true`, `jwk_cache_ttl: 5m`
E. **Write-in**:

**Answer**:

---

## Completion Status

**How many questions are answered**: _____ / 14

**Percentage complete**: _____ %

**Estimated time to complete remaining**: _____ hours

**Follow-up needed**: YES / NO

---

**End of CLARIFY-QUIZME-v3**
