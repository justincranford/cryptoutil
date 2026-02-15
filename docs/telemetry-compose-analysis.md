# Telemetry Compose Configuration Analysis

**Last Updated**: 2026-02-14
**Purpose**: Analyze and document the current state of OpenTelemetry and Grafana container configurations across the repository to identify consolidation opportunities.

---

## Executive Summary

**Single Source of Truth**: `deployments/telemetry/compose.yml` is the canonical source for telemetry services, successfully reused by most product services via Docker Compose `include:` directive.

**Duplications Found**: 3 files duplicate telemetry service definitions instead of using `include:`:
1. `deployments/template/compose.yml` - Full duplication of otel-collector
2. `deployments/cipher-im/` - Full duplication of both otel-collector and grafana
3. `deployments/compose/compose.yml` - Override pattern (intended, not a bug)

**Missing Service**: `healthcheck-opentelemetry-collector-contrib` is referenced as a dependency but never defined.

---

## Single Source of Truth: deployments/telemetry/compose.yml

**File**: `deployments/telemetry/compose.yml`
**Purpose**: Provides shared OpenTelemetry Collector and Grafana OTEL LGTM stack for all cryptoutil services.

### Services Defined

```yaml
services:
  opentelemetry-collector-contrib:
    image: otel/opentelemetry-collector-contrib:latest
    # Ports NOT exposed to host (allows multiple deployments)
    # Services use "opentelemetry-collector-contrib:4317" via telemetry-network

  grafana-otel-lgtm:
    image: grafana/otel-lgtm:latest
    profiles: ["with-grafana"]  # Optional
    ports:
      - "3000:3000"    # Grafana UI
      - "14317:4317"   # OTLP gRPC
      - "14318:4318"   # OTLP HTTP
```

### Design Patterns

| Pattern | Rationale |
|---------|-----------|
| No host port exposure for otel-collector | Allows multiple product deployments simultaneously without port conflicts |
| Container-to-container communication | Services reference `opentelemetry-collector-contrib:4317` via `telemetry-network` |
| Grafana as optional profile | Enable with `--profile with-grafana` when needed |
| Resource limits | 256M/128M for otel, 512M/256M for grafana |

---

## Correct Usage: Services Using Include

These services correctly reuse the canonical telemetry compose:

| Service | Compose File | Include Path | Status |
|---------|--------------|--------------|--------|
| KMS | `deployments/sm-kms/compose.yml` | `../telemetry/compose.yml` | ✅ Correct |
| PKI-CA | `deployments/pki-ca/compose.yml` | `../telemetry/compose.yml` | ✅ Correct |
| JOSE-JA | `deployments/jose-ja/compose.yml` | `../telemetry/compose.yml` | ✅ Correct |
| Identity (simple) | `deployments/identity/compose.simple.yml` | `../telemetry/compose.yml` | ✅ Correct |
| Identity (e2e) | `deployments/identity/compose.e2e.yml` | `../telemetry/compose.yml` | ✅ Correct |
| KMS Demo | `deployments/sm-kms/compose.demo.yml` | `../telemetry/compose.yml` | ✅ Correct |
| Template (docs) | `docs/compose-PRODUCT-SERVICE.yml` | `../telemetry/compose.yml` | ✅ Correct |

**Pattern**: All use `include: - path: ../telemetry/compose.yml` and reference services via `depends_on: opentelemetry-collector-contrib: condition: service_started`.

---

## Duplication Issues

### 1. deployments/template/compose.yml

**Status**: ❌ **DUPLICATION** - Should use `include:` instead

**Current State**:
- Defines its own `opentelemetry-collector-contrib` service (lines 81-99)
- Does NOT use `include:` to reference canonical telemetry compose
- REASON: Template is internal library for testing, isolated from product deployments

**Recommendation**:
- **KEEP AS-IS** - Template is intentionally isolated for validation
- Deviations are expected as it tests infrastructure patterns independently
- Document in comments why duplication exists

### 2. deployments/cipher-im/compose.yml

**Status**: ❌ **DUPLICATION** - Should use `include:` instead

**Current State**:
- Defines its own `opentelemetry-collector-contrib` (lines 67-85)
- Defines its own `grafana-otel-lgtm` (lines 86-108)
- Does NOT use `include:` to reference canonical telemetry compose

**Impact**:
- Config drift risk between cipher and canonical telemetry
- Maintenance burden (changes must be duplicated)
- Violates DRY principle

**Recommendation**:
- **REFACTOR** - Replace duplicated services with `include: - path: ../telemetry/compose.yml`
- Update references from local services to included services
- Test to ensure no behavioral changes

**Refactoring Steps**:
```yaml
# ADD at top of file
include:
  - path: ../telemetry/compose.yml

# REMOVE services section for:
#   - opentelemetry-collector-contrib (lines 67-85)
#   - grafana-otel-lgtm (lines 86-108)

# UPDATE depends_on references (likely no changes needed)
```

### 3. deployments/compose/compose.yml (E2E Testing)

**Status**: ✅ **INTENTIONAL OVERRIDE** - Not a duplication issue

