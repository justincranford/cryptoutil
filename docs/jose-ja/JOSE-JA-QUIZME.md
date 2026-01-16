# JOSE-JA Refactoring QUIZME

**Last Updated**: 2026-01-15
**Purpose**: Critical unknowns, risks, gaps, and prioritization decisions requiring user input
**Format**: A-D options + E write-in for each question

## Instructions

**This document contains questions about UNKNOWNS discovered during deep analysis that require user clarification before proceeding with refactoring.**

After user answers, update JOSE-JA-REFACTORING-PLAN.md and JOSE-JA-REFACTORING-TASKS.md accordingly.

---

## CRITICAL ERROR IN ANALYSIS - CORRECTION REQUIRED

**ANALYSIS ERROR IDENTIFIED**:
- ❌ **WRONG**: Analysis presented "stateless vs database" as an architectural decision/option
- ✅ **CORRECT**: **ALL services MUST use service-template which MANDATES database persistence (SQLite/PostgreSQL)**

**SERVICE-TEMPLATE MANDATE** (from 02-02.service-template.instructions.md):
- ALL services MUST use service-template builder pattern
- Service-template PROVIDES: DB (GORM), migrations, sessions, barrier, telemetry, JWKGen
- cipher-im DEMONSTRATES: Uses builder, migrations, database persistence (MessageRepository, UserRepository)

**CORRECTED UNDERSTANDING**:
- jose-ja MUST migrate from in-memory KeyStore to database-backed JWKRepository
- jose-ja MUST use GORM for persistence (same as cipher-im)
- jose-ja MUST support SQLite + PostgreSQL (same as service-template)
- jose-ja MUST use barrier encryption for private keys at rest

---

## Section 1: KeyStore → Database Migration Strategy

### Q1: Private Key Storage Format

**Current**: In-memory `map[string]*StoredKey` with `PrivateJWK joseJwk.Key` (in-memory object)

**Question**: How should private keys be stored in database?

**A)** Store raw JWK JSON in TEXT column, encrypt entire JSON with barrier
- **Pros**: Simple, preserves all JWK fields, easy serialization
- **Cons**: Entire JWK encrypted (harder to query metadata), larger ciphertext

**B)** Store PEM-encoded private key in TEXT column, encrypt PEM with barrier
- **Pros**: Standard format, works with x509/crypto packages
- **Cons**: Loses JWK-specific fields (kid, use, alg), conversion overhead

**C)** Store JWK JSON in TEXT column, encrypt ONLY the private key material (d, p, q, etc)
- **Pros**: Metadata queryable, smaller ciphertext, granular encryption
- **Cons**: Complex serialization/deserialization, JWK structure manipulation

**D)** Store encrypted blob with custom format (metadata + encrypted key material)
- **Pros**: Optimized for query + encryption
- **Cons**: Custom format, maintenance burden

**E)** Write-in (describe alternative approach): Use barrier service to encrypt, just like in cipher-im, which is based on sm-kms service.

---

### Q2: Public Key Storage Format

**Current**: In-memory `PublicJWK joseJwk.Key` (in-memory object)

**Question**: How should public keys be stored in database?

**A)** Store JWK JSON in TEXT column (plaintext, no encryption)
- **Pros**: Simple, queryable, JWKS generation fast
- **Cons**: JSON parsing overhead on retrieval

**B)** Store PEM-encoded public key in TEXT column (plaintext)
- **Pros**: Standard format
- **Cons**: Conversion overhead, loses JWK metadata

**C)** Store as separate columns (kty, n, e for RSA; kty, crv, x, y for EC)
- **Pros**: Fully queryable, no JSON parsing
- **Cons**: Complex schema, algorithm-dependent columns

**D)** Store compressed/binary format for efficiency
- **Pros**: Smaller storage
- **Cons**: Compression overhead, binary handling

**E)** Write-in (describe alternative approach): JWE output from barrier service

---

### Q3: Key Metadata Duplication

**Current KeyStore.StoredKey fields**: KID, PrivateJWK, PublicJWK, KeyType, Algorithm, Use, CreatedAt

**Question**: Should key metadata (kty, alg, use) be stored in separate columns OR only in JWK JSON?

**A)** Separate columns + JWK JSON (duplication)
- **Pros**: Fast queries (SELECT * WHERE use='sig'), indexable
- **Cons**: Duplication (metadata in both columns AND JWK JSON), sync issues

