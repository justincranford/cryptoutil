# Certificate Authority Expansion Plan

## Purpose

- Deliver an independent CA subsystem under `internal/ca/` with no hidden dependencies on other domains.
- Provide YAML-driven configuration for cryptographic materials, subject profiles, and certificate profiles.
- Ensure every CA workflow is automated, tested, and ready for production and compliance audits.

## Guiding Principles

- Use deterministic YAML schemas for crypto, subject, and certificate policies.
- Enforce strict support for CA/Browser Forum Baseline Requirements and RFC 5280.
- Design for multi-backend support (PostgreSQL, SQLite) with identical behavior.
- Integrate security tooling (gosec, golangci-lint) and observability from day one.

## Implementation Status

| Task | Status | Package/Location |
|------|--------|------------------|
| Task 1: Domain Charter | ✅ Complete | `docs/05-ca/charter.md` |
| Task 2: Config Schema | ✅ Complete | `configs/ca/`, `internal/ca/config/` |
| Task 3: Crypto Provider | ✅ Complete | `internal/ca/crypto/` |
| Task 4: Subject Profiles | ✅ Complete | `internal/ca/profile/subject/` |
| Task 5: Certificate Profiles | ✅ Complete | `internal/ca/profile/certificate/` |
| Task 6: Root CA Bootstrap | ✅ Complete | `internal/ca/bootstrap/` |
| Task 7: Intermediate CA | ✅ Complete | `internal/ca/issuance/` |
| Task 8: Issuing CA Lifecycle | ✅ Complete | `internal/ca/lifecycle/` |
| Task 9: Enrollment API | ✅ Complete | `internal/ca/enrollment/` |
| Task 10: Revocation Services | ✅ Complete | `internal/ca/service/revocation/` |
| Task 11: Time-Stamping | ✅ Complete | `internal/ca/service/timestamp/` |
| Task 12: RA Workflows | ✅ Complete | `internal/ca/service/ra/` |
| Task 13: Profile Library | ✅ Complete | `configs/ca/profiles/` (24 profiles) |
| Task 14: Storage Layer | ✅ Complete | `internal/ca/storage/` |
| Task 15: CLI Tooling | ✅ Complete | `internal/ca/cli/` |
| Task 16: Observability | ✅ Complete | `internal/ca/observability/` |
| Task 17: Security Hardening | ⏳ Pending | - |
| Task 18: Compliance/Audit | ⏳ Pending | - |
| Task 19: Deployment Bundles | ⏳ Pending | - |
| Task 20: Final Handover | ⏳ Pending | - |

## Package Structure

```text
internal/ca/
├── bootstrap/        # Root CA bootstrap workflow
├── cli/              # CLI tooling for CA operations
├── config/           # Configuration types and loading
├── crypto/           # Crypto provider abstractions
├── enrollment/       # Certificate enrollment service
├── issuance/         # Certificate issuance service
├── lifecycle/        # CA lifecycle management
├── observability/    # Metrics, tracing, audit logging
├── profile/
│   ├── certificate/  # Certificate profile engine
│   └── subject/      # Subject profile engine
├── service/
│   ├── ra/           # Registration Authority workflows
│   ├── revocation/   # CRL and OCSP services
│   └── timestamp/    # Time-stamping service
└── storage/          # Certificate storage layer

configs/ca/
├── crypto/           # Crypto configuration YAML files
├── profiles/         # 24 certificate profile YAML files
└── subjects/         # Subject template YAML files
```

## Task Breakdown

### Task 1: Domain Charter and Scope Definition

- Analysis Focus: Capture the target capabilities, compliance obligations, and non-goals.
- Issues to Address: Lack of formal requirements for CA features.
- Implementation Notes: Interview stakeholders, map compliance references.
- Deliverables: `docs/ca/charter.md`, scope matrix, glossary.
- Validation: Stakeholder sign-off on charter.

### Task 2: Configuration Schema Design

