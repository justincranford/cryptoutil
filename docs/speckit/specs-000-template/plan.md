# cryptoutil Implementation Plan Template - Iteration NNN

## Overview

This plan outlines the technical implementation approach for Iteration NNN deliverables:

1. [Deliverable 1]
2. [Deliverable 2]
3. [Deliverable 3]

**Estimated Total Effort**: ~[X] hours

---

## CRITICAL: Test Concurrency Requirements

**!!! NEVER use `-p=1` or `-parallel=1` in test commands !!!**
**!!! ALWAYS use concurrent test execution with `-shuffle=on` !!!**

**Mandatory Test Execution**:

```bash
# CORRECT
go test ./... -cover -shuffle=on

# WRONG - NEVER DO THIS
go test ./... -p=1  # ❌ Hides concurrency bugs
```

**Test Data Isolation**:

- ✅ ALWAYS use UUIDv7 for test data uniqueness
- ✅ ALWAYS use dynamic ports (port 0 pattern)
- ✅ ALWAYS use TestMain for shared dependencies
- ✅ Real dependencies preferred (test containers, in-memory services)
- ✅ Mocks ONLY for hard-to-reach corner cases

**Rationale**: Concurrent tests provide fastest execution and reveal production concurrency bugs.

---

## Phase 1: [Phase Name] (Week X-Y, ~[H] hours)

### 1.1 Project Structure

```
internal/product/
├── server/
│   ├── application/      # Server lifecycle
│   │   ├── application.go
│   │   ├── application_test.go
│   │   └── config.go
│   ├── handler/          # HTTP handlers
│   │   ├── handler.go
│   │   ├── handler_test.go
│   │   └── dto/
│   ├── middleware/       # Auth, logging, rate limit
│   │   ├── auth.go
│   │   ├── logging.go
│   │   └── ratelimit.go
│   └── service/          # Business logic
│       ├── service.go
│       └── service_test.go
├── repository/           # Data persistence
│   ├── repository.go
│   ├── repository_test.go
│   └── orm/
│       ├── orm.go
│       └── orm_test.go
├── domain/               # Domain models
│   ├── models.go
│   └── nullable_uuid.go
└── magic/                # Magic constants
    └── magic.go
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

### 1.2 Implementation Order

| Order | Task ID | Task | Dependencies | LOE |
|-------|---------|------|--------------|-----|
| 1 | TASK-1 | [Task description] | None | 2h |
| 2 | TASK-2 | [Task description] | TASK-1 | 4h |
| 3 | TASK-3 | [Task description] | TASK-2 | 3h |
| 4 | TASK-4 | [Task description] | TASK-2 | 3h |
| 5 | TASK-5 | [Task description] | TASK-3, TASK-4 | 6h |

**Phase 1 Subtotal**: ~[H] hours

### 1.3 Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| HTTP Framework | Fiber | Consistency with existing services |
| Database | PostgreSQL + SQLite | Production + dev/test |
| ORM | GORM | Cross-database compatibility |
| Authentication | mTLS + JWT | Security requirements |
| Configuration | YAML files | No env vars per constitution |
| Testing | Table-driven + parallel | Constitution requirement |

### 1.4 Risk Mitigation

| Risk | Impact | Mitigation Strategy |
|------|--------|---------------------|
| [Risk 1] | HIGH/MEDIUM/LOW | [Mitigation approach] |
| [Risk 2] | HIGH/MEDIUM/LOW | [Mitigation approach] |

---

## Phase 2: [Phase Name] (Week Y-Z, ~[H] hours)

### 2.1 Project Structure

[Any new directories/files specific to this phase]

### 2.2 Implementation Order

| Order | Task ID | Task | Dependencies | LOE |
|-------|---------|------|--------------|-----|
| 1 | TASK-6 | [Task description] | Phase 1 complete | 4h |
| 2 | TASK-7 | [Task description] | TASK-6 | 6h |
| 3 | TASK-8 | [Task description] | TASK-7 | 3h |

**Phase 2 Subtotal**: ~[H] hours

### 2.3 Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| [Decision 1] | [Choice] | [Rationale] |
| [Decision 2] | [Choice] | [Rationale] |

### 2.4 Risk Mitigation

| Risk | Impact | Mitigation Strategy |
|------|--------|---------------------|
| [Risk 1] | HIGH/MEDIUM/LOW | [Mitigation approach] |

---

## Phase 3: [Phase Name] (Week Z-W, ~[H] hours)

### 3.1 Integration Architecture

```yaml
# deployments/compose/compose.yml additions
services:
  product-server:
    build:
      context: .
      dockerfile: deployments/product/Dockerfile.product
    ports:
      - "8080:8080"  # Public API
      - "9090:9090"  # Admin API
    secrets:
      - database_url
      - unseal_secret_1
      - unseal_secret_2
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/v1/livez"]
      interval: 10s
      timeout: 5s
      retries: 3
