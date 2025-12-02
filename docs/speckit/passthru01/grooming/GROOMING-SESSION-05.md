# Grooming Session 05: Infrastructure and Quality Assurance

## Overview

- **Focus Area**: Infrastructure components, CI/CD, testing patterns, code quality standards
- **Related Spec Section**: Infrastructure I1-I16, Quality Gates, Testing Requirements
- **Prerequisites**: Sessions 01-04 completed, understanding of DevOps and CI/CD

---

## Questions

### Q1: What is the purpose of I1: Configuration component?

A) Network configuration only
B) Config files, env vars, secrets, feature flags, validation
C) Database configuration only
D) Telemetry configuration only

**Answer**: B
**Explanation**: I1: Configuration handles config files, environment variables (for non-secrets), secrets, feature flags, and validation.

---

### Q2: Which infrastructure component handles HTTP, HTTPS, and gRPC?

A) I1: Configuration
B) I2: Networking
C) I5: Telemetry
D) I9: Deployment

**Answer**: B
**Explanation**: I2: Networking handles HTTP, HTTPS, gRPC, REST, load balancing, and firewalls.

---

### Q3: What coverage target applies to utility code?

A) ≥75%
B) ≥80%
C) ≥85%
D) ≥95%

**Answer**: D
**Explanation**: Utility code requires ≥95% coverage. Production is ≥80%, infrastructure is ≥85%.

---

### Q4: Which tool is used for Go static analysis?

A) go vet only
B) golangci-lint v2.6.2+
C) staticcheck only
D) gosec only

**Answer**: B
**Explanation**: golangci-lint v2.6.2+ is the required static analysis tool, which includes multiple linters.

---

### Q5: What pre-commit hook enforces file encoding?

A) encoding-check
B) utf8-validator
C) cicd all-enforce-utf8
D) bom-check

**Answer**: C
**Explanation**: The `cicd all-enforce-utf8` command enforces UTF-8 without BOM for all text files.

---

### Q6: What is the soft file size limit?

A) 200 lines
B) 300 lines
C) 400 lines
D) 500 lines

**Answer**: B
**Explanation**: File size limits: 300 (soft), 400 (medium), 500 (hard requiring refactor).

---

### Q7: Which linter requires periods at end of comments?

A) gocritic
B) godot
C) gofumpt
D) revive

**Answer**: B
**Explanation**: godot linter requires periods at the end of comments for proper documentation.

---

### Q8: What is the wsl linter configuration key in golangci-lint v2?

A) wsl
B) wsl_v4
C) wsl_v5
D) whitespace

**Answer**: C
**Explanation**: In golangci-lint v2, wsl configuration uses the `wsl_v5` key.

---

### Q9: Which CI workflow handles linting and formatting?

A) ci-coverage
B) ci-quality
C) ci-security
D) ci-lint

**Answer**: B
**Explanation**: ci-quality handles linting, formatting, and builds (no database services required).

---

### Q10: What services are required for ci-dast workflow?

A) None
B) SQLite only
C) PostgreSQL
D) Full Docker stack

**Answer**: C
**Explanation**: ci-dast requires PostgreSQL for dynamic application security testing.

---

### Q11: Which workflow tool is used for local CI testing?

A) Jenkins
B) GitLab Runner
C) Act (via cmd/workflow)
D) CircleCI local

**Answer**: C
**Explanation**: Act is used via `go run ./cmd/workflow` for local GitHub Actions workflow testing.

---

### Q12: What command runs a quick DAST scan?

A) `go run ./cmd/workflow -workflows=dast`
B) `go run ./cmd/workflow -workflows=dast -inputs="scan_profile=quick"`
C) `nuclei -quick`
D) `zap -quick-scan`

**Answer**: B
**Explanation**: Quick DAST scan uses `-inputs="scan_profile=quick"` (3-5 min) vs full profile (10-15 min).

---

### Q13: What is the purpose of the autoapprove wrapper?

