# ARCHITECTURE.md - Agent Lookup Reference

**Purpose**: Efficient LLM agent reference to semantic topics in ARCHITECTURE.md using line number ranges.

**Last Updated**: 2026-02-26
**Source**: docs/ARCHITECTURE.md (4219 lines)

---

## How to Use

1. Find the topic you need in the **Semantic Topic Index** below
2. Note the line number range (e.g., Lines 99-209)
3. Read the relevant section of ARCHITECTURE.md using that range
4. Cross-reference related sections via the **Quick Reference by Theme** at the bottom

**Important**: Line numbers may drift as ARCHITECTURE.md evolves. If a section header doesn't match, search for the section title instead.

---

## Semantic Topic Index

### 1. Executive Summary (Lines 99-209)

**Topics**: Vision, cryptographic standards, API architecture, security features, core principles, success metrics

**Subsections**:
- 1.1 Vision Statement (101-111): FIPS 140-3 crypto, PKI, JOSE, identity
- 1.2 Key Architectural Characteristics (112-148): Crypto standards, API design, security features, observability, production readiness
- 1.3 Core Principles (149-178): Security-first, maximum quality, microservices, developer experience
- 1.4 Success Metrics (179-209): Code quality, performance, security, operational excellence

---

### 2. Strategic Vision & Principles (Lines 211-406)

**Topics**: Agent orchestration, architecture strategy, design strategy, implementation strategy, quality strategy

**Subsections**:
- 2.1 Agent Orchestration Strategy (213-259): Agent architecture, catalog, handoff flow, instruction files
- 2.2 Architecture Strategy (260-289): Monorepo, service template, evolutionary design
- 2.3 Design Strategy (290-334): Core principles, autonomous execution principles
- 2.4 Implementation Strategy (335-373): Spec-Driven Development, domain-driven design
- 2.5 Quality Strategy (374-406): Continuous quality, maximum quality mandate

---

### 3. Product Suite Architecture (Lines 408-687)

**Topics**: Product overview, service catalog (PKI, JOSE, SM, Identity), product-service relationships, port assignments

**Subsections**:
- 3.1 Product Overview (410-441): 5 products, 10 services
- 3.2 Service Catalog (442-597): PKI-CA, JOSE-JA, SM-IM, SM-KMS, Identity (Authz, IdP, RS, RP, SPA)
- 3.3 Product-Service Relationships (598-620): 1-to-1, 1-to-N, N-to-N patterns
- 3.4 Port Assignments & Networking (621-687): Port design, PostgreSQL ports, telemetry ports

---

### 4. System Architecture (Lines 689-953)

**Topics**: System context, container architecture, layered architecture, dependency injection, Go project structure, CLI patterns

**Subsections**:
- 4.1 System Context (691-700): External/internal actors, service boundaries
- 4.2 Container Architecture (701-717): Docker, Kubernetes, service communication
- 4.3 Component Architecture (718-731): Layered architecture, dependency injection
- 4.4 Code Organization (732-953): Go project structure, directory rules, CLI entry points, service implementations, shared utilities, Docker Compose, CLI patterns

---

### 5. Service Architecture (Lines 955-1121)

**Topics**: Service template, builder pattern, dual HTTPS endpoints, dual API paths, health checks

