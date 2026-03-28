# Framework v7 - Ongoing Reference Documentation

**Status**: Living documentation — ongoing, continuous, iterative
**Created**: 2026-03-28
**Purpose**: Consolidation of all ongoing and living reference documentation from prior framework iterations (v3-v6). All completed work from prior iterations has been deleted — git history preserves everything.

## Contents

### target-structure.md

Canonical target repository structure — defines the complete, parameterized target state of every directory and file in the repository. Post-v6 implementation state. Living spec: updated when architectural decisions change.

### gremlins/

Mutation testing reference documentation — ongoing quality improvement.

| File | Purpose |
|------|---------|
| MUTATIONS-HOWTO.md | Quick start guide for gremlins mutation testing |
| MUTATIONS-TASKS.md | Comprehensive task list with commands for all packages |
| mutation-analysis.md | Detailed analysis of lived mutations and recommended fixes |
| mutation-baseline-results.md | Baseline results and per-package improvement tracking |

### workflow-runtimes/

Operational reference for CI/CD workflow performance and GitHub storage management.

| File | Purpose |
|------|---------|
| README.md | Workflow runtime statistics, success/failure rates, optimization recommendations |
| GITHUB-STORAGE-CLEANUP.md | GitHub Actions storage cleanup best practices and automation |

## Prior Iterations (Deleted — Git History Preserves)

| Directory | Status | Summary |
|-----------|--------|---------|
| framework-brainstorm/ | COMPLETED | Initial research (00-overview through 08-recommendations), consumed by v1-v6 |
| framework-v3/ | COMPLETED | Aggressive standardization, 10 services as thin wrappers, all phases done |
| framework-v4/ | COMPLETED | Anti-drift fitness linter expansion, 43 fitness checks operational |
| framework-v5/ | COMPLETED | Archive cleanup, configs standardization, ARCHITECTURE.md consolidation (49/49 tasks) |
| framework-v6/ | COMPLETED | Corrective standardization, secret/config/deployment fixes (63/63 tasks) |
| LESSONS/ | COMPLETED | 8 lessons propagated to instruction files and ARCHITECTURE.md |
