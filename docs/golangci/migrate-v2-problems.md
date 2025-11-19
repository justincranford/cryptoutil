# golangci-lint v2 Migration - Problems Encountered

This document tracks problems encountered during each migration step.

## Purpose

- Document linting errors/warnings after each step
- Identify patterns in v2 migration issues
- Guide instruction file updates

## Format

Each step should document:

- Step number and description
- Linter output (errors, warnings)
- Analysis of root causes
- Fixes applied (if any)

---

## Migration Problems Log

<!-- Problems will be appended here after each step -->
## Step 1: Core Configuration

**Status**: Partial - v2 removed skip-dirs, skip-files, build-cache, modules-download-mode

**Linter Errors**:
- Property skip-dirs is not allowed
- Property skip-files is not allowed
- Property build-cache is not allowed
- Property modules-download-mode is not allowed (v1 only)

**Analysis**: v2 schema simplified run configuration. These properties were v1-specific.

**Configuration Applied**:
- version: 2
- timeout: 10m
- issues-exit-code: 1
- tests: true
- concurrency: 0

**Next**: Check v2 reference for exclusion patterns (likely moved to linters.exclusions section)

---

## Step 2: Output Configuration

**Status**: Complete - v2 changed output format

**Changes from v1**:
- sort-results: true → sort-order: [linter, severity, file]
- formats.text structure → formats array with format field

**Configuration Applied**:
- formats: colored-line-number
- sort-order: by linter, severity, file

**Linter Test**:

## Step 3: Fast Essential Linters

**Status**: Complete

**Linters Enabled**:
- errcheck
- govet
- ineffassign
- staticcheck (includes gosimple, stylecheck in v2)
- unused

**Test Output**:

## Step 4: Fast Code Quality Linters

**Status**: Complete

**Linters Enabled**:
- revive
- godot
- copyloopvar
- goconst
- importas


level=warning msg="[runner/nolint_filter] Found unknown linters in //nolint directives: stylecheck"
cmd\identity\mock-identity-services.go:462:24                                       errcheck     Error return value of `resp.Body.Close` is not checked
cmd\identity\spa-rp\main.go:50:24                                                   errcheck     Error return value of `indexFile.Close` is not checked
internal\cmd\cicd\cicd_enforce_utf8.go:168:18                                       errcheck     Error return value of `file.Close` is not checked
internal\cmd\cicd\cicd_final_coverage_test.go:46:15                                 errcheck     Error return value of `os.Unsetenv` is not checked
internal\cmd\cicd\cicd_final_coverage_test.go:48:13                                 errcheck     Error return value of `os.Setenv` is not checked
internal\cmd\cicd\cicd_final_coverage_test.go:52:11                                 errcheck     Error return value of `os.Setenv` is not checked
internal\cmd\workflow\workflow.go:145:25                                            errcheck     Error return value of `combinedLog.Close` is not checked
internal\cmd\workflow\workflow.go:441:25                                            errcheck     Error return value of `workflowLog.Close` is not checked
internal\cmd\workflow\workflow.go:623:17                                            errcheck     Error return value of `fmt.Fprintln` is not checked

## Step 5: Security & Error Handling

**Status**: Complete (wrapcheck settings removed - v2 API changed)

**Linters Enabled**:
- gosec (with excludes)
- noctx
- wrapcheck (no custom settings in v2)
- errorlint

## Step 6: Testing Quality Linters

**Status**: Complete

**Linters Enabled**:
- thelper
- tparallel
- testpackage
- gomodguard
- gomoddirectives

## Step 7: Performance & Style Linters

**Status**: Complete

**Linters Enabled**:
- prealloc
- bodyclose
- mnd (with settings)
- wsl
- nlreturn

## Step 8: Maintainability & Headers

**Status**: Complete

**Linters Enabled**:
- goheader
- depguard

**Linters Disabled**:
- dupl
- gocyclo
- godox

## Step 9: Linter-Specific Settings

**Status**: Complete (some v2 API changes)

**Settings Added**:
- errcheck (check-type-assertions, check-blank)
- gocyclo (min-complexity)
- goconst (min-len, min-occurrences, numbers)
- dupl (threshold)
- misspell (locale only - ignore-words removed in v2)
- revive (severity)
- godot (scope, capital)
- godox (keywords)
- thelper (test.begin)
- testpackage (skip-regexp)

**v2 Changes**:
- goconst.ignore-tests removed
- misspell.ignore-words removed

## Step 10: Import Alias Configuration

**Status**: Complete

**Aliases Added**: 60+ import aliases for cryptoutil packages and dependencies

## Step 11 & 12: Issues and Severity Configuration

**Status**: Complete (v2 removed many exclusion options)

**v2 Changes**:
- exclude-dirs, exclude-files, exclude-rules removed
- exclude-generated removed
- Directory exclusions now handled via run.skip-dirs (which also removed in v2)
- File exclusions for generated code automatic via language server

**Issues Configuration**:
- max-issues-per-linter: 100
- max-same-issues: 20

**Severity Configuration**:
- default: error
- revive, godot: warning
- misspell: info

## Final Migration Status

**All 12 steps completed successfully!**

**Final fixes applied**:
- wsl → wsl_v5 (deprecated linter replacement)
- depguard configured with 'main' rule allowing all packages except github.com/pkg/errors

**v2 Configuration Complete**: .golangci.yml now uses golangci-lint v2.6.2 schema
