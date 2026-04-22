# Lessons - Framework V15: Pre-Flight Gap Fixes + OTel/Grafana mTLS + Public App TLS Trust

**Created**: 2026-04-22
**Last Updated**: 2026-04-21

---

> **MANDATORY per-phase structure** (fill during execution, not planning):
>
> **What Worked** — patterns, approaches, and decisions that proved effective
>
> **What Didn't Work** — approaches tried that failed, wrong assumptions, pitfalls
>
> **Root Causes** — underlying causes for failures; root cause analysis, not symptoms
>
> **Patterns for Future Phases** — lessons to carry forward into the next phase and into
> permanent artifacts (ENG-HANDBOOK.md, agents, skills, instructions)

---

## Phase 0: Pre-Flight Gap Fixes

### What Worked

- **Batch signal-handling fix**: Adding `close(sigChan)` after `signal.Stop(sigChan)` across all
  10 service entry points in a single commit was fast and zero-risk — the pattern is identical in
  every file, making batch editing the right approach.
- **Reference implementation lookup**: `sm-im/im.go` had the correct shutdown timeout pattern.
  Reading it first before editing `sm-kms/kms.go` gave immediate confidence — no guessing.
- **`MustStartAndWaitForDualPorts` helper**: The helper existed and was well-documented. Replacing
  pki-ca's 300-attempt polling loop reduced testmain_test.go by ~20 lines with zero behavior change.
- **`lint-docs` and `lint-deployments` already present**: Task 0.1 only needed the top-level
  permissions block — both CI steps were already wired. Investigation saved from unnecessary work.
- **gofumpt auto-fix**: After adding the `e2e_helpers` import, running `gofumpt -w` auto-fixed
  the import ordering without manual intervention.

### What Didn't Work

- **Task 0.5 scope underestimate**: The plan estimated 1.5h for usage.go deduplication. The actual
  scope (8 `const` blocks → `var`, new shared package, 7 files across 4 product trees) revealed a
  `const`→`var` conversion that introduces non-obvious linter surface. Deferred to V16.
- **V13 stale content in lessons.md**: The lessons.md file had a stale V13 phase list appended
  after the V15 content. Root cause: the file was not cleaned up when V15 planning was created.

### Root Causes

- **`sm-kms` shutdown timeout missing**: The `sm-kms` entry point was created by copying `sm-im`
  but the shutdown block was simplified without carrying forward the `context.WithTimeout` pattern.
  Pattern drift between near-identical files is the recurring cause.
- **`close(sigChan)` missing in all 10 entry points**: The canonical pattern was never established
  in a shared location or checked by a fitness linter. Each file evolved independently.
- **`continue-on-error: true` on coverage gates**: These were added as temporary suppressors during
  initial CI setup but never removed after coverage targets were met. Suppressor debt accumulates
  when there's no automated check that they're removed.
- **`pull-requests: write` over-scope**: The workflow-level permission was copied from a template
  that needed PR comments and never scoped down when that feature was removed.

### Patterns for Future Phases

- **Fitness linter for signal handling pattern**: The `close(sigChan)` gap will recur as new
  services are added. Add a `lint-fitness` sub-linter checking that all `signal.Stop(sigChan)` are
  followed by `close(sigChan)` within 3 lines in service entry points.
- **Coverage gate audit**: Any `continue-on-error: true` on a step that ends with `exit 1` is a
  suppressor that MUST have a removal ticket. Add to Phase 12 knowledge propagation.
- **Shutdown pattern enforcement**: The shutdown pattern (`context.WithTimeout` + error log + cancel
  defer) should be extracted to a shared helper and linted for consistency.
- **`const` vs `var` for CLI strings**: Usage strings are pure data. The correct long-term approach
  is `var` initialized by a parameterized builder, but this is a breaking change for existing
  `const` usage. Introduce in V16 with a fitness linter to enforce the pattern going forward.

---

## Phase 1: pki-init Patch — Cat 2, Cat 3, Cat 4, Cat 8, Cat 9 app

### What Worked

- **Category-based generator pattern**: pki-init uses a structured generator that processes
  certificate categories by number (Cat 2, Cat 3, Cat 4, Cat 8, Cat 9). Adding new cert categories
  required only adding new `GenerateCert` calls with the appropriate inputs — the pattern was
  clean and self-contained.
- **14 cert categories proved sufficient**: The V15 cert hierarchy (server entity, client entity,
  issuing CA, truststore) maps cleanly onto the 14 category slots. No new abstractions needed.
