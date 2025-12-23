# Constitution Structure Proposals

**Document Purpose**: Five alternative organizational structures for `.specify/memory/constitution.md`

**Created**: 2025-12-22
**Current Constitution Version**: 3.0.0 (1257 lines)
**Context**: Proposing reorganization strategies to improve clarity, maintainability, and navigation

---

## Current Structure Analysis

### Existing Organization (.specify/memory/constitution.md v3.0.0)

**Section Hierarchy**:

```
I. Product Delivery Requirements
   - Four Working Products Goal
   - Complete Service Architecture (9 Services)
   - Service Status and Implementation Priority
   - Standalone Mode Requirements
   - United Mode Requirements
   - Architecture Clarity

II. Cryptographic Compliance and Standards
   - CGO Ban
   - FIPS 140-3 Compliance
   - Data at Rest Encryption
   - Docker/Kubernetes Secrets

III. KMS Hierarchical Key Security

IV. Go Testing Requirements
   - Test Concurrency
   - Test Data Isolation
   - Test Requirements
   - Test Execution Time Targets
   - Probability-Based Test Execution
   - main() Function Testability Pattern
   - Real Dependencies Preferred
   - Race Condition Prevention

V. Service Architecture - Dual HTTPS Endpoint Pattern
   - Deployment Environments
   - CA Architecture Pattern
   - TLS Certificate Configuration
   - Private HTTPS Endpoint
   - Public HTTPS Endpoint
   - Service Examples
   - Critical Rules

VA. Service Federation and Discovery
   - Federation Architecture
   - Service Discovery Mechanisms
   - Graceful Degradation Patterns
   - Federation Health Monitoring
   - Cross-Service Authentication
   - MFA Factor Priority
   - Federation Testing Requirements

VB. Performance, Scaling, and Resource Management
   - Vertical Scaling
   - Horizontal Scaling
   - Backup and Recovery
   - Quality Tracking Documentation

VI. CI/CD Workflow Requirements
   - GitHub Actions Service Dependencies
```

### Issues with Current Structure

1. **Inconsistent Depth**: Some sections have 3+ nesting levels (V.VA.1.a), others are flat (III has no subsections)
2. **Mixed Concerns**: Section V mixes deployment (environments), architecture (dual HTTPS), and operational (critical rules)
3. **Sequential Numbering Breaks**: V → VA → VB creates confusion (why not VI, VII, VIII?)
4. **Scattered Related Topics**: Testing (IV), CI/CD (VI), Quality Tracking (VB.4) are separated despite tight coupling
5. **No Clear Grouping**: Technical requirements, operational requirements, and governance requirements are intermixed

---

## Proposal 1: Layered Architecture Pattern

**Philosophy**: Organize by architectural layers (Application → Infrastructure → Operations → Governance)

### Proposed Structure

