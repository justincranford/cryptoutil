# Feature Template Usage Example

**Purpose**: Demonstrates how to use the feature template for a realistic feature implementation
**Example Feature**: Identity OAuth 2.1 Authorization Server
**Complexity**: Medium-High (demonstrates most template sections)

---

## How This Example Maps to Template

This document shows a **filled-out version** of the feature template for a real feature. Compare this to `feature-template.md` to understand:

1. **What to fill in**: Each section populated with realistic content
2. **What to remove**: Sections marked "N/A - Not applicable" when not needed
3. **What to customize**: Domain-specific additions not in base template
4. **What to scale**: How granularity changes based on complexity

---

## Executive Summary

### Feature Overview

**Feature Name**: OAuth 2.1 Authorization Server
**Feature ID**: IDENTITY-AUTHZ-V1
**Status**: IN_PROGRESS
**Priority**: 🔴 CRITICAL

### Current Reality

**Problem Statement**:
The identity module lacks a fully compliant OAuth 2.1 authorization server, preventing secure API access delegation and third-party integrations. Current authentication is limited to basic username/password with no token-based delegation.

**Current State Analysis**:

- Existing: Basic user authentication, session management
- Missing: OAuth 2.1 flows (authorization code, client credentials, device flow)
- Impact: Cannot support SPA applications, mobile apps, or third-party integrations
- Technical Debt: Authentication coupled tightly to session cookies, not stateless

**Production Blockers**:

1. 🔴 Authorization code flow non-functional (16 TODO comments)
2. 🔴 PKCE validation missing (security vulnerability for public clients)
3. 🔴 Token lifecycle management incomplete (no cleanup, tokens persist indefinitely)

### Completion Metrics

| Metric | Count | Percentage | Status |
|--------|-------|------------|--------|
| **Fully Complete** | 3/8 | 37.5% | ⚠️ |
| **Documented Complete but Has Gaps** | 2/8 | 25% | ⚠️ |
| **Incomplete/Not Started** | 3/8 | 37.5% | ❌ |
| **Total Tasks** | 8 | 100% | - |

### Production Readiness Assessment

**Production Ready**: ❌ NO

**Rationale**: Authorization code flow is the foundational OAuth 2.1 flow required for web/mobile applications. Without it, no secure API access delegation is possible.

### Remediation Approach

**Strategy**: Foundation First - Complete core OAuth 2.1 flows before advanced features (device flow, PKCE, mTLS)

**Timeline**: 12 days (assumes full-time focus, 8 hours/day)

**Effort Distribution**:

- Foundation: 25% (3 days) - Domain models, database schema
- Core Features: 50% (6 days) - Authorization code flow, token endpoints
- Advanced Features: 15% (2 days) - Client authentication methods
- Integration & Testing: 10% (1 day) - E2E test suite

---

## Goals and Objectives

### Primary Goals

**Goal 1**: Implement OAuth 2.1 Authorization Code Flow

- **Description**: Enable web applications to obtain access tokens on behalf of users via authorization code flow with PKCE
- **Success Criteria**: Full authorization code flow functional (authorize → login → consent → code → token)
- **Priority**: CRITICAL
- **Dependencies**: Domain models (Task 01), Storage layer (Task 02)
- **Risk Level**: MEDIUM (well-defined specification, but integration complexity)

**Goal 2**: Implement Token Lifecycle Management

- **Description**: Issue, validate, refresh, and revoke access/refresh tokens with proper expiration
- **Success Criteria**: Tokens expire correctly, refresh flow works, revocation immediate
- **Priority**: HIGH
- **Dependencies**: Token repository (Task 02), JWT operations (Task 03)
- **Risk Level**: LOW (standard patterns)

**Goal 3**: Support Multiple Client Authentication Methods

- **Description**: client_secret_basic, client_secret_post, private_key_jwt
- **Success Criteria**: Each method validates client credentials correctly
- **Priority**: HIGH
- **Dependencies**: Client repository (Task 02), Crypto operations (existing)
- **Risk Level**: MEDIUM (private_key_jwt requires JWT validation)

### Secondary Goals (Nice-to-Have)

**Goal 4**: Device Authorization Flow (RFC 8628)

- **Description**: Enable devices without browsers (smart TVs, CLI tools) to obtain tokens
- **Success Criteria**: Device flow functional end-to-end
- **Priority**: LOW
- **Dependencies**: Primary goals 1-3 complete

