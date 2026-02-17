# Implementation Plan - Configs/Deployments/CICD Rigor & Consistency v3

**Status**: Planning Complete
**Created**: 2026-02-17
**Last Updated**: 2026-02-17
**Purpose**: Achieve absolute rigor and consistency for configs/, deployments/, and CICD linting to fully comply with ARCHITECTURE.md standards and eliminate all inconsistencies.

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL documentation must be accurate and complete
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation
- ❌ **Premature Completion**: NEVER mark complete without verification

**ALL issues are blockers - NO exceptions:**

- ✅ **Fix issues immediately** - When unknowns discovered, blockers identified, unit/integration/E2E/mutations/fuzz/bench/race/SAST/DAST/load/any tests fail, or quality gates are not met, STOP and address
- ✅ **Treat as BLOCKING** - ALL issues block progress to next phase or task
- ✅ **Document root causes** - Root cause analysis is part of planning AND implementation, not optional; planning blockers must be resolved during planning, implementation blockers MUST be resolved during implementation
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER de-prioritize quality** - Evidence-based verification is ALWAYS highest priority

**Rationale**: Maintaining maximum quality prevents cascading failures and rework.

---

## Overview

This plan addresses structural inconsistencies in configs/ and deployments/ directories, implements comprehensive CICD validation for deployment and config files, and ensures full compliance with ARCHITECTURE.md patterns.

**Key Goals**:
1. Restructure configs/ to mirror deployments/ hierarchy (SERVICE, PRODUCT, SUITE levels)
2. Implement comprehensive CICD validation (8 types: naming, kebab-case, schema, ports, telemetry, admin, consistency, secrets)
3. Create PRODUCT-level and SUITE-level configs following template patterns
4. Update ARCHITECTURE.md with deployment/config patterns
5. Propagate ARCHITECTURE.md changes to all linked instruction files

---

## Background

**Prior Work**:
- V1: Completed Docker E2E test fixes and documentation consistency
- V2: Completed deployment/config refactoring with template pattern and orphaned config handling

**Current State Issues**:
1. configs/ missing SERVICE-level subdirs (cipher-im/, jose-ja/, pki-ca/, sm-kms/)
2. configs/ missing PRODUCT-level configs (cipher/, jose/, identity/, sm/, pki/)
3. configs/ missing SUITE-level configs (cryptoutil/)
4. configs/identity/ has mixed authz/idp/rs files (should be separate subdirs: identity-{authz,idp,rp,rs,spa})
5. CICD linting incomplete (missing 8 validation types)
6. ARCHITECTURE.md missing deployment/config rigor patterns
7. Instruction files not updated with ARCHITECTURE.md deployment patterns

**What V3 Must Achieve**:
- Absolute structural consistency between configs/ and deployments/
- Comprehensive CICD validation covering all deployment and config aspects
- Template-driven config generation for PRODUCT and SUITE levels
- ARCHITECTURE.md documentation of all deployment/config patterns
- Propagation of ARCHITECTURE.md patterns to all instruction files

---

## Executive Summary

**Critical Context**:
- User demands **"rigorous!!!"** validation - NO shortcuts, NO minimal approaches
- configs/ must mirror deployments/ structure (SERVICE, PRODUCT, SUITE)
- Template pattern MANDATORY for all PRODUCT/SUITE configs
- All 8 CICD validation types required (naming, kebab-case, schema, ports, telemetry, admin, consistency, secrets)
- ARCHITECTURE.md is single source of truth - updates MUST propagate to instruction files

**Assumptions & Risks**:
- **Assumption**: Template pattern from deployments/template/ applies to all PRODUCT/SUITE configs
- **Assumption**: Environment-specific files (development.yml, production.yml, test.yml) stay at parent product level
- **Risk**: Large file rewrites (20+ config files) could introduce errors -> **Mitigation**: Validate after each task
- **Risk**: CICD validation too strict breaks existing workflows -> **Mitigation**: Validate against all existing deployments first
- **Risk**: ARCHITECTURE.md updates not propagated completely -> **Mitigation**: Automated grep checks for instruction file references

