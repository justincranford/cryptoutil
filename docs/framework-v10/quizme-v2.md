# Quizme v2 — Framework v10: Canonical Template Registry

**Context**: quizme-v1 had 10 questions; 8 answered (Q1-Q7, Q9), 2 unanswered (Q8, Q10).
Deep analysis of `./configs/` and `./deployments/` revealed 5 additional ambiguities.
This quizme has 7 questions: 2 carry-forwards (Q8/Q10) + 5 new from deep analysis.

**Instructions**: For each question, write your answer letter (A, B, C, D, or E) on the
**Answer:** line. Option E is blank — write your custom answer if A-D don't fit.

---

## Question 1: Standalone Config Domain Extensions (Q4 Revisited)

**Background**: In quizme-v1 Q4, you chose E ("I assume no domain-specific additions are
required. All standalone configs would match the template exactly. If this causes issues,
I need concrete examples."). Deep analysis shows your assumption is **incorrect** — 6 of 10
PS-IDs have domain-specific content after the shared template portion.

**Concrete evidence** from actual `configs/<ps-id>/<ps-id>.yml` files:

| PS-ID | Domain Extension Content |
|-------|--------------------------|
| sm-kms | `database-url:` (1 line) |
| sm-im | `database-url:` + 6 session config lines (`browser-session-algorithm`, `service-session-algorithm`, etc.) |
| pki-ca | Extensive: `dev:`, `ca:` (nested), `storage:` (nested), `revocation:`, `tsa:`, `est:`, `profiles:` |
| identity-authz | `database-url:`, `issuer:`, `token-lifetime:`, `refresh-token-lifetime:`, `authorization-code-ttl:`, `enable-discovery:`, `enable-dynamic-registration:` |
| identity-idp | Same OAuth settings as identity-authz |
| identity-rp | `database-url:` |
| identity-rs | `database-url:` |
| identity-spa | `database-url:` |
| jose-ja | Clean — no extensions (ends at `log-level: "INFO"`) |
| skeleton-template | Clean — no extensions |

**Question**: Given that 6/10 PS-IDs have domain extensions in their standalone configs,
how should the template handle this?

**A)** **Prefix-only template**: The `__PS_ID__.yml` template covers ONLY the shared prefix
(bind-*, tls-*, cors-*, otlp-*, log-level) — identical across all 10 PS-IDs. Domain extensions
are extra content appended by each PS-ID. The linter does **prefix comparison**: template
content must match the FIRST N lines of the actual file; extra lines after the template portion
are silently allowed. (Simplest: no domain knowledge in templates.)

**B)** **Suffix marker**: Same as A, but each PS-ID standalone config has a `# --- domain ---`
marker line separating template content from domain content. The linter compares everything
ABOVE the marker (or the whole file if no marker exists). (Explicit boundary, easy to validate.)

**C)** **Split template + domain**: The template covers only the shared prefix. Each PS-ID MAY
have an additional `configs/__PS_ID__/domain/` directory containing domain-specific YAML
fragments that are NOT compared against any template. The standalone config file itself is
100% template-generated; domain extensions are in separate files loaded at runtime.
(Clean separation, but requires runtime config merge changes.)

**D)** **Per-PS-ID domain template**: In addition to the shared `__PS_ID__.yml` template, PS-IDs
with domain extensions get a SECOND template file with domain-specific content. Registry.yaml
tracks which PS-IDs have domain templates. (Most precise, but high maintenance burden —
every domain config change requires template update.)

**E)**

**Answer:**

**Rationale**: This determines whether the standalone config template is a full-file comparison
or a prefix-only comparison. Most of v10 template infrastructure (Dockerfile, compose, deployment
configs) does full-file comparison. Standalone configs may need a different comparison mode.

---

## Question 2: Config Hierarchy Scope (Q5 Clarification)

