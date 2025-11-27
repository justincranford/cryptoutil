# Tools and Utilities

This document describes the tools and utilities available in the cryptoutil project.

## Code Generation

### go generate

Run `go generate ./...` to regenerate code from OpenAPI specs and other generators.

## Automation Tools

### go-generate-postmortem

**Purpose**: Automated post-mortem generation from commit history and task documents

**Usage**:

```bash
go run ./cmd/cicd go-generate-postmortem --start-task P5.01 --end-task P5.05 --output path/to/POSTMORTEM.md
```

**Inputs**:

- Commit history (git log)
- Task documents (docs/02-identityV2/passthru5/P5.XX-*.md)
- Requirements coverage (REQUIREMENTS-COVERAGE.md)
- Project status (PROJECT-STATUS.md)

**Outputs**:

- Post-mortem markdown (8 sections: Executive Summary, Task-by-Task, Patterns, Improvements, Gaps, Evidence Quality)

**Benefit**: 50% reduction in post-mortem creation time (15 min → 7.5 min)

### go-update-project-status-v2

**Purpose**: Automated PROJECT-STATUS.md updates from requirements coverage metrics

**Usage**:

```bash
go run ./cmd/cicd go-update-project-status-v2
```

**Inputs**:

- Requirements coverage (REQUIREMENTS-COVERAGE.md)
- Project status (PROJECT-STATUS.md)

**Outputs**:

- Updated PROJECT-STATUS.md (requirements table, task-specific coverage, production readiness)

**Benefit**: Eliminates manual PROJECT-STATUS.md updates (5 min → 0 min)

## Testing Tools

See README.md for test execution instructions.

## Development Tools

See DEV-SETUP.md for development environment setup and tooling.
