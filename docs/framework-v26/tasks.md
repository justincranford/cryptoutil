# Framework v26 — Tasks

**Status**: Phase 1 lifetime/placement closure is complete. Phase 2 was authorized on 2026-11-01
and remains in progress, with no framework integration checkpoint or runtime acceptance. The
accepted isolated tooling checkpoint is `e9f312cfb214eb9bf1fed49e2ffdb617ca93d537`
(`fix(tooling): pin supported Semgrep runtime`): its sole file is `.pre-commit-config.yaml`,
pinning Python 3.14+ and Semgrep 1.183.0. Supported Semgrep and normal configured commit hooks
passed, its isolated worktree was clean, and it weakened no policy or bypassed hooks. The active
scanner worker owns the next isolated eight-file checkpoint; this is not root integration. The
latest root `TestLint_Integration` run failed at its 600s timeout in `require-framework-naming`
filesystem walking; a later rerun was aborted with exit 137. The old helper was killed and
destroyed, which is not proof of OOM. The focused naming-package test passes in 5.768s after the
traversal change, but does not replace root integration or broad gates. Earlier broad passes
predate substantial source changes and are historical. An earlier disk read found 940,208,128
bytes free on `R:` and 621,630,750,720 bytes free on `C:`; the parent later reported 493,735,936
bytes free on `R:` (~471 MiB), and the adapter-test worker subsequently observed about 397 MiB
free. The parent then verified 17,607,114,752 bytes free and removed 48 explicitly inventoried
stale `go-build*` directories under `R:\Go\GOTMP`, reclaiming 62,537,728 bytes; no active build
processes, reparse points, caches, modules, binaries, or other temporary files were removed.
Go is 1.27.3 on Windows/amd64 with inherited `CGO_ENABLED=0`. Historical normal/tagged
full-repository build passes used process-scoped
`GOCACHE`/`GOTMPDIR` overrides; they are not current evidence. The user prohibits changing any
inherited Go/tool environment values, including `GOCACHE`, `GOTMPDIR`, `TMP`, `TEMP`,
`CGO_ENABLED`, and `GOTOOLCHAIN`. No shared Go caches, persistent environment variables, or global
Docker resources were changed. The Docker host reports 14 CPUs and 8.33 GB
total; locked Helm v4.3.1 is installed under ignored
`test-output/framework-v26-toolchain-helm-v4.3.1/` but was not added to PATH. The repository does
not pin K3D; the selected local target runs pinned K3s directly through Compose, so no unpinned
K3D binary was installed. The helper is still available and untouched. The primary/framework line remains at `491cfd7fa` with unaccepted
implementation; no primary integration is active. The five
older component workers are frozen/idle. Graph/chart/projection, protected-import and admission
prerequisites have focused evidence below; no K3S/runtime, rollback, workflow, app-image, or
cluster acceptance is claimed. No
app image or cluster deployment has started.
There is no Phase 2 integration/runtime checkpoint or push. The registry-backed three-product
scope is unchanged.

**2026-11-03 pre-cleanup root validation**: `go build ./...` and a serialized
`go build -p 1 ./...` both failed while linking multiple command binaries, with the toolchain
reporting `resize output file` under its configured `R:\GO\MOD_CACHE` path. `R:` had
656,957,440 bytes free at inspection; `go env` reported the inherited cache and temporary paths
on `R:`. No Go/environment variables, persistent configuration, or shared caches were changed.
The focused `go test ./internal/apps-framework/service/server/readiness` passed (10 tests).
`docker ps` reported no running containers; Helm and K3D were not found on PATH at that time.
The normal-build link failure was later resolved by parent-authorized cleanup of explicitly
inventoried stale Go scratch directories; see the current post-cleanup result below.

