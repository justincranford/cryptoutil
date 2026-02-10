# cryptoutil Constitution

## I. Product Delivery Requirements

cryptoutil MUST deliver four Products (9 total services: 8 product services + 1 demo service) that are independently or jointly deployable.

| Product | Services | Standalone | United |
|---------|----------|------------|--------|
| P1: JOSE | 1 service (JWK Authority) | ✅ | ✅ |
| P2: Identity | 5 services (AuthZ, IdP, RS, RP, SPA) | ✅ | ✅ |
| P3: KMS | 1 service (Key Management Service) | ✅ | ✅ |
| P4: CA | 1 service (Certificate Authority) | ✅ | ✅ |
| Demo: Cipher | 1 service (InstantMessenger) | ✅ | ✅ |

**See**: [architecture.instructions.md](/.github/instructions/02-01.architecture.instructions.md) for complete service catalog and port allocations

### Standalone Mode Requirements

Each product MUST:

- Start independently without other products
- Have working Docker Compose deployments
- Pass all tests in isolation (unit, integration, fuzz, bench, E2E)
- Support SQLite (dev) and PostgreSQL (prod)
- Support configuration via environment variables, CLI parameters, and YAML files
- Default to dev mode when started without configuration

### United Mode Requirements

All four products MUST:

- Deploy together via single Docker Compose without port conflicts
- Share telemetry infrastructure (reusable Docker Compose)
- Support optional inter-product federation via configuration
- Share reusable crypto implementation (NOT external dependencies)
- Pass all isolated AND federated E2E test suites

### Architecture Clarity

Clear separation between infrastructure and products:

- **Infrastructure** (`internal/infra/*`): Reusable building blocks
- **Products** (`internal/product/*`): Deployable services

## II. Cryptographic Compliance

### CGO Ban - ABSOLUTE REQUIREMENT

**CGO_ENABLED=0 is MANDATORY** except for race detector (Go toolchain limitation).

- NEVER use dependencies requiring CGO (e.g., `github.com/mattn/go-sqlite3`)
- ALWAYS use CGO-free alternatives (e.g., `modernc.org/sqlite`)
- Rationale: Maximum portability, static linking, cross-compilation

**See**: [golang.instructions.md](/.github/instructions/03-03.golang.instructions.md) for CGO enforcement patterns

### FIPS 140-3 Compliance

All cryptographic operations MUST use NIST FIPS 140-3 approved algorithms. FIPS mode is ALWAYS enabled.

**Approved**: RSA ≥2048, AES ≥128, EC NIST curves, EdDSA, PBKDF2-HMAC-SHA256/384/512, SHA-256/384/512
**BANNED**: MD5, SHA-1, bcrypt, scrypt, Argon2, 3DES, DES, RSA <2048

Algorithm agility is MANDATORY: all crypto operations MUST support configurable algorithms with FIPS-approved defaults.

**See**: [cryptography.instructions.md](/.github/instructions/02-07.cryptography.instructions.md) for algorithm requirements

### Data Protection

All secret data (passwords, keys) or sensitive data (PII) MUST be encrypted or hashed at rest.

**Hash Registry Selection**:

- **Searchable, no decryption** (PII): Deterministic hash (PBKDF2 with pepper)
- **Searchable, needs decryption** (PII): Deterministic cipher (AES-GCM convergent encryption)
- **Non-searchable, no decryption** (passwords): Non-deterministic hash (PBKDF2 with random salt + pepper)
- **Non-searchable, needs decryption** (keys): Non-deterministic cipher (AES-GCM)

**Secret Management**: ALWAYS use Docker/Kubernetes secrets. NEVER use environment variables.

**See**: [hashes.instructions.md](/.github/instructions/02-08.hashes.instructions.md) for complete hash registry architecture

### Certificate Validation

Full cert chain validation, MinVersion TLS 1.3+, NEVER InsecureSkipVerify. mTLS MUST implement BOTH CRLDP and OCSP for revocation checking.

**See**: [pki.instructions.md](/.github/instructions/02-09.pki.instructions.md) for CA/Browser Forum compliance

## III. Service Architecture Requirements

### Dual HTTPS Endpoints - MANDATORY

ALL services MUST implement TWO separate HTTPS servers:

**Private Endpoint** (Admin Server):