---

## Technical Context

- **Language**: Go 1.25.5
- **Framework**: CICD utility (cmd/cicd/)
- **Validation**: YAML parsing, schema validation, port consistency checks
- **Dependencies**: gopkg.in/yaml.v3, Go standard library
- **Related Files**:
  - `cmd/cicd/` - CICD command implementation
  - `internal/cmd/cicd/` - CICD business logic
  - `docs/ARCHITECTURE.md` - Architecture reference
  - `.github/instructions/04-01.deployment.instructions.md` - Deployment patterns
  - `configs/` - Configuration templates
  - `deployments/` - Runtime deployments

---

## Phases

### Phase 0: Pre-Planning Research (COMPLETED INTERNALLY)

**Objective**: Internal research to inform plan creation (NOT documented as output phase)

**Research Completed**:
- Analyzed configs/ current structure (7 subdirs, missing SERVICE/PRODUCT/SUITE levels)
- Analyzed deployments/ structure (SERVICE/PRODUCT/SUITE hierarchy)
- Identified 8 CICD validation types needed
- Analyzed quizme-v1.md user answers
- Documented findings in: `test-output/fixes-v3-quizme-analysis/`

**Success**: Internal research complete, quizme-v1.md answered, ready for implementation

---

### Phase 1: configs/ Directory Restructuring (12h) [Status: ☐ TODO]

**Objective**: Restructure configs/ to mirror deployments/ hierarchy at SERVICE level

**Key Tasks**:
1. Rename cipher/ → cipher-im/ (SERVICE-level)
2. Rename pki/ → pki-ca/ (SERVICE-level)
3. Restructure identity/ → identity-{authz,idp,rp,rs,spa}/ (5 SERVICE subdirs)
4. Create sm-kms/ under configs/ (SERVICE-level)
5. Rename jose/ → jose-ja/ (SERVICE-level)
6. Preserve shared files (identity/policies/, identity/profiles/, environment yamls at identity/ parent)
7. Update all file references in code
8. Create configs/README.md documenting structure

**Success Criteria**:
- configs/ has 9 SERVICE-level subdirs: cipher-im/, pki-ca/, identity-{authz,idp,rp,rs,spa}/, sm-kms/, jose-ja/
- Shared files preserved at parent levels
- All code references updated
- Tests pass: `go test ./...`
- Build clean: `go build ./...`

---

### Phase 2: PRODUCT/SUITE Config Creation (6h) [Status: ☐ TODO]

**Objective**: Create PRODUCT-level (cipher/, jose/, identity/, sm/, pki/) and SUITE-level (cryptoutil/) configs following template patterns

**Key Tasks**:
1. Create cipher/ PRODUCT configs (delegates to cipher-im/)
2. Create pki/ PRODUCT configs (delegates to pki-ca/)
3. Create identity/ PRODUCT configs (delegates to 5 services)
4. Create sm/ PRODUCT configs (delegates to sm-kms/)
5. Create jose/ PRODUCT configs (delegates to jose-ja/)
6. Create cryptoutil/ SUITE configs (delegates to all 5 products)
7. Add README.md to each PRODUCT/SUITE directory (minimal: purpose, delegation, ARCHITECTURE.md link)

**Template Pattern** (from quizme Q4, Q7 answers):
- All PRODUCT/SUITE configs MUST follow deployments/template/ patterns
- Naming: `{PRODUCT}-app-{common,sqlite-1,postgresql-1,postgresql-2}.yml`
- Validation: Template patterns propagate to all generated configs

**Success Criteria**:
- All PRODUCT dirs exist: configs/{cipher,pki,identity,sm,jose}/
- SUITE dir exists: configs/cryptoutil/
- Each has 4 config variants: common, sqlite-1, postgresql-1, postgresql-2
- Each has README.md (minimal content)
- Template patterns validated: `cicd lint-deployments validate-config`
- Tests pass: `go test ./...`