A) Auto-approve pull requests
B) Bypass VS Code safety blockers for loopback commands
C) Auto-approve deployments
D) Approve code changes

**Answer**: B
**Explanation**: autoapprove bypasses VS Code Copilot's hardcoded safety blockers for loopback network commands (127.0.0.1, ::1, localhost).

---

### Q14: Where are autoapprove logs stored?

A) `./logs/autoapprove/`
B) `./test-output/autoapprove/`
C) `/var/log/autoapprove/`
D) `~/.autoapprove/logs/`

**Answer**: B
**Explanation**: Autoapprove creates timestamped directories in `./test-output/autoapprove/`.

---

### Q15: What HTTP tool should be used in Docker healthchecks?

A) curl
B) wget
C) httpie
D) Invoke-WebRequest

**Answer**: B
**Explanation**: wget is available in Alpine containers; curl is not by default. Use wget for healthchecks.

---

### Q16: What is the correct table-driven test pattern requirement?

A) Optional for simple tests
B) Mandatory with t.Parallel()
C) Only for complex tests
D) Deprecated in favor of subtests

**Answer**: B
**Explanation**: Table-driven tests with t.Parallel() are mandatory to reveal concurrency bugs.

---

### Q17: What happens when a parallel test fails?

A) Remove t.Parallel()
B) Skip the test
C) Fix the production bug revealed
D) Run sequentially instead

**Answer**: C
**Explanation**: Failing parallel tests reveal race conditions - fix the bug, don't remove parallelism.

---

### Q18: What is the correct approach for test data values?

A) Hardcode UUIDs for consistency
B) Magic package constants OR runtime UUIDv7
C) Sequential integers
D) Random strings

**Answer**: B
**Explanation**: Use magic package constants OR generate UUIDv7 once and reuse. Never hardcode or generate multiple times expecting same value.

---

### Q19: What port allocation pattern is required for server tests?

A) Hard-coded port numbers
B) Sequential port assignment
C) Dynamic port allocation (port 0)
D) External port registry

**Answer**: C
**Explanation**: Use port 0 and extract actual assigned port for concurrent test execution.

---

### Q20: What is the correct import alias for googleUuid?

A) `uuid`
B) `guuid`
C) `googleUuid`
D) `cryptoutilUuid`

**Answer**: C
**Explanation**: Third-party aliases: googleUuid for google UUID package (not cryptoutil prefix pattern).

---

### Q21: What are the crypto acronyms casing rules?

A) camelCase (Rsa, Aes)
B) lowercase (rsa, aes)
C) ALL CAPS (RSA, AES, ECDSA)
D) Mixed case per context

**Answer**: C
**Explanation**: Crypto acronyms are ALL CAPS: RSA, EC, ECDSA, ECDH, HMAC, AES, JWA, JWK, JWS, JWE, ED25519, PKCS8, PEM, DER.

---

### Q22: Where should magic values be defined?

A) Inline in code
B) `internal/common/magic/magic_*.go`
C) Environment variables
D) Configuration files only

**Answer**: B
**Explanation**: Magic values go in `internal/common/magic/magic_*.go` or identity-specific `magic*.go` files.

---

### Q23: What is the SQLite busy_timeout setting?

A) 5000ms
B) 10000ms
C) 30000ms
D) 60000ms

**Answer**: C
**Explanation**: busy_timeout should be 30000ms (30 seconds) via `PRAGMA busy_timeout = 30000`.

---

### Q24: What journal mode is required for SQLite concurrency?

A) DELETE
B) TRUNCATE
C) WAL
D) MEMORY

**Answer**: C
**Explanation**: WAL (Write-Ahead Logging) mode enables better concurrency with multiple readers.

---

### Q25: What MaxOpenConns is required for SQLite with GORM transactions?

A) 1
B) 5
C) 10
D) 25

**Answer**: B
**Explanation**: GORM transactions need multiple connections. MaxOpenConns=1 causes deadlocks; use 5 for GORM.

---

