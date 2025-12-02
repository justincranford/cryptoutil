# Grooming Session 04: Certificate Authority Planning

## Overview

- **Focus Area**: P4 Certificate Authority expansion, CA/Browser Forum compliance, PKI architecture
- **Related Spec Section**: Spec P4: Certificates, docs/05-ca/README.md (20 tasks)
- **Prerequisites**: Sessions 01-03 completed, understanding of X.509, PKI concepts

---

## Questions

### Q1: How many tasks are defined in the CA expansion plan?

A) 10 tasks
B) 15 tasks
C) 20 tasks
D) 25 tasks

**Answer**: C
**Explanation**: The CA expansion plan defines 20 tasks covering domain charter through final handover.

---

### Q2: Which compliance standard is primary for the CA subsystem?

A) ISO 27001
B) SOC 2
C) CA/Browser Forum Baseline Requirements
D) PCI DSS

**Answer**: C
**Explanation**: CA/Browser Forum Baseline Requirements and RFC 5280 are the primary compliance standards for the CA subsystem.

---

### Q3: What configuration approach is specified for the CA?

A) Environment variables
B) Database-driven configuration
C) YAML-driven configuration
D) Hard-coded defaults

**Answer**: C
**Explanation**: Deterministic YAML schemas are specified for crypto, subject, and certificate policies.

---

### Q4: Which task defines the scope and compliance obligations?

A) Task 1: Domain Charter
B) Task 2: Configuration Schema
C) Task 17: Security Hardening
D) Task 18: Compliance

**Answer**: A
**Explanation**: Task 1: Domain Charter captures target capabilities, compliance obligations, and non-goals.

---

### Q5: What validation approach is recommended for the configuration schema?

A) Manual review only
B) JSON Schema or Cue for validation
C) XML Schema validation
D) No formal validation

**Answer**: B
**Explanation**: Task 2 specifies JSON Schema or Cue for configuration validation with documented defaults.

---

### Q6: Which cryptographic algorithms must the CA crypto providers support?

A) RSA only
B) RSA and ECDSA only
C) RSA, ECDSA, EdDSA, HMAC, and future PQC stubs
D) Any algorithm available in Go stdlib

**Answer**: C
**Explanation**: Task 3 requires support for RSA, ECDSA, EdDSA, HMAC, and future PQC (Post-Quantum Cryptography) stubs.

---

### Q7: What is the purpose of the Subject Profile Engine?

A) Manage user authentication
B) Implement subject template resolution with fields, SANs, constraints
C) Store subject certificates
D) Validate subject names only

**Answer**: B
**Explanation**: Task 4: Subject Profile Engine implements subject template resolution including fields, SANs, and constraints.

---

### Q8: How many certificate profile archetypes must be supported?

A) 5+ profiles
B) 10+ profiles
C) 15+ profiles
D) 20+ profiles

**Answer**: D
**Explanation**: Task 5 requires support for 20+ profile archetypes for different certificate types.

---

### Q9: What type of Root CA bootstrap is planned?

A) Online root CA only
B) Offline root CA with CLI and library support
C) Cloud-hosted root CA
D) HSM-only root CA

**Answer**: B
**Explanation**: Task 6 provides CLI and library support for offline root CA bootstrap.

---

### Q10: What validation method is specified for certificate chain verification?

A) Internal Go crypto only
B) OpenSSL or Go crypto verification
C) External CA validation service
D) Manual certificate inspection

**Answer**: B
**Explanation**: Task 7 integration tests verify chain building and validation using openssl or Go crypto.

---

### Q11: What lifecycle management features are planned for Issuing CAs?

A) Manual rotation only
B) Rotation scheduler, status reporting via OTEL
C) No lifecycle management
D) External rotation service

**Answer**: B
**Explanation**: Task 8 includes rotation scheduler and status reporting via OpenTelemetry.

---

### Q12: What API design approach is specified for the Enrollment API?