**2026-11-03 historical identity-only lifecycle snapshot (superseded below)**: After the parent-authorized scratch cleanup, the
final seven-service startupz adapter test command passed (35 tests), and
`go test ./internal/apps-tools/deploymentgen -shuffle=on -count=1` passed (588 tests). Focused
combined Helm parity and min/max replica integration tests passed; the `sm-kms` PostgreSQL
lifecycle preview reported 13 external Secret identities, 5 lifecycle Job identities, and 6
bootstrap RBAC identities. The final shuffled deploymentgen coverage run passed at 98.8%
statement coverage; the new quota calculation and memory formatter are each at 100%, and the
focused eight-package golangci-lint run reported zero new issues. Generated runtime/infrastructure
charts passed strict Helm lint, direct render, and mirror `-check`. The lifecycle inventory is
identity-only and non-deployable; actual external Secret provisioning, PKI/DB Job
specifications/execution, bootstrap RBAC rendering, and runtime acceptance remain open.
Post-cleanup `go build ./...` and
`go build -tags e2e,integration ./...` both passed with the inherited environment unchanged.
The broad repository test suite and broad linters have not been run in this bounded slice.
**Created**: 2026-10-21
**Last Updated**: 2026-11-03

**2026-11-03 deployable lifecycle prerequisite and protected-import continuation**:
The lifecycle slice now has an exact-graph typed operator contract, explicit digest-pinned
suite Job inputs, bounded resources/deadlines, restricted named Secret RBAC, content-derived
revisions, and a third canonical lifecycle chart with generated mirrors. Actual in-memory
fixture Secret imports plus rendered lifecycle/runtime/infrastructure objects and controller
PVCs match the independent full registry, PG/SQLite selections, and min=1/max=3 graph.
Rendered mount/key references resolve to imported Secrets; forged chart permissions are rejected.
The prior closure passed full normal/tagged builds with inherited environment unchanged,
785 normal cases across deploymentgen/CLI, 142 focused tagged cases, zero skips,
deploymentgen 98.3% / CLI 100% coverage, scoped normal/tagged lint, all 52 current deployment
validators, and chart/tree reproducibility checks. These are prerequisite checks, not broad
repository tests or runtime acceptance.

The current continuation adds `ImportPhysicalLifecycleSecrets` using `os.Root` confinement,
opened-handle identity/type checks, current-user ownership, no symlink/reparse/hard-linked
files, bounded reads, and explicit classified failures. Windows validates the opened handle's
DACL (owner and SYSTEM only); portable mode bits are normalized only after ACL verification,
not used as an ACL substitute. Unix enforces owner-only modes and current effective UID.
Production import never changes permissions. `lifecycle -format import-check -secret-root`
reads protected material into memory and emits only a validated count, never Secret bytes,
files, or cluster mutations. Native Windows ACL and actual public CLI positive/negative checks
passed 33 focused cases; current combined continuation evidence is recorded below.

Admission now requires explicit lifecycle operator input and conservatively includes every
selected Job request alongside maximum application/init and infrastructure requests and
explicit host/K3s reserves; quota headroom and physical host budgeting are separate checks.
The existing read-only Docker ownership/host/port probes are retained. The coupled preflight
and physical checks passed 52 cases, and focused Job-budget/import CLI boundaries passed
20 cases. The orthogonal execution worker completed only new `deploymentrun/lifecycle`
apply/wait DAG files; the outer CLI handoff is now integrated.

**Current coupled execution/import/admission validation**: full normal and tagged builds
passed after production CLI wiring. The three owning packages passed shuffled `-count=3`
tests with deploymentgen 98.1%, outer CLI 98.8%, and execution adapter 98.5% coverage in
`test-output/coverage-analysis/lifecycle-physical-admission.out`. A combined run exposed
adapter test deadlines expiring before Prepare under parallel package scheduling, followed
by unsafe last-call indexing. The owner increased only non-timeout test budgets, retained
real operation/Job deadline waits, rebuilt the one-second Job's revision/inventory, and added
an explicit nonempty-call assertion. The combined repeated pass is after this fix; earlier
456-check worker evidence predates it. A subsequent focused 12-case check also passed,
including actual timeout behavior. Current focused tagged lifecycle/bootstrap/execution/
kubectl tests, scoped normal/tagged lint and vet, and production chart/tree `-check` pass.
No broad repository test/lint or runtime acceptance is claimed.

