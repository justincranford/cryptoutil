# Speckit Passthru02 Grooming Session 03: Directory Structure & 4-Products Architecture

**Purpose**: Structured questions to refine directory organization, product separation, and code architecture alignment with P1-P4 product strategy.
**Created**: 2025-12-02
**Status**: AWAITING ANSWERS

---

## Instructions

Select your answer by changing `[ ]` to `[x]` for each question. Add comments in the "Notes" field if needed. Multiple selections allowed where indicated.

---

## Section 1: Current Structure Assessment (Q1-8)

### Q1. Current `internal/` Organization

Current structure: `internal/{client,cmd,common,crypto,identity,infra,server,test}`. How well does this align with the 4-products strategy?

- [ ] A. Well aligned - minor adjustments needed
- [ ] B. Partially aligned - significant refactoring needed
- [ ] C. Poorly aligned - major restructuring required
- [ ] D. Not aligned - complete redesign needed

**Notes**:

---

### Q2. KMS Code Location

KMS code is in `internal/server/`. Should it be moved?

- [ ] A. Keep in `internal/server/` - it's working well
- [ ] B. Move to `internal/kms/` - align with product strategy
- [ ] C. Move to `internal/product/kms/` - explicit product separation
- [ ] D. Move to `pkg/kms/` - make it externally importable

**Notes**:

---

### Q3. JOSE Code Location

JOSE code is in `internal/common/crypto/jose/`. Should it be moved?

- [ ] A. Keep in `internal/common/` - it's shared infrastructure
- [ ] B. Move to `internal/jose/` - treat as product
- [ ] C. Move to `internal/product/jose/` - explicit product separation
- [ ] D. Split: core primitives in `internal/infra/jose/`, service in `internal/product/jose/`

**Notes**:

---

### Q4. Identity Code Location

Identity code is in `internal/identity/`. Is this appropriate?

- [ ] A. Keep in `internal/identity/` - current structure is fine
- [ ] B. Move to `internal/product/identity/` - explicit product separation
- [ ] C. Split into `internal/identity/authz/` and `internal/identity/idp/`
- [ ] D. Split into `internal/product/authz/` and `internal/product/idp/` (separate products)

**Notes**:

---

### Q5. Infrastructure vs Product Boundary

What's the right boundary between infrastructure and products?

- [ ] A. `internal/infra/` for all shared code, `internal/product/` for products
- [ ] B. `internal/common/` for shared, `internal/{jose,identity,kms,ca}/` for products
- [ ] C. Current: `internal/common/` + `internal/server/` + `internal/identity/`
- [ ] D. Flat: `internal/{jose,identity,kms,ca,config,telemetry,crypto,...}`

**Notes**:

---

### Q6. `internal/common/` Scope

`internal/common/` contains: apperr, config, container, crypto, magic, pool, telemetry, testutil, util. Is this appropriate?

- [ ] A. All belong in `internal/common/` - shared infrastructure
- [ ] B. Move crypto to `internal/infra/crypto/` - it's infrastructure
- [ ] C. Move telemetry to `internal/infra/telemetry/` - it's infrastructure
- [ ] D. Rename `internal/common/` to `internal/infra/` - clearer naming
- [ ] E. Multiple: crypto, telemetry, config to `internal/infra/`, keep util, magic, testutil in `internal/common/`

**Notes**:

---

### Q7. `api/` Directory Organization

`api/` contains: authz, client, fix_external_ref, identity, idp, model, server. How should this be organized?

- [ ] A. Current structure is fine
- [ ] B. Organize by product: `api/{jose,identity,kms,ca}/`
- [ ] C. Split: `api/openapi/` for specs, `api/generated/` for generated code
- [ ] D. Per-product OpenAPI: `api/jose/openapi/`, `api/kms/openapi/`, etc.

**Notes**:

---

### Q8. `cmd/` Directory Organization

`cmd/` contains: cicd, cryptoutil, demo, identity, identity-demo, identity-orchestrator, workflow. Is this appropriate?

- [ ] A. Current structure is fine
- [ ] B. Consolidate demos: `cmd/demo/{kms,identity,integration}`
- [ ] C. Per-product binaries: `cmd/{jose,identity,kms,ca}/`
- [ ] D. Single binary with subcommands: `cmd/cryptoutil/{serve,demo,admin}`

**Notes**:

---

## Section 2: Target Structure (Q9-16)

### Q9. Proposed `internal/` Structure

Which target structure is most appropriate?

- [ ] A. `internal/{infra,product}/{subpackages}`
- [ ] B. `internal/{jose,identity,kms,ca,infra}/{subpackages}`
- [ ] C. `internal/{common,server,client,identity}/{subpackages}` (current)
- [ ] D. `internal/{core,products}/{subpackages}`

