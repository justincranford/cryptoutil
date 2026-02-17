# Implementation Plan - Configs/Deployments/CICD Rigor & Consistency v3

**Status**: Planning Complete (Quizme-v2 Answered)
**Created**: 2026-02-16
**Last Updated**: 2026-02-17
**Purpose**: Establish rigorous configs/, deployments/, and C ICD validation patterns with comprehensive template definitions, aggressive quality standards, and minimal documentation overhead

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL documentation must be accurate and complete
- ✅ **Completeness**: NO steps skipped, NO steps de-prioritized, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark steps complete without verification

**ALL issues are blockers - NO exceptions:**

- ✅ **Fix issues immediately** - When unknowns discovered, blockers identified, unit/integration/E2E/mutations/fuzz/bench/race/SAST/DAST/load/any tests fail, or quality gates are not met, STOP and address
- ✅ **Treat as BLOCKING** - ALL issues block progress to next phase or task
- ✅ **Document root causes** - Root cause analysis is part of planning AND implementation, not optional; planning blockers must be resolved during planning, implementation blockers MUST be resolved during implementation
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER de-prioritize quality** - Evidence-based verification is ALWAYS highest priority

**Rationale**: Maintaining maximum quality prevents cascading failures and rework.

---

## Overview

**Problem**: configs/ and deployments/ lack rigorous structure, validation, and propagation to ARCHITECTURE.md/instruction files. Current state has 7 SERVICE-level subdirs but missing PRODUCT/SUITE hierarchy, no systematic CICD validation (8 types needed), and incomplete documentation propagation.

**Solution**: Comprehensive 6-phase plan to restructure configs/, add PRODUCT/SUITE templates, implement 8 CICD validators with ≥98% coverage/mutation, update ARCHITECTURE.md with minimal but precise docs (ASCII diagrams), and propagate via chunk-based verbatim copying to instruction files.

**Scope**:
- Phase 1: configs/ restructuring (9 SERVICE subdirs)
- Phase 2: PRODUCT/SUITE config creation (6 new dirs)
- Phase 3: 8 CICD validators (naming, kebab-case, schema, ports, telemetry, admin, consistency, secrets) with parallel execution, aggressive secrets detection, NO mutation exemptions
- Phase 4: ARCHITECTURE.md updates (minimal depth, ASCII diagrams)
- Phase 5: Instruction file propagation (chunk-based verbatim copying)
- Phase 6: E2E validation (100% pass rate)

**Out of Scope**: Docker E2E tests (handled in V1), deployment restructuring (handled in V2), comprehensive ARCHITECTURE.md documentation (quizme-v2 Q1: minimal preferred)

---

## Background

**Prior Work**:
- V1: Docker E2E test fixes (cipher-im timeout resolution, health check improvements)
- V2: Deployment refactoring (validation foundation, listings, mirror detection)

**Phase 0 Research** (internal, completed): Analyzed configs/ structure (identified 7 SERVICE subdirs, missing PRODUCT/SUITE), analyzed deployments/ structure (SERVICE/PRODUCT/SUITE hierarchy present), identified 8 CICD validation types, analyzed quizme-v1/v2 user answers (18 total decisions), documented findings in `test-output/fixes-v3-quizme-analysis/`, `test-output/fixes-v3-quizme-v2-analysis/`.

**Current State Issues**:
1. configs/ has only SERVICE level (cipher/, pki/, identity/, jose/, sm-kms/ planned)
2. Missing PRODUCT (cipher/, pki/, identity/, sm/, jose/) and SUITE (cryptoutil/) configs
3. identity/ is flat (needs 5 SERVICE subdirs: authz, idp, rp, rs, spa)
4. No CICD validators exist (need 8 types with ≥98% coverage/mutation, NO exemptions)
5. ARCHITECTURE.md missing deployment/config sections (need minimal docs per Q1:A)
6. Instruction files outdated (need chunk-based propagation per Q5:E)
7. CONFIG-SCHEMA.md integration unclear (Q2 unanswered, assuming D: embed + parse)

---

## Technical Context

- **Language**: Go 1.25.5
- **Framework**: CICD utility (cmd/cicd/) with validator pattern
- **Current CICD**: 2 foundation subcommands (generate-listings, validate-mirror)
- **Target CICD**: 8 validator types (naming, kebab-case, schema, ports, telemetry, admin, consistency, secrets)
- **Database**: N/A (file-based validation only)
- **Dependencies**: gopkg.in/yaml.v3 for config parsing, entropy analysis library for secrets detection
- **Related Files**:
  - configs/ (7 SERVICE subdirs currently, 15 total after Phase 1+2: 9 SERVICE + 5 PRODUCT + 1 SUITE)
  - deployments/ (SERVICE/PRODUCT/SUITE hierarchy exists)
  - cmd/cicd/ (2 subcommands currently, 10 total after Phase 3: +8 validators)
  - docs/ARCHITECTURE.md (missing sections 12.4-12.6)
  - docs/CONFIG-SCHEMA.md (embeddable schema reference)
  - .github/instructions/04-01.deployment.instructions.md (needs propagation)
  - .github/instructions/02-01.architecture.instructions.md (needs propagation)

---

## Phases

### Phase 0: Pre-Planning Research (COMPLETED INTERNALLY)

**Objective**: Internal research to inform plan creation (NOT documented as output phase per quizme-v2 Q10:E)

**Research Completed**:
- Analyzed configs/ current structure (7 subdirs, mixed SERVICE/incomplete hierarchy)
- Analyzed deployments/ structure (SERVICE/PRODUCT/SUITE hierarchy complete)
- Identified 8 CICD validation types needed
- Analyzed quizme-v1.md user answers (8 questions)
- Analyzed quizme-v2.md user answers (10 questions, incl. Q2 blank→default D)
- Documented findings in: `test-output/fixes-v3-quizme-analysis/`, `test-output/fixes-v3-quizme-v2-analysis/`
- Synthesized findings into this plan.md per quizme-v2 Q10:E (not just test-output/ references)

**Key Findings Synthesized**:
1. **Template Pattern** (Decision 4A): Needs concrete validation rules (naming patterns, structure requirements, value patterns including port offset calculations)
2. **Performance Critical** (Decision 11): Pre-commit must use parallel validators, target <5s incremental
3. **Documentation Philosophy** (Decision 9): Minimal ARCHITECTURE.md depth preferred (brief overview, defer to code)
4. **Propagation Method** (Decision 13): Chunk-based verbatim copying from ARCHITECTURE.md to instruction files
5. **Quality Rigor** (Decision 17): NO mutation testing exemptions, ALL validators ≥98%
6. **Secrets Detection** (Decision 15): Aggressive with entropy analysis, not just pattern matching