`lifecycle-execute` now requires explicit owned paths/context/durations and prior-admission
operator attestation; it revalidates current bindings/revision/quota/request totals before
protected import and guarded execution. The real native Windows filesystem/ACL/schema/
admission/outer-dispatch test injects only the final executor; arbitrary downstream errors
and private bytes never become CLI output. The flag does not authenticate a host measurement
artifact or reprobe Docker/reserves/ports after prerequisites are installed.

**Concrete remaining runtime requirements**:

- The checked-in model remains `review`/`r1`; approve an owned run/selection and explicitly
  update its revision from the supplied canonical lifecycle operator input. No ready-to-run
  lifecycle or runtime input JSON exists under the inspected api/configs/deployments paths.
- Existing toolchain/infra image pins do not supply certified application, suite Job or PKI
  materializer image digests, per-service framework/domain configs, executable probes/drain
  commands, runtime resources/storage, or exact lifecycle command/Secret/ConfigMap bindings.
- Supply a current-user-private material directory with every selected authority/client/TLS/
  credential data key in its required format. File/privacy checks do not certify cryptographic
  validity, endpoint credentials or trust relationships.
- Supply certified absolute tools/repository/kubeconfig, an explicit owned context, policy
  durations and reserves, then perform fresh host admission before installing owned Active
  namespaces, ConfigMaps, lifecycle-aware quotas and infrastructure StatefulSets.
- Certify the actual pinned commands. Current suite `pki-init` requires `--domain` and
  `--output-dir` and publishes filesystem certificates; merely naming it or using positional
  test fixture arguments does not implement Kubernetes Secret publication. Union database,
  migrations and replication need real supported per-service command behavior, not synthetic
  test argument labels. The generic executor intentionally cannot invent those adapters.
- Execute on the owned target and prove PKI publication/trust, leader/follower readiness,
  union database/migration/replication semantics, multi-pod authenticated client operations,
  rollback and the remaining framework/workflow acceptance gates. These are still open.

**2026-11-03 production host PKI publication continuation**:
The missing filesystem-output-to-Secret bridge is now implemented, not an operator-answer
blocker. `cicd-deployment pki-publish` uses the same actual rooted loader, lifecycle-aware
admission and physical importer, plus an explicit non-secret publication manifest covering
every reserved runtime pod. It reuses the unchanged materializer certificate/key/trust/SAN/
EKU/PostgreSQL CN policy, rejects incomplete/forged/oversized material and conflicting labels,
adds the exact materializer role/ownership labels, and uses the lifecycle client's namespace
guard and verify-or-create Secrets. No bytes, crypto policy, permissions, CA keys or credentials
are changed; no bridge temporary material, logging, overwrite or cluster launch occurs.

The typed lifecycle contract now accepts materializer Opaque ordinal `*-tls` bundle keys
(`public.crt/key`, `admin.crt/key`, PG `postgres.crt/key`) alongside distinct legacy TLS
resources; the consumer still enforces exact role keys. PKI fixtures now use real supported
`pki-init --domain=... --output-dir=...` flags rather than positional placeholder arguments.
Full graph/injected-renderer publication tests prove create-only first publication and exact
existing-data verification; namespace/prepare/transport failures reject. The independent
materializer builder tests cover malformed certificates, wrong keys/labels/types/identities,
missing roots/keys, oversized material and no input alias mutation.

Current coupled four-package shuffled `-count=3` tests pass with CLI 99.5%, deploymentgen
98.1%, materializer 98.3% and execution 98.5% coverage at
`test-output/coverage-analysis/pki-publication.out`; every new publication function is 100%.
Full normal/tagged builds, focused tagged coupled tests and normal/tagged vet passed.
Final scoped normal/tagged lint reports zero findings after the test-only constant/whitespace
correction; focused publication/native CLI checks and chart/tree consistency checks also pass.
No inherited index work is staged, committed or treated as owned checkpoint permission.