**Current State**:
- Overrides `opentelemetry-collector-contrib` to expose ports (lines 46-72)
- Redefines `grafana-otel-lgtm` for E2E testing (lines 74-98)
- Contains comment explaining why: "E2E tests run from host and need to reach the health endpoint"

**Pattern**: Docker Compose override strategy
- When a service is defined in both included file and main file, main file wins
- This is intentional for E2E testing where host needs access to health endpoints

**Recommendation**:
- **KEEP AS-IS** - This is correct pattern for E2E testing
- The override is documented and intentional
- Consider adding `include: ../telemetry/compose.yml` comment for clarity

---

## Missing Service Definition: healthcheck-opentelemetry-collector-contrib

**Status**: ❌ **BUG** - Service referenced but never defined

**Files Referencing Missing Service**:
- `deployments/sm-kms/compose.yml` (lines 115, 164, 211)
- `docs/compose-PRODUCT-SERVICE.yml` (lines 147, 196, 243)

**Current Pattern**:
```yaml
depends_on:
  opentelemetry-collector-contrib:
    condition: service_started
  healthcheck-opentelemetry-collector-contrib:
    condition: service_completed_successfully  # ❌ Service doesn't exist
```

**Root Cause**:
- `opentelemetry-collector-contrib` image is minimal (no curl/wget)
- Cannot define healthcheck using standard patterns
- Intended to use ephemeral healthcheck job, but never implemented

**Impact**:
- Compose will fail with "service healthcheck-opentelemetry-collector-contrib not found"
- Unless Docker Compose silently ignores missing dependencies (unlikely)
- Likely causing deployment failures

**Recommendations**:

**Option 1: Add Ephemeral Healthcheck Job** (Preferred)
```yaml
# Add to deployments/telemetry/compose.yml
services:
  healthcheck-opentelemetry-collector-contrib:
    image: alpine:latest
    command:
      - sh
      - -c
      - |
        apk add --no-cache wget
        wget --quiet --tries=10 --retry-connrefused --waitretry=2 \
             http://opentelemetry-collector-contrib:13133/
    depends_on:
      opentelemetry-collector-contrib:
        condition: service_started
    networks:
      - telemetry-network
```

**Option 2: Remove healthcheck-opentelemetry-collector-contrib References** (Simpler)
```yaml
# Remove from all compose files
depends_on:
  opentelemetry-collector-contrib:
    condition: service_started
  # DELETE: healthcheck-opentelemetry-collector-contrib dependency
```

**Option 3: Use HTTP Healthcheck in otel-collector**
```yaml
# Add healthcheck to opentelemetry-collector-contrib in telemetry/compose.yml
healthcheck:
  test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://127.0.0.1:13133/"]
  start_period: 10s
  interval: 5s
  timeout: 3s
  retries: 5
```
⚠️ **Problem**: otel-collector-contrib image doesn't have wget/curl installed

**Recommended Solution**: **Option 2** (Remove references)
- Simplest and least fragile
- `condition: service_started` is sufficient for otel-collector
- No complex healthcheck needed for telemetry sidecar

---

## Telemetry Network Architecture

All services using telemetry connect to two networks:

```yaml
networks:
  - {product}-network      # Product-specific network (e.g., kms-network, ca-network)
  - telemetry-network      # Shared telemetry network (defined in telemetry/compose.yml)
```

**Communication Flow**:
```
cryptoutil-service → opentelemetry-collector-contrib:4317 → grafana-otel-lgtm:4317
```

**Benefits**:
- Network isolation between products
- Shared telemetry infrastructure
- Container-to-container communication (no host port exposure)

---

## Recommended Refactoring Plan

### Phase 1: Fix Missing Service (Immediate)

**Action**: Remove `healthcheck-opentelemetry-collector-contrib` references
**Files**:
- `deployments/sm-kms/compose.yml` (3 locations)
- `docs/compose-PRODUCT-SERVICE.yml` (3 locations)

**Impact**: Low risk, simplifies dependencies

### Phase 2: Consolidate Cipher (High Priority)

**Action**: Replace duplicated services with `include:` directive
**Files**:
- `deployments/cipher-im/compose.yml`

**Benefits**:
- Single source of truth
- Config consistency
- Easier maintenance

**Testing**:
- `docker compose -f deployments/cipher-im/compose.yml up -d`
- Verify cipher-im connects to otel and grafana
- Check `docker network ls` shows telemetry-network

### Phase 3: Document Template (Low Priority)

**Action**: Add comments explaining intentional duplication
**Files**:
- `deployments/template/compose.yml`

**Rationale**: Template is internal library, isolation is intentional

### Phase 4: Template Update (Optional)

**Action**: Create `compose-PRODUCT.yml` template for multiple services per product
**Files**:
- `docs/compose-PRODUCT.yml` (new file, copy from `compose-PRODUCT-SERVICE.yml`)

**Changes**:
- Support 1-5 services per product
- 3 instances per service (1 SQLite + 2 PostgreSQL)
- Shared PostgreSQL database across all services in product
- Dynamic port allocation based on product base port

