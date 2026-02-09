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

| Product | Service | Product-Service Identifier | Host Public Address | Host Port Range | Container Public Address | Container Public Port Range | Container Admin Private Address | Container Admin Port Range | Description |
|---------|----------------|-----------------|------------|----------|------------|----------------|-------------------|----------|
| **Private Key Infrastructure (PKI)** | **Certificate Authority (CA)** | **pki-ca** | 127.0.0.1 | 8050-8059 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | X.509 certificates, EST, SCEP, OCSP, CRL |
| **JSON Object Signing and Encryption (JOSE)** | **JWK Authority (JA)** | **jose-ja** | 127.0.0.1 | 8060-8069 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | JWK/JWS/JWE/JWT operations |
| **Cipher** | **Instant Messenger (IM)** | **cipher-im** | 127.0.0.1 | 8070-8079 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | E2E encrypted messaging, encryption-at-rest |
| **Secrets Manager (SM)** | **Key Management Service (KMS)** | **sm-kms** | 127.0.0.1 | 8080-8089 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | Elastic key management, encryption-at-rest |
| **Identity** | **Authorization Server (Authz)** | **identity-authz** | 127.0.0.1 | 8100-8109 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | OAuth 2.1 authorization server |
| **Identity** | **Identity Provider (IdP)** | **identity-idp** | 127.0.0.1 | 8110-8119 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | OIDC 1.0 Identity Provider |
| **Identity** | **Resource Server (RS)** | **identity-rs** | 127.0.0.1 | 8120-8129 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | OAuth 2.1 Resource Server |
| **Identity** | **Relying Party (RP)** | **identity-rp** | 127.0.0.1 | 8130-8139 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | OAuth 2.1 Relying Party |
| **Identity** | **Single Page Application (SPA)** | **identity-spa** | 127.0.0.1 | 8140-8149 | 0.0.0.0 | 8080 | 127.0.0.1 | 9090 | OAuth 2.1 Single Page Application |

#### 3.2.1 PKI Product

- Certificate Authority (CA): X.509 certificates, EST, SCEP, OCSP, CRL
- Product-Service Identifier: pki-ca
- Host Public Address: 127.0.0.1
- Container Public Address: 0.0.0.0
- Container Admin Private Address: 127.0.0.1
- Public Port Range: 8050-8059 (host), 8080 (container)
- Private Admin Port: 9090 (container only)

#### 3.2.2 JOSE Product

- JWK Authority (JA): JWK/JWS/JWE/JWT operations
- Product-Service Identifier: jose-ja
- Host Public Address: 127.0.0.1
- Container Public Address: 0.0.0.0
- Container Admin Private Address: 127.0.0.1
- Public Port Range: 8060-8069 (host), 8080 (container)
- Private Admin Port: 9090 (container only)

#### 3.2.3 Cipher Product

- Instant Messenger (IM): E2E encrypted messaging, encryption-at-rest
- Product-Service Identifier: cipher-im
- Host Public Address: 127.0.0.1
- Container Public Address: 0.0.0.0
- Container Admin Private Address: 127.0.0.1
- Public Port Range: 8070-8079 (host), 8080 (container)
- Private Admin Port: 9090 (container only)

#### 3.2.4 Secrets Manager (SM) Product

- Key Management Service (KMS): Elastic key management, encryption-at-rest
- Product-Service Identifier: sm-kms
- Host Public Address: 127.0.0.1
- Container Public Address: 0.0.0.0
- Container Admin Private Address: 127.0.0.1
- Public Port Range: 8080-8089 (host), 8080 (container)
- Private Admin Port: 9090 (container only)

#### 3.2.5 Identity Product

- Authorization Server (Authz): OAuth 2.1 authorization server
- Product-Service Identifier: identity-authz
- Host Public Address: 127.0.0.1
- Container Public Address: 0.0.0.0
- Container Admin Private Address: 127.0.0.1
- Public Port Range: 8100-8109 (host), 8080 (container)
- Private Admin Port: 9090 (container only)