**Goal 5**: Mutual TLS Client Authentication

- **Description**: Support certificate-based client authentication for high-security scenarios
- **Success Criteria**: mTLS client authentication validates certificates
- **Priority**: LOW
- **Dependencies**: Certificate infrastructure, Primary goals 1-3

### Non-Goals (Out of Scope)

- **Non-Goal 1**: OpenID Connect (deferred to separate feature - IDENTITY-OIDC-V1)
- **Non-Goal 2**: OAuth 1.0 compatibility (deprecated standard, not implementing)
- **Non-Goal 3**: Social login integrations (Google, Facebook, GitHub - deferred to Task 15)

### Constraints and Boundaries

**Technical Constraints**:

- Use ONLY existing go.mod dependencies (no new OAuth libraries)
- Use lestrrat-go/jwx/v3 for JWT operations (already in use)
- GORM for database access (already in use)
- Fiber for HTTP endpoints (already in use)

**Architectural Constraints**:

- Authorization server CANNOT import Identity Provider packages (domain isolation)
- Must use `internal/identity/authz/` namespace
- Import alias: `cryptoutilIdentityAuthz` (defined in .golangci.yml)
- Magic values in `internal/identity/authz/magic*.go`

**Security Constraints**:

- FIPS 140-3 compliance MANDATORY (no bcrypt, use PBKDF2-HMAC-SHA256 for secrets)
- PKCE MANDATORY for all public clients
- Authorization code single-use enforcement
- State parameter validation MANDATORY

**Operational Constraints**:

- Backward compatible with existing session-based authentication
- Support both PostgreSQL (production) and SQLite (dev/test)
- Observability: OTLP traces for all flows, metrics for token issuance rates

---

## Context and Baseline

### Historical Context

**Previous Attempts**:

- Attempt 1 (2024 Q3): Started OAuth implementation but abandoned due to complexity
  - Why it failed: Tried to build custom OAuth library instead of using established patterns
  - Lessons learned: Use proven patterns from RFC specs, don't reinvent the wheel

**Related Work**:

- IDENTITY-IDP-V1: Identity Provider (separate feature, provides login/consent UI)
- IDENTITY-RS-V1: Resource Server (separate feature, validates tokens)

**Evolution Timeline**:

- Phase 1 (2024 Q1-Q2): Basic username/password authentication with sessions
- Phase 2 (2024 Q3): Attempted OAuth 2.0 implementation (abandoned)
- Current State (2025 Q4): Restarting with OAuth 2.1 spec, cleaner architecture

### Baseline Assessment

**Current Implementation Status**:

- ✅ Complete: User domain models, session management, database repositories
- ⚠️ Partial: Client domain models exist but missing OAuth-specific fields
- ❌ Missing: Authorization request storage, PKCE validation, token endpoints

**Code Analysis**:

- Total files: 15 files (5 domain models, 5 repositories, 5 handlers)
- Total lines: 2,400 lines (1,200 production, 800 test, 400 docs)
- Test coverage: 72% overall (95% domain, 65% repositories, 55% handlers)
- TODO count: 16 critical (authorization flow), 4 high (token lifecycle), 8 medium (client auth)

**Dependency Analysis**:

- External: lestrrat-go/jwx/v3 (JWT), gorm.io/gorm (ORM), gofiber/fiber/v2 (HTTP)
- Internal: internal/common/crypto (key operations), internal/common/util (helpers)
- Circular: NONE (clean dependency graph)
- Coupling: LOW (well-isolated authz package)

**Technical Debt Assessment**:

- Architecture debt: Authorization code stored in-memory (not persistent)
- Code quality debt: 12 linting issues (wsl spacing, godot periods)
- Test debt: Missing E2E tests for full authorization flow
- Documentation debt: OpenAPI spec incomplete (missing PKCE parameters)

### Stakeholder Analysis

**Primary Stakeholders**:

- Application Developers: Need OAuth 2.1 for secure API access from SPAs/mobile apps
- Security Team: Require PKCE, state validation, short-lived tokens
- Operations Team: Need monitoring, token cleanup jobs, incident response runbooks

**User Impact**:

- End Users: Better experience with SSO, less frequent re-authentication
- Third-Party Developers: Can build integrations using standard OAuth 2.1

---

## Architecture and Design

### System Architecture

**High-Level Architecture**:

