# DETAILED Implementation Tracking

**Project**: cryptoutil
**Spec**: 002-cryptoutil
**Status**: Phase 2 (Service Template Extraction) - READY TO START
**Last Updated**: 2025-12-25

---

## Section 1: Task Checklist

Tracks implementation progress from [tasks.md](../tasks.md). Updated continuously during implementation.

### Phase 2: Service Template Extraction ⏸️ PENDING

#### P2.1: Template Extraction

- ⚠️ **P2.1.1**: Extract service template from KMS
  - **Status**: IN PROGRESS
  - **Effort**: L (14-21 days)
  - **Dependencies**: None (Phase 1 complete)
  - **Coverage**: Target ≥98%
  - **Mutation**: Target ≥98%
  - **Blockers**: None
  - **Notes**: CRITICAL - Blocking all service migrations (Phases 3-6)
  - **Commits**: (pending)

### Phase 3: Learn-IM Demonstration Service ⏸️ PENDING

#### P3.1: Learn-IM Implementation

- ❌ **P3.1.1**: Implement learn-im encrypted messaging service
  - **Status**: BLOCKED BY P2.1.1
  - **Effort**: L (21-28 days)
  - **Dependencies**: P2.1.1 (template extracted)
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P2.1.1 (template extraction)
  - **Notes**: CRITICAL - First real-world template validation, blocks all production migrations
  - **Commits**: (pending)

### Phase 4: Migrate jose-ja to Template ⏸️ PENDING

#### P4.1: JA Service Migration

- ❌ **P4.1.1**: Migrate jose-ja admin server to template
  - **Status**: BLOCKED BY P3.1.1
  - **Effort**: M (5-7 days)
  - **Dependencies**: P3.1.1 (learn-im validates template)
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P3.1.1 (learn-im validates template)
  - **Notes**: First production service migration, will drive JOSE pattern refinements
  - **Commits**: (pending)

### Phase 5: Migrate pki-ca to Template ⏸️ PENDING

#### P5.1: CA Service Migration

- ❌ **P5.1.1**: Migrate pki-ca admin server to template
  - **Status**: BLOCKED BY P4.1.1
  - **Effort**: M (5-7 days)
  - **Dependencies**: P4.1.1 (JOSE migrated)
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P4.1.1 (JOSE migrated)
  - **Notes**: Second production migration, will drive CA/PKI pattern refinements
  - **Commits**: (pending)

### Phase 6: Identity Services Enhancement ⏸️ PENDING

#### P6.1: Admin Server Implementation

- ❌ **P6.1.1**: RP admin server with template
  - **Status**: BLOCKED BY P5.1.1
  - **Effort**: M (3-5 days)
  - **Dependencies**: P5.1.1 (template mature after CA migration)
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P5.1.1 (template mature after CA migration)
  - **Commits**: (pending)

- ❌ **P6.1.2**: SPA admin server with template
  - **Status**: BLOCKED BY P6.1.1
  - **Effort**: M (3-5 days)
  - **Dependencies**: P6.1.1
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P6.1.1
  - **Commits**: (pending)

- ❌ **P6.1.3**: Migrate authz, idp, rs to template
  - **Status**: BLOCKED BY P6.1.2
  - **Effort**: M (4-6 days)
  - **Dependencies**: P6.1.2
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P6.1.2
  - **Commits**: (pending)

#### P6.2: E2E Path Coverage

- ❌ **P6.2.1**: Browser path E2E tests
  - **Status**: BLOCKED BY P6.1.3
  - **Effort**: M (5-7 days)
  - **Dependencies**: P6.1.3
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P6.1.3
  - **Notes**: BOTH `/service/**` and `/browser/**` paths required
  - **Commits**: (pending)

### Phase 7: Advanced Identity Features ⏸️ FUTURE

#### P7.1: Multi-Factor Authentication

- ❌ **P7.1.1**: TOTP implementation
  - **Status**: BLOCKED BY P6.2.1
  - **Effort**: M (7-10 days)
  - **Dependencies**: P6.2.1
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P6.2.1
  - **Commits**: (pending)

#### P7.2: WebAuthn

- ❌ **P7.2.1**: WebAuthn support
  - **Status**: BLOCKED BY P7.1.1
  - **Effort**: L (14-21 days)
  - **Dependencies**: P7.1.1
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P7.1.1
  - **Commits**: (pending)

