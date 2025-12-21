# Health Check Endpoint Analysis

## Summary

Investigation into `/admin/v1/healthz` vs `/admin/v1/livez` + `/admin/v1/readyz` pattern reveals inconsistencies between documentation and implementation.

## Key Findings

### 1. KMS (Reference Implementation) - Original Pattern

**Implementation Reality** (source: internal/kms/server/application/application_listener.go, magic_network.go):

- Uses **gofiber middleware** for health check pattern
- Implements **ONLY livez + readyz**, NO healthz
- Endpoints: `/admin/v1/livez` and `/admin/v1/readyz`
- Magic constants: `PrivateAdminLivezRequestPath = "/livez"`, `PrivateAdminReadyzRequestPath = "/readyz"`
- Context path: `DefaultPrivateAdminAPIContextPath = "/admin/v1"`
- Full paths: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/shutdown`

**Docker Compose healthcheck** (deployments/compose/compose.yml lines 237, 321, 409):

```yaml
healthcheck:
  test: ["CMD", "curl", "-k", "-f", "-s", "https://127.0.0.1:9090/admin/v1/livez"]
```

### 2. CA Service - Inconsistent Pattern

**Implementation** (source: internal/ca/server/server.go, admin.go):

Public server (server.go lines 197-199):

- `/health`, `/livez`, `/readyz` - NO /admin/v1 prefix

Admin server (admin.go lines 77-78):

- `/livez`, `/readyz` - NO /admin/v1 prefix

**Docker Compose healthcheck** (deployments/ca/compose.yml line 133):

```yaml
# Comment: "CA server routes: /health, /livez, /readyz (no /admin/v1/ prefix)"
test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:8443/livez"]
```

**Issue**: CA doesn't use `/admin/v1` context path at all - uses root-level endpoints

### 3. JOSE Service

**Implementation**: Uses gofiber similar to KMS (likely livez/readyz pattern)

**Docker Compose healthcheck** (deployments/jose/compose.yml line 64-67):

```yaml
healthcheck:
  # Missing test definition - compose file incomplete
```

### 4. Identity Services (authz, idp, rs)

**Implementation**: Dual-server pattern like KMS (likely livez/readyz pattern)

**Docker Compose healthcheck** (deployments/identity/compose.advanced.yml):

```yaml
healthcheck:
  test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/v1/livez"]
```

### 5. Third-Party Services

**OpenTelemetry Collector** (deployments/telemetry/compose.yml):

- Port 13133 for health checks (OTEL standard)
- NOT exposed to host (container-only)
- Pattern: OpenTelemetry standard health endpoint

**PostgreSQL** (deployments/compose/compose.yml):

- Uses native `pg_isready` health check
- Pattern: Database-specific health check

**Grafana OTEL LGTM**:

- No explicit health check in compose file
- Likely uses Grafana standard health endpoints

## Documentation Inconsistencies

### constitution.md (.specify/memory/constitution.md lines 404-408)

**INCORRECT Documentation**:

```markdown
**Endpoints**:

- `/admin/v1/livez` - Liveness probe (service running)
- `/admin/v1/readyz` - Readiness probe (service ready to accept traffic)
- `/admin/v1/healthz` - Health check (service healthy)  <-- WRONG: Not implemented
- `/admin/v1/shutdown` - Graceful shutdown trigger
```

### spec.md (specs/002-cryptoutil/spec.md)

**INCORRECT Documentation** (multiple locations):

- Line 416: "Poll `/admin/v1/healthz` endpoint before resuming traffic"
- Line 496: "`/admin/v1/healthz` - Combined health check"
- Lines 1200-1209: Table showing `/admin/v1/healthz` for all services
- Line 1325: Lists `/admin/v1/healthz` alongside livez/readyz
- Line 1647: "Add `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/healthz` endpoints"

### copilot-instructions.md (.github/copilot-instructions.md)

**MISSING**: No mention of health check endpoint pattern at all

### Instruction Files (.github/instructions/)

**01-01.architecture.instructions.md**:

- Line 41: "Health Check Endpoints: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/healthz`"
- Line 149-151: Lists all three endpoints
- Line 287: References `/admin/v1/healthz`

