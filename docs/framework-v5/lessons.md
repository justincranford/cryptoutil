# Lessons - Framework v5: Rigid Standardization & Cleanup

**Created**: 2026-03-21
**Last Updated**: 2026-03-24

## Phase 1: Archive, Dead Code, and Legacy Cleanup

- **Root-level junk cleanup**: ~80+ files (*.exe, *.py, coverage_*,*.log) at repo root required cleanup. Prevention: add root-level file type restrictions to pre-commit hooks.
- **Archive vs delete decision**: Permanent deletion chosen over archival. Git history preserves everything. This avoids "dead code resurrection" from hot-exit.
- **VS Code hot-exit file resurrection**: Deleted files reappear after VS Code reload due to buffer recovery. Must close tabs before deleting files.
- **Batch deletion efficiency**: `git rm` with glob patterns faster than individual file deletions.

## Phase 2: Non-Standard Entry Rationalization

- **cmd/cicd to cmd/cicd-lint rename**: Cascading reference updates across 30+ files (workflows, instructions, agents, Makefiles). Always grep entire repo for old references.
- **PowerShell terminal truncation**: Long heredoc scripts appear to produce no output but execute correctly. Verify with `Test-Path` and content checks rather than trusting terminal display.
- **Pre-commit hook behavior**: When pre-commit hooks modify files (e.g., gofmt, trailing whitespace), first commit attempt fails with "files were modified by hook." Re-stage and re-commit.

## Phase 3: Configs Standardization

- **Config file naming**: {PS-ID}-{purpose}.yml pattern enforced (e.g., sm-kms-main.yml). Fitness linter validates.
- **Reference update ordering**: When renaming directories, update all internal references BEFORE renaming to avoid broken imports during intermediate states.

## Phase 4: Deployments Refinement

- **Compose multi-deploy merge**: ARCHITECTURE-COMPOSE-MULTIDEPLOY.md content merged into ARCHITECTURE.md Section 12.3. Standalone doc deleted. Prevents doc sprawl.
- **Deployment validator count**: 68+ validators across lint-deployments. New linters must integrate without breaking existing validators.

## Phase 5: ARCHITECTURE.md Roadmap Consolidation

- **Propagation integrity**: @propagate/@source block system requires byte-for-byte matching. `validate-propagation` CI command catches drift.
- **Workflow consolidation**: ci-cicd-lint.yml merged into ci-quality.yml. Reduces workflow maintenance overhead.

## Phase 6: Fitness Linter Expansion

- **Magic constant cascading**: Adding `CICDConfigsDir = "configs"` to magic_cicd.go instantly created 73 blocking `literal-use` violations across 22+ files. Must fix ALL in same commit to pass lint-go.
- **Seam test pattern**: Internal package tests (same package, not `_test` suffix) access unexported seam variables for error injection. Mark `// Sequential:` for non-parallel tests that modify global state.
- **Gremlins mutation timeout interpretation**: Timed-out mutants count as "caught" — the mutation prevents tests from passing within the timeout, meaning tests detect the mutation. Zero lived mutants = effective 100% mutation coverage.
- **Import block editing hazard**: When using replace_string_in_file on import blocks ending with `)`, the closing paren can be consumed by the replacement. Always verify `)` is preserved in both old and new strings.
- **Permission constants**: 0o755 and 0o644 literals must use `CICDTempDirPermissions` and `CICDOutputFilePermissions` from magic_cicd.go.
- **Entity registry as source of truth**: Product-service registry (10 PS entries, 5 products, 1 suite) used by multiple linters (configs_naming, configs_deployments_consistency, entity_registry_completeness). Changes to registry cascade to all dependent linters.

## Phase 7: Knowledge Propagation

*(To be filled during Phase 7 execution)*
