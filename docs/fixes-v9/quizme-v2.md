# Fixes v9 - Quiz v2 - Copilot Skills Candidates

**Instructions**: For each question, mark `[x]` for your choice. Multiple selections allowed where noted.

**Context**: VS Code Copilot Skills live in `.github/skills/SKILLNAME.md`. A skill is an on-demand specialized prompt invoked by referencing `#SKILLNAME` in chat or via agent `skills:` frontmatter. Unlike instructions (always loaded) and agents (full autonomous workflows), skills are focused, single-purpose prompt templates with embedded examples/templates.

**Example skill file structure**:
```
.github/skills/
  test-table-driven.md        ← skill definition
  test-table-driven/          ← optional examples
    example_test.go
```

**Example SKILL.md format** (abridged):
```markdown
---
name: test-table-driven
description: Generate table-driven Go tests following project conventions
---
Generate a table-driven Go test for the given function. Follow these rules:
- t.Parallel() on all tests and subtests
- Use googleUuid.NewV7() for test IDs, never hardcoded UUIDs
- Use require over assert
- [example follows...]
```

---

## Section 0: Skills Infrastructure

### Q0: .github/skills/ Directory

Should we create the `.github/skills/` directory and infrastructure regardless of individual skill decisions?

- [ ] **A)** YES — create directory + README.md + SKILL-TEMPLATE.md template even before implementing any skills
- [ ] **B)** NO — create it only when first skill is approved
- [ ] **C)**
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

## Section 1: Group A — Test Generation Skills

### Q1: test-table-driven skill

**Purpose**: Generate table-driven Go tests following project conventions (t.Parallel, googleUuid.NewV7, require over assert, subtests).

**Would create**:
```
.github/skills/test-table-driven.md
.github/skills/test-table-driven/example_test.go   ← shows pattern
```

**What it does** when invoked: Given a function signature, generates complete table-driven test file with correct imports (cryptoutil alias pattern), UUID test data, parallel subtests, error checking patterns.

**vs. instructions**: 03-02.testing.instructions.md already describes the PATTERN. A skill would add: embedded ready-to-copy CODE TEMPLATE.

- [ ] **A)** YES — implement
- [ ] **B)** NO — testing instructions are sufficient, skill adds no value
- [ ] **C)** DEFER — implement after seeing it used in practice first
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

### Q2: test-fuzz-gen skill

**Purpose**: Generate fuzz test files following project conventions (file suffix `_fuzz_test.go`, minimum 15s fuzz time, corpus examples).

**Would create**:
```
.github/skills/test-fuzz-gen.md
.github/skills/test-fuzz-gen/example_fuzz_test.go
```

**What it does**: Given a function to fuzz, generates `_fuzz_test.go` with seed corpus, correct build tags, fuzz target registration.

- [ ] **A)** YES — implement
- [ ] **B)** NO — testing instructions sufficient
- [ ] **C)** DEFER — implement after seeing how fuzz skill is used
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

### Q3: test-benchmark-gen skill

**Purpose**: Generate benchmark test files following project conventions (file suffix `_bench_test.go`, mandatory for crypto operations, reset timer patterns).

**Would create**:
```
.github/skills/test-benchmark-gen.md
.github/skills/test-benchmark-gen/example_bench_test.go
```

- [ ] **A)** YES — implement
- [ ] **B)** NO — testing instructions sufficient
- [ ] **C)** DEFER
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

## Section 2: Group B — Infrastructure / Deployment Skills

### Q4: compose-validator skill (wraps cicd lint-deployments)

**Purpose**: Validate Docker Compose files and deployment structure. Wraps `go run ./cmd/cicd lint-deployments` but adds interactive guidance on fixing violations.

**Would create**:
```
.github/skills/compose-validator.md
```

**What it does**: Runs validation, interprets errors, suggests specific fixes with code examples. Goes beyond just running the command — explains WHY a rule exists.

**vs. lint-deployments**: The cicd command finds problems; the skill explains AND fixes them.