**B)** Only in JWK JSON (no separate columns)
- **Pros**: No duplication, single source of truth
- **Cons**: Requires JSON extraction for queries (slower), harder to index

**C)** Separate columns as primary, derive JWK JSON on read
- **Pros**: Query performance, no JSON storage overhead
- **Cons**: JWK reconstruction logic, migration complexity

**D)** Use PostgreSQL JSONB column with GIN indexes
- **Pros**: Queryable JSON, indexable, no duplication
- **Cons**: PostgreSQL-specific (SQLite compatibility issue)

**E)** Write-in (describe alternative approach): Same format as in cipher-im; it is JWE encrypted blob, plus subset of metadata; i think just kid (uuidv7) and create timestamp, but you can look it up to confirm

---

## Section 2: Multi-Tenancy & Realms

### Q4: Tenant/Realm Architecture for JWKs

**Current**: jose-ja has NO tenant/realm architecture (single KeyStore)

**cipher-im pattern**: Uses `tenant_id` and `realm_id` foreign keys in messages table

**Question**: Should JWKs be isolated by tenant + realm?

**A)** YES - Add tenant_id + realm_id to jwks table (multi-tenancy)
- **Pros**: Consistent with cipher-im, supports multi-tenant deployments, key isolation
- **Cons**: Migration complexity, existing stateless API breaks, realm management required

**B)** NO - Keep global key namespace (single tenant, single realm)
- **Pros**: Simpler migration, backward compatible, no realm management
- **Cons**: Inconsistent with other services, can't support multi-tenant later

**C)** OPTIONAL - Add tenant_id + realm_id columns with nullable values, default to single tenant
- **Pros**: Incremental migration path, backward compatible
- **Cons**: Partial multi-tenancy, nullable foreign keys, mixed semantics

**D)** DEFERRED - Implement single-tenant now, add multi-tenancy in Phase 2
- **Pros**: Faster initial migration, validate pattern first
- **Cons**: Database schema migration later, potential breaking changes

**E)** Write-in (describe alternative approach):

A is mandatory

---

### Q5: Default Tenant/Realm for jose-ja

**Question**: If using multi-tenancy, what should default tenant/realm IDs be?

**A)** Use `cryptoutilMagic.DefaultTenantID` and `cryptoutilMagic.DefaultRealmID` (shared with other services)
- **Pros**: Consistency, shared magic constants
- **Cons**: Cross-service tenant/realm sharing (may not be desired)

**B)** Create `cryptoutilMagic.JoseDefaultTenantID` and `cryptoutilMagic.JoseDefaultRealmID` (jose-specific)
- **Pros**: Service isolation, clear ownership
- **Cons**: More magic constants, potential confusion

**C)** Use environment variable (JOSE_DEFAULT_TENANT_ID, JOSE_DEFAULT_REALM_ID)
- **Pros**: Runtime configurability
- **Cons**: Configuration complexity, not compile-time constant

**D)** Hardcode UUIDs in migration (0001_jose_keys.up.sql inserts default tenant/realm)
- **Pros**: Database-driven, visible in schema
- **Cons**: Migration dependency, not Go constant

**E)** Write-in (describe alternative approach): No default realm allowed in service-template, cipher-im, jose-ja, or any product-services!!! Service-template allows browser users or service clients to register, with option to generate a new tenant or ask to join an existing tenant. If new tenant, the user or client is the admin, and subsequent requests to join the tenant must be authorized by an admin. Allof this should already be clear from service-template and cipher-im, if not them service-template and cipher-im need fixing.

---

## Section 3: Audit Logging

### Q6: Audit Log Granularity

**Current**: No audit logging

**Question**: What operations should be logged in audit table?

**A)** ALL operations (generate, get, list, delete, sign, verify, encrypt, decrypt)
- **Pros**: Complete audit trail, compliance-ready
- **Cons**: High volume (especially for sign/verify/encrypt/decrypt), storage overhead

**B)** ONLY mutating operations (generate, delete)
- **Pros**: Lower volume, focuses on key lifecycle
- **Cons**: Missing usage patterns (who's using which keys for what)

**C)** CONFIGURABLE - Environment variable controls logging level
- **Pros**: Flexibility, performance tuning
- **Cons**: Configuration complexity, inconsistent audit trails across environments

**D)** TIERED - Generate/delete logged always, usage operations logged with sampling (10% of sign/verify/encrypt/decrypt)
- **Pros**: Balance between completeness and performance
- **Cons**: Sampling logic complexity, incomplete audit trail

