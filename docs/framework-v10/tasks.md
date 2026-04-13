# Tasks — Framework v10: Canonical Template Registry

**Status**: 0 of 33 tasks complete (0%)
**Created**: 2026-04-12
**Last Updated**: 2026-04-13

---

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥98% coverage/mutation for infrastructure)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions:**
- ✅ **Fix issues immediately** — when unknowns discovered, blockers identified, any tests fail, or quality gates not met, STOP and address
- ✅ **Treat as BLOCKING**: ALL issues block progress to next task
- ✅ **NEVER defer**: No "fix later", no "non-critical", no shortcuts

---

## Phase 1: Create Canonical Template Directory

**Phase Objective**: Populate `api/cryptosuite-registry/templates/` with all parameterized
template files (~63 files). These are plain configuration files — NOT Go code. They mirror the
structure of `./deployments/` and `./configs/` with `__KEY__` placeholders in both directory paths
and content. Includes secrets directory templates (Decision 14), shared-postgres templates
(Decision 10), framework/domain config split (Decision 7/16), and postgresql conf files.
Also fix actual deployment files to match quizme-v1/v2/v3 decisions: pki-init pflag rewrite
(Decision 15), shell-form command with `$SUITE_ARGS` (Decision 8), deployment config
framework/domain split (Decision 16), postgres secrets uniformity, image removal, standalone
config split, and stale postgres URL fix (Decision 17).
Delete the obsolete `template_drift/templates/` directory.

### Task 1.1: Read and understand all existing template content

- **Status**: ✅
- **Estimated**: 0.5h
- **Dependencies**: None
- **Description**: Before creating any template files, read all 6 existing `.tmpl` files in
  `internal/.../template_drift/templates/` to understand exact content and placeholder usage.
  Also read the actual product/suite compose files to understand product template content.
- **Acceptance Criteria**:
  - [ ] All 6 `.tmpl` files read and understood
  - [ ] Actual `deployments/{sm,jose,pki,identity,skeleton}/compose.yml` files read
  - [ ] Actual `deployments/cryptoutil/Dockerfile` and `compose.yml` read
  - [ ] Param usage per template type noted

### Task 1.2: Create `api/cryptosuite-registry/templates/` directory tree

- **Status**: ✅
- **Estimated**: 0.25h
- **Dependencies**: Task 1.1
- **Description**: Create the directory structure. No `.go` files. No Go package. Plain directories only.
- **Acceptance Criteria**:
  - [ ] `api/cryptosuite-registry/templates/deployments/__PS_ID__/` exists
  - [ ] `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/` exists
  - [ ] `api/cryptosuite-registry/templates/deployments/__PS_ID__/secrets/` exists
  - [ ] `api/cryptosuite-registry/templates/deployments/__PRODUCT__/` exists (single product template dir)
  - [ ] `api/cryptosuite-registry/templates/deployments/__PRODUCT__/secrets/` exists
  - [ ] `api/cryptosuite-registry/templates/deployments/__SUITE__/` exists (single suite template dir)
  - [ ] `api/cryptosuite-registry/templates/deployments/__SUITE__/secrets/` exists
  - [ ] `api/cryptosuite-registry/templates/deployments/shared-postgres/` exists (static path)
  - [ ] `api/cryptosuite-registry/templates/deployments/shared-postgres/secrets/` exists
  - [ ] `api/cryptosuite-registry/templates/configs/__PS_ID__/` exists
  - [ ] NO `.go` files anywhere in `api/cryptosuite-registry/`

### Task 1.3: Create PS-ID level template files (7 deployment + 1 config)

- **Status**: ✅
- **Estimated**: 1.5h
- **Dependencies**: Task 1.2
- **Description**: Convert existing `.tmpl` files into parameterized template files.
  The content is identical to current `.tmpl` content but placed in parameterized subdirectory
  paths using `__PS_ID__` (no `.tmpl` extension). `config-sqlite.yml.tmpl` becomes TWO files
  (instance-1 and instance-2); same for config-postgresql. Config filenames use `__PS_ID__`
  prefix with `framework` qualifier to match Decision 16 naming convention
  (e.g., `sm-kms-app-framework-common.yml`).

  **Dockerfile** ENTRYPOINT changed from `["/sbin/tini", "--", "/app/__PS_ID__"]` to
  `["/sbin/tini", "--"]` (Decision 8, quizme-v4 Q1). App binary moves into compose command.
  Tini is already installed: `apk --no-cache add tini` in runtime-deps stage.

  **compose.yml** MUST include `pki-init` service with pflag format and app binary in command:
  `["/app/__PS_ID__", "init", "--domain=__PS_ID__", "--output-dir=/certs"]` (Decision 4/15;
  app binary in command because ENTRYPOINT is tini-only).
  **compose.yml** MUST use shell-form command for app service:
  `command: /bin/sh -c "exec /app/__PS_ID__ server --config=... $SUITE_ARGS"` (Decision 8).
  Config merge order (quizme-v4 Q3): later `--config=` files override earlier for same keys.
  Specify all configs in precedence order (lowest first, highest last):
  TLS → framework-common → framework-variant → domain-common → domain-variant → otel.

  **Domain deployment configs** (`__PS_ID__-app-domain-*.yml`) are per-PS-ID and NOT templated
  (Decision 16). They MUST exist on disk (initially empty except pki-ca `crl-directory`),
  but are not in the template directory. Created by Task 1.8.

  Files to create in `api/cryptosuite-registry/templates/deployments/__PS_ID__/`:
  - `Dockerfile` (from `Dockerfile.tmpl`)
  - `compose.yml` (from `compose.yml.tmpl`; pflag pki-init + shell-form command)
  - `config/__PS_ID__-app-framework-common.yml` (from `config-common.yml.tmpl`, Decision 16)
  - `config/__PS_ID__-app-framework-sqlite-1.yml` (from `config-sqlite.yml.tmpl`, instance-1)
  - `config/__PS_ID__-app-framework-sqlite-2.yml` (from `config-sqlite.yml.tmpl`, instance-2)
  - `config/__PS_ID__-app-framework-postgresql-1.yml` (from `config-postgresql.yml.tmpl`, instance-1)
  - `config/__PS_ID__-app-framework-postgresql-2.yml` (from `config-postgresql.yml.tmpl`, instance-2)

  File to create in `api/cryptosuite-registry/templates/configs/__PS_ID__/`:
  - `__PS_ID__-framework.yml` — framework settings only (Decision 7/13: framework/domain split).
    Contains: `bind-*`, `tls-*`, `cors-*`, `otlp-*`, `log-level`, `database-url`,
    session algorithms, `enable-dynamic-registration`.
    Domain config (`__PS_ID__-domain.yml`) is per-PS-ID and NOT templated.