- Purpose: Administration, health checks, graceful shutdown
- Bind: ALWAYS `127.0.0.1:9090` (NEVER configurable, NEVER exposed)
- Endpoints: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/shutdown`

**Public Endpoint** (Public Server):

- Purpose: Business APIs, browser UIs, external client access
- Bind: Configurable (container default: `0.0.0.0`, test/dev default: `127.0.0.1`)
- Port Ranges: Service-specific (8080-8089 KMS, 8180-8189 Identity, etc.)
- Paths: `/service/**` (headless clients) vs `/browser/**` (browser clients)

**See**: [https-ports.instructions.md](/.github/instructions/02-03.https-ports.instructions.md) for complete binding patterns

### Container Support - MANDATORY

ALL services MUST support running as containers (preferred for production and E2E testing).

**See**: [docker.instructions.md](/.github/instructions/04-02.docker.instructions.md) for Docker patterns

### Service Federation - MANDATORY

Services MUST support configurable federation for cross-service communication (NEVER hardcoded URLs).

**Federation Patterns**:

- **Service Discovery**: Config file, Docker Compose service names, Kubernetes DNS
- **Graceful Degradation**: Circuit breakers, fallback modes, retry strategies
- **Cross-Service Auth**: mTLS (preferred) or OAuth 2.1 client credentials

**See**: [architecture.instructions.md](/.github/instructions/02-01.architecture.instructions.md) for federation requirements

## IV. Testing Requirements

### Test Concurrency - MANDATORY

**ALWAYS use concurrent test execution** (NEVER `-p=1` or `-parallel=1`). ALWAYS use `-shuffle=on`.

**Requirements**:

- ALL tests MUST use `t.Parallel()` in test functions and sub-tests
- Test data MUST be isolated (UUIDv7 for uniqueness, port 0 for dynamic allocation)
- Real dependencies preferred over mocks (PostgreSQL test containers, real crypto, real HTTP servers)
- Race detector MUST keep probabilistic execution enabled

**Rationale**: Reveals production concurrency bugs, validates thread safety, ensures faster test execution.

**See**: [testing.instructions.md](/.github/instructions/03-02.testing.instructions.md) for complete testing patterns

### Coverage and Quality Gates

**Coverage Targets** (NO EXCEPTIONS):

- Production code: ≥95%
- Infrastructure/utility code: ≥98%
- main() functions: 0% acceptable if internalMain() ≥95%

**Mutation Testing**: ≥85% Phase 4, ≥98% Phase 5+ gremlins score per package

**Test Execution Time**:

- Unit test packages: <15 seconds per package
- Full unit test suite: <180 seconds (3 minutes) total
- Probabilistic execution MANDATORY for packages approaching 15s limit

**See**: [testing.instructions.md](/.github/instructions/03-02.testing.instructions.md) for timing targets and evidence requirements

## V. Code Quality Requirements

### Continuous Work Mandate - ABSOLUTE ENFORCEMENT

**STOP CONDITIONS (ONLY 2)**:

1. Token usage ≥ 990,000 (NOT 90k - ACTUAL 990,000!)
2. User types "STOP" or "HALT" explicitly

**WORK CONTINUOUSLY**: Execute tool → Execute next tool → Repeat WITHOUT STOPPING. NO pauses, NO status updates mid-session, NO stopping after completing tasks.

**IF TASK COMPLETE**: Immediately start next task from todo list. IF TODO EMPTY: Create new tasks and execute.

**See**: [beast-mode.instructions.md](/.github/instructions/01-02.beast-mode.instructions.md) for enforcement details

### Linting and Formatting

ALWAYS fix linting/formatting errors - NO EXCEPTIONS (production, tests, docs, configs, workflows - ALL must pass).

- NEVER use `//nolint:` except for documented linter bugs
- ALWAYS use UTF-8 without BOM for ALL text files
- File size limits: 300 (soft), 400 (medium), 500 (hard → refactor REQUIRED)
- Pre-commit hooks: ALWAYS fix all errors

**See**: [linting.instructions.md](/.github/instructions/03-07.linting.instructions.md) for golangci-lint configuration

### Evidence-Based Completion

No task is complete without objective, verifiable evidence:

- **Code**: `go build ./...` clean, `golangci-lint run` clean, coverage ≥95%/98%
- **Tests**: All tests passing, no skips without tracking, mutation ≥85%/98%
- **Git**: Conventional commits, clean working tree, changes align with task

Quality gates are MANDATORY - task NOT complete until all checks pass.

**See**: [evidence-based.instructions.md](/.github/instructions/06-01.evidence-based.instructions.md) for completion criteria

## VI. Development Workflow

### Spec Kit Iteration Lifecycle - MANDATORY

Every iteration MUST follow this sequence:

```
1. /speckit.constitution  → Review/update principles (first iteration only)
2. /speckit.specify       → Define/update requirements (spec.md)
3. /speckit.clarify       → Resolve ALL ambiguities (MANDATORY)
4. /speckit.plan          → Technical implementation plan
5. /speckit.tasks         → Generate task breakdown
6. /speckit.analyze       → Coverage check (MANDATORY before implement)
7. /speckit.implement     → Execute implementation
8. /speckit.checklist     → Validate completion (MANDATORY after implement)
```

**CRITICAL**: Steps 3, 6, and 8 are MANDATORY (NOT optional).

**See**: [speckit.instructions.md](/.github/instructions/01-03.speckit.instructions.md) for complete workflow

### Pre-Implementation Gates

Before running `/speckit.implement`:

- [ ] All `[NEEDS CLARIFICATION]` markers resolved in spec.md
- [ ] `/speckit.clarify` executed if spec was created/modified
- [ ] `/speckit.analyze` executed after `/speckit.tasks`
- [ ] All requirements have corresponding tasks
- [ ] No orphan tasks without requirement traceability

### Post-Implementation Gates

Before marking iteration complete:

- [ ] `go build ./...`, `go test ./...`, `golangci-lint run` all pass
- [ ] `/speckit.checklist` executed and all items verified
- [ ] Coverage targets maintained (95% production, 98% infrastructure/utility)
- [ ] Mutation score ≥85%/98% per package
- [ ] All spec.md status markers accurate

**See**: [speckit.instructions.md](/.github/instructions/01-03.speckit.instructions.md) for gate requirements

## VII. Service Template Requirements

### Template Extraction - Phase 6 MANDATORY

Phase 6 MUST extract reusable service template from proven implementations (KMS, JOSE, Identity).

**Template Components**:

- Dual HTTPS servers (public + admin)
- Health checks (`/livez`, `/readyz`, `/shutdown`)
- Graceful shutdown
- Telemetry integration (OTLP)
- Middleware pipeline
- Configuration management (YAML + Docker secrets)

**Template Parameterization**: Constructor injection for handlers, middleware, configuration (business logic separated from infrastructure).

**See**: [service-template.instructions.md](/.github/instructions/02-02.service-template.instructions.md) for complete template requirements

### Migration Priority - MANDATORY

**Decision Source**: CLARIFY-QUIZME-01 Q1

1. **cipher-im FIRST** (Phase 7): Implement using template, validate reusability
2. **JOSE and CA NEXT** (Phases 8-9): jose-ja → pki-ca (sequential, drive refinements)
3. **Identity services LAST** (Phases 10-14): authz → idp → rs → rp → spa
4. **sm-kms NEVER**: Reference implementation remains on current code

**See**: [service-template.instructions.md](/.github/instructions/02-02.service-template.instructions.md) for migration strategy

## VIII. Governance and Standards

### Decision Authority

- **Technical decisions**: Follow copilot instructions in `.github/instructions/`
- **Architectural decisions**: Document in ADRs, follow Standard Go Project Layout
- **Compliance decisions**: CA/Browser Forum Baseline Requirements, RFC 5280, FIPS 140-3, NIST SP 800-57

### Documentation Standards

- PROGRESS.md (in `specs/NNN-cryptoutil/`) is authoritative status source for spec kit iterations
- Keep docs in 2 main files: README.md (main), docs/README.md (deep dive)
- NEVER create separate documentation files for scripts or tools

### Terminology Standards - RFC 2119

**RFC 2119 Keywords**:

- **MUST** / **REQUIRED** / **MANDATORY** / **CRITICAL**: Absolute requirement (all 4 are synonyms)
- **MUST NOT** / **SHALL NOT**: Absolute prohibition
- **SHOULD** / **RECOMMENDED**: Strong recommendation (exceptions require justification)
- **MAY** / **OPTIONAL**: Truly optional (implementer's choice)

**User Intent Clarification**: MUST, REQUIRED, MANDATORY, and CRITICAL are intentionally treated as complete synonyms - all indicate absolute, non-negotiable requirements.

**See**: [terminology.instructions.md](/.github/instructions/01-01.terminology.instructions.md) for RFC 2119 compliance

## IX. Amendment Process

This constitution may be amended only by:

1. **Unanimous consent** of all maintainers for Sections I-VII (core principles)
2. **Majority consent** for Sections VIII-IX (governance)
3. **Automatic updates** for version references following documented update process

**Amendment Procedure**: Proposal → Review (48-hour minimum) → Discussion → Approval → Documentation → Communication

**Ratified**: 2025-12-01 | **Latest amendments**: 2025-12-22
