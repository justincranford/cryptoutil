# Framework v26 — Kubernetes HA Deployment and In-Process PKI CA Reload

**Status**: Phase 1 lifetime/placement closure is complete. Phase 2 is user-authorized and
underway, but no framework integration checkpoint or runtime acceptance exists. The accepted
isolated tooling checkpoint is `e9f312cfb214eb9bf1fed49e2ffdb617ca93d537`
(`fix(tooling): pin supported Semgrep runtime`): it changes only `.pre-commit-config.yaml` to
pin Python 3.14+ and Semgrep 1.183.0; supported Semgrep and normal configured commit hooks passed,
and its isolated worktree was clean. It weakened no source/policy and bypassed no hook. The active
scanner worker owns the next isolated eight-file checkpoint; the five older component workers are
frozen/idle, and no primary integration is active. Scanner work does not establish root integration.
The latest root `TestLint_Integration` run failed at its 600s timeout in
`require-framework-naming` filesystem walking; a later rerun was aborted with exit 137. The old
helper was explicitly killed and destroyed, which is not proof of OOM. The focused package test
passes in 5.768s after changing per-entry `Lstat` to `WalkDir`; existing artifact pruning remains,
and a regression confirms `test/` sources are still scanned. Neither result establishes root
integration or a broad-gate pass. Earlier broad validation predates substantial source changes
and is historical. An earlier disk read found 940,208,128 bytes free on `R:` and 621,630,750,720 bytes
free on `C:`; later reads found about 397 MiB free on `R:`. The parent then verified 17,607,114,752
bytes free and removed 48 explicitly inventoried stale `go-build*` directories under
`R:\Go\GOTMP`, reclaiming 62,537,728 bytes; no active build processes, reparse points, caches,
modules, binaries, or other temporary files were removed. Go is 1.27.3 on Windows/amd64 with
inherited `CGO_ENABLED=0`. Historical normal and
tagged full-repository build passes used process-scoped Go cache/temp overrides; they are not current
gates. The user prohibits changing any inherited Go/tool environment values, including `GOCACHE`,
`GOTMPDIR`, `TMP`, `TEMP`, `CGO_ENABLED`, and `GOTOOLCHAIN`. No shared caches, persistent
environment variables, helper container, or global Docker resources were changed. The Docker host
reports 14 CPUs and 8.33 GB total. Locked Helm v4.3.1 is installed under ignored
`test-output/framework-v26-toolchain-helm-v4.3.1/` but was not added to PATH. The repository does
not pin K3D; the selected Compose-hosted target uses locked K3s, so no unpinned K3D was installed.
Production graph/tooling, canonical charts, typed projections, protected import, lifecycle-aware
admission and a node-only Compose template have prerequisite validation described below.
Full rendered lifecycle parity is validated; PKI/migration/ownership execution acceptance, framework integration, runtime, workflow, Compose
cutover, and the main-agent postmortem remain incomplete. No app image or cluster deployment has
started; no Phase 2 integration/runtime checkpoint or push exists.

Earlier broad validation ran in the owned clean Linux validator while `R:` was reported
constrained; the latest read confirms that `R:` remains constrained. No host cache/temp paths,
Docker VM allocation, or container CPU quota were changed. Earlier
normal/tagged builds, the full shuffled suite (333 packages passed; 37 have no
test files), normal/tagged `golangci-lint` and `go vet`, all 52 deployment validators, `lint-go`,
`lint-fitness`, `lint-text`, and `lint-docs` passed before the latest source changes; these are
historical, not current gate evidence. `deploymentgen` statement coverage was 99.0%. An earlier
production CLI run passed chart-mirror, infrastructure-values, and graph checks; Helm 4.3.1
strict lint/render passed for generated mirrors with canonical infrastructure
values and the supported runtime fixture. Real configured pre-commit hooks passed for the
isolated tooling checkpoint; framework implementation commits and a clean integration checkpoint
remain pending. Capacity estimation rejects the full matrix against the measured host budget
after explicit reserves and accepts a selected PostgreSQL projection, but no live host probe or
scheduler admission has run. No runtime/lifecycle acceptance is claimed.
**Created**: 2026-10-21
**Last Updated**: 2026-11-03

