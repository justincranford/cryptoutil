# cryptoutil CLARIFY-QUIZME-05.md

**Last Updated**: 2025-12-24
**Purpose**: Multiple choice questions for UNKNOWN answers requiring user input
**Format**: A-D options + E write-in for each question

## Instructions

**CRITICAL**: This file contains ONLY questions with UNKNOWN answers that require user clarification.

**Questions with KNOWN answers belong in clarify.md, NOT here.**

**When adding questions**:

1. Search copilot instructions, constitution.md, spec.md, codebase FIRST
2. If answer is KNOWN: Add Q&A to clarify.md and update constitution/spec as needed
3. If answer is UNKNOWN: Add question HERE with NO pre-filled answers
4. After user answers: Refactor clarify.md to cover answered questions, update constitution.md with architecture decisions, and update spec.md with finalized requirements

---

## Open Questions Requiring User Input

### Architecture & Deployment

#### Q1: Horizontal Scaling - Session State Management Decision

Current state: clarify.md lists 4 session state management patterns for horizontal scaling but NO clear decision on which to implement or prioritize.

**Options listed** (spec.md lines 1638-1642):

- Stateless sessions (Preferred): JWT tokens, no server-side storage
- Sticky sessions: Load balancer affinity based on session cookie
- Distributed session store: Redis cluster for shared session state
- Database-backed sessions: PostgreSQL with connection pooling

**UNKNOWN**: Which pattern(s) should be implemented in cryptoutil services?

A. Implement ALL 4 patterns with configuration-driven selection per deployment
B. Implement ONLY stateless sessions (JWT) + database-backed sessions (PostgreSQL)
C. Implement ONLY stateless sessions (JWT) + sticky sessions (load balancer affinity)
D. Implement ONLY database-backed sessions (PostgreSQL/SQLite) - no JWT, no Redis, no sticky sessions
E.

**Context**: Constitution.md Section VB mentions patterns but doesn't mandate implementation. Spec.md shows options but no decision. Clarify.md lacks Q&A on this.

---

#### Q2: Database Sharding - Timeline and Implementation Approach

Current state: spec.md line 1649 mentions "Database sharding: Partition data by tenant ID or key range (future consideration)" but NO decision on WHEN or HOW.

**UNKNOWN**: When should database sharding be implemented and what partition strategy?

A. Implement sharding in Phase 4 alongside multi-tenancy features, partition by tenant ID
B. Implement sharding in Phase 6+ only when performance metrics show bottleneck, partition by key range
C. Defer sharding indefinitely - use read replicas and connection pooling only until proven insufficient
D. Implement sharding in Phase 3 before Identity product goes production, partition by service type
E.

**Context**: Constitution.md Section VB lists it as pattern but no timeline. No clarify.md Q&A exists. No code implementation found.

---

#### Q3: Multi-Tenancy Isolation - Schema vs Table-Level Decision

Current state: clarify.md Q10.2 mentions "Preferred: Schema-level" and "Acceptable: Tenant ID column" but NO mandate or implementation decision.

**UNKNOWN**: Which multi-tenancy isolation pattern MUST be implemented?

A. Implement ONLY schema-level isolation (tenant_a.users, tenant_b.users)
B. Implement BOTH schema-level AND table-level with configuration selection
C. Implement ONLY table-level isolation with row-level security (RLS)
D. Defer multi-tenancy entirely - single-tenant deployments only for now
E.

**Context**: Constitution.md lacks multi-tenancy section. Spec.md mentions it but no mandate. Clarify.md says "Preferred" without implementation requirement.

---

### Security & Cryptography

#### Q4: TLS Client Certificate Authentication - Revocation Checking Implementation

Current state: constitution.md line 111 mentions "Certificate serial numbers... valid... revoked status" and authentication methods include mTLS, but NO specification for HOW revocation is checked (CRL, OCSP, OCSP stapling).

**UNKNOWN**: What revocation checking mechanisms MUST be implemented for mTLS client certificates?