---

### Phase 3: CICD Validation Implementation (23h) [Status: ☐ TODO]

**Objective**: Implement comprehensive cicd lint-deployments validation (8 types from quizme Q5, Q6 answers)

**8 Validation Types** (MANDATORY):
1. **Naming**: File names follow {PRODUCT-SERVICE}-{app|compose}-{variant}.{yml|yaml} pattern
2. **Kebab-Case**: All keys in kebab-case (no snake_case, no camelCase)
3. **Schema**: All config keys match CONFIG-SCHEMA.md, all compose keys valid
4. **Ports**: Public 8XXX, Admin 9090, PostgreSQL 543XX, telemetry standard ports
5. **Telemetry**: OTLP endpoint, protocol, service-name, insecure flag
6. **Admin**: bind-private-address ALWAYS 127.0.0.1, bind-private-port 9090
7. **Consistency**: Configs match deployed compose services
8. **Secrets**: NO inline credentials, ALL secrets use Docker secrets pattern

**Key Tasks**:
1. Implement ValidateNaming (file naming patterns)
2. Implement ValidateKebabCase (key naming conventions)
3. Implement ValidateSchema (config/compose schema compliance)
4. Implement ValidatePorts (port assignment rules)
5. Implement ValidateTelemetry (OTLP configuration)
6. Implement ValidateAdmin (admin endpoint security)
7. Implement ValidateConsistency (config-compose matching)
8. Implement ValidateSecrets (credential handling)
9. Add comprehensive tests (≥98% coverage for infrastructure code)
10. Add mutation testing (≥98% for infrastructure code)

**Success Criteria**:
- All 8 validation types implemented
- Tests pass with ≥98% coverage
- Mutation testing ≥98%
- Validates ALL configs/ files successfully
- Validates ALL deployments/ files successfully
- Pre-commit hook integration complete
- Documentation: README.md in cmd/cicd/lint-deployments/

---

### Phase 4: ARCHITECTURE.md Updates (8h) [Status: ☐ TODO]

**Objective**: Document deployment/config rigor patterns in ARCHITECTURE.md

**Key Sections to Add/Update**:
1. **Section 12.4**: Deployment Validation Architecture
   - 8 validation types (naming, kebab-case, schema, ports, telemetry, admin, consistency, secrets)
   - cicd lint-deployments command usage
   - Pre-commit hook integration
2. **Section 12.5**: Config File Architecture
   - SERVICE/PRODUCT/SUITE hierarchy
   - Template pattern requirements (Decision 4, Decision 4A)
   - Naming conventions
   - Schema compliance (CONFIG-SCHEMA.md reference)
3. **Section 12.6**: Secrets Management
   - Docker secrets pattern (MANDATORY)
   - File permissions (440)
   - NO inline credentials
4. **Config Pattern Examples**: Add examples for each validation type
5. **Validate cross-references**: Ensure consistency with instruction files - PRIORITY 1

**Success Criteria**:
- ARCHITECTURE.md sections 12.4-12.6 complete
- CONFIG-SCHEMA.md referenced correctly
- Examples validate against actual configs
- Table of contents updated
- ARCHITECTURE-INDEX.md updated with new sections

---

### Phase 5: Instruction File Propagation (6h) [Status: ☐ TODO]

**Objective**: Propagate ARCHITECTURE.md deployment patterns to all instruction files

**Files to Update**:
1. `.github/instructions/04-01.deployment.instructions.md`
   - Add 8 CICD validation types
   - Add SERVICE/PRODUCT/SUITE hierarchy
   - Add template pattern requirements
   - Add secrets management rules
2. `.github/instructions/02-01.architecture.instructions.md`
   - Add deployment architecture quick reference
   - Update config patterns
3. Update agent files if needed:
   - `.github/agents/implementation-planning.agent.md` (if deployment patterns referenced)
   - `.github/agents/implementation-execution.agent.md` (if deployment patterns referenced)
4. **Automated consistency check**: Verify propagation completeness - PRIORITY 1