Host publication is explicitly ordered after fresh admission/owned prerequisites and before
the existing infra PKI -> release PKI -> DB/migration/replication completion DAG. It is not
a new PKI Job completion attestation: the executor still observes owned Job Complete/failure/
UID barriers and does not prove Job-generated bytes equal the immutable host publication.
The in-Job writer alternative is not implemented and no named bootstrap create restriction
was broadened. Subsequent lifecycle imports must exactly match the published bytes/key sets.

Remaining missing **code**, not generic missing operator answers: `cmd/pki-materializer` does
not exist despite the canonical `/app/pki-materializer materialize` command; its init chart
mounts token/CA at `/var/run/secrets/kubernetes.io/serviceaccount` while the consumer library
reads `/run/cryptoutil/pki/input/{token,ca.crt}`, and the init lacks issuer trust mounts and
explicit required DNS/PostgreSQL binding inputs. That consumer/chart boundary needs separate
typed wiring; real database migration/replication executable adapters remain separate work.
Operator pins/material/owned target inputs are still required, but do not substitute for
these code surfaces. Stock pki-init Unix permissions also require explicit source protection;
the bridge intentionally does not repair operator files.

Real operator pins/material/commands remain required inputs. The source owner must not mutate
protected files concurrently. Unix code is not runtime-certified by Windows tests. Race checks
remain unavailable with inherited CGO disabled; no override is authorized. No cluster, app image,
runtime acceptance, tracking/index checkpoint, staging, or commit is claimed.

**2026-11-03 continuation**: Shared framework lifecycle work now includes startupz CLI/router
support, bounded health requests, a startup/readiness monitor, DB schema/ping and verified public
listener observations, an mTLS receiver-health check when OTLP is enabled, sanitized admin
snapshots, and graceful listener drain ordering. OTLP now requires an explicit HTTPS receiver
health URL and CA/client certificate/key settings. The 21-package server/CLI/usage/config shuffled
slice passes 2,746 tests; the focused config/readiness/listener/server slice passes 494 tests in
four packages, and the listener package passes another 165 tests after the latest status-literal
cleanup. The historical full-repository build pass used process-scoped cache/temp overrides and is
not current gate evidence or permitted to repeat. Framework-slice `golangci-lint run --new` is
clean; full lint still
reports 15 existing `wrapcheck` findings and one existing `errcheck` finding in untouched files
(the tagged run reports the same 15 wrapcheck findings): `businesslogic/session_manager_jwe.go`,
`apis/sessions.go`, `repository/tenant_repository.go`, and `public_server_base_test.go`. No
full-repository shuffled test run,
full-repository lint, deployment-validator rerun, framework integration checkpoint, K3S/runtime
execution, all-seven-service integration, upgrade/cleanup acceptance, or Compose cutover has
passed. A normal-hook path-limited commit attempt for the parent-owned linter fix was rejected
because the hook saw `.pre-commit-config.yaml` as unstaged in the temporary path-only index. No
hook was bypassed and no commit was created; inherited staged files and ownership-unclear
docs/config changes remain untouched.

**2026-11-03 listener startup follow-up**: Review found that the listener `Started()` accessors
returned nil before bind, while `Application.Start` interprets a nil channel as already started.
Both accessors now lazily initialize and return a stable pending channel under the listener mutex;
the existing idempotence test verifies it stays open until signaled. `go test -shuffle=on -count=1`
passed 519 tests across config, readiness, listener, and server packages. Targeted normal/tagged
builds, `go vet`, and focused listener/readiness golangci-lint passed. Full-repository builds remain
unavailable with the inherited `R:` scratch constraints; no Go/environment values, caches, or
Docker resources were modified. Runtime/cluster acceptance and tagged full-repository build remain
unverified. The shared worktree still contains 360 changed entries (336 modified, 24 untracked),
including unrelated staged/unstaged work; no paths were staged and no commit was made.

