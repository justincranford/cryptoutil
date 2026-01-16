# JOSE-JA Round 2 QUIZME - Implementation Details

**Last Updated**: 2026-01-16
**Purpose**: Deep implementation questions discovered after refining plan with Round 1 answers
**Format**: A-E options with **YOUR ANSWER: __** field for each question

## Instructions

**This document contains IMPLEMENTATION-LEVEL questions requiring user decisions before proceeding.**

After answering, these will inform detailed task breakdowns in JOSE-JA-REFACTORING-TASKS-V2.md.

**ANSWER FORMAT**: For each question, write your choice (A, B, C, D, or E) on the **YOUR ANSWER:__ ** line.
- If choosing E (write-in), also provide your custom answer below the question.

---

## Section 1: Elastic JWK Material Management

### Q1: Material JWK Lifecycle - When to Delete Retired Material?

**Context**: Elastic JWK contains active material + retired material JWKs. Retired materials needed for decrypt/verify historical data.

**Question**: When should retired material JWKs be deleted from database?

**A)** NEVER delete - keep all historical material JWKs forever
- **Pros**: Can always decrypt/verify historical data
- **Cons**: Database growth over time, potential compliance issues

**B)** CONFIGURABLE retention period (e.g., 90 days, 1 year, 5 years)
- **Pros**: Balance between history and storage, compliance-friendly
- **Cons**: Configuration complexity, may lose decrypt/verify ability

**C)** DELETE on elastic JWK deletion (cascade delete all materials)
- **Pros**: Clean deletion, simple logic
- **Cons**: Lose all historical materials when elastic JWK deleted

**D)** SOFT DELETE (mark as deleted, actual delete after retention period)
- **Pros**: Gradual cleanup, recovery window
- **Cons**: Requires cleanup job, more complex queries

**E)** Write-in (describe approach):

**YOUR ANSWER: __** A, max 1000 material JWKs per elastic JWK. When limit reached, user or client must generate a new elastic JWK, and manually migrate to use it.

---

### Q2: Material JWK Rotation Trigger

**Context**: Active material JWK should rotate periodically for security.

**Question**: What should trigger material JWK rotation?

**A)** TIME-BASED - Automatic rotation every N days (e.g., 90 days)
- **Pros**: Predictable, automated, compliance-friendly
- **Cons**: May rotate unnecessarily, cron job complexity

**B)** USAGE-BASED - Rotate after N operations (e.g., 10,000 signs)
- **Pros**: Efficient (only when needed)
- **Cons**: High-traffic JWKs rotate frequently, low-traffic never rotate

**C)** MANUAL ONLY - Admin explicitly triggers rotation via API
- **Pros**: Full control, no surprise rotations
- **Cons**: Manual process, may forget to rotate

**D)** HYBRID - Time-based default + manual override
- **Pros**: Automated with control, best of both
- **Cons**: More configuration, complexity

**E)** Write-in (describe approach):

**YOUR ANSWER: __** D

---

### Q3: Active Material Selection on Rotation

**Context**: When rotating, new material JWK becomes active. Old material JWK becomes retired.

**Question**: Should rotation be INSTANT or GRADUAL?

**A)** INSTANT - New material active immediately, old material retired
- **Pros**: Simple, clean cutover
- **Cons**: In-flight operations may fail if using old material

**B)** GRADUAL - Overlap period (both old and new active for N minutes)
- **Pros**: Smooth transition, no operation failures
- **Cons**: Two active materials simultaneously (which to use for sign/encrypt?)

**C)** BLUE-GREEN - Generate new material, test, then activate
- **Pros**: Validate before activation, safe rollback
- **Cons**: Manual step, delayed activation

**D)** CANARY - Route small percentage to new material, gradually increase
- **Pros**: Controlled rollout, early issue detection
- **Cons**: Complex routing logic, partial consistency

**E)** Write-in (describe approach):

**YOUR ANSWER: __** Old material JWK is never retired. For asymmetric signing, the public material JWK remains available for verification of historical signatures. For asymmetric encryption, the old private material JWK can still be used to decrypt data encrypted with it. For symmetric encryption, the old material JWK can still be used to decrypt, but not for encryption. For symmetric signing, the old material JWK can still be used to verify historical signatures, but not for new signing operations.