**Propagation Verification**:
- Search for ARCHITECTURE.md deployment sections in instruction files
- Verify consistency with ARCHITECTURE.md content
- Check for outdated patterns
- Run: `grep -r "deployment\|config\|validation" .github/instructions/ | grep -v ".md:"`

**Success Criteria**:
- All instruction files updated
- No conflicting patterns between ARCHITECTURE.md and instructions
- Agent files updated if needed
- Documentation cross-references correct

---

### Phase 6: E2E Validation (3h) [Status: ☐ TODO]

**Objective**: Validate all configs and deployments pass comprehensive validation

**Key Tasks**:
1. Run cicd lint-deployments against ALL configs/
2. Run cicd lint-deployments against ALL deployments/
3. Verify pre-commit hooks work correctly
4. Test sample violations to verify detection
5. Document validation results

**Success Criteria**:
- ALL configs/ files pass validation (100%)
- ALL deployments/ files pass validation (100%)
- Sample violations detected correctly
- Pre-commit integration functional
- Evidence collected: `test-output/phase6/validation-results.txt`

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
- A: Move to each service subdir (duplicate across 5 subdirs)
- **B: Preserve in identity/ parent with cross-references** ✓ **SELECTED**
- C: Create identity/shared/ subdir
- D: Merge into service configs (no separate env files)

**Decision**: Option B selected - Keep at identity/ parent level

**Rationale**:
- development.yml, production.yml, test.yml are PRODUCT-level concerns (affect all 5 services)
- Avoids duplication across 5 SERVICE subdirs
- Service configs can reference parent env files via relative paths
- Matches shared policies/ and profiles/ pattern (also at parent level)

**Alternatives Rejected**:
- Option A: Duplication violates DRY principle
- Option C: Adds unnecessary shared/ subdir layer
- Option D: Loses environment-specific flexibility

**Impact**:
- Phase 1 Task 1.3: Preserve identity/{development,production,test}.yml at parent
- Service configs reference: `../development.yml` pattern
- README.md documents shared file locations

**Evidence**: User selected "B" in quizme-v1.md Q2

---

### Decision 3: Config Validation Strictness (quizme-v1 Q3)

**Options**:
- A: Basic YAML validity only
- B: Moderate validation (naming + schema)
- **C: Strict validation against deployment constraints** ✓ **SELECTED**
- D: Permissive with warnings

**Decision**: Option C selected - Strict validation

**Rationale**:
- User demands **"rigorous!!!"** approach (quizme historical pattern)
- Must catch ALL misconfigurations before deployment
- Validates: naming, schema, ports, telemetry, admin, secrets, consistency
- Prevents runtime errors from config mistakes

**Alternatives Rejected**:
- Option A: Too minimal, misses critical issues
- Option B: Moderate misses security issues (secrets, admin bindings)
- Option D: Warnings ignored in practice, defeats purpose

**Impact**:
- Phase 3: All 8 validation types MANDATORY (no optional validations)
- Validation failures block commits (pre-commit hook)
- Higher implementation complexity (18h vs 8h for basic)

**Evidence**: User selected "C" in quizme-v1.md Q3

---

### Decision 4: PRODUCT/SUITE Config Template Pattern (quizme-v1 Q4)

**Options**:
- A: Manual creation (copy-paste from SERVICE configs)
- B: Basic delegation (minimal config pointing to services)
- C: Reference-based (symbolic links to SERVICE configs)
- **D: Template-driven generation with shared conventions** ✓ **SELECTED**

**Decision**: Option D selected - Template-driven generation

**Rationale**:
- Consistent patterns across ALL PRODUCT/SUITE configs
- Leverages deployments/template/ established patterns
- Enforces conventions: naming, delegation, port offsets, secrets sharing
- Validates generated configs against template rules (Decision 3 strict validation)

**Alternatives Rejected**:
- Option A: Manual creation error-prone, inconsistent
- Option B: Basic delegation lacks rigor
- Option C: Symbolic links break Windows compatibility