- Analysis Focus: Define YAML schema covering crypto parameters, subject templates, and certificate profiles.
- Issues to Address: Missing authoritative schema; risk of inconsistent configs.
- Implementation Notes: Use JSON Schema or Cue for validation; document defaults.
- Deliverables: `docs/ca/config-schema.yaml`, validation utilities.
- Validation: Schema unit tests; sample configs validated in CI.

### Task 3: Crypto Provider Abstractions

- Analysis Focus: Build provider interfaces for key generation, storage, and signing.
- Issues to Address: Prevent direct coupling to existing crypto modules.
- Implementation Notes: Support RSA, ECDSA, EdDSA, HMAC, and future PQC stubs.
- Deliverables: `internal/ca/crypto/provider.go`, memory and filesystem implementations.
- Validation: `go test ./internal/ca/crypto/...`; lint checks.

### Task 4: Subject Profile Engine

- Analysis Focus: Implement subject template resolution (fields, SANs, constraints).
- Issues to Address: Manual duplication of subject details.
- Implementation Notes: Provide library to render templates from YAML with validation.
- Deliverables: `internal/ca/profile/subject` package, examples.
- Validation: Unit tests with fixtures from `configs/ca/subjects/`.

### Task 5: Certificate Profile Engine

- Analysis Focus: Implement certificate policy rendering (key usage, extensions, lifetimes).
- Issues to Address: Inconsistent policy enforcement.
- Implementation Notes: Support 20+ profile archetypes as mandated.
- Deliverables: `internal/ca/profile/certificate` package, library of predefined profiles.
- Validation: Profile validation tests; golden files for DER comparisons.

### Task 6: Root CA Bootstrap Workflow

- Analysis Focus: Provide CLI and library support to bootstrap offline root CAs.
- Issues to Address: Manual root creation with inconsistent metadata.
- Implementation Notes: Support deterministic serial numbers, key storage, audit logs.
- Deliverables: `cmd/ca/root-bootstrap`, documentation, sealed storage strategy.
- Validation: Integration tests using temporary directories; metadata verification.

### Task 7: Intermediate CA Provisioning

- Analysis Focus: Build workflow to create and sign intermediate CAs.
- Issues to Address: Missing automation for subordinate hierarchies.
- Implementation Notes: Support cross-signing, path length constraints, emergency rollover.
- Deliverables: CLI command, policy templates, runbooks.
- Validation: Integration tests verifying chain building and validation using `openssl` or Go crypto.

### Task 8: Issuing CA Lifecycle Management

- Analysis Focus: Manage issuing CAs (TLS server/client, code signing, etc.).
- Issues to Address: Lack of rotation policies and monitoring hooks.
- Implementation Notes: Provide rotation scheduler, status reporting via OTEL.
- Deliverables: Lifecycle controller, metrics dashboard, rotation documentation.
- Validation: Automated rotation simulation; metrics verification.

### Task 9: End-Entity Enrollment API

- Analysis Focus: Implement REST API for certificate enrollment (CSR submission, issuance).
- Issues to Address: Missing API contract, error handling, audit logging.
- Implementation Notes: Use OpenAPI-first approach; generate handlers.
- Deliverables: `api/ca/openapi_spec.yaml`, generated code, handler implementations.
- Validation: Contract tests (client/server), integration suite with mock CSRs.

### Task 10: Revocation Services (CRL & OCSP)

- Analysis Focus: Implement revocation operations including CRL generation and OCSP responders.
- Issues to Address: No revocation infrastructure; compliance gap.
- Implementation Notes: Support incremental CRL updates, delta CRLs, OCSP responder with caching.
- Deliverables: Revocation services, CLI tooling, monitoring dashboards.
- Validation: Conformance tests with OpenSSL, OCSP integration tests.

### Task 11: Time-Stamping and Time Signer Support

- Analysis Focus: Provide TSA-like functionality for timestamp tokens.
- Issues to Address: Missing support for time-based assertions.
- Implementation Notes: Implement RFC 3161-like endpoints, integrate with signing keys.
- Deliverables: Time-signing service, schemas, documentation.
- Validation: Unit and integration tests using RFC-compliant fixtures.

