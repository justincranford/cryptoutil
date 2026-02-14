# ARCHITECTURE.md - Agent Lookup Reference

**Purpose**: Efficient LLM agent reference to semantic topics in ARCHITECTURE.md using line number ranges.

**Last Updated**: 2026-02-14
**Source**: docs/ARCHITECTURE.md (2790 lines)

---

## How to Use

Agents can retrieve specific sections using line number ranges:
```powershell
Get-Content docs\ARCHITECTURE.md | Select-Object -Skip 66 -First 113
```

**Pattern**: `-Skip (StartLine-1) -First (EndLine-StartLine+1)`

---

## Semantic Topic Index

### 1. Executive Summary (Lines 67-179)

**Topics**: Vision, cryptographic standards, API architecture, security features, core principles, success metrics

**Subsections**:
- 1.1 Vision Statement (69-80)
- 1.2 Key Architectural Characteristics (81-117)
- 1.3 Core Principles (118-147)
- 1.4 Success Metrics (148-179)

---

### 2. Strategic Vision & Principles (Lines 180-365)

**Topics**: Agent orchestration, architecture strategy, design strategy, implementation strategy, quality strategy

**Subsections**:
- 2.1 Agent Orchestration Strategy (182-219): Agent architecture, catalog, handoff flow, instruction files
- 2.2 Architecture Strategy (220-249): Monorepo, service template, evolutionary design
- 2.3 Design Strategy (250-294): Core principles, autonomous execution principles
- 2.4 Implementation Strategy (295-333): Spec-Driven Development, domain-driven design
- 2.5 Quality Strategy (334-365): Continuous quality, maximum quality mandate

---

### 3. Product Suite Architecture (Lines 366-563)

**Topics**: Product overview, service catalog (PKI, JOSE, Cipher, SM, Identity), product-service relationships, port assignments

**Subsections**:
- 3.1 Product Overview (368-406): 5 products, 9 services
- 3.2 Service Catalog (407-502): PKI-CA, JOSE-JA, Cipher-IM, SM-KMS, Identity (Authz, IdP, RS, RP, SPA)
- 3.3 Product-Service Relationships (503-525): 1-to-1, 1-to-N, N-to-N patterns
- 3.4 Port Assignments & Networking (526-563): Port design, PostgreSQL ports, telemetry ports

---

### 4. System Architecture (Lines 564-828)

**Topics**: System context, container architecture, layered architecture, dependency injection, Go project structure, CLI patterns

**Subsections**:
- 4.1 System Context (566-574): External/internal actors, service boundaries
- 4.2 Container Architecture (575-591): Docker, Kubernetes, service communication
- 4.3 Component Architecture (592-605): Layered architecture, dependency injection
- 4.4 Code Organization (606-828): Go project structure, directory rules, CLI entry points, service implementations, shared utilities, Docker Compose, CLI patterns

---

### 5. Service Architecture (Lines 829-993)

**Topics**: Service template, builder pattern, dual HTTPS endpoints, dual API paths, health checks

