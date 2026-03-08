# Lessons Learned - Framework Brainstorm Execution

**Purpose**: Persistent memory for the framework brainstorm execution plan. Records
what worked, what did not, root causes, and patterns for use throughout the plan and
for knowledge propagation to ARCHITECTURE.md, agents, skills, and instructions.

---

## Phase 1: Framework v1 Closure

**Completed**: 2026-03-08

### What Worked

- **Bulk PowerShell replace** for updating 48 stale `❌` statuses across 4 duplicate blocks in tasks.md — far faster than per-occurrence string replacement
- **`replace_string_in_file` for unique context** — using surrounding unique content (file paths, task descriptions) as the oldString discriminator works well for single-block updates
- **ARCHITECTURE.md multi-section edits in sequence** — doing all 6 Phase 8.2 edits in one session before committing prevents partial-state commits
- **`lint-docs` as propagation gate** — `go run ./cmd/cicd lint-docs` caught the stale `@source` mismatch (10→11 linters) before commit

### What Didn't Work

- **Token budget mid-edit** — Session was cut off mid-way through ARCHITECTURE.md Section 13.5.5 insertion; `@propagate` marker was accidentally corrupted during header restore after a failed edit attempt
- **`replace_string_in_file` with duplicate blocks** — Cannot use when same block appears multiple times; must use PowerShell bulk replace or add surrounding unique context
- **Race detector on Windows** — `go test -race` requires `CGO_ENABLED=1` + GCC; GCC not available on Windows dev machines; Linux CI/CD is the required platform for race detection

### Root Causes

- **4 duplicate Phase 5-8 blocks in tasks.md**: Each session appended a new completion block rather than editing in place; creates maintenance overhead but is acceptable as long as the LAST block is the authoritative one
- **propagate marker corruption**: `replace_string_in_file` with a long oldString that accidentally included the separator line merged the header with the `@propagate to` text; fix is to verify both before and after the header

### Patterns Observed

- **Pre-commit failures on Windows**: `check-todo-severity` (WSL bash), `actionlint` (intentional `if: false`), `hadolint` (Docker daemon) are all pre-existing and not caused by our work
- **ARCHITECTURE.md Section 13.5.4**: The `@propagate` marker was corrupted by a prior session fix; always verify the FULL marker line after any edit to headers near propagate markers
- **tasks.md duplicate block pattern**: When a task file accumulates iterative blocks, PowerShell `$content -replace` is more reliable than per-occurrence string matching

### Knowledge Propagation Done

- Section 9.11: Architecture Fitness Functions (23 sub-linters)
- Section 9.10.2: lint-fitness count 10→11
- Section 9.10.4: lint_fitness/ directory structure
- Section 10.3.5: testserver.StartAndWait (was SetupTestServer typo)
- Section 10.3.6: Shared Test Infrastructure APIs
- Section 13.5.5: Air Live Reload (SERVICE=sm-im air)
- `.github/instructions/04-01.deployment.instructions.md`: @source cicd-command-naming updated (10→11)

---

## Phase 2: Copilot Skills Audit

*(To be filled after Phase 2 completes)*

---

## Phase 3: Copilot Agents Audit

*(To be filled after Phase 3 completes)*

---

## Phase 4: Instructions Audit

*(To be filled after Phase 4 completes)*

---

## Phase 5: ARCHITECTURE.md Completeness

*(To be filled after Phase 5 completes)*

---

## Cross-Phase Lessons

### Session 1 (2026-03-07): Context Recovery + Agent Updates

**What was found**:
1. Framework v1 implementation is 58% complete (28/48 tasks). Phase 5 shared test
   packages actually exist in the codebase but task status fields were never updated.
2. framework-brainstorm/ had 9 research docs but no plan.md or tasks.md — the
   brainstorm research was disconnected from an executable plan.
3. SKILL.md files had invalid `metadata: domain:` frontmatter blocks not in VS Code spec.
   These caused YAML validation warnings.