A) SOAP/XML API
B) OpenAPI-first with generated handlers
C) GraphQL API
D) gRPC only

**Answer**: B
**Explanation**: Task 9 uses OpenAPI-first approach with generated handlers for certificate enrollment.

---

### Q13: What revocation mechanisms are included in Task 10?

A) CRL only
B) OCSP only
C) CRL, delta CRLs, and OCSP
D) Certificate transparency only

**Answer**: C
**Explanation**: Task 10 includes CRL generation, delta CRLs, and OCSP responders with caching.

---

### Q14: What standard does the Time-Stamping service implement?

A) RFC 2560
B) RFC 3161
C) RFC 5280
D) RFC 6960

**Answer**: B
**Explanation**: Task 11 implements RFC 3161-like endpoints for timestamp tokens.

---

### Q15: What is the purpose of the Registration Authority (RA)?

A) Generate certificates directly
B) Request validation, identity proofing, approval routing
C) Revoke certificates only
D) Store audit logs

**Answer**: B
**Explanation**: Task 12: RA handles request validation, identity proofing, and approval routing.

---

### Q16: Which certificate types are included in the 20+ profile library?

A) TLS server/client only
B) Root, intermediate, TLS, S/MIME, code signing, and many more
C) Only code signing certificates
D) Only email certificates

**Answer**: B
**Explanation**: Task 13 covers root, intermediate, TLS server/client, S/MIME, code signing, document signing, VPN, IoT, SAML, JWT, OCSP, RA, TSA, CT log, ACME, SCEP, EST, CMP, enterprise custom profiles.

---

### Q17: What persistence requirements apply to CA storage?

A) File-based storage only
B) ACID guarantees with audit-friendly schema
C) In-memory storage only
D) No specific requirements

**Answer**: B
**Explanation**: Task 14 requires ACID guarantees and audit-friendly schema with PostgreSQL/SQLite adapters.

---

### Q18: What is included in the CLI tooling suite?

A) Bootstrap only
B) Bootstrap, issuance, revocation, reporting
C) Reporting only
D) No CLI planned

**Answer**: B
**Explanation**: Task 15 covers `cmd/ca` CLI with bootstrap, issuance, revocation, and reporting commands.

---

### Q19: What observability features are specified for CA operations?

A) Logging only
B) OTLP metrics, Grafana dashboards, alert rules
C) No observability planned
D) External monitoring only

**Answer**: B
**Explanation**: Task 16 includes OTLP metrics export, Grafana dashboards, and alert rules.

---

### Q20: What threat modeling approach is specified?

A) PASTA
B) STRIDE-based review
C) DREAD
D) No formal threat modeling

**Answer**: B
**Explanation**: Task 17 requires STRIDE-based threat modeling review with HSM adapter planning.

---

### Q21: What compliance documentation is required for Task 18?

A) Internal audit only
B) Audit evidence folder, policy documents, ceremony checklists
C) No documentation required
D) External audit report only

**Answer**: B
**Explanation**: Task 18 requires audit evidence folder, policy documents, and ceremony checklists for CA/Browser Forum audits.

---

### Q22: What deployment options are specified in Task 19?

A) Docker only
B) Docker Compose and Kubernetes manifests
C) Bare metal only
D) Cloud-only deployment

**Answer**: B
**Explanation**: Task 19 provides Docker Compose, Kubernetes manifests, and automation scripts.

---

### Q23: What is included in Task 20: Final Readiness?

A) Code review only
B) Full regression, handoff documentation, operator training
C) Security review only
D) Performance testing only

**Answer**: B
**Explanation**: Task 20 includes full system regression, support handoff documentation, and operator training.

---

### Q24: What is the minimum certificate serial number entropy?

A) 32 bits
B) 64 bits
C) 128 bits
D) 256 bits

**Answer**: B
**Explanation**: CA/Browser Forum requires minimum 64 bits from CSPRNG for serial numbers.