- Identity Provider (IdP): OIDC 1.0 Identity Provider
- Product-Service Identifier: identity-idp
- Host Public Address: 127.0.0.1
- Container Public Address: 0.0.0.0
- Container Admin Private Address: 127.0.0.1
- Public Port Range: 8110-8119 (host), 8080 (container)
- Private Admin Port: 9090 (container only)

- Resource Server (RS): OAuth 2.1 Resource Server
- Product-Service Identifier: identity-rs
- Host Public Address: 127.0.0.1
- Container Public Address: 0.0.0.0
- Container Admin Private Address: 127.0.0.1
- Public Port Range: 8120-8129 (host), 8080 (container)
- Private Admin Port: 9090 (container only)

- Relying Party (RP): OAuth 2.1 Relying Party
- Product-Service Identifier: identity-rp
- Host Public Address: 127.0.0.1
- Container Public Address: 0.0.0.0
- Container Admin Private Address: 127.0.0.1
- Public Port Range: 8130-8139 (host), 8080 (container)
- Private Admin Port: 9090 (container only)

- Single Page Application (SPA): OAuth 2.1 Single Page Application
- Product-Service Identifier: identity-spa
- Host Public Address: 127.0.0.1
- Container Public Address: 0.0.0.0
- Container Admin Private Address: 127.0.0.1
- Public Port Range: 8140-8149 (host), 8080 (container)
- Private Admin Port: 9090 (container only)

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

| Product-Service Identifier | Host Address | Host Port | Container Address | Container Port |
|---------|-----------|----------------|----------|----------------|
| **pki-ca** | 127.0.0.1 | 54320 | 0.0.0.0 | 5432 |
| **jose-ja** | 127.0.0.1 | 54321 | 0.0.0.0 | 5432 |
| **cipher-im** | 127.0.0.1 | 54322 | 0.0.0.0 | 5432 |
| **sm-kms** | 127.0.0.1 | 54323 | 0.0.0.0 | 5432 |
| **identity-authz** | 127.0.0.1 | 54324 | 0.0.0.0 | 5432 |
| **identity-idp** | 127.0.0.1 | 54325 | 0.0.0.0 | 5432 |
| **identity-rs** | 127.0.0.1 | 54326 | 0.0.0.0 | 5432 |
| **identity-rp** | 127.0.0.1 | 54327 | 0.0.0.0 | 5432 |
| **identity-spa** | 127.0.0.1 | 54328 | 0.0.0.0 | 5432 |

#### 3.4.3 Telemetry Ports (Shared)

| Service | Host Port | Container Port | Protocol |
|---------|-----------|----------------|----------|
| opentelemetry-collector-contrib | 127.0.0.1:4317 | 0.0.0.0:4317 | OTLP gRPC (no TLS) |
| opentelemetry-collector-contrib | 127.0.0.1:4318 | 0.0.0.0:4318 | OTLP HTTP (no TLS) |
| grafana-otel-lgtm | 127.0.0.1:3000 | 0.0.0.0:3000 | HTTP (no TLS) |
| grafana-otel-lgtm | 127.0.0.1:4317 | 0.0.0.0:4317 | OTLP gRPC (no TLS) |
| grafana-otel-lgtm | 127.0.0.1:4318 | 0.0.0.0:4318 | OTLP HTTP (no TLS) |

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

#### 4.4.3 CLI Entry Points

