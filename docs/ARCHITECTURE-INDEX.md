# ARCHITECTURE.md - Agent Lookup Reference

**Purpose**: Efficient LLM agent reference to semantic topics in ARCHITECTURE.md using line number ranges.

**Last Updated**: 2026-02-16
**Source**: docs/ARCHITECTURE.md (3356 lines)

---

## How to Use

Agents can retrieve specific sections using line number ranges:
```powershell
Get-Content docs\ARCHITECTURE.md | Select-Object -Skip 66 -First 113
```

**Pattern**: `-Skip (StartLine-1) -First (EndLine-StartLine+1)`

---

## Semantic Topic Index

### 1. Executive Summary (Lines 69-180)

**Topics**: Vision, cryptographic standards, API architecture, security features, core principles, success metrics

**Subsections**:
- 1.1 Vision Statement (69-80)
- 1.2 Key Architectural Characteristics (81-117)
- 1.3 Core Principles (118-147)
- 1.4 Success Metrics (148-179)

---

### 2. Strategic Vision & Principles (Lines 182-366)

**Topics**: Agent orchestration, architecture strategy, design strategy, implementation strategy, quality strategy

**Subsections**:
- 2.1 Agent Orchestration Strategy (182-219): Agent architecture, catalog, handoff flow, instruction files
- 2.2 Architecture Strategy (220-249): Monorepo, service template, evolutionary design
- 2.3 Design Strategy (250-294): Core principles, autonomous execution principles
- 2.4 Implementation Strategy (295-333): Spec-Driven Development, domain-driven design
- 2.5 Quality Strategy (334-365): Continuous quality, maximum quality mandate

---

### 3. Product Suite Architecture (Lines 368-666)

**Topics**: Product overview, service catalog (PKI, JOSE, Cipher, SM, Identity), product-service relationships, port assignments

**Subsections**:
- 3.1 Product Overview (368-406): 5 products, 9 services
- 3.2 Service Catalog (407-502): PKI-CA, JOSE-JA, Cipher-IM, SM-KMS, Identity (Authz, IdP, RS, RP, SPA)
- 3.3 Product-Service Relationships (503-525): 1-to-1, 1-to-N, N-to-N patterns
- 3.4 Port Assignments & Networking (526-563): Port design, PostgreSQL ports, telemetry ports

---

### 4. System Architecture (Lines 668-932)

**Topics**: System context, container architecture, layered architecture, dependency injection, Go project structure, CLI patterns

**Subsections**:
- 4.1 System Context (566-574): External/internal actors, service boundaries
- 4.2 Container Architecture (575-591): Docker, Kubernetes, service communication
- 4.3 Component Architecture (592-605): Layered architecture, dependency injection
- 4.4 Code Organization (606-828): Go project structure, directory rules, CLI entry points, service implementations, shared utilities, Docker Compose, CLI patterns

---

### 5. Service Architecture (Lines 934-1097)

**Topics**: Service template, builder pattern, dual HTTPS endpoints, dual API paths, health checks

