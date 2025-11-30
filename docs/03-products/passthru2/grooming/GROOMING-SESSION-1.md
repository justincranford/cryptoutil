# Passthru2 Grooming Questions (25 Q&A)

**Purpose**: Structured questions to capture decisions and priorities for `passthru2` workstream.
**Created**: 2025-11-30
**Status**: AWAITING ANSWERS

---

## Section 1: Vision & Strategy (Q1-5)

### Q1. Demonstrability Priority
What is the #1 goal for the `passthru2` effort?
- [ ] A. Stabilize product code and tests (KMS, Identity)
- [ ] B. Improve developer experience and demo UX (one-command demos)
- [ ] C. Reorganize repo for infra/product split
- [ ] D. Improve test coverage and CI quality gates
- [ ] E. All of the above

Notes:

---

### Q2. Required Demo Parity
Which demos must be feature-parity (parity meaning same UX/responsiveness)?
- [ ] A. KMS (existing manual demo) only
- [ ] B. Identity only
- [ ] C. Both KMS and Identity must be equal parity
- [ ] D. Integration demo must have equal or better parity
- [ ] E. All demos + JOSE Authority must be consistent

Notes:

---

### Q3. Timeline & Pace
What timeline and cadence do you prefer for the high-priority items?
- [ ] A. Aggressive: Complete in 1-2 weeks
- [ ] B. Moderate: Complete in 2-4 weeks
- [ ] C. Slow: Complete in 4-8 weeks with thorough testing
- [ ] D. Minimal: Only a subset of features (DX + KMS parity)
- [ ] E. Custom: Please specify in Notes

Notes:

---

### Q4. Acceptance for Breaks
What's your tolerance for breaking changes during this phase?
- [ ] A. Zero tolerance (no breaking changes allowed)
- [ ] B. Minor breaks OK - maintain compatibility where possible
- [ ] C. Major reworks allowed (v2 rewrite, migration docs will be prepared)
- [ ] D. If breaking, must be reversible or have migration steps
- [ ] E. Other (explain in Notes)

Notes:

---

### Q5. Demo Audiences
Who are the primary demo audiences for `passthru2` (choose one or more)?
- [ ] A. Contributors / developers
- [ ] B. Interviewers / Recruiters (portfolio)
- [ ] C. Potential users / early adopters
- [ ] D. Internal stakeholders / auditors
- [ ] E. All of the above

Notes:

---

## Section 2: Infra & Deployment (Q6-10)

### Q6. Telemetry Extraction
Should we centralize telemetry into `deployments/telemetry/compose.yml` and require products to opt-in?
- [ ] A. Yes - single source for telemetry
- [ ] B. No - keep telemetry per product for dev ease
- [ ] C. Hybrid - default to shared, product opt-out
- [ ] D. Not sure - prefer discussion

Notes:

---

### Q7. Config and Secrets
Which configuration model should we standardize for product Compose files?
- [ ] A. Product-only config under `deployments/<product>/config/`
- [ ] B. Centralized config folder with product overlays
- [ ] C. Use Docker Secrets for all secrets (recommended)
- [ ] D. Mix: secrets via Docker secrets, defaults in config folder

Notes:

---

### Q8. Compose Profiles
Which compose profiles should be available per product?
- [ ] A. `dev` for local development (no telemetry)
- [ ] B. `demo` for seeded demo data and guidance
- [ ] C. `ci` minimal run for CI tests
- [ ] D. All above

Notes:

---

### Q9. Telemetry as Opt-In
Should a product be able to run without telemetry for ease of local dev?
- [ ] A. Yes, default to no telemetry (local dev focus)
- [ ] B. No, default to telemetry on for demo and CI parity
- [ ] C. Configurable per profile
- [ ] D. Not sure

Notes:

---

### Q10. Secret Management
Do we standardize secret usage in Compose to Docker secrets across all products?
- [ ] A. Yes, standardize to Docker secrets
- [ ] B. No, environment variables are okay for local dev
- [ ] C. Hybrid: secrets in prod, env vars for dev
- [ ] D. Not sure

Notes:

---

## Section 3: Products & Parity (Q11-15)

### Q11. Pre-seeded Demo Accounts
Do we require pre-seeded demo accounts across all products (admin, user, service)?
- [ ] A. Yes - all products must support seeded demo data
- [ ] B. No - only Identity needs seeded demo accounts
- [ ] C. No - KMS can be seeded via integration steps only
- [ ] D. Selective - depends on product

Notes:

---

### Q12. Demo Script Standardization
Which format for demo scripts do you prefer?
- [ ] A. `make demo` (simple Makefile wrapper)
- [ ] B. CLI `./bin/demo` program (go-based orchestration)
- [ ] C. Shell scripts per product (example: `./scripts/demo-kms.sh`)
- [ ] D. Docker Compose wrappers with health checks (recommended)

Notes:

---

### Q13. KMS Parity with Identity
What KMS feature parity should be added to match Identity demo? (choose multiple)
- [ ] A. Demo mode auto-seed (users & clients) - YES
- [ ] B. Swagger UI "Try it out" with demo credentials - YES
- [ ] C. CLI demo orchestration that also seeds keys - YES
- [ ] D. All above