**01-07.security.instructions.md**:

- Line 47: "ALWAYS use HTTPS 127.0.0.1:9090 for admin APIs (/shutdown, /livez, /readyz)" - CORRECT (no healthz)

**02-02.docker.instructions.md**:

- Lines 90, 159, 198: Use `/admin/v1/livez` - CORRECT

**06-01.speckit.instructions.md**:

- Lines 423-425: Use `/admin/v1/healthz` - INCORRECT

## Analysis: Where Did healthz Come From?

### Hypothesis 1: Kubernetes Influence

- Kubernetes uses `/healthz` as de facto standard
- May have been added to spec without verifying KMS implementation
- KMS uses gofiber which provides `/livez` and `/readyz` (Kubernetes standard)

### Hypothesis 2: Third-Party Service Pattern Copy

- OpenTelemetry Collector uses port 13133 for health (different pattern)
- Grafana uses standard Grafana health endpoints
- PostgreSQL uses `pg_isready` (not HTTP endpoints)
- **NOT FOUND**: No third-party service using `/admin/v1/healthz`

### Hypothesis 3: Documentation Drift

- KMS was implemented first with livez/readyz (gofiber standard)
- Documentation added healthz without implementation
- Pattern propagated to spec.md, constitution.md, instructions

## Correct Pattern for Proprietary Services

### All 9 Proprietary Services SHOULD Use

**Admin HTTPS Endpoint** (127.0.0.1:9090):

- `/admin/v1/livez` - Liveness probe (lightweight, service running)
- `/admin/v1/readyz` - Readiness probe (heavyweight, service ready for traffic)
- `/admin/v1/shutdown` - Graceful shutdown trigger
- **NO** `/admin/v1/healthz` - This is not implemented and not needed

**Rationale**:

1. Kubernetes standard: livez (liveness) + readyz (readiness) - TWO separate concerns
2. Gofiber middleware provides livez/readyz out-of-box
3. KMS reference implementation uses this pattern
4. Docker health checks use livez (lightweight, fast response)
5. Kubernetes probes use livez (liveness) and readyz (readiness) separately

### CA Service Exception

**Current CA Implementation**:

- Public server: `/health`, `/livez`, `/readyz` (no /admin/v1 prefix)
- Admin server: `/livez`, `/readyz` (no /admin/v1 prefix)
- Uses DIFFERENT port (8443) for health checks in Docker Compose

**Should CA Align?** - YES

