# Review Follow-Up Plan — Framework v25

**Status**: Partially verified; dispositions and remaining validation are recorded below.
**Created**: 2026-10-15
**Scope**: Prior review recommendations 2–14 and additional recommendations 15–38.

[`plan.md`](./plan.md) records the implemented 8-to-7 PS-ID consolidation; Phase 5 quality-gate
acceptance remains in progress. This document tracks review follow-up only. Recommendation 1 is
omitted as requested. Recommendation 28 remains pending as described below.

## Status Definitions

- **Complete**: The recommendation’s acceptance criteria are met and evidence is recorded.
- **Partially verified**: Implementation or focused evidence exists, but a stated gap remains.
- **Pending**: Required review, implementation, or validation has not been evidenced.

Items 15–27 and 29–37 are accepted into this follow-up plan, but inclusion does not mean that
implementation is authorized or complete. Update a status only with evidence recorded in
[`tasks.md`](./tasks.md) or its referenced `test-output/review-followup/` artifacts. The task table
is authoritative for implementation results; this plan is authoritative for the recommendation
scope and current disposition. The recommendation numbers preserve the prior review’s numbering
and are independent of the numbered rows in the task evidence table.

## Prior Review Recommendations

- **2 — Complete.** SQLite migration tests and PostgreSQL up/down SQL checks cover migrations
  2012–2015, and the PostgreSQL migration-runner test passed against an isolated testcontainers
  PostgreSQL container; its log records successful termination of that container. No Compose
  lifecycle operations were run after the recorded volume incident. Legacy duplicate-row
  remediation is intentionally excluded under item 28.
- **3 — Complete.** The durable SQLite Compose instance started healthy and authenticated admin
  readiness returned HTTP 200; deployment validators pass.
- **4 — Complete.** The inventory maintenance procedure is documented in ENG-HANDBOOK §10.2.3
  and propagated to testing instructions. The workflow uploads both inventories; docs, workflow,
  and text validators pass.
- **5 — Partially verified.** Six synthetic result/profile edge cases pass. The final
  profile-driven run now verifies all 352 pinned packages, 27 no-test packages, both expected
  no-statement packages, and all 323 threshold-eligible packages with zero threshold failures;
  aggregate coverage is 81.6% (informational). The profile and gate summary are recorded in
  `tasks.md`. Earlier 71.5% and 79.5% collection results are historical. Keep the pinned
  inventory and thresholds unchanged. The current run is not proven like-for-like with the
  earlier 79.9%/78.3% aggregate and 95.3%/95.2% businesslogic baselines, so that comparison
  remains inconclusive. The deployment-linter package's 98.1% result and synthetic gate cases
  remain as previously evidenced.
- **6 — Partially verified.** A focused shuffled Linux race run passes for Authz, IDP, RS, and
  framework listener packages after fixing unsynchronized admin-listener publication in Authz and
  IDP (`test-output/reviewfollowup/race-go-linux-admin-listener-fixed.log`). The later full race
  failure at `TestCheckInDir_RealWorkspace` came from its five-parent relative root resolving to
  `/internal/apps`; the test now discovers the nearest `go.mod` from its working directory, and
  the shuffled health-path package passes
  (`test-output/reviewfollowup/health-path-real-workspace-shuffle.log`). A local race rerun is
  unavailable because this Windows host has no GCC/Clang
  (`health-path-real-workspace-race-shuffle.log`); the current Docker Engine 29.8.2 is below the
  documented 29.9.1 minimum, and no supported Linux runner is available in this continuation.
  Full-repository race validation remains outstanding.
- **7 — Complete.** TLS outputs are staged and validated, publication failures roll back, and
  Compose consumers are gated on successful `pki-init`; platform permission and failure-path
  tests pass. Rotation/reload remains unsupported and is documented.
- **8 — Partially verified.** The load workflow uses authenticated mTLS readiness; live authenticated
  readiness and workflow validation pass. An earlier `ServiceApiSimulation` `complete` profile
  attempt, before the current headless registration/Basic-credentials flow, failed all 24
  key-creation requests because the simulation supplied no authorization token
  (`test-output/review-follow-up/v25-runtime-load-complete-20261016-0401/runtime-load-summary.log`
  and `load-auth-source.log`). `PublicServerBase` now maps typed application errors to their
  declared HTTP status, so missing authorization returns 401 instead of 500; this does not make the
  load workload valid. Headless client registration and Basic-credential exchange now produce a
  service session, which the current Gatling scenario supplies to each virtual user's
  cryptographic requests; compilation, Spotless, and Checkstyle pass
  (`test-output/reviewfollowup/v25-gatling-auth-compile.log`).
  Remaining evidence: run the complete workload against the isolated runtime and record its result;
  do not disable authentication to make it pass. The 2026-10-16 live attempt did run against a
  healthy isolated product runtime but failed all 48 requests before the current headless
  client-registration/Basic-credentials session flow was implemented. That attempt does not
  validate the current flow or establish that a current route is absent. Live authenticated load
  against the current implementation remains unverified; see
  `test-output/review-follow-up/v25-runtime-current/gatling-complete-retry1.log`.
