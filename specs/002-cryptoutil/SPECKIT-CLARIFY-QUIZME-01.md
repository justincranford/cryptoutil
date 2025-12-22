# cryptoutil CLARIFY-QUIZME-01.md

**Last Updated**: 2025-12-22
**Purpose**: Multiple choice questions for UNKNOWN answers requiring user input
**Format**: A-D options + E write-in for each question

## Instructions

**CRITICAL**: This file contains ONLY questions with UNKNOWN answers that require user clarification.

**Questions with KNOWN answers belong in clarify.md, NOT here.**

**When adding questions**:

1. Search copilot instructions, constitution.md, spec.md, codebase FIRST
2. If answer is KNOWN: Add Q&A to clarify.md and update constitution/spec as needed
3. If answer is UNKNOWN: Add question HERE with NO pre-filled answers
4. After user answers: Refactor clarify.md to cover answered questions, update constitution.md with architecture decisions, and update spec.md with finalized requirements

---

## Open Questions Requiring User Input

### Q1: Service Template Migration Priority for Identity Services

**Context**: Constitution.md Section I states identity services have COMPLETE dual-server implementation. However, 02-02.service-template.instructions.md mandates extracting service template from KMS and validating with Learn-PS BEFORE migrating any production services.

**Question**: Should identity services (authz, idp, rs) be refactored to use the extracted service template immediately after Learn-PS validation, or should they remain with their current dual-server implementation until after all other services (JOSE, CA) are migrated?

**Options**:

A) Refactor identity services immediately after Learn-PS validation (highest priority after demo service)
B) Refactor identity services only after JOSE and CA are migrated (lower priority)
C) Keep identity services on current implementation indefinitely (never migrate to template)
D) Refactor identity services in parallel with JOSE/CA migration (all at once)
E) Other approach (please specify): _______________

**User Answer**: ___

---

### Q2: Admin Port Mapping Strategy for External Monitoring

**Context**: Constitution.md Section V states private endpoints MUST ALWAYS use 127.0.0.1:9090 and are "NOT exposed in Docker port mappings". However, external monitoring tools (Prometheus, Grafana) may need to scrape `/admin/v1/metrics` endpoint.

**Question**: Should admin ports be exposed via Docker Compose port mappings for external monitoring, or should monitoring tools run inside the Docker network?

**Options**:

A) NEVER expose admin ports externally; monitoring tools MUST run inside Docker network (strict security)
B) Expose admin ports ONLY in development/testing Docker Compose files, NEVER in production
C) Expose admin ports with unique external mappings per service (e.g., 19090 for KMS, 19091 for Identity)
D) Use a sidecar metrics exporter that forwards metrics to external Prometheus
E) Other approach (please specify): _______________

**User Answer**: ___

---

### Q3: SQLite Production Readiness for Single-Instance Deployments

**Context**: Constitution.md mandates dual support for PostgreSQL (prod) and SQLite (dev). Current implementation uses SQLite for development only. However, small-scale deployments may prefer SQLite for simplicity.

**Question**: Should SQLite be supported and tested for production single-instance deployments, or remain strictly development-only?

**Options**:

A) SQLite is ONLY for development/testing, NEVER for production (current approach)
B) SQLite is acceptable for production single-instance deployments with <1000 requests/day
C) SQLite is acceptable for production if properly configured (WAL mode, busy timeout, connection limits)
D) SQLite is production-ready for all deployment sizes (equivalent to PostgreSQL)
E) Other approach (please specify): _______________

**User Answer**: ___

---

### Q4: MFA Factor Priority and Implementation Sequence

**Context**: Constitution.md does not specify MFA factor implementation priority. Spec.md lists 9 MFA factors with varying priority levels (HIGHEST to LOW). Some factors are marked NIST deprecated but MANDATORY (SMS OTP).

**Question**: What is the mandatory implementation sequence for MFA factors in Phase 2? Should deprecated factors (SMS OTP) be skipped or implemented for compatibility?

**Options**:

A) Implement only HIGHEST/HIGH priority factors (Passkey, TOTP, Hardware Keys), skip deprecated factors
B) Implement all factors except NIST-deprecated ones (skip SMS OTP, Phone Call OTP)
C) Implement all factors including deprecated ones for backward compatibility (SMS OTP required)
D) Implement factors incrementally: Phase 2.1 (Passkey, TOTP), Phase 2.2 (Email OTP, Recovery Codes), Phase 3 (remaining)
E) Other sequence (please specify): _______________

**User Answer**: ___

---

### Q5: Certificate Profile Library Extensibility

**Context**: Spec.md Section P4 states "24 predefined certificate profiles" are implemented. However, organizations may need custom profiles for specific use cases.

**Question**: Should the CA support custom certificate profiles via configuration, or are the 24 predefined profiles sufficient for all use cases?

**Options**:

A) 24 predefined profiles are sufficient; NO custom profile support needed
B) Support custom profiles via YAML configuration files (file-based extensibility)
C) Support custom profiles via database-driven configuration (dynamic runtime extensibility)
D) Support custom profiles via plugin system (external Go modules loaded at runtime)
E) Other approach (please specify): _______________

**User Answer**: ___

---

### Q6: Telemetry Data Retention and Privacy

**Context**: Constitution.md requires OpenTelemetry integration. However, telemetry data may contain sensitive information (user IDs, request paths, IP addresses) that must comply with data privacy regulations (GDPR, CCPA).

**Question**: What data retention policy should be enforced for telemetry data, and should sensitive fields be redacted/hashed?

**Options**:

A) Retain all telemetry data indefinitely; no redaction (full observability)
B) Retain telemetry data for 30 days; redact PII fields (user IDs, IP addresses) automatically
C) Retain telemetry data for 90 days; allow configurable redaction patterns per deployment
D) Retain telemetry data for 7 days only; aggressive redaction of all potentially sensitive data
E) Other policy (please specify): _______________

