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