- **Acceptance Criteria**:
  - [ ] 7 files in `templates/deployments/__PS_ID__/` (Dockerfile + compose + 5 framework configs)
  - [ ] Config filenames use `__PS_ID__-app-framework-` prefix (Decision 16)
  - [ ] Dockerfile ENTRYPOINT is `["/sbin/tini", "--"]` (NOT `["/sbin/tini", "--", "/app/__PS_ID__"]`)
  - [ ] compose.yml includes `pki-init` service with `["/app/__PS_ID__", "init", "--domain=__PS_ID__", "--output-dir=/certs"]`
  - [ ] compose.yml uses shell-form command with `$SUITE_ARGS` (Decision 8)
  - [ ] 1 file in `templates/configs/__PS_ID__/` (`__PS_ID__-framework.yml`)
  - [ ] Framework config template has all framework settings, NO domain settings
  - [ ] Content matches existing `.tmpl` files; instance-1 vs instance-2 files differ only in port param
- **Files**:
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/Dockerfile` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-common.yml` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-sqlite-1.yml` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-sqlite-2.yml` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-postgresql-1.yml` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/__PS_ID__-app-framework-postgresql-2.yml` (CREATE)
  - `api/cryptosuite-registry/templates/configs/__PS_ID__/__PS_ID__-framework.yml` (CREATE)

### Task 1.4: Create product compose template file (1 file, `__PRODUCT__` in path)