A. Implement OCSP only (RFC 6960) with soft-fail if responder unavailable
B. Implement CRL only with periodic refresh (24-hour cache)
C. Implement BOTH OCSP and CRL with OCSP preferred, CRL fallback
D. Implement OCSP stapling (RFC 6066) - server provides OCSP response to avoid client lookup
E.

**Context**: Constitution.md mandates validation but not HOW. Spec.md lacks revocation details. Clarify.md has no Q&A on this.

---

#### Q5: Unseal Secrets - Key Derivation vs Direct JWK Storage

Current state: constitution.md Section III mentions "derive same JWKs with same kids, OR use same JWKs in enclave" but NO decision on which approach is REQUIRED.

**UNKNOWN**: Should unseal secrets derive keys deterministically OR store pre-generated JWKs directly?

A. ALWAYS derive keys from unseal secrets using HKDF (deterministic, reproducible)
B. ALWAYS store pre-generated JWKs in Docker/Kubernetes secrets (no derivation)
C. Support BOTH approaches with configuration selection per deployment
D. Use derivation for development/testing, pre-generated JWKs for production
E.

**Context**: Constitution.md line 121 says "or use same JWKs" suggesting choice. No clarify.md Q&A. No code archaeology confirms pattern.

---

#### Q6: Pepper Rotation Procedure - Downtime and Migration Strategy

Current state: constitution.md, spec.md, and clarify.md all mention "Changing pepper REQUIRES version bump, re-hash all records" but NO specification for HOW to do this without downtime or data loss.

**UNKNOWN**: What is the REQUIRED procedure for rotating pepper without service interruption?

A. Blue-green deployment: Migrate data to new version in parallel, cut over atomically
B. Rolling update: Accept both old and new versions during grace period, gradually re-hash records
C. Scheduled maintenance window: Take service offline, re-hash all records, restart service
D. Lazy migration: Re-hash records opportunistically as users re-authenticate
E.

**Context**: All docs mention version bump requirement but no migration procedure. Clarify.md lacks Q&A on operational procedure.

---

### Testing & Quality

#### Q7: Race Detector Workflow - Probabilistic Test Execution Conflict

Current state: constitution.md mandates probabilistic execution (TestProbQuarter, TestProbTenth) to stay under 15s/package limit. Race detector adds ~10× overhead.

**UNKNOWN**: Should race detector workflow disable probabilistic execution and accept longer runtimes?

A. Disable probabilistic execution in race detector - run ALL test cases with 10× longer timeout (150s/package)
B. Keep probabilistic execution enabled - accept that some race conditions may not be caught
C. Use HIGHER probability in race detector (TestProbHalf instead of TestProbTenth) - balance coverage vs time
D. Run race detector on subset of packages only (high-risk packages like crypto, concurrency)
E.

**Context**: Constitution.md mandates parallel tests + shuffle + probabilistic execution. Race detector ~10× overhead conflicts with 15s limit. No clarify.md Q&A.

---

#### Q8: E2E Test Scope - Browser-Based vs Service-to-Service API Coverage