---

## Section 2: Audit Logging Configuration

### Q4: Audit Log Retention per Tenant

**Context**: High-frequency operations (sign/verify) generate large audit logs.

**Question**: How should per-tenant audit log retention be configured?

**A)** GLOBAL retention (all tenants same, e.g., 90 days)
- **Pros**: Simple, consistent, easy to enforce
- **Cons**: One-size-fits-all, may not fit all compliance needs

**B)** PER-TENANT retention (each tenant configures own, e.g., 30-365 days)
- **Pros**: Flexible, tenant-specific compliance
- **Cons**: Configuration complexity, tenant management overhead

**C)** PER-OPERATION retention (e.g., generate=5 years, sign=90 days)
- **Pros**: Granular control, optimize storage
- **Cons**: Complex configuration, hard to reason about

**D)** TIERED retention (critical ops=5 years, routine ops=90 days)
- **Pros**: Balance between granularity and simplicity
- **Cons**: Requires operation classification

**E)** Write-in (describe approach):

**YOUR ANSWER: __** D

---

### Q5: Audit Log Sampling for High-Volume Operations

**Context**: sign/verify operations may happen 1000s of times per second.

**Question**: Should high-volume operations be sampled (not all logged)?

**A)** NO sampling - Log ALL operations (100%)
- **Pros**: Complete audit trail, compliance-friendly
- **Cons**: Massive storage for high-traffic tenants

**B)** YES sampling - Configurable sample rate (e.g., 10% of sign/verify)
- **Pros**: Reduces storage, maintains statistical visibility
- **Cons**: Incomplete audit trail, may miss incidents

**C)** ADAPTIVE sampling - Sample more when idle, less when busy
- **Pros**: Balances completeness with performance
- **Cons**: Complex algorithm, unpredictable behavior

**D)** TIERED sampling - Different rates per operation type
- **Pros**: Optimize per operation characteristics
- **Cons**: Configuration complexity

**E)** Write-in (describe approach):

**YOUR ANSWER: __** B, 1% default per elastic JWK, can be overridden

---

### Q6: Audit Log Metadata Verbosity

**Context**: Audit logs can include request headers, IP, user-agent, etc.

**Question**: How verbose should audit log metadata be?

**A)** MINIMAL - operation, kid, user_id, session_id, timestamp ONLY
- **Pros**: Low storage, fast writes, privacy-friendly
- **Cons**: Limited forensics capability

**B)** STANDARD - Minimal + IP address, user-agent
- **Pros**: Good balance, useful for investigations
- **Cons**: PII concerns (IP address)

**C)** VERBOSE - Standard + request headers, payload hashes
- **Pros**: Complete forensics capability
- **Cons**: High storage, potential secrets in headers

**D)** CONFIGURABLE - Tenant chooses verbosity level
- **Pros**: Flexible, tenant-specific needs
- **Cons**: Configuration complexity, inconsistent logs

**E)** Write-in (describe approach):

**YOUR ANSWER: __** A

---

## Section 3: Per-JWK JWKS Endpoint Design

### Q7: JWKS Endpoint Caching Strategy

**Context**: JWKS endpoint (`/service/api/v1/jose/{kid}/.well-known/jwks.json`) may be called frequently.

**Question**: Should JWKS responses be cached?

**A)** NO caching - Always query database for freshness
- **Pros**: Always current, no stale data
- **Cons**: High DB load for popular JWKs

**B)** YES caching - Cache JWKS for N minutes (e.g., 5 min TTL)
- **Pros**: Reduces DB load, fast responses
- **Cons**: Stale data during rotation

**C)** CONDITIONAL caching - Cache until next rotation event
- **Pros**: Fresh during stability, efficient
- **Cons**: Requires cache invalidation on rotation

**D)** ETAG-based caching - Client caches with ETag, 304 Not Modified
- **Pros**: Bandwidth efficient, standard HTTP caching
- **Cons**: Still requires DB query to check ETag