- **9 — Complete.** PKI tests were checked against parallelism, fixtures, UUIDs, cleanup, seams,
  and naming conventions; focused shuffled tests and the PostgreSQL continuity test pass.
- **10 — Partially verified.** The final authoritative profile now passes the pinned 352-package
  inventory gate: 27 no-test packages and two expected no-statement packages match, and all 323
  threshold-eligible packages pass with no threshold failures. Aggregate coverage is 81.6%
  (informational). The current scoped results for PKI, framework/shared, fitness, remaining
  apps-tools, and non-PKI/non-identity apps are retained in `tasks.md` and
  `test-output/coverage-analysis/`. The historical like-for-like comparison against earlier
  aggregate and package-specific baselines remains inconclusive. Do not change the pinned
  inventory or thresholds.
- **11 — Complete.** The PostgreSQL continuity test cold-starts authority components, stops and
  restarts PostgreSQL, then verifies authority-key restoration and persisted revocation state.
- **12 — Complete.** Plan 25 is restored as the consolidation plan, with implementation phases
  1–4 and knowledge propagation complete and Phase 5 acceptance correctly marked in progress.
  This separate plan carries review dispositions and links to the evidence tracker.
- **13 — Partially verified.** Current-tree normal/tagged builds and full normal/tagged
  golangci-lint pass; the 11 relevant repository validators also pass after the tracker updates.
  Evidence is recorded in
  `test-output/reviewfollowup/framework-v25-build-after-fixes-20261017.log`,
  `test-output/reviewfollowup/framework-v25-golangci-after-fixes-20261017.log`, and
  `test-output/reviewfollowup/framework-v25-cicd-lint-final-doc-updates-20261017.log`. The full
  shuffled suite exits 0 with no failed tests or packages; the latest run records 84 test-level
  skips across 325 passing packages and 41 package-level no-test events. The zero-skip criterion
  remains open. The authoritative pinned coverage gate passes 323 eligible packages, with 27
  no-test packages, two expected no-statement packages, and zero threshold failures; see
  `tasks.md`. Linux race, authenticated load, mutation, and supported-Engine teardown/legacy-volume
  gates remain open. The current focused server-package run passes
  (`test-output/reviewfollowup/public_server_server_package_shuffle.log`). Keep final validation
  open until the remaining gates are resolved. An initial
  full-suite run concurrent with repository validators timed out in two RP/RS health tests; the
  isolated tests, shuffled packages, and serial full-suite rerun passed. A later full shuffled run
  had one transient admin-listener shutdown assertion; the isolated test and listener package
  passed, followed by another successful full shuffled run. The latest full shuffled run and
  cancellation-path fix are recorded in
  `test-output/reviewfollowup/framework-v25-go-test-current-20261016.log`; the suite exits 0,
  while the earlier audited skip total and unresolved gaps keep the zero-skip gate open. No push
  is authorized by this plan.
- **14 — Complete.** `test-output/reviewfollowup/commit-trailer-audit.log` verifies the required
  co-author trailer on every task commit from the recorded baseline through the implementation,
  disposition, and post-review fix commits.

## Additional Recommendations

- **15 — Complete.** Grant application roles access to databases created by
  `deployments/shared-postgres/upgrade-existing-volume.sql`.
- **16 — Complete.** Make the PostgreSQL upgrade validate or repair existing database ownership,
  grants, and publication properties rather than skipping by name alone.
- **17 — Complete.** Exercise the existing-volume upgrade with the supported PostgreSQL image,
  psql invocation, and required role privileges.
- **18 — Complete.** Durable SQLite overrides were exercised against the application image,
  including initialization, runtime identity, file creation, and authenticated readiness.
