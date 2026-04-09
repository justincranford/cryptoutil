# Plan — Framework v9: Carryover & Quality

**Status**: Not Started
**Created**: 2026-04-08

---

## Overview

Framework v9 carries forward deferred items from framework-v8 (Deployment Parameterization)
and addresses quality improvements identified during framework-v8 execution.

Framework-v8 completed 43/43 tasks (100%) across 10 phases. The recursive Docker Compose
include architecture is fully operational at all 3 deployment tiers (SERVICE, PRODUCT, SUITE).

---

## Carryover Items from Framework-v8

### 1. Load Test Refactoring: All Tiers [LOW]

**Source**: framework-v8 carryover Item 7

**Current state**: `test/load/` (Gatling, Java 21, Maven) covers only some service-level
scenarios. The target is:
- 10 service-level load test scenarios (one per PS-ID)
- 5 product-level load test scenarios (one per product)
- 1 suite-level load test scenario

**Why LOW**: Load tests do not block CI/CD and require Java/Gatling expertise to extend.
However, the gap means production throughput characteristics at product and suite levels are
unknown until these are created.

**Action**: Extend `test/load/src/` to add product-level and suite-level simulation classes.
Ensure `pom.xml` is updated with the new simulation entry points.

### 2. Docker Compose v2.24+ Version Check [LOW]

**Source**: framework-v8 tasks.md deferred work note

**Current state**: The recursive include architecture (Approach C with `!override` YAML tag)
requires Docker Compose v2.24+. This minimum version is documented in `docs/DEV-SETUP.md`
but is not enforced by any validator.

**Action**: Consider adding a Docker Compose version check to `lint-deployments` or
documenting the minimum version requirement more prominently.