```
I. PRODUCT LAYER - Application Architecture
   A. Product Suite Overview
      1. Four Products (JOSE, Identity, KMS, CA)
      2. Nine Services (8 product + 1 demo)
      3. Standalone vs United Deployment Modes

   B. Service Architecture Patterns
      1. Dual HTTPS Endpoints (Public + Admin)
      2. Request Path Prefixes (/browser vs /service)
      3. Authentication/Authorization Matrix
      4. Service Federation and Discovery

   C. Product-Specific Requirements
      1. KMS Hierarchical Key Security
      2. CA Certificate Profiles
      3. Identity MFA Factor Priority
      4. JOSE Algorithm Agility

II. INFRASTRUCTURE LAYER - Technical Foundations
   A. Cryptographic Compliance
      1. FIPS 140-3 Requirements
      2. CGO Ban (Absolute Requirement)
      3. TLS Configuration (MinVersion 1.3+, cert validation)
      4. Data at Rest Encryption

   B. Database and Persistence
      1. PostgreSQL (production) vs SQLite (dev/test/small-scale)
      2. GORM ORM Patterns
      3. Database Realm vs File Realm
      4. Backup and Recovery

   C. Networking and Communication
      1. HTTPS Binding (Public vs Private)
      2. Service Discovery Mechanisms
      3. Federation Fallback Modes
      4. Health Check Semantics (livez vs readyz)

   D. Observability and Monitoring
      1. OpenTelemetry OTLP (Traces, Metrics, Logs)
      2. Telemetry Forwarding Architecture
      3. Retention Policy (90 days default)
      4. Resource Limits

III. OPERATIONS LAYER - Quality and Delivery
   A. Testing Requirements
      1. Test Concurrency (MANDATORY -shuffle, t.Parallel)
      2. Coverage Targets (95% production, 98% infrastructure)
      3. Mutation Testing (≥85% Phase 4, ≥98% Phase 5+)
      4. Test Execution Time Targets (<15s per package)
      5. Probability-Based Execution
      6. Real Dependencies vs Mocks

   B. CI/CD Workflows
      1. GitHub Actions Service Dependencies
      2. PostgreSQL Service Container Requirements
      3. Workflow Matrix (quality, test, coverage, race, etc.)
      4. Artifact Management

   C. Performance and Scaling
      1. Vertical Scaling (Resource Limits)
      2. Horizontal Scaling (Load Balancing, Session State)
      3. Database Scaling (Read Replicas, Connection Pooling)
      4. Distributed Caching

IV. GOVERNANCE LAYER - Policies and Standards
   A. Development Standards
      1. Go Project Structure (cmd, internal, pkg)
      2. Coding Conventions (gofumpt, importas)
      3. File Size Limits (300 soft, 500 hard)
      4. Error Handling Patterns

   B. Security Policies
      1. Docker/Kubernetes Secrets (MANDATORY, never env vars)
      2. IP Allowlisting
      3. Rate Limiting
      4. Windows Firewall Prevention (127.0.0.1 for tests)

   C. Quality Assurance
      1. Quality Tracking Documentation (docs/QUALITY-TODOs.md)
      2. Mutation Testing Targets
      3. Benchmark Requirements
      4. Fuzz Testing Requirements
```

### Advantages

- **Clear Separation of Concerns**: Application logic vs infrastructure vs operations vs governance
- **Logical Flow**: Top-down from product requirements to implementation details
- **Easier Navigation**: Related topics grouped together (all testing in III.A, all security in IV.B)
- **Scalability**: Easy to add new sections without breaking numbering scheme

### Disadvantages

- **Major Reorganization**: Requires moving many sections, high disruption risk
- **Deeper Nesting**: Some sections may have 4 levels (I.A.1.a), harder to reference
- **Learning Curve**: Users familiar with current structure need to relearn navigation

---

## Proposal 2: Domain-Driven Design (DDD) Pattern

**Philosophy**: Organize by business domains (Identity, Crypto, Networking, Data, Quality)

### Proposed Structure

```
I. IDENTITY DOMAIN - Authentication and Authorization
   A. Authentication Methods
      1. Browser-Based (Passkey, TOTP, Basic, etc.)
      2. Headless-Based (mTLS, Client Credentials, etc.)
      3. Realm Types (File vs Database)
      4. MFA Factor Priority (9 factors)

   B. Authorization Methods
      1. Scope-Based Authorization
      2. Role-Based Access Control (RBAC)
      3. Resource-Level Access Control
      4. Consent Tracking

   C. Session Management
      1. Session Tokens (Browser: Cookie, Headless: JWT)
      2. Token Validation (Signature, Expiry, Scope)
      3. Session Store (Redis, Database)

II. CRYPTO DOMAIN - Cryptographic Requirements
   A. Compliance Standards
      1. FIPS 140-3 (Approved Algorithms)
      2. CGO Ban (Absolute Requirement)
      3. CA/Browser Forum Baseline Requirements

   B. Key Management
      1. KMS Hierarchical Key Security (Unseal → Root → Intermediate → Content)
      2. Algorithm Agility (Configurable with FIPS defaults)
      3. Key Versioning and Rotation

   C. Data Protection
      1. Data at Rest Encryption (Deterministic vs Non-Deterministic)
      2. TLS Configuration (MinVersion 1.3+, cert validation)
      3. Docker/Kubernetes Secrets (MANDATORY)

III. NETWORKING DOMAIN - Service Communication
   A. HTTPS Endpoints
      1. Public Endpoint (Public APIs, UI)
      2. Private Endpoint (Admin APIs, Health Checks)
      3. TLS Certificate Configuration

   B. Service Federation
      1. Federation Architecture (Identity, JOSE, CA)
      2. Service Discovery (Config, Docker Compose, Kubernetes)
      3. Graceful Degradation (Circuit Breaker, Fallback Modes)

   C. Request Routing
      1. Path Prefixes (/browser vs /service)
      2. Middleware Stacks (CORS, CSRF, CSP)
      3. IP Allowlisting and Rate Limiting

IV. DATA DOMAIN - Persistence and State
   A. Database Support
      1. PostgreSQL (Production)
      2. SQLite (Dev/Test/Small-Scale)
      3. GORM ORM Patterns

   B. Schema Management
      1. Database Migrations (golang-migrate)
      2. Realm Types (File vs Database)
      3. Multi-Tenancy Patterns

   C. Backup and Recovery
      1. Database Backups (Daily, 30-day retention)
      2. Key Recovery (Version-based KeyRing)
      3. Disaster Recovery Procedures

V. QUALITY DOMAIN - Testing and Validation
   A. Testing Requirements
      1. Test Concurrency (MANDATORY -shuffle, t.Parallel)
      2. Coverage Targets (95%/98%)
      3. Mutation Testing (≥85%/≥98%)
      4. Test Execution Time (<15s per package)

   B. CI/CD Workflows
      1. GitHub Actions Service Dependencies
      2. Workflow Matrix (quality, test, coverage, race)
      3. Artifact Management

   C. Performance and Scaling
      1. Vertical Scaling (Resource Limits)
      2. Horizontal Scaling (Load Balancing)
      3. Performance Baselines (No Hard Targets)

VI. DEPLOYMENT DOMAIN - Operations and Infrastructure
   A. Product Suite
      1. Four Products (JOSE, Identity, KMS, CA)
      2. Nine Services (8 product + 1 demo)
      3. Standalone vs United Modes

   B. Service Architecture
      1. Dual HTTPS Endpoints
      2. Docker/Kubernetes Deployment
      3. Observability (OTLP, Telemetry Retention)

   C. Development Standards
      1. Go Project Structure
      2. Coding Conventions
      3. File Size Limits
```