- Migrate CA to standard `/admin/v1/livez` and `/admin/v1/readyz` pattern
- Move health checks to admin server on 127.0.0.1:9090
- Remove public server health endpoints (security: don't expose internal status)

## Recommendations

### 1. Update Documentation (MANDATORY)

**constitution.md**:

- Remove `/admin/v1/healthz` from endpoint list
- Clarify livez (lightweight liveness) vs readyz (heavyweight readiness)
- Add CA alignment requirement

**spec.md**:

- Remove all `/admin/v1/healthz` references
- Update health check tables to show only livez/readyz
- Update retry logic to use `/admin/v1/livez` (fast check)

**copilot-instructions.md**:

- Add explicit health check endpoint pattern section
- Document livez (liveness) vs readyz (readiness) semantics
- Clarify third-party vs proprietary patterns

**Instruction Files**:

- 01-01.architecture.instructions.md: Remove healthz references
- 06-01.speckit.instructions.md: Change healthz to livez
- Add explicit third-party service health check documentation

### 2. Align CA Implementation (RECOMMENDED)

**Phase 1: Add /admin/v1 context path**:

- Update CA admin server to use `/admin/v1/livez`, `/admin/v1/readyz`
- Keep existing endpoints for backward compatibility (deprecate in Phase 2)

**Phase 2: Remove public server health endpoints**:

- Remove `/health`, `/livez`, `/readyz` from public server (security)
- Update Docker Compose to use admin server health check
- Update documentation to match

### 3. Third-Party Service Documentation (NEW)

**Add explicit documentation for third-party dependencies**:

| Service | Health Check Pattern | Port | Endpoint | Notes |
|---------|---------------------|------|----------|-------|
| OpenTelemetry Collector | OTEL standard | 13133 | Container-only | NOT exposed to host |
| PostgreSQL | pg_isready | N/A | Native health check | Database-specific |
| Grafana OTEL LGTM | Grafana standard | 3000 | TBD | Grafana-specific |

## User Review Prompts

### Proprietary Services (9 Services)

Please verify the following health check configuration is correct for each proprietary service:

| Service | Expected Health Checks | Admin Port | Notes |
|---------|----------------------|------------|-------|
| sm-kms | /admin/v1/livez, /admin/v1/readyz | 127.0.0.1:9090 | ✅ Reference implementation |
| pki-ca | /livez, /readyz (no /admin/v1) | 127.0.0.1:9090 | ⚠️ Needs alignment |
| jose-ja | /admin/v1/livez, /admin/v1/readyz | 127.0.0.1:9090 | ⏳ Verify implementation |
| identity-authz | /admin/v1/livez, /admin/v1/readyz | 127.0.0.1:9090 | ⏳ Verify implementation |
| identity-idp | /admin/v1/livez, /admin/v1/readyz | 127.0.0.1:9090 | ⏳ Verify implementation |
| identity-rs | /admin/v1/livez, /admin/v1/readyz | 127.0.0.1:9090 | ⏳ Verify implementation |
| identity-rp | /admin/v1/livez, /admin/v1/readyz | 127.0.0.1:9090 | ❌ Not started |
| identity-spa | /admin/v1/livez, /admin/v1/readyz | 127.0.0.1:9090 | ❌ Not started |
| learn-ps | /admin/v1/livez, /admin/v1/readyz | 127.0.0.1:9090 | ❌ Not started (Phase 7) |

**Questions**:

1. Should CA align to `/admin/v1/livez` + `/admin/v1/readyz` pattern? (Recommendation: YES)
2. Should CA remove public server health endpoints for security? (Recommendation: YES)
3. Are JOSE, Identity services already using livez/readyz? (Need verification)

### Third-Party Services

Please verify the following third-party health check configuration:

| Service | Health Check Type | Port | Endpoint | Exposed to Host? |
|---------|------------------|------|----------|------------------|
| OpenTelemetry Collector | OTEL standard | 13133 | Internal | ❌ No (container-only) |
| PostgreSQL | pg_isready | N/A | N/A | N/A |
| Grafana OTEL LGTM | Grafana standard | 3000 | TBD | ✅ Yes |

**Questions**:

1. Should OpenTelemetry Collector port 13133 be exposed to host? (Current: No)
2. What is Grafana OTEL LGTM health check endpoint? (Need documentation)
3. Are there other third-party services with health checks to document?

## Next Steps

1. ✅ **Document Analysis** - This document
2. ⏳ **Update constitution.md** - Remove healthz, clarify livez/readyz semantics
3. ⏳ **Update spec.md** - Remove healthz references, update tables
4. ⏳ **Update copilot-instructions.md** - Add health check pattern section
5. ⏳ **Update instruction files** - Remove healthz, document third-party patterns
6. ⏳ **Align CA implementation** - Migrate to `/admin/v1/livez` + `/admin/v1/readyz`
7. ⏳ **User review** - Verify proprietary and third-party health check configurations

## References

- KMS implementation: `internal/kms/server/application/application_listener.go`
- CA implementation: `internal/ca/server/server.go`, `internal/ca/server/admin.go`
- Magic constants: `internal/shared/magic/magic_network.go`
- Docker Compose: `deployments/compose/compose.yml`, `deployments/ca/compose.yml`
- Gofiber middleware: github.com/gofiber/fiber/v2 (provides livez/readyz)
