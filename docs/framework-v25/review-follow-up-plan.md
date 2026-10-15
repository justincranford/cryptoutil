# Review Follow-Up Plan — Framework v25

**Status**: Partially verified; dispositions and remaining validation are recorded below.
**Created**: 2026-10-15
**Scope**: Prior review recommendations 2–14 and additional recommendations 15–27, 29–37.

[`plan.md`](./plan.md) remains the completed 8-to-7 PS-ID consolidation plan. This document tracks
review follow-up only. Recommendation 1 is omitted as requested. Recommendation 28 is omitted
because there are no existing deployments or deployed data to remediate.

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

- **2 — Partially verified.** SQLite migration tests and PostgreSQL up/down SQL checks cover
  migrations 2012–2015; the PostgreSQL migration-runner test is present but was not run. Running
  it requires a separate Docker-enabled environment; no further Docker lifecycle operations were
  run after the recorded volume incident. Legacy duplicate-row remediation is intentionally
  excluded under item 28.
- **3 — Complete.** The durable SQLite Compose instance started healthy and authenticated admin
  readiness returned HTTP 200; deployment validators pass.
- **4 — Complete.** The inventory maintenance procedure is documented in ENG-HANDBOOK §10.2.3
  and propagated to testing instructions. The workflow uploads both inventories; docs, workflow,
  and text validators pass.
- **5 — Partially verified.** Six synthetic result/profile edge cases pass and the actual
  profile-driven gate parses the complete inventory without parser errors. It reports 152/320
  packages meeting thresholds and 168/320 below threshold (71.5% aggregate statement coverage).
  Failures span 84 `internal/apps-tools`, 54 `internal/apps`, 18 `internal/apps-framework`, and
  12 `internal/shared` packages. In-scope PKI packages remain below 95%: handler 90.8%, bootstrap
  93.2%, server 52.4%, revocation 80.4%, and storage 69.4%. No package ceiling exception is
  justified by the current evidence; targeted coverage remediation and a passing CI runner remain
  outstanding. Do not lower thresholds to clear the gate.
- **6 — Partially verified.** Focused listener tests cover shutdown paths; full race validation
  remains outstanding because this Windows host has no GCC compiler. Run it on a supported
  CGO-enabled host.
- **7 — Complete.** TLS outputs are staged and validated, publication failures roll back, and
  Compose consumers are gated on successful `pki-init`; platform permission and failure-path
  tests pass. Rotation/reload remains unsupported and is documented.
- **8 — Complete.** The load workflow uses authenticated mTLS readiness; live authenticated
  readiness and workflow validation pass. The full load workload was not run.
- **9 — Complete.** PKI tests were checked against parallelism, fixtures, UUIDs, cleanup, seams,
  and naming conventions; focused shuffled tests and the PostgreSQL continuity test pass.
- **10 — Complete.** Package-level PKI coverage evidence is recorded in `tasks.md` and
  `test-output/review-followup/`; this does not claim a full repository test run.
- **11 — Complete.** The PostgreSQL continuity test cold-starts authority components, stops and
  restarts PostgreSQL, then verifies authority-key restoration and persisted revocation state.
- **12 — Complete.** Plan 25 is restored as the completed consolidation plan; this separate plan
  carries review dispositions and links to the evidence tracker.
- **13 — Complete.** Normal/tagged builds, normal/tagged golangci-lint, the full shuffled suite,
  and applicable repository validators pass; current logs are recorded in `tasks.md`. An initial
  full-suite run concurrent with repository validators timed out in two RP/RS health tests; the
  isolated tests, shuffled packages, and serial full-suite rerun passed. No push is authorized by
  this plan.
- **14 — Pending.** Verify the required co-author trailer on every implementation commit.

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
  and distinct service/product/suite project scoping; product/suite runtime stacks were not run.
  Runtime verification requires a separate Docker-enabled environment; no further Docker
  lifecycle operations were run after the recorded volume incident.
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
- **30 — Complete.** Authority ID, issuer key ID, and issuer DN consistency are checked across
  issuance/revocation, with focused tests.
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

## Excluded Recommendation

**28 — Omitted.** Legacy duplicate request-ID remediation is not planned because the project has
no existing deployments or deployed data. Reconsider this only if persistent databases with
pre-migration records become part of the supported upgrade path.