**E)** Write-in (describe alternative approach):

A, with each operation individually configurable on a per-tenant basis, but all default to enabled. TBD, I may tweak it so some operations default their audit log to off, so you need to ensure table-driven tests cover each operation as individually configurable for auditlogging on or off per-tenant.

---

### Q7: Audit Log User Attribution

**Question**: How should audit logs identify the user/caller?

**A)** Extract from JWT claims (sub, email) if authentication enabled
- **Pros**: Real user identity
- **Cons**: Requires authentication, nullable if API key auth

**B)** Store API key ID or client certificate subject
- **Pros**: Works with current API key middleware
- **Cons**: Service account, not real user

**C)** Store IP address + User-Agent
- **Pros**: Always available
- **Cons**: Not reliable identity, privacy concerns

**D)** Leave user_id nullable, populate when available (best-effort)
- **Pros**: Flexible, no breaking changes
- **Cons**: Incomplete attribution

**E)** Write-in (describe alternative approach): Auditlogs must link to primary key of the user or client in service-template. The user||client can therefore be looked up for additional information, such as tenant and identifiers and attributes. Auditlogs must also link to primary key of session from session manager as provided by service-template.

---

## Section 4: Backward Compatibility

### Q8: Migration Strategy for Existing Deployments

**Current**: jose-ja stateless (in-memory KeyStore), keys lost on restart

**Question**: How should migration handle existing deployments?

**A)** BREAKING CHANGE - Require database from day 1, no backward compatibility
- **Pros**: Clean migration, no technical debt
- **Cons**: Breaks existing deployments, requires immediate database setup

**B)** DUAL MODE - Support both in-memory (deprecated) and database modes via flag
- **Pros**: Gradual migration, backward compatible
- **Cons**: Maintenance burden (two code paths), deprecation timeline unclear

**C)** AUTO-DETECT - Use database if DATABASE_URL present, fallback to in-memory
- **Pros**: Seamless migration, no flag needed
- **Cons**: Silent fallback behavior, hard to debug

**D)** STAGED ROLLOUT - Version N supports both modes, version N+1 database-only
- **Pros**: Clear deprecation path, migration window
- **Cons**: Multi-version support complexity

**E)** Write-in (describe alternative approach): A, no backwards compatibility. This is a non-released alpha project.

---

### Q9: Key Export/Import for Migration

**Question**: Should jose-ja provide key export/import tools for migration?

**A)** YES - Provide CLI commands (jose-ja export-keys, jose-ja import-keys)
- **Pros**: Easy migration, admin tooling
- **Cons**: Development time, testing burden

**B)** NO - Expect users to re-generate keys after migration
- **Pros**: Clean slate, no export logic needed
- **Cons**: Service disruption, key rotation required

**C)** ADMIN API - Add /admin/v1/keys/export and /admin/v1/keys/import endpoints
- **Pros**: No CLI needed, API-driven
- **Cons**: Security risk (private key exposure), requires authentication

**D)** DOCUMENTATION ONLY - Provide SQL queries for manual migration
- **Pros**: Low effort, flexibility
- **Cons**: Error-prone, manual process

**E)** Write-in (describe alternative approach): A, but `jose-ja client subcommand` format, and other command formats per service-template

---

## Section 5: Session Management

### Q10: Session Manager Integration

**service-template provides**: SessionManager (from builder.Build().SessionManager)

**cipher-im uses**: SessionManager for user sessions

**Question**: Does jose-ja need SessionManager integration?

**A)** YES - Use SessionManager for API client sessions (track API key usage, rate limiting)
- **Pros**: Consistent with other services, session-based rate limiting
- **Cons**: jose-ja currently stateless API (no sessions), paradigm shift

**B)** NO - jose-ja uses stateless API key auth only (no sessions)
- **Pros**: Simpler, maintains stateless API paradigm
- **Cons**: Inconsistent with service-template pattern, no session tracking

**C)** OPTIONAL - Provide SessionManager for OAuth flows (future), skip for current API key auth
- **Pros**: Future-proof for OAuth, doesn't change current API
- **Cons**: Dead code until OAuth implemented

**D)** DEFERRED - Skip SessionManager for initial migration, revisit when adding OAuth
- **Pros**: Faster migration, incremental adoption
- **Cons**: Incomplete service-template integration