### Task 12: Registration Authority (RA) Workflows

- Analysis Focus: Build RA service for request validation, identity proofing, and approval routing.
- Issues to Address: Missing RA layer; manual validation risk.
- Implementation Notes: Provide queue-backed workflow, policy-driven approvals.
- Deliverables: RA microservice, admin UI stubs, audit logging.
- Validation: Workflow integration tests; manual RA simulation scenarios.

### Task 13: Automation for 20+ Profile Library

- Analysis Focus: Deliver catalog covering mandated CA profiles (root, intermediate, TLS server/client, S/MIME, code signing, document signing, VPN, IoT, SAML, JWT, OCSP, RA, TSA, CT log, ACME, SCEP, EST, CMP, enterprise custom).
- Issues to Address: Guarantee library completeness and documentation.
- Implementation Notes: Generate templates from schema; publish reference table.
- Deliverables: Profile catalog in `docs/ca/profiles.md`, YAML samples.
- Validation: Automated validation ensuring every profile passes schema and smoke issuance tests.

### Task 14: Storage and Persistence Layer

- Analysis Focus: Implement persistence for keys, certificates, and audit trails.
- Issues to Address: Need ACID guarantees and audit-friendly schema.
- Implementation Notes: Provide repository interfaces with PostgreSQL/SQLite adapters.
- Deliverables: `internal/ca/storage` packages, migrations, ERD diagrams.
- Validation: Integration tests for both databases; migration dry runs.

### Task 15: CLI Tooling Suite

- Analysis Focus: Provide `cmd/ca` CLI covering bootstrap, issuance, revocation, reporting.
- Issues to Address: Missing operator tooling, inconsistent CLI ergonomics.
- Implementation Notes: Align flag naming with existing cryptoutil conventions.
- Deliverables: CLI commands, PowerShell/Bash examples, completion scripts.
- Validation: Automated CLI tests; manual smoke tests via `scripts/` harness.

### Task 16: Observability and Telemetry

- Analysis Focus: Add tracing, metrics, and logging for CA operations.
- Issues to Address: Lack of visibility for issuance and revocation events.
- Implementation Notes: Export OTLP metrics, create Grafana dashboards, add alert rules.
- Deliverables: Telemetry instrumentation, `deployments/compose/ca-observability.yml`.
- Validation: `go test ./internal/ca/...` with telemetry hooks; dashboard verification.

### Task 17: Security Hardening and Threat Modeling

- Analysis Focus: Conduct threat modeling for CA services, including HSM integration planning.
- Issues to Address: Potential attack surfaces not documented.
- Implementation Notes: Perform STRIDE-based review, add gosec rules, plan HSM adapter.
- Deliverables: Threat model document, security backlog, gosec configuration updates.
- Validation: Security review sign-off; gosec and golangci-lint runs clean.

### Task 18: Compliance and Audit Readiness

- Analysis Focus: Prepare evidence packs for CA/Browser Forum audits.
- Issues to Address: Missing documentation and logging retention policies.
- Implementation Notes: Generate audit trails, retention policies, signing ceremonies.
- Deliverables: Audit evidence folder, policy documents, ceremony checklists.
- Validation: Internal audit dry run; retention automation tests.

### Task 19: Deployment Bundles and Automation

- Analysis Focus: Provide Docker Compose, Kubernetes manifests, and automation scripts.
- Issues to Address: Inconsistent deployment story between environments.
- Implementation Notes: Reuse existing workflow utilities; add CA-specific actions.
- Deliverables: Deployment manifests, CI workflow updates, act profiles.
- Validation: Local `docker compose` smoke test; workflow dry run via `cmd/workflow`.

### Task 20: Final Readiness and Handover

- Analysis Focus: Execute full system regression, document support handoff, and train operators.
- Issues to Address: Ensure sustainability and knowledge transfer.
- Implementation Notes: Produce runbooks, on-call rotations, DR drills.
- Deliverables: Release checklist, operator handbook, training session recordings (references).
- Validation: Final sign-off meeting; DR rehearsal results archived.