**Success**: Internal research complete, quizme-v1 + quizme-v2 answered (18 decisions total), ready for implementation

---

### Phase 1: configs/ Directory Restructuring (12h) [Status: ☐ TODO]

**Objective**: Restructure configs/ to mirror deployments/ hierarchy at SERVICE level

**Key Tasks**:
1. Rename cipher/ → cipher-im/ (SERVICE-level)
2. Rename pki/ → pki-ca/ (SERVICE-level)
3. Restructure identity/ → identity-{authz,idp,rp,rs,spa}/ (5 SERVICE subdirs per Decision 1)
4. Create sm-kms/ under configs/ (SERVICE-level, new directory)
5. Rename jose/ → jose-ja/ (SERVICE-level)
6. Preserve shared files (identity/policies/, identity/profiles/, identity/{development,production,test}.yml at identity/ parent per Decision 2)
7. Update all file references in Go code (search for old paths, update imports/file opens)
8. Create configs/README.md documenting SERVICE/PRODUCT/SUITE structure (minimal per Decision 8)

**Success Criteria**:
- configs/ has 9 SERVICE-level subdirs: cipher-im/, pki-ca/, identity-{authz,idp,rp,rs,spa}/, sm-kms/, jose-ja/
- Shared files preserved at parent levels (identity/ parent has policies/, profiles/, environment yamls)
- All code references updated (no broken imports/file paths)
- Tests pass: `go test ./...`
- Build clean: `go build ./...`
- Evidence collected: `test-output/phase1/restructure-verification.log`

---

### Phase 2: PRODUCT/SUITE Config Creation (6h) [Status: ☐ TODO]

**Objective**: Create PRODUCT-level (cipher/, jose/, identity/, sm/, pki/) and SUITE-level (cryptoutil/) configs following template patterns

**Key Tasks**:
1. Create cipher/ PRODUCT configs (delegates to cipher-im/ per Decision 4, 4A, 12)
2. Create pki/ PRODUCT configs (delegates to pki-ca/)
3. Create identity/ PRODUCT configs (delegates to 5 services: authz, idp, rp, rs, spa)
4. Create sm/ PRODUCT configs (delegates to sm-kms/)
5. Create jose/ PRODUCT configs (delegates to jose-ja/)
6. Create cryptoutil/ SUITE configs (delegates to all 5 products)
7. Add README.md to each PRODUCT/SUITE directory (minimal per Decision 8: purpose, delegation, ARCHITECTURE.md link)

**Template Pattern** (from Decision 4, 4A, 12):
- **Naming**: `{product-level-name}/config.yml` or `{suite-level-name}/config.yml`
- **Structure**: Required keys (service-name, delegation: [service-ids]), optional keys (ports with offset, telemetry endpoints)
- **Values**: Port offset calculations (PRODUCT = SERVICE + 10000, SUITE = SERVICE + 20000), delegation patterns, secrets handling
- **Validation**: Naming + structure + values per Decision 12

**Success Criteria**:
- 5 PRODUCT-level configs created (cipher/, pki/, identity/, sm/, jose/)
- 1 SUITE-level config created (cryptoutil/)
- All configs follow template pattern (validated by Task 3.4)
- README.md in each directory (1 paragraph purpose, delegation list, ARCHITECTURE.md link)
- Tests pass: `go test ./...`
- Build clean: `go build ./...`
- Evidence collected: `test-output/phase2/template-validation.log`

---

### Phase 3: CICD Validation Implementation (23h) [Status: ☐ TODO]

**Objective**: Implement 8 CICD validators with ≥98% coverage/mutation, parallel execution, aggressive secrets detection

**Key Tasks**:
1. Task 3.1: ValidateNaming (2h) - filename patterns (cipher-im-app.yml, cipher-im-app-common.yml)
2. Task 3.2: ValidateKebabCase (2h) - keys/values kebab-case enforced
3. Task 3.3: ValidateSchema (3h) - embed CONFIG-SCHEMA.md, parse at init per Decision 10 (Q2 blank→default D)
4. Task 3.4: ValidateTemplatePattern (3h) - naming + structure + values per Decision 12
5. Task 3.5: ValidatePorts (2h) - port ranges, offsets (PRODUCT +10000, SUITE +20000), conflicts
6. Task 3.6: ValidateTelemetry (2h) - OTLP endpoints, sidecar config
7. Task 3.7: ValidateAdmin (2h) - admin API patterns (127.0.0.1:9090)
8. Task 3.8: ValidateSecrets (2h) - aggressive detection with entropy per Decision 15
9. Task 3.9: Pre-commit integration (1h) - parallel validators per Decision 11
10. Task 3.10: Mutation testing (3h) - ALL validators ≥98% per Decision 17 (NO exemptions)
11. Task 3.11: Performance benchmarks (1h) - target <5s pre-commit incremental per Decision 11
12. Task 3.12: Validation caching (added Priority 1, not directly from quizme-v2)

**Validation Requirements** (from Decision 3, 12, 14, 15, 17):
- **Strictness**: All constraints enforced (Decision 3:C)
- **Scope**: Naming + structure + values (Decision 12:C)
- **Error Messages**: Moderate verbosity - core info + suggested fix + file/line (Decision 14:B)
- **Secrets Detection**: Aggressive with entropy analysis (Decision 15:C)
- **Mutation Coverage**: ≥98% for ALL validators, NO exemptions (Decision 17:A)

**Success Criteria**:
- 8 validators implemented with unit tests
- Integration tests pass (cross-validator interactions)
- Coverage ≥98% for ALL validators (cmd/cicd/ is infrastructure code)
- Mutation testing ≥98% for ALL validators (NO exemptions per Decision 17)
- Pre-commit runs validators in parallel per Decision 11
- Performance <5s for incremental validation per Decision 11
- Evidence collected: `test-output/phase3/validator-results.log`, coverage reports, mutation scores

---

### Phase 4: ARCHITECTURE.md Updates (6h) [Status: ☐ TODO]

