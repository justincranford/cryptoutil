
---

## Appendix B: Reference Tables

### B.1 Technology Stack

<!-- @to-appendix as="minimum-versions" appendixes=".github/instructions/02-02.versions.instructions.md" -->
**CRITICAL: ALWAYS use the same version everywhere** (dev, CI/CD, Docker, workflows, docs)

- Go: 1.26.1+
- Python: 3.14+
- golangci-lint: v2.12.2+
- Node: v24.11.1+ LTS
- Java: 21 LTS (Gatling load tests)
- Maven: 3.9+
- pre-commit: 2.20.0+
- Docker: 27+
- Docker Compose: v5+
<!-- @/to-appendix -->

**Languages**: Go 1.26.1 (services), Python 3.14+ (utilities), Node v24.11.1+ (CLI tools)
**Databases**: PostgreSQL 18, SQLite (modernc.org/sqlite, CGO-free)
**Frameworks**: Fiber (HTTP), GORM (ORM), oapi-codegen (OpenAPI)
**Container Base**: Alpine Linux latest (all Dockerfiles, unpinned for security patches)
**Observability**: OpenTelemetry (otel-collector-contrib:latest), Grafana LGTM (grafana/otel-lgtm:latest)
**Security**: FIPS 140-3 approved algorithms, Docker/Kubernetes secrets
**Testing**: testify, gremlins (mutation), Nuclei/ZAP (DAST), Gatling (load)

### B.2 Dependency Matrix

**Core Dependencies**:

- github.com/gofiber/fiber/v3 (HTTP framework)
- gorm.io/gorm (ORM)
- github.com/google/uuid/v7 (UUIDv7)
- go.opentelemetry.io/otel (telemetry)
- github.com/go-jose/go-jose/v4 (JOSE)

**Test Dependencies**: testify, testcontainers-go, httptest

### B.3 Configuration Reference

**Priority Order**: Docker secrets > YAML > CLI parameters (NO env vars for secrets)

**Standard Files**:

- config.yml: Main configuration
- secrets/*.secret: Credentials (chmod 440)

### B.4 Instruction File Reference

**See .github/copilot-instructions.md** for complete table of 18 instruction files

**Summary**: 01-terminology/beast-mode, 02-architecture (5 files), 03-development (4 files), 04-deployment (1 file), 05-platform (2 files), 06-evidence (2 files)

### B.5 Agent Catalog & Handoff Matrix

| Copilot Agent | Claude Code Agent | Description | Handoffs |
|--------------|------------------|-------------|----------|
| `copilot-implementation-planning` | `claude-implementation-planning` | Planning and task decomposition | → implementation-execution |
| `copilot-implementation-execution` | `claude-implementation-execution` | Autonomous implementation execution | → fix-workflows |
| `copilot-fix-workflows` | `claude-fix-workflows` | Workflow repair and validation | None defined |
| `copilot-beast-mode` | `claude-beast-mode` | Continuous execution mode | None defined |

See `.github/agents/*.agent.md` `tools:` frontmatter for the authoritative Copilot per-agent tool list. The parallel `.claude/agents/*.md` files omit `tools:` — Claude Code inherits all tools by default. `cicd-lint lint-docs` (`lint-agent-drift`) enforces that description, argument-hint, and body are verbatim identical across each pair.

### B.6 CI/CD Workflow Catalog

| Workflow | Purpose | Dependencies | Duration | Timeout |
|----------|---------|--------------|----------|---------|
| ci-coverage | Test coverage collection, enforce ≥95%/98% | None | 5-6min | 20min |
| ci-mutation | Mutation testing with gremlins | None | 15-20min | 45min |
| ci-race | Race condition detection | None | 10-15min | 20min |
| ci-benchmark | Performance benchmarking | None | 8-15min | 30min |
| ci-quality | Linting and code quality | None | 3-5min | 15min |
| ci-sast | Static security analysis | None | 5-10min | 20min |
| ci-dast | Dynamic security testing | PostgreSQL | 10-20min | 30min |
| ci-e2e | End-to-end integration tests | Docker Compose | 20-40min | 60min |
| ci-load | Load testing with Gatling | Docker Compose | 15-30min | 45min |
| ci-gitleaks | Secret detection | None | 2-3min | 10min |
| release | Automated release workflows | ci-* passing | 5-10min | 30min |

### B.7 Reusable Action Catalog

All reusable actions live in `.github/actions/`. Each action is a composite action with a `README.md` and `action.yml`.

| Action | Description |
|--------|-------------|
| `docker-compose-build` | Build Docker images for Compose services |
| `docker-compose-down` | Stop and remove Docker Compose services |
| `docker-compose-logs` | Retrieve logs from Docker Compose services |
| `docker-compose-up` | Start Docker Compose services |
| `docker-compose-verify` | Verify Docker Compose service health |
| `docker-images-pull` | Parallel Docker image pre-fetching (inputs: `images` newline-separated list) |
| `download-cicd` | Download cicd-lint binary from GitHub Releases |
| `fuzz-test` | Run Go fuzz tests with configurable duration |
| `go-setup` | Go toolchain setup with module cache (replaces manual `actions/setup-go` + cache) |
| `golangci-lint` | golangci-lint v2 execution (wraps `golangci-lint run` with `::group::` output) |
| `security-scan-gitleaks` | Secret detection scan (gitleaks) |
| `security-scan-trivy` | Manual Trivy install + CLI (supports `scan-files` mode for multiple target types) |
| `security-scan-trivy2` | Official `aquasecurity/trivy-action` (simpler, SARIF output to GitHub Security tab) |
| `workflow-job-begin` | Job telemetry start (records job start time, emits OTel span) |
| `workflow-job-end` | Job telemetry end (records duration, emits OTel span with status) |

See `.github/actions/` for the authoritative action catalog and per-action `action.yml` inputs/outputs.

### B.8 Linter Rule Reference

| Linter | Purpose | Enabled | Auto-Fix | Exclusions |
|--------|---------|---------|----------|------------|
| errcheck | Unchecked errors | ✅ | ❌ | Test helpers |
| govet | Suspicious code | ✅ | ❌ | None |
| staticcheck | Static analysis | ✅ | ❌ | Generated code |
| wsl_v5 | Whitespace linting | ✅ | ✅ | None |
| godot | Comment periods | ✅ | ✅ | None |
| gosec | Security issues | ✅ | ❌ | Justified cases |

See `.golangci.yml` for the authoritative linter configuration with all 30+ active linters.