- **19 — Partially verified.** All 14 SQLite instances have distinct PS-ID/instance data volumes
  and distinct service/product/suite project scoping. Product and suite Compose configurations
  parse successfully (`product-compose-config-services.log` and `suite-compose-config-services.log`).
  The isolated suite runtime now has 28 healthy app containers and healthy PostgreSQL leader and
  follower containers (`test-output/review-follow-up/v25-replication-current/runtime-health-current.log`);
  the earlier bounded run recorded 56 public health probes returning HTTP 200 in
  `test-output/review-follow-up/v25-runtime-20261016-0130/runtime-summary.log`, before the follower
  schema fix. PostgreSQL replication now verifies 7/7 subscriptions, 146/146 ready relations, and
  sentinel insert/delete propagation (`test-output/review-follow-up/v25-replication-current/validation-results.txt`).
  The isolated product Compose runtime has since started healthy after that fix
  (`test-output/review-follow-up/v25-runtime-current/product-runtime-up.log`). The current
  headless client-registration and Basic-credentials session flow is described under
  Recommendation 8; live authenticated load against it remains unverified, and the earlier failure
  predates the current flow. Do not operate on globally named or historical volumes.
- **20 — Complete.** Load-workflow PostgreSQL readiness uses authenticated mTLS; the matching
  readiness command was exercised against a live PostgreSQL-backed service.
- **21 — Complete.** Workflow inputs are passed through environment variables and validated
  against numeric bounds before use; workflow validation passes.
- **22 — Complete.** Publication uses staged, validated files with rollback, and Compose consumers
  cannot start unless `pki-init` completes successfully; rotation/reload is explicitly unsupported.
- **23 — Complete.** Deterministic filesystem failure-path tests and effective platform-specific
  permission checks pass.
- **24 — Complete.** Tests cover Compose JSON arrays, a single object, and newline-delimited
  objects, including malformed and empty input. Both workflow callers provide the action's
  required `required-completed-services` input.
- **25 — Complete.** CRL allocation checks the SQLite version required for `UPDATE ... RETURNING`;
  SQLite allocation tests pass.
- **26 — Complete.** PostgreSQL retry classification uses SQLSTATE and bounded retries; both
  PostgreSQL and SQLite allocation tests pass.
- **27 — Complete.** The contract permits CRL-number gaps after failed generation/publication and
  never reuses reserved numbers; documented and tested.
- **29 — Complete.** The request-ID sentinel policy is enforced at persistence; zero UUID and
  duplicate cases are covered by focused tests.
- **30 — Complete.** Issuance and revocation enforce the configured authority ID, issuer key ID,
  and issuer DN. Pre-migration records without an authority ID can be revoked only after their
  embedded certificate, serial, issuer, authority key ID, and signature are verified against the
  configured issuer; focused tests cover valid and invalid legacy records.
- **31 — Complete.** CRLs are explicitly issuer-wide for the shared authority; tests verify
  revocation lookup and publication semantics without tenant-scoped CRL partitioning.
- **32 — Complete.** CRL/OCSP URLs and profile metadata are deployment-time configuration, not
  persisted authority state; restart and configuration precedence are tested.
- **33 — Complete.** Deployment endpoint hostnames, ports, and TLS paths are validated across
  PS-ID configs and template instances.
- **34 — Complete.** The encrypted authority-key envelope is versioned and the version is
  persisted through migrations and tested across PostgreSQL restart.
- **35 — Complete.** Authority creation/encryption failures and concurrency are covered, including
  persistence failure after encryption.
- **36 — Complete.** Revocation lifecycle transitions are explicit; stale updates cannot erase
  revocation metadata, as verified by focused storage and handler tests.
- **37 — Complete.** PostgreSQL continuity tests exercise concurrent multi-connection CRL number
  allocation, persisted revocation, and authority-key restoration across container restart.
- **38 — Partially verified.** Code review found that an already-enabled logical subscription
  skipped follower-schema preparation after migration 1006 renamed `client_secret` to
  `client_secret_hash` on the leader. The canonical and instantiated setup scripts now require the
  migrated leader schema and reconcile the rename on both follower tables before accepting an
  existing subscription; Bash syntax, deployment validation, template fitness, and text checks
  pass (`test-output/reviewfollowup/framework-v25-follower-schema-reconciliation-20261017.log`).
  A persistent legacy-follower upgrade and replication smoke test remains unverified because the
  local Docker Engine is below the documented minimum; do not claim the live upgrade path passes.

## Excluded Recommendation

