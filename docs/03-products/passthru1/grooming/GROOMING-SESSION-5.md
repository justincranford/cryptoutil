# Grooming Session 5: Aggressive Refactoring Plan

## Purpose

Session 5 focuses on the aggressive refactoring to reorganize KMS into product-based `./internal/` directory structure while preserving manual KMS implementation, putting LLM-generated identity in separate product directory, and preparing for certificate authority and JOSE authority products.

**Date**: November 30, 2025
**Status**: ANSWERED

---

## Section A: Scope & Goals Clarification (Q1-6)

### Q1. KMS Protection Scope

You said "preserves my KMS manual implementation." What specifically should be protected?

- [ ] A. Only `internal/server/` package (barrier, businesslogic, handler, repository)
- [x] B. `internal/server/` + `internal/client/`
- [ ] C. `internal/server/` + `internal/client/` + `internal/common/crypto/`
- [ ] D. All of the above + `cmd/cryptoutil/`
- [ ] E. Every file I manually wrote (vs LLM-generated)

**Notes**: Server and client packages are the core KMS implementation to protect.

---

### Q2. Identity Treatment

The Identity code is "LLM-generated across 6 passthrus" and needs fixing. Your approach:

- [x] A. Move to product directory AS-IS, then fix incrementally
- [ ] B. Audit and prune broken code FIRST, then move to product directory
- [ ] C. Keep in current location, fix completely, THEN reorganize
- [ ] D. Delete and rewrite from scratch in new product directory
- [ ] E. Depends on severity of issues (need assessment first)

**Notes**: Move to product dir as-is, figure out the code that needs to be extracted, extract chunks of code from identity to common code directories I identified and ensure the extracted code and identity code tests/e2e/workflow checks still pass.

---

### Q3. Refactoring vs Feature Work Ratio

During this aggressive refactoring phase, what's your tolerance for feature work?

- [ ] A. 100% refactoring - zero new features until structure is solid
- [ ] B. 90/10 - minimal bug fixes only
- [x] C. 80/20 - small features if they help demos
- [ ] D. 60/40 - refactoring shouldn't block demos
- [ ] E. Interleaved - alternate between refactoring and features

**Notes**:

---

### Q4. Directory Naming Convention

The plan shows `internal/infra/` and `internal/product/`. Do you prefer:

- [x] A. `internal/infra/` + `internal/product/` (current plan)
- [ ] B. `internal/infrastructure/` + `internal/products/` (explicit)
- [ ] C. `internal/shared/` + `internal/services/` (service-oriented)
- [ ] D. `internal/lib/` + `internal/apps/` (app-focused)
- [ ] E. Keep `internal/common/` for infra, create `internal/products/`

**Notes**:

---

### Q5. Certificate Authority Urgency

You mentioned "prepare to add certificate authority and jose authority product directories." How urgent?

- [ ] A. Part of this aggressive refactoring - create directories now, implement later
- [x] B. Create after KMS + Identity are reorganized (Phase 3+)
- [ ] C. Just placeholder directories - actual implementation is months away
- [ ] D. Not urgent - don't create until there's code to put in them
- [ ] E. High priority - want to start CA implementation within 2-4 weeks

**Notes**: Priority order is KMS, Identity, JOSE Authority, Certificate Authority.

---

### Q6. JOSE Authority Definition

You mentioned "jose authority" - what does this mean to you?

- [x] A. JWK/JWKS key server (distribute public keys for verification)
- [x] B. JWT issuing service (centralized token minting)
- [ ] C. Complete JWS/JWE/JWT library product (like current `internal/common/crypto/`)
- [ ] D. JOSE-based signing authority (like code signing)
- [ ] E. All JOSE operations unified under one product
- [ ] F. Something else (explain in notes)

**Notes**: Basic JOSE operations can be in the common shared crypto code, but JOSE Authority would contain:

- Embedded services for server-side issuance and client-side verification/usage
- Microservices for server and client
- Working demo of many different servers and clients running with different PKC domains

---

## Section B: Current Structure Analysis (Q7-12)

### Q7. `internal/common/` Disposition

`internal/common/` contains: `apperr/`, `config/`, `container/`, `crypto/`, `magic/`, `pool/`, `telemetry/`, `testutil/`, `util/`. What happens to it?