**Impact**:
- Phase 2: ALL PRODUCT/SUITE configs follow template pattern
- Validation (Phase 3) checks template compliance
- Template updates propagate to all configs automatically

**Evidence**: User selected "D" in quizme-v1.md Q4

---

### Decision 4A: Template Pattern Definition (CRITICAL - From Analysis)

**Question**: What concrete rules define "follows template pattern" for validation?

**Context**: Decision 4 mandates template-driven generation. Need concrete validation rules for Phase 2 implementation and Phase 3 ValidateSchema.

**Template Pattern Compliance Rules**:

```yaml
Template Pattern Validation:
  Naming:
    - Pattern: {PRODUCT}-app-{variant}.yml
    - Variants: common, sqlite-1, postgresql-1, postgresql-2
    - Example: cipher-app-common.yml ✓, cipher_app.yml ✗, Cipher-App.yml ✗
  
  Structure:
    - Required keys: service-name, bind-public-port, bind-private-port, database-url, observability
    - Key naming: ALL kebab-case (no snake_case, no camelCase)
    - Nesting: Max 3 levels deep
  
  Value Patterns:
    - Port offsets:
      - SERVICE: Base range (e.g., cipher-im 8700-8799)
      - PRODUCT: SERVICE + 10000 (e.g., cipher 18700-18799)
      - SUITE: SERVICE + 20000 (e.g., cryptoutil 28700-28799)
    - Delegation:
      - PRODUCT configs delegate to SERVICE configs (relative paths: ../cipher-im/)
      - SUITE configs delegate to PRODUCT configs (relative paths: ../cipher/)
    - Secrets:
      - ALL credentials via file:///run/secrets/{secret-name}
      - NO inline passwords/tokens/keys
    - Service names:
      - Match directory name: identity/ → service-name: identity
      - Match delegation target: cipher/ → delegates-to: cipher-im
```

**Validation Implementation**:
- Phase 2 Tasks 2.1-2.6: Manual checklist validation (validators not ready)
- Phase 3 Task 3.3 ValidateSchema: Automated template pattern checks
- Phase 6 Task 6.1: Validate all PRODUCT/SUITE configs against template rules

**Examples**:

**Valid PRODUCT Config** (configs/cipher/cipher-app-common.yml):
```yaml
service-name: cipher
bind-public-port: 18080  # SERVICE 8080 + 10000
bind-private-port: 9090
delegates-to: ../cipher-im/cipher-im-app-common.yml
password-file: file:///run/secrets/cipher_db_password
```

**Invalid PRODUCT Config** (violations):
```yaml
serviceName: cipher  # ✗ camelCase (should be service-name)
bind_public_port: 18080  # ✗ snake_case (should be bind-public-port)
bind-public-port: 8080  # ✗ No offset (should be 18080 = 8080 + 10000)
password: "secret123"  # ✗ Inline credential (should be file:///run/secrets/)
```

**Impact**:
- Phase 2: Tasks 2.1-2.6 acceptance criteria updated with concrete checklist
- Phase 3: Task 3.3 ValidateSchema implements template pattern checks
- Phase 6: Task 6.1 validates ALL PRODUCT/SUITE configs pass template rules

**Evidence**: Identified in deep-analysis.md as CRITICAL improvement (Priority 1)

---

### Decision 5: CICD Validation Scope (quizme-v1 Q5)

**Options**:
- A: Minimal (naming + YAML validity)
- B: Moderate (A + schema + ports)
- **C: Full suite (naming, kebab-case, schema, ports, telemetry, admin, consistency, secrets)** ✓ **SELECTED**
- D: Staged approach (implement A now, B later, C final)

**Decision**: Option C selected - Full 8 validation types

**Rationale**:
- User demands **"rigorous!!!"** validation (consistent with historical quizme answers)
- Catches ALL config/deployment issues at commit time
- Prevents: naming errors, schema violations, port conflicts, telemetry gaps, admin security issues, secret leaks
- Aligns with FIPS 140-3 and Zero Trust principles (no shortcuts on security)

