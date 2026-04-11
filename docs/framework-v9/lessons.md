# Lessons — Framework v9: Quality & Consistency

*This file is maintained by the implementation-execution agent. Each section is filled in after
the corresponding phase completes its quality gates.*

---

## Phase 1: Dockerfile & Deployment Fixes

*(To be filled during Phase 1 execution)*

## Phase 2: Config Key Naming Migration

*(To be filled during Phase 2 execution)*

## Phase 3: Linter Configuration

*(To be filled during Phase 3 execution)*

## Phase 4: Test Quality

*(To be filled during Phase 4 execution)*

## Phase 5: Low-Priority Improvements

*(To be filled during Phase 5 execution)*

## Phase 6: Dockerfile Template Enforcement

*(To be filled during Phase 6 execution)*

## Phase 7: Config Standardization

*(To be filled during Phase 7 execution)*

## Phase 8: Enforcement Linters

*(To be filled during Phase 8 execution)*

## Phase 9: Knowledge Propagation

*(To be filled during Phase 9 execution)*

---

## Pre-Implementation Research Findings

The following findings were discovered during the Phase 0 deep analysis that created
`docs/deployment-templates.md`. They are recorded here as input for the execution agent.

### Finding 1: Three Divergent Dockerfile Patterns

**Issue**: 10 PS-ID Dockerfiles use 3 fundamentally different patterns, making it impossible
to reason about or enforce consistency.

- **Pattern A** (5 services): 4-stage but with curl, GOMODCACHE/GOCACHE in final, USER commented out
- **Pattern B** (3 services): 3-stage (no runtime-deps), adduser-based nonroot, CMD with config path
- **Pattern C** (1 service — sm-im): 2-stage (no validation), no BuildKit caches, user 1000:1000

**Root cause**: Each service was implemented independently with no template enforcement.
Only 2 of 73 fitness linters validate Dockerfile content.

**Lesson**: Small per-rule sublinters are insufficient when there's no structural template
validation. Need a "shape" linter that validates the overall Dockerfile architecture,
not just individual properties.

### Finding 2: Copy-Paste Contamination

**Issue**: skeleton-template is contaminated with jose-ja artifacts at every level:
- Dockerfile header: "JOSE Authority Server"
- Dockerfile username: `jose`
- Dockerfile paths: `/etc/jose/`, `jose.yml`
- Standalone config header: "JOSE Authority Server configuration"
- Deployment common config header: "JOSE Common Configuration"
- Deployment common config otlp-service: `jose-e2e`

**Root cause**: skeleton-template was created by copying jose-ja files with incomplete
search-and-replace. No linter validates that file content identity matches the PS-ID.

**Lesson**: Add a `config_header_identity` fitness linter that validates every config and
Dockerfile header contains the correct PS-ID display name, not a copy-paste source.

### Finding 3: identity-spa COPY Bug

**Issue**: identity-spa Dockerfile builder builds `./cmd/identity-spa` into `/app/identity-spa`,
but the COPY instruction copies `/app/cryptoutil` into the final stage.

**Root cause**: identity-spa was likely copy-pasted from the suite Dockerfile without
updating the binary name.

**Lesson**: Add a `dockerfile_binary_name` fitness linter that validates the COPY source
path matches the PS-ID binary path (`/app/{PS-ID}`).

### Finding 4: Config Key Naming Split

**Issue**: 7 of 10 PS-IDs use snake_case config keys. Only jose-ja, pki-ca, skeleton-template
use kebab-case (correct per ENG-HANDBOOK §13.2).

**Root cause**: Early services (sm-kms, sm-im) used snake_case before the kebab-case
convention was established. Identity services followed sm-kms pattern.

**Lesson**: Enforce config key naming with a `config_key_naming` fitness linter.
The snake_case → kebab-case migration must coordinate config files + Go struct tags + tests.

### Finding 5: Config Overlay Duplication

**Issue**: jose-ja and skeleton-template deployment instance configs duplicate shared
settings (security-headers, rate-limiting) that should be in the common config only.

**Root cause**: No clear definition of what goes in common vs instance configs.
`docs/deployment-templates.md` now defines this boundary.

**Lesson**: Instance configs should be minimal — only cors-origins, otlp-service,
otlp-hostname, and database-url (for SQLite). Everything else goes in common.