## 2026-11-03 Lifecycle Prerequisite Continuation

The earlier identity-only lifecycle notes below are historical. The typed lifecycle operator
contract, restricted Job/RBAC chart, file-backed in-memory importer, content revisions, Job quota
headroom, generated third-chart mirrors, full-registry combined parity, and rendered Secret
reference checks are now implemented. Prior closure evidence is recorded in `tasks.md`.

This continuation implements a protected physical source boundary and actual CLI import-check:
clean absolute root, rooted filesystem, checked handles, strict current-user ownership,
Windows DACL or Unix mode enforcement, no symlink/reparse/hard-linked files, bounded sizes,
classified errors, and no Secret output/persistence. The source owner must not concurrently
replace material; production must not repair permissions on the operator's behalf.

Physical host admission must require the explicit lifecycle input and include all selected Jobs
concurrently with maximum application/init and infrastructure requests plus explicit reserves.
Quota projection alone cannot prove host capacity. Existing read-only resource-collision and
port checks remain mandatory. The operator must provide pins, protected material, commands,
configuration, and a model revision matching the non-secret contract.

The orthogonal execution worker owns only new `deploymentrun/lifecycle` transport/DAG files.
The outer `lifecycle-execute` CLI now loads current rooted projections, gates current lifecycle
budgets, performs protected physical imports, and invokes that adapter without a dependency cycle.
It requires explicit owned target paths/context/policy and prior-admission operator attestation,
not a live host probe. The real Windows filesystem/ACL/schema/admission/dispatch pipeline passes
with only the final executor injected. Combined validation exposed scheduling-sensitive short
test deadlines before Prepare; the owner corrected test policy budgets while preserving real
timeout cases and asserting the call was reached. Post-fix combined shuffled repeated tests,
subsequent focused timeouts, tagged coupled checks, scoped lint/vet and artifact checks pass;
current coverage is deploymentgen 98.1%, outer CLI 98.8%, adapter 98.5%. `tasks.md` records
the exact remaining operator inputs and runtime command/publication acceptance requirements.
No cluster is started and no runtime acceptance is claimed.
Inherited staged tracking/source work is preserved; no commit slot has been authorized.

## 2026-11-03 Host PKI Publication

The production `pki-publish` host bridge now validates existing protected pki-init outputs
against the unchanged materializer consumer policy, constructs exact graph-bound materializer
labels/key bundles in memory and uses the existing lifecycle namespace/verify-or-create client.
An explicit non-secret manifest binds every reserved pod to approved DNS/PG identities and
same-namespace imported trust anchors. The command retains explicit admission/target/policy
inputs and performs no key generation, secret rewriting, file staging or broader RBAC.
The supported suite generation command uses `--domain`/`--output-dir`; test fixtures no longer
pretend positional PKI arguments are executable.

This is an explicit pre-Job host prerequisite, not an in-Job publisher or replacement for the
existing completion DAG. In-Job Secret create remains forbidden; Job Complete is not a
certificate-output attestation. The remaining consumer binary/chart contract (absent entry,
token path mismatch, missing init trust and identity inputs) and executable database migration/
replication adapters are missing production wiring, separate from explicit operator certification.
Current new publication functions have 100% coverage; four owning packages pass shuffled
repeated tests at 99.5%/98.1%/98.3%/98.5%. Evidence and limitations are recorded in tasks.md
and the lifecycle contract. Runtime acceptance and safe parent-owned integration remain open.

**Purpose**: Replace Docker Compose as the service deployment/test path with K3D/K3S and
Helm, prove real client operations across multiple service pods sharing PostgreSQL, and
define safe in-process activation of supported PKI CA configuration changes.

