# Lessons — Framework v10: Canonical Template Relocation

## Pre-Implementation Research Findings

These findings were discovered during plan creation (Phase 0 research) and informed the v10 architecture.

### Finding 1: v9 Task 8.1 Implementation Failure

**Root Cause**: v9 chose the path of least resistance — embedding templates in the same package
(`template_drift/templates/`) avoids Go's `//go:embed` restriction against `../` traversal.
The v9 plan claimed "templates at canonical location" but implemented the easy path instead.

**Go embed constraint**: `//go:embed` CANNOT reference paths outside the package directory (no `../`).
The only correct solution is creating a new Go package at `api/cryptosuite-registry/` that exports
`TemplatesFS embed.FS` and is imported by `template_drift.go`.

**Lesson**: When a requirement blocks a simple implementation path, document the constraint explicitly
and implement the correct (harder) solution instead of silently choosing a non-compliant path.

### Finding 2: Product Compose Templates Cannot Be Generic

**Root Cause**: Product compose files (`deployments/sm/compose.yml`, etc.) list per-product PS-IDs
as service names. The identity product has 5 PS-IDs; sm has 2. A single generic template cannot
represent this diversity without complex conditional logic.

**Solution**: Per-product static templates (`product-sm-compose.yml.tmpl` etc.) with only
`__SUITE__` and `__IMAGE_TAG__` as substitutable parameters. Product-specific PS-ID names,
ports, and includes are hardcoded in each per-product template.

**Lesson**: Template parameterization is limited to values that vary uniformly across all instances.
Per-product templates that are mostly static (2 substitution params) are correct; forcing generic
template reuse by complicating parameter handling creates fragile designs.

### Finding 3: Suite Dockerfile Reuses PS-ID Template (Decision 3)

**Root Cause**: The suite Dockerfile (`deployments/cryptoutil/Dockerfile`) follows the same
4-stage build pattern as every PS-ID Dockerfile. The only differences are in the binary name
(`cryptoutil` instead of `sm-kms` etc.) and display labels.

**Solution**: `CheckSuiteDockerfile` reuses `Dockerfile.tmpl` with params:
- `__PS_ID__` = `cryptoutil`
- `__PRODUCT_DISPLAY_NAME__` = `Cryptoutil`
- `__SERVICE_DISPLAY_NAME__` = `Suite`

No separate `suite-Dockerfile.tmpl` file is needed.

**Lesson**: Before creating a new template file, verify whether an existing template can be
parameterized to serve the new use case. "Template reuse via different params" avoids file
proliferation and simplifies maintenance.

---

## Phase 1: Template Relocation

*(To be filled during Phase 1 execution)*

---

## Phase 2: Missing Templates and Linters

*(To be filled during Phase 2 execution)*

---

## Phase 3: Documentation Update

*(To be filled during Phase 3 execution)*

---

## Phase 4: Quality Gates

*(To be filled during Phase 4 execution)*

---

## Phase 5: Knowledge Propagation

*(To be filled during Phase 5 execution)*