**Objective**: Add minimal ARCHITECTURE.md sections 12.4-12.6 with ASCII diagrams (reduced from 8h due to Decision 9:A)

**Key Tasks**:
1. Task 4.1: Section 12.4 - Deployment Validation (2h, minimal depth per Decision 9)
   - Brief overview of 8 validators
   - ASCII diagram of validation flow per Decision 16
   - Defer details to code comments per Decision 9
2. Task 4.2: Section 12.5 - Config File Architecture (2h, minimal depth)
   - SERVICE/PRODUCT/SUITE hierarchy summary
   - ASCII diagram of config delegation per Decision 16
   - Template pattern reference (link to Decision 4A)
3. Task 4.3: Section 12.6 - Secrets Management (2h, minimal depth)
   - Docker secrets priority
   - Brief validation approach mention
   - Defer to ValidateSecrets implementation
4. Task 4.4: Add cross-references (optional, minimal maintenance)
   - Link sections 12.4-12.6 to existing ARCHITECTURE.md sections
5. Task 4.5: Cross-reference validation tool (added Priority 1)
   - Automated tool to verify section number consistency

**Documentation Philosophy** (Decision 9:A):
- **Depth**: Minimal - brief overview, defer to code comments and validator implementations
- **Diagrams**: ASCII only per Decision 16:B (Git-friendly, limited expressiveness acceptable)
- **Maintenance**: Low burden preferred over comprehensive guidance
- **Rationale**: Code is primary documentation, ARCHITECTURE.md provides high-level structure only

**Success Criteria**:
- Sections 12.4, 12.5, 12.6 added to ARCHITECTURE.md
- ASCII diagrams present (not Mermaid per Decision 16)
- Minimal depth maintained (1-2 paragraphs + diagram per section)
- Cross-references validated (Task 4.5 tool)
- Evidence collected: `test-output/phase4/architecture-sections.md`

---

### Phase 5: Instruction File Propagation (7h) [Status: ☐ TODO]

**Objective**: Propagate ARCHITECTURE.md patterns to instruction files via chunk-based verbatim copying (updated from 6h due to Decision 13:E)

**Key Tasks**:
1. Task 5.1: Update 04-01.deployment.instructions.md (2h)
   - Copy ARCHITECTURE.md Section 12.4 chunks verbatim per Decision 13
   - Add deployment validation patterns
   - Add references to 8 validators
2. Task 5.2: Update 02-01.architecture.instructions.md (2h)
   - Copy ARCHITECTURE.md Section 12.5 chunks verbatim
   - Add config file architecture patterns
   - Add SERVICE/PRODUCT/SUITE hierarchy
3. Task 5.3: Chunk-based verification (2h, changed from checklist per Decision 13:E)
   - Implement tool to verify ARCHITECTURE.md chunks present in instruction files
   - Verify verbatim copying (not just keyword presence)
   - ARCHITECTURE.md is single source of truth per Decision 13
4. Task 5.4: Doc consistency check (added Priority 1, 1h)
   - Checklist-based tool for systematic verification
   - Ensures propagation completeness

**Propagation Method** (Decision 13:E):
- **Chunk-Based**: ARCHITECTURE.md chunks copied verbatim to instruction files
- **Single Source**: ARCHITECTURE.md is single source of trust per Decision 13
- **Verification**: Tool checks chunk presence in instruction files (not just keywords)
- **Clarification**: If needed, add note to ARCHITECTURE.md (not instruction files)

**Success Criteria**:
- 04-01.deployment.instructions.md updated with verbatim chunks
- 02-01.architecture.instructions.md updated with verbatim chunks
- Chunk verification tool implemented (Task 5.3)
- ALL chunks from ARCHITECTURE.md present in instruction files
- Doc consistency check passed (Task 5.4)
- Evidence collected: `test-output/phase5/propagation-verification.log`

---

### Phase 6: E2E Validation (3h) [Status: ☐ TODO]

**Objective**: Validate all configs and deployments pass comprehensive validation

**Key Tasks**:
1. Run cicd lint-deployments against ALL configs/ files (100% pass required)
2. Run cicd lint-deployments against ALL deployments/ files (100% pass required)
3. Verify pre-commit hooks work correctly (parallel execution per Decision 11)
4. Test sample violations to verify detection (naming, kebab-case, schema, ports, telemetry, admin, consistency, secrets)
5. Document validation results (evidence-based completion)

**Success Criteria**:
- ALL configs/ files pass validation (100%, zero failures)
- ALL deployments/ files pass validation (100%, zero failures)
- Sample violations detected correctly (8/8 validator types)
- Pre-commit integration functional (parallel, <5s incremental per Decision 11)
- Evidence collected: `test-output/phase6/validation-results.txt`, `test-output/phase6/sample-violations.log`

---

## Executive Decisions

### Decision 1: identity/ Directory Restructuring (quizme-v1 Q1)

**Options**:
- A: Keep flat structure, rename files with prefixes
- B: Partial restructure (split authz/idp only)
- **C: Separate subdirs for each service (authz, idp, rp, rs, spa)** ✓ **SELECTED**
- D: Single subdir identity/services/ with all files

**Decision**: Option C selected - 5 separate SERVICE subdirs

**Rationale**:
- Mirrors deployments/identity/ structure exactly
- Clear SERVICE-level separation (required by ARCHITECTURE.md)
- Scales to PRODUCT-level identity/ parent configs in Phase 2
- Eliminates ambiguity (no mixed authz/idp/rs files)

**Alternatives Rejected**:
- Option A: Flat structure harder to navigate, filename prefixes brittle
- Option B: Partial split still leaves mixed files
- Option D: Nested services/ subdir adds unnecessary hierarchy

**Impact**:
- Phase 1 Task 1.3: Create 5 subdirs, move/rename 15+ config files
- Shared files (policies/, profiles/, environment yamls) preserved at identity/ parent
- Code references must be updated

**Evidence**: User selected "C" in quizme-v1.md Q1

---

### Decision 2: Environment-Specific Files Handling (quizme-v1 Q2)

**Options**:
- A: Move to each service subdir (duplicates shared files)
- **B: Keep at parent identity/ level, services reference via ../development.yml** ✓ **SELECTED**
- C: Single shared/ directory for all products
- D: Hardcode in service configs (eliminates files)

**Decision**: Option B selected - Preserve at parent, relative references