**E)** Write-in (describe alternative approach): A, SessionManager reuse from service-template is mandatory

---

## Section 6: Route Organization

### Q11: Public vs Service API Paths

**service-template pattern**: `/browser/**` (sessions, CSRF) vs `/service/**` (tokens, no CSRF)

**jose-ja current**: `/jose/v1/**` (stateless API key auth, no CSRF)

**Question**: Should jose-ja adopt `/browser/**` and `/service/**` path split?

**A)** YES - Migrate to `/browser/api/v1/jose/**` and `/service/api/v1/jose/**`
- **Pros**: Consistent with service-template, supports future browser UI
- **Cons**: Breaking API change, URL migration for clients

**B)** NO - Keep `/jose/v1/**` paths (grandfathered exception)
- **Pros**: Backward compatible, no client changes
- **Cons**: Inconsistent with other services, harder to add CSRF later

**C)** HYBRID - Add new `/browser/**` and `/service/**` paths, keep `/jose/v1/**` as deprecated alias
- **Pros**: Backward compatible + new pattern, gradual migration
- **Cons**: Three path prefixes, deprecation timeline, routing complexity

**D)** NAMESPACE - Use `/service/api/v1/jose/**` only (no browser paths until OAuth)
- **Pros**: Partial compliance, clear API semantics
- **Cons**: Halfway solution, still inconsistent

**E)** Write-in (describe alternative approach): A, yes 100%, same path separation for browser users vs session clients.

---

### Q12: JWKS Discovery Endpoint Path

**Current**: `/.well-known/jwks.json` (OpenID Connect standard location)

**Question**: Should JWKS endpoint path change with new routing?

**A)** KEEP - `/.well-known/jwks.json` (standard location, no change)
- **Pros**: OIDC compliance, client compatibility
- **Cons**: Inconsistent with service-template routing

**B)** MOVE - `/service/api/v1/jose/.well-known/jwks.json`
- **Pros**: Consistent routing
- **Cons**: Breaks OIDC spec, non-standard

**C)** BOTH - Keep standard location + add namespaced alias
- **Pros**: OIDC compliance + consistency
- **Cons**: Duplication, routing complexity

**D)** SPECIAL CASE - `/.well-known/**` routes exempt from template routing pattern
- **Pros**: Standards compliance, clear exception
- **Cons**: Special case handling

**E)** Write-in (describe alternative approach): B, yes 100%, same path separation for browser users vs session clients. Also, per-JWK entry scoping. Each JWK is an elastic JWK (aka keyring, proxy JWK) containing time-ordered list of material JWKs. The JWKS URL per elastic JWK will only contain the public material JWKs of that elastic JWK. If JWK is a secret JWK, it will be empty.

---

## Section 7: Testing Strategy

### Q13: Test Data Isolation for Parallel Tests

**Question**: How should parallel tests avoid key ID collisions?

**A)** UUIDv7 for all key IDs in tests (like cipher-im pattern)
- **Pros**: Thread-safe, no collisions, consistent with other services
- **Cons**: Non-deterministic key IDs

**B)** Test-specific prefixes (test1_, test2_) + sequential IDs
- **Pros**: Deterministic, easier debugging
- **Cons**: Manual coordination, collision risk

**C)** Separate test databases per test (testcontainers)
- **Pros**: Complete isolation, no coordination needed
- **Cons**: Slower tests, resource overhead

**D)** Magic constants for test keys (TestKeyID1, TestKeyID2)
- **Pros**: Predictable, magic constant pattern
- **Cons**: Limited key count, collision risk in parallel

**E)** Write-in (describe alternative approach): A, UUIDv7 for all key IDs in tests (like cipher-im pattern)

---

### Q14: Mock vs Real Dependencies in Tests

**Question**: Should repository tests use mocks or real database?

**A)** Real database (testcontainers PostgreSQL + SQLite) - NO mocks
- **Pros**: Tests real GORM behavior, catches SQL bugs
- **Cons**: Slower tests, container dependency

**B)** Mock GORM interface for unit tests, real DB for integration tests
- **Pros**: Fast unit tests, comprehensive integration tests
- **Cons**: Mock maintenance, two test types

**C)** In-memory SQLite for all tests
- **Pros**: Fast, no container dependency
- **Cons**: May miss PostgreSQL-specific issues

**D)** Use builder pattern for tests (let builder handle DB setup)
- **Pros**: Consistent with service-template, less test boilerplate
- **Cons**: Heavyweight for unit tests