**Background**: In quizme-v1 Q5, you chose E ("SUITE/PRODUCT compose files should
prepend/append extra YAML config file parameters to PS-ID startup commands. Config hierarchy:
PS-ID → PRODUCT → SUITE").

**Analysis**: This describes a **new runtime feature** — product/suite compose files would
need `command:` overrides that layer additional config on top of PS-ID configs. Currently:
- PS-ID compose: `command: ["server", "--config=/app/config/sm-kms-app-sqlite-1.yml", ...]`
- Product compose: overrides ONLY `ports:` and `secrets:` — does NOT touch `command:`
- Suite compose: overrides ONLY `ports:` and `secrets:` — does NOT touch `command:`

Implementing config hierarchy requires:
1. Creating product-level and suite-level config files (don't exist today)
2. Modifying product/suite compose templates to add `command:` overrides appending those config files
3. Ensuring the Go application supports multiple `--config=` flags with last-wins merging

**Question**: Is the config hierarchy a v10 scope item, or future work?

**A)** **v10 scope**: Implement the full config hierarchy in v10. Product/suite compose templates
add `command:` overrides to append product/suite config files. Requires creating the config files
and testing the Go app's multi-config merge. (~2-3h additional LOE.)

**B)** **Future work — document only**: Config hierarchy is NOT v10 scope. v10 focuses on template
linting. Document the design in plan.md as a future enhancement. Product/suite compose templates
continue to override only `ports:` and `secrets:` (no `command:` changes).

**C)** **v10 scope — compose only, no Go changes**: Add config file volume mounts and `command:`
overrides to product/suite compose templates, but do NOT modify the Go application. The config
files are created but not loaded until a future phase adds multi-config support. (Partial
implementation — compose infrastructure ready, app support deferred.)

**D)** **Future work — separate plan**: Create a separate `docs/config-hierarchy/plan.md` for this
feature. v10 templates are designed to NOT conflict with future config hierarchy additions (e.g.,
reserved `command:` override slots or documented extension points).

**E)**

**Answer:**

**Rationale**: Config hierarchy is a significant scope expansion. v10's primary objective is
template linting infrastructure. Adding runtime config behavior changes risks scope creep and
delays template linting delivery.

---

## Question 3: shared-postgres Template Scope (Q8 Carry-Forward)

**Background**: quizme-v1 Q8 was left unanswered. shared-postgres is under
`deployments/shared-postgres/` and provides PostgreSQL infrastructure shared by multiple
PS-ID services.

**Current structure**:
```
deployments/shared-postgres/
  compose.yml                    ← PostgreSQL leader + follower + init-db
  config/
    postgresql.conf              ← PostgreSQL tuning
  init/
    01-create-databases.sql      ← Creates all 10 PS-ID databases
```

**Key observations**:
- `compose.yml` references service names like `cryptoutil-shared-postgres-leader` (suite-specific)
- `01-create-databases.sql` creates 10 databases matching PS-ID names + `_database` suffix
- `postgresql.conf` is generic PostgreSQL tuning (no suite/PS-ID references)

**Question**: How should shared-postgres be handled in the template system?

**A)** **Full template**: All 3 files become templates with `__SUITE__` substitution. The SQL
file uses `__PS_ID_DATABASE_LIST__` (a multi-line param generated from registry.yaml listing
all PS-ID database names). (Most complete — validates all shared-postgres content against
the registry.)

**B)** **Compose-only template**: Only `compose.yml` gets a template (with `__SUITE__`
substitution). `postgresql.conf` and `01-create-databases.sql` are static files — too complex
or too volatile to template. (Partial coverage — validates compose structure but not database
creation script.)

**C)** **No template**: shared-postgres is infrastructure, not a service. It has no `__PS_ID__`
variation — there is exactly one shared-postgres per deployment. Comparing it against a template
adds complexity without value. Leave it outside the template system entirely.

**D)** **Static comparison only**: Copy the current shared-postgres files as-is into
`templates/deployments/shared-postgres/`. No parameterization — just exact byte-for-byte
comparison. Detects any accidental drift. (Simple but rigid — any intentional change requires
updating both the template and the actual file.)

**E)**

**Answer:**

**Rationale**: shared-postgres is the only infrastructure component that creates resources based
on the PS-ID registry (the SQL init script). Deciding its template scope affects how tightly
the template system validates infrastructure consistency.

---

## Question 4: Registry Categorization (Q10 Carry-Forward)

**Background**: quizme-v1 Q10 was left unanswered. The current `registry.yaml` has `suites`,
`products`, and `product_services` sections. It does NOT model shared infrastructure
(shared-postgres, shared-telemetry) or build tools (cicd-lint, cicd-workflow).