### Advantages

- **Business-Oriented**: Domains align with business capabilities (Identity, Crypto, etc.)
- **Domain Experts**: Easier for domain experts to find relevant sections (e.g., security team → Crypto Domain)
- **Cohesive**: Related concepts grouped together (all authentication in Identity Domain)

### Disadvantages

- **Cross-Domain Dependencies**: Some topics span multiple domains (e.g., TLS in Crypto + Networking)
- **Duplication Risk**: May need to reference same concept in multiple domains
- **Less Technical Flow**: Doesn't follow typical software architecture layers

---

## Proposal 3: Lifecycle-Based Pattern

**Philosophy**: Organize by development lifecycle stages (Design → Build → Test → Deploy → Operate)

### Proposed Structure

```
I. DESIGN PHASE - Architecture and Requirements
   A. Product Suite Design
      1. Four Products (JOSE, Identity, KMS, CA)
      2. Nine Services (8 product + 1 demo)
      3. Standalone vs United Deployment Modes

   B. Service Architecture Design
      1. Dual HTTPS Endpoints (Public + Admin)
      2. Request Path Prefixes (/browser vs /service)
      3. Authentication/Authorization Patterns
      4. Service Federation Architecture

   C. Security Design
      1. FIPS 140-3 Compliance
      2. KMS Hierarchical Key Security
      3. TLS Configuration Requirements
      4. Data Protection Patterns

II. BUILD PHASE - Implementation Standards
   A. Development Standards
      1. Go Project Structure (cmd, internal, pkg)
      2. CGO Ban (Absolute Requirement)
      3. Coding Conventions (gofumpt, importas)
      4. File Size Limits (300 soft, 500 hard)

   B. Database Implementation
      1. PostgreSQL vs SQLite Support
      2. GORM ORM Patterns
      3. Database Migrations (golang-migrate)
      4. Realm Types (File vs Database)

   C. Cryptographic Implementation
      1. FIPS 140-3 Approved Algorithms
      2. Algorithm Agility
      3. Key Versioning and Rotation
      4. Secrets Management (Docker/Kubernetes)

III. TEST PHASE - Quality Assurance
   A. Test Requirements
      1. Test Concurrency (MANDATORY -shuffle, t.Parallel)
      2. Coverage Targets (95%/98%)
      3. Test Execution Time (<15s per package)
      4. Probability-Based Execution

   B. Test Types
      1. Unit Tests (table-driven, t.Parallel)
      2. Integration Tests (PostgreSQL test containers)
      3. Benchmark Tests (crypto operations)
      4. Fuzz Tests (parsers, validators)
      5. Property-Based Tests (gopter)
      6. Mutation Tests (gremlins ≥85%/≥98%)

   C. Test Patterns
      1. main() Testability (internalMain pattern)
      2. Real Dependencies vs Mocks
      3. Test Data Isolation (UUIDv7, port 0)
      4. Race Condition Prevention

IV. DEPLOY PHASE - CI/CD and Packaging
   A. CI/CD Workflows
      1. GitHub Actions Service Dependencies
      2. PostgreSQL Service Container Requirements
      3. Workflow Matrix (quality, test, coverage, race, etc.)
      4. Artifact Management

   B. Docker/Kubernetes Deployment
      1. Multi-Stage Builds
      2. Docker Secrets (MANDATORY, never env vars)
      3. Health Check Configuration
      4. Telemetry Sidecar (otel-collector-contrib)

   C. Service Configuration
      1. HTTPS Binding (Public vs Private)
      2. TLS Certificate Configuration
      3. Federation Configuration
      4. Database Configuration

V. OPERATE PHASE - Runtime and Monitoring
   A. Observability
      1. OpenTelemetry OTLP (Traces, Metrics, Logs)
      2. Telemetry Forwarding Architecture
      3. Retention Policy (90 days default)
      4. Resource Limits

   B. Service Federation
      1. Service Discovery Mechanisms
      2. Graceful Degradation (Circuit Breaker, Fallback Modes)
      3. Federation Health Monitoring
      4. Cross-Service Authentication

   C. Performance and Scaling
      1. Vertical Scaling (Resource Limits)
      2. Horizontal Scaling (Load Balancing, Session State)
      3. Database Scaling (Read Replicas, Connection Pooling)
      4. Backup and Recovery

   D. Health and Reliability
      1. Health Check Semantics (livez vs readyz)
      2. Graceful Shutdown
      3. Error Handling and Logging
      4. Incident Response
```

