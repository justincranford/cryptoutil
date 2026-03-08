# Tasks - Framework Brainstorm Execution

**Status**: 7 of 38 tasks complete (18%)
**Created**: 2026-03-07
**References**: docs/framework-brainstorm/plan.md, docs/framework-v1/tasks.md

## Quality Mandate - MANDATORY

- ✅ **Correctness**: ALL changes accurate and semantically correct
- ✅ **Completeness**: NO phases or tasks skipped
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: build, lint, tests, lint-docs all pass
- ✅ **Accuracy**: Root cause addressed, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation

---

## Task Checklist

### Phase 1: Framework v1 Closure

**Phase Objective**: Complete all unclosed Framework v1 work.

#### Task 1.1: Update Stale Task Status Fields

- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Description**: In framework-v1/tasks.md, update Phase 5 tasks 5.1-5.6 from ❌ to ✅.
  All packages exist in `internal/apps/template/service/testing/` but status wasn't updated.
- **Acceptance Criteria**:
  - [x] Tasks 5.1-5.6 updated from ❌ to ✅ with actual completion evidence
  - [x] Task count header updated from 28/48 to 48/48 (100%)
  - [x] Acceptance criteria checkboxes checked for existing implementations

#### Task 1.2: Run Phase 7 Quality Gates

- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Description**: Execute all Phase 7 acceptance criteria: tests passing, coverage, lint.
- **Acceptance Criteria**:
  - [x] `go test ./...` — pre-existing Windows failures documented (TestFix_WalkDirError/WalkError, chmod doesn't block reads); all CI/CD (Linux) passes
  - [x] `golangci-lint run` clean
  - [x] `golangci-lint run --build-tags e2e,integration` clean
  - [x] `go run ./cmd/cicd lint-fitness` passes
  - [x] New shared test infrastructure packages at ≥98% coverage (assertions/fixtures/healthclient/testserver 100%)
  - [x] Evidence recorded in framework-v1/tasks.md Phase 7 section

#### Task 1.3: Complete Phase 8.2 - Update ARCHITECTURE.md

- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Description**: Add missing ARCHITECTURE.md documentation from framework-v1 patterns.
- **Acceptance Criteria**:
  - [x] Architecture fitness functions documented: Section 9.11 added
    - lint-fitness command added to Section 9.10.2 (11 total linters)
    - 23 sub-linters in 3 groups documented in Section 9.11
    - Pre-commit hook + CI/CD integration documented
  - [x] Shared test infrastructure documented: Section 10.3.6 added
    - `testdb.NewInMemorySQLiteDB(t)` documented
    - `testserver.StartAndWait(ctx, t, srv)` documented
    - `fixtures.CreateTestTenant/Realm/User` documented
    - `assertions.AssertHealthy/AssertErrorResponse` documented
    - `healthclient.NewHealthClient` documented
  - [x] Air live reload documented: Section 13.5.5 added
  - [x] `go run ./cmd/cicd lint-docs` passes (validate-propagation: 0 broken refs)

#### Task 1.4: Verify Phase 8.3 - Skills Already Created

- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Description**: Both skills required by Phase 8.3 already exist.
- **Acceptance Criteria**:
  - [x] `contract-test-gen/SKILL.md` exists
  - [x] `fitness-function-gen/SKILL.md` exists

#### Task 1.5: Verify Phase 8.4 - Agents Updated

- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Description**: Both agents were updated in this session with post-mortem scope.
- **Acceptance Criteria**:
  - [x] implementation-planning.agent.md: post-mortem added, Knowledge Propagation expanded
  - [x] implementation-execution.agent.md: Phase-Based Post-Mortem expanded

#### Task 1.6: Run Phase 8.5 - Verify Propagation and Final Commit

- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Description**: Verify propagation passes, mark framework-v1 tasks complete.
- **Acceptance Criteria**:
  - [x] `go run ./cmd/cicd lint-docs` passes (all 3 sub-linters: chunk-verify, validate-chunks, validate-propagation)
  - [x] `go build ./...` clean
  - [x] framework-v1/tasks.md updated to 48/48 complete
  - [x] framework-v1/plan.md marked COMPLETE
  - [x] Git commits: `docs(framework-v1): mark all Phase 5+7 tasks complete, update task statuses to 48/48`

#### Task 1.7: Phase 1 Post-Mortem

- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Description**: Update lessons.md with Phase 1 lessons.
- **Acceptance Criteria**:
  - [x] lessons.md updated with: what worked, what didn't, root causes, patterns
  - [x] Evaluate lessons for immediate ARCHITECTURE.md contradictions/omissions

---

### Phase 2: Copilot Skills Completeness Audit

**Phase Objective**: Audit all 14 skills for cross-artifact completeness.

#### Task 2.1: Audit test-table-driven, test-fuzz-gen, test-benchmark-gen

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45m
- **Description**: Check if these testing skills mention when their patterns apply to
  test helper code (not just domain code tests). Verify they don't assume code-only context.
- **Acceptance Criteria**:
  - [ ] Each skill reviewed against checklist: code ✓, tests ✓, implicit code-only assumptions identified
  - [ ] Skills updated where cross-artifact implications are missing
  - [ ] Rationale documented for code-only skills

#### Task 2.2: Audit coverage-analysis, migration-create, fips-audit

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45m
- **Description**: Check these workflow/utility skills for cross-artifact scope.
  - coverage-analysis: does it cover test helper packages?
  - migration-create: does it reference docs/CONFIG-SCHEMA.md updates?
  - fips-audit: is it clearly scoped to Go code only?
- **Acceptance Criteria**:
  - [ ] Each skill reviewed and updated where needed
  - [ ] migration-create explicitly mentions CONFIG-SCHEMA.md cross-update if applicable

#### Task 2.3: Audit propagation-check, openapi-codegen

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45m
- **Description**: These skills span multiple artifact types already. Verify they're complete.
  - propagation-check: covers ARCHITECTURE.md → instruction files → agents?
  - openapi-codegen: covers spec, server gen, model gen, client gen?
- **Acceptance Criteria**:
  - [ ] Each skill reviewed for completeness across artifact types it claims to cover
  - [ ] Cross-references to related skills present (e.g., propagation-check → doc-sync agent)

#### Task 2.4: Audit agent-scaffold, instruction-scaffold, skill-scaffold

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45m
- **Description**: These scaffolding skills create artifacts. Verify they cross-reference:
  - agent-scaffold: mentions updating copilot-instructions.md skills table?
  - instruction-scaffold: mentions updating copilot-instructions.md instruction table?
  - skill-scaffold: mentions updating copilot-instructions.md skills table + ARCHITECTURE.md skills catalogue?
- **Acceptance Criteria**:
  - [ ] Each scaffolding skill has explicit post-creation checklist for cross-artifact updates
  - [ ] Skills updated where checklist is missing

#### Task 2.5: Audit new-service, contract-test-gen, fitness-function-gen

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45m
- **Description**: These are framework-specific skills. Verify completeness:
  - new-service: covers ALL artifact types (service code, tests, config, compose, CI, docs)?
  - contract-test-gen: covers all 3 contract groups (health, server isolation, response format)?
  - fitness-function-gen: covers all 8 fitness function categories?
- **Acceptance Criteria**:
  - [ ] new-service skill has explicit checklist covering: service code, tests, compose, config, CI workflow, docs
  - [ ] contract-test-gen references all 3 RunContractTests contract groups
  - [ ] fitness-function-gen references all 8 fitness function categories

#### Task 2.6: Phase 2 Quality Gate

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Acceptance Criteria**:
  - [ ] All 14 skills reviewed (Tasks 2.1-2.5)
  - [ ] All updates committed with conventional commits
  - [ ] `go run ./cmd/cicd lint-docs validate-propagation` passes (if ARCHITECTURE.md skills catalogue updated)

#### Task 2.7: Phase 2 Post-Mortem

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Acceptance Criteria**:
  - [ ] lessons.md updated with Phase 2 observations
  - [ ] Contradictions/omissions identified → fix tasks created immediately

---

### Phase 3: Copilot Agents Completeness Audit

**Phase Objective**: Audit all 5 agents for cross-artifact post-mortem scope.

#### Task 3.1: Audit beast-mode.agent.md

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Description**: Verify beast-mode covers all artifact types in quality gate checklist
  and commit strategy. Note: Semantic Grouping & Periodic Commits added this session.
- **Acceptance Criteria**:
  - [ ] Quality gate checklist includes: code, tests, config, deployments, docs
  - [ ] Commit strategy mentions: one artifact type changed = one commit (not just code commits)
  - [ ] Self-evaluation criteria mention checking docs/skills/agents/instructions for consistency

#### Task 3.2: Audit doc-sync.agent.md

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Description**: Verify doc-sync covers all document types in its sync workflow.
- **Acceptance Criteria**:
  - [ ] Sync workflow explicitly covers: ARCHITECTURE.md → instruction files → agent files → skill files
  - [ ] Step 5 commit guidance includes per-type semantic grouping
  - [ ] Anti-patterns section covers common doc-sync failure modes

#### Task 3.3: Audit fix-workflows.agent.md

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Description**: Verify fix-workflows covers all CI/CD artifact types (not just workflow .yml files).
- **Acceptance Criteria**:
  - [ ] Scope covers: workflow files, docker files, compose files, pre-commit config
  - [ ] Semantic Grouping guidance covers config file changes, not just workflow file changes

#### Task 3.4: Audit implementation-execution.agent.md

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Description**: Verify post-mortem and knowledge extraction cover all artifact types.
  Note: expanded this session to include code/tests/workflows/docs.
- **Acceptance Criteria**:
  - [ ] Phase-Based Post-Mortem covers: code, tests, config, deployments, workflows, docs, agents, skills, instructions
  - [ ] Extract Lessons covers: ARCHITECTURE.md, agents, skills, instructions, code, tests, workflows

#### Task 3.5: Audit implementation-planning.agent.md

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Description**: Verify phase templates and knowledge propagation cover all artifact types.
  Note: expanded this session to include code/tests/workflows/docs.
- **Acceptance Criteria**:
  - [ ] Every phase template post-mortem step mentions: code, tests, config, deployments, workflows, docs
  - [ ] Knowledge Propagation Phase covers all 9 artifact types

#### Task 3.6: Phase 3 Quality Gate + Post-Mortem

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Acceptance Criteria**:
  - [ ] All 5 agents reviewed (Tasks 3.1-3.5) and updated where needed
  - [ ] All updates committed
  - [ ] lessons.md updated with Phase 3 observations

---

### Phase 4: Copilot Instructions Completeness Audit

**Phase Objective**: Audit all 18 instruction files for cross-artifact consistency.

#### Task 4.1: Audit architecture + security + authn + observability + openapi + versions

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Description**: Review 6 instruction files for implicit code-only rules that should
  apply to config/docs/deployments but don't.
- **Acceptance Criteria**:
  - [ ] 02-01.architecture: patterns apply to ALL service artifacts (code, config, compose, docs)
  - [ ] 02-05.security: secret rules cover config files, not just code
  - [ ] 02-03.observability: telemetry patterns documented for config AND code
  - [ ] No implicit code-only assumptions found without explicit scope

#### Task 4.2: Audit coding + testing + golang + data-infrastructure + linting

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Description**: Review 5 instruction files for gaps in test helper/infra coverage.
- **Acceptance Criteria**:
  - [ ] 03-02.testing: shared test infrastructure packages mentioned (testdb, testserver, fixtures, assertions, healthclient)
  - [ ] 03-02.testing: cross-service contract test pattern documented
  - [ ] 03-02.testing: TestMain integration pattern mentions SetupTestServer helper
  - [ ] 03-04.data-infrastructure: testdb helper referenced

#### Task 4.3: Audit deployment + git + cross-platform + evidence-based + beast-mode + agent-format

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45m
- **Description**: Review remaining instruction files for cross-artifact gaps.
- **Acceptance Criteria**:
  - [ ] 04-01.deployment: config schema validation documented (lint-deployments)
  - [ ] 05-02.git: periodic commits language propagated from ARCHITECTURE.md 13.2.2
  - [ ] 06-01.evidence-based: evidence checklist covers docs/config/deployments not just code
  - [ ] 06-02.agent-format: agent self-containment checklist references all artifact types

#### Task 4.4: Phase 4 Quality Gate + Post-Mortem

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Acceptance Criteria**:
  - [ ] All 18 instruction files reviewed (Tasks 4.1-4.3) and updated where needed
  - [ ] `go run ./cmd/cicd lint-docs validate-propagation` passes
  - [ ] All updates committed
  - [ ] lessons.md updated with Phase 4 observations

---

### Phase 5: ARCHITECTURE.md Cross-Artifact Completeness

**Phase Objective**: Document framework-v1 patterns in ARCHITECTURE.md; ensure all
patterns cover ALL artifact types.

#### Task 5.1: Document Architecture Fitness Functions

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Description**: Add lint-fitness documentation to ARCHITECTURE.md.
- **Files**: ARCHITECTURE.md Section 9.10 or new Section 11.4
- **Acceptance Criteria**:
  - [ ] lint-fitness command documented: `go run ./cmd/cicd lint-fitness`
  - [ ] All 8 sub-linters described: service-contract-compliance, parallel-tests,
    sequential-test-comment, admin-bind-policy, health-endpoints, port-assignments,
    import-isolation, file-size
  - [ ] Pre-commit hook integration documented
  - [ ] CI/CD workflow integration mentioned
  - [ ] @propagate markers added for propagation to relevant instruction files

#### Task 5.2: Document Shared Test Infrastructure

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Description**: Add shared test infrastructure packages to ARCHITECTURE.md Section 10.3.
- **Acceptance Criteria**:
  - [ ] `testdb.NewInMemorySQLiteDB(t *testing.T) *gorm.DB` documented
  - [ ] `testdb.NewPostgresTestContainer(ctx, t)` documented
  - [ ] `testserver.SetupTestServer(t, constructor)` documented
  - [ ] `fixtures.CreateTestTenant/Realm/User` documented
  - [ ] `assertions.AssertHealthy/AssertErrorResponse/AssertJSONContentType` documented
  - [ ] `healthclient.NewHealthClient` documented

#### Task 5.3: Document Air Live Reload

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Description**: Add air live reload to ARCHITECTURE.md Section 13 (Development Practices).
- **Acceptance Criteria**:
  - [ ] Brief section explaining air live reload
  - [ ] Command: `SERVICE=sm-im air` documented
  - [ ] `.air.toml` structure mentioned
  - [ ] Link to air live reload docs

#### Task 5.4: Phase 5 Quality Gate + Post-Mortem

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd lint-docs validate-propagation` passes
  - [ ] `go build ./...` clean
  - [ ] Updates committed with semantic commits
  - [ ] lessons.md updated

---

### Phase 6: Knowledge Propagation

**Phase Objective**: Ensure all changes from Phases 1-5 are consistently propagated
to @source blocks in instruction files.

#### Task 6.1: Run validate-propagation and Fix Drift

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd lint-docs validate-propagation` passes with 0 errors
  - [ ] Any drift in @source blocks corrected to match ARCHITECTURE.md
  - [ ] All fixes committed

#### Task 6.2: Final Quality Gates

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Acceptance Criteria**:
  - [ ] `go build ./...` exits 0
  - [ ] `golangci-lint run` exits 0
  - [ ] `go test ./...` exits 0 (100%, zero skips)
  - [ ] plan.md success criteria all checked
  - [ ] Git commit: `docs(framework-brainstorm): complete execution plan phases 1-6`

#### Task 6.3: Phase 6 Post-Mortem

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Acceptance Criteria**:
  - [ ] lessons.md final update
  - [ ] Remaining deferred items documented in Phase 7

---

### Phase 7: Future Brainstorm Items (Reference Tasks)

**Phase Objective**: Document deferred items with actionable context for future plans.
These tasks are NOT executed now — they are reference stubs for future plans.

#### Task 7.1: Document cicd new-service Scaffolding (P1-1)

- **Status**: 📋 DEFERRED (not executing now)
- **Trigger**: When 10th+ service needs to be added
- **Effort**: 2-3 weeks
- **Approach**: See docs/framework-brainstorm/08-recommendations.md P1-1 for design.
  Use `text/template` + skeleton-template as source, generate into `internal/apps/PRODUCT/SERVICE/`.

#### Task 7.2: Document Extract Framework Module (P3-2)

- **Status**: 📋 DEFERRED (not executing now)
- **Trigger**: When external consumers of the framework are identified, OR when >15 services exist
- **Effort**: 1-4 weeks
- **Approach**: Move `internal/apps/template/` to separate Go module `github.com/user/cryptoutil-framework`.

---

## Evidence Archive

- `test-output/framework-brainstorm/phase1/` - Phase 1 evidence
- `test-output/framework-brainstorm/phase2/` - Skills audit evidence
- `docs/framework-brainstorm/lessons.md` - Persistent lessons memory
