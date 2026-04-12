# Quizme v1 — Framework v10: Canonical Template Registry

**Purpose**: Deep analysis of `./configs/` and `./deployments/` revealed 10 scope and design
ambiguities that must be resolved before Phase 1 implementation begins. Each question has
exactly one correct answer for this project. Fill in the `Answer:` field (A, B, C, D, or E).

**Scope**: Answers feed directly into plan.md / tasks.md updates. The execution agent must
read all answers before writing any template files or starting Task 1.4 (product compose).

---

## Question 1: Product Compose `pki-init` Override — Structural Non-Uniformity

**Context**: The `sm` product compose and the `cryptoutil` suite compose each have a
`pki-init:` service override:
- `sm/compose.yml`: `pki-init: { command: ["init", "--domain=sm"], ... }`
- `cryptoutil/compose.yml`: `pki-init: { command: ["init", "--domain=cryptoutil"], ... }`

The other 4 products (`identity`, `jose`, `pki`, `skeleton`) have NO `pki-init` override.
A single `templates/deployments/__PRODUCT__/compose.yml` template must handle both cases.

**Question**: How should the single product compose template handle the structural
difference between products that need a `pki-init` override and those that do not?

**A)** Conditional block: add `__PRODUCT_PKI_INIT_BLOCK__` placeholder that expands to the
full `pki-init:` service override for `sm` (with `--domain=__PRODUCT__`) and expands to an
empty string for `identity`, `jose`, `pki`, `skeleton`. Add boolean `has_pki_init: true` to
`registry.yaml` `products` entries for `sm` only.

**B)** Remove `pki-init` from `sm` product compose — the override is incorrect or redundant
with the `pki-ca` PS-ID compose already defining `pki-init`. All 5 products use the same
template with no `pki-init` block.

**C)** Two product templates: `templates/deployments/__PRODUCT__/compose-with-pki.yml` for `sm`
and `templates/deployments/__PRODUCT__/compose-no-pki.yml` for the other 4. Registry specifies
which template each product uses via a `compose_template: with-pki|no-pki` field.

**D)** All 5 products should have a `pki-init` override. Add `pki-init` to `identity`, `jose`,
`pki`, `skeleton` composes with their correct `--domain=<product>` values. One uniform template
includes `pki-init` for all products.

**E)**

**Answer**:

**Rationale**: This directly controls whether Task 1.4 writes 1 template file or 2, and whether
`registry.yaml` needs a new boolean field. See `deployments/sm/compose.yml` lines 1-200.

---

## Question 2: Product Compose `image:` Explicit References in Service Overrides

**Context**: The `sm` product compose specifies `image: cryptoutil-sm-kms:dev` on each service
override. The other 4 products (`identity`, `jose`, `pki`, `skeleton`) do NOT include `image:`
in their service overrides — they rely on the PS-ID `compose.yml` to supply the image reference.

**Question**: Should the product compose template include `image:` in service overrides?

**A)** Intentional asymmetry — keep `sm` as-is with explicit image references because it has
multiple PS-IDs. The other 4 products correctly omit image references. Template uses a
`__PRODUCT_SERVICE_IMAGE_LINE__` placeholder that expands to `image: __SUITE__-__PS_ID__:__IMAGE_TAG__`
for `sm` and empty string for others. Registry adds `override_image: true` per product.

**B)** Artifact / inconsistency — remove `image:` from `sm` product compose to match the other
4. PS-ID-level `compose.yml` files already set the image; product-level overrides should not
duplicate it. One uniform template with NO `image:` at product level.

**C)** All 5 products should have explicit `image:` in their service overrides for clarity.
Add `image: __SUITE__-__PS_ID__:__IMAGE_TAG__` to `identity`, `jose`, `pki`, `skeleton`
service overrides. One uniform template with `image:` for all.

**D)** All 5 products should NOT have `image:` in service overrides. Remove from `sm` to
match others. The PS-ID compose is the single source of the image reference.

