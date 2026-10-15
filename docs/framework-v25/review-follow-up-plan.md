# Review Follow-Up Plan — Framework v25

**Status**: Open; review recommendations are tracked separately from the completed Framework v25 consolidation.
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

- **2 — Partially verified.** Review migrations 2012–2014 for upgrade/down behavior on SQLite and
  PostgreSQL. Persistence evidence exists; PostgreSQL runtime remains unverified. Legacy
  duplicate-row remediation is intentionally excluded under item 28.
- **3 — Partially verified.** Deployment lint evidence exists; live Compose startup and health
  verification remain outstanding.
- **4 — Complete.** The inventory maintenance procedure is documented in ENG-HANDBOOK §10.2.3
  and propagated to testing instructions. The workflow uploads both inventories; docs, workflow,
  and text validators pass.
- **5 — Partially verified.** Six synthetic result/profile edge cases pass and the actual
  profile-driven gate runs without parser or inventory errors. The run reports 152/320 packages
  meeting thresholds and 168/320 below threshold (71.5% total statement coverage); the CI runner
  has not passed the gate, so threshold remediation and runner confirmation remain outstanding.
- **6 — Partially verified.** Focused listener tests cover shutdown paths; full race validation is
  outstanding.
- **7 — Pending.** Focused TLS tests exist, but whole-set publication atomicity and platform
  permission behavior still need resolution.
- **8 — Partially verified.** Workflow lint evidence exists; live Compose validation and the load
  workflow’s authenticated readiness parity remain outstanding.
- **9 — Pending.** Review the new PKI tests against parallelism, fixture, UUID, cleanup, seam, and
  filename conventions; the prior static review did not inspect every test file.
- **10 — Complete.** Package-level PKI coverage evidence is recorded in `tasks.md` and
  `test-output/review-followup/`; this does not claim a full repository test run.
- **11 — Partially verified.** Focused authority/configuration persistence evidence exists; live
  PostgreSQL cold-start and restore behavior remain outstanding.
- **12 — Complete.** Plan 25 is restored as the completed consolidation plan; this separate plan
  carries review dispositions and links to the evidence tracker.
- **13 — Pending before any future push.** Full normal/tagged builds, lint, shuffled tests, and
  required repository validators have not all been evidenced for this commit set. No push is
  authorized by this plan.
- **14 — Pending.** Verify the required co-author trailer on every reviewed implementation commit.

## Additional Recommendations

- **15 — Complete.** Grant application roles access to databases created by
  `deployments/shared-postgres/upgrade-existing-volume.sql`.
- **16 — Complete.** Make the PostgreSQL upgrade validate or repair existing database ownership,
  grants, and publication properties rather than skipping by name alone.
- **17 — Complete.** Exercise the existing-volume upgrade with the supported PostgreSQL image,
  psql invocation, and required role privileges.
- **18 — Pending.** Validate durable SQLite overrides against the actual image, including shell
  availability, CLI URL precedence, runtime identity, file creation, and permissions.
- **19 — Pending.** Verify durable SQLite override parity and volume isolation across all services
  and product/suite Compose include paths.
- **20 — Pending.** Align load-workflow readiness with the authenticated mTLS readiness contract,
  or document and justify the weaker check.
- **21 — Pending.** Pass workflow inputs through environment variables and validate numeric ranges
  before shell use instead of interpolating expressions directly into commands.
- **22 — Pending.** Publish staged TLS outputs atomically as one set, or require an activation
  marker that prevents consumers observing partial publication.
- **23 — Pending.** Add deterministic TLS filesystem failure-path checks and verify the effective
  platform-specific permissions on staged/published material.
- **24 — Complete.** Tests cover Compose JSON arrays, a single object, and newline-delimited
  objects, including malformed and empty input. Both workflow callers provide the action's
  required `required-completed-services` input.
- **25 — Pending.** Enforce a minimum SQLite version supporting `UPDATE ... RETURNING`, or add a
  tested dialect-specific fallback for CRL number allocation.
- **26 — Pending.** Classify PostgreSQL CRL allocation retries using SQLSTATE codes as well as
  driver text, with bounded retry tests for both database dialects.
- **27 — Pending.** Define whether consumed CRL numbers may have gaps after failed generation, or
  persist generation/publication state if stronger continuity is required.
- **29 — Pending.** Enforce the request-ID sentinel policy at the write boundary so only the
  intended zero UUID bypasses uniqueness.
- **30 — Pending.** Validate authority ID, issuer key ID, and issuer DN consistency during
  issuance/revocation; add persistence constraints if multiple authorities or rotations are
  supported.
- **31 — Pending.** Specify and test whether CRLs are global per shared authority or tenant-scoped;
  apply tenant filtering if the intended contract requires it.
- **32 — Pending.** Decide whether CRL/OCSP URLs and profile metadata are persisted authority state
  or deployment-time metadata, and validate restart/configuration-change behavior accordingly.
- **33 — Pending.** Verify CRL/OCSP URL hostnames, ports, and TLS paths across deployment
  configurations and PS-ID template instances.
- **34 — Pending.** Version the encrypted authority-key envelope and persist the metadata needed
  for key rotation, migration, and recovery.
- **35 — Pending.** Document and test authority create/encrypt failure and concurrency behavior,
  including database-write failure after encryption.
- **36 — Partially verified.** Focused tests protect revocation state from ordinary updates; define
  the lifecycle-transition contract and verify stale updates cannot erase revocation metadata.
- **37 — Pending.** Exercise PostgreSQL multi-instance CRL allocation and issuance/revocation
  continuity across restarts.

## Excluded Recommendation

**28 — Omitted.** Legacy duplicate request-ID remediation is not planned because the project has
no existing deployments or deployed data. Reconsider this only if persistent databases with
pre-migration records become part of the supported upgrade path.
