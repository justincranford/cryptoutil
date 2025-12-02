# Speckit Passthru02 Grooming Session 01: Constitution & Spec Refinement

**Purpose**: Structured questions to refine constitution principles, product definitions, and strategic priorities for cryptoutil Spec Kit.
**Created**: 2025-12-02
**Status**: AWAITING ANSWERS

---

## Instructions

Select your answer by changing `[ ]` to `[x]` for each question. Add comments in the "Notes" field if needed. Multiple selections allowed where indicated.

---

## Section 1: Constitution Principles (Q1-8)

### Q1. FIPS 140-3 Strictness

The constitution mandates FIPS 140-3 compliance. How strict should enforcement be?

- [ ] A. Strict: ONLY FIPS-approved algorithms, no exceptions, compile-time enforcement
- [ ] B. Default: FIPS-approved by default, allow non-FIPS via explicit opt-in flag for testing
- [ ] C. Flexible: FIPS-approved recommended, allow non-FIPS algorithms when justified
- [ ] D. Configurable: Runtime toggle between FIPS-strict and FIPS-relaxed modes

**Notes**:

---

### Q2. Evidence-Based Completion Threshold

What is the minimum evidence required to mark a task "complete"?

- [ ] A. Code compiles (`go build ./...` clean)
- [ ] B. Code compiles + linting passes (`golangci-lint run` clean)
- [ ] C. Code compiles + linting + tests pass
- [ ] D. Code compiles + linting + tests pass + coverage maintained + E2E demo works

**Notes**:

---

### Q3. Coverage Target Adjustment

Current targets are 80% production, 85% infrastructure, 95% utility. Are these appropriate?

- [ ] A. Too high - lower to 70%/75%/85%
- [ ] B. Just right - keep 80%/85%/95%
- [ ] C. Too low - raise to 85%/90%/98%
- [ ] D. Eliminate fixed targets - use risk-based coverage decisions

**Notes**:

---

### Q4. File Size Limits

Current limits: 300 (soft), 400 (medium), 500 (hard). Are these appropriate?

- [ ] A. Too restrictive - raise to 400/500/600
- [ ] B. Just right - keep 300/400/500
- [ ] C. Too permissive - lower to 200/300/400
- [ ] D. Replace line counts with cyclomatic complexity metrics

**Notes**:

---

### Q5. KMS Key Hierarchy Depth

The constitution specifies 4 layers: Unseal → Root → Intermediate → Content. Is this sufficient?

- [ ] A. Too complex - simplify to 3 layers (Unseal → Master → Content)
- [ ] B. Just right - keep 4 layers
- [ ] C. Need more - add Tenant layer (Unseal → Root → Tenant → Intermediate → Content)
- [ ] D. Configurable - allow 3-5 layers based on deployment complexity

**Notes**:

---

### Q6. Product Architecture Split

Constitution defines P1-P4 products. Should products share code or be independent?

- [ ] A. Monolith: All products in single deployable binary
- [ ] B. Shared: Products share infrastructure, deployed separately
- [ ] C. Independent: Products are fully independent, no shared code
- [ ] D. Hybrid: P1 (JOSE) embedded in all, P2-P4 deploy separately

**Notes**:

---

### Q7. Secret Management Strategy

Constitution mandates Docker/Kubernetes secrets only. Is this too restrictive?

- [ ] A. Too restrictive - allow environment variables for local development
- [ ] B. Just right - Docker/K8s secrets only, even locally
- [ ] C. Add vault support - HashiCorp Vault, AWS Secrets Manager, etc.
- [ ] D. Tiered: Secrets for production, env vars for dev, config files for tests

**Notes**:

---

### Q8. Governance Decision Authority

Who has authority to change constitution principles?

- [ ] A. Only project owner (you) can modify constitution
- [ ] B. Major changes require PR with justification, minor clarifications allowed
- [ ] C. Constitution is immutable - create new version for breaking changes
- [ ] D. Living document - update freely as understanding evolves

**Notes**:

---

## Section 2: Product Specification (Q9-16)

### Q9. P1 JOSE Scope

P1 (JOSE) is marked as ✅ Implemented. What changes are needed?

- [ ] A. Complete - no changes needed
- [ ] B. Add algorithm agility - runtime algorithm selection
- [ ] C. Add key import/export - interoperability with external systems
- [ ] D. Add key wrapping APIs - wrap/unwrap operations

**Notes**:

---

### Q10. P2 Identity Priority

P2 (Identity) has partial implementation. Which area needs most focus?

- [ ] A. Login/Consent UI - user-facing flows
- [ ] B. Token lifecycle - revocation, cleanup, introspection
- [ ] C. MFA completion - TOTP, Passkey, Recovery codes
- [ ] D. Security hardening - secret hashing, rate limiting, audit logging

**Notes**:

---

### Q11. P2 Identity Authentication Methods

Which client authentication methods should be prioritized?

- [ ] A. Basic (client_secret_basic, client_secret_post) - already working
- [ ] B. JWT-based (client_secret_jwt, private_key_jwt) - partially implemented
- [ ] C. mTLS (tls_client_auth, self_signed_tls_client_auth) - not implemented
- [ ] D. All methods should have equal priority

**Notes**:

---

### Q12. P2 Identity MFA Priority

Which MFA factors should be completed first? (Select top 2)

- [ ] A. TOTP - Time-based OTP (marked ✅ Working)
- [ ] B. Passkey - WebAuthn/FIDO2 (marked ✅ Working)
- [ ] C. Email OTP - (marked ⚠️ Partial)
- [ ] D. SMS OTP - (marked ⚠️ Partial)
- [ ] E. Recovery Codes - (marked ❌ Not Implemented)
- [ ] F. Hardware Security Keys - (marked ❌ Not Implemented)