**User Answer**: ___

---

### Q7: Federation Fallback Mode for Production Deployments

**Context**: Clarify.md Section "Service Federation Configuration" describes three fallback modes when Identity service is unavailable: local_validation (cached keys), reject_all (strict), allow_all (dev only).

**Question**: What is the MANDATORY fallback mode for production deployments when the Identity service is unavailable?

**Options**:

A) reject_all (strict mode) - Deny all requests until Identity service recovers (maximum security)
B) local_validation (cached keys) - Continue serving with cached public keys for token validation (high availability)
C) Configurable per deployment - Allow operators to choose based on security vs availability trade-off
D) Hybrid mode - Use local_validation for first N minutes, then switch to reject_all (graceful degradation)
E) Other approach (please specify): _______________

**User Answer**: ___

---

### Q8: Docker Secrets vs Kubernetes Secrets Priority

**Context**: Constitution.md mandates "Docker/Kubernetes secrets" for all sensitive data. However, implementation details differ significantly between Docker Compose and Kubernetes.

**Question**: Should the codebase prioritize Docker secrets integration (file:///run/secrets/*) or Kubernetes secrets integration (env vars from secrets, volume mounts)?

**Options**:

A) Docker secrets ONLY; Kubernetes deployments must use Docker-compatible secret mounting
B) Kubernetes secrets ONLY; Docker Compose deployments must emulate Kubernetes patterns
C) Support both equally; detect runtime environment and use appropriate secret provider
D) Use external secret management (HashiCorp Vault, AWS Secrets Manager) for both Docker and Kubernetes
E) Other approach (please specify): _______________

**User Answer**: ___

---

### Q9: Load Testing Target Performance Metrics

**Context**: Spec.md Section I3 identifies missing load test coverage for Browser API and Admin API. However, target performance metrics (requests/second, latency percentiles, error rates) are not specified.

**Question**: What are the target performance metrics for load testing across all API types?

**Options**:

A) Service API: 1000 req/s @ p95 <100ms, Browser API: 100 req/s @ p95 <200ms, Admin API: 10 req/s @ p95 <50ms
B) Service API: 500 req/s @ p95 <200ms, Browser API: 50 req/s @ p95 <500ms, Admin API: 5 req/s @ p95 <100ms
C) Service API: 100 req/s @ p95 <500ms, Browser API: 10 req/s @ p95 <1s, Admin API: 1 req/s @ p95 <200ms (relaxed)
D) No hard targets; load tests validate scalability trends and identify bottlenecks only
E) Other metrics (please specify): _______________

**User Answer**: ___

---

### Q10: E2E Test Workflow Coverage Priority

**Context**: Spec.md Section "E2E Test Scope" lists missing critical workflows: OAuth 2.1 flow, certificate issuance, KMS operations, JOSE signing. However, implementing all workflows is time-intensive.

**Question**: What is the minimum viable E2E test coverage for Phase 2 completion?

**Options**:

A) ONLY OAuth 2.1 authorization code flow (identity product validation)
B) OAuth 2.1 flow + KMS encryption/decryption (2 core product workflows)
C) OAuth 2.1 + KMS + CA certificate issuance (3 core product workflows)
D) OAuth 2.1 + KMS + CA + JOSE signing (all 4 core product workflows)
E) Other minimum viable coverage (please specify): _______________

**User Answer**: ___

---

### Q11: Mutation Testing Enforcement Strategy

**Context**: Constitution.md mandates ≥85% gremlins score (Phase 4) and ≥98% (Phase 5+). However, some packages may have inherently low mutant killability (generated code, simple CRUD operations).

**Question**: Should mutation testing targets be enforced strictly per package, or should exemptions be allowed for specific package types?

**Options**:

A) Strict enforcement: ALL packages MUST meet ≥85%/≥98% targets, NO exceptions
B) Allow exemptions for generated code (e.g., OpenAPI-generated models) with documentation
C) Allow exemptions for simple CRUD repositories with justification in clarify.md
D) Use package-level targets: 98% for business logic, 85% for infrastructure, 70% for generated code
E) Other enforcement approach (please specify): _______________

**User Answer**: ___

---

### Q12: Probabilistic Testing Seed Management

**Context**: Constitution.md mandates probabilistic testing (TestProbQuarter, TestProbTenth) for algorithm variants. However, seed management for deterministic random selection is not specified.

**Question**: Should probabilistic test execution use fixed seeds for reproducibility, or random seeds for broader coverage?

**Options**:

A) Fixed seed per test run (SEED=12345) - Reproducible subset selection for debugging
B) Random seed per test run - Broader coverage over multiple CI/CD executions
C) Fixed seed in CI/CD, random seed in local development - Balance reproducibility and coverage
D) Date-based seed (YYYYMMDD) - Deterministic daily rotation of test subset
E) Other seed strategy (please specify): _______________

**User Answer**: ___

---

## Status Summary

**Total Questions**: 12
**Answered**: 0
**Pending User Input**: 12

**Next Steps**:

1. User reviews questions and provides answers (A/B/C/D/E with write-in if E)
2. Agent refactors clarify.md to incorporate answers into topical Q&A format
3. Agent updates constitution.md with architectural decisions
4. Agent updates spec.md with finalized requirements
5. Agent updates plan.md and tasks.md based on clarifications
6. Agent removes answered questions from this file or marks as resolved

---

**Note**: All questions in this file represent genuine UNKNOWNS requiring user decisions. Questions with answers already available in copilot instructions, constitution.md, spec.md, or codebase have been excluded and documented in clarify.md instead.