### Advantages

- **Developer-Friendly**: Follows natural development workflow (design → build → test → deploy → operate)
- **Phase-Specific**: Easy to find requirements for current phase of work
- **Clear Milestones**: Each phase has completion criteria

### Disadvantages

- **Cross-Phase Topics**: Some topics span multiple phases (e.g., security in Design + Build + Operate)
- **Repetition**: May need to reference same concept in multiple phases
- **Less Intuitive for Reference**: Harder to find topic if you don't know which phase it belongs to

---

## Proposal 4: Compliance-First Pattern

**Philosophy**: Organize by compliance requirements and non-negotiable rules (MUST, SHOULD, MAY)

### Proposed Structure

```
I. ABSOLUTE REQUIREMENTS (MUST / SHALL / MANDATORY)
   A. Security Mandates
      1. CGO Ban (Absolute Requirement)
      2. FIPS 140-3 Compliance (No Non-Approved Algorithms)
      3. TLS 1.3+ Minimum (Never InsecureSkipVerify)
      4. Docker/Kubernetes Secrets (NEVER Environment Variables)
      5. Test Concurrency (NEVER -p=1 or -parallel=1)

   B. Architecture Mandates
      1. Dual HTTPS Endpoints (Public + Admin) for ALL Services
      2. Service Federation Support (Configurable, Never Hardcoded)
      3. PostgreSQL Support (ALL Workflows Running go test)
      4. Health Check Implementation (livez, readyz, shutdown)

   C. Quality Mandates
      1. Coverage Targets (95%/98% MANDATORY)
      2. Mutation Testing (≥85%/≥98% MANDATORY)
      3. Test Execution Time (<15s per package MANDATORY)
      4. Race Detection (CGO_ENABLED=1 for -race workflow ONLY)

   D. Operational Mandates
      1. Graceful Shutdown Support
      2. OTLP Telemetry Forwarding (NEVER Bypass Sidecar)
      3. Database Migrations on Startup
      4. Configuration Validation on Startup

II. STRONG RECOMMENDATIONS (SHOULD / RECOMMENDED)
   A. Architecture Recommendations
      1. File Realm Type (Disaster Recovery)
      2. Service Template Usage (learn-ps validates reusability)
      3. mTLS for Service-to-Service (Preferred over OAuth 2.1)
      4. PostgreSQL for Production (SQLite acceptable <1000 req/day)

   B. Testing Recommendations
      1. Real Dependencies (Preferred over Mocks)
      2. Property-Based Tests (gopter for invariants)
      3. Benchmark Tests (Crypto operations, hot paths)
      4. Probability-Based Execution (Packages approaching 15s limit)

   C. Security Recommendations
      1. Passkey MFA (Highest priority for user-facing auth)
      2. mTLS (Highest priority for service-to-service auth)
      3. Scope-Based Authorization (Preferred over RBAC only)
      4. IP Allowlisting (Additional security layer)

III. OPTIONAL FEATURES (MAY / OPTIONAL)
   A. Optional Services
      1. Identity-RP (Reference Implementation)
      2. Identity-SPA (Reference Implementation)
      3. Learn-PS (Educational Demonstration)

   B. Optional Federation
      1. CA Federation (KMS can use internal TLS certs)
      2. JOSE Federation (Services can use internal crypto)

   C. Optional MFA Factors
      1. Phone Call OTP (NIST deprecated, backward compatibility)
      2. SMS OTP (NIST deprecated, accessibility)
      3. Magic Link (Passwordless alternative)
      4. Push Notification (Requires mobile app)

   D. Optional Scaling Features
      1. Horizontal Scaling (Load balancing)
      2. Database Sharding (Future consideration)
      3. Distributed Caching (Redis/Memcached)

IV. IMPLEMENTATION DETAILS
   A. Product Suite
      1. Four Products (JOSE, Identity, KMS, CA)
      2. Nine Services (8 product + 1 demo)
      3. Standalone vs United Deployment Modes

   B. Service Architecture
      1. Dual HTTPS Endpoints (Public + Admin)
      2. Request Path Prefixes (/browser vs /service)
      3. Authentication/Authorization Matrix
      4. Service Federation and Discovery

   C. Database and Persistence
      1. PostgreSQL (Production) vs SQLite (Dev/Test/Small-Scale)
      2. GORM ORM Patterns
      3. Database Migrations
      4. Backup and Recovery

   D. Cryptographic Implementation
      1. FIPS 140-3 Approved Algorithms
      2. KMS Hierarchical Key Security
      3. Algorithm Agility
      4. Key Versioning and Rotation

   E. Testing Implementation
      1. Test Concurrency Patterns
      2. Coverage Targets and Measurement
      3. Mutation Testing with Gremlins
      4. Probability-Based Execution

   F. CI/CD Implementation
      1. GitHub Actions Workflows
      2. Docker/Kubernetes Deployment
      3. Observability and Monitoring
      4. Performance and Scaling
```