Current state: clarify.md Q10 says "JOSE + CA + KMS first, Identity later" but NO specification of WHICH API paths to test (/browser/api/v1/*vs /service/api/v1/* vs both).

**UNKNOWN**: Which API paths MUST be covered by E2E tests for each product?

A. Test ONLY /service/api/v1/*paths (headless clients, simpler authentication)
B. Test ONLY /browser/api/v1/* paths (browser clients, full middleware stack)
C. Test BOTH /service/*and /browser/* paths with separate E2E scenarios per path type
D. Test /service/*paths only in Phase 2, add /browser/* paths in Phase 3+
E.

**Context**: Spec.md defines dual paths but E2E tests currently test health checks only. Clarify.md Q10 lists workflows but not path coverage.

---

#### Q9: Mutation Testing - Generated Code Exemption Criteria

Current state: clarify.md Q11 says "Allow exemptions for generated code with ramp-up plan" but NO specific criteria for WHAT constitutes "generated code" warranting exemption.

**UNKNOWN**: What code qualifies as "generated code" eligible for mutation testing exemption?

A. ONLY OpenAPI-generated models/clients (oapi-codegen output)
B. OpenAPI-generated code + GORM auto-migration code + protobuf (if used)
C. Any code not written by human developers (includes third-party libraries)
D. No exemptions - ALL code including generated code MUST meet 85%/98% mutation targets
E.

**Context**: Clarify.md mentions exemption but no criteria. Constitution.md says "≥85% per package" with no exemptions listed.

---

### Observability & Operations

#### Q10: Telemetry Sampling Strategy - Adaptive Sampling Configuration

Current state: spec.md line 1747 mentions "Sampling strategy: Adaptive based on throughput (100% at low load, 10% at high load)" but NO specification of thresholds or algorithm.

**UNKNOWN**: What are the exact thresholds and algorithm for adaptive sampling?

A. Linear scaling: 100% at 0-100 req/s, 50% at 100-1000 req/s, 10% at >1000 req/s
B. Exponential decay: 100% until 500 req/s, then halve every 2× traffic increase (50% at 1000, 25% at 2000, etc.)
C. Head-based sampling: Always sample first request in trace, probabilistically sample remaining spans
D. Tail-based sampling: Sample based on error status, latency >1s, or other quality signals
E.

**Context**: Spec.md mentions adaptive sampling but no algorithm. Clarify.md lacks Q&A. OpenTelemetry Collector config not found.

---

#### Q11: Health Check Failure Tolerance - Kubernetes vs Docker Behavior

Current state: constitution.md and spec.md define health check retries (5 retries, 5s interval) but NO specification of what happens AFTER all retries exhausted.

**UNKNOWN**: What action should orchestrator take when health checks fail after all retries?

A. Kubernetes: Remove from load balancer + restart pod. Docker Compose: Mark unhealthy + continue running
B. Kubernetes: Remove from load balancer only (no restart). Docker Compose: Stop container
C. Both: Restart service immediately after failure
D. Both: Mark unhealthy but take no action (manual intervention required)
E.

**Context**: Health check config defined but failure action not specified. Different for K8s (liveness vs readiness) vs Docker Compose.

---

### Federation & Service Integration

#### Q12: Federation Timeout Configuration - Per-Service vs Global

Current state: federation config examples show per-service timeouts (identity_timeout: 10s, jose_timeout: 10s) but NO decision on whether this is REQUIRED or if global timeout acceptable.

**UNKNOWN**: Should federation timeouts be configurable per federated service or use single global timeout?

A. REQUIRED: Per-service timeout configuration (identity_timeout, jose_timeout, ca_timeout separate)
B. ACCEPTABLE: Single global federation_timeout with per-service override option
C. SIMPLE: Single global timeout only, no per-service configuration
D. DYNAMIC: Auto-tune timeout based on observed latency percentiles (P99)
E.

**Context**: Examples show per-service but no mandate. Clarify.md Q8.1 mentions circuit breaker but not timeout granularity.

---

#### Q13: Cross-Service API Versioning - Backward Compatibility Strategy

Current state: APIs use /v1/ path prefix but NO specification for how to handle API version upgrades when services are deployed independently.

**UNKNOWN**: How should API versioning be handled across federated services?

A. Strict version matching: Identity v2 ONLY works with JOSE v2, KMS v2 (synchronized releases)
B. Backward compatible: Identity v2 works with JOSE v1 OR v2 (support N-1 version)
C. API gateway translation: Gateway translates between API versions transparently
D. Version negotiation: Services advertise supported versions, negotiate at runtime
E.

**Context**: Spec.md shows /v1/ paths but no upgrade strategy. Federation examples assume same version. No clarify.md Q&A.

---

#### Q14: Service Discovery - DNS TTL and Caching Behavior

Current state: Service discovery mechanisms listed (config file, Docker Compose DNS, K8s DNS) but NO specification of DNS caching or refresh intervals.

**UNKNOWN**: How should services handle DNS caching for federated service URLs?

A. Respect DNS TTL with minimum 30s refresh (prevent stale DNS)
B. Cache DNS lookups for 5 minutes, refresh on connection failure
C. No caching - perform DNS lookup on every request (highest reliability, higher latency)
D. Use service mesh (Istio/Linkerd) for DNS management (external to services)
E.

**Context**: Spec.md lists discovery mechanisms but not caching behavior. Clarify.md lacks Q&A on DNS TTL.

---

### Performance & Scalability

#### Q15: Connection Pool Sizing - Formula and Tuning Guidance

Current state: spec.md line 1630 says "Connection pool sizing: Based on workload (PostgreSQL 10-50, SQLite 5)" but NO formula or tuning guidance.

**UNKNOWN**: What is the REQUIRED formula for determining connection pool size?

A. Formula: (max_concurrent_requests / avg_request_duration_sec) × safety_factor_1.5
B. Fixed values: PostgreSQL always 25, SQLite always 5 (no dynamic sizing)
C. Workload-based: Low (<100 req/min) = 10, Medium (100-1000) = 25, High (>1000) = 50
D. Auto-tune: Start at 10, increment by 5 when pool exhaustion detected, decrement by 5 when idle
E.

**Context**: Spec.md mentions ranges but no formula. Constitution.md Section VB lacks formula. Clarify.md has no Q&A.

---

#### Q16: Read Replica Strategy - Lag Tolerance and Fallback

Current state: spec.md line 1648 mentions "Read replicas: Route read-only queries to PostgreSQL replicas" but NO specification of replication lag tolerance or fallback to primary.

**UNKNOWN**: What replication lag is acceptable and when should reads fall back to primary?

A. Max lag 1s - if replica lags >1s, route read to primary
B. Max lag 5s - acceptable for non-critical reads (reports, dashboards)
C. No lag limit - always use replica for reads (accept stale data)
D. Dynamic: Strong consistency reads → primary, eventual consistency reads → replica
E.

**Context**: Spec.md mentions read replicas but no lag tolerance. Clarify.md lacks Q&A on consistency requirements.

---

### CI/CD & Workflows

#### Q17: GitHub Actions Workflow Dependencies - Required vs Optional Services

Current state: ci-race, ci-mutation, ci-coverage require PostgreSQL service, but NO specification of which OTHER workflows need PostgreSQL.

**UNKNOWN**: Which workflows beyond ci-race/ci-mutation/ci-coverage REQUIRE PostgreSQL service container?

A. REQUIRED for: ci-race, ci-mutation, ci-coverage, ci-e2e, ci-identity-validation (5 workflows)
B. REQUIRED for: Only ci-race, ci-mutation, ci-coverage (3 workflows as documented)
C. REQUIRED for: ALL workflows that run `go test` command (conservative, safe)
D. OPTIONAL for: All workflows - use test-containers library instead of service container
E.

**Context**: Spec.md line 1704 says "ANY workflow executing go test" but table shows only 5 workflows with PostgreSQL=✅. Ambiguous.

---

#### Q18: Docker Image Pre-Pull Strategy - Performance vs Reliability Trade-off

Current state: .github/actions/docker-images-pull exists for parallel pre-pull but NO decision on WHEN to use vs letting workflows pull on-demand.

**UNKNOWN**: When should workflows use docker-images-pull action vs on-demand pulling?

A. ALWAYS use pre-pull for ALL workflows (consistent startup time)
B. Use pre-pull ONLY for E2E/load testing workflows (high image count)
C. Use pre-pull ONLY when workflow timeout is concern (>20 images)
D. NEVER use pre-pull - on-demand pulling with Docker layer caching sufficient
E.

**Context**: cross-platform.md describes action but no usage policy. Some workflows use it, others don't. Inconsistent.

---

---

## Next Steps After User Answers

1. **Refactor clarify.md**: Merge all answered QUIZME questions into clarify.md with topical organization
2. **Update constitution.md**: Incorporate architectural decisions and mandates from answers
3. **Update spec.md**: Add finalized requirements and detailed specifications
4. **Update copilot instructions**: Ensure .github/instructions/*.instructions.md reflect decisions
5. **Generate/update plan.md**: Adjust implementation plan based on clarifications
6. **Generate/update tasks.md**: Create concrete tasks with dependencies and acceptance criteria

---

**Status**: 18 questions identified requiring user clarification (architecture: 3, security: 3, testing: 3, observability: 2, federation: 3, performance: 2, CI/CD: 2)