**Subsections**:
- 5.1 Service Template Pattern (957-983): Template components, benefits, mandatory usage
- 5.2 Service Builder Pattern (984-1039): Builder methods, merged migrations, ServiceResources, database compatibility
- 5.3 Dual HTTPS Endpoint Pattern (1040-1061): Public (0.0.0.0:8080), Private (127.0.0.1:9090), binding defaults
- 5.4 Dual API Path Pattern (1062-1087): /service/** (headless), /browser/** (browser), mutual exclusivity
- 5.5 Health Check Patterns (1088-1121): /livez, /readyz, /shutdown, Kubernetes standard

---

### 6. Security Architecture (Lines 1123-1447)

**Topics**: FIPS 140-3, SDLC security, product security, cryptography, PKI, JOSE, KMS, MFA, authn/authz, secrets detection

**Subsections**:
- 6.1 FIPS 140-3 Compliance Strategy (1125-1133): Approved/banned algorithms, ALWAYS enabled
- 6.2 SDLC Security Strategy (1134-1153): Shift-left, security gates, SAST/DAST
- 6.3 Product Security Strategy (1154-1178): Network security, web security headers, audit logging
- 6.4 Cryptographic Architecture (1179-1272): Key hierarchy, hash service, FIPS compliance, algorithm agility
- 6.5 PKI Architecture & Strategy (1273-1295): CA/BF Baseline Requirements, EST, SCEP, OCSP, CRL
- 6.6 JOSE Architecture & Strategy (1296-1317): JWK, JWS, JWE, JWT, elastic key ring
- 6.7 Key Management System Architecture (1318-1339): Hierarchical key barriers, unseal key interoperability
- 6.8 Multi-Factor Authentication Strategy (1340-1362): TOTP, HOTP, WebAuthn, Passkeys, Push
- 6.9 Authentication & Authorization (1363-1432): 13 headless methods, 28 browser methods, zero-trust, MFA step-up
- 6.10 Secrets Detection Strategy (1433-1447): Length-based threshold, inline secret detection, safe references

---

### 7. Data Architecture (Lines 1449-1590)

**Topics**: Database schema, multi-tenancy, dual database strategy, migrations, data security

**Subsections**:
- 7.1 Database Schema Patterns (1451-1496): GORM patterns, UUID fields, nullable UUIDs, JSON arrays, database isolation
- 7.2 Multi-Tenancy Architecture & Strategy (1497-1547): Schema-level isolation, authentication realms, realm principals
- 7.3 Dual Database Strategy (1548-1560): PostgreSQL (production) + SQLite (development/testing)
- 7.4 Migration Strategy (1561-1577): golang-migrate, embedded files, merged migrations
- 7.5 Data Security & Encryption (1578-1590): Encryption at rest, key hierarchy integration

---

### 8. API Architecture (Lines 1592-1720)

**Topics**: OpenAPI-first, REST conventions, API versioning, error handling, API security

**Subsections**:
- 8.1 OpenAPI-First Design (1594-1618): OpenAPI 3.0.3, strict-server, oapi-codegen
- 8.2 REST Conventions (1619-1649): Resource naming, HTTP methods, idempotency, pagination
- 8.3 API Versioning (1650-1663): N-1 backward compatibility, deprecation policy
- 8.4 Error Handling (1664-1699): Standard error schema, HTTP status codes, request IDs
- 8.5 API Security (1700-1720): IP allowlisting, rate limiting, CORS, CSRF, CSP

---

### 9. Infrastructure Architecture (Lines 1722-2045)

**Topics**: CLI patterns, configuration, observability, telemetry, containers, orchestration, CI/CD, reusable actions, pre-commit hooks

**Subsections**:
- 9.1 CLI Patterns & Strategy (1724-1743): Product-service pattern, suite pattern, subcommands
- 9.2 Configuration Architecture & Strategy (1744-1811): Docker secrets > YAML > CLI, NO env vars
- 9.3 Observability Architecture (1812-1834): Telemetry flow, sidecar pattern
- 9.4 Telemetry Strategy (1835-1901): Structured logging, Prometheus metrics, OpenTelemetry, OTel collector constraints
- 9.5 Container Architecture (1902-1936): Multi-stage Dockerfile, Docker secrets, healthchecks
- 9.6 Orchestration Patterns (1937-1959): Docker Compose, Kubernetes, service discovery
- 9.7 CI/CD Workflow Architecture (1960-2007): GitHub Actions, workflow matrix, reusable workflows
- 9.8 Reusable Action Patterns (2008-2028): Docker image pre-pull, shared action patterns
- 9.9 Pre-Commit Hook Architecture (2029-2045): golangci-lint, gofumpt, goimports, UTF-8 BOM

---

### 10. Testing Architecture (Lines 2047-2569)

**Topics**: Testing strategy, unit/integration/E2E/mutation/load/fuzz/bench/race/SAST/DAST/workflow testing

**Subsections**:
- 10.1 Testing Strategy Overview (2049-2073): Test pyramid, coverage targets, timing, file organization
- 10.2 Unit Testing Strategy (2074-2220): Table-driven, app.Test(), TestMain, t.Parallel(), forbidden patterns, coverage ceiling analysis, test seam injection
- 10.3 Integration Testing Strategy (2221-2284): PostgreSQL test containers, SQLite in-memory, database testing
- 10.4 E2E Testing Strategy (2285-2464): Docker Compose, ComposeManager, service-under-test patterns, E2E test organization
- 10.5 Mutation Testing Strategy (2465-2489): gremlins, >=95%/98% targets, exempt patterns
- 10.6 Load Testing Strategy (2490-2501): Gatling, Java 21, performance baselines
- 10.7 Fuzz Testing Strategy (2502-2510): _fuzz_test.go, 15s minimum fuzz time
- 10.8 Benchmark Testing Strategy (2511-2527): _bench_test.go, crypto operation benchmarks
- 10.9 Race Detection Strategy (2528-2535): -race flag, CGO_ENABLED=1 requirement
- 10.10 SAST Strategy (2536-2547): gosec, static analysis, secret detection
- 10.11 DAST Strategy (2548-2562): Nuclei scanning, dynamic security testing
- 10.12 Workflow Testing Strategy (2563-2569): Workflow test patterns, CI/CD testing

---

### 11. Quality Architecture (Lines 2571-2814)

**Topics**: Maximum quality strategy, quality gates, code quality standards, documentation standards, review processes

**Subsections**:
- 11.1 Maximum Quality Strategy (2573-2631): Quality attributes, magic values, import aliases, CGO ban, file size limits, conditional patterns, format_go protection
- 11.2 Quality Gates (2632-2745): Pre-commit gates, coverage targets, mutation testing, file size limits, conditional patterns, baseline restore
- 11.3 Code Quality Standards (2746-2775): golangci-lint v2, linter configuration, code quality enforcement
- 11.4 Documentation Standards (2776-2796): Code comments, API docs, architecture docs
- 11.5 Review Processes (2797-2814): PR format, review checklist, evidence-based approval

---

### 12. Deployment Architecture (Lines 2816-3779)

**Topics**: CI/CD automation, build pipeline, deployment patterns, multi-level deployment hierarchy, deployment structure validation, config file architecture, secrets management, documentation propagation, environment strategy, release management

**Subsections**:
- 12.1 CI/CD Automation Strategy (2818-2823): GitHub Actions, automated quality gates
- 12.2 Build Pipeline (2824-2847): Build, test, coverage, mutation, SAST, DAST
- 12.3 Deployment Patterns (2848-3244): Docker Compose, Docker secrets, health checks, multi-stage Dockerfile, secrets coordination, multi-level hierarchy (SUITE/PRODUCT/SERVICE)
- 12.4 Deployment Structure Validation (3245-3531): 8 validators, naming, kebab-case, schema, template, ports, telemetry, admin, secrets
- 12.5 Config File Architecture (3532-3589): Flat kebab-case YAML, service template configs, domain configs, environment configs
- 12.6 Secrets Management in Deployments (3590-3601): Docker secrets enforcement, inline secret detection
- 12.7 Documentation Propagation Strategy (3602-3754): @source/@propagate markers, CI/CD validation, instruction file sync
- 12.8 Validator Error Aggregation Pattern (3755-3766): Sequential execution, aggregated errors, single unified report
- 12.9 Environment Strategy (3767-3772): Development, staging, production environments
- 12.10 Release Management (3773-3779): Versioning, changelog, release process

**Companion Docs**:
- [ARCHITECTURE-COMPOSE-MULTIDEPLOY.md](/docs/ARCHITECTURE-COMPOSE-MULTIDEPLOY.md) - Comprehensive multi-level deployment hierarchy documentation

---

### 13. Development Practices (Lines 3781-3989)

**Topics**: Coding standards, version control, branching strategy, code review, development workflow, plan lifecycle, infrastructure blocker escalation

**Subsections**:
- 13.1 Coding Standards (3783-3787): Go 1.25.7, CGO ban, import aliases, file size limits
- 13.2 Version Control (3788-3827): Conventional commits, incremental commits, baseline restore pattern
- 13.3 Branching Strategy (3828-3833): Main branch, feature branches, release branches
- 13.4 Code Review (3834-3839): PR descriptions, review checklist, evidence-based approval
- 13.5 Development Workflow (3840-3933): Local dev, testing, linting, git flow, Docker Desktop startup
- 13.6 Plan Lifecycle Management (3934-3953): Plan creation, tracking, completion criteria
- 13.7 Infrastructure Blocker Escalation (3954-3989): Three-encounter rule, mandatory Phase 0 fix, infrastructure categories

---

### 14. Operational Excellence (Lines 3991-4024)

**Topics**: Monitoring, incident management, performance, capacity planning, disaster recovery

**Subsections**:
- 14.1 Monitoring & Alerting (3993-3999): Prometheus, Grafana, alerting rules
- 14.2 Incident Management (4000-4005): Runbooks, escalation, post-incident review
- 14.3 Performance Management (4006-4011): Benchmarking, profiling, optimization
- 14.4 Capacity Planning (4012-4017): Resource estimation, scaling strategy
- 14.5 Disaster Recovery (4018-4024): Backup strategy, RTO/RPO, failover procedures

---

### Appendices (Lines 4026-4219)

**Topics**: Decision records, reference tables, compliance matrix

**Subsections**:
- Appendix A: Decision Records (4026-4055): ADRs, technology selection, pattern selection
- Appendix B: Reference Tables (4057-4166): Service ports, DB ports, tech stack, dependencies, config reference, instruction files, agents, CI/CD workflows, reusable actions, linter rules
- Appendix C: Compliance Matrix (4168-4202): FIPS 140-3, PKI standards, OAuth 2.1/OIDC 1.0, security standards

---

## Quick Reference by Theme

### Security Topics
- FIPS 140-3 Compliance: Section 6.1 (Lines 1125-1133)
- Cryptographic Architecture: Section 6.4 (Lines 1179-1272)
- PKI Architecture: Section 6.5 (Lines 1273-1295)
- JOSE Architecture: Section 6.6 (Lines 1296-1317)
- Key Management: Section 6.7 (Lines 1318-1339)
- Authentication & Authorization: Section 6.9 (Lines 1363-1432)
- Secrets Detection: Section 6.10 (Lines 1433-1447)
- Secrets Management: Section 12.6 (Lines 3590-3601)
- API Security: Section 8.5 (Lines 1700-1720)

### Testing Topics
- Testing Strategy Overview: Section 10.1 (Lines 2049-2073)
- Unit Testing: Section 10.2 (Lines 2074-2220)
- Integration Testing: Section 10.3 (Lines 2221-2284)
- E2E Testing: Section 10.4 (Lines 2285-2464)
- Mutation Testing: Section 10.5 (Lines 2465-2489)
- Fuzz Testing: Section 10.7 (Lines 2502-2510)
- Race Detection: Section 10.9 (Lines 2528-2535)
- Workflow Testing: Section 10.12 (Lines 2563-2569)

### Architecture Topics
- Service Template Pattern: Section 5.1 (Lines 957-983)
- Builder Pattern: Section 5.2 (Lines 984-1039)
- Dual HTTPS Endpoint: Section 5.3 (Lines 1040-1061)
- Dual API Path: Section 5.4 (Lines 1062-1087)
- Port Assignments: Section 3.4 (Lines 621-687)
- Multi-Level Deployment: Section 12.3 (Lines 2848-3244)

### Configuration Topics
- Configuration Architecture: Section 9.2 (Lines 1744-1811)
- Config File Architecture: Section 12.5 (Lines 3532-3589)
- Docker Compose: Section 12.3 (Lines 2848-3244)
- CLI Patterns: Section 9.1 (Lines 1724-1743)

### Quality Topics
- Maximum Quality Strategy: Section 11.1 (Lines 2573-2631)
- Quality Gates: Section 11.2 (Lines 2632-2745)
- Code Quality Standards: Section 11.3 (Lines 2746-2775)
- Coverage Targets: Section 2.5 (Lines 374-406)
- Documentation Propagation: Section 12.7 (Lines 3602-3754)
- Infrastructure Blocker Escalation: Section 13.7 (Lines 3954-3989)