- **`export_test.go` seam for generator tests**: All new generator tests use exported test seams
  (`ExportedNewTestGenerator`, `ExportedProductionNewGenerator`) rather than modifying production
  code, keeping test logic cleanly separated from production paths.
- **Table-driven category mapping tests**: The `TestCategoryMapping` table-driven pattern caught
  several CN template mismatches before any certs were generated.

### What Didn't Work

- **Initial CN template naming**: The first draft used `public-https-server-entity-sm-kms-postgres`
  as a CN for both postgres-1 and postgres-2. This had to be corrected to variant-specific CNs
  (`-postgres-1`, `-postgres-2`) to match what the app config templates reference.

### Root Causes

- **CN template drift between pki-init and deployment configs**: The cert CN is defined in pki-init
  but consumed in both the app TLS config and the test `ServerCertCN` magic constants. A naming
  drift would cause TLS verification failures that are silent in development but fatal in E2E.

### Patterns for Future Phases

- **CN constants in magic package**: The Cat 3 server cert CNs are now magic constants — use them
  in both pki-init tests and E2E tests so a CN rename is a single-point change.
- **All 14 cert categories documented with `// Cat N: <name>` comments**: Maintained in
  `generator.go` call sites so reviewers can cross-reference without mental mapping.

---

## Phase 2: OTel Collector Server TLS

### What Worked

- **OTel Collector `tls:` block pattern**: The `receivers.otlp.protocols.grpc.tls` and
  `receivers.otlp.protocols.http.tls` blocks accept `cert_file`, `key_file`, `ca_file`, and
  `client_ca_file`. Placing `client_ca_file` forces mTLS (RequireAndVerifyClientCert equivalent).
- **Cert paths via mounted bind volume**: OTel config files reference cert paths under `/certs/`,
  which compose mounts from `./certs:/certs:ro`. This pattern is consistent across all services.
- **40 deployment config files updated atomically**: Using `multi_replace_string_in_file` with up
  to 10 replacements per call kept the OTel config changes to a few tool invocations.

### What Didn't Work

- **`OTELCOL_EXTRA_ARGS` for Grafana integration**: The first approach tried setting Grafana's
  env vars directly. `OTELCOL_EXTRA_ARGS` (for passing flags to otelcol inside Grafana) is a
  cleaner integration point that avoids forking the Grafana LGTM image.

### Root Causes

- **OTel config schema is untyped YAML**: Mistakes in the `tls:` block (wrong key names, wrong
  indent) silently fail at runtime. The fix is always to check the official OTel Collector contrib
  docs for the exact key names.

### Patterns for Future Phases

- **OTel `client_ca_file` enables server-side mTLS**: The `client_ca_file` key under `tls:` is
  the mechanism for requiring client certs. Document this in §9.4 of ENG-HANDBOOK.md.
- **Canonical template sync MUST happen in the same commit**: If a deploy config changes, the
  canonical template in `api/cryptosuite-registry/templates/` changes in the same commit.

---

## Phase 3: App→OTel Client mTLS

### What Worked

- **Framework OTLP TLS config fields**: The framework config struct already had placeholder
  fields for `otlp-tls-cert-file`, `otlp-tls-key-file`, `otlp-tls-ca-file`. Adding Cat 9 cert
  paths to all 40 deployment configs required only editing the per-variant config files.
- **Cat 9 app cert path pattern**: `certs/{PS-ID}/otel-collector-contrib-https-client-entity-{PS-ID}-{variant}/`
  is consistent across all services. Path templating was easy once the pattern was clear.
- **Per-variant Cat 9 certs**: Each sm-kms variant (sqlite-1, sqlite-2, postgresql-1, postgresql-2)
  gets its own Cat 9 cert. The OTel Collector uses the Cat 8 CA to verify these at ingest.

### What Didn't Work

- **Config key name spelling**: The first draft had `otlp-client-cert-file` instead of
  `otlp-tls-cert-file`. The schema validator caught this immediately via `lint-deployments`.

### Root Causes

- **Config schema is the authority**: The framework's `validate_schema.go` defines the canonical
  key names. Always check it before editing config files to avoid schema validation failures.

### Patterns for Future Phases

- **`lint-deployments` as immediate feedback loop**: Run `go run ./cmd/cicd-lint lint-deployments`
  after every config file change. It catches key naming and structural issues in seconds.

---

## Phase 4: Verify OTel Standalone

### What Worked

- **Go `crypto/tls` dial pattern**: `tls.DialWithDialer` with a custom `tls.Config` (RootCAs +
  Certificates) is the canonical Go E2E TLS verification approach. No curl, no openssl.