**E)**

**Answer**:

**Rationale**: This determines whether the product compose template uses `image:` lines in
service overrides (uniform) vs. a conditional block (non-uniform). See `deployments/sm/compose.yml`.

---

## Question 3: Product Compose PostgreSQL Secrets — `sm` Has 4, Others Have 1

**Context**: The `sm` product compose `secrets:` section includes four PostgreSQL-related secrets:
`postgres-url.secret`, `postgres-username.secret`, `postgres-password.secret`,
`postgres-database.secret`. The `identity`, `jose`, `pki`, and `skeleton` product composes
include only `postgres-url.secret`.

**Question**: Should the product compose template secrets section be uniform across all products?

**A)** Intentional — `sm` needs individual credentials in addition to the connection URL; the
other 4 products only need the URL. Template uses `__PRODUCT_POSTGRES_SECRETS_BLOCK__` with
per-product expansion. Registry adds `postgres_secrets: url_only | all_credentials` per product.

**B)** Artifact — `sm` should only expose `postgres-url.secret` like the other 4. The individual
credential secrets (`username`, `password`, `database`) are redundant when `postgres-url` contains
the full connection string. Remove them from `sm`. One uniform template with 1 postgres secret.

**C)** All 5 products should include all 4 postgres secrets for maximum operational flexibility.
Add `postgres-username.secret`, `postgres-password.secret`, `postgres-database.secret` to the
`identity`, `jose`, `pki`, `skeleton` composes. One uniform template with 4 postgres secrets.

**D)** Individual credential secrets (`username`, `password`, `database`) belong at PS-ID level,
not product level. Move them from `deployments/sm/` to `deployments/sm-kms/` and
`deployments/sm-im/`. Product template keeps only `postgres-url.secret`.

**E)**

**Answer**:

**Rationale**: Determines the secrets section shape in the product compose template and whether
registry.yaml needs per-product metadata. See `deployments/sm/compose.yml` secrets section.

---

## Question 4: Template Comparison Mode — Standalone Configs With Domain-Specific Extensions

**Context**: The v9 `template_drift` package used `comparePrefix` for standalone config files:
the template specifies the required beginning of the file; actual files may append domain-specific
content at the end. Example: `configs/sm-kms/sm-kms.yml` ends with `database-url:` after the
template's last line (`log-level: "INFO"`). The `pki-ca` standalone config will similarly have
profile-directory settings at the end not present in the baseline template.

The v10 plan specifies `CompareExpectedFS` (exact comparison after normalization) but does not
define how to handle files that legitimately extend the template with domain-specific additions.

**Question**: How should the v10 template compliance engine handle standalone configs that have
allowed domain-specific additions after the required template prefix?

**A)** Prefix comparison for standalone configs — template defines the required beginning; any
content after the template's last line is allowed. Encode via a `# END_TEMPLATE` sentinel comment
at the end of each `configs/__PS_ID__/__PS_ID__.yml` template; content after this marker in actual
files is ignored during comparison.

**B)** Exact match everywhere — extend the `standalone-config.yml` template with all PS-ID-specific
additions as conditional blocks (e.g., `__PS_ID_EXTRA_CONFIG__` placeholder that expands to the
pki-ca-specific lines for `pki-ca` and empty string for others). All actual file content is
accounted for in the template.

**C)** Exact match for most templates; prefix match for files under `configs/` only. No sentinel
comment needed — the engine uses the directory path to select comparison mode: `deployments/` → exact,
`configs/` → prefix (only header/baseline validated, trailing lines ignored).

**D)** No comparison for standalone configs — they are developer-maintained files that legitimately
vary too much. Remove `configs/__PS_ID__/__PS_ID__.yml` from the template tree entirely. Only
`deployments/` files are template-linted.

**E)**

**Answer**:

**Rationale**: This is a Phase 2 design decision that affects `CompareExpectedFS` implementation.
Without this answer, Task 2.1 cannot correctly implement the engine.

