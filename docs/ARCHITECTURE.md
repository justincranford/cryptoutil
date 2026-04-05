---
title: cryptoutil Architecture - Single Source of Truth
version: 2.0
date: 2026-03-29
status: Draft
audience:
  - Copilot Instructions
  - Copilot Agents
  - Copilot Skills
  - Copilot Prompts
  - Development Team
  - Technical Stakeholders
references:
  - .github/copilot-instructions.md
  - .github/instructions/*.instructions.md
  - .github/agents/*.agent.md
  - .claude/agents/*.md
  - .github/skills/NAME/SKILL.md
  - .github/prompts/NAME.prompt.md
  - .github/workflows/*.yml
  - .github/actions/NAME/action.yml
maintainers:
  - cryptoutil Development Team
tags:
  - architecture
  - strategy
  - design
  - implementation
  - testing
  - security
  - compliance
  - automation
---

# cryptoutil Architecture - Single Source of Truth

**Last Updated**: March 29, 2026
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
- [13. Deployment Tooling & Validation](#13-deployment-tooling--validation)
- [14. Development Practices](#14-development-practices)
- [15. Operational Excellence](#15-operational-excellence)
- [Appendix A: Decision Records](#appendix-a-decision-records)
- [Appendix B: Reference Tables](#appendix-b-reference-tables)
- [Appendix C: Compliance Matrix](#appendix-c-compliance-matrix)

### Cross-Reference Index

| Topic | Primary Section | Also Referenced In |
|-------|----------------|-------------------|
| Secrets management | [12.3.3](#1233-secrets-coordination-strategy) | [6.10](#610-secrets-detection-strategy), [13.3](#133-secrets-management-in-deployments), [4.4.6](#446-deployments) |
| TLS / mTLS | [6.11](#611-tls-certificate-configuration) | [5.3](#53-dual-https-endpoint-pattern), [6.5](#65-pki-architecture--strategy) |
| Port assignments | [3.4](#34-port-assignments--networking) | [3.4.1](#341-port-design-principles), [12.3.4](#1234-multi-level-deployment-hierarchy) |
| Health checks | [5.5](#55-health-check-patterns) | [10.3.5](#1035-cross-service-contract-test-pattern) |
| Testing database tiers | [10.1](#101-testing-strategy-overview) | [7.3](#73-dual-database-strategy) |
| @propagate system | [13.4](#134-documentation-propagation-strategy) | [13.4.7](#1347-propagation-coverage-accounting) |
| Key rotation | [6.4.2](#642-key-hierarchy-barrier-service) | [6.7](#67-key-management-system-architecture) |
| Multi-tenancy | [7.2](#72-multi-tenancy-architecture--strategy) | [2.2](#22-architecture-strategy) |
| FIPS 140-3 | [6.1](#61-fips-140-3-compliance-strategy) | [6.4.1](#641-fips-140-3-compliance-always-enabled) |
| Fitness linters | [9.11.1](#9111-fitness-sub-linter-catalog) | [9.11](#911-architecture-fitness-functions) |

### Document Conventions

#### RFC 2119 Keywords

<!-- @propagate to=".github/instructions/01-01.terminology.instructions.md" as="rfc-2119-keywords" -->
- **MUST** = **REQUIRED** = **MANDATORY** = **SHALL** - Absolute requirement
- **MUST NOT** = **SHALL NOT** - Absolute prohibition
- **SHOULD** = **RECOMMENDED** - Highly desirable (may ignore with justification)
- **SHOULD NOT** = **NOT RECOMMENDED** - Not advisable (may do with justification)
- **MAY** = **OPTIONAL** - Truly optional (implementer decides)
<!-- @/propagate -->

#### Emphasis Keywords

<!-- @propagate to=".github/instructions/01-01.terminology.instructions.md" as="emphasis-keywords" -->
- **CRITICAL** - Historically regression-prone areas requiring extra attention
- **ALWAYS** / **NEVER** - Emphatic MUST / MUST NOT (no exceptions)
<!-- @/propagate -->

#### Abbreviations

<!-- @propagate to=".github/instructions/01-01.terminology.instructions.md" as="abbreviations" -->
**CRITICAL: NEVER use ambiguous `auth` abbreviation to mean either authentication or authorization**

- **authn** = Authentication
- **authz** = Authorization

**Rationale**: Prevents confusion filenames, variable names, and documentation.
<!-- @/propagate -->

---

## 1. Executive Summary

### 1.1 Vision Statement

**cryptoutil** is a production-ready suite of five cryptographic-based products, designed with enterprise-grade security, **FIPS 140-3** standards compliance, Zero-Trust principles, and security-on-by-default:

1. **Private Key Infrastructure (PKI)** - X.509 certificate management with EST, SCEP, OCSP, and CRL support
2. **JSON Object Signing and Encryption (JOSE)** - JWK/JWS/JWE/JWT cryptographic operations
3. **Secrets Manager (SM)** - Elastic key management service with hierarchical key barriers; also hosts the encrypted messaging service
4. **Identity** - OAuth 2.1, OIDC 1.0, WebAuthn, and Passkeys authentication and authorization
5. **Skeleton** - Best-practice stereotype product-service template for service-framework usage reference

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

See [Section 11.1 Maximum Quality Strategy](#111-maximum-quality-strategy---mandatory) for complete quality attributes (NO EXCEPTIONS). Evidence-based validation is mandatory for all task completion.

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
- **Secret Management**: 100% Docker secrets (zero inline credentials, zero environment variable credentials)
- **TLS Configuration**: TLS 1.3+ only, full certificate chain validation, for protect-in-transit
- **JWE/JWS Configuration**: JOSE+JWT for protection-at-rest
- **Authentication**: Multi-factor support across all service
- **Authorization**: OAuth 2.1 access control with least privilege
- **Identification**: OIDC 1.0 identity

#### Operational Excellence

- **Availability**: Health checks respond <100ms
- **Observability**: 100% operations traced and logged
- **Documentation**: Every feature documented in OpenAPI specs
- **CI/CD**: All workflows passing, <15 min total pipeline time

### 1.5 Architecture at a Glance

**Entity Hierarchy** (parameterized naming conventions used throughout this document):

| Parameter | Meaning | Count | Values |
|-----------|---------|-------|--------|
| `{SUITE}` | Suite name | 1 | `cryptoutil` |
| `{PRODUCT}` | Product name | 5 | `identity`, `jose`, `pki`, `skeleton`, `sm` |
| `{PS-ID}` | Product-Service Identifier | 10 | `identity-authz`, `identity-idp`, `identity-rp`, `identity-rs`, `identity-spa`, `jose-ja`, `pki-ca`, `skeleton-template`, `sm-im`, `sm-kms` |
| `{PS_ID}` | Underscore variant (SQL, secrets) | 10 | Same as `{PS-ID}` with `_` replacing `-` |
| `{INFRA-TOOL}` | Infrastructure tooling | 2 | `cicd-lint`, `cicd-workflow` |

**1 Suite → 5 Products → 10 Services**:

```
cryptoutil (suite)
├── PKI           → pki-ca
├── JOSE          → jose-ja
├── SM            → sm-kms, sm-im
├── Identity      → identity-authz, identity-idp, identity-rs, identity-rp, identity-spa
└── Skeleton      → skeleton-template
```

**Deployment Tiers** (independent, stackable — see [Section 3.4.1](#341-port-design-principles)):

| Tier | Scope | Host Port Offset | compose.yml Location |
|------|-------|------------------|---------------------|
| SERVICE | Single service | +0 (8XXX) | `deployments/{PS-ID}/` |
| PRODUCT | All services in one product | +10000 (18XXX) | `deployments/{PRODUCT}/` |
| SUITE | All 10 services across all 5 products | +20000 (28XXX) | `deployments/{SUITE}-suite/` |

**Service Independence**: Each `{PS-ID}` service is a standalone binary with its own HTTPS listeners (public :8080, admin :9090), database (PostgreSQL or SQLite), config (`configs/{PS-ID}/`), Docker Compose file (`deployments/{PS-ID}/compose.yml`), and deployment secrets (`deployments/{PS-ID}/secrets/`). Services communicate via mTLS or OAuth 2.1 client credentials (see [Section 3.3](#33-product-service-relationships)).

**Product & Suite Convenience**: Each `{PRODUCT}` product and `{SUITE}` suite are also available as all-in-one binaries for optional, alternate packaging for ease of distribution. They also come with convenience `{PRODUCT}` product and `{SUITE}` suite Docker Compose files for ease of e2e testing.

**Migration Priority** (to service framework — see [Section 5.1.3](#513-mandatory-usage)): sm-im → jose-ja → sm-kms → pki-ca → identity services. SM services (sm-im/jose-ja/sm-kms) migrate first; pki-ca second; identity last.

**Federation**: Services fail over through FEDERATED → DATABASE → FILE realms with no retry logic or circuit breakers. FILE realms (local, always available) are the last-resort failsafe.

---

## 2. Strategic Vision & Principles

### 2.1 Agent Orchestration Strategy

#### Copilot Customization Types Decision Matrix

VS Code Copilot supports exactly 3 customization file types:

| Type | Pattern | Trigger | Best For |
|------|---------|---------|----------|
| **Instructions** | `.github/instructions/*.instructions.md` | Always loaded automatically | Passive context, standards, constraints |
| **Agents (Copilot)** | `.github/agents/*.agent.md` | `/agent-name` invocation | Complex multi-step autonomous tasks — Copilot canonical with explicit `tools:` whitelist |
| **Agents (Claude Code)** | `.claude/agents/*.md` | `/agent-name` invocation | Same agents, Claude Code canonical — omits `tools:` (inherits all tools) |
| **Skills** | `.github/skills/NAME/SKILL.md` | `/skill-name` slash command or auto-loaded | On-demand templates, code generation, analysis |

**Dual Canonical Strategy**: `.github/agents/*.agent.md` is the Copilot-authoritative source (with `tools:` whitelist, `handoffs:`, `skills:`). `.claude/agents/*.md` is the Claude Code-authoritative source (omits `tools:` — Claude inherits all tools). Both files must be kept in sync when agent content changes. Use `/agent-scaffold` to create both simultaneously.

See [Section 2.1.5 Copilot Skills](#215-copilot-skills) for skill catalogue and `.github/skills/` organization.

#### 2.1.1 Agent Architecture

- Agent isolation principle (agents do NOT inherit copilot instructions)
- **Dual canonical files**: `.github/agents/*.agent.md` (Copilot) and `.claude/agents/*.md` (Claude Code) — both must exist and stay in sync
- **Copilot format** (`.github/agents/*.agent.md`): YAML frontmatter with `name`, `description`, `tools` (whitelist — required for full tool access), `handoffs`, `argument-hint`
- **Claude Code format** (`.claude/agents/*.md`): YAML frontmatter with `name`, `description`, `argument-hint` only — omit `tools:` so Claude inherits all tools; Copilot-only fields (`handoffs`, `skills`) not applicable
- Autonomous execution mode patterns
- Quality over speed enforcement

**Implementation Plan File Structure**:

Implementation plans use the following files in `<work-dir>/`:

**Core** (created by implementation-planning, updated by implementation-execution):
- `plan.md` — Phase plan with scope, LOE, rationale, and constraints
- `tasks.md` — Task breakdown with checkbox tracking (updated continuously during execution)
- `lessons.md` — Phase post-mortem lessons: what worked, what didn't, root causes, patterns observed (scaffold created by planning, populated after each phase by execution)

**Ephemeral** (temporary, session-scoped):
- `quizme-v#.md` — Unknowns clarification during planning only (A-D options + E blank; deleted after answers merged)

<!-- @propagate to=".github/instructions/06-02.agent-format.instructions.md" as="agent-self-containment" -->
**Agent Self-Containment Checklist** (MANDATORY):
- Agents generating implementation plans MUST reference ARCHITECTURE.md testing (Section 10), quality gates (Section 11), coding standards (Section 14)
- Agents modifying code MUST reference coding standards (Sections 11, 14)
- Agents modifying deployments MUST reference deployment architecture (Sections 12, 13)
- Agents modifying CI/CD workflows or infrastructure MUST reference infrastructure architecture (Section 9)
- Agents modifying documentation or copilot artifacts (skills, instructions, agents) MUST reference Section 2.1 (Agent/Skill/Instruction catalog) and Section 13.4 (Documentation Propagation)
- ALL agents MUST reference Section 2.5 (Quality Strategy) for coverage and mutation targets
- Agents with ZERO ARCHITECTURE.md references are NON-COMPLIANT and MUST be updated
<!-- @/propagate -->

#### 2.1.2 Agent Catalog

**Naming convention**: Copilot agents are named `copilot-NAME`; Claude Code agents are named `claude-NAME`. Invoke via `@copilot-NAME` in Copilot Chat or `/claude-NAME` in Claude Code.

| Agent (Copilot) | Agent (Claude Code) | Description |
|----------------|--------------------|--------------|
| `copilot-implementation-planning` | `claude-implementation-planning` | Planning and task decomposition, quizme, and lessons.md scaffold |
| `copilot-implementation-execution` | `claude-implementation-execution` | Autonomous implementation execution, plan.md, tasks.md (phases and tasks and post-mortems and lessons.md), lessons.md updates |
| `copilot-fix-workflows` | `claude-fix-workflows` | Workflow repair and validation |
| `copilot-beast-mode` | `claude-beast-mode` | Continuous execution mode |
| Explore (built-in) | Explore (built-in) | Fast read-only codebase exploration and Q&A subagent (quick/medium/thorough) |

**Drift prevention**: `cicd-lint lint-docs` runs two drift sub-linters:
- **`lint-agent-drift`**: enforces that each Copilot agent (`copilot-NAME.agent.md`) has a matching Claude agent (`claude-NAME.md`) with identical `description:`, `argument-hint:`, and body. Only `name:` prefix and Copilot-only fields (`tools:`, `handoffs:`, `skills:`) may differ.
- **`lint-skill-command-drift`**: enforces that each Copilot skill in `.github/skills/NAME/SKILL.md` has a corresponding Claude Code slash command in `.claude/commands/NAME.md` that references the skill path string. Detects missing commands, missing skill path references, and orphaned commands.

#### 2.1.3 Agent Handoff Flow

- Planning → Implementation → Fix handoff chains
- Explicit handoff triggers and conditions
- State preservation across handoffs

#### 2.1.4 Instruction File Organization

- Hierarchical numbering scheme (01-01 through 07-01)
- Auto-discovery and alphanumeric ordering
- Single responsibility per file
- Cross-reference patterns
- Propagations from this document (i.e. docs/ARCHITECTURE.md)

#### 2.1.5 Copilot Skills

Skills live in `.github/skills/NAME/SKILL.md` — each skill in its own subdirectory where the directory name matches the `name` field in the SKILL.md YAML frontmatter. Invoked via `/skill-name` slash command or auto-loaded by Copilot when the request matches the skill description. See [VS Code Agent Skills reference](https://code.visualstudio.com/docs/copilot/customization/agent-skills).

**SKILL.md Frontmatter Requirements**: `name` (required, matches directory name, max 64 chars, lowercase-hyphens), `description` (required, max 1024 chars, specific about both capabilities and use cases), `argument-hint` (optional, hint shown in chat input), `user-invocable` (optional, defaults true; set false to hide from / menu), `disable-model-invocation` (optional, defaults false; set true to require manual /skill invocation only). The `metadata:` sub-key is NOT a valid SKILL.md frontmatter field and MUST NOT be used.

**Claude Code slash commands**: Each Copilot skill has a corresponding Claude Code slash command file at `.claude/commands/NAME.md` that references the skill path string `".github/skills/NAME/SKILL.md"` in its body. The `lint-skill-command-drift` sub-linter (part of `cicd-lint lint-docs`) enforces this 1:1 correspondence — missing commands, missing skill references, and orphaned commands all produce errors.

**Claude Command Frontmatter Requirements** (`.claude/commands/NAME.md`): YAML frontmatter (`---`) is REQUIRED. Fields: `name` (bare skill name, NOT the `claude-` prefix — e.g., `test-table-driven` not `claude-test-table-driven`), `description` (IDENTICAL to the corresponding Copilot skill's `description`), `argument-hint` (IDENTICAL to the Copilot skill's `argument-hint` when the skill has one). NEVER include `disable-model-invocation` — that field is Copilot-ONLY. The `lint-skill-command-drift` linter validates frontmatter presence, `description` match, and `argument-hint` match.

**Key Rules Section**: Both `.github/skills/NAME/SKILL.md` AND `.claude/commands/NAME.md` MUST contain a `## Key Rules` section with the essential rules for using the skill/command correctly. The linter enforces this requirement and errors if either file is missing the section.

**Skill Catalogue**:

| Skill | Domain | Purpose | File |
|-------|--------|---------|------|
| `test-table-driven` | testing | Generate table-driven Go tests (t.Parallel, UUIDv7 data, subtests) | [SKILL.md](.github/skills/test-table-driven/SKILL.md) |
| `test-fuzz-gen` | testing | Generate `_fuzz_test.go` (15s fuzz time, corpus examples, build tags) | [SKILL.md](.github/skills/test-fuzz-gen/SKILL.md) |
| `test-benchmark-gen` | testing | Generate `_bench_test.go` (mandatory for crypto, reset timer pattern) | [SKILL.md](.github/skills/test-benchmark-gen/SKILL.md) |
| `coverage-analysis` | testing | Analyze coverprofile output, categorize uncovered lines, suggest tests | [SKILL.md](.github/skills/coverage-analysis/SKILL.md) |
| `fips-audit` | security | Detect FIPS 140-3 violations and provide fix guidance | [SKILL.md](.github/skills/fips-audit/SKILL.md) |
| `openapi-codegen` | api | Generate three oapi-codegen configs (server/model/client) + OpenAPI spec skeleton | [SKILL.md](.github/skills/openapi-codegen/SKILL.md) |
| `migration-create` | data | Create numbered golang-migrate SQL files (template 1001-1999, domain 2001+) | [SKILL.md](.github/skills/migration-create/SKILL.md) |
| `new-service` | architecture | Guide service creation from skeleton-template: copy, rename, register, migrate, test | [SKILL.md](.github/skills/new-service/SKILL.md) |
| `propagation-check` | docs | Detect @propagate/@source drift, generate corrected @source blocks | [SKILL.md](.github/skills/propagation-check/SKILL.md) |
| `contract-test-gen` | testing | Generate cross-service contract compliance tests for framework behavioral contracts | [SKILL.md](.github/skills/contract-test-gen/SKILL.md) |
| `fitness-function-gen` | tooling | Create new architecture fitness function (linter) for lint-fitness framework | [SKILL.md](.github/skills/fitness-function-gen/SKILL.md) |
| `agent-scaffold` | tooling | Create both `.github/agents/NAME.agent.md` (Copilot, with `tools:`) and `.claude/agents/NAME.md` (Claude Code, without `tools:`) with all mandatory sections | [SKILL.md](.github/skills/agent-scaffold/SKILL.md) |
| `instruction-scaffold` | tooling | Create conformant `.github/instructions/NN-NN.name.instructions.md` | [SKILL.md](.github/skills/instruction-scaffold/SKILL.md) |
| `skill-scaffold` | tooling | Create conformant `.github/skills/NAME/SKILL.md` with proper YAML frontmatter | [SKILL.md](.github/skills/skill-scaffold/SKILL.md) |

#### 2.1.6 Agent Tool Discovery

**Four tool sources** — each requires a different discovery method.

<!-- @propagate to=".github/instructions/06-02.agent-format.instructions.md" as="agent-tool-discovery" -->
**Tool discovery by source type**:

| Source | How to Discover | Tool ID Format in Agent `tools:` |
|--------|----------------|----------------------------------|
| Built-in documented | [VS Code Agent Tools doc](https://code.visualstudio.com/docs/copilot/agents/agent-tools) | `category/toolReferenceName` (e.g., `edit/createFile`) |
| Built-in undocumented *(u)* | Empirical: check deferred tools list in active agent session | `category/toolReferenceName` (e.g., `web/githubRepo`) |
| Extension tools | Scan `~/.vscode/extensions/*/package.json` for `contributes.languageModelTools` | `toolReferenceName` (camelCase); use `name` (snake_case) if no `toolReferenceName` |
| MCP server tools | `%APPDATA%\Code\User\mcp.json` or `.vscode/mcp.json` | Tool name as shown in MCP server config |

**Extension scan script** (Python — cross-platform):

```python
import json, pathlib

ext_dir = pathlib.Path.home() / ".vscode" / "extensions"
for d in sorted(ext_dir.iterdir()):
    pkg = d / "package.json"
    if pkg.is_file():
        data = json.loads(pkg.read_text(encoding="utf-8"))
        tools = data.get("contributes", {}).get("languageModelTools")
        if tools:
            print(f"=== {d.name} ===")
            for t in tools:
                print(f"  name={t.get('name', '')}  toolReferenceName={t.get('toolReferenceName', '')}")
```

**Category disambiguation**: `github.copilot-chat` extension tools use `category/toolReferenceName` (categories: `agent`, `browser`, `edit`, `execute`, `read`, `search`, `vscode`, `web`). All other extensions use bare `toolReferenceName`.

**Maintenance**: Re-run the extension scan after any VS Code update, extension install/update, or MCP server change.
<!-- @/propagate -->

### 2.2 Architecture Strategy

#### Service Framework Pattern

- **Single Reusable Template**: All 10 services across 5 products inherit from `internal/apps/framework/`
- **Eliminates 48,000+ lines per service**: TLS setup, dual HTTPS servers, database, migrations, sessions, barrier
- **Merged Migrations**: Template (1001-1999) + Domain (2001+) for golang-migrate validation
- **Builder Pattern**: Fluent API with `NewServerBuilder(ctx, cfg).WithDomainMigrations(...).Build()`

#### Microservices Architecture

- **10 Services across 5 Products**: Independent deployment, scaling, and lifecycle
- **Dual HTTPS Endpoints**: Public (0.0.0.0:8080) for business APIs, Private (127.0.0.1:9090) for admin operations
- **Service Discovery**: Config file → Docker Compose → Kubernetes DNS (no caching)
- **Multi-Level Failover**: Services attempt credential validators in priority order (FEDERATED → DATABASE → FILE), with FILE realms as CRITICAL failsafe guaranteeing admin access

#### Multi-Tenancy

- **Schema-Level Isolation**: Each tenant gets separate schema (`tenant_<uuid>.users`)
- **tenant_id Scoping**: ALL data access filtered by tenant_id (not realm_id)
- **Realm-Based Authentication**: Authentication context only, NOT for data filtering
- **Registration Flow**: Create new tenant OR join existing tenant

#### Database Strategy

- **Dual Database Support**: PostgreSQL (production, e2e) + SQLite (dev/integration)
- **Cross-DB Compatibility**: UUID as TEXT, serializer:json for arrays, NullableUUID for optional FKs
- **GORM Always**: Never raw database/sql for consistency
- **TestMain Pattern**: Heavyweight resources initialized once per package; integration package starts one instance of service with SQLite database, for use by all tests within that package
- **Docker Compose Pattern**: Each service has its own `docker-compose.yml` for e2e testing, with 2x applications using shared PostgreSQL instance and 2x applications using separate in-memory SQLite instances; separate compose files for product-level and suite-level deployments

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

See [Section 11.1 Maximum Quality Strategy](#111-maximum-quality-strategy---mandatory) for the 8 quality attributes enforced without exception. Key principles:

- **Quality over speed**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- **Evidence-based validation**: Objective evidence required before marking any task complete
- **All issues are blockers**: Fix immediately, NEVER defer ("fix later"), NEVER skip
- **Continuous improvement**: Post-mortem on every phase, extract lessons to permanent homes

#### 2.3.2 Autonomous Execution Principles

See the [beast-mode agent](.github/agents/beast-mode.agent.md) for the full autonomous execution contract:

- Continuous work until ALL tasks complete or user clicks STOP
- Zero permission requests, zero status updates between tasks
- Blocker documentation with parallel unblocked work
- Mandatory review passes (see [Section 2.5](#25-quality-strategy)) before task completion

### 2.4 Implementation Strategy

#### Go Best Practices

- **Go Version**: 1.26.1+ (same everywhere: dev, CI/CD, Docker)
- **CGO Ban**: CGO_ENABLED=0 (except race detector) for maximum portability
- **Import Aliases**: `cryptoutil<Package>` for internal, `<vendor><Package>` for external
- **Magic Values**: `internal/shared/magic/magic_*.go` for shared, package-specific for domain

#### Testing Strategy

- **Table-Driven Tests**: ALWAYS use for multiple test cases (NOT standalone functions)
- **app.Test() Pattern**: ALL HTTP handler tests use in-memory testing (NO real servers)
- **TestMain Pattern**: Heavyweight resources (SQLite in-memory databases, application servers) initialized once per package; PostgreSQL initialized indirectly ONLY by E2E tests via Docker Compose
- **Dynamic Test Data**: UUIDv7 for all test values (thread-safe, process-safe, time-ordered)
- **t.Parallel()**: ALWAYS use in test functions and subtests for concurrency validation

#### Incremental Commits

- **Conventional Commits**: `<type>[scope]: <description>` format mandatory
- **Commit Strategy**: Incremental commits (NOT amend) preserve history for bisect
- **Restore from Clean**: When fixing regressions, restore known-good baseline first
- **Quality Gates**: Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`), linting clean (`golangci-lint run --fix` AND `golangci-lint run --build-tags e2e,integration`), tests pass (`go test ./... -shuffle=on`), deployment validators pass (when `deployments/` or `configs/` changed)

#### Continuous Execution