- **Single TestMain approach (D8=E)**: Having all TLS E2E tests in one package with one TestMain
  avoids redundant compose stack startups. The same OTel + Grafana stack is shared by Phase 4,
  Phase 7, and Phase 11 tests.
- **`loadCACertPool` + `loadClientCert` helpers**: These two helpers cover all TLS configuration
  patterns needed across all E2E test files.
- **`waitForOtelHealth` time-based polling**: Using `time.Now().UTC().Before(deadline)` with
  `time.Sleep` is the correct pattern — avoids duration-multiplication anti-pattern.

### What Didn't Work

- **Port conflict discovered late**: OTel test-expose ports 14317/14318/14133 conflicted with
  Grafana's host-side ports 14317/14318 from shared-telemetry. This was only discovered in
  Phase 11 planning when both services were added to the same compose stack.

### Root Causes

- **Port conflicts when multiple services share a compose stack**: When Phase 4 was designed,
  only OTel was in scope. When Phase 11 added Grafana, the port overlap became blocking.
  Port assignments should be reviewed against the full service catalog at planning time.

### Patterns for Future Phases

- **Offset test-expose ports by +10000 from Grafana**: OTel test-expose now uses 24317/24318/24133
  (not 14317/14318/14133) to avoid Grafana overlap. Document the convention in service port tables.
- **Build tag `e2e` MUST be on every file in the package**: All helper files AND test files get
  `//go:build e2e` — the tag is package-wide for Docker-dependent tests.

---

## Phase 5: Grafana LGTM HTTPS + OTLP Ingest TLS

### What Worked

- **`grafana.ini` approach for Grafana HTTPS (D1=A)**: Mounting a custom `grafana.ini` with
  `[server] protocol = https`, `cert_file`, and `cert_key` is the standard way to enable HTTPS
  in Grafana. No custom image required.
- **`OTELCOL_EXTRA_ARGS` for Grafana's embedded OTel Collector (D6=A)**: Grafana LGTM embeds
  an `otelcol` binary. Passing `--config=file://<path>` via `OTELCOL_EXTRA_ARGS` overlays the
  custom OTel config with mTLS receiver settings onto the default config.
- **Cat 2 cert for both Grafana UI HTTPS and OTLP gRPC TLS**: A single Cat 2 cert covers both
  the Grafana web UI (:3000) and the OTel ingest endpoint (:4317/:4318). The same CA pool verifies
  both connections from the test client.

### What Didn't Work

- **Grafana LGTM image's embedded OTel config discovery**: The embedded otelcol in Grafana uses
  a non-standard config path. `OTELCOL_EXTRA_ARGS` with `--config=file://` was the correct
  workaround once the standard env vars were found not to override the embedded config.

### Root Causes

- **Grafana LGTM bundles its own OTel instance**: Unlike a standalone OTel Collector, the
  bundled otelcol is managed by the Grafana process. TLS config must go through the Grafana
  environment, not through a separate compose service.

### Patterns for Future Phases

- **Document `OTELCOL_EXTRA_ARGS` pattern in §9.4**: This is a non-obvious integration point
  specific to Grafana LGTM. New engineers will spend time rediscovering it without documentation.

---

## Phase 6: OTel→Grafana Client mTLS

### What Worked

- **Cat 9 infra cert in OTel exporter config**: The OTel Collector's `exporters.otlphttp.tls`
  or `exporters.otlp.tls` block uses `cert_file`/`key_file`/`ca_file` for client auth. Setting
  `cert_file` and `key_file` to the Cat 9 infra cert paths enables mTLS in the OTel→Grafana leg.
- **Shared Cat 1 CA pool for all server cert verification**: Both Grafana and OTel server certs
  are signed by the same Cat 1 issuing CA. The test client loads only one CA pool and verifies
  all server certs with it.

### What Didn't Work

- **OTel OTLP exporter endpoint port confusion**: The OTel exporter must target Grafana's container
  port 4317 (not the host-side 14317). Inside the Docker network, service names resolve to internal
  ports. The host-mapped port (14317) is only for test clients running outside Docker.

### Root Causes

- **Container-internal vs host-external port distinction**: The OTel compose config uses
  `endpoint: grafana-otel-lgtm:4317` (container port), while test clients use `127.0.0.1:14317`
  (host-mapped port). Mixing the two in the same config is a common source of connection errors.

### Patterns for Future Phases

- **Container endpoint = service-name:container-port, test endpoint = 127.0.0.1:host-port**:
  This distinction must be explicit in compose config comments. Added to deployment-templates.md.

---