**E)** Write-in (describe approach):

**YOUR ANSWER: __** B, default 5 min TTL

---

### Q8: JWKS Endpoint for Symmetric/Secret JWKs

**Context**: Symmetric JWKs (oct type) have no public key. Should JWKS be empty or error?

**Question**: What should `/service/api/v1/jose/{symmetric_kid}/.well-known/jwks.json` return?

**A)** EMPTY JWKS - `{"keys": []}` (no public keys for symmetric)
- **Pros**: Standard JWKS format, clear semantics
- **Cons**: Might confuse clients expecting keys

**B)** ERROR 404 - JWKS endpoint does not exist for symmetric keys
- **Pros**: Clear error signal, prevents misuse
- **Cons**: Inconsistent (some JWKs have endpoint, some don't)

**C)** ERROR 403 - Forbidden to access symmetric JWK public keys
- **Pros**: Security-focused message
- **Cons**: Implies public keys exist but are restricted

**D)** METADATA ONLY - Return JWK without key material (kid, alg, use only)
- **Pros**: Useful for key discovery
- **Cons**: Non-standard JWKS format

**E)** Write-in (describe approach):

**YOUR ANSWER: __** B; is there a more descriptive status code, or should we stick with 404? If HTTP 404, is there a way to communicate extra information in the response body to indicate that the endpoint exists but deliverately empty? And security concern about that, such as enumeration, or does UUIDv7 in path mitigate that? Any other security concerns or considerations?

---

### Q9: JWKS Endpoint Cross-Tenant Access

**Context**: Tenant A has public RSA JWK. Should Tenant B be able to access Tenant A's JWKS endpoint?

**Question**: Should JWKS endpoints be cross-tenant accessible?

**A)** YES - Public JWKs accessible to all tenants (public means public)
- **Pros**: Enables cross-tenant verification (e.g., federated identity)
- **Cons**: Tenant isolation weakened, information disclosure

**B)** NO - JWKS endpoints only accessible within same tenant
- **Pros**: Strict tenant isolation, privacy
- **Cons**: Can't verify JWTs from other tenants

**C)** CONFIGURABLE - Tenant chooses "public" or "private" JWKs
- **Pros**: Flexible, tenant control
- **Cons**: Configuration complexity, per-JWK setting

**D)** WHITELIST - Tenant A explicitly shares JWKs with Tenant B
- **Pros**: Explicit sharing, controlled access
- **Cons**: Management overhead, ACL complexity

**E)** Write-in (describe approach):

**YOUR ANSWER: __** B is default, but can be overridden per elastic JWK if needed. Options include making the elastic JWK public to all, sharing with specific tenant(s), or keeping it private within the tenant.

---

## Section 4: Tenant Registration & Management

### Q10: Tenant Registration UI

**Context**: Service-template requires tenant registration flow (create or join).

**Question**: Should jose-ja provide browser UI for tenant management?

**A)** YES - Full browser UI at `/browser/admin/tenants/**`
- **Pros**: User-friendly, self-service, admin tooling
- **Cons**: Development time, UI maintenance

**B)** NO - API-only, expect external UI or CLI
- **Pros**: Focus on API, less maintenance
- **Cons**: Users must build own UI or use CLI

**C)** MINIMAL UI - Basic tenant list + create, no advanced features
- **Pros**: Quick to build, covers common cases
- **Cons**: Limited functionality, may need expansion

**D)** DEFERRED - Phase 2 feature, API-only for Phase 1
- **Pros**: Faster Phase 1 completion
- **Cons**: Poor UX until Phase 2

**E)** Write-in (describe approach):

**YOUR ANSWER: __** A; it is MANDATORY for service-template to provide browser UI and service API for registering users || clients. Tenants can only be created via registering a new user || client, and specifying the option to create a new tenant or join an existing tenant. Tenant management user UI and service APIs must support viewing tenant, managing tenant details, managing tenant members, or deleting the tenant. Tenants MUST never be created directly without going through the registration flow.

---

### Q11: Tenant Admin Authorization Model

**Context**: Tenant creator becomes admin. Subsequent join requests need admin approval.