```
┌─────────────┐       ┌──────────────────┐       ┌─────────────┐
│   Client    │──────▶│ Authorization    │──────▶│  Identity   │
│  (SPA/App)  │       │     Server       │       │  Provider   │
└─────────────┘       │   (This Feature) │       │  (Separate) │
                      └──────────────────┘       └─────────────┘
                              │
                              ▼
                      ┌──────────────────┐
                      │    Database      │
                      │ (clients, tokens,│
                      │  auth requests)  │
                      └──────────────────┘
```

**Component Breakdown**:

- Authorization Handler: Receives authorization requests, validates parameters, redirects to IdP
- Token Handler: Exchanges authorization codes for tokens, validates PKCE
- Client Authenticator: Validates client credentials (basic, post, jwt)
- Token Repository: Stores/retrieves access/refresh tokens
- Auth Request Repository: Stores authorization requests with PKCE challenges

**Data Flow**:

```
Client → /authorize (with PKCE challenge) → AuthZ validates → Store request → Redirect to IdP login
      ← 302 redirect to IdP

User → IdP login → IdP consent → Generate auth code → Redirect to client callback
    ← 302 redirect with code

Client → /token (with code + PKCE verifier) → AuthZ validates PKCE → Issue tokens → Return JWT
      ← 200 OK with access_token, refresh_token
```

**Technology Stack**:

- Language: Go 1.25.4+
- HTTP Framework: Fiber v2
- Database: GORM with PostgreSQL/SQLite
- JWT: lestrrat-go/jwx/v3
- Observability: OpenTelemetry (OTLP)

### Design Patterns

**Pattern 1**: Repository Pattern

- **Use Case**: Data access abstraction for clients, tokens, authorization requests
- **Implementation**: Interface in `repository/repository.go`, GORM impl in `repository/gorm/`
- **Benefits**: Database-agnostic, testable with mocks, swap PostgreSQL ↔ SQLite
- **Trade-offs**: Extra abstraction layer, but well worth it for testing

**Pattern 2**: Strategy Pattern (Client Authentication)

- **Use Case**: Multiple client authentication methods (basic, post, jwt, mtls)
- **Implementation**: `ClientAuthenticator` interface with method-specific implementations
- **Benefits**: Easily add new auth methods, test each independently
- **Trade-offs**: More files, but cleaner than if/else chains

**Pattern 3**: Factory Pattern (Repository Creation)

- **Use Case**: Create repositories with proper database connection
- **Implementation**: `NewRepositoryFactory(db *gorm.DB)` returns configured repositories
- **Benefits**: Consistent initialization, dependency injection
- **Trade-offs**: One extra layer, but simplifies setup

### Directory Structure

**Target Directory Layout**:

```
internal/identity/authz/
├── handler/                      # HTTP handlers
│   ├── authorize.go             # GET/POST /authorize
│   ├── token.go                 # POST /token
│   ├── introspect.go            # POST /introspect
│   └── revoke.go                # POST /revoke
├── clientauth/                   # Client authentication
│   ├── authenticator.go         # Interface
│   ├── basic.go                 # client_secret_basic
│   ├── post.go                  # client_secret_post
│   ├── jwt.go                   # private_key_jwt
│   └── mtls.go                  # tls_client_auth (future)
├── pkce/                         # PKCE validation
│   └── pkce.go                  # S256 challenge/verifier
├── repository/                   # Data access
│   ├── repository.go            # Interfaces
│   └── gorm/                    # GORM implementations
│       ├── client_repository.go
│       ├── token_repository.go
│       └── auth_request_repository.go
├── domain/                       # Domain models
│   ├── client.go                # OAuth client
│   ├── token.go                 # Access/refresh tokens
│   └── auth_request.go          # Authorization requests
├── service/                      # Business logic
│   └── authz_service.go         # Core authorization logic
├── config/                       # Configuration
│   └── config.go                # AuthZ server config
└── magic/                        # Constants
    └── magic_oauth.go           # OAuth grant types, scopes, etc.
```

**No Migration** (new feature, not refactoring existing code)

### API Design

**Endpoint Structure**:

- `GET /oauth2/v1/authorize` - Initiate authorization code flow
- `POST /oauth2/v1/authorize` - Submit authorization decision (from IdP)
- `POST /oauth2/v1/token` - Exchange code for tokens
- `POST /oauth2/v1/introspect` - Validate token
- `POST /oauth2/v1/revoke` - Revoke token