**Subsections**:
- 5.1 Service Template Pattern (831-854): Template components, benefits, mandatory usage
- 5.2 Service Builder Pattern (855-910): Builder methods, merged migrations, ServiceResources, database compatibility
- 5.3 Dual HTTPS Endpoint Pattern (911-932): Public (0.0.0.0:8080), Private (127.0.0.1:9090), binding defaults
- 5.4 Dual API Path Pattern (933-958): /service/**(headless), /browser/** (browser), mutual exclusivity
- 5.5 Health Check Patterns (959-993): /livez, /readyz, /shutdown, Kubernetes standard

---

### 6. Security Architecture (Lines 994-1272)

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

### 7. Data Architecture (Lines 1273-1388)

**Topics**: Database schema, multi-tenancy, dual database strategy, migrations, data security

**Subsections**:
- 7.1 Database Schema Patterns (1275-1293): GORM tags, cross-DB compatibility, UUIDs, JSON arrays
- 7.2 Multi-Tenancy Architecture & Strategy (1294-1344): Schema-level isolation, tenant_id vs realm_id
- 7.3 Dual Database Strategy (1345-1357): PostgreSQL (distributed), SQLite (single-node), connection pooling
- 7.4 Migration Strategy (1358-1374): golang-migrate, embedded FS, merged migrations (template 1001-1999, domain 2001+)
- 7.5 Data Security & Encryption (1375-1388): Encryption-at-rest, barrier service, transparent data encryption

---

### 8. API Architecture (Lines 1389-1502)

**Topics**: OpenAPI-first, REST conventions, API versioning, error handling, API security

**Subsections**:
- 8.1 OpenAPI-First Design (1391-1415): OpenAPI 3.0.3, strict-server, oapi-codegen
- 8.2 REST Conventions (1416-1446): Resource naming, HTTP methods, idempotency, pagination
- 8.3 API Versioning (1447-1460): N-1 backward compatibility, deprecation policy
- 8.4 Error Handling (1461-1480): Standard error schema, HTTP status codes, request IDs
- 8.5 API Security (1481-1502): IP allowlisting, rate limiting, CORS, CSRF, CSP

---

### 9. Infrastructure Architecture (Lines 1503-1783)

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

### 10. Testing Architecture (Lines 1784-2131)

**Topics**: Testing strategy, unit/integration/E2E/mutation/load/fuzz/benchmark/race/SAST/DAST/workflow testing

**Subsections**:
- 10.1 Testing Strategy Overview (1786-1800): Pyramid, coverage targets, mutation targets
- 10.2 Unit Testing Strategy (1801-1903): Table-driven, t.Parallel(), Fiber app.Test(), TestMain
- 10.3 Integration Testing Strategy (1904-1967): Test containers, database, real dependencies
- 10.4 E2E Testing Strategy (1968-2026): Docker Compose, health checks, dual paths
- 10.5 Mutation Testing Strategy (2027-2050): gremlins, 95% production, 98% infrastructure
- 10.6 Load Testing Strategy (2051-2062): Gatling, concurrent users, response times
- 10.7 Fuzz Testing Strategy (2063-2071): go test -fuzz, 15s minimum
- 10.8 Benchmark Testing Strategy (2072-2088): crypto operations, performance baselines
- 10.9 Race Detection Strategy (2089-2096): go test -race, concurrency safety
- 10.10 SAST Strategy (2097-2108): gosec, golangci-lint, pre-commit hooks
- 10.11 DAST Strategy (2109-2123): Nuclei scanning, E2E environment
- 10.12 Workflow Testing Strategy (2124-2131): GitHub Actions, matrix testing

---

### 11. Quality Architecture (Lines 2132-2357)

**Topics**: Maximum quality strategy, quality gates, code quality standards, documentation standards, review processes

**Subsections**:
- 11.1 Maximum Quality Strategy - MANDATORY (2134-2187): ALL issues are blockers, NO exceptions
- 11.2 Quality Gates (2188-2287): Per-action, per-phase, overall project quality gates
- 11.3 Code Quality Standards (2288-2317): File size limits, linting, complexity, maintainability
- 11.4 Documentation Standards (2318-2338): README, architecture, instructions, inline comments
- 11.5 Review Processes (2339-2357): Pre-commit hooks, PR reviews, evidence-based validation

---

### 12. Deployment Architecture (Lines 2358-2452)

**Topics**: CI/CD automation, build pipeline, deployment patterns, environment strategy, release management

**Subsections**:
- 12.1 CI/CD Automation Strategy (2360-2365): GitHub Actions, automated quality gates
- 12.2 Build Pipeline (2366-2389): Build, test, coverage, mutation, SAST, DAST
- 12.3 Deployment Patterns (2390-2438): Docker Compose, Docker secrets, health checks, multi-stage Dockerfile
- 12.4 Environment Strategy (2439-2444): Dev, CI/CD, Docker, production
- 12.5 Release Management (2445-2452): Semantic versioning, changelog

---

### 13. Development Practices (Lines 2453-2520)

**Topics**: Coding standards, version control, branching strategy, code review, development workflow

**Subsections**:
- 13.1 Coding Standards (2455-2459): Go 1.25.5, CGO ban, import aliases, file size limits
- 13.2 Version Control (2460-2483): Conventional commits, incremental commits, pre-commit hooks
- 13.3 Branching Strategy (2484-2489): Main branch, feature branches, release branches
- 13.4 Code Review (2490-2495): PR descriptions, review checklist, evidence-based approval
- 13.5 Development Workflow (2496-2520): Local dev, testing, linting, git flow

---

### 14. Operational Excellence (Lines 2521-2554)

**Topics**: Monitoring, incident management, performance, capacity planning, disaster recovery

**Subsections**:
- 14.1 Monitoring & Alerting (2523-2529): Prometheus, Grafana, OTLP
- 14.2 Incident Management (2530-2535): Incident response, post-mortems
- 14.3 Performance Management (2536-2541): Benchmarking, optimization
- 14.4 Capacity Planning (2542-2547): Resource scaling, load testing
- 14.5 Disaster Recovery (2548-2554): Backup, restore, failover

---

### Appendices (Lines 2555-2772)

**Topics**: Decision records, reference tables, compliance matrix

**Subsections**:
- Appendix A: Decision Records (2557-2639): ADRs, architectural decisions
- Appendix B: Reference Tables (2641-2728): Service catalog, port assignments, file size limits, coverage targets
- Appendix C: Compliance Matrix (2730-2772): FIPS 140-3, CA/BF Baseline, OAuth 2.1, OIDC 1.0, WebAuthn, NIST SP 800-63B

---

## Quick Reference by Theme

### Security Topics

- FIPS 140-3: Lines 996-1004, 1052-1064
- Cryptographic Architecture: Lines 1050-1143
- PKI: Lines 1144-1166
- JOSE: Lines 1167-1188
- KMS: Lines 1189-1210
- MFA: Lines 1211-1233
- Auth/Authz: Lines 1234-1272
- Data Security: Lines 1375-1388

### Testing Topics

- Testing Strategy: Lines 1786-2131
- Unit Testing: Lines 1801-1903
- Integration Testing: Lines 1904-1967
- E2E Testing: Lines 1968-2026
- Mutation Testing: Lines 2027-2050
- Quality Gates: Lines 2188-2287

### Architecture Topics

- Service Template: Lines 831-854
- Service Builder: Lines 855-910
- Dual HTTPS: Lines 911-932
- Dual API Paths: Lines 933-958
- Health Checks: Lines 959-993

### Configuration Topics

- Configuration Strategy: Lines 1525-1592
- Docker Secrets: Lines 1542-1561
- CLI Patterns: Lines 1505-1524

### Quality Topics

- Maximum Quality Strategy: Lines 2134-2187
- Quality Gates: Lines 2188-2287
- Code Quality: Lines 2288-2317
- Documentation: Lines 2318-2338