```
cmd/
├── cryptoutil/main.go         # Suite-level CLI (all products): Thin main() call to `internal/apps/cryptoutil.go`
├── cipher/main.go             # Product-level Cipher CLI: Thin main() call to `internal/apps/cipher/cipher.go`
├── jose/main.go               # Product-level JOSE CLI: Thin main() call to `internal/apps/jose/jose.go`
├── pki/main.go                # Product-level PKI CLI: Thin main() call to `internal/apps/pki/pki.go`
├── identity/main.go           # Product-level Identity CLI: Thin main() call to `internal/apps/identity/identity.go`
├── sm/main.go                 # Product-level SM CLI: Thin main() call to `internal/apps/sm/sm.go`
├── cipher-im/main.go          # Service-level Cipher-IM CLI: Thin main() call to `internal/apps/cipher/im/im.go`
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
2. `cmd/<product>/` for product-level CLI
```go
func main() {
    os.Exit(cryptoutilApps<PRODUCT>.<PRODUCT>(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
```
3. `cmd/<product>/<service>/` for service-level CLI
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
├── cipher/
│   └── im/                    # Cipher-IM service
│       ├── domain/            # Domain models (Message, Recipient)
│       ├── repository/        # Domain repos + migrations (2001+)
│       ├── server/            # CipherIMServer, PublicServer
│       │   ├── config/        # CipherImServerSettings embeds template
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
└── cipher/
    └── ... (same structure)
```

#### 4.4.7 CLI Patterns

### CLI Hierarchy

```
# Product-Service pattern (preferred)
cipher-im server --config=/etc/cipher/im.yml

# Service pattern
im server --config=/etc/cipher/im.yml

# Product pattern (routes to service)
cipher im server --config=/etc/cipher/im.yml

# Suite pattern (routes to product, then service)
cryptoutil cipher im server --config=/etc/cipher/im.yml
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
| `server` | CLI server start with dual HTTPS listeners, for Private Admin APIs vs Public Business Logic APIs |
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

[To be populated]

#### 5.1.1 Template Components

- Two HTTPS Listeners: Public (business APIs) + Admin (health checks)
- Two Public Paths: /browser/** (sessions) vs /service/** (tokens)
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
- Migration priority: cipher-im FIRST (validation) → jose-ja → pki-ca → identity services → sm-kms LAST

### 5.2 Service Builder Pattern

[To be populated]

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
                └── Domain Data (encrypted-at-rest with content key) - Examples: Cipher-IM messages, SM-KMS JWKs, JOSE-JA JWKs, PKI-CA private keys, Identity user credentials
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
7. Every service is RECOMMENDED to include at least one file-based factor realm for fallback session creation, plus at least one file-based session realm for session use.

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
| 2001+ | Domain | cipher-im messages (2001), jose JWKs (2001) |

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
func NewTestSettings() *CipherImServerSettings {
    return &CipherImServerSettings{
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

### 10.3 Integration Testing Strategy

[To be populated]

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

[To be populated]

#### 10.4.1 Docker Compose Orchestration

**Use ComposeManager for E2E testing**:

```go
func TestE2E_SendMessage(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    ctx := context.Background()

    // Start Docker Compose stack
    manager := e2e.NewComposeManager(t, "../../../deployments/cipher-im")
    manager.Up(ctx)
    defer manager.Down(ctx)

    // Wait for service healthy
    manager.WaitForHealthy(ctx, "cipher-im", 60*time.Second)

    // Get TLS-enabled HTTP client
    client := manager.HTTPClient()

    // Test API
    resp, err := client.Post(
        manager.ServiceURL("cipher-im") + "/browser/api/v1/messages",
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

#### 11.2.1 MANDATORY Pre-Commit Quality Gates

`golangci-lint run --fix` → zero warnings
`go build ./...` → clean build
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
If not running, start it via command line.

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
  * ❌ Modify comments in enforce_any.go without reading full package context
  * ❌ Change backticked `interface{}` to `any` in format_go package
  * ❌ Refactor code in isolation (single-file view)
  * ❌ Simplify "verbose" CRITICAL comments
- ALWAYS DO:
  * ✅ Read complete package context before refactoring self-modifying code
  * ✅ Check for CRITICAL/SELF-MODIFICATION tags in comments
  * ✅ Verify self-exclusion patterns exist and are respected
  * ✅ Run tests after ANY changes to format_go package

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

```yaml
healthcheck:
  test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null",
         "https://127.0.0.1:9090/admin/api/v1/livez"]
  start_period: 60s
  interval: 5s
```

**Use wget (Alpine), 127.0.0.1 (not localhost), port 9090.**

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
