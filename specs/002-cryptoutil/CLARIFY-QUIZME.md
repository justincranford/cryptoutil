# cryptoutil CLARIFY-QUIZME.md

**Generated**: 2025-12-21
**Purpose**: Multiple choice questions identifying problems, omissions, ambiguities, conflicts, risks, and validation gaps in constitution.md, spec.md, and PLAN.md
**Format**: A-D options + E write-in for each question

## Service Architecture Questions

### Question 1: Port Range Conflicts

Do the assigned port ranges for all 9 services create any conflicts or overlaps?

- A) No conflicts - all ranges are unique and properly spaced
- B) Minor overlap between JOSE (9443-9449) and CA (8443-8449) ranges
- C) Identity services have overlapping ranges within the 18000-18409 block
- D) Learn-PS range (8888-8889) conflicts with other services
- E) Write-in: [specific conflict description]

**Correct Answer**: A, all port ranges for public HTTPS ports are unique and properly spaces; because private HTTPS ports are never exposed from containers, and private HTTPS ports in unit/integration tests are port 0 (dynamic)

### Question 2: Admin Port Standardization

Is the standardization of all admin ports to 127.0.0.1:9090 appropriate for service isolation?

- A) Yes - single admin port simplifies monitoring and firewall rules
- B) No - each service should have unique admin port for better isolation
- C) Partially - core services can share, but identity services need separate ports
- D) No - conflicts with existing service assignments (9091, 9092, 9093)
- E) Write-in: [alternative approach]

**Correct Answer**: A, private HTTPS ports are never exposed from containers, and private HTTPS ports in unit/integration tests are port 0 (dynamic)

### Question 3: Service Implementation Status Accuracy

Does the constitution.md service status table accurately reflect current implementation?

- A) Yes - all status indicators match file system verification
- B) No - RS shows "IN PROGRESS" but evidence claims "IMPLEMENTED"
- C) No - authz/idp show complete but may be missing federation configs
- D) No - CA shows complete but constitution notes "missing admin server"
- E) Write-in: [specific inaccuracy]

**Correct Answer**: B

## Federation and Integration Questions

### Question 4: Cross-Service Communication

Are federation patterns adequately specified for all service interactions?

- A) Yes - all services have clear federation URLs and timeout configs
- B) No - missing federation configs for JOSE ↔ Identity communication
- C) No - CA service federation not specified for TLS certificate operations
- D) No - KMS federation fallback modes incomplete
- E) Write-in: [missing federation pattern]

**Correct Answer**: A

### Question 5: Service Discovery Mechanisms

Do the specified service discovery mechanisms cover all deployment scenarios?

- A) Yes - config files, Docker Compose, Kubernetes DNS all supported
- B) No - missing environment variable override patterns
- C) No - no specification for service mesh discovery (Istio, Linkerd)
- D) No - missing DNS-SD (DNS Service Discovery) support
- E) Write-in: [missing discovery mechanism]

**Correct Answer**: A

## Testing and Quality Questions

### Question 6: Coverage Targets Consistency

Are the coverage targets consistently applied across all documents?

- A) Yes - 95% production, 100% infrastructure/utility in all docs
- B) No - constitution mentions different targets than testing instructions
- C) No - mutation testing targets vary between Phase 4 (85%) and Phase 5 (98%)
- D) No - some services may require different targets based on complexity
- E) Write-in: [inconsistency description]

**Correct Answer**: A

### Question 7: Test Timing Requirements

Are the test timing requirements realistic for the current codebase size?

- A) Yes - <15s per package, <180s total aligns with current performance
- B) No - current tests already exceed 15s for some packages
- C) No - <180s total too aggressive for full test suite with integration tests
- D) No - probabilistic execution needed but not specified for slow packages
- E) Write-in: [timing issue]

**Correct Answer**: A

## Security and Compliance Questions

### Question 8: FIPS Algorithm Restrictions

Are the FIPS 140-3 algorithm restrictions clearly communicated?

- A) Yes - banned algorithms (bcrypt, scrypt, MD5, SHA-1) explicitly listed
- B) No - missing guidance on approved key sizes for RSA/ECDSA
- C) No - no clarification on when non-FIPS algorithms acceptable (dev/test)
- D) No - password hashing requirements could be more specific
- E) Write-in: [missing FIPS guidance]

**Correct Answer**: A

### Question 9: TLS Configuration Patterns

