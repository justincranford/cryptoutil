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
  - docs/speckit/constitution.md
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

**Companion Document**: [ARCHITECTURE-INDEX.md](ARCHITECTURE-INDEX.md) provides a semantic topic index with line number ranges for efficient agent lookups. MUST be kept in sync with this document when sections are added, removed, or significantly reorganized.

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

**cryptoutil** is a production-ready suite of four cryptographic-based products, designed with enterprise-grade security, **FIPS 140-3** standards compliance, Zero-Trust principles, and security-on-by-default:

1. **Private Key Infrastructure (PKI)** - X.509 certificate management with EST, SCEP, OCSP, and CRL support
2. **JSON Object Signing and Encryption (JOSE)** - JWK/JWS/JWE/JWT cryptographic operations
3. **Secrets Manager (SM)** - Elastic key management service with hierarchical key barriers; also hosts the encrypted messaging service
4. **Identity** - OAuth 2.1, OIDC 1.0, WebAuthn, and Passkeys authentication and authorization

**Purpose**: This project is **for fun** while providing a comprehensive learning experience with LLM agents for Spec-Driven Development (SDD) and delivering modern, enterprise-ready security products.

### 1.2 Key Architectural Characteristics

#### 🔐 Cryptographic Standards

- **FIPS 140-3 Compliance**: Only NIST-approved algorithms (RSA ≥2048, AES ≥128, NIST curves, EdDSA)
- **Key Generation**: RSA, ECDSA, ECDH, EdDSA, AES, HMAC, UUIDv7 with concurrent key pools
- **JWE/JWS Support**: Full JSON Web Encryption and Signature implementation
- **Hierarchical Key Management**: Multi-tier barrier system (unseal → root → intermediate → content keys)

#### 🌐 API Architecture

- **Dual Context Design**: Browser API (`/browser/api/v1/*`) with CORS/CSRF/CSP vs Service API (`/service/api/v1/*`) for service-to-service
- **Management Interface** (`127.0.0.1:9090`): Private health checks and graceful shutdown
- **OpenAPI-Driven**: Auto-generated handlers, models, and interactive Swagger UI

#### 🛡️ Security Features

- **Multi-layered IP allowlisting**: Individual IPs + CIDR blocks
- **Per-IP rate limiting**: Separate thresholds for browser (100 req/sec) vs service (25 req/sec) APIs
- **CSRF protection** with secure token handling for browser clients
- **Content Security Policy (CSP)** for XSS prevention
- **Encrypted key storage** with barrier system protection

#### 📊 Observability & Monitoring

- **OpenTelemetry integration**: Traces, metrics, logs via OTLP
- **Structured logging** with slog
- **Kubernetes-ready health endpoints**: `/admin/api/v1/livez`, `/admin/api/v1/readyz`
- **Grafana-OTEL-LGTM stack**: Integrated Grafana, Loki, Tempo, and Prometheus

#### 🏗️ Production Ready

- **Database support**: PostgreSQL (production), SQLite (development/testing)
- **Container deployment**: Docker Compose with secret management
- **Configuration management**: YAML files + CLI parameters
- **Graceful shutdown**: Signal handling and connection draining

### 1.3 Core Principles

#### Security-First Design

- **Zero-Trust Architecture**: Never trust, always verify
- **Security-on-by-default**: Secure configurations without manual hardening, encryption-at-rest barrier layer mandatory, encryption-in-transit TLS mandatory
- **FIPS 140-3 Compliance**: Mandatory approved algorithms only
- **Defense in Depth**: Multiple security layers (encryption-at-rest, encryption-in-transit, barrier system)

#### Maximum Quality Strategy

- **NO EXCEPTIONS**: Quality, correctness, completeness, thoroughness, reliability, efficiency, and accuracy are ALL mandatory
- **Evidence-based validation**: Objective proof required for task completion
- **Reliability and efficiency**: Optimized for maintainability and performance, NOT implementation speed
- **Time/token pressure does NOT exist**: Work spans hours/days/weeks as needed

#### Microservices Architecture

- **Service isolation**: Each product-service runs independently
- **Dual HTTPS endpoints**: Public (business) + Private (admin) mandatory
- **Container-first**: Mandatory Docker support for production and E2E testing
- **Multi-tenancy**: Schema-level isolation with tenant_id scoping

#### Developer Experience

- **OpenAPI-first**: Auto-generated code from specifications
- **Comprehensive testing**: Build, unit, integration, E2E, code coverage, mutation, benchmark, fuzzing, race condition, property-based, load, SAST, DAST, gitleaks, linting, formatting
- **Observability built-in**: Structured logging, tracing, metrics from day one
- **Documentation as code**: Single source of truth for architecture and implementation

### 1.4 Success Metrics

#### Code Quality

- **Test Coverage**: ≥95% production code, ≥98% infrastructure/utility code
- **Mutation Testing**: ≥95% efficacy production, ≥98% infrastructure/utility
- **Linting**: Zero golangci-lint violations across all code (`golangci-lint run` and `golangci-lint run --build-tags e2e,integration`)
- **Build**: Clean `go build ./...` and `go build -tags e2e,integration ./...` with no errors or warnings

#### Performance

- **Test Execution**: <15s per package unit tests, <180s full suite
- **API Response**: <100ms p95 for cryptographic operations
- **Startup Time**: <10s server ready state
- **Container Build**: <60s multi-stage Docker build

#### Security

- **Vulnerability Scanning**: Zero high/critical CVEs in dependencies
- **Secret Management**: 100% Docker secrets (zero inline credentials)
- **TLS Configuration**: TLS 1.3+ only, full certificate chain validation
- **Authentication**: Multi-factor support across all services

#### Operational Excellence

- **Availability**: Health checks respond <100ms
- **Observability**: 100% operations traced and logged
- **Documentation**: Every feature documented in OpenAPI specs
- **CI/CD**: All workflows passing, <15 min total pipeline time

---

## 2. Strategic Vision & Principles

### 2.1 Agent Orchestration Strategy

#### 2.1.1 Agent Architecture

- Agent isolation principle (agents do NOT inherit copilot instructions)
- YAML frontmatter requirements (name, description, tools, handoffs)
- Autonomous execution mode patterns
- Quality over speed enforcement

**Implementation Plan File Structure**:

Implementation plans are composed of 4 files in `<work-dir>/`:
- `quizme-v#.md` - Ephemeral, only during implementation-planning.agent.md (A-D options + E blank + Answer field)
- `plan.md` - Created/updated during implementation-planning.agent.md, implemented during implementation-execution.agent.md
- `tasks.md` - Created/updated during implementation-planning.agent.md, implemented during implementation-execution.agent.md (phases and tasks as checkboxes, updated continuously)
- `memory.md` - Ephemeral, only during implementation-execution.agent.md (NOT in .github/instructions/memory.instruction.md - copilot instruction files are not loaded by agents)

#### 2.1.2 Agent Catalog

- implementation-planning: Planning and task decomposition
- implementation-execution: Autonomous implementation execution
- doc-sync: Documentation synchronization
- fix-workflows: Workflow repair and validation
- beast-mode: Continuous execution mode

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

#### Service Template Pattern

- **Single Reusable Template**: All 9 services across 5 products inherit from `internal/apps/template/`
- **Eliminates 48,000+ lines per service**: TLS setup, dual HTTPS servers, database, migrations, sessions, barrier
- **Merged Migrations**: Template (1001-1999) + Domain (2001+) for golang-migrate validation
- **Builder Pattern**: Fluent API with `NewServerBuilder(ctx, cfg).WithDomainMigrations(...).Build()`

#### Microservices Architecture

- **9 Services across 5 Products**: Independent deployment, scaling, and lifecycle
- **Dual HTTPS Endpoints**: Public (0.0.0.0:8080) for business APIs, Private (127.0.0.1:9090) for admin operations
- **Service Discovery**: Config file → Docker Compose → Kubernetes DNS (no caching)
- **Multi-Level Failover**: Services attempt credential validators in priority order (FEDERATED → DATABASE → FILE), with FILE realms as CRITICAL failsafe guaranteeing admin access

#### Multi-Tenancy

- **Schema-Level Isolation**: Each tenant gets separate schema (`tenant_<uuid>.users`)
- **tenant_id Scoping**: ALL data access filtered by tenant_id (not realm_id)
- **Realm-Based Authentication**: Authentication context only, NOT for data filtering
- **Registration Flow**: Create new tenant OR join existing tenant

#### Database Strategy

- **Dual Database Support**: PostgreSQL (production) + SQLite (dev/test)
- **Cross-DB Compatibility**: UUID as TEXT, serializer:json for arrays, NullableUUID for optional FKs
- **GORM Always**: Never raw database/sql for consistency
- **TestMain Pattern**: Heavyweight resources initialized once per package

### 2.3 Design Strategy

#### Domain-Driven Design

- **Layered Architecture**: main() → Application → Business Logic → Repositories → Database/External Systems
- **Domain Isolation**: Identity domain cannot import server/client/api layers
- **Bounded Contexts**: Each product-service has clear boundaries and responsibilities
- **Repository Pattern**: Abstract data access, enable testing with real databases

#### API-First Development

- **OpenAPI 3.0.3**: Single source of truth for API contracts
- **Code Generation**: oapi-codegen strict-server for type safety and validation
- **Dual Path Prefixes**: `/browser/**` (session-based) vs `/service/**` (session-based)
- **Consistent Error Schemas**: Unified error response format across all services

#### Configuration Management

- **Priority Order**: Docker secrets (highest) → YAML files → CLI parameters (lowest)
- **NO Environment Variables**: For configuration or secrets (security violation)
- **file:// Pattern**: Reference secrets as `file:///run/secrets/secret_name`
- **Hot-Reload Support**: Connection pool settings reconfigurable without restart

#### Security by Design

- **Barrier Layer Key Hierarchy**: Unseal → Root → Intermediate → Content keys
- **Elastic Key Rotation**: Active key for encrypt, historical keys for decrypt
- **PBKDF2 for Low-Entropy**: Passwords, PII (≥600k iterations)
- **HKDF for High-Entropy**: API keys, config blobs (deterministic derivation)
- **Pepper MANDATORY**: All hash inputs peppered before processing

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

#### Go Best Practices

- **Go Version**: 1.25.5+ (same everywhere: dev, CI/CD, Docker)
- **CGO Ban**: CGO_ENABLED=0 (except race detector) for maximum portability
- **Import Aliases**: `cryptoutil<Package>` for internal, `<vendor><Package>` for external
- **Magic Values**: `internal/shared/magic/magic_*.go` for shared, package-specific for domain

#### Testing Strategy

- **Table-Driven Tests**: ALWAYS use for multiple test cases (NOT standalone functions)
- **app.Test() Pattern**: ALL HTTP handler tests use in-memory testing (NO real servers)
- **TestMain Pattern**: Heavyweight resources (PostgreSQL, servers) initialized once per package
- **Dynamic Test Data**: UUIDv7 for all test values (thread-safe, process-safe, time-ordered)
- **t.Parallel()**: ALWAYS use in test functions and subtests for concurrency validation

#### Incremental Commits