### Advantages

- **Compliance-Oriented**: Immediately clear what is non-negotiable vs recommended vs optional
- **Risk Management**: Critical requirements highlighted at top (Section I)
- **Decision Support**: Helps prioritize what to implement first (MUST > SHOULD > MAY)
- **Audit-Friendly**: Easy to verify compliance with absolute requirements

### Disadvantages

- **Fragmented Topics**: Same topic split across sections (e.g., testing in I.C, II.B, III.D, IV.E)
- **Harder Navigation**: Need to check multiple sections to understand full topic
- **Less Intuitive**: Doesn't follow typical software architecture organization

---

## Proposal 5: Hybrid Pattern (Recommended)

**Philosophy**: Combine best aspects of Layered (clear hierarchy) + DDD (domain grouping) + Compliance (priority levels)

### Proposed Structure

```
I. EXECUTIVE SUMMARY - Critical Requirements
   A. Absolute Requirements (MUST / MANDATORY)
      1. CGO Ban (Production builds, testing, Docker)
      2. FIPS 140-3 Compliance (No non-approved algorithms)
      3. Dual HTTPS Endpoints (ALL services)
      4. Test Concurrency (NEVER disable parallelism)
      5. Docker/Kubernetes Secrets (NEVER environment variables)

   B. Product Suite Overview
      1. Four Products: JOSE, Identity, KMS, CA
      2. Nine Services: 8 product + 1 demo (learn-ps)
      3. Deployment Modes: Standalone vs United

   C. Quality Gates
      1. Coverage: 95% production, 98% infrastructure/utility
      2. Mutation: ≥85% Phase 4, ≥98% Phase 5+
      3. Test Timing: <15s per package, <180s full suite

II. PRODUCT ARCHITECTURE - Service Design
   A. Service Portfolio
      1. P1: JOSE (jose-ja)
      2. P2: Identity (authz, idp, rs, rp, spa)
      3. P3: KMS (sm-kms)
      4. P4: CA (pki-ca)
      5. Demo: Learn-PS (learn-ps)

   B. Dual HTTPS Endpoint Pattern
      1. Public Endpoint (Public APIs, Browser UI)
      2. Private Endpoint (Admin APIs, Health Checks)
      3. Request Path Prefixes (/browser vs /service)
      4. Middleware Stacks (CORS, CSRF, CSP, Token Validation)

   C. Service Federation
      1. Federation Architecture (Identity, JOSE, CA)
      2. Service Discovery (Config, Docker Compose, Kubernetes)
      3. Graceful Degradation (Circuit Breaker, Fallback Modes)
      4. Cross-Service Authentication (mTLS, OAuth 2.1)

III. SECURITY ARCHITECTURE - Cryptography and Access Control
   A. Cryptographic Compliance
      1. FIPS 140-3 Requirements (Approved Algorithms)
      2. CA/Browser Forum Baseline Requirements
      3. Algorithm Agility (Configurable with FIPS defaults)

   B. Key Management
      1. KMS Hierarchical Key Security (Unseal → Root → Intermediate → Content)
      2. Key Versioning and Rotation
      3. Elastic Keys (Active key for encrypt/sign, historical for decrypt/verify)

   C. Authentication and Authorization
      1. Authentication Methods (17 methods, priority-ordered)
      2. Authorization Methods (Scope, RBAC, Resource-Level)
      3. Realm Types (File > Database for disaster recovery)
      4. MFA Factor Priority (9 factors including deprecated SMS/Phone OTP)

   D. Data Protection
      1. Data at Rest Encryption (Deterministic vs Non-Deterministic)
      2. TLS Configuration (MinVersion 1.3+, cert validation)
      3. Docker/Kubernetes Secrets (MANDATORY)

IV. DATA ARCHITECTURE - Persistence and State
   A. Database Support
      1. PostgreSQL (Production default)
      2. SQLite (Dev/test/small-scale <1000 req/day production acceptable)
      3. GORM ORM Patterns

   B. Schema Management
      1. Database Migrations (golang-migrate, embedded SQL)
      2. Realm Types (File vs Database)
      3. Multi-Tenancy Patterns (Tenant ID in all tables)

   C. Backup and Recovery
      1. Database Backups (Daily, 30-day retention)
      2. Key Recovery (Version-based KeyRing pattern)
      3. Disaster Recovery (File Realm Type for admin access)

V. QUALITY ARCHITECTURE - Testing and Validation
   A. Testing Requirements
      1. Test Concurrency (MANDATORY -shuffle, t.Parallel)
      2. Coverage Targets (95%/98%)
      3. Test Execution Time (<15s per package)
      4. Probability-Based Execution (TestProbAlways/Quarter/Tenth)

   B. Test Types and Patterns
      1. Unit Tests (table-driven, t.Parallel)
      2. Integration Tests (PostgreSQL test containers)
      3. Benchmark Tests (crypto operations, hot paths)
      4. Fuzz Tests (parsers, validators, 15s minimum)
      5. Property-Based Tests (gopter for invariants)
      6. Mutation Tests (gremlins ≥85%/≥98%)

   C. Test Implementation Patterns
      1. main() Testability (internalMain pattern)
      2. Real Dependencies vs Mocks (prefer real)
      3. Test Data Isolation (UUIDv7, port 0)
      4. Race Condition Prevention (never share state in parallel tests)

VI. OPERATIONS ARCHITECTURE - Deployment and Runtime
   A. CI/CD Workflows
      1. GitHub Actions Service Dependencies (PostgreSQL container)
      2. Workflow Matrix (quality, test, coverage, race, mutation, etc.)
      3. Docker Multi-Stage Builds
      4. Secrets Management (Docker secrets, never env vars)

   B. Observability and Monitoring
      1. OpenTelemetry OTLP (Traces, Metrics, Logs)
      2. Telemetry Forwarding (MANDATORY through otel-collector sidecar)
      3. Retention Policy (90 days default, no redaction)
      4. Health Check Semantics (livez vs readyz)

   C. Performance and Scaling
      1. Vertical Scaling (Resource Limits: CPU, Memory)
      2. Horizontal Scaling (Load Balancing, Session State)
      3. Database Scaling (Read Replicas, Connection Pooling)
      4. Performance Baselines (No hard targets, track trends)

VII. DEVELOPMENT STANDARDS - Implementation Guidelines
   A. Go Project Structure
      1. Standard Layout (cmd, internal, pkg, api, configs, deployments)
      2. CGO Ban (CGO_ENABLED=0 MANDATORY except -race workflow)
      3. Import Alias Conventions (cryptoutil prefix, googleUuid, etc.)

   B. Coding Conventions
      1. File Size Limits (300 soft, 400 medium, 500 hard)
      2. Formatting (gofumpt, not gofmt)
      3. Error Handling (wrap errors, context propagation)
      4. Magic Values (internal/shared/magic packages)

   C. Quality Tracking
      1. docs/QUALITY-TODOs.md (coverage/gremlins challenges)
      2. Mutation Testing Targets (per package)
      3. Lessons Learned Documentation
```