**Subsections**:
- 5.1 Service Template Pattern (831-854): Template components, benefits, mandatory usage
- 5.2 Service Builder Pattern (855-910): Builder methods, merged migrations, ServiceResources, database compatibility
- 5.3 Dual HTTPS Endpoint Pattern (911-932): Public (0.0.0.0:8080), Private (127.0.0.1:9090), binding defaults
- 5.4 Dual API Path Pattern (933-958): /service/**(headless), /browser/** (browser), mutual exclusivity
- 5.5 Health Check Patterns (959-993): /livez, /readyz, /shutdown, Kubernetes standard

---

### 6. Security Architecture (Lines 1099-1388)

**Topics**: FIPS 140-3 compliance, SDLC security, product security, cryptographic architecture, PKI, JOSE, KMS, MFA, auth/authz

**Subsections**:
- 6.1 FIPS 140-3 Compliance Strategy (996-1004): Approved algorithms, no bcrypt/scrypt/Argon2
- 6.2 SDLC Security Strategy (1005-1024): Secret scanning, SAST, DAST, dependency scanning
- 6.3 Product Security Strategy (1025-1049): TLS 1.3+, IP allowlisting, rate limiting, CSRF, CSP, CORS
- 6.4 Cryptographic Architecture (1050-1143): FIPS algorithms, key hierarchy (barrier), hash service, unseal modes, key rotation
- 6.5 PKI Architecture & Strategy (1144-1166): CA/BF Baseline Requirements, EST, SCEP, OCSP, CRL
- 6.6 JOSE Architecture & Strategy (1167-1188): JWK, JWS, JWE, JWT, elastic key ring
- 6.7 Key Management System Architecture (1189-1210): Hierarchical key barriers, unseal key interoperability
- 6.8 Multi-Factor Authentication Strategy (1211-1233): TOTP, HOTP, WebAuthn, Passkeys, Push
- 6.9 Authentication & Authorization (1234-1272): 13 headless methods, 28 browser methods, zero-trust, MFA step-up

---

### 7. Data Architecture (Lines 1390-1531)

**Topics**: Database schema, multi-tenancy, dual database strategy, migrations, data security

**Subsections**:
- 7.1 Database Schema Patterns (1275-1293): GORM tags, cross-DB compatibility, UUIDs, JSON arrays
- 7.2 Multi-Tenancy Architecture & Strategy (1294-1344): Schema-level isolation, tenant_id vs realm_id
- 7.3 Dual Database Strategy (1345-1357): PostgreSQL (distributed), SQLite (single-node), connection pooling
- 7.4 Migration Strategy (1358-1374): golang-migrate, embedded FS, merged migrations (template 1001-1999, domain 2001+)
- 7.5 Data Security & Encryption (1375-1388): Encryption-at-rest, barrier service, transparent data encryption

---

### 8. API Architecture (Lines 1533-1645)

**Topics**: OpenAPI-first, REST conventions, API versioning, error handling, API security

**Subsections**:
- 8.1 OpenAPI-First Design (1391-1415): OpenAPI 3.0.3, strict-server, oapi-codegen
- 8.2 REST Conventions (1416-1446): Resource naming, HTTP methods, idempotency, pagination
- 8.3 API Versioning (1447-1460): N-1 backward compatibility, deprecation policy
- 8.4 Error Handling (1461-1480): Standard error schema, HTTP status codes, request IDs
- 8.5 API Security (1481-1502): IP allowlisting, rate limiting, CORS, CSRF, CSP

---

### 9. Infrastructure Architecture (Lines 1647-1929)

**Topics**: CLI patterns, configuration, observability, telemetry, containers, orchestration, CI/CD, pre-commit hooks

**Subsections**:
- 9.1 CLI Patterns & Strategy (1505-1524): Product-service pattern, suite pattern, subcommands
- 9.2 Configuration Architecture & Strategy (1525-1592): Docker secrets > YAML > CLI, NO env vars
- 9.3 Observability Architecture (OTLP) (1593-1615): Telemetry flow, sidecar pattern
- 9.4 Telemetry Strategy (1616-1641): Structured logging, Prometheus metrics, OpenTelemetry
- 9.5 Container Architecture (1642-1676): Multi-stage Dockerfile, Docker secrets, healthchecks
- 9.6 Orchestration Patterns (1677-1699): Docker Compose, Kubernetes, service discovery
- 9.7 CI/CD Workflow Architecture (1700-1744): Workflow matrix (CI, security, integration), path filters
- 9.8 Reusable Action Patterns (1745-1765): Docker pre-pull, test execution, artifact management
- 9.9 Pre-Commit Hook Architecture (1766-1783): golangci-lint, formatters, security checks

---

### 10. Testing Architecture (Lines 1931-2277)

**Topics**: Testing strategy, unit/integration/E2E/mutation/load/fuzz/benchmark/race/SAST/DAST/workflow testing

**Subsections**:
- 10.1 Testing Strategy Overview (1933-1946): Pyramid, coverage targets, mutation targets
- 10.2 Unit Testing Strategy (1948-2049): Table-driven, t.Parallel(), Fiber app.Test(), TestMain
- 10.3 Integration Testing Strategy (2051-2113): Test containers, database, real dependencies
- 10.4 E2E Testing Strategy (2115-2172): Docker Compose, health checks, dual paths
- 10.5 Mutation Testing Strategy (2174-2196): gremlins, 95% production, 98% infrastructure
- 10.6 Load Testing Strategy (2198-2208): Gatling, concurrent users, response times
- 10.7 Fuzz Testing Strategy (2210-2217): go test -fuzz, 15s minimum
- 10.8 Benchmark Testing Strategy (2219-2234): crypto operations, performance baselines
- 10.9 Race Detection Strategy (2236-2242): go test -race, concurrency safety
- 10.10 SAST Strategy (2244-2254): gosec, golangci-lint, pre-commit hooks
- 10.11 DAST Strategy (2256-2269): Nuclei scanning, E2E environment
- 10.12 Workflow Testing Strategy (2271-2277): GitHub Actions, matrix testing

---

### 11. Quality Architecture (Lines 2279-2508)

**Topics**: Maximum quality strategy, quality gates, code quality standards, documentation standards, review processes

**Subsections**:
- 11.1 Maximum Quality Strategy - MANDATORY (2281-2302): ALL issues are blockers, NO exceptions
- 11.2 Quality Gates (2335-2440): Per-action, per-phase, overall project quality gates
- 11.3 Code Quality Standards (2440-2468): File size limits, linting, complexity, maintainability
- 11.4 Documentation Standards (2470-2491): README, architecture, instructions, inline comments
- 11.5 Review Processes (2491-2508): Pre-commit hooks, PR reviews, evidence-based validation

---

### 12. Deployment Architecture (Lines 2510-3300)

**Topics**: CI/CD automation, build pipeline, deployment patterns, multi-level deployment hierarchy, deployment structure validation, config file architecture, secrets management, environment strategy, release management

**Subsections**:
- 12.1 CI/CD Automation Strategy (2512-2516): GitHub Actions, automated quality gates
- 12.2 Build Pipeline (2518-2540): Build, test, coverage, mutation, SAST, DAST
- 12.3 Deployment Patterns (2542-2905): Docker Compose, Docker secrets, health checks, multi-stage Dockerfile, secrets coordination, multi-level hierarchy
  - 12.3.4 Multi-Level Deployment Hierarchy (2732-2905): SUITE/PRODUCT/SERVICE tiers, layered pepper strategy, port offset strategy, linter validation
- 12.4 Deployment Structure Validation (2907-3192): Automated validation, deployment types (SUITE/PRODUCT/PRODUCT-SERVICE), validation rules, CI/CD integration
- 12.5 Config File Architecture (3194-3246): Service template configs, domain configs, environment configs
- 12.6 Secrets Management in Deployments (3248-3258): Docker secrets enforcement
- 12.7 Documentation Propagation Strategy (3260-3276): Chunk-based propagation mapping
- 12.8 Validator Error Aggregation Pattern (3278-3286): Sequential execution, aggregated errors
- 12.9 Environment Strategy (3288-3292): Dev, CI/CD, Docker, production
- 12.10 Release Management (3294-3300): Semantic versioning, changelog

**Cross-References**:
- [ARCHITECTURE-COMPOSE-MULTIDEPLOY.md](/docs/ARCHITECTURE-COMPOSE-MULTIDEPLOY.md) - Comprehensive multi-level deployment hierarchy documentation

---

### 13. Development Practices (Lines 3302-3431)

**Topics**: Coding standards, version control, branching strategy, code review, development workflow

**Subsections**:
- 13.1 Coding Standards (3304-3307): Go 1.25.5, CGO ban, import aliases, file size limits
- 13.2 Version Control (3309-3331): Conventional commits, incremental commits, pre-commit hooks
- 13.3 Branching Strategy (3333-3337): Main branch, feature branches, release branches
- 13.4 Code Review (3339-3343): PR descriptions, review checklist, evidence-based approval
- 13.5 Development Workflow (3345-3431): Local dev, testing, linting, git flow, Docker Desktop startup

---

### 14. Operational Excellence (Lines 3433-3466)

**Topics**: Monitoring, incident management, performance, capacity planning, disaster recovery

**Subsections**:
- 14.1 Monitoring & Alerting (3435-3440): Prometheus, Grafana, OTLP
- 14.2 Incident Management (3442-3446): Incident response, post-mortems
- 14.3 Performance Management (3448-3452): Benchmarking, optimization
- 14.4 Capacity Planning (3454-3458): Resource scaling, load testing
- 14.5 Disaster Recovery (3460-3466): Backup, restore, failover

---

### Appendices (Lines 3468-3683)

**Topics**: Decision records, reference tables, compliance matrix

**Subsections**:
- Appendix A: Decision Records (3468-3497): ADRs, architectural decisions
- Appendix B: Reference Tables (3499-3631): Service catalog, port assignments, file size limits, coverage targets
- Appendix C: Compliance Matrix (3633-3667): FIPS 140-3, CA/BF Baseline, OAuth 2.1, OIDC 1.0, WebAuthn, NIST SP 800-63B

---

## Quick Reference by Theme

### Security Topics

- FIPS 140-3: Lines 1101-1108, 1157-1168
- Cryptographic Architecture: Lines 1155-1248
- PKI: Lines 1249-1292
- JOSE: Lines 1272-1315
- KMS: Lines 1294-1338
- MFA: Lines 1316-1374
- Auth/Authz: Lines 1339-1388
- Data Security: Lines 1519-1531

### Testing Topics

- Testing Strategy: Lines 1931-2277
- Unit Testing: Lines 1948-2049
- Integration Testing: Lines 2051-2115
- E2E Testing: Lines 2117-2175
- Mutation Testing: Lines 2177-2200
- Quality Gates: Lines 2335-2440

### Architecture Topics

- Service Template: Lines 936-958
- Service Builder: Lines 960-1014
- Dual HTTPS: Lines 1016-1036
- Dual API Paths: Lines 1038-1062
- Health Checks: Lines 1064-1097

### Configuration Topics

- Configuration Strategy: Lines 1647-1740
- Docker Secrets: Lines 2591-2730
- CLI Patterns: Lines 1647-1680

### Quality Topics

- Maximum Quality Strategy: Lines 2281-2333
- Quality Gates: Lines 2335-2440
- Code Quality: Lines 2440-2468
- Documentation: Lines 2470-2508
