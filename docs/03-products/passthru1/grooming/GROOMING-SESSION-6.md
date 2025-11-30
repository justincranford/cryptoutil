# Grooming Session 6: Implementation Details

## Purpose

Session 6 focuses on implementation details for the aggressive refactoring plan - package migration order, identity code extraction, JOSE Authority implementation, E2E test organization, and demo orchestration.

**Date**: November 30, 2025
**Status**: ANSWERED

---

## Section E: Package Migration Order (Q26-31)

### Q26. First Package to Move

Which package should be moved FIRST to establish the pattern?

- [x] A. `internal/common/magic/` → `internal/infra/magic/` (smallest, lowest risk)
- [x] B. `internal/common/apperr/` → `internal/infra/apperr/` (widely used)
- [ ] C. `internal/common/config/` → `internal/infra/config/` (foundational)
- [x] D. `internal/common/crypto/` → `internal/infra/crypto/` (core functionality)
- [x] E. `internal/common/telemetry/` → `internal/infra/telemetry/` (cross-cutting)

**Notes**: Move multiple packages together - magic, apperr, crypto, telemetry as first batch.

---

### Q27. KMS Migration Sequence

For moving KMS to `internal/product/kms/`, what order?

- [ ] A. server → client → barrier (dependency order)
- [x] B. barrier → server → client (bottom-up)
- [ ] C. All at once (single commit)
- [ ] D. repository → businesslogic → handler → application (layer order)
- [ ] E. Start with server/application/, expand from there

**Notes**: Bottom-up approach starting with barrier.

---

### Q28. Identity Migration Sequence

For moving Identity to `internal/product/identity/`, what order?

- [ ] A. Move all at once, fix in place
- [x] B. Move working packages first, broken packages later
- [ ] C. Move by dependency order (domain → repository → service → handler)
- [ ] D. Extract infra candidates first, move remaining to product
- [ ] E. Move authz/ first (OAuth core), then supporting packages

**Notes**: Working packages first for stability.

---

### Q29. Import Alias Updates

When updating import paths, should aliases change?

- [x] A. Keep existing aliases (e.g., `cryptoutilMagic`)
- [x] B. Update to reflect new structure (e.g., `infraMagic`, `productKMS`)
- [ ] C. Simplify aliases during migration
- [ ] D. No aliases - use full package names
- [ ] E. Different convention for infra vs product

**Notes**: Keep existing aliases but can update to reflect new structure.

---

### Q30. Circular Dependency Prevention

Strategy to prevent circular imports during migration?

- [ ] A. Strict layering: infra → product (never reverse)
- [ ] B. Interface packages to break cycles
- [ ] C. Dependency injection for cross-cutting concerns
- [ ] D. All of the above
- [x] E. Address cycles as they appear

**Notes**: Address issues as they appear rather than over-engineering upfront.

---

### Q31. Migration Verification

After each package move, what verification?

- [ ] A. `go build ./...` only
- [ ] B. `go build ./...` + `go test ./...`
- [x] C. Above + `golangci-lint run`
- [ ] D. Above + specific package coverage check
- [ ] E. Full CI pipeline equivalent

**Notes**: Build + test + lint after each migration step.

---

## Section F: Identity Code Extraction Candidates (Q32-37)

### Q32. Identity `apperr/` Package

`internal/identity/apperr/` duplicates `internal/common/apperr/`. Action:

- [ ] A. Delete identity version, use common
- [ ] B. Merge identity additions into common
- [x] C. Keep both, consolidate later
- [ ] D. Compare and pick best implementation
- [ ] E. Need to analyze differences first

**Notes**: Consolidation can happen after migration stabilizes.

---

### Q33. Identity `config/` Package

`internal/identity/config/` vs `internal/common/config/`. Action:

- [ ] A. Identity config is product-specific, keep in product
- [x] B. Extract common patterns to infra, keep product-specific in product
- [ ] C. Merge into unified config system in infra
- [ ] D. Keep separate - different enough to warrant both
- [ ] E. Need to analyze differences first

**Notes**: Common patterns to infra, product-specific stays in product.

---

### Q34. Identity `magic/` Package

`internal/identity/magic/` vs `internal/common/magic/`. Action:

