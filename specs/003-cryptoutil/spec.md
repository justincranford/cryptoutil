# cryptoutil Specification - Iteration 3

## Overview

**Iteration 3** focuses on **CI/CD reliability, test optimization, and completing deferred work from iteration 2**.

**Context**: Iteration 2 delivered JOSE Authority (10 endpoints), CA Server (9 endpoints), and comprehensive observability infrastructure. However, **8 of 11 CI/CD workflows are failing (27% pass rate)**, coverage is below new 95% threshold, and several key features were deferred.

**Critical Issues**:

- DATA RACE in CA handler tests (production bug)
- Coverage below 95% threshold (Identity ORM at 67.5%)
- E2E/Load/DAST workflows failing due to service startup issues
- Mutation testing tool (gremlins) blocked by upstream bug

---

## Iteration Scope

### Primary Goals

1. **Fix Critical CI/CD Failures**: Achieve 100% workflow pass rate (currently 27%)
2. **Increase Test Coverage**: Identity ORM 67.5% → 95%+, overall 90% → 95%+
3. **Optimize Test Runtime**: Reduce feedback loop from 28min → <10min
4. **Complete Deferred Work**: JOSE Docker, CA OCSP/EST handlers, unified testing suite

### Success Criteria

| Criterion | Target Metric |
|-----------|---------------|
| Workflow Pass Rate | 11/11 workflows passing (100%) |
| Test Coverage | ≥95% production, ≥100% infrastructure |
| Mutation Testing | Manual for critical paths (gremlins blocked) |
| CI Feedback Loop | <10 minutes for full suite |
| Deferred Tasks | 4/4 major deferred tasks completed |

---

## Iteration 3 Deliverables

### D1: CI/CD Reliability Fixes

**Purpose**: Achieve 100% workflow pass rate and <10min feedback loop

**Current State**: 8/11 workflows failing (27% pass rate), 28min bottleneck (race detection)

**Target State**: All workflows passing, optimized feedback loop

#### Critical Fixes

| Issue | Location | Priority | Status |
|-------|----------|----------|--------|
| DATA RACE in CA handler | `internal/ca/api/handler/handler_comprehensive_test.go:1502` | 🔴 CRITICAL | ❌ Not Started |
| Identity ORM test failures | `internal/identity/repository/orm/consent_decision_repository_test.go:160` | 🔴 HIGH | ❌ Not Started |
| Identity ORM low coverage | 67.5% → 95%+ | 🔴 HIGH | ❌ Not Started |
| E2E/Load/DAST startup failures | Docker Compose health checks | 🔴 CRITICAL | ❌ Not Started |
| Quality/lint errors | golangci-lint violations | ⚠️ MEDIUM | ❌ Not Started |

#### Optimization Opportunities

| Optimization | Impact | Effort | Priority | Status |
|--------------|--------|--------|----------|--------|
| Path filters | Skip docs-only workflows | Low | 🟡 P1 | ❌ |
| Parallel lint+test | -3min feedback | Medium | 🟡 P1 | ❌ |
| Cache Go modules | -25s/run | Low | 🟢 P2 | ❌ |

---

### D2: Deferred Work from Iteration 2

**Purpose**: Complete incomplete features from iteration 2

**Current State**: 4 major features deferred due to time constraints

**Target State**: All deferred features completed and production-ready

#### Deferred Features

| Feature | Package | Priority | Estimate | Status |
|---------|---------|----------|----------|--------|
| JOSE Docker Integration | `deployments/jose/` | HIGH | 2h | ❌ Not Started |
| CA OCSP Handler | `internal/ca/api/handler/` | HIGH | 6h | ❌ Not Started |
| CA EST Handler | `internal/ca/api/handler/` | MEDIUM | 4h | ❌ Not Started |
| Unified E2E Test Suite | `internal/test/e2e/` | MEDIUM | 3h | ❌ Not Started |

---

### D3: Test Methodology Enhancements

**Purpose**: Comprehensive test coverage with modern testing techniques