**Notes**:

---

### Q10. Infrastructure Packages

What should be in `internal/infra/` (or equivalent)?

- [ ] A. config, telemetry, crypto, database, networking
- [ ] B. config, telemetry, crypto, database, networking, security
- [ ] C. Only low-level: crypto, database, networking
- [ ] D. Everything shared: config, telemetry, crypto, database, networking, util, magic, pool

**Notes**:

---

### Q11. Product Package Structure

What should each product package contain?

- [ ] A. Flat: `internal/product/kms/{handler,service,repository,domain}`
- [ ] B. Layered: `internal/product/kms/api/`, `internal/product/kms/business/`, `internal/product/kms/data/`
- [ ] C. Current KMS pattern: `application, barrier, businesslogic, handler, middleware, repository`
- [ ] D. Simplified: `internal/product/kms/{server,client,model}`

**Notes**:

---

### Q12. Shared Domain Models

Where should shared domain models live?

- [ ] A. Each product has its own domain models
- [ ] B. `internal/domain/` for shared models
- [ ] C. `internal/infra/domain/` for shared models
- [ ] D. `pkg/domain/` - externally importable

**Notes**:

---

### Q13. `deployments/` Organization

`deployments/` contains: compose.integration.yml, Dockerfile, identity/, kms/, telemetry/. How should this be organized?

- [ ] A. Current structure is fine
- [ ] B. Per-product: `deployments/{jose,identity,kms,ca}/`
- [ ] C. By type: `deployments/{docker,kubernetes,terraform}/`
- [ ] D. Flat with product prefixes: `deployments/{compose-kms.yml,compose-identity.yml,...}`

**Notes**:

---

### Q14. Docker Compose Organization

Should each product have its own compose file?

- [ ] A. Single compose.yml with all services
- [ ] B. Per-product compose files that can be combined
- [ ] C. Base compose + per-product overrides
- [ ] D. Current: separate directories (identity/, kms/, telemetry/)

**Notes**:

---

### Q15. Dockerfile Strategy

How many Dockerfiles should there be?

- [ ] A. Single multi-stage Dockerfile with build args for product selection
- [ ] B. Per-product Dockerfiles: `deployments/{jose,identity,kms,ca}/Dockerfile`
- [ ] C. Current: single Dockerfile at `deployments/Dockerfile`
- [ ] D. Base + per-product: `deployments/Dockerfile.base`, `deployments/kms/Dockerfile`

**Notes**:

---

### Q16. Config Files Location

Where should config files live?

- [ ] A. `configs/` at root (current)
- [ ] B. Per-product: `deployments/{product}/config/`
- [ ] C. In `internal/{product}/config/` - closer to code
- [ ] D. Both: defaults in code, overrides in `configs/` or `deployments/`

**Notes**:

---

## Section 3: Migration Strategy (Q17-22)

### Q17. Migration Approach

How should the directory restructuring be performed?

- [ ] A. Big bang - refactor everything at once
- [ ] B. Incremental - one product at a time
- [ ] C. Copy-first - create new structure, copy code, then remove old
- [ ] D. Gradual - create new structure, use both during transition

**Notes**:

---

### Q18. Migration Order

If incremental, what order?

- [ ] A. Infrastructure first, then products
- [ ] B. KMS first (working, stable), then Identity, then JOSE, then CA
- [ ] C. JOSE first (foundational), then KMS, then Identity, then CA
- [ ] D. Identity first (most active development)

**Notes**:

---

### Q19. Import Path Changes

How to handle import path changes during migration?