**2026-11-03 Helm and disk follow-up**: Installed locked Helm v4.3.1 from the official Windows
archive into ignored `test-output/framework-v26-toolchain-helm-v4.3.1/`, after verifying its
SHA-256 against `api/cryptosuite-registry/k8s/k8s-toolchain.lock.json`; no PATH or environment
setting changed. Strict lint and direct renders passed for the runtime fixture and infrastructure
full/SQLite-only selections after correcting two fixture mismatches. The runtime fixture had
non-contract SAN expectation fields rejected by its strict values schema; those external
lifecycle expectations were removed from the chart-only fixture. The SQLite-only fixture had
inherited the default non-empty PostgreSQL database catalog; explicitly setting `databases: []`
made its render match the disabled PostgreSQL projection. `helm lint` alone passed despite that
template failure, so direct render remains required. K3D was not installed: the lock pins Helm,
K3s image and kubectl but no K3D version, and the selected local target runs pinned K3s directly
through Compose; do not introduce an unpinned optional adapter.

The authorized seven-package adapter test command passed five packages; `identity-idp` and
`skeleton-template` failed while linking `.test.exe` files with `resize output file` /
“There is not enough space on the disk” in the configured `R:` Go temporary area. At the worker's
check, `R:` had about 397 MiB free. Earlier full normal builds likewise failed resizing linker
outputs on `R:` (including under `R:\GO\MOD_CACHE`); the tagged full-repository build remains
unverified. These are storage/linker failures, not demonstrated source compile failures. No
environment values, shared Go caches/temp files, or Docker resources were changed or deleted.
The adapter owner was instructed not to retry until the owner resolves disk capacity. No staging
or commit was performed; the shared index/worktree still includes unrelated inherited changes.

**2026-11-03 chart mirror follow-up**: The production chart parity check found the runtime mirror's
`values.schema.json` stale relative to its canonical schema, which already constrained the PKI
materializer command/arguments. Regenerated the mirror only through
`go run ./cmd/cicd-deployment -format charts -write`; the production `-check` then passed. Strict
Helm lint and direct runtime/infrastructure renders also passed against the generated mirror copies,
including the full and SQLite-only infrastructure fixtures. The remaining chart parity gap is
combined chart/lifecycle-object comparison, not canonical-to-generated file drift.

**2026-11-03 startupz adapter review**: The bounded adapter slice changed seven service entry
points and tests plus the two canonical app templates (16 files total). The help assertions use
shared product/service magic constants and verify endpoint, usage, TLS flags, and output stream;
comments and template instantiations were reviewed for parity. Gofmt, whitespace, and UTF-8/LF
checks passed. The one authorized seven-package test invocation passed five packages;
`identity-idp` and `skeleton-template` failed at test-binary linking with `R:` disk exhaustion.
The final strengthened assertions were added after that run, so complete executable validation is
not established. No further test run or staging was authorized while `R:` remains full. Current
worktree review found 376 changed paths (305 staged, 66 unstaged, 24 untracked; some paths overlap).
These include unrelated inherited staged work; no commit is safe without an owner-serialized
normal-hook slot and explicit path ownership review.

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
      framework startup/liveness/readiness semantics in the
      [durable deployment contract](../../docs/kubernetes-deployment-contract.md).
- [x] Generate and validate the exhaustive
      [k8s-deploy-tree.md](../../api/cryptosuite-registry/k8s/k8s-deploy-tree.md) and
      persistent [k8s-expected-inventory.json](../../api/cryptosuite-registry/k8s/k8s-expected-inventory.json) from the same
      planning model. Current evidence and exact coverage are recorded in
      [the Phase 1 evidence log](../../test-output/framework-v26-phase1-evidence.log).
- [x] Adopt the registry's three products as the explicit draft scope assumption; do not invent
      a fourth group. Adding one requires a separately defined identity/membership scope change.
- [x] Complete the main-agent Phase 1 postmortem and technical acceptance review. This does
      not by itself approve runtime work; the explicit user authorization is recorded separately below.
