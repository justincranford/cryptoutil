# cryptoutil Implementation Plan - Iteration 2

## Overview

This plan outlines the technical implementation approach for Iteration 2 deliverables:

1. JOSE Authority standalone service
2. CA Server REST API
3. Unified 4-product suite

---

## Phase 1: JOSE Authority (Weeks 1-2)

### 1.1 Project Structure

```
internal/jose/
├── server/
│   ├── application/      # Server lifecycle
│   ├── handler/          # HTTP handlers
│   ├── middleware/       # Auth, logging, rate limit
│   ├── service/          # Business logic
│   └── repository/       # Key storage (optional DB)
api/jose/
├── openapi_spec_components.yaml
├── openapi_spec_paths.yaml
└── openapi-gen_config_*.yaml
cmd/jose-server/
└── main.go               # Entry point
configs/jose/
└── jose-server.yml       # Configuration
deployments/jose/
├── Dockerfile.jose
└── compose.jose.yml
```

### 1.2 Implementation Order

| Order | Task | Dependencies | Assignee |
|-------|------|--------------|----------|
| 1 | OpenAPI spec (JOSE-1) | None | - |
| 2 | Server scaffolding (JOSE-2) | JOSE-1 | - |
| 3 | Key handlers (JOSE-3, JOSE-4) | JOSE-2 | - |
| 4 | Sign/Verify handlers (JOSE-5, JOSE-6) | JOSE-3 | - |
| 5 | Encrypt/Decrypt handlers (JOSE-7, JOSE-8) | JOSE-3 | - |
| 6 | JWT handlers (JOSE-9, JOSE-10) | JOSE-5, JOSE-6 | - |
| 7 | Integration tests (JOSE-11) | All handlers | - |
| 8 | Docker integration (JOSE-12) | JOSE-11 | - |

### 1.3 Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| HTTP Framework | Fiber | Consistent with Identity/KMS |
| Key Storage | In-memory + optional DB | Simple start, DB for persistence |
| Algorithm Validation | FIPS-only | Constitution requirement |
| Auth Methods | API Key, mTLS | External applications |

### 1.4 Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Algorithm incompatibility | Use existing keygen package |
| Key serialization errors | Extensive test coverage |
| Performance bottleneck | Pool-based key generation |

---

## Phase 2: CA Server (Weeks 3-5)

### 2.1 Project Structure

```
internal/ca/
├── server/               # NEW - HTTP layer
│   ├── application/      # Server lifecycle
│   ├── handler/          # HTTP handlers
│   ├── middleware/       # mTLS, auth, logging
│   └── dto/              # Request/response types
├── service/              # EXISTING - business logic
├── crypto/               # EXISTING - crypto operations
├── profile/              # EXISTING - cert profiles
├── storage/              # EXISTING - persistence
└── compliance/           # EXISTING - validation
api/ca/
├── openapi_spec_components.yaml  # EXISTING - update
├── openapi_spec_paths.yaml       # NEW
└── openapi-gen_config_*.yaml     # NEW
deployments/ca/
├── Dockerfile.ca
└── compose.ca.yml
```

### 2.2 Implementation Order

| Order | Task | Dependencies |
|-------|------|--------------|
| 1 | OpenAPI spec update (CA-1) | None |
| 2 | Server scaffolding (CA-2) | CA-1 |
| 3 | Health handler (CA-3) | CA-2 |
| 4 | CA handlers (CA-4, CA-5) | CA-3 |
| 5 | Certificate handlers (CA-6, CA-7, CA-8, CA-9) | CA-4 |
| 6 | OCSP handler (CA-10) | CA-4 |
| 7 | Profile handlers (CA-11) | CA-4 |
| 8 | EST handlers (CA-12, CA-13, CA-14, CA-15) | CA-4 |
| 9 | TSA handler (CA-16) | CA-4 |
| 10 | Integration tests (CA-17) | All handlers |
| 11 | Docker integration (CA-18) | CA-17 |

### 2.3 Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| HTTP Framework | Fiber | Consistency |
| mTLS | Required for certificate ops | Security |
| EST Protocol | RFC 7030 compliance | Industry standard |
| OCSP | RFC 6960 compliance | Required for revocation |

### 2.4 Risk Mitigation

| Risk | Mitigation |
|------|------------|
| mTLS configuration | Use existing TLS infra |
| OCSP signing latency | Cache OCSP responses |
| CRL size growth | Delta CRLs support |

---

## Phase 3: Unified Suite (Week 6)

### 3.1 Architecture Updates

```yaml
# deployments/compose.yml additions
services:
  jose-authority:
    build: ./jose
    ports: ["8083:8083", "9093:9093"]
    depends_on:
      - postgres

  ca-server:
    build: ./ca
    ports: ["8084:8084", "9094:9094"]
    depends_on:
      - postgres
```

### 3.2 Implementation Order

| Order | Task | Dependencies |
|-------|------|--------------|
| 1 | compose.yml update (UNIFIED-1) | JOSE-12, CA-18 |
| 2 | Shared secrets config (UNIFIED-2) | UNIFIED-1 |
| 3 | Service discovery (UNIFIED-3) | UNIFIED-2 |
| 4 | Health checks (UNIFIED-4) | UNIFIED-3 |
| 5 | E2E demo script (UNIFIED-5) | UNIFIED-4 |
| 6 | Documentation (UNIFIED-6) | UNIFIED-5 |

### 3.3 Inter-Service Communication

| From | To | Purpose |
|------|-----|---------|
| KMS | Identity | Token validation |
| CA | JOSE | JWT signing for CA tokens |
| All | OTEL | Telemetry export |

---

## Testing Strategy

### Unit Tests

- Table-driven with `t.Parallel()`
- Coverage ≥80% for handlers
- Mock external dependencies

### Integration Tests

- Docker Compose environment
- Real database (PostgreSQL)
- Full API workflows

### E2E Tests

- Full 4-service stack
- Demo script automation
- `cmd/demo unified` command

---

## Success Milestones

| Week | Milestone | Deliverable |
|------|-----------|-------------|
| 1 | JOSE OpenAPI + Server | Scaffolding complete |
| 2 | JOSE Handlers | All 10 endpoints working |
| 3 | CA OpenAPI + Server | Scaffolding complete |
| 4 | CA Core Handlers | Issue/revoke/status working |
| 5 | CA EST/OCSP | Protocol compliance |
| 6 | Unified Suite | 4-product deployment |

---

## Dependencies and Prerequisites

### External Dependencies

| Dependency | Version | Status |
|------------|---------|--------|
| Go | 1.25.4+ | ✅ Available |
| Fiber | v2.52+ | ✅ Available |
| GORM | v1.25+ | ✅ Available |
| PostgreSQL | 17+ | ✅ Available |

### Internal Dependencies

| Dependency | Location | Status |
|------------|----------|--------|
| keygen package | internal/common/crypto/keygen | ✅ Complete |
| JOSE primitives | internal/jose | ✅ Complete |
| CA services | internal/ca/service | ✅ Complete |
| TLS infrastructure | internal/infra/tls | ✅ Complete |

---

## Rollback Plan

| Phase | Rollback Strategy |
|-------|-------------------|
| JOSE Authority | Revert to embedded JOSE |
| CA Server | Revert to CLI-only CA |
| Unified Suite | Deploy services individually |

---

*Plan Version: 2.0.0*
*Created: January 15, 2025*
