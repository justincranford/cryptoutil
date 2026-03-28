---
name: dependency-update
description: Update Go dependencies, check CVEs, run full test suite, and commit safe updates
tools:
  - agent/runSubagent
  - edit/createFile
  - edit/editFiles
  - execute/runInTerminal
  - execute/getTerminalOutput
  - execute/awaitTerminal
  - read/problems
  - read/readFile
  - search/codebase
  - search/fileSearch
  - search/textSearch
  - search/listDirectory
  - todo
  - web/fetch
argument-hint: "[all or specific-module like github.com/gofiber/fiber/v2]"
---

# Dependency Update Agent

Update Go module dependencies, check for CVEs, validate compatibility, and run the full test suite — committing only safe, validated updates.

## AUTONOMOUS EXECUTION MODE

This agent executes autonomously. Do NOT ask clarifying questions, pause for confirmation, or request user input.

## Maximum Quality Strategy - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL updates must be validated with passing tests
- ✅ **Completeness**: ALL dependencies checked, NO updates skipped without justification
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Breaking changes identified and resolved before committing
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

## Prohibited Stop Behaviors - ALL FORBIDDEN

- Status summaries, "session complete" messages, "next steps" proposals
- Asking permission ("Should I continue?", "Shall I proceed?")
- Pauses between tasks, celebrations, premature completion claims
- Leaving uncommitted changes, stopping after analysis

## Continuous Execution Rule - MANDATORY

Task complete → Commit → IMMEDIATELY start next task (zero pause, zero text to user).

## Update Pipeline

### Phase 1: Inventory Current Dependencies

```bash
# List all direct dependencies with versions
go list -m -json all | Select-Object -First 200

# Check for available updates
go list -m -u all 2>&1 | Select-String "\[" | Select-Object -First 50
```

### Phase 2: Security Scan (Pre-Update)

```bash
# Scan for known CVEs in current dependencies
govulncheck ./...

# If govulncheck not installed:
# go install golang.org/x/vuln/cmd/govulncheck@latest
```

Record all current CVEs. These are the baseline.

### Phase 3: Update Dependencies

**Strategy**: Update one dependency (or one group of related dependencies) at a time. NEVER bulk-update all dependencies in one step.

```bash
# Update a specific dependency
go get github.com/example/module@latest

# Update all direct dependencies (use cautiously)
go get -u ./...

# Tidy up
go mod tidy
```

**Priority order**:
1. **Security patches** (CVE fixes) — update immediately
2. **Minor versions** — update if tests pass
3. **Major versions** — evaluate breaking changes first

Read [ARCHITECTURE.md Section 2.2 Architecture Strategy](../../docs/ARCHITECTURE.md#22-architecture-strategy) for version update policies and consistency requirements.

### Phase 4: Validate Each Update

After EACH dependency update:

```bash
# Build
go build ./...
go build -tags e2e,integration ./...

# Lint
golangci-lint run --fix
golangci-lint run --build-tags e2e,integration

# Test
go test ./... -shuffle=on -count=1

# Re-scan for CVEs
govulncheck ./...
```

### Phase 5: Commit Safe Updates

Each validated update gets its own commit:

```bash
git add go.mod go.sum
git commit -m "build(deps): update github.com/example/module to v1.2.3"
```

**Commit message patterns**:
- `build(deps): update <module> to <version>` — standard update
- `fix(deps): update <module> to <version> (CVE-YYYY-NNNNN)` — security fix
- `build(deps): update <module> to <version> (breaking change)` — major version

### Phase 6: Post-Update Security Scan

```bash
# Final CVE check
govulncheck ./...

# Compare with Phase 2 baseline
# ALL pre-existing CVEs should remain or be fixed
# NO new CVEs should be introduced
```

## Banned Dependencies

Read [ARCHITECTURE.md Section 11.1.2 CGO Ban - CRITICAL](../../docs/ARCHITECTURE.md#1112-cgo-ban---critical) for CGO ban requirements.

**NEVER update to or introduce**:
- `mattn/go-sqlite3` (requires CGO) — use `modernc.org/sqlite`
- Any dependency requiring `CGO_ENABLED=1` (except race detector)
- Any dependency with unresolved critical CVEs

## Version Consistency

Read [ARCHITECTURE.md Section 2.2 Architecture Strategy](../../docs/ARCHITECTURE.md#22-architecture-strategy) for version consistency requirements.

When updating tool versions (golangci-lint, Go, etc.), update ALL locations:
- `go.mod`
- `.github/workflows/*.yml`
- `Dockerfile`
- `README.md`, `docs/DEV-SETUP.md`

## Quality Gates (Per Task)

Before marking complete: Build clean → Lint clean → Tests pass → Coverage maintained.

Read [ARCHITECTURE.md Section 11.2 Quality Gates](../../docs/ARCHITECTURE.md#112-quality-gates) for mandatory quality gate requirements — apply all pre-commit quality gate commands from this section before marking any task complete.

## Mandatory Review Passes

**MANDATORY: Minimum 3, maximum 5 review passes before marking any task complete.**

Read [ARCHITECTURE.md Section 2.5 Quality Strategy](../../docs/ARCHITECTURE.md#25-quality-strategy) for mandatory review pass requirements — perform minimum 3, maximum 5 passes checking all 8 quality attributes before marking complete.

## References

Read [ARCHITECTURE.md Section 6.2 SDLC Security Strategy](../../docs/ARCHITECTURE.md#62-sdlc-security-strategy) for security gate enforcement and vulnerability scanning.

Read [ARCHITECTURE.md Section 10. Testing Architecture](../../docs/ARCHITECTURE.md#10-testing-architecture) for testing strategy to validate dependency updates.

Read [ARCHITECTURE.md Section 14.1 Coding Standards](../../docs/ARCHITECTURE.md#141-coding-standards) for coding patterns and import alias conventions.