```

### 3.2 Implementation Order

| Order | Task ID | Task | Dependencies | LOE |
|-------|---------|------|--------------|-----|
| 1 | TASK-9 | Docker Compose integration | Phase 1 & 2 | 4h |
| 2 | TASK-10 | Service discovery | TASK-9 | 2h |
| 3 | TASK-11 | Health checks | TASK-10 | 2h |
| 4 | TASK-12 | E2E demo script | TASK-11 | 6h |
| 5 | TASK-13 | Documentation | All tasks | 4h |

**Phase 3 Subtotal**: ~[H] hours

### 3.3 Inter-Service Communication

| From | To | Purpose | Protocol |
|------|----|---------|----------|
| Product | Identity | Token validation | HTTPS/gRPC |
| Product | KMS | Key operations | HTTPS |
| Product | Database | Data persistence | PostgreSQL |
| Product | OTEL | Telemetry export | OTLP/gRPC |

---

## Testing Strategy

### Unit Tests (≥95% coverage production, ≥98% infrastructure/utility)

**File Naming**: `*_test.go`

**Requirements**:

- Table-driven tests with `t.Parallel()`
- Test helpers marked with `t.Helper()`
- No magic values (use runtime UUIDv7 or magic constants)
- Dynamic port allocation (port 0 pattern)

**Coverage Targets by Package Type**:

- Production code: ≥95%
- Infrastructure (cicd): ≥98%
- Utility code: 100%

### Integration Tests

**File Naming**: `*_integration_test.go`

**Build Tag**: `//go:build integration`

**Requirements**:

- Docker Compose environment
- Real database (PostgreSQL + SQLite tests)
- Full API workflows
- Cleanup after tests

### Benchmark Tests (All hot paths)

**File Naming**: `*_bench_test.go`

**Requirements**:

- All cryptographic operations
- All API endpoints
- Database operations
- Baseline metrics documented

**Execution**:

```bash
go test -bench=. -benchmem ./internal/product/...
```

### Fuzz Tests (All parsers/validators)

**File Naming**: `*_fuzz_test.go`

**Requirements**:

- Unique fuzz function names (not substrings)
- All input parsers
- All validators
- Minimum 15s fuzz time

**Execution**:

```bash
go test -fuzz=FuzzFunctionName -fuzztime=15s ./internal/product/...
```

### Property-Based Tests (Invariants)

**Library**: gopter

**Requirements**:

- Round-trip encoding/decoding
- Cryptographic properties
- Invariant validation

### Mutation Tests (≥98% mutation score)

**Tool**: gremlins

**Requirements**:

- Baseline per package
- Target ≥80% mutation score
- Regular execution (weekly/monthly)

**Execution**:

```bash
gremlins unleash
```

### E2E Tests

**File Naming**: `*_e2e_test.go` or in `test/e2e/`

**Requirements**:

- Full service stack
- Real telemetry infrastructure
- Demo script automation
- `cmd/demo product` command

---

## Quality Gates

### Pre-Commit Gates

- [ ] `go build ./...` passes clean
- [ ] `golangci-lint run --fix` resolves all issues
- [ ] `golangci-lint run` passes with 0 errors
- [ ] File sizes ≤500 lines (refactor if exceeded)
- [ ] UTF-8 without BOM encoding
- [ ] No new TODOs without tracking

### Pre-Push Gates

- [ ] `go test ./...` passes all tests
- [ ] Coverage ≥95% production, ≥98% infrastructure/utility
- [ ] All benchmarks run successfully
- [ ] Dependency checks pass
- [ ] Pre-commit hooks pass

### Pre-Merge Gates

- [ ] All CI workflows passing
- [ ] Code review approved
- [ ] Integration tests passing
- [ ] E2E tests passing
- [ ] Docker Compose deployment healthy
- [ ] Documentation updated

---

## Success Milestones

| Week | Milestone | Deliverable | Verification |
|------|-----------|-------------|--------------|
| X | Phase 1 Complete | [Deliverable] | [How to verify] |
| Y | Phase 2 Complete | [Deliverable] | [How to verify] |
| Z | Phase 3 Complete | [Deliverable] | [How to verify] |
| W | Iteration NNN Done | All tasks complete | CHECKLIST-ITERATION-NNN.md ✅ |