**Your Q10 E) text** (from quizme-v1): "cicd-lint and cicd-workflow... are infra tools, NOT
shared services. shared-postgres and shared-telemetry are shared services, NOT infra tools."

**Question**: How should `registry.yaml` be extended to model non-PS-ID deployment artifacts?

**A)** **Add `shared_services` section**: Add a top-level `shared_services:` section to
registry.yaml listing `shared-postgres` and `shared-telemetry` with their compose paths and
relationship to the suite. Template expansion uses this section for shared service templates.
`cicd-lint` and `cicd-workflow` are NOT added to registry.yaml (they are build tools, not
deployment artifacts).

**B)** **Add both sections**: Add `shared_services:` (shared-postgres, shared-telemetry) AND
`infra_tools:` (cicd-lint, cicd-workflow) to registry.yaml. Both have compose paths.
Template expansion handles both categories.

**C)** **No registry changes**: shared-postgres and shared-telemetry templates use static paths
(no registry expansion). Registry.yaml stays focused on suites/products/PS-IDs. Template files
for shared services are at hardcoded paths like `templates/deployments/shared-telemetry/...`
(already decided in Decision 11).

**D)** **Add `infrastructure` section**: Single section for ALL non-PS-ID deployment artifacts:
shared-postgres, shared-telemetry, and any future shared infrastructure. No separate categories.
cicd-lint/cicd-workflow are excluded (build tools, not deployment infrastructure).

**E)**

**Answer:**

**Rationale**: Decision 11 already placed shared-telemetry templates at a static path. This
question determines whether registry.yaml becomes a broader deployment manifest or stays focused
on the product/service hierarchy. The answer affects `__INFRA_TOOL__` expansion (currently
unused in the template directory structure).

---

## Question 5: Config File Naming Convention

**Background**: The actual deployment config files use a **PS-ID prefix** naming convention:
```
deployments/sm-kms/config/sm-kms-app-common.yml
deployments/sm-kms/config/sm-kms-app-sqlite-1.yml
deployments/sm-kms/config/sm-kms-app-postgresql-1.yml
```

The original plan.md used a **generic** naming convention:
```
templates/deployments/__PS_ID__/config/config-common.yml
templates/deployments/__PS_ID__/config/config-sqlite-1.yml
```

**Question**: Which naming convention should the template files use?

**A)** **PS-ID prefix (matches actual)**: Template filenames use `__PS_ID__-app-common.yml`,
`__PS_ID__-app-sqlite-1.yml`, etc. After expansion, `__PS_ID__` = `sm-kms` produces
`sm-kms-app-common.yml` — matching the actual filenames exactly. (Correct — templates must
produce file paths that match actual files on disk.)

**B)** **Generic prefix (config-)**: Template filenames use `config-common.yml`,
`config-sqlite-1.yml`, etc. Rename all actual deployment config files to drop the PS-ID prefix.
(Breaking change — requires renaming 50 actual config files.)

**C)** **PS-ID prefix with `app` configurable**: Template filenames use
`__PS_ID__-__CONFIG_INFIX__-common.yml` with `__CONFIG_INFIX__` = `app` as a registry param.
(Over-parameterized — `app` has never varied.)

**D)** **Both formats**: Template supports both; actual files can use either naming convention.
Linter normalizes filenames for comparison. (Adds unnecessary complexity.)

**E)**

**Answer:**

**Rationale**: The template must produce filenames that match the actual files on disk.
The plan.md and tasks.md have already been updated to use `__PS_ID__-app-*.yml` (Option A),
but this question confirms the decision explicitly.

---

## Question 6: Product Compose `pki-init` Image and Builder Dependency

**Background**: Decision 4 (Q1) requires all product compose files to include a `pki-init`
service with `--domain=<product>`. Currently, only the SM product compose has `pki-init`:

```yaml
# From deployments/sm/compose.yml
pki-init:
  image: cryptoutil-sm-kms:dev
  depends_on:
    builder-sm-kms:
      condition: service_completed_successfully
  command: ["init", "--output-dir=/certs", "--domain=sm"]
```