### Phase 8: Scale & Multi-Tenancy ⏸️ FUTURE

#### P8.1: Database Sharding

- ❌ **P8.1.1**: Tenant ID partitioning
  - **Status**: BLOCKED BY P7.2.1
  - **Effort**: L (14-21 days)
  - **Dependencies**: P7.2.1
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P7.2.1
  - **Notes**: Multi-tenancy dual-layer (per-row tenant_id + schema-level for PostgreSQL)
  - **Commits**: (pending)

### Phase 9: Production Readiness ⏸️ FUTURE

#### P9.1: Security Hardening

- ❌ **P9.1.1**: SAST/DAST security audit
  - **Status**: BLOCKED BY P8.1.1
  - **Effort**: M (7-10 days)
  - **Dependencies**: P8.1.1
  - **Coverage**: N/A (security audit)
  - **Blockers**: P8.1.1
  - **Commits**: (pending)

#### P9.2: Production Monitoring

- ❌ **P9.2.1**: Observability enhancement
  - **Status**: BLOCKED BY P9.1.1
  - **Effort**: M (5-7 days)
  - **Dependencies**: P9.1.1
  - **Coverage**: N/A (monitoring)
  - **Blockers**: P9.1.1
  - **Commits**: (pending)

---

## Section 2: Append-Only Timeline

Chronological implementation log with mini-retrospectives. NEVER delete entries - append only.

### 2025-12-25: Started P2.1.1 - Extract Service Template from KMS

**Work Completed**:

- Analyzed tasks.md P2.1.1 requirements (12 files to create, 98% coverage/mutation targets)
- Updated DETAILED.md Section 1: P2.1.1 status changed from ❌ NOT STARTED to ⚠️ IN PROGRESS
- Code archaeology on KMS and newer services (JOSE, Identity AuthZ/IdP/RS/RP)
- Key findings:
  - KMS uses complex `application_listener.go` with dual Fiber apps (public + private) - **LEGACY PATTERN**
  - JOSE/Identity services use cleaner pattern: `Application`, `PublicServer`, `AdminServer` - **PREFERRED PATTERN**
  - JOSE `application.go`: Unified app managing both servers with `Start()`, `Shutdown()`, port getters
  - JOSE `admin.go`: 325 lines - livez/readyz/shutdown endpoints, self-signed TLS, mutex-protected state
  - JOSE `server.go`: 336 lines - public server with business logic, TLS config, dynamic port allocation
  - Identity services follow same pattern with config-driven binding (not hardcoded ports)
- Next: Extract reusable patterns into `internal/template/server/` package

**Coverage/Quality Metrics**:

- Before: N/A (template package doesn't exist yet)
- Target: ≥98% coverage, ≥98% mutation score

**Lessons Learned**:

- Two distinct patterns exist in codebase:
  - **KMS**: Complex single-file `application_listener.go` (legacy, ~458 lines, harder to maintain)
  - **JOSE/Identity**: Clean separation (`Application` + `PublicServer` + `AdminServer`, easier to test and reuse)
- Template should extract JOSE/Identity pattern (cleaner, newer, more maintainable)
- Key abstraction points identified:
  1. **Dual HTTPS servers**: Public (business) + Admin (health checks) with independent lifecycles
  2. **TLS generation**: Self-signed cert generation pattern (ECDSA P-384, 1-year validity)
  3. **Admin endpoints**: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/shutdown` (standardized)
  4. **Dynamic port allocation**: `port 0` pattern for tests, configured ports for production
  5. **Graceful shutdown**: Context-based shutdown with timeout, mutex-protected state transitions
  6. **Health check semantics**: Liveness (process alive) vs Readiness (dependencies healthy)

**Constraints Discovered**:

- Must support BOTH blocking and non-blocking server startup modes
- Admin server ALWAYS binds to `127.0.0.1:9090` (hardcoded, NOT configurable per security requirements)
- Public server binding configurable (`127.0.0.1` for tests, `0.0.0.0` for containers)

**Requirements Discovered**:

- Template must support constructor injection (config, handlers, middleware)
- Template must allow service-specific customization (OpenAPI specs, business logic routes)
- Template must support dual request paths (`/service/**` vs `/browser/**` middleware stacks)

**Related Commits**: ca555b29 (started P2.1.1 tracking)

---