**Current State**: Traditional unit/integration tests only

**Target State**: Benchmarks, fuzz tests, property-based tests (where applicable)

#### Test Enhancements

| Enhancement | Scope | Priority | Status |
|-------------|-------|----------|--------|
| Benchmark Tests | All crypto operations | 🔴 MANDATORY | ❌ Not Started |
| Fuzz Tests | All parsers/validators | 🔴 MANDATORY | ❌ Not Started |
| Property-Based Tests | Crypto invariants | 🟡 RECOMMENDED | ❌ Not Started |
| Manual Mutation Tests | Critical paths | 🔴 REQUIRED (gremlins blocked) | ❌ Not Started |

#### Benchmark Targets

| Package | Operations | Target |
|---------|------------|--------|
| `internal/common/crypto/keygen` | Key generation (RSA, EC, EdDSA, AES, HMAC) | All algorithms |
| `internal/crypto` | Encrypt/decrypt, sign/verify, hash | Happy + sad paths |
| `internal/jose` | JWE compact/JSON, JWS compact/JSON | All serializations |

#### Fuzz Test Targets

| Package | Handlers | Runtime |
|---------|----------|---------|
| `internal/identity/authz` | Token parsers | ≥15s |
| `internal/jose` | JOSE parsers | ≥15s |
| `internal/ca` | Certificate request parsers | ≥15s |

---

### D4: Documentation and Knowledge Transfer

**Purpose**: Clean up temporary docs, consolidate lessons learned

**Current State**: 2 DELETE-ME files, scattered notes

**Target State**: Clean docs structure, comprehensive runbooks

#### Documentation Tasks

| Task | Files | Priority | Status |
|------|-------|----------|--------|
| Process DELETE-ME files | `docs/DELETE-ME-LATER-*.md` (2 files) | MEDIUM | ❌ Not Started |
| Process NOT-FINISHED | `docs/NOT-FINISHED.md` | LOW | ❌ Not Started |
| Update runbooks | `docs/runbooks/*.md` | MEDIUM | ❌ Not Started |
| Workflow optimization guide | Based on `docs/workflow-analysis.md` | MEDIUM | ❌ Not Started |

---

## Architecture

### System Topology