### Advantages

- **Best of All Worlds**: Combines compliance priority (I), domain grouping (III, IV), lifecycle flow (V, VI), and standards (VII)
- **Executive Summary**: Section I provides quick reference for critical requirements
- **Logical Hierarchy**: Clear flow from high-level (products) to low-level (code standards)
- **Domain Cohesion**: Related topics grouped together (all crypto in III, all testing in V)
- **Easier Navigation**: Predictable structure, clear section purposes

### Disadvantages

- **Most Complex**: Requires understanding multiple organizational principles
- **Deeper Nesting**: Some sections may have 4 levels (III.C.4.a), harder to reference
- **Migration Effort**: Significant reorganization from current structure

---

## Comparison Matrix

| Criteria | Current | Layered | DDD | Lifecycle | Compliance | Hybrid |
|----------|---------|---------|-----|-----------|------------|--------|
| **Clarity** | 6/10 | 8/10 | 7/10 | 7/10 | 6/10 | 9/10 |
| **Navigation** | 5/10 | 7/10 | 8/10 | 6/10 | 5/10 | 8/10 |
| **Compliance-Friendly** | 6/10 | 6/10 | 5/10 | 6/10 | 10/10 | 8/10 |
| **Developer-Friendly** | 7/10 | 7/10 | 6/10 | 9/10 | 5/10 | 8/10 |
| **Maintainability** | 5/10 | 8/10 | 7/10 | 7/10 | 6/10 | 9/10 |
| **Migration Effort** | 0/10 (baseline) | 8/10 | 7/10 | 8/10 | 9/10 | 9/10 |
| **Risk of Errors** | 0/10 (baseline) | 6/10 | 5/10 | 6/10 | 7/10 | 7/10 |
| **Total Score** | 29/60 | 50/60 | 45/60 | 49/60 | 48/60 | 58/60 |