**Problem**: The pki-init service needs:
1. An `image:` reference — currently `cryptoutil-sm-kms:dev` (hardcoded to a specific PS-ID's image)
2. A `depends_on:` to a builder service — currently `builder-sm-kms` (from the included PS-ID compose)

For the template, both need to be parameterized. But products have different numbers of PS-IDs:
- SM: 2 PS-IDs (sm-kms, sm-im)
- Jose: 1 PS-ID (jose-ja)
- PKI: 1 PS-ID (pki-ca)
- Identity: 5 PS-IDs (identity-authz, identity-idp, identity-rs, identity-rp, identity-spa)
- Skeleton: 1 PS-ID (skeleton-template)

**Question**: How should the product template parameterize pki-init's image and builder?

**A)** **First PS-ID convention**: Registry.yaml designates a "primary PS-ID" per product (the
first one listed). Template uses `__PRIMARY_PS_ID__` for pki-init image and builder dependency.
SM primary = sm-kms, Identity primary = identity-authz, etc.

**B)** **Dedicated pki-init image**: Create a lightweight `pki-init` Docker image (or use the
suite-level image `cryptoutil:dev`) for product-level pki-init. Template uses
`__SUITE__-pki-init:dev` or `__SUITE__:dev` as the image. Builder dependency removed — pki-init
uses a pre-built image. (Decouples pki-init from PS-ID builders.)

**C)** **Product-level builder**: Add a `builder-__PRODUCT__` service to the product compose
template. Template uses `__PRODUCT__` for both the builder service name and the image. Product
compose builds its own image for pki-init. (Clean but adds a build step at product level.)

**D)** **Registry `pki_init_ps_id` field**: Add a `pki_init_ps_id` field to each product in
registry.yaml specifying which PS-ID's builder/image to use for that product's pki-init.
Template uses `__PKI_INIT_PS_ID__` for image and builder. (Explicit and flexible but adds
registry complexity.)

**E)**

**Answer:**

**Rationale**: This is a BLOCKING question for Task 1.4 (product compose template). The
pki-init service cannot be fully parameterized until this is decided. SM product currently
hardcodes `sm-kms` — the template needs a general solution.

---

## Question 7: Secrets Directory Template Scope

**Background**: Each deployment level has a `secrets/` directory:

**PS-ID level** (`deployments/sm-kms/secrets/`) — 14 files:
- 5 unseal keys (`unseal-1of5.secret` through `unseal-5of5.secret`)
- `hash-pepper-v3.secret`
- 4 postgres secrets (`postgres-url.secret`, `-username`, `-password`, `-database`)
- `browser-password.secret`, `browser-username.secret`
- `service-password.secret`, `service-username.secret`

**Product level** (`deployments/sm/secrets/`) — 14 files:
- Same filenames but browser/service credentials use `.secret.never` suffix (marker files
  indicating these are PS-ID-level concerns, not product-level)

**Suite level** (`deployments/cryptoutil/secrets/`) — 14 files:
- Same pattern as product: `.secret.never` for browser/service credentials

**Secret file CONTENTS contain actual random values** — they cannot be compared byte-for-byte.
Only the **filename structure** (which files exist, their naming pattern) can be validated.

**Question**: Should the template system validate secrets directory structure?

**A)** **Full directory template**: Template includes a `secrets/` directory listing all expected
filenames per level. Linter checks existence only (file must exist, content ignored).
Different filename lists for PS-ID (`.secret`) vs product/suite (`.secret` + `.secret.never`).

**B)** **Filename list in registry.yaml**: Registry.yaml defines expected secret filenames per
tier (ps_id_secrets, product_secrets, suite_secrets). Linter validates files exist but compares
no content. No template files for secrets — the registry is the spec.

**C)** **No secrets validation**: Secrets directories are outside template scope. They contain
actual secret material and are managed by deployment operators, not validated by a linter.
(Simplest — but risks secrets directory inconsistency across PS-IDs.)

**D)** **Existence audit only**: No templates or registry entries. Add a separate
`lint-fitness` check (not template-compliance) that validates: every PS-ID deployment dir
has a `secrets/` subdirectory with the expected 14 files. Product/suite have their 14 files.
(Dedicated linter, not part of template drift.)

**E)**

**Answer:**

**Rationale**: Secrets directories have unique constraints: content is random (cannot be
compared), but structure is uniform (same filenames across all PS-IDs). The template system is
designed for content comparison — extending it to existence-only checks requires a different
comparison mode.