- **Conventional Commits**: `<type>[scope]: <description>` format mandatory
- **Commit Strategy**: Incremental commits (NOT amend) preserve history for bisect
- **Restore from Clean**: When fixing regressions, restore known-good baseline first
- **Quality Gates**: Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`), linting clean (`golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`), tests pass before commit

#### Continuous Execution

- **Beast Mode**: Work autonomously until problem completely solved
- **Maximum Quality Strategy**: Correctness, completeness, thoroughness (NO EXCEPTIONS)
- **NO Stopping**: Task complete → Commit → IMMEDIATELY start next task (zero pause)
- **Blocker Handling**: Document blocker, switch to unblocked tasks, return when resolved

##### Autonomous Execution Principles

- **Quality Attributes**: Correctness (functionally correct), Completeness (all tasks done), Thoroughness (evidence-based validation), Reliability (quality gates enforced), Efficiency (maintainable/performance-optimized), Accuracy (root cause fixes)
- **Prohibited Stop Behaviors**: No status summaries, no "ready to proceed" questions, no strategic pivots with handoff, no time/token justifications, no pauses between tasks, no asking permission, no leaving uncommitted changes, no ending with analysis, no celebrations followed by stopping, no premature completion claims, no "current task done, moving to next" announcements
- **Execution Workflow**: Complete task → Commit → Next tool invocation (zero text, zero questions); Todo list empty → Check tracking docs → Find next incomplete task → Start immediately; All tasks done/blocked → Find quality improvements → Scan for technical debt → Review recent commits → Ask user if nothing left
- **Completion Verification Checklist**: Build clean, linting clean, tests pass (100%, zero skips), coverage maintained, mutation testing passes, evidence exists, git commit ready; After substantive change, run relevant build/tests/linters, validate code works (fast, minimal input), provide optional fenced commands for larger runs; Fix failures up to three targeted fixes, summarize root cause if still failing
- **Blocker Resolution**: Document in tracking doc, continue with ALL unblocked tasks, maximize progress, return to blocker when resolved; NO waiting for external dependencies

### 2.5 Quality Strategy

#### Coverage Targets

- **Production Code**: ≥95% minimum coverage
- **Infrastructure/Utility**: ≥98% minimum coverage
- **main() Functions**: 0% (exempt if internalMain() ≥95%)
- **Generated Code**: 0% (excluded - OpenAPI, GORM models, protobuf)

#### Mutation Testing

- **Category-Based Targets**: ≥98% ideal efficacy (all packages), ≥95% mandatory minimum
- **Tool**: gremlins v0.6.0+ (Linux CI/CD for Windows compatibility)
- **Execution**: `gremlins unleash --tags=!integration` per package
- **Timeouts**: 4-6 packages per parallel job, <20 minutes total

#### Linting Standards

- **Zero Exceptions**: ALL code must pass linting (production, tests, demos, utilities)
- **golangci-lint v2**: v2.7.2+ with wsl_v5, built-in formatters
- **Auto-Fixable**: Run `--fix` first (gofumpt, goimports, wsl, godot, importas)
- **Critical Rules**: wsl (no suppression), godot (periods required), mnd (magic constants)

#### Pre-Commit Hooks

- **Same as CI/CD**: golangci-lint, gofumpt, goimports, cicd-enforce-internal
- **Auto-Conversions**: `time.Now()` → `time.Now().UTC()` for SQLite compatibility
- **UTF-8 without BOM**: All text files mandatory enforcement
- **Hook Documentation**: Update `docs/pre-commit-hooks.md` with config changes

---

## 3. Product Suite Architecture

### 3.1 Product Overview

**cryptoutil** comprises five independent products, each providing specialized cryptographic capabilities:

#### 1. Private Key Infrastructure (PKI)

- **Service**: Certificate Authority (CA)
- **Capabilities**: X.509 certificate lifecycle management, EST, SCEP, OCSP, CRL
- **Use Cases**: TLS certificate issuance, client authentication, code signing
- **Architecture**: 3-tier CA hierarchy (Offline Root → Online Root → Issuing CA)

#### 2. JSON Object Signing and Encryption (JOSE)

- **Service**: JWK Authority (JA)
- **Capabilities**: JWK/JWS/JWE/JWT cryptographic operations, elastic key rotation
- **Use Cases**: API token generation, data encryption, digital signatures
- **Key Features**: Per-message key rotation, automatic key versioning

#### 3. Secrets Manager (SM)

- **Services**: Key Management Service (KMS), Instant Messenger (IM; renamed from sm-im)
- **Capabilities**: Elastic key management, hierarchical key barriers, encryption-at-rest, end-to-end encrypted messaging
- **Use Cases**: Application secrets, database encryption keys, API key management, secure communications
- **Key Features**: Unseal-based bootstrapping, automatic key rotation, message-level JWKs

#### 4. Identity

- **Services**: Authorization Server (Authz), Identity Provider (IdP), Resource Server (RS), Relying Party (RP), Single Page Application (SPA)
- **Capabilities**: OAuth 2.1, OIDC 1.0, WebAuthn, Passkeys, multi-factor authentication
- **Use Cases**: User authentication, API authorization, SSO, passwordless login
- **Key Features**: 41 authentication methods (13 headless + 28 browser), multi-tenancy

### 3.2 Service Catalog

| Product | Service | Product-Service Identifier | Address (Container) [Admin] | Address (Container) [Public] | Address (Host) [Public] | Port Value (Container) [Admin] | Port Value (Container) [Public] | Port Range (Host) [Service Deployment] | Port Range (Host) [Product Deployment] | Port Range (Host) [Suite Deployment] | Description |
|---------|---------|----------------------------|-----------------------------|-----------------------------|-------------------------|--------------------------------|---------------------------------|----------------------------------------|----------------------------------------|--------------------------------------|-------------|
| **Secrets Manager (SM)** | **Key Management Service (KMS)** | **sm-kms** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8000-8099 | 18000-18099 | 28000-28099 | Elastic key management, encryption-at-rest |
| **Private Key Infrastructure (PKI)** | **Certificate Authority (CA)** | **pki-ca** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8100-8199 | 18100-18199 | 28100-28199 | X.509 certificates, EST, SCEP, OCSP, CRL |
| **Identity** | **Authorization Server (Authz)** | **identity-authz** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8200-8299 | 18200-18299 | 28200-28299 | OAuth 2.1 authorization server |
| **Identity** | **Identity Provider (IdP)** | **identity-idp** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8300-8399 | 18300-18399 | 28300-28399 | OIDC 1.0 Identity Provider |
| **Identity** | **Resource Server (RS)** | **identity-rs** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8400-8499 | 18400-18499 | 28400-28499 | OAuth 2.1 Resource Server |
| **Identity** | **Relying Party (RP)** | **identity-rp** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8500-8599 | 18500-18599 | 28500-28599 | OAuth 2.1 Relying Party |
| **Identity** | **Single Page Application (SPA)** | **identity-spa** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8600-8699 | 18600-18699 | 28600-28699 | OAuth 2.1 Single Page Application |
| **Secrets Manager (SM)** | **Instant Messenger (IM)** | **sm-im** (renamed from sm-im) | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8700-8799 | 18700-18799 | 28700-28799 | E2E encrypted messaging, encryption-at-rest |
| **JSON Object Signing and Encryption (JOSE)** | **JWK Authority (JA)** | **jose-ja** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8800-8899 | 18800-18899 | 28800-28899 | JWK/JWS/JWE/JWT operations |

**Implementation Status**:

| Product-Service Identifier | Status | Completion | Notes |
|----------------------------|--------|------------|-------|
| **sm-kms** | ✅ Complete | 100% | Reference implementation with dual servers, Docker Compose |
| **pki-ca** | ⚠️ Partial | ~85% | Missing admin server, Docker Compose needs update |
| **jose-ja** | ⚠️ Partial | ~85% | Missing admin server, Docker Compose needs update |
| **sm-im** | ✅ Complete | 100% | Phase 8: renamed from sm-im |
| **identity-authz** | ✅ Complete | 100% | Dual servers, Docker Compose working |
| **identity-idp** | ✅ Complete | 100% | Dual servers, Docker Compose working |
| **identity-rs** | ✅ Complete | 100% | Dual servers, Docker Compose working |
| **identity-rp** | ❌ Not Started | 0% | Planned for Phase 6 of implementation |
| **identity-spa** | ❌ Not Started | 0% | Planned for Phase 6 of implementation |

**Legend**: ✅ Complete (production-ready), ⚠️ Partial (functional but missing features), ❌ Not Started

**See Also**: [docs/fixes-v1/](../fixes-v1/) for current implementation work and [docs/speckit/specs-002-cryptoutil/](../speckit/specs-002-cryptoutil/) for detailed specifications.

#### 3.2.1 Secrets Manager (SM) Product (2 Services)

##### 3.2.1.1 Key Management Service (KMS) Service

- Product-Service (Unique Identifier): sm-kms
- Service Name: Key Management Service (KMS)
- Service Description: Elastic key management, encryption-at-rest
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8000-8099
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18000-18099
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28000-28099

#### 3.2.2 SM Instant Messenger (IM) Service (moved from former Cipher product (now SM))

##### 3.2.2.1 Instant Messenger (IM) Service

- Product-Service (Unique Identifier): sm-im (renamed from sm-im) JSON Object Signing & Encryption (JOSE) Product

##### 3.2.3.1 JWK Authority (JA) Service

- Product-Service (Unique Identifier): jose-ja
- Service Name: JWK Authority (JA)
- Service Description: JWK/JWS/JWE/JWT operations
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8800-8899
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18800-18899
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28800-28899

#### 3.2.4 Public Key Infrastructure (PKI) Product

##### 3.2.4.1 Certificate Authority (CA) Service

- Product-Service (Unique Identifier): pki-ca
- Service Name: Certificate Authority (CA)
- Service Description: X.509 certificates, EST, SCEP, OCSP, CRL
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8100-8199
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18100-18199
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28100-28199

#### 3.2.5 Identity Product

##### 3.2.5.1 OAuth 2.1 Authorization Server (Authz) Service

- Product-Service (Unique Identifier): identity-authz
- Service Name: OAuth 2.1 Authorization Server (Authz)
- Service Description: OAuth 2.1 authorization server
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8200-8299
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18200-18299
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28200-28299

##### 3.2.5.2 OIDC 1.0 Identity Provider (IdP) Service

- Product-Service (Unique Identifier): identity-idp
- Service Name: OIDC 1.0 Identity Provider (IdP)
- Service Description: OIDC 1.0 Identity Provider
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8300-8399
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18300-18399
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28300-28399

##### 3.2.5.3 OAuth 2.1 Resource Server (RS) Service

- Product-Service (Unique Identifier): identity-rs
- Service Name: OAuth 2.1 Resource Server (RS)
- Service Description: OAuth 2.1 Resource Server
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8400-8499
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18400-18499
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28400-28499

##### 3.2.5.4 OAuth 2.1 Relying Party (RP) Service

- Product-Service (Unique Identifier): identity-rp
- Service Name: OAuth 2.1 Relying Party (RP)
- Service Description: OAuth 2.1 Relying Party
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8500-8599
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18500-18599
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28500-28599

##### 3.2.5.5 OAuth 2.1 Single Page Application (SPA) Service

- Product-Service (Unique Identifier): identity-spa
- Service Name: OAuth 2.1 Single Page Application (SPA)
- Service Description: OAuth 2.1 Single Page Application
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8600-8699
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18600-18699
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28600-28699

### 3.3 Product-Service Relationships

**Federation Patterns**:

- **Identity ↔ JOSE**: Identity services use JOSE service for JWK/JWT operations
- **All Services ↔ JOSE**: All services may federate to JOSE for cryptographic operations
- **All Services ↔ Identity**: Optional OAuth 2.1 federation for authentication
- **Immediate Failover**: Services attempt credential validators in priority order (no retry logic, no circuit breakers)
  - **FEDERATED unreachable** → fail over to DATABASE and FILE realms
  - **DATABASE unreachable** → fail over to FEDERATED and FILE realms
  - **FEDERATED + DATABASE unreachable** → fail over to FILE realms (CRITICAL failsafe)
- **FILE Realms**: Local to service, always available, MANDATORY minimum 1 FACTOR realm + 1 SESSION realm for admin/DevOps access

**Service Discovery**:

- Configuration file → Docker Compose DNS → Kubernetes DNS
- MUST NOT cache DNS results (for dynamic scaling)

**Cross-Service Authentication**:

- mTLS (preferred) or OAuth 2.1 client credentials
- Federation timeout: Configurable per-service (default: 10s)

### 3.4 Port Assignments & Networking

#### 3.4.1 Port Design Principles

**Container Port Bindings**:

- HTTPS protocol for all public and admin port bindings
- Same HTTPS 127.0.0.1:9090 for Private HTTPS Admin APIs inside Docker Compose and Kubernetes (never localhost due to IPv4 vs IPv6 dual stack issues)
- Same HTTPS 0.0.0.0:8080 for Public HTTPS APIs inside Docker Compose and Kubernetes
- Different HTTPS 127.0.0.1 port range mappings for Public APIs on Docker host (to avoid conflicts)

**Standard Health Check Paths**:

- Same health check paths: `/browser/api/v1/health`, `/service/api/v1/health` on Public HTTPS listeners
- Same health check paths: `/admin/api/v1/livez`, `/admin/api/v1/readyz` on Private HTTPS Admin listeners
- Same graceful shutdown path: `/admin/api/v1/shutdown` on Private HTTPS Admin listeners

**Deployment Type Port Allocation Strategy**:

Three deployment scenarios each use distinct host port ranges to enable concurrent operation:

1. **Service Deployment** (8XXX): Single isolated service
   - Port Range: Service-specific base (e.g., 8100-8199 for pki-ca)
   - Use Case: Independent service development, testing, or production deployment
   - Example: `pki-ca` alone uses host ports 8100-8199

2. **Product Deployment** (18XXX): All services within a product
   - Port Range: Service-specific base + 10000 offset (e.g., 18100-18199 for pki-ca)
   - Use Case: Product-level integration testing, product-only deployments
   - Example: All PKI services (currently only pki-ca) use host ports 18100-18199

3. **Suite Deployment** (28XXX): All services across all products
   - Port Range: Service-specific base + 20000 offset (e.g., 28100-28199 for pki-ca)
   - Use Case: Full system integration, E2E testing, complete production suite
   - Example: All 9 services across 5 products use host ports 28000-28899

**Port Allocation Benefits**:

- No port conflicts between deployment types (all three can run simultaneously)
- Consistent port offsets simplify troubleshooting (service port + offset = deployment type)
- Clear separation enables independent CI/CD pipelines per deployment type

#### 3.4.2 PostgreSQL Ports

| Product-Service Identifier | Address (Host) | Host Port | Container Address | Port Value (Container) |
|---------|-----------|----------------|----------|----------------|
| **pki-ca** | 127.0.0.1 | 54320 | 0.0.0.0 | 5432 |
| **jose-ja** | 127.0.0.1 | 54321 | 0.0.0.0 | 5432 |
| **sm-im** | 127.0.0.1 | 54322 | 0.0.0.0 | 5432 |
| **sm-kms** | 127.0.0.1 | 54323 | 0.0.0.0 | 5432 |
| **identity-authz** | 127.0.0.1 | 54324 | 0.0.0.0 | 5432 |
| **identity-idp** | 127.0.0.1 | 54325 | 0.0.0.0 | 5432 |
| **identity-rs** | 127.0.0.1 | 54326 | 0.0.0.0 | 5432 |
| **identity-rp** | 127.0.0.1 | 54327 | 0.0.0.0 | 5432 |
| **identity-spa** | 127.0.0.1 | 54328 | 0.0.0.0 | 5432 |

#### 3.4.3 Telemetry Ports (Shared)

| Service | Host Port | Port Value (Container) | Protocol |
|---------|-----------|----------------|----------|
| opentelemetry-collector-contrib | 127.0.0.1:4317 | 0.0.0.0:4317 | OTLP gRPC (no TLS) |
| opentelemetry-collector-contrib | 127.0.0.1:4318 | 0.0.0.0:4318 | OTLP HTTP (no TLS) |
| grafana-otel-lgtm | 127.0.0.1:3000 | 0.0.0.0:3000 | HTTP (no TLS) |
| grafana-otel-lgtm | 127.0.0.1:4317 | 0.0.0.0:4317 | OTLP gRPC (no TLS) |
| grafana-otel-lgtm | 127.0.0.1:4318 | 0.0.0.0:4318 | OTLP HTTP (no TLS) |

---

## 4. System Architecture

### 4.1 System Context

**cryptoutil** operates as a suite of independent microservices with optional federation:

- **Architectural Deployment Models**: Standalone services, federated suite, Kubernetes pods
- **Docker Compose Deployment Types**: Service (single service), Product (all services in one product), Suite (all services across all products) - see section 3.4.1 for port allocation strategy
- **External Dependencies**: PostgreSQL (production), SQLite (dev/test), OpenTelemetry Collector, Grafana OTEL-LGTM
- **Client Types**: Browser clients (session-based), service clients (token-based), CLI tools
- **Integration Points**: REST APIs, Docker secrets, configuration files, health endpoints

### 4.2 Container Architecture

**Container Strategy**:

- **Base Image**: Alpine Linux (minimal attack surface)
- **Multi-Stage Builds**: Builder → Validator → Runtime (secrets validation mandatory)
- **Runtime User**: Non-root (security best practice)
- **Health Checks**: Integrated liveness and readiness probes
- **Secret Management**: Docker/Kubernetes secrets mounted at /run/secrets/
- **Network Isolation**: Service-specific networks, admin endpoints localhost-only

**Docker Compose Patterns**:

- Single build, shared image (prevents 3× build time)
- Health check dependencies (service_healthy, not service_started)
- Latency hiding: First instance initializes DB, others wait

### 4.3 Component Architecture

#### 4.3.1 Layered Architecture

- main() `cmd/` → Application `internal/*/application/` → Business Logic `internal/*/service/`, `internal/*/domain/` → Repositories `internal/*/repository/` → Database/External Systems
- Dependency flow: One-way only (top → bottom)
- Cross-cutting concerns: Telemetry, logging, error handling

#### 4.3.2 Dependency Injection

- Constructor injection pattern: NewService(logger, repo, config)
- Factory pattern: *FromSettings functions for configuration-driven initialization
- Context propagation: Pass context.Context to all long-running operations

### 4.4 Code Organization

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

#### 4.4.3 CLI Entry Points

```
cmd/
├── cryptoutil/main.go         # Suite-level CLI (all products): Thin main() call to `internal/apps/cryptoutil.go`
# SM product removed - im moved to sm product
├── jose/main.go               # Product-level JOSE CLI: Thin main() call to `internal/apps/jose/jose.go`
├── pki/main.go                # Product-level PKI CLI: Thin main() call to `internal/apps/pki/pki.go`
├── identity/main.go           # Product-level Identity CLI: Thin main() call to `internal/apps/identity/identity.go`
├── sm/main.go                 # Product-level SM CLI: Thin main() call to `internal/apps/sm/sm.go`
├── sm-im/main.go            # Service-level SM-IM CLI: Thin main() call to `internal/apps/sm/im/im.go`
├── jose-ja/main.go            # Service-level JOSE-JA CLI: Thin main() call to `internal/apps/jose/ja/ja.go`
├── pki-ca/main.go             # Service-level PKI-CA CLI: Thin main() call to `internal/apps/pki/ca/ca.go`
├── identity-authz/main.go     # Service-level Identity-Authz CLI: Thin main() call to `internal/apps/identity/authz/authz.go`
├── identity-idp/main.go       # Service-level Identity-IDP CLI: Thin main() call to `internal/apps/identity/idp/idp.go`
├── identity-rp/main.go        # Service-level Identity-RP CLI: Thin main() call to `internal/apps/identity/rp/rp.go`
├── identity-rs/main.go        # Service-level Identity-RS CLI: Thin main() call to `internal/apps/identity/rs/rs.go`
├── identity-spa/main.go       # Service-level Identity-SPA CLI: Thin main() call to `internal/apps/identity/spa/spa.go`
└── sm-kms/main.go             # Service-level SM-KMS CLI (legacy): Thin main() call to `internal/apps/sm/kms/kms.go`
```

**Pattern**: Thin `main()` pattern for all cmd/ CLIs, with all logic in `internal/apps/` for maximum code reuse and testability.

1. `cmd/cryptoutil/` for suite-level CLI
```go
func main() {
    os.Exit(cryptoutilAppsSuite.Suite(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
```
1. `cmd/<product>/` for product-level CLI
```go
func main() {
    os.Exit(cryptoutilApps<PRODUCT>.<PRODUCT>(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
```
1. `cmd/<product>/<service>/` for service-level CLI
```go
func main() {
    os.Exit(cryptoutilApps<PRODUCT><SERVICE>.<SERVICE>(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
```

#### 4.4.4 Service Implementations

```
internal/apps/
├── template/                  # REUSABLE product-service template (all 9 services for all 5 products MUST reuse this template for maximum consistency and minimum duplication)
│   ├── service/
│   │   ├── config/            # ServiceTemplateServerSettings
│   │   ├── server/            # Application, PublicServerBase, AdminServer
│   │   │   ├── application/   # ApplicationCore, ApplicationBasic
│   │   │   ├── builder/       # ServerBuilder fluent API
│   │   │   ├── listener/      # AdminHTTPServer
│   │   │   ├── barrier/       # Encryption-at-rest service
│   │   │   ├── businesslogic/ # SessionManager, TenantRegistration
│   │   │   ├── repository/    # TenantRepo, RealmRepo, SessionRepo
│   │   │   └── realms/        # Authentication realm implementations
│   │   └── testutil/          # Test helpers (NewTestSettings)
│   └── testing/
│       └── e2e/               # ComposeManager for E2E orchestration

│   └── im/                    # SM-IM service
│       ├── domain/            # Domain models (Message, Recipient)
│       ├── repository/        # Domain repos + migrations (2001+)
│       ├── server/            # SMIMServer, PublicServer
│       │   ├── config/        # SMImServerSettings embeds template
│       │   └── apis/          # HTTP handlers
│       ├── client/            # API client
│       ├── e2e/               # E2E tests (Docker Compose)
│       └── integration/       # Integration tests
├── jose/
│   └── ja/                    # JOSE-JA service (same structure)
├── pki/
│   └── ca/                    # PKI-CA service (same structure)
├── sm/
│   └── jose/                  # SM-KMS service (same structure)
└── identity/
    ├── authz/                 # OAuth 2.1 Authorization Server (same structure)
    ├── idp/                   # OIDC 1.0 Identity Provider (same structure)
    ├── rs/                    # OAuth 2.1 Resource Server (same structure)
    ├── rp/                    # OAuth 2.1 Relying Party (same structure)
    └── spa/                   # OAuth 2.1 Single Page Application (same structure)
```

#### 4.4.5 Shared Utilities

```
internal/shared/
├── apperr/                  # Application errors
├── container/               # Dependency injection container
├── config/                  # Configuration helpers
├── crypto/                  # Cryptographic utilities
├── magic/                   # Named constants (ports, timeouts, paths)
├── pool/                    # Generator pool utilities
├── pwdgen/                  # Password generator utilities
├── telemetry/               # OpenTelemetry integration
└── testutil/                # Shared test utilities
```

#### 4.4.6 Docker Compose

```
deployments/
├── telemetry/
│   └── compose.yml
├── sm-kms/
│   ├── config/
|   │   ├── common.yml        # common configuration for all 3 sm-kms instances
|   │   ├── postgresql-1.yml  # instance 1 of sm-kms; uses shared sm-kms PostgreSQL
|   │   ├── postgresql-2.yml  # instance 2 of sm-kms; uses shared sm-kms PostgreSQL
|   │   └── sqlite.yml        # instance 3 of sm-kms; uses non-shared in-memory sm-kms SQLite
│   ├── secrets/
|   │   ├──postgres_url.secret      # Docker Compose secret shared by 2 instances of sm-kms; PostgreSQL instances only
|   │   ├──postgres_database.secret # Docker Compose secret shared by 2 instances of sm-kms; PostgreSQL instances only
|   │   ├──postgres_username.secret # Docker Compose secret shared by 2 instances of sm-kms; PostgreSQL instances only
|   │   ├──postgres_password.secret # Docker Compose secret shared by 2 instances of sm-kms; PostgreSQL instances only
|   │   ├──unseal_1of5.secret       # Docker Compose secret shared by 3 instances of sm-kms; unseal service
|   │   ├──unseal_2of5.secret       # Docker Compose secret shared by 3 instances of sm-kms; unseal service
|   │   ├──unseal_3of5.secret       # Docker Compose secret shared by 3 instances of sm-kms; unseal service
|   │   ├──unseal_4of5.secret       # Docker Compose secret shared by 3 instances of sm-kms; unseal service
|   │   ├──unseal_5of5.secret       # Docker Compose secret shared by 3 instances of sm-kms; unseal service
|   │   └──hash_pepper.secret       # Docker Compose secret shared by 3 instances of sm-kms; hash registries of hash algorithms
│   ├── compose.yml                 # Docker Compose config: `builder-cryptoutil` builds Dockerfile, 3 instances of sm-kms depend on it
│   └── Dockerfile                  # Dockerfile: compose.yml `builder-cryptoutil` builds this Dockerfile
├── <PRODUCT>/
│   └── ... (same structure)
├── jose/
│   └── ... (same structure)
├── ca/
│   └── ... (same structure)
├── identity/
│   └── ... (same structure)

    └── ... (same structure)
```

#### 4.4.7 CLI Patterns

### CLI Hierarchy

```
# Product-Service pattern (preferred)
sm-im server --config=/etc/sm/im.yml

# Service pattern
im server --config=/etc/sm/im.yml

# Product pattern (routes to service)
sm im server --config=/etc/sm/im.yml

# Suite pattern (routes to product, then service)
cryptoutil sm im server --config=/etc/sm/im.yml
```

```
# Product-Service pattern (preferred)
jose-ja server --config=/etc/jose/ja.yml

# Service pattern
ja server --config=/etc/jose/ja.yml

# Product pattern (routes to service)
jose ja server --config=/etc/jose/ja.yml

# Suite pattern (routes to product, then service)
cryptoutil jose ja server --config=/etc/jose/ja.yml
```

```
# Product-Service pattern (preferred)
sm-kms server --config=/etc/sm/kms.yml

# Service pattern
kms server --config=/etc/sm/kms.yml

# Product pattern (routes to service)
sm kms server --config=/etc/sm/kms.yml

# Suite pattern (routes to product, then service)
cryptoutil sm kms server --config=/etc/sm/kms.yml
```

### CLI Subcommand

All CLIs for all 9 services MUST support these subcommands, with consistent behavior and config parsing and flag parsing.
Consistency MUST be guaranteed by inheriting from service-template, which will reuse `internal/apps/template/service/<SUBCOMMAND>/` packages:

| Subcommand | Description |
|------------|-------------|
| `server` | CLI server start with dual HTTPS listeners, for Private Admin Compose+K8s APIs vs Public Business Logic APIs |
| `health` | CLI client for Public health endpoint API check |
| `livez` | CLI client for Private liveness endpoint API check |
| `readyz` | CLI client for Private readiness endpoint API check |
| `shutdown` | CLI client for Private graceful shutdown endpoint API trigger |
| `client` | CLI client for Business Logic API interaction (n.b. domain-specific for each of the 9 services) |
| `init` | CLI client for Initialize static config, like TLS certificates |
| `demo` | CLI client for start server, inject Demo data, and run clients |

---

## 5. Service Architecture

### 5.1 Service Template Pattern

#### 5.1.1 Template Components

- Two HTTPS Listeners: Public (business APIs) + Admin (health checks)
- Two Public Paths: `/browser/**` (sessions) vs `/service/**` (tokens)
- Three Admin APIs: /admin/api/v1/livez, /admin/api/v1/readyz, /admin/api/v1/shutdown
- Database: PostgreSQL || SQLite with GORM
- Telemetry: OTLP → otel-collector-contrib → Grafana LGTM
- Config Priority: Docker secrets > YAML > CLI parameters (NO environment variables)

#### 5.1.2 Template Benefits

- Eliminates 48,000+ lines of boilerplate per service
- Consistent infrastructure across all 9 services
- Proven patterns: TLS setup, middleware stacks, health checks, graceful shutdown
- Parameterization: OpenAPI specs, handlers, middleware chains injected via constructor

#### 5.1.3 Mandatory Usage

- ALL new services MUST use `internal/apps/template/service/` (consistency, reduced duplication)
- ALL existing services MUST be refactored to use `internal/apps/template/service/` (iterative migration)
- Migration priority: sm-im → jose-ja → sm-kms → pki-ca → identity services
  - sm-im/jose-ja/sm-kms migrate first (SM product); pki-ca second; identity last

### 5.2 Service Builder Pattern

#### 5.2.1 Builder Methods

- NewServerBuilder(ctx, cfg): Create builder with `internal/apps/template/service/` config
- WithDomainMigrations(fs, path): Register domain migrations (2001+)
- WithPublicRouteRegistration(func): Register domain-specific public routes
- Build(): Construct complete infrastructure and return ServiceResources

#### 5.2.2 Merged Migrations

*See section 7.4 Migration Strategy for details on the merged migrations pattern.*

#### 5.2.3 ServiceResources

- Returns initialized infrastructure: DB (GORM), TelemetryService, JWKGenService, BarrierService, UnsealKeysService, SessionManager, RealmService, Application
- Shutdown functions: ShutdownCore(), ShutdownContainer()
- Domain code receives all dependencies ready-to-use

#### 5.2.4 Database Compatibility Rules

##### Cross-DB Compatibility Rules

```go
// UUID fields: ALWAYS type:text (SQLite has no native UUID)
ID googleUuid.UUID `gorm:"type:text;primaryKey"`

// Nullable UUIDs: Use NullableUUID (NOT *googleUuid.UUID)
ClientProfileID NullableUUID `gorm:"type:text;index"`

// JSON arrays: ALWAYS serializer:json (NOT type:json)
AllowedScopes []string `gorm:"serializer:json"`
```

##### SQLite Configuration

```go
sqlDB.Exec("PRAGMA journal_mode=WAL;")       // Concurrent reads + 1 writer
sqlDB.Exec("PRAGMA busy_timeout = 30000;")   // 30s retry on lock
sqlDB.SetMaxOpenConns(5)                     // GORM transactions need multiple
```

##### SQLite DateTime (CRITICAL)

**ALWAYS use `.UTC()` when comparing with SQLite timestamps**:

```go
// ❌ WRONG: time.Now() without .UTC()
if session.CreatedAt.After(time.Now()) { ... }

// ✅ CORRECT: Always use .UTC()
if session.CreatedAt.After(time.Now().UTC()) { ... }
```

**Pre-commit hook auto-converts** `time.Now()` → `time.Now().UTC()`.

### 5.3 Dual HTTPS Endpoint Pattern

#### 5.3.1 Public HTTPS Endpoint

- Purpose: Business APIs, browser UIs, external client access
- Default Binding: 127.0.0.1 (dev/test), 0.0.0.0 (containers)
- Default Port: Service-specific ranges (8080-8089 KMS, 8100-8149 Identity, etc.)
- Request Paths: `/service/**` (headless clients) and `/browser/**` (browser clients)

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

- SAME OpenAPI Specification served at both `/service/**` and `/browser/**` paths
- API contracts identical, only middleware/authentication differ
- Middleware enforces authorization mutual exclusivity (headless → /service/**, browser → /browser/**)
- E2E tests MUST verify BOTH path prefixes

### 5.5 Health Check Patterns

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

**CRITICAL: FIPS 140-3 mode is ALWAYS enabled by default and MUST NEVER be disabled**

- **Approved Algorithms**: RSA ≥2048, ECDSA (P-256/384/521), ECDH, EdDSA (25519/448), AES ≥128 (GCM, CBC+HMAC), SHA-256/384/512, HMAC-SHA256/384/512, PBKDF2, HKDF
- **BANNED Algorithms**: bcrypt, scrypt, Argon2, MD5, SHA-1, RSA <2048, DES, 3DES
- **Algorithm Agility**: All cryptographic operations support configurable algorithms with FIPS-approved defaults
- **Compliance Validation**: Automated tests verify only FIPS-approved algorithms used

### 6.2 SDLC Security Strategy

**Security Gates**:

- **Pre-Commit**: golangci-lint (including gosec), UTF-8 validation, auto-formatters
- **CI/CD**: SAST (gosec), dependency scanning (govulncheck), secret detection, DAST (Nuclei, OWASP ZAP)
- **Testing**: Security test cases, fuzzing (15s minimum), mutation testing (≥95% production, ≥98% infrastructure)

**Vulnerability Management**:

- Weekly `govulncheck ./...` execution
- Sources: <https://pkg.go.dev/vuln/list>, GitHub Advisories, CVE Details
- Incremental updates with testing before deployment

**Code Review**:

- Security-focused reviews for all cryptographic changes
- Mandatory for changes to authentication, authorization, key management
- Peer review for all production changes

### 6.3 Product Security Strategy

**Defense in Depth**:

- **Multi-Layer Keys**: Unseal → Root → Intermediate → Content (hierarchical encryption)
- **Network Security**: IP allowlisting + per-IP rate limiting + CORS + CSRF + CSP
- **Secret Management**: Docker/Kubernetes secrets (NEVER environment variables)
- **TLS Everywhere**: TLS 1.3+ with full certificate chain validation
- **Audit Logging**: All security events, 90-day retention minimum

**Secure Defaults**:

- FIPS 140-3 enabled (cannot disable)
- TLS required for all endpoints
- IP allowlisting enabled by default
- Session timeouts: 30 minutes with MFA step-up
- Auto-lock on multiple failed authentication attempts

**Zero Trust**:

- No caching of authorization decisions (always re-evaluate)
- Mutual TLS for service-to-service communication
- Least privilege principle for all operations
- Timeout configuration mandatory for all network operations

### 6.4 Cryptographic Architecture

#### 6.4.1 FIPS 140-3 Compliance (ALWAYS Enabled)

**Approved Algorithms**:

| Category | Algorithms |
|----------|------------|
| Asymmetric | RSA ≥2048, DH ≥2048, ECDSA P256/P384/P521, ECDH P256/P384/P521, EdDSA 25519/448, EdDH X25519/X448 |
| Symmetric | AES-128/192/256 (GCM, CBC+HMAC, CMAC) |
| Digest | SHA-256/384/512, HMAC-SHA-256/384/512 |
| KDF | PBKDF2-HMAC-SHA256/384/512, HKDF-SHA256/384/512 |

**Banned**: bcrypt, scrypt, Argon2, MD5, SHA-1, RSA <2048, DH <2048, EC < P256, DES, 3DES.

#### 6.4.2 Key Hierarchy (Barrier Service)

- Unseal keys (Docker secrets, NEVER stored in database)
- Root keys (encrypted-at-rest with unseal keys)
- Intermediate keys (encrypted-at-rest with root keys)
- Content keys (encrypted-at-rest with intermediate keys)
- Domain data encryption (encrypted-at-rest with content keys)

```
Unseal Key (Docker secrets, NEVER stored)
    └── Root Key (encrypted-at-rest with unseal key(s), rotated manually or automatically annually)
        └── Intermediate Key (encrypted-at-rest with root key, rotated manually or automatically quarterly)
            └── Content Key (encrypted-at-rest with intermediate key, rotated manually or automatically monthly)
                └── Domain Data (encrypted-at-rest with content key) - Examples: SM-IM messages, SM-KMS JWKs, JOSE-JA JWKs, PKI-CA private keys, Identity user credentials
```

Design Intent: Unseal secret(s) or unseal key(s) are loaded by service instances at startup. To decrypt and reuse existing, sealed root keys in a database, each service instance MUST use unseal credentials to unseal the root keys. This is design intent for barrier service.

#### 6.4.3 Hash Service (Version-Based)

Hash service supports 4 hash types.
1. Low-entropy, random-salt => Used for short values that DON'T need to be indexed or searched in a database (e.g. Passwords)
2. Low-entropy, fixed-salt => Used for short values that DO need to be indexed and searched in a database (e.g. PII, Usernames, Emails, Addresses, Phone Numbers, SIN/SSN/NIN, IPs, MACs)
3. High-entropy, random-salt => Used for long values that DON'T need to be indexed or searched in a database (e.g. Private Keys)
4. High-entropy, fixed-salt => Used for long values that DO need to be indexed and searched in a database, inputs MUST have a minimum of 256-bits (32-bytes) of entropy

##### Low-entropy vs High-entropy

Low entropy: Values with >= 256-bits (32-bytes) or higher of brute-force search space; values are hashed with high-iterations PBKDF2 to mitigate brute-force attacks, because small search spaces are not big enough to mitigate brute-force attacks on their own; do not use HKDF, it does not add sufficient security for low-entropy values

High entropy: Values with < 256-bits (32-bytes) of brute-force search space; values are hashed with one-iteration HKDF, because large search space is big enough to mitigate brute-force attacks on its own; do not use PBKDF2, extra iterations do not add meaningful security

##### Random-salt vs Fixed-salt

Random salt: Used for values that DON'T require indexing or searching in a database; non-deterministic hash outputs for the same input is best practice for security

Fixed-salt: Used for values that DO require indexing or searching in a database; deterministic hash outputs for the same input are required for indexing and searching, which overrides best practice for security; to mitigate reduced security of using fixed-salt, pepper MUST be applied to all values before passing them into hash functions

##### Pepper

Pepper MUST be used on all values passed into hash functions that use fixed-salt.
Pepper SHOULD be used on all values passed into hash functions that use random-salt
For consistency, pepper usage WILL be used on all values passed to all hash functions, regardless of salt type.

Pepper before deterministic hashing MUST use AES-GCM-SIV.
Pepper before non-deterministic hashing MUST use AES-GCM-SIV or AES-GCM. The AES-256 key MUST be generated and used for the lifetime of the hash.

##### Low-Entropy Hash Format

```
Format: {pepperTypeAndVersion}base64(optionalPepperNonce):base64(optionalPepperAAD)#{hashTypeAndVersion}:{algorithm}:{iterations}:base64(salt):base64(hash)
Deterministic Example:     {d2}#{f5}:PBKDF2-HMAC-SHA256:600000:abc123...:def456...
Non-Deterministic Example: {n2}nonce#{f5}PBKDF2-HMAC-SHA256:600000:abc123...:def456...
Non-Deterministic Example: {n2}nonce:aad#{f5}PBKDF2-HMAC-SHA256:600000:abc123...:def456...
```

##### High-Entropy Hash Format

```
Format: {pepperTypeAndVersion}base64(optionalPepperNonce):base64(optionalPepperAAD)#{hashTypeAndVersion}:{algorithm}:base64(salt):base64(info):base64(hash)
Deterministic Example:     {d2}#{F5}:HKDF-HMAC-SHA256:abc123...:def456...:ghi789...
Non-Deterministic Example: {n2}nonce#{R5}HKDF-HMAC-SHA256:abc123...:def456...:ghi789...
Non-Deterministic Example: {n2}nonce:aad#{R5}HKDF-HMAC-SHA256:abc123...:def456...:ghi789...
```

#### 6.4.4 Unseal Modes

- Simple keys: File-based unseal key loading
- Shared secrets: M-of-N Shamir secret sharing (e.g., 3-of-5)
- System fingerprinting: Device-specific unseal key derivation
- High availability patterns for multi-instance deployments

#### 6.4.5 Key Rotation Strategies

- Root keys: Annual rotation (manual or automatic)
- Intermediate keys: Quarterly rotation
- Content keys: Monthly or per-operation rotation
- Elastic key pattern: Active key + historical keys for decryption

### 6.5 PKI Architecture & Strategy

**CA/Browser Forum Baseline Requirements**:

- **Serial Number**: ≥64 bits CSPRNG, non-sequential, >0, <2^159
- **Algorithms**: RSA ≥2048, ECDSA P-256/384/521, EdDSA, SHA-256/384/512 (NEVER MD5/SHA-1)
- **Validity**: Subscriber certs ≤398 days, Intermediate CA 5-10 years, Root CA 20-25 years
- **Extensions**: Key Usage (critical), EKU, SAN, AKI, SKI, CRL Distribution Points, OCSP
- **CRL/OCSP**: Update ≤7 days, OCSP response ≤7-10 days validity
- **Audit Logging**: 7-year retention minimum

**CA Architecture Patterns** (highest to lowest preference):

1. **Offline Root → Online Root → Issuing CA** (Maximum security)
2. **Online Root → Issuing CA** (Balanced)
3. **Online Root** (Simple, dev/test acceptable)

**Certificate Lifecycle**:

- **Issuance**: CSR validation, identity verification, generation, signing, publication
- **Renewal**: Pre-expiration notification (60/30/7 days), re-validation if required
- **Revocation**: Request validation, CRL/OCSP update, notification

### 6.6 JOSE Architecture & Strategy

**Elastic Key Rotation** (per-message):

- **Active Key**: Current key for signing/encrypting (new JWK per message)
- **Historical Keys**: Previous keys for verifying/decrypting (preserved indefinitely)
- **Key ID Embedding**: Ciphertext/signature includes key ID for deterministic lookup
- **Rotation Trigger**: Per-message, hourly, or on-demand

**JOSE Operations**:

- **JWK Generation**: RSA, EC, ED, symmetric keys with configurable algorithms
- **JWS**: Signing with RS256/384/512, ES256/384/512, EdDSA
- **JWE**: Encryption with RSA-OAEP, ECDH-ES, A128/192/256GCM
- **JWT**: Access tokens (JWS), refresh tokens (opaque), ID tokens (JWS)

**Storage Pattern**:

- Domain-specific JWK tables (encrypted with Barrier service)
- Key versioning for rotation support
- Per-tenant key isolation

### 6.7 Key Management System Architecture

**Hierarchical Key Structure**:

- **Unseal Keys**: Never stored in app, provided via Docker secrets (5-of-5 required)
- **Root Keys**: Encrypted with unseal keys, rotated annually
- **Intermediate Keys**: Encrypted with root keys, rotated quarterly
- **Content Keys**: Encrypted with intermediate keys, rotated per-operation or hourly

**Barrier Service**:

- HKDF-based deterministic key derivation (instances with same unseal secrets derive same keys)
- AES-256-GCM encryption at rest
- Multi-layer protection cascade

**Elastic Key Management**:

- Active key for encrypt/sign operations
- Historical keys for decrypt/verify operations
- Key ID embedded in ciphertext for deterministic lookup
- Lazy migration on rotation (re-encrypt on next write)

### 6.8 Multi-Factor Authentication Strategy

**MFA Methods**:

- **Time-Based**: TOTP (Google Authenticator, Authy)
- **Event-Based**: HOTP (YubiKey, RSA SecurID)
- **Biometric**: WebAuthn with Passkeys (Face ID, Touch ID, Windows Hello)
- **Hardware**: WebAuthn security keys (YubiKey, Titan Key)
- **Push**: Mobile app push notifications
- **OTP**: Email, SMS, phone call one-time passwords
- **Recovery**: Backup single-use codes

**Step-Up Authentication**:

- Re-authentication MANDATORY every 30 minutes for high-sensitivity operations
- MFA enrollment optional during setup, access limited until enrolled
- Adaptive based on risk (IP change, unusual access patterns)

**Common Combinations**:

- Password + TOTP/WebAuthn/Push (browser clients)
- Client ID/Secret + mTLS/Bearer (service clients)

### 6.9 Authentication & Authorization

**Authentication Methods**: 41 total (13 headless + 28 browser)
**Authorization**: Scope-based, RBAC, resource-level ACLs, zero-trust (no caching)

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

### 6.10 Secrets Detection Strategy

**Purpose**: Detect inline secrets in deployment compose files to enforce Docker secrets usage.

**Detection Approach**: Length-based threshold (≥32 bytes raw, ≥43 characters base64-encoded) identifies high-entropy inline values in environment variables matching secret-pattern names (PASSWORD, SECRET, TOKEN, KEY, API_KEY). No entropy calculation is used - it produces too many false positives on non-secret configuration values.

**Safe References** (excluded from detection): Docker secret paths (`/run/secrets/`), short development defaults (< threshold), empty values, variable references (`${VAR}`).

**Trade-offs**: Length threshold catches most real secrets (UUIDs, tokens, hashes) while allowing short developer passwords (`admin`, `dev123`). Infrastructure deployments (Grafana, OTLP collector) are excluded since they intentionally use inline dev credentials.

**Cross-References**: Implementation in [validate_secrets.go](/internal/cmd/cicd/lint_deployments/validate_secrets.go). Deployment secrets management in [Section 12.6](#126-secrets-management-in-deployments).

---

## 7. Data Architecture

### 7.1 Database Schema Patterns

**MANDATORY: Schema-Level Isolation ONLY**

- Each tenant gets separate schema: `tenant_<uuid>.users`, `tenant_<uuid>.sessions`
- NEVER use row-level multi-tenancy (single schema, tenant_id column)
- Reason: Data isolation, compliance, performance (per-tenant indexes)
- Pattern: Set `search_path` on connection: `SET search_path TO tenant_abc123`

**Database Query Pattern:**

```go
// ✅ CORRECT: Schema isolation (separate schemas per tenant)
db.Where("user_id = ?", userID).Find(&messages)

// ❌ WRONG: Row-level isolation (single schema, tenant_id column)
db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).Find(&messages)
```

### 7.1.1 Database Isolation for Microservices

**MANDATORY: Each Service MUST Have Isolated Database Storage**

**Rationale**: Microservice architecture principles require independent data stores to ensure:
1. Service autonomy and independent scaling
2. Failure isolation (one service's DB issues don't affect others)
3. Technology flexibility (each service can choose optimal storage)
4. Clear ownership boundaries and access control

**Requirements** (enforced by linter):
- **Unique Database Name**: Each of 9 services MUST have unique `postgres_database.secret`
  - Example: `authz_db`, `idp_db`, `rp_db`, `rs_db`, `spa_db` (NOT shared `identity_db`)
- **Unique Username**: Each service MUST have unique `postgres_username.secret`
  - Example: `authz_user`, `idp_user`, `rp_user`, `rs_user`, `spa_user`
- **Unique Password**: Each service MUST have unique `postgres_password.secret`
- **Unique Connection URL**: Each service MUST have unique `postgres_url.secret`

**Linter Enforcement** (`cicd lint-deployments`):
- Scans ALL 9 service directories for database credential secrets
- ERRORS on duplicate database names or usernames across services
- Validates credentials are isolated regardless of deployment level (SUITE/PRODUCT/SERVICE)

**Exception**: Leader-follower PostgreSQL replication where logical schemas are replicated but services maintain separate schema namespaces within the same physical server.

**Cross-Service Communication**: Services needing data from other services MUST use REST APIs, NEVER direct database access.

### 7.2 Multi-Tenancy Architecture & Strategy

#### 7.2.1 Authentication Realms

**CRITICAL**: Realms define authentication METHOD and POLICY, NOT data scoping.

**Realms do NOT scope data** - all realms in same tenant see same data. Only `tenant_id` scopes data access.

#### Authentication Realm Types

| Realm Type | Purpose | Scheme | Credential | Credential Validators |
|------|--------|------------|-------------------------|-----------------|
| `https-client-cert-factor` | Create or Upgrade Session | HTTP/mTLS Handshake | HTTPS Client Certificate | File, Database, Federated |
| `webauthn-resident-synced-factor` | Create or Upgrade Session | WebAuthn L2 Resident Synced (aka Passkeys) | Local PublicKeyCredential | File, Database, Federated |
| `webauthn-resident-unsynced-factor` | Create or Upgrade Session | WebAuthn L2 Resident Unsynced (e.g. Windows Hello) | Local PublicKeyCredential | File, Database, Federated |
| `webauthn-nonresident-synced-factor` | Create or Upgrade Session | WebAuthn L2 Non-Resident Synced (e.g. Azure AD) | Cloud PublicKeyCredential | File, Database, Federated |
| `webauthn-nonresident-unsynced-factor` | Create or Upgrade Session | WebAuthn L2 Non-Resident Unsynced (e.g. YubiKey) | Cloud PublicKeyCredential | File, Database, Federated |
| `authorization-code-opaque-factor` | Create or Upgrade Session | OAuth 2.1 Authorization Code Flow + PKCE | Opaque | File, Database, Federated |
| `authorization-code-jwe-factor` | Create or Upgrade Session | OAuth 2.1 Authorization Code Flow + PKCE | JWE | File, Database, Federated |
| `authorization-code-jws-factor` | Create or Upgrade Session | OAuth 2.1 Authorization Code Flow + PKCE | JWS | File, Database, Federated |
| `bearer-token-opaque-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Bearer | Opaque Token | File, Database, Federated |
| `bearer-token-jwe-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Bearer | JWE Token | File, Database, Federated |
| `bearer-token-jws-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Bearer | JWS Token | File, Database, Federated |
| `basic-username-password-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic | Username/Password | File, Database, Federated |
| `basic-email-password-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic | Email/Password | File, Database, Federated |
| `basic-email-otp-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic | Email/RandomOTP | File, Database, Federated |
| `basic-email-magiclink-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic + Query String | Email/Nothing & QueryParameter | File, Database, Federated |
| `basic-sms-password-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic | Phone/Password | File, Database, Federated |
| `basic-sms-otp-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic | Phone/RandomOTP | File, Database, Federated |
| `basic-sms-magiclink-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic + Query String | Phone/Nothing & QueryParameter | File, Database, Federated |
| `basic-voice-otp-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic | Phone/RandomOTP | File, Database, Federated |
| `basic-id-otp-factor` | Create or Upgrade Session | HTTP Header 'Authorize' Basic | ID/Nothing & HOTP/TOTP | File, Database, Federated |
| `cookie-token-opaque-session` | Use Session | HTTP Header 'Cookie' Token | Opaque Token | File, Database, Federated |
| `cookie-token-jwe-session` | Use Session | HTTP Header 'Cookie' Token | JWE Token | File, Database, Federated |
| `cookie-token-jws-session` | Use Session | HTTP Header 'Cookie' Token | JWS Token | File, Database, Federated |

### Authentication Realm Principals

1. Every service MUST configure a prioritized list of realm instances; multiple realm instances of same realm type are allowed.
2. Every service MUST configure one or more factor realms, for creating or upgrading sessions; zero factor realms is NOT allowed.
3. Every service MUST configure one or more session realms, for using sessions; zero session realms is NOT allowed.
4. Every realm instance MUST specify one-and-only-one credential validator; the only valid credential validator options are file-backed, database-backed, or federated.
5. Every factor realm instance MUST return a created or rotated session cookie on successful authentication.
6. Every session realm instance MAY return a rotated session cookie on successful authentication; mitigates session fixation.
7. Every service MUST include at least one FILE-based factor realm for fallback session creation, plus at least one FILE-based session realm for session use. FILE realms are CRITICAL failsafes - local to the service, always available, ensuring admin/DevOps can always access the service.
8. **Multi-Level Failover Pattern**: Services attempt credential validators in configured priority order (no circuit breakers, no retry logic):
   - **FEDERATED unreachable** → services continue with DATABASE and FILE realms
   - **DATABASE unreachable** → services continue with FEDERATED and FILE realms
   - **FEDERATED + DATABASE unreachable** → services continue with FILE realms (CRITICAL failsafe)
   - FILE realms provide disaster recovery / high availability guarantees for administrative access even when all external dependencies are unavailable.

### 7.3 Dual Database Strategy

All 9 services MUST support using one of PostgreSQL or SQLite, specified via configuration at startup.

Typical usages for each database for different purposes:
- Unit tests, Fuzz tests, Benchmark tests, Mutations tests => Ephemeral SQLite instance (e.g. in-memory)
- Integration tests, Load tests => Ephemeral PostgreSQL instance (i.e. test-container)
- End-to-End tests => Static PostgreSQL instance (e.g. Docker Compose)
- Production => Static PostgreSQL instance (e.g. Cloud hosted)
- Local Development => Static SQLite instance (e.g. file); used for local development

Caveat: End-to-End Docker Compose tests use both PostgreSQL and SQLite, for isolation testing; 3 service instances, 2 using a shared PostgreSQL container, and 1 using in-memory SQLite

### 7.4 Migration Strategy

#### Merged Migrations Pattern

| Range | Owner | Examples |
|-------|-------|----------|
| 1001-1999 | Service Template | Sessions (1001), Barrier (1002), Realms (1003), Tenants (1004), PendingUsers (1005) |
| 2001+ | Domain | sm-im messages (2001), jose JWKs (2001) |

- mergedMigrations type: Implements fs.FS interface, unifies both for golang-migrate validation
- Prevention: Solves "no migration found for version X" validation errors

**Migration Process**:
1. Extract shared infrastructure migrations (1001-1999) to service template
2. Domain services start migrations at 2001+ (never conflicts)
3. Use mergedMigrations for unified validation during service initialization

### 7.5 Data Security & Encryption

cryptoutil ensures enterprise-grade data security through:

- **FIPS 140-3 Compliance**: NIST-approved algorithms only (see [6.4.1](#641-fips-140-3-compliance-always-enabled))
- **Hierarchical Key Management**: Unseal → Root → Intermediate → Content keys (see [6.4.2](#642-key-hierarchy-barrier-service))
- **Version-Based Hashing**: Entropy-based algorithm selection (see [6.4.3](#643-hash-service-version-based))
- **Encryption at Rest/In Transit**: AES-GCM and TLS 1.3+ with full validation
- **Zero-Trust Authorization**: No caching; re-evaluation per request

Implemented via Barrier Service, Hash Service, PKI-CA, and audit logging.

---

## 8. API Architecture

### 8.1 OpenAPI-First Design

cryptoutil follows an OpenAPI-first design approach, ensuring all APIs are defined in OpenAPI 3.0.3 specifications before implementation. This enables type-safe code generation, consistent validation, and automatic documentation.

**Version**: OpenAPI 3.0.3 | **Structure**: components.yaml + paths.yaml | **Gen**: oapi-codegen strict-server | **Content**: application/json only

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

#### 8.3.1 Versioning Strategy

- Path-based versioning: /api/v1/, /api/v2/
- N-1 backward compatibility (current + previous version)
- Major version for breaking changes, minor version within same major

#### 8.3.2 Version Lifecycle

- Deprecation warning period: 6+ months
- Documentation of migration path
- Parallel operation of N and N-1 versions

### 8.4 Error Handling

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

#### 9.2.1 Configuration Priority Order

1. **Docker Secrets** (`file:///run/secrets/secret_name`) - Sensitive values
2. **YAML Configuration** (`--config=/path/to/config.yml`) - Primary configuration
3. **CLI Parameters** (`--bind-public-port=8080`) - Overrides

**CRITICAL: Environment variables NOT desirable for configuration** (security risk, not scalable, auditability).

#### 9.2.2 *FromSettings Factory Pattern (PREFERRED)

Services should use settings-based factories for testability and consistency:

```go
// ✅ PREFERRED: Settings-based factory
type UnsealKeysSettings struct {
    KeyPaths []string `yaml:"key_paths"`
}

func NewUnsealKeysServiceFromSettings(settings *UnsealKeysSettings) (*UnsealKeysService, error) {
    if settings == nil {
        return nil, errors.New("settings required")
    }
    return &UnsealKeysService{
        keyPaths: settings.KeyPaths,
    }, nil
}

// Usage in ServerBuilder
builder.WithUnsealKeysService(func(settings *UnsealKeysSettings) (*UnsealKeysService, error) {
    return NewUnsealKeysServiceFromSettings(settings)
})
```

**Benefits**:
- All configuration in one struct
- Easy to test (pass test settings)
- Consistent initialization across codebase
- Self-documenting dependencies

#### 9.2.3 Test Settings Factory

Every service config should have a test settings factory:

```go
// NewTestSettings returns configuration suitable for testing
func NewTestSettings() *SMImServerSettings {
    return &SMImServerSettings{
        ServiceTemplateServerSettings: cryptoutilTemplateTestutil.NewTestSettings(),
        MaxMessageSize:                65536,
    }
}
```

**NewTestSettings() configures**:
- SQLite in-memory (`:memory:`)
- Port 0 (dynamic allocation, no conflicts)
- Auto-generated TLS certificates
- Disabled telemetry export
- Short timeouts for fast tests

#### 9.2.4 Secret Management Patterns

- Docker/Kubernetes secrets mounting
- File-based secret references (file://)
- Secret rotation strategies

### 9.3 Observability Architecture (OTLP)

**Telemetry Flow** (MANDATORY sidecar pattern):

```
cryptoutil services → opentelemetry-collector (OTLP :4317/:4318) → grafana-otel-lgtm (:14317/:14318)
```

**NEVER Direct**: ❌ cryptoutil → grafana-otel-lgtm (bypasses sidecar)

**Configuration**:

```yaml
telemetry:
  otlp_endpoint: "opentelemetry-collector:4317"  # Sidecar, not upstream
  protocol: "grpc"  # grpc (default) or http (firewall-friendly)
```

**Protocols**:

- **gRPC** (:4317) - Default, efficient binary protocol
- **HTTP** (:4318) - Firewall-friendly, universal compatibility

### 9.4 Telemetry Strategy

**Structured Logging** (MANDATORY):

```go
log.Info("operation completed",
  "user_id", userID,
  "operation", "key_generation",
  "duration_ms", duration.Milliseconds(),
)
```

**Standard Fields**: timestamp, level, message, trace_id, span_id, service.name, service.version

**Prometheus Metrics** (MANDATORY categories):

- **HTTP**: http_requests_total, http_request_duration_seconds, http_requests_in_flight
- **Database**: db_connections_open/idle, db_query_duration_seconds, db_errors_total
- **Crypto**: crypto_operations_total, crypto_operation_duration_seconds, crypto_errors_total
- **Keys**: keys_total, key_rotations_total, key_usage_total

**Sensitive Data Protection**:

- NEVER log: Passwords, API keys, tokens, private keys, PII
- Safe to log: Key IDs, user IDs, resource IDs, operation types, durations, counts

### 9.5 Container Architecture

**Multi-Stage Dockerfile Pattern**:

```dockerfile
# Global ARGs
ARG GO_VERSION=1.25.5
ARG VCS_REF
ARG BUILD_DATE

# Builder stage
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src
# Build logic...

# Validator stage (secrets validation MANDATORY)
FROM alpine:3.19 AS validator
WORKDIR /validation
RUN echo "🔍 Validating Docker secrets..."
# Validation logic...

# Runtime stage
FROM alpine:3.19 AS runtime
WORKDIR /app
COPY --from=validator /app/cryptoutil /app/cryptoutil
```

**Container Best Practices**:

- Base image: Alpine Linux (minimal attack surface)
- Runtime user: Non-root (security)
- Health checks: Integrated liveness/readiness
- Secret validation: MANDATORY validator stage
- Network isolation: Admin endpoints localhost-only

### 9.6 Orchestration Patterns

**Docker Compose**:

- Single build, shared image (prevents 3× build time)
- Health check dependencies (service_healthy, not service_started)
- Latency hiding: First instance initializes DB, others wait
- Expected startup: builder 30-60s, postgres 5-30s, cryptoutil 10-35s

**Kubernetes**:

- StatefulSet for services with persistent state
- Deployment for stateless services
- ConfigMaps for non-sensitive config
- Secrets for credentials (base64 encoded)
- Service mesh: Istio/Linkerd for mTLS

**Service Discovery**:

- Config file → Docker Compose DNS → Kubernetes DNS
- MUST NOT cache DNS results (for dynamic scaling)
- Health monitoring: Poll federated services every 30s

### 9.7 CI/CD Workflow Architecture

**NEVER DEFER Principle**: CI/CD workflow integration is non-negotiable. Every validator, quality gate, and enforcement tool MUST have a corresponding GitHub Actions workflow from the moment it is implemented. Deferring CI/CD integration to "later phases" is explicitly forbidden - it creates drift between local validation and CI enforcement.

**Workflow Categories**:

- **CI**: ci-quality (lint/format/build), ci-test (unit tests), ci-coverage (≥95%/98%), ci-benchmark, ci-mutation (≥95%/98%), ci-race (concurrency)
- **Security**: ci-sast (gosec), ci-gitleaks (secret detection), ci-dast (Nuclei/ZAP)
- **Integration**: ci-e2e (Docker Compose), ci-load (Gatling)
- **Deployment**: cicd-lint-deployments (8 validators on deployments/ and configs/)

**Quality Gates** (MANDATORY before merge):

- Build clean: `go build ./...` and `go build -tags e2e,integration ./...`
- Linting clean: `golangci-lint run` and `golangci-lint run --build-tags e2e,integration`
- Tests pass: 100%, zero skips
- Coverage: ≥95% production, ≥98% infrastructure/utility
- Mutation: ≥95% production, ≥98% infrastructure/utility
- Security: SAST/DAST clean

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

**GitHub Actions** reusable actions provide consistency across workflows:

- **docker-images-pull**: Parallel image pre-fetching to avoid rate limits
- **Input/Output patterns**: Parameter passing, cross-platform compatibility
- **Composite steps**: Shell selection, error handling

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

**Testing Pyramid**:

- **Unit Tests**: Fast (<15s per package), isolated, table-driven, t.Parallel()
- **Integration Tests**: TestMain pattern, shared resources, GORM repositories
- **E2E Tests**: Docker Compose, production-like, cross-service validation

**Coverage Requirements**:

- Production code: ≥95%
- Infrastructure/utility: ≥98%
- Mutation testing: ≥95% production, ≥98% infrastructure/utility
- Race detection: go test -race -count=2 (probabilistic execution)

### 10.2 Unit Testing Strategy

#### 10.2.1 Table-Driven Test Pattern

**MANDATORY for multiple test cases**:

```go
func TestSendMessage_Validation(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        request SendMessageRequest
        wantErr string
    }{
        {
            name:    "empty content",
            request: SendMessageRequest{Content: ""},
            wantErr: "content required",
        },
        {
            name:    "no recipients",
            request: SendMessageRequest{Content: "hello", Recipients: nil},
            wantErr: "at least one recipient",
        },
        {
            name: "valid request",
            request: SendMessageRequest{
                Content:    "hello",
                Recipients: []string{googleUuid.NewV7().String()},
            },
            wantErr: "",
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            // Use unique test data
            tenantID := googleUuid.NewV7()

            err := testServer.SendMessage(ctx, tenantID, tc.request)

            if tc.wantErr != "" {
                require.Error(t, err)
                require.Contains(t, err.Error(), tc.wantErr)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

**Benefits**:
- Single test function for multiple validation cases
- Easy to add new test cases (just add table row)
- t.Parallel() for concurrent execution
- UUIDv7 for dynamic, conflict-free test data

#### 10.2.2 Fiber Handler Testing (app.Test())

**ALWAYS use Fiber's in-memory testing for ALL HTTP handler tests**:

```go
func TestListMessages_Handler(t *testing.T) {
    t.Parallel()

    // Create standalone Fiber app
    app := fiber.New(fiber.Config{DisableStartupMessage: true})

    // Register handler under test
    msgRepo := repository.NewMessageRepository(testDB)
    handler := NewPublicServer(nil, msgRepo, nil, nil, nil)
    app.Get("/browser/api/v1/messages", handler.ListMessages)

    // Create HTTP request (no network call)
    req := httptest.NewRequest("GET", "/browser/api/v1/messages", nil)
    req.Header.Set("X-Tenant-ID", testTenantID.String())

    // Test handler in-memory
    resp, err := app.Test(req, -1)
    require.NoError(t, err)
    defer resp.Body.Close()

    require.Equal(t, 200, resp.StatusCode)
}
```

**Benefits**:
- In-memory testing (<1ms, no network binding)
- Prevents Windows Firewall popups
- Reliable, no port conflicts
- Test middleware, routing, and response handling

#### 10.2.3 Coverage Targets

- ≥95% production code
- ≥98% infrastructure/utility code
- 0% acceptable for main() if internalMain() ≥95%
- Generated code excluded from coverage
- **`internal/shared/magic/` excluded**: constants-only package, no executable logic

### 10.3 Integration Testing Strategy

#### 10.3.1 TestMain Pattern

**ALL integration tests MUST use TestMain for heavyweight dependencies**:

```go
var (
    testDB     *gorm.DB
    testServer *Server
)

func TestMain(m *testing.M) {
    ctx := context.Background()

    // Create server with test configuration
    cfg := config.NewTestSettings()
    var err error
    testServer, err = NewFromConfig(ctx, cfg)
    if err != nil {
        log.Fatalf("Failed to create test server: %v", err)
    }

    // Start server
    go func() {
        if err := testServer.Start(); err != nil {
            log.Printf("Server error: %v", err)
        }
    }()

    // Wait for ready
    if err := testServer.WaitForReady(ctx, 10*time.Second); err != nil {
        log.Fatalf("Server not ready: %v", err)
    }

    // Run tests
    exitCode := m.Run()

    // Cleanup
    testServer.Shutdown(ctx)
    os.Exit(exitCode)
}
```

**Benefits**:
- Start heavyweight resources (PostgreSQL containers, servers) ONCE per package
- Share testDB, testServer across all tests
- Prevents repeated 10-30s startup overhead
- Proper cleanup with defer statements

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

**Production-Like Environment**:

- Docker Compose with PostgreSQL, OpenTelemetry, services
- Docker secrets for credentials
- TLS-enabled HTTP client for secure testing
- Health check polling before test execution

**E2E Test Scope**: MUST test BOTH `/service/**` and `/browser/**` paths, verify middleware (IP allowlist, CSRF, CORS), cross-service integration

#### 10.4.1 Docker Compose Orchestration

**Use ComposeManager for E2E testing**:

```go
func TestE2E_SendMessage(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    ctx := context.Background()

    // Start Docker Compose stack
    manager := e2e.NewComposeManager(t, "../../../deployments/sm-im")
    manager.Up(ctx)
    defer manager.Down(ctx)

    // Wait for service healthy
    manager.WaitForHealthy(ctx, "sm-im", 60*time.Second)

    // Get TLS-enabled HTTP client
    client := manager.HTTPClient()

    // Test API
    resp, err := client.Post(
        manager.ServiceURL("sm-im") + "/browser/api/v1/messages",
        "application/json",
        strings.NewReader(`{"content":"hello","recipients":["user-id"]}`),
    )
    require.NoError(t, err)
    defer resp.Body.Close()

    require.Equal(t, 201, resp.StatusCode)
}
```

**Benefits**:
- Production-like environment (Docker secrets, TLS)
- Automatic lifecycle management (up/down)
- Health check polling
- TLS-enabled HTTP client for secure testing

#### 10.4.2 Docker Compose Health Check Patterns

**Three Healthcheck Use Cases**:

Docker Compose health checking supports three distinct patterns. The `docker compose up --wait` flag ONLY works with containers that have native `HEALTHCHECK` instructions in their Dockerfile. Many third-party containers (e.g., `otel/opentelemetry-collector-contrib`) lack native healthchecks, requiring alternative approaches.

##### Use Case 1: Job-Only Healthchecks

**Pattern**: Standalone job that must exit successfully (ExitCode=0)

**Examples**: `healthcheck-secrets`, `builder-cryptoutil`

**Usage**:
```go
services := []e2e.ServiceAndJob{
    {Service: "", Job: "healthcheck-secrets"},
}
err := composeManager.WaitForServicesHealthy(ctx, services)
```

**Docker Compose Pattern**:
```yaml
healthcheck-secrets:
  image: alpine:latest
  command:
    - sh
    - -c
    - |
      for secret in unseal_1 unseal_2 unseal_3 unseal_4 unseal_5 hash_pepper_v3 postgres_url; do
        test -f /run/secrets/$${secret}.secret || exit 1;
      done
      echo 'All secrets validated'
  secrets:
    - unseal_1.secret
    - unseal_2.secret
    - unseal_3.secret
```

##### Use Case 2: Service-Only Healthchecks

**Pattern**: Services with native HEALTHCHECK instructions in their container image or Dockerfile

**Examples**: `cryptoutil-sqlite`, `cryptoutil-postgres-1`, `postgres`, `grafana-otel-lgtm`

**Usage**:
```go
services := []e2e.ServiceAndJob{
    {Service: "cryptoutil-sqlite", Job: ""},
    {Service: "postgres", Job: ""},
}
err := composeManager.WaitForServicesHealthy(ctx, services)
```

**Docker Compose Pattern**:
```yaml
cryptoutil-sqlite:
  image: cryptoutil:latest
  healthcheck:
    test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null",
           "https://127.0.0.1:8080/admin/api/v1/livez"]
    interval: 10s
    timeout: 5s
    retries: 3
    start_period: 20s
```

##### Use Case 3: Service with Healthcheck Job

**Pattern**: Services without native healthchecks use external sidecar job for health verification

**Example**: `opentelemetry-collector-contrib` with `healthcheck-opentelemetry-collector-contrib`

**Usage**:
```go
services := []e2e.ServiceAndJob{
    {Service: "opentelemetry-collector-contrib", Job: "healthcheck-opentelemetry-collector-contrib"},
}
err := composeManager.WaitForServicesHealthy(ctx, services)
```

**Docker Compose Pattern**:
```yaml
opentelemetry-collector-contrib:
  image: otel/opentelemetry-collector-contrib:latest
  # No native HEALTHCHECK in container image

healthcheck-opentelemetry-collector-contrib:
  image: alpine:latest
  command:
    - sh
    - -c
    - |
      apk add --no-cache wget
      for i in $(seq 1 30); do
        if wget --quiet --tries=1 --spider --timeout=2 http://opentelemetry-collector-contrib:13133/ 2>/dev/null; then
          echo "OpenTelemetry Collector is ready"
          exit 0
        fi
        sleep 2
      done
      exit 1
  depends_on:
    opentelemetry-collector-contrib:
      condition: service_started
```

**ServiceAndJob Struct**:

```go
// Three use cases:
// 1. Job-only: ServiceAndJob{Service: "", Job: "job-name"}
// 2. Service-only: ServiceAndJob{Service: "service-name", Job: ""}
// 3. Service with job: ServiceAndJob{Service: "service-name", Job: "job-name"}
type ServiceAndJob struct {
    Service string // Service name (empty for standalone jobs)
    Job     string // Healthcheck job name (empty if service has native healthcheck)
}
```

**Implementation**: See `internal/apps/template/service/testing/e2e_infra/docker_health.go` for parseDockerComposePsOutput(), determineServiceHealthStatus(), WaitForServicesHealthy().

#### 10.4.3 E2E Test Scope

- MUST test BOTH `/service/**` and `/browser/**` paths
- Verify middleware behavior (IP allowlist, CSRF, CORS)
- Production-like environment (Docker secrets, TLS)

### 10.5 Mutation Testing Strategy

**Efficacy Targets**:

- Production code: ≥95%
- Infrastructure/utility: ≥98%

**Parallelization**: 4-6 packages per GitHub Actions matrix job (sequential 45min → parallel 15-20min)

**Exemptions**: OpenAPI-generated, GORM models, Protobuf stubs

#### 10.5.1 Gremlins Configuration

- Package-level parallelization (4-6 packages per job)
- Exclude tests, generated code, vendor
- Efficacy targets: ≥95% production, ≥98% infrastructure
- Timeout optimization: sequential 45min → parallel 15-20min

#### 10.5.2 Mutation Exemptions

- OpenAPI-generated code (stable, no business logic)
- GORM models (database schema definitions)
- Protobuf-generated code (gRPC/protobuf stubs)
- **`internal/shared/magic/`**: constants-only package, no executable logic to mutate

### 10.6 Load Testing Strategy

**Tool**: Gatling (Scala-based, Java 21 LTS required)

**Scenarios**:

- Baseline load: 100 users, 5 min duration
- Peak load: 1000 users, 10 min duration
- Stress test: Ramp to failure

**Metrics**: Response time (p50/p95/p99), throughput (req/s), error rate

### 10.7 Fuzz Testing Strategy

**Requirements**:

- Fuzz test files: *_fuzz_test.go (ONLY fuzz functions, exclude property tests with //go:build !fuzz)
- Minimum fuzz time: 15 seconds per test
- Always run from project root: go test -fuzz=FuzzXXX -fuzztime=15s ./path
- Unique function names: MUST NOT be substrings of others (e.g., FuzzHKDFAllVariants not FuzzHKDF)

### 10.8 Benchmark Testing Strategy

**Mandatory for crypto operations**:

```go
func BenchmarkAESEncrypt(b *testing.B) {
    key := make([]byte, 32)
    plaintext := make([]byte, 1024)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = encrypt(key, plaintext)
    }
}
```

**Execution**: go test -bench=. -benchmem ./pkg/crypto

### 10.9 Race Detection Strategy

**Probabilistic Execution**: go test -race -count=2 ./... (requires CGO_ENABLED=1)

**Why count=2+**: Race detector uses randomization, single execution may miss races

**CI/CD**: go test -race -count=5 ./... for more coverage

### 10.10 SAST Strategy

**Tool**: gosec (part of golangci-lint)

**Key Checks**:

- G401: Weak crypto (MD5, DES)
- G501: Import blocklist
- G505: Weak random (math/rand)

**Suppressions**: MUST include justification (#nosec G401 -- MD5 used for non-cryptographic checksums only)

### 10.11 DAST Strategy

**Tools**: Nuclei, OWASP ZAP

**Nuclei**:

- Quick scan: nuclei -target URL -severity info,low (1-2 min)
- Comprehensive: nuclei -target URL -severity medium,high,critical (5-15 min)
- Performance: nuclei -target URL -c 25 -rl 100

**OWASP ZAP**:

- Baseline: zap-baseline.py -t URL (5-10 min, passive)
- Full scan: zap-full-scan.py -t URL (30-60 min, active)

### 10.12 Workflow Testing Strategy

**Local Testing**: go run ./cmd/workflow -workflows=dast,e2e -inputs="key=value"

**Act Compatibility**: NEVER use -t timeout, ALWAYS specify -workflows, use -inputs for params

---

## 11. Quality Architecture

### 11.1 Maximum Quality Strategy - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ Correctness: ALL code functionally correct with comprehensive tests
- ✅ Completeness: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- ✅ Thoroughness: Evidence-based validation at every step
- ✅ Reliability: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ Efficiency: Optimized for maintainability and performance NOT implementation speed
- ✅ Accuracy: Changes must address root cause, not just symptoms
- ❌ Time Pressure: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ Premature Completion: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions:**

- ✅ **Fix issues immediately** - When unknowns discovered, blockers identified, unit/integration/E2E/mutations/fuzz/bench/race/SAST/DAST/load/any tests fail, or quality gates are not met, STOP and address
- ✅ **Treat as BLOCKING**: ALL issues block progress to next task
- ✅ **Document root causes** - Root cause analysis is part of planning AND implementation, not optional; planning blockers must be resolved during planning, implementation blockers MUST be resolved during implementation
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark phase or task complete with known issues
- ✅ **NEVER de-prioritize quality** - Evidence-based verification is ALWAYS highest priority

**Rationale**: Maintaining maximum quality prevents cascading failures and rework.

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

- **ALL magic constants and variables MUST be consolidated in `internal/shared/magic/`**; domain-specific sub-files allowed (e.g., `magic_domain*.go`) but NEVER in scattered package-local files
- Shared constants: internal/shared/magic/magic_*.go (network, database, cryptography, testing)
- Pattern: Declare as named variables, NEVER inline literals
- Rationale: mnd (magic number detector) linter enforcement
- **Coverage/Mutation Exemption**: `internal/shared/magic/` is **excluded from all code coverage and mutation testing thresholds**; it contains only named constants and variables with no executable logic to test

### 11.2 Quality Gates

#### 11.2.1 MANDATORY Pre-Commit Quality Gates

`go build ./...` → clean build (all non-tagged files)
`go build -tags e2e,integration ./...` → clean build (all build-tagged files)
`golangci-lint run --fix` → zero warnings (all non-tagged files)
`golangci-lint run --build-tags e2e,integration --fix` → zero warnings (all build-tagged files)
`go test -cover -shuffle=on ./...` → MANDATORY 100% tests pass, and ≥98% coverage per package

#### 11.2.2 RECOMMENDED Pre-Commit Quality Gates

`go get -u ./...` → direct dependency updates only
`go mod tidy` → dependency tidy-up; must run after update dependencies
`govulncheck ./...` → vulnerability scan

#### 11.2.3 RECOMMENDED Pre-push Quality Gates

`gremlins unleash --tags=!integration` → mutation testing
`govulncheck ./...` → vulnerability scan

#### 11.2.4 SUGGESTED Pre-push Quality Gates

`go get -u all` → all dependency updates, including transitive dependencies
`go test -bench=. -benchmem ./pkg/path` → benchmark tests
`go test -fuzz=FuzzTestName -fuzztime=15s ./pkg/path` → fuzz tests
`go test -race -count=3 ./...` → race detection

#### 11.2.5 CI/CD

***Docker Desktop** MUST be running locally, because workflows are run locally by `act` in containers.

See [Section 13.5.4 Docker Desktop Startup](#1354-docker-desktop-startup---critical) for cross-platform startup instructions (Windows, macOS, Linux).

Here are local convenience commands to run the workflows locally for Development and Testing.

`go run ./cmd/workflow -workflows=build`     → build check
`go run ./cmd/workflow -workflows=coverage`  → workflow coverage check; ≥98% required
`go run ./cmd/workflow -workflows=quality`   → workflow quality check
`go run ./cmd/workflow -workflows=lint`      → linting check
`go run ./cmd/workflow -workflows=benchmark` → workflow benchmark check
`go run ./cmd/workflow -workflows=fuzz`      → workflow fuzz check
`go run ./cmd/workflow -workflows=race`      → workflow race check
`go run ./cmd/workflow -workflows=sast`      → static security analysis
`go run ./cmd/workflow -workflows=gitleaks`  → secrets scanning
`go run ./cmd/workflow -workflows=dast`      → dynamic security testing
`go run ./cmd/workflow -workflows=mutation`  → mutation testing; ≥95% required
`go run ./cmd/workflow -workflows=e2e`       → end-to-end tests; BOTH `/service/**` AND `/browser/**` paths
`go run ./cmd/workflow -workflows=load`      → load testing
`go run ./cmd/workflow -workflows=ci`        → full CI workflow

**Mutation Testing Scope**: ALL `cmd/cicd/` packages (including `lint_deployments/`) require ≥98% mutation testing efficacy. This includes test infrastructure, CLI wiring, and validator implementations. Mutation testing validates test quality, not just test coverage.

**Deployment Validation CI/CD**: The `cicd-lint-deployments` workflow runs `validate-all` on every push/PR affecting `deployments/**`, `configs/**`, or validator source code. Validation failures block merges. CI/CD integration is NEVER deferred - all validators must have workflow coverage from the moment they are implemented.

#### 11.2.6 File Size Limits

| Threshold | Lines | Action |
|-----------|-------|--------|
| Soft | 300 | Ideal target |
| Medium | 400 | Acceptable with justification |
| Hard | 500 | NEVER EXCEED - refactor required |

- Rationale: Faster LLM processing, easier review, better organization, forces logical grouping

#### 11.2.7 Conditional Statement Patterns

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

#### 11.2.8 format_go Self-Modification Protection - CRITICAL

- Root Cause: LLM agents lose exclusion context during narrow-focus refactoring
- NEVER DO:
  - ❌ Modify comments in enforce_any.go without reading full package context
  - ❌ Change backticked `interface{}` to `any` in format_go package
  - ❌ Refactor code in isolation (single-file view)
  - ❌ Simplify "verbose" CRITICAL comments
- ALWAYS DO:
  - ✅ Read complete package context before refactoring self-modifying code
  - ✅ Check for CRITICAL/SELF-MODIFICATION tags in comments
  - ✅ Verify self-exclusion patterns exist and are respected
  - ✅ Run tests after ANY changes to format_go package

#### 11.2.9 Restore from Clean Baseline Pattern

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

**Pull Request Requirements**:

- Title: type(scope): description (<72 chars)
- Tests added: ≥95%/98% coverage
- Linting passes: golangci-lint clean
- Docs updated: README, ARCHITECTURE.md, instruction files
- Security: No sensitive data, input validation, secure defaults

**PR Size Guidelines**:

- <200 lines: Small (ideal)
- 200-500 lines: Medium (acceptable)
- 500-1000 lines: Large (consider splitting)
- 1000+ lines: Epic (MUST break down)

---

## 12. Deployment Architecture

### 12.1 CI/CD Automation Strategy

**Workflow Matrix**: ci-quality (lint/format/build), ci-coverage/mutation/race (test analysis), ci-sast/dast (security), ci-e2e/benchmark/load (integration/performance)

**Quality Gates before merge**: Build clean, linting clean, tests pass (100%), coverage (≥95%/98%), mutation (≥95%/98%), security (SAST/DAST clean)

### 12.2 Build Pipeline

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

**Environments**: Development (SQLite, local), Testing (test-containers, CI), Staging (Docker Compose, TLS), Production (Kubernetes, cloud)

**Secret Management**: Docker/Kubernetes secrets (MANDATORY), NEVER inline environment variables

#### 12.3.1 Docker Compose Deployment

- Secret management via Docker secrets (MANDATORY)
- Health check configuration (interval, timeout, retries, start-period)
- Dependency ordering (depends_on with service_healthy)
- Network isolation patterns

##### Docker Secrets (MANDATORY)

```yaml
secrets:
  postgres_password.secret:
    file: ./secrets/postgres_password.secret  # chmod 440

services:
  postgres:
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password.secret
    secrets:
      - postgres_password.secret
```

**NEVER use inline environment variables for credentials.**

##### Health Checks

**Three Healthcheck Patterns** (see Section 10.4.2 for details):

1. **Service-only** (native HEALTHCHECK):
```yaml
cryptoutil-service:
  healthcheck:
    test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null",
           "https://127.0.0.1:9090/admin/api/v1/livez"]
    start_period: 60s
    interval: 5s
```

2. **Job-only** (validation job, ExitCode=0 required):
```yaml
healthcheck-secrets:
  image: alpine:latest
  command: ["sh", "-c", "test -f /run/secrets/unseal_1.secret || exit 1"]
```

3. **Service with healthcheck job** (external sidecar):
```yaml
otel-collector:
  image: otel/opentelemetry-collector-contrib:latest
  # No native healthcheck

healthcheck-otel-collector:
  image: alpine:latest
  command: ["sh", "-c", "wget --spider http://otel-collector:13133/"]
  depends_on:
    otel-collector:
      condition: service_started
```

**Use wget (Alpine), 127.0.0.1 (not localhost), port 9090 for services. Job-only and service-with-job patterns validate availability before services start.**

#### 12.3.2 Kubernetes Deployment

- secretKeyRef for environment variables from secrets
- Volume mount for file-based secrets
- Health probes (liveness vs readiness)
- Resource limits and requests

#### 12.3.3 Secrets Coordination Strategy

**MANDATORY: All secrets stored in Docker/Kubernetes secrets, NEVER inline environment variables.**

**Secrets Structure**: Each service requires 10 secrets (5 unseal keys, 1 hash pepper, 4 PostgreSQL credentials).

##### Secret Naming Suffixes (4 Levels)

**MANDATORY**: All secret files MUST use a level suffix to indicate deployment scope:

| Suffix | Location | Scope | Example |
|--------|----------|-------|---------|
| `-SERVICEONLY.secret` | `deployments/PRODUCT-SERVICE/secrets/` | Single service only | `unseal_1of5-SERVICEONLY.secret` |
| `-PRODUCTONLY.secret` | `deployments/PRODUCT/secrets/` | Product services only | `postgres_url-PRODUCTONLY.secret` |
| `-SUITEONLY.secret` | `deployments/cryptoutil-suite/secrets/` | Suite-wide only | `admin_api_key-SUITEONLY.secret` |
| `-SHARED.secret` | Any level | Shared across multiple levels | `hash_pepper_v3-SHARED.secret` |

**Rules**:

- **`-SERVICEONLY`**: Secret exists at exactly one service level. UNIQUE per service.
- **`-PRODUCTONLY`**: Secret exists at product level only. NOT inherited by service-level deployments.
- **`-SUITEONLY`**: Secret exists at suite level only. NOT inherited by product/service deployments.
- **`-SHARED`**: Secret is shared across multiple deployment levels (e.g., hash pepper for SSO across identity services). The SAME value MUST be used at all levels where it appears.

**Transition**: Existing service-level secrets without suffixes are valid during transition. New secrets MUST use level suffixes. When SUITE/PRODUCT directories are created, all secrets MUST use appropriate suffixes.

##### SUITE-Level Deployment (cryptoutil)

**Location**: `deployments/cryptoutil-suite/secrets/` (template pattern applied)

**Consistency Requirements**:

- **hash_pepper_v3.secret**: SAME value across ALL 9 services (enables cross-service PII deduplication, SSO username lookup)
- **unseal_*.secret**: UNIQUE per service (barrier encryption independence, security isolation)
- **postgres_*.secret**: Service-specific (27 logical databases: 9 suite + 9 product + 9 service)

**Rationale**: Unified hash pepper allows username@domain lookups across identity services while maintaining per-service encryption boundaries.

##### PRODUCT-Level Deployment (identity, jose, pki, sm)

**Location**: `deployments/PRODUCT/secrets/` (multiple services per product)

**Consistency Requirements**:

- **hash_pepper_v3.secret**: SAME value within product services (identity-{authz,idp,rs,rp,spa} share pepper for SSO)
- **unseal_*.secret**: UNIQUE per service within product (independent barrier hierarchies)
- **postgres_*.secret**: Service-specific (each service has 3 databases: suite, product, service levels)

**Example**: identity-authz, identity-idp, identity-rs, identity-rp, identity-spa share hash_pepper for unified user lookups.

##### SERVICE-Level Deployment (single service)

**Location**: `deployments/PRODUCT-SERVICE/secrets/` (e.g., `deployments/jose-ja/secrets/`)

**Consistency Requirements**:

- **hash_pepper_v3.secret**: UNIQUE per service (no cross-service lookups needed)
- **unseal_*.secret**: UNIQUE per service (barrier encryption independence)
- **postgres_*.secret**: Service-specific (3 databases per service)

**Rationale**: Maximum isolation for standalone deployments.

##### Secret File Format

**Unseal Keys** (`unseal_{1,2,3,4,5}.secret`):

```
jose-ja-40c8c0f3c1c3b9c3f3c3b9c3f3c3b9c3f3c3b9c3
```

- Format: `{product}-{service}-{32 hex chars}`
- Generation: `secrets.token_hex(16)`
- Purpose: HKDF deterministic derivation for Root Key encryption

**Hash Pepper** (`hash_pepper_v3.secret`):

```
dGhpcyBpcyBhIDMyLWJ5dGUgcGVwcGVyIGZvciBoYXNoaW5nIQ==
```

- Format: Base64-encoded 32 bytes
- Generation: `base64.b64encode(secrets.token_bytes(32))`
- Purpose: PBKDF2/HKDF salt/pepper for PII hashing (username, email deduplication)

**PostgreSQL Credentials** (`postgres_{url,username,password,database}.secret`):

```
# postgres_url.secret
postgres://jose_ja_user:jose-ja-pass@postgres-leader:5432/jose_ja_db

# postgres_username.secret
jose_ja_user

# postgres_password.secret
jose-ja-pass-40c8c0f3c1c3b9c3f3c3b9c3f3c3b9c3

# postgres_database.secret
jose_ja_db
```

- Naming: `{product}_{service}_user`, `{product}_{service}_db`
- Password: `{product}-{service}-pass-{32 hex chars}`

##### Secret Validation

**Healthcheck-Secrets Service** (in all compose templates):

```yaml
healthcheck-secrets:
  image: alpine:3.19
  command: >
    sh -c "
      for secret in unseal_1 unseal_2 unseal_3 unseal_4 unseal_5
                   hash_pepper_v3
                   postgres_url postgres_username postgres_password postgres_database; do
        test -f /run/secrets/$${secret}.secret || exit 1;
      done;
      echo 'All secrets validated';
    "
  secrets:
    - unseal_1.secret
    - unseal_2.secret
    - unseal_3.secret
    - unseal_4.secret
    - unseal_5.secret
    - hash_pepper_v3.secret
    - postgres_url.secret
    - postgres_username.secret
    - postgres_password.secret
    - postgres_database.secret
```

**Purpose**: Fail fast on missing secrets, prevent runtime errors.

##### Cross-Reference Documentation

- Secrets generation scripts: `internal/cmd/cicd/secrets/` (Python secrets module)
- Security architecture: [02-05.security.instructions.md](../.github/instructions/02-05.security.instructions.md#secret-management---mandatory)
- Hash service patterns: [02-05.security.instructions.md](../.github/instructions/02-05.security.instructions.md#hash-service-architecture)
- Docker Compose templates: `deployments/template/compose-cryptoutil*.yml`

#### 12.3.4 Multi-Level Deployment Hierarchy

**Purpose**: Enable flexible deployment granularity from single-service development to full suite production deployment while maintaining consistent configuration and secret management across all levels.

**Implementation Status**: Implemented and validated (2026-02-16)

**MANDATORY: Rigid Delegation Pattern** (enforced by linter):
- **SUITE → PRODUCT → SERVICE**: ALL deployments MUST follow this delegation chain
- **Rationale**:
  1. Products can scale from 1 to N services without breaking suite-level deployment
  2. Suite can scale from 5 to N products without hardcoded service dependencies
  3. Enables independent testing at each level (service, product, suite)

**Linter Enforcement** (`cicd lint-deployments`):
- Suite compose MUST include product-level (e.g., `../sm/compose.yml`), NOT service-level (e.g., `../sm-kms/compose.yml`)
- Product compose MUST include service-level (e.g., `../sm-kms/compose.yml`)
- Violations are ERRORS that block CI/CD

**Three-Tier Hierarchy**:

| Level | Directory | Scope | Services | Use Cases |
|-------|-----------|-------|----------|-----------|
| **SERVICE** | `deployments/{PRODUCT}-{SERVICE}/` | Single service | 1 | Development, testing, isolated deployment |
| **PRODUCT** | `deployments/{PRODUCT}/` | Product services | 1-5 | Product-level testing, SSO within product |
| **SUITE** | `deployments/cryptoutil-suite/` | All services | 9 | Full integration, cross-product federation |

##### SUITE-Level Deployment (cryptoutil)

**Location**: `deployments/cryptoutil-suite/compose.yml`

**Composition Pattern**: Includes all PRODUCT-level composes via Docker Compose `include` directive.

```yaml
include:
  - path: ../sm/compose.yml          # sm-kms, sm-im services
  - path: ../pki/compose.yml         # pki-ca service
  - path: ../jose/compose.yml        # jose-ja service
  - path: ../identity/compose.yml    # identity-authz, -idp, -rp, -rs, -spa

secrets:
  cryptoutil-hash_pepper.secret:
    file: ./secrets/cryptoutil-hash_pepper.secret
```

**Purpose**: Deploy all 9 services with unified hash pepper for cross-product SSO and PII deduplication.

**Secret Sharing**: `cryptoutil-hash_pepper.secret` shared by ALL services enables username@domain lookups across identity-authz, identity-idp, jose-ja, etc.

**Port Assignments**: Suite deployment uses offset +20000 (e.g., sm-kms: 28080 public, 29090 admin instead of 8080/9090).

##### PRODUCT-Level Deployment (Multi-Service Products)

**Example**: `deployments/identity/compose.yml` (5 services: authz, idp, rp, rs, spa)

**Composition Pattern**: Includes all SERVICE-level composes for product services.

```yaml
include:
  - path: ../identity-authz/compose.yml
  - path: ../identity-idp/compose.yml
  - path: ../identity-rp/compose.yml
  - path: ../identity-rs/compose.yml
  - path: ../identity-spa/compose.yml

secrets:
  identity-hash_pepper.secret:
    file: ./secrets/identity-hash_pepper.secret
```

**Purpose**: Deploy all identity services with shared hash pepper for SSO within product.

**Secret Sharing**: `identity-hash_pepper.secret` shared by all 5 identity services enables unified username lookups for authentication/authorization.

**Port Assignments**: Product deployment uses offset +10000 (e.g., identity-authz: 18200 public, 19290 admin instead of 8200/9290).

**Other Products**:
- `deployments/sm/compose.yml` → includes `../sm-kms/` (currently single service)
- `deployments/pki/compose.yml` → includes `../pki-ca/` (currently single service)
- `deployments/sm/compose.yml` → includes sm-kms and sm-im services
- `deployments/jose/compose.yml` → includes `../jose-ja/` (currently single service)

##### SERVICE-Level Deployment (Individual Services)

**Example**: `deployments/sm-kms/compose.yml` (standalone sm-kms service)

**Composition Pattern**: Direct service deployment with NO includes.

```yaml
services:
  sm-kms:
    build:
      context: ../..
      dockerfile: ./deployments/sm-kms/Dockerfile
    ports:
      - "8080:8080"   # Public API
      - "9090:9090"   # Admin API

secrets:
  sm-kms-hash_pepper.secret:
    file: ./secrets/sm-kms-hash_pepper.secret
```

**Purpose**: Deploy single service with unique hash pepper (NO cross-service sharing).

**Secret Uniqueness**: Each SERVICE-level deployment uses unique `{PRODUCT}-{SERVICE}-hash_pepper.secret` for maximum isolation.

**Port Assignments**: Service deployment uses base ports (e.g., sm-kms: 8080 public, 9090 admin).

##### Layered Pepper Strategy

**Three Tiers** (from most isolated to most shared):

1. **SERVICE pepper** (`{PRODUCT}-{SERVICE}-hash_pepper.secret`): Unique per service, NO cross-service lookups
2. **PRODUCT pepper** (`{PRODUCT}-hash_pepper.secret`): Shared within product services (e.g., 5 identity services)
3. **SUITE pepper** (`cryptoutil-hash_pepper.secret`): Shared by ALL 9 services for cross-product federation

**Selection Logic** (service configures which pepper to use):

```yaml
# SERVICE-only deployment (isolated)
HASH_PEPPER_FILE: /run/secrets/sm-kms-hash_pepper.secret

# PRODUCT deployment (SSO within product)
HASH_PEPPER_FILE: /run/secrets/identity-hash_pepper.secret

# SUITE deployment (cross-product federation)
HASH_PEPPER_FILE: /run/secrets/cryptoutil-hash_pepper.secret
```

**Rationale**: Layered peppers enable flexible deployment modes while maintaining security isolation at SERVICE level and enabling federation at PRODUCT/SUITE levels.

##### Port Offset Strategy

**Three Port Ranges** (prevents conflicts when multiple deployment levels running simultaneously):

| Level | Offset | Example (sm-kms base 8080/9090) |
|-------|--------|----------------------------------|
| SERVICE | +0 | 8080 (public), 9090 (admin) |
| PRODUCT | +10000 | 18080 (public), 19090 (admin) |
| SUITE | +20000 | 28080 (public), 29090 (admin) |

**Why**: Enables simultaneous SERVICE, PRODUCT, SUITE deployments on same host for testing without port conflicts.

##### Linter Validation

**PRODUCT Deployment Validation** (`cicd lint-deployments deployments/identity`):

```
✅ Required: compose.yml exists
✅ Required: secrets/ directory exists
✅ Required: identity-hash_pepper.secret exists in secrets/
✅ Forbidden: unseal_*.secret MUST NOT exist (documented by .never files)
```

**SUITE Deployment Validation** (`cicd lint-deployments deployments/cryptoutil-suite`):

```
✅ Required: compose.yml exists
✅ Required: secrets/ directory exists
✅ Required: cryptoutil-hash_pepper.secret exists in secrets/
✅ Forbidden: unseal_*.secret MUST NOT exist (documented by .never files)
```

**Enforcement**: Linter validates ALL 19 deployments (9 SERVICE, 5 PRODUCT, 1 SUITE, 1 template, 3 infrastructure).

**Implementation**: [internal/cmd/cicd/lint_deployments/lint_deployments.go](/internal/cmd/cicd/lint_deployments/) with `validateProductSecrets()` and `validateSuiteSecrets()` functions.

##### Cross-Reference Documentation

- **Comprehensive hierarchy documentation**: [ARCHITECTURE-COMPOSE-MULTIDEPLOY.md](/docs/ARCHITECTURE-COMPOSE-MULTIDEPLOY.md)
- **Secrets coordination**: [12.3.3 Secrets Coordination Strategy](#1233-secrets-coordination-strategy)
- **Deployment validation**: [12.4 Deployment Structure Validation](#124-deployment-structure-validation)
- **Port assignments**: [3.4.1 Port Design Principles](#341-port-design-principles)

### 12.4 Deployment Structure Validation

**Purpose**: Automated enforcement of consistent deployment directory structures across all services to prevent configuration drift and deployment failures.

**Linter Tool**: `cicd lint-deployments <directory>` validates all deployments in a directory tree.

**Implementation**: [internal/cmd/cicd/lint_deployments](/internal/cmd/cicd/lint_deployments/) package with comprehensive table-driven tests.

#### 12.4.1 Deployment Types

**SUITE** (e.g., cryptoutil - all 9 services):
- Required directories: `secrets/`
- Required files: `compose.yml`
- Optional files: `README.md`
- Required secrets: 1 file (hash pepper ONLY, NO unseal keys)
  - `cryptoutil-hash_pepper.secret` (shared by all 9 services)
- Forbidden secrets (documented by `.never` files):
  - `unseal_1of5-SUITEONLY.never` through `unseal_5of5-SUITEONLY.never`
  - Rationale: Unseal keys MUST be unique per service (security isolation)
- Validation function: `validateSuiteSecrets()` in lint_deployments.go

**PRODUCT** (e.g., identity, sm, pki, jose):
- Required directories: `secrets/`
- Required files: `compose.yml`
- Optional files: `README.md`
- Required secrets: 1 file (hash pepper ONLY, NO unseal keys)
  - `{PRODUCT}-hash_pepper.secret` (shared within product services)
- Forbidden secrets (documented by `.never` files):
  - `unseal_1of5-PRODUCTONLY.never` through `unseal_5of5-PRODUCTONLY.never`
  - Rationale: Unseal keys MUST be unique per service (security isolation)
- Validation function: `validateProductSecrets()` in lint_deployments.go

**PRODUCT-SERVICE** (e.g., sm-im, jose-ja, pki-ca, sm-kms, identity-authz/idp/rp/rs/spa):
- Required directories: `secrets/`, `config/`
- Required files: `compose.yml`, `Dockerfile`
- Optional files: `compose.demo.yml`, `otel-collector-config.yaml`, `README.md`
- Required secrets: 14 files (5 unseal, 1 hash pepper, 4 PostgreSQL, 4 auth credentials)
  - `unseal_1of5.secret` through `unseal_5of5.secret`
  - `hash_pepper_v3.secret`
  - `postgres_url.secret`, `postgres_username.secret`, `postgres_password.secret`, `postgres_database.secret`
  - `browser_username.secret`, `browser_password.secret` (web UI auth)
  - `service_username.secret`, `service_password.secret` (headless/API auth)

**template** (deployment template for new services):
- Required directories: `secrets/`
- Required files: `compose.yml`
- Optional files: `compose.demo.yml`, `Dockerfile`, `README.md`
- Required secrets: Same 14 files as PRODUCT-SERVICE

**infrastructure** (shared-postgres, shared-citus, shared-telemetry):
- Required directories: none
- Required files: `compose.yml`
- Optional files: `init-db.sql`, `init-citus.sql`, `README.md`
- Required secrets: none (infrastructure secrets are optional)

#### 12.4.2 Validation Rules

**Directory Structure**: Each deployment type enforces specific required/optional directories.

**File Requirements**: compose.yml is MANDATORY for all types; Dockerfile MANDATORY only for PRODUCT-SERVICE.

**Secret Validation**: For PRODUCT-SERVICE and template types, all 14 required secrets MUST exist in `secrets/` directory.

**Error Reporting**: Linter identifies missing directories, missing files, and missing secrets with actionable error messages.

**Rigid Delegation Pattern** (NEW - enforced 2026-02-16):
- **SUITE Compose**: MUST include PRODUCT-level paths (e.g., `../sm/compose.yml`), NEVER service-level (e.g., `../sm-kms/compose.yml`)
- **PRODUCT Compose**: MUST include SERVICE-level paths (e.g., `../sm-kms/compose.yml`)
- **Validation Function**: `checkDelegationPattern()` in lint_deployments.go
- **Failure Mode**: Violations are ERRORS that block CI/CD

**Database Isolation** (NEW - enforced 2026-02-16):
- Each of 9 services MUST have unique `postgres_database.secret` value
- Each of 9 services MUST have unique `postgres_username.secret` value
- Duplicate database names or usernames across services are ERRORS
- **Validation Function**: `checkDatabaseIsolation()` in lint_deployments.go
- **Cross-Service Check**: Runs after all deployments validated to detect sharing violations

**Authentication Credentials** (NEW - enforced 2026-02-16):
- Each service MUST have 4 credential files: `browser_username.secret`, `browser_password.secret`, `service_username.secret`, `service_password.secret`
- **Validation Function**: `checkBrowserServiceCredentials()` in lint_deployments.go
- **Rationale**: No hardcoded credentials in config files (E2E testing requires Docker secrets)

**OTLP Protocol Override** (NEW - enforced 2026-02-16):
- Config files SHOULD NOT specify protocol in `otlp-endpoint` (no `grpc://` or `http://` prefixes)
- Use hostname:port format (e.g., `opentelemetry-collector-contrib:4317`)
- **Validation Function**: `checkOTLPProtocolOverride()` in lint_deployments.go
- **Failure Mode**: Violations are WARNINGS (non-blocking)

#### 12.4.3 CI/CD Integration

**GitHub Actions Workflow**: [cicd-lint-deployments.yml](/.github/workflows/cicd-lint-deployments.yml) runs on all changes to `deployments/**`.

**Quality Gate**: Deployment structure validation is MANDATORY before merge; violations block CI/CD pipeline.

**Artifact Upload**: Validation reports uploaded as GitHub Actions artifacts for 7-day retention.

#### 12.4.4 Cross-Reference Documentation

- Secrets coordination strategy: [12.3.3 Secrets Coordination Strategy](#1233-secrets-coordination-strategy)
- Docker Compose deployment patterns: [12.3.1 Docker Compose Deployment](#1231-docker-compose-deployment)
- Secret management instructions: [02-05.security.instructions.md](../.github/instructions/02-05.security.instructions.md#secret-management---mandatory)

#### 12.4.5 Config File Naming Strategy

**MANDATORY Pattern**: All service config files MUST use full `{PRODUCT}-{SERVICE}-app-{variant}.yml` naming.

**Standard Config File Set** (required for ALL services):

- `{PRODUCT}-{SERVICE}-app-common.yml` - Shared configuration for all deployment modes (SQLite, PostgreSQL-1, PostgreSQL-2)
- `{PRODUCT}-{SERVICE}-app-sqlite-1.yml` - SQLite in-memory configuration for single-instance development/testing
- `{PRODUCT}-{SERVICE}-app-postgresql-1.yml` - PostgreSQL instance 1 configuration (shared database)
- `{PRODUCT}-{SERVICE}-app-postgresql-2.yml` - PostgreSQL instance 2 configuration (shared database, high-availability pair)

**Optional Config Files**:

- `{PRODUCT}-{SERVICE}-e2e.yml` - End-to-end test-specific overrides
- `{PRODUCT}-{SERVICE}-demo.yml` - Demo environment settings

**Examples**:

```
deployments/sm-kms/config/
├── sm-kms-app-common.yml
├── sm-kms-app-sqlite-1.yml
├── sm-kms-app-postgresql-1.yml
├── sm-kms-app-postgresql-2.yml
├── sm-kms-e2e.yml          (optional)
└── sm-kms-demo.yml         (optional)

deployments/jose-ja/config/
├── jose-ja-app-common.yml
├── jose-ja-app-sqlite-1.yml
├── jose-ja-app-postgresql-1.yml
└── jose-ja-app-postgresql-2.yml
```

**Rationale**:

- **Explicit Product-Service Coupling**: Prevents config file collisions when multiple services deployed together
- **Variant Clarity**: `app-sqlite-1` vs `app-postgresql-1` makes deployment mode immediately obvious
- **Instance Numbering**: `-1` and `-2` suffixes enable horizontal scaling with unique configs per instance
- **Tooling Support**: Linter validates presence of 4 required files, flags non-conformant naming

**Migration Strategy** (Q9 Answer: Break Immediately):

- NO backward compatibility period - rename all files immediately
- NO symlinks or aliases - clean cutover
- **Rationale**: Pre-production repository with zero deployed instances, rigid enforcement prevents future drift

#### 12.4.6 Demo and Integration File Handling

**Decision** (Q3 Answer: Remove from service directories):

- **demo-seed.yml**: Remove from all `deployments/{PRODUCT}-{SERVICE}/config/` directories
- **integration.yml**: Remove from all `deployments/{PRODUCT}-{SERVICE}/config/` directories

**Replacement Pattern**:

- Demo-specific settings → `{PRODUCT}-{SERVICE}-demo.yml` (optional file in config/)
- E2E test settings → `{PRODUCT}-{SERVICE}-e2e.yml` (optional file in config/)
- Integration test data → Use TestMain with test-containers (NOT Docker Compose)

**Rationale**:

- Ambiguous naming (`demo-seed`, `integration`) caused confusion about purpose
- New naming (`-demo.yml`, `-e2e.yml`) aligns with PRODUCT-SERVICE prefix pattern
- Optional nature prevents bloat when not needed

**Linter Enforcement** (Q5 Answer: Warning Mode Transition):

```go
// Phase 1: Warning Mode (current)
if file == "demo-seed.yml" || file == "integration.yml" {
    warnings = append(warnings, fmt.Sprintf("DEPRECATED: %s should be removed or renamed to %s-demo.yml / %s-e2e.yml",
        file, productService, productService))
}

// Phase 2: Error Mode (after transition period)
if file == "demo-seed.yml" || file == "integration.yml" {
    errors = append(errors, fmt.Sprintf("FORBIDDEN: %s must be removed", file))
}
```

**Transition Period**: Completed. Strict enforcement mode is now active.

#### 12.4.7 Linter Validation Modes

**Current Mode**: Strict (all violations block CI/CD)

**ALL violations are errors (blocking)**:

- Config files not matching `{PRODUCT}-{SERVICE}-app-{variant}.yml` pattern
- Presence of deprecated `demo-seed.yml` or `integration.yml` files
- Missing required config files (4 standard files)
- Missing required secrets (10 secret files)
- Missing required directories (`secrets/`, `config/`)
- Missing required compose/Dockerfile files
- Single-part deployment names (must be `PRODUCT-SERVICE` format)
- Wrong product prefix in config file names

#### 12.4.8 Config File Content Validation

**Implementation**: `ValidateConfigFile()` in [internal/cmd/cicd/lint_deployments/validate_config.go](/internal/cmd/cicd/lint_deployments/validate_config.go)

**Schema Reference**: See [CONFIG-SCHEMA.md](/docs/CONFIG-SCHEMA.md) for complete config file schema with all supported keys, types, and valid values.

**Validation Rules**:

1. **YAML Syntax**: File must parse as valid YAML
2. **Bind Address Format**: Must be valid IPv4 (via `net.ParseIP`)
3. **Port Range**: 1-65535 inclusive
4. **Protocol**: Must be `https` (TLS required)
5. **Admin Bind Policy**: `bind-private-address` MUST be `127.0.0.1`
6. **Secret References**: `database-url` must use `file:///run/secrets/` or `sqlite://` (never inline `postgres://`)
7. **OTLP Consistency**: When `otlp: true`, `otlp-service` and `otlp-endpoint` are required

**CLI Usage**: `cicd lint-deployments validate-config <config-file.yml>`

#### 12.4.9 Compose File Content Validation

**Implementation**: `ValidateComposeFile()` in [internal/cmd/cicd/lint_deployments/validate_compose.go](/internal/cmd/cicd/lint_deployments/validate_compose.go)

**Validation Rules**:

1. **Port Conflicts**: No duplicate host port bindings across services
2. **Health Checks**: All non-exempt services must have healthcheck configuration
3. **Dependency Chains**: `depends_on` references must resolve to defined services
4. **Secret References**: Referenced secrets must be defined in the `secrets:` section
5. **Hardcoded Credentials**: Environment variables must not contain inline passwords
6. **Bind Mount Security**: Host paths must use relative paths (no absolute paths)
7. **Include Resolution**: Docker Compose `include` directives are resolved for cross-file validation

**CLI Usage**: `cicd lint-deployments validate-compose <compose-file.yml>`

#### 12.4.10 Structural Mirror Validation

**Implementation**: `ValidateStructuralMirror()` in [internal/cmd/cicd/lint_deployments/validate_mirror.go](/internal/cmd/cicd/lint_deployments/validate_mirror.go)

**Direction**: `deployments/` → `configs/` (one-way). Every deployment directory MUST have a `configs/` counterpart.

**Mapping Rules**:

- `PRODUCT-SERVICE` (e.g., `jose-ja`) → product name (e.g., `jose`)
- `PRODUCT` (e.g., `sm`) → same name
- Explicit overrides: `pki`/`pki-ca` → `ca`, `sm`/`sm-kms` → `sm`

**Exclusions**: Infrastructure deployments (`shared-postgres`, `shared-citus`, `shared-telemetry`, `compose`, `template`)

**Orphan Handling**: Orphaned `configs/` directories produce warnings (not errors). Archived orphans go to `configs/orphaned/`.

**CLI Usage**: `cicd lint-deployments validate-mirror [deployments-dir configs-dir]`

#### 12.4.11 Validation Pipeline Architecture

**Execution Model**: All 8 validators run sequentially with aggregated error reporting. Each validator produces a `ValidationResult` with pass/fail status, error list, and timing metrics. The `validate-all` orchestrator collects all results and prints a unified summary.

```
┌─────────────────────────────────────────────────────────┐
│                  validate-all orchestrator               │
│                                                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────┐ │
│  │ naming   │→ │kebab-case│→ │  schema  │→ │template │ │
│  └──────────┘  └──────────┘  └──────────┘  └─────────┘ │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────┐ │
│  │  ports   │→ │telemetry │→ │  admin   │→ │ secrets │ │
│  └──────────┘  └──────────┘  └──────────┘  └─────────┘ │
│                                                         │
│  Result: N passed / M failed (Xms)                      │
└─────────────────────────────────────────────────────────┘
```

**8-Validator Reference**:

| # | Validator | Scope | Purpose | Key Rules |
|---|-----------|-------|---------|-----------|
| 1 | **ValidateNaming** | `deployments/`, `configs/` | Enforce kebab-case directory and file naming | All directories/files must be lowercase kebab-case; template directory skipped (intentional uppercase placeholders). |
| 2 | **ValidateKebabCase** | `configs/` YAML files | Enforce kebab-case in YAML keys and compose service names | Top-level YAML keys must use kebab-case; `x-` extension keys and `services` map entries validated. |
| 3 | **ValidateSchema** | `configs/` `config-*.yml` files | Validate service template config files against hardcoded schema | Required keys (`bind-public-protocol`, `bind-public-address`, `bind-public-port`, etc.); protocol must be `https`; addresses must be valid IP/hostname. |
| 4 | **ValidateTemplatePattern** | `deployments/template/` | Validate template directory naming, structure, and placeholder values | compose.yml must use `PRODUCT-SERVICE` placeholders; required directories/files present; secrets follow template naming. |
| 5 | **ValidatePorts** | `deployments/` compose files | Validate port assignments per deployment type | SERVICE: 8000-8999, PRODUCT: 18000-18999, SUITE: 28000-28999; admin always 9090; no port conflicts. |
| 6 | **ValidateTelemetry** | `configs/` YAML files | Validate OTLP endpoint consistency | OTLP endpoint required in service configs; hostname:port format (no protocol prefix); consistent collector naming. |
| 7 | **ValidateAdmin** | `deployments/` compose files | Validate admin endpoint bind policy | Admin port must bind to `127.0.0.1:9090` (never `0.0.0.0`); ensures admin API never exposed outside container. |
| 8 | **ValidateSecrets** | `deployments/` compose files | Detect inline secrets in environment variables | Secret-pattern env vars (PASSWORD, SECRET, TOKEN, KEY, API_KEY) must use Docker secrets (`/run/secrets/`); length threshold ≥32/43 chars for non-reference values. Infrastructure deployments excluded. |

**Cross-References**: Each validator is implemented in `internal/cmd/cicd/lint_deployments/validate_<name>.go` with comprehensive table-driven tests in `validate_<name>_test.go`. See code comments for detailed validation rules (per Decision 9:A minimal docs, comprehensive code comments).

### 12.5 Config File Architecture

**Purpose**: Centralized configuration management for all services with a consistent directory hierarchy mirroring the deployment structure.

**Schema Strategy**: Config file schema is HARDCODED in Go (`validate_schema.go`) with comprehensive code comments. No external schema files (e.g., JSON Schema, CONFIG-SCHEMA.md) are maintained. The validator source code is the single source of truth for schema rules.

**Directory Structure**:

```
configs/
├── ca/                          # PKI product (maps to pki-ca service)
│   ├── ca-server.yml            # CA-specific nested config
│   ├── ca-config-schema.yaml    # CA certificate schema
│   └── profiles/                # X.509 certificate profiles
│       ├── tls-server.yaml
│       └── root-ca.yaml

├── sm/
│   ├── kms/                     # SM KMS service configs
│   │   ├── config-pg-1.yml      # PostgreSQL instance 1 (flat kebab-case)
│   │   ├── config-pg-2.yml      # PostgreSQL instance 2 (flat kebab-case)
│   │   └── config-sqlite.yml    # SQLite development (flat kebab-case)
│   └── im/                      # SM IM service configs (renamed from sm-im)
│       ├── config-pg-1.yml      # PostgreSQL instance 1 (flat kebab-case)
│       ├── config-pg-2.yml      # PostgreSQL instance 2 (flat kebab-case)
│       └── config-sqlite.yml    # SQLite development (flat kebab-case)
├── identity/
│   ├── development.yml          # Environment-specific
│   ├── production.yml           # Environment-specific
│   ├── test.yml                 # Environment-specific
│   ├── policies/                # Shared authentication policies
│   │   ├── adaptive-auth.yml
│   │   └── step-up.yml
│   ├── profiles/                # Deployment profiles
│   │   ├── full-stack.yml
│   │   └── ci.yml
│   └── authz/                   # Per-service configs
│       └── authz.yml
├── jose/
│   └── jose-server.yml
├── cryptoutil/
│   └── cryptoutil.yml           # Suite-level config
└── orphaned/                    # Archived configs (no active deployment)
    └── template/                # Orphaned template configs
```

**File Types**:

| Type | Pattern | Schema | Example |
|------|---------|--------|---------|
| Service template config | `config-*.yml` | Flat kebab-case, validated by `ValidateSchema` | `config-pg-1.yml` |
| Domain-specific config | `{service}.yml` | Nested YAML, service-specific | `ca-server.yml`, `authz.yml` |
| Environment config | `{env}.yml` | Product-level deployment settings | `development.yml`, `production.yml` |
| Certificate profile | `profiles/*.yaml` | X.509 certificate definitions | `tls-server.yaml` |
| Auth policy | `policies/*.yml` | Authentication/authorization rules | `adaptive-auth.yml` |

**Cross-References**: Schema validation rules in [validate_schema.go](/internal/cmd/cicd/lint_deployments/validate_schema.go). Config naming in [Section 12.4.5](#1245-config-file-naming-strategy).

### 12.6 Secrets Management in Deployments

**Purpose**: Enforce Docker secrets usage for all credentials in compose files, preventing inline secret exposure in version-controlled YAML.

**Docker Secrets Pattern**: All secret-bearing environment variables (PASSWORD, SECRET, TOKEN, KEY, API_KEY) MUST reference Docker secrets via `/run/secrets/<name>` or `file:///run/secrets/<name>`. Inline values are violations.

**File Permissions**: All `.secret` files MUST have 440 (r--r-----) permissions. Never commit actual secret values to version control.

**Detection Strategy**: Length-based threshold (≥32 bytes / ≥43 base64 chars) identifies high-entropy inline values. Safe references (`/run/secrets/`, Docker secret names, short dev defaults) are excluded. No entropy calculation (too many false positives). Infrastructure deployments (Grafana, OTLP collector) are excluded from secrets validation since they use intentional inline dev credentials.

**Cross-References**: Secrets coordination strategy in [Section 12.3.3](#1233-secrets-coordination-strategy). Validator implementation in [validate_secrets.go](/internal/cmd/cicd/lint_deployments/validate_secrets.go).

### 12.7 Documentation Propagation Strategy

**Purpose**: Keep instruction files (`.github/instructions/`) synchronized with ARCHITECTURE.md using chunk-based verbatim copying of semantic units.

**Propagation Model**: ARCHITECTURE.md is the single source of truth. Instruction files contain compressed summaries with `See [ARCHITECTURE.md Section X.Y]` cross-references. When ARCHITECTURE.md sections change, corresponding instruction file sections MUST be updated.

**Mapping**:

| ARCHITECTURE.md Section | Instruction File |
|------------------------|------------------|
| 12.4 Deployment Structure Validation | 04-01.deployment.instructions.md |
| 12.5 Config File Architecture | 04-01.deployment.instructions.md |
| 12.6 Secrets Management | 02-05.security.instructions.md |
| 6.X Secrets Detection | 02-05.security.instructions.md |
| 9.7 CI/CD Workflow Architecture | 04-01.deployment.instructions.md |

**Semantic Units**: Propagation copies complete sections (not individual sentences). Each section is a self-contained semantic unit with purpose, rules, and cross-references.

### 12.8 Validator Error Aggregation Pattern

**Purpose**: All validators run to completion (never short-circuit) and aggregate errors for a single unified report.

**Execution Model**: Sequential execution of all 8 validators. Each validator produces a `ValidationResult` containing: valid/invalid status, error list, and execution duration. The orchestrator (`ValidateAll`) collects all results and produces a summary with pass/fail counts and total duration.

**Rationale**: Sequential execution (not parallel) ensures deterministic output ordering and simplifies debugging. Aggregated errors (not fail-fast) show ALL problems in one run, reducing fix-test-fix cycles.

**Exit Code**: `validate-all` returns exit code 0 if all validators pass, exit code 1 if any validator fails. CI/CD workflows use this to block merges on validation failures.

### 12.9 Environment Strategy

**Development**: SQLite in-memory, port 0, auto-generated TLS, disabled telemetry
**Testing**: test-containers (PostgreSQL), dynamic ports, ephemeral instances
**Production**: PostgreSQL (cloud), static ports, full telemetry, TLS required

### 12.10 Release Management

**Versioning**: Semantic versioning (major.minor.patch)
**Release Process**: Tag creation, CHANGELOG generation, artifact publishing
**Rollback Strategy**: Previous version stable, blue-green deployment

---

## 13. Development Practices

### 13.1 Coding Standards

**Go Best Practices**: Effective Go, Code Review Comments, Go Proverbs
**Project Patterns**: See [03-01.coding.instructions.md](../.github/instructions/03-01.coding.instructions.md) for file size limits, default values, conditional statements

### 13.2 Version Control

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

**Main branch**: Always stable, deployable, protected
**Feature branches**: Short-lived (<7 days), rebased on main before merge
**Release branches**: For production releases, cherry-pick hotfixes

### 13.4 Code Review

**Requirements**: 2+ approvals for core changes, 1 approval for docs/tests
**Checklist**: Tests added, docs updated, linting passes, security reviewed
**Size limits**: <500 lines ideal, >1000 lines requires breakdown

### 13.5 Development Workflow

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

#### 13.5.4 Docker Desktop Startup - CRITICAL

**MANDATORY**: Docker Desktop MUST be running before executing any Docker-dependent operations (E2E tests, Docker Compose, container builds).

**Cross-Platform Verification**:

```bash
# Check if Docker is running (all platforms)
docker ps

# Expected: list of containers or empty table
# Error: "Cannot connect to the Docker daemon" = Docker not running
```

**Platform-Specific Startup**:

**Windows**:
```powershell
# Start Docker Desktop
Start-Process -FilePath "C:\Program Files\Docker\Docker\Docker Desktop.exe"

# Wait for initialization (30-60 seconds)
Start-Sleep -Seconds 45

# Verify Docker is ready
docker ps
```

**macOS**:
```bash
# Start Docker Desktop
open -a Docker

# Wait for initialization (30-60 seconds)
sleep 45

# Verify Docker is ready
docker ps
```

**Linux**:
```bash
# Start Docker service (systemd)
sudo systemctl start docker

# Enable Docker to start on boot
sudo systemctl enable docker

# Verify Docker is ready
docker ps

# Alternative: Docker Desktop for Linux (if installed)
systemctl --user start docker-desktop
```

**Why Critical**: All workflow testing infrastructure, E2E tests, and Docker Compose operations require Docker daemon. Without Docker running:
- `docker ps` fails with "Cannot connect to the Docker daemon" error
- `docker compose up` fails with pipe/socket errors
- E2E tests skip with environmental warnings
- Integration test containers cannot start

See: [Section 11.2.5 CI/CD](#1125-cicd) for local workflow testing commands that require Docker.

---

## 14. Operational Excellence

### 14.1 Monitoring & Alerting

**Metrics**: Prometheus (HTTP, DB, crypto, keys)
**Logging**: Structured logs via OpenTelemetry
**Tracing**: Distributed traces via OTLP
**Dashboards**: Grafana LGTM (Loki, Tempo, Prometheus)

### 14.2 Incident Management

**Post-Mortem Template**: docs/P0.X-INCIDENT_NAME.md - Summary, root cause, timeline, impact, lessons, action items
**Severity Levels**: P0 (critical), P1 (high), P2 (medium), P3 (low)
**Response Time**: P0 <15min, P1 <1hr, P2 <1 day, P3 <1 week

### 14.3 Performance Management

**Benchmarks**: Crypto operations, HTTP endpoints, database queries
**Load Testing**: Gatling scenarios (baseline, peak, stress)
**Optimization**: Profile hot paths, caching strategies, connection pooling

### 14.4 Capacity Planning

**Resource Limits**: Memory, CPU, disk, network
**Scaling Triggers**: >70% utilization sustained >5min
**Horizontal Scaling**: Stateless services, PostgreSQL read replicas (future)

### 14.5 Disaster Recovery

**Backup Strategy**: PostgreSQL daily snapshots, 30-day retention
**Recovery Time**: RTO <4 hours, RPO <1 hour
**Testing**: Quarterly DR drills, documented runbooks

---

## Appendix A: Decision Records

### A.1 Architectural Decision Records (ADRs)

**ADR Template**:

- Title: ADR-NNNN-descriptive-name
- Status: Proposed, Accepted, Deprecated, Superseded
- Context: Problem statement, constraints, requirements
- Decision: Chosen approach with rationale
- Consequences: Trade-offs, benefits, risks

**Location**: docs/adr/

### A.2 Technology Selection Decisions

**Go 1.25.5**: Static typing, fast compilation, excellent concurrency, CGO-free (portability)
**PostgreSQL + SQLite**: Production (ACID, scalability) + Dev/Test (zero-config, in-memory)
**GORM**: Cross-DB compatibility, migrations, type-safe queries
**Fiber**: Fast HTTP framework, Express-like API, low memory footprint
**OpenTelemetry**: Vendor-neutral observability, OTLP standard, future-proof

### A.3 Pattern Selection Decisions

**Service Template**: Eliminates 48,000+ lines per service, ensures consistency
**Dual HTTPS**: Security (public vs admin), network isolation, health checks
**Multi-Tenancy**: Schema-level isolation (not row-level), compliance, performance
**Hierarchical Keys**: Defense in depth, key rotation, compliance (FIPS 140-3)

---

## Appendix B: Reference Tables

### B.1 Service Port Assignments

**See Section 3.2 Product-Service Port Assignments** for complete table

**Summary**: pki-ca (8050-8059), jose-ja (8060-8069), sm-im (8070-8079), sm-kms (8080-8089), identity-authz (8100-8109), identity-idp (8110-8119), identity-rs (8120-8129), identity-rp (8130-8139), identity-spa (8140-8149)

### B.2 Database Port Assignments

**See Section 3.4.2 PostgreSQL Ports** for complete table

**Summary**: Host ports 54320-54328 map to container port 5432 for 9 services

### B.3 Technology Stack

**Languages**: Go 1.25.5 (services), Python 3.14+ (utilities), Node v24.11.1+ (CLI tools)
**Databases**: PostgreSQL 18, SQLite (modernc.org/sqlite, CGO-free)
**Frameworks**: Fiber (HTTP), GORM (ORM), oapi-codegen (OpenAPI)
**Observability**: OpenTelemetry, Grafana LGTM (Loki, Tempo, Prometheus)
**Security**: FIPS 140-3 approved algorithms, Docker/Kubernetes secrets
**Testing**: testify, gremlins (mutation), Nuclei/ZAP (DAST), Gatling (load)

### B.4 Dependency Matrix

**Core Dependencies**:

- github.com/gofiber/fiber/v3 (HTTP framework)
- gorm.io/gorm (ORM)
- github.com/google/uuid/v7 (UUIDv7)
- go.opentelemetry.io/otel (telemetry)
- github.com/go-jose/go-jose/v4 (JOSE)

**Test Dependencies**: testify, testcontainers-go, httptest

### B.5 Configuration Reference

**Priority Order**: Docker secrets > YAML > CLI parameters (NO env vars for secrets)

**Standard Files**:

- config.yml: Main configuration
- secrets/*.secret: Credentials (chmod 440)

### B.6 Instruction File Reference

**See .github/copilot-instructions.md** for complete table of 18 instruction files

**Summary**: 01-terminology/beast-mode, 02-architecture (5 files), 03-development (4 files), 04-deployment (1 file), 05-platform (2 files), 06-evidence (2 files)

### B.7 Agent Catalog & Handoff Matrix

**Agents Available**:

- implementation-planning: Planning and task decomposition → hands off to implementation-execution
- implementation-execution: Autonomous implementation execution
- doc-sync: Documentation synchronization
- fix-workflows: Workflow repair and validation
- beast-mode: Continuous execution mode

**See .github/agents/*.agent.md** for complete agent definitions

### B.8 CI/CD Workflow Catalog

**See Section 9.7.1 Workflow Catalog** for complete list

**Summary**: ci-quality, ci-coverage, ci-mutation, ci-race, ci-benchmark, ci-sast, ci-dast, ci-e2e, ci-load, ci-gitleaks

### B.9 Reusable Action Catalog

**See Section 9.8.1 Action Catalog** for complete list

**Summary**: docker-images-pull (parallel image pre-fetching), cache-go (Go module cache)

### B.10 Linter Rule Reference

**See .golangci.yml** for complete configuration

**Summary**: 30+ linters enabled (errcheck, govet, staticcheck, wsl_v5, godot, gosec, etc.)

| File | Description |
|------|-------------|
| 01-01.terminology | RFC 2119 keywords (MUST, SHOULD, MAY, CRITICAL) |
| 01-02.beast-mode | Beast mode directive |
| 02-01.architecture | Products and services architecture patterns |
| ... | (25 total instruction files) |

### B.7 Agent Catalog & Handoff Matrix

| Agent | Description | Tools | Handoffs |
|-------|-------------|-------|----------|
| implementation-planning | Planning and task decomposition | edit, execute, read, search, web | → implementation-execution |
| implementation-execution | Autonomous implementation execution | edit, execute, read, search, web | → doc-sync, fix-workflows |
| doc-sync | Documentation synchronization | TBD | TBD |
| fix-workflows | Workflow repair and validation | TBD | TBD |
| beast-mode | Continuous execution mode | TBD | TBD |

### B.8 CI/CD Workflow Catalog

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

| Action | Description | Inputs | Outputs |
|--------|-------------|--------|---------|
| docker-images-pull | Parallel Docker image pre-fetching | images (newline-separated list) | None |
| Additional actions | TBD | TBD | TBD |

### B.10 Linter Rule Reference

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

**Status**: ALWAYS enabled, NEVER disabled

**Approved Algorithms**: RSA ≥2048, ECDSA (P-256/384/521), ECDH, EdDSA (25519/448), AES ≥128 (GCM, CBC+HMAC), SHA-256/384/512, HMAC-SHA256/384/512, PBKDF2, HKDF

**BANNED**: bcrypt, scrypt, Argon2, MD5, SHA-1, RSA <2048, DES, 3DES

### C.2 PKI Standards Compliance

**CA/Browser Forum Baseline Requirements**:

- Serial ≥64 bits CSPRNG, >0, <2^159
- Validity ≤398 days (subscriber), 5-10 years (intermediate), 20-25 years (root)
- Extensions: Key Usage (critical), EKU, SAN, AKI, SKI, CRL, OCSP
- Audit logging: 7-year retention

**Validation**: DV (<30 days), OV/EV (<13 months), CAA DNS checks

### C.3 OAuth 2.1 / OIDC 1.0 Compliance

**OAuth 2.1**: Authorization Code + PKCE, Client Credentials, Token Exchange
**OIDC 1.0**: ID Token (JWS), UserInfo endpoint, Discovery (.well-known/openid-configuration)

**Security**: HTTPS required, state parameter, nonce in ID tokens, consent tracking

### C.4 Security Standards Compliance

**OWASP**: Password Storage Cheat Sheet (peppering, PBKDF2), Top 10
**NIST**: FIPS 140-3 (crypto), SP 800-63 (digital identity)
**Zero Trust**: No caching authz, mTLS, least privilege, audit logging (90-day retention)

---

## Document Metadata

**Revision History**:
- v2.0 (2026-02-08): Initial skeleton structure
- v1.0 (historical): Original ARCHITECTURE.md

**Related Documents**:
- `.github/copilot-instructions.md` - Copilot configuration
- `.github/instructions/*.instructions.md` - Detailed instructions
- `docs/speckit/constitution.md` - Project constitution
- `docs/ARCHITECTURE.md` - Legacy architecture document

**Cross-References**:
- All sections maintain stable anchor links for referencing
- Machine-readable YAML frontmatter for metadata
- Consistent section numbering for navigation
