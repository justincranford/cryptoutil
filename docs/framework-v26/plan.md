# Framework v26 — Kubernetes HA Deployment and In-Process PKI CA Reload

**Status**: Kubernetes planning draft for human review; not implementation-ready.
**Created**: 2026-10-21
**Last Updated**: 2026-10-23
**Purpose**: Replace Docker Compose as the service deployment/test path with K3D/K3S and
Helm, prove real client operations across multiple service pods sharing PostgreSQL, and
define safe in-process activation of supported PKI CA configuration changes.

## Current Planning Deliverables

- [Kubernetes design, grammar, probes and migration strategy](k8s-design.md)
- [Versioned, non-secret deployment model](k8s-model.json)
- [Deterministically generated exhaustive deployment tree](k8s-deploy-tree.md)
- [Planning-only Go generator](k8s-gen.go) and [tests](k8s-model_test.go)

Only design and the planning inventory generator are authorized in the current request.
No Helm runtime, cluster bootstrap, reusable-framework code, deployment config or workflow
is implemented here. The existing PKI reload track remains future scope.

**Draft scope assumption**: use the registry's three products (sm, identity, skeleton),
treating the requested fourth as a count mismatch rather than inventing membership.
The tree covers 7 + 3 + 1 = 11 registered scopes, not a fabricated twelfth.
Adding a fourth requires an explicit identity/membership and registry/group scope change.
The current draft is handoff-ready for human review, not approved for implementation;
no deployment-runtime dependency is a planning blocker.

## Scope

### Kubernetes Deployment and Multi-Pod Verification

Compose-hosted K3S clusters and Helm charts are the target for local and CI service deployment
and end-to-end testing. Docker Compose is replaced as the application runtime and E2E
orchestration path, but retained exclusively for local K3S node infrastructure. Native K3D
is an optional alternative bootstrap adapter, not nested inside Compose. Future Kubernetes
uses the same charts with provider configuration; local testing does not certify that target.

Configuration defaults to min=2,max=2 per PS-ID/backend pool; both bounds can be changed
explicitly later, including an intentional minimum of 1. The default PG mode has two shared-state
pods per service; mixed modes have two PG plus two independent SQLite pods, preserving the
four legacy instance identities. In-memory and durable SQLite stay diagnostic, not falsely
advertised as shared-state HA. Product/suite compositions use the same configurable rules.
TLS/mTLS, secrets, all three probe types, migration and persistence intent are preserved.

Persistent-storage ownership and lifecycle MUST be explicit for every chart: identify
shared versus release-owned data, PVC retention on upgrade/uninstall, PV reclaim policy,
and backup/restore expectations. Routine pod replacement, rollout, or chart upgrade MUST
not delete a claim or silently initialize an empty database. Destructive cleanup is limited
to dedicated disposable K3S namespaces and resources owned by that test run; resource names
alone are not proof of ownership or data continuity.

The current Compose topology already starts two PostgreSQL-backed `sm-kms` app instances
using the same `postgres-url.secret`; its E2E suite checks health and TLS for both. The
PostgreSQL storage E2E also uses two independent connections/stores and verifies concurrent
CRL allocation, cross-connection store/revoke, and restart continuity. However, the app E2E
does not issue functional client API requests to both app instances. The K3D/K3S E2E must
close that precise gap: authenticated clients must exercise each pod, and state written
through one pod must be observable through another pod backed by the same PostgreSQL
database.

### In-Process PKI CA Configuration Reload

The current PKI component graph is assembled at startup. It includes the active issuer,
CRL and OCSP services, request handler, and public revocation endpoint map. See
[`pki_components.go`](../../internal/apps/sm-kms/server/pki/server/pki_components.go).
Authority and distribution settings are read from configuration files at startup; see
[`pki_authority_config.go`](../../internal/apps/sm-kms/server/pki/server/pki_authority_config.go).
The existing supported activation path for CA changes is coordinated process replacement.