**Question**: How should admin authorization be managed?

**A)** SINGLE ADMIN - Only tenant creator is admin, can't delegate
- **Pros**: Simple, clear authority
- **Cons**: Single point of failure, no delegation

**B)** MULTIPLE ADMINS - Creator can promote other users to admin
- **Pros**: Delegation, redundancy
- **Cons**: Admin management complexity, permission conflicts

**C)** ROLE-BASED - Admin, moderator, member roles with different permissions
- **Pros**: Granular control, flexible permissions
- **Cons**: Complex RBAC implementation

**D)** INHERIT SERVICE-TEMPLATE - Use whatever service-template provides
- **Pros**: Consistent with other services, no duplication
- **Cons**: Limited to service-template capabilities

**E)** Write-in (describe approach): C and D; this is what session-template is supposed to support already. If not, then fix session-template to support RBAC for tenant members, so it can be reused by jose-ja and cipher-im.

**YOUR ANSWER: __**

---

### Q12: Cross-Realm JWK Visibility

**Context**: Tenant can have multiple realms. Should JWKs be visible across realms within same tenant?

**Question**: Are JWKs scoped to tenant only, or tenant + realm?

**A)** TENANT-SCOPED - All realms in tenant share same JWKs
- **Pros**: Simple, easy key sharing across realms
- **Cons**: No realm isolation, can't have realm-specific keys

**B)** REALM-SCOPED - JWKs isolated per (tenant, realm) pair
- **Pros**: Strict isolation, realm-specific keys
- **Cons**: Can't share keys across realms in same tenant

**C)** CONFIGURABLE - JWK has "scope" field (tenant or realm)
- **Pros**: Flexible, per-JWK decision
- **Cons**: Configuration complexity, scope conflicts

**D)** REALM-SCOPED with sharing - Realms can explicitly share JWKs
- **Pros**: Isolation with controlled sharing
- **Cons**: Sharing mechanism complexity

**E)** Write-in (describe approach):

**YOUR ANSWER: __** A; realms are for authentication and authorization purposes only, so tenant-scoped JWKs MUST be available with the tenant

---

## Section 5: Session Management Integration

### Q13: Session Timeout Differentiation

**Context**: SessionManager tracks both browser and service sessions.

**Question**: Should browser and service sessions have different timeouts?

**A)** YES - Browser shorter (e.g., 1 hour), service longer (e.g., 24 hours)
- **Pros**: UX for browsers, stability for services
- **Cons**: Different timeout logic, configuration complexity

**B)** NO - Same timeout for both (e.g., 8 hours)
- **Pros**: Consistent, simple configuration
- **Cons**: May not fit both use cases well

**C)** CONFIGURABLE - Per-tenant timeout settings
- **Pros**: Tenant-specific needs
- **Cons**: Configuration proliferation

**D)** ACTIVITY-BASED - Extend timeout on activity (sliding window)
- **Pros**: UX-friendly, automatic extension
- **Cons**: Complex expiration logic

**E)** Write-in (describe approach): A; this is how service-template is supposed to work, and offer that reusable functionality for all 9 product-services, including jose-ja and cipher-im.

**YOUR ANSWER: __**

---

### Q14: Session Invalidation on Key Rotation

**Context**: When material JWK rotates, should sessions using that JWK be invalidated?

**Question**: Should key rotation invalidate sessions?

**A)** YES - Force re-authentication after rotation for security
- **Pros**: Clean break, security-focused
- **Cons**: UX disruption, mass logout

**B)** NO - Sessions continue, use new material automatically
- **Pros**: Smooth UX, no disruption
- **Cons**: Old material still in use by active sessions

**C)** GRACE PERIOD - Sessions valid for N minutes after rotation
- **Pros**: Balance between security and UX
- **Cons**: Window where old material still used

**D)** OPTIONAL - Tenant chooses strict or lenient rotation
- **Pros**: Flexibility
- **Cons**: Configuration complexity, security risk if chosen poorly

**E)** Write-in (describe approach):

**YOUR ANSWER: __** B, session management for product-services is completely independent of the elastic JWKs managed inside the jose-ja service