- **Status**: ✅
- **Estimated**: 1.5h
- **Dependencies**: Task 1.2
- **Description**: Create ONE parameterized template file at
  `api/cryptosuite-registry/templates/deployments/__PRODUCT__/compose.yml`.
  The directory name `__PRODUCT__` is the expansion placeholder; the linter expands it × 5
  (sm, jose, pki, identity, skeleton) using per-product params from registry.yaml.

  The template content uses:
  - `__PRODUCT__` — substituted with product name (sm, jose, etc.)
  - `__SUITE__` — substituted with suite name
  - `__IMAGE_TAG__` — substituted with image tag
  - `__PRODUCT_INCLUDE_LIST__` — multi-line YAML include entries generated from registry.yaml
    (each product's PS-IDs produce include lines like `- path: ../sm-kms/compose.yml`)

  **Decisions applied**:
  - **Decision 4 (Q1) + Decision 15**: Template MUST include `pki-init` service with full definition:
    - `image: cryptoutil-__PRODUCT_INIT_PS_ID__:dev`
    - `command: ["/app/__PRODUCT_INIT_PS_ID__", "init", "--domain=__PRODUCT__", "--output-dir=/certs"]`
      (app binary in command because ENTRYPOINT is tini-only)
    - `volumes: [./certs/:/certs/:rw]`
    - `depends_on: builder-__PRODUCT_INIT_PS_ID__: { condition: service_completed_successfully }`
    `__PRODUCT_INIT_PS_ID__` = first PS-ID of the product from registry.yaml (sm→sm-kms,
    jose→jose-ja, pki→pki-ca, identity→identity-authz, skeleton→skeleton-template).
    **Product-level override (quizme-v3 Q6 + quizme-v4 Q2)**: Product compose uses standard
    Docker Compose service-name override. Product includes PS-ID composes (which define
    `pki-init`), then redefines `pki-init` service entirely with the product-level definition.
    This is a clean full override — the PS-ID-level pki-init is fully replaced, not left as
    a no-op. SM product compose already uses this pattern correctly.
    Docker Compose profiles are BANNED (Decision 18).
  - **Decision 5 (Q2)**: Template MUST NOT include `image:` on APP service overrides (port-only).
    PS-ID compose is the single source for APP image references. SM product currently has `image:`
    on port overrides — must be removed from actual `deployments/sm/compose.yml`.
    Note: `pki-init` IS a full service definition (not a port override) and DOES need `image:`.
  - **Decision 6 (Q3)**: Template MUST reference all 4 postgres secrets:
    `postgres-url.secret`, `postgres-username.secret`, `postgres-password.secret`,
    `postgres-database.secret`. Currently jose/pki/identity/skeleton only reference
    `postgres-url.secret`.
- **Acceptance Criteria**:
  - [ ] Exactly ONE template file: `templates/deployments/__PRODUCT__/compose.yml`
  - [ ] Template uses `__PRODUCT__`, `__SUITE__`, `__IMAGE_TAG__`, `__PRODUCT_INCLUDE_LIST__`
  - [ ] Template includes `pki-init` with full definition: `image`, `command` with `/app/__PRODUCT_INIT_PS_ID__`, `volumes`, `depends_on`
  - [ ] Template fully overrides PS-ID pki-init services via service-name override (NOT no-op, NOT profiles)
  - [ ] Template includes all 4 postgres secrets (Decision 6)
  - [ ] Template has NO `image:` on service overrides (Decision 5)
  - [ ] No literal product names (`sm`, `jose`, `pki`, `identity`, `skeleton`) in the path
  - [ ] `go build ./...` succeeds
- **Files**:
  - `api/cryptosuite-registry/templates/deployments/__PRODUCT__/compose.yml` (CREATE)

### Task 1.5: Create suite template files (2 files, `__SUITE__` in path)

- **Status**: ✅
- **Estimated**: 0.5h
- **Dependencies**: Task 1.2
- **Description**: Create suite-level template files in `templates/deployments/__SUITE__/`.
  The directory name `__SUITE__` is the expansion placeholder; the linter expands it × 1
  (currently `cryptoutil`). Parameterized even though there is only one suite value today.

  Template content uses `__SUITE__` throughout (not the hard-coded literal `cryptoutil`).
  These are compared against `deployments/cryptoutil/{Dockerfile,compose.yml}` (after expansion).

  **Decisions applied**:
  - **Decision 4 (Q1) + Decision 15**: compose.yml MUST include `pki-init` with full definition:
    - `image: cryptoutil-__SUITE_INIT_PS_ID__:dev`
    - `command: ["/app/__SUITE_INIT_PS_ID__", "init", "--domain=__SUITE__", "--output-dir=/certs"]`
      (app binary in command because ENTRYPOINT is tini-only)
    - `volumes: [./certs/:/certs/:rw]`
    - `depends_on: builder-__SUITE_INIT_PS_ID__: { condition: service_completed_successfully }`
    `__SUITE_INIT_PS_ID__` = `sm-kms` (first PS-ID of first product).
    Actual suite compose already has `--domain=cryptoutil` ✔ — update to pflag format.
  - **Decision 6 (Q3)**: compose.yml MUST reference all 4 postgres secrets.
    Actual suite compose already has all 4 ✔.
  - **Decision 8**: compose.yml MUST use shell-form command for app services:
    `command: /bin/sh -c "exec /app/__PS_ID__ server --config=... $SUITE_ARGS"`.
- **Acceptance Criteria**:
  - [ ] `templates/deployments/__SUITE__/Dockerfile` created with `__SUITE__` and suite-level params
  - [ ] `templates/deployments/__SUITE__/compose.yml` created with `__SUITE__` and suite-level params
  - [ ] compose.yml includes `pki-init` with full definition: `image`, `command` with `/app/__SUITE_INIT_PS_ID__`, `volumes`, `depends_on`
  - [ ] compose.yml includes all 4 postgres secrets (Decision 6)
  - [ ] compose.yml uses shell-form command with `$SUITE_ARGS` (Decision 8)
  - [ ] Literal `cryptoutil` does NOT appear in either template file content (use `__SUITE__`)
  - [ ] Manually verifying: substituting `__SUITE__`=`cryptoutil` produces content identical to actual files
- **Files**:
  - `api/cryptosuite-registry/templates/deployments/__SUITE__/Dockerfile` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/__SUITE__/compose.yml` (CREATE)

### Task 1.6: Delete the obsolete `template_drift/templates/` directory

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Tasks 1.3, 1.4, 1.5, 1.7 (all canonical templates exist first)
- **Description**: Remove the 6 old `.tmpl` files from `internal/.../template_drift/templates/`.
  Their content is superseded by the canonical files in `api/cryptosuite-registry/templates/`.
- **Acceptance Criteria**:
  - [ ] `internal/apps/tools/cicd_lint/lint_fitness/template_drift/templates/` does NOT exist
  - [ ] `go build ./...` succeeds (no dangling references to removed directory)

### Task 1.7: Create shared-telemetry template files (2 files, static path)

- **Status**: ✅
- **Estimated**: 0.5h
- **Dependencies**: Task 1.2
- **Description**: Create 2 template files for shared-telemetry infrastructure (Decision 11).
  These use a static path (NOT expanded with `__INFRA_TOOL__`). Content uses `__SUITE__`
  substitution for service names and OTLP endpoints.

  Grafana dashboards, provisioning YAML, alerts, and `cryptoutil-otel.yml` are NOT templated
  (Decision 11: too complex/fragile). These are extra files on disk, allowed by the
  one-directional comparison (every expected file must exist; extra files are silently allowed).

  Files to create:
  - `api/cryptosuite-registry/templates/deployments/shared-telemetry/compose.yml`
  - `api/cryptosuite-registry/templates/deployments/shared-telemetry/otel/otel-collector-config.yaml`
- **Acceptance Criteria**:
  - [ ] Both files created with `__SUITE__` substitution in content
  - [ ] Manually verifying: substituting `__SUITE__`=`cryptoutil` produces content matching actual files
  - [ ] Grafana/alerts files are NOT referenced as templates (they are allowed as extra files)
- **Files**:
  - `api/cryptosuite-registry/templates/deployments/shared-telemetry/compose.yml` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/shared-telemetry/otel/otel-collector-config.yaml` (CREATE)

### Task 1.8: Fix actual deployment files to match quizme-v1/v2/v3 decisions

- **Status**: ❌
- **Estimated**: 2.5h
- **Dependencies**: Tasks 1.3, 1.4, 1.5, 1.15, 1.16, 1.17 (template content and code fixes finalized)
- **Description**: Update actual deployment files on disk to match all decisions from
  quizme-v1, quizme-v2, and quizme-v3. These fixes ensure the expanded templates match the
  actual files. Changes needed:

  **Decision 4 (Q1) + Decision 15 — pki-init pflag at all levels**:
  - Add `pki-init` service with pflag format
    `["/app/<ps-id>", "init", "--domain=<product>", "--output-dir=/certs"]`
    to `deployments/jose/compose.yml`, `deployments/pki/compose.yml`,
    `deployments/identity/compose.yml`, `deployments/skeleton/compose.yml` (currently missing)
  - Update all 10 PS-ID compose `pki-init` commands to pflag format:
    `["/app/<ps-id>", "init", "--domain=<ps-id>", "--output-dir=/certs"]` (app binary in command)
  - Product compose uses full service-name override for pki-init (Decision 4/Q2, quizme-v4 Q2):
    product defines its own `pki-init` service that fully replaces PS-ID-level pki-init.
    Docker Compose profiles are BANNED (Decision 18).

  **Decision 5 (Q2) — remove image from sm product**:
  - Remove `image: cryptoutil-sm-*:dev` from service overrides in `deployments/sm/compose.yml`

  **Decision 6 (Q3) — all 4 postgres secrets everywhere**:
  - Add `postgres-username.secret`, `postgres-password.secret`, `postgres-database.secret`
    to secrets sections in `deployments/jose/compose.yml`, `deployments/pki/compose.yml`,
    `deployments/identity/compose.yml`, `deployments/skeleton/compose.yml`

  **Decision 8 — shell-form command with `$SUITE_ARGS` + ENTRYPOINT change**:
  - Change all 10 PS-ID Dockerfiles: ENTRYPOINT from `["/sbin/tini", "--", "/app/<ps-id>"]`
    to `["/sbin/tini", "--"]` (quizme-v4 Q1). App binary moves into compose command.
  - Update all 10 PS-ID compose app service commands from exec-form to shell-form:
    `command: /bin/sh -c "exec /app/<ps-id> server --config=... $SUITE_ARGS"`
  - Config merge order (quizme-v4 Q3): specify all `--config=` files in precedence order
    (lowest first, highest last): TLS → framework-common → framework-variant →
    domain-common → domain-variant → otel

  **Decision 16 — deployment config framework/domain split**:
  - Rename all 10 PS-ID deployment config files from `<ps-id>-app-common.yml` etc. to
    `<ps-id>-app-framework-common.yml` etc.
  - Create domain deployment config files (`<ps-id>-app-domain-*.yml`) — initially empty
    except pki-ca which has `crl-directory` (see Task 1.17)

  **Decision 7 — standalone config split** (quizme-v2 Q1):
  - Split all 10 `configs/<ps-id>/<ps-id>.yml` into `<ps-id>-framework.yml` + `<ps-id>-domain.yml`
  - Framework file gets: `bind-*`, `tls-*`, `cors-*`, `otlp-*`, `log-level`, `database-url`,
    session algorithms, `enable-dynamic-registration`
  - Domain file gets: everything else (PS-ID-specific business logic settings)

  **Decision 14 — secrets file uniformity** (quizme-v2 Q7):
  - All PS-ID secrets directories must match template structure from Decision 14

  **Decision 17 — stale postgres-url.secret hostname fix** (quizme-v3 Q8):
  - Fix all stale postgres-url.secret values to use `shared-postgres-leader:5432`
    (see Task 1.16 for details)

  **Registry update — registry.yaml entrypoint field**:
  - Update all 10 PS-ID `entrypoint` fields in `api/cryptosuite-registry/registry.yaml`
    from `["/sbin/tini", "--", "/app/<ps-id>"]` to `["/sbin/tini", "--"]` (Decision 8)
  - Add `init_ps_id` field to each product in registry.yaml (used for `__PRODUCT_INIT_PS_ID__`):
    sm→sm-kms, jose→jose-ja, pki→pki-ca, identity→identity-authz, skeleton→skeleton-template
  - Add `init_ps_id` field to suite in registry.yaml (used for `__SUITE_INIT_PS_ID__`): sm-kms

  **⚠️ NOTE**: This task modifies actual deployment files, not templates. Template files define
  what actual files SHOULD look like; this task brings actual files into compliance.
- **Acceptance Criteria**:
  - [ ] All 5 product compose files have `pki-init` with `["/app/<ps-id>", "init", "--domain=<product>", "--output-dir=/certs"]`
  - [ ] All 10 PS-ID compose files have `pki-init` with `["/app/<ps-id>", "init", "--domain=<ps-id>", "--output-dir=/certs"]`
  - [ ] All 10 PS-ID compose files use shell-form command with `$SUITE_ARGS` (Decision 8)
  - [ ] All 10 PS-ID Dockerfiles have ENTRYPOINT `["/sbin/tini", "--"]` (NOT including app binary)
  - [ ] SM product compose has NO `image:` on service overrides
  - [ ] All 5 product compose files reference all 4 postgres secrets
  - [ ] All 10 PS-ID deployment configs renamed to `framework` pattern (Decision 16)
  - [ ] Domain deployment config files created (initially empty except pki-ca)
  - [ ] All 10 standalone configs split into framework + domain files
  - [ ] All secrets directories conform to Decision 14 template
  - [ ] All stale postgres-url.secret values fixed (Decision 17)
  - [ ] `api/cryptosuite-registry/registry.yaml` entrypoint fields updated to `["/sbin/tini", "--"]`
  - [ ] `api/cryptosuite-registry/registry.yaml` has `init_ps_id` for each product and suite
  - [ ] `docker compose -f deployments/<ps-id>/compose.yml config` succeeds for all 10 PS-IDs
- **Evidence**:
  - `test-output/framework-v10/phase1/deployment-fixes.log`

### Task 1.9: Create PS-ID secrets directory templates

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: Task 1.2
- **Description**: Create parameterized secrets template files in
  `api/cryptosuite-registry/templates/deployments/__PS_ID__/secrets/`.
  Content follows Decision 14 (quizme-v2 Q7): 14 `.secret` files with `__PS_ID__`-prefixed
  values and `BASE64_CHAR43` placeholder for secret content.

  Files to create (14 files):
  - `unseal-1of5.secret` through `unseal-5of5.secret` (5 files)
  - `hash-pepper-v3.secret`
  - `postgres-url.secret`, `postgres-username.secret`, `postgres-password.secret`,
    `postgres-database.secret` (4 files)
  - `browser-session-secret.secret`, `browser-csrf-secret.secret`
  - `service-session-secret.secret`
  - `issuing-ca-key.secret`

  Each file's value uses format: `__PS_ID__-<purpose>-<detail>-BASE64_CHAR43`
  (e.g., `__PS_ID__-unseal-key-1-of-5-BASE64_CHAR43`).
  `postgres-url.secret` value: `postgres://__PS_ID___database_user:BASE64_CHAR43@shared-postgres-leader:5432/__PS_ID___database?sslmode=disable`
- **Acceptance Criteria**:
  - [ ] 14 `.secret` files created in `templates/deployments/__PS_ID__/secrets/`
  - [ ] All files use `__PS_ID__` prefix in content values
  - [ ] All secret content uses `BASE64_CHAR43` placeholder (no real secrets)
  - [ ] `postgres-url.secret` references `shared-postgres-leader:5432` (not per-PS-ID postgres)
- **Files**:
  - `api/cryptosuite-registry/templates/deployments/__PS_ID__/secrets/*.secret` (14 files CREATE)

### Task 1.10: Create product and suite secrets directory templates

- **Status**: ✅
- **Estimated**: 0.5h
- **Dependencies**: Task 1.9
- **Description**: Create secrets templates for product-level and suite-level deployments.

  Product level (`templates/deployments/__PRODUCT__/secrets/`):
  - Same 14-file structure as PS-ID but with `__PRODUCT__`-prefixed values
  - Only unseal, hash-pepper, and postgres secrets are real; browser/service/issuing-ca use
    `.secret.never` marker extension (product level doesn't serve requests directly)

  Suite level (`templates/deployments/__SUITE__/secrets/`):
  - Same pattern with `__SUITE__`-prefixed values and `.secret.never` markers
- **Acceptance Criteria**:
  - [ ] Product secrets template directory created with correct files
  - [ ] Suite secrets template directory created with correct files
  - [ ] `.secret.never` marker files used for browser/service/issuing-ca at product/suite level
  - [ ] `__PRODUCT__` and `__SUITE__` placeholders in content values
- **Files**:
  - `api/cryptosuite-registry/templates/deployments/__PRODUCT__/secrets/*` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/__SUITE__/secrets/*` (CREATE)

### Task 1.11: Create shared-postgres template files

- **Status**: ✅
- **Estimated**: 1.5h
- **Dependencies**: Task 1.2
- **Description**: Create template files for the shared-postgres infrastructure service
  (Decision 10, quizme-v2 Q3, quizme-v3 Q3). These use `__SUITE__` substitution in content.

  **Architecture (resolved quizme-v3 Q3, quizme-v4 Q4/Q5)**:
  - **Container topology**: Exactly TWO containers — ONE `postgres-leader` container and
    ONE `postgres-follower` container. ALL databases below are LOGICAL databases within
    their respective single PostgreSQL container (NOT separate containers per database).
  - **30 leader databases**: 3 tiers (PS-ID, PRODUCT, SUITE) × 10 PS-IDs (all LOGICAL)
  - **16 follower databases**: 1 SUITE-level follower (replicates all 10) +
    5 PRODUCT-level followers (one per product, replicates that product's PS-IDs)
    - 10 PS-ID-level followers (all LOGICAL)
  - **Admin user model**: 1 admin user per leader DB (e.g., `sm_kms_database_user`).
    Follower DBs reuse the same admin users as their corresponding leaders.
    No DDL/DML user separation — single admin user per DB simplifies management.
  - **PostgreSQL conf files**: `postgresql-leader.conf` and `postgresql-follower.conf`
    with appropriate tuning parameters and replication settings.
  - **Template parameterization (quizme-v4 Q5)**: ALL template files — including conf files,
    SQL scripts, shell scripts — MUST use ALL applicable `__KEY__` placeholders. NO EXCEPTIONS.

  Files to create in `api/cryptosuite-registry/templates/deployments/shared-postgres/`:
  - `compose.yml` — leader + follower PostgreSQL containers
  - `postgresql-leader.conf` — leader-specific PostgreSQL configuration
  - `postgresql-follower.conf` — follower-specific PostgreSQL configuration
  - `init-databases.sql` — CREATE DATABASE statements for all 30 leader databases
  - `init-users.sql` — CREATE USER/ROLE statements (1 admin per leader DB)
  - `setup-logical-replication.sh` — replication config for ALL 10 PS-IDs across all 3 tiers
- **Acceptance Criteria**:
  - [ ] 6 template files created in `templates/deployments/shared-postgres/`
  - [ ] `init-databases.sql` covers all 30 leader databases (3 tiers × 10 PS-IDs)
  - [ ] `init-users.sql` creates 1 admin user per leader DB (no DDL/DML separation)
  - [ ] `setup-logical-replication.sh` covers all 10 PS-IDs and all 3 tiers
  - [ ] `postgresql-leader.conf` and `postgresql-follower.conf` included
  - [ ] Content uses `__SUITE__` substitution where appropriate
- **Files**:
  - `api/cryptosuite-registry/templates/deployments/shared-postgres/compose.yml` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/shared-postgres/postgresql-leader.conf` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/shared-postgres/postgresql-follower.conf` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/shared-postgres/init-databases.sql` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/shared-postgres/init-users.sql` (CREATE)
  - `api/cryptosuite-registry/templates/deployments/shared-postgres/setup-logical-replication.sh` (CREATE)

### Task 1.12: Create shared-postgres secrets templates

- **Status**: ✅
- **Estimated**: 0.25h
- **Dependencies**: Task 1.11
- **Description**: Create secrets templates for shared-postgres in
  `api/cryptosuite-registry/templates/deployments/shared-postgres/secrets/`.
  Includes postgres admin credentials and replication credentials.
- **Acceptance Criteria**:
  - [ ] Secrets template files created for shared-postgres
  - [ ] Uses `BASE64_CHAR43` placeholder for secret content
  - [ ] Admin and replication credentials included
- **Files**:
  - `api/cryptosuite-registry/templates/deployments/shared-postgres/secrets/*` (CREATE)

### Task 1.13: Split all 10 actual standalone configs into framework/domain

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 1.3 (template defines framework settings)
- **Description**: Split each of the 10 actual `configs/<ps-id>/<ps-id>.yml` files into
  `<ps-id>-framework.yml` + `<ps-id>-domain.yml` per Decision 7 (quizme-v2 Q1).

  Framework settings (same across all PS-IDs, templatable):
  `bind-*`, `tls-*`, `cors-*`, `otlp-*`, `log-level`, `database-url`,
  session algorithms, `enable-dynamic-registration`

  Domain settings (PS-ID-specific, NOT templated):
  Everything else — `issuer`, `token-lifetime`, `ca:` block (pki-ca),
  `authz-server-url` (identity-rp), etc.

  **Edge cases resolved (quizme-v3 Q4)**:
  - pki-ca `storage.type`: **REMOVED** — framework abstracts storage backend differences.
    Not a valid config key.
  - identity-rp `authz-server-url`, `client-id`, `redirect-uri`: **DOMAIN** — these are
    identity-rp-specific OIDC configuration, not generic framework connectivity.
  - identity-spa `static-files-path`: **FRAMEWORK** — any PS-ID may optionally support
    static file serving. This is a framework gap to be addressed (add to framework
    config schema `validate_schema.go` and pflag registration in v10 scope — quizme-v4 Q6/A).
- **Acceptance Criteria**:
  - [ ] All 10 PS-IDs have `<ps-id>-framework.yml` + `<ps-id>-domain.yml`
  - [ ] Original `<ps-id>.yml` removed (breaking change, acceptable per Decision 7)
  - [ ] Framework files contain ONLY framework settings
  - [ ] Domain files contain ONLY domain settings
  - [ ] pki-ca has NO `storage.type` in either file (removed)
  - [ ] identity-rp `authz-server-url`, `client-id`, `redirect-uri` in domain file
  - [ ] identity-spa `static-files-path` in framework file (gap: add to framework schema)
  - [ ] `go build ./...` succeeds (config loading must support new file names)
- **Files**:
  - `configs/<ps-id>/<ps-id>-framework.yml` × 10 (CREATE)
  - `configs/<ps-id>/<ps-id>-domain.yml` × 10 (CREATE)
  - `configs/<ps-id>/<ps-id>.yml` × 10 (DELETE)

### Task 1.14: Add shared_services and infra_tools to registry.yaml

- **Status**: ✅
- **Estimated**: 0.5h
- **Dependencies**: None
- **Description**: Add `shared_services` and `infra_tools` sections to
  `api/cryptosuite-registry/registry.yaml` per Decision 12 (quizme-v2 Q4).

  `shared_services`:
  - `shared-postgres` with leader/follower ports, database count, replication config

  `infra_tools`:
  - `shared-telemetry` with collector ports (4317/4318), Grafana port (3000)

  These entries model all parts of the repository, even if not all are PS-IDs.
  The linter uses these for template expansion of static-path templates.
- **Acceptance Criteria**:
  - [ ] `shared_services` section added with `shared-postgres` entry
  - [ ] `infra_tools` section added with `shared-telemetry` entry
  - [ ] Existing PS-ID entries unchanged
  - [ ] YAML is valid: `python -c "import yaml; yaml.safe_load(open('api/cryptosuite-registry/registry.yaml'))"`
- **Files**:
  - `api/cryptosuite-registry/registry.yaml` (MODIFY)

### Task 1.15: Rewrite pki-init init.go with pflag

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: None
- **Description**: Rewrite `internal/apps/framework/tls/init.go` to use pflag for argument
  parsing instead of positional args (Decision 15, quizme-v3 Q1).

  **Current**: `init.go` expects 2 positional args: `<tier-id> <target-dir>`
  (e.g., `init sm-kms /certs`).

  **New**: pflag-based with `--domain=<id>` and `--output-dir=<dir>` flags.
  Compose services call: `["init", "--domain=sm-kms", "--output-dir=/certs"]`.

  This resolves the compose/code mismatch where compose was passing `["init", "--output-dir=/certs"]`
  but init.go expected positional args.

  Implementation:
  - Replace `os.Args` positional parsing with pflag `StringVar`
  - `--domain` (required): tier identifier (PS-ID, product, or suite name)
  - `--output-dir` (required): certificate output directory
  - Error on unknown flags, missing required flags
  - Keep all existing TLS certificate generation logic unchanged
- **Acceptance Criteria**:
  - [x] `init.go` uses pflag for `--domain` and `--output-dir` (no positional args)
  - [x] `go build ./...` succeeds
  - [x] Existing tests updated to use new flag format
  - [x] Tests ≥95% coverage for init.go
  - [x] `golangci-lint run` clean
- **Files**:
  - `internal/apps/framework/tls/init.go` (MODIFY)
  - `internal/apps/framework/tls/init_test.go` (MODIFY)

### Task 1.16: Fix stale postgres-url.secret hostnames (Decision 17)

- **Status**: ✅
- **Estimated**: 0.5h
- **Actual**: 0.25h
- **Dependencies**: None
- **Description**: Fix all stale `postgres-url.secret` values across all 10 PS-ID deployment
  secret directories. Some files reference per-PS-ID postgres hostnames (e.g.,
  `sm-kms-postgres:5432`) that were removed when shared-postgres was introduced—they should
  reference `shared-postgres-leader:5432` instead (Decision 17, quizme-v3 Q8).

  Also verify ALL documentation uses the correct PostgreSQL leader hostname:
  - `docs/ENG-HANDBOOK.md` shared-postgres sections
  - `docs/deployment-templates.md` compose examples
  - `docs/tls-structure.md` if it references postgres
- **Acceptance Criteria**:
  - [x] All 10 `deployments/<ps-id>/secrets/postgres-url.secret` files use `shared-postgres-leader:5432`
  - [x] No stale per-PS-ID postgres hostnames in any secrets file
  - [x] Documentation references verified for correct hostname
- **Files**:
  - `deployments/*/secrets/postgres-url.secret` × 16 (MODIFY — 10 PS-ID + 5 product + 1 suite)
  - `docs/ENG-HANDBOOK.md` (MODIFY — 3 hostname references fixed)

### Task 1.17: Create domain deployment config files (Decision 16)

- **Status**: ✅
- **Estimated**: 0.5h
- **Actual**: 0.15h
- **Dependencies**: Task 1.3 (framework config templates exist)
- **Description**: Create domain deployment config files for all 10 PS-IDs (Decision 16,
  quizme-v3 Q5). These are NOT templated — they are per-PS-ID and initially empty.
  The file structure mirrors the framework deployment configs:
  `<ps-id>-app-domain-common.yml`, `<ps-id>-app-domain-sqlite-1.yml`,
  `<ps-id>-app-domain-sqlite-2.yml`, `<ps-id>-app-domain-postgresql-1.yml`,
  `<ps-id>-app-domain-postgresql-2.yml`.

  All files start empty (YAML comment header only) except:
  - pki-ca: `pki-ca-app-domain-common.yml` includes `crl-directory` setting
- **Acceptance Criteria**:
  - [x] 50 domain deployment config files created (10 PS-IDs × 5 variants)
  - [x] All files initially empty (YAML comment header only) except pki-ca
  - [x] pki-ca domain common config includes `crl-directory`
  - [x] File naming follows `<ps-id>-app-domain-<variant>.yml` pattern
- **Files**:
  - `deployments/<ps-id>/config/<ps-id>-app-domain-*.yml` × 50 (CREATE)

---

## Phase 2: Rewrite Template Linter

**Phase Objective**: Rewrite `template_drift.go` to read templates from disk at runtime using
`os.WalkDir`, build an in-memory expected filesystem, and compare against actual `./deployments/`
and `./configs/` directories. Replace all per-file `Check*` functions with a single
`CheckTemplateCompliance` linter.

### Task 2.1: Rewrite core engine in `template_drift.go`

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Phase 1 complete
- **Description**: Rewrite `template_drift.go` to implement the new in-memory FS approach:

  1. `LoadTemplatesDir(projectRoot string) (map[string]string, error)` — walks
     `{projectRoot}/api/cryptosuite-registry/templates/` using `os.WalkDir`, returns map
     of template-relative path → raw file content (e.g., `"deployments/__PS_ID__/Dockerfile"` → content)

  2. `BuildExpectedFS(templates map[string]string, registry *Registry) (map[string]string, error)` —
     for each template path in the map, detect the expansion key:
     - `__PS_ID__` in path: iterate over all 10 PS-IDs from registry; substitute `__PS_ID__` in
       BOTH the path AND call `buildParams(psID)` for full content substitution
     - `__PRODUCT__` in path: iterate over all 5 products from registry; substitute `__PRODUCT__`
       in path AND call `buildProductParams(product)` for content (includes `__PRODUCT_INCLUDE_LIST__`
       computed from registry PS-IDs for that product)
     - `__SUITE__` in path: iterate over suite name(s) from registry (currently just `cryptoutil`);
       substitute `__SUITE__` in path AND call `buildSuiteParams()` for content
     - `__INFRA_TOOL__` in path: iterate over infra tools from registry; substitute accordingly
     - No expansion key: substitute generic params in content; use template-relative path directly

  3. `CompareExpectedFS(expected map[string]string, projectRoot string) error` — for each entry
     in expected FS: resolve `{projectRoot}/{resolvedPath}`, read actual file, compare after
     normalization; collect all diffs; return aggregated error listing all mismatches

  4. REMOVE: `//go:embed templates/*`, `var templatesFS embed.FS`, `instantiate()` function,
     all individual `Check*` per-file wrappers

  5. KEEP: `buildParams(psID string) map[string]string`, `normalizeCommentAlignment`,
     `normalizeLineEndings`

  6. ADD: `buildProductParams(product string, registry *Registry) map[string]string` —
     builds product-level substitution map including `__PRODUCT_INCLUDE_LIST__`
     (multi-line include entries for each of that product's PS-IDs)

  7. ADD: `buildSuiteParams(registry *Registry) map[string]string` —
     builds suite-level substitution map (`__SUITE__`, display names, etc.)
- **Acceptance Criteria**:
  - [ ] `template_drift.go` has NO `//go:embed` directive, NO `embed.FS`, NO `embed` import
  - [ ] `LoadTemplatesDir` correctly discovers all ~63 template files
  - [ ] `BuildExpectedFS` correctly expands `__PS_ID__` templates × 10: 70 deployment files + 10 config files
  - [ ] `BuildExpectedFS` correctly expands `__PRODUCT__` template × 5 product compose files
  - [ ] `BuildExpectedFS` correctly expands `__SUITE__` templates × 1 (Dockerfile + compose.yml)
  - [ ] `BuildExpectedFS` correctly handles shared-telemetry static templates (no expansion, `__SUITE__` content sub)
  - [ ] `CompareExpectedFS` returns aggregated error with description for each mismatch
  - [ ] `go build ./...` succeeds
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/template_drift.go` (REWRITE)

### Task 2.2: Implement `CheckTemplateCompliance` single linter in `checks.go`

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 2.1
- **Description**: Rewrite `checks.go` to remove all per-file `Check*` functions and add a single
  `CheckTemplateCompliance(logger *Logger) error`. Use seam injection for testing:

  ```
  type templateComplianceFn func(projectRoot string) (map[string]string, error)
  func CheckTemplateCompliance(logger *Logger) error {
      return checkTemplateComplianceInDir(logger, ".", defaultComplianceFn)
  }
  func checkTemplateComplianceInDir(logger *Logger, projectRoot string, fn templateComplianceFn) error {
      ...
  }
  ```

  `defaultComplianceFn` calls `LoadTemplatesDir → BuildExpectedFS → CompareExpectedFS`.
- **Acceptance Criteria**:
  - [ ] All old `Check*` functions removed from `checks.go`
  - [ ] `CheckTemplateCompliance` added
  - [ ] `checkTemplateComplianceInDir` is the seam-injectable private function
  - [ ] `go build ./...` succeeds
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/checks.go` (REWRITE)

### Task 2.3: Update `lint-fitness-registry.yaml` and dispatch

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 2.2
- **Description**: Replace all per-file template linter entries with a single `template-compliance`
  entry. Update dispatch in `lint_fitness.go` to wire `CheckTemplateCompliance`.
- **Acceptance Criteria**:
  - [ ] All old per-file entries removed (`template-dockerfile`, `template-compose`, etc.)
  - [ ] Single `template-compliance` entry added, wired to `CheckTemplateCompliance`
  - [ ] `go run ./cmd/cicd-lint lint-fitness` runs without unknown command errors
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/lint-fitness-registry.yaml` (MODIFY)
  - `internal/apps/tools/cicd_lint/lint_fitness/lint_fitness.go` (MODIFY)

### Task 2.4: Write tests (≥98% coverage, seam injection pattern)

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Task 2.2
- **Description**: Write comprehensive tests using `t.TempDir()` for file system test isolation.
  All tests must call `t.Parallel()`. Use table-driven pattern for multi-variant cases.

  Test cases needed:
  - `TestLoadTemplatesDir_Happy`: temp dir with a few template files → correct map returned
  - `TestLoadTemplatesDir_NonExistentRoot`: error returned for missing directory
  - `TestBuildExpectedFS_PSIDExpansion`: 1 template with `__PS_ID__` path → N expansions
  - `TestBuildExpectedFS_StaticPath`: 1 static template → 1 entry without expansion
  - `TestBuildExpectedFS_ContentSubstitution`: verify `__PS_ID__` substituted in content too
  - `TestCompareExpectedFS_AllMatch`: expected FS matches temp dir → nil error
  - `TestCompareExpectedFS_ContentMismatch`: one file has wrong content → error with diff description
  - `TestCompareExpectedFS_MissingFile`: expected file does not exist → error
  - `TestCheckTemplateComplianceInDir_Success`: inject fn returning valid expected FS → passes
  - `TestCheckTemplateComplianceInDir_ComplianceError`: inject fn returning mismatch → error
  - `TestCheckTemplateComplianceInDir_LoadError`: inject fn returning error → propagated
  - `TestBuildExpectedFS_SecretsExpansion`: secrets template with `BASE64_CHAR43` →
    verify `BASE64_CHAR43` is NOT substituted (it's a length-check placeholder, not a value)
  - `TestCompareExpectedFS_SecretsPlaceholder`: verify `BASE64_CHAR43` placeholder comparison
    uses length-based matching, not exact byte comparison
- **Acceptance Criteria**:
  - [ ] All test functions have `t.Parallel()` at top
  - [ ] Table-driven for multi-variant tests
  - [ ] `go test ./internal/apps/tools/cicd_lint/lint_fitness/template_drift/...` → 100% pass
  - [ ] Coverage ≥98% for `template_drift` package
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/template_drift/template_drift_test.go` (MODIFY/ADD)

### Task 2.5: Fix actual deployment/config files if templates reveal drift

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 2.3
- **Description**: Run `go run ./cmd/cicd-lint lint-fitness` to detect any drift between the new
  templates and the actual `./deployments/` and `./configs/` files. Templates ARE the canonical
  spec — if a file on disk differs from the template-expanded expected content, fix the file on
  disk (not the template).
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes with zero template-compliance violations
- **Evidence**:
  - `test-output/framework-v10/phase2/lint-fitness-output.txt`

### Task 2.6: Create secrets-compliance sublinter

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 1.9 (secrets templates exist)
- **Description**: Create a new lint-fitness sublinter that validates secrets directories
  match the template structure from Decision 14. The linter checks:

  1. All 14 expected `.secret` files exist in each PS-ID secrets directory
  2. Secret values use correct PS-ID prefix format
  3. `BASE64_CHAR43` placeholder positions contain values ≥43 characters (length check only)
  4. No extra unexpected `.secret` files
  5. Product/suite levels use `.secret.never` markers where required

  Uses seam injection pattern for testability.
- **Acceptance Criteria**:
  - [ ] `secrets-compliance` linter registered in `lint-fitness-registry.yaml`
  - [ ] Catches missing secrets files, wrong prefixes, too-short secrets
  - [ ] Allows extra non-.secret files (e.g., README)
  - [ ] ≥98% test coverage
  - [ ] `go build ./...` succeeds
- **Files**:
  - `internal/apps/tools/cicd_lint/lint_fitness/secrets_compliance/secrets_compliance.go` (CREATE)
  - `internal/apps/tools/cicd_lint/lint_fitness/secrets_compliance/secrets_compliance_test.go` (CREATE)

---

## Phase 3: Update Documentation

**Phase Objective**: Ensure docs accurately describe the new parameterized template architecture.
Template files are the source of truth; docs describe them.

### Task 3.1: Update `deployment-templates.md` template content sections

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 1 complete
- **Description**: For Sections B.1, C.1, D.1-5, E, G.1, I, J: replace embedded template code
  blocks with a reference to the canonical file in `api/cryptosuite-registry/templates/{path}`.
  Keep all rule tables, parameter tables, and rationale text.
- **Acceptance Criteria**:
  - [ ] No embedded template code blocks remain in those sections
  - [ ] Each section references its canonical template file path
  - [ ] All rule tables and rationale text intact
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
- **Files**:
  - `docs/deployment-templates.md` (MODIFY)

### Task 3.2: Update Section O.2 template catalog

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 3.1
- **Description**: Update the template file catalog in Section O.2 to list all canonical
  template files with their parameterized paths, expansion behavior (PS-ID × 10, product × 5,
  suite × 1, shared-telemetry static, shared-postgres static, secrets), and the actual files
  they validate.
- **Acceptance Criteria**:
  - [ ] Section O.2 correctly lists all template files grouped by category
  - [ ] Expansion behavior documented for each file type
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
- **Files**:
  - `docs/deployment-templates.md` (MODIFY)

### Task 3.3: Update `target-structure.md`

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 3.2
- **Description**: Add `api/cryptosuite-registry/templates/` tree to `target-structure.md`
  listing all template files and the full directory structure including secrets subdirectories.
- **Acceptance Criteria**:
  - [ ] `target-structure.md` shows `api/cryptosuite-registry/templates/` with full subtree
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
- **Files**:
  - `docs/target-structure.md` (MODIFY)

### Task 3.4: Annotate v9 plan and tasks with correction note

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: None
- **Description**: Add a `⚠️ v9 IMPLEMENTATION ERROR` note to v9 Task 8.1 and v9 plan Item 23
  indicating templates were placed at wrong location; v10 corrects this.
- **Acceptance Criteria**:
  - [ ] `docs/framework-v9/plan.md` Item 23 has correction note
  - [ ] `docs/framework-v9/tasks.md` Task 8.1 has correction note
- **Files**:
  - `docs/framework-v9/plan.md` (MODIFY)
  - `docs/framework-v9/tasks.md` (MODIFY)

---

## Phase 4: Quality Gates

**Phase Objective**: Full validation that all quality gates pass end-to-end.

### Task 4.1: Full build and lint validation

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: All Phase 1–3 tasks
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean
  - [ ] `golangci-lint run` zero warnings/errors
  - [ ] `golangci-lint run --build-tags e2e,integration` zero warnings/errors
- **Evidence**: `test-output/framework-v10/phase4/build-lint.txt`

### Task 4.2: Full test suite and coverage

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 4.1
- **Acceptance Criteria**:
  - [ ] `go test ./...` 100% pass, zero skips
  - [ ] `go test -race -count=2 ./internal/apps/tools/...` race-free
  - [ ] ≥98% coverage for `template_drift` package
- **Evidence**:
  - `test-output/framework-v10/phase4/test-results.txt`
  - `test-output/framework-v10/phase4/coverage-template-drift.out`

### Task 4.3: cicd-lint end-to-end validation

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 4.1
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes (template-compliance green)
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
- **Evidence**: `test-output/framework-v10/phase4/cicd-lint-output.txt`

---

## Phase 5: Knowledge Propagation

**Phase Objective**: Apply lessons learned. NEVER skip.

### Task 5.1: Update ENG-HANDBOOK.md

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: All Phase 1–4 tasks
- **Acceptance Criteria**:
  - [ ] Section 9.11.1 fitness linter catalog: old per-file entries removed, `template-compliance` added
  - [ ] Section 13.6 describes parameterized template directory structure and `os.WalkDir` approach
  - [ ] `go run ./cmd/cicd-lint lint-docs validate-propagation` passes

### Task 5.2: Update instruction files if new patterns discovered

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 5.1
- **Acceptance Criteria**:
  - [ ] Any new patterns documented in relevant instruction files
  - [ ] `go run ./cmd/cicd-lint lint-docs validate-propagation` passes

### Task 5.3: Final clean commit

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Task 5.2
- **Acceptance Criteria**:
  - [ ] All changes committed with conventional commit messages (one commit per semantic group)
  - [ ] `git status --porcelain` returns empty
  - [ ] Final `go run ./cmd/cicd-lint lint-fitness` passes

---

## Cross-Cutting Tasks

### Code Quality (enforced per task)

- [ ] No new `//nolint:` directives without GitHub issue reference
- [ ] All magic constants in `internal/shared/magic/` (no new magic literals in Go code)
- [ ] `gofumpt -w .` before committing any Go file
- [ ] No `.go` files added to `api/cryptosuite-registry/` (CRITICAL — it MUST NOT be a Go package)

### Testing (enforced per task)

- [ ] `t.Parallel()` on all new tests and subtests
- [ ] Seam injection pattern for all new linter functions
- [ ] Table-driven for multi-variant tests
- [ ] `t.TempDir()` for file system test isolation
- [ ] ≥98% line coverage for `template_drift` package

---

## Notes

**Critical constraint**: `api/cryptosuite-registry/` MUST NOT contain any `.go` files.
The templates are plain configuration files, not Go source code. Any `.go` file discovered
there during implementation is an error and must be removed immediately.

**Expansion rules**:
- `__PS_ID__` in path → expand for all 10 PS-IDs from `registry.yaml`
- `__PRODUCT__` in path → expand for all 5 products from `registry.yaml`
- `__SUITE__` in path → expand for suite name(s) from `registry.yaml` (currently 1: `cryptoutil`)
- No expansion key in path → static path, content-only substitution (e.g., shared-telemetry)

**Config file naming**: Actual deployment config files use PS-ID prefix with `framework`
qualifier: `sm-kms-app-framework-common.yml` (Decision 16). Template files use
`__PS_ID__-app-framework-common.yml` with the prefix parameterized. Domain deployment configs
use `<ps-id>-app-domain-<variant>.yml` and are NOT templated.

**ENTRYPOINT change (quizme-v4 Q1)**: ALL PS-ID Dockerfiles change from
`ENTRYPOINT ["/sbin/tini", "--", "/app/<ps-id>"]` to `ENTRYPOINT ["/sbin/tini", "--"]`.
App binary moves into compose command. Tini already installed via `apk --no-cache add tini`.

**Config merge order (quizme-v4 Q3)**: Framework `config_parse.go` uses `ReadInConfig` for
first file, `MergeInConfig` for subsequent. Later files override earlier for same keys.
Shell-form command specifies all config files in precedence order (lowest first, highest last).

**Docker Compose profiles BANNED (Decision 18)**: NEVER use Docker Compose `profiles:` feature.
Use service-name override instead (product compose redefines PS-ID `pki-init` service).

**Container topology (quizme-v4 Q4)**: shared-postgres = 2 containers (leader + follower).
ALL 30 leader databases and 16 follower databases are LOGICAL databases within PostgreSQL,
NOT separate containers.

**Template parameterization invariant (quizme-v4 Q5)**: ALL template files use ALL applicable
`__KEY__` placeholders. NO EXCEPTIONS. Never ask whether to parameterize a template file.

**Total template files**: ~63 physical files → ~381 expected files after full expansion
(7 PS-ID deployment templates × 10 = 70, + 14 PS-ID secrets × 10 = 140, + 1 PS-ID config × 10 = 10,
- 1 product compose × 5 = 5, + product secrets × 5, + 2 suite files × 1 = 2, + suite secrets × 1,
- 2 shared-telemetry static = 2, + 6 shared-postgres static = 6, + shared-postgres secrets)
Plus ~50 non-template domain deployment config files and ~10 non-template domain standalone
config files.

---

## Evidence Archive

- `test-output/framework-v10/phase2/` — lint-fitness output after Phase 2
- `test-output/framework-v10/phase4/` — full quality gate evidence