**What was done**:
1. Fixed: removed `metadata:` blocks from all 14 SKILL.md files (commit 2acabc715)
2. Added: lessons.md as 5th companion document to implementation-planning and implementation-execution agents
3. Added: post-mortem step to every phase template in implementation-planning agent
4. Expanded: implementation-execution Phase-Based Post-Mortem to include code/tests/workflows/docs
5. Expanded: ARCHITECTURE.md 13.8.1 and 13.8.2 to include all artifact types
6. Added: Semantic Grouping & Periodic Commits directives to beast-mode, doc-sync, fix-workflows agents
7. Created: framework-brainstorm/plan.md and tasks.md (this session)

**Root Causes**:
- SKILL.md `metadata:` blocks: VS Code agent skills spec was not consulted during initial skill creation; a project-specific convention was invented without verifying against the actual spec
- Disconnected brainstorm docs: The brainstorm research was done but never translated into an executable plan with task tracking
- Stale task status: Phase 5 packages were implemented as part of tasks 5.7/5.8 work but individual task statuses 5.1-5.6 were not back-filled

**Patterns**:
- Always verify toolkit/platform specs (VS Code, GitHub, etc.) before adding custom conventions
- Research docs MUST be followed by plan.md + tasks.md before ending a session
- When a "higher-level" task (5.7: migrate) is marked done, verify that "lower-level" prerequisite tasks (5.1-5.6: create) are also marked done

**Candidates for Knowledge Propagation**:
- SKILL.md frontmatter: ARCHITECTURE.md 2.1.5 already updated ✅
- lessons.md as 5th companion doc: ARCHITECTURE.md 13.8 and agents already updated ✅
- Post-mortem across all artifact types: ARCHITECTURE.md 13.8.1/13.8.2 and agents already updated ✅
- Periodic commits: ARCHITECTURE.md 13.2.2 already has it; propagated to agents ✅
- lint-fitness not in ARCHITECTURE.md: Task 1.3 / Phase 5 Task 5.1 to address
- Shared test infra packages not in ARCHITECTURE.md: Task 1.3 / Phase 5 Task 5.2 to address

---

## Phase 2 Post-Mortem: Copilot Skills Completeness Audit

**What Worked**:
1. Reading all 14 skills in one pass gave complete picture before making changes.
2. Grouping findings by artifact type made gaps obvious.
3. `multi_replace_string_in_file` with 7 items handled all Phase 2 skill changes in one call — efficient.

**What Was Found (6 skills needed updates)**:
1. `test-table-driven`: Missing `Section 10.3.6 Shared Test Infrastructure` reference.
2. `coverage-analysis`: Missing `Main Functions | 0%` row in coverage targets table.
3. `migration-create`: Missing CONFIG-SCHEMA.md cross-update note for new domain config keys.
4. `propagation-check`: Did not clarify that agent files have `@source` blocks; missing doc-sync cross-ref.
5. `new-service`: Missing Step 9 (update ARCHITECTURE.md service catalog after creating service).
6. `fitness-function-gen`: Missing Section 9.11 reference (23 sub-linters in 3 groups).

**What Was Complete (8 skills, no changes needed)**:
`test-fuzz-gen`, `test-benchmark-gen`, `fips-audit`, `openapi-codegen`, `agent-scaffold`, `instruction-scaffold`, `skill-scaffold`, `contract-test-gen`.

**Root Causes**:
- Skills were created as focused "how-to" guides without checking if their outputs affect other artifacts.
- `propagation-check` assumed only instruction files have `@source` blocks; agents were added later and the skill was not updated.
- `new-service` stopped at code/tests/config/compose/CI without checking docs/ARCHITECTURE.md.

**Patterns**:
- When creating a skill, ask: does following this skill's steps affect ARCHITECTURE.md, instruction files, agent files, config files, or deployment files? If yes, add a post-creation checklist.

---

## Phase 3 Post-Mortem: Copilot Agents Completeness Audit