**Alternatives Rejected**:
- Option A: Too minimal, misses critical issues
- Option B: Moderate misses security-critical validations (secrets, admin)
- Option D: Staged approach deferred problems (user wants rigor NOW)

**Impact**:
- Phase 3: Implement ALL 8 validation types (18h vs 6h for minimal)
- Higher test coverage requirements (≥98% for infrastructure)
- Pre-commit performance impact (~30-60s for full validation suite)

**Evidence**: User selected "C" in quizme-v1.md Q5

---

### Decision 6: CICD Implementation Strategy (quizme-v1 Q6)

**Options**:
- A: Quick bash scripts (minimal Go integration)
- B: Pre-commit hooks only (no cicd command)
- C: Separate validation binaries per type
- **D: Comprehensive cicd lint-deployments command** ✓ **SELECTED**

**Decision**: Option D selected - cicd lint-deployments with 8 subcommands

**Rationale**:
- Unified command interface: `cicd lint-deployments <type> <file>`
- Supports all 8 validation types under single command tree
- Testable (≥98% coverage with unit + integration tests)
- Pre-commit hook calls cicd command (DRY principle)
- Extensible for future validation types

**Alternatives Rejected**:
- Option A: Bash scripts not testable, error-prone
- Option B: Pre-commit-only lacks standalone validation capability
- Option C: Separate binaries fragment tooling

**Impact**:
- Phase 3: Focus on robust cicd lint-deployments implementation
- Command structure: `cicd lint-deployments {validate-naming|validate-schema|validate-ports|...}`
- Pre-commit hook: thin wrapper calling cicd command
- Documentation: cmd/cicd/lint-deployments/README.md

**Evidence**: User selected "D" in quizme-v1.md Q6

---

### Decision 7: Template Propagation Strategy (quizme-v1 Q7)

**Options**:
- A: Manual template updates (no propagation)
- B: Document conventions (rely on developer discipline)
- C: Validate templates only (not generated configs)
- **D: Propagate template patterns to all generated configs with validation** ✓ **SELECTED**

**Decision**: Option D selected - Full propagation with validation

**Rationale**:
- Changes to deployments/template/ MUST propagate to ALL PRODUCT/SUITE configs
- Validation (Decision 3) enforces template compliance
- Prevents drift between template and generated configs
- Automated validation catches violations immediately

**Alternatives Rejected**:
- Option A: Manual updates error-prone, drift guaranteed
- Option B: Developer discipline unreliable
- Option C: Template-only validation misses generated config violations

**Impact**:
- Phase 2: All PRODUCT/SUITE configs MUST follow template patterns
- Phase 3: Template validation checks ALL configs against template rules
- Template updates trigger config regeneration + validation

**Evidence**: User selected "D" in quizme-v1.md Q7

---

### Decision 8: README.md Content Requirements (quizme-v1 Q8)

**Options**:
- **A: Minimal (purpose, delegation, link to ARCHITECTURE.md)** ✓ **SELECTED**
- B: Comprehensive (purpose, delegation, config examples, secret sharing, port offsets)
- C: Template-based (use template from deployments/template/README.md)
- D: Generated (auto-generate from compose.yml and config files)

**Decision**: Option A selected - Minimal README.md

**Rationale**:
- Quick overview sufficient for PRODUCT/SUITE directories
- Detailed docs live in ARCHITECTURE.md (single source of truth)
- Reduces maintenance burden (fewer docs to keep in sync)
- ARCHITECTURE.md link provides comprehensive reference

**Alternatives Rejected**:
- Option B: Comprehensive README.md duplicates ARCHITECTURE.md content
- Option C: Template README.md too detailed for simple delegation configs
- Option D: Generated docs miss context and rationale

**Impact**:
- Phase 2: README.md tasks create minimal content only
- Content: Purpose paragraph + delegation pattern + ARCHITECTURE.md link
- Reduced Phase 2 LOE (6h vs 10h for comprehensive READMEs)