- [ ] A. Merge all magic values into single infra package
- [ ] B. Keep product-specific magic in products
- [x] C. Shared magic → infra, product magic → product
- [ ] D. Consolidate naming, split by domain
- [ ] E. Need to analyze what's in each

**Notes**: Split by scope - shared values in infra, product-specific in product.

---

### Q35. Identity `domain/` Models

Which Identity domain models might be reusable?

- [x] A. None - all Identity-specific
- [ ] B. User, Session - generic auth concepts
- [ ] C. Token, Key - shared with KMS/JOSE
- [ ] D. Client, Scope - OAuth generic
- [x] E. Need to analyze for reuse candidates

**Notes**: Likely Identity-specific, but worth analyzing.

---

### Q36. Identity Repository Patterns

`internal/identity/repository/` has ORM patterns. Reusable?

- [ ] A. Extract generic repository interfaces to infra
- [x] B. Extract GORM helpers/utilities to infra
- [ ] C. Keep all in identity - too product-specific
- [ ] D. Transaction patterns could be shared
- [x] E. Need to analyze what's generic vs specific

**Notes**: GORM helpers likely reusable, need analysis.

---

### Q37. Identity Test Utilities

`internal/identity/test/` has test helpers. Action:

- [ ] A. Move all to `internal/infra/testing/`
- [ ] B. Keep identity-specific in product, move generic to infra
- [ ] C. Merge with `internal/common/testutil/`
- [ ] D. Each product should have own test utilities
- [x] E. Need to analyze what's reusable

**Notes**: Analyze and determine reusability.

---

## Section G: JOSE Authority Implementation (Q38-43)

### Q38. JOSE Authority Initial Source

Where does initial JOSE Authority code come from?

- [ ] A. Extract from `internal/common/crypto/jose/`
- [x] B. Extract from `internal/identity/issuer/`
- [ ] C. Extract from `internal/identity/jwks/`
- [ ] D. Combination of B and C
- [ ] E. New implementation (existing code not suitable)

**Notes**: Identity issuer has good foundation for JOSE Authority.

---

### Q39. JOSE Authority vs Identity Issuer

Relationship between JOSE Authority and Identity's token issuing?

- [x] A. Identity uses JOSE Authority for all token operations
- [ ] B. Identity has own issuer, JOSE Authority is separate product
- [ ] C. JOSE Authority replaces Identity issuer entirely
- [ ] D. Shared interfaces, separate implementations
- [ ] E. Need to understand current Identity issuer first

**Notes**: Identity delegates token operations to JOSE Authority.

---

### Q40. JOSE Authority Templates/Profiles

You mentioned "templates/profiles like a CA." Examples:

- [ ] A. Access Token profile, ID Token profile, Refresh Token profile
- [ ] B. Service-to-service JWS, User-facing JWT, Encrypted JWE
- [ ] C. Short-lived tokens, Long-lived tokens, One-time tokens
- [x] D. All of the above
- [ ] E. Different categorization (explain in notes)

**Notes**: Comprehensive template/profile support.

---

### Q41. JOSE Authority Key Management

How does JOSE Authority manage signing keys?

- [ ] A. Own key storage (separate from KMS)
- [ ] B. Delegates to KMS for all key operations
- [ ] C. Uses KMS for storage, own logic for JOSE operations
- [x] D. Configurable: standalone or KMS-backed
- [ ] E. Embedded keys only (no external storage)

**Notes**: Flexibility - can run standalone or with KMS backing.

---

### Q42. JOSE Authority JWKS Endpoint

JWKS endpoint ownership:

- [x] A. JOSE Authority owns JWKS, Identity consumes it
- [ ] B. Each product (Identity, JOSE Auth) has own JWKS endpoint
- [ ] C. Shared JWKS infrastructure in infra
- [x] D. JOSE Authority is THE JWKS server, others delegate
- [ ] E. Need to understand current JWKS implementation

**Notes**: JOSE Authority is the authoritative JWKS server.

---

### Q43. JOSE Authority Demo Scenario

What's the JOSE Authority demo?

- [ ] A. Issue tokens for multiple audiences (KMS, Identity, external)
- [ ] B. JWKS rotation demonstration
- [ ] C. Multi-tenant token issuance (different PKC domains)
- [ ] D. Token verification/validation service
- [x] E. All of the above

