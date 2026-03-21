# Quizme v1 - Framework v5 Decisions

**Purpose**: Clarify 4 strategic decisions before execution begins. All are currently tentative in plan.md.

---

## Question 1: configs/ vs deployments/config/ Relationship

**Question**: Two config systems coexist. `deployments/*/config/` has Docker-specific configs (standardized `{PS-ID}-app-{variant}.yml`). `configs/` has standalone/development/domain configs (inconsistent naming). What should the relationship be?

**A)** Merge all configs/ into deployments/*/config/ (single config location, one system)
**B)** Keep both with clear separation: configs/ = domain/standalone/development, deployments/config/ = Docker-specific. Standardize configs/ naming to `{PS-ID}-{purpose}.yml`. RECOMMENDED — both serve distinct purposes.
**C)** Move deployments/*/config/ into configs/ (centralize everything under configs/)
**D)** Deprecate configs/ entirely, only use deployments/config/ for everything
**E)**

**Answer**:

**Rationale**: This is the highest-impact decision. It determines the naming standard for 76+ config files and whether we maintain one or two config locations.

---

## Question 2: Non-Standard cmd/ Entry Disposition

**Question**: `cmd/identity-compose/` and `cmd/identity-demo/` violate the anti-pattern rule (Section 4.4.7: "NO executables for subcommands"). How should they be handled?

**A)** Keep all as-is, just document them as intentional exceptions
**B)** Merge identity-compose and identity-demo into `cmd/identity compose` and `cmd/identity demo` subcommands. Keep cmd/cicd, cmd/demo, cmd/workflow as documented infrastructure. RECOMMENDED.
**C)** Archive all demo entries (cmd/demo, cmd/identity-demo, internal/apps/demo) — demos are not needed. Keep only cicd/workflow.
**D)** Move cicd/workflow/demo under a new `cmd/tools/` pattern (separate from product/service binaries)
**E)**

**Answer**:

**Rationale**: Determines whether we enforce strict CLI patterns or allow documented exceptions for non-product tools.

---

## Question 3: Archive Deletion Policy

**Question**: 161+ files across 9 archived/orphaned directories. All are dead code superseded by framework v3/v4 work. How should they be handled?

**A)** Delete all permanently from main branch. Git history preserves content. Entity registry and fitness linters confirm active services. RECOMMENDED.
**B)** Move to a separate `archive` branch before deleting from main
**C)** Keep archived directories but add fitness linter to prevent growth
**D)** Compress into .tar.gz files in docs/ for offline reference
**E)**

**Answer**:

**Rationale**: 161+ dead files add search noise, confuse LLM agents analyzing the codebase, and increase cognitive load with zero value.

---

## Question 4: ARCHITECTURE-COMPOSE-MULTIDEPLOY.md Disposition

**Question**: `docs/ARCHITECTURE-COMPOSE-MULTIDEPLOY.md` (872 lines) contains detailed compose tier patterns (SERVICE/PRODUCT/SUITE delegation, override hierarchies, multi-deploy architecture). This content is NOT in ARCHITECTURE.md but is referenced in development. What should happen to it?

**A)** Merge content into ARCHITECTURE.md Section 12.3 and delete the file. ARCHITECTURE.md is the SSOT. RECOMMENDED.
**B)** Keep as supplementary document, add cross-reference from ARCHITECTURE.md Section 12.3
**C)** Convert to an instruction file (`.github/instructions/04-02.compose-multideploy.instructions.md`)
**D)** Delete without merging — ARCHITECTURE.md Section 12 already covers the essentials
**E)**

**Answer**:

**Rationale**: ARCHITECTURE.md is documented as the SSOT. Satellite documents create information silos that LLM agents may miss during analysis.