---

## File Location Reference

| File | Purpose | Include? | Status |
|------|---------|----------|--------|
| `deployments/telemetry/compose.yml` | Canonical source | N/A (is source) | ✅ Source of Truth |
| `deployments/sm-kms/compose.yml` | KMS deployment | ✅ Yes | ✅ Correct |
| `deployments/pki-ca/compose.yml` | PKI-CA deployment | ✅ Yes | ✅ Correct |
| `deployments/jose-ja/compose.yml` | JOSE-JA deployment | ✅ Yes | ✅ Correct |
| `deployments/cipher-im/compose.yml` | Cipher-IM deployment | ❌ No | ❌ Needs Refactor |
| `deployments/template/compose.yml` | Template validation | ❌ No | ⚠️ Intentional |
| `deployments/compose/compose.yml` | E2E testing | ❌ Override | ✅ Intentional |
| `deployments/identity/compose.simple.yml` | Identity simple | ✅ Yes | ✅ Correct |
| `deployments/identity/compose.e2e.yml` | Identity E2E | ✅ Yes | ✅ Correct |
| `docs/compose-PRODUCT-SERVICE.yml` | Template doc | ✅ Yes | ✅ Correct |

---

## Conclusion

**Overall Assessment**: ✅ Mostly Good

**Strengths**:
1. Clear single source of truth (`deployments/telemetry/compose.yml`)
2. Majority of services correctly use `include:` directive
3. Good network isolation and resource limits

**Improvements Needed**:
1. **Critical**: Fix missing `healthcheck-opentelemetry-collector-contrib` service (blocks deployments)
2. **High**: Consolidate `deployments/cipher-im/compose.yml` to use `include:`
3. **Low**: Document intentional duplication in template

**Next Steps**: Execute Phase 1 (remove missing healthcheck references) immediately, then Phase 2 (consolidate cipher) in next sprint.

---

## UPDATE 2026-02-14: PostgreSQL Single Source of Truth

**New Infrastructure**: Created `deployments/postgres/compose.yml` as canonical source for PostgreSQL infrastructure.

### PostgreSQL Services

- **postgres-leader**: OLTP read-write primary (port 5432, 2GB RAM, 27 logical databases)
- **postgres-follower**: OLAP read-only replica (port 5433, 3GB RAM, logical replication)
- **citus-coordinator**: Distributed PostgreSQL coordinator (port 5434)
- **citus-worker-1/2**: Worker nodes for row-level sharding (ports 5435/5436)
- **citus-setup**: Ephemeral configuration job

### Database Architecture (27 Logical Databases)

**Suite Level** (1 database):
- `suitedeployment-cryptoutil`  9 schemas (all services)

**Product Level** (5 databases):
- `productdeployment-pki`  1 schema (ca)
- `productdeployment-jose`  1 schema (ja)
- `productdeployment-cipher`  1 schema (im)
- `productdeployment-sm`  1 schema (kms)
- `productdeployment-identity`  5 schemas (authz, idp, rs, rp, spa)

**Service Level** (9 databases):
- `servicedeployment-pki-ca`, `servicedeployment-jose-ja`, `servicedeployment-cipher-im`, `servicedeployment-sm-kms`
- `servicedeployment-identity-authz`, `servicedeployment-identity-idp`, `servicedeployment-identity-rs`, `servicedeployment-identity-rp`, `servicedeployment-identity-spa`

### Template Compose Files Created

1. **deployments/template/compose-cryptoutil-PRODUCT-SERVICE.yml**: Single-service deployment (3 instances)
2. **deployments/template/compose-cryptoutil-PRODUCT.yml**: Product deployment (1-5 services, 3 instances each)
3. **deployments/template/compose-cryptoutil.yml**: Suite deployment (all 9 services, 27 total instances)

### Migration Status

**Completed**:
- deployments/telemetry/compose.yml - Added healthcheck-opentelemetry-collector-contrib ephemeral job
- deployments/postgres/compose.yml - Full leader/follower/Citus infrastructure
- deployments/template/compose.yml - Made generic (removed cipher-im specificity)
- deployments/compose/compose.yml - Already uses postgres include
- deployments/sm-kms/compose.yml - Added postgres include

**Pending** (local postgres service removal + depends_on updates required):
- deployments/pki-ca/compose.yml
- deployments/pki-ca/compose/compose.yml
- deployments/cipher-im/compose.yml
- deployments/jose-ja/compose.yml
- deployments/identity/compose.yml

### Required Updates

For each pending file:
1. Add `- path: ../postgres/compose.yml` to include section
2. Remove local postgres service definitions (e.g., `sm-kms-db-postgres-1`)
3. Update `depends_on`: local postgres names  `postgres-leader`
4. Add `postgres-network` to services connecting to postgres
5. Remove local postgres volumes

### New Best Practice

**ALL compose files MUST include both**:
```yaml
include:
  - path: ../telemetry/compose.yml
  - path: ../postgres/compose.yml
```

This eliminates duplicate service definitions and ensures consistent infrastructure across all deployments.