```
┌─────────────────────────────────────────────────────────────────┐
│                    [Architecture Diagram]                        │
├─────────────────────────────────────────────────────────────────┤
│  [Component boxes showing relationships]                         │
│  [Database connections]                                          │
│  [External service integrations]                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Service Ports

| Service | Public API | Admin API | Database | Status |
|---------|------------|-----------|----------|--------|
| service-name | 8080 | 9090 | PostgreSQL | ❌ |

### Project Structure

```
internal/product/
├── server/
│   ├── application/      # Server lifecycle, config
│   ├── handler/          # HTTP handlers, routing
│   ├── middleware/       # Auth, logging, rate limit
│   └── dto/              # Request/response types
├── service/              # Business logic
├── repository/           # Data persistence
├── domain/               # Domain models
└── magic/                # Magic constants
api/product/
├── openapi_spec_components.yaml
├── openapi_spec_paths.yaml
└── openapi-gen_config_*.yaml
cmd/product-server/
└── main.go               # Entry point
configs/product/
└── product-server.yml    # Configuration
deployments/product/
├── Dockerfile.product
└── compose.product.yml
```

---

## Functional Requirements

### FR1: [Requirement Category]

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR1.1 | [Detailed requirement] | HIGH | ❌ |
| FR1.2 | [Detailed requirement] | MEDIUM | ❌ |

### FR2: [Requirement Category]

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR2.1 | [Detailed requirement] | HIGH | ❌ |
| FR2.2 | [Detailed requirement] | LOW | ❌ |

---

## Non-Functional Requirements

### NFR1: Security

| ID | Requirement | Target | Measurement |
|----|-------------|--------|-------------|
| NFR1.1 | FIPS 140-3 compliance | 100% algorithms | Algorithm validation tests |
| NFR1.2 | Secret management | Docker/K8s secrets | No env vars for secrets |
| NFR1.3 | TLS minimum version | TLS 1.3+ | Configuration enforcement |
| NFR1.4 | Audit logging | All operations | Telemetry audit logs |

### NFR2: Performance

| ID | Requirement | Target | Measurement |
|----|-------------|--------|-------------|
| NFR2.1 | API response time | p95 < 100ms | Benchmarks |
| NFR2.2 | Throughput | ≥1000 req/sec | Load tests |
| NFR2.3 | Database queries | p95 < 50ms | Query performance |

### NFR3: Reliability

| ID | Requirement | Target | Measurement |
|----|-------------|--------|-------------|
| NFR3.1 | Uptime | ≥99.9% | Health checks |
| NFR3.2 | Error rate | <0.1% | Error logs |
| NFR3.3 | Graceful shutdown | <5s | Shutdown tests |

### NFR4: Quality

| ID | Requirement | Target | Measurement |
|----|-------------|--------|-------------|
| NFR4.1 | Code coverage | ≥95% production | `go test -cover` |
| NFR4.2 | Infrastructure coverage | ≥100% cicd | `go test -cover ./internal/cmd/cicd/...` |
| NFR4.3 | Utility coverage | 100% | `go test -cover ./internal/common/util/...` |
| NFR4.4 | Linting | 0 errors | `golangci-lint run` |
| NFR4.5 | File size limits | <500 lines | Manual review |
| NFR4.6 | Mutation score | ≥80% | `gremlins unleash` |

### NFR5: Testability

| ID | Requirement | Target | Measurement |
|----|-------------|--------|-------------|
| NFR5.1 | Table-driven tests | 100% test files | Code review |
| NFR5.2 | Parallel tests | `t.Parallel()` | Test execution |
| NFR5.3 | Benchmark tests | All hot paths | `*_bench_test.go` files |
| NFR5.4 | Fuzz tests | All parsers/validators | `*_fuzz_test.go` files |
| NFR5.5 | Property tests | Invariants | gopter tests |
| NFR5.6 | Integration tests | E2E workflows | `*_integration_test.go` |

### NFR6: Observability

| ID | Requirement | Target | Measurement |
|----|-------------|--------|-------------|
| NFR6.1 | Structured logging | All operations | Log format validation |
| NFR6.2 | OpenTelemetry tracing | All requests | OTLP export |
| NFR6.3 | Prometheus metrics | Key metrics | Metrics endpoint |
| NFR6.4 | Health endpoints | /livez, /readyz | HTTP checks |

### NFR7: Deployment

| ID | Requirement | Target | Measurement |
|----|-------------|--------|-------------|
| NFR7.1 | Docker image | Multi-stage, static | `ldd` check |
| NFR7.2 | Container size | <100MB | Image inspection |
| NFR7.3 | Startup time | <10s | Deployment logs |
| NFR7.4 | Configuration | YAML files | No env vars |

---

## Dependencies

### From Previous Iterations

| Dependency | Status | Version | Notes |
|------------|--------|---------|-------|
| Iteration 1 - Identity V2 | ✅ Complete | 1.0.0 | OAuth 2.1 + OIDC |
| Iteration 1 - KMS | ✅ Complete | 1.0.0 | Key management |
| Iteration 2 - JOSE Authority | ⚠️ 92% | 2.0.0 | Needs Docker |
| Iteration 2 - CA Server | ⚠️ 65% | 2.0.0 | Needs OCSP/EST |

### External Dependencies

| Package | Version | Purpose | License |
|---------|---------|---------|---------|
| github.com/gofiber/fiber/v2 | v2.x | HTTP framework | MIT |
| gorm.io/gorm | v1.x | ORM | MIT |
| modernc.org/sqlite | Latest | CGO-free SQLite | BSD |

---

## Testing Strategy

### Unit Tests

- Table-driven with `t.Parallel()`
- Coverage ≥95% for production code
- Coverage ≥100% for infrastructure/utility code
- Mock external dependencies
- Test both happy and sad paths

### Integration Tests

- Docker Compose environment
- Real database (PostgreSQL and SQLite)
- Full API workflows
- Tagged with `//go:build integration`