**Notes**: Comprehensive demo covering all scenarios.

---

## Section H: E2E Test Organization (Q44-47)

### Q44. Product E2E Test Scope

What should each product's E2E tests cover?

- [x] A. Just that product in isolation
- [ ] B. Product + its infra dependencies
- [ ] C. Product + mock external dependencies
- [ ] D. Product + real external dependencies (Docker)
- [ ] E. Varies by product

**Notes**: Product isolation for E2E tests.

---

### Q45. Cross-Product E2E Location

Where do tests that span multiple products go?

- [ ] A. `internal/e2e/` (new top-level)
- [ ] B. `internal/infra/testing/e2e/`
- [ ] C. `test/e2e/` (root level)
- [ ] D. `internal/product/integration/` (integration as product)
- [ ] E. In the "primary" product being tested

**Notes**: `internal/product/e2e/` for cross-product tests.

---

### Q46. E2E Test Dependencies

How do E2E tests get their dependencies?

- [x] A. Docker Compose per product
- [ ] B. Shared Docker Compose for all products
- [ ] C. TestContainers (programmatic)
- [ ] D. Mix: simple deps inline, complex deps compose
- [ ] E. In-memory/mock for speed, Docker for CI

**Notes**: Docker Compose per product for consistency.

---

### Q47. E2E Test Naming Convention

How to name E2E test files?

- [x] A. `*_e2e_test.go` (current convention)
- [ ] B. `*_integration_test.go`
- [ ] C. In `e2e/` subdirectory with regular `_test.go`
- [ ] D. Build tag based (`// +build e2e`)
- [ ] E. Multiple conventions depending on scope

**Notes**: Keep existing `*_e2e_test.go` convention.

---

## Section I: Demo Orchestration (Q48-50)

### Q48. Per-Product Demo Executable

What does the per-product demo executable do?

- [ ] A. Just `docker compose up/down` wrapper
- [x] B. Compose + health checks + sample API calls
- [ ] C. Compose + seed data + interactive prompts
- [ ] D. Full guided walkthrough with explanations
- [ ] E. Configurable verbosity levels

**Notes**: Compose management with health checks and sample API calls.

---

### Q49. Cross-Product Demo Configuration

How is delegation between products configured?

- [x] A. YAML config file specifying product relationships
- [ ] B. Environment variables passed between compose files
- [ ] C. Service discovery (products find each other)
- [ ] D. Hardcoded for demo (not production-ready)
- [ ] E. CLI flags to demo orchestrator

**Notes**: YAML configuration for product relationships.

---

### Q50. Demo Success Criteria

What must work for demo to be "complete"?

- [ ] A. All API endpoints return expected responses
- [ ] B. Full auth flow: token → API call → authorized response
- [ ] C. Error cases handled gracefully with clear messages
- [ ] D. Can run offline (no external dependencies)
- [x] E. All of the above

**Notes**: Complete demo requires all criteria met.

---

## Analysis Summary from Session 6

### Key Implementation Decisions

| Area | Decision |
|------|----------|
| **First Packages** | magic, apperr, crypto, telemetry together |
| **KMS Migration** | Bottom-up (barrier → server → client) |
| **Identity Migration** | Working packages first |
| **Import Aliases** | Keep existing, update incrementally |
| **Circular Deps** | Address as they appear |
| **Verification** | build + test + lint after each step |
| **Identity Duplicates** | Keep both, consolidate later |
| **JOSE Authority** | Extract from identity/issuer |
| **E2E Location** | `internal/product/e2e/` for cross-product |
| **Demo Config** | YAML-based product relationships |

### JOSE Authority Vision Confirmed

- Source: Extract from `internal/identity/issuer/`
- Relationship: Identity delegates token ops to JOSE Authority
- Profiles: Access, ID, Refresh tokens; S2S JWS; User JWT; Encrypted JWE
- Key Management: Configurable standalone or KMS-backed
- JWKS: JOSE Authority is THE authoritative JWKS server

### E2E Test Strategy

- Product E2E: Isolation tests per product
- Cross-Product E2E: `internal/product/e2e/`
- Dependencies: Docker Compose per product
- Naming: `*_e2e_test.go` convention

---

**Status**: ANSWERED
**Next Step**: Generate Grooming Session 7 or begin implementation
