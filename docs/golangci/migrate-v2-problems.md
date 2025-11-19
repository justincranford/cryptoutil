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
