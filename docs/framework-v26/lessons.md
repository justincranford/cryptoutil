# Framework v26 — Lessons

**Status**: Phase 1 lifetime/placement closure is complete. The user authorized Phase 2 on
2026-11-01 with: “Then implement phase 2, including post-mortem and fixes at the end. Ensure all
work is committed.” Phase 2 is authorized but not started.
**Created**: 2026-10-21
**Last Updated**: 2026-11-01

## Carry-Forward Lesson from Framework v25

Volume names and metadata do not prove data continuity or ownership after recreation or cleanup.
Define Kubernetes PVC retention, PV reclaim, backup/restore, and namespace-scoped test cleanup
before implementation; verify durable application state across pod replacement and Helm upgrade.

## Phase 1: Define Kubernetes Deployment Contract

- Registry evidence is 7 PS-IDs + 3 products + 1 suite, not four products. Do not instantiate
  obsolete PKI as another product after consolidation into sm-kms. Use the three registered
  products as the explicit draft scope assumption; adding a fourth requires a defined
  identity/membership scope change, not guessed membership or a claim of user approval.
- Abstract identity is (run, tier, scope, mode, PS-ID, backend pool, ordinal). Define that
  taxonomy before enumerating resources; namespace, DB, SAN and config generation share it.
- Replica pools preserve legacy slots: two PG replicas and two independent SQLite replicas
  in mixed mode, not two replicas of each of the four existing slots.
- A second SQLite pod is independent data, not HA. Stable ordinal endpoints are necessary;
  routing independent databases through one pooled Service is a correctness bug.
- Effective product/suite Compose secret mounts share unseal/pepper at tier scope despite
  conflicting Compose-era handbook prose. Migration must preserve HKDF inputs and imported
  material; defer permanent handbook propagation to Phase 7 after actual runtime adoption.
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
- See the [durable deployment contract](../../docs/kubernetes-deployment-contract.md),
  [model](../../api/cryptosuite-registry/k8s/k8s-model.json),
  [tree](../../api/cryptosuite-registry/k8s/k8s-deploy-tree.md),
  [expected inventory](../../api/cryptosuite-registry/k8s/k8s-expected-inventory.json),
  and named-file input, inventory and CLI tests in
  [`api/cryptosuite-registry/tools/k8s-plan`](../../api/cryptosuite-registry/tools/k8s-plan/).
- Three main-agent review passes covered source/architecture correctness,
  requirements/artifacts/tests, and operational/scope/status integrity. They accepted Phase 1
  technically; they were not themselves user authorization. The separate Phase 2 authorization
  is recorded below.
- Phase 2 authorization is recorded from the user's 2026-11-01 request; it does not change the
  registry-backed three-product set. Phase 2 implementation items remain unchecked.

### Phase 1 Lifetime/Placement Postmortem

#### Gap

The Phase 1 planning generator, model, generated tree and expected inventory lived only inside
`docs/framework-v26`, a tracking directory the user states is ephemeral and deleted after
completion. The contract itself also lived there. Reusing the prototype or its generated inputs
therefore required a directory that is not a durable repository owner.

#### Root Cause

The planning prototype was treated as disposable along with its tracking directory, but reusable
inputs/code and generated reference artifacts had no separate durable owner. Fixed CLI input and
output paths reinforced the misplaced ownership, while no file-I/O seam test asserted their
repository-root locations.

#### Fix

Move the eight `//go:build ignore` Go source/test files to
`api/cryptosuite-registry/tools/k8s-plan/`; place the non-secret model, generated human tree and
machine inventory in `api/cryptosuite-registry/k8s/`; and place the contract in
`docs/kubernetes-deployment-contract.md`. Keep only plan/tasks/lessons in the ephemeral directory,
with links to durable artifacts. Make CLI paths repository-root-relative and test expected input
and output paths through injected read/write functions. Keep the three registered products and
all inventory semantics unchanged; this corrects artifact lifetime, not Phase 2 runtime behavior.

#### Verification

Run named-file tests, vet and scoped golangci-lint against all eight moved files; maintain at least
98% coverage; run generator write/check and compare tree/JSON bytes; verify a path-seam test uses
the durable locations without renaming or deleting the tracking directory; run relevant build and
text/docs/fitness validators. Record exact commands and UTF-8/LF evidence in the durable test
evidence log. Phase 2 is authorized but implementation remains unstarted.

### Phase 1 Postmortem

#### What Worked

- Registry and effective Compose-source evidence preserved the two shared-PostgreSQL legacy
  replicas and independent SQLite semantics.
- One pure graph generates both checked-in tree and JSON. Named-file shuffled tests, vet,
  98.3% coverage, and exact parity verify the planning prototype without runtime success claims.
- A focused injected writer test surfaced and fixed dropped CLI status-write errors. An initial
  dependency-proxy failure resolved on an unchanged retry; cache/temp environment was preserved.

#### What Didn't Work

- Initial check/write success status ignored stdout errors.
- The original generated tree had no persistent expected JSON artifact; the first commit-hook
  formatter also changed generated bytes because key sorting and indentation differed.
- A literal duplicate-key scan missed case aliases that `encoding/json` matches
  case-insensitively. The 500-line test cap also escaped the normal docs-excluded fitness scan.
- Planning docs incorrectly deferred technical orchestration acceptance to a future user
  follow-up, conflating worker readiness with external human approval.

#### Root Causes

- Check/write discarded `fmt` writer errors. The model was treated as sufficient machine
  inventory, so the required expected JSON was omitted. The serializer did not match the shared
  formatter's sorted keys and two-space indentation.
- The duplicate scanner compared literal key casing, while `encoding/json` performs
  case-insensitive struct-field matching. Build-ignored named-file sources are absent from
  package-discovery tests/lints; planning-file limits were not explicitly enforced.
- Worker readiness was mistaken for a reason to defer technical Phase 1 acceptance. Runtime
  authorization is a separate gate; the user explicitly supplied that authorization on
  2026-11-01 without changing registry membership.

#### Patterns

- Verify every required output and exact measurement, including all 55 selector projections,
  the 84 active SQLite PVC identities, and both checked-in artifacts.
- Keep generated JSON byte-identical to repository formatting; validate semantic schema field
  names before standard-library decoding.
- Enforce line caps on every touched planning file and explicitly run tests, vet, coverage, and
  lint with the named `//go:build ignore` source/test files.
- Distinguish technical phase acceptance from human Phase 2 runtime authorization; neither a
  registry-backed scope assumption nor a completed prototype implies external approval.
- Preserve existing Compose-era handbook text during planning; reconcile it in Phase 7 only
  after actual runtime adoption, not as a Phase 1 blocker.

## Phase 2: Replace Compose Service Deployment and Test Orchestration

Authorized by the user on 2026-11-01; not started. Keep the actual registry membership (three
products) unless the user specifies otherwise. All implementation tasks remain pending.

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