- [ ] **A)** YES — implement
- [ ] **B)** NO — `cicd lint-deployments` error messages are self-explanatory
- [ ] **C)** DEFER — evaluate after seeing concrete use cases
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

### Q5: migration-create skill

**Purpose**: Create properly numbered golang-migrate migration files with correct naming, paired up/down files, and standard boilerplate.

**Would create**:
```
.github/skills/migration-create.md
```

**What it does**: Given a migration description (e.g., "add sessions table"), generates:
- `migrations/1005_add_sessions.up.sql` (next available number in correct range: template 1001-1999, domain 2001+)
- `migrations/1005_add_sessions.down.sql`
- Standard SQL boilerplate with correct constraints

**vs. instructions**: 03-04.data-infrastructure.instructions.md describes numbering rules but doesn't generate files.

- [ ] **A)** YES — implement
- [ ] **B)** NO — the numbering rules are simple enough to follow manually
- [ ] **C)** DEFER
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

### Q6: service-scaffold skill

**Purpose**: Scaffold a new service following the template builder pattern (all required files, correct directory structure, registered routes).

**Would create**:
```
.github/skills/service-scaffold.md
.github/skills/service-scaffold/
  template_service.go
  template_handler.go
  template_repository.go
```

**What it does**: Given a service name (e.g., "sm-im"), generates all scaffolding files following `internal/apps/template/service/` pattern.

**vs. implementation-planning agent**: agent creates plan.md/tasks.md; skill creates actual code scaffold.

- [ ] **A)** YES — implement
- [ ] **B)** NO — too complex for a skill; use implementation-planning agent instead
- [ ] **C)** DEFER
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

## Section 3: Group C — Code Quality Skills

### Q7: coverage-analysis skill

**Purpose**: Analyze test coverage output, identify which lines are uncovered, categorize them (unreachable vs. testable), and generate targeted test suggestions.

**Would create**:
```
.github/skills/coverage-analysis.md
```

**What it does**: Given a `go test -coverprofile` output, identifies RED lines, categorizes them (error paths, test seam candidates, unreachable), and generates specific test cases to improve coverage.

**vs. instructions**: 03-02.testing.instructions.md describes coverage targets but doesn't analyze results.

- [ ] **A)** YES — implement
- [ ] **B)** NO — developers can read coverage HTML directly
- [ ] **C)** DEFER
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

### Q8: fips-audit skill

**Purpose**: Audit Go code for FIPS 140-3 compliance — find banned algorithms (bcrypt, scrypt, MD5, SHA-1), check crypto library usage, verify approved algorithms used.

**Would create**:
```
.github/skills/fips-audit.md
```

**What it does**: Scans a package/file for FIPS violations: banned imports, non-FIPS algorithm usage, incorrect key sizes, missing CSPRNG usage.

**vs. lint_go**: `cicd lint-go non-fips-algorithms` already runs this. Skill adds: interactive explanation and fix suggestions per violation.

- [ ] **A)** YES — implement (adds fix guidance on top of detection)
- [ ] **B)** NO — `cicd lint-go non-fips-algorithms` is sufficient
- [ ] **C)** DEFER
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

## Section 4: Group D — Documentation Skills

### Q9: propagation-check skill (wraps cicd lint-docs)

**Purpose**: Check @propagate/@source sync status, explain which chunks are out of sync, and generate the corrected @source block content.

**Would create**:
```
.github/skills/propagation-check.md
```

**What it does**: Runs `cicd lint-docs`, identifies out-of-sync chunks, shows diff between ARCHITECTURE.md source and propagated @source blocks, generates corrected text.

**vs. lint-docs**: The cicd command detects; the skill explains AND generates the fix.

- [ ] **A)** YES — implement
- [ ] **B)** NO — lint-docs error messages + manual edit is sufficient
- [ ] **C)** DEFER
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

### Q10: openapi-codegen skill

**Purpose**: Generate the three oapi-codegen config files (server, model, client) for a new service, and generate the OpenAPI spec structure following project conventions.

**Would create**:
```
.github/skills/openapi-codegen.md
.github/skills/openapi-codegen/
  openapi-gen_config_server.yaml
  openapi-gen_config_model.yaml
  openapi-gen_config_client.yaml
```