### Q26: What is the correct GORM annotation for JSON arrays in SQLite?

A) `gorm:"type:json"`
B) `gorm:"serializer:json"`
C) `gorm:"type:jsonb"`
D) `gorm:"type:text"`

**Answer**: B
**Explanation**: Use `serializer:json` for cross-DB compatibility; SQLite lacks native JSON type.

---

### Q27: What nullable UUID type should be used with GORM?

A) `*googleUuid.UUID`
B) `sql.NullString`
C) Custom NullableUUID type
D) `string` with empty check

**Answer**: C
**Explanation**: Pointer UUIDs cause SQLite errors. Use custom NullableUUID implementing sql.Scanner and driver.Valuer.

---

### Q28: What telemetry forwarding architecture is required?

A) Direct app to Grafana
B) App → otel-collector-contrib → Grafana
C) Prometheus scraping only
D) No telemetry forwarding

**Answer**: B
**Explanation**: All telemetry MUST go through otel-collector-contrib sidecar; never bypass it.

---

### Q29: What OTLP port is for gRPC?

A) 4317
B) 4318
C) 9090
D) 13133

**Answer**: A
**Explanation**: OTLP gRPC uses 4317; OTLP HTTP uses 4318; 13133 is collector health check.

---

### Q30: What conventional commit type is for new features?

A) feature
B) feat
C) new
D) add

**Answer**: B
**Explanation**: Conventional commits use `feat` for new features, `fix` for bugs, etc.

---

### Q31: What commit message format is required?

A) Free-form description
B) `<type>[scope]: <description>`
C) JIRA-123: description
D) [FEATURE] description

**Answer**: B
**Explanation**: Conventional Commits format: `<type>[optional scope]: <description>`.

---

### Q32: Which GitHub CLI command shows failed workflow logs?

A) `gh workflow logs --failed`
B) `gh run view <id> --log-failed`
C) `gh actions logs --failures`
D) `gh log --workflow-failed`

**Answer**: B
**Explanation**: `gh run view <run-id> --log-failed` shows logs for failed jobs in a workflow run.

---

### Q33: What artifact retention is recommended for temporary files?

A) 7 days
B) 30 days
C) 1 day
D) 90 days

**Answer**: C
**Explanation**: 1 day for temporary artifacts; 1-30 days for valuable reports.

---

### Q34: What SARIF upload action is used for security findings?

A) `security/upload-sarif`
B) `github/codeql-action/upload-sarif@v3`
C) `actions/upload-sarif`
D) `sarif/upload@v1`

**Answer**: B
**Explanation**: `github/codeql-action/upload-sarif@v3` uploads SARIF to GitHub Security tab.

---

### Q35: What is the purpose of diagnostic timing in workflows?

A) Track build performance
B) Debug slow steps (>10s should include timing)
C) Billing calculation
D) Resource allocation

**Answer**: B
**Explanation**: Steps >10s MUST include timing with START_TIME and DURATION calculations.

---

### Q36: What emojis indicate workflow status?

A) ✓/✗
B) 📋 start, ✅ success, ❌ error
C) [OK]/[FAIL]
D) +/-

**Answer**: B
**Explanation**: Workflow diagnostics use: 📋 start, 📅 timestamps, ⏱️ duration, ✅ success, ❌ error.

---

### Q37: Where should coverage files be placed?

A) Root directory
B) `./test-output/`
C) `/tmp/coverage/`
D) `./coverage/`

**Answer**: B
**Explanation**: Coverage files go in `./test-output/`: `go test -coverprofile=test-output/coverage_pkg.out`.

---

### Q38: What cicd command self-exclusion pattern is required?

A) No self-exclusion needed
B) Every command excludes its own subdirectory
C) Exclude all cicd directories
D) Include all directories equally

**Answer**: B
**Explanation**: EVERY cicd command MUST exclude its own subdirectory; define in `magic_cicd.go`.

---

### Q39: What script preference order is specified?

