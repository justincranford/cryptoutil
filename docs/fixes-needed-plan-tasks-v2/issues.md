# V2 Session Issues Tracker

## Purpose

Track specific issues fixed during V2 workflow testing and container mode work.

## Session: 2026-01-24 - Workflow Test Fixes

### Issue #1: Container Mode - Explicit Database URL Support

**Status**: Fixed (commit 9e9da31c)

**Problem**:
Container mode requires 0.0.0.0 binding, but the only SQLite path required `dev: true`, which is rejected with 0.0.0.0 (security restriction).

**Solution**:
Added explicit `sqlite://` URL support so containers can specify database type independently of dev mode.

**Key Insight**:
Database choice (SQLite vs PostgreSQL) is orthogonal to bind address (127.0.0.1 vs 0.0.0.0).

---

### Issue #2: mTLS Container Mode

**Status**: Fixed (commit f58c6ff6)

**Problem**:
Container healthchecks failed because admin server required client certificates, but Docker healthcheck (wget) doesn't provide them.

**Solution**:
Detect container mode via 0.0.0.0 binding and disable mTLS for both dev mode AND container mode.

**Test Gap**:
ZERO tests for mTLS configuration logic (security-critical code untested).

---

### Issue #3: DAST Workflow Diagnostics

**Status**: Fixed (commit 80a69d18)

**Problem**:
DAST workflow failures didn't upload artifacts, making diagnosis impossible.

**Solution**:
Added `if: always()` to artifact upload step and inline diagnostic output on health check failures.

---

### Issue #4: DAST Configuration Field Mapping

**Status**: Under investigation

**Problem**:
YAML shows `dev-mode: true` but application logs show `Dev mode (-d): false`.

**Hypothesis**:
Potential kebab-case → PascalCase field mapping issue in config loading.