**What it does**: Given a service name and path prefix, generates all three codegen configs with correct output paths; also generates OpenAPI spec skeleton (paths + components files) following project conventions.

- [ ] **A)** YES — implement
- [ ] **B)** NO — instructions + existing examples are sufficient
- [ ] **C)** DEFER
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

## Section 5: Group E — Scaffolding Skills

### Q11: agent-scaffold skill

**Purpose**: Create a new agent file from template with correct YAML frontmatter, mandatory sections (autonomous execution directive, quality mandate, prohibited behaviors), and ARCHITECTURE.md cross-references.

**Would create**:
```
.github/skills/agent-scaffold.md
.github/skills/agent-scaffold/template.agent.md
```

**What it does**: Given an agent name + purpose, generates a conformant `.github/agents/AGENT-NAME.agent.md` with all mandatory sections pre-filled.

**Rationale**: Agents frequently violate agent-format.instructions.md rules when created manually.

- [ ] **A)** YES — implement (catches common agent creation errors)
- [ ] **B)** NO — 06-02.agent-format.instructions.md is sufficient reference
- [ ] **C)** DEFER
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

### Q12: instruction-scaffold skill

**Purpose**: Create a new instruction file from template with correct YAML frontmatter, `applyTo` pattern, and mandatory ARCHITECTURE.md cross-reference format.

**Would create**:
```
.github/skills/instruction-scaffold.md
.github/skills/instruction-scaffold/template.instructions.md
```

**What it does**: Given an instruction file name + description, generates a conformant `.github/instructions/NN-NN.name.instructions.md`.

- [ ] **A)** YES — implement
- [ ] **B)** NO — existing instructions are clear enough to follow manually
- [ ] **C)** DEFER
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

## Section 6: Skills Naming and Organization

### Q13: Skill file naming convention

How should we name skill files?

- [ ] **A)** `SKILLNAME.md` in `.github/skills/` (flat, e.g., `test-table-driven.md`)
- [ ] **B)** Grouped by category subdirectory (e.g., `.github/skills/test/table-driven.md`)
- [ ] **C)** Match instruction file naming convention (`NN-NN.skillname.md`)
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

### Q14: Should skills be inventoried in ARCHITECTURE.md skills section?

The new skills section in ARCHITECTURE.md (Phase 3, Task 3.5) — should it include a catalogue table of all implemented skills?

- [ ] **A)** YES — include a catalogue table (name, purpose, source instruction)
- [ ] **B)** NO — skills are a VS Code feature layer; ARCHITECTURE.md should only describe the PATTERN, not catalogue each skill
- [ ] **C)** PARTIAL — only list skill categories, not individual skills
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

### Q15: Should any existing agents gain a `skills:` reference?

Some agents could reference relevant skills via `skills:` frontmatter, making skills auto-suggested when the agent is running.

- [ ] **A)** YES — update relevant agents with `skills:` references after skills are created
- [ ] **B)** NO — keep agents and skills independent
- [ ] **C)** YES but only for beast-mode (most generic, would benefit from test/coverage skills)
- [ ] **D)**
- [ ] **E)**

**Answer**:

---

## Summary

After completing your selections:
1. Save this file
2. Agent will merge answers into plan.md/tasks.md and begin implementation

**Skill candidate decision tracking:**

| # | Skill | Your Choice |
|---|-------|-------------|
| Q1 | test-table-driven | |
| Q2 | test-fuzz-gen | |
| Q3 | test-benchmark-gen | |
| Q4 | compose-validator | |
| Q5 | migration-create | |
| Q6 | service-scaffold | |
| Q7 | coverage-analysis | |
| Q8 | fips-audit | |
| Q9 | propagation-check | |
| Q10 | openapi-codegen | |
| Q11 | agent-scaffold | |
| Q12 | instruction-scaffold | |
| Q13 | naming convention | |
| Q14 | ARCHITECTURE.md catalogue | |
| Q15 | agent skills: references | |