Do the TLS configuration patterns adequately cover all deployment scenarios?

- A) Yes - external, mixed, auto-generated patterns cover production/dev/test
- B) No - missing specification for client certificate requirements
- C) No - no guidance on certificate chain validation depth
- D) No - OCSP stapling configuration not specified
- E) Write-in: [missing TLS pattern]

**Correct Answer**: A

## Implementation Risk Questions

### Question 10: Phase Dependency Risks

Are the strict phase dependencies (1→2→3) creating unacceptable risks?

- A) No - dependencies are logical and reduce integration risks
- B) Yes - Phase 3 blocked by Phase 2 completion creates timeline risk
- C) Yes - identity-rp/spa as "Phase 3+" may delay production readiness
- D) Yes - Learn-PS as Phase 7 creates long feedback cycle
- E) Write-in: [dependency risk]

**Correct Answer**: A

### Question 11: Database Compatibility

Are cross-database compatibility requirements sufficiently specified?

- A) Yes - TEXT type for UUID, GORM serializer:json clearly documented
- B) No - missing guidance on nullable foreign keys in SQLite
- C) No - no specification for handling database-specific SQL features
- D) No - concurrent write handling for SQLite WAL mode incomplete
- E) Write-in: [missing database guidance]

**Correct Answer**: A

## Documentation Consistency Questions

### Question 12: Service Naming Conventions

Are service naming conventions consistent across all documents?

- A) Yes - sm-kms, pki-ca, jose-ja, identity-* patterns uniform
- B) No - constitution uses "Full Name" format, spec uses "Service" format
- C) No - port assignments vary between documents
- D) No - admin port assignments inconsistent (9090 vs 9091-9093)
- E) Write-in: [naming inconsistency]

**Correct Answer**: A (after recent fixes)

### Question 13: Success Criteria Clarity

Are success criteria clearly defined and measurable?

- A) Yes - coverage %, mutation %, test passing, workflow status all measurable
- B) No - "production ready" criteria too vague
- C) No - "comprehensive security solution" lacks specific metrics
- D) No - E2E demo requirements not detailed enough
- E) Write-in: [unclear criteria]

**Correct Answer**: A

## Operational Questions

### Question 14: Health Check Patterns

Are health check patterns adequately specified for all services?

- A) Yes - /admin/v1/livez, readyz, healthz, shutdown clearly defined
- B) No - missing specification for startup probe timing
- C) No - no guidance on health check frequency in production
- D) No - container healthcheck commands not specified for all services
- E) Write-in: [missing health pattern]

**Correct Answer**: A

### Question 15: Log Aggregation Strategy

Is the log aggregation strategy clear for distributed deployments?

- A) Yes - otel-collector sidecar pattern with OTLP forwarding specified
- B) No - missing specification for log levels and structured logging
- C) No - no guidance on log retention and rotation
- D) No - error correlation across services not specified
- E) Write-in: [missing logging guidance]

**Correct Answer**: A

## Risk Assessment Questions

### Question 16: Single Points of Failure

Are single points of failure adequately identified and mitigated?

- A) Yes - federation fallback modes, circuit breakers, graceful degradation specified
- B) No - database single point of failure not addressed
- C) No - otel-collector as single telemetry aggregation point
- D) No - shared admin port creates monitoring blind spots
- E) Write-in: [unidentified SPOF]

**Correct Answer**: A

### Question 17: Performance Scaling

Are performance scaling requirements specified for production deployments?

- A) Yes - concurrent testing, timing targets, resource pooling addressed
- B) No - missing horizontal scaling guidance for services
- C) No - no specification for database connection pooling limits
- D) No - load testing scenarios not detailed enough
- E) Write-in: [missing scaling guidance]

**Correct Answer**: B

### Question 18: Backup and Recovery

Are backup and recovery procedures specified?

- A) Yes - database migrations, key versioning, rotation patterns cover recovery
- B) No - missing specification for configuration backup
- C) No - no guidance on log backup and point-in-time recovery
- D) No - certificate revocation and re-issuance not detailed
- E) Write-in: [missing backup guidance]

**Correct Answer**: A

## Validation Questions

### Question 19: Integration Testing Scope

Is the integration testing scope comprehensive enough?

- A) Yes - E2E workflows, Docker Compose, health checks cover integration
- B) No - missing specification for API contract testing
- C) No - no guidance on chaos engineering or fault injection
- D) No - federation integration tests not specified
- E) Write-in: [missing integration test]

