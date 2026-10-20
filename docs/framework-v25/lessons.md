# Lessons â€” Framework v25: 8-to-7 PS-ID Consolidation

**Created**: 2026-05-25
**Purpose**: Lessons learned from each phase. Filled by `implementation-execution` agent
after each phase's quality gates pass.

---

## Executive Summary

1. [Phase 1](#phase-1-port-pki-ca-packages-to-sm-kms): The PKI packages, migration, and certificate profiles are present under sm-kms and pass the final build and test suite.
2. [Phase 2](#phase-2-extend-sm-kms-api-with-pki-endpoints): All 18 PKI routes are registered under both public API prefixes; OpenAPI and handler tests pass.
3. [Phase 3](#phase-3-delete-pki-ca-service-artifacts): The standalone pki-ca service artifacts are absent and normal/tagged builds pass.
4. [Phase 4](#phase-4-delete-pki-product-and-update-registry): The canonical registry, runtime registry, and fitness validation reflect 3 products and 7 PS-IDs.
5. [Phase 5](#phase-5-quality-gates): Normal/tagged builds, the default-concurrency shuffled suite, linters, and applicable cicd-lint checks pass. The latest pinned coverage profile meets all package thresholds; skips are classified under the repository's probability-based policy. Coverage, Mutation, and Load passed on push `c42c09e5359270bce889411bb8fc8755dd0669e3`; Quality and E2E failed there, and their fixes were pushed in `b0d954d22`. The 11 non-Race workflows passed on prior push `7ef7c857`, but CI Race run `39321755986` is a false green because its failing test step was non-blocking. On `e819239de`, CI Quality, Load, and E2E completed successfully; blocking Race run `39356665257` was cancelled at its 30-minute job timeout after reporting `identity/mfa` at 1199.098 seconds and `identity-authz/clientauth` at 311.174 seconds. CI Load run `39361288540` passed quick, standard, and stress. The full-module Race matrix run `39362007528` is queued; its six-root package-set comparison covers all 368 integration-tagged packages exactly once. Docker lifecycle and legacy-volume provenance remain open.
6. [Phase 6](#phase-6-knowledge-propagation): README and handbook topology documentation are reconciled, lessons are recorded, and docs propagation validation passes.

---

## Actions

1. The earlier PKI persistence follow-up was marked resolved prematurely: implementation claims did not establish issuer/tenant CRL consistency or production persistence-chain continuity. See the October 2026 follow-up status and evidence in `tasks.md`.
2. Keep the retired PKI port 8300, PostgreSQL host port 54323, and migration range 5001-5999 unassigned until an explicit canonical registry change authorizes reuse.
3. For future migrations, create and update each phase's evidence directory and task statuses as its quality gates complete instead of reconstructing status from the final tree.

---

> **Mandatory per-phase structure** (each `## Phase N:` section MUST use these 4 sub-headings):
>
> ### What Worked
>
> ### What Didn't Work
>
> ### Root Causes
>
> ### Patterns for Future Phases

---

## Phase 1: Port pki-ca Packages to sm-kms

### What Worked

The package tree under `internal/apps/sm-kms/server/pki/` keeps PKI implementation within the owning service. The full shuffled test suite and package coverage profile verify the ported packages in their current location.

**Current post-push status (2026-10-20)**: The 11 non-Race workflows recorded above passed on
prior push `7ef7c857`. After the fail-closed fixes were pushed in `e819239de`, CI Quality run
`39356665316`, CI Load run `39356665340`, and CI E2E run `39356665264` passed; CI Race run
`39356665257` was cancelled at its 30-minute job timeout. Its test step ran for about 29 minutes
and did not finish the package suite. CI Load run `39361288540` passed quick, standard, and stress.
The full-module Race matrix run `39362007528` is queued after commit `2557168fd`; the interrupted
monolithic local race command is superseded by this exact-once partitioned matrix, not considered a
passing result.

### What Didn't Work

No Phase 1-specific build or test logs were retained, so the original execution sequence and intermediate failures cannot be reconstructed.

### Root Causes

The phase-level evidence archive was not maintained alongside the individual task statuses. The existing migration and code are present, but historical per-phase validation evidence is missing.

### Patterns for Future Phases

Run package-scoped build/tests while porting, preserve the logs under the phase directory, and update each task status immediately after its acceptance criteria pass. The original in-memory port is retained only as a test store; production persistence belongs to the shared sm-kms database.

---

## Phase 2: Extend sm-kms API with PKI Endpoints

### What Worked

The dual-prefix route test enumerates all 18 PKI paths under both `/service/` and `/browser/` prefixes and checks that the profiles endpoint responds successfully at both mounts. OpenAPI generation and handler package tests provide additional validation.

### What Didn't Work

No Phase 2-specific logs were retained. The mounted-route test checks registration for every route, while endpoint request/response behavior is primarily exercised by the PKI handler package tests.

### Root Causes

The evidence archive did not distinguish route-registration checks from handler-level behavior checks, and the phase logs were not persisted.

### Patterns for Future Phases

Keep OpenAPI validation, generated-code builds, route registration, and handler behavior tests as separate evidence entries. Use in-memory `app.Test()` requests for HTTP behavior tests and avoid treating route-table inspection alone as an endpoint behavior test.

---

## Phase 3: Delete pki-ca Service Artifacts

### What Worked

The current tree has no standalone pki-ca API, service, command, config, or deployment directories. Full normal/tagged builds and deployment validation pass after removal.

### What Didn't Work

No Phase 3-specific deletion or build logs were retained.

### Root Causes

The execution tracker did not preserve per-phase evidence when the deletion work was completed.

### Patterns for Future Phases

Pair large artifact removals with a bounded existence check, repository-wide build, and deployment validator; save the exact command output in the phase evidence directory.

---

## Phase 4: Delete pki Product and Update Registry

### What Worked

The machine-readable registry, magic tier constants, and lint-fitness entity registry now agree on the 3-product / 7-PS-ID topology. Fitness and deployment validation pass.

### What Didn't Work

No Phase 4-specific registry or validator logs were retained. The retired pki-ca sentinel constants remain referenced by regression tests by design, so broad text searches need to distinguish retained retirement guards from live registrations.

### Root Causes

The original checks did not document the intentional retired-ID exception, and the phase-specific evidence directory was not kept.

### Patterns for Future Phases

Treat the machine-readable registry as the source of truth, validate all derived registries with fitness checks, and retain retired identifiers only when regression guards require them. State those exceptions explicitly in completion criteria.

---

## Phase 5: Quality Gates

### What Worked

Recorded normal/tagged builds, the default-concurrency shuffled full suite, normal/tagged golangci-lint, and applicable cicd-lint checks passed. The latest detailed full run has 17,736 passing test events, no failures, 36 explained skips, and 41 packages without test files. The current pinned coverage profile passes all 323 eligible thresholds; the focused sysinfo profile reaches 100.0%. The local workflow fixes pass normal/tagged builds, a fresh shuffled full-suite run, all 14 repository validators, and focused E2E-infrastructure tests/lint (evidence is recorded under Task 5.8). The original Quality/E2E fixes were pushed in `b0d954d22`; the stale checksum correction was pushed in `45c2757`; bounded retries for transient module-download failures were pushed in `7ef7c857`. The 11 non-Race workflows passed on the latest push. CI Race run `39321755986` had a success conclusion because its test failure was tolerated by `continue-on-error`; it is not accepted as passing. The fail-closed Race workflow and test fixes pass the post-fix Linux CI-equivalent full internal race run (`-tags=integration -count=2`). The initial standalone root run failed chmod-based permission tests; the first non-root run failed because Git rejected the root-owned `/workspace` checkout as dubious ownership while `go list` collected VCS metadata. Adding `/workspace` to the disposable container's system `safe.directory` made `git status` and `go list ./...` pass as UID 65534. The corrected full standalone race run was stopped before completion after resource monitoring showed 43% CPU and about 1.5 GiB; its partial output is not a passing result. The corrected blocking GitHub Race rerun and runtime gates remain open.

### What Didn't Work

Push `c42c09e5359270bce889411bb8fc8755dd0669e3` exposed two issues: CI Quality #953
(`39165120901`) reported `goconst` for repeated `"cp"` and
`"test-service-pki-init-1:/certs/."` literals in the E2E infrastructure test; CI E2E #952
(`39165121059`) could not remove the generated certificate bind mount because files retained
container ownership. These fixes were pushed in `b0d954d22`. Eight workflows on that revision then
failed before their substantive steps because `go.sum` retained an unused
`github.com/grpc-ecosystem/grpc-gateway v1.16.0` content checksum that `go mod tidy` deletes.
The checksum correction was pushed in `45c2757`; transient proxy HTTP/2 failures prompted bounded
module-download retries in `7ef7c857`. The 11 non-Race workflows passed on that push, but CI Race
run `39321755986` was a false green: its test step was non-blocking and reported test failures.
GORM error-path tests shared mutable fixtures across parallel subtests; JWT claim error tests
computed expirations before delayed parallel execution; and registration rate-limit tests reused
tenant identifiers and depended on token-bucket refill timing under race instrumentation. These
test defects were fixed. A local CI-equivalent run then exposed that `./scripts/...` matches no Go
packages and makes `go test` exit 1. The workflow fails closed and has an enforced timeout; the
latest version partitions all 368 integration-tagged packages across six non-overlapping roots.
Run `39362007528` for commit `2557168fd` is queued, so the corrected blocking Race result is pending.
The earlier CI-equivalent coverage baseline reported 152/320 executable
packages meeting thresholds and 168/320 below (71.5% aggregate statement coverage); this is
historical and superseded by the pinned 352-package profile, which now passes all thresholds. The
user explicitly dropped the historical aggregate comparison, so it will not be pursued; the
profiles were not established as like-for-like. The standalone full-repository race command
remains unverified. The Windows host has no GCC compiler; Docker Engine 29.8.2 meets the corrected
minimum aligned with Docker Desktop 4.95.0. An initial inline Windows coverage-profile argument
was parsed as an extra package.

### Root Causes

The original coverage gate pinned a stale no-test inventory and the initial profile did not pass
package thresholds; after reconciliation, the current pinned gate validates 352 packages and all
323 eligible thresholds. The Quality failure came from repeated literals in an expected command
argument list. The E2E failure came from `cp -a` preserving UID/GID 65532 on a host bind mount that
the runner later removes. The follow-up failures share a separate cause: an unused grpc-gateway
v1.16.0 content checksum remained in `go.sum`, so CI's `go mod tidy` check failed before test,
lint, and runtime steps. The repeated literals, runner ownership, and stale checksum were
corrected; bounded retries now handle transient module proxy HTTP/2 failures.

The prior CI Race workflow conclusion did not prove a passing test: `continue-on-error` masked
race-test failures. The failing tests shared mutable GORM fixtures, used stale JWT expiration
values in delayed parallel subtests, and reused rate-limit tenant names while relying on refill
timing. Isolate mutable fixtures, compute time inside each parallel subtest, and use unique tenant
IDs with a non-refilling limiter. Race workflows must not tolerate test failures, and package
selectors must match actual Go packages (`./internal/...`, not the empty `./scripts/...`). The
first Linux non-root standalone rerun also exposed an environment issue: Git refused the
root-owned mounted checkout as dubious ownership, causing `go list` VCS metadata collection to
fail. Adding the checkout to the disposable container's system `safe.directory` made the direct
Git and `go list` checks pass. The corrected full run was then stopped before completion after it
consumed 43% CPU and about 1.5 GiB; interrupting the PowerShell client alone left test processes
running inside the container, so those exact processes and the task-specific validation container
were explicitly stopped. A separate existing `sm` PostgreSQL app container was observed in an
on-failure restart loop (333 restarts) because `shared-postgres-leader` and the OTel collector did
not resolve in its Compose network; it was left untouched. Go's race detector requires CGO and a C
compiler; the Windows host has neither, so the monolithic race attempt was stopped to avoid
unnecessary laptop load and is superseded by the six-partition hosted Linux matrix. A quoted
coverage-profile path is required by the Windows shell
invocation.

### Patterns for Future Phases

Keep the pinned package inventory and thresholds unchanged; rerun the coverage gate after test changes. Run `go mod tidy -diff` before pushing dependency-affecting work so CI and local module metadata stay consistent. Accept documented probability-sampled skips, but investigate every unexplained/actionable skip. Run full race tests in a supported Go/Linux environment with CGO and a C compiler when the Windows host lacks GCC. For container-to-host file handoffs, verify ownership as well as permission bits before runner cleanup. Pass the coverage-profile path as a quoted argument and validate it with `go tool cover`. Treat function-level coverage as a diagnostic, inspect uncovered behavior, test meaningful branches, and document structural or external-boundary limits without weakening package targets.

---

## Phase 6: Knowledge Propagation

### What Worked

The README and ENG-HANDBOOK now describe three products and seven services, position PKI as an sm-kms capability, and use corrected directory-count formulas. `lint-docs` and `lint-text` pass.

### What Didn't Work

The plan and task trackers still reflected stale task-level labels, and the older phase evidence folders were absent.

### Root Causes

Tracker and evidence updates had not been part of the phase completion workflow.

### Patterns for Future Phases

Reconcile phase status from current implementation and executable evidence; do not interpret stale task labels as proof that work was not done. Preserve missing historical evidence as an explicit limitation instead of fabricating it.

## Review Follow-Up

The October 2026 review corrected premature persistence-completion claims. Focused validation
now exercises the real migration, barrier, file-backed SQLite restart, CRL/OCSP continuity,
and historical issuer publication chain. It does not establish PostgreSQL multi-instance or
Docker deployment behavior; the evidence and those remaining checks are enumerated in
`tasks.md`.