---

## Section 6: Rate Limiting Strategy

### Q15: Rate Limiting Scope

**Context**: Rate limiting prevents abuse of JOSE operations.

**Question**: What should rate limiting be scoped to?

**A)** PER-TENANT - All users in tenant share rate limit
- **Pros**: Prevents tenant-level abuse, simple
- **Cons**: One heavy user affects all in tenant

**B)** PER-SESSION - Each session has independent rate limit
- **Pros**: Fairness, user isolation
- **Cons**: Can't limit tenant-wide abuse (many sessions)

**C)** PER-USER - Each user has rate limit across all sessions
- **Pros**: Consistent per-user experience
- **Cons**: User tracking complexity, cross-session coordination

**D)** TIERED - Different limits per operation type (generate=10/min, sign=1000/min)
- **Pros**: Optimized per operation cost
- **Cons**: Configuration complexity

**E)** Write-in (describe approach): Rate limiting MUST be per-browser or per-service at the Fiber level. This is how I originally implemented in sm-kms, and extracted for reuse in service-template. If it is not working that way, then service-template needs to be fixed for work like originally implemented in sm-kms.

**YOUR ANSWER: __**

---

### Q16: Rate Limit Enforcement Action

**Context**: When rate limit exceeded, what should happen?

**Question**: How should rate limit violations be handled?

**A)** REJECT - HTTP 429 Too Many Requests, client must retry later
- **Pros**: Standard HTTP semantics, clear signal
- **Cons**: Client retry logic required

**B)** QUEUE - Queue request, process when rate limit resets
- **Pros**: No dropped requests, smooth experience
- **Cons**: Queue management complexity, memory usage

**C)** THROTTLE - Slow down requests (delay responses)
- **Pros**: Gradual degradation, no errors
- **Cons**: Slow UX, resource holding

**D)** PRIORITIZE - Accept but deprioritize rate-limited requests
- **Pros**: Best effort, no hard failures
- **Cons**: Complex priority queue

**E)** Write-in (describe approach):

**YOUR ANSWER: __** A; that is how sm-kms was implemented, and the same logic should be extracted for reuse in service-template. If it is not working that way, then service-template needs to be fixed for work like originally implemented in sm-kms..

---

## Section 7: Barrier Key Rotation

### Q17: Barrier Key Rotation Strategy

**Context**: Barrier service encrypts private JWKs. Rotating barrier key requires re-encrypting ALL JWKs.

**Question**: How should barrier key rotation be handled?

**A)** BULK RE-ENCRYPT - Rotate barrier key, immediately re-encrypt all JWKs
- **Pros**: Clean cutover, single barrier key
- **Cons**: Downtime during re-encryption, database-heavy

**B)** LAZY RE-ENCRYPT - Rotate barrier key, re-encrypt JWKs on next use
- **Pros**: No downtime, gradual migration
- **Cons**: Old and new barrier keys coexist, long migration period

**C)** BACKGROUND JOB - Rotate barrier key, background job re-encrypts gradually
- **Pros**: No downtime, controlled migration
- **Cons**: Job complexity, partial consistency during migration

**D)** DUAL-KEY PERIOD - Both old and new barrier keys valid for N days
- **Pros**: Smooth transition, retry safety
- **Cons**: Two keys active, key management complexity

**E)** Write-in (describe approach): WRONG context. Barrier keys encrypt new content using latest material AES JWK, and support historical decrypt. There is no need to re-encrypt existing JWKs when rotating barrier keys. After rotating the barrier key, only new material JWKs will be encrypted with the new barrier key, while existing JWKs remain encrypted with their original barrier key.

**YOUR ANSWER: __**

---

### Q18: Barrier Key Version Tracking

**Context**: JWKs encrypted with different barrier key versions need version tracking.

**Question**: How should barrier key version be tracked per JWK?

**A)** COLUMN - Add `barrier_key_version` column to jwks table
- **Pros**: Explicit tracking, easy queries
- **Cons**: Schema change, migration required