### Benchmark Tests

- All cryptographic operations
- All hot path handlers
- Database queries
- Baseline metrics documented

### Fuzz Tests

- All input parsers
- All validators
- All cryptographic operations
- 15s minimum fuzz time

### Property-Based Tests

- Invariant validation
- Round-trip encoding/decoding
- Cryptographic properties (gopter)

### Mutation Tests

- Gremlins on all packages
- Target ≥80% mutation score
- Baseline report per package

### E2E Tests

- Full service stack
- Real telemetry infrastructure
- Demo script automation

---

## Risk Management

### Technical Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| [Risk 1] | HIGH/MEDIUM/LOW | HIGH/MEDIUM/LOW | [Mitigation strategy] |
| [Risk 2] | HIGH/MEDIUM/LOW | HIGH/MEDIUM/LOW | [Mitigation strategy] |

### Schedule Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| [Risk 1] | [Impact] | [Mitigation] |

### Quality Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Test coverage below target | Regression bugs | Enforce pre-commit coverage checks |
| Mutation score low | Weak tests | Run gremlins regularly |

---

## Timeline

### Phase 1: [Phase Name] (Week X-Y)

- [Task group 1]
- [Task group 2]
- Target: [Deliverable]

### Phase 2: [Phase Name] (Week Y-Z)

- [Task group 1]
- [Task group 2]
- Target: [Deliverable]

### Phase 3: [Phase Name] (Week Z-W)

- [Task group 1]
- [Task group 2]
- Target: [Deliverable]

---

## Completion Criteria

### Pre-Implementation Gates

- [ ] All `[NEEDS CLARIFICATION]` markers resolved
- [ ] CLARIFICATIONS.md created and reviewed
- [ ] All requirements have corresponding tasks
- [ ] Dependencies identified and available

### Post-Implementation Gates

- [ ] `go build ./...` passes clean
- [ ] `golangci-lint run` passes with 0 errors
- [ ] `go test ./...` passes with ≥95% coverage (production), ≥100% (infra/util)
- [ ] All benchmarks run successfully
- [ ] All fuzz tests run for ≥15s
- [ ] Gremlins mutation score ≥80%
- [ ] Docker Compose deployment healthy
- [ ] Integration tests passing
- [ ] E2E demo script working
- [ ] Documentation updated (README, runbooks)
- [ ] PROGRESS.md up-to-date
- [ ] EXECUTIVE-SUMMARY.md created
- [ ] CHECKLIST-ITERATION-NNN.md complete

---

## Version Control

| Attribute | Value |
|-----------|-------|
| **Specification Version** | N.0.0 |
| **Created** | [Date] |
| **Last Updated** | [Date] |
| **Status** | ❌ Not Started / ⚠️ In Progress / ✅ Complete |

---

## Template Usage Notes

**For LLM Agents**: This template includes:
- ✅ Mandatory functional requirements section
- ✅ Mandatory non-functional requirements section (security, performance, quality, testability, observability, deployment)
- ✅ Coverage targets: 95% production, 100% infrastructure/utility
- ✅ Testing requirements: unit, integration, benchmark, fuzz, property, mutation, E2E
- ✅ Quality gates: build, lint, test, coverage, mutation score
- ✅ FIPS 140-3 compliance requirements
- ✅ Docker/Kubernetes secret management requirements
- ✅ File size limits (300/400/500 lines)
- ✅ Table-driven test requirements with `t.Parallel()`

**Customization**:
- Replace [PLACEHOLDERS] with actual values
- Remove unused sections for simpler iterations
- Add iteration-specific sections as needed
- Keep scope focused and achievable

**References**:
- Constitution: `.specify/memory/constitution.md`
- Copilot Instructions: `.github/instructions/*.md`
- Feature Template: `docs/feature-template/`
- Previous Iterations: `specs/001-cryptoutil/`, `specs/002-cryptoutil/`