- **Beast Mode**: Work autonomously until problem completely solved
- **Maximum Quality Strategy**: All 8 quality attributes enforced — see [Section 11.1](#111-maximum-quality-strategy---mandatory) (NO EXCEPTIONS)
- **NO Stopping**: Task complete → Commit → IMMEDIATELY start next task (zero pause)
- **Blocker Handling**: Document blocker, switch to unblocked tasks, return when resolved
- **Mandatory Review Passes**: Minimum 3 review passes before marking any task complete — see [Section 2.5](#25-quality-strategy)

##### Autonomous Execution Principles

- **Quality Attributes**: Correctness (functionally correct), Completeness (all tasks done), Thoroughness (evidence-based validation), Reliability (quality gates enforced), Efficiency (maintainable/performance-optimized), Accuracy (root cause fixes)
- **Prohibited Stop Behaviors**: No status summaries, no "ready to proceed" questions, no strategic pivots with handoff, no time/token justifications, no pauses between tasks, no asking permission, no leaving uncommitted changes, no ending with analysis, no celebrations followed by stopping, no premature completion claims, no "current task done, moving to next" announcements
- **Execution Workflow**: Complete task → Commit → Next tool invocation (zero text, zero questions); Todo list empty → Check tracking docs → Find next incomplete task → Start immediately; All tasks done/blocked → Find quality improvements → Scan for technical debt → Review recent commits → Ask user if nothing left
- **Completion Verification Checklist**: Build clean, linting clean, tests pass (100%, zero skips), coverage maintained, mutation testing passes, evidence exists, git commit ready; After substantive change, run relevant build/tests/linters, validate code works (fast, minimal input), provide optional fenced commands for larger runs; Fix failures up to three targeted fixes, summarize root cause if still failing
- **Blocker Resolution**: Document in tracking doc, continue with ALL unblocked tasks, maximize progress, return to blocker when resolved; NO waiting for external dependencies

##### End-of-Turn Commit Protocol

<!-- NOTE: This @propagate target is the beast-mode instruction file, which is injected as modeInstructions at runtime (not via the standard instructions directory scan). This means the chunk is consumed in the mode prompt, not in the standard instructions context — a different injection path than all other @propagate targets. -->
<!-- @propagate to=".github/instructions/01-02.beast-mode.instructions.md" as="end-of-turn-commit-protocol" -->
**MANDATORY: NEVER end a turn with uncommitted changes. Your ABSOLUTE LAST TOOL INVOCATION before yielding to the user MUST be running `git status --porcelain`. NEVER assume the worktree is clean — always RUN the command as a tool call.**

If `git status --porcelain` returns ANY output:

1. Stage all changes: `git add -A`
2. Commit with a conventional commit message: `git commit -m "type(scope): description"`
3. Verify clean: `git status --porcelain` MUST return empty
4. Only then end the turn

**Critical violations** (any one = turn is NOT complete):

- Ending a turn with `M`, `A`, `D`, or `?` in `git status --porcelain` output
- Stopping after analysis without committing changes
- Marking work "done" with unstaged or uncommitted files
- Responding to the user without committing work in progress
- Assuming the worktree is clean without running the command as a tool call

**Pattern**: `git status --porcelain` returns empty → End turn. Any output → Commit first.
<!-- @/propagate -->

### 2.5 Quality Strategy

**Quality Attributes**: See [Section 11.1 Maximum Quality Strategy](#111-maximum-quality-strategy---mandatory) for the complete quality attributes list (NO EXCEPTIONS).

#### Coverage Targets

- **Production Code**: ≥95% minimum coverage
- **Infrastructure/Utility**: ≥98% minimum coverage
- **main() Functions**: 0% (exempt if internalMain() ≥95%)
- **Generated Code**: 0% (excluded - OpenAPI, GORM models, protobuf)

**Package-Level Exceptions**: Packages MAY have targets below mandatory minimum IF a coverage ceiling analysis (see [Section 10.2.3](#1023-coverage-targets)) documents the structural ceiling with justification.

#### Mutation Testing

- **Category-Based Targets**: ≥98% ideal efficacy (all packages), ≥95% mandatory minimum
- **Tool**: gremlins v0.6.0+ (Linux CI/CD for Windows compatibility)
- **Execution**: `gremlins unleash --tags=!integration` per package
- **Timeouts**: 4-6 packages per parallel job, <20 minutes total

#### Linting Standards

- **Zero Exceptions**: ALL code must pass linting (production, tests, examples, utilities)
- **golangci-lint v2**: v2.7.2+ with wsl_v5, built-in formatters
- **Auto-Fixable**: Run `--fix` first (gofumpt, goimports, wsl, godot, importas)
- **Critical Rules**: wsl (no suppression), godot (periods required), mnd (magic constants)

#### Pre-Commit Hooks

- **Same as CI/CD**: golangci-lint, gofumpt, goimports, cicd-enforce-internal
- **Auto-Conversions**: `time.Now()` → `time.Now().UTC()` for SQLite compatibility
- **UTF-8 without BOM**: All text files mandatory enforcement
- **Hook Documentation**: Update `docs/pre-commit-hooks.md` with config changes

#### Mandatory Review Passes

<!-- @propagate to=".github/instructions/01-02.beast-mode.instructions.md, .github/instructions/06-01.evidence-based.instructions.md, .github/agents/beast-mode.agent.md, .github/agents/fix-workflows.agent.md, .github/agents/implementation-execution.agent.md, .github/agents/implementation-planning.agent.md, .claude/agents/beast-mode.md, .claude/agents/fix-workflows.md, .claude/agents/implementation-execution.md, .claude/agents/implementation-planning.md" as="mandatory-review-passes" -->
**MANDATORY: Minimum 3, maximum 5 review passes before marking any task complete.**

Copilot and AI agents have a tendency to partially fulfill requested work, accidentally omitting or skipping items per request. To counter this, every task completion MUST include at least 3 review passes, each checking ALL 8 quality attributes:

**Each pass checks ALL 8 attributes** (fresh perspective per pass):
1. ✅ **Correctness** — code/docs correct, no regressions
2. ✅ **Completeness** — all tasks/steps/items addressed, nothing skipped
3. ✅ **Thoroughness** — evidence-based validation, all edge cases covered
4. ✅ **Reliability** — build, lint, test, coverage, mutation all pass
5. ✅ **Efficiency** — optimized for maintainability, not implementation speed
6. ✅ **Accuracy** — root cause addressed, not just symptoms
7. ❌ **NO Time Pressure** — NEVER rushed, NEVER cutting corners
8. ❌ **NO Premature Completion** — objective evidence required before marking complete

**Continuation rule**: If pass 3 finds ANY issue, continue to pass 4. If pass 4 still finds issues, continue to pass 5. Diminishing returns = done.

**Scope**: ALL work types — code, docs, config, tests, infrastructure, deployments.
<!-- @/propagate -->

---

## 3. Product Suite Architecture

### 3.1 Product Overview

**cryptoutil** comprises five independent products, each providing specialized cryptographic capabilities:

#### 3.1.1 Private Key Infrastructure (PKI)

- **Service**: Certificate Authority (CA)
- **Capabilities**: X.509 certificate lifecycle management, EST, SCEP, OCSP, CRL
- **Use Cases**: TLS certificate issuance, client authentication, code signing
- **Architecture**: 3-tier CA hierarchy (Offline Root → Online Root → Issuing CA)

#### 3.1.2 JSON Object Signing and Encryption (JOSE)

- **Service**: JWK Authority (JA)
- **Capabilities**: JWK/JWS/JWE/JWT cryptographic operations, elastic key rotation
- **Use Cases**: API token generation, data encryption, digital signatures
- **Key Features**: Per-message key rotation, automatic key versioning

#### 3.1.3 Secrets Manager (SM)

- **Services**: Key Management Service (KMS), Instant Messenger (IM)
- **Capabilities**: Elastic key management, hierarchical key barriers, encryption-at-rest, end-to-end encrypted messaging
- **Use Cases**: Application secrets, database encryption keys, API key management, secure communications
- **Key Features**: Unseal-based bootstrapping, automatic key rotation, message-level JWKs

#### 3.1.4 Identity

- **Services**: Authorization Server (Authz), Identity Provider (IdP), Resource Server (RS), Relying Party (RP), Single Page Application (SPA)
- **Capabilities**: OAuth 2.1, OIDC 1.0, WebAuthn, Passkeys, multi-factor authentication
- **Use Cases**: User authentication, API authorization, SSO, passwordless login
- **Key Features**: 41 authentication methods (13 headless + 28 browser), multi-tenancy

#### 3.1.5 Skeleton

- **Service**: Template
- **Capabilities**: Best-practice stereotype product-service showcasing all service-framework patterns
- **Use Cases**: Reference implementation for new product-service creation, developer onboarding
- **Key Features**: Minimal domain logic, full service-framework integration, deployment and config examples

**Skeleton / lint-fitness / `/new-service` Relationship**:

| Component | Role | Scope |
|-----------|------|-------|
| `skeleton-template` | Reference implementation | Source code copied by `/new-service` skill |
| `lint-fitness` | Automated enforcement | Validates ALL services (including skeleton) conform to structure rules |
| `/new-service` skill | Generation guide | Step-by-step instructions to copy skeleton-template and customize for a new service |

- **skeleton-template** is the canonical 8-file starter service using the latest builder API (`Build()` with `DomainConfig`). It showcases domain model, repository, migrations, server, config, and test patterns.
- **lint-fitness** enforces structural invariants (file limits, import isolation, test patterns, PostgreSQL isolation) across ALL services independently of how they were created.
- **`/new-service`** skill guides developers through copying skeleton-template, renaming identifiers, assigning ports, and registering with CI/CD. The skeleton is the INPUT; lint-fitness validates the OUTPUT.

### 3.2 Service Catalog

| Product | Service | Product-Service Identifier | Address (Container) [Admin] | Address (Container) [Public] | Address (Host) [Public] | Port Value (Container) [Admin] | Port Value (Container) [Public] | Port Range (Host) [Service Deployment] | Port Range (Host) [Product Deployment] | Port Range (Host) [Suite Deployment] | Description |
|---------|---------|----------------------------|-----------------------------|-----------------------------|-------------------------|--------------------------------|---------------------------------|----------------------------------------|----------------------------------------|--------------------------------------|-------------|
| **Secrets Manager (SM)** | **Key Management Service (KMS)** | **sm-kms** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8000-8099 | 18000-18099 | 28000-28099 | Elastic key management, encryption-at-rest |
| **Secrets Manager (SM)** | **Instant Messenger (IM)** | **sm-im** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8100-8199 | 18100-18199 | 28100-28199 | E2E encrypted messaging, encryption-at-rest |
| **JSON Object Signing and Encryption (JOSE)** | **JWK Authority (JA)** | **jose-ja** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8200-8299 | 18200-18299 | 28200-28299 | JWK/JWS/JWE/JWT operations |
| **Private Key Infrastructure (PKI)** | **Certificate Authority (CA)** | **pki-ca** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8300-8399 | 18300-18399 | 28300-28399 | X.509 certificates, EST, SCEP, OCSP, CRL |
| **Identity** | **Authorization Server (Authz)** | **identity-authz** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8400-8499 | 18400-18499 | 28400-28499 | OAuth 2.1 authorization server |
| **Identity** | **Identity Provider (IdP)** | **identity-idp** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8500-8599 | 18500-18599 | 28500-28599 | OIDC 1.0 Identity Provider |
| **Identity** | **Resource Server (RS)** | **identity-rs** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8600-8699 | 18600-18699 | 28600-28699 | OAuth 2.1 Resource Server |
| **Identity** | **Relying Party (RP)** | **identity-rp** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8700-8799 | 18700-18799 | 28700-28799 | OAuth 2.1 Relying Party |
| **Identity** | **Single Page Application (SPA)** | **identity-spa** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8800-8899 | 18800-18899 | 28800-28899 | OAuth 2.1 Single Page Application |
| **Skeleton** | **Template** | **skeleton-template** | 127.0.0.1 | 0.0.0.0 | 127.0.0.1 | 9090 | 8080 | 8900-8999 | 18900-18999 | 28900-28999 | Best-practice stereotype product-service template |

**Implementation Status**:

| Product-Service Identifier | Status | Completion | Notes |
|----------------------------|--------|------------|-------|
| **sm-kms** | ✅ Complete | 100% | Reference implementation with dual servers, Docker Compose |
| **sm-im** | ✅ Complete | 100% | E2E encrypted messaging, Docker Compose working |
| **jose-ja** | ✅ Complete | ~95% | Dual HTTPS servers, Docker Compose, E2E tests |
| **pki-ca** | ⚠️ Partial | ~40% | Domain under active development; E2E tests pending |
| **identity-authz** | ⚠️ Partial | ~50% | Domain under active development; partial E2E tests |
| **identity-idp** | ⚠️ Partial | ~50% | Domain under active development; partial E2E tests |
| **identity-rs** | ⚠️ Partial | ~20% | Domain implementation started; E2E tests pending |
| **identity-rp** | ⚠️ Partial | ~15% | Domain implementation started; E2E tests pending |
| **identity-spa** | ⚠️ Partial | ~15% | Domain implementation started; E2E tests pending |
| **skeleton-template** | ✅ Complete | ~95% | Best-practice stereotype template, dual HTTPS, Docker Compose, E2E tests |

**Legend**: ✅ Complete (production-ready), ⚠️ Partial (functional but missing features), ❌ Not Started

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

##### 3.2.1.2 Instant Messenger (IM) Service

- Product-Service (Unique Identifier): sm-im
- Service Name: Instant Messenger (IM)
- Service Description: E2E encrypted messaging, encryption-at-rest
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8100-8199
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18100-18199
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28100-28199

#### 3.2.2 JSON Object Signing and Encryption (JOSE) Product (1 Service)

##### 3.2.2.1 JWK Authority (JA) Service

- Product-Service (Unique Identifier): jose-ja
- Service Name: JWK Authority (JA)
- Service Description: JWK/JWS/JWE/JWT operations
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8200-8299
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18200-18299
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28200-28299

#### 3.2.3 Public Key Infrastructure (PKI) Product

##### 3.2.3.1 Certificate Authority (CA) Service

- Product-Service (Unique Identifier): pki-ca
- Service Name: Certificate Authority (CA)
- Service Description: X.509 certificates, EST, SCEP, OCSP, CRL
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8300-8399
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18300-18399
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28300-28399

#### 3.2.4 Identity Product

##### 3.2.4.1 OAuth 2.1 Authorization Server (Authz) Service

**Architecture note**: `identity-authz` is an **authorization-only** service. It implements OAuth 2.1 token issuance, scope enforcement, and client management. It does **NOT** authenticate users directly — it delegates user authentication to `identity-idp` (the OIDC Identity Provider). In error messages, logs, and variable names ALWAYS use `authz` (never the ambiguous `auth`).

- Product-Service (Unique Identifier): identity-authz
- Service Name: OAuth 2.1 Authorization Server (Authz)
- Service Description: OAuth 2.1 authorization server (authz-only; delegates user authn to identity-idp)
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8400-8499
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18400-18499
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28400-28499

##### 3.2.4.2 OIDC 1.0 Identity Provider (IdP) Service

**Architecture note**: `identity-idp` is the **authentication and identification** service. It implements all 41 authentication methods (13 headless + 28 browser) and issues OIDC ID tokens. `identity-authz` calls `identity-idp` to authenticate users as part of the OAuth 2.1 authorization code flow. At the Identity **product** level, saying "authentication and authorization" is correct since the product provides both capabilities via separate services.

- Product-Service (Unique Identifier): identity-idp
- Service Name: OIDC 1.0 Identity Provider (IdP)
- Service Description: OIDC 1.0 Identity Provider (authentication + identification; all 41 authn methods)
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8500-8599
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18500-18599
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28500-28599

##### 3.2.4.3 OAuth 2.1 Resource Server (RS) Service

- Product-Service (Unique Identifier): identity-rs
- Service Name: OAuth 2.1 Resource Server (RS)
- Service Description: OAuth 2.1 Resource Server
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8600-8699
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18600-18699
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28600-28699

##### 3.2.4.4 OAuth 2.1 Relying Party (RP) Service

- Product-Service (Unique Identifier): identity-rp
- Service Name: OAuth 2.1 Relying Party (RP)
- Service Description: OAuth 2.1 Relying Party
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8700-8799
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18700-18799
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28700-28799

##### 3.2.4.5 OAuth 2.1 Single Page Application (SPA) Service

- Product-Service (Unique Identifier): identity-spa
- Service Name: OAuth 2.1 Single Page Application (SPA)
- Service Description: OAuth 2.1 Single Page Application
- Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (container loopback only, IPv4 only)
- Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
- Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
- Port Value (Container): Private Admin Compose+K8s APIs: 9090
- Port Value (Container): Public Browser+Service APIs: 8080
- Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8800-8899
- Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18800-18899
- Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28800-28899

#### 3.2.5 Skeleton Product

##### 3.2.5.1 Template Service

- **Product-Service Identifier**: skeleton-template
- **Purpose**: Best-practice stereotype product-service template for service-framework usage reference
- **Capabilities**: Showcases all service-framework patterns with minimal domain logic
- **Use Cases**: Reference implementation for new product-service creation, developer onboarding, service-framework validation
- **Status**: ✅ Complete (~95%)
- **Network Configuration**:
  - Address (Container): Private Admin Compose+K8s APIs: 127.0.0.1 (IPv4 only)
  - Address (Container): Public Browser+Service APIs: 0.0.0.0 (all interfaces, IPv4 only)
  - Address (Host): Public Browser+Service APIs: 127.0.0.1 (IPv4 only), localhost
  - Port Value (Container): Private Admin Compose+K8s APIs: 9090
  - Port Value (Container): Public Browser+Service APIs: 8080
  - Port Range (Host): Public Browser+Service APIs (Isolated Service Deployment): 8900-8999
  - Port Range (Host): Public Browser+Service APIs (Isolated Product Deployment): 18900-18999
  - Port Range (Host): Public Browser+Service APIs (Suite Deployment): 28900-28999

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
   - Port Range: Service-specific base (e.g., 8300-8399 for pki-ca)
   - Use Case: Independent service development, testing, or production deployment
   - Example: `pki-ca` alone uses host ports 8300-8399

2. **Product Deployment** (18XXX): All services within a product
   - Port Range: Service-specific base + 10000 offset (e.g., 18300-18399 for pki-ca)
   - Use Case: Product-level integration testing, product-only deployments
   - Example: All PKI services (currently only pki-ca) use host ports 18300-18399

3. **Suite Deployment** (28XXX): All services across all products
   - Port Range: Service-specific base + 20000 offset (e.g., 28300-28399 for pki-ca)
   - Use Case: Full system integration, E2E testing, complete production suite
   - Example: All 10 services across 5 products use host ports 28000-28999

**Port Allocation Benefits**:

- No port conflicts between deployment types (all three can run simultaneously)
- Consistent port offsets simplify troubleshooting (service port + offset = deployment type)
- Clear separation enables independent CI/CD pipelines per deployment type

**App Service Variant Port Formula**:

Each PS-ID compose.yml defines 4 app service instances. The host port for each instance is:

```
host_port = base_port + tier_offset + variant_offset
```

Where:
- `base_port` is the PS-ID's assigned base host port (see §3.2 Service Catalog)
- `tier_offset` is 0 (Service), 10000 (Product), or 20000 (Suite)
- `variant_offset` is the instance variant offset:

| Variant | Compose Service Suffix | Variant Offset | Database Backend |
|---------|------------------------|----------------|-----------------|
| sqlite-1 | `-app-sqlite-1` | +0 | SQLite in-memory (dev/CI primary) |
| sqlite-2 | `-app-sqlite-2` | +1 | SQLite in-memory (dev/CI secondary) |
| postgresql-1 | `-app-postgresql-1` | +2 | PostgreSQL (primary) |
| postgresql-2 | `-app-postgresql-2` | +3 | PostgreSQL (secondary) |

**Example** (sm-kms, SERVICE tier, base_port=8000, tier_offset=0):

| Compose Service | host_port formula | Host Port |
|----------------|------------------|-----------|
| `sm-kms-app-sqlite-1` | 8000 + 0 + 0 | 8000 |
| `sm-kms-app-sqlite-2` | 8000 + 0 + 1 | 8001 |
| `sm-kms-app-postgresql-1` | 8000 + 0 + 2 | 8002 |
| `sm-kms-app-postgresql-2` | 8000 + 0 + 3 | 8003 |

The `compose-port-formula` fitness linter (`go run ./cmd/cicd-lint lint-fitness`) validates all compose port bindings against this formula at CI time.

#### 3.4.2 PostgreSQL Ports

| Product-Service Identifier | Address (Host) | Host Port | Container Address | Port Value (Container) |
|---------|-----------|----------------|----------|----------------|
| **sm-kms** | 127.0.0.1 | 54320 | 0.0.0.0 | 5432 |
| **sm-im** | 127.0.0.1 | 54321 | 0.0.0.0 | 5432 |
| **jose-ja** | 127.0.0.1 | 54322 | 0.0.0.0 | 5432 |
| **pki-ca** | 127.0.0.1 | 54323 | 0.0.0.0 | 5432 |
| **identity-authz** | 127.0.0.1 | 54324 | 0.0.0.0 | 5432 |
| **identity-idp** | 127.0.0.1 | 54325 | 0.0.0.0 | 5432 |
| **identity-rs** | 127.0.0.1 | 54326 | 0.0.0.0 | 5432 |
| **identity-rp** | 127.0.0.1 | 54327 | 0.0.0.0 | 5432 |
| **identity-spa** | 127.0.0.1 | 54328 | 0.0.0.0 | 5432 |
| **skeleton-template** | 127.0.0.1 | 54329 | 0.0.0.0 | 5432 |

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

- main() `cmd/` → Application `internal/*/application/` → Business Logic `internal/*/service/`, `internal/*/model/` → Repositories `internal/*/repository/` → Database/External Systems
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
- internal/apps/: Applications (suite at {SUITE}/, products at {PRODUCT}/, services at flat {PS-ID}/)
- internal/shared/: Shared utilities (apperr, config, crypto, magic, pool, telemetry, testutil, util)
- api/: OpenAPI specs, generated code
- configs/: Configuration files
- deployments/: Docker Compose, Kubernetes manifests
- docs/: Documentation
- pkg/: Public library code (intentionally empty - all code is internal)
- scripts/: Placeholder only (`.gitkeep`). All tooling lives in `cmd/` or `internal/apps/tools/`.
- test/: Additional test files (Gatling load tests — Java/Maven only)

#### 4.4.2 Directory Rules

- ❌ Avoid /src directory (redundant in Go)
- ❌ Avoid deep nesting (>8 levels indicates design issue)
- ✅ Use /internal for private code (enforced by compiler)
- ✅ Use /pkg for public libraries (safe for external import - currently empty by design)
- ✅ Use `model/` (not `domain/`) for packages containing GORM-tagged structs — these are persistence models, not pure domain types

#### 4.4.3 CLI Entry Points

**18 flat entries**: 1 suite + 5 products + 10 services + 2 infra tools.

```
cmd/
│   # Suite (×1, {SUITE}=cryptoutil)
├── cryptoutil/main.go                  # Suite CLI → internal/apps/cryptoutil/cryptoutil.go
│
│   # Products (×5, {PRODUCT}=identity|jose|pki|skeleton|sm)
├── identity/main.go                    # Product CLI → internal/apps/identity/identity.go
├── jose/main.go                        # Product CLI → internal/apps/jose/jose.go
├── pki/main.go                         # Product CLI → internal/apps/pki/pki.go
├── skeleton/main.go                    # Product CLI → internal/apps/skeleton/skeleton.go
├── sm/main.go                          # Product CLI → internal/apps/sm/sm.go
│
│   # Services (×10, {PS-ID}={PRODUCT}-{SERVICE})
├── identity-authz/main.go             # Service CLI → internal/apps/identity-authz/identity-authz.go
├── identity-idp/main.go               # Service CLI → internal/apps/identity-idp/identity-idp.go
├── identity-rp/main.go                # Service CLI → internal/apps/identity-rp/identity-rp.go
├── identity-rs/main.go                # Service CLI → internal/apps/identity-rs/identity-rs.go
├── identity-spa/main.go               # Service CLI → internal/apps/identity-spa/identity-spa.go
├── jose-ja/main.go                     # Service CLI → internal/apps/jose-ja/jose-ja.go
├── pki-ca/main.go                      # Service CLI → internal/apps/pki-ca/pki-ca.go
├── skeleton-template/main.go           # Service CLI → internal/apps/skeleton-template/skeleton-template.go
├── sm-im/main.go                       # Service CLI → internal/apps/sm-im/sm-im.go
├── sm-kms/main.go                      # Service CLI → internal/apps/sm-kms/sm-kms.go
│
│   # Infra tools (×2, {INFRA-TOOL}=cicd-lint|cicd-workflow)
├── cicd-lint/main.go                   # CICD lint CLI → internal/apps/tools/cicd_lint/cicd.go
└── cicd-workflow/main.go               # Workflow CLI → internal/apps/tools/cicd_workflow/workflow.go
```

**Pattern**: Thin `main()` pattern for all cmd/ CLIs, with all logic in `internal/apps/` for maximum code reuse and testability.

1. `cmd/{SUITE}/` for suite-level CLI (e.g., `cmd/cryptoutil/`)
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
1. `cmd/<ps-id>/` for service-level CLI
```go
func main() {
    os.Exit(cryptoutilApps<PS-ID>.<PS-ID>(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
```

#### 4.4.4 Service Implementations

Services live at flat `internal/apps/{PS-ID}/`. Product directories contain only
product-level code (`{PRODUCT}.go`, shared packages) — NO service subdirectories.

```
internal/apps/
│
│   # Suite orchestration
├── cryptoutil/
│   └── cryptoutil.go                     # Suite CLI dispatch
│
│   # Product level (product.go + shared packages only; NO service subdirs)
├── identity/
│   ├── identity.go                       # Product CLI dispatch
│   └── (shared: domain/, repository/, config/, apperr/, email/, issuer/, jobs/, mfa/, ratelimit/, rotation/)
├── jose/
│   └── jose.go
├── pki/
│   └── pki.go
├── skeleton/
│   └── skeleton.go
├── sm/
│   └── sm.go
│
│   # Service level (flat PS-ID directories, ×10)
├── identity-authz/
│   └── identity-authz.go                 # Service entry point (seam pattern)
├── identity-idp/
│   └── identity-idp.go
├── identity-rp/
│   └── identity-rp.go
├── identity-rs/
│   └── identity-rs.go
├── identity-spa/
│   └── identity-spa.go
├── jose-ja/
│   └── jose-ja.go
├── pki-ca/
│   └── pki-ca.go
├── skeleton-template/
│   └── skeleton-template.go
├── sm-im/
│   └── sm-im.go
├── sm-kms/
│   └── sm-kms.go
│
│   # Framework & tools
├── framework/                            # Service framework (shared by all services)
└── tools/                                # Infrastructure tooling (cicd_lint, cicd_workflow)
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

#### 4.4.6 Deployments

**Structure**: Parameterized by `{PS-ID}` (service), `{PRODUCT}` (product), and `{SUITE}` (suite). All 10 services follow an identical pattern.

##### Per-Service Deployment (`deployments/{PS-ID}/`) — ×10

```
deployments/{PS-ID}/
├── compose.yml                    # Docker Compose service definition
├── Dockerfile                     # Service Docker image build
├── config/                        # 5 config overlay files
│   ├── {PS-ID}-app-common.yml
│   ├── {PS-ID}-app-sqlite-1.yml
│   ├── {PS-ID}-app-sqlite-2.yml
│   ├── {PS-ID}-app-postgresql-1.yml
│   └── {PS-ID}-app-postgresql-2.yml
└── secrets/                       # 14 secret files (see table below)
```

**All 10 PS-IDs**: `identity-authz`, `identity-idp`, `identity-rp`, `identity-rs`, `identity-spa`, `jose-ja`, `pki-ca`, `skeleton-template`, `sm-im`, `sm-kms`.

##### Tier Differences (Product ×5 / Suite ×1)

| Component | Service (`{PS-ID}/`) | Product (`{PRODUCT}/`) | Suite (`cryptoutil/`) |
|-----------|---------------------|----------------------|---------------------------|
| `compose.yml` | Direct service definition | Includes service composes | Includes product composes |
| `Dockerfile` | ✅ | ✅ | ❌ (no separate image) |
| `config/` | 5 overlay files | ❌ (uses service configs) | ❌ (uses service configs) |
| `secrets/` | 14 `.secret` files | `.secret` + `.secret.never` | `.secret` + `.secret.never` |
| Browser/service creds | Real `.secret` files | `.secret.never` markers | `.secret.never` markers |
| Value prefix | `{PS-ID}-` / `{PS_ID}_` | `{PRODUCT}-` / `{PRODUCT}_` | `{SUITE}-` / `{SUITE}_` |

**All 5 products**: `identity`, `jose`, `pki`, `skeleton`, `sm`.

##### Shared Infrastructure Deployments

`deployments/shared-telemetry/compose.yml` (otel-collector-contrib + grafana-otel-lgtm) and `deployments/shared-postgres/compose.yml` (shared PostgreSQL container).

##### Secret File Naming Convention

All tiers (service, product, suite) use **identical `{purpose}.secret` filenames** — no tier prefix on filenames. The **value inside** each secret contains the tier-specific prefix. `.secret.never` marker files exist ONLY at product and suite tiers.

| Secret Purpose | Filename | Service Value | Product Value | Suite Value |
|---------------|----------|---------------|---------------|-------------|
| Hash pepper v3 | `hash-pepper-v3.secret` | `{PS-ID}-hash-pepper-v3-{base64-random-32-bytes}` | `{PRODUCT}-hash-pepper-v3-{base64-random-32-bytes}` | `{SUITE}-hash-pepper-v3-{base64-random-32-bytes}` |
| Browser username | `browser-username.secret` | `{PS-ID}-browser-user` | `.never` marker | `.never` marker |
| Browser password | `browser-password.secret` | `{PS-ID}-browser-pass-{base64-random-32-bytes}` | `.never` marker | `.never` marker |
| Service username | `service-username.secret` | `{PS-ID}-service-user` | `.never` marker | `.never` marker |
| Service password | `service-password.secret` | `{PS-ID}-service-pass-{base64-random-32-bytes}` | `.never` marker | `.never` marker |
| PostgreSQL username | `postgres-username.secret` | `{PS_ID}_database_user` | `{PRODUCT}_database_user` | `{SUITE}_database_user` |
| PostgreSQL password | `postgres-password.secret` | `{PS_ID}_database_pass-{base64-random-32-bytes}` | `{PRODUCT}_database_pass-{base64-random-32-bytes}` | `{SUITE}_database_pass-{base64-random-32-bytes}` |
| PostgreSQL database | `postgres-database.secret` | `{PS_ID}_database` | `{PRODUCT}_database` | `{SUITE}_database` |
| PostgreSQL URL | `postgres-url.secret` | `postgres://{PS_ID}_database_user:{password}@{PS-ID}-postgres:5432/{PS_ID}_database?sslmode=disable` | `...@{PRODUCT}-postgres:5432/...` | `...@{SUITE}-postgres:5432/...` |
| Unseal shard N | `unseal-{N}of5.secret` | `{PS-ID}-unseal-key-N-of-5-{base64-random-32-bytes}` | `{PRODUCT}-unseal-key-N-of-5-{base64-random-32-bytes}` | `{SUITE}-unseal-key-N-of-5-{base64-random-32-bytes}` |

#### 4.4.7 CLI Patterns

### CLI Hierarchy

Four equivalent invocation forms route to the same service handler:

```
# Product-Service pattern (preferred) — binary name IS the {PS-ID}
{PS-ID} <subcommand> --config=/etc/{PRODUCT}/{SERVICE}.yml

# Service pattern — short alias when product is unambiguous
{SERVICE} <subcommand> --config=/etc/{PRODUCT}/{SERVICE}.yml

# Product pattern — product binary routes to service
{PRODUCT} {SERVICE} <subcommand> --config=/etc/{PRODUCT}/{SERVICE}.yml

# Suite pattern — suite binary routes to product, then service
{SUITE} {PRODUCT} {SERVICE} <subcommand> --config=/etc/{PRODUCT}/{SERVICE}.yml
```

**Concrete example** (sm-kms, {SUITE}=cryptoutil, {PRODUCT}=sm, {SERVICE}=kms, {PS-ID}=sm-kms):

```
sm-kms  server --config=/etc/sm/kms.yml   # {PS-ID} pattern
kms     server --config=/etc/sm/kms.yml   # {SERVICE} pattern
sm kms  server --config=/etc/sm/kms.yml   # {PRODUCT} {SERVICE} pattern
cryptoutil sm kms server --config=/etc/sm/kms.yml   # {SUITE} {PRODUCT} {SERVICE} pattern
```

### CLI Subcommand

All CLIs for all 10 services MUST support these subcommands, with consistent behavior and config parsing and flag parsing.
Consistency MUST be guaranteed by inheriting from service-framework, which will reuse `internal/apps/framework/service/<SUBCOMMAND>/` packages:

| Subcommand | Description |
|------------|-------------|
| `server` | CLI server start with dual HTTPS listeners, for Private Admin Compose+K8s APIs vs Public Business Logic APIs |
| `health` | CLI client for Public health endpoint API check |
| `livez` | CLI client for Private liveness endpoint API check |
| `readyz` | CLI client for Private readiness endpoint API check |
| `shutdown` | CLI client for Private graceful shutdown endpoint API trigger |
| `client` | CLI client for Business Logic API interaction (n.b. domain-specific for each of the 10 services) |
| `init` | CLI client for Initialize static config, like TLS certificates |

#### Framework Tier Routing

The suite → product → service CLI hierarchy is implemented by three framework routing packages:

| Package | Function | Call Pattern |
|---------|----------|-------------|
| `internal/apps/framework/suite/cli/` | `RouteSuite(cfg, args, stdin, stdout, stderr, products)` | `{SUITE} {PRODUCT} {SERVICE} <subcommand>` |
| `internal/apps/framework/product/cli/` | `RouteProduct(cfg, args, stdin, stdout, stderr, services)` | `{PRODUCT} {SERVICE} <subcommand>` |
| `internal/apps/framework/service/` | Service-level subcommand dispatch | `{PS-ID} <subcommand>` |

**RouteSuite** accepts a `[]ProductEntry` (name + handler). Matches `args[0]` to a product name, delegates remaining args to the product handler.

**RouteProduct** accepts a `[]ServiceEntry` (name + handler). Supports `--version`/`--help` flags. Matches `args[0]` to a service name, delegates remaining args to the service handler.

**Convention**: Each `cmd/{PRODUCT}/main.go` calls `RouteProduct()`. Each `cmd/{SUITE}/main.go` calls `RouteSuite()`, which delegates to the same product handlers.

#### Anti-Patterns

**NEVER** create `cmd/{PRODUCT}-{SUBCOMMAND}/` executables for subcommands:

- `cmd/sm-im-server/main.go` → **WRONG**: Use `cmd/sm-im server` subcommand instead (i.e., `cmd/{PS-ID} server`).
- `cmd/cryptoutil-health/main.go` → **WRONG**: Use `cmd/cryptoutil health` subcommand instead (i.e., `cmd/{SUITE} health`).

Each product-service binary (`cmd/{PS-ID}/main.go`) routes subcommands internally via the framework. Multiple executables for the same binary's subcommands create maintenance burden and violate the single-binary principle.

#### Infrastructure CLI Tools (Intentional Exceptions)

Two `cmd/` entries exist as deliberate **infrastructure tools**, NOT product/service CLIs:

| Entry | Internal Package | Purpose |
|-------|-----------------|---------|
| `cmd/cicd-lint/` | `internal/apps/tools/cicd_lint/` | CI/CD quality tooling: 11 linters, 2 formatters, 1 script |
| `cmd/cicd-workflow/` | `internal/apps/tools/cicd_workflow/` | GitHub Actions workflow testing infrastructure |

These are **intentional exceptions** to the product/service CLI pattern (`{INFRA-TOOL}=cicd-lint|cicd-workflow`). They serve **repository infrastructure**, not business domain concerns:

- MUST NOT be merged into product/service CLIs.
- MUST NOT be subcommands of `cmd/{SUITE}/` (suite CLI).
- MUST be documented here to prevent confusion about "non-standard" entries.

See [Section 9.10 CICD Command Architecture](#910-cicd-command-architecture) for the `cmd/cicd-lint/` four-layer dispatch pattern, command catalog, and enforcement rules.

---

## 5. Service Architecture

### 5.1 Service Framework Pattern

#### 5.1.1 Framework Components

<!-- @propagate to=".github/instructions/02-01.architecture.instructions.md" as="service-framework-components" -->
- Two HTTPS Listeners: Public (business APIs) + Admin (health checks)
- Two Public Paths: `/browser/**` (session cookies) vs `/service/**` (session tokens)
- Three Admin APIs: /admin/api/v1/livez, /admin/api/v1/readyz, /admin/api/v1/shutdown
- Database: PostgreSQL || SQLite with GORM
- Telemetry: OTLP → otel-collector-contrib → Grafana LGTM
- Config Priority: Docker secrets > YAML > CLI parameters (NO environment variables)
<!-- @/propagate -->

#### 5.1.2 Framework Benefits

- Eliminates 48,000+ lines of boilerplate per service
- Consistent infrastructure across all 10 services
- Proven patterns: TLS setup, middleware stacks, health checks, graceful shutdown
- Parameterization: OpenAPI specs, handlers, middleware chains injected via constructor

#### 5.1.3 Mandatory Usage

- ALL new services MUST use `internal/apps/framework/service/` (consistency, reduced duplication)
- ALL existing services MUST be refactored to use `internal/apps/framework/service/` (iterative migration)
- Migration priority: sm-im → jose-ja → sm-kms → pki-ca → identity services
  - sm-im/jose-ja/sm-kms migrate first (SM product); pki-ca second; identity last

### 5.2 Service Builder Pattern

#### 5.2.1 Builder Methods

- NewServerBuilder(ctx, cfg): Create builder with `internal/apps/framework/service/` config
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

##### SQLite + Barrier Outside Transactions (CRITICAL)

<!-- @propagate to=".github/instructions/03-04.data-infrastructure.instructions.md" as="sqlite-barrier-outside-tx" -->
**MANDATORY**: ALL calls to `barrier.EncryptContentWithContext` or `barrier.DecryptContentWithContext` MUST be outside any ORM `WithTransaction` scope.

**Root cause**: The barrier service opens its own internal read/write transaction. SQLite WAL mode allows only one writer at a time. Nesting two write transactions on the same connection pool causes deadlock: all connections are held by the outer ORM transaction, so the inner barrier transaction cannot acquire one.

**Correct pattern** — barrier after ORM commit:
```
ORM.Create(plainRecord) → commit → (outside tx) barrier.Encrypt → ORM.Update(encryptedRecord)
```

This is a **correctness requirement**, not a performance concern. Barrier calls inside ORM transactions are a guaranteed SQLite deadlock.
<!-- @/propagate -->

##### SQLite DateTime (CRITICAL)

**ALWAYS use `.UTC()` when comparing with SQLite timestamps**:

```go
// ❌ WRONG: time.Now() without .UTC()
if session.CreatedAt.After(time.Now()) { ... }

// ✅ CORRECT: Always use .UTC()
if session.CreatedAt.After(time.Now().UTC()) { ... }
```

**Pre-commit hook auto-converts** `time.Now()` → `time.Now().UTC()`.

#### 5.2.5 Framework Shared Types

The `internal/apps/framework/service/` package provides reusable infrastructure types that all services share. These types eliminate boilerplate and ensure architectural consistency.

##### Config Types (`internal/apps/framework/service/config/`)

| Type | Purpose | Key Fields |
|------|---------|-----------|
| `ServerConfig` | HTTP server settings | Name, BindAddress, Port, TLS fields, Admin fields |
| `DatabaseConfig` | Database connection settings | Type (postgres/sqlite), DSN, MaxOpenConns, MaxIdleConns |
| `SessionConfig` | Session management settings | SessionLifetime, IdleTimeout, Cookie fields |
| `ObservabilityConfig` | Logging and telemetry settings | LogLevel, LogFormat, MetricsEnabled, TracingEnabled |

Each type has a `Validate()` method that enforces field constraints.

**Usage pattern** (type alias for backward compatibility in service-specific config packages):

```go
import cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"

// Type alias: backward compatible — all importers unchanged, Validate() inherited
type ServerConfig = cryptoutilAppsFrameworkServiceConfig.ServerConfig
type DatabaseConfig = cryptoutilAppsFrameworkServiceConfig.DatabaseConfig
type SessionConfig = cryptoutilAppsFrameworkServiceConfig.SessionConfig
type ObservabilityConfig = cryptoutilAppsFrameworkServiceConfig.ObservabilityConfig
```

Service-specific types (e.g., `identity.TokenConfig`, `identity.SecurityConfig`) remain in their own packages.

##### Rate Limiter (`internal/apps/framework/service/ratelimit/`)

The `RateLimiter` type provides per-key token-bucket rate limiting with configurable windows. Used by services to throttle operations like email OTP sending.

```go
// Create with max requests per window
rateLimiter := cryptoutilFrameworkServiceRatelimit.NewRateLimiter(maxCount, windowSize)

// Allow returns nil or an error if the rate limit is exceeded
if err := rateLimiter.Allow(userID.String()); err != nil {
    return fmt.Errorf("%w: %w", apperr.ErrRateLimitExceeded, err)
}
```

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

**mTLS Deployment Strategy**:

| Stage | TLS Mode | Authentication | Status |
|-------|----------|---------------|--------|
| Current | Unilateral TLS | Bearer token / API key over TLS 1.3+ | ✅ Complete |
| Future | Mutual TLS (mTLS) | Client certificate + server certificate | ⏳ Deferred until PKI service completes |

- **Current**: Server authenticates to client. Service-to-service calls use Bearer token or API key over TLS 1.3+.
- **Future (mTLS)**: Both parties authenticate via client certificates (mTLS). All services require PKI-issued client certificates.
- **Production Goal**: All internal service-to-service calls use mTLS (client certificate auth over TLS 1.3+). Revocation checked via CRLDP + OCSP (both must be checked; fail if both unreachable). See [Section 6.5](#65-pki-architecture--strategy) for PKI architecture.

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

<!-- @propagate to=".github/instructions/02-06.authn.instructions.md" as="key-principles" -->
- **Zero Trust**: NO caching of authorization decisions. Re-evaluate every request.
- **MFA Step-Up**: Re-authentication MANDATORY every 30 minutes for high-sensitivity operations.
- **Session Storage**: SQL databases ONLY (PostgreSQL or SQLite with ACID). NEVER Redis/Memcached.
- **mTLS Revocation**: MUST check BOTH CRLDP and OCSP. Fail if BOTH unreachable.
<!-- @/propagate -->

#### 6.9.1 Authentication Realm Architecture

<!-- @propagate to=".github/instructions/02-06.authn.instructions.md" as="session-token-formats" -->
**Opaque** (UUID), **JWE** (encrypted JWT), **JWS** (signed JWT). Storage: PostgreSQL (distributed) or SQLite (single-node). NO Redis/Memcached.
<!-- @/propagate -->

- Realm types and purposes
- Credential validators (File, Database, Federated)
- Session creation vs session upgrade flows
- Multi-tenancy isolation via realms

#### 6.9.2 Headless Authentication Methods (13 Total)

<!-- @propagate to=".github/instructions/02-06.authn.instructions.md" as="headless-authn" -->
**Non-Federated (6)**: JWE Session Token, JWS Session Token, Opaque Session Token, Basic (Client ID/Secret), Bearer (API Token), HTTPS Client Certificate.

**Federated (7)**: Basic/Bearer/ClientCert via OAuth 2.1, JWE/JWS/Opaque Access Token, Opaque Refresh Token.

**Storage**: YAML + SQL (Config > DB priority) for all methods.
<!-- @/propagate -->

#### 6.9.3 Browser Authentication Methods (28 Total)

<!-- @propagate to=".github/instructions/02-06.authn.instructions.md" as="browser-authn" -->
**Non-Federated (6)**: JWE/JWS/Opaque Session Cookie, Basic (Username/Password), Bearer (API Token), HTTPS Client Certificate.

**Federated (22)**: All non-federated methods PLUS:
- **MFA Factors**: TOTP, HOTP, Recovery Codes, WebAuthn (with/without Passkeys), Push Notification
- **Passwordless**: Email/Password, Magic Link (Email/SMS), Random OTP (Email/SMS/Phone)
- **Social Login**: Google, Microsoft, GitHub, Facebook, Apple, LinkedIn, Twitter/X, Amazon, Okta
- **Enterprise**: SAML 2.0

**Storage**: YAML + SQL (Config > DB) for static credentials. SQL ONLY for dynamic user data (OTPs, enrollments, magic links).
<!-- @/propagate -->

#### 6.9.4 Multi-Factor Authentication (MFA)

<!-- @propagate to=".github/instructions/02-06.authn.instructions.md" as="mfa-combinations" -->
**Browser**: Password + TOTP/WebAuthn/Push/OTP.
**Headless**: Client ID/Secret + mTLS/Bearer.
<!-- @/propagate -->

- Step-up authentication (re-auth every 30min for high-sensitivity operations)
- Factor enrollment workflows
- MFA bypass policies and emergency access

#### 6.9.5 Authorization Patterns

<!-- @propagate to=".github/instructions/02-06.authn.instructions.md" as="authz-methods" -->
**Headless**: Scope-based, RBAC.
**Browser**: Scope-based, RBAC, resource-level ACLs, consent tracking (scope+resource granularity).
<!-- @/propagate -->

- Zero trust: NO caching of authorization decisions
- Scope-based authorization (headless)
- Resource-based ACLs (browser)
- Consent tracking at scope + resource granularity

### 6.10 Secrets Detection Strategy

<!-- @propagate to=".github/instructions/02-05.security.instructions.md" as="secrets-detection-strategy" -->
**Detection**: Length-based threshold (≥32 bytes / ≥43 base64 chars) for inline secrets in compose files. NO entropy calculation (too many false positives). Safe references (`/run/secrets/`, short dev defaults) excluded. Infrastructure deployments excluded.
<!-- @/propagate -->

**Detection Approach**: Length-based threshold (≥32 bytes raw, ≥43 characters base64-encoded) identifies high-entropy inline values in environment variables matching secret-pattern names (PASSWORD, SECRET, TOKEN, KEY, API_KEY). No entropy calculation is used - it produces too many false positives on non-secret configuration values.

**Safe References** (excluded from detection): Docker secret paths (`/run/secrets/`), short development defaults (< threshold), empty values, variable references (`${VAR}`).

**Trade-offs**: Length threshold catches most real secrets (UUIDs, tokens, hashes) while allowing short developer passwords (`admin`, `dev123`). Infrastructure deployments (Grafana, OTLP collector) are excluded since they intentionally use inline dev credentials.

**Cross-References**: Implementation in [validate_secrets.go](/internal/apps/tools/cicd_lint/lint_deployments/validate_secrets.go). Deployment secrets management in [Section 12.6](#133-secrets-management-in-deployments).

---

### 6.11 TLS Certificate Configuration

**Service Template uses a 3-mode auto-detect strategy** based on credentials provided at startup:

| Environment | Cert Chain | TLS Key | Issuing CA Key | TLS Mode | Outcome |
|-------------|-----------|---------|----------------|----------|---------|
| Production | Provided | Docker Secret | Not provided | Static | Use as-is |
| E2E Dev | Provided | Not provided | Docker Secret | Mixed | Generate + sign TLS cert |
| Unit/Integration | Not provided | Not provided | Not provided | Auto | Auto-create all certs |

#### 6.11.1 TLS Mode Taxonomy (Static / Mixed / Auto)

The `GenerateTLSMaterial()` function in `internal/apps/framework/service/config/tls_generator.go` selects one of three modes based on available credentials:

**TLSModeStatic** (Production):
- Provides pre-generated certificate chain + private key via Docker secrets
- No key generation at runtime — fastest, most deterministic
- Requires: `tls_server_cert.secret` + `tls_server_key.secret`

**TLSModeMixed** (E2E Dev):
- Provides CA certificate + CA private key; server certificate generated at startup
- Server private key generated in-memory (not stored)
- Requires: `tls_ca_cert.secret` + `tls_issuing_ca_key.secret`

**TLSModeAuto** (Unit/Integration Tests):
- Fully auto-generates 3-tier CA hierarchy (root → intermediate → server)
- All keys generated in-memory; ephemeral per process start
- Requires: no TLS secrets (any absent → Auto mode)

**Detection Logic**: `StaticCertPEM + StaticKeyPEM` provided → **Static**. `MixedCACertPEM + MixedCAKeyPEM` provided → generate server cert then treat as **Static**. Nothing provided → **Auto**.

#### 6.11.2 Test TLS Bundle

**Unit/Integration tests MUST use Auto TLS** (no Docker secrets needed). The server auto-generates a complete ephemeral PKI chain per test run. Test HTTP clients must use `TLSRootCAPool()` / `AdminTLSRootCAPool()` from the started server to trust the ephemeral CA. See [Section 10.3.7](#1037-tls-test-bundle-pattern) for the TestMain pattern.

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
- **Unique Database Name**: Each of 10 services MUST have unique `postgres-database.secret`
  - Example: `identity_authz_database`, `identity_idp_database`, `jose_ja_database` (NOT shared `identity_database`)
- **Unique Username**: Each service MUST have unique `postgres-username.secret`
  - Example: `identity_authz_database_user`, `identity_idp_database_user`, `jose_ja_database_user`
- **Unique Password**: Each service MUST have unique `postgres-password.secret`
- **Unique Connection URL**: Each service MUST have unique `postgres-url.secret`

**Linter Enforcement** (`cicd-lint lint-deployments`):
- Scans ALL 10 service directories for database credential secrets
- ERRORS on duplicate database names or usernames across services
- Validates credentials are isolated regardless of deployment level (SUITE/PRODUCT/SERVICE)

**Exception**: Leader-follower PostgreSQL replication where logical schemas are replicated but services maintain separate schema namespaces within the same physical server.

**Cross-Service Communication**: Services needing data from other services MUST use REST APIs, NEVER direct database access.

### 7.2 Multi-Tenancy Architecture & Strategy

#### 7.2.1 Authentication Realms

**CRITICAL**: Realms define authentication METHOD and POLICY, NOT data scoping.

**Realms do NOT scope data** - all realms in same tenant see same data. Only `tenant_id` scopes data access.

#### Authentication Realm Types

| Category | Realm Types | Scheme | Credential Validators |
|----------|------------|--------|----------------------|
| **mTLS** | `https-client-cert-factor` | HTTP/mTLS Handshake | File, Database, Federated |
| **WebAuthn** (4 variants) | `webauthn-{resident,nonresident}-{synced,unsynced}-factor` | WebAuthn L2 | File, Database, Federated |
| **OAuth 2.1 AuthZ Code** (3 variants) | `authorization-code-{opaque,jwe,jws}-factor` | AuthZ Code Flow + PKCE | File, Database, Federated |
| **Bearer Token** (3 variants) | `bearer-token-{opaque,jwe,jws}-factor` | HTTP `Authorization: Bearer` | File, Database, Federated |
| **Basic Auth** (8 variants) | `basic-{username,email,sms}-password-factor`, `basic-{email,sms}-{otp,magiclink}-factor`, `basic-voice-otp-factor`, `basic-id-otp-factor` | HTTP `Authorization: Basic` | File, Database, Federated |
| **Session Cookies** (3 variants) | `cookie-token-{opaque,jwe,jws}-session` | HTTP `Cookie` | File, Database, Federated |

All factor realm types create or upgrade sessions. All session realm types use existing sessions. Total: 22 factor types + 1 session type (with 3 token format variants each = 3 session realm types). All realm types support File, Database, and Federated credential validators.

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

**Supported Engines**: PostgreSQL and SQLite ONLY. No other database engines (Citus, CockroachDB, Redis, etc.) are supported or planned. This is a deliberate constraint to reduce complexity and ensure consistent cross-database testing.

All 10 services MUST support using one of PostgreSQL or SQLite, specified via configuration at startup.

Typical usages for each database for different purposes (MANDATORY — see [Section 10.1](#101-testing-strategy-overview) for 3-Tier Database Strategy):
- Unit tests, Fuzz tests, Benchmark tests, Mutation tests => SQLite in-memory (NEVER PostgreSQL)
- Integration tests => SQLite in-memory via TestMain shared instance (NEVER PostgreSQL)
- Load tests => Ephemeral PostgreSQL instance (i.e. test-container)
- End-to-End tests => Static PostgreSQL instance (e.g. Docker Compose)
- Production => Static PostgreSQL instance (e.g. Cloud hosted)
- Local Development => Static SQLite instance (e.g. file)

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
- **MANDATORY**: Handler DTOs MUST come from generated `api/*/server/` and `api/model/` packages. NEVER hand-roll request/response structs that duplicate generated models.

#### 8.1.3 Validation Rules

- String: format (uuid), enum, minLength/maxLength, pattern
- Number: minimum/maximum, multipleOf
- Array: minItems/maxItems, uniqueItems, items
- Object: required, nested properties

#### 8.1.4 Canonical Base Initialisms List

<!-- @propagate to=".github/instructions/02-04.openapi.instructions.md" as="base-initialisms" -->
All `openapi-gen_config*.yaml` files MUST include the full base initialisms list in their `additional-initialisms` section. Domain-specific additions follow the base list.

**Base initialisms (mandatory in every gen config)**:

| Initialism | Meaning |
|------------|---------|
| IDS | Intrusion Detection System |
| JWT | JSON Web Token |
| JWK | JSON Web Key |
| JWE | JSON Web Encryption |
| JWS | JSON Web Signature |
| OIDC | OpenID Connect |
| SAML | Security Assertion Markup Language |
| AES | Advanced Encryption Standard |
| GCM | Galois/Counter Mode |
| CBC | Cipher Block Chaining |
| RSA | Rivest-Shamir-Adleman |
| EC | Elliptic Curve |
| HMAC | Hash-based Message Authentication Code |
| SHA | Secure Hash Algorithm |
| TLS | Transport Layer Security |
| IP | Internet Protocol |
| AI | Artificial Intelligence |
| ML | Machine Learning |
| KEM | Key Encapsulation Mechanism |
| PEM | Privacy Enhanced Mail |
| DER | Distinguished Encoding Rules |
| DSA | Digital Signature Algorithm |
| IKM | Input Keying Material |

**Domain-specific additions by service**:

| Service | Domain Additions |
|---------|----------------|
| `jose-ja` | JWKS, OKP, URI |
| `pki-ca` | CSR, CA, CRL, OCSP, URI, SAN, DN, CN, OU |
| `sm-im` | IM, SM, URI |
| `sm-kms` | URI |
| `skeleton-template` | (none — base list only) |
<!-- @/propagate -->

**Enforcement**: `lint-fitness gen-config-initialisms` verifies every `openapi-gen_config_server.yaml` contains the full base list.

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

<!-- @propagate to=".github/instructions/02-04.openapi.instructions.md" as="http-status-codes" -->
| Code | Usage |
|------|-------|
| 200 | GET, PUT, PATCH successful |
| 201 | POST (resource created) |
| 204 | DELETE successful |
| 400 | Validation error |
| 401 | Missing/invalid auth |
| 403 | Insufficient permissions |
| 404 | Resource not found |
| 409 | Duplicate/conflict |
| 422 | Semantic validation error |
| 500 | Unhandled server error |
| 503 | Temporary unavailability |
<!-- @/propagate -->

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

Three hierarchical levels: Suite (`cmd/{SUITE}/`), Product (`cmd/{PRODUCT}/`), Service (`cmd/{PS-ID}/`). All delegate to `internal/apps/` layers with subcommands: server, client, health, livez, readyz, shutdown, init, compose, e2e.

See [Section 4.4.7 CLI Patterns](#447-cli-patterns) for complete hierarchy, routing rules, and anti-patterns (no executables for subcommands).

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
        ServiceFrameworkServerSettings: cryptoutilFrameworkTestutil.NewTestSettings(),
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

#### 9.4.1 OTel Collector Processor Constraints

<!-- @propagate to=".github/instructions/02-03.observability.instructions.md" as="otel-collector-constraints" -->
| Processor | Requirement | Dev/CI | Production |
|-----------|------------|--------|------------|
| resourcedetection/docker | Docker socket `/var/run/docker.sock` | NEVER use | Use when socket available |
| resourcedetection/env | Environment variables | ALWAYS | ALWAYS |
| resourcedetection/system | OS hostname, IP | ALWAYS | ALWAYS |

**MANDATORY for dev/CI**: Use `detectors: [env, system]`. NEVER include `docker` detector without verified socket access.

**CRITICAL**: NEVER defer OTel or infrastructure configuration issues as "pre-existing." Infrastructure blockers are ALWAYS MANDATORY BLOCKING.
<!-- @/propagate -->

**Anti-Pattern**: NEVER defer OTel or infrastructure configuration issues as "pre-existing." Infrastructure blockers that prevent E2E validation MUST be fixed immediately — they are BLOCKING, not "nice-to-have."

#### 9.4.2 Docker Desktop and Testcontainers API Compatibility

**CRITICAL: Docker Desktop upgrades can break testcontainers.**

Docker Desktop version upgrades (e.g., .55 → .62) may introduce Docker API version mismatches with testcontainers libraries. Symptoms include:

- Containers fail to start with API version errors
- Health checks time out despite correct configuration
- Test infrastructure works locally but fails in CI/CD (or vice versa)
- Intermittent "connection refused" or "daemon not responding" errors

**Diagnosis Checklist** (when Docker-related tests fail after an upgrade):

1. Check Docker Desktop version: `docker version` (note both Client and Server API versions)
2. Check testcontainers-go version in `go.mod` — verify compatibility with Docker API version
3. Check Docker Compose version: `docker compose version`
4. Verify Docker daemon health: `docker info` (check for warnings)
5. Check for Docker Desktop settings changes (resource limits, WSL backend, etc.)
6. Try `docker system prune -f` to clear stale state
7. Verify network connectivity: `docker network ls`

**Resolution**: Update testcontainers-go dependency to match Docker Desktop API version. If testcontainers-go hasn't released a compatible version yet, pin Docker Desktop to the last working version until alignment.

**MANDATORY**: After any Docker Desktop upgrade, run the full E2E test suite before considering the upgrade complete. Docker API mismatches are BLOCKING issues.

### 9.5 Container Architecture

**Multi-Stage Dockerfile Pattern**:

```dockerfile
# Global ARGs
ARG GO_VERSION=1.26.1
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

| Workflow | Purpose |
|----------|---------|
| `ci-quality` | Linting, code quality, and CICD lint validation |
| `ci-coverage` | Test coverage collection and enforcement |
| `ci-mutation` | Mutation testing with gremlins |
| `ci-race` | Race condition detection |
| `ci-benchmark` | Performance benchmarking |
| `ci-fuzz` | Fuzz testing |
| `ci-fitness` | Architecture fitness function validation |
| `ci-sast` | Static application security testing |
| `ci-dast` | Dynamic application security testing |
| `ci-gitleaks` | Secret detection |
| `ci-e2e` | End-to-end integration testing |
| `ci-load` | Load testing with Gatling |
| `ci-identity-validation` | Identity service requirements validation |
| `release` | Automated release workflows |

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

| Action | Purpose |
|--------|---------|
| docker-compose-build | Build Docker Compose services |
| docker-compose-down | Tear down Docker Compose environment |
| docker-compose-logs | Collect Docker Compose logs |
| docker-compose-up | Start Docker Compose services |
| docker-compose-verify | Verify Docker Compose service health |
| docker-images-pull | Parallel Docker image pre-fetching |
| download-cicd | Download CI/CD tooling artifacts |
| fuzz-test | Run fuzz tests with controlled duration |
| go-setup | Go toolchain configuration |
| golangci-lint | Run golangci-lint with project config |
| security-scan-gitleaks | Scan for leaked secrets (Gitleaks) |
| security-scan-trivy | Scan Docker images for vulnerabilities (Trivy) |
| security-scan-trivy2 | Scan filesystem for vulnerabilities (Trivy) |
| workflow-job-begin | Standard job prologue (checkout, setup) |
| workflow-job-end | Standard job epilogue (artifacts, cleanup) |

See `.github/actions/` for the authoritative action catalog.

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

#### 9.9.3 UTF-8 Without BOM Enforcement

<!-- @propagate to=".github/instructions/03-05.linting.instructions.md" as="utf8-without-bom" -->
**MANDATORY**: UTF-8 without BOM for ALL text files. Enforcement via pre-commit hook `fix-byte-order-marker` (auto-fix) and `lint-text` sub-linter (in `cicd-lint-all` hook).

**PowerShell file writing MUST use UTF-8 without BOM — `Set-Content -Encoding UTF8` adds BOM in PowerShell 5.1:**

```powershell
# CORRECT — UTF-8 without BOM
[System.IO.File]::WriteAllText($path, $content, [System.Text.UTF8Encoding]::new($false))

# WRONG — adds BOM in PowerShell 5.1
Set-Content -Path $path -Value $content -Encoding UTF8  # ❌ BOM
```
<!-- @/propagate -->

#### 9.9.4 Platform Line-Ending Policy

The repository stores files with LF line endings (`\n`). Working-tree line endings are platform-native on developer machines.

**Policy** (MANDATORY):

- **Repository storage**: Always LF (`\n`). Git normalizes on commit.
- **Windows developers**: `git config --global core.autocrlf true` — git converts LF→CRLF on checkout, CRLF→LF on commit. Working tree has CRLF.
- **Linux/macOS developers**: `git config --global core.autocrlf input` — git converts CRLF→LF on commit; no conversion on checkout. Working tree has LF.
- **Local repo override BANNED**: `git config core.autocrlf false` in `.git/config` overrides the global setting and prevents CRLF working-tree checkout on Windows. NEVER set this local override.
- **AI agent behavior**: LLMs always write `\n` (LF) regardless of platform. With `core.autocrlf=true`, files written by AI agents are LF on disk until the next `git checkout`. This is acceptable — the `mixed-line-ending` pre-commit hook (default "auto") only modifies files with **mixed** CRLF+LF content; consistently LF-only files are not touched.
- **`mixed-line-ending` hook**: MUST NOT have `--fix lf` arg. Keep default "auto" mode.

**To fix if local override was set**:

```bash
git config --unset core.autocrlf          # remove local override
git config core.autocrlf                  # verify: empty = global takes effect
git config --global core.autocrlf         # verify: true (Windows) or input (Linux)
```

### 9.10 CICD Command Architecture

The `cicd-lint` CLI tool implements a strict directory-driven code organization pattern. Every command is enforced through a consistent four-layer dispatch, with three command naming categories.

#### 9.10.1 Code Flow

```
cmd/cicd-lint/main.go                          # Layer 1: Thin main(), os.Exit(Cicd(...))
  → internal/cmd/cicd_lint/cicd.go             # Layer 2: Validates command name, delegates to apps
    → internal/apps/tools/cicd_lint/cicd.go          # Layer 3: Unified dispatch switch, run()
      → internal/apps/tools/cicd_lint/<command>/     # Layer 4: Registered linters/formatters/scripts
        → internal/apps/tools/cicd_lint/<command>/<sub>/  # Sub-linters/formatters/scripts
```

**Strict Enforcement Rules**:

- Layer 1 (`cmd/cicd-lint/main.go`): ONLY `os.Exit()` + delegate. Zero logic.
- Layer 2 (`internal/cmd/cicd_lint/cicd.go`): ONLY command validation + usage display + delegate to Layer 3. Zero business logic.
- Layer 3 (`internal/apps/tools/cicd_lint/cicd.go`): Unified `run()` switch for ALL commands. Each command has a `const` declaration. `ValidCommands` in `internal/shared/magic/magic_cicd.go` MUST match the switch cases 1:1.
- Layer 4 (`internal/apps/tools/cicd_lint/<command>/`): Package-per-command. Entry point is `Lint()`, `Format()`, or `Cleanup()`/script-specific. Internal sub-linters/formatters registered in a `registeredLinters`/`registeredFormatters`/`registeredCleaners` slice.

#### 9.10.2 Command Naming Patterns

<!-- @propagate to=".github/instructions/04-01.deployment.instructions.md" as="cicd-command-naming" -->
**Three command categories** with strict naming and directory conventions:

| Category | Naming Pattern | Directory Pattern | Entry Function | Registration |
|----------|---------------|-------------------|----------------|-------------|
| **Linters** | `lint-<target>` | `lint_<target>/` | `Lint(logger)` | `registeredLinters` |
| **Formatters** | `format-<target>` | `format_<target>/` | `Format(logger, ...)` | `registeredFormatters` |
| **Scripts** | `<action>-<target>` | `<action>_<target>/` | Script-specific | `registeredCleaners` etc. |

**Linter commands** (13): `lint-text`, `lint-go`, `lint-go-test`, `lint-go-mod`, `lint-golangci`, `lint-compose`, `lint-ports`, `lint-workflow`, `lint-deployments`, `lint-docs`, `lint-fitness`, `lint-java-test`, `lint-python-test`
**Formatter commands** (2): `format-go`, `format-go-test`
**Script commands** (1): `github-cleanup`
<!-- @/propagate -->

#### 9.10.3 Registered Sub-Command Pattern

Each command package follows the **registered sub-command pattern**:

```go
// Type alias for sub-command functions.
type LinterFunc func(logger *common.Logger) error

// Registered sub-commands executed sequentially.
var registeredLinters = []struct {
    name   string
    linter LinterFunc
}{
    {"sub-linter-name", subPackage.Check},
}

// Entry point aggregates errors from all registered sub-commands.
func Lint(logger *common.Logger) error {
    var errors []error
    for _, l := range registeredLinters {
        if err := l.linter(logger); err != nil {
            errors = append(errors, err)
        }
    }
    // Return aggregated errors
}
```

**Sub-command packages**: Each sub-linter/formatter/script lives in its own subdirectory with a single `Check()` or `Fix()` or equivalent entry point. Example: `lint_go/circular_deps/circular_deps.go` exports `Check(logger) error`.

#### 9.10.4 Directory Structure

```
cmd/cicd-lint/main.go                                    # Thin main()
internal/cmd/cicd_lint/cicd.go                           # Validation + delegation
internal/apps/tools/cicd_lint/
├── cicd.go                                         # Unified dispatch (ALL commands)
├── common/                                         # Shared logger, summary, filter
├── docs_validation/                                # Core docs validation logic
├── lint_text/                                       # lint-text command
│   ├── lint_text.go                                # Lint() + registeredLinters
│   └── utf8/utf8.go                                # Sub-linter: UTF-8 enforcement
├── lint_go/                                          # lint-go command
│   ├── lint_go.go                                  # Lint() + registeredLinters (7 sub-linters)
│   ├── common/                                     # Shared helpers
│   ├── function_var_redeclaration/                 # Sub-linter: forbid var xxxFn = pkg.Func
│   ├── leftover_coverage/                          # Sub-linter: detect _coverage_ test files
│   ├── magic_aliases/                              # Sub-linter: import alias enforcement
│   ├── magic_duplicates/                           # Sub-linter: duplicate magic constants
│   ├── magic_usage/                                # Sub-linter: literal-use violations
│   ├── no_unaliased_cryptoutil_imports/            # Sub-linter: require aliases on internal imports
│   └── test_presence/                              # Sub-linter: require test files per package
├── lint_go_mod/                                      # lint-go-mod command
│   ├── lint_go_mod.go                              # Lint() + registeredLinters
│   └── outdated_deps/                              # Sub-linter
├── lint_gotest/                                      # lint-go-test command
│   ├── lint_gotest.go                              # Lint() + registeredLinters (4 sub-linters)
│   ├── common/                                     # Shared helpers (ListAllFiles, FilterExcludedTestFiles)
│   ├── require_over_assert/                        # Sub-linter: forbid assert.*; use require.*
│   ├── lint_gotest_hardcoded_uuid/                 # Sub-linter: forbid uuid.MustParse(literal)
│   ├── lint_gotest_real_http_server/               # Sub-linter: forbid httptest.NewServer(
│   └── lint_gotest_test_sleep/                     # Sub-linter: forbid time.Sleep(
├── lint_golangci/                                    # lint-golangci command
│   ├── lint_golangci.go                            # Lint() + registeredLinters
│   └── golangci_config/                            # Sub-linter
├── lint_compose/                                     # lint-compose command
│   ├── lint_compose.go                             # Lint() + registeredLinters
│   └── admin_port_exposure/                        # Sub-linter
├── lint_ports/                                       # lint-ports command
│   ├── lint_ports.go                               # Lint() + registeredLinters
│   ├── health_paths/                               # Sub-linter
│   ├── host_port_ranges/                           # Sub-linter
│   └── legacy_ports/                               # Sub-linter
├── lint_workflow/                                     # lint-workflow command
│   ├── lint_workflow.go                            # Lint() + registeredLinters
│   └── github_actions/                             # Sub-linter
├── lint_deployments/                                 # lint-deployments command
│   ├── lint.go                                     # Lint() entry point
│   ├── validate_all.go                             # ValidateAll() orchestrator
│   └── validate_*.go                               # 8 registered validators
├── lint_docs/                                         # lint-docs command
│   ├── lint_docs.go                                # Lint() + registeredLinters
│   ├── check_chunk_verification/                   # Sub-linter
│   ├── validate_chunks/                            # Sub-linter
│   └── validate_propagation/                       # Sub-linter
├── lint_fitness/                                      # lint-fitness command
│   ├── lint_fitness.go                             # Lint() + registeredLinters (68 sub-linters)
│   └── ... (68 sub-linters, see Section 9.11)
├── lint_javatest/                                     # lint-java-test command
│   └── lint_javatest.go                            # Lint() + CheckInsecureRandom()
├── lint_pythontest/                                   # lint-python-test command
│   └── lint_pythontest.go                          # Lint() + CheckUnittestAntipattern()
├── format_go/                                        # format-go command
│   ├── format_go.go                                # Format() + registeredFormatters
│   ├── copyloopvar/                                # Sub-formatter
│   ├── enforce_any/                                # Sub-formatter
│   └── enforce_time_now_utc/                       # Sub-formatter
├── format_gotest/                                    # format-go-test command
│   ├── format_gotest.go                            # Format() + registeredFormatters
│   └── thelper/                                    # Sub-formatter
└── github_cleanup/                                   # github-cleanup command
    ├── cleanup_api.go                              # Cleanup() + registeredCleaners
    ├── cleanup_runs.go                             # Sub-script
    ├── cleanup_artifacts.go                        # Sub-script
    └── cleanup_caches.go                           # Sub-script
```

#### 9.10.5 Enforcement Invariants

1. **1:1 mapping**: Every `const cmd*` in `cicd.go` MUST have a matching `switch case` in `run()`, a matching entry in `ValidCommands`, and a matching directory under `internal/apps/tools/cicd_lint/`.
2. **No scattered commands**: ALL cicd commands MUST be dispatched through the single `run()` function. No secondary dispatch paths.
3. **Directory = Command**: Directory name (with underscores) MUST match command name (with hyphens). Example: command `lint-go` → directory `lint_go/`.
4. **Entry point naming**: Linters export `Lint()`, formatters export `Format()`, scripts export their action verb (e.g., `Cleanup()`).
5. **Sub-commands are registered**: Sub-linters/formatters MUST be registered in a slice within the parent command's entry point file. No ad-hoc invocations.
6. **Test presence**: Every package under `internal/apps/tools/cicd_lint/` MUST have at least one `_test.go` file (enforced by `lint_go/test_presence` sub-linter).

#### 9.10.6 cicd-lint Command Constraints — MANDATORY

<!-- @propagate to=".github/instructions/04-01.deployment.instructions.md" as="cicd-lint-constraints" -->
**Purpose**: `cicd-lint` is exclusively for linting, formatting, and operational cleanup. It NEVER generates files, scaffolds content, or transforms the repository.

**Constraints** (NO EXCEPTIONS):

1. **Subcommands only**: `go run ./cmd/cicd-lint <subcommand> [<subcommand2> ...]` — the ONLY accepted arguments are subcommand names. No `--flags`, no `--ps-id=`, no customization parameters of any kind.
2. **Linting and formatting only**: Linter commands detect deviations from expected structure and return errors. Formatter commands auto-fix style issues. Neither generates new content.
3. **No content generation**: cicd-lint NEVER creates Dockerfiles, compose files, config overlays, secrets, migration files, or any other repository artifacts. The strategy is detect-and-error, not generate-and-apply.
4. **No Python under cicd_lint**: `internal/apps/tools/cicd_lint/` is pure Go. No Python scripts, modules, or helpers.
5. **Codify as validators**: When a new invariant is identified, implement it as a fitness linter that validates the actual state against expected state and returns descriptive errors. NEVER implement it as a generator that creates the expected state.
<!-- @/propagate -->

**Rationale**: The single source of truth is `docs/ARCHITECTURE.md` (prose). Its invariants are codified by a combination of pre-commit and pre-push hooks, including many `cicd-lint` subcommands. This strategy means ARCHITECTURE.md drives the repository, not generated files that can drift from the prose.

---

### 9.11 Architecture Fitness Functions

Architecture fitness functions are automated checks that enforce ARCHITECTURE.md invariants on every commit via `go run ./cmd/cicd-lint lint-fitness`. Violations are caught at pre-commit time and in CI, preventing architectural drift.

**Command**: `go run ./cmd/cicd-lint lint-fitness`
**Pre-commit hook**: `lint-fitness` (runs on `.go`, `.yml`, `.sql` changes)
**CI/CD integration**: `ci-quality` workflow includes lint-fitness

**Adding new fitness functions**: Use the `fitness-function-gen` Copilot skill — see `.github/skills/fitness-function-gen/SKILL.md`.

#### 9.11.1 Fitness Sub-Linter Catalog

**Category summary**:

| Category | Count | Examples |
|----------|-------|---------|
| Security | 7 | `crypto-rand`, `tls-minimum-version`, `non-fips-algorithms` |
| Architecture | 12 | `circular-deps`, `cmd-entry-whitelist`, `api-path-registry`, `subcommand-completeness` |
| Deployment & Config | 14 | `compose-service-names`, `secret-naming`, `unseal-secret-content` |
| Code Quality | 9 | `file-size-limits`, `cgo-free-sqlite`, `banned-product-names` |
| Testing | 7 | `parallel-tests`, `no-unit-test-real-db`, `test-patterns` |
| Service Framework | 6 | `health-endpoint-presence`, `health-path-completeness`, `service-contract-compliance` |
| Database & Migrations | 3 | `migration-numbering`, `migration-range-compliance` |

**Full catalog** (ordered by category):

| Sub-Linter | Rule Enforced |
|-----------|--------------|
| | **Security** |
| `admin-bind-address` | Go source: `BindPrivateAddress` must be `127.0.0.1`, never `0.0.0.0` (complements `admin-port-exposure` and `validate-admin` in lint_deployments) |
| `bind-address-safety` | Tests bind to `127.0.0.1`, never `0.0.0.0` |
| `crypto-rand` | Use `crypto/rand`, NEVER `math/rand` |
| `insecure-skip-verify` | No `InsecureSkipVerify: true` in production TLS config |
| `no-hardcoded-passwords` | No hardcoded credentials in test files |
| `non-fips-algorithms` | No bcrypt, scrypt, argon2, MD5, SHA-1 |
| `tls-minimum-version` | TLS config must specify `tls.VersionTLS13` minimum |
| | **Architecture** |
| `circular-deps` | No circular imports between packages |
| `cmd-anti-pattern` | `cmd/` directories must follow `cmd/{name}/main.go` pattern, no banned names |
| `cmd-entry-whitelist` | Only 18 allowed `cmd/` entries (1 suite + 5 products + 10 services + 2 infra tools) |
| `cmd-main-pattern` | `cmd/*/main.go` must delegate to `internalMain()`, no logic |
| `cross-service-import-isolation` | Service packages must not import other service packages |
| `domain-layer-isolation` | Domain layer must not import server/client/API packages |
| `entity-registry-completeness` | All PS in registry must have required magic constants |
| `product-structure` | Product packages must follow PRODUCT/SERVICE hierarchy |
| `product-wiring` | Product wiring must delegate to service entry points |
| `service-structure` | Service packages must follow PRODUCT/SERVICE layout convention |
| `api-path-registry` | OpenAPI specs must have paths matching the registry entry for each PS-ID; no paths allowed outside the declared `api_resources` list |
| `subcommand-completeness` | Service `cmd/*/main.go` must use the `route.Service()` entry point for all registered PS-IDs |
| | **Deployment & Configuration** |
| `compose-db-naming` | Compose DB service names must use `sqlite`/`postgres` not `pg` |
| `compose-header-format` | Compose files must have canonical comment header |
| `compose-service-names` | Compose service names must match `{ps-id}-{db}-N` pattern |
| `configs-deployments-consistency` | Every `deployments/{PS-ID}/` must have matching `configs/{PS-ID}/` |
| `configs-empty-dir` | `configs/` directories must not be empty (require `.gitkeep` or files) |
| `configs-naming` | `configs/` directories must follow flat `configs/{PS-ID}/` pattern from entity registry |
| `deployment-dir-completeness` | Every PS must have Dockerfile, compose.yml, secrets/, and config/ |
| `dockerfile-labels` | Dockerfile `org.opencontainers.image.title` LABEL matches deployment tier; `image.description` is non-empty |
| `otlp-service-name-pattern` | OTLP service names must match `{ps-id}-{db}-N` pattern |
| `secret-content` | Non-unseal secret values match documented format patterns: `{PREFIX}-hash-pepper-v3-{base64-random-32-bytes}`, `{PREFIX}-browser-user`, `{PREFIX}-browser-pass-{base64-random-32-bytes}`, `{PREFIX}-service-user`, `{PREFIX}-service-pass-{base64-random-32-bytes}`, `{PREFIX_UNDERSCORE}_database[_user/_pass-{base64-random-32-bytes}]`, postgres-url composition; `.secret.never` marker content enforced at product/suite tiers |
| `secret-naming` | All tiers use `{purpose}.secret` filenames; `.secret.never` markers enforced at product/suite tiers |
| `standalone-config-otlp-names` | Config file `otlp-service` values must match `{ps-id}-{db}-N` pattern |
| `standalone-config-presence` | All PS must have 5 config overlay files: `{PS-ID}-app-common.yml`, `{PS-ID}-app-sqlite-1.yml`, `{PS-ID}-app-sqlite-2.yml`, `{PS-ID}-app-postgresql-1.yml`, `{PS-ID}-app-postgresql-2.yml` |
| `template-consistency` | `deployments/skeleton-template/` uses hyphenated secret names (not underscores) |
| `unseal-secret-content` | Unseal secret values match `{TIER-PREFIX}-unseal-key-N-of-5-{base64-random-32-bytes}` (base64-encoded 32 random bytes, unique per shard, tier prefix matches deployment directory) |
| | **Code Quality** |
| `archive-detector` | `_archived/`, `archived/`, `orphaned/` directories must not exist |
| `banned-product-names` | Old product names (`Cipher IM`, `Cipher KMS`, etc.) banned from source |
| `cgo-free-sqlite` | Use `modernc.org/sqlite` not `mattn/go-sqlite3` (CGO banned) |
| `check-skeleton-placeholders` | No skeleton placeholder text left in service code |
| `file-size-limits` | Files ≤500 lines (warning at >300, error at >500) |
| `gen-config-initialisms` | Config field names must use correct Go initialisms |
| `legacy-dir-detection` | Legacy directories (`internal/apps/cipher/`) must not exist |
| `require-framework-naming` | Framework packages must use canonical naming conventions |
| `root-junk-detection` | No `*.exe`, `*.py`, `coverage*`, `*.test.exe` at project root |
| | **Testing** |
| `cicd-coverage` | `cicd-lint` sub-commands must have test coverage for registered linters/validators |
| `no-local-closed-db-helper` | No local `closedDB` helper — use shared test utility |
| `no-postgres-in-non-e2e` | PostgreSQL must not appear in non-E2E test files |
| `no-unit-test-real-db` | Unit tests must use in-memory SQLite, never real DB |
| `no-unit-test-real-server` | Unit tests must use `app.Test()`, never real server |
| `parallel-tests` | All tests must call `t.Parallel()` (with `// Sequential:` exemption) |
| `test-patterns` | Test file naming, table-driven structure compliance |
| | **Service Framework** |
| `health-endpoint-presence` | Services must expose `/admin/api/v1/livez` and `/admin/api/v1/readyz` |
| `health-path-completeness` | Services must document all 5 standard health paths in top-level Go files: `/service/api/v1/health`, `/browser/api/v1/health`, `/admin/api/v1/livez`, `/admin/api/v1/readyz`, `/admin/api/v1/shutdown` |
| `magic-e2e-compose-path` | `*E2EComposePath` constants must point to existing compose files |
| `magic-e2e-container-names` | `*E2ESQLiteContainer`/`*E2EPostgresContainer` constants must match compose names |
| `require-api-dir` | Services must have an `api/` directory |
| `service-contract-compliance` | Services must implement `ServiceServer` interface (`PublicBaseURL`, `AdminBaseURL`, `SetReady`) |
| | **Database & Migrations** |
| `migration-comment-headers` | Domain migrations (2001+) first comment must be `{DisplayName} database schema` |
| `migration-numbering` | Migration files must use `NNNN_name.up.sql` format |
| `migration-range-compliance` | Template migrations use 1001-1999; domain migrations use 2001+ |

#### 9.11.2 Fitness Linter Contract — Hard Error on Absent Dirs

**MANDATORY: ALL fitness linters must return a hard error when a required directory is absent.**

The majority pattern (`os.IsNotExist(err) → return nil`) silently hides compliance gaps and is CATEGORICALLY WRONG for fitness linters. Strict enforcement requires every linter to fail loudly when its required input does not exist.

**Rule**: When a `CheckInDir()` or equivalent function encounters an absent required directory, it MUST return `fmt.Errorf("directory not found: ...")` (or similar). It MUST NOT return `nil`.

**Rationale**: A missing directory is a signal of structural non-compliance, not a vacuously true state. Silently passing on absent input makes the linter useless for catching drift.

**Test Pattern**: Unit tests that create isolated `t.TempDir()` workspaces must create stubs for ALL directories that the linter requires to exist. For linters that iterate the entity registry (e.g., `database_key_uniformity`, `config_overlay_freshness`), use a registry-iterating helper to create stubs for all 10 PS-IDs. For template-directory linters (`migration_numbering`, `migration_range_compliance`), use a `createTemplateMigrationsDirStub()` helper.

**Structural ceiling**: When a template stub necessarily creates a parent directory that another absent-dir check depends on, direct testing via `CheckInDir` is structurally impossible. In these cases, the equivalent coverage MUST be provided by a direct unit test of the inner function (e.g., `TestFindDomainMigrationDirs_NoAppsDir_ReturnsError` covers the code path that `TestCheckInDir_NoInternalAppsDir_Fails` cannot reach via `CheckInDir`).

#### 9.11.3 Entity Registry

**Location**: `internal/apps/tools/cicd_lint/lint_fitness/registry/registry.go`

**YAML Source**: `api/cryptosuite-registry/registry.yaml` — single source of truth for all products and product-services. Loaded at startup via `gopkg.in/yaml.v3` with JSON Schema validation at `api/cryptosuite-registry/registry-schema.json`.

**Purpose**: All registry-driven fitness checks iterate `AllProductServices()` to detect drift without hardcoding names. The registry is loaded once via `init()` — malformed YAML panics at program start (fail-fast pattern).

**Types**:
- `Product` — ID, DisplayName, InternalAppsDir, CmdDir
- `ProductService` — PSID, Product, Service, DisplayName, InternalAppsDir, MagicFile, Entrypoint, APIResources

**Update Procedure**: When adding a new product-service:
1. Add entry to `api/cryptosuite-registry/registry.yaml` with all required fields
2. Run `go run ./cmd/cicd-lint lint-fitness` — `entity-registry-completeness` and `entity-registry-schema` will catch gaps
3. Add the corresponding magic constants to `internal/shared/magic/magic_*.go`
4. Add required deployment artifacts (Dockerfile, compose.yml, configs, secrets)

**Current registry**: 5 products, 10 product-services

#### 9.11.4 Naming Convention Catalog

**Key conventions**:
- Compose service names: `{ps-id}-{db}-N` (e.g. `sm-im-postgres-1`)
- OTLP service names: `{ps-id}-sqlite-1`, `{ps-id}-sqlite-2`, `{ps-id}-postgres-1`, `{ps-id}-postgres-2`
- Migration comment headers: `-- {DisplayName} database schema` / `-- {DisplayName} database schema rollback`
- Config overlay files: `{PS-ID}-app-common.yml`, `{PS-ID}-app-sqlite-1.yml`, `{PS-ID}-app-sqlite-2.yml`, `{PS-ID}-app-postgresql-1.yml`, `{PS-ID}-app-postgresql-2.yml`

---

## 10. Testing Architecture

### 10.1 Testing Strategy Overview

<!-- @propagate to=".github/instructions/03-02.testing.instructions.md" as="test-file-suffixes" -->
| Type | Suffix |
|------|--------|
| Unit | `_test.go` |
| Bench | `_bench_test.go` |
| Fuzz | `_fuzz_test.go` |
| Property | `_property_test.go` |
| Integration | `_integration_test.go` |
<!-- @/propagate -->

**Testing Pyramid**:

- **Unit Tests**: Fast (<15s per package), isolated, table-driven, t.Parallel()
- **Integration Tests**: TestMain pattern, shared resources, GORM repositories
- **E2E Tests**: Docker Compose, production-like, cross-service validation

<!-- @propagate to=".github/instructions/03-02.testing.instructions.md, .github/instructions/03-04.data-infrastructure.instructions.md" as="three-tier-database-strategy" -->
**3-Tier Database Strategy (MANDATORY)**:

| Tier | Database | Pattern | PostgreSQL? |
|------|----------|---------|-------------|
| Unit | SQLite in-memory | `testdb.NewInMemorySQLiteDB(t)` | NEVER |
| Integration | SQLite in-memory via TestMain | ONE shared instance per package | NEVER |
| E2E | Docker Compose PostgreSQL | 3 app instances (2 PostgreSQL + 1 SQLite) | YES (only here) |

**Key Rules**:
- NEVER use PostgreSQL in unit or integration tests — PostgreSQL tested ONLY in E2E.
- NEVER create DB per-test in integration tests (use TestMain shared instance).
- NEVER start real servers in unit tests (use Fiber app.Test()).
- E2E tests use Docker Compose with 3 service instances: 2 sharing a PostgreSQL container, 1 using in-memory SQLite, validating cross-database compatibility.
<!-- @/propagate -->

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
- **UUID Literal Construction**: Use `googleUuid.UUID{}` for nil UUID and `googleUuid.UUID{0xff, 0xff, ...}` for max UUID instead of `googleUuid.MustParse("00000000-...")` to satisfy the `test-patterns` fitness linter

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

**Coverage Ceiling Analysis** (MANDATORY before setting per-package targets):

1. Generate `go tool cover -html=coverage.out`
2. Categorize every uncovered line:
   - **Structurally testable**: Error paths reachable via seam injection or input manipulation
   - **Structurally unreachable**: Default switch cases for exhaustive type switches, dead code paths
   - **Third-party boundary**: Library return errors that require internal library state manipulation
3. Calculate ceiling: `ceiling = (total - unreachable) / total`
4. Set target at ceiling - 2% (safety margin)
5. Document exceptions in package README or coverage analysis file

**Package-Level Exceptions**: Packages MAY have targets below the mandatory minimum IF a coverage ceiling analysis documents the structural ceiling. Exception format:

| Package | Standard Target | Actual Target | Ceiling | Justification |
|---------|----------------|---------------|---------|---------------|
| internal/shared/crypto/jose | 95% | 95% | ~96% | JWE OKP branches unreachable |

#### 10.2.4 Test Seam Injection Pattern

**Purpose**: Enable error path testing for third-party library wrappers and dependency injection without interfaces or mocks.

**Standard: Function-Parameter Injection (MANDATORY)**

All production code MUST use function-parameter injection (passing `fn func(...)` parameters or struct fields) as the seam mechanism. Package-level `var xxxFn = pkg.Func` declarations are FORBIDDEN in production code.

```go
// Production code — function fields on struct
type SessionManager struct {
    generateRSAJWKFn func(rsaBits int) (joseJwk.Key, error)
    encryptBytesFn   func(jwks []joseJwk.Key, clear []byte) (*joseJwe.Message, []byte, error)
}

func NewSessionManager(ctx context.Context) (*SessionManager, error) {
    return &SessionManager{
        generateRSAJWKFn: joseJwkUtil.GenerateRSAJWK,
        encryptBytesFn:   joseJweUtil.EncryptBytes,
    }, nil
}

func (sm *SessionManager) initKey(bits int) (joseJwk.Key, error) {
    return sm.generateRSAJWKFn(bits) // indirected through struct field
}

// Test code — per-test struct field mutation (parallel-safe: sm is per-test instance)
func TestGenerateKey_Error(t *testing.T) {
    t.Parallel()
    sm := setupSessionManager(t)
    sm.generateRSAJWKFn = func(_ int) (joseJwk.Key, error) {
        return nil, fmt.Errorf("injected generate error")
    }
    _, err := sm.initKey(2048)
    require.ErrorContains(t, err, "injected generate error")
}
```

**For standalone functions** (not struct methods), pass the fn as a parameter:

```go
// Production code
func Lint(ctx context.Context, walkFn filepath.WalkFunc, readFileFn func(string) ([]byte, error)) error { ... }

// Test code
func TestLint_WalkError(t *testing.T) {
    t.Parallel()
    err := Lint(ctx, func(root string, fn fs.WalkDirFunc) error { return errors.New("walk fail") }, os.ReadFile)
    require.ErrorContains(t, err, "walk fail")
}
```

**Restricted Exception: Package-Level Vars for OS/Process Exits**

Package-level `var` is permitted ONLY for process-exit functions that cannot be injected through normal call paths (`os.Exit`, `log.Fatal`). Use `export_test.go` (not `seams.go`) to expose the var for testing:

```go
// Production code — osExit only, no other package-level seam vars
var osExit = os.Exit

// export_test.go — expose for tests only
var OsExit = &osExit

// Test code
func TestShutdownError(t *testing.T) {
    // Sequential: mutates osExit package-level var — cannot use t.Parallel()
    orig := *OsExit
    defer func() { *OsExit = orig }()
    *OsExit = func(code int) { panic(fmt.Sprintf("exit %d", code)) }
    require.Panics(t, func() { callShutdown() })
}
```

**Decision Matrix**:

| Scenario | Pattern | Parallel-Safe? |
|----------|---------|---------------|
| Struct method calls fn dep | Struct field (`sm.xxxFn`) | ✅ Yes (per-test instance) |
| Standalone function calls fn dep | fn parameter | ✅ Yes (per-call) |
| `os.Exit` / `log.Fatal` | Package-level var + `export_test.go` | ❌ No (Sequential:) |
| Business logic substitution | Interface injection | ✅ Yes |

**Coverage Impact**: Function-param injection enables full parallel error-path testing (3-8% coverage gain) without the sequential test constraint imposed by package-level seam variable mutation.

**FORBIDDEN**: `var xxxFn = pkg.Func` in non-test production files (except `osExit` / `log.Fatal` patterns). Use `golangci-lint` `no-pkg-seam-vars` fitness linter to enforce.

#### 10.2.5 Sequential Test Exemption

<!-- @propagate to=".github/instructions/03-02.testing.instructions.md" as="sequential-test-exemption" -->
Tests that mutate **package-level state** (e.g., `os.Chdir()`, global registries) MUST NOT call `t.Parallel()`. Add a `// Sequential:` comment within 10 lines before the function to exempt it from the `parallel_tests` linter:

```go
// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestMyFunction_ChangeDir(t *testing.T) {
    // no t.Parallel() here
}
```

```go
// Sequential: mutates registeredHandlers package-level state.
func TestRegisterHandler_Duplicate(t *testing.T) {
    // no t.Parallel() here
}
```

**Rule**: Comment MUST be within 10 lines before function declaration. Include a reason after the colon.
<!-- @/propagate -->

Seam variables (see §10.2.4) are a common cause of sequential tests.

#### 10.2.6 Test File Consolidation

**MANDATORY**: Prefer one test file per source file. Avoid proliferating small error-path test files (e.g., `*_error_mapping_test.go`, `*_gorm_errors_test.go`, `*_postgres_errors_test.go`). Instead, consolidate related error-path tests into thematic files grouped by error category.

**500-line hard limit per file**: When merging would exceed 500 lines, keep files separate with clear thematic grouping.

**Naming**: Use descriptive semantic names (`*_error_paths_test.go`, `*_factory_test.go`, `*_db_errors_test.go`) that describe WHAT is being tested, not WHY the test was written. NEVER use `*_coverage_test.go` or `*_gaps_test.go` — these describe motivation (hitting coverage) rather than domain. Use `*_test_util_test.go` to test test-utility functions.

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

#### 10.3.4 Test HTTP Client Patterns

<!-- @propagate to=".github/instructions/03-02.testing.instructions.md" as="disable-keep-alives-test-transport" -->
NEVER use a default `http.Transport` in integration tests calling a real server. ALWAYS set `DisableKeepAlives: true`:

```go
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // test certs only
        DisableKeepAlives: true, // REQUIRED: prevents 90-second shutdown hang
    },
    Timeout: 5 * time.Second,
}
```

**Why**: Fasthttp (Fiber) keeps an `open` counter > 0 while keep-alive connections remain open. `ShutdownWithContext` hangs for 90 seconds waiting for the counter to reach zero.
<!-- @/propagate -->

**Symptom**: Tests pass but teardown is extremely slow (≥90s per test binary); `TestMain` never completes in a reasonable time.

<!-- @propagate to=".github/instructions/03-02.testing.instructions.md" as="timeout-double-multiplication-antipattern" -->
NEVER multiply a `time.Duration` constant by `time.Second`. Magic constants that are already `time.Duration` (e.g., `DefaultDataServerShutdownTimeout = 5 * time.Second`) produce ~158-year values when multiplied again:

```go
// WRONG: DefaultDataServerShutdownTimeout is already time.Duration
ctx, cancel := context.WithTimeout(ctx, magic.DefaultDataServerShutdownTimeout * time.Second) // ~158 years!

// CORRECT: use directly
ctx, cancel := context.WithTimeout(ctx, magic.DefaultDataServerShutdownTimeout) // 5 seconds
```
<!-- @/propagate -->

#### 10.3.5 Cross-Service Contract Test Pattern

**Purpose**: Validate that all services conform to shared behavioral contracts (health endpoints, dual-server isolation, response format) without duplicating test code in each service.

**Entry Point**:

```go
import cryptoutilContract "cryptoutil/internal/apps/framework/service/testing/contract"

func TestMyService_ContractCompliance(t *testing.T) {
    t.Parallel()
    cryptoutilContract.RunContractTests(t, testServer)
}
```

**`ServiceServer` Interface** (required by `RunContractTests`):

```go
type ServiceServer interface {
    PublicBaseURL() string                 // base URL e.g. "https://127.0.0.1:8080"
    AdminBaseURL() string                  // base URL e.g. "https://127.0.0.1:9090"
    SetReady(ready bool)                   // used by RunReadyzNotReadyContract
    TLSRootCAPool() *x509.CertPool         // public server CA pool (no InsecureSkipVerify)
    AdminTLSRootCAPool() *x509.CertPool    // admin server CA pool (no InsecureSkipVerify)
    // ... (full interface: see internal/apps/framework/service/server/contract.go)
}
```

**TLS Accessors**: `TLSRootCAPool()` and `AdminTLSRootCAPool()` return the root CA certificate pools for the auto-generated ephemeral TLS chains. Test infrastructure uses these to create properly-trusted HTTP clients without `InsecureSkipVerify: true`. See [Section 10.3.7](#1037-tls-test-bundle-pattern) for the full pattern.

**What contracts are verified**: health endpoint liveness, readiness, dual-server isolation (public vs. admin port separation), and response format correctness.

**`SetReady(true)` Requirement**: If using `MustStartAndWaitForDualPorts` in a manual `TestMain`, you MUST call `server.SetReady(true)` explicitly after the server starts. The `testserver.StartAndWait` helper does this automatically.

#### 10.3.6 Shared Test Infrastructure

Shared test packages in `internal/apps/framework/service/testing/` eliminate TestMain boilerplate by providing reusable setup helpers.

**testdb** — Database setup helpers:

```go
import cryptoutilTestdb "cryptoutil/internal/apps/framework/service/testing/testdb"

// In-memory SQLite (fast, no cleanup needed)
db := cryptoutilTestdb.NewInMemorySQLiteDB(t)

// With auto-migrate for given models
db := cryptoutilTestdb.RequireNewInMemorySQLiteDB(t, &MyModel{})

// Pre-closed SQLite DB for error-path testing (returns already-closed *gorm.DB)
db := cryptoutilTestdb.NewClosedSQLiteDB(t, nil) // nil = no migrations
db := cryptoutilTestdb.NewClosedSQLiteDB(t, func(sqlDB *sql.DB) error { return migrate(sqlDB) })

// PostgreSQL test container (Docker required)
db := cryptoutilTestdb.RequireNewPostgresTestContainer(ctx, t, &MyModel{})
```

**testserver** — Test server lifecycle:

```go
import cryptoutilTestserver "cryptoutil/internal/apps/framework/service/testing/testserver"

// Start server with port 0 (dynamic), wait for ready, register cleanup
srv := cryptoutilTestserver.StartAndWait(ctx, t, myServiceServer)
```

**fixtures** — Test data creation:

```go
import cryptoutilFixtures "cryptoutil/internal/apps/framework/service/testing/fixtures"

tenant := cryptoutilFixtures.CreateTestTenant(t, db)
realm  := cryptoutilFixtures.CreateTestRealm(t, db, tenant.ID)
user   := cryptoutilFixtures.CreateTestUser(t, db, tenant.ID)
```

**assertions** — HTTP response assertions:

```go
import cryptoutilAssertions "cryptoutil/internal/apps/framework/service/testing/assertions"

cryptoutilAssertions.AssertHealthy(t, resp)                         // 200 OK
cryptoutilAssertions.AssertErrorResponse(t, resp, http.StatusBadRequest)
cryptoutilAssertions.AssertJSONContentType(t, resp)
cryptoutilAssertions.AssertTraceID(t, resp)
```

**healthclient** — HTTPS health endpoint client:

```go
import cryptoutilHealthclient "cryptoutil/internal/apps/framework/service/testing/healthclient"

client := cryptoutilHealthclient.NewHealthClient(srv.PublicBaseURL(), srv.AdminBaseURL())
livezResp, err  := client.Livez()
readyzResp, err := client.Readyz()
```

**coverage ceiling**: `testdb` (57.5%) and `e2e_infra` (37.3%) have documented ceilings due to Docker-dependent code paths unreachable in unit tests. All other packages: ≥95% production, ≥98% infrastructure.

#### 10.3.7 TLS Test Bundle Pattern

**Problem**: Tests starting real HTTPS servers need valid TLS certificate chains. `InsecureSkipVerify: true` is **prohibited** (gosec G402, semgrep `no-tls-insecure-skip-verify`).

**Solution**: Service servers expose `TLSRootCAPool()` and `AdminTLSRootCAPool()` accessors that return the x509.CertPool for the auto-generated ephemeral TLS chain. Test infrastructure uses these pools to create properly-trusted HTTP clients.

**TestMain Pattern**:

```go
// In TestMain: create service-specific HTTP clients after server starts.
testPublicHTTPClient = &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            MinVersion: tls.VersionTLS13,
            RootCAs:    testServer.TLSRootCAPool(),       // trusts the server's CA
        },
        DisableKeepAlives: true, // prevents 90-second shutdown hang
    },
    Timeout: 5 * time.Second,
}
testAdminHTTPClient = &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            MinVersion: tls.VersionTLS13,
            RootCAs:    testServer.AdminTLSRootCAPool(),  // trusts the admin CA
        },
        DisableKeepAlives: true, // prevents 90-second shutdown hang
    },
    Timeout: 5 * time.Second,
}
```

**Testutil Helpers** (shared test infrastructure for template-derived tests):

```go
import cryptoutilTestutil "cryptoutil/internal/apps/framework/service/server/testutil"

// Pre-configured cert pools from the shared test TLS bundle.
publicPool := cryptoutilTestutil.PublicRootCAPool()   // public server CA
adminPool  := cryptoutilTestutil.PrivateRootCAPool()  // admin server CA
```

**Rules** (ALL MANDATORY):

- ALWAYS use `testServer.TLSRootCAPool()` / `testServer.AdminTLSRootCAPool()` for server-specific tests
- ALWAYS use `cryptoutilTestutil.PublicRootCAPool()` / `cryptoutilTestutil.PrivateRootCAPool()` for template-derived tests
- ALWAYS set `DisableKeepAlives: true` to prevent fasthttp 90-second shutdown hang (see [Section 10.3.4](#1034-test-http-client-patterns))
- NEVER use `InsecureSkipVerify: true` in any test (caught by gosec G402 + semgrep `no-tls-insecure-skip-verify`)

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
      for secret in unseal-1of5 unseal-2of5 unseal-3of5 unseal-4of5 unseal-5of5 hash-pepper-v3 browser-username browser-password service-username service-password postgres-url postgres-username postgres-password postgres-database; do
        test -f /run/secrets/$${secret}.secret || exit 1;
      done
      echo 'All secrets validated'
  secrets:
    - unseal-1of5.secret
    - unseal-2of5.secret
    - unseal-3of5.secret
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

**Implementation**: See `internal/apps/framework/service/testing/e2e_infra/docker_health.go` for parseDockerComposePsOutput(), determineServiceHealthStatus(), WaitForServicesHealthy().

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
- Build tag: `//go:build fuzz` is an optional marker; fuzz tests run without it using `-fuzz=FuzzXxx` directly
- Property tests that must not run during fuzzing: use `//go:build !fuzz` at the top of `_property_test.go`

### 10.8 Benchmark Testing Strategy

**Mandatory for crypto operations**:

```go
func BenchmarkAESEncrypt(b *testing.B) {
    key := make([]byte, 32)
    plaintext := make([]byte, 1024)
    b.SetBytes(1024)     // enables MB/s throughput reporting
    b.ReportAllocs()     // report allocations per op
    b.ResetTimer()       // exclude setup time
    for i := 0; i < b.N; i++ {
        _, _ = encrypt(key, plaintext)
    }
}

// Per-iteration setup (use StopTimer/StartTimer)
func BenchmarkWithSetup(b *testing.B) {
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        b.StopTimer()               // pause timer for setup
        input := prepareInput()
        b.StartTimer()              // resume timer for measured work
        _, _ = processInput(input)
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

**Local Testing**: go run ./cmd/cicd-workflow -workflows=dast,e2e -inputs="key=value"

**Act Compatibility**: NEVER use -t timeout, ALWAYS specify -workflows, use -inputs for params

### 10.13 Java / Gatling Load Test Standards

Java Gatling simulations in `test/load/src/test/java/cryptoutil/` MUST follow these standards, enforced by `cicd-lint lint-java-test`.

**Secure RNG (MANDATORY)**:

- ALWAYS use `java.security.SecureRandom` — FIPS 140-3 compliance requires it
- NEVER use `new Random()` (java.util.Random) — non-cryptographic, predictable
- NEVER use `Math.random()` — delegates to java.util.Random internally

```java
// CORRECT
import java.security.SecureRandom;
private static final SecureRandom SECURE_RANDOM = new SecureRandom();

// WRONG — detected by lint-java-test
private static final Random RANDOM = new Random();          // new Random()
private double roll() { return Math.random(); }             // Math.random()
```

**Parameterization — MANDATORY**:

All configurable values (base URLs, user counts, durations) MUST use `System.getProperty()` with a default:

```java
private static final String BASE_URL = System.getProperty("baseUrl", "https://localhost:8080");
private static final int    USERS    = Integer.parseInt(System.getProperty("users", "10"));
private static final int    DURATION = Integer.parseInt(System.getProperty("durationSeconds", "60"));
```

**Simulation pattern**:

```java
import java.security.SecureRandom;
import io.gatling.javaapi.core.*;
import io.gatling.javaapi.http.*;
import static io.gatling.javaapi.core.CoreDsl.*;
import static io.gatling.javaapi.http.HttpDsl.*;

public class ApiSimulation extends Simulation {
    private static final SecureRandom SECURE_RANDOM = new SecureRandom();
    private static final String BASE_URL = System.getProperty("baseUrl", "https://localhost:8080");
    private static final int    USERS    = Integer.parseInt(System.getProperty("users", "1"));

    HttpProtocolBuilder protocol = http.baseUrl(BASE_URL).disableFollowRedirect();

    ScenarioBuilder scn = scenario("Health check")
        .exec(http("health").get("/service/api/v1/health").check(status().is(200)));

    { setUp(scn.injectOpen(atOnceUsers(USERS))).protocols(protocol); }
}
```

**Violations detected by `lint-java-test`**:

| Pattern | Violation | Fix |
|---------|-----------|-----|
| `new Random()` | Non-FIPS RNG | Replace with `new SecureRandom()` |
| `Math.random()` | Non-FIPS RNG | Replace with `secureRandom.nextDouble()` |

**Execution**: `mvn gatling:test -pl test/load -Dgatling.simulationClass=cryptoutil.ApiSimulation`

### 10.14 Python / pytest Standards

Python test files (when present in `test/` or elsewhere) MUST use pytest style, enforced by `cicd-lint lint-python-test`. The linter checks files named `test_*.py` or `*_test.py` only.

**Required: pytest standalone functions**:

```python
# CORRECT — pytest style
import pytest

@pytest.mark.parametrize("value,expected", [
    ("valid",   True),
    ("invalid", False),
])
def test_validate_input(value, expected):
    result = validate_input(value)
    assert result == expected
```

**Prohibited: unittest.TestCase inheritance**:

```python
# WRONG — detected by lint-python-test
import unittest

class MyTest(unittest.TestCase):
    def test_something(self):
        self.assertEqual(result, expected)   # self.assert*()
```

**pytest fixtures (parameterization)**:

```python
@pytest.fixture
def api_client(base_url):
    return ApiClient(base_url)

def test_health_check(api_client):
    resp = api_client.get("/service/api/v1/health")
    assert resp.status_code == 200
```

**Violations detected by `lint-python-test`** (in `test_*.py` / `*_test.py` files only):

| Pattern | Violation | Fix |
|---------|-----------|-----|
| `class X(unittest.TestCase)` | unittest inheritance | Use standalone `def test_*()` |
| `from unittest import TestCase` | unittest import | Use `import pytest` |
| `self.assertEqual(...)` etc. | self.assert* calls | Use bare `assert` or `pytest.raises()` |

**Execution**: `pytest test/ -v --tb=short`

---

## 11. Quality Architecture

### 11.1 Maximum Quality Strategy - MANDATORY

<!-- @propagate to=".github/instructions/01-02.beast-mode.instructions.md" as="quality-attributes" -->
**Quality Attributes (NO EXCEPTIONS)**:
- ✅ Correctness: ALL code functionally correct with comprehensive tests
- ✅ Completeness: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- ✅ Thoroughness: Evidence-based validation at every step
- ✅ Reliability: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ Efficiency: Optimized for maintainability and performance NOT implementation speed
- ✅ Accuracy: Changes must address root cause, not just symptoms
- ❌ Time Pressure: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ Premature Completion: NEVER mark phases or tasks or steps complete without objective evidence
<!-- @/propagate -->

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
- Current Version: 1.26.1 (check go.mod)
- Enforcement Locations: go.mod (go 1.26.1), .github/workflows/*.yml (GO_VERSION: '1.26.1'), Dockerfile (FROM golang:1.26.1-alpine), README.md (document Go 1.26.1+ requirement)
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

<!-- @propagate to=".github/instructions/03-03.golang.instructions.md" as="crypto-acronyms-caps" -->
**Crypto Acronyms**: ALWAYS ALL CAPS: RSA, EC, ECDSA, ECDH, HMAC, AES, JWA, JWK, JWS, JWE, ED25519, PKCS8, PEM, DER.
<!-- @/propagate -->

**Import Safety When Replacing Function Bodies**: When replacing function bodies, imports used by the OLD body may be accidentally removed even though they are still needed elsewhere in the file. ALWAYS run `go vet` after import changes to catch missing or unused imports.

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

See [Section 14.5.5 Docker Desktop Startup](#1455-docker-desktop-startup---critical) for cross-platform startup instructions (Windows, macOS, Linux).

Here are local convenience commands to run the workflows locally for Development and Testing.

`go run ./cmd/cicd-workflow -workflows=build`     → build check
`go run ./cmd/cicd-workflow -workflows=coverage`  → workflow coverage check; ≥98% required
`go run ./cmd/cicd-workflow -workflows=quality`   → workflow quality check
`go run ./cmd/cicd-workflow -workflows=lint`      → linting check
`go run ./cmd/cicd-workflow -workflows=benchmark` → workflow benchmark check
`go run ./cmd/cicd-workflow -workflows=fuzz`      → workflow fuzz check
`go run ./cmd/cicd-workflow -workflows=race`      → workflow race check
`go run ./cmd/cicd-workflow -workflows=sast`      → static security analysis
`go run ./cmd/cicd-workflow -workflows=gitleaks`  → secrets scanning
`go run ./cmd/cicd-workflow -workflows=dast`      → dynamic security testing
`go run ./cmd/cicd-workflow -workflows=mutation`  → mutation testing; ≥95% required
`go run ./cmd/cicd-workflow -workflows=e2e`       → end-to-end tests; BOTH `/service/**` AND `/browser/**` paths
`go run ./cmd/cicd-workflow -workflows=load`      → load testing
`go run ./cmd/cicd-workflow -workflows=ci`        → full CI workflow

**Mutation Testing Scope**: ALL `cmd/cicd-lint/` packages (including `lint_deployments/`) require ≥98% mutation testing efficacy. This includes test infrastructure, CLI wiring, and validator implementations. Mutation testing validates test quality, not just test coverage.

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

<!-- @propagate to=".github/instructions/03-01.coding.instructions.md" as="format-go-protection" -->
**MANDATORY Prevention Rules**:
- NEVER change ` +""+interface{}+""+ ` to ` +""+ny+""+ ` in format_go package
- NEVER simplify CRITICAL/SELF-MODIFICATION comments
- ALWAYS read complete package context (enforce_any.go, filter.go, magic_cicd.go, format_go_test.go, self_modification_test.go) before modifying
<!-- @/propagate -->

**Root Cause**: LLM agents lose exclusion context during narrow-focus refactoring
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

**Dockerfile Parameterization by Tier**:

All Dockerfiles follow identical multi-stage structure. Parameterized fields differ by deployment tier:

| Field | Service (`{PS-ID}`) | Product (`{PRODUCT}`) | Suite (`{SUITE}-suite`) |
|-------|---------------------|----------------------|-------------------------|
| `image.title` LABEL | `{SUITE}-{PS-ID}` | `{SUITE}-{PRODUCT}` | `{SUITE}` |
| Binary built | `./cmd/{SUITE}` (always suite binary) | `./cmd/{SUITE}` | `./cmd/{SUITE}` |
| `EXPOSE` | 8080 (container public) | Service-range (e.g., 18000) | Suite-range (e.g., 28000) |
| `HEALTHCHECK` | `wget --no-check-certificate -qO- https://127.0.0.1:8080/browser/api/v1/health` | Same, product port | Same, suite port |
| `ENTRYPOINT` | `["/app/{SUITE}", "{PS-ID}", "start"]` | `["/app/{SUITE}", "{PRODUCT}", "start"]` | `["/app/{SUITE}"]` |

**Current state**: 10 service-level + 1 suite-level Dockerfiles exist. 0 product-level Dockerfiles exist (v6 CREATE).

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

<!-- @propagate to=".github/instructions/04-01.deployment.instructions.md" as="docker-compose-rules" -->
- Use `docker compose` (NOT `docker-compose`)
- ALWAYS relative paths in compose.yml (NEVER absolute)
- ALWAYS `127.0.0.1` in containers (NOT `localhost` - Alpine resolves to IPv6)
- Use `wget` for healthchecks (available in Alpine)
- Healthcheck fields use hyphens: `start-period` (NOT `start_period`)
<!-- @/propagate -->

- Secret management via Docker secrets (MANDATORY)
- Health check configuration (interval, timeout, retries, start-period)
- Dependency ordering (depends_on with service_healthy)
- Network isolation patterns

##### Docker Secrets (MANDATORY)

```yaml
secrets:
  postgres-password.secret:
    file: ./secrets/postgres-password.secret  # chmod 440

services:
  postgres:
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres-password.secret
    secrets:
      - postgres-password.secret
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

1. **Job-only** (validation job, ExitCode=0 required):
```yaml
healthcheck-secrets:
  image: alpine:latest
  command: ["sh", "-c", "test -f /run/secrets/unseal-1of5.secret || exit 1"]
```

1. **Service with healthcheck job** (external sidecar):
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

**Secrets Structure**: Each service requires 14 secrets (5 unseal keys, 1 hash pepper, 2 browser credentials, 2 service credentials, 4 PostgreSQL credentials).

##### Secret Naming Convention

**MANDATORY**: All secret files use **identical `{purpose}.secret` filenames** at ALL deployment tiers (service, product, suite). NO tier prefix on filenames. The **value inside** each secret file contains the tier-specific prefix (`{PS-ID}-`, `{PRODUCT}-`, or `{SUITE}-`).

**`.secret.never` Marker Files**: Product and suite tiers include `.secret.never` files to document that browser and service credentials are intentionally service-level only. These markers are NOT present at the service tier.

See [Section 4.4.6](#446-deployments) for the complete secret file listing at each tier.

##### Secret Value Format By Tier

| Secret | Service Value | Product Value | Suite Value |
|--------|---------------|---------------|-------------|
| `hash-pepper-v3.secret` | `{PS-ID}-hash-pepper-v3-{base64-random-32-bytes}` | `{PRODUCT}-hash-pepper-v3-{base64-random-32-bytes}` | `{SUITE}-hash-pepper-v3-{base64-random-32-bytes}` |
| `browser-username.secret` | `{PS-ID}-browser-user` | `.never` marker | `.never` marker |
| `browser-password.secret` | `{PS-ID}-browser-pass-{base64-random-32-bytes}` | `.never` marker | `.never` marker |
| `service-username.secret` | `{PS-ID}-service-user` | `.never` marker | `.never` marker |
| `service-password.secret` | `{PS-ID}-service-pass-{base64-random-32-bytes}` | `.never` marker | `.never` marker |
| `postgres-username.secret` | `{PS_ID}_database_user` | `{PRODUCT}_database_user` | `{SUITE}_database_user` |
| `postgres-password.secret` | `{PS_ID}_database_pass-{base64-random-32-bytes}` | `{PRODUCT}_database_pass-{base64-random-32-bytes}` | `{SUITE}_database_pass-{base64-random-32-bytes}` |
| `postgres-database.secret` | `{PS_ID}_database` | `{PRODUCT}_database` | `{SUITE}_database` |
| `postgres-url.secret` | `postgres://{PS_ID}_database_user:{password}@{PS-ID}-postgres:5432/{PS_ID}_database?sslmode=disable` | `...@{PRODUCT}-postgres:5432/...` | `...@{SUITE}-postgres:5432/...` |
| `unseal-{N}of5.secret` | `{PS-ID}-unseal-key-N-of-5-{base64-random-32-bytes}` | `{PRODUCT}-unseal-key-N-of-5-{base64-random-32-bytes}` | `{SUITE}-unseal-key-N-of-5-{base64-random-32-bytes}` |

**Encoding Notation**: `{base64-random-32-bytes}` = base64 encoding of 32 cryptographically random bytes. `{password}` = the full `{PS_ID}_database_pass-{base64-random-32-bytes}` value from `postgres-password.secret`.

**`.secret.never` Marker Content**: Product-level markers contain `MUST NEVER be used at product level. Use service-specific secrets.` Suite-level markers contain `MUST NEVER be used at suite level. Use service-specific secrets.`

**Note**: `{PS_ID}` uses underscores (e.g., `jose_ja`) for PostgreSQL identifiers; `{PS-ID}` uses hyphens (e.g., `jose-ja`) for all other contexts.

##### SUITE-Level Deployment (cryptoutil)

**Location**: `deployments/cryptoutil/secrets/`

**Consistency Requirements**:

- **hash-pepper-v3.secret**: SAME value across ALL 10 services (enables cross-service PII deduplication, SSO username lookup, for suite-level PostgreSQL OLAP database).
- **unseal-{N}of5.secret**: UNIQUE per service (barrier encryption independence, security isolation).
- **postgres-*.secret**: Service-specific.

**Rationale**: Unified hash pepper allows username@domain lookups across identity services while maintaining per-service encryption boundaries.

##### PRODUCT-Level Deployment (identity, jose, pki, skeleton, sm)

**Location**: `deployments/{PRODUCT}/secrets/`

**Consistency Requirements**:

- **hash-pepper-v3.secret**: SAME value within product services (identity-{authz,idp,rs,rp,spa} share pepper for SSO, for product-level PostgreSQL OLAP database).
- **unseal-{N}of5.secret**: UNIQUE per service within product (independent barrier hierarchies).
- **postgres-*.secret**: Service-specific.

**Example**: identity-authz, identity-idp, identity-rs, identity-rp, identity-spa share hash pepper for unified user lookups, for product-level PostgreSQL OLAP database.

##### SERVICE-Level Deployment (single service)

**Location**: `deployments/{PS-ID}/secrets/` (e.g., `deployments/jose-ja/secrets/`)

**Consistency Requirements**:

- **hash-pepper-v3.secret**: UNIQUE per service (no cross-service lookups needed, design intent is to block service-level PostgreSQL OLAP database).
- **unseal-{N}of5.secret**: UNIQUE per service (barrier encryption independence).
- **postgres-*.secret**: Service-specific.

**Rationale**: Maximum isolation for standalone deployments in non-federated mode, even if they share the same infrastructure.

##### Secret File Format

| Secret | Format | Generation |
|--------|--------|-----------|
| `unseal-{N}of5.secret` | `{PREFIX}-unseal-key-N-of-5-{base64-random-32-bytes}` | `base64.b64encode(secrets.token_bytes(32)).decode()` |
| `hash-pepper-v3.secret` | `{PREFIX}-hash-pepper-v3-{base64-random-32-bytes}` | `base64.urlsafe_b64encode(secrets.token_bytes(32)).rstrip(b'=')` |
| `browser-username.secret` | `{PREFIX}-browser-user` | Static |
| `browser-password.secret` | `{PREFIX}-browser-pass-{base64-random-32-bytes}` | `base64.urlsafe_b64encode(secrets.token_bytes(32)).rstrip(b'=')` |
| `service-username.secret` | `{PREFIX}-service-user` | Static |
| `service-password.secret` | `{PREFIX}-service-pass-{base64-random-32-bytes}` | `base64.urlsafe_b64encode(secrets.token_bytes(32)).rstrip(b'=')` |
| `postgres-database.secret` | `{PREFIX_UNDERSCORE}_database` | Static |
| `postgres-username.secret` | `{PREFIX_UNDERSCORE}_database_user` | Static |
| `postgres-password.secret` | `{PREFIX_UNDERSCORE}_database_pass-{base64-random-32-bytes}` | `base64.urlsafe_b64encode(secrets.token_bytes(32)).rstrip(b'=')` |
| `postgres-url.secret` | `postgres://{user}:{pass}@{PREFIX}-postgres:5432/{db}?sslmode=disable` | Composed from above |

**Encoding Notation**: `{base64-random-32-bytes}` = base64 encoding of 32 cryptographically random bytes. `{PREFIX}` = deployment-specific prefix (PS-ID, PRODUCT, or SUITE). `{PREFIX_UNDERSCORE}` = prefix with hyphens replaced by underscores for PostgreSQL identifiers.

**Enforcement**: `secret-content` fitness linter validates ALL secret file content matches these patterns. `unseal-secret-content` fitness linter validates unseal-specific patterns (value uniqueness, shard matching). `secret-naming` fitness linter validates filenames.

**Note**: `{PREFIX_UNDERSCORE}` uses underscores for PostgreSQL identifiers; `{PREFIX}` uses hyphens for all other contexts. Unseal keys are used for HKDF deterministic derivation; hash pepper for PBKDF2/HKDF PII hashing.

##### TLS Secrets (Static/Mixed Modes Only)

Unit/integration tests use Auto TLS (no secrets needed). See [Section 6.11](#611-tls-certificate-configuration) for TLS mode taxonomy.

| Secret File | TLS Mode | Purpose |
|-------------|----------|---------|
| `tls_server_cert.secret` | Static | PEM server certificate |
| `tls_server_key.secret` | Static + Mixed | PEM server private key |
| `tls_ca_cert.secret` | Static + Mixed | PEM CA certificate chain |
| `tls_issuing_ca_key.secret` | Mixed | PEM CA signing key (runtime generation) |

**TLS Mode Detection**: Static = cert + key provided; Mixed = CA cert + CA key provided; Auto = no TLS secrets → self-generated ephemeral certificates.

##### Secret Validation

**Healthcheck-Secrets Service**: All compose templates include a `healthcheck-secrets` service (Alpine) that validates all 14 secrets exist at `/run/secrets/` on startup. Fails fast on missing secrets to prevent runtime errors. See any `compose.yml` in `deployments/` for the canonical implementation.

##### Cross-Reference Documentation

- Security architecture: [02-05.security.instructions.md](../.github/instructions/02-05.security.instructions.md#secret-management---mandatory)
- Hash service patterns: [02-05.security.instructions.md](../.github/instructions/02-05.security.instructions.md#hash-service-architecture)
- Deployment structure: [Section 4.4.6 Deployments](#446-deployments)

#### 12.3.4 Multi-Level Deployment Hierarchy

**Purpose**: Enable flexible deployment granularity from single-service development to full suite production deployment while maintaining consistent configuration and secret management across all levels.

**Implementation Status**: Implemented and validated (2026-02-16)

**MANDATORY: Rigid Delegation Pattern** (enforced by linter):
- **SUITE → PRODUCT → SERVICE**: ALL deployments MUST follow this delegation chain
- **Rationale**:
  1. Products can scale from 1 to N services without breaking suite-level deployment
  2. Suite can scale from 5 to N products without hardcoded service dependencies
  3. Enables independent testing at each level (service, product, suite)

**Linter Enforcement** (`cicd-lint lint-deployments`):
- Suite compose MUST include product-level (e.g., `../sm/compose.yml`), NOT service-level (e.g., `../sm-kms/compose.yml`)
- Product compose MUST include service-level (e.g., `../sm-kms/compose.yml`)
- Violations are ERRORS that block CI/CD

**Three-Tier Hierarchy**:

| Level | Directory | Scope | Services | Use Cases |
|-------|-----------|-------|----------|-----------|
| **SERVICE** | `deployments/{PS-ID}/` | Single service | 1 | Development, testing, isolated deployment |
| **PRODUCT** | `deployments/{PRODUCT}/` | Product services | 1-5 | Product-level testing, SSO within product |
| **SUITE** | `deployments/cryptoutil/` | All services | 10 | Full integration, cross-product federation |

##### Docker Compose `include` Semantics

Docker Compose `include` merges services from different compose files into a single project with shared networking. Key behaviors (validated with Docker Compose v2.40+):

| Scenario | Result |
|----------|--------|
| Same secret name + same file path across includes | Merged (deduplicated) |
| Same secret name + different file paths across includes | **CONFLICT ERROR** |
| Different secret names + different files | Works |
| Same infrastructure service included by multiple files | Correctly deduplicated |

**Implications**:
- `depends_on` references work across included files within the same project.
- Secret names MUST be globally unique within a compose project unless pointing to the same file.
- Infrastructure services (telemetry, postgres) included by multiple service files appear once in the merged configuration.

##### Deployment Composition Patterns

All tiers follow the same Docker Compose pattern with `include` + shared `hash-pepper-v3.secret`. The difference is which level they include and what pepper value prefix they use:

| Tier | `include:` targets | Pepper value prefix | Hash pepper scope |
|------|-------------------|--------------------|--------------------|
| **SUITE** | Product composes (`../sm/`, `../pki/`, etc.) | `cryptoutil-` | Cross-product SSO, PII dedup |
| **PRODUCT** | Service composes (`../identity-authz/`, etc.) | `{PRODUCT}-` | SSO within product |
| **SERVICE** | None (direct service definition) | `{PS-ID}-` | Maximum isolation |

**Port Offset Strategy** (prevents conflicts when running multiple tiers simultaneously):

| Level | Offset | Example (sm-kms base 8080/9090) |
|-------|--------|----------------------------------|
| SERVICE | +0 | 8080 (public), 9090 (admin) |
| PRODUCT | +10000 | 18080 (public), 19090 (admin) |
| SUITE | +20000 | 28080 (public), 29090 (admin) |

##### Layered Pepper Strategy

1. **SERVICE** (`{PS-ID}-` prefix): Unique per service, NO cross-service lookups.
2. **PRODUCT** (`{PRODUCT}-` prefix): Shared within product services (e.g., 5 identity services for SSO).
3. **SUITE** (`{SUITE}-` prefix): Shared by ALL 10 services for cross-product federation.

All tiers use the identical filename `hash-pepper-v3.secret` — the tier is selected by which `secrets/` directory the compose file references. This simplifies compose `include` merging.

##### Linter Validation

`cicd-lint lint-deployments` validates ALL deployment tiers automatically (no parameters).

**PRODUCT Deployment Validation Rules**:

```
✅ Required: compose.yml exists
✅ Required: Dockerfile exists
✅ Required: secrets/ directory exists
✅ Required: hash-pepper-v3.secret exists in secrets/
✅ Required: browser-username.secret.never marker exists in secrets/
✅ Required: service-username.secret.never marker exists in secrets/
✅ Forbidden: unseal-*.secret MUST NOT exist (service-level only)
```

**SUITE Deployment Validation Rules**:

```
✅ Required: compose.yml exists
✅ Required: secrets/ directory exists
✅ Required: hash-pepper-v3.secret exists in secrets/
✅ Required: browser-username.secret.never marker exists in secrets/
✅ Required: service-username.secret.never marker exists in secrets/
✅ Forbidden: unseal-*.secret MUST NOT exist (service-level only)
```

**Coverage**: Linter validates ALL deployments (10 SERVICE, 5 PRODUCT, 1 SUITE, 2 shared infrastructure).

**Implementation**: [internal/apps/tools/cicd_lint/lint_deployments/lint_deployments.go](/internal/apps/tools/cicd_lint/lint_deployments/) with `validateProductSecrets()` and `validateSuiteSecrets()` functions.

##### Cross-Reference Documentation

- **Secrets coordination**: [12.3.3 Secrets Coordination Strategy](#1233-secrets-coordination-strategy)
- **Deployment validation**: [13.1 Deployment Structure Validation](#131-deployment-structure-validation)
- **Port assignments**: [3.4.1 Port Design Principles](#341-port-design-principles)

## 13. Deployment Tooling & Validation

### 13.1 Deployment Structure Validation

**Purpose**: Automated enforcement of consistent deployment directory structures across all services to prevent configuration drift and deployment failures.

**Linter Tool**: `cicd-lint lint-deployments` validates ALL deployments in `deployments/` and `configs/` directories.

**Implementation**: [internal/apps/tools/cicd_lint/lint_deployments](/internal/apps/tools/cicd_lint/lint_deployments/) package with comprehensive table-driven tests.

#### 13.1.1 Deployment Types

**SUITE** (e.g., cryptoutil - all 10 services):
- Required directories: `secrets/`
- Required files: `compose.yml`
- Optional files: `README.md`
- Required secrets: hash pepper, postgres credentials, unseal keys (14 `.secret` files, same as service tier)
- `.secret.never` marker files: `browser-username.secret.never`, `browser-password.secret.never`, `service-username.secret.never`, `service-password.secret.never`
  - Rationale: Browser and service credentials are service-level only
- Validation function: `validateSuiteSecrets()` in lint_deployments.go

**PRODUCT** (e.g., identity, sm, pki, jose, skeleton):
- Required directories: `secrets/`
- Required files: `compose.yml`, `Dockerfile`
- Optional files: `README.md`
- Required secrets: hash pepper, postgres credentials, unseal keys (14 `.secret` files, same as service tier)
- `.secret.never` marker files: `browser-username.secret.never`, `browser-password.secret.never`, `service-username.secret.never`, `service-password.secret.never`
  - Rationale: Browser and service credentials are service-level only
- Validation function: `validateProductSecrets()` in lint_deployments.go

**PRODUCT-SERVICE** (e.g., sm-im, jose-ja, pki-ca, sm-kms, identity-authz/idp/rp/rs/spa, skeleton-template):
- Required directories: `secrets/`, `config/`
- Required files: `compose.yml`, `Dockerfile`
- Optional files: `otel-collector-config.yaml`, `README.md`
- Required secrets: 14 files
  - `unseal-1of5.secret` through `unseal-5of5.secret`
  - `hash-pepper-v3.secret`
  - `postgres-url.secret`, `postgres-username.secret`, `postgres-password.secret`, `postgres-database.secret`
  - `browser-username.secret`, `browser-password.secret` (web UI auth)
  - `service-username.secret`, `service-password.secret` (headless/API auth)

**infrastructure** (shared-postgres, shared-telemetry):
- Required directories: none
- Required files: `compose.yml`
- Optional files: `init-db.sql`, `README.md`
- Required secrets: none (infrastructure secrets are optional)

#### 13.1.2 Validation Rules

**Directory Structure**: Each deployment type enforces specific required/optional directories.

**File Requirements**: compose.yml is MANDATORY for all types; Dockerfile MANDATORY only for PRODUCT-SERVICE.

**Secret Validation**: For PRODUCT-SERVICE types, all 14 required secrets MUST exist in `secrets/` directory. For PRODUCT and SUITE types, all 14 `.secret` files plus 4 `.secret.never` marker files MUST exist.

**Error Reporting**: Linter identifies missing directories, missing files, and missing secrets with actionable error messages.

**Rigid Delegation Pattern** (enforced):
- **SUITE Compose**: MUST include PRODUCT-level paths (e.g., `../sm/compose.yml`), NEVER service-level (e.g., `../sm-kms/compose.yml`)
- **PRODUCT Compose**: MUST include SERVICE-level paths (e.g., `../sm-kms/compose.yml`)
- **Validation Function**: `checkDelegationPattern()` in lint_deployments.go
- **Failure Mode**: Violations are ERRORS that block CI/CD

**Database Isolation** (enforced):
- Each of 10 services MUST have unique `postgres-database.secret` value
- Each of 10 services MUST have unique `postgres-username.secret` value
- Duplicate database names or usernames across services are ERRORS
- **Validation Function**: `checkDatabaseIsolation()` in lint_deployments.go
- **Cross-Service Check**: Runs after all deployments validated to detect sharing violations

**Authentication Credentials** (enforced):
- Each service MUST have 4 credential files: `browser-username.secret`, `browser-password.secret`, `service-username.secret`, `service-password.secret`
- **Validation Function**: `checkBrowserServiceCredentials()` in lint_deployments.go
- **Rationale**: No hardcoded credentials in config files (E2E testing requires Docker secrets)

**OTLP Protocol Override** (NEW - enforced 2026-02-16):
- Config files SHOULD NOT specify protocol in `otlp-endpoint` (no `grpc://` or `http://` prefixes)
- Use hostname:port format (e.g., `opentelemetry-collector-contrib:4317`)
- **Validation Function**: `checkOTLPProtocolOverride()` in lint_deployments.go
- **Failure Mode**: Violations are WARNINGS (non-blocking)

#### 13.1.3 CI/CD Integration

**GitHub Actions Workflow**: [cicd-lint-deployments.yml](/.github/workflows/cicd-lint-deployments.yml) runs on all changes to `deployments/**`.

**Quality Gate**: Deployment structure validation is MANDATORY before merge; violations block CI/CD pipeline.

**Artifact Upload**: Validation reports uploaded as GitHub Actions artifacts for 7-day retention.

#### 13.1.4 Cross-Reference Documentation

- Secrets coordination strategy: [12.3.3 Secrets Coordination Strategy](#1233-secrets-coordination-strategy)
- Docker Compose deployment patterns: [12.3.1 Docker Compose Deployment](#1231-docker-compose-deployment)
- Secret management instructions: [02-05.security.instructions.md](../.github/instructions/02-05.security.instructions.md#secret-management---mandatory)

#### 13.1.5 Config File Naming Strategy

**MANDATORY Pattern**: All service config files MUST use full `{PS-ID}-app-{variant}.yml` naming.

**Standard Config File Set** (5 required for ALL services):

- `{PS-ID}-app-common.yml` - Shared configuration for all deployment modes (bind addresses, TLS, network).
- `{PS-ID}-app-sqlite-1.yml` - SQLite in-memory instance 1 configuration.
- `{PS-ID}-app-sqlite-2.yml` - SQLite in-memory instance 2 configuration (REQUIRED for parity with PostgreSQL instances).
- `{PS-ID}-app-postgresql-1.yml` - PostgreSQL logical instance 1 configuration (shared database).
- `{PS-ID}-app-postgresql-2.yml` - PostgreSQL logical instance 2 configuration (shared database, high-availability pair).

**Optional Config Files**:

- Additional domain-specific files may exist with valid `{PS-ID}-` prefix naming.

**Examples**:

```
deployments/{PS-ID}/config/         # e.g., {PS-ID}=sm-kms
├── {PS-ID}-app-common.yml         # sm-kms-app-common.yml
├── {PS-ID}-app-sqlite-1.yml       # sm-kms-app-sqlite-1.yml
├── {PS-ID}-app-sqlite-2.yml       # sm-kms-app-sqlite-2.yml
├── {PS-ID}-app-postgresql-1.yml   # sm-kms-app-postgresql-1.yml
└── {PS-ID}-app-postgresql-2.yml   # sm-kms-app-postgresql-2.yml
```

**Rationale**:

- **Explicit PS-ID Coupling**: Prevents config file collisions when multiple services deployed together.
- **Variant Clarity**: `app-sqlite-1` vs `app-postgresql-1` makes deployment mode immediately obvious.
- **Instance Numbering**: `-1` and `-2` suffixes enable horizontal scaling with unique configs per instance.
- **SQLite Parity**: Two SQLite instances (`sqlite-1`, `sqlite-2`) mirror the two PostgreSQL instances for consistent testing across both database backends, but with different database data sharing vs non-sharing goals.
- **Tooling Support**: Linter validates presence of 5 required files, flags non-conformant naming.

**Migration Strategy** (Q9 Answer: Break Immediately):

- NO backward compatibility period - rename all files immediately
- NO symlinks or aliases - clean cutover
- **Rationale**: Pre-production repository with zero deployed instances, rigid enforcement prevents future drift

#### 13.1.6 Deprecated File Handling

**Decision**: Remove deprecated files from all `deployments/{PS-ID}/config/` directories.

**Removed Files**:

- `demo-seed.yml` — FORBIDDEN (removed; demo orchestration deferred until E2E foundation mature).
- `integration.yml` — FORBIDDEN (removed; integration test data uses TestMain with test-containers, NOT Docker Compose).
- `{PS-ID}-demo.yml` — FORBIDDEN (removed; demo orchestration deferred until E2E foundation mature).

**Rationale**:

- Demo orchestration remains deferred; the E2E orchestration foundation is now established (sm-im, sm-kms, jose-ja, and skeleton-template have full E2E test suites). Demo support will be designed to reuse E2E patterns when prioritized.
- Integration test data belongs in Go test code (TestMain + test-containers), not Docker Compose config files.

**Linter Enforcement**: Strict mode. Presence of any deprecated file is an ERROR that blocks CI/CD.

#### 13.1.7 Linter Validation Modes

**Current Mode**: Strict (all violations block CI/CD)

**ALL violations are errors (blocking)**:

- Config files not matching `{PS-ID}-app-{variant}.yml` pattern.
- Presence of deprecated `demo-seed.yml`, `integration.yml`, or `{PS-ID}-demo.yml` files.
- Missing required config files (5 standard files).
- Missing required secrets (14 secret files for service tier).
- Missing required directories (`secrets/`, `config/`).
- Missing required compose/Dockerfile files.
- Single-part deployment names (must be `{PS-ID}` format, e.g., `sm-kms`, `jose-ja`).
- Wrong PS-ID prefix in config file names.

#### 13.1.8 Config File Content Validation

**Implementation**: `ValidateConfigFile()` in [internal/apps/tools/cicd_lint/lint_deployments/validate_config.go](/internal/apps/tools/cicd_lint/lint_deployments/validate_config.go)

**Validation Rules**:

1. **YAML Syntax**: File must parse as valid YAML
2. **Bind Address Format**: Must be valid IPv4 (via `net.ParseIP`)
3. **Port Range**: 1-65535 inclusive
4. **Protocol**: Must be `https` (TLS required)
5. **Admin Bind Policy**: `bind-private-address` MUST be `127.0.0.1`
6. **Secret References**: `database-url` must use `file:///run/secrets/` or `sqlite://` (never inline `postgres://`)
7. **OTLP Consistency**: When `otlp: true`, `otlp-service` and `otlp-endpoint` are required

**Triggered by**: `cicd-lint lint-deployments` (no parameters; validates all config files automatically)

#### 13.1.9 Compose File Content Validation

**Implementation**: `ValidateComposeFile()` in [internal/apps/tools/cicd_lint/lint_deployments/validate_compose.go](/internal/apps/tools/cicd_lint/lint_deployments/validate_compose.go)

**Validation Rules**:

1. **Port Conflicts**: No duplicate host port bindings across services
2. **Health Checks**: All non-exempt services must have healthcheck configuration
3. **Dependency Chains**: `depends_on` references must resolve to defined services
4. **Secret References**: Referenced secrets must be defined in the `secrets:` section
5. **Hardcoded Credentials**: Environment variables must not contain inline passwords
6. **Bind Mount Security**: Host paths must use relative paths (no absolute paths)
7. **Include Resolution**: Docker Compose `include` directives are resolved for cross-file validation

**Triggered by**: `cicd-lint lint-deployments` (no parameters; validates all compose files automatically)

#### 13.1.10 Structural Mirror Validation

**Implementation**: `ValidateStructuralMirror()` in [internal/apps/tools/cicd_lint/lint_deployments/validate_mirror.go](/internal/apps/tools/cicd_lint/lint_deployments/validate_mirror.go)

**Direction**: `deployments/` → `configs/` (one-way). Every deployment directory MUST have a `configs/` counterpart.

**Mapping Rules**:

- Identity mapping (1:1): deployment directory name = `configs/` directory name
- Only exception: `cryptoutil` → `cryptoutil`

**Exclusions**: Infrastructure deployments (`shared-postgres`, `shared-telemetry`, `compose`, `template`)

**Orphan Handling**: Orphaned `configs/` directories produce **errors** (blocking). Each `configs/` directory MUST have a corresponding `deployments/` directory. No archive directory — git history preserves all deleted files (see Section 14.9).

**Triggered by**: `cicd-lint lint-deployments` (no parameters; validates `deployments/` → `configs/` mirror automatically)

#### 13.1.11 Validation Pipeline Architecture

**Execution Model**: All 8 validators run sequentially with aggregated error reporting. Sequential execution ensures deterministic output ordering — each validator's errors appear in the same order across runs, simplifying CI/CD log analysis. The internal `ValidateAll()` Go function orchestrates the pipeline; it is NOT a CLI subcommand. Each validator produces a `ValidationResult` with pass/fail status, error list, and timing metrics. Results are collected and a unified summary printed at the end.

**Note on Concurrency**: Sequential execution was chosen for output determinism. A concurrent collection model (validators run in parallel, errors merged after join) is architecturally feasible and would improve throughput for large deployments — but is not currently implemented.

```
┌─────────────────────────────────────────────────────────┐
│                  validate-all orchestrator               │
│                                                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │ naming   │→ │kebab-case│→ │  schema  │→ │ template │ │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘ │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │  ports   │→ │telemetry │→ │  admin   │→ │ secrets  │ │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘ │
│                                                         │
│  Result: N passed / M failed (Xms)                      │
└─────────────────────────────────────────────────────────┘
```

**8-Validator Reference**:

| # | Validator | Scope | Purpose | Key Rules |
|---|-----------|-------|---------|-----------|
| 1 | **ValidateNaming** | `deployments/`, `configs/` | Enforce kebab-case directory and file naming | All directories/files must be lowercase kebab-case. |
| 2 | **ValidateKebabCase** | `configs/` YAML files | Enforce kebab-case in YAML keys and compose service names | Top-level YAML keys must use kebab-case; `x-` extension keys and `services` map entries validated. |
| 3 | **ValidateSchema** | `configs/` `config-*.yml` files | Validate service template config files against hardcoded schema | Required keys (`bind-public-protocol`, `bind-public-address`, `bind-public-port`, etc.); protocol must be `https`; addresses must be valid IP/hostname. |
| 4 | **ValidateTemplatePattern** | `deployments/template/` | Validate skeleton-template structure, naming, and placeholder values | Required files present, config files use placeholder names, secret files use hyphenated names. |
| 5 | **ValidatePorts** | `deployments/` compose files | Validate port assignments per deployment type | SERVICE: 8000-8999, PRODUCT: 18000-18999, SUITE: 28000-28999; admin always 9090; no port conflicts. |
| 6 | **ValidateTelemetry** | `configs/` YAML files | Validate OTLP endpoint consistency | OTLP endpoint required in service configs; hostname:port format (no protocol prefix); consistent collector naming. |
| 7 | **ValidateAdmin** | `deployments/` compose files | Validate admin endpoint bind policy | Admin port must bind to `127.0.0.1:9090` (never `0.0.0.0`); ensures admin API never exposed outside container. |
| 8 | **ValidateSecrets** | `deployments/` compose files | Detect inline secrets in environment variables | Secret-pattern env vars (PASSWORD, SECRET, TOKEN, KEY, API_KEY) must use Docker secrets (`/run/secrets/`); length threshold ≥32/43 chars for non-reference values. Infrastructure deployments excluded. |

**Cross-References**: Each validator is implemented in `internal/apps/tools/cicd_lint/lint_deployments/validate_<name>.go` with comprehensive table-driven tests in `validate_<name>_test.go`. See code comments for detailed validation rules (per Decision 9:A minimal docs, comprehensive code comments).

### 13.2 Config File Architecture

**Purpose**: Centralized configuration management for all services with a consistent directory hierarchy mirroring the deployment structure.

**Schema Strategy**: Config file schema is HARDCODED in Go (`validate_schema.go`) with comprehensive code comments. The validator source code is the single source of truth for schema rules.

**Directory Structure** (flat `configs/{PS-ID}/` pattern):

```
configs/
├── cryptoutil/
│   └── cryptoutil.yml                          # Suite-level config
├── identity-authz/
│   ├── identity-authz.yml                      # Domain config (nested YAML)
│   └── domain/
│       └── policies/                           # Authentication/authorization policies
│           ├── adaptive-authorization.yml
│           ├── risk-scoring.yml
│           └── step-up.yml
├── identity-idp/
│   └── identity-idp.yml                        # Domain config (nested YAML)
├── identity-rp/
│   └── identity-rp.yml                         # Domain config (nested YAML)
├── identity-rs/
│   └── identity-rs.yml                         # Domain config (nested YAML)
├── identity-spa/
│   └── identity-spa.yml                        # Domain config (nested YAML)
├── jose-ja/
│   └── jose-ja.yml                             # Domain config (nested YAML)
├── pki-ca/
│   ├── pki-ca.yml                              # Domain config (nested YAML)
│   └── profiles/                               # X.509 certificate profiles (25 files)
│       ├── tls-server.yaml
│       ├── root-ca.yaml
│       └── ...
├── skeleton-template/
│   └── skeleton-template.yml                   # Domain config (nested YAML)
├── sm-im/
│   └── sm-im.yml                               # Domain config (nested YAML)
└── sm-kms/
    └── sm-kms.yml                              # Domain config (nested YAML)
```

**Key Rules**:

- **Flat {PS-ID} directories**: `configs/{PS-ID}/` (e.g., `configs/jose-ja/`, `configs/sm-kms/`). NOT nested `configs/{PRODUCT}/{SERVICE}/`.
- **Domain config naming**: `{PS-ID}.yml` (e.g., `jose-ja.yml`, `sm-im.yml`).
- **Special subdirectories**: `configs/pki-ca/profiles/` for X.509 profiles, `configs/identity-authz/domain/policies/` for auth policies.

**Config File Naming Conventions**:

| Type | Naming Pattern | Schema Format | Examples |
|------|---------------|---------------|----------|
| Domain config | `{PS-ID}.yml` | Nested YAML, service-specific | `jose-ja.yml`, `sm-im.yml` |
| Suite config | `cryptoutil.yml` | Suite-level settings | `configs/cryptoutil/cryptoutil.yml` |
| Certificate profile | `profiles/*.yaml` | X.509 certificate definitions | `tls-server.yaml` |
| Auth policy | `domain/policies/*.yml` | Authentication/authorization rules | `adaptive-authorization.yml` |
| Certificate schema | `*-config-schema.yaml` | CA certificate schema definitions | `pki-ca-config-schema.yaml` |

**Dual configs/ vs deployments/config/ Relationship**: The `configs/` directory holds **standalone development configs** for direct `go run` usage. The `deployments/{PS-ID}/config/` directories hold **Docker Compose deployment configs** (`{PS-ID}-app-{variant}.yml`) with 5 required variant files (common, sqlite-1, sqlite-2, postgresql-1, postgresql-2). Both follow the same flat kebab-case schema for service framework configs.

**Cross-References**: Schema validation rules in [validate_schema.go](/internal/apps/tools/cicd_lint/lint_deployments/validate_schema.go). Config naming in [Section 12.4.5](#1245-config-file-naming-strategy).

### 13.3 Secrets Management in Deployments

**Purpose**: Enforce Docker secrets usage for all credentials in compose files, preventing inline secret exposure in version-controlled YAML.

**Docker Secrets Pattern**: All secret-bearing environment variables (PASSWORD, SECRET, TOKEN, KEY, API_KEY) MUST reference Docker secrets via `/run/secrets/<name>` or `file:///run/secrets/<name>`. Inline values are violations.

**File Permissions**: All `.secret` files MUST have 440 (r--r-----) permissions. Never commit actual secret values to version control.

**Detection Strategy**: Length-based threshold (≥32 bytes / ≥43 base64 chars) identifies high-entropy inline values. Safe references (`/run/secrets/`, Docker secret names, short dev defaults) are excluded. No entropy calculation (too many false positives). Infrastructure deployments (Grafana, OTLP collector) are excluded from secrets validation since they use intentional inline dev credentials.

**Secret Naming Convention**: All tiers use **identical `{purpose}.secret` filenames** — no tier prefix on filenames. The VALUE inside each secret contains the tier-specific prefix (`{PS-ID}-`, `{PRODUCT}-`, or `{SUITE}-`). See [Section 4.4.6](#446-deployments) for the complete secret file listing at each tier.

**`.secret.never` Marker Files**: Product and suite tiers include `.secret.never` files (e.g., `browser-username.secret.never`, `service-password.secret.never`) to document that browser and service credentials are intentionally service-level only. These markers are enforced by `validateProductSecrets()` and `validateSuiteSecrets()`.

**Legacy Patterns (BANNED)**:

- Tier-suffixed secrets: `unseal_1of5-SERVICEONLY.secret`, `hash_pepper_v3-SHARED.secret` (use simple `{purpose}.secret` naming).
- Underscore-based secrets: `hash_pepper_v3.secret`, `unseal_1of5.secret` (use hyphenated `hash-pepper-v3.secret`, `unseal-1of5.secret`).
- Product-prefixed filenames: `sm-hash-pepper.secret`, `jose-unseal-1of5.secret` (prefix goes in VALUE, not filename).

**Cross-References**: Secrets coordination strategy in [Section 12.3.3](#1233-secrets-coordination-strategy). Validator implementation in [validate_secrets.go](/internal/apps/tools/cicd_lint/lint_deployments/validate_secrets.go).

### 13.4 Documentation Propagation Strategy

#### 13.4.1 Design Intent

**Problem**: Different Copilot modes of operation (VS Code Chat, CLI agents, Cloud agents, custom agents) read different file sets. Custom agents do NOT read `.github/instructions/*.instructions.md`. CLI and Cloud agents may not read `.github/copilot-instructions.md`. Keeping all file sets synchronized is error-prone.

**Solution**: ARCHITECTURE.md is the **absolute single source of truth**. Content is propagated to downstream files using **chunk-based verbatim copying** with HTML comment markers. A deterministic CI/CD validator verifies propagation integrity.

**MANDATORY**: Changes to ARCHITECTURE.md MUST be propagated to ALL downstream files in the SAME commit. Infrastructure changes (Docker, OTel, testcontainers, CI/CD) are ALWAYS BLOCKING — NEVER deferred.

#### 13.4.2 Propagation Marker System

**Marker Format in ARCHITECTURE.md** (source):

```html
<!-- @propagate to=".github/instructions/02-06.authn.instructions.md" as="key-principles" -->
content here (verbatim body text)
<!-- @/propagate -->
```

**Marker Format in Instruction Files** (target):

```html
<!-- @source from="docs/ARCHITECTURE.md" as="key-principles" -->
content here (verbatim copy of source)
<!-- @/source -->
```

**Attributes**:
- `to`: Relative path from repository root to the target file (ARCHITECTURE.md markers only)
- `from`: Relative path from repository root to the source file (target file markers only)
- `as`: Unique chunk identifier within the source-target pair (kebab-case)

**Content Between Markers**:
- MUST be identical in source and target (byte-for-byte after whitespace normalization)
- MUST NOT contain section headings (headings go OUTSIDE markers, allowing different heading levels)
- MUST NOT contain `See [ARCHITECTURE.md ...]` cross-reference links (those go OUTSIDE markers as glue)
- MUST be self-contained body text: paragraphs, bullet lists, tables, code blocks, bold/italic

**Content Outside Markers** (non-propagated glue):
- Section headings (## in instruction files, #### in ARCHITECTURE.md)
- `See [ARCHITECTURE.md Section X.Y](...)` cross-reference links
- Transitional paragraphs connecting propagated chunks
- YAML frontmatter (instruction files only)

**Formal Grammar** (BNF-like, for validator implementors):

```
@propagate-open  ::= '<!-- @propagate to="' PATH_LIST '" as="' CHUNK_ID '" -->'
@propagate-close ::= '<!-- @/propagate -->'
@source-open     ::= '<!-- @source from="' PATH '" as="' CHUNK_ID '" -->'
@source-close    ::= '<!-- @/source -->'
PATH_LIST        ::= PATH ( ', ' PATH )*
PATH             ::= [a-zA-Z0-9_./-]+
CHUNK_ID         ::= [a-z0-9-]+
```

Any variant not matching the above grammar (e.g., `@propagate from=...`, `@source to=...`, alternative quoting) will be silently missed by the validator — this grammar defines the enforced contract.

#### 13.4.3 Propagation Rules

**One-to-many**: One ARCHITECTURE.md chunk MAY propagate to multiple target files. Use a comma-separated `to` attribute: `to="file-a.md, file-b.md"`. The validator splits on comma-space and creates one propagation block per target with identical content. Avoid separate duplicate blocks.

**Chunk granularity**: Propagate the smallest self-contained unit. Prefer one chunk per logical concept (a table, a rule set, a code block with explanation). Do NOT wrap entire sections in a single marker.

**Heading-agnostic**: Headings are NEVER inside markers. This allows ARCHITECTURE.md to use `####` while instruction files use `##` for the same content.

**Link transformation**: Content inside markers uses NO internal cross-references. All `See [Section X.Y](#anchor)` references go outside markers as instruction file glue.

**Whitespace normalization**: CI/CD comparison ignores leading/trailing blank lines within markers and normalizes line endings (CRLF → LF). All other whitespace (indentation, inline spaces) MUST match exactly.

#### 13.4.3.1 Anchor Stability Policy

**After any section renumbering**:
- MUST grep `.github/**/*.md` and `docs/**/*.md` for old anchor patterns and update all broken references.
- SHOULD prefer stable named anchors (`#documentation-propagation-strategy`) over numbered anchors (`#134-documentation-propagation-strategy`) for sections frequently referenced from agent, skill, or instruction files.
- The `lint-docs validate-propagation` broken-reference check is the safety net — treat any broken references as blocking.

#### 13.4.4 Section-Level Mapping

| ARCHITECTURE.md Section | Primary Instruction File(s) | Agent File(s) |
|------------------------|----------------------------|---------------|
| 1. Executive Summary | (none — context only) | — |
| 2. Strategic Vision | 01-01.terminology, 01-02.beast-mode, 02-02.versions | implementation-planning, implementation-execution, beast-mode |
| 3. Product Suite | 02-01.architecture | — |
| 4. System Architecture | 02-01.architecture, 03-03.golang | — |
| 5. Service Architecture | 02-01.architecture, 03-04.data-infrastructure | — |
| 6. Security Architecture | 02-05.security, 02-06.authn | — |
| 7. Data Architecture | 03-04.data-infrastructure | — |
| 8. API Architecture | 02-04.openapi | — |
| 9. Infrastructure Architecture | 02-03.observability, 04-01.deployment, 03-05.linting | fix-workflows |
| 10. Testing Architecture | 03-02.testing | implementation-execution |
| 11. Quality Architecture | 03-05.linting, 03-01.coding, 06-01.evidence-based | implementation-planning, beast-mode |
| 12. Deployment Architecture | 04-01.deployment, 02-05.security | fix-workflows |
| 13. Deployment Tooling & Validation | 04-01.deployment | fix-workflows |
| 14. Development Practices | 05-02.git, 03-01.coding, 03-03.golang, 05-01.cross-platform | implementation-planning, implementation-execution |
| 15. Operational Excellence | 02-03.observability | — |
| Appendix A-C | (reference only) | — |

#### 13.4.5 CI/CD Validation

**Command**: `cicd-lint lint-docs`

**Algorithm**:
1. Parse all `@propagate` markers in ARCHITECTURE.md → extract (target_file, chunk_id, content)
2. For each target, parse `@source` markers → extract (source_file, chunk_id, content)
3. Normalize whitespace (trim leading/trailing blank lines, LF line endings)
4. Compare source content with target content byte-for-byte
5. Report mismatches with diff output showing exact divergence

**Exit codes**: 0 = all chunks match, 1 = divergence detected

**CI/CD integration**: Runs in `cicd-lint-docs` workflow on every push/PR. Blocks merge on divergence.

#### 13.4.6 Feasibility Constraints

**Verbatim propagation IS feasible** for ~80% of content with the following constraints:

| Content Type | Verbatim? | Rationale |
|-------------|-----------|-----------|
| Rules, requirements, constraints | ✅ Yes | Self-contained, no context dependency |
| Tables (algorithms, ports, methods) | ✅ Yes | Structured data, works in any context |
| Code blocks with annotations | ✅ Yes | Self-contained examples |
| Bullet lists of specifications | ✅ Yes | Enumerated facts |
| Prose with internal cross-references | ❌ No | `See Section X.Y` links differ between source/target |
| Section headings | ❌ No | Different heading levels per document |
| Transitional paragraphs | ❌ No | Context-dependent narrative flow |

**Non-propagated content** (~20%) is structural glue: headings, `See` references, transitions. This content exists in both documents but is NOT required to match verbatim.

#### 13.4.7 Migration Strategy

Propagation markers are added incrementally:
1. Start with highest-value instruction files (most-referenced, most-divergent)
2. For each section: reconcile content direction (ARCHITECTURE.md → instruction file)
3. Add `@propagate` markers in ARCHITECTURE.md
4. Add `@source` markers in instruction files with verbatim copy
5. Run `validate-propagation` to confirm match
6. Repeat for remaining sections

**Completed propagation chunks**:

| Chunk ID | ARCHITECTURE.md Section | Target File(s) |
|----------|------------------------|----------------|
| rfc-2119-keywords | 1.2 | 01-01.terminology |
| emphasis-keywords | 1.2 | 01-01.terminology |
| abbreviations | 1.2 | 01-01.terminology |
| quality-attributes | 11.1 | 01-02.beast-mode |
| end-of-turn-commit-protocol | 2.4 | 01-02.beast-mode |
| mandatory-review-passes | 2.5 | 01-02.beast-mode, 06-01.evidence-based, .github/agents/beast-mode, .github/agents/fix-workflows, .github/agents/implementation-execution, .github/agents/implementation-planning, .claude/agents/beast-mode, .claude/agents/fix-workflows, .claude/agents/implementation-execution, .claude/agents/implementation-planning |
| infrastructure-blocker-escalation | 14.7 | 01-02.beast-mode, 06-01.evidence-based |
| service-framework-components | 5.1.1 | 02-01.architecture |
| minimum-versions | B.1 | 02-02.versions |
| otel-collector-constraints | 9.4.1 | 02-03.observability |
| base-initialisms | 8.1.4 | 02-04.openapi |
| http-status-codes | 8.4 | 02-04.openapi |
| secrets-detection-strategy | 6.10 | 02-05.security |
| key-principles | 6.9 | 02-06.authn |
| session-token-formats | 6.9.1 | 02-06.authn |
| headless-authn | 6.9.2 | 02-06.authn |
| browser-authn | 6.9.3 | 02-06.authn |
| mfa-combinations | 6.9.4 | 02-06.authn |
| authz-methods | 6.9.5 | 02-06.authn |
| validator-error-aggregation | 13.5 | 03-01.coding |
| format-go-protection | 11.2.8 | 03-01.coding |
| test-file-suffixes | 10.1 | 03-02.testing |
| three-tier-database-strategy | 10.1 | 03-02.testing, 03-04.data-infrastructure |
| sequential-test-exemption | 10.2.5 | 03-02.testing |
| disable-keep-alives-test-transport | 10.3.4 | 03-02.testing |
| timeout-double-multiplication-antipattern | 10.3.4 | 03-02.testing |
| crypto-acronyms-caps | 11.1.3 | 03-03.golang |
| sqlite-barrier-outside-tx | 5.2.4 | 03-04.data-infrastructure |
| utf8-without-bom | 9.9.3 | 03-05.linting |
| docker-compose-rules | 12.3.1 | 04-01.deployment |
| cicd-command-naming | 9.10.2 | 04-01.deployment |
| cicd-lint-constraints | 9.10.6 | 04-01.deployment |
| docker-desktop-startup | 14.5.5 | 05-01.cross-platform |
| docker-desktop-upgrade | 14.5.5 | 05-01.cross-platform |
| scripting-language-policy | 14.9 | 05-01.cross-platform |
| conventional-commits | 14.2.1 | 05-02.git |
| incremental-commits | 14.2.2 | 05-02.git |
| restore-from-baseline | 14.2.3 | 05-02.git |
| agent-self-containment | 2.1.1 | 06-02.agent-format |
| agent-tool-discovery | 2.1.6 | 06-02.agent-format |

**Instruction file coverage**: All 18 instruction files analyzed. 17 files have 1+ propagation chunks (38 unique chunks, 41 total source-target pairs). 1 file (copilot-instructions) is structural glue only — its content is a condensed quick-reference summary and instruction file table, not verbatim ARCHITECTURE.md content.

**Structural glue** (~20% of instruction file content) remains non-propagated: condensed quick-reference summaries, section headings, `See` cross-references, transitional text, tables in different formats, and code examples unique to instruction file context.

### 13.5 Validator Error Aggregation Pattern

<!-- @propagate to=".github/instructions/03-01.coding.instructions.md" as="validator-error-aggregation" -->
All validators run to completion (never short-circuit) and aggregate errors for a single unified report. Sequential execution ensures deterministic output ordering. Aggregated errors (not fail-fast) show ALL problems in one run, reducing fix-test-fix cycles. `validate-all` returns exit code 0 if all pass, exit code 1 if any fail.
<!-- @/propagate -->

**Execution Model**: Sequential execution of all 8 validators. Each validator produces a `ValidationResult` containing: valid/invalid status, error list, and execution duration. The orchestrator (`ValidateAll`) collects all results and produces a summary with pass/fail counts and total duration.

**Rationale**: Sequential execution (not parallel) ensures deterministic output ordering and simplifies debugging. Aggregated errors (not fail-fast) show ALL problems in one run, reducing fix-test-fix cycles.

**Exit Code**: `validate-all` returns exit code 0 if all validators pass, exit code 1 if any validator fails. CI/CD workflows use this to block merges on validation failures.

#### 12.3.5 Canonical Docker Compose Service Command Pattern

**MANDATORY**: All 10 PS-ID Docker Compose app services (all 4 variants each) MUST use the following canonical command array structure. This pattern is enforced by the `compose-entrypoint-uniformity` fitness linter.

**Canonical Command Format**:

```yaml
command: ["server", "--bind-public-port=8080", "--config=/certs/tls-config.yml",
          "--config=/app/config/{PS-ID}-app-{variant}.yml",
          "--config=/app/config/{PS-ID}-app-common.yml",
          "--config=/app/otel/otel.yml",
          "-u", "{DATABASE_URL}"]
```

**Argument Order** (MUST NOT deviate):

1. `server` — service subcommand (ALWAYS first)
2. `--bind-public-port=8080` — public HTTPS port binding (ALWAYS second)
3. `--config=/certs/tls-config.yml` — TLS certificate configuration (ALWAYS third)
4. `--config=/app/config/{PS-ID}-app-{variant}.yml` — variant-specific config overlay
5. `--config=/app/config/{PS-ID}-app-common.yml` — common config overlay
6. `--config=/app/otel/otel.yml` — OpenTelemetry/OTLP configuration
7. `-u {DATABASE_URL}` — database connection URL (ALWAYS last)

**Database URL by Variant**:

| Variant | `-u` Value |
|---------|-----------|
| `sqlite-1` | `sqlite://file::memory:?cache=shared` |
| `sqlite-2` | `sqlite://file::memory:?cache=shared` |
| `postgresql-1` | `file:///run/secrets/postgres-url.secret` |
| `postgresql-2` | `file:///run/secrets/postgres-url.secret` |

**Dockerfile ENTRYPOINT**:
- Each PS-ID Dockerfile builds its own service binary: `go build ... -o /app/{PS-ID} ./cmd/{PS-ID}`
- ENTRYPOINT: `["/sbin/tini", "--", "/app/{PS-ID}"]`
- The `command:` array in compose.yml is appended to the ENTRYPOINT as arguments

**Rationale**:
- `server` subcommand is explicit (not relying on default empty-args behavior)
- `--bind-public-port=8080` before configs allows configs to override if needed
- TLS config always present (dual HTTPS servers require certificates)
- `-u` flag always last, always present (explicit database URL)
- Each service builds its own binary (not the suite binary) for minimal image size

### 12.4 Environment Strategy

**Development**: SQLite in-memory, port 0, auto-generated TLS, disabled telemetry
**Testing**: test-containers (PostgreSQL), dynamic ports, ephemeral instances
**Production**: PostgreSQL (cloud), static ports, full telemetry, TLS required

### 12.5 Release Management

**Versioning**: Semantic versioning (major.minor.patch)
**Release Process**: Tag creation, CHANGELOG generation, artifact publishing
**Rollback Strategy**: Previous version stable, blue-green deployment

---

## 14. Development Practices

### 14.1 Coding Standards

**Go Best Practices**: Effective Go, Code Review Comments, Go Proverbs
**Project Patterns**: See [03-01.coding.instructions.md](../.github/instructions/03-01.coding.instructions.md) for file size limits, default values, conditional statements

#### 14.1.1 Opportunistic Quality Fixes — MANDATORY

**CRITICAL: ALL linter violations, code quality issues, and pre-existing defects discovered during ANY work MUST be fixed immediately — even when not part of the original request, phase, task, or plan.**

This applies to ALL issue types including but not limited to:

- `goconst`: Repeated string literals must become constants
- `noctx`: Missing context in database/HTTP calls (`Ping()` → `PingContext(ctx)`)
- `lint-go literal-use`: Magic constants must use `cryptoutilSharedMagic` values
- `wsl`, `godot`, `gofumpt`: Formatting and style violations
- Import ordering and unused imports
- Pre-commit hook findings from any linter

**Rationale**: Quality is paramount. Deferring discovered issues creates technical debt that compounds. Each linter pass may discover new issues from different linters — fix ALL before re-staging. Incremental lint discovery is normal and expected.

**Anti-Pattern**: Tagging discovered issues as "pre-existing" or "not part of this task" to justify deferral. If an issue is discovered, it is blocking regardless of origin.

**Atomic Staging for Cross-Cutting Changes**: When a refactor touches imports across multiple packages AND renames/moves directories, ALL changes MUST be staged together. Pre-commit hooks run against the staged state, not the working directory — partial staging of cross-cutting changes will fail type-checking.

### 14.2 Version Control

#### 14.2.1 Conventional Commits

<!-- @propagate to=".github/instructions/05-02.git.instructions.md" as="conventional-commits" -->
**Format**: `<type>[optional scope]: <description>`

**Types**: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert

**Examples**:

```bash
feat(auth): add OAuth2 client credentials flow
fix(database): prevent connection pool exhaustion
feat(api)!: remove deprecated v1 endpoints  # Breaking change
```
<!-- @/propagate -->

#### 14.2.2 Incremental Commit Strategy

<!-- @propagate to=".github/instructions/05-02.git.instructions.md" as="incremental-commits" -->
- ALWAYS commit incrementally (NOT amend) - preserves history for bisect, selective revert.
- NEVER repeatedly amend - loses context, hard to bisect.
- Amend ONLY for immediate typo fixes (<1 min, before push).
- **Semantic Grouping**: Commit each semantically coherent unit of work as it completes. NEVER accumulate changes for different semantic groups into a bulk commit. Semantic boundaries: one feature, one bug fix, one refactor, one test suite, one doc update = each gets its own commit.
- **Periodic Commits**: Prefer frequent small commits over rare large commits. A completed task = a commit. Push every 5-10 commits.
<!-- @/propagate -->

#### 14.2.3 Restore from Clean Baseline Pattern

<!-- @propagate to=".github/instructions/05-02.git.instructions.md" as="restore-from-baseline" -->
**When fixing regressions, ALWAYS restore clean baseline FIRST**:

1. Find last known-good commit (`git log --oneline --grep="baseline"`)
2. Restore package (`git checkout <hash> -- path/to/package/`)
3. Verify baseline works (`go test`)
4. Apply ONLY the new fix (minimal change)
5. Commit as NEW commit (NOT amend)

**Why**: HEAD may be corrupted from previous failed attempts. Start from known-good state.
<!-- @/propagate -->

### 14.3 Branching Strategy

**Main branch**: Always stable, deployable, protected
**Feature branches**: Short-lived (<7 days), rebased on main before merge
**Release branches**: For production releases, cherry-pick hotfixes

### 14.4 Code Review

**Requirements**: 2+ approvals for core changes, 1 approval for docs/tests
**Checklist**: Tests added, docs updated, linting passes, security reviewed
**Size limits**: <500 lines ideal, >1000 lines requires breakdown

### 14.5 Development Workflow

#### 14.5.1 Spec Structure Patterns

- plan.md: Vision, phases, success criteria, anti-patterns
- tasks.md: Phase breakdown, task checklist, dependencies
- DETAILED.md: Session timeline with date-stamped entries
- Coverage tracking by package

#### 14.5.2 Terminal Command Auto-Approval

- Pattern checking against .vscode/settings.json
- Auto-enable: Read-only and build operations
- Auto-disable: Destructive operations
- autoapprove wrapper for loopback network commands

#### 14.5.3 Session Documentation Strategy

- MANDATORY: Append to DETAILED.md Section 2 timeline
- Format: `### YYYY-MM-DD: Title`
- NEVER create standalone session docs
- DELETE completed tasks immediately from todos-*.md

#### 14.5.4 Air Live Reload

Use `air` for live-reload development of individual services. Air watches Go source files and automatically rebuilds/restarts the service binary on changes.

**Prerequisites**: `go install github.com/air-verse/air@latest`

**Usage**: `SERVICE=<service-name> air`

- Example: `SERVICE=sm-im air` — starts the sm-im service with live reload
- Valid SERVICE values: `sm-im`, `sm-kms`, `jose-ja`, `pki-ca`, `skeleton-template`, `identity-authz`, `identity-idp`, `identity-rp`, `identity-rs`, `identity-spa`

**Configuration** (`.air.toml` at repo root):

- Binary output: `./tmp/main`; tmp dir: `./tmp/`
- Watch directories: `internal/`, `cmd/`, `pkg/`, `api/`
- Watch extensions: `.go`, `.tpl`, `.tmpl`, `.html`
- Excludes: `_test.go` files (test changes don't trigger rebuild)
- Service args: `server --dev` (enables dev mode)
- Graceful shutdown: SIGTERM with 500ms kill delay

**Anti-patterns**:

- NEVER use `air` for running tests — use `go test ./...`
- NEVER hard-code SERVICE in scripts — `SERVICE` env var is required

#### 14.5.5 Docker Desktop Startup - CRITICAL

<!-- @propagate to=".github/instructions/05-01.cross-platform.instructions.md" as="docker-desktop-startup" -->
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
<!-- @/propagate -->

See: [Section 11.2.5 CI/CD](#1125-cicd) for local workflow testing commands that require Docker.

<!-- @propagate to=".github/instructions/05-01.cross-platform.instructions.md" as="docker-desktop-upgrade" -->
**Docker Desktop Upgrade Warning**: After ANY Docker Desktop or testcontainers upgrade, run the full E2E test suite. Upgrades MAY break API compatibility between testcontainers-go and Docker Desktop — symptoms may include socket errors, container startup failures, and general Docker API issues.
<!-- @/propagate -->

See [Section 9.4.2 Docker Desktop and Testcontainers API Compatibility](#942-docker-desktop-and-testcontainers-api-compatibility) for diagnosis checklist and resolution guidance.

### 14.6 Plan Lifecycle Management

**Single Living Plan**: Each project MUST have exactly one active plan document (`plan.md`) and one active task list (`tasks.md`). Creating versioned successor plans (e.g., `plan-v2.md`, `fixes-v8/`) is an anti-pattern.

**Plan Lifecycle**:
- **Active**: Currently executing. Single `plan.md` + `tasks.md` in project directory.
- **Archived**: Completed or superseded. Move entire directory to `archive/` subdirectory.
- **NEVER**: Create parallel/successor plans. Update the existing plan instead.

**Anti-Patterns** (FORBIDDEN):
- Creating `vN+1` plan directories when `vN` has remaining work
- Mixing analysis prose with task checkboxes in the same file
- Task lists exceeding 300 lines (split into phases with separate files)
- Leaving archived plans in active directories without moving to `archive/`

**Task Document Rules**:
- `tasks.md`: Checkboxes ONLY (`- [ ]` / `- [x]`), grouped by phase
- `plan.md`: Strategy, architecture decisions, phase descriptions (NO checkboxes)
- Analysis results go in `research/` subdirectory, NOT in plan or task files

**Knowledge Propagation — Every Plan MUST**:
- Include a final "Knowledge Propagation" phase that updates ARCHITECTURE.md, agents, skills, and instructions with new patterns discovered
- Conduct post-mortems after EVERY phase to self-evaluate artifacts for contradictions or omissions
- Document all architectural decisions in plan.md before archiving the plan

### 14.7 Infrastructure Blocker Escalation

<!-- @propagate to=".github/instructions/06-01.evidence-based.instructions.md, .github/instructions/01-02.beast-mode.instructions.md" as="infrastructure-blocker-escalation" -->
**MANDATORY: ALL infrastructure issues are BLOCKING. NEVER defer, deprioritize, skip, or tag as "pre-existing."**

Three-encounter rule: 1st → document, 2nd → create fix task, 3rd → MANDATORY Phase 0 fix (block ALL other work). Infrastructure blockers (OTel, Docker, testcontainers, CI/CD) take priority over feature work.
<!-- @/propagate -->

**Three-Encounter Escalation Rule**:

| Encounter | Action | Example |
|-----------|--------|---------|
| 1st | Document in tracking doc with root cause hypothesis | "OTel collector fails: docker detector requires socket" |
| 2nd | Create dedicated fix task, assign to current phase | "Task 11.3: Fix OTel docker detector configuration" |
| 3rd | MANDATORY Phase 0 fix — block ALL other work until resolved | "Phase 0: OTel infrastructure fix (BLOCKING)" |

**Anti-Pattern**: Tagging infrastructure blockers as "pre-existing" or "not caused by our changes" to justify deferral. If an infrastructure issue blocks E2E tests, it blocks the ENTIRE project regardless of origin.

**Infrastructure Categories** (ALL are blocking):
- OTel/telemetry configuration errors
- Docker socket/daemon connectivity
- Docker Desktop version incompatibility (see [Section 9.4.2](#942-docker-desktop-and-testcontainers-api-compatibility))
- Testcontainers API version mismatch
- CI/CD workflow failures
- Database connectivity or migration errors
- TLS certificate generation failures

**Resolution Priority**: Infrastructure blockers take priority over feature work. A broken test infrastructure means ALL test results are unreliable.

### 14.8 Phase Post-Mortem & Knowledge Propagation

#### 14.8.1 Phase Post-Mortem — MANDATORY

At the end of EVERY phase's quality gates, conduct a post-mortem **before starting the next phase**:

1. **lessons.md** (in `<work-dir>/`): Record lessons learned — what worked, what didn't, root causes, patterns observed.
2. **Artifact Self-Evaluation**: Actively evaluate whether phase lessons expose contradictions or omissions in:
   - `docs/ARCHITECTURE.md` — architecture decisions, patterns, strategies
   - `.github/agents/*.agent.md` — agent guidance and workflows
   - `.github/skills/*/SKILL.md` — skill templates and guidance
   - `.github/instructions/*.instructions.md` — coding, testing, security guidelines
   - Production code — missed abstractions, incorrect patterns, technical debt
   - Tests — missing coverage, weak assertions, deprecated test patterns
   - CI/CD workflows — missing steps, incorrect gates, outdated tooling
   - Project documentation — README, docs/, comments that contradict new patterns
3. **Create Fix Tasks**: If contradictions or omissions are found, create new phase tasks to fix them — NEVER defer artifact updates.
4. **Identify new phases**: Create follow-up phases for any blockers, gaps, or artifact fixes discovered.

Skipping post-mortems is FORBIDDEN. This is continuous self-improvement.

#### 14.8.2 Plan Completion Knowledge Propagation — MANDATORY

After ALL plan tasks are complete, apply accumulated lessons to permanent artifacts:

1. **ARCHITECTURE.md**: Update with new patterns, strategies, and architectural decisions discovered.
2. **Agents**: Update `.github/agents/*.agent.md` with improved guidance and workflows.
3. **Skills**: Add or update `.github/skills/*/SKILL.md` to capture new patterns and templates.
4. **Instructions**: Update `.github/instructions/*.instructions.md` with new coding/testing patterns.
5. **Code**: Apply patterns discovered during the plan back to production code where appropriate.
6. **Tests**: Improve test suites where plan work exposed incomplete coverage or weak assertions.
7. **Workflows**: Update CI/CD workflows to reflect any new quality gates or tooling discovered.
8. **Documentation**: Update README, inline comments, and docs/ to reflect new patterns.
9. **Verify propagation**: Run `go run ./cmd/cicd-lint lint-docs validate-propagation` to ensure `@source` blocks are in sync with `@propagate` blocks.
10. Commit all artifact updates with separate semantic commits per artifact type.

**Every plan MUST include a final "Knowledge Propagation" phase** that executes these steps. This phase is NOT optional.

### 14.9 Scripting Language Policy — MANDATORY

<!-- @propagate to=".github/instructions/05-01.cross-platform.instructions.md" as="scripting-language-policy" -->
**MANDATORY: Choose scripting language in priority order. Lower-priority choices require justification.**

| Priority | Language | When to Use |
|----------|----------|-------------|
| 1 (primary) | **Go** | ALL tooling, automation, scripts — compiled, cross-platform, static binary |
| 2 (exception) | **Java** | Gatling load tests in `test/load/` ONLY |
| 3 (last resort) | **Python** | Quick one-offs where Go is not suitable; one file, no maintenance burden |
| ❌ (BANNED) | **Bash** | BANNED everywhere except Docker container init scripts (`docker-entrypoint-initdb.d/`) |
| ❌ (BANNED) | **PowerShell** | BANNED everywhere, no exceptions |

**Rationale**: Go and Java are compiled languages that produce cross-platform static binaries with proper type safety and testability. Python scripts tend to accumulate without lifecycle management. Bash/PowerShell are platform-specific and error-prone.

**Docker container init exception**: The official PostgreSQL Docker image's `docker-entrypoint-initdb.d/` mechanism runs shell scripts natively. Shell scripts in this specific directory are the only permitted Bash exception. Minimize logic in these scripts; prefer `.sql` files where possible.

**NO Python under `internal/apps/tools/cicd_lint/`**: The `cicd_lint` tool is pure Go. No Python scripts, generation helpers, or utility modules belong here. If a capability requires Python (rare), it belongs outside the Go module.
<!-- @/propagate -->

### 14.10 Archive and Dead Code Policy

**MANDATORY: Code is DELETED, not archived. Git history preserves everything.**

| Action | Correct | Incorrect |
|--------|---------|-----------|
| Remove unused code | `git rm file.go` | Move to `_archived/` or `archived/` |
| Remove legacy configs | `git rm config.yml` | Move to `configs/orphaned/` |
| Remove old docs | `git rm docs/OLD.md` | Rename to `docs/OLD.archived.md` |
| Remove dead packages | `git rm -r pkg/` | Move to `internal/_archived/pkg/` |

**Rules**:

1. **No archive directories**: `_archived/`, `archived/`, `orphaned/` directories MUST NOT exist anywhere in the repository. The `archive-detector` fitness linter enforces this.
2. **Git history is the archive**: Any deleted file is recoverable via `git log --diff-filter=D -- path/to/file` and `git show <hash>:path/to/file`.
3. **No "soft delete" patterns**: Do not comment out large blocks of code, wrap in `if false {}`, or use build tags to hide dead code.
4. **Satellite docs are merged, then deleted**: When consolidating documentation, merge unique content into ARCHITECTURE.md (the SSOT), then delete the satellite file. Do not keep both.

**Rationale**: Archive directories accumulate stale code that confuses search results, inflates repository size, creates false positives in linters, and misleads developers into thinking archived code is maintained. Git provides complete history for recovery when needed.

---

### 14.11 Claude Code Autonomous Execution

Claude Code supports three execution modes for cryptoutil development work.

**Mode 1: Beast-Mode Agent (Sustained autonomous work)**

Invoke `/claude-beast-mode` (or `@claude-beast-mode` in chat) for continuous autonomous execution without interruptions. NEVER-STOP behavior: commits after every completed task, enforces all quality gates, and continues until ALL tasks done. Agent file: `.claude/agents/beast-mode.md`.

Use for: large multi-step implementations, refactoring sessions, any task requiring sustained uninterrupted progress across many files.

**Mode 2: Implementation Execution Workflow (Plan-then-execute)**

Two-phase approach for complex, high-risk changes:

1. `/claude-implementation-planning <work-dir>` — research + create `<work-dir>/plan.md` and `<work-dir>/tasks.md` with phases, decisions, and quality gates
2. `/claude-implementation-execution <work-dir>` — execute ALL tasks in plan.md/tasks.md continuously, committing after each task

Use for: features requiring upfront design decisions, multi-phase implementations, or when quizme-style Q&A is needed before coding starts.

**Mode 3: Standard Chat (Interactive)**

Default conversational mode. Claude asks clarifying questions and waits for confirmation. Use for: single-file edits, Q&A, quick targeted fixes, code review.

**`.claude/settings.local.json` Configuration**

`.claude/settings.local.json` configures Claude Code workspace behavior. This file is tracked in git but contains machine-local paths where needed.

```json
{
  "permissions": {
    "additionalDirectories": ["/path/to/memory/dir"]
  }
}
```

Key settings:

| Setting | Purpose |
|---------|---------|
| `permissions.additionalDirectories` | Extra directories Claude can read/write (e.g., memory store) |
| `permissions.allow` | Tool patterns to allow without prompting (e.g., `"Bash(go test*)"`) |
| `permissions.deny` | Tool patterns to deny unconditionally |

**CLAUDE.md** (`.claude/agents/`, `.claude/commands/`) registers all Claude Code agents and slash commands. Update CLAUDE.md when adding new agents or commands.

**Enforcement**

All autonomous execution modes enforce the same quality gates as Section 11.2 and the same commit discipline as Section 14.2. Beast-mode and implementation-execution agents are held to identical standards as interactive chat — the difference is only in interruption behavior, not in quality requirements.

---

## 15. Operational Excellence

### 15.1 Monitoring & Alerting

**Metrics**: Prometheus (HTTP, DB, crypto, keys)
**Logging**: Structured logs via OpenTelemetry
**Tracing**: Distributed traces via OTLP
**Dashboards**: Grafana LGTM (Loki, Tempo, Prometheus)

### 15.2 Incident Management

**Post-Mortem Template**: docs/P0.X-INCIDENT_NAME.md - Summary, root cause, timeline, impact, lessons, action items
**Severity Levels**: P0 (critical), P1 (high), P2 (medium), P3 (low)
**Response Time**: P0 <15min, P1 <1hr, P2 <1 day, P3 <1 week

### 15.3 Performance Management

**Benchmarks**: Crypto operations, HTTP endpoints, database queries
**Load Testing**: Gatling scenarios (baseline, peak, stress)
**Optimization**: Profile hot paths, caching strategies, connection pooling

### 15.4 Capacity Planning

**Resource Limits**: Memory, CPU, disk, network
**Scaling Triggers**: >70% utilization sustained >5min
**Horizontal Scaling**: Stateless services, PostgreSQL read replicas (future)

### 15.5 Disaster Recovery

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

**Go 1.26.1**: Static typing, fast compilation, excellent concurrency, CGO-free (portability)
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

### B.1 Technology Stack

<!-- @propagate to=".github/instructions/02-02.versions.instructions.md" as="minimum-versions" -->
**CRITICAL: ALWAYS use the same version everywhere** (dev, CI/CD, Docker, workflows, docs)

- Go: 1.26.1+
- Python: 3.14+
- golangci-lint: v2.7.2+
- Node: v24.11.1+ LTS
- Java: 21 LTS (Gatling load tests)
- Maven: 3.9+
- pre-commit: 2.20.0+
- Docker: 24+
- Docker Compose: v2+
<!-- @/propagate -->

**Languages**: Go 1.26.1 (services), Python 3.14+ (utilities), Node v24.11.1+ (CLI tools)
**Databases**: PostgreSQL 18, SQLite (modernc.org/sqlite, CGO-free)
**Frameworks**: Fiber (HTTP), GORM (ORM), oapi-codegen (OpenAPI)
**Observability**: OpenTelemetry, Grafana LGTM (Loki, Tempo, Prometheus)
**Security**: FIPS 140-3 approved algorithms, Docker/Kubernetes secrets
**Testing**: testify, gremlins (mutation), Nuclei/ZAP (DAST), Gatling (load)

### B.2 Dependency Matrix

**Core Dependencies**:

- github.com/gofiber/fiber/v3 (HTTP framework)
- gorm.io/gorm (ORM)
- github.com/google/uuid/v7 (UUIDv7)
- go.opentelemetry.io/otel (telemetry)
- github.com/go-jose/go-jose/v4 (JOSE)

**Test Dependencies**: testify, testcontainers-go, httptest

### B.3 Configuration Reference

**Priority Order**: Docker secrets > YAML > CLI parameters (NO env vars for secrets)

**Standard Files**:

- config.yml: Main configuration
- secrets/*.secret: Credentials (chmod 440)

### B.4 Instruction File Reference

**See .github/copilot-instructions.md** for complete table of 18 instruction files

**Summary**: 01-terminology/beast-mode, 02-architecture (5 files), 03-development (4 files), 04-deployment (1 file), 05-platform (2 files), 06-evidence (2 files)

### B.5 Agent Catalog & Handoff Matrix

| Copilot Agent | Claude Code Agent | Description | Handoffs |
|--------------|------------------|-------------|----------|
| `copilot-implementation-planning` | `claude-implementation-planning` | Planning and task decomposition | → implementation-execution |
| `copilot-implementation-execution` | `claude-implementation-execution` | Autonomous implementation execution | → fix-workflows |
| `copilot-fix-workflows` | `claude-fix-workflows` | Workflow repair and validation | None defined |
| `copilot-beast-mode` | `claude-beast-mode` | Continuous execution mode | None defined |

See `.github/agents/*.agent.md` `tools:` frontmatter for the authoritative Copilot per-agent tool list. The parallel `.claude/agents/*.md` files omit `tools:` — Claude Code inherits all tools by default. `cicd-lint lint-docs` (`lint-agent-drift`) enforces that description, argument-hint, and body are verbatim identical across each pair.

### B.6 CI/CD Workflow Catalog

| Workflow | Purpose | Dependencies | Duration | Timeout |
|----------|---------|--------------|----------|---------|
| ci-coverage | Test coverage collection, enforce ≥95%/98% | None | 5-6min | 20min |
| ci-mutation | Mutation testing with gremlins | None | 15-20min | 45min |
| ci-race | Race condition detection | None | 10-15min | 20min |
| ci-benchmark | Performance benchmarking | None | 8-15min | 30min |
| ci-quality | Linting and code quality | None | 3-5min | 15min |
| ci-sast | Static security analysis | None | 5-10min | 20min |
| ci-dast | Dynamic security testing | PostgreSQL | 10-20min | 30min |
| ci-e2e | End-to-end integration tests | Docker Compose | 20-40min | 60min |
| ci-load | Load testing with Gatling | Docker Compose | 15-30min | 45min |
| ci-gitleaks | Secret detection | None | 2-3min | 10min |
| release | Automated release workflows | ci-* passing | 5-10min | 30min |

### B.7 Reusable Action Catalog

| Action | Description | Inputs | Outputs |
|--------|-------------|--------|---------|
| docker-images-pull | Parallel Docker image pre-fetching | images (newline-separated list) | None |

See `.github/actions/` for the authoritative action catalog.

### B.8 Linter Rule Reference

| Linter | Purpose | Enabled | Auto-Fix | Exclusions |
|--------|---------|---------|----------|------------|
| errcheck | Unchecked errors | ✅ | ❌ | Test helpers |
| govet | Suspicious code | ✅ | ❌ | None |
| staticcheck | Static analysis | ✅ | ❌ | Generated code |
| wsl_v5 | Whitespace linting | ✅ | ✅ | None |
| godot | Comment periods | ✅ | ✅ | None |
| gosec | Security issues | ✅ | ❌ | Justified cases |

See `.golangci.yml` for the authoritative linter configuration with all 30+ active linters.

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

**Related Documents**:
- `.github/copilot-instructions.md` - Copilot configuration
- `.github/instructions/*.instructions.md` - Detailed instructions

**Cross-References**:
- All sections maintain stable anchor links for referencing
- Consistent section numbering for navigation