**E)** Write-in (describe alternative approach): A, Real database (SQLite) - NO mocks!!!

---

## Section 8: Performance & Scalability

### Q15: Key Caching Strategy

**Question**: Should frequently-accessed keys be cached in memory?

**A)** YES - LRU cache (size=1000) for key lookups, cache-aside pattern
- **Pros**: Faster lookups, reduced DB load
- **Cons**: Cache invalidation complexity, stale data risk

**B)** NO - Always query database (no caching)
- **Pros**: Simple, always fresh, consistent across instances
- **Cons**: Higher DB load, slower response times

**C)** READ-THROUGH cache - Cache on read, TTL=5 minutes
- **Pros**: Automatic population, bounded staleness
- **Cons**: TTL tuning, memory usage

**D)** WRITE-THROUGH cache - Update cache on write, invalidate on delete
- **Pros**: Always fresh, no staleness
- **Cons**: Complex invalidation, multi-instance coordination

**E)** Write-in (describe alternative approach): B, Always query database (no caching)

---

### Q16: Connection Pool Configuration

**Question**: What should database connection pool settings be?

**A)** Same as cipher-im (MaxOpenConns=25, MaxIdleConns=10)
- **Pros**: Proven configuration, consistency
- **Cons**: May not match jose-ja usage patterns

**B)** Higher (MaxOpenConns=50, MaxIdleConns=20) - expect high concurrency
- **Pros**: Better under load
- **Cons**: More resource usage

**C)** Lower (MaxOpenConns=10, MaxIdleConns=5) - expect low concurrency
- **Pros**: Lower resource usage
- **Cons**: Connection exhaustion risk

**D)** Configurable via environment variables
- **Pros**: Tunable per deployment
- **Cons**: More configuration surface

**E)** Write-in (describe alternative approach): Defaults are supposed to be in service-template so they are reusable by all product-services. I don't understand why you are saying cipher-im has defaults, the defaults must be inherited from service-template. If cipher-im is specifying defaults, that is a mistake that needs to be fixed, so that jose-ja and cipher-im can reuse the same database setups. The only thing unique I expect in services like cipher-im and jose-ja are domain-specific migrations and repositories, not PostgreSQL/SQLite/Gorm setup.

---

## Section 9: Prioritization & Phasing

### Q17: Phase 1 Scope - Minimum Viable Migration

**Question**: What is minimum viable Phase 1 for production deployment?

**A)** Database schema + repositories + JWK CRUD only (no sign/verify/encrypt/decrypt)
- **Pros**: Smallest increment, validate persistence first
- **Cons**: Incomplete API, can't deploy to production

**B)** Full JWK lifecycle + JWS/JWE/JWT operations (complete API)
- **Pros**: Production-ready, complete feature set
- **Cons**: Large Phase 1, longer time to validate

**C)** JWK CRUD + audit logging (no crypto operations)
- **Pros**: Validates persistence + compliance, incremental
- **Cons**: Crypto operations deferred, incomplete API

**D)** JWK CRUD + ONE crypto operation (sign/verify) as proof-of-concept
- **Pros**: End-to-end validation, manageable scope
- **Cons**: Partial API, other operations deferred

**E)** Write-in (describe alternative approach): B

---

### Q18: Builder Integration Timing

**Question**: When should jose-ja adopt ServerBuilder pattern?

**A)** BEFORE database migration (Phase 0: Builder, then Phase 1: Database)
- **Pros**: Infrastructure ready, then add domain logic
- **Cons**: Two-phase migration, more coordination

**B)** CONCURRENT with database migration (Phase 1: Builder + Database together)
- **Pros**: Single migration, faster completion
- **Cons**: Larger change set, harder to debug

**C)** AFTER database migration (Phase 1: Database, then Phase 2: Builder)
- **Pros**: Incremental, validate database first
- **Cons**: Temporary state with database but manual infrastructure

**D)** INCREMENTAL - Builder provides DB/telemetry/JWKGen in Phase 1, eliminate admin.go in Phase 2
- **Pros**: Gradual adoption, smaller changes
- **Cons**: Multiple phases, incomplete builder usage initially

**E)** Write-in (describe alternative approach): D

---

## Section 10: Risk Assessment

### Q19: Highest Risk Area

**Question**: What is the highest risk area for this refactoring?

