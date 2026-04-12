# Lessons — Framework v10: Canonical Template Registry

## Pre-Implementation Research Findings

These findings were discovered during plan creation (Phase 0 research) and informed the v10 architecture.

### Finding 1: v9 Had Two Compounding Failures

**Wrong location**: v9 placed templates at `internal/.../template_drift/templates/` — the
same-package path required by `//go:embed`. The required location is `api/cryptosuite-registry/templates/`.

**Wrong mechanism**: v9 used `//go:embed` + `embed.FS`. This was architecturally incorrect.
`api/cryptosuite-registry/` is NOT a Go package — it is a plain directory of actual configuration
files (Dockerfiles, compose.yml, config YAML) that serve as the canonical spec for all deployment
artifacts. cicd-lint reads these files at runtime using `os.WalkDir`, not via Go's build system.

**Correct approach**: `os.WalkDir("api/cryptosuite-registry/templates")` at runtime. No Go package,
no `.go` files, no `embed.FS`, no import aliases. The templates directory is part of the source
tree, readable like any other config file when cicd-lint runs from the project root.

**Lesson**: When a requirement specifies a location for non-Go artifacts, do not force Go build
tooling (embed) onto them. Runtime file reading is appropriate for linter tools that always run
from a known working directory.

### Finding 2: Product Compose Templates Cannot Use a Single Generic Path

**Root Cause**: Product compose files (`deployments/sm/compose.yml`, etc.) list the specific
PS-IDs for each product. The identity product has 5 PS-IDs; sm has 2. A single
`deployments/__PRODUCT__/compose.yml` template cannot represent this diversity without
conditional logic that templates cannot express.

**Solution**: Per-product static template files with STATIC paths in the template directory
(e.g., `templates/deployments/sm/compose.yml`). Only `__SUITE__` and `__IMAGE_TAG__` are
substituted — all product-specific PS-ID names, port numbers, and service references are
hardcoded. These files are compared directly against their `./deployments/{product}/compose.yml`
counterpart (no expansion loop).

**Lesson**: Template parameterization should cover only values that vary uniformly across all
instances of the same file type. Per-product templates that are mostly literal content with 2
substitution points are simpler, more readable, and less error-prone than a generic expansion.

### Finding 3: Suite Files Are Standalone Templates (Not Runtime-Derived)

**Root Cause**: The suite Dockerfile (`deployments/cryptoutil/Dockerfile`) follows the same
4-stage build pattern as every PS-ID Dockerfile. An initial impulse was to derive the suite
Dockerfile at runtime from the `deployments/__PS_ID__/Dockerfile` template with
`__PS_ID__`=`cryptoutil`. This would save one file at the cost of complexity.

**Solution**: Store `templates/deployments/cryptoutil/Dockerfile` as a complete, standalone
template file — explicit rather than derived. The BuildExpectedFS expansion loop ignores it
(no `__PS_ID__` in the path); it substitutes suite-level params in the content and compares
directly against `deployments/cryptoutil/Dockerfile`. Same pattern for `compose.yml`.

**Lesson**: Derivation at runtime saves one file but adds a special-case expansion rule and
makes the template directory harder to inspect ("where is the suite Dockerfile template?").
Standalone files are more transparent and directly readable.

---

## Phase 1: Create Canonical Template Directory

*(To be filled during Phase 1 execution)*

---

## Phase 2: Rewrite Template Linter

*(To be filled during Phase 2 execution)*

---

## Phase 3: Update Documentation

*(To be filled during Phase 3 execution)*

---

## Phase 4: Quality Gates

*(To be filled during Phase 4 execution)*

---

## Phase 5: Knowledge Propagation

*(To be filled during Phase 5 execution)*
