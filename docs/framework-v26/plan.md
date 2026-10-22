# Framework v26 — Kubernetes HA Deployment and In-Process PKI CA Reload

**Status**: Parked for future planning; not implementation-ready.
**Created**: 2026-10-21
**Purpose**: Replace Docker Compose as the service deployment/test path with K3D/K3S and
Helm, prove real client operations across multiple service pods sharing PostgreSQL, and
define safe in-process activation of supported PKI CA configuration changes.

## Scope

### Kubernetes Deployment and Multi-Pod Verification

K3D-managed K3S clusters and Helm charts are the target for local and CI service deployment
and end-to-end testing. Docker Compose is to be replaced as the service runtime and E2E
orchestration path, not retained as a parallel supported deployment path.

Helm chart defaults MUST ensure at least two pods for each of the seven PS-ID services.
When autoscaling is enabled, its minimum replica count must remain at least two. Each
service's pods must use the configured shared PostgreSQL deployment and preserve existing
TLS/mTLS, secrets, readiness, migration, and persistence requirements. Product and suite
charts must preserve this invariant for every included service.

Persistent-storage ownership and lifecycle MUST be explicit for every chart: identify
shared versus release-owned data, PVC retention on upgrade/uninstall, PV reclaim policy,
and backup/restore expectations. Routine pod replacement, rollout, or chart upgrade MUST
not delete a claim or silently initialize an empty database. Destructive cleanup is limited
to dedicated disposable K3D namespaces and resources owned by that test run; resource names
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
not assume that every PKI setting or CA key can be changed at runtime. The first phase
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
- Every service chart defaults to at least two ready replicas; no chart or product/suite
  overlay may silently reduce that minimum.
- Persistent data ownership, PVC retention, PV reclaim, and backup/restore behavior are
  documented; routine pod replacement and Helm upgrades preserve persistent state.
- Destructive test cleanup is namespace-scoped to disposable K3D resources owned by that run;
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

Map the existing Compose service, shared PostgreSQL, telemetry, certificate, secrets,
migration, health-check, and persistent-storage dependencies into K3D/K3S and Helm. Define
chart ownership, the default two-pod floor for all seven PS-ID services, product/suite
composition, PVC retention and PV reclaim policies, backup/restore expectations, the
local/CI cluster lifecycle, and safe direct-per-pod client access for E2E while preserving
production TLS/authentication behavior.

### Phase 2: Replace Compose Service Deployment and Test Orchestration

Implement Helm deployment for all seven services with a minimum of two pods per service,
shared PostgreSQL connectivity, persistent storage where required, and existing TLS/mTLS
and secret handling. Routine upgrades and pod replacement must preserve data claims;
destructive cleanup must be limited to run-owned disposable namespaces. Replace
Compose-driven service startup and E2E orchestration with K3D/K3S workflows; remove Compose
as a parallel supported service deployment path.

### Phase 3: Verify Multi-Pod Client Behavior

Use authenticated clients against each `sm-kms` pod and at least one other service pair.
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

- K3D/K3S and Helm replace Docker Compose for service deployment and E2E orchestration.
- Helm defaults and product/suite charts guarantee at least two pods for every PS-ID service.
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