**Rationale**:
- Eliminates duplication (5 services share 3 environment files)
- Relative path references simple (`../development.yml`)
- Scales to multiple products (each product can have parent-level shared files)
- Maintains flexibility (per-service overrides still possible)

**Alternatives Rejected**:
- Option A: 15 duplicate files (5 services × 3 environments) = maintenance burden
- Option C: Cross-product sharing too broad, breaks SERVICE isolation
- Option D: Hardcoding reduces flexibility, complicates testing

**Impact**:
- Phase 1 Task 1.3: Preserve identity/{development,production,test}.yml at parent
- Service configs reference: `../development.yml` pattern
- README.md documents shared file locations

**Evidence**: User selected "B" in quizme-v1.md Q2

---

### Decision 3: Config Validation Strictness (quizme-v1 Q3)

**Options**:
- A: Permissive (warnings only, no errors)
- B: Moderate (required keys enforced, optional keys flexible)
- **C: Strict (all constraints enforced, no bypasses)** ✓ **SELECTED**
- D: Configurable per-file (complexity burden)

**Decision**: Option C selected - Strict validation, all constraints enforced

**Rationale**:
- Maximum rigor (consistent with "most awesome implementation plan" requirement)
- Prevents config drift (all files validated identically)
- Early error detection (fails fast on invalid configs)
- Simplifies debugging (validation errors explicit, not runtime failures)