**B)** EMBEDDED - Embed version in encrypted JWE (JWE header)
- **Pros**: No schema change, self-describing
- **Cons**: Requires JWE parsing to determine version

**C)** MAGIC PREFIX - Prepend version to ciphertext (e.g., "v2:")
- **Pros**: Simple, fast version check
- **Cons**: Non-standard, parsing logic

**D)** INHERIT BARRIER SERVICE - Barrier service tracks versions internally
- **Pros**: No jose-ja changes needed
- **Cons**: Tight coupling to barrier service internals

**E)** Write-in (describe approach): Barrier service encrypted content in JWE format, which contains a copy of the barrier key's in the JWE kid header. That JWE kid can be used to look up the corresponding barrier key. This should already be implemented in the current barrier service in service-template, which was extracted from the original sm-ks for reuse by all product-services built on top of service-template.

**YOUR ANSWER: __**

---

## Section 8: OpenAPI & API Versioning

### Q19: OpenAPI Spec Organization

**Context**: jose-ja has many endpoints (JWK, JWS, JWE, JWT, JWKS, tenant management).

**Question**: How should OpenAPI specs be organized?

**A)** SINGLE FILE - One openapi.yaml with all endpoints
- **Pros**: Simple, single source of truth
- **Cons**: Large file, hard to navigate

**B)** SPLIT BY FEATURE - Separate files (jwk.yaml, jws.yaml, jwe.yaml, etc.) merged
- **Pros**: Modular, easier reviews
- **Cons**: Merge complexity, reference resolution

**C)** SPLIT BY PATH PREFIX - browser.yaml + service.yaml
- **Pros**: Matches routing structure
- **Cons**: Feature split across files

**D)** COMPONENTS + PATHS - Shared components.yaml + paths per feature
- **Pros**: Reusability, modular
- **Cons**: Multiple file management

**E)** Write-in (describe approach):

**YOUR ANSWER: __** D; look at the openapi docs for sm-kms, that is how all product-service OpenAPI specs should be organized; components and paths are in separate files, and paths doc references components doc.

---

### Q20: API Versioning Strategy

**Context**: API may evolve (e.g., elastic JWK changes, new operations).

**Question**: How should API versioning be handled?

**A)** URL VERSIONING - `/browser/api/v2/jose/**` when breaking changes
- **Pros**: Explicit, clear cutover
- **Cons**: Multiple versions to maintain

**B)** HEADER VERSIONING - `Accept: application/vnd.jose.v2+json`
- **Pros**: Same URLs, version in header
- **Cons**: Less discoverable, client must know to send header

**C)** NO VERSIONING - Maintain backward compatibility always
- **Pros**: No version proliferation
- **Cons**: Technical debt, constrained evolution

**D)** DEFERRED - No versioning until first breaking change needed
- **Pros**: YAGNI, simpler initially
- **Cons**: May regret when breaking change needed

**E)** Write-in (describe approach):

**YOUR ANSWER: __** A

---

## Section 9: Testing & Validation

### Q21: Multi-Tenant Test Isolation

**Context**: Parallel tests with multiple tenants may conflict.

**Question**: How should multi-tenant tests be isolated?

**A)** UNIQUE TENANT IDs - UUIDv7 for each test tenant
- **Pros**: No collisions, parallel-safe
- **Cons**: Cleanup complexity, orphaned tenants

**B)** TEST DATABASE per test - Each test gets own database
- **Pros**: Complete isolation
- **Cons**: Slow, resource-heavy

**C)** TRANSACTION ROLLBACK - Use transactions, rollback after test
- **Pros**: Fast, clean state
- **Cons**: Doesn't test transaction behavior, may mask bugs

**D)** CLEANUP HOOKS - Create tenants, delete in defer/cleanup
- **Pros**: Explicit cleanup, test real create/delete
- **Cons**: If test crashes, orphaned tenants

**E)** Write-in (describe approach):

**YOUR ANSWER: __** A; it is extremely important to use UUIDv7 for all parallel testing collision avoidance. Maximum concurrency in all tests is mandatory for fastest test execution, and for revealing concurrency issues in main product code.

---