**Correct Answer**: A

### Question 20: Documentation Maintenance

Is the documentation maintenance process clear?

- A) Yes - living documents, mini-cycle feedback, DETAILED.md tracking specified
- B) No - missing specification for document version control
- C) No - no guidance on when to update vs create new documents
- D) No - post-mortem integration not detailed enough
- E) Write-in: [missing maintenance guidance]

**Correct Answer**: A

---

### Q1.3: Relying Party (identity-rp) and SPA (identity-spa) Mandatory Status

The spec labels identity-rp and identity-spa as "reference implementations" but constitution lists them as core services.

**Are identity-rp and identity-spa MANDATORY for all cryptoutil deployments?**

A. MANDATORY: Both services MUST be deployed in every cryptoutil installation (production, dev, test)
B. OPTIONAL: Both services are reference implementations; customers may omit if building custom clients
C. CONDITIONAL: Required for browser-based deployments, optional for service-to-service only deployments
D. PARTIAL: identity-rp is MANDATORY (BFF pattern), identity-spa is OPTIONAL (static example)
E. Other (please specify):

**Answer**: _____

**Follow-up if B or C or D**: How should Docker Compose configurations handle optional services?

A. Separate compose files: `compose.yml` (core only), `compose.full.yml` (includes optional services)
B. Profile-based: Use `--profile full` to enable optional services
C. Always include: Deploy all services by default, document how to remove optional ones
D. Environment variable: Set `DEPLOY_REFERENCE_IMPLEMENTATIONS=true|false`
E. Other (please specify):

**Answer**: _____

---

### Q1.4: Learn-PS Pet Store Service Deployment Environments

The constitution and spec list learn-ps as a demonstration service but don't specify deployment requirements.

**In which environments should learn-ps be deployed?**

A. ALL ENVIRONMENTS: Production, staging, development, testing (always deployed)
B. DEV/TEST ONLY: Development and testing environments (NEVER production or staging)
C. OPTIONAL DEMO: Separate Docker Compose file (`compose.demo.yml`), opt-in deployment
D. DOCUMENTATION ONLY: Code exists in repository but no deployable artifacts (customer copy-paste pattern)
E. Other (please specify):

**Answer**: _____

**Follow-up if A or B or C**: Should learn-ps share infrastructure (PostgreSQL, OTLP) with product services or have isolated stack?

A. Shared: learn-ps uses same PostgreSQL and OTLP collector as product services
B. Isolated: learn-ps has dedicated PostgreSQL and OTLP collector (separate compose network)
C. Hybrid: Shared OTLP for telemetry, isolated PostgreSQL for data separation
D. Configurable: Support both shared and isolated modes via Docker Compose profiles
E. Other (please specify):

**Answer**: _____

---

## Section 2: Authentication and Authorization

### Q2.1: Browser Client Authentication Methods in Standalone vs Federated Mode

The spec states "different option depending if a service is deployed standalone vs federated with Identity product" but lists overlapping methods.