**OpenAPI Specification**:

- Location: `api/identity/authz/openapi.yaml`
- Generation: `oapi-codegen` with strict server pattern
- Validation: Request/response validation middleware

**Authentication/Authorization**:

- Authorize endpoint: Session-based (user must be logged in via IdP)
- Token endpoint: Client authentication (basic, post, jwt)
- Introspect/Revoke: Bearer token (resource server validates tokens)

### Database Schema

**Entity Relationship Diagram**:

```
┌─────────────────┐       ┌──────────────────────┐
│     Client      │───────│   AuthRequest        │
├─────────────────┤       ├──────────────────────┤
│ id (PK)         │       │ id (PK)              │
│ client_id       │       │ client_id (FK)       │
│ client_secret   │       │ user_id (FK)         │
│ auth_methods    │       │ pkce_challenge       │
│ redirect_uris   │       │ state                │
└─────────────────┘       │ created_at           │
                          └──────────────────────┘
                                    │
                                    ▼
                          ┌──────────────────────┐
                          │      Token           │
                          ├──────────────────────┤
                          │ id (PK)              │
                          │ client_id (FK)       │
                          │ user_id (FK)         │
                          │ token_type           │
                          │ expires_at           │
                          └──────────────────────┘
```

**Schema Evolution Strategy**:

- Migration tool: GORM AutoMigrate for development, golang-migrate for production
- Versioning: `0001_initial_schema.up.sql`
- Rollback: Down migrations for each change
- Zero-downtime: N/A (new tables, no existing data)

**Cross-Database Compatibility**:

