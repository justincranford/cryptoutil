# Framework v26 — Tasks

**Status**: Kubernetes planning draft for human review. Only design and the planning generator
are authorized; runtime implementation remains pending. Fourth-product identity is unresolved.
**Created**: 2026-10-21
**Last Updated**: 2026-10-22

## Phase 1: Define Kubernetes Deployment Contract

- [x] Verify the current Compose baseline: two PostgreSQL-backed `sm-kms` instances share the
      configured database secret; E2E checks their health/TLS, but does not send functional client
      requests to both; PostgreSQL storage E2E covers two connections/stores.
- [x] Map Compose dependencies to K3S and Helm, including shared PostgreSQL, TLS/mTLS, secrets,
      migrations, readiness, telemetry, and persistent storage ownership.
- [x] Define PVC retention for upgrades/uninstalls, PV reclaim policy, backup/restore expectations,
      and a namespace-scoped destructive-reset policy for disposable K3S runs.
- [x] Define chart ownership and configurable default min/max=2/2 per PS-ID/backend pool,
      preserving all four legacy slots in mixed modes, without false SQLite HA.
- [x] Define local/CI Compose-hosted K3S lifecycle and how E2E can target each pod without bypassing production
      TLS/authentication.
- [x] Record cutover replacing application Compose, retaining only the local K3S bootstrap.
- [x] Define grammar, exact selectors, five-mode matrix, network/DB/crypto isolation, and
      framework startup/liveness/readiness semantics in [k8s-design.md](k8s-design.md).
- [x] Generate and validate the exhaustive [k8s-deploy-tree.md](k8s-deploy-tree.md) plus
      machine-readable expected inventory from the same planning model. Named-file shuffled
      tests: 99.1% coverage; vet, deterministic tree check, JSON preview and lint-text pass.
      [Evidence](../../test-output/framework-v26-k8s-planning-evidence.log).
- [ ] Resolve requested fourth product identity/membership versus the registry's three products.
- [ ] Complete human design review and authorize runtime implementation.

## Phase 2: Replace Compose Service Deployment and Test Orchestration

- [ ] Promote the planning graph into production Go generation/check tooling with typed schemas,
      canonical chart templates, deterministic chart/values generation and independent Helm parity tests.
- [ ] Implement Helm charts and values for all services, tiers and modes, with configurable default
      min/max=2/2; no product/service list duplication or manual generated-file edits.
- [ ] Extend shared framework startupz/livez/readyz, background dependency monitor, graceful
      draining, CLI/router/usage/constants, templates, test helpers and all seven services.
- [ ] Implement K3S-only Compose bootstrap, locked prerequisites/toolchain, loopback API/gateway,
      TLS-passthrough SNI routing and capacity/ownership preflight.
- [ ] Implement restricted PKI provisioning, per-pod init materialization, migration Jobs and
      post-migration logical replication; prevent dependency cycles.
- [ ] Implement an explicit old-instance/data/secret-reference migration map, writer quiescence,
      backup/restore, traffic switch and rollback checks; never implicitly fork data/material
      into every new mode.
- [ ] Render/schema-check provider-configured future Kubernetes target, explicitly not locally
      runtime-certified.
- [ ] Deploy pods against the shared PostgreSQL deployment with existing TLS/mTLS, Docker/Kubernetes
      secret, migration, health, and persistence requirements intact.
- [ ] Ensure routine Helm upgrade and pod replacement preserve persistent claims; constrain
      destructive test cleanup to per-run disposable namespaces and resources.
- [ ] Replace Compose-driven app startup and E2E orchestration with K3S/Helm workflows,
      retaining only the local cluster-node Compose bootstrap.

## Phase 3: Verify Multi-Pod Client Behavior

- [ ] Exercise authenticated functional APIs on each pod of all seven services and both
      browser/service paths, not only health/TLS endpoints.
- [ ] Deploy each selector and the simultaneous full matrix; compare actual and rendered graphs
      against independently generated expected inventory, including isolation-negative tests.
- [ ] Test all three probes, dependency/public-listener failures, stale observations, startup
      budget, latency bounds, draining, scaling, CNI enforcement and end-to-end telemetry.
- [ ] Verify independent SQLite endpoints, memory-loss semantics and durable ordinal continuity.
- [ ] Prove writes through one pod are visible through another pod using the shared database.
- [ ] Exercise PKI issuance, revocation, and CRL allocation across pods and verify persisted state.
- [ ] Verify a durable API record survives pod replacement and Helm upgrade, and validate PVC
      retention/reclaim behavior against the documented lifecycle policy.
- [ ] Verify readiness, rollout/restart, and that test cleanup cannot target shared storage.

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