**For standalone mode (without Identity product), which authentication methods are MANDATORY for /browser/** paths?**

A. Basic (Username/Password) ONLY - minimal viable authentication
B. Basic (Username/Password) + Bearer (API Token) - two methods required
C. Basic (Username/Password) + Session Cookie - username/password login creates session
D. ALL LISTED: Basic (Username/Password), Basic (Email/Password), Bearer (API Token), Session Cookie
E. Other (please specify):

**Answer**: _____

**Follow-up**: For federated mode (with Identity product), is the complete MFA list MANDATORY or OPTIONAL?

A. MANDATORY: All MFA methods (WebAuthn, TOTP, Email OTP, SMS OTP, HOTP, Recovery Codes, Push, Phone Call) MUST be implemented
B. MINIMUM VIABLE: WebAuthn (Passkeys) + TOTP + Email OTP required, others optional
C. TIERED: Tier 1 (WebAuthn + TOTP), Tier 2 (+ Email OTP + SMS OTP), Tier 3 (+ all others)
D. CONDITIONAL: Based on NIST AAL level (AAL1=password, AAL2=2FA, AAL3=hardware token)
E. Other (please specify):

**Answer**: _____

---

### Q2.2: Service Client Authentication Methods Priority and Defaults

The spec lists multiple authentication methods for /service/** paths but doesn't specify defaults or priority order.

**What is the default authentication method for /service/** paths in standalone mode?**

A. Basic (clientid,clientsecret) - HTTP Basic Auth with client credentials
B. Bearer (API Token) - Pre-provisioned API keys
C. mTLS (client certificate) - Certificate-based authentication
D. None by default - user MUST explicitly configure authentication method in YAML
E. Other (please specify):

**Answer**: _____

**Follow-up**: If multiple authentication methods are configured, what is the priority order?

A. First match: Try methods in YAML order, accept first successful authentication
B. Most secure first: mTLS > JWT > Basic > API Token (fallback to least secure)
C. Explicit per-endpoint: YAML config specifies which method per endpoint/scope
D. Require all: Client MUST satisfy all configured methods (additive, not alternative)
E. Other (please specify):

**Answer**: _____

---

### Q2.3: Session Token Format and Storage

The spec mentions "session cookie (opaque||JWE||JWS non-OAuth 2.1)" but doesn't specify defaults.

**What is the default session token format?**

A. Opaque: Random UUID stored in server-side session store (Redis/database)
B. JWE (JSON Web Encryption): Encrypted JWT, server-side session store not required
C. JWS (JSON Web Signature): Signed JWT, server-side session store not required
D. Configurable: YAML setting `session_format: opaque|jwe|jws`, default is opaque
E. Other (please specify):

**Answer**: _____

**Follow-up if A or D (opaque)**: Which session store backend is default?

A. In-memory: Go map with sync.RWMutex (single-instance only, lost on restart)
B. Redis: External Redis server (supports multi-instance, persists across restarts)
C. Database: PostgreSQL/SQLite table (same DB as application data)
D. Configurable: YAML setting `session_store: memory|redis|database`, default is memory for dev, redis for prod
E. Other (please specify):

**Answer**: _____

---

## Section 3: Database and Schema Management

### Q3.1: Multi-Instance PostgreSQL Deployment Pattern

The spec shows multiple PostgreSQL instances (kms-postgres-1, kms-postgres-2) but doesn't clarify deployment strategy.

**What is the purpose of multiple PostgreSQL instances in Docker Compose?**

A. High Availability: Primary-replica setup with automatic failover
B. Database Sharding: kms-postgres-1 handles tenants A-M, kms-postgres-2 handles tenants N-Z
C. Backend Demonstration: Show same service works with multiple DB instances (fixed, not dynamic)
D. Load Balancing: Round-robin distribution of read queries across instances
E. Other (please specify):

**Answer**: _____

**Follow-up if A or D**: How is failover/load balancing configured?

A. PgPool-II: PostgreSQL connection pooler and load balancer (separate container)
B. Application-level: Service implements connection retry and failover logic
C. Kubernetes Operator: Use postgres-operator for HA management
D. Manual: No automatic failover, operator manually switches connection strings
E. Other (please specify):

**Answer**: _____

---

### Q3.2: SQLite vs PostgreSQL Feature Parity

The constitution mandates dual support for SQLite and PostgreSQL but doesn't define feature parity requirements.

**Must ALL features work identically on SQLite and PostgreSQL?**

A. STRICT PARITY: Every feature MUST work identically on both backends (no exceptions)
B. BEST EFFORT: Features use SQL common subset, database-specific features gracefully degrade
C. POSTGRESQL PRIMARY: PostgreSQL is primary, SQLite is development/testing only (subset of features OK)
D. FEATURE FLAGS: YAML config declares feature availability per database backend
E. Other (please specify):

**Answer**: _____

**Follow-up**: Which database-specific features (if any) are allowed to diverge?

A. None: Strict parity required (see answer A above)
B. Performance optimizations only: PostgreSQL uses partitioning/indexing not available in SQLite, but API behavior identical
C. Advanced features optional: Full-text search, JSONB queries, stored procedures allowed on PostgreSQL only
D. Documented exceptions: Maintain compatibility matrix in docs (e.g., "recursive CTEs require PostgreSQL")
E. Other (please specify):

**Answer**: _____

---

### Q3.3: Database Migration Strategy for Multi-Service Deployments

With 9 services sharing infrastructure, database migration coordination is critical.

**How should database migrations be coordinated across 9 services?**

A. Independent schemas: Each service has dedicated PostgreSQL schema (service1.*, service2.*), migrations run independently
B. Independent databases: Each service has dedicated PostgreSQL database, migrations run independently
C. Shared database: All services share single database, migrations coordinated via migration tool (Flyway/Liquibase)
D. First-instance pattern: First service instance to start runs all migrations, subsequent instances skip
E. Other (please specify):

**Answer**: _____

**Follow-up if C or D**: How are migration conflicts prevented when multiple instances start simultaneously?

A. Advisory locks: PostgreSQL advisory locks ensure only one instance runs migrations
B. Migration table: Checksum-based migration tracking prevents duplicate execution
C. Leader election: Kubernetes leader election sidecar ensures single migration runner
D. Sequential startup: Docker Compose depends_on with health checks enforces startup order
E. Other (please specify):

**Answer**: _____

---

## Section 4: Cryptography and Key Management

### Q4.1: FIPS 140-3 Module Certification vs Algorithm Compliance

The constitution states "FIPS 140-3 mode is ALWAYS enabled" but doesn't clarify certification vs algorithm compliance.

**What does "FIPS 140-3 compliance" mean for cryptoutil?**

A. Certified module: Use FIPS 140-3 validated cryptographic module (e.g., OpenSSL FIPS, BoringCrypto)
B. Algorithm compliance: Use only FIPS-approved algorithms, but not necessarily certified module
C. Aspirational: Follow FIPS guidelines, plan for future certification (currently algorithm compliance only)
D. Conditional: Offer FIPS-certified build variant alongside standard build
E. Other (please specify):

**Answer**: _____

**Follow-up if A or D**: Which FIPS 140-3 validated module should be used?

A. OpenSSL FIPS Module: Use FIPS-validated OpenSSL 3.x
B. BoringCrypto: Use Google's FIPS-validated BoringSSL fork
C. Go Crypto: Wait for Go standard library FIPS validation
D. External HSM: Delegate all crypto operations to FIPS-validated HSM (PKCS#11)
E. Other (please specify):

**Answer**: _____

---

### Q4.2: Unseal Secret Interoperability Requirement

The constitution states "ALL cryptoutil instances using the same unseal secrets MUST derive the same unseal JWKs, including KIDs".

**Does this requirement apply across different product types or only within the same product?**

A. Cross-product: KMS instance and CA instance using same unseal secrets MUST derive same JWKs
B. Same product only: KMS instance 1 and KMS instance 2 MUST match, but KMS and CA can diverge
C. Tenant-scoped: Within same tenant/database, instances MUST match; cross-tenant can diverge
D. Optional interoperability: Configuration flag enables/disables cross-instance JWK derivation
E. Other (please specify):

**Answer**: _____

**Follow-up**: What happens if unseal secrets differ between instances in a unified deployment?

A. FAIL FAST: Service startup fails with error "Unseal secret mismatch detected"
B. WARN AND CONTINUE: Log warning, allow startup (data encrypted by one instance won't decrypt on another)
C. AUTO-SYNC: First instance writes derived JWKs to database, subsequent instances read from DB
D. ISOLATED ENCLAVES: Each instance has isolated enclave, no cross-instance decryption required
E. Other (please specify):

**Answer**: _____

---

### Q4.3: Hash Service Version Update Trigger and Migration Strategy

The constitution defines hash versioning (v1, v2, v3) but doesn't specify update triggers.

**When should hash versions be updated (e.g., v1 → v2)?**

A. Manual operator decision: Admin manually updates config and restarts service
B. Automatic calendar-based: Every 12 months, auto-upgrade to next version
C. Security event-driven: When NIST/OWASP publishes new guidelines, prompt for upgrade
D. Gradual rollout: Blue-green deployment with version upgrade, validate before full switch
E. Other (please specify):

**Answer**: _____

**Follow-up**: How should existing hashes be migrated after version update?

A. No migration: Old hashes stay on original version, new hashes use new version (gradual transition)
B. Lazy migration: On next authentication, re-hash with new version and update database
C. Bulk migration: Background job re-hashes all passwords/PII during maintenance window
D. Force re-authentication: Invalidate all sessions, require users to re-authenticate (generates new hashes)
E. Other (please specify):

**Answer**: _____

---

## Section 5: Testing and Quality Gates

### Q5.1: Mutation Testing Threshold Enforcement

The constitution specifies "≥85% Phase 4, ≥98% Phase 5+" but doesn't define enforcement mechanism.

**How should mutation testing thresholds be enforced?**

A. CI/CD blocking: Workflow fails if mutation score < threshold
B. Pull request comments: Bot comments with score, but doesn't block merge
C. Dashboard tracking: Grafana dashboard shows trends, manual enforcement
D. Pre-commit hook: Local mutation testing runs before commit (blocks commit if < threshold)
E. Other (please specify):

**Answer**: _____

**Follow-up**: Should mutation testing run on every commit or only on main branch?

A. Every commit: All branches, all commits (slow feedback but maximum coverage)
B. Main branch only: Feature branches skip mutation testing (fast iteration)
C. Pull request gating: Run on PR creation and updates, not every commit
D. Scheduled: Nightly cron job runs mutation testing, reports next day
E. Other (please specify):

**Answer**: _____

---

### Q5.2: Test Execution Time Violation Remediation

The constitution mandates "<15s per package" but doesn't specify remediation when exceeded.

**What should happen when a test package exceeds 15-second limit?**

A. CI/CD failure: Workflow fails with error "Test package X exceeded 15s limit"
B. Warning only: Log warning, allow workflow to pass (soft limit)
C. Issue creation: Automatically create GitHub issue for remediation
D. Probabilistic enforcement: Enable TestProbTenth for slow packages until optimized
E. Other (please specify):

**Answer**: _____

**Follow-up if A or C**: What is the grace period before enforcement?

A. Immediate: Enforce limit starting today (may break existing workflows)
B. 30-day grace: Warning for 30 days, then enforcement
C. Per-package baseline: Grandfather existing slow packages, enforce for new packages only
D. Version-based: Enforce starting with next major version (2.0.0)
E. Other (please specify):

**Answer**: _____

---

### Q5.3: Coverage Target Exceptions for Generated Code

The constitution mandates "95%+ production, 100% infrastructure/utility" but doesn't address generated code (OpenAPI clients, GORM models).

**How should coverage be measured for generated code?**

A. Include in coverage: Generated code counts toward 95% target (must test generated code)
B. Exclude from coverage: Use `//go:generate` comments to exclude from coverage reports
C. Separate target: Generated code has lower target (e.g., 70%) separate from hand-written code
D. No generation: Avoid code generation, hand-write all code (ensures testability)
E. Other (please specify):

**Answer**: _____

**Follow-up if A or C**: How should generated OpenAPI server stubs be tested?

A. Integration tests only: Test via HTTP requests, don't unit test stubs
B. Mock business logic: Unit test stubs with mocked handlers
C. Contract testing: Use OpenAPI spec as contract, validate responses match spec
D. Ignore stubs: Focus testing on handlers, skip generated boilerplate
E. Other (please specify):

**Answer**: _____

---

## Section 6: CI/CD and Deployment

### Q6.1: Workflow Failure Notification Strategy

With 13 workflows, failure notifications could become noisy. What is the notification strategy?

A. All failures: Email/Slack notification for every workflow failure (13 potential notifications per push)
B. Critical only: Notify only for ci-quality, ci-coverage, ci-race failures (3 workflows)
C. Threshold-based: Notify if ≥3 workflows fail on same commit (aggregate)
D. Pull request only: Notify on PR workflow failures, ignore main branch noise
E. Other (please specify):

**Answer**: _____

**Follow-up**: Should workflow failures block deployment?

A. Strict blocking: ANY workflow failure blocks deployment (13 gates)
B. Critical workflows only: ci-quality, ci-coverage, ci-race, ci-sast MUST pass (4 gates)
C. Manual override: Failures block deployment unless admin explicitly approves
D. Conditional: Block production deployment, allow staging deployment with failures
E. Other (please specify):

**Answer**: _____

---

### Q6.2: Docker Compose Health Check Retry Strategy Standardization

The spec mentions "start_period: 30s, interval: 5s, retries: 10" but doesn't standardize across all services.

**Should all 9 services use identical health check retry configuration?**

A. Identical: All services use same values (30s start, 5s interval, 10 retries)
B. Tiered: Database-dependent services (KMS, Identity) use longer timeouts, others use shorter
C. Service-specific: Each service tunes based on measured startup time (manual configuration)
D. Auto-tuning: Health check script measures startup time, adjusts retries dynamically
E. Other (please specify):

**Answer**: _____

**Follow-up**: What is the maximum acceptable health check window (start_period + interval × retries)?

A. 60 seconds: Short window for fast feedback (may cause false failures)
B. 90 seconds: Balanced for typical startup (current spec: 30s + 5s×10 = 80s)
C. 120 seconds: Conservative for slow environments (CI/CD, low-resource VMs)
D. 180 seconds: Maximum tolerance for cold starts (container image pull + DB migration)
E. Other (please specify):

**Answer**: _____

---

## Section 7: Documentation and Spec Kit Process

### Q7.1: Constitution Amendment Authority for Service Count Changes

The constitution lists "Four Working Products Goal" but now has 9 services. Is this an amendment or interpretation?

**How should service count changes be handled in the constitution?**

A. Amendment required: Adding services (9th service learn-ps) requires Section I amendment
B. Interpretation: "Four Products" is immutable, service count within products is implementation detail
C. Living document: Section I is "Living Section" updated during implementation without formal amendment
D. Version control: Constitution version increments (3.1.0 → 3.2.0) for service additions
E. Other (please specify):

**Answer**: _____

**Follow-up**: Should future products (P5: Secrets Vault, P6: Policy Engine) require constitutional amendment?

A. YES: Adding P5/P6 requires formal amendment with stakeholder approval
B. NO: Constitution's "Four Products" is historical, expand freely based on implementation needs
C. CONDITIONAL: New products require amendment if cross-cutting concerns (database, crypto), otherwise implementation detail
D. VERSIONING: Major version increment (4.0.0) allows product expansion without amendment
E. Other (please specify):

**Answer**: _____

---

### Q7.2: Spec Kit Clarify Document Update Frequency

Spec Kit methodology includes /speckit.clarify but doesn't specify update cadence.

**When should clarify.md be updated during implementation?**

A. Pre-implementation only: Run /speckit.clarify once before /speckit.implement, never update after
B. Continuous: Update clarify.md every time ambiguity is discovered during implementation
C. Milestone-based: Update at end of each Phase (Phase 1 done → update clarify.md)
D. Issue-driven: Update only when blocker encountered (implementation cannot proceed without clarification)
E. Other (please specify):

**Answer**: _____

**Follow-up**: Should clarify.md be regenerated or manually updated?

A. Always regenerate: Run /speckit.clarify to regenerate entire document (overwrites manual changes)
B. Manual append: Add new Q&A entries manually to preserve existing answers
C. Hybrid: Regenerate for new topics, preserve manually added deep-dive sections
D. Version control: Keep clarify-v1.md, clarify-v2.md, clarify-v3.md as historical record
E. Other (please specify):

**Answer**: _____

---

### Q7.3: CLARIFY-QUIZME.md Integration with Spec Kit Workflow

This document (CLARIFY-QUIZME.md) is not mentioned in Spec Kit methodology. How should it be integrated?

**What is the role of CLARIFY-QUIZME.md in the Spec Kit workflow?**

A. Pre-clarify: User answers CLARIFY-QUIZME.md questions BEFORE running /speckit.clarify (input to clarify)
B. Post-clarify: CLARIFY-QUIZME.md is generated AFTER /speckit.clarify to identify remaining gaps
C. Continuous: CLARIFY-QUIZME.md is living document updated whenever ambiguity discovered
D. Optional: CLARIFY-QUIZME.md is project-specific extension, not part of standard Spec Kit workflow
E. Other (please specify):

**Answer**: _____

**Follow-up**: After user answers CLARIFY-QUIZME.md, how should answers be integrated?

A. Manual: User copies answers into clarify.md manually
B. Automated: Script parses CLARIFY-QUIZME-ANSWERS.md and updates clarify.md
C. Regenerate: User provides answers as input to /speckit.clarify, regenerates clarify.md with answers
D. Separate: CLARIFY-QUIZME.md and clarify.md remain separate documents (no integration)
E. Other (please specify):

**Answer**: _____

---

## Section 8: Observability and Telemetry

### Q8.1: OTLP Collector Resource Allocation

With 9 services sending telemetry to single OTLP collector, resource limits may be needed.

**Should OTLP collector have resource limits in Docker Compose?**

A. No limits: Allow OTLP collector to consume unbounded CPU/memory (risk: starve other services)
B. Fixed limits: Set CPU=1, memory=512Mi (risk: collector drops telemetry under load)
C. Dynamic limits: Use Docker Compose `deploy.resources.limits` with autoscaling (requires Swarm mode)
D. Separate collector per service: 9 services → 9 OTLP collectors (increased resource usage)
E. Other (please specify):

**Answer**: _____

**Follow-up**: What is the telemetry sampling strategy under high load?

A. No sampling: Send 100% of traces/metrics/logs (risk: OTLP collector overwhelmed)
B. Tail-based sampling: Sample 10% of traces, keep all errors (requires stateful collector)
C. Probabilistic sampling: Sample 1% of all telemetry uniformly (risk: miss important events)
D. Adaptive sampling: Adjust sampling rate based on OTLP collector queue depth
E. Other (please specify):

**Answer**: _____

---

## Section 9: Security and Secrets Management

### Q9.1: Docker Secrets File Permissions

The constitution mandates Docker secrets for sensitive data but doesn't specify file permissions.

**What permissions should Docker secrets have in /run/secrets/?**

A. 400 (r--------): Owner read-only, no write/execute
B. 440 (r--r-----): Owner + group read, no write/execute
C. 600 (rw-------): Owner read/write, no execute (allows in-place updates)
D. 444 (r--r--r--): World-readable (insecure, but simplifies debugging)
E. Other (please specify):

**Answer**: _____

**Follow-up**: Should secrets be validated at startup or runtime?

A. Startup validation: Service reads all secrets on startup, fails fast if missing/invalid
B. Lazy loading: Secrets loaded on first use (allows service to start without secrets)
C. Health check: Secrets validated during /readyz probe (delays readiness until secrets valid)
D. No validation: Assume Docker Compose correctly mounts secrets (fail on first crypto operation)
E. Other (please specify):

**Answer**: _____

---

## Section 10: Product-Specific Questions

### Q10.1: Identity Service Session Sharing Across Services

With 5 Identity services, should sessions be shared or isolated?

**Can a session created by identity-idp be used to authenticate to identity-authz?**

A. Shared session store: All 5 Identity services share Redis/database session store (sessions work across services)
B. Service-isolated sessions: Each service has isolated session store (session only valid for originating service)
C. SSO pattern: identity-idp creates SSO session, other services validate via SSO token
D. Federated authentication: Each service authenticates independently, no session sharing
E. Other (please specify):

**Answer**: _____

---

### Q10.2: KMS ElasticKey Tenant Isolation Strategy

The spec mentions "per-tenant isolation" but doesn't define tenant boundaries.

**What is the tenant isolation boundary for KMS?**

A. Database-level: Each tenant has separate PostgreSQL database
B. Schema-level: Each tenant has separate PostgreSQL schema (tenant1.*, tenant2.*)
C. Table-level: Single schema, tenant_id column in all tables with RLS (Row-Level Security)
D. Instance-level: Each tenant gets dedicated KMS instance (separate containers)
E. Other (please specify):

**Answer**: _____

---

### Q10.3: CA Certificate Profile Customization

The spec lists "24 predefined certificate profiles" but doesn't clarify if custom profiles are allowed.

**Can administrators create custom certificate profiles beyond the 24 predefined ones?**

A. NO: Only 24 predefined profiles allowed, no customization (strict compliance)
B. YES: Admin API allows creating/updating/deleting custom profiles (full flexibility)
C. EXTEND ONLY: Predefined profiles can be extended with additional fields, not created from scratch
D. APPROVAL REQUIRED: Custom profiles require approval workflow (security review before activation)
E. Other (please specify):

**Answer**: _____

---

## Instructions for Completing This Quiz

1. **Read the context**: Each question references specific sections of constitution.md, spec.md, clarify.md, or copilot instructions.
2. **Select best answer**: Choose A-D based on your understanding of the project's intent.
3. **Use option E for gaps**: If none of A-D fit, describe the correct answer in option E.
4. **Answer follow-up questions**: If your answer triggers a follow-up, answer that as well.
5. **Return completed quiz**: Save answers in CLARIFY-QUIZME-ANSWERS.md or directly update this file.

**Next Steps After Completion**:

1. LLM agent will parse your answers
2. Answers will be integrated into clarify.md (canonical Q&A document)
3. Constitution and spec will be updated based on clarifications
4. CLARIFY-QUIZME.md will be archived as CLARIFY-QUIZME-2025-12-19.md

---

**Quiz Version**: 1.0.0
**Created**: 2025-12-19
**Covers**: constitution.md v3.0.0, spec-incomplete.md v1.2.0, copilot-instructions.md, 01-01.architecture.instructions.md