---

## Question 5: `configs/cryptoutil/cryptoutil.yml` — Suite-Level Standalone Config

**Context**: `configs/cryptoutil/cryptoutil.yml` exists with a deeply nested YAML schema
(different from PS-ID flat kebab-case configs):
```yaml
# Cryptoutil Suite-Level Configuration
observability:
  telemetry_enabled: true
  otlp_endpoint: "opentelemetry-collector-contrib:4317"
  ...
security:
  tls_min_version: "1.3"
  ...
admin:
  bind_address: "127.0.0.1"
  ...
```
This file is NOT covered in the v10 plan. It uses the old snake_case nested schema (never migrated
to flat kebab-case in v9) and has a completely different structure from `configs/sm-kms/sm-kms.yml`.

**Question**: What is the correct treatment of `configs/cryptoutil/cryptoutil.yml` in v10?

**A)** Add a `templates/configs/__SUITE__/__SUITE__.yml` template. Also migrate
`configs/cryptoutil/cryptoutil.yml` from nested snake_case to flat kebab-case to match PS-ID
standalone format (this file was missed in the v9 config standardization). Add corresponding
task to Phase 1.

**B)** Keep the file as-is (nested schema is intentional — it describes suite-wide shared defaults),
but exclude it from template drift linting. No template created. File is manually maintained.

**C)** Delete the file — it is not read by any running service. Each PS-ID has its own standalone
config; a suite-level shared config concept is not implemented anywhere in the Go code.

**D)** Treat as an infrastructure config (not a service deployment config). Move it to a new
`configs/infra/` or keep at `configs/cryptoutil/` but exclude from the template drift linter
scope entirely. Log it as a separate clean-up task outside v10.

**E)**

**Answer**:

**Rationale**: Determines whether Phase 1 needs a 3rd suite template file in addition to
`Dockerfile` and `compose.yml`. Not answering this leaves an untemplated file that will fail
`CompareExpectedFS` unless explicitly excluded.

---

## Question 6: `configs/pki-ca/profiles/` — 25 Certificate Profile YAML Files

**Context**: `configs/pki-ca/profiles/` contains 25 YAML files (`tls-server.yaml`,
`code-signing.yaml`, `database-server.yaml`, etc.) defining certificate profile constraints
(validity, key algorithms, key usage, SANs, extensions). These are pki-ca-domain-specific
and are NOT referenced in the v10 plan at all.

Sample structure of each profile file:
```yaml
profile:
  name: "tls-server"
  validity: { max_days: 398, default_days: 365 }
  key: { allowed_algorithms: [...], default_algorithm: "ECDSA" }
  san: { allow_dns_names: true, dns_patterns: ["*.example.com"] }
  ...
```

**Question**: How should the 25 `configs/pki-ca/profiles/` files be treated in v10?

**A)** Completely static — zero parameterization, zero template drift validation. Exclude the
entire `configs/pki-ca/profiles/` subtree from `CompareExpectedFS` scope. No changes needed.

**B)** Include in template compliance as static files: add `templates/configs/__PS_ID__/profiles/`
to the template tree for pki-ca only. Files have no `__KEY__` placeholders but ARE copied into the
template directory and compared byte-for-byte to detect unintended modifications.

**C)** Out of v10 scope — a future pki-ca-specific initiative will decide whether these files
need parameterization or standardization. Explicitly exclude them from `CompareExpectedFS` and
log as a known exclusion in plan.md.

**D)** Parameterized templates: DNS name patterns (`*.example.com`) and IP ranges are
environment-specific. Add `__DOMAIN__` and `__IP_RANGE__` placeholders. Include in the template
tree. This requires a new registry.yaml field `domain:` per suite or product.

**E)**

**Answer**:

**Rationale**: If A or C, the linter must explicitly exclude `configs/pki-ca/profiles/` to avoid
false positives. If B or D, new template files must be created in Phase 1.