## Phase 7: Verify OTel→Grafana Pipeline

### What Worked

- **`grafana_tls_e2e_test.go` reuses `loadCACertPool` and `loadClientCert`**: The Phase 4 helpers
  generalize across all TLS verification patterns. Grafana tests required zero new helper functions.
- **`TestGrafanaHTTPS_APIHealth` uses standard HTTPS client with Cat 1 CA pool**: Grafana's web
  UI health endpoint (`/api/health`) is an HTTPS endpoint — verified using the same TLS dial
  pattern as OTel. No special Grafana client required.
- **`TestGrafanaOTLP_GRPC_mTLS_Rejected` tests the failure path**: Connecting without a client
  cert to the Grafana OTLP gRPC endpoint must fail. This test validates Cat 8 enforcement.

### What Didn't Work

- Nothing significant. The Phase 4 patterns transferred cleanly to Phase 7.

### Root Causes

- N/A — Phase 7 implementation was smooth because Phase 4 established all necessary patterns.

### Patterns for Future Phases

- **All rejection tests must assert `err.Error()` contains `"tls"`**: Network errors (e.g.
  connection refused) look like TLS errors without this assertion. The assertion distinguishes a
  proper TLS rejection from a misconfigured service that isn't listening.

---

## Phase 8: Public PS-ID App Server TLS

### What Worked

- **Framework config fields `server-public-tls-cert-file`, `server-public-tls-key-file`,
  `server-public-tls-ca-file`**: These three fields in the framework config struct map directly
  to Cat 3 (server cert), Cat 3 key, and Cat 4 (client CA). Setting them enables
  `tls.RequireAndVerifyClientCert` on the public HTTPS listener.
- **`applyPublicMTLS` in `internal/apps/framework/service/server/listener/public.go`**: The
  framework already had the mTLS application logic. Phase 8 only required providing the cert paths
  in the deployment configs.
- **Per-variant Cat 3/4 certs**: Each sm-kms variant (sqlite-1, sqlite-2, postgresql-1/2) gets
  its own Cat 3 server cert. Postgres-1 and postgres-2 share a Cat 4 client CA
  (`public-https-client-issuing-ca-sm-kms-postgres`) since they belong to the same postgres tier.

### What Didn't Work

- **40-file config update was verbose**: Updating all 40 deployment configs across 10 PS-IDs
  required careful verification that the cert path patterns were correct per variant. A
  lint-deployments validator for cert path existence would have caught any typos immediately.

### Root Causes

- **Cat 4 CA scope decision**: Sharing one Cat 4 CA between postgres-1 and postgres-2 is
  correct (they're in the same trust domain — both are PostgreSQL-backed instances of the same
  service). SQLite variants each get their own Cat 4 CA for stricter isolation.

### Patterns for Future Phases

- **Cat 3/4 cert mount paths in compose.yml**: The `./certs:/certs:ro` bind mount must be present
  in all service containers for TLS cert access. This is a structural requirement in compose files.
- **`server-public-tls-ca-file` = Cat 4 (client CA)**: Not to be confused with Cat 1 (server CA).
  The naming distinction (server issuing CA vs client issuing CA) must be explicit in all config
  documentation.

---

## Phase 9: Deployment Templates

### What Worked

- **`lint-fitness` as the canonical validator for template compliance**: The fitness linter
  (`entity-registry-completeness`, `template-compliance`) catches any drift between canonical
  templates and deployed files before it reaches CI/CD. Running it after every template change is
  the correct workflow.
- **Canonical templates in `api/cryptosuite-registry/templates/`**: Updating the canonical
  templates and running `lint-fitness` in the same phase ensured no drift accumulated.

### What Didn't Work

- Nothing significant. Template compliance enforcement was already working from previous versions.

### Root Causes

- N/A — Phase 9 was a verification phase that required no new patterns.

### Patterns for Future Phases

- **Always run `go run ./cmd/cicd-lint lint-fitness` after any template change**: This is the
  single command that validates all canonical template compliance in one shot.

---

## Phase 10: Deployment Linting

### What Worked

- **54/54 validators passing in `lint-deployments`**: All 8 validators (naming, kebab-case, schema,
  template, ports, telemetry, admin, secrets) were green. Running `lint-deployments` after each
  config change batch kept the validation feedback loop tight.
- **Admin bind validator**: The validator correctly caught that admin endpoints must bind to
  `127.0.0.1:9090` (never `0.0.0.0:9090`). All configs were already compliant.

### What Didn't Work

- Nothing. All configs were already compliant from prior phases.

### Root Causes