**Evidence**: User selected "A" in quizme-v1.md Q8

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Large file restructure introduces errors | Medium | High | Validate after each task, incremental commits, comprehensive tests |
| CICD validation too strict breaks workflows | Medium | High | Validate against ALL existing deployments first, adjust rules if needed |
| Template pattern incomplete for all products | Low | Medium | Review deployments/template/ thoroughly before Phase 2 |
| ARCHITECTURE.md updates not propagated fully | Medium | Medium | Automated grep checks for instruction file references, manual review |
| Pre-commit performance impact | Low | Low | Optimize validation code, cache validation results where possible |
| Config schema changes during implementation | Low | High | Version CONFIG-SCHEMA.md, validate backward compatibility |

---

## Quality Gates - MANDATORY

**Per-Task Quality Gates**:
- ✅ All tests pass (`go test ./...`) - 100% passing, zero skips
- ✅ Build clean (`go build ./...`) - zero errors
- ✅ Linting clean (`golangci-lint run`) - zero warnings
- ✅ No new TODOs without tracking in tasks.md

**Coverage Targets** (from copilot instructions):
- ✅ Production code: ≥95% line coverage
- ✅ Infrastructure/utility code (CICD): ≥98% line coverage
- ✅ main() functions: 0% acceptable if internalMain() ≥95%
- ✅ Generated code: Excluded (OpenAPI stubs, GORM models)

**Mutation Testing Targets** (from copilot instructions):
- ✅ Production code: ≥95% minimum, ≥98% ideal
- ✅ Infrastructure/utility code (CICD): ≥98% (NO EXCEPTIONS)

**Per-Phase Quality Gates**:
- ✅ Unit tests complete before moving to next phase
- ✅ Integration tests pass (config validation against real files)
- ✅ All code references updated (no broken imports)
- ✅ Documentation updated (README.md, ARCHITECTURE.md)

**Overall Project Quality Gates**:
- ✅ All phases complete with evidence
- ✅ All test categories passing (unit, integration)
- ✅ Coverage and mutation targets met
- ✅ All configs/ and deployments/ pass validation
- ✅ Pre-commit hooks functional
- ✅ ARCHITECTURE.md and instruction files in sync

---

## Success Criteria

- [ ] configs/ mirrors deployments/ structure (SERVICE, PRODUCT, SUITE levels)
- [ ] All 9 SERVICE subdirs exist: cipher-im/, pki-ca/, identity-{authz,idp,rp,rs,spa}/, sm-kms/, jose-ja/
- [ ] All 5 PRODUCT dirs exist: cipher/, pki/, identity/, sm/, jose/
- [ ] SUITE dir exists: cryptoutil/
- [ ] All PRODUCT/SUITE configs follow template patterns
- [ ] CICD validation implements all 8 types (naming, kebab-case, schema, ports, telemetry, admin, consistency, secrets)
- [ ] All configs/ pass validation (100%)
- [ ] All deployments/ pass validation (100%)
- [ ] ARCHITECTURE.md sections 12.4-12.6 complete
- [ ] Instruction files updated with ARCHITECTURE.md patterns
- [ ] Pre-commit hooks functional
- [ ] Tests pass: `go test ./...` (≥98% coverage for CICD code)
- [ ] Mutation testing: ≥98% for CICD code
- [ ] Build clean: `go build ./...`
- [ ] Linting clean: `golangci-lint run`
- [ ] Evidence archived: `test-output/fixes-v3/`

---

## Evidence Archive

- `test-output/fixes-v3-quizme-analysis/` - Quizme-v1 answers and impact analysis
- `test-output/phase1/` - configs/ restructuring verification
- `test-output/phase2/` - PRODUCT/SUITE config creation
- `test-output/phase3/` - CICD validation implementation + testing
- `test-output/phase4/` - ARCHITECTURE.md updates verification
- `test-output/phase5/` - Instruction file propagation checks
- `test-output/phase6/` - E2E validation results
