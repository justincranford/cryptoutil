# Quizme v2 — Target Structure Open Questions

**Plan**: docs/framework-v5/plan.md
**Context**: target-structure.md has been rewritten incorporating all v1 answers.
These questions surfaced during the rewrite and need your decisions.

---

## Question 1: tools/cicd/ Naming

**Question**: The `internal/apps/tools/cicd/` directory only contains custom linting
and formatting tools. The name `cicd/` is misleading. Should it be renamed?

**A)** Keep `tools/cicd/` — the name is familiar and changing it breaks too many imports
**B)** Rename to `tools/lint/` — since it only contains linters and formatters
**C)** Rename to `tools/quality/` — broader scope for future quality tooling
**D)** Rename to `tools/checks/` — neutral name that covers lint + format + fitness
**E)**

**Answer**:

**Rationale**: This rename affects `cmd/cicd/main.go`, all import paths, CI/CD
workflows, pre-commit hooks, and documentation. A rename should only happen if
the benefit clearly outweighs the migration cost.

---

## Question 2: TLS Generator Restructuring

**Question**: `framework/service/config/tls_generator/` currently generates TLS
certs only at the service level. Suite and product deployments also need TLS.
Should the TLS logic be extracted to a shared location?

**A)** Keep in `framework/service/config/tls_generator/` — suite/product can call it
**B)** Move to `framework/tls/` — shared across suite/product/service tiers
**C)** Move to `internal/shared/crypto/tls/` — it is a crypto primitive, not framework
**D)** Move to `framework/tls/` AND merge `internal/apps/pkiinit/` into it
**E)**

**Answer**:

**Rationale**: Currently `pkiinit/` and `tls_generator/` both generate TLS certs
but for different contexts. Merging avoids duplication. The question is whether
TLS generation belongs in framework or shared crypto.

---

## Question 3: apperr/ Location

**Question**: `internal/shared/apperr/` contains application-level error types
used across all services. Is it truly "shared utility" or is it "framework"?

**A)** Keep in `internal/shared/apperr/` — it is a general-purpose utility
**B)** Move to `internal/apps/framework/apperr/` — it is part of the application framework
**C)** Move to `internal/apps/framework/service/apperr/` — it is specifically for service-level errors
**D)** Split: generic errors stay in shared, framework-specific errors move to framework
**E)**

**Answer**:

**Rationale**: If apperr only contains HTTP status mapping and application error
wrapping, it belongs with the framework. If it contains general Go error utilities,
it belongs in shared.

---

## Question 4: framework/suite/ and framework/product/ Scope

**Question**: The target structure adds `framework/suite/` and `framework/product/`
alongside the existing `framework/service/`. What goes in them?

**A)** Empty scaffolds for now — create when suite/product CLI patterns emerge
**B)** Suite orchestration CLI + product aggregation CLI (extract from current inline code)
**C)** Suite/product lifecycle (start/stop/health aggregation) + CLI
**D)** Full parity with service/ (CLI, config, server, testing subdirs)
**E)**

**Answer**:

**Rationale**: Currently suite/product CLI code lives inline in
`internal/apps/{SUITE}/{SUITE}.go` and `internal/apps/{PRODUCT}/{PRODUCT}.go`.
Extracting shared patterns to framework avoids duplication across 5 products.

---

## Question 5: testdata/ Directory

**Question**: `testdata/` currently contains only `adaptive-sim/sample-auth-logs.json`.
Should it be kept or cleaned up?

**A)** Keep as-is — the sample data is useful for adaptive auth simulation testing
**B)** Delete entirely — test data should live next to tests (`*_test.go` directories)
**C)** Keep the directory but move sample data to the relevant package's testdata/
**D)** Keep and expand — add more shared test fixtures here
**E)**

**Answer**:

**Rationale**: Go convention is to place `testdata/` next to the package that uses
it. A root-level `testdata/` is unusual but acceptable for cross-package test data.