**What Worked**:
1. Reading full agent files sequentially gave complete picture of what was missing.
2. All 5 agents had the same class of gap: config/deployments missing from post-mortem scope.
3. Targeted `replace_string_in_file` per agent was precise and fast.

**What Was Found (5 agents needed updates)**:
1. `beast-mode`: Requirements Validation missing config/deployments + cross-artifact consistency self-check.
2. `doc-sync`: Step 3 listed instruction files only — missing agents and skill files as sync targets.
3. `fix-workflows`: No scope section; implied only `.github/workflows/*.yml` in scope.
4. `implementation-execution`: Phase-Based Post-Mortem missing config/deployments bullets.
5. `implementation-planning`: Phase Post-Mortem Self-Evaluation missing config/deployments bullets.

**Root Causes**:
- Agents were created incrementally; each time a new artifact type was added to the framework, not all agents were checked.
- `doc-sync` Step 3 was updated for instruction files but not revisited when agent/skill propagation was added.

**Patterns**:
- After adding any new artifact type to the framework, audit ALL 5 agents for scope completeness.
- `doc-sync` Step 3 is the canonical "document update order" and must always reflect the full propagation chain.

---

## Phase 4 Post-Mortem: Copilot Instructions Completeness Audit

**What Worked**:
1. The 6 "02-xx" files had no gaps — cross-artifact by design.
2. Targeted reading of the 4 known-gap files was efficient.
3. lint-docs validate-propagation immediately verified the @source sync was correct.

**What Was Found (4 files needed updates)**:
1. `03-02.testing`: `testdb`, `testserver`, `fixtures`, `assertions`, `healthclient` packages not mentioned.
2. `03-04.data-infrastructure`: testdb/testserver helpers not in Test Compatibility section.
3. `06-01.evidence-based`: Missing Config/Deployments and Docs in Mandatory Evidence checklist.
4. `06-02.agent-format`: Agent Self-Containment Checklist missing bullet for docs/skills/agents/instructions modifications.

**Key Technical Detail — @source blocks in agent files**:
`06-02.agent-format` has a `<!-- @source -->` block. Updating it requires updating ARCHITECTURE.md `<!-- @propagate -->` FIRST, then syncing `@source` to match byte-for-byte. lint-docs validate-propagation enforces this.

**Root Causes**:
- Shared test infrastructure was added to ARCHITECTURE.md in Phase 8.2 but instruction files were not checked for sync.
- Mandatory Evidence checklist predated config/deployments validation becoming a first-class quality gate.

**Patterns**:
- After updating ARCHITECTURE.md with new content, always check instruction files for content that should be added.
- When adding a new quality gate command, update 06-01.evidence-based immediately.

---

## Phase 5 Post-Mortem: ARCHITECTURE.md Cross-Artifact Completeness

**Completed in commit 75edd8400 (prior session)**. See framework-v1 tasks.md for full evidence.

**Summary**: Section 9.11 (23 fitness function sub-linters), Section 10.3.6 (shared test infra APIs), Section 13.5.5 (air live reload) — all added and propagated.

---

## Phase 6 Post-Mortem: Knowledge Propagation Final

**What Worked**:
1. Running lint-docs as first quality gate caught @source drift immediately.
2. Sequential commits per artifact type (skills → agents → instructions) enabled clean bisect history.
3. 3-pass review caught Tasks 4.1 checkbox acceptance criteria as needing verification before marking complete.

**Final Quality Gate Results**:
- `go run ./cmd/cicd lint-docs`: 3/3 sub-linters pass.
- `go build ./...`: exit 0.
- All 38 tasks: ✅ complete.

**Overall Execution Summary**:
- 6 skills updated for cross-artifact completeness.
- 5 agents updated for config/deployments/scope completeness.
- 4 instruction files updated for shared test infrastructure and evidence checklist.
- ARCHITECTURE.md: `agent-self-containment` @propagate block updated + synced to @source.
- All documentation now consistently covers all artifact types: code, tests, config, deployments, docs, skills, agents, instructions, CI/CD workflows.
