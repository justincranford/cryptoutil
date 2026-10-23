# Framework v26 — Lessons

**Status**: Kubernetes planning draft for human review; no runtime implementation has started.
**Created**: 2026-10-21
**Last Updated**: 2026-10-22

## Carry-Forward Lesson from Framework v25

Volume names and metadata do not prove data continuity or ownership after recreation or cleanup.
Define Kubernetes PVC retention, PV reclaim, backup/restore, and namespace-scoped test cleanup
before implementation; verify durable application state across pod replacement and Helm upgrade.

## Phase 1: Define Kubernetes Deployment Contract

- Registry evidence is 7 PS-IDs + 3 products + 1 suite, not four products. Do not instantiate
  obsolete PKI as another product after consolidation into sm-kms. The requested fourth
  product remains an explicit unresolved planning input.
- Abstract identity is (run, tier, scope, mode, PS-ID, backend pool, ordinal). Define that
  taxonomy before enumerating resources; namespace, DB, SAN and config generation share it.
- Replica pools preserve legacy slots: two PG replicas and two independent SQLite replicas
  in mixed mode, not two replicas of each of the four existing slots.
- A second SQLite pod is independent data, not HA. Stable ordinal endpoints are necessary;
  routing independent databases through one pooled Service is a correctness bug.
- Effective product/suite Compose secret mounts share unseal/pepper at tier scope despite
  conflicting handbook prose. Migration must preserve HKDF inputs and imported material.
- Shared PostgreSQL logical follower is not a failover primary; wait for migrations before
  subscriptions, without making OLTP readiness depend on the follower.
- Zero-latency remote health cannot be stateless in the literal sense. Fast read-only
  handlers use bounded-freshness snapshots; monitoring state is explicit and testable.
- Admin loopback and mTLS require exec probes with the application's client binary, not
  kubelet HTTP probes or a newly exposed unauthenticated admin port.
- Deterministic generated names do not mean deterministic secret values or predictable
  controller UIDs. Expected inventory must distinguish active/min and reserved/max identities.
- Full matrix costs are derived, not guessed: `21 member occurrences * 7 pools across modes *
  2 pods = 294 app pods`. Local capacity failure must not silently weaken replica settings.
- See [design](k8s-design.md), [model](k8s-model.json), [tree](k8s-deploy-tree.md) and
  [tests](k8s-model_test.go). Generator shuffled tests pass at 99.1% coverage; vet,
  deterministic tree comparison, CLI JSON subset counts and lint-text pass. See
  [validation evidence](../../test-output/framework-v26-k8s-planning-evidence.log).
- Review passes covered source/registry fidelity, exhaustive graph/legacy-slot mapping,
  replica/SQLite semantics, probe transport and freshness, RBAC pod-versus-container
  boundaries, source-to-target data migration, deterministic output and triad consistency.
  Runtime and human-review tasks remain unchecked rather than borrowing prototype evidence.

## Phase 2: Replace Compose Service Deployment and Test Orchestration

Pending.

## Phase 3: Verify Multi-Pod Client Behavior

Pending.

## Phase 4: Define PKI Reload Contract

Pending.

## Phase 5: Implement Safe Per-Pod Reload

Pending.

## Phase 6: Verify and Document PKI Reload Operations

Pending.

## Phase 7: Propagate Lessons

Pending.