Notes:

---

### Q14. JOSE Authority Demo
Should JOSE Authority be delivered as part of passthru2 or left as future work?
- [ ] A. Include JOSE as a demo (fast extraction from identity issuer)
- [ ] B. Leave JOSE to Phase 3 - not in immediate scope
- [ ] C. Create placeholders and a README only
- [ ] D. Other (specify)

Notes:

---

### Q15. Embedded Identity Option
Should KMS continue to support an embedded Identity option for development?
- [ ] A. Yes - continue embedded mode for dev convenience
- [ ] B. No - prefer external-only to avoid circular deps
- [ ] C. Keep but limit to `demo` profile only
- [ ] D. Other (specify)

Notes:

---

## Section 4: Design & Security (Q16-20)

### Q16. PBKDF2 & FIPS Compliance
Does Identity meet FIPS compliance for hashing and crypto operations?
- [ ] A. Yes - PBKDF2 and other FIPS-approved algorithms used
- [ ] B. No - we need to replace some hashing functions and re-audit
- [ ] C. Partial - some components comply, others do not
- [ ] D. Unsure - need a security audit

Notes:

---

### Q17. Token Management
Which token handling approach should KMS use for validation and caching?
- [ ] A. Introspection endpoint (live validation) + optional caching
- [ ] B. Local JWT validation (by verifying signatures & expiry) - avoids network calls
- [ ] C. Mix: introspection for revocation, local validation for speed
- [ ] D. Other (specify)

Notes:

---

### Q18. Scope Granularity
What scope granularity model would you adopt for KMS? (choose one)
- [ ] A. Coarse granularity (`kms:admin`, `kms:read`, `kms:write`)
- [ ] B. Fine-grained per operation (`kms:encrypt`, `kms:decrypt`, `kms:sign`)
- [ ] C. Hybrid - coarse plus optional finer scopes
- [ ] D. Other (specify)

Notes:

---

### Q19. Audit Logging Requirements
What audit logging level is required for demos and initial implementation?
- [ ] A. Minimal: success/failure events (recommended developer priority)
- [ ] B. Detailed: include payloads for full traceability (sensitive data redaction applied)
- [ ] C. Compliance-level: immutable logs and export (FedRAMP/SOC2) - future work
- [ ] D. Mix: minimal now, compliance later

Notes:

---

### Q20. Database & Transaction Patterns
Which database patterns should be enforced across products for reliability and parallel testability?
- [ ] A. SQLite shared memory with WAL + busy_timeout + MaxOpenConns=5 for GORM packages
- [ ] B. Always require PostgreSQL in CI/Integration testing for realism
- [ ] C. Mixed: SQLite for unit tests, Postgres for integration tests
- [ ] D. Other (specify)

Notes:

---

## Section 5: Tests, CI, & Migration (Q21-25)

### Q21. Coverage Targets & Quality Gates
What coverage targets should be enforced in passthru2 CI?
- [ ] A. Infrastructure: 95%+; Products: 85%+ (per instructions)
- [ ] B. All code: 90%+
- [ ] C. Minimum: 80% for now, iterative improvement
- [ ] D. No hard target - rely on tests & PR reviews

Notes:

---

### Q22. Migration Strategy
What's your migration style preference for moving packages to `internal/infra/` and `internal/product/`?
- [ ] A. Big bang - move all at once and update imports in single PR
- [ ] B. One package at a time - move, run build/test/lint, commit
- [ ] C. Hybrid: move low-risk infra first, then products in batches
- [ ] D. Use aliases for imports and slowly migrate

Notes:

---

### Q23. E2E Test Location & Service Scopes
Where should E2E tests live and what should they test?
- [ ] A. `internal/product/<product>/e2e/` - product-specific E2E
- [ ] B. `internal/product/e2e/` - cross-product E2E
- [ ] C. `test/e2e/` - root-level for orchestration and CI
- [ ] D. All of the above with clear separation

Notes:

---

### Q24. CI Changes
How should CI be adjusted to accommodate the refactor and demo parity work?
- [ ] A. Add demo/compose runs in CI for `demo` profile tests
- [ ] B. Ensure per-product `go test` and coverage remain green per package
- [ ] C. Add matrix runs for SQLite/PG backends in CI
- [ ] D. All of the above

Notes:

---

### Q25. Final Acceptance Criteria for Passthru2
Before closing passthru2, what must be true (select at least 2)?
- [ ] A. KMS and Identity demos both start with `docker compose` and include seeded data
- [ ] B. KMS and Identity both have interactive demo scripts and Swagger UI usable with demo creds
- [ ] C. Integration demo runs and validates token-based auth and scopes
- [ ] D. All product tests pass with coverage targets achieved
- [ ] E. Telemetry extracted to shared compose and secrets standardized

Notes:

---

**Status**: AWAITING YOUR ANSWERS (Change [ ] to [x] as applicable and add notes if needed)

