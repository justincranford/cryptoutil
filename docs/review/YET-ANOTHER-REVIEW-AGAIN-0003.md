# Review 0003: Multi-Tenancy Architecture Contradiction (Schema-Only vs Dual-Layer)

**Date**: 2025-12-24
**Severity**: CRITICAL
**Category**: Multi-Tenancy Architecture
**Status**: FOUND - NOT FIXED

---

## Issue Description

`spec.md` specifies **schema-level isolation ONLY**, explicitly stating "NOT SUPPORTED: Row-level security (RLS) with tenant ID columns", directly contradicting constitution.md, clarify.md, and plan.md which mandate **dual-layer isolation** (per-row tenant_id + schema-level for PostgreSQL).

---

## Evidence

### Files Using CORRECT multi-tenancy spec (Dual-Layer)

1. **constitution.md** (Lines 170-178, Section "Database Architecture"):

   ```markdown
   - Multi-tenancy (Dual-Layer Isolation):
     - Layer 1 (PostgreSQL + SQLite): Per-row tenant_id column (UUIDv4, FK to tenants.id) in all tables
     - Layer 2 (PostgreSQL only): Schema-level isolation (CREATE SCHEMA tenant_<UUID>)
     - NEVER use row-level security (RLS) - per-row tenant_id + schema isolation provides sufficient protection
   ```

2. **plan.md** (Lines 69-73):

   ```markdown
   - Multi-tenancy (Dual-Layer Isolation):
     - Layer 1 (PostgreSQL + SQLite): Per-row tenant_id column (UUIDv4, FK to tenants.id) in all tables
     - Layer 2 (PostgreSQL only): Schema-level isolation (CREATE SCHEMA tenant_<UUID>)
     - NEVER use row-level security (RLS) - per-row tenant_id + schema isolation provides sufficient protection
   ```

3. **clarify.md** (Multi-Tenancy section, comprehensive examples with tenant_id column + schema switching)

4. **User's explicit instruction (commit message 9105bf68)**:

   ```
   CRITICAL FIXES:
   - Multi-tenancy: Dual-layer (per-row tenant_id + schema-level for PostgreSQL)
   ```

### Files Using INCORRECT multi-tenancy spec (Schema-Only)

1. **spec.md** (Lines 2387-2388):

   ```markdown
   - **REQUIRED**: Schema-level tenant isolation (e.g., `tenant_a.users`, `tenant_b.users`)
   - **NOT SUPPORTED**: Row-level security (RLS) with tenant ID columns
   ```

2. **spec.md** (Line 2391):

   ```markdown
   **Schema Isolation Architecture**:
   ```

3. **spec.md** (Line 2408):

   ```markdown
   tenant_id_header: X-Tenant-ID
   ```

   (Mentions tenant_id but doesn't specify per-row column requirement)

4. **spec.md** (Line 2419):

   ```markdown
   **Rationale**: Schema isolation provides database-level isolation without separate database connections, balancing security and resource efficiency.
   ```

   (Only mentions schema isolation, omits per-row tenant_id layer)

---

## Root Cause

User clarified multi-tenancy in QUIZME-05, constitution.md, clarify.md, and plan.md as **dual-layer**:

1. **Layer 1** (PostgreSQL + SQLite): Per-row `tenant_id` column (UUIDv4, FK to `tenants.id`)
2. **Layer 2** (PostgreSQL only): Schema-level isolation (`CREATE SCHEMA tenant_<UUID>`)

However, `spec.md` was NOT updated and still specifies **schema-level ONLY** with explicit prohibition of row-level tenant ID columns.

---

## Impact

- **CRITICAL**: Developers implementing from spec.md will use schema-level isolation ONLY, missing per-row tenant_id requirement
- **Data Isolation**: Without per-row tenant_id, queries won't filter by tenant at SQL level
- **SQLite Support**: Schema-level isolation doesn't work with SQLite (no schema support), requires per-row tenant_id
- **Security**: Single-layer isolation is insufficient per constitution.md
- **Divergence**: spec.md contradicts constitution.md, clarify.md, and plan.md

---

## Fix Required

1. **spec.md**: Replace "Schema-level isolation ONLY" with "Dual-layer isolation" specification
2. **spec.md**: Remove "NOT SUPPORTED: Row-level security (RLS) with tenant ID columns"
3. **spec.md**: Add Layer 1 requirement (per-row tenant_id for ALL tables in PostgreSQL + SQLite)
4. **spec.md**: Clarify Layer 2 (schema-level) is PostgreSQL-only enhancement
5. **spec.md**: Update rationale to explain dual-layer defense-in-depth

---

## Correct Specification (from constitution.md)

**Multi-tenancy MUST implement dual-layer isolation**:

**Layer 1: Per-Row Tenant ID** (PostgreSQL + SQLite):

- ALL tables MUST have `tenant_id UUID NOT NULL` column
- `tenant_id` is foreign key to `tenants.id` (UUIDv4)
- ALL queries MUST filter by `WHERE tenant_id = $1`
- Enforced at application layer (SQL query construction)

**Layer 2: Schema-Level Isolation** (PostgreSQL ONLY):

- Each tenant gets separate schema: `CREATE SCHEMA tenant_<UUID>`
- Connection sets search_path: `SET search_path TO tenant_<UUID>`
- Provides database-level isolation for PostgreSQL deployments
- NOT applicable to SQLite (no schema support)

**Rationale**:

- Layer 1 (per-row tenant_id): Works on PostgreSQL + SQLite, mandatory defense
- Layer 2 (schema-level): PostgreSQL-only enhancement, additional isolation
- Dual-layer provides defense-in-depth (application-level + database-level)
- NEVER use row-level security (RLS) - layers 1+2 sufficient

---

## SpecKit Divergence Pattern

**Observation**: User clarified multi-tenancy architecture in constitution.md (Step 1) and clarify.md (Step 3), but spec.md (Step 2) was NOT updated.

**Root Cause**: spec.md is an authoritative source generated BEFORE clarify.md, but clarifications from clarify.md are NOT automatically backported to spec.md.

**Fundamental Flaw**: SpecKit treats spec.md as "frozen" after Step 2, but clarifications in Step 3 (clarify.md) often contradict or refine spec.md. There's no bidirectional feedback loop.

**Recommendation**: Add Step 2.5 "Update spec.md from clarify.md" to incorporate clarifications back into spec.md before generating plan.md/tasks.md.