- [ ] A. Update all imports immediately
- [ ] B. Use type aliases during transition
- [ ] C. Create forwarding packages that re-export
- [ ] D. IDE refactoring (Go's gopls rename)

**Notes**:

---

### Q20. Test Migration Strategy

How should tests be migrated?

- [ ] A. Move with their packages
- [ ] B. Keep tests in place, update imports only
- [ ] C. Move to central `test/` directory
- [ ] D. Mix: unit tests with code, integration/e2e in `test/`

**Notes**:

---

### Q21. Breaking Changes

How to handle breaking changes to external APIs?

- [ ] A. No external APIs - all internal, break freely
- [ ] B. Maintain backwards compatibility during migration
- [ ] C. Version bump with documented breaking changes
- [ ] D. Deprecation warnings, removal in next major version

**Notes**:

---

### Q22. Validation Strategy

How to validate migration correctness?

- [ ] A. All tests pass
- [ ] B. All tests pass + all demos work
- [ ] C. All tests pass + demos work + Docker deployments work
- [ ] D. All above + manual E2E testing

**Notes**:

---

## Section 4: Code Organization Patterns (Q23-28)

### Q23. Package Naming Convention

What naming convention for packages?

- [ ] A. Short: `kms`, `identity`, `jose`
- [ ] B. Prefixed: `cryptoutilkms`, `cryptoutilidentity`
- [ ] C. Descriptive: `keymanagement`, `identityserver`
- [ ] D. Current: mixed (some short, some descriptive)

**Notes**:

---

### Q24. Internal vs Pkg

Should any code move from `internal/` to `pkg/`?

- [ ] A. No - everything should be internal
- [ ] B. Yes - domain models for external integration
- [ ] C. Yes - client SDKs for products
- [ ] D. Yes - utility packages that might be useful externally

**Notes**:

---

### Q25. Dependency Direction

What's the allowed dependency direction?

- [ ] A. Products can depend on infra, not on each other
- [ ] B. Products can depend on infra and other products
- [ ] C. Strict layering: infra → domain → product → cmd
- [ ] D. Acyclic: no circular dependencies, otherwise flexible

**Notes**:

---

### Q26. Interface Placement

Where should interfaces be defined?

- [ ] A. In the package that uses them (consumer)
- [ ] B. In the package that implements them (provider)
- [ ] C. In a central interfaces package
- [ ] D. Depends on the interface scope

**Notes**:

---

### Q27. Error Handling Packages

How should error handling be organized?

- [ ] A. Central `internal/infra/apperr/` for all errors
- [ ] B. Per-product error packages
- [ ] C. Mix: common errors in infra, product-specific in product
- [ ] D. No separate error packages - errors where they occur

**Notes**:

---

### Q28. Magic Values Organization

Where should magic values/constants live?

- [ ] A. Central `internal/infra/magic/` for all
- [ ] B. Per-product magic packages
- [ ] C. Current: `internal/common/magic/` + per-identity `internal/identity/magic/`
- [ ] D. In-package constants, no central magic package

**Notes**:

---

## Section 5: Specific Refactoring Decisions (Q29-35)

### Q29. `internal/server/` Disposition

What should happen to `internal/server/`?

- [ ] A. Rename to `internal/kms/` - it is KMS
- [ ] B. Move to `internal/product/kms/`
- [ ] C. Keep as-is - "server" is accurate
- [ ] D. Split: KMS-specific to `internal/kms/`, generic server infra to `internal/infra/server/`

**Notes**:

---

### Q30. `internal/client/` Disposition

What should happen to `internal/client/`?

- [ ] A. Keep as-is - it's client infrastructure
- [ ] B. Move to `internal/infra/httpclient/`
- [ ] C. Move per-product clients to product packages
- [ ] D. Move to `pkg/client/` - externally importable SDK

**Notes**:

---

### Q31. `internal/crypto/` Disposition

`internal/crypto/` contains registry.go, secret.go. What should happen?

- [ ] A. Move to `internal/infra/crypto/`
- [ ] B. Merge with `internal/common/crypto/`
- [ ] C. Keep separate - it's different from common/crypto
- [ ] D. Move to `internal/kms/crypto/` - KMS-specific

**Notes**:

---

### Q32. `internal/infra/` Creation

Should a new `internal/infra/` package be created?

- [ ] A. Yes - move appropriate packages from common
- [ ] B. No - rename `internal/common/` to `internal/infra/`
- [ ] C. No - keep current structure
- [ ] D. Yes - but only for new infrastructure code

**Notes**:

---

### Q33. Test Utilities Location

Where should test utilities live?

- [ ] A. `internal/testutil/` - central location
- [ ] B. `internal/infra/testutil/` - it's infrastructure
- [ ] C. `internal/common/testutil/` - current location
- [ ] D. Per-package `testutil_test.go` files

**Notes**:

---

### Q34. Demo Code Location

Where should demo code live?

- [ ] A. `cmd/demo/` - current location
- [ ] B. Per-product: `internal/{product}/demo/`
- [ ] C. `examples/` at project root
- [ ] D. Split: cmd for binary, internal for demo logic

**Notes**:

---

### Q35. Generated Code Location

Where should generated code (oapi-codegen) live?

- [ ] A. `api/` - current location
- [ ] B. `internal/generated/` - clearly marked as generated
- [ ] C. Per-product: `internal/{product}/api/generated/`
- [ ] D. `pkg/api/` - externally importable

**Notes**:

---

## Summary & Next Steps

After completing this grooming session:

1. Review answers for consistency
2. Identify migration priorities
3. Create refactoring task list
4. Update PROJECT-STATUS.md with refactoring plan
5. Share answers with Copilot for implementation guidance

---

*Session Created: 2025-12-02*
*Expected Completion: [DATE]*
