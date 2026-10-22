# Framework v26 — Tasks

**Status**: Parked; no implementation tasks are authorized until Phase 1 defines the Kubernetes
deployment and PKI reload contracts.
**Created**: 2026-10-21

## Phase 1: Define Kubernetes Deployment Contract

- [x] Verify the current Compose baseline: two PostgreSQL-backed `sm-kms` instances share the
      configured database secret; E2E checks their health/TLS, but does not send functional client
      requests to both; PostgreSQL storage E2E covers two connections/stores.
- [ ] Map Compose dependencies to K3D/K3S and Helm, including shared PostgreSQL, TLS/mTLS, secrets,
      migrations, readiness, telemetry, and persistent storage.
- [ ] Define chart ownership and guarantee a minimum of two pods per each of the seven PS-ID
      services in standalone, product, and suite charts.
- [ ] Define local/CI K3D/K3S lifecycle and how E2E can target each pod without bypassing production
      TLS/authentication.
- [ ] Record the cutover plan that replaces Compose as the service deployment and E2E orchestration
      path.

## Phase 2: Replace Compose Service Deployment and Test Orchestration

- [ ] Implement Helm charts and values for all seven service deployments with a default minimum of
      two pods per service.
- [ ] Deploy pods against the shared PostgreSQL deployment with existing TLS/mTLS, Docker/Kubernetes
      secret, migration, health, and persistence requirements intact.
- [ ] Replace Compose-driven service startup and E2E orchestration with K3D/K3S workflows.

## Phase 3: Verify Multi-Pod Client Behavior

- [ ] Exercise authenticated functional APIs on each pod, not only health/TLS endpoints.
- [ ] Prove writes through one pod are visible through another pod using the shared database.
- [ ] Exercise PKI issuance, revocation, and CRL allocation across pods and verify persisted state.
- [ ] Verify readiness, rollout/restart, and persistent data behavior in the K3D/K3S environment.

## Phase 4: Define PKI Reload Contract

- [ ] Inventory active and historical PKI component state and its configuration sources.
- [ ] Specify reloadable settings and which authority/key changes remain restart-only or require a
      distinct rollover workflow.
- [ ] Compare trigger choices and document authorization, validation, status reporting, rollback,
      in-flight request, and per-pod semantics.
- [ ] Review the contract against the safety invariants in `plan.md` and record decisions.

## Phase 5: Implement Safe Per-Pod Reload

- [ ] Construct and validate a complete candidate PKI component generation before activation.
- [ ] Activate candidate state atomically for new requests while allowing in-flight requests to finish
      on their existing generation.
- [ ] Preserve the last known-good generation and expose an explicit failure result when reload fails.
- [ ] Retain historical authority behavior and avoid logging key or credential material.

## Phase 6: Verify and Document PKI Reload Operations

- [ ] Add table-driven tests for successful reload, invalid configuration, construction failures,
      and rollback to the prior generation.
- [ ] Add concurrency tests proving requests cannot observe mixed generations; run under `-race`.
- [ ] Verify historical certificate verification and revocation behavior after active configuration
      changes.
- [ ] Document the trigger, reloadable fields, per-pod behavior, recovery, and restart fallback.
- [ ] Run normal/tagged builds, focused tests, linters, and applicable repository/CI quality gates.

## Phase 7: Propagate Lessons

- [ ] Update applicable handbook, Kubernetes, deployment, and PKI/configuration guidance with
      reusable operational rules.
- [ ] Record phase post-mortems, validation evidence, and the completion summary before archiving.
