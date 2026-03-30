# Parameterization Opportunities

**Status**: Reference — Ideas Backlog
**Created**: 2026-04-14
**Purpose**: Exhaustive catalogue of opportunities to apply 10×–100× more rigid parameterization
across the cryptoutil repository. Based on deep analysis of `docs/ARCHITECTURE.md` (§1–15,
Appendices A–C), `internal/apps/tools/cicd_lint/lint_fitness/registry/registry.go`, all 18
instruction files, all 14 skills, `deployments/`, `configs/`, and `docs/framework-v7/target-structure.md`.
**Scope**: All 5 products, 10 PS-IDs, 2 infra tools, 1 suite, 1 framework.

---

## Executive Summary

Eighteen ranked opportunities to implement (Part A) plus two deferred items (Part B). Each Part A
item moves one or more currently prose-described or manually-maintained invariants into a machine
that validates and rejects deviations automatically.

1. **[#01] Machine-Readable Entity Registry Schema** — replace the Go-struct registry with a canonical
   YAML schema that becomes the single data source for every downstream artifact in the repo.
2. **[#02] Generative Deployment Scaffold Command** — *(Deferred — see Part B. Violates cicd-lint
   no-generation and no-parameter constraints.)*
3. **[#03] @propagate Coverage Completeness Matrix** — formalize which ARCHITECTURE.md section must
   propagate to which instruction/agent file; lint-docs reports missing blocks, not only content drift.
4. **[#04] Port Formula Codification** — store base port per PS-ID in the entity registry; compute all
   host/container port appearances as `base + tier_offset`; fitness linter validates the math cross-file.
5. **[#05] Parameterized Secret Value Generation** — a single script generates all 420 (14 × 3 tiers × 10
   services) secret instances from one parameterized schema; the `secret-content` linter validates every instance.
6. **[#06] Config Overlay Generation Template** — replace 50 hand-crafted `{PS-ID}-app-{variant}.yml` overlay
   files with outputs of a single Go template, regenerated on demand from the entity registry.
7. **[#07] Per-PS-ID Migration Range Reservation** — extend migration numbering from a vague "2001+ domain"
   to per-PS-ID explicit bands (sm-kms 2001–2099, sm-im 2101–2199, …); fitness linter enforces bands.
8. **[#08] Dockerfile Label Derivation** — all Dockerfile LABELs (`org.opencontainers.*`, ENTRYPOINT, EXPOSE)
   derived deterministically from the entity registry, eliminating per-service manual authorship.
9. **[#09] CLI Subcommand Completeness Matrix** — 10 PS-IDs × 8 subcommands = 80 required handlers declared
   as a completeness table; fitness linter verifies every entry exists in code.
10. **[#10] OTLP / Compose Naming Formula** — `{ps-id}-{db}-{N}` codified as a bounded tuple (ps-id ∈
    registry, db ∈ {sqlite,postgresql}, N ∈ {1,2}); fitness linter rejects any name outside the valid set.
11. **[#11] SQL Identifier Derivation Function** — a single `PSIDToSQLIdent(psid)` registry function computes
    all four SQL-safe identifiers (`{PS_ID}_database`, `{PS_ID}_database_user`, host `{ps-id}-postgres`, …).
12. **[#12] Secret File Content Schema** — 14 secret value regex/template patterns stored in one machine-readable
    schema used by both the generation tool and the `secret-content` fitness linter.
13. **[#13] API Path Parameter Registry** — each PS-ID declares its resource names as typed tuples
    `(path_type, api_version, resource)`; a fitness linter validates OpenAPI spec paths match the declarations.
14. **[#14] Instruction File Slot Reservation Table** — *(Deferred — see Part B. Not planned at this time.)*
15. **[#15] Fitness Sub-Linter Category Registry** — the 57 sub-linters expressed as a structured registry
    (category, name, scope, enforcement level) enabling automated docs generation and gap detection.
16. **[#16] Compose Service Instance Naming Schema** — `{ps-id}-{db}-{N}` stored as an explicit enumeration
    of all valid (ps-id, db, N) tuples; invalid names are errors, missing names are warnings.
17. **[#17] Health Path Completeness Matrix** — 6 mandatory paths × 10 services = 60 required entries
    declared as a completeness table; fitness linter verifies all 60 are reachable in running services.
18. **[#18] Test File Suffix Registry** — 5 suffixes with machine-enforced structural content rules
    (e.g., `_bench_test.go` MUST contain `Benchmark*`, MUST NOT contain non-bench `Test*`).
19. **[#19] Import Alias Formula Enforcement** — `cryptoutil{Package}` / `{vendor}{Package}` alias patterns
    validated automatically by a fitness linter over every import block in every `.go` file.
20. **[#20] X.509 Certificate Profile Schema** — the 25 PKI profiles in `configs/pki-ca/profiles/` expressed
    as typed instances of a single strict YAML schema with all required fields validated on every change.

---

## Part A: To Implement

---

### #01 — Machine-Readable Entity Registry Schema

**Impact**: 100× — eliminates the root cause of every consistency gap.

**Current state**: The registry lives in
`internal/apps/tools/cicd_lint/lint_fitness/registry/registry.go` as Go structs populated with
`cryptoutilSharedMagic.*` constants. The structs define `PSID`, `Product`, `Service`,
`DisplayName`, `InternalAppsDir`, and `MagicFile`. Every fitness linter that needs to iterate
services imports this Go package at compile time.

This approach works today (10 services), but it has critical limitations:

- Non-Go tooling (GitHub Actions expressions, future Terraform modules) cannot read the
  Go registry without a separate extraction step.
- Adding a new PS-ID requires: (a) adding a magic constant, (b) updating the Go registry struct,
  (c) manually creating every downstream artifact — there is no generative feedback loop.
- The registry contains only structural fields; derived fields (base port, SQL identifier,
  display name title-case, OTLP prefix, compose service name) must be re-derived independently
  in every consumer.

**Opportunity**: Introduce a canonical `registry.yaml` (location TBD — see quizme-v1.md
Question 1) with the full entity schema capturing suite, products, and product-services:

```yaml
suite:
  id: cryptoutil
  display_name: "Cryptoutil"
  cmd_dir: cryptoutil/

products:
  - id: sm
    display_name: "Secrets Manager"
    internal_apps_dir: sm/
    cmd_dir: sm/
  - id: jose
    display_name: "JOSE"
    internal_apps_dir: jose/
    cmd_dir: jose/
  - id: pki
    display_name: "PKI"
    internal_apps_dir: pki/
    cmd_dir: pki/
  - id: identity
    display_name: "Identity"
    internal_apps_dir: identity/
    cmd_dir: identity/
  - id: skeleton
    display_name: "Skeleton"
    internal_apps_dir: skeleton/
    cmd_dir: skeleton/

product_services:
  - ps_id: sm-kms
    product: sm
    service: kms
    display_name: "Secrets Manager Key Management"
    internal_apps_dir: sm-kms/
    magic_file: magic_sm.go
    base_port: 8000           # ← New: drives all port derivations
    pg_host_port: 54320       # ← New: PostgreSQL host port assignment
    migration_range_start: 2001  # ← New: explicit domain migration band
    migration_range_end: 2099
    api_resources:            # ← New: from api/sm-kms/openapi_spec_paths.yaml
      - path: /elastickey
      - path: /elastickey/{elasticKeyID}
      - path: /elastickeys
      - path: /elastickey/{elasticKeyID}/materialkey
      - path: /elastickey/{elasticKeyID}/materialkey/{materialKeyID}
      - path: /elastickey/{elasticKeyID}/materialkey/{materialKeyID}/revoke
      - path: /elastickey/{elasticKeyID}/materialkeys
      - path: /materialkeys
      - path: /elastickey/{elasticKeyID}/generate
      - path: /elastickey/{elasticKeyID}/encrypt
      - path: /elastickey/{elasticKeyID}/decrypt
      - path: /elastickey/{elasticKeyID}/sign
      - path: /elastickey/{elasticKeyID}/verify
      - path: /elastickey/{elasticKeyID}/import
  - ps_id: sm-im
    product: sm
    service: im
    display_name: "Secrets Manager Instant Messenger"
    internal_apps_dir: sm-im/
    magic_file: magic_sm_im.go
    base_port: 8100
    pg_host_port: 54321
    migration_range_start: 2101
    migration_range_end: 2199
    api_resources:            # ← from api/sm-im/openapi_spec.yaml
      - path: /messages/tx
      - path: /messages/rx
      - path: /messages/{id}
  - ps_id: jose-ja
    product: jose
    service: ja
    display_name: "JOSE JWK Authority"
    internal_apps_dir: jose-ja/
    magic_file: magic_jose.go
    base_port: 8200
    pg_host_port: 54322
    migration_range_start: 2201
    migration_range_end: 2299
    api_resources:            # ← from api/jose-ja/openapi_spec.yaml
      - path: /elastic-jwks
      - path: /elastic-jwks/{kid}
      - path: /elastic-jwks/{kid}/materials
      - path: /elastic-jwks/{kid}/materials/active
      - path: /elastic-jwks/{kid}/rotate
      - path: /jwk/generate
      - path: /jwk/{kid}
      - path: /jwk
      - path: /jwks
      - path: /jws/sign
      - path: /jws/verify
      - path: /jwe/encrypt
      - path: /jwe/decrypt
      - path: /jwt/create
      - path: /jwt/verify
  # ... remaining 7 PS-IDs follow same pattern
```

The Go registry structs become **generated code** — `go generate` reads `registry.yaml` and
emits `registry.go`. All other downstream consumers (GitHub Actions, Dockerfile templates)
read `registry.yaml` directly.

**Quantified scope**: Replaces manual authorship in 10 PS-ID Dockerfiles, 10 PS-ID compose files,
50 config overlay files, 140 secret files (14 × 10), 30 port rows in docs tables, 57 fitness
linters (all iterate `AllProductServices()`), and 18 instruction file tables.

**Pre-requisites**: None — can be introduced incrementally. Start by adding derived fields to the
YAML while keeping the Go struct in sync; then flip to code generation.

**Fitness enforcement**: New `entity-registry-schema` linter validates `registry.yaml` against a
JSON Schema on every commit.

---

> **[#02] Generative Deployment Scaffold Command** — Deferred. See **Part B** below for rationale.

---

### #03 — @propagate Coverage Completeness Matrix

**Impact**: 50× — prevents knowledge rot and compliance gaps across all 18 instruction files.

**Current state**: The `@propagate` marker system (Section 13.4) is the repository's most powerful
existing parameterization mechanism. When an ARCHITECTURE.md block changes, the `lint-docs`
command detects drift in any `@source` block across instruction/agent files. Currently 38 unique
chunk IDs are propagated across 17 instruction files.

However, the system only validates **existing** propagations — it does not detect when a new
ARCHITECTURE.md section has no propagation target at all. The Section-Level Mapping table
(Section 13.4.4) documents which sections map to which instruction files, but this table is prose,
not machine-checked. A new section can be added to ARCHITECTURE.md and silently never reach any
instruction file.

**Opportunity**: Introduce a `required-propagations.yaml` manifest that explicitly declares:

```yaml
required_propagations:
  - section: "6.9"
    description: "Authentication key principles"
    chunk_id: key-principles
    targets:
      - .github/instructions/02-06.authn.instructions.md
  - section: "14.7"
    description: "Infrastructure blocker escalation rule"
    chunk_id: infrastructure-blocker-escalation
    targets:
      - .github/instructions/06-01.evidence-based.instructions.md
      - .github/instructions/01-02.beast-mode.instructions.md
  # ... all 38 current chunks plus new ones
```

The `lint-docs validate-propagation` command then checks not only drift but also
**coverage completeness**: every chunk in the manifest MUST exist as a `@propagate` block in
ARCHITECTURE.md, and every chunk MUST have a corresponding `@source` block in each declared target.

Additionally, add a **section coverage report**: for every top-level H2 and H3 section in
ARCHITECTURE.md, report whether zero propagation chunks exist. Sections with zero chunks are
candidates for the manifest.

**Quantified scope**: 15 H2 sections × avg. 3 H3 sub-sections = ~45 section blocks. Current
coverage: 38 chunks across 17 files. Estimated gap: 15–20 uncovered sections that have
instruction-worthy content.

**Pre-requisites**: None — purely additive to existing `lint-docs` infrastructure.

**Fitness enforcement**: `lint-docs check-coverage` exits 1 if any manifest entry lacks a
`@propagate` block, or if any declared target lacks the corresponding `@source` block.

---

### #04 — Port Formula Codification

**Impact**: 30× — eliminates the only remaining manual source of port conflicts across
the 10-service, 3-tier port matrix.

**Current state**: The port design principle is documented in Section 3.4.1:
- SERVICE tier: base port (e.g., 8000 for sm-kms)
- PRODUCT tier: base + 10000 (e.g., 18000)
- SUITE tier: base + 20000 (e.g., 28000)

The PostgreSQL host port assignments (54320–54329) and the container ports (8080/9090) are also
documented. However, the formula is only prose-described; the actual port values in every
compose.yml and config overlay file are hand-entered. The `ValidatePorts` lens checks ranges but
does not verify that the actual port value equals `base_port + tier_offset`.

**Current port inventory**:
- 10 services × 3 tiers × 2 ports (public+admin) = 60 host port bindings in compose files
- 10 PostgreSQL host port bindings
- 10 base ports in docs tables
- 10 config overlay common files specifying `bind-public-port`

Total: ~90 port values that should be derivable from 10 base ports + 3 formulas.

**Opportunity**: Store `base_port` and `pg_host_port` per PS-ID in `registry.yaml` (see #01).
Generate a `ports.go` file in the scaffold package that exposes:

```go
func PublicPort(psid string, tier Tier) (uint16, error)
func AdminPort(psid string, tier Tier) (uint16, error)
func PostgresPort(psid string) (uint16, error)
```

Extend the `ValidatePorts` fitness linter to call these functions and verify that every port
declaration in every compose file equals the computed value — not just falls within a range.

Also generate a `ports-reference.md` table under `docs/` from the registry, replacing the
hand-maintained table in ARCHITECTURE.md Section 3.4.

**Quantified scope**: 90 manually-maintained port values reduced to 10 source values + 3
computed formulas + automated cross-file validation.

**Pre-requisites**: Opportunity #01 (base_port in registry.yaml).

**Fitness enforcement**: Enhanced `ValidatePorts` with formula-level verification; new
`port-formula-consistency` check.

---

### #05 — Parameterized Secret Value Generation

**Impact**: 30× — eliminates the most error-prone part of onboarding: generating 140 correct secrets.

**Current state**: Section 12.3.3 defines a complete secret value format table for all 14 secret
types at all 3 deployment tiers. The format for each type is:
- Static prefix: `{PREFIX}-`, `{PREFIX_UNDERSCORE}_`
- Purpose string: e.g., `hash-pepper-v3`, `unseal-key-N-of-5`, `database_user`
- Optional random suffix: base64url-encoded 32 random bytes for password-like secrets
- Static suffix: none for identifiers

No automated generation tooling exists. Secrets are authored manually following the format
conventions documented in Section 12.3.3, but the format schemas are not machine-readable —
they live only in ARCHITECTURE.md prose and in ad-hoc regex inside the linters.

**Current gap**: No single command generates all secrets for a new deployment tier; no validator
checks that existing secrets match their declared format. The `secret-content` fitness linter
validates some patterns; the `unseal-secret-content` linter validates unseal key uniqueness. But
the format schemas live only in ARCHITECTURE.md prose, not in machine-readable form.

**Opportunity**: Define a `secret-schemas.yaml` manifest:

```yaml
secrets:
  - name: hash-pepper-v3.secret
    tiers: [service, product, suite]
    format: "{PREFIX}-hash-pepper-v3-{base64url-random-32-bytes}"
    prefix_separator: "-"
    random_bytes: 32
    encoding: base64url_no_padding
  - name: unseal-{N}of5.secret
    tiers: [service]
    n_range: [1, 5]
    format: "{PREFIX}-unseal-key-{N}-of-5-{base64-random-32-bytes}"
    prefix_separator: "-"
    random_bytes: 32
    encoding: base64_standard
    uniqueness: required  # Each shard value MUST be different
  - name: postgres-username.secret
    tiers: [service, product, suite]
    format: "{PREFIX_UNDERSCORE}_database_user"
    prefix_separator: "_"
    random_bytes: 0      # Static — no random component
```

Use this schema to:
1. **Validate** all existing secrets in CI: `cicd-lint lint-fitness` runs `secret-content` linter against schemas.
2. **Document** the format table in ARCHITECTURE.md via `@propagate` from the YAML.
3. **Guide** manual secret creation: developers follow the schema when onboarding a new tier or PS-ID.

**Quantified scope**: 14 format schemas × 10 PS-IDs × 3 tiers = 420 secret instances all
derivable from 14 schemas + entity registry. Replaces 420 hand-crafted secret files.

**Pre-requisites**: Opportunity #01 for prefix derivation.

**Fitness enforcement**: `secret-content` and `unseal-secret-content` linters rewritten to
validate against `secret-schemas.yaml` rather than hardcoded regex.

---

### #06 — Config Overlay Template Validation

**Impact**: 30× — validates 50 deployment config overlay files against canonical templates.

**Current state**: Every service requires exactly 5 config overlay files in
`deployments/{PS-ID}/config/`:
- `{PS-ID}-app-common.yml` — bind addresses, TLS settings, network
- `{PS-ID}-app-sqlite-1.yml` — SQLite instance 1
- `{PS-ID}-app-sqlite-2.yml` — SQLite instance 2
- `{PS-ID}-app-postgresql-1.yml` — PostgreSQL instance 1
- `{PS-ID}-app-postgresql-2.yml` — PostgreSQL instance 2

All 5 files follow a strict schema (`bind-public-protocol`, `bind-public-address`,
`bind-public-port`, `bind-private-address`, `bind-private-port`, `tls-public-dns-names`,
`database-url`, `otlp-endpoint`, `otlp-service`, etc.). The values differ by only the PS-ID,
the database variant, the instance number, and the port binding.

Currently 50 files exist (10 services × 5 variants). Each was hand-crafted, some with subtle
structural differences (capitalization, key ordering, optional OTLP fields). The `ValidateSchema`
linter enforces required keys but not value derivations.

**Opportunity**: Two Go templates per variant:

```
internal/apps/tools/cicd_lint/scaffold/templates/
├── config-common.yml.tmpl
├── config-sqlite.yml.tmpl
└── config-postgresql.yml.tmpl
```

Each template references entity registry values (PS-ID, base port, OTLP service name pattern)
and instance parameters (variant: sqlite/postgresql, N: 1/2).

A `config-overlay-freshness` fitness check verifies that each overlay file matches what the
template defines for its PS-ID and variant — catching drift when template patterns change
without corresponding overlay updates. When onboarding a new service, the developer copies the
template and substitutes the PS-ID manually; the linter validates correctness.

**Quantified scope**: 50 files × avg. 15 lines = 750 YAML lines covered by 3 templates × 15
lines = 45 lines + registry data. New service onboarding: 5 files validated by `cicd-lint lint-fitness`.

**Pre-requisites**: Opportunities #01 (base_port), #04 (port formula), #10 (OTLP naming).

**Fitness enforcement**: New `config-overlay-freshness` linter; existing `ValidateSchema` linter
continues to provide independent content validation.

---

### #07 — Per-PS-ID Migration Range Reservation

**Impact**: 20× — prevents migration number collisions when multiple services evolve concurrently.

**Current state**: Section 5.2 defines migration ranges:
- Framework template: 1001–1999 (shared: sessions, barrier, realms, tenants, pending users)
- Domain: 2001+ (application-specific, "never conflicts with template")

The "never conflicts" claim holds only if all 10 services use separate, non-overlapping sub-ranges
within 2001+. Currently there is no explicit per-service sub-range declaration. The
`migration-range-compliance` fitness linter checks that domain migrations are ≥ 2001 and template
migrations are in 1001–1999, but does not check for cross-service collisions.

A collision would only manifest at E2E time when two services with overlapping migration numbers
share a PostgreSQL database — which is exactly the scenario under the SUITE deployment tier,
where all 10 services connect to the same shared database.

**Opportunity**: Each PS-ID entry in `registry.yaml` declares an explicit reservation:

```yaml
product_services:
  - ps_id: sm-kms
    migration_range_start: 2001
    migration_range_end: 2099
  - ps_id: sm-im
    migration_range_start: 2101
    migration_range_end: 2199
  - ps_id: jose-ja
    migration_range_start: 2201
    migration_range_end: 2299
  # ... etc.
```

99-migration-per-service bands (10 × 100 = 1000; fits comfortably in a 2001–2999 domain space
with room for 9 new services before the range needs extending).

The `migration-range-compliance` linter is enhanced to:
1. Load reservations from `registry.yaml`
2. Verify each service's migration files fall within its declared band
3. Verify no two services have overlapping bands (cross-service check)

Additionally, the `migration_comment_headers` linter verifies that the top-of-file comment
in each migration SQL file includes the service's display name (enabling quick identification of
which service owns a migration in a shared database state).

**Quantified scope**: 10 services, each with a guaranteed collision-free band. Prevents a class
of database bugs that currently only surface at E2E test time. Also documents expected migration
count per service as a code archaeology signal.

**Pre-requisites**: Opportunity #01 (registry.yaml).

**Fitness enforcement**: Enhanced `migration-range-compliance` linter with cross-service collision
detection; enhanced `migration-numbering` linter reading bands from registry.

---

### #08 — Dockerfile Label Derivation

**Impact**: 20× — eliminates 10–11 hand-maintained Dockerfile sections with divergent labels.

**Current state**: Section 12.2.1 defines the Dockerfile parameterization by tier:
- `image.title` LABEL: `{SUITE}-{PS-ID}` at service tier
- Binary: always `./cmd/{SUITE}` (the suite binary)
- EXPOSE: 8080 at service tier
- ENTRYPOINT: `["/app/{SUITE}", "{PS-ID}", "start"]` at service tier

The LABELs must follow OCI Image Specification (`org.opencontainers.image.*`). Each Dockerfile
currently has these labels hand-written. The `dockerfile-labels` fitness linter validates label
presence but does not verify that the label values match registry-derived values.

Labels affected per service:
- `org.opencontainers.image.title` — must equal `{SUITE}-{PS-ID}`
- `org.opencontainers.image.description` — must equal `{DisplayName}`
- `org.opencontainers.image.source` — must equal the canonical repository URL + ps-id path
- `org.opencontainers.image.vendor` — must equal `{SUITE}`
- `EXPOSE` value — must equal container public port (always 8080 at service tier)
- `ENTRYPOINT` — must equal `["/app/{SUITE}", "{PS-ID}", "start"]`
- `HEALTHCHECK` — must reference `127.0.0.1:8080` (container port, always same)

**Opportunity**: Regenerate Dockerfiles (or just the label+entrypoint sections) from a shared
template in the scaffold system. The `dockerfile-labels` fitness linter is extended to read
expected values from the entity registry and fail if any label value diverges.

Also add a `dockerfile-entrypoint-formula` fitness check that verifies `ENTRYPOINT` follows
the `["/app/{SUITE}", "{PS-ID}", "start"]` formula rather than a hardcoded string.

**Quantified scope**: 10 service Dockerfiles + 5 product Dockerfiles + 1 suite Dockerfile =
16 files; each has 5–8 parameterized lines. All 80–128 lines become computable from 2 registry
fields per PS-ID.

**Pre-requisites**: Opportunity #01 (display_name in registry.yaml).

**Fitness enforcement**: Enhanced `dockerfile-labels` linter with registry-derived expected values.

---

### #09 — CLI Subcommand Completeness Matrix

**Impact**: 15× — guarantees every PS-ID exposes every required subcommand with zero manual tracking.

**Current state**: Section 9.1 defines the CLI hierarchy:
- PS-ID subcommands: `server`, `client`, `health`, `livez`, `readyz`, `shutdown`, `init`,
  `compose`, `e2e` (documented; exact list varies by service implementation maturity)
- Product commands: subset of PS-ID subcommands delegating to all child services
- Suite commands: product commands delegating to all products

Currently there is no machine-enforced completeness check — a PS-ID can exist in the registry
with only 3 of the 8 required subcommands, and no fitness linter will catch the gap. The
`cmd-entry-whitelist` linter checks that only valid cmd entry points exist; the `cmd-main-pattern`
linter checks internal structure. Neither checks that all required subcommands are registered.

**Opportunity**: Define the required subcommand set per role in `registry.yaml`:

```yaml
subcommand_requirements:
  product_service:
    required: [server, health, livez, readyz, shutdown]
    optional: [client, init, compose, e2e]
  product:
    required: [health, livez, readyz, shutdown]
    optional: [server, client, init, compose, e2e]
  suite:
    required: [health, livez, readyz, shutdown]
    optional: [server, client, init, compose, e2e]
```

New `subcommand-completeness` fitness linter traverses the Go AST of each
`internal/apps/{PS-ID}/{PS-ID}.go` file and verifies that all required subcommand handlers
are registered. Reports missing handlers as errors with the exact function signature expected.

**Quantified scope**: 10 PS-IDs × 5 required subcommands = 50 required handlers. Currently tracked
implicitly; fitness linter would surface the ~15 gaps in less-mature services.

**Pre-requisites**: Opportunity #01 (subcommand_requirements in registry.yaml).

**Fitness enforcement**: New `subcommand-completeness` linter.

---

### #10 — OTLP / Compose Naming Formula

**Impact**: 15× — eliminates ~50 string literals that must match a precise formula.

**Current state**: Section 9.10 defines OTLP service names follow `{ps-id}-{db}-{N}` where
`db ∈ {sqlite, postgresql}` and `N ∈ {1, 2}`. Compose service names follow the same formula.
These names appear in:
- 50 config overlay files (one `otlp-service` value per file)
- 10+ compose.yml files (service name keys)
- 30+ fitness linter test fixtures

The `otlp-service-name-pattern` fitness linter validates the `{ps-id}-{db}-{N}` format via
regex. The `compose-service-names` linter validates compose service name format. However,
neither linter cross-references the entity registry to verify that the ps-id component is a
*valid registered PS-ID* — a typo like `sm-km-sqlite-1` (hyphen instead of dash) would pass
the regex but reference a non-existent service.

**Opportunity**: Generate the OTLP service name and compose service name in the entity registry:

```go
func OTLPServiceName(psid, db string, n int) string
func ComposeServiceName(psid, db string, n int) string
func ValidOTLPServiceNames(psid string) []string // returns all 4 valid names
```

Enhance `otlp-service-name-pattern` and `compose-service-names` to:
1. Extract the ps-id component from the name
2. Verify it exists in `AllProductServices()`
3. Verify db ∈ {sqlite, postgresql}
4. Verify N ∈ {1, 2}

Also generate `standalone-config-otlp-names` validation from the registry rather than a
hardcoded list of valid names.

**Quantified scope**: 50 config overlay files + 10 compose service name sections = 60 instances
all validated against a formula with entity registry cross-reference, rather than regex alone.

**Pre-requisites**: Opportunity #01 (entities in registry.yaml).

**Fitness enforcement**: Enhanced `otlp-service-name-pattern` and `compose-service-names` linters
with entity registry validation; new `standalone-config-otlp-names` derivation from registry.

---

### #11 — SQL Identifier Derivation Function

**Impact**: 10× — provides a single canonical function for all database identifier computations.

**Current state**: PostgreSQL identifiers follow a clear pattern but are computed independently in
multiple places:
- `{PS_ID}_database` — PostgreSQL database name (e.g., `sm_kms_database`)
- `{PS_ID}_database_user` — database username
- `{PS_ID}_database_pass-{random}` — database password pattern
- `{PS_ID}_database_user` — in postgres-url.secret
- `{ps-id}-postgres` — Docker Compose PostgreSQL service name (hyphens, not underscores)
- Port: `{pg_host_port}` — from the port assignment table

The conversion rule `{PS-ID}` → `{PS_ID}` (replace `-` with `_`) is simple but critical. Any
deviation causes authentication failures in PostgreSQL (which is case-insensitive but
hyphen-intolerant in unquoted identifiers).

Currently: the `ValidateKebabCase` linter checks YAML keys; the `compose-db-naming` fitness linter
checks Compose db service names. But no single function is the authoritative source for all
derived identifiers.

**Opportunity**: Add to the entity registry package:

```go
// PSIDToSQLID converts hyphenated PS-ID to underscore SQL identifier component.
// "sm-kms" → "sm_kms"
func PSIDToSQLID(psid string) string

// DatabaseName returns the PostgreSQL database name for a PS-ID.
// "sm-kms" → "sm_kms_database"
func DatabaseName(psid string) string

// DatabaseUser returns the PostgreSQL username for a PS-ID.
// "sm-kms" → "sm_kms_database_user"
func DatabaseUser(psid string) string

// PostgresServiceName returns the Compose PostgreSQL service name for a PS-ID.
// "sm-kms" → "sm-kms-postgres"
func PostgresServiceName(psid string) string
```

Both the secret generation tooling and the deployment fitness linters import this package.
Eliminates all hard-coded `strings.ReplaceAll(psid, "-", "_")` calls scattered across the codebase.

**Quantified scope**: 4 derived identifiers × 10 PS-IDs = 40 values scattered across 50+ config
overlay files, 10 compose files, and 14 secret value formats — all derived from a single function.

**Pre-requisites**: Opportunity #01 is optional but recommended; this function can stand alone.

**Fitness enforcement**: Existing `compose-db-naming` and `secret-content` linters updated to call
these functions rather than regex-match independently.

---

### #12 — Secret File Content Schema

**Impact**: 10× — closes the validation gap between "file exists" and "file contains correct content".

**Current state**: The `secret-naming` fitness linter validates that secret filenames follow the
`{purpose}.secret` or `{purpose}.secret.never` naming convention. The `secret-content` fitness
linter validates some content patterns. The `unseal-secret-content` linter validates unseal key
patterns and uniqueness.

However, the format schemas for all 14 secret types (Section 12.3.3, "Secret File Format" table)
live only in ARCHITECTURE.md prose and in ad-hoc regex inside the linters. There is no single
machine-readable source of truth for:
- Which secret types use base64url vs. base64 standard encoding
- Which have a random component vs. are fully static
- Which are identity-mapped `{PREFIX_UNDERSCORE}_...` vs. hyphen-prefixed `{PREFIX}-...`
- Which require uniqueness across shards (unseal keys)

**Opportunity**: Introduce a `secret-schemas.yaml` referenced in §05, but focused here on the
validation side. The `secret-content` linter reads this schema and generates its validation
rules at startup. Adding or changing a secret type requires only a schema update, not a linter
code change.

Schema per secret:
```yaml
- name: browser-username.secret
  tiers: [service]  # not:never at product/suite
  pattern: ^{PREFIX}-browser-user$
  random_bytes: 0
  description: "Browser auth username, static"
```

**Quantified scope**: 14 schema entries replace disparate regex in 2 fitness linters. Reduces
the cost of adding a new secret type from "modify linter code + tests" to "append YAML entry".

**Pre-requisites**: None — can be introduced as a standalone YAML file.

**Fitness enforcement**: Enhanced `secret-content` linter with schema-driven validation rules.

---

### #13 — API Path Parameter Registry

**Impact**: 10× — ensures OpenAPI specs are structurally consistent across all 10 PS-IDs.

**Current state**: All services expose APIs under exactly two path prefixes:
- `/browser/api/v1/{resource}` — browser clients only
- `/service/api/v1/{resource}` — headless clients only

Section 8.2 requires plural noun resource names in kebab-case (e.g., `/keys`, `/elastic-jwks`,
`/messages`). SAME OpenAPI spec at both paths; only middleware/auth differs. The `require-api-dir`
fitness linter verifies the `api/{PS-ID}/` directory exists. But no linter checks that the paths
in OpenAPI specs follow the `/{prefix}/api/v{N}/{resource}` formula or that resource names are
correctly plural and kebab-case.

**Opportunity**: Each PS-ID entry in `registry.yaml` declares its API resources (using actual
paths from existing OpenAPI specs):

```yaml
product_services:
  - ps_id: sm-kms
    api_version: 1
    api_resources:              # from api/sm-kms/openapi_spec_paths.yaml
      - path: /elastickey
      - path: /elastickey/{elasticKeyID}
      - path: /elastickeys
      - path: /elastickey/{elasticKeyID}/materialkey
      - path: /elastickey/{elasticKeyID}/materialkeys
      - path: /materialkeys
      - path: /elastickey/{elasticKeyID}/generate
      - path: /elastickey/{elasticKeyID}/encrypt
      - path: /elastickey/{elasticKeyID}/decrypt
      - path: /elastickey/{elasticKeyID}/sign
      - path: /elastickey/{elasticKeyID}/verify
  - ps_id: jose-ja
    api_version: 1
    api_resources:              # from api/jose-ja/openapi_spec.yaml
      - path: /elastic-jwks
      - path: /elastic-jwks/{kid}
      - path: /elastic-jwks/{kid}/materials
      - path: /elastic-jwks/{kid}/rotate
      - path: /jwk/generate
      - path: /jwk/{kid}
      - path: /jwk
      - path: /jwks
      - path: /jws/sign
      - path: /jws/verify
      - path: /jwe/encrypt
      - path: /jwe/decrypt
```

A new `api-path-formula` fitness linter reads the PS-ID's OpenAPI spec
and verifies:
1. Every path matches `/browser/api/v{N}/{resource}` or `/service/api/v{N}/{resource}`
2. N matches the declared `api_version`
3. The resource segment is plural-noun kebab-case
4. Both `/browser/` and `/service/` prefixes exist for every resource

**Quantified scope**: 10 services × ~5 resource endpoints × 2 path prefixes = ~100 API paths
validated against a formula rather than inspected manually during code review.

**Pre-requisites**: Opportunity #01 (api_resources in registry.yaml).

**Fitness enforcement**: New `api-path-formula` linter; enhanced `require-api-dir` linter.

---

> **[#14] Instruction File Slot Reservation Table** — Deferred. See **Part B** below for rationale.

---

### #15 — Fitness Sub-Linter Category Registry

**Impact**: 5× — enables automated documentation and gap detection for the 57-linter catalog.

**Current state**: Section 9.11 documents all 57 fitness sub-linters in 7 categories
(Security, Architecture, Deployment & Config, Code Quality, Testing, Service Framework,
Database & Migrations). This catalog lives entirely in ARCHITECTURE.md prose. The actual
linter implementations live in 57 directories under `lint_fitness/`. There is no machine-readable
link between the documentation table and the implementation files.

Consequences:
- Adding a new linter requires both updating the implementation AND updating the ARCHITECTURE.md
  table; neither step enforces the other.
- The `entity-registry-completeness` linter and `lint_fitness.go` registration are the
  authoritative source of registered linters — but neither has a declared expected count.
- There is no check that all 57 documented linters are actually registered, or that all
  registered linters are documented.

**Opportunity**: Introduce a `lint-fitness-registry.yaml`:

```yaml
categories:
  security:
    linters:
      - name: crypto-rand
        dir: crypto_rand/
        scope: "*.go"
        enforcement: error
        description: "Enforce crypto/rand over math/rand"
      - name: non-fips-algorithms
        dir: non_fips_algorithms/
        ...
```

The `lint_fitness.go` registration step reads this YAML to:
1. Verify every declared linter directory exists
2. Verify every registered linter appears in the YAML
3. Generate the catalog table in ARCHITECTURE.md via `@propagate`

**Quantified scope**: 57 linters × 2 attributes (doc ↔ code) = 114 cross-references that are
currently manually maintained, reduced to one authoritative YAML.

**Pre-requisites**: None — purely additive.

**Fitness enforcement**: New `lint-fitness-registry-completeness` check; generated catalog table
validated via `@propagate`.

---

### #16 — Compose Service Instance Naming Schema

**Impact**: 5× — closes the validation gap between formula description and formula enforcement.

**Current state**: The `compose-service-names` fitness linter validates that Docker Compose
service names in service-tier compose files follow certain patterns. The `otlp-service-name-pattern`
linter validates OTLP service names. But neither linter maintains an explicit enumeration of all
valid (ps-id, db, N) combinations.

As a result, a compose.yml with service name `sm-kms-mariadb-1` (invalid db type) would pass
the current regex (`.*-(sqlite|postgresql)-[0-9]+`) if the regex were slightly wrong, but there
is no cross-reference to the entity registry to confirm `sm-kms` is a real PS-ID.

**Opportunity**: Generate the complete valid-name set from the entity registry:

```go
ValidComposeServiceNames() []string
// returns: ["sm-kms-sqlite-1", "sm-kms-sqlite-2", "sm-kms-postgresql-1", ...]
```

`compose-service-names` linter changes from regex-match to set-membership check. A name in a
compose file either appears in `ValidComposeServiceNames()` or is an invalid name (error).

For the PostgreSQL service name (`{ps-id}-postgres`), add `PostgresServiceName(psid)` (see #11)
and validate compose PostgreSQL service names against the formula.

**Quantified scope**: 10 PS-IDs × 4 names each = 40 valid compose service names; 10 PostgreSQL
service names. All 50 validated by set membership rather than regex.

**Pre-requisites**: Opportunity #10 (OTLP/compose naming formula).

**Fitness enforcement**: Enhanced `compose-service-names` and `otlp-service-name-pattern` linters.

---

### #17 — Health Path Completeness Matrix

**Impact**: 5× — guarantees every service exposes every required health endpoint.

**Current state**: Section 5.5 defines 6 health endpoints per service:
- Public: `/browser/api/v1/health`, `/service/api/v1/health`
- Admin: `/admin/api/v1/livez`, `/admin/api/v1/readyz`, `/admin/api/v1/shutdown`
- (Plus the Fiber router health)

The `health-endpoint-presence` fitness linter checks that health endpoint handlers are
registered. The `service-contract-compliance` linter verifies the contract test is present.
But no linter checks that all 6 paths are present in the OpenAPI spec or that compose
`healthcheck` configurations reference the correct admin path (`/admin/api/v1/livez`).

**Opportunity**: Generate a health path completeness check from the 6 required paths:

```go
RequiredHealthPaths() []HealthPath
// returns 6 entries: {prefix, version, path, port, role}
```

New `health-path-completeness` linter reads each service's `openapi_spec_paths.yaml` and
compose healthcheck config, verifying the 6 required paths are present.

Also add a cross-tier check: compose healthcheck in the product-tier compose must reference
the service's admin path, not a product-level path.

**Quantified scope**: 10 services × 6 paths = 60 required health path appearances. Currently
partially checked (3 of 6 have fitness coverage); this completes the matrix.

**Pre-requisites**: Opportunity #13 (API path registry) is complementary but not required.

**Fitness enforcement**: New `health-path-completeness` linter; enhanced `service-contract-compliance`
linter checking OpenAPI spec completeness.

---

### #18 — Test File Suffix Registry

**Impact**: 3× — enforces structural content rules for each of the 5 test file types.

**Current state**: Section 10.1 defines 5 test file suffixes:
- `_test.go` — unit tests (`func Test*`)
- `_bench_test.go` — benchmarks (`func Benchmark*`)
- `_fuzz_test.go` — fuzz tests (`func Fuzz*`)
- `_property_test.go` — property tests (`func Test*` with `//go:build !fuzz`)
- `_integration_test.go` — integration tests (with `//go:build integration` tag)

The `test-patterns` fitness linter enforces some test patterns (parallel, UUID usage). But no
linter enforces suffix-specific structural rules, e.g.:
- `_bench_test.go` MUST contain at least one `func Benchmark*` and MUST NOT contain `func Test*`
- `_fuzz_test.go` MUST contain at least one `func Fuzz*` and MUST have a minimum fuzz time seed
- `_integration_test.go` MUST have `//go:build integration` tag at the top

**Opportunity**: Introduce a `test-file-suffix-rules.yaml`:

```yaml
suffixes:
  _bench_test.go:
    must_contain_pattern: "^func Benchmark"
    must_not_contain_pattern: "^func Test[^M]"   # Test* except TestMain
    required_build_tag: ~
  _fuzz_test.go:
    must_contain_pattern: "^func Fuzz"
    required_comment_pattern: "fuzztime=15s"
    required_build_tag: ~
  _integration_test.go:
    required_build_tag: "integration"
```

New `test-file-suffix-structure` fitness linter reads these rules and enforces them.

**Quantified scope**: 5 suffix rules × ~20 files each = ~100 files validated for structural
content compliance, catching common errors (benchmark in non-bench file, missing build tags).

**Pre-requisites**: None.

**Fitness enforcement**: New `test-file-suffix-structure` linter.

---

### #19 — Import Alias Formula Enforcement

**Impact**: 3× — makes the `cryptoutil{Package}` / `{vendor}{Package}` convention machine-enforced.

**Current state**: Section 11.1.3 defines import aliases:
- Internal packages: `cryptoutil{Package}` (e.g., `cryptoutilMagic "cryptoutil/internal/shared/magic"`)
- External packages: `{vendor}{Package}` (e.g., `googleUuid "github.com/google/uuid"`)

The `importas` golangci-lint plugin is configured to enforce some aliases. But:
- The list of required aliases in `.golangci.yml` is manually maintained and does not
  automatically include new packages as they are added.
- There is no guarantee that every `cryptoutil/internal/*` package has a declared alias.
- The vendor prefix convention (`google`, `fiber`, `gorm`, etc.) is documented as prose with
  no mechanical enforcement beyond the manually-maintained `importas` list.

**Opportunity**: Generate the `importas` rules from the entity registry and a package alias map:

```yaml
internal_alias_prefix: cryptoutil
packages:
  - path: cryptoutil/internal/shared/magic
    alias_suffix: SharedMagic
  - path: cryptoutil/internal/apps/framework/service
    alias_suffix: FrameworkService
  # ... all internal packages

external_aliases:
  - path: github.com/google/uuid
    vendor: google
    package: Uuid
  # ... all external packages
```

The `importas` golangci-lint config is maintained separately, but a new `import-alias-formula`
fitness linter reads the alias map YAML and validates every import block in the codebase directly,
ensuring actual usage matches the declared convention.

**Quantified scope**: ~50 internal packages + ~20 external packages = 70 import alias rules.
Currently only ~20 are declared in `.golangci.yml`; the rest are convention without enforcement.

**Pre-requisites**: None.

**Fitness enforcement**: New `import-alias-formula` fitness linter reads the alias map YAML and
validates import blocks across all `.go` files; `importas` golangci-lint config updated manually
when adding new packages.

---

### #20 — X.509 Certificate Profile Schema

**Impact**: 2× — validates 25 certificate profiles against a strict typed schema.

**Current state**: `configs/pki-ca/profiles/` contains 25 X.509 certificate profile files
(e.g., `tls-server.yaml`, `root-ca.yaml`, `intermediate-ca.yaml`, client cert profiles, etc.).
Each file defines a certificate profile with fields like:
- Key algorithm, key size
- Validity period (days)
- Key usage extensions
- Extended key usage
- SAN types allowed

The `ValidateTemplatePattern` linter checks some profile structure but does not validate against a
typed JSON Schema. A misconfigured profile (e.g., `validity_days: 500` for a subscriber cert
that must be ≤398 days) currently only fails at runtime when a certificate is actually issued.

**Opportunity**: Define `pki-ca-profile-schema.json` (JSON Schema draft-07) capturing:
- Required fields: `key_algorithm`, `key_size`, `validity_days`, `key_usage`, `extended_key_usage`
- Type constraints: `validity_days` is integer, `key_algorithm` is enum(`RSA`, `ECDSA`, `EdDSA`)
- Business constraints: subscriber certs `validity_days ≤ 398`, intermediate `≤ 3650`, root `≤ 9125`
- Key size constraints: RSA ≥ 2048, ECDSA P-256/384/521 only
- Compliance: SHA-1 and MD5 MUST NOT appear in `signature_algorithm`

A pre-commit hook and CI step validate all 25 profiles against this schema on every change.

**Quantified scope**: 25 profile files × avg. 15 fields = 375 field values validated against
typed constraints, catching FIPS or CA/Browser Forum violations at commit time rather than at
certificate issuance time.

**Pre-requisites**: None — JSON Schema is self-contained.

**Fitness enforcement**: New `pki-profile-schema-validation` step in `cicd-lint lint-deployments`
or as a standalone pre-commit hook using `check-jsonschema`.

---

## Part B: Deferred — Not Planned

The following items are **not planned for implementation** at this time.

---

### #02 — Generative Deployment Scaffold Command

**Status**: Deferred.

**Reason**: cicd-lint is exclusively for linting, formatting, and operational cleanup. It NEVER
generates content and NEVER accepts customization parameters beyond subcommand names. The proposed
`scaffold --ps-id=NEW-SVC` command violates both constraints:

- `--ps-id=NEW-SVC` is a customization parameter (violates subcommands-only rule).
- Generating Dockerfiles, compose files, config overlays, and secrets is content generation
  (violates no-generation rule).

**Alternative**: New services are onboarded manually following the patterns documented in
ARCHITECTURE.md. The `entity-registry-completeness`, `deployment-dir-completeness`, and
`config-overlay-freshness` fitness linters detect gaps and missing artifacts after manual
creation. The `skeleton-template` service provides a copy-and-modify reference.

---

### #14 — Instruction File Slot Reservation Table

**Status**: Deferred — not sure if valuable enough to implement.

**Reason**: The `NN-NN.name.instructions.md` numbering convention is a lightweight implicit
scheme. A formal reservation table adds governance overhead without a clear benefit given the
current small number of files (18). May reconsider if the number grows significantly or if
numbering collisions occur in practice.

**Alternative**: The existing `copilot-instructions.md` table serves as an informal slot registry.
Conflicts are rare enough that manual coordination suffices.

---

## Impact Ranking Summary

| Rank | ID | Title | Impact | Primary Benefit |
|------|----|-------|--------|-----------------|
| 1 | #01 | Machine-Readable Entity Registry Schema | 100× | Foundation for all other opportunities |
| 2 | #03 | @propagate Coverage Completeness Matrix | 50× | Zero knowledge rot across 18 instruction files |
| 3 | #04 | Port Formula Codification | 30× | 90 port values → 10 base values + 3 formulas |
| 4 | #05 | Parameterized Secret Value Generation | 30× | 420 secret instances validated against 14 schemas |
| 5 | #06 | Config Overlay Freshness Validation | 30× | 50 overlay files validated against templates |
| 6 | #07 | Per-PS-ID Migration Range Reservation | 20× | Cross-service collision prevention |
| 7 | #08 | Dockerfile Label Derivation | 20× | 128 parameterized lines from registry |
| 8 | #09 | CLI Subcommand Completeness Matrix | 15× | 80 required handlers tracked automatically |
| 9 | #10 | OTLP / Compose Naming Formula | 15× | 60 name instances from bounded formula |
| 10 | #11 | SQL Identifier Derivation Function | 10× | 40 SQL identifiers from 1 function |
| 11 | #12 | Secret File Content Schema | 10× | Schema-driven validation replaces ad-hoc regex |
| 12 | #13 | API Path Parameter Registry | 10× | 100 API paths validated against formula |
| 13 | #15 | Fitness Sub-Linter Category Registry | 5× | Doc ↔ code cross-reference for 57 linters |
| 14 | #16 | Compose Service Instance Naming Schema | 5× | 50 names validated by set membership |
| 15 | #17 | Health Path Completeness Matrix | 5× | 60 health path appearances verified |
| 16 | #18 | Test File Suffix Registry | 3× | Structural content rules per suffix type |
| 17 | #19 | Import Alias Formula Enforcement | 3× | 70 import aliases machine-validated |
| 18 | #20 | X.509 Certificate Profile Schema | 2× | 375 profile field values type-validated |

---

## Dependency Graph

```
#01 Registry YAML
├── → #04 Port Formula
│       └── → #06 Config Overlay Template
├── → #05 Secret Generation
│       └── → #12 Secret Schema
├── → #07 Migration Ranges
├── → #08 Dockerfile Labels
├── → #09 CLI Completeness
├── → #10 OTLP / Compose Naming
│       └── → #16 Compose Instance Naming
├── → #11 SQL Identifiers
└── → #13 API Path Registry
        └── → #17 Health Path Matrix

#03 @propagate Coverage ── standalone ──→  (no external deps)
#15 Fitness Registry ────── standalone ──→  (no external deps)
#18 Test Suffix Rules ───── standalone ──→  (no external deps)
#19 Import Alias Gen ────── standalone ──→  (no external deps)
#20 PKI Profile Schema ──── standalone ──→  (no external deps)
```

**Recommended implementation order**:

```
#01 → #03 → #04 → #11 → #12 → #05 → #07 → #08 → #10
→ #16 → #06 → #09 → #13 → #17 → #15 → #18 → #19 → #20
```

---

## Currently Parameterized (Baseline)

The following patterns are already parameterized and enforced — these are **not** opportunities
but are listed to establish the baseline from which the 20 ideas above extend:

| Pattern | Formula | Enforced By |
|---------|---------|-------------|
| PS-ID format | `{PRODUCT}-{SERVICE}` | `service-structure` fitness linter |
| Config overlay naming | `{PS-ID}-app-{variant}.yml` | `configs-naming` fitness linter |
| OTLP name format | `{ps-id}-{db}-N` (regex) | `otlp-service-name-pattern` linter |
| Compose service name | `{ps-id}-{db}-N` (regex) | `compose-service-names` linter |
| Secret filename | `{purpose}.secret` (no tier prefix) | `secret-naming` fitness linter |
| Unseal key value uniqueness | `N` shards MUST differ | `unseal-secret-content` linter |
| Delegation pattern | SUITE→PRODUCT→SERVICE | `checkDelegationPattern()` in lint-deployments |
| Admin bind address | MUST be `127.0.0.1:9090` | `ValidateAdmin` + `admin-bind-address` linter |
| Import alias prefix | `cryptoutil*` for internal | `importas` golangci-lint plugin (partial) |
| Migration range split | template 1001-1999, domain 2001+ | `migration-range-compliance` linter |
| @propagate drift | verbatim content match | `lint-docs validate-propagation` |
| Port range by tier | SERVICE 8k, PRODUCT 18k, SUITE 28k | `ValidatePorts` in lint-deployments |
| Health endpoint presence | handlers registered | `health-endpoint-presence` linter |
| API directory existence | `api/{PS-ID}/` exists | `require-api-dir` fitness linter |
| Entity completeness | all PS-IDs in registry | `entity-registry-completeness` linter |
| Dockerfile labels | labels present | `dockerfile-labels` fitness linter |
| Test parallel() | t.Parallel() in Test* | `parallel-tests` fitness linter |
| SQLite CGO-free | modernc.org/sqlite only | `cgo-free-sqlite` fitness linter |