A) Bash > PowerShell > Go
B) Go > Python > (BANNED: PowerShell/Bash)
C) PowerShell > Bash > Python
D) Python > Go > Bash

**Answer**: B
**Explanation**: Preference: Go > Python. PowerShell and Bash scripts are BANNED for cross-platform.

---

### Q40: What action is used for parallel Docker image pulls?

A) `docker/pull-action`
B) `.github/actions/docker-images-pull`
C) `actions/docker-pull`
D) Manual docker pull commands

**Answer**: B
**Explanation**: Use `.github/actions/docker-images-pull` for parallel image downloads in workflows.

---

### Q41: What is prohibited in Docker Compose files?

A) Volume mounts
B) Absolute paths
C) Environment variables
D) Network definitions

**Answer**: B
**Explanation**: NEVER use absolute paths in compose.yml; use relative paths like `file: ./postgres/secret.secret`.

---

### Q42: How should Docker secrets be accessed?

A) Environment variables
B) File URLs from /run/secrets/
C) Command line arguments directly
D) Config files in image

**Answer**: B
**Explanation**: Use secrets with file:// URLs from `/run/secrets/`, not environment variables.

---

### Q43: What healthcheck command works in Alpine containers?

A) `curl -f http://localhost/health`
B) `wget --no-check-certificate -q -O /dev/null https://127.0.0.1:9090/livez`
C) `http GET localhost/health`
D) `/bin/health-check`

**Answer**: B
**Explanation**: wget is available in Alpine; use wget with --no-check-certificate for HTTPS healthchecks.

---

### Q44: What documentation file is the single source of truth for status?

A) README.md
B) CHANGELOG.md
C) PROJECT-STATUS.md
D) STATUS.md

**Answer**: C
**Explanation**: PROJECT-STATUS.md is the ONLY authoritative source for project status.

---

### Q45: How many main documentation files should exist?

A) One per feature
B) Two main files (README.md, docs/README.md)
C) Unlimited
D) One file only

**Answer**: B
**Explanation**: Keep docs in 2 main files: README.md (main), docs/README.md (deep dive).

---

### Q46: What pre-commit config changes require documentation updates?

A) No documentation required
B) Update docs/pre-commit-hooks.md
C) Update README only
D) Update CHANGELOG only

**Answer**: B
**Explanation**: When modifying `.pre-commit-config.yaml`, `.golangci.yml`, or `cicd.go`, update `docs/pre-commit-hooks.md`.

---

### Q47: What is the approach for fixing many lint issues?

A) Fix one at a time
B) Use multi_replace_string_in_file for batch fixes
C) Suppress all issues
D) Create separate PR per fix

**Answer**: B
**Explanation**: Use multi_replace_string_in_file for efficiency when fixing many similar issues.

---

### Q48: What detect-secrets inline allowlist format is used?

A) `// noqa: secrets`
B) `// pragma: allowlist secret`
C) `// nosec`
D) `// allow-secret`

**Answer**: B
**Explanation**: detect-secrets uses `// pragma: allowlist secret` for inline allowlisting.

---

### Q49: What tool should be used over terminal commands?

A) Terminal is preferred
B) Built-in tools (create_file, read_file, runTests)
C) External scripts
D) IDE features only

**Answer**: B
**Explanation**: ALWAYS use built-in tools over terminal commands: create_file, read_file, runTests, etc.

---

### Q50: What is the token budget target for continuous work?

A) 50% of budget
B) 75% of budget
C) 99% of 1M-token budget
D) No specific target

**Answer**: C
**Explanation**: Target 99% of the 1M-token budget; keep working even if task appears complete.

---

## Session Summary

**Topics Covered**:

- Infrastructure components (I1-I16)
- CI/CD workflows and testing
- Code quality standards and linting
- Database configuration (SQLite, PostgreSQL)
- Telemetry and observability architecture
- Documentation standards
- Git workflow and conventional commits
- Tool preferences and automation

**Completion**: All 5 grooming sessions complete
