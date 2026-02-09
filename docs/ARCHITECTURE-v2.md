---
title: cryptoutil Architecture - Single Source of Truth
version: 2.0
date: 2026-02-08
status: Draft
audience:
  - Copilot Instructions
  - GitHub Copilot Agents
  - Prompts and Skills
  - Development Team
  - Technical Stakeholders
references:
  - .github/copilot-instructions.md
  - .github/instructions/*.instructions.md
  - .github/agents/*.agent.md
  - .github/workflows/*.yml
  - .specify/memory/constitution.md
maintainers:
  - cryptoutil Development Team
tags:
  - architecture
  - design
  - implementation
  - testing
  - security
  - compliance
---

# cryptoutil Architecture - Single Source of Truth

**Last Updated**: February 8, 2026
**Version**: 2.0
**Purpose**: Comprehensive architectural reference for the cryptoutil product suite, serving as the canonical source for all architectural decisions, patterns, strategies, and implementation guidelines.

---

## Document Organization

This document is structured to serve multiple audiences:
- **Copilot Instructions & Agents**: Machine-parseable sections with clear directives
- **Developers**: Detailed implementation patterns and examples
- **Architects**: High-level design decisions and trade-offs
- **Stakeholders**: Strategic context and rationale

### Navigation Guide

- [1. Executive Summary](#1-executive-summary)
- [2. Strategic Vision & Principles](#2-strategic-vision--principles)
- [3. Product Suite Architecture](#3-product-suite-architecture)
- [4. System Architecture](#4-system-architecture)
- [5. Service Architecture](#5-service-architecture)
- [6. Security Architecture](#6-security-architecture)
- [7. Data Architecture](#7-data-architecture)
- [8. API Architecture](#8-api-architecture)
- [9. Infrastructure Architecture](#9-infrastructure-architecture)
- [10. Testing Architecture](#10-testing-architecture)
- [11. Quality Architecture](#11-quality-architecture)
- [12. Deployment Architecture](#12-deployment-architecture)
- [13. Development Practices](#13-development-practices)
- [14. Operational Excellence](#14-operational-excellence)
- [Appendix A: Decision Records](#appendix-a-decision-records)
- [Appendix B: Reference Tables](#appendix-b-reference-tables)
- [Appendix C: Compliance Matrix](#appendix-c-compliance-matrix)

---

## 1. Executive Summary

### 1.1 Vision Statement

[To be populated with vision]

### 1.2 Key Architectural Characteristics

[To be populated with architectural characteristics]

### 1.3 Core Principles

[To be populated with principles]

### 1.4 Success Metrics

[To be populated with metrics]

---

## 2. Strategic Vision & Principles

### 2.1 Agent Orchestration Strategy

[To be populated]

#### 2.1.1 Agent Architecture

- Agent isolation principle (agents do NOT inherit copilot instructions)
- YAML frontmatter requirements (name, description, tools, handoffs)
- Autonomous execution mode patterns
- Quality over speed enforcement

#### 2.1.2 Agent Catalog

- plan-tasks-quizme: Planning and task decomposition
- plan-tasks-implement: Autonomous implementation execution
- doc-sync: Documentation synchronization
- fix-github-workflows: Workflow repair and validation
- fix-tool-names: Tool name consistency enforcement
- beast-mode-custom: Continuous execution mode

#### 2.1.3 Agent Handoff Flow

- Planning → Implementation → Documentation → Fix handoff chains
- Explicit handoff triggers and conditions
- State preservation across handoffs

#### 2.1.4 Instruction File Organization

- Hierarchical numbering scheme (01-01 through 07-01)
- Auto-discovery and alphanumeric ordering
- Single responsibility per file
- Cross-reference patterns

### 2.2 Architecture Strategy

[To be populated]

### 2.3 Design Strategy

[To be populated]

#### 2.3.1 Core Principles

- Quality over speed (NO EXCEPTIONS)
- Evidence-based validation
- Correctness, completeness, thoroughness
- Reliability and efficiency

#### 2.3.2 Autonomous Execution Principles

- Continuous work without stopping
- No permission requests between tasks
- Blocker documentation and parallel work
- Task completion criteria

### 2.4 Implementation Strategy

[To be populated]

### 2.5 Quality Strategy

[To be populated]

---

## 3. Product Suite Architecture

### 3.1 Product Overview

[To be populated]

### 3.2 Service Catalog

[To be populated]

#### 3.2.1 PKI Product

- Certificate Authority (CA): X.509 certificates, EST, SCEP, OCSP, CRL
- Product-Service Identifier: pki-ca
- Public Port Range: 8050-8059 (host), 8080 (container)
- Private Admin Port: 9090 (container only)

#### 3.2.2 JOSE Product

- JWK Authority (JA): JWK/JWS/JWE/JWT operations
- Product-Service Identifier: jose-ja
- Public Port Range: 8060-8069 (host), 8080 (container)
- Private Admin Port: 9090 (container only)

#### 3.2.3 Cipher Product

- Instant Messenger (IM): E2E encrypted messaging, encryption-at-rest
- Product-Service Identifier: cipher-im
- Public Port Range: 8070-8079 (host), 8080 (container)
- Private Admin Port: 9090 (container only)

#### 3.2.4 Secrets Manager (SM) Product

- Key Management Service (KMS): Elastic key management, encryption-at-rest
- Product-Service Identifier: sm-kms
- Public Port Range: 8080-8089 (host), 8080 (container)
- Private Admin Port: 9090 (container only)

#### 3.2.5 Identity Product

- Authorization Server (Authz): OAuth 2.1 authorization server
- Identity Provider (IdP): OIDC 1.0 Identity Provider
- Resource Server (RS): OAuth 2.1 Resource Server
- Relying Party (RP): OAuth 2.1 Relying Party
- Single Page Application (SPA): OAuth 2.1 Single Page Application
- Product-Service Identifiers: identity-authz, identity-idp, identity-rs, identity-rp, identity-spa
- Public Port Ranges: 8100-8109 (authz), 8110-8119 (idp), 8120-8129 (rs), 8130-8139 (rp), 8140-8149 (spa)
- Private Admin Port: 9090 (all services, container only)

### 3.3 Product-Service Relationships

[To be populated]

### 3.4 Port Assignments & Networking

[To be populated]

#### 3.4.1 Port Design Principles

- HTTPS protocol for all public and admin port bindings
- Same HTTPS 127.0.0.1:9090 for Private HTTPS Admin APIs inside Docker Compose and Kubernetes (never localhost due to IPv4 vs IPv6 dual stack issues)
- Same HTTPS 0.0.0.0:8080 for Public HTTPS APIs inside Docker Compose and Kubernetes
- Different HTTPS 127.0.0.1 port range mappings for Public APIs on Docker host (to avoid conflicts)
- Same health check paths: `/browser/api/v1/health`, `/service/api/v1/health` on Public HTTPS listeners
- Same health check paths: `/admin/api/v1/livez`, `/admin/api/v1/readyz` on Private HTTPS Admin listeners
- Same graceful shutdown path: `/admin/api/v1/shutdown` on Private HTTPS Admin listeners

#### 3.4.2 PostgreSQL Ports

- Container address: 0.0.0.0:5432 (all services)
- Host address: 127.0.0.1 (all services)
- Host port ranges: 54320-54328 (one per service to avoid conflicts)
- Examples: pki-ca (54320), jose-ja (54321), cipher-im (54322), sm-kms (54323), identity-authz (54324)

#### 3.4.3 Telemetry Ports (Shared)

- opentelemetry-collector-contrib: 4317 (OTLP gRPC), 4318 (OTLP HTTP)
- grafana-otel-lgtm: 3000 (HTTP UI), 4317 (OTLP gRPC), 4318 (OTLP HTTP)

---

## 4. System Architecture

### 4.1 System Context

[To be populated]

### 4.2 Container Architecture

[To be populated]

### 4.3 Component Architecture

[To be populated]

#### 4.3.1 Layered Architecture

- main() [cmd/] → Application [internal/*/application/] → Business Logic [internal/*/service/, internal/*/domain/] → Repositories [internal/*/repository/] → Database/External Systems
- Dependency flow: One-way only (top → bottom)
- Cross-cutting concerns: Telemetry, logging, error handling

#### 4.3.2 Dependency Injection

- Constructor injection pattern: NewService(logger, repo, config)
- Factory pattern: *FromSettings functions for configuration-driven initialization
- Context propagation: Pass context.Context to all long-running operations

### 4.4 Code Organization

[To be populated]

#### 4.4.1 Go Project Structure

Based on golang-standards/project-layout:
- cmd/: Applications (external entry points for binary executables)
- internal/apps/: Applications (internal implementations organized as PRODUCT/SERVICE/)
- internal/shared/: Shared utilities (apperr, config, crypto, magic, pool, telemetry, testutil, util)
- api/: OpenAPI specs, generated code
- configs/: Configuration files
- deployments/: Docker Compose, Kubernetes manifests
- docs/: Documentation
- pkg/: Public library code (intentionally empty - all code is internal)
- scripts/: Build/test scripts (minimal, prefer Go applications)
- test/: Additional test files (Gatling load tests)

#### 4.4.2 Directory Rules

- ❌ Avoid /src directory (redundant in Go)
- ❌ Avoid deep nesting (>8 levels indicates design issue)
- ✅ Use /internal for private code (enforced by compiler)
- ✅ Use /pkg for public libraries (safe for external import - currently empty by design)

---

## 5. Service Architecture

### 5.1 Service Template Pattern

[To be populated]

#### 5.1.1 Template Components

- Two HTTPS Listeners: Public (business APIs) + Admin (health checks)
- Two Public Paths: /browser/** (sessions) vs /service/** (tokens)
- Three Admin APIs: /admin/api/v1/livez, /admin/api/v1/readyz, /admin/api/v1/shutdown
- Database: PostgreSQL || SQLite with GORM
- Telemetry: OTLP → otel-collector-contrib → Grafana LGTM
- Config Priority: Docker secrets > YAML > CLI parameters (NO environment variables)

#### 5.1.2 Template Benefits

- Eliminates 260+ lines of boilerplate per service
- Consistent infrastructure across all 9 services
- Proven patterns: TLS setup, middleware stacks, health checks, graceful shutdown
- Parameterization: OpenAPI specs, handlers, middleware chains injected via constructor

#### 5.1.3 Mandatory Usage

- ALL new services MUST use template (consistency, reduced duplication)
- ALL existing services MUST be refactored to use template (iterative migration)
- Migration priority: cipher-im FIRST (validation) → jose-ja → pki-ca → identity services → sm-kms LAST

### 5.2 Service Builder Pattern

[To be populated]

#### 5.2.1 Builder Methods

- NewServerBuilder(ctx, cfg): Create builder with template config
- WithDomainMigrations(fs, path): Register domain migrations (2001+)
- WithPublicRouteRegistration(func): Register domain-specific public routes
- Build(): Construct complete infrastructure and return ServiceResources

#### 5.2.2 Merged Migrations

- Template migrations: 1001-1004 (sessions, barrier, realms, tenants)
- Domain migrations: 2001+ (application-specific tables)
- mergedMigrations type: Implements fs.FS interface, unifies both for golang-migrate validation
- Prevention: Solves "no migration found for version X" validation errors

#### 5.2.3 ServiceResources

- Returns initialized infrastructure: DB (GORM), TelemetryService, JWKGenService, BarrierService, UnsealKeysService, SessionManager, RealmService, Application
- Shutdown functions: ShutdownCore(), ShutdownContainer()
- Domain code receives all dependencies ready-to-use

### 5.3 Dual HTTPS Endpoint Pattern

[To be populated]

#### 5.3.1 Public HTTPS Endpoint

- Purpose: Business APIs, browser UIs, external client access
- Default Binding: 127.0.0.1 (dev/test), 0.0.0.0 (containers)
- Default Port: Service-specific ranges (8080-8089 KMS, 8100-8149 Identity, etc.)
- Request Paths: /service/** (headless clients) and /browser/** (browser clients)

#### 5.3.2 Private HTTPS Endpoint (Admin Server)

- Purpose: Administration, health checks, graceful shutdown
- Default Binding: 127.0.0.1 (ALWAYS - admin localhost-only in all environments)
- Default Port: 9090 (standard, never exposed to host)
- APIs: /admin/api/v1/livez, /admin/api/v1/readyz, /admin/api/v1/shutdown

#### 5.3.3 Binding Defaults by Environment

- Unit/Integration Tests: Port 0 (dynamic allocation), 127.0.0.1 (prevents Windows Firewall prompts), HTTPS (production parity)
- Docker Containers: Public 0.0.0.0:8080 (external access), Private 127.0.0.1:9090 (admin isolated)
- Production: Public configurable, Private ALWAYS 127.0.0.1:9090 (security isolation)

### 5.4 Dual API Path Pattern

[To be populated]

#### 5.4.1 Service-to-Service APIs (/service/**)

- Access: Service clients ONLY (headless, non-browser)
- Authentication: Bearer tokens, mTLS, OAuth 2.1 client credentials
- Middleware: IP allowlist → Rate limiting → Request logging → Authentication → Authorization (scope-based)
- Examples: /service/api/v1/keys, /service/api/v1/tokens
- Browser clients: BLOCKED

#### 5.4.2 Browser-to-Service APIs/UI (/browser/**)

- Access: Browser clients ONLY (user-facing UIs)
- Authentication: Session cookies, OAuth tokens, social login
- Middleware: IP allowlist → CSRF protection → CORS policies → CSP headers → Rate limiting → Request logging → Authentication → Authorization (resource-level)
- Additional Content: HTML pages, JavaScript, CSS, images, fonts
- Examples: /browser/api/v1/keys, /browser/login, /browser/assets/app.js
- Service clients: BLOCKED

#### 5.4.3 API Consistency & Mutual Exclusivity

- SAME OpenAPI Specification served at both /service/** and /browser/** paths
- API contracts identical, only middleware/authentication differ
- Middleware enforces authorization mutual exclusivity (headless → /service/**, browser → /browser/**)
- E2E tests MUST verify BOTH path prefixes

### 5.5 Health Check Patterns

[To be populated]

#### 5.5.1 Liveness Check (/admin/api/v1/livez)

- Purpose: Is process alive?
- Check: Lightweight (goroutines not deadlocked, minimal logic)
- Response: HTTP 200 OK (healthy) or 503 Unavailable (unhealthy)
- Failure Action: Restart container
- Kubernetes: liveness probe

#### 5.5.2 Readiness Check (/admin/api/v1/readyz)

- Purpose: Is service ready for traffic?
- Check: Heavyweight (database connected, dependent services healthy, critical resources available)
- Response: HTTP 200 OK (ready) or 503 Unavailable (not ready)
- Failure Action: Remove from load balancer (do NOT restart)
- Kubernetes: readiness probe

#### 5.5.3 Graceful Shutdown (/admin/api/v1/shutdown)

- Purpose: Trigger shutdown sequence
- Actions: Stop accepting new requests, drain active requests (up to 30s timeout), close database connections, release resources, exit process
- Response: HTTP 200 OK
- Kubernetes: preStop hook

#### 5.5.4 Why Two Separate Health Endpoints (Kubernetes Standard)

| Scenario | Liveness | Readiness | Action |
|----------|----------|-----------|--------|
| Process alive, dependencies healthy | ✅ Pass | ✅ Pass | Serve traffic |
| Process alive, dependencies down | ✅ Pass | ❌ Fail | Remove from LB, don't restart |
| Process stuck/deadlocked | ❌ Fail | ❌ Fail | Restart container |

---

## 6. Security Architecture

### 6.1 FIPS 140-3 Compliance Strategy

[To be populated]

### 6.2 SDLC Security Strategy

[To be populated]

### 6.3 Product Security Strategy

[To be populated]

### 6.4 Cryptographic Architecture

[To be populated]

#### 6.4.1 Barrier Service (Multi-Layer Key Hierarchy)

- Unseal keys (Docker secrets, NEVER stored in database)
- Root keys (encrypted-at-rest with unseal keys)
- Intermediate keys (encrypted-at-rest with root keys)
- Content keys (encrypted-at-rest with intermediate keys)
- Domain data encryption (encrypted-at-rest with content keys)

#### 6.4.2 Unseal Modes

- Simple keys: File-based unseal key loading
- Shared secrets: M-of-N Shamir secret sharing (e.g., 3-of-5)
- System fingerprinting: Device-specific unseal key derivation
- High availability patterns for multi-instance deployments

#### 6.4.3 Key Rotation Strategies

- Root keys: Annual rotation (manual or automatic)
- Intermediate keys: Quarterly rotation
- Content keys: Monthly or per-operation rotation
- Elastic key pattern: Active key + historical keys for decryption

### 6.5 PKI Architecture & Strategy

[To be populated]

### 6.6 JOSE Architecture & Strategy

[To be populated]

### 6.7 Key Management System Architecture

[To be populated]

### 6.8 Multi-Factor Authentication Strategy

[To be populated]

### 6.9 Authentication & Authorization

[To be populated]

#### 6.9.1 Authentication Realm Architecture

- Realm types and purposes
- Credential validators (File, Database, Federated)
- Session creation vs session upgrade flows
- Multi-tenancy isolation via realms

#### 6.9.2 Headless Authentication Methods (13 Total)

- Non-Federated: JWE/JWS/Opaque session tokens, Basic (client ID/secret), Bearer (API token), HTTPS client cert
- Federated: OAuth 2.1 client credentials, Bearer (federated), mTLS, JWE/JWS/Opaque access tokens, Opaque refresh tokens

#### 6.9.3 Browser Authentication Methods (28 Total)

- All headless methods PLUS: TOTP, HOTP, Recovery codes, WebAuthn (with/without passkeys), Push notification
- Email/Phone factors: Password, OTP, Magic link (email/SMS/voice)
- Social login: Google, Microsoft, GitHub, Facebook, Apple, LinkedIn, Twitter/X, Amazon, Okta
- SAML 2.0 federation

#### 6.9.4 Multi-Factor Authentication (MFA)

- Step-up authentication (re-auth every 30min for high-sensitivity operations)
- Factor enrollment workflows
- MFA bypass policies and emergency access

#### 6.9.5 Authorization Patterns

- Zero trust: NO caching of authorization decisions
- Scope-based authorization (headless)
- Resource-based ACLs (browser)
- Consent tracking at scope + resource granularity

---

## 7. Data Architecture

### 7.1 Multi-Tenancy Architecture & Strategy

[To be populated]

#### 7.1.1 Schema-Level Isolation

- Each tenant gets separate schema (NOT row-level with tenant_id column)
- Schema naming: `tenant_<uuid>.users`, `tenant_<uuid>.sessions`
- Rationale: Data isolation, compliance, performance (per-tenant indexes)
- Pattern: Set search_path on connection

#### 7.1.2 Tenant Registration Flow

- tenant_id absent in register request → Create new tenant
- tenant_id present in register request → Join existing tenant
- Realm-based authentication (NOT for data filtering)
- tenant_id scopes ALL data access (keys, sessions, audit logs)

#### 7.1.3 Realm vs Tenant Isolation

- realm_id: Authentication context ONLY (session lifetimes, auth policies)
- tenant_id: Data isolation (MANDATORY for all database queries)
- Cross-realm access within same tenant: Supported
- Cross-tenant access: FORBIDDEN (security boundary)

### 7.2 Dual Database Strategy

[To be populated]

### 7.3 Database Schema Patterns

[To be populated]

### 7.4 Migration Strategy

[To be populated]

### 7.5 Data Security & Encryption

[To be populated]

---

## 8. API Architecture

### 8.1 OpenAPI-First Design

[To be populated]

#### 8.1.1 Specification Organization

- Split specifications: components.yaml + paths.yaml
- Benefits: Easier reviews, smaller diffs, reduced git conflicts, better IDE performance
- Version: OpenAPI 3.0.3 (NOT 2.0/Swagger, NOT 3.1.x)

#### 8.1.2 Code Generation with oapi-codegen

- Three config files: Server, Model, Client
- strict-server: true (type safety + validation)
- Single source of truth for models (prevents drift)

#### 8.1.3 Validation Rules

- String: format (uuid), enum, minLength/maxLength, pattern
- Number: minimum/maximum, multipleOf
- Array: minItems/maxItems, uniqueItems, items
- Object: required, nested properties

### 8.2 REST Conventions

[To be populated]

#### 8.2.1 Resource Naming

- Plural nouns: /keys, /certificates
- Singular for singletons: /config, /health
- Kebab-case: /api-keys

#### 8.2.2 HTTP Method Semantics

- GET /keys: List (with pagination)
- POST /keys: Create
- GET /keys/{id}: Get
- PUT /keys/{id}: Replace (full update)
- PATCH /keys/{id}: Update (partial)
- DELETE /keys/{id}: Delete

#### 8.2.3 Status Codes

- 200 OK: GET, PUT, PATCH successful
- 201 Created: POST successful (resource created)
- 204 No Content: DELETE successful
- 400 Bad Request: Validation error
- 401 Unauthorized: Missing/invalid authentication
- 403 Forbidden: Valid auth but insufficient permissions
- 404 Not Found: Resource does not exist
- 409 Conflict: Duplicate resource
- 422 Unprocessable Entity: Semantic validation error
- 500 Internal Server Error: Unhandled server error
- 503 Service Unavailable: Temporary unavailability

### 8.3 API Versioning

[To be populated]

#### 8.3.1 Versioning Strategy

- Path-based versioning: /api/v1/, /api/v2/
- N-1 backward compatibility (current + previous version)
- Major version for breaking changes, minor version within same major

#### 8.3.2 Version Lifecycle

- Deprecation warning period: 6+ months
- Documentation of migration path
- Parallel operation of N and N-1 versions

### 8.4 Error Handling

[To be populated]

#### 8.4.1 Error Schema Format

- Consistent format across all services
- Required fields: code, message
- Optional fields: details (object), requestId (uuid)
- Example: {"code": "INVALID_KEY_SIZE", "message": "Key size must be 2048, 3072, or 4096 bits", "details": {...}, "requestId": "uuid"}

#### 8.4.2 Error Code Naming

- Pattern: ERR_<CATEGORY>_<SPECIFIC>
- Examples: ERR_INVALID_REQUEST, ERR_AUTHENTICATION_FAILED, ERR_KEY_NOT_FOUND

#### 8.4.3 Error Details

- Provide actionable information
- Include field-level validation errors
- Never expose internal implementation details

### 8.5 API Security

[To be populated]

#### 8.5.1 Dual Path Security

- `/service/**` paths: IP allowlist, rate limiting, Bearer token or mTLS authentication
- `/browser/**` paths: CSRF protection, CORS policies, CSP headers, session cookie authentication
- Middleware enforces mutual exclusivity

#### 8.5.2 Rate Limiting

- Public APIs: 100 req/min per IP (burst: 20)
- Admin APIs: 10 req/min per IP (burst: 5)
- Login endpoints: 5 req/min per IP (burst: 2)
- Token bucket algorithm

#### 8.5.3 Content Type

- ALWAYS application/json (NEVER text/plain, application/xml)
- Consistent error responses in JSON format

---

## 9. Infrastructure Architecture

### 9.1 CLI Patterns & Strategy

[To be populated]

#### 9.1.1 Suite-Level CLI Pattern

- Unified cryptoutil executable
- Product → Service → Subcommand routing
- Zero-dependency binary distribution

#### 9.1.2 Product-Level CLI Pattern

- Separate executable per product
- Service delegation patterns
- Multi-service orchestration

#### 9.1.3 Service-Level CLI Pattern

- Standalone service executables
- Direct subcommand execution
- Container deployment patterns

### 9.2 Configuration Architecture & Strategy

[To be populated]

#### 9.2.1 Configuration Priority Order

- Docker secrets (highest priority)
- YAML configuration files
- CLI arguments
- Environment variables (NEVER for credentials)

#### 9.2.2 Secret Management Patterns

- Docker/Kubernetes secrets mounting
- File-based secret references (file://)
- Secret rotation strategies

### 9.3 Observability Architecture (OTLP)

[To be populated]

### 9.4 Telemetry Strategy

[To be populated]

### 9.5 Container Architecture

[To be populated]

### 9.6 Orchestration Patterns

[To be populated]

### 9.7 CI/CD Workflow Architecture

[To be populated]

#### 9.7.1 Workflow Catalog

- ci-coverage: Test coverage collection and enforcement
- ci-mutation: Mutation testing with gremlins
- ci-race: Race condition detection
- ci-benchmark: Performance benchmarking
- ci-quality: Linting and code quality
- ci-sast: Static application security testing
- ci-dast: Dynamic application security testing
- ci-e2e: End-to-end integration testing
- ci-load: Load testing with Gatling
- ci-gitleaks: Secret detection
- release: Automated release workflows

#### 9.7.2 Workflow Optimization Patterns

- Path filters to skip irrelevant changes
- Matrix strategies for package-level parallelization
- Docker image pre-pull for faster execution
- Dependency chain optimization
- Timeout tuning (20min CI, 45min mutation, 60min E2E)

#### 9.7.3 Workflow Dependencies

- Critical path: Test execution for largest packages
- Expected durations by workflow type
- Parallel vs sequential execution patterns

### 9.8 Reusable Action Patterns

[To be populated]

#### 9.8.1 Action Catalog

- docker-images-pull: Parallel Docker image pre-fetching
- setup-go: Go toolchain configuration
- cache-go: Go module and build cache management
- Additional actions TBD

#### 9.8.2 Action Composition Patterns

- Composite steps with shell selection
- Input/output parameter passing
- Cross-platform compatibility

### 9.9 Pre-Commit Hook Architecture

[To be populated]

#### 9.9.1 Hook Execution Flow

- Formatting hooks (gofmt, goimports, gofumpt)
- Linting hooks (golangci-lint)
- Security hooks (gitleaks, detect-secrets)
- Custom enforcement hooks (cicd-enforce-internal)

#### 9.9.2 Hook Configuration Patterns

- Repository-level .pre-commit-config.yaml
- Language-specific hooks
- Skip patterns and exclusions
- Fail-fast vs continue-on-error strategies

---

## 10. Testing Architecture

### 10.1 Testing Strategy Overview

[To be populated]

### 10.2 Unit Testing Strategy

[To be populated]

#### 10.2.1 Table-Driven Test Pattern

- MANDATORY for multiple test cases
- Single test function with test table
- t.Parallel() for concurrent execution
- UUIDv7 for dynamic test data

#### 10.2.2 Fiber Handler Testing (app.Test())

- MANDATORY for ALL HTTP handler tests (unit and integration)
- In-memory testing (NO real HTTPS listeners)
- Fast (<1ms), reliable, no network binding
- Prevents Windows Firewall popups

#### 10.2.3 Coverage Targets

- ≥95% production code
- ≥98% infrastructure/utility code
- 0% acceptable for main() if internalMain() ≥95%
- Generated code excluded from coverage

### 10.3 Integration Testing Strategy

[To be populated]

#### 10.3.1 TestMain Pattern

- MANDATORY for heavyweight dependencies (PostgreSQL, servers)
- Start resources ONCE per package
- Share testDB, testServer across all tests
- Prevents repeated 10-30s container startup overhead

#### 10.3.2 Test Isolation with t.Parallel()

- MANDATORY in ALL test functions and subtests
- Reveals race conditions, deadlocks, data conflicts
- If tests can't run concurrently, production can't either
- Dynamic test data (UUIDv7) prevents conflicts

#### 10.3.3 Database Testing Patterns

- Use real databases (testcontainers) NOT mocks
- SQLite WAL mode + busy_timeout for concurrent writes
- MaxOpenConns=5 for GORM (transaction wrapper needs separate connection)
- Cross-database compatibility (PostgreSQL + SQLite)

### 10.4 E2E Testing Strategy

[To be populated]

#### 10.4.1 Docker Compose Orchestration

- ComposeManager for lifecycle management
- Health check polling with TLS client
- Sequential startup (builder → postgres → app)
- Latency hiding strategies

#### 10.4.2 E2E Test Scope

- MUST test BOTH `/service/**` and `/browser/**` paths
- Verify middleware behavior (IP allowlist, CSRF, CORS)
- Production-like environment (Docker secrets, TLS)

### 10.5 Mutation Testing Strategy

[To be populated]

#### 10.5.1 Gremlins Configuration

- Package-level parallelization (4-6 packages per job)
- Exclude tests, generated code, vendor
- Efficacy targets: ≥95% production, ≥98% infrastructure
- Timeout optimization: sequential 45min → parallel 15-20min

#### 10.5.2 Mutation Exemptions

- OpenAPI-generated code (stable, no business logic)
- GORM models (database schema definitions)
- Protobuf-generated code (gRPC/protobuf stubs)

### 10.6 Load Testing Strategy

[To be populated]

### 10.7 Fuzz Testing Strategy

[To be populated]

### 10.8 Benchmark Testing Strategy

[To be populated]

### 10.9 Race Detection Strategy

[To be populated]

### 10.10 SAST Strategy

[To be populated]

### 10.11 DAST Strategy

[To be populated]

### 10.12 Workflow Testing Strategy

[To be populated]

---

## 11. Quality Architecture

### 11.1 Maximum Quality Strategy

[To be populated]

#### 11.1.1 Go Version Consistency

- MANDATORY: Use same Go version everywhere (development, CI/CD, Docker, documentation)
- Current Version: 1.25.5 (check go.mod)
- Enforcement Locations: go.mod (go 1.25.5), .github/workflows/*.yml (GO_VERSION: '1.25.5'), Dockerfile (FROM golang:1.25.5-alpine), README.md (document Go 1.25.5+ requirement)
- Update Policy: Security patches (apply immediately), minor versions (update monthly), major versions (evaluate quarterly)

#### 11.1.2 CGO Ban - CRITICAL

- MANDATORY: CGO_ENABLED=0 for all builds, tests, Docker, production
- ONLY EXCEPTION: Race detector requires CGO_ENABLED=1 (Go toolchain limitation)
- NEVER use CGO-dependent packages (e.g., github.com/mattn/go-sqlite3)
- ALWAYS use CGO-free alternatives (e.g., modernc.org/sqlite)
- Rationale: Maximum portability (no C toolchain), static linking (single binary), cross-compilation, no C library version conflicts
- Detection: `go list -u -m all | grep '\[.*\]$'` (shows CGO dependencies)
- Enforcement: Custom lint_go checker validates go.mod and imports

#### 11.1.3 Import Alias Conventions

- Internal packages: cryptoutil<Package> (camelCase) - Example: cryptoutilMagic, cryptoutilServer
- Third-party packages: <vendor><Package> - Examples: crand (crypto/rand), googleUuid (github.com/google/uuid)
- Configuration: .golangci.yml importas section enforces consistency
- Rationale: Avoids naming conflicts, improves readability

#### 11.1.4 Magic Values Organization

- Shared constants: internal/shared/magic/magic_*.go (network, database, cryptography, testing)
- Domain-specific constants: internal/<package>/magic*.go
- Pattern: Declare as named variables, NEVER inline literals
- Rationale: mnd (magic number detector) linter enforcement

### 11.2 Quality Gates

[To be populated]

#### 11.2.1 File Size Limits

| Threshold | Lines | Action |
|-----------|-------|--------|
| Soft | 300 | Ideal target |
| Medium | 400 | Acceptable with justification |
| Hard | 500 | NEVER EXCEED - refactor required |

- Rationale: Faster LLM processing, easier review, better organization, forces logical grouping

#### 11.2.2 Conditional Statement Patterns

- PREFER switch statements over if/else if/else chains for cleaner, more maintainable code
- Pattern for mutually exclusive conditions:
  ```go
  switch {
  case ctx == nil:
      return nil, fmt.Errorf("nil context")
  case logger == nil:
      return nil, fmt.Errorf("nil logger")
  default:
      return processValid(ctx, logger, description)
  }
  ```
- When NOT to chain: Independent conditions, error accumulation, early returns

#### 11.2.3 format_go Self-Modification Protection - CRITICAL

- Root Cause: LLM agents lose exclusion context during narrow-focus refactoring
- NEVER DO:
  * ❌ Modify comments in enforce_any.go without reading full package context
  * ❌ Change backticked `interface{}` to `any` in format_go package
  * ❌ Refactor code in isolation (single-file view)
  * ❌ Simplify "verbose" CRITICAL comments
- ALWAYS DO:
  * ✅ Read complete package context before refactoring self-modifying code
  * ✅ Check for CRITICAL/SELF-MODIFICATION tags in comments
  * ✅ Verify self-exclusion patterns exist and are respected
  * ✅ Run tests after ANY changes to format_go package

#### 11.2.4 Restore from Clean Baseline Pattern

- When: Fixing regressions, multiple failed attempts, uncertain HEAD state
- Steps:
  1. Find last known-good commit: `git log --grep="baseline"` or `git bisect`
  2. Restore package: `git checkout <hash> -- path/to/package/`
  3. Verify baseline works: `go test`
  4. Apply ONLY new fix (targeted change)
  5. Verify fix works
  6. Commit as NEW commit (NOT amend)
- Rationale: HEAD may be corrupted by failed attempts, start from verified clean state

### 11.3 Code Quality Standards

[To be populated]

#### 11.3.1 Linter Configuration Architecture

- golangci-lint v2 configuration
- Enabled linters: errcheck, govet, staticcheck, unused, revive, gosec, etc.
- Disabled linters with justification
- Version-specific syntax (wsl → wsl_v5 in v2)

#### 11.3.2 Linter Exclusion Patterns

- Path-based exclusions (api/*, test-output/*, vendor/)
- Rule-based exclusions (nilnil, wrapcheck for specific files)
- Test file exemptions

#### 11.3.3 Code Quality Enforcement

- Zero linting errors policy (NO exceptions)
- Auto-fixable linters (--fix workflow)
- Manual fix requirements
- Pre-commit hook integration

#### 11.3.4 Mutation Testing Architecture

- gremlins configuration and execution
- Package-level parallelization strategy
- Efficacy targets: ≥95% production, ≥98% infrastructure
- Timeout optimization (sequential 45min → parallel 15-20min)
- Exclusion patterns (generated code, test utilities)

### 11.4 Documentation Standards

[To be populated]

#### 11.4.1 Documentation Organization

- Primary: README.md and docs/README.md (keep in 2 files)
- Spec structure: plan.md and tasks.md patterns
- NEVER create standalone session/analysis docs
- Append to existing docs instead of creating new files

#### 11.4.2 Documentation Frontmatter

- YAML metadata for machine readability
- Title, version, date, status, audience, tags
- Cross-reference patterns

#### 11.4.3 Lean Documentation Principle

- Avoid duplication across files
- Reference external resources
- Single source of truth patterns

### 11.5 Review Processes

[To be populated]

---

## 12. Deployment Architecture

### 12.1 CI/CD Automation Strategy

[To be populated]

### 12.2 Build Pipeline

[To be populated]

#### 12.2.1 Multi-Stage Dockerfile Pattern

- Global ARGs at top (GO_VERSION, VCS_REF, BUILD_DATE)
- Builder stage (compile Go binaries)
- Validator stage (secrets validation MANDATORY)
- Runtime stage (Alpine-based minimal image)
- LABELs in final published image only

#### 12.2.2 Build Optimization

- Single build, shared image (prevents 3× build time)
- Docker image pre-pull for faster workflow execution
- BuildKit caching strategies
- Cross-platform build support

#### 12.2.3 Secret Validation Stage

- MANDATORY in all Dockerfiles
- Validates Docker secrets existence and permissions
- Fails fast on missing/misconfigured secrets
- Prevents runtime secret access errors

### 12.3 Deployment Patterns

[To be populated]

#### 12.3.1 Docker Compose Deployment

- Secret management via Docker secrets (MANDATORY)
- Health check configuration (interval, timeout, retries, start-period)
- Dependency ordering (depends_on with service_healthy)
- Network isolation patterns

#### 12.3.2 Kubernetes Deployment

- secretKeyRef for environment variables from secrets
- Volume mount for file-based secrets
- Health probes (liveness vs readiness)
- Resource limits and requests

### 12.4 Environment Strategy

[To be populated]

### 12.5 Release Management

[To be populated]

---

## 13. Development Practices

### 13.1 Coding Standards

[To be populated]

### 13.2 Version Control

[To be populated]

#### 13.2.1 Conventional Commits

- Format: `<type>[scope]: <description>`
- Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert
- Breaking changes: Use `!` or `BREAKING CHANGE:`
- Examples: `feat(auth): add OAuth2 flow`, `fix(database): prevent pool exhaustion`

#### 13.2.2 Incremental Commit Strategy

- ALWAYS commit incrementally (NOT amend)
- Preserve timeline for bisect and selective revert
- Commit each logical unit independently
- Avoid: Repeated amend, hiding fixes, amend after push

#### 13.2.3 Restore from Clean Baseline Pattern

- Find last known-good commit
- Restore entire package from clean commit
- Verify baseline works
- Apply targeted fix ONLY
- Commit as NEW commit (not amend)

### 13.3 Branching Strategy

[To be populated]

### 13.4 Code Review

[To be populated]

### 13.5 Development Workflow

[To be populated]

#### 13.5.1 Spec Structure Patterns

- plan.md: Vision, phases, success criteria, anti-patterns
- tasks.md: Phase breakdown, task checklist, dependencies
- DETAILED.md: Session timeline with date-stamped entries
- Coverage tracking by package

#### 13.5.2 Terminal Command Auto-Approval

- Pattern checking against .vscode/settings.json
- Auto-enable: Read-only and build operations
- Auto-disable: Destructive operations
- autoapprove wrapper for loopback network commands

#### 13.5.3 Session Documentation Strategy

- MANDATORY: Append to DETAILED.md Section 2 timeline
- Format: `### YYYY-MM-DD: Title`
- NEVER create standalone session docs
- DELETE completed tasks immediately from todos-*.md

---

## 14. Operational Excellence

### 14.1 Monitoring & Alerting

[To be populated]

### 14.2 Incident Management

[To be populated]

### 14.3 Performance Management

[To be populated]

### 14.4 Capacity Planning

[To be populated]

### 14.5 Disaster Recovery

[To be populated]

---

## Appendix A: Decision Records

### A.1 Architectural Decision Records (ADRs)

[To be populated]

### A.2 Technology Selection Decisions

[To be populated]

### A.3 Pattern Selection Decisions

[To be populated]

---

## Appendix B: Reference Tables

### B.1 Service Port Assignments

[To be populated]

### B.2 Database Port Assignments

[To be populated]

### B.3 Technology Stack

[To be populated]

### B.4 Dependency Matrix

[To be populated]

### B.5 Configuration Reference

[To be populated]

### B.6 Instruction File Reference

[To be populated - Mirror table from .github/copilot-instructions.md]

| File | Description |
|------|-------------|
| 01-01.terminology | RFC 2119 keywords (MUST, SHOULD, MAY, CRITICAL) |
| 01-02.beast-mode | Beast mode directive |
| 02-01.architecture | Products and services architecture patterns |
| ... | (25 total instruction files) |

### B.7 Agent Catalog & Handoff Matrix

[To be populated]

| Agent | Description | Tools | Handoffs |
|-------|-------------|-------|----------|
| plan-tasks-quizme | Planning and task decomposition | edit, execute, read, search, web | → plan-tasks-implement |
| plan-tasks-implement | Autonomous implementation execution | edit, execute, read, search, web | → doc-sync, fix-github-workflows |
| doc-sync | Documentation synchronization | TBD | TBD |
| fix-github-workflows | Workflow repair and validation | TBD | TBD |
| fix-tool-names | Tool name consistency enforcement | TBD | TBD |
| beast-mode-custom | Continuous execution mode | TBD | TBD |

### B.8 CI/CD Workflow Catalog

[To be populated]

| Workflow | Purpose | Dependencies | Duration | Timeout |
|----------|---------|--------------|----------|---------|
| ci-coverage | Test coverage collection, enforce ≥95%/98% | None | 5-6min | 20min |
| ci-mutation | Mutation testing with gremlins | None | 15-20min | 45min |
| ci-race | Race condition detection | None | TBD | 20min |
| ci-benchmark | Performance benchmarking | None | TBD | 30min |
| ci-quality | Linting and code quality | None | 3-5min | 15min |
| ci-sast | Static security analysis | None | TBD | 20min |
| ci-dast | Dynamic security testing | PostgreSQL | TBD | 30min |
| ci-e2e | End-to-end integration tests | Docker Compose | TBD | 60min |
| ci-load | Load testing with Gatling | Docker Compose | TBD | 45min |
| ci-gitleaks | Secret detection | None | 2-3min | 10min |
| release | Automated release workflows | ci-* passing | TBD | 30min |

### B.9 Reusable Action Catalog

[To be populated]

| Action | Description | Inputs | Outputs |
|--------|-------------|--------|---------|
| docker-images-pull | Parallel Docker image pre-fetching | images (newline-separated list) | None |
| Additional actions | TBD | TBD | TBD |

### B.10 Linter Rule Reference

[To be populated]

| Linter | Purpose | Enabled | Auto-Fix | Exclusions |
|--------|---------|---------|----------|------------|
| errcheck | Unchecked errors | ✅ | ❌ | Test helpers |
| govet | Suspicious code | ✅ | ❌ | None |
| staticcheck | Static analysis | ✅ | ❌ | Generated code |
| wsl_v5 | Whitespace linting | ✅ | ✅ | None |
| godot | Comment periods | ✅ | ✅ | None |
| gosec | Security issues | ✅ | ❌ | Justified cases |
| ... | (30+ total linters) | TBD | TBD | TBD |

---

## Appendix C: Compliance Matrix

### C.1 FIPS 140-3 Compliance

[To be populated]

### C.2 PKI Standards Compliance

[To be populated]

### C.3 OAuth 2.1 / OIDC 1.0 Compliance

[To be populated]

### C.4 Security Standards Compliance

[To be populated]

---

## Document Metadata

**Revision History**:
- v2.0 (2026-02-08): Initial skeleton structure
- v1.0 (historical): Original ARCHITECTURE.md

**Related Documents**:
- `.github/copilot-instructions.md` - Copilot configuration
- `.github/instructions/*.instructions.md` - Detailed instructions
- `.specify/memory/constitution.md` - Project constitution
- `docs/ARCHITECTURE.md` - Legacy architecture document

**Cross-References**:
- All sections maintain stable anchor links for referencing
- Machine-readable YAML frontmatter for metadata
- Consistent section numbering for navigation