---

## Question 7: `configs/identity-authz/domain/policies/` — 3 Policy YAML Files

**Context**: `configs/identity-authz/domain/policies/` contains 3 files:
`adaptive-authorization.yml`, `risk-scoring.yml`, `step-up.yml`. These define risk-based
authentication policy (risk thresholds, required MFA methods, session durations). They are
specific to `identity-authz` and are NOT referenced in the v10 plan.

**Question**: How should the 3 `configs/identity-authz/domain/policies/` files be treated in v10?

**A)** Completely static — zero parameterization, zero template drift validation. Exclude the
entire `configs/identity-authz/domain/policies/` subtree from `CompareExpectedFS` scope.

**B)** Include in template compliance as static files: add
`templates/configs/__PS_ID__/domain/policies/` to the template tree for `identity-authz` only.
Files have no placeholders but ARE compared byte-for-byte to detect unintended modifications.

**C)** Out of v10 scope — a future identity-specific initiative will handle these. Explicitly
exclude from `CompareExpectedFS` and log as a known exclusion in plan.md.

**D)** Parameterized templates — risk thresholds and session durations are operation-environment
specific. Add placeholder substitution (e.g., `__RISK_THRESHOLD_LOW__`). Include in template tree.

**E)**

**Answer**:

**Rationale**: Same structural situation as Q6. If A or C, the linter must exclude this subtree.
If B, template files must be created. `CompareExpectedFS` will fail on unexpected files unless
scope is explicit.

---

## Question 8: `deployments/shared-postgres/` — INFRA_TOOL Template Scope

**Context**: `deployments/shared-postgres/compose.yml` is a complex file with:
- 30-logical-database architecture (10 PS-IDs × 3 deployment tiers = suitedeployment-*, productdeployment-*, servicedeployment-*)
- 2 services: `postgres-leader` (OLTP) and `postgres-follower` (OLAP)
- Volume mounts referencing `init-leader-databases.sql`, `init-follower-databases.sql`, `setup-logical-replication.sh`
- Container names hardcoded: `cryptoutil-postgres-leader`, `cryptoutil-postgres-follower`

The v10 plan lists `__INFRA_TOOL__` as a 4th expansion key but does NOT specify what template
files are needed, or whether they exist at all.

**Question**: What template treatment does `deployments/shared-postgres/` get in v10?

**A)** Compose-only, `__SUITE__`-parameterized: `templates/deployments/__INFRA_TOOL__/compose.yml`
template with `__SUITE__` substitutions for container names (`cryptoutil-postgres-*`). SQL init
files and shell scripts are static (not templates). The 30 database names in comments are
auto-generated from registry PS-IDs using a `__SUITE_PS_ID_DB_LIST__` placeholder.

**B)** Compose-only, static content validation: template file is byte-for-byte copy of actual
`compose.yml` with NO `__KEY__` placeholders. Engine compares actual file against template for
drift detection. SQL init files and shell scripts are excluded from comparison.

**C)** Out of v10 scope — shared-postgres is infrastructure managed separately. Explicitly
exclude `deployments/shared-postgres/` from `CompareExpectedFS`. Note as known exclusion.

**D)** No template, no exclusion — `CompareExpectedFS` only checks files that have a template
counterpart. Files with no template entry in the expected FS are silently ignored (the linter
is a one-directional check: all expected files must exist, but extra disk files are allowed).

**E)**

**Answer**:

**Rationale**: v10 plan Phase 2 describes `CompareExpectedFS` as one-directional (extra files on
disk are allowed). If D, no action needed for shared-postgres. If A/B/C, new tasks are required.

---

## Question 9: `deployments/shared-telemetry/` — Grafana/OTel Subdirectory Scope