**28 - Pending / fail-closed.** Legacy duplicate request-ID remediation remains unauthorized because
  the previously deleted historical global volumes have unknown contents and provenance.
  `test-output/reviewfollowup/legacy-volume-readonly-inventory.log` shows the three known historical
  global volume names absent from the current Docker inventory, but cannot recover deleted contents
  or establish the state of other deployments. The user's later Docker Desktop screenshot showed
  13 Postgres entries and an active `v25runtime20261016productretest` Compose project. The earlier
  `docker-created-postgres-follow-up-20261016.log` says 12 Testcontainers-labelled Postgres
  candidates and that project remained present, with no Docker resources changed. The
  `docker-container-cleanup-20261016-1148.log` records removal of 64 containers from two other
  named Compose projects and says the original Testcontainers candidates remained. A fresh
  read-only inventory in `docker-live-inventory-continuation-20261016.log` at 18:48 local found no
  containers, running stats, or Compose projects in either Docker context and retained the visible
  named volumes; no Docker resources were changed in that continuation. Earlier event queries for
  17:45–18:05 local across both contexts returned no matching events, and do not establish whether
  the events did not occur or had expired from retained history. A later read-only inventory at
  19:17 local (`docker-live-inventory-current-20261016-191749.log`) again found no containers,
  running stats, or Compose projects in either context. Both contexts resolved to the same Docker
  Engine ID and version 29.8.2; the snapshot retained the complete listing of 63 visible volumes
  per context and changed no Docker resources. The latest read-only inventory at 20:49 local
  (`docker-live-inventory-current-20261016-204951.log`) again found no containers, running stats,
  or Compose projects in either context. Both contexts resolve to the same Engine ID and
  version 29.8.2 and report 64 visible volumes, including all 11 named framework
  follow-up/product-retest volumes with their Compose project labels; no Docker resources were
  changed. The count change from 63 to 64 visible volumes is unexplained. As no containers were
  present, there were no current IDs, mounts, networks, logs, or per-container resource
  measurements to inspect. The evidence does not identify who or what removed the screenshot's
  containers. A further read-only inventory at 21:03 local
  (`docker-live-inventory-current-20261016-210349.log`) again found zero containers and Compose
  projects in both contexts, 64 visible volumes per context, and all 11 tracked framework
  follow-up/product-retest volumes present. It made no resource changes and does not establish
  disappearance provenance. A source audit found a credible destructive
  teardown path: `ci-load.yml` previously invoked `docker-compose-down` without its `compose-file`
  input, and the action defaulted to the suite Compose file and ran `docker compose down -v`. The
  action now requires an explicit Compose file and runs `docker compose down` without volume
  deletion; `ci-load.yml` passes the same sm-kms Compose file it starts. Static workflow validation
  passes (`framework-v25-compose-teardown-validation-20261016.log`), but live teardown validation
  is blocked because Docker Engine 29.8.2 is below the documented 29.9.1 minimum. No evidence
  attributes historical deletions to the corrected action. Do not claim this session removed the
  screenshot's Postgres entries or product-retest project, and do not delete preserved volumes. Do
  not perform destructive deduplication while legacy-volume contents and provenance are unknown.
  Remaining evidence: resolve the disappearance provenance and perform any eventual teardown test
  only on a supported engine and isolated task-owned project; use non-destructive inspection for
  legacy-volume state.

## Post-Implementation Deep-Review Findings

These findings were raised by the requested deep review of the implementation commits. Their
identifiers are separate from the numbered recommendations above; disposition evidence is recorded
in `tasks.md`.

- **F-1 — Complete.** Compose credential validation requires `_FILE` values to resolve to declared,
  service-mounted Docker secrets; rejects inline values and all host-environment interpolation for
  credential variables; and recognizes generic `_KEY` names without treating `MONKEY` as a
  credential. Focused tests and all deployment validators pass.
- **F-2 — Complete.** Revocation remains available for pre-migration certificates without a stable
  authority ID only after the embedded certificate identity, serial, issuer, authority key ID, and
  signature are verified against the configured issuer. Focused regression tests pass, and the
  compatibility rule is documented in ENG-HANDBOOK §6.5.
- **F-3 — Complete.** Compose `env_file` contents are scanned for credential variables, with paths
  resolved relative to the declaring Compose file (including included files), and `_FILE`
  references validated against mounted Docker secrets. Regression tests cover both Compose forms,
  optional/missing files, unresolved path interpolation, credential values, and valid secret
  references. The deployment, docs, and text validators pass; evidence is recorded in
  `test-output/reviewfollowup/post-review-env-file-tests.log` and
  `test-output/reviewfollowup/post-review-env-file-validation.log`.