---

### Q25: What constraints apply to certificate serial numbers?

A) Must be sequential
B) Non-sequential, >0, <2^159
C) Any positive integer
D) Must be even numbers

**Answer**: B
**Explanation**: Serial numbers must be non-sequential, greater than 0, and less than 2^159.

---

### Q26: What is the maximum validity period for subscriber certificates?

A) 90 days
B) 180 days
C) 398 days
D) 825 days

**Answer**: C
**Explanation**: CA/Browser Forum limits subscriber certificate validity to 398 days (post-2020-09-01).

---

### Q27: Which RFC defines X.509 certificate format?

A) RFC 5246
B) RFC 5280
C) RFC 6125
D) RFC 7469

**Answer**: B
**Explanation**: RFC 5280 defines Internet X.509 PKI Certificate and CRL Profile.

---

### Q28: What priority level is assigned to the CA expansion?

A) LOW
B) MEDIUM
C) HIGH
D) CRITICAL

**Answer**: C
**Explanation**: P4: Certificates is marked HIGH priority in the product table.

---

### Q29: What infrastructure dependencies are identified for P4?

A) I1, I2, I3
B) I6, I7, I11
C) I4, I5, I6
D) I8, I9, I10

**Answer**: B
**Explanation**: P4 Certificates depends on I6 (crypto), I7 (database), and I11 (auditing) infrastructure.

---

### Q30: What golden file testing approach is specified for certificates?

A) No golden file testing
B) DER comparisons for profile validation
C) PEM comparisons only
D) Hash comparisons only

**Answer**: B
**Explanation**: Task 5 uses golden files for DER comparisons in certificate profile validation.

---

### Q31: What cross-signing support is required for Intermediate CAs?

A) No cross-signing support
B) Cross-signing with path length constraints
C) Cross-signing without constraints
D) External cross-signing only

**Answer**: B
**Explanation**: Task 7 includes cross-signing support with path length constraints and emergency rollover.

---

### Q32: What queue mechanism is specified for RA workflows?

A) No queue mechanism
B) Queue-backed workflow with policy-driven approvals
C) Synchronous processing only
D) External queue service only

**Answer**: B
**Explanation**: Task 12 specifies queue-backed workflow with policy-driven approvals for RA.

---

### Q33: What deliverable location is specified for CA configuration schema?

A) `internal/ca/config.yaml`
B) `docs/ca/config-schema.yaml`
C) `configs/ca/schema.yaml`
D) `api/ca/config.yaml`

**Answer**: B
**Explanation**: Task 2 deliverable is `docs/ca/config-schema.yaml` with validation utilities.

---

### Q34: What crypto provider implementations are required initially?

A) HSM only
B) Memory and filesystem implementations
C) Cloud KMS only
D) Software TPM only

**Answer**: B
**Explanation**: Task 3 requires memory and filesystem implementations, with HSM as future planning.

---

### Q35: What test fixtures location is specified for subject profiles?

A) `test/fixtures/subjects/`
B) `testdata/ca/subjects/`
C) `configs/ca/subjects/`
D) `internal/ca/test/subjects/`

**Answer**: C
**Explanation**: Task 4 uses fixtures from `configs/ca/subjects/` for subject profile testing.

---

### Q36: What sealed storage strategy is required for Root CA?

A) Encrypted database only
B) Sealed storage for offline root keys
C) HSM-only storage
D) No special storage requirements

**Answer**: B
**Explanation**: Task 6 includes sealed storage strategy for offline root CA keys.

---

### Q37: What conformance testing is specified for revocation services?

A) No conformance testing
B) Conformance tests with OpenSSL
C) Internal testing only
D) Manual conformance verification

**Answer**: B
**Explanation**: Task 10 includes conformance tests with OpenSSL for revocation services.

---

### Q38: What profile catalog documentation is required?

A) Inline code comments only
B) `docs/ca/profiles.md` with YAML samples
C) No documentation required
D) README only