**Context**: `deployments/shared-telemetry/` contains:
- `compose.yml` — 3 services (healthcheck-otel, opentelemetry-collector-contrib, grafana-otel-lgtm)
- `alerts/cryptoutil.yml` — Prometheus alerting rules
- `grafana-otel-lgtm/dashboards/*.json` — 3 Grafana dashboard JSON files
- `grafana-otel-lgtm/provisioning/dashboards/dashboards.yaml`
- `grafana-otel-lgtm/provisioning/datasources/prometheus.yaml`
- `otel/cryptoutil-otel.yml` — suite-specific otel config
- `otel/otel-collector-config.yaml` — otel collector config

Many files contain `cryptoutil` hardcoded (otel service names, alert labels, etc.).

**Question**: What template treatment does `deployments/shared-telemetry/` get in v10?

**A)** Compose + otel configs only: `templates/deployments/__INFRA_TOOL__/compose.yml` and
`templates/deployments/__INFRA_TOOL__/otel/otel-collector-config.yaml` get templates with
`__SUITE__` substitution. Grafana JSON dashboards and provisioning YAML are static (too
complex/fragile to template).

**B)** Compose only: only `compose.yml` gets a template. All subdirectory files (grafana, otel,
alerts) are static and excluded from template drift.

**C)** Full parameterization: all text files that contain `cryptoutil` get templates with `__SUITE__`
substitution. Grafana JSON files also get `__SUITE__` in service name strings.

**D)** Out of v10 scope / one-directional: same reasoning as Q8 Option D — extra files on disk
are silently ignored by `CompareExpectedFS`. No template files needed, no exclusion needed.

**E)**

**Answer**:

**Rationale**: Controls how many template files Phase 1 must create for the infra-tool tier.
See `deployments/shared-telemetry/otel/otel-collector-config.yaml` for cryptoutil references.

---

## Question 10: `registry.yaml` — Adding `infra_tools` Section

**Context**: The v10 plan's `BuildExpectedFS` loops over `__INFRA_TOOL__` values from registry.yaml
for expansion. But the current `registry.yaml` (266 lines) has NO `infra_tools` section. The two
known infra tools are `shared-postgres` and `shared-telemetry`.

If Questions 8 and 9 result in infra tool templates being needed, registry.yaml must provide
the list. The `BuildExpectedFS` function needs: infra tool id (e.g., `shared-postgres`),
and optionally display_name and compose_dir.

**Question**: Should `registry.yaml` have an `infra_tools` section, and on what timeline?

**A)** Add `infra_tools` section to `registry.yaml` in Phase 1 with id and display_name for
`shared-postgres` and `shared-telemetry`. `BuildExpectedFS` reads this in Phase 2 for `__INFRA_TOOL__`
expansion. This is required regardless of Q8/Q9 answers so the engine design is complete.

**B)** Skip registry.yaml changes — if Q8 and Q9 both result in "out of scope" (Options C or D),
no infra_tools section is needed in v10. Defer until a future phase when infra tool templates
are actually implemented.

**C)** Add infra_tools to registry.yaml but hardcode the list in `template_drift.go` as backup.
If registry section exists, use it; if not, fall back to hardcoded list. Avoids blocking
implementation on registry schema decisions.

**D)** Add infra_tools to registry.yaml AND add template files for each infra tool in Phase 1
(compose.yml at minimum), resolving Q8 and Q9 as scope-included rather than deferred.

**E)**

**Answer**:

**Rationale**: `BuildExpectedFS` design depends on whether it reads an `infra_tools` list from
registry or handles it differently. Answering this before Phase 2 coding begins prevents rework.
Add task to Phase 1 if A or D. Skip if B.

---

## Post-Answer Actions

After you fill in all 10 `Answer:` fields above, the execution agent will:
1. Delete this quizme file
2. Update `docs/framework-v10/plan.md` Architecture Decisions section with each decision
3. Update `docs/framework-v10/tasks.md` to add/modify tasks for Phase 1 and Phase 2 based on answers
4. Commit with `docs(framework-v10): merge quizme-v1 answers into plan/tasks`