**Alternatives Rejected**:
- Option A: Too permissive, allows invalid configs to reach production
- Option B: Moderate strictness leaves ambiguity (what's required vs optional?)
- Option D: Per-file configuration adds complexity without clear benefit

**Impact**:
- Phase 3 Tasks 3.1-3.8: All validators enforce strict rules
- No bypasses or warnings-only mode
- Validation failures block CI/CD pipeline

**Evidence**: User selected "C" in quizme-v1.md Q3

---

### Decision 4: PRODUCT/SUITE Config Template Pattern (quizme-v1 Q4)

**Options**:
- A: Manual creation (no templates, full flexibility)
- B: Copy-paste from SERVICE with edits (simple, error-prone)
- **C: Template-driven generation with validation** ✓ **SELECTED**
- D: Fully automated (infrastructure-as-code approach)

**Decision**: Option C selected - Template-driven with comprehensive validation

**Rationale**:
- Consistency (all PRODUCT/SUITE configs follow same pattern)
- Validation enforced (Decision 4A defines concrete rules)
- Balance of flexibility and rigor (templates guide, validation enforces)
- Scales to future products/services (template reusable)

**Alternatives Rejected**:
- Option A: Manual creation too error-prone (no consistency checks)
- Option B: Copy-paste brittle (divergence over time)
- Option D: Full automation too rigid (limits product-specific customization)

**Impact**:
- Phase 2 all tasks: Generate PRODUCT/SUITE configs from templates
- Phase 3 Task 3.4: ValidateTemplatePattern enforces compliance
- Decision 4A: Concrete template validation rules defined

**Evidence**: User selected "C" in quizme-v1.md Q4

---

### Decision 4A: Template Pattern Definition (CRITICAL - From Analysis)

**Options**:
- A: Abstract principle ("must follow template pattern") - NOT SELECTED
- B: Concrete rules (naming, structure, values) - ✓ **SELECTED** (implicitly via analysis)

**Decision**: Concrete template validation rules defined

**Rationale**:
- "Follows template pattern" too vague for validation automation
- Concrete rules enable ValidateTemplatePattern implementation
- Eliminates ambiguity in what constitutes valid template compliance

**Concrete Rules**:

1. **Naming Patterns**:
   - SERVICE: `{service-id}/{app-type}.yml` (e.g., `cipher-im/app.yml`)
   - PRODUCT: `{product-id}/config.yml` (e.g., `cipher/config.yml`)
   - SUITE: `cryptoutil/config.yml`

2. **Structure Requirements**:
   - SERVICE: Required keys (service-name, ports), optional keys (telemetry, admin-api)
   - PRODUCT: Required keys (product-name, delegation: [service-ids]), optional keys (shared-ports)
   - SUITE: Required keys (suite-name, delegation: [product-ids]), optional keys (shared-telemetry)

3. **Value Patterns**:
   - Port offsets: PRODUCT = SERVICE + 10000, SUITE = SERVICE + 20000
   - Delegation: Array of child service/product IDs
   - Secrets: MUST use file:///run/secrets/ paths (NEVER inline)

4. **Validation Enforcement**:
   - Task 3.4 (ValidateTemplatePattern) checks naming, structure, value patterns
   - Task 3.5 (ValidatePorts) verifies offset calculations
   - Task 3.8 (ValidateSecrets) enforces secrets file usage

**Impact**:
- Task 3.4: Implementation now has concrete acceptance criteria
- Phase 2: Template generation uses these rules
- Documentation: Rules documented in ARCHITECTURE.md Section 12.5

**Evidence**: Added in Priority 1 improvements from ANALYSIS.md

---

### Decision 5: CICD Validation Scope (quizme-v1 Q5)

**Options**:
- A: Minimal (naming + basic structure only)
- B: Moderate (naming + structure + ports + telemetry)
- **C: Comprehensive (8 types: naming, kebab-case, schema, ports, telemetry, admin, consistency, secrets)** ✓ **SELECTED**
- D: Exhaustive (all above + performance + security + compliance)

**Decision**: Option C selected - Full suite of 8 validator types

**Rationale**:
- Maximum rigor without over-engineering (exhaustive adds diminishing returns)
- Covers critical validation categories (naming, structure, security, consistency)
- Aligns with "most awesome implementation plan" requirement
- Balanced LOE (23h for Phase 3 vs >40h for exhaustive)

**Alternatives Rejected**:
- Option A: Minimal scope insufficient (misses schema, secrets, consistency)
- Option B: Moderate scope misses admin patterns and consistency checks
- Option D: Exhaustive overkill (performance/compliance better in dedicated tools)

**Impact**:
- Phase 3: 8 validator implementations (2-3h each)
- Phase 3 LOE: 23h total (includes mutation testing, pre-commit, benchmarks)
- Quality standard: ALL validators ≥98% coverage/mutation (Decision 17)

**Evidence**: User selected "C" in quizme-v1.md Q5

---

### Decision 6: CICD Implementation Strategy (quizme-v1 Q6)

**Options**:
- **A: Comprehensive cicd lint-deployments with multiple subcommands** ✓ **SELECTED**
- B: Simple validation script (shell/Python)
- C: Pre-commit hooks only (no standalone CLI)
- D: External tool integration (golangci-lint plugin)

**Decision**: Option A selected - Full-featured cicd lint-deployments CLI

**Rationale**:
- Consistent with existing cicd utility architecture (cmd/cicd/)
- Extensible (easy to add new validators as subcommands)
- Testable (Go tests for each validator)
- Reusable (CLI usable in CI/CD, pre-commit, manual validation)

**Alternatives Rejected**:
- Option B: Shell/Python scripts harder to test, less type-safe
- Option C: Pre-commit only limits CI/CD reusability
- Option D: External integration adds dependency, less control

**Impact**:
- Phase 3: All validators implemented as cicd subcommands
- Architecture: cmd/cicd/lint-deployments/{validate-naming, validate-schema, ...}
- Integration: Pre-commit calls cicd CLI (not separate scripts)

**Evidence**: User selected "A" in quizme-v1.md Q6

---

### Decision 7: Template Propagation Strategy (quizme-v1 Q7)

**Options**:
- **A: All configs validated against template patterns** ✓ **SELECTED**
- B: Only PRODUCT/SUITE configs validated
- C: Validation optional (warnings only)
- D: No template validation (trust manual creation)

**Decision**: Option A selected - Universal template validation

**Rationale**:
- Consistency across all config levels (SERVICE, PRODUCT, SUITE)
- Early detection of template violations (fails fast)
- Enforces architectural patterns (prevents drift)
- Complements Decision 3 (strict validation) and Decision 4 (template-driven)

**Alternatives Rejected**:
- Option B: SERVICE configs also need validation (not just PRODUCT/SUITE)
- Option C: Warnings-only too permissive (conflicts with Decision 3 strictness)
- Option D: No validation defeats purpose of templates

**Impact**:
- Phase 3 Task 3.4: ValidateTemplatePattern checks ALL config files
- Phase 2: All generated configs must pass Task 3.4 validation
- Phase 6: E2E validation includes template compliance

**Evidence**: User selected "A" in quizme-v1.md Q7

---

### Decision 8: README.md Content Requirements (quizme-v1 Q8)

**Options**:
- **A: Minimal (purpose + delegation + ARCHITECTURE.md link)** ✓ **SELECTED**
- B: Moderate (add common patterns + examples)
- C: Comprehensive (full reference with all config keys documented)
- D: Generated (auto-generate from config files)

**Decision**: Option A selected - Minimal READMEs

**Rationale**:
- Aligns with Decision 9:A (minimal documentation depth)
- Reduces maintenance burden (comprehensive docs require updates)
- ARCHITECTURE.md is primary reference (READMEs just signposts)
- Purpose paragraph sufficient for orientation

**Alternatives Rejected**:
- Option B: Moderate docs duplicate ARCHITECTURE.md content
- Option C: Comprehensive docs high maintenance, outdated quickly
- Option D: Generated docs miss context and rationale

**Impact**:
- Phase 2: README.md tasks create minimal content only
- Content: Purpose paragraph + delegation pattern + ARCHITECTURE.md link (3-5 lines)
- Reduced Phase 2 LOE (6h vs 10h for comprehensive READMEs)

**Evidence**: User selected "A" in quizme-v1.md Q8

---

### Decision 9: ARCHITECTURE.md Documentation Depth (quizme-v2 Q1)

**Options**:
- **A: Minimal (brief overview, defer to code comments)** ✓ **SELECTED**
- B: Moderate (core principles, key patterns, examples)
- C: Comprehensive (detailed rules, extensive examples, decision rationale, edge cases)
- D: Reference-heavy (link to external docs)

**Decision**: Option A selected - Minimal documentation depth

**Rationale**:
- Code is primary documentation (comments explain implementation details)
- ARCHITECTURE.md provides high-level structure only (not detailed reference)
- Low maintenance burden preferred over comprehensive guidance
- Aligns with Decision 8:A (minimal READMEs)

**Alternatives Rejected**:
- Option B: Moderate depth still requires significant maintenance for examples
- Option C: Comprehensive docs too high maintenance burden (not "awesome" if outdated)
- Option D: Reference-heavy loses cohesion (reader must jump between docs)

**Impact**:
- Phase 4 Tasks 4.1-4.3: Minimal depth (1-2 paragraphs + ASCII diagram per section)
- Phase 4 LOE reduced: 8h → 6h (2h savings from less documentation)
- Sections 12.4-12.6: Brief overview, defer to ValidateXXX implementations

**Evidence**: User selected "A" in quizme-v2.md Q1

---

### Decision 10: CONFIG-SCHEMA.md Integration (quizme-v2 Q2)

**Options**:
- A: Parse CONFIG-SCHEMA.md markdown at runtime
- B: Generate Go types from CONFIG-SCHEMA.md
- C: Hardcode schema in Go
- **D: Embed CONFIG-SCHEMA.md as string, parse once at init** ✓ **SELECTED** (DEFAULT, Q2 blank)
- E: Delete CONFIG-SCHEMA.md and hardcode schema in Go

**Decision**: Option D selected (DEFAULT) - Embed + parse at init

**Rationale**:
- ⚠️ **Q2 was blank (unanswered)** - using recommended default D
- Balanced approach: Compiled-in doc (no file I/O runtime), human-readable markdown (CONFIG-SCHEMA.md kept)
- Parse once at init (performance acceptable, <10ms overhead)
- Prevents drift (embedded doc matches code at build time)

**Alternatives Considered**:
- Option E: User may have intended "code only" approach (aligns with Q1:A minimal docs)
- Option D: More flexible than E (keeps CONFIG-SCHEMA.md for reference)

**Impact**:
- Phase 3 Task 3.3: Embed CONFIG-SCHEMA.md, parse at Init() function
- Add `//go:embed` directive in ValidateSchema
- Parser library: Use gopkg.in/yaml.v3 or lightweight markdown parser
- ⚠️ **User should confirm**: If Option E preferred (delete CONFIG-SCHEMA.md), revise Task 3.3

**Evidence**: ⚠️ User left Q2 blank in quizme-v2.md - default D assumed (needs confirmation)

---

### Decision 11: Pre-Commit Performance Optimization (quizme-v2 Q3)

**Options**:
- A: Run all validators sequentially
- B: Run validators in parallel
- C: Cache validation results (file hash-based)
- D: Validate only staged files
- **E: Parallel validators + optimize individual validator performance** ✓ **SELECTED** (CUSTOM)

**Decision**: Option E selected (user custom answer) - Parallel execution + optimization

**Rationale**:
- User noted: "B; this is how all cicd linters work. It is optimized to do this efficiently. We can also optimize validator performance if needed."
- Parallel execution (Option B) is standard practice for linters
- Per-validator optimization (beyond parallelization) addresses individual bottlenecks
- Target: <5s pre-commit for incremental validation

**Implementation**:
- Run 8 validators in parallel (using goroutines, errgroup pattern)
- Optimize individual validators: minimize file I/O, efficient parsing, early exits
- Add benchmarks (Task 3.11) to measure per-validator performance
- Consider caching (Task 3.12 added in Priority 1) as additional optimization

**Alternatives Rejected**:
- Option A: Sequential too slow (30-60s for 50+ files)
- Option C: Caching alone insufficient without parallelization
- Option D: Staged-only validation misses cross-file consistency

**Impact**:
- Phase 3 Task 3.9: Implement parallel validator execution (add goroutine orchestration)
- Phase 3 Task 3.11: Performance benchmarks (measure per-validator and total time)
- Target: <5s incremental validation (developer experience priority)

**Evidence**: User selected "E" (custom) in quizme-v2.md Q3

---

### Decision 12: Template Validation Scope (quizme-v2 Q4)

**Options**:
- A: Naming only (filename patterns)
- B: Naming + structure (required/optional keys)
- **C: Naming + structure + values (port offsets, delegation, secrets)** ✓ **SELECTED**
- D: Full semantic validation (cross-file relationships)

**Decision**: Option C selected - Naming + structure + values

**Rationale**:
- Comprehensive validation without cross-file complexity
- Catches common errors: wrong port offsets (PRODUCT=+10000, SUITE=+20000), missing delegation, inline secrets
- Aligns with Decision 4A (concrete template rules)
- Feasible LOE (3h for Task 3.4 vs >5h for Option D)

**Validation Details**:
1. **Naming**: `{id}/{type}.yml` patterns per Decision 4A
2. **Structure**: Required keys (service-name, ports, delegation), optional keys validated if present
3. **Values**: Port offset calculations, delegation array format, secrets file paths (not inline)

**Alternatives Rejected**:
- Option A: Naming only too shallow (misses structural errors)
- Option B: Structure only misses value errors (wrong ports, inline secrets)
- Option D: Full semantic validation (e.g., verify delegated services exist) too complex for Phase 3

**Impact**:
- Phase 3 Task 3.4: ValidateTemplatePattern checks all 3 aspects (naming, structure, values)
- Task 3.4 LOE: 3h (comprehensive but focused)
- Acceptance criteria: Test all 3 validation types (naming, structure, value patterns)

**Evidence**: User selected "C" in quizme-v2.md Q4

---

### Decision 13: Propagation Verification Method (quizme-v2 Q5)

**Options**:
- A: Manual review (read ARCHITECTURE.md, check each instruction file)
- B: Keyword search (grep for deployment/config terms)
- C: Semantic diff (extract concepts, verify in instructions)
- D: Checklist-based (pre-defined list of patterns)
- **E: Chunk-based - ARCHITECTURE.md chunks copied verbatim, verify presence** ✓ **SELECTED** (CUSTOM)

**Decision**: Option E selected (user custom answer) - Chunk-based verbatim copying

**Rationale**:
- User noted: "Chunk-based, ARCHITECTURE.md chunks are intended to be copied verbatim into instruction files, custom agents, etc. ARCHITECTURE.md is the single source of trust, and copied-based propagation ensures consistency. If clarification is necessary, add note to ARCHITECTURE.md. We can verify that each chunk is present in the instruction files. This is a direct way to verify propagation."
- ARCHITECTURE.md is single source of truth (no paraphrasing in instruction files)
- Verbatim copying eliminates interpretation errors (no semantic drift)
- Verification tool checks chunk presence (simpler than semantic diff)

**Implementation**:
- Identify "chunk boundaries" in ARCHITECTURE.md (e.g., subsections, code blocks, diagrams)
- Copy chunks verbatim to instruction files (exact text match required)
- Verification tool: Extract chunks from ARCHITECTURE.md, search for exact match in instruction files
- If clarification needed: Add to ARCHITECTURE.md (not instruction files)

**Alternatives Rejected**:
- Option A: Manual review error-prone (no systematic verification)
- Option B: Keyword search finds references but not completeness
- Option C: Semantic diff too complex (requires NLP or manual concept mapping)
- Option D: Checklist-based still requires defining "what to check" (chunk method more direct)

**Impact**:
- Phase 5 Task 5.1-5.2: Copy ARCHITECTURE.md chunks verbatim (not paraphrase or summarize)
- Phase 5 Task 5.3: Implement chunk verification tool (extracts + matches chunks)
- Phase 5 LOE adjusted: +1h for chunk tool implementation (6h → 7h)

**Evidence**: User selected "E" (custom) in quizme-v2.md Q5

---

### Decision 14: Error Message Verbosity (quizme-v2 Q6)

**Options**:
- A: Terse (file, line, code only)
- **B: Moderate (core info + suggested fix + file/line)** ✓ **SELECTED**
- C: Verbose (stack traces + ARCHITECTURE.md refs + examples)
- D: Configurable (user-selectable verbosity)

**Decision**: Option B selected - Moderate error messages

**Rationale**:
- Actionable errors (suggested fix) without overwhelming output
- File/line info sufficient for navigation
- No stack traces (not debugging internal validator errors)
- No exhaustive examples (keep output concise)

**Error Format**:
```
ERROR: [validator-name] {file}:{line}
  Issue: {core problem description}
  Fix: {suggested resolution}
  Reference: ARCHITECTURE.md Section {X.Y} (optional, only if directly relevant)
```

**Alternatives Rejected**:
- Option A: Terse errors unhelpful (no fix suggestions, harder to debug)
- Option C: Verbose errors overwhelming (stack traces for user errors inappropriate)
- Option D: Configurable adds complexity without clear benefit

**Impact**:
- Phase 3 Tasks 3.1-3.8: All validators use moderate error format
- Consistent error structure across all validators
- No CLI flag for verbosity (simplifies implementation)

**Evidence**: User selected "B" in quizme-v2.md Q6

---

### Decision 15: Secrets Detection Strategy (quizme-v2 Q7)

**Options**:
- A: Secrets files only (.secret, environment: *_FILE)
- B: Secrets files + inline values (pattern matching only)
- **C: Aggressive - secrets files + inline values + entropy analysis** ✓ **SELECTED**
- D: Integrated secrets scanner (external tool like gitleaks)

**Decision**: Option C selected - Aggressive detection with entropy analysis

**Rationale**:
- Maximum security rigor (detects high-entropy strings that aren't pattern matches)
- Reduces false negatives (catches base64 secrets, API keys without standard patterns)
- Accept some false positives (UUIDs, base64 data) - better than missing secrets
- Aligns with "most awesome implementation plan" security focus

**Detection Layers**:
1. **Secrets files**: Check for .secret extensions, environment: *_FILE patterns
2. **Pattern matching**: Check known patterns (AWS keys, GitHub tokens, passwords)
3. **Entropy analysis**: Shannon entropy for inline string values (threshold: >4.5 bits/char)

**Alternatives Rejected**:
- Option A: Secrets files only misses inline secrets (common anti-pattern)
- Option B: Pattern matching misses custom secrets (no standard pattern)
- Option D: External tool adds dependency, less control over detection rules

**Impact**:
- Phase 3 Task 3.8: Implement ValidateSecrets with 3-layer detection
- Add entropy analysis library (or implement Shannon entropy)
- Test false positives: UUIDs, base64 encoded data (should NOT trigger, adjust threshold)
- Evidence: `test-output/phase3/secrets-detection-tests.log`

**Evidence**: User selected "C" in quizme-v2.md Q7

---

### Decision 16: ARCHITECTURE.md Diagram Format (quizme-v2 Q8)

**Options**:
- A: No diagrams (text-only)
- **B: ASCII diagrams (simple text diagrams)** ✓ **SELECTED**
- C: Mermaid diagrams (code-based, expressive)
- D: External diagrams (draw.io/Excalidraw)

**Decision**: Option B selected - ASCII diagrams

**Rationale**:
- Git-friendly (text-based, diff-able)
- No external tools required (renders in any text editor)
- Simple and maintainable (no Mermaid syntax to learn)
- Aligns with Decision 9:A (minimal documentation overhead)

**Limitations Accepted**:
- Limited expressiveness (complex relationships hard to visualize)
- Less visually appealing than Mermaid/external diagrams
- Trade-off: Simplicity over aesthetics

**Example ASCII Diagram**:
```
SERVICE/PRODUCT/SUITE Hierarchy:

cryptoutil/ (SUITE)
├── configs/
│   ├── cryptoutil/           [SUITE-level]
│   │   └── config.yml        (delegates to 5 PRODUCTs)
│   ├── cipher/               [PRODUCT-level]
│   │   └── config.yml        (delegates to cipher-im)
│   ├── cipher-im/            [SERVICE-level]
│   │   └── app.yml
│   └── ...
```

**Alternatives Rejected**:
- Option A: No diagrams misses visual clarity (some patterns easier to show than describe)
- Option C: Mermaid requires renderer (GitHub/VS Code support, but plain text viewing harder)
- Option D: External diagrams separate maintenance (PNG/SVG not Git-diffable)

**Impact**:
- Phase 4 Tasks 4.1-4.3: Include ASCII diagrams (not Mermaid)
- Diagram examples: Section 12.4 (validation flow), Section 12.5 (config hierarchy)
- No Mermaid fences (```mermaid) in ARCHITECTURE.md

**Evidence**: User selected "B" in quizme-v2.md Q8

---

### Decision 17: Mutation Testing Policy (quizme-v2 Q9)

**Options**:
- **A: No exemptions - ALL validators ≥98%** ✓ **SELECTED**
- B: Exempt trivial validators (kebab-case, naming)
- C: Exempt by complexity (only complex validators ≥98%)
- D: User decision per validator

**Decision**: Option A selected - NO exemptions, ALL validators ≥98%

**Rationale**:
- Strictest rigor (consistent with "most awesome implementation plan")
- Infrastructure code (CICD validators) CRITICAL - requires maximum quality
- Even trivial validators have edge cases (regex escaping, boundary conditions)
- Simplifies policy (same standard for ALL validators)

**Justification**:
- **ValidateKebabCase**: Regex edge cases (e.g., `kebab-case-123` vs `kebab_case`, `-case-` vs `case-`)
- **ValidateNaming**: Filename edge cases (e.g., `cipher-im-app.yml` vs `cipher-im-app-common.yml`)
- **ValidateSchema**: Complex logic (schema parsing, nested validation) - obviously ≥98%
- **ValidateSecrets**: Entropy calculation edge cases (ASCII vs UTF-8, special chars)

**Alternatives Rejected**:
- Option B: "Trivial" is subjective (who defines trivial?), exemptions create inconsistency
- Option C: Complexity-based exemptions still require categorization decisions
- Option D: Per-validator discretion defeats purpose of universal quality standard

**Impact**:
- Phase 3 Task 3.10: Mutation testing for ALL 8 validators (NO exemptions)
- Estimated 3h total (includes mutation testing for "trivial" validators)
- Quality gate: ≥98% mutation score for cmd/cicd/ (infrastructure code)

**Evidence**: User selected "A" in quizme-v2.md Q9

---

### Decision 18: Phase 0 Research Documentation (quizme-v2 Q10)

**Options**:
- A: Internal only (findings in test-output/, NOT in plan.md)
- B: Summary in plan.md (brief findings in Background/Executive Summary)
- C: Full documentation (Phase 0 findings fully in plan.md)
- D: Reference-based (plan.md references test-output/ for details)
- **E: Synthesize into plan.md - research findings integrated into plan.md sections** ✓ **SELECTED** (CUSTOM)

**Decision**: Option E selected (user custom answer) - Synthesized findings in plan.md

**Rationale**:
- User noted: "Research must be completed as part of creating the plan.md and tasks.md. If this is unclear in .github/agents/implementation-planning.agent.md (and .github/agents/implementation-execution.agent.md), then update the agent instructions to do the research as part of creating the plan.md and tasks.md. This ensures that all research findings are captured in the plan.md and tasks.md, which are the primary documentation for the implementation. The test-output/phase0-research/ can be used for raw data and notes, but the synthesized findings should be in the plan.md."
- Plan.md is comprehensive documentation (not minimal/reference-only)
- Research findings inform phases, decisions, tasks (integrated throughout plan.md)
- test-output/ for raw data, plan.md for synthesized insights

**Implementation**:
- Phase 0 section: Documents that research was completed (not detailed findings)
- Findings synthesized into: Technical Context, Executive Decisions rationales, Phase descriptions
- Example: "Key Findings Synthesized" subsection lists 6 major insights from research
- test-output/phase0-research/ still exists for raw notes/logs

**Alternatives Rejected**:
- Option A: Internal-only loses context (plan.md missing research insights)
- Option B: Summary too brief (doesn't capture all findings adequately)
- Option C: Full documentation clutters plan.md (verbose Phase 0 section)
- Option D: Reference-based makes plan.md incomplete (reader must check test-output/)

**Impact**:
- Phase 0 section: Includes "Key Findings Synthesized" subsection
- Executive Decisions: Rationales reference research findings
- Agent instructions: May need updating to clarify research synthesis expectation
- No separate "Research" phase in output (Phase 1 is first implementation phase)

**Evidence**: User selected "E" (custom) in quizme-v2.md Q10

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Q2 blank (CONFIG-SCHEMA.md integration unclear) | High | Medium | Used default D (embed+parse), flagged for user confirmation, Task 3.3 flexible to change |
| Phase 1 code reference updates break builds | Medium | High | Incremental refactoring, run tests after each rename, rollback plan if failures |
| Template validation too strict (blocks valid configs) | Medium | Medium | Task 3.4 comprehensive tests, sample configs validated before enforcement |
| Parallel validators introduce race conditions | Low | High | Use errgroup pattern, validator isolation (no shared state), race detector in tests |
| Entropy analysis false positives (UUIDs detected as secrets) | Medium | Low | Tune entropy threshold (>4.5 bits/char), whitelist UUID patterns, test false positive rate |
| Chunk-based propagation breaks on ARCHITECTURE.md edits | Medium | Medium | Chunk verification tool (Task 5.3) detects drift, clear chunk boundaries, versioning |
| Minimal ARCHITECTURE.md insufficient for developers | Low | Medium | Decision 9:A accepted (code comments primary), monitor developer feedback post-implementation |
| ASCII diagrams too limited for complex patterns | Low | Low | Decision 16:B accepted (simplicity over expressiveness), supplement with text explanations |
| Mutation testing ≥98% for trivial validators infeas ible | Low | Medium | Decision 17:A accepted (NO exemptions), add edge case tests, invest LOE if needed |
| Phase 5 LOE underestimated (chunk verification complex) | Medium | Low | 7h estimated (includes 2h for chunk tool), buffer in Phase 6 if overrun |

---

## Quality Gates - MANDATORY

**Per-Task Quality Gates**:
- ✅ All tests pass (`go test ./...`) - 100% passing, zero skips
- ✅ Build clean (`go build ./...`) - zero errors
- ✅ Linting clean (`golangci-lint run`) - zero warnings
- ✅ No new TODOs without tracking in tasks.md

**Coverage Targets** (from copilot instructions):
- ✅ Production code: ≥95% line coverage
- ✅ Infrastructure/utility code (cmd/cicd/): ≥98% line coverage (MANDATORY)
- ✅ main() functions: 0% acceptable if internalMain() ≥95%
- ✅ Generated code: Excluded (OpenAPI stubs, GORM models, protobuf)

**Mutation Testing Targets** (from Decision 17):
- ✅ ALL validators: ≥98% mutation score (NO exemptions)
- ✅ cmd/cicd/: ≥98% (infrastructure code, CRITICAL)

**Per-Phase Quality Gates**:
- ✅ Unit + integration tests complete before moving to next phase
- ✅ E2E tests pass (Phase 6: ALL configs/ + deployments/ validate)
- ✅ Pre-commit integration functional (Phase 3: parallel, <5s incremental)
- ✅ Race detector clean (`go test -race -count=2 ./...`)

**Overall Project Quality Gates**:
- ✅ All phases complete with evidence
- ✅ All test categories passing (unit, integration, E2E)
- ✅ Coverage and mutation targets met (≥98% for cmd/cicd/)
- ✅ CI/CD workflows green
- ✅ Documentation updated (ARCHITECTURE.md, instruction files)

---

## Success Criteria

- [ ] All 6 phases complete (Phase 1-6)
- [ ] configs/ restructured (9 SERVICE + 5 PRODUCT + 1 SUITE dirs)
- [ ] 8 CICD validators implemented (naming, kebab-case, schema, ports, telemetry, admin, consistency, secrets)
- [ ] ALL validators ≥98% coverage + mutation (NO exemptions)
- [ ] Pre-commit <5s incremental (parallel validators)
- [ ] ARCHITECTURE.md updated (minimal depth, ASCII diagrams, Sections 12.4-12.6)
- [ ] Instruction files updated (chunk-based verbatim copying)
- [ ] E2E validation 100% pass (ALL configs + deployments)
- [ ] Documentation consistent (ARCHITECTURE.md ↔ instruction files verified)
- [ ] CI/CD workflows green
- [ ] Evidence archived (test-output/phase1-6/)

---

## Evidence Archive

**Planning Evidence**:
- `test-output/fixes-v3-quizme-analysis/` - Quizme-v1 answers + analysis (8 questions)
- `test-output/fixes-v3-quizme-v2-analysis/` - Quizme-v2 answers + analysis (10 questions)
- `docs/fixes-v3/quizme-v1.md` - DELETED (merged into plan.md Decisions 1-8)
- `docs/fixes-v3/quizme-v2.md` - TO BE DELETED (will merge into plan.md Decisions 9-18)
- `docs/fixes-v3/ANALYSIS.md` - Deep analysis (15 improvements identified, Priority 1 applied)
- `docs/fixes-v3/COMPLETION-STATUS.md` - V2 completion summary

**Implementation Evidence** (to be created):
- `test-output/phase1/` - Phase 1 restructuring logs
- `test-output/phase2/` - Phase 2 template generation logs
- `test-output/phase3/` - Phase 3 validator implementation + coverage + mutation
- `test-output/phase4/` - Phase 4 ARCHITECTURE.md updates
- `test-output/phase5/` - Phase 5 propagation verification logs
- `test-output/phase6/` - Phase 6 E2E validation results