**Notes**:

---

### Q13. P3 KMS Feature Completeness

P3 (KMS) has basic operations. What should be added?

- [ ] A. Update/Delete operations for ElasticKey
- [ ] B. Import/Revoke operations for MaterialKey
- [ ] C. Key rotation automation
- [ ] D. Multi-tenant isolation
- [ ] E. All of the above

**Notes**:

---

### Q14. P3 KMS Authentication Strategy

How should KMS authenticate requests?

- [ ] A. No authentication - trust network isolation
- [ ] B. API key - simple bearer token
- [ ] C. OAuth 2.1 - federate to Identity (P2)
- [ ] D. mTLS - client certificate authentication
- [ ] E. Configurable - support multiple methods

**Notes**:

---

### Q15. P4 CA Priority

P4 (Certificates) is PLANNED. What's the implementation priority?

- [ ] A. High - start CA foundation immediately (Phase 4 per plan)
- [ ] B. Medium - complete P2/P3 first, then CA
- [ ] C. Low - CA is nice-to-have, defer indefinitely
- [ ] D. Reconsider - use existing CA (Let's Encrypt, Vault PKI) instead

**Notes**:

---

### Q16. P4 CA Scope

If P4 is implemented, what scope is appropriate?

- [ ] A. Internal-only - PKI for internal services, not public TLS
- [ ] B. Private CA - enterprise PKI for org-internal certificates
- [ ] C. Public CA - CA/Browser Forum compliant for public TLS
- [ ] D. Hybrid - internal CA first, public CA compliance later

**Notes**:

---

## Section 3: Infrastructure Priorities (Q17-22)

### Q17. Database Backend Priority

Which database backend should be primary?

- [ ] A. PostgreSQL only - enterprise-grade, remove SQLite
- [ ] B. SQLite only - simplicity, embed in binary
- [ ] C. Both equally - PostgreSQL for production, SQLite for dev/test
- [ ] D. Add more - MySQL, CockroachDB, etc.

**Notes**:

---

### Q18. Telemetry Completeness

Current telemetry uses OTLP → Collector → Grafana. Is this sufficient?

- [ ] A. Sufficient - current stack is adequate
- [ ] B. Add Prometheus direct - for environments without OTLP
- [ ] C. Add structured logging - JSON logs for log aggregation
- [ ] D. Add distributed tracing - full request tracing across services

**Notes**:

---

### Q19. CI/CD Pipeline Completeness

Which CI/CD workflows need improvement? (Select all that apply)

- [ ] A. ci-quality - linting, formatting, builds
- [ ] B. ci-coverage - test coverage
- [ ] C. ci-dast - dynamic security testing
- [ ] D. ci-e2e - end-to-end testing
- [ ] E. ci-load - performance/load testing

**Notes**:

---

### Q20. Load Testing Priority

Gatling load tests exist in test/load/. How important are they?

- [ ] A. Critical - must pass before any release
- [ ] B. Important - run regularly, investigate regressions
- [ ] C. Nice-to-have - run occasionally, not blocking
- [ ] D. Not needed - remove load testing infrastructure

**Notes**:

---

### Q21. Demo Experience Priority

Which demo experience improvements are most important?

- [ ] A. One-command demos (`go run ./cmd/demo all`)
- [ ] B. Swagger UI with pre-filled credentials
- [ ] C. Docker Compose with health checks
- [ ] D. Interactive tutorials/walkthroughs
- [ ] E. All of the above

**Notes**:

---

### Q22. Documentation Strategy

How should documentation be organized?

- [ ] A. Current: README.md + docs/README.md only
- [ ] B. Add ADRs: Architecture Decision Records in docs/adr/
- [ ] C. Add API guides: Per-product API documentation
- [ ] D. Add runbooks: Operational procedures in docs/runbooks/
- [ ] E. All of the above

**Notes**:

---

## Section 4: Risk & Prioritization (Q23-25)

### Q23. Biggest Risk to Project Success

What is the single biggest risk to cryptoutil success?

- [ ] A. Complexity - too many products, too ambitious
- [ ] B. Security - cryptographic implementation errors
- [ ] C. Adoption - no users, no contributors
- [ ] D. Maintenance - solo project, burnout risk
- [ ] E. Competition - existing solutions (Vault, Keycloak) are good enough

**Notes**:

---

### Q24. Timeline Realism

The plan shows 6 phases over 20+ weeks. Is this realistic?

- [ ] A. Too aggressive - double the estimates
- [ ] B. About right - achievable with focused effort
- [ ] C. Too conservative - can move faster
- [ ] D. Not sure - need to complete Phase 1 first to calibrate

**Notes**:

---

### Q25. Investment Allocation

If you had 100 hours to invest, how would you allocate them?

| Area | Hours |
|------|-------|
| P2 Identity completion | __ |
| P3 KMS stabilization | __ |
| P4 CA foundation | __ |
| Infrastructure improvements | __ |
| Documentation | __ |
| Testing/quality | __ |

**Notes**:

---

## Summary & Next Steps

After completing this grooming session:

1. Review your answers for consistency
2. Identify any conflicting priorities
3. Update constitution.md if principles need refinement
4. Update spec.md if product scope changes
5. Update plan.md if priorities shift
6. Share answers with Copilot for next iteration

---

*Session Created: 2025-12-02*
*Expected Completion: [DATE]*