## Current Planning Deliverables

- [Durable Kubernetes deployment contract, grammar, probes and migration strategy](../../docs/kubernetes-deployment-contract.md)
- [Versioned, non-secret deployment model](../../api/cryptosuite-registry/k8s/k8s-model.json)
- [Deterministically generated exhaustive deployment tree](../../api/cryptosuite-registry/k8s/k8s-deploy-tree.md)
- [Deterministically generated machine-readable expected inventory](../../api/cryptosuite-registry/k8s/k8s-expected-inventory.json)
- Production Go generator under
  [`internal/apps-tools/deploymentgen`](../../internal/apps-tools/deploymentgen/), with a thin
  [`cmd/cicd-deployment`](../../cmd/cicd-deployment/) entry, typed model schema, pinned local
  toolchain inputs, registry-derived graph, typed runtime and infrastructure chart-values
  projections, canonical runtime/infra Helm charts, generated chart mirrors, and K3s-node-only
  Compose template. The
  obsolete build-ignored prototype was removed after promotion. Focused Go tests validate the
  runtime projection against its strict schema across all 55 releases and the infrastructure
  projection against its schema for full and SQLite-only inventories. Registry-derived API edge
  projection and negative dependency validation are also covered. Production CLI chart-values
  preview/write/check is now implemented and focused-tested. The production `charts` mode writes
  and checks exact canonical chart mirrors. An identity-only, non-deployable lifecycle inventory
  now projects external Secret, PKI/DB Job, and bootstrap RBAC identities from the validated
  graph. A focused combined parity check joins those identities to chart-rendered runtime and
  infrastructure objects and controller-created PVC identities for selected PostgreSQL and
  SQLite releases plus full infrastructure. Actual Secret provisioning, restricted Job specs and
  execution, bootstrap RBAC policy/rendering, full-matrix combined parity, complete
  lifecycle/orchestration, and runtime acceptance remain pending. Helm integration tests now
  render all 55 runtime releases and match chart-owned object identities to the independent
  graph; infrastructure chart-owned identities match for full and SQLite-only inventories.
  The typed infrastructure projection rejects a collector replica mismatch against canonical
  chart defaults and PostgreSQL pod identities in a SQLite-only selection. Earlier focused
  package coverage was 99.0%; the capacity estimator additionally tests projection mismatches, malformed
  resources, arithmetic overflow, max-replica budgets, and explicit host reserves. Both canonical
  charts passed strict Helm lint when supplied their supported values inputs. An earlier
  standalone Linux CLI run passed exact chart-mirror, generated infrastructure-values, and graph
  checks; it predates the latest source changes and is not current integration evidence. The
  An earlier native Windows CLI-link attempt failed for insufficient `R:` scratch space. At the
  current checkpoint, the full normal build also failed while resizing linker output on `R:`; a
  tagged full-repository build has not passed under the unchanged inherited environment.
- [Phase 1 validation evidence](../../test-output/framework-v26-phase1-evidence.log)

Phase 1 completion is technical planning acceptance only; it is not runtime certification or
approval of a new product. The user explicitly authorized Phase 2 on 2026-11-01 with: “Then
implement phase 2, including post-mortem and fixes at the end. Ensure all work is committed.”
Phase 2 implementation is underway, but the isolated tooling checkpoint is not a clean framework
integration checkpoint or runtime acceptance. The framework worker owns existing shared application
lifecycle/configuration files; the runtime owner must coordinate any typed migration policy
interface without changing those files. The existing PKI hot-reload track remains future scope.