**A)** Private key encryption/decryption with barrier (data loss if wrong)
- **Risk Level**: CRITICAL
- **Mitigation**: Comprehensive tests, backup/restore tools, gradual rollout

**B)** Multi-instance coordination (cache invalidation, lock contention)
- **Risk Level**: HIGH
- **Mitigation**: Optimistic locking, database-level locks, thorough load testing

**C)** API backward compatibility (breaking existing clients)
- **Risk Level**: HIGH
- **Mitigation**: Versioning, deprecation warnings, migration guide

**D)** Migration path from in-memory to database (state transfer)
- **Risk Level**: MEDIUM
- **Mitigation**: Export/import tools, clear documentation, rollback plan

**E)** Write-in (describe alternative risk): A

---

### Q20: Rollback Strategy

**Question**: If refactoring fails in production, how to rollback?

**A)** Database schema supports rollback - down migrations restore schema
- **Pros**: Clean rollback
- **Cons**: Requires down migrations tested

**B)** Blue-green deployment - keep old version running during migration
- **Pros**: Zero-downtime rollback
- **Cons**: Requires dual deployment, state sync

**C)** Feature flag - Database code present but disabled by default
- **Pros**: Instant rollback (flip flag)
- **Cons**: Dead code in production, flag management

**D)** Snapshot database before migration, restore on failure
- **Pros**: Full state recovery
- **Cons**: Downtime during restore, lost writes

**E)** Write-in (describe alternative approach): Not applicable, there is no backwards compatibility or upgrade, this is an alpha non-released suite of services

---

## Section 11: Gap Analysis

### Q21: Missing Service-Template Features

**Question**: Are there service-template features jose-ja currently doesn't need?

**A)** SessionManager (jose-ja uses stateless API, no sessions)
- **Gap**: Not using SessionManager from builder
- **Impact**: Inconsistent with template pattern

**B)** Multi-tenancy (jose-ja currently single-tenant)
- **Gap**: No tenant_id/realm_id in schema
- **Impact**: Can't support multi-tenant deployments

**C)** Browser UI (jose-ja is API-only service)
- **Gap**: No `/browser/**` paths, no CSRF middleware
- **Impact**: Can't add admin UI later without refactoring

**D)** ALL of the above are gaps that should be addressed
- **Gap**: Incomplete service-template adoption
- **Impact**: Divergence from standard pattern

**E)** Write-in (describe other gaps): All service-template features MUST be used for jose-ja. No exceptions. Service-template is the baseline for all 9 product-services.

---

### Q22: Documentation Gaps

**Question**: What documentation is missing for successful migration?

**A)** Database schema evolution guide (how to add columns/tables later)
- **Gap**: No guide for post-migration schema changes
- **Impact**: Future developers don't know migration patterns

**B)** Multi-instance deployment guide (load balancer config, session affinity)
- **Gap**: No guidance for horizontal scaling
- **Impact**: Production deployment issues

**C)** Key rotation procedures (how to rotate keys in database)
- **Gap**: No operational runbook
- **Impact**: Key lifecycle management unclear

**D)** Disaster recovery procedures (backup/restore, key export/import)
- **Gap**: No DR documentation
- **Impact**: Data loss risk, unclear recovery path

**E)** Write-in (describe other documentation gaps): No product or services docs, only plan and task-tracking docs in docs\jose-ja

---

## Status

**Total Questions**: 22 across 11 sections
**Questions Requiring User Input**: ALL 22 (analysis uncovered fundamental unknowns)

**After user answers**:
1. Update JOSE-JA-REFACTORING-PLAN.md with decisions
2. Update JOSE-JA-REFACTORING-TASKS.md with concrete implementation details
3. Proceed with refactoring execution

---

## Cross-References

- **Analysis**: [JOSE-JA-ANALYSIS.md](JOSE-JA-ANALYSIS.md) (to be updated after quiz answers)
- **Plan**: [JOSE-JA-REFACTORING-PLAN.md](JOSE-JA-REFACTORING-PLAN.md) (to be updated after quiz answers)
- **Tasks**: [JOSE-JA-REFACTORING-TASKS.md](JOSE-JA-REFACTORING-TASKS.md) (to be updated after quiz answers)
- **Service-Template**: [03-08.server-builder.instructions.md](../../.github/instructions/03-08.server-builder.instructions.md)
- **cipher-im Reference**: [internal/apps/cipher/im/](../../internal/apps/cipher/im/)