**Scoring**:

- 10/10 = Excellent
- 8-9/10 = Good
- 6-7/10 = Acceptable
- 4-5/10 = Needs Improvement
- 0-3/10 = Poor

---

## Recommendation

**Recommended**: **Proposal 5 (Hybrid Pattern)** with phased migration

### Migration Strategy

**Phase 1: Non-Disruptive Additions** (Low Risk):

1. Add Section I: Executive Summary at beginning
2. Keep existing structure intact (sections II-VI)
3. Cross-reference new Executive Summary to existing sections

**Phase 2: Incremental Reorganization** (Medium Risk):

1. Consolidate testing sections (IV + VB.4 → V)
2. Consolidate security sections (II + VA.5 → III)
3. Update cross-references

**Phase 3: Full Reorganization** (High Risk):

1. Implement full Hybrid structure
2. Update all cross-references
3. Update copilot instructions
4. Notify all stakeholders

### Alternative: Minimal Changes

If full reorganization is too risky, apply **Quick Wins** from Hybrid proposal:

1. Add Executive Summary (Section I) at beginning
2. Fix numbering inconsistencies (VA/VB → VI/VII)
3. Add subsection summaries for long sections
4. Create cross-reference index at end

---

## References

- Current Constitution: `.specify/memory/constitution.md` (Version 3.0.0, 1257 lines)
- RFC 2119: Key words for use in RFCs to Indicate Requirement Levels
- NIST SP 800-63B: Digital Identity Guidelines
- ISO/IEC 27001: Information Security Management Systems

---

**Document Version**: 1.0.0
**Last Updated**: 2025-12-22
**Maintainer**: Spec Kit AI Agent
