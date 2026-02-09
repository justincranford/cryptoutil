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

### 3.3 Product-Service Relationships

[To be populated]

### 3.4 Port Assignments & Networking

[To be populated]

---

## 4. System Architecture

### 4.1 System Context

[To be populated]

### 4.2 Container Architecture

[To be populated]

### 4.3 Component Architecture

[To be populated]

### 4.4 Code Organization

[To be populated]

---

## 5. Service Architecture

### 5.1 Service Template Pattern

[To be populated]

### 5.2 Service Builder Pattern

[To be populated]

### 5.3 Dual HTTPS Endpoint Pattern

[To be populated]

### 5.4 Dual API Path Pattern

[To be populated]

### 5.5 Health Check Patterns

[To be populated]

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

---

## 7. Data Architecture

### 7.1 Multi-Tenancy Architecture & Strategy

[To be populated]

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

### 8.2 REST Conventions

[To be populated]

### 8.3 API Versioning

[To be populated]

### 8.4 Error Handling

[To be populated]

### 8.5 API Security

[To be populated]

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

### 11.2 Quality Gates

[To be populated]

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