**Answer**: B
**Explanation**: Task 13 requires profile catalog in `docs/ca/profiles.md` with YAML samples.

---

### Q39: What database migration approach is specified?

A) Manual SQL scripts
B) Repository interfaces with PostgreSQL/SQLite adapters and migrations
C) ORM auto-migration only
D) No migrations required

**Answer**: B
**Explanation**: Task 14 requires repository interfaces with proper migrations and ERD diagrams.

---

### Q40: What CLI completion scripts are required?

A) Bash only
B) PowerShell and Bash
C) No completion scripts
D) Zsh only

**Answer**: B
**Explanation**: Task 15 includes PowerShell/Bash examples and completion scripts.

---

### Q41: What deployment composition file is specified for CA observability?

A) `compose.ca.yml`
B) `deployments/compose/ca-observability.yml`
C) `docker-compose.ca.yml`
D) `observability/ca.yml`

**Answer**: B
**Explanation**: Task 16 deliverable includes `deployments/compose/ca-observability.yml`.

---

### Q42: What gosec configuration updates are required?

A) No gosec updates
B) Security-specific gosec rules for CA code
C) Disable gosec for CA
D) External security scanner only

**Answer**: B
**Explanation**: Task 17 includes gosec configuration updates for CA-specific security rules.

---

### Q43: What retention automation is required for audit logs?

A) Manual retention only
B) Retention policy automation with tests
C) No retention requirements
D) External retention service

**Answer**: B
**Explanation**: Task 18 requires retention automation tests for audit logs.

---

### Q44: What local testing validation is specified for deployments?

A) CI/CD only testing
B) Local docker compose smoke test
C) No local testing
D) Manual deployment only

**Answer**: B
**Explanation**: Task 19 validation includes local docker compose smoke test.

---

### Q45: What disaster recovery validation is required?

A) No DR validation
B) DR rehearsal with archived results
C) Documentation only
D) External DR service

**Answer**: B
**Explanation**: Task 20 requires DR (disaster recovery) rehearsal results to be archived.

---

### Q46: What is the OCSP responder caching requirement?

A) No caching
B) OCSP responder with caching
C) External caching only
D) Cache invalidation only

**Answer**: B
**Explanation**: Task 10 specifies OCSP responder with caching for performance.

---

### Q47: What emergency rollover support is required?

A) No emergency procedures
B) Emergency rollover for Intermediate CA
C) Emergency procedures only for Root CA
D) External rollover service

**Answer**: B
**Explanation**: Task 7 includes emergency rollover support for Intermediate CA provisioning.

---

### Q48: What admin UI is specified for RA?

A) Full-featured admin UI
B) Admin UI stubs for future development
C) No admin UI planned
D) Command-line only

**Answer**: B
**Explanation**: Task 12 includes admin UI stubs as a deliverable for future development.

---

### Q49: What act profiles are required for deployment testing?

A) No act profiles
B) Act profiles for workflow dry runs
C) External testing only
D) Manual workflow execution

**Answer**: B
**Explanation**: Task 19 includes act profiles for local workflow dry run testing.

---

### Q50: What is the Phase 4 timeline in the implementation plan?

A) 1-2 weeks
B) 2-4 weeks
C) 4-8 weeks
D) 8-12 weeks

**Answer**: C
**Explanation**: Phase 4: Certificate Authority Foundation has a 4-8 week duration covering Tasks 1-10.

---

## Session Summary

**Topics Covered**:

- CA expansion plan (20 tasks)
- CA/Browser Forum Baseline Requirements
- X.509 and RFC 5280 compliance
- Certificate profiles (20+ types)
- CA hierarchy (Root, Intermediate, Issuing)
- Enrollment API and revocation services
- Security hardening and threat modeling
- Audit and compliance documentation
- Deployment and operational readiness

**Next Session**: GROOMING-SESSION-05 - Infrastructure and Quality Assurance