- **F-4 — Complete.** Certificate mounts are parsed as exact source/target pairs in both short and
  long Compose syntax. Host bind mounts or unexpected sources targeting certificate paths are
  rejected, and the activation barrier is checked for each consumer of the expected named volume.
  Regression tests cover the expected-name bind-source bypass and long-syntax named/bind mounts.
  Focused tests and all deployment, fitness, docs, text, Go, and Go-test validators pass; evidence is
  recorded in `test-output/reviewfollowup/post-review-cert-volume-tests.log` and
  `test-output/reviewfollowup/post-review-cert-volume-validation.log`.
- **F-5 — Complete.** Windows mode-bit checks do not enforce confidentiality against inherited
  filesystem ACLs. `pki-init` now applies and verifies a protected DACL granting access only to
  the current user and LocalSystem for staged private key files, keystore PKCS#12 files, and
  `tls-config.yml` before publication. Windows-specific tests reject a broad ACL and verify the
  published private files; TLS tests, normal/tagged builds and lint, and repository validators
  pass. Evidence is recorded in `test-output/reviewfollowup/post-review-windows-acl-*.log`.
- **F-6 — Complete.** The staging directory now receives a protected, inheritable DACL for the
  current user and LocalSystem before generation starts. Windows regression tests verify that
  broad ACLs are replaced, child private files inherit only authorized principals, and `Generate`
  validates the staging ACL before its first write. The shuffled TLS package tests, normal and
  `e2e,integration` builds, focused normal/tagged golangci-lint, and Go, Go-test, fitness, docs, and
  text validators pass. Evidence is recorded in `test-output/reviewfollowup/post-review-windows-staging-*.log`.
- **F-7 — Complete.** Certificate volume targets are now normalized with POSIX path semantics.
  Only the expected named volume mounted at an exact certificate root is accepted after
  normalization; trailing-slash aliases of that named-volume target are accepted, while bind
  mounts and overlapping ancestor/descendant paths are rejected. Table-driven tests cover short
  and long Compose syntax, valid trailing-slash named mounts, forbidden bind mounts,
  root/ancestor/descendant overlaps, and a non-overlapping sibling path. The shuffled
  deployment-linter tests, `lint-deployments`, normal and tagged builds/lint, and Go, Go-test,
  fitness, docs, and text validators pass. Evidence is recorded in
  `test-output/reviewfollowup/post-review-cert-volume-*.log`.
- **F-8 — Complete.** Compose volume targets with unresolved variable interpolation are rejected
  in both short and long syntax because their resolved paths cannot be established safely.
  Regression tests cover braced, unbraced, and defaulted short targets, long-syntax targets,
  escaped dollar literals, and source interpolation with a known non-certificate target. CO-21 now documents this fail-closed
  policy. F-4/F-7 evidence and the F-7 disposition are updated to match the tests: trailing-slash
  aliases are normalized and accepted for the expected named volume, while overlapping bind
  mounts remain rejected. Focused and shuffled deployment-linter tests, normal/tagged builds and
  lint, and repository validators pass; evidence is recorded in
  `test-output/reviewfollowup/post-review-compose-target-interpolation-*.log`.
- **F-9 — Complete.** `PublicServerBase` now preserves the HTTP status and safe JSON summary
  declared by typed application errors. Missing session authorization returns HTTP 401 rather than
  Fiber's default HTTP 500. Regression coverage passes in the focused server tests and the shuffled
  server subtree (2,216 tests across 17 packages); evidence is recorded in
  `test-output/reviewfollowup/public-server-server-package-shuffle.log`. Repository-wide final
  validation remains open under Recommendation 13.
- **F-10 - Static validation passes; live authenticated load remains unverified.**
  The current headless service flow registers a client and exchanges HTTP Basic credentials for a
  service session; the Gatling scenario attaches that session to cryptographic requests.
  `ServiceApiSimulation` explicitly handles the CI `standard` and `stress` profiles and rejects
  unsupported profile names instead of silently falling back to `quick`. The CI workflow targets
  PostgreSQL-1 on port 8002, matching its readiness target. Forced clean Maven compilation,
  Spotless, and Checkstyle pass in
  `test-output/reviewfollowup/v25-gatling-auth-compile.log`; workflow and text validation pass in
  `test-output/reviewfollowup/v25-load-workflow-validation.log`. The retained 48/48 live failure
  predates the current auth flow and is not evidence that the current routes are absent or that the
  current workload fails. Run and retain the authenticated `complete` profile against the isolated
  product runtime; keep F-10 open until the configured success-rate threshold passes. Evidence for
  the earlier attempt is `test-output/review-follow-up/v25-runtime-current/gatling-complete-retry1.log`.