This plan is for a future, explicitly triggered, in-process reload capability. It does
not assume that every PKI setting or CA key can be changed at runtime. The reload design phase
must define the reloadable setting allowlist and distinguish distribution/profile changes
from issuer certificate or signing-key rollover. Changes that conflict with persisted
authority identity or key material must continue to fail closed unless a separate,
reviewed rollover contract authorizes them.

The reload trigger is deliberately undecided. The design phase must compare supported
configuration and administration patterns before selecting a trigger. Reload is per pod
unless an explicit cluster-wide coordination contract is designed; the Kubernetes
deployment must expose partial rollout/reload state rather than imply that independently
reloaded pods are atomically synchronized.

## Safety Invariants

- Kubernetes workloads retain the current security and persistence requirements while
  Compose-based service deployment and E2E orchestration are retired.
- Every backend pool defaults to min=2,max=2 from configuration. Explicit overrides are
  validated, never silently reduced to fit local resources.
- Exact tier/scope/mode/run identity isolates names, selectors, credentials, DBs, certificates,
  storage and permitted network edges, including simultaneous all-variation deployment.
- Startup is a one-time local initialization latch; liveness never checks remote dependencies;
  readiness uses fresh bounded dependency observations through fast read-only handlers.
- SQLite instances remain independently addressable; no pooled traffic implies shared state.
- Persistent data ownership, PVC retention, PV reclaim, and backup/restore behavior are
  documented; routine pod replacement and Helm upgrades preserve persistent state.
- Destructive test cleanup is namespace-scoped to disposable K3S resources owned by that run;
  shared or persistent data is never removed based only on resource names.
- Functional E2E traffic reaches each replica, not only a load-balanced service endpoint
  that might repeatedly route to one pod.
- Cross-pod state tests use the shared PostgreSQL deployment and real supported client
  authentication; health-only success is insufficient.
- Parse and validate a complete candidate configuration and component graph before activation.
- A failed reload leaves the last known-good active generation serving requests.
- Each request uses one internally consistent generation; issuer, CRL, OCSP, handler, and
  endpoint routing must not be mixed across generations.
- Historical authorities needed to verify or serve existing certificates remain available
  after an active-authority change.
- Reload outcomes are observable without logging private keys, credentials, or other secrets.
- Existing startup validation and coordinated process replacement remain supported.

## Phases

### Phase 1: Define Kubernetes Deployment Contract

Map Compose dependencies and effective override semantics into the registry-driven grammar
and five-mode matrix in [k8s-design.md](k8s-design.md). Define owners, configurable replica
bounds, all three probes, storage/crypto continuity, local cluster lifecycle and direct-per-pod
client testing. Generate the human-review tree and JSON expected inventory using the same
pure model. Review the explicit three-product scope assumption and design before authorization.

### Phase 2: Replace Compose Service Deployment and Test Orchestration

Implement reviewed Go graph/rendering tooling and canonical Helm templates, rather than
hand-editing generated charts/values. Extend shared-framework startup/liveness/readiness
and every CLI/template surface. Implement K3S-only Compose bootstrap and Helm lifecycle/DAG
for apps, PKI, DB migrations/replication and telemetry; default bounds are 2/2.
Preserve data claims and material, and constrain destructive cleanup to verified disposable
ownership. Render/schema-check future Kubernetes provider values without claiming runtime
certification. Retire application Compose only after replacement acceptance.

### Phase 3: Verify Multi-Pod Client Behavior

Use authenticated functional clients against each pod across all seven services.
Prove that a write through one pod is visible through another, exercise shared-PG PKI
issuance/revocation and CRL allocation across pods, and verify readiness, restart,
rollout, Helm upgrade, PVC retention, and persistence behavior. Tests must distinguish each
pod rather than relying only on load-balanced health probes, and cleanup must target only
the disposable namespace and resources owned by that test run.

### Phase 4: Define PKI Reload Contract

Research configuration lifecycle, PKI component ownership, route registration, and
administrative operations. Specify reloadable fields, trigger, authorization, validation,
activation, rollback, in-flight request behavior, per-pod semantics, and historical
authority retention before implementation begins.

### Phase 5: Implement Safe Per-Pod Reload