---

## Evidence-Based Completion Checklist

### Code Quality

- [ ] `go build ./...` clean
- [ ] `golangci-lint run` clean (0 errors)
- [ ] No new TODOs without tracking in tasks.md
- [ ] File sizes within limits (≤500 lines)
- [ ] UTF-8 without BOM encoding

### Test Coverage

- [ ] `go test ./... -shuffle=on` passes (concurrent execution)
- [ ] Coverage ≥95% production: `go test -cover ./internal/product/...`
- [ ] Coverage ≥98% infrastructure: `go test -cover ./internal/cmd/cicd/...`
- [ ] Coverage 98% utility: `go test -cover ./internal/common/util/...`
- [ ] No skipped tests without tracking

### Benchmarks

- [ ] All cryptographic operations benchmarked
- [ ] All hot path handlers benchmarked
- [ ] Baseline metrics documented
- [ ] No performance regressions

### Fuzz Tests

- [ ] All parsers fuzzed for ≥15s
- [ ] All validators fuzzed for ≥15s
- [ ] Crash-free fuzz execution

### Mutation Tests

- [ ] Gremlins baseline report created
- [ ] Mutation score ≥98% per package
- [ ] Weak tests identified and improved

### Integration

- [ ] Docker Compose deploys successfully
- [ ] All services report healthy
- [ ] E2E demo script passes
- [ ] Inter-service communication working

### Documentation

- [ ] README.md updated with new features
- [ ] API documentation generated (OpenAPI)
- [ ] Runbooks created for operations
- [ ] implement/DETAILED.md Section 2 (timeline) updated
- [ ] implement/EXECUTIVE.md created/updated
- [ ] CHECKLIST-ITERATION-NNN.md complete

---

## Dependency Management

### Version Requirements

- Go: 1.25.5+
- PostgreSQL: 14+
- Docker: 24+
- Docker Compose: v2+
- golangci-lint: v2.6.2+
- Node: v24.11.1+ (for Gatling load tests)

### Updating Dependencies

```bash
# Check for updates
go list -u -m all | grep '\[.*\]$'

# Update incrementally
go get -u [package]
go mod tidy

# Test after each update
go test ./...
golangci-lint run
```

---

## Workflow Integration

### CI/CD Pipelines

| Workflow | Trigger | Purpose | Duration |
|----------|---------|---------|----------|
| ci-quality | PR, push | Linting, formatting, builds | ~5 min |
| ci-coverage | PR, push | Test coverage | ~10 min |
| ci-benchmark | PR, push | Performance benchmarks | ~5 min |
| ci-fuzz | Scheduled | Fuzz testing | ~15 min |
| ci-race | PR, push | Race detection | ~15 min |
| ci-sast | PR, push | Static security | ~5 min |
| ci-dast | Scheduled | Dynamic security | ~10 min |
| ci-e2e | PR, push | Integration tests | ~20 min |

### Artifact Management

- Upload test reports: `actions/upload-artifact@v5.0.0`
- Upload SARIF: `github/codeql-action/upload-sarif@v3`
- Retention: 1 day (temporary), 30 days (reports)

---

## Post-Mortem Template

### What Went Well

- [Success 1]
- [Success 2]

### What Needs Improvement

- [Issue 1]
- [Issue 2]

### Action Items for Next Iteration

- [Action 1]
- [Action 2]

---

## Template Usage Notes

**For LLM Agents**: This plan template includes:

- ✅ Phase breakdown with LOE estimates
- ✅ Project structure showing file organization
- ✅ Implementation order with dependencies
- ✅ Technical decisions documentation
- ✅ Risk mitigation strategies
- ✅ Comprehensive testing strategy (unit, integration, benchmark, fuzz, property, mutation, E2E)
- ✅ Quality gates (pre-commit, pre-push, pre-merge)
- ✅ Success milestones with verification
- ✅ Evidence-based completion checklist
- ✅ CI/CD workflow integration

**Customization**:

- Adjust phase count based on iteration complexity
- Update LOE estimates based on actual experience
- Add/remove technical decisions as needed
- Tailor risk mitigation to specific iteration

**References**:

- spec.md: Requirements and functional specifications
- tasks.md: Detailed task breakdown
- Constitution: Quality and testing requirements
- Copilot Instructions: Implementation patterns