- N/A — Phase 10 was a validation phase that confirmed prior work.

### Patterns for Future Phases

- **`lint-deployments` is a complete validation suite**: Use it as the post-phase gate for
  any work touching `deployments/` or `configs/`. It subsumes manual config review.

---

## Phase 11: Deployment Verification — Full Telemetry Stack

### What Worked

- **Port offset pattern (+10000 from Grafana)**: Changing OTel test-expose ports from
  14317/14318/14133 to 24317/24318/24133 resolved the Grafana port conflict cleanly. The
  `+10000` offset is easy to remember and avoids the standard OTLP port range.
- **Single TestMain extending to all pipeline services**: The D8=E decision (all TLS E2E tests
  in one package/TestMain) means Phase 11 tests required no new TestMain — only extending the
  `compose up` command to include the 5 additional services.
- **Table-driven `allAppVariants()`**: Centralizing the 4 app variant test cases in one function
  keeps `TestFullPipeline_AppPublicHTTPS_*` tests DRY. Adding a new variant requires only one
  change to `allAppVariants()`.
- **`waitForAppsHealthy` time-based polling**: Consistently using `time.Now().UTC().Before(deadline)`
  matches the pattern in `waitForGrafanaHealth` and `waitForOtelHealth`. The codebase is now
  consistent across all wait functions.

### What Didn't Work

- **Initial `full_pipeline_test.go` had double `package` declaration**: The file was created with
  a `package e2e` on line 1 before the build tag comment, causing a compile error. The fix was
  straightforward (move copyright header above build tag, remove duplicate package declaration).
- **`copyloopvar` auto-fix left blank lines**: `golangci-lint --fix` removed `v := v` captures
  but left blank lines after the for-range opening, causing `wsl_v5` lint failures. Required a
  manual second pass to remove the orphaned blank lines.

### Root Causes

- **`create_file` tool prepends `package` before file content**: When using the `create_file` tool,
  the tool adds a `package` declaration at the start before the copyright header, resulting in
  two `package` declarations. The fix is to ensure the copyright header is truly the first line
  when creating Go test files.
- **Auto-fix tools interact**: `golangci-lint --fix` runs multiple fixers in sequence. `copyloopvar`
  removes the variable capture line but the subsequent `wsl_v5` check runs against the modified
  file — the blank line left by the removal is only detected in the next `golangci-lint run` pass.

### Patterns for Future Phases

- **After `golangci-lint --fix`, always run `golangci-lint run` again**: Auto-fix may create new
  violations (e.g., orphaned blank lines after removed statements). Two-pass linting is standard.
- **Verify file header when creating Go test files**: Copyright comment must be BEFORE the build
  tag. Build tag must be BEFORE the package declaration. This is enforced by `gofumpt` but not
  checked by `create_file`.
- **Port conflict review at planning time**: When adding services to an existing compose test
  stack, check all host-side port bindings in all included compose files for overlap.

---

## Phase 12: Knowledge Propagation

### What Worked

- **Lessons.md as implementation log**: Recording lessons phase-by-phase throughout execution
  rather than retrospectively ensures accuracy. Filling placeholder sections with concrete facts
  (not generalities) gives future developers actionable patterns.
- **ENG-HANDBOOK.md §9.4 as the canonical OTel/Grafana mTLS reference**: Adding OTel receiver
  `tls:` config blocks, Grafana's `OTELCOL_EXTRA_ARGS` pattern, and the OTel→Grafana mTLS
  forwarding pattern to §9.4 ensures they're discoverable from the primary architecture reference.
- **deployment-templates.md cert mount table**: Documenting the combined V12+V15 cert mount
  least-privilege table in one place makes cert audit straightforward.

### What Didn't Work

- **Placeholder structure in lessons.md**: All phase sections were created as placeholders during
  planning and required filling during execution. The MANDATORY BLOCKING quality gate (per the
  implementation-execution agent spec) enforces this correctly — but it means lessons.md has high
  maintenance cost if phases are numerous.

### Root Causes

- **Lessons are only valuable if specific**: Generic lessons ("test your code") add no value.
  The quality gate requiring substantive content (not just the placeholder) forces specificity.

### Patterns for Future Phases

- **lessons.md structure is correct**: The 4-section format (What Worked, What Didn't Work,
  Root Causes, Patterns for Future Phases) consistently captures the right information.
- **Extract permanent knowledge to ENG-HANDBOOK.md after each plan**: The lessons in phases 1-11
  above contain patterns that belong in §9.4, §6, §10 of ENG-HANDBOOK.md. Always extract before
  the plan is archived.