**Draft scope assumption**: use the registry's three products (sm, identity, skeleton),
treating the requested fourth as a count mismatch rather than inventing membership.
The tree covers 7 + 3 + 1 = 11 registered scopes, not a fabricated twelfth.
Adding a fourth requires an explicit identity/membership and registry/group scope change.
The three-product set is registry-backed, not a claim of human approval. Phase 1 is complete
for planning; no deployment-runtime dependency is a planning blocker.

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
and five-mode matrix in the
[durable deployment contract](../../docs/kubernetes-deployment-contract.md). Define owners, configurable replica
bounds, all three probes, storage/crypto continuity, local cluster lifecycle and direct-per-pod
client testing. Generate the human-review tree and JSON expected inventory using the same
pure model. Keep the three-product scope assumption explicit; it is registry-backed, not user
approval. The main-agent technical postmortem and Phase 1 acceptance are complete. Close the
prototype's lifetime/placement gap by keeping reusable code and planning inputs outside ephemeral
tracking directories. The four-section lifetime/placement postmortem is in
[lessons.md](lessons.md#phase-1-lifetimeplacement-postmortem). Phase 2 is user-authorized and
underway; its runtime acceptance remains incomplete.

### Phase 2: Replace Compose Service Deployment and Test Orchestration

The user explicitly authorized Phase 2 on 2026-11-01 with: “Then implement phase 2, including
post-mortem and fixes at the end. Ensure all work is committed.” Authorization and partial
implementation progress are recorded separately. The current generator/chart/bootstrap work is
not runtime certification, does not include app-image deployment, and is not the required clean
framework-merge checkpoint.

The production graph/CLI and model schema have been promoted; canonical runtime and infrastructure
charts, typed runtime/infra projections, a toolchain lock, K3S-only Compose template, and pure
request-capacity estimator/probe are partial foundation work. A deterministic identity-only
lifecycle inventory and runtime ResourceQuota projection/template now close selected identity
gaps; neither is full lifecycle provisioning. Remaining work includes provider-configured chart
values and actual combined lifecycle-resource specifications/parity. Controller-created PVC identities
match the graph, while actual capacity/lifecycle execution and ownership controls, restricted PKI
and migration helpers,
framework-owned startup/drain integration, actual post-merge K3S acceptance, and only then
Compose cutover. Default bounds remain 2/2. Preserve data claims and material, constrain
destructive cleanup to verified disposable ownership, and render future Kubernetes provider
values without claiming runtime certification. Keep workflow integration and the final
postmortem with the main agent after interfaces stabilize.

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

After actual runtime adoption and validation, reconcile Compose-era handbook assumptions and
update deployment, Kubernetes, testing, and PKI/configuration guidance with source-backed
patterns. This permanent propagation is intentionally deferred to Phase 7, not a Phase 1
blocker. Record phase post-mortems and completion evidence before archiving this plan.

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

## Historical Phase 1 Planning Verification

Named-file shuffled planner tests pass at **98.3% statement coverage**, with named-file `go vet`.
The generator checks exact bytes for both the exhaustive tree and full expected JSON; repeated
`-write` runs produced identical SHA-256 digests. Tests project all 55 selectors from the same
full graph and cover strict JSON/YAML input, expanded DNS/SQL identity constraints, capacity
preflight, isolation, scaling, and CLI failures. The identity mixed-disk CLI preview is
1 release / 10 pools / 20 pod slots / 5 databases. The full expected inventory is 55 releases /
147 pools / 294 maximum pod slots (294 active at default min=max=2) / 63 PostgreSQL databases
on each logical-replication node / 84 active SQLite claims; it contains Secret names only.
Repository `go build ./...`, tagged build, shuffled full tests, `lint-text`, `lint-docs`,
`lint-fitness`, and `lint-go` passed. The build-ignored planner also passed scoped
`golangci-lint` when invoked with its exact named source/test files; package discovery alone does
not inspect it. Applying `--build-tags=ignore` globally is invalid because it selects unrelated
ignored standard-library packages. Exact commands, digests, scoped-lint findings, and the
initial transient module-proxy failure/recovery are recorded in
[the Phase 1 evidence log](../../test-output/framework-v26-phase1-evidence.log).
These remain planning/prototype checks, not Helm, Docker, Kubernetes or framework runtime tests.

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