### Q22: E2E Test Multi-Instance Coordination

**Context**: E2E tests with multiple jose-ja instances sharing database.

**Question**: How should E2E tests coordinate multiple instances?

**A)** DOCKER COMPOSE - Declarative multi-instance deployment
- **Pros**: Production-like, simple config
- **Cons**: Slower startup, port management

**B)** IN-PROCESS - Start multiple servers in same test process (different ports)
- **Pros**: Fast, easy debugging
- **Cons**: Not production-like, shared process memory

**C)** GOROUTINES - Start instances in goroutines
- **Pros**: Concurrency testing, fast
- **Cons**: Race conditions, shared state bugs

**D)** SEQUENTIAL - Start instance 1, test, stop, start instance 2, test
- **Pros**: Simple, no coordination needed
- **Cons**: Doesn't test true multi-instance scenarios

**E)** Write-in (describe approach):

**YOUR ANSWER: __** A; e2e tests should use Docker Compose in TestMain to start all jose-ca services and dependencies once, and then reuse them for all tests in the e2e package. That is the same pattern used in cipher-im.

---

## Section 10: Performance Optimization

### Q23: Database Connection Pool Tuning

**Context**: High-concurrency jose-ja may exhaust database connections.

**Question**: How should connection pool be configured?

**A)** INHERIT SERVICE-TEMPLATE - Use service-template defaults (no override)
- **Pros**: Consistent, tested configuration
- **Cons**: May not fit jose-ja usage patterns

**B)** JOSE-SPECIFIC TUNING - Override based on jose-ja benchmarks
- **Pros**: Optimized for jose-ja workload
- **Cons**: Duplication, per-service tuning

**C)** DYNAMIC TUNING - Adjust pool size based on load
- **Pros**: Adaptive, efficient
- **Cons**: Complex algorithm, unpredictable

**D)** TENANT-BASED POOLING - Separate pools per tenant
- **Pros**: Tenant isolation, fairness
- **Cons**: High resource usage, complex management

**E)** Write-in (describe approach):

**YOUR ANSWER: __** A; it worked for cipher-im which inherits from service-template, and it worked for sm-kms which was extracted to create service-template.

---

### Q24: Barrier Encryption Performance

**Context**: Encrypting/decrypting JWKs on every operation may be slow.

**Question**: Should there be a performance optimization for barrier encryption?

**A)** NO OPTIMIZATION - Always encrypt/decrypt (security first)
- **Pros**: Maximum security, simple
- **Cons**: Potential performance bottleneck

**B)** MEMORY CACHE - Cache decrypted JWKs in memory (TTL=5 min)
- **Pros**: Fast, reduces crypto overhead
- **Cons**: Plaintext in memory, invalidation complexity

**C)** PROCESS-LOCAL CACHE - Cache per jose-ja instance (no sharing)
- **Pros**: No cross-instance coordination
- **Cons**: Inconsistent across instances, memory usage

**D)** LAZY ENCRYPTION - Store plaintext, encrypt only for backup/export
- **Pros**: Fast operations
- **Cons**: **SECURITY VIOLATION** - defeats barrier purpose

**E)** Write-in (describe approach):

**YOUR ANSWER: __** A

---

## Summary

**Total Questions**: 24 across 10 sections
**Answer Format**: Write A, B, C, D, or E in **YOUR ANSWER: __** field
**Write-In Instructions**: If choosing E, provide detailed custom answer below question

**After Answering**:
1. Review all answers for consistency
2. Update JOSE-JA-REFACTORING-TASKS-V2.md with implementation details
3. Begin Phase 1 implementation

---

## Cross-References

- **Round 1 Quiz**: [JOSE-JA-QUIZME.md](JOSE-JA-QUIZME.md)
- **Refined Plan**: [JOSE-JA-REFACTORING-PLAN-V2.md](JOSE-JA-REFACTORING-PLAN-V2.md)
- **Service-Template**: [03-08.server-builder.instructions.md](../../.github/instructions/03-08.server-builder.instructions.md)
- **cipher-im Reference**: [internal/apps/cipher/im/](../../internal/apps/cipher/im/)
