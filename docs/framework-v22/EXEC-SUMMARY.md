# EXEC-SUMMARY - framework-v22

Status: Complete.
Date: 2026-05-14
Reviewer model intent: objective post-implementation audit.

## Scope and Evidence

- Plan artifacts reconciled:
  - docs/framework-v22/plan.md
  - docs/framework-v22/tasks.md
  - docs/framework-v22/lessons.md
- Lessons inclusion status:
  - Included and reconciled: docs/framework-v22/lessons.md executive summary, per-phase root causes, and actions were reviewed and reflected in this EXEC-SUMMARY completion narrative.
- Validation evidence from current execution:
  - `go build ./...` (pass)
  - `go build -tags e2e,integration ./...` (pass)
  - `golangci-lint run` (pass, 0 issues)
  - `go test -tags integration ./...` (pass)
  - `go test -tags e2e ./internal/apps/sm-kms/e2e/... -v` (pass, archived)
  - `go test -tags e2e ./internal/apps/sm-im/e2e/... -v` (pass, archived)
  - `go run ./cmd/cicd-lint lint-deployments` (pass)
  - `go test ./internal/apps-framework/service/testing/e2e_infra/...` (pass)
  - `go test ./internal/apps-tools/cicd_lint/lint_fitness/template_drift -run TestCheckTemplateCompliance -v` (pass)
- Evidence archive:
  - `test-output/v22-e2e/sm-kms.log`
  - `test-output/v22-e2e/sm-im.log`

## Completion Validation

Overall finding: Framework-v22 implementation scope is complete with the previously blocked Docker/E2E phase now resolved.

Phase-level validation:

1. Phases 1-8: complete from existing implementation and prior phase evidence.
2. Phase 9: complete. Both `sm-kms` and `sm-im` E2E suites pass and are archived.
3. Phase 10: complete. Inventory artifact present with formula-derived total (54).
4. Phase 11: complete. Knowledge-propagation artifacts remain valid; final status reconciled.

Quality-gate validation:

1. Build gates pass (`go build ./...`, tagged build pass).
2. Lint gate passes (`golangci-lint run` = 0 issues).
3. Integration and E2E gates pass in current run.
4. Deployment and template compliance gates pass.
5. Remaining race detector check is CI-scoped by policy on Linux runners and is documented as such.

## Post-Implementation Issues

1. Windows bind-mounted cert directory write failures
- Symptoms: `pki-init` failed with permission denied when writing `/certs/tls-config.yml` and related cert artifacts.
- Root Cause: Read-only file attributes persisted on host bind-mounted `deployments/*/certs` contents.
- Fix: Added cert-directory sanitization and writable normalization in E2E compose startup path.

1. Compose startup failure leaked partial stack state on start errors
- Symptoms: Failed startup left running/restarting containers and volume state that contaminated subsequent runs.
- Root Cause: E2E TestMain factory returned immediately on `startFn` error without invoking `stopFn` cleanup.
- Fix: Execute `stopFn` on startup failure path and validate with dedicated unit test assertions.

1. Shared-postgres credential source mismatch
- Symptoms: PostgreSQL app containers failed with role/password errors (`Role "sm_kms_database_user" does not exist`) while leader was healthy.
- Root Cause: Shared-postgres used `deployments/shared-postgres/secrets` by default while PS-ID apps used PS-ID secret files.
- Fix: Parameterized shared-postgres secret file paths with `POSTGRES_SECRETS_DIR` and wired all PS-ID `.env.postgres` files (template + 10 instantiations).

1. CRLF secret file encoding on Windows corrupted credentials
- Symptoms: PostgreSQL authn failures persisted even after startup ordering fixes; role/user comparisons diverged.
- Root Cause: CRLF in PS-ID `postgres-*.secret` files injected hidden `\r` via `_FILE` environment loading.
- Fix: Normalized `sm-kms` and `sm-im` PostgreSQL secret files to LF endings and revalidated E2E.

1. Shared-postgres transient init readiness race
- Symptoms: Leader reported healthy during init phase before stable post-init TCP readiness, causing early app connection failures.
- Root Cause: Health probe accepted transient init conditions.
- Fix: Hardened leader healthcheck to require process identity plus explicit TCP probe on `127.0.0.1:5432`.

## Auto-Mode Quality Gate Evaluation

1. Correctness: improved; fixes addressed concrete root causes validated by passing E2E suites.
2. Completeness: improved to full phase completion after reconciling blocked Phase 9 tasks.
3. Thoroughness: strong; runtime logs, container state, and file-encoding bytes were used as evidence.
4. Reliability: strong within framework-v22 scope (build/lint/integration/e2e/deployment gates passing).
5. Efficiency: good; smallest viable changes in orchestration, deployment templates, and secret wiring.
6. Accuracy: strong; memory-only hypothesis was rejected in favor of bind-mount and credential-line-ending root causes.
7. No time pressure: maintained; repeated validation runs were executed until blockers cleared.
8. No premature completion: maintained in final state; prior contradictions removed.

## Recommended Improvements (Highest to Lowest Priority)

1. Add deployment fitness lint for LF-only `.secret` files under `deployments/*/secrets/`.
2. Add explicit Windows bind-mount cert-dir cleanup guidance to deployment/E2E orchestration docs.
3. Add CI check that `POSTGRES_SECRETS_DIR` exists in every PS-ID `.env.postgres` and template instantiation.
4. Add a startup-failure cleanup test pattern requirement for all orchestration factories.
5. Track and reduce `SKIP` cases in `internal/apps/sm-im/e2e/...` for broader runtime assurance.

## Propagation Candidates

1. Deployment instructions: codify LF-only `.secret` enforcement and `POSTGRES_SECRETS_DIR` pattern for shared-postgres include consumers.
2. Testing instructions: add Windows bind-mount writable cleanup pattern for Docker-based E2E TestMain orchestration.
3. Agent execution guidance: require startup-failure cleanup hooks in orchestration code paths by default.
