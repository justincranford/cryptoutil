# Review 0002: Admin Port Specification Inconsistency (9090 vs 9091/9092/9093)

**Date**: 2025-12-24
**Severity**: CRITICAL
**Category**: Admin Ports
**Status**: FOUND - NOT FIXED

---

## Issue Description

`spec.md` specifies UNIQUE admin ports per product (9090/9091/9092/9093), directly contradicting constitution.md, plan.md, tasks.md, clarify.md, and HTTPS ports memory instructions which mandate **127.0.0.1:9090 for ALL services**.

---

## Evidence

### Files Using CORRECT admin port spec (9090 for ALL)

1. **constitution.md** (Service Catalog table, Column "Admin Port"):

   ```markdown
   | Service | Product | Public Ports | Admin Port | Status | Notes |
   |---------|---------|--------------|------------|--------|-------|
   | sm-kms | Secrets Manager | 8080-8089 | 9090 | ✅ COMPLETE | Reference implementation |
   | pki-ca | PKI | 8443-8449 | 9090 | ⚠️ PARTIAL | Needs dual-server |
   | jose-ja | JOSE | 9443-9449 | 9090 | ⚠️ PARTIAL | Needs dual-server |
   | identity-authz | Identity | 18000-18009 | 9090 | ✅ COMPLETE | Dual servers |
   | identity-idp | Identity | 18100-18109 | 9090 | ✅ COMPLETE | Dual servers |
   | identity-rs | Identity | 18200-18209 | 9090 | ⏳ IN PROGRESS | Public server pending |
   | identity-rp | Identity | 18300-18309 | 9090 | ❌ NOT STARTED | Reference implementation |
   | identity-spa | Identity | 18400-18409 | 9090 | ❌ NOT STARTED | Reference implementation |
   | learn-im | Learn | 8888-8889 | 9090 | ❌ NOT STARTED | Phase 3 validation |
   ```

2. **plan.md** (Line 63):

   ```markdown
   - Private HTTPS Server: 127.0.0.1:9090 (admin endpoints, ALL services use same port)
   - Admin Port Configuration: 127.0.0.1:9090 inside container (NEVER exposed to host), or 127.0.0.1:0 for tests (dynamic allocation)
   ```

3. **.specify/memory/https-ports.md** (Line 23):

   ```markdown
   | Port | `9090` | `9090` |
   ```

4. **clarify.md** (Lines 29-37):

   ```markdown
   **Private HTTPS Server** (Admin endpoints):

   - Purpose: Internal admin tasks, health checks, metrics
   - Admin Port Assignments:
     - KMS: 9090 (all KMS instances share, bound to 127.0.0.1)
     - Identity: 9091 (all 5 Identity services share)
     - CA: 9092 (all CA instances share)
     - JOSE: 9093 (all JOSE instances share)
   ```

**WAIT - clarify.md ALSO has the per-product ports issue!**

### Files Using INCORRECT admin port spec (9090/9091/9092/9093)

1. **spec.md** (Lines 660-662):

   ```markdown
   - **Identity**: Admin port 9091 (all 5 Identity services share)
   - **CA**: Admin port 9092 (all CA instances share)
   - **JOSE**: Admin port 9093 (all JOSE instances share)
   ```

2. **spec.md** (Line 760, ASCII diagram):

   ```
   │ Admin:9093    │ Admin:9091     │ Admin:9090   │Admin:9092
   ```

3. **spec.md** (Lines 879-883):

   ```markdown
   1. **AuthZ Server**: OAuth 2.1 Authorization Server (identity-authz, port 8180, admin 9091)
   2. **IdP Server**: OIDC Identity Provider (identity-idp, port 8181, admin 9091)
   3. **Resource Server**: Protected API with token validation (identity-rs, port 8182, admin 9091)
   4. **Relying Party**: Backend-for-Frontend pattern (identity-rp, port 8183, admin 9091)
   5. **Single Page Application**: Static hosting for SPA clients (identity-spa, port 8184, admin 9091)
   ```

4. **spec.md** (Lines 1188-1190):

   ```markdown
   - **ca-sqlite**: Port 8380 (public API), Port 9092 (admin), SQLite backend
   - **ca-postgres-1**: Port 8381 (public API), Port 9092 (admin), PostgreSQL backend
   - **ca-postgres-2**: Port 8382 (public API), Port 9092 (admin), PostgreSQL backend
   ```

5. **clarify.md** (Lines 29-37) - CONTRADICTS constitution.md:

   ```markdown
   - Admin Port Assignments:
     - KMS: 9090 (all KMS instances share, bound to 127.0.0.1)
     - Identity: 9091 (all 5 Identity services share)
     - CA: 9092 (all CA instances share)
     - JOSE: 9093 (all JOSE instances share)
   ```

---

## Root Cause

User clarified in QUIZME-05 and explicit instruction:

> "CRITICAL FIXES:
>
> - Admin ports: 127.0.0.1:9090 for ALL services (not per-service ports)"

This was backported to constitution.md and plan.md, but **spec.md** and **clarify.md** were NOT updated.

---

## Impact

- **CRITICAL**: Developers implementing from spec.md will use wrong admin ports (9091/9092/9093 instead of 9090)
- **Docker Compose Conflicts**: Per-service admin ports could cause port conflicts when multiple services run
- **Configuration Complexity**: Unnecessary per-service admin port configuration instead of single 9090 port
- **Divergence**: spec.md and clarify.md contradict constitution.md (authoritative source)

---

## Fix Required

1. **spec.md**: Replace ALL admin port references (9091, 9092, 9093) with 9090
2. **clarify.md**: Replace per-product admin port assignments with single 9090 port for ALL services
3. **Validate**: Ensure ASCII diagrams, tables, and prose all reference 127.0.0.1:9090

---

## Correct Specification (from constitution.md)

**All services MUST use**:

- Admin Port: **127.0.0.1:9090** (inside container, NEVER exposed to host)
- Tests: **127.0.0.1:0** (dynamic allocation)
- Admin Endpoints: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/shutdown`

**Rationale**:

- Admin endpoints bound to 127.0.0.1 only (not externally accessible)
- Docker Compose: Each service = separate container with isolated network namespace
- Same admin port (9090) can be reused across ALL instances without collision
- Simplifies configuration (no per-service admin port tracking)

---

## SpecKit Divergence Pattern

**Observation**: User fixed constitution.md and plan.md, but spec.md and clarify.md were missed.

**Root Cause**: clarify.md is supposed to be an authoritative source (Step 3 of SpecKit), but it contradicts constitution.md (Step 1).

**Fundamental Flaw**: SpecKit has THREE authoritative sources (constitution.md, spec.md, clarify.md) with NO automated cross-validation to detect contradictions.

**Recommendation**: Add pre-generation validation that greps for admin port patterns across ALL authoritative sources, flags contradictions BEFORE generating plan.md/tasks.md.