- PostgreSQL: Production database (UUID native type)
- SQLite: Development/testing (UUID as TEXT)
- Type mapping: `gorm:"type:text"` for UUIDs (works on both)
- JSON fields: `gorm:"serializer:json"` (SQLite doesn't have native JSON)
- Connection pool: MaxOpenConns=5 for SQLite (GORM + transactions), unlimited for PostgreSQL

### Security Design

**Threat Model**:

- Threat 1: Authorization code interception → Mitigate with PKCE (S256 required)
- Threat 2: Client impersonation → Mitigate with client authentication (secret/jwt/mtls)
- Threat 3: Token theft → Mitigate with short-lived tokens, refresh rotation

**Security Controls**:

- PKCE: Mandatory for public clients, optional for confidential (but recommended)
- State parameter: Validated to prevent CSRF attacks
- Redirect URI: Exact match required (no substring matching)
- Client secrets: Hashed with PBKDF2-HMAC-SHA256 (FIPS 140-3 compliant)
- Token signing: RS256, ES256 (FIPS 140-3 approved algorithms)

**Compliance Requirements**:

- FIPS 140-3: All cryptographic operations use approved algorithms
- OAuth 2.1: Follow draft specification (includes PKCE, deprecates implicit flow)
- Security best practices: Short-lived access tokens (15 min), refresh tokens rotated

---

## Implementation Tasks

### Task Organization

**Task Numbering Convention**:

- Primary tasks: `01-foundation.md` through `08-integration.md`
- No sub-tasks in this feature (tasks atomic enough)

**Task Categories**:

- **Foundation**: Core infrastructure (Tasks 01-02)
- **Core Features**: OAuth flows (Tasks 03-05)
- **Advanced Features**: Client auth methods (Tasks 06-07)
- **Integration**: Testing (Task 08)

### Implementation Tasks Table

| Task | File | Status | Priority | Effort | Dependencies | Risk | Description |
|------|------|--------|----------|--------|--------------|------|-------------|
| 01 | `01-domain-models.md` | ✅ | 🔴 CRITICAL | 4 hours | None | LOW | Domain models for Client, Token, AuthRequest |
| 02 | `02-database-repositories.md` | ✅ | 🔴 CRITICAL | 6 hours | 01 | MEDIUM | GORM repositories, migrations |
| 03 | `03-authorization-flow.md` | ⚠️ | 🔴 CRITICAL | 12 hours | 01, 02 | HIGH | Authorization code flow with PKCE |
| 04 | `04-token-endpoints.md` | ⚠️ | 🔴 CRITICAL | 8 hours | 03 | MEDIUM | Token issuance, refresh, revoke |
| 05 | `05-introspection.md` | ❌ | ⚠️ HIGH | 4 hours | 04 | LOW | Token introspection endpoint |
| 06 | `06-client-auth-basic.md` | ❌ | ⚠️ HIGH | 6 hours | 02 | LOW | client_secret_basic, client_secret_post |
| 07 | `07-client-auth-jwt.md` | ❌ | 🟡 MEDIUM | 8 hours | 06 | MEDIUM | private_key_jwt authentication |
| 08 | `08-e2e-tests.md` | ❌ | ⚠️ HIGH | 8 hours | 03, 04, 05 | LOW | End-to-end test suite |

**Status Legend**: ✅ COMPLETE | ⚠️ PARTIAL | 🔄 IN PROGRESS | 📋 PLANNED | ❌ BLOCKED | 🗄️ ARCHIVED

### Task Dependencies Graph

**Critical Path**:

```
01 → 02 → 03 → 04 → 05 → 08
         ↓
         06 → 07
```

**Parallel Execution**:

- Week 1: Tasks 01-02 (sequential foundation)
- Week 2: Task 03 (authorization flow) + Task 06 (client auth basic) - parallel
- Week 3: Task 04 (token endpoints) + Task 07 (client auth jwt) - parallel
- Week 4: Tasks 05, 08 (sequential wrap-up)

### Implementation Phases

**Phase 1: Foundation** (Days 1-2)

- Focus: Domain models, database schema, repositories
- Tasks: 01, 02
- Deliverables: Working database with CRUD operations
- Exit Criteria: All repository tests passing, migrations applied

**Phase 2: Authorization Flow** (Days 3-6)

- Focus: Authorization code flow with PKCE
- Tasks: 03, 06 (parallel)
- Deliverables: Functional /authorize and /token endpoints
- Exit Criteria: E2E authorization flow works (manual testing)

**Phase 3: Token Management** (Days 7-9)

- Focus: Token refresh, revocation, introspection
- Tasks: 04, 05, 07 (parallel)
- Deliverables: Complete token lifecycle
- Exit Criteria: All token operations functional

**Phase 4: Integration Testing** (Days 10-12)

- Focus: Automated E2E test suite
- Tasks: 08
- Deliverables: Comprehensive test coverage
- Exit Criteria: 95%+ coverage, all tests passing, ≥80% mutation score

---

## Task Execution Instructions

### LLM Agent Directives

**Follow continuous work pattern from template** (see feature-template.md "Task Execution Instructions")

**Key directives for this feature**:

1. NEVER stop between tasks - commit → immediately start next task
2. Work until 950k tokens used (95% of 1M budget)
3. Use `runTests` tool exclusively (NEVER `go test` in terminal)
4. Commit with conventional messages: `feat(authz): complete task 01 - domain models`
5. Create post-mortem after each task: `01-domain-models-POSTMORTEM.md`

### Task-Specific Notes

**Task 03 (Authorization Flow)**:

- CRITICAL: PKCE S256 mandatory for public clients
- Store authorization requests with expiration (5 minutes)
- Single-use authorization codes (delete after exchange)
- Validate redirect_uri exact match

**Task 04 (Token Endpoints)**:

- Access tokens: 15 minute expiration (short-lived)
- Refresh tokens: 30 day expiration, rotate on use
- Use RS256 for JWS (FIPS 140-3 compliant)
- Include OTLP tracing for all token operations

**Task 07 (Client Auth JWT)**:

- Validate JWT signature using client's registered public key
- Check `aud` claim matches token endpoint
- Enforce JWT expiration (max 5 minutes)
- Use lestrrat-go/jwx/v3 for validation

---

## Post-Mortem and Corrective Actions

**See feature-template.md for complete post-mortem structure**

**Example corrective action from Task 03**:

### Task 03: Authorization Flow - Corrective Actions

**Issue**: PKCE validation was implemented incorrectly - used MD5 instead of SHA-256 for S256 challenge method

**Root Cause**: Misread RFC 7636 specification, confused hash algorithm naming

**Immediate Fix**:

- Replaced MD5 with SHA-256 in `pkce/pkce.go`
- Added test vectors from RFC 7636 Appendix B
- Verified against reference implementation

**Deferred Actions**:

- Task 08.1: Add PKCE compliance test suite with RFC test vectors
- Task 09: Security audit of all crypto operations (ensure FIPS 140-3 compliance)

**Pattern Improvement**:

- Created checklist for crypto implementations: verify algorithm, test with spec examples, cross-reference FIPS 140-3 approved list

---

## Quality Gates and Acceptance Criteria

**Use universal acceptance criteria from template** (see feature-template.md "Quality Gates and Acceptance Criteria")

**Feature-specific acceptance criteria**:

### OAuth 2.1 Compliance

- [ ] Authorization code flow matches RFC 6749 Section 4.1
- [ ] PKCE implementation matches RFC 7636
- [ ] Token endpoint matches RFC 6749 Section 3.2
- [ ] Introspection matches RFC 7662
- [ ] Revocation matches RFC 7009

### Security Requirements

- [ ] PKCE mandatory for public clients
- [ ] Client secrets hashed with PBKDF2-HMAC-SHA256 (FIPS 140-3)
- [ ] JWT signatures use RS256 or ES256 (FIPS 140-3)
- [ ] State parameter validated (CSRF protection)
- [ ] Redirect URI exact match enforced

### Performance Requirements

- [ ] /authorize endpoint: < 50ms P95
- [ ] /token endpoint: < 100ms P95 (includes JWT signing)
- [ ] /introspect endpoint: < 50ms P95
- [ ] Token cleanup job: < 1 second for 10k expired tokens

### Operational Requirements

- [ ] OTLP traces for all OAuth flows
- [ ] Metrics: token_issued_total, token_revoked_total, authorization_code_generated_total
- [ ] Runbook: "OAuth Authorization Flow Troubleshooting"
- [ ] Rollback procedure: Disable new endpoints via feature flag

---

## Risk Management

### Risk Assessment Matrix

| Risk ID | Description | Probability | Impact | Severity | Mitigation | Owner | Status |
|---------|-------------|-------------|--------|----------|------------|-------|--------|
| R01 | PKCE implementation bug | MEDIUM | HIGH | HIGH | RFC test vectors, security review | Dev Team | ⚠️ ACTIVE |
| R02 | Token database contention | LOW | MEDIUM | MEDIUM | Connection pooling, indexing | Ops Team | 🟢 MITIGATED |
| R03 | Breaking change to client API | LOW | HIGH | MEDIUM | Versioned API, deprecation period | Product | 🟢 MITIGATED |

### Risk Monitoring

- Daily: R01 (PKCE implementation) during Task 03 development
- Weekly: R02 (database performance) after Task 02 complete
- Monthly: R03 (API compatibility) during beta period

---

## Success Metrics

### Completion Metrics

- Target: 8/8 tasks complete by end of month
- Current: 2/8 complete (25%)
- Trend: +1 task/week (on track)

### Performance Metrics

- Token issuance: Target 100 tokens/sec, Current N/A (not launched)
- P95 latency: Target <100ms, Current N/A (not launched)
- Error rate: Target <0.1%, Current N/A (not launched)

### Quality Metrics

- Code coverage: Target ≥95%, Current 92% (Tasks 01-02)
- Mutation score: Target ≥80%, Current 75% (Tasks 01-02)
- Linting issues: Target 0, Current 0
- TODO comments: Target 0, Current 16 (in Task 03)

---

## Template Comparison

### What Was Removed from Template

- **Non-Goals Section**: Simplified (only 3 non-goals, template has placeholder for many)
- **Historical Context**: Condensed (only 1 previous attempt vs. template's multiple phases)
- **Appendix B (References)**: Removed (not needed for this example)

### What Was Added Beyond Template

- **PKCE-specific security notes**: Not in template, domain-specific
- **FIPS 140-3 compliance checks**: Not in template, project-specific
- **Cross-database compatibility details**: Expanded from template (SQLite vs. PostgreSQL specifics)

### What Was Customized

- **Task granularity**: 8 tasks (vs. template's 20 task example) - scaled to feature complexity
- **Risk assessment**: 3 risks (vs. template's extensive matrix) - appropriate for medium feature
- **Success metrics**: Focused on OAuth-specific KPIs (token issuance rate) vs. generic metrics

---

## Key Takeaways for Template Users

1. **Scale appropriately**: This feature used 8 tasks vs. 20 in template - adjust to complexity
2. **Remove unused sections**: Removed appendix sections not needed for this feature
3. **Add domain-specific content**: Added PKCE, FIPS 140-3 details not in base template
4. **Customize acceptance criteria**: OAuth 2.1 compliance checks specific to this domain
5. **Simplify where possible**: Condensed historical context, non-goals, risk matrix

**Template is a STARTING POINT** - customize based on feature needs, remove what doesn't apply, add domain-specific sections.