Build and validate a replacement component generation off the request path, then activate
it as one coherent generation. Preserve the active generation on any parse, validation,
key-load, or component-construction failure. Provide an explicit and observable reload
result using the trigger selected in Phase 4.

### Phase 6: Verify and Document PKI Reload Operations

Test successful activation, rejected configuration, rollback, historical certificate
verification/revocation, and concurrent requests during reload. Document the selected
trigger, supported settings, per-pod behavior, failure recovery, and the continued
process-replacement fallback.

### Phase 7: Propagate Lessons

Update deployment, Kubernetes, testing, and PKI/configuration guidance with reusable
patterns. Record phase post-mortems and completion evidence before archiving this plan.

## Acceptance Criteria

- K3S and Helm replace application Compose; Compose remains only for local cluster nodes.
- Generated graphs cover every registered tier/mode and exact selected subsets, with names,
  network/data/trust isolation and full-matrix coexistence verified.
- Configurable min/max default to 2/2 per backend pool; no extra replica group per legacy slot.
- Shared-framework startup, liveness and dependency-aware readiness work through loopback
  authenticated exec probes, with tested handler latency and bounded dependency staleness.
- No SQLite pooled-service or multi-pod shared-state HA claim; ordinal persistence is tested.
- The E2E suite performs authenticated functional client requests against each replica; it
  proves cross-pod persistence and shared-PostgreSQL PKI behavior, not only health/TLS.
- PostgreSQL, migration, certificate, secret, readiness, and TLS/mTLS behavior pass on K3D/K3S.
- A durable record remains available across pod replacement and Helm upgrade; PVC retention
  and reclaim behavior match the documented policy, and test cleanup cannot delete shared data.
- The reloadable setting allowlist and explicit non-reloadable settings are documented.
- A failed candidate cannot partially replace live issuer, CRL, OCSP, handler, or endpoint state.
- In-flight requests observe one complete generation, and concurrent reload/request tests pass
  under the race detector.
- Existing and historical authority behavior remains correct after activation.
- Reload success/failure and per-pod rollout state are actionable to operators and do not disclose
  secret material.
- Normal and tagged builds, focused tests, repository validators, and applicable CI quality gates
  pass; operations documentation describes both deployment and reload paths.

## Planning Verification

The planning generator's shuffled named-file tests pass at **99.1% statement coverage**;
`go vet`, exact generated-tree comparison and repository `lint-text` pass.
The identity mixed-disk CLI JSON preview contains 1 release, 10 pools, 20 pods and 5 DBs.
Full-matrix and subset-projection tests validate 55 releases, 147 pools, 294 app pods and
63 PostgreSQL DBs. Repeated generation produces identical bytes.
Evidence: [planning validation log](../../test-output/framework-v26-k8s-planning-evidence.log).
These are planning/prototype checks, not Helm, Docker, Kubernetes or framework runtime tests.

## References

- [PKI component construction](../../internal/apps/sm-kms/server/pki/server/pki_components.go)
- [PKI authority configuration](../../internal/apps/sm-kms/server/pki/server/pki_authority_config.go)
- [sm-kms Compose application services](../../deployments/sm-kms/compose-apps.yml)
- [sm-kms Compose E2E tests](../../internal/apps/sm-kms/e2e/)
- [PostgreSQL cross-connection PKI E2E](../../internal/apps/sm-kms/server/pki/storage/e2e/postgres_continuity_e2e_test.go)
- [TLS structure and rotation constraints](../tls-structure.md)
- [ENG-HANDBOOK.md §6 Security Architecture](../ENG-HANDBOOK.md#6-security-architecture)
- [ENG-HANDBOOK.md §8 API Architecture](../ENG-HANDBOOK.md#8-api-architecture)
- [ENG-HANDBOOK.md §10 Testing Architecture](../ENG-HANDBOOK.md#10-testing-architecture)
- [ENG-HANDBOOK.md §11 Quality Strategy](../ENG-HANDBOOK.md#11-quality-strategy)
- [ENG-HANDBOOK.md §14 Coding Standards](../ENG-HANDBOOK.md#14-development-practices)