- [x] A. Move entirely to `internal/infra/` (rename common → infra)
- [ ] B. Split: crypto → product, rest → infra
- [ ] C. Split by usage: multi-product things → infra, product-specific → products
- [ ] D. Keep as `internal/common/` - don't rename during this pass
- [ ] E. Analyze each subdirectory individually before deciding

**Notes**: Move to infra first, then later Identity code will be analyzed to see what can be extracted to augment it.

---

### Q8. Duplication Analysis

The README says Identity has its own versions of: `apperr/`, `config/`, `magic/`, etc. Strategy:

- [ ] A. Consolidate immediately - one version in `internal/infra/`
- [ ] B. Keep separate for now - consolidate after products work
- [ ] C. Analyze differences first - may need both versions
- [ ] D. Identity versions replace common versions (they're newer)
- [x] E. Common versions replace Identity versions (they're proven)

**Notes**: KMS common is initial priority, Identity code will be analyzed, extracted, and merged into the original KMS common base later.

---

### Q9. Test Code Disposition

Test utilities are scattered: `internal/common/testutil/`, `internal/identity/test/`, `internal/test/`. Strategy:

- [x] A. All to `internal/infra/testing/`
- [ ] B. Keep product-specific tests with products, shared utils to infra
- [ ] C. Keep current locations - not worth the churn
- [ ] D. Create `internal/testing/` as top-level (not under infra)
- [ ] E. Depends on what's actually in each directory

**Notes**:

---

### Q10. `cmd/` Structure

Currently: `cmd/cicd/`, `cmd/cryptoutil/`, `cmd/identity/`, `cmd/identity-demo/`, `cmd/identity-orchestrator/`, `cmd/workflow/`. After refactoring:

- [ ] A. Keep as-is - `cmd/` stays separate from `internal/product/`
- [ ] B. Move under products: `internal/product/kms/cmd/`, `internal/product/identity/cmd/`
- [ ] C. Rename for clarity: `cmd/kms/` instead of `cmd/cryptoutil/`
- [x] D. Create new: `cmd/ca/`, `cmd/jose/` for new products
- [ ] E. Multiple changes (specify in notes)

**Notes**:

---

### Q11. E2E Test Location

`internal/test/e2e/` contains integration tests. After refactoring:

- [ ] A. Stay at `internal/test/e2e/` (cross-product tests)
- [ ] B. Move to `internal/infra/testing/e2e/`
- [x] C. Split: product-specific E2E → under product, cross-product → infra
- [ ] D. Move to `test/e2e/` (root level, not internal)
- [ ] E. Create `internal/integration/` for all integration tests

**Notes**: Each product gets their own e2e subdirectory; and later an `internal/e2e/` for cross-product orchestration of the product-specific e2e subdirectories.

---

### Q12. Deployment Artifacts

`deployments/compose/` has Docker Compose files. After refactoring:

- [ ] A. Keep at `deployments/` - works fine
- [ ] B. Move to `internal/infra/containers/`
- [x] C. Split by product: `deployments/kms/`, `deployments/identity/`
- [ ] D. Combine: `deployments/compose-kms.yml`, `deployments/compose-identity.yml`
- [ ] E. No opinion - deployments aren't the focus right now

**Notes**:

---

## Section C: Risk & Dependencies (Q13-18)

### Q13. Import Path Change Strategy

Moving packages changes import paths. 100+ files affected. Strategy:

- [ ] A. Big bang - all changes in one massive commit
- [ ] B. Create aliases first, migrate gradually, remove aliases
- [x] C. Move one package at a time, fix imports, commit, repeat
- [ ] D. Use `goimports` and IDE refactoring for bulk changes
- [ ] E. Script the import path changes (Go AST manipulation)

**Notes**:

---

### Q14. Test Coverage During Refactoring

Current coverage varies (27-85%). During refactoring:

- [x] A. Coverage must stay same or improve - no regressions allowed
- [ ] B. Minor regressions OK (<5%) if they're temporary
- [ ] C. Focus on refactoring - coverage is secondary for now
- [ ] D. Set floor: nothing below 70% allowed
- [ ] E. Different standards: products 80%+, infra 90%+

**Notes**: Add task after all product and infra refactoring, to analyze and create list of independent Cloud Agent Session prompts, that I can submit for concurrent processing in GitHub PRs, and merge without conflicts. Goal is to have prompts that can achieve iterative implementation, fixing, verification, and working code with high test coverage 90%. The prompts have to be good enough that they will work with a single request each, because if they require many iterations each they will be unmanageable and wasteful of premium requests. Max 6 prompts, with the more straightforward and less risky package codecov stuff prioritized first.

---

### Q15. CI/CD Impact Assessment

Refactoring paths affects CI/CD workflows. Your approach:

- [ ] A. Update workflows as part of each refactoring PR
- [x] B. Do all refactoring first, fix CI/CD in one pass
- [ ] C. Maintain parallel paths during transition
- [ ] D. Disable CI/CD checks temporarily during refactoring
- [ ] E. Haven't thought about this - need to investigate

**Notes**:

---

### Q16. Breaking Changes Communication

This is a "v2 rewrite" (per GROOMING-QUESTIONS.md). How to communicate:

- [x] A. No external users - don't worry about breaking changes
- [ ] B. CHANGELOG.md documents all breaking changes
- [ ] C. Migration guide with step-by-step instructions
- [ ] D. Keep old paths as aliases for backward compatibility
- [ ] E. Git tag before refactoring (v1.x), after refactoring (v2.x)

**Notes**:

---

### Q17. Parallel Development Risk

What if urgent bug/feature needed during refactoring?

- [ ] A. Refactoring continues - bugs wait for completion
- [ ] B. Pause refactoring - fix bug in current structure
- [ ] C. Branch: fix bug in main, continue refactoring in feature branch
- [x] D. Fix bug in refactored structure (forward-only)
- [ ] E. Depends on severity - critical bugs interrupt, others wait

**Notes**:

---

### Q18. Rollback Threshold

At what point would you abandon the refactoring and revert?

- [x] A. Never - committed to the new structure
- [ ] B. If tests can't pass after 1 week
- [ ] C. If CI/CD broken for more than 2 days
- [ ] D. If demos can't run by deadline
- [ ] E. If effort exceeds 2x estimate

**Notes**:

---

## Section D: Product-Specific Decisions (Q19-25)

### Q19. KMS Product Boundary

What exactly goes in `internal/product/kms/`?

- [ ] A. Just `internal/server/` (current KMS server code)
- [ ] B. `internal/server/` + `internal/client/` (server + client)
- [ ] C. Server + client + related crypto operations
- [x] D. Server + client + crypto + barrier/unseal logic
- [ ] E. Need to analyze dependencies first

**Notes**:

---

### Q20. Identity Product Boundary

What exactly goes in `internal/product/identity/`?

- [ ] A. Everything in current `internal/identity/`
- [ ] B. Only working parts of `internal/identity/`
- [x] C. `internal/identity/` minus stuff that should be infra
- [ ] D. Restructure significantly while moving
- [ ] E. Need to audit before deciding

**Notes**:

---

### Q21. Cross-Product Sharing

If KMS and Identity both need crypto operations, where does crypto live?

- [x] A. `internal/infra/crypto/` - both products import it
- [ ] B. `internal/common/crypto/` - keep current name
- [ ] C. Embedded in each product - duplicate is OK for isolation
- [x] D. `internal/product/jose/` - crypto is part of JOSE product
- [ ] E. Need to understand actual dependencies first

**Notes**: Both A and D apply - basic crypto in infra, JOSE-specific operations in JOSE product.

---

### Q22. JOSE Product Scope

Creating `internal/product/jose/` - what goes in it?

- [ ] A. JWK, JWS, JWE, JWT operations (current `internal/common/crypto/jose/`)
- [ ] B. Above + JWKS server endpoints
- [ ] C. Above + OAuth2.1/OIDC token generation
- [ ] D. Just library code - no server/endpoints (infra, not product)
- [ ] E. This is the same as Identity's token issuing - don't need separate product

**Notes**: None of these apply. Current JOSE package goes in `internal/infra/jose/`. `internal/product/jose/` will be cross-product reusable in `internal/product/identity/`. It will support many templates/profiles like a CA, but for JWKs, JWSs, JWEs, etc. instead of certificates. Hoping previously created code in current Identity can be extracted and used as the initial basis of `internal/product/jose/` (JOSE Authority).

---

### Q23. CA Product Placeholder

Creating `internal/product/ca/` - what should exist initially?

- [ ] A. Empty directory with README.md describing vision
- [ ] B. Basic package structure (cmd/, domain/, handler/, etc.)
- [x] C. Move existing X.509 code if any exists
- [ ] D. Don't create until actual implementation starts
- [ ] E. Create with interface definitions only (no implementation)

**Notes**: There is code for issuing TLS server/client certificates that can be moved into it. Subsequent functionality would be to expand the number of Subject profiles, to support N certificates per subject (e.g., 1 cert per TLS server subject, 1 cert per TLS client subject, 2 cert per SCEP server, 2 cert per SCEP client, 3 cert per user Subject for authn/encrypt@rest/nonrepudiation, and many other typical use cases in CA products, etc.).

---

### Q24. Product Independence

Should products be importable by other products?

- [ ] A. Yes - `product/identity` can import `product/jose`
- [x] B. No - products only import infra, never each other
- [ ] C. One-way only: smaller products (jose) can be imported, larger can't
- [ ] D. Through embedded package only: `product/identity/embedded`
- [ ] E. Case-by-case - document dependencies explicitly

**Notes**: For now B. Later, we might need to think about a suite manager, that is responsible for managing N deployments of 1-4 products per deployment (e.g., KMS deployment, with delegation of Authn/Authz to an Identity deployment, and the Identity deployment delegates JWS Issuance to JOSE Authority deployment). Haven't really thought that through yet, but for now that might be the general approach unless future grooming comes up with a different outcome.

---

### Q25. Demo Target After Refactoring

After refactoring completes, what's the demo?

- [ ] A. Same as passthru1: KMS encrypt/decrypt, Identity OAuth, Integration
- [ ] B. New demo: showcase the clean product architecture
- [ ] C. All existing tests pass (demos are tests)
- [ ] D. Docker compose up brings up all products
- [x] E. Multiple demos for different audiences

**Notes**:

1. Docker compose per product, so they can be setup/torn down independently, for manual testing each one
2. Demo executable to orchestrate docker compose per product, to simplify ease of use for demo
3. Overarching demo executable that will orchestrate the per-product demos, with configuration that sets up delegation between the independent product clusters

---

## Analysis Summary

### Key Decisions Captured

| Area | Decision |
|------|----------|
| **KMS Scope** | `internal/server/` + `internal/client/` protected |
| **Identity Approach** | Move as-is to product dir, extract to infra incrementally |
| **Directory Structure** | `internal/infra/` + `internal/product/` |
| **Product Priority** | KMS → Identity → JOSE Authority → Certificate Authority |
| **Common Code** | KMS common base prioritized, Identity merged into it later |
| **Import Strategy** | One package at a time, commit, repeat |
| **Coverage** | No regressions allowed; post-refactor task for 90% coverage prompts |
| **CI/CD** | Fix in one pass after refactoring complete |
| **Breaking Changes** | No external users, don't worry about it |
| **Rollback** | Never - committed to new structure |

### JOSE Authority Vision

- Basic JOSE operations → `internal/infra/jose/`
- JOSE Authority product → `internal/product/jose/`
  - Embedded services for server-side issuance
  - Client-side verification/usage
  - Microservices for server and client
  - Multi-PKC domain demo support
  - Cross-product reusable by Identity

### CA Product Vision

- Move existing TLS cert code to `internal/product/ca/`
- Expand Subject profiles:
  - TLS server (1 cert)
  - TLS client (1 cert)
  - SCEP server (2 certs)
  - SCEP client (2 certs)
  - User authn/encrypt/nonrepudiation (3 certs)
  - Other CA product use cases

### Demo Architecture

```
┌─────────────────────────────────────────────────────────┐
│           Overarching Demo Orchestrator                  │
│  (configures delegation between product clusters)        │
└─────────────────────────────────────────────────────────┘
                           │
       ┌───────────────────┼───────────────────┐
       ▼                   ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ KMS Demo    │     │ Identity    │     │ JOSE Auth   │
│ Orchestrator│     │ Demo Orch   │     │ Demo Orch   │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │
       ▼                   ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ deployments/│     │ deployments/│     │ deployments/│
│ kms/        │     │ identity/   │     │ jose/       │
│ compose.yml │     │ compose.yml │     │ compose.yml │
└─────────────┘     └─────────────┘     └─────────────┘
```

### Post-Refactoring Coverage Task

Create max 6 independent Cloud Agent Session prompts for concurrent GitHub PRs:

- Single request each (no multi-iteration)
- Merge without conflicts
- Target 90% coverage
- Prioritize straightforward, less risky packages first

---

**Status**: ANSWERED
**Next Step**: See GROOMING-SESSION-6.md for Q26-50 (Implementation Details)