- [x] Close Phase 1 prototype lifetime/placement: the eight build-ignored Go files and durable
      inputs were initially moved outside the ephemeral tracking directory, with injected-I/O
      path tests. Phase 2 subsequently promoted the generator into
      [`internal/apps-tools/deploymentgen`](../../internal/apps-tools/deploymentgen/) and removed
      the obsolete prototype sources. The model/tree/expected JSON remain in
      [`api/cryptosuite-registry/k8s`](../../api/cryptosuite-registry/k8s/); the durable contract
      remains at [docs/kubernetes-deployment-contract.md](../../docs/kubernetes-deployment-contract.md).
      See the four-section postmortem in [lessons.md](lessons.md#phase-1-lifetimeplacement-postmortem).

## Phase 2: Replace Compose Service Deployment and Test Orchestration

- [x] Record explicit user authorization dated 2026-11-01: “Then implement phase 2, including
      post-mortem and fixes at the end. Ensure all work is committed.” Authorization is recorded.
- [ ] Promote the planning graph into production Go generation/check tooling with typed schemas,
      canonical chart templates, deterministic chart/values generation and independent Helm parity tests.
      **Progress**: production package/CLI, strict typed model inputs/schema, registry-derived graph,
      typed runtime and infrastructure values projections, and canonical runtime and infrastructure
      charts with strict schemas and focused Helm fixtures exist. Focused Go tests validate runtime
      projection/schema across all 55 releases and infrastructure projection/schema for the full
      inventory plus a SQLite-only selection. The CLI now previews, writes, and checks runtime
      values for one selected release from explicit non-secret input; it also previews/writes/checks
      complete infrastructure values. Focused mode, schema, output-path, and failure-path tests
      pass. Helm integration tests render all 55 runtime releases
      and compare chart-owned identities to the independent graph; infrastructure chart-owned
      identity parity is verified for full and SQLite-only inventories. Canonical and generated
      chart copies now pass production byte-exact `-check`; strict Helm lint and direct render pass
      against the generated runtime/infrastructure copies, including full and SQLite-only
      infrastructure fixtures. A generated lifecycle identity preview partitions externally
      managed Secret identities, PKI/DB Job identities, and bootstrap SA/Role/RoleBinding
      identities from the validated deployment graph without including Secret material. Focused
      Helm parity combines runtime and infrastructure objects, controller-generated PVC
      identities, and lifecycle identities for the `sm-kms` PostgreSQL and SQLite selections
      plus full infrastructure. This is selected identity parity only, not deployable lifecycle
      manifests or execution acceptance. A focused
      controller-ownership check
      expands StatefulSet claim templates into PVC identities and compares active runtime ordinals
      plus infrastructure claims to the independent graph, including exact namespace identity.
      Wrong-namespace rendered-object and controller-PVC cases fail as expected. The 55-release
      and infrastructure-selector parity loops are named parallel subtests; all split Helm
      integration test files are below the 500-line limit. An earlier focused snapshot reported
      deploymentgen statement coverage of 99.0% and passed Helm 4.3.1 strict lint for both
      canonical charts with the runtime chart's required schema fixture.
- [ ] Implement Helm charts and values for all services, tiers and modes, with configurable default
      min/max=2/2; no product/service list duplication or manual generated-file edits.
      **Progress**: runtime and generic PG/telemetry infrastructure chart templates now cover the
      partial resource surfaces and focused render/schema fixtures; neither chart is
      runtime-accepted. Runtime values now preserve per-pool args/config files, framework/domain
      ConfigMaps, credentials, per-ordinal TLS/health/OTLP Secret names, and an optional explicit
      API-server audience. Registry-approved within-release API edges are projected to scoped
      ingress/egress policies. Infrastructure values now project run identity, selected release
      clients, PostgreSQL enablement, and canonical component settings into the strict chart
      schema. Focused integration checks rendered the full 55-release runtime matrix and matched
      chart-owned resource identities to the independent graph. Infrastructure chart-owned
      identities match the graph for the full and SQLite-only selections. A min=1/max=3 Helm render
      verifies one active pod/PVC claim, all three reserved ordinal Secret names per role, and
      exact get-only RBAC. Controller-generated PVC names now match active graph ordinals across
      the runtime matrix and full/SQLite-only infra renders; these identity checks do not prove
      persisted volume behavior or runtime acceptance. Runtime values retain approved identity API edges, optional explicit
      API-server audiences, and backend-specific arguments/configuration. The typed infrastructure
      projection rejects collector replica counts that differ from canonical chart defaults and
      stale PostgreSQL pod identities when PostgreSQL is disabled.
      Runtime Helm negative renders reject missing/wrong-role ConfigMap references and duplicate
      ConfigMap names; these are cross-object template checks, not expected-inventory parity.
      Infrastructure Helm negative renders reject PostgreSQL without catalog entries and catalog
      entries without PostgreSQL.
      **Lifecycle prerequisite boundary (updated 2026-11-03)**: `lifecycle-identities` remains
      an identity-only preview, but the separate typed `lifecycle` contract now renders restricted
      PKI/DB Jobs and bootstrap RBAC and imports actual fixture Secrets for full-matrix parity.
      Protected physical import and lifecycle-aware admission are implemented with focused checks
      above; final continuation gates and execution-worker handoff remain pending. Actual
      Kubernetes provisioning, command execution and runtime acceptance are not established.
      The canonical chart files are copied by the production `charts` CLI mode
      into `deployments/k8s/charts`; byte-exact `-check` rejects missing, stale, extra, or
      non-regular mirrored files. The production CLI regenerated the stale runtime schema mirror
      and its subsequent `-check` passed; strict Helm lint/render against generated chart copies
      also pass. The PostgreSQL database catalog is now projected from independent inventory
      into infrastructure values and rendered as a ConfigMap; Helm integration decodes and
      compares its full database list, and confirms SQLite-only selections omit it. StatefulSet PVCs are controller-created and their
      names now match independent active-ordinal identities, but durable claim retention/runtime
      behavior is untested. Combined identity parity for the selected lifecycle-aware slice now
      passes; full 55-release combined lifecycle identity coverage now passes, while real runtime
      execution using operator values remains unaccepted. The planning CLI's
      tree/expected-inventory artifacts were refreshed by `go run ./cmd/cicd-deployment -write`;
      an earlier standalone Linux binary `-check` passed, and repeated generation preserved both
      SHA-256 digests. Do not hand-edit generated inventory.
      **Historical validation snapshot (before the latest source changes; not current gate
      evidence)**: an earlier Linux run passed normal/tagged builds, the full shuffled suite,
      normal/tagged `golangci-lint` and `go vet`, `lint-go`, `lint-fitness`, `lint-text`,
      `lint-docs`, and all 52 deployment validators. That suite exposed the required Helm
      `Chart.yaml` filename exception and a test-literal lint issue, which were corrected before
      that historical rerun. `deploymentgen` coverage was then 99.0%. The latest root integration
      failure is recorded in the current status above; no later broad pass is established.
      Combined lifecycle-object parity, PKI/migration/ownership execution, and runtime acceptance
      remain outstanding.
- [ ] Extend shared framework startupz/livez/readyz, background dependency monitor, graceful
      draining, CLI/router/usage/constants, templates, test helpers and all seven services.
      **Progress**: shared CLI/router and usage expose startupz; the framework application waits
      for both listeners and initial DB/public/required-collector observations before opening the
      startup/readiness latches. Admin handlers use sanitized snapshots; liveness remains available
      during drain; the legacy listener drains public work before admin shutdown. Collector
      readiness verifies the configured HTTPS endpoint with TLS 1.3, CA verification, client
      certificate, no redirects/proxy, and HTTP 200. The shared implementation has shuffled
      package test coverage above. Seven service entrypoints and their canonical templates now
      document startupz and contain help-output regression assertions for the endpoint, usage, TLS
      flags, and output stream; the final strengthened assertions passed the serialized seven-
      package test command (35 tests). The full-repository normal/tagged builds and broad gates
      have not yet been rerun after the user-authorized disk cleanup.
      Chart exec fixtures, every service's production launch, measured handler/exec latency,
      lifecycle race/timeout acceptance, and cluster fault tests remain.
- [ ] Implement K3S-only Compose bootstrap, locked prerequisites/toolchain, loopback API/gateway,
      TLS-passthrough SNI routing and capacity/ownership preflight.
      **Progress**: a pinned toolchain input and deterministic K3s-node-only Compose template exist;
      the actual read-only `bootstrap-preflight` CLI checks stopped/existing containers, networks,
      volumes and Compose-project labels, one Docker host measurement, explicit reserves and ports.
      Admission now requires explicit lifecycle input and budgets all selected Job requests
      alongside maximum app/init and infrastructure requests; current quota/revision checks are
      revalidated. Focused host-that-only-fits-apps and invalid-contract boundaries reject rather
      than silently downsize or omit Jobs. These injected checks are not live cluster acceptance.
      Locked Helm v4.3.1 was installed under ignored `test-output/` storage after digest
      verification; no PATH/environment setting changed. K3D was not installed because no version
      is pinned and the selected local target runs pinned K3s directly through Compose. Executable
      startup, gateway chart, certified live measured-host admission, scheduler/runtime capacity admission, and
      ownership-guarded cleanup remain.
- [ ] Implement restricted PKI provisioning, per-pod init materialization, migration Jobs and
      post-migration logical replication; prevent dependency cycles.
      **Progress**: runtime chart now declares exact max-ordinal server/health-client/OTLP-client
      Secret names for init-only `get`, per-ordinal materialization mapping, read-only framework/
      domain ConfigMaps, and explicit credential Secret mounts. Its chart fixture verifies
      min=1/max=3 exposes only one active replica and no reserved active PVCs. The generated
      typed runtime projection now supplies the required names and per-pool configuration.
      The typed lifecycle contract now renders restricted PKI/database/migration Jobs and exact
      bootstrap RBAC, while protected physical import reads operator-provided material in memory.
      The orthogonal execution transport/DAG library has reported focused injected tests, but its
      final handoff and outer CLI integration are pending; no real suite command is certified.
      Materializer implementation, actual PKI publication, runnable migration-only helpers, framework
      verify-only startup coordination, and follower reconciliation remain outstanding.
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
      **Cutover inventory (do not retire before runtime acceptance)**:
      `deployments/{7 PS-IDs,3 products,cryptoutil}/compose.yml` form the current service-to-product-to-suite
      include graph, with shared PostgreSQL/telemetry Compose under `deployments/shared-*`.
      Tagged framework TLS orchestration (`internal/apps-framework/service/test_orch_e2e`)
      selects PS-ID Compose files and test overrides; the shared E2E factory
      (`internal/apps-framework/service/testing/e2e_infra`) starts/stops configured Compose
      stacks. `ci-e2e.yml` builds/starts `deployments/sm/compose.yml`; `ci-load.yml` builds,
      starts, verifies, and tears down `deployments/sm-kms/compose.yml`. The DAST workflow
      currently includes a `deployments/cryptoutil` Compose `down -v` cleanup that must be
      reconciled during workflow cutover.
      **Enforcement owners for post-acceptance reconciliation**:
      `lint_deployments/lint_required_contents_deployments.go` requires Compose artifacts across
      service/product/suite/infra/template levels; `lint_deployments/validate_deployments.go`
      enforces product/suite include delegation; template-drift and Compose content/security
      checks currently validate the source set. Retain the old files as explicit read-only
      migration sources until rollback is verified, then replace active-required checks with
      K8s generated-artifact/schema/graph validators—not blanket validator removal. Workflow
      edits belong to the later workflow owner; no workflow file is changed by this runtime
      owner.

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
      source-backed reusable operational rules after actual runtime adoption; permanent handbook
      propagation is deferred to Phase 7 and is not a Phase 1 blocker.
- [ ] Record phase post-mortems, validation evidence, and the completion summary before archiving.
