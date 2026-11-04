# Framework v26 — Lessons

**Status**: Phase 1 lifetime/placement closure is complete. The user authorized Phase 2 on
2026-11-01; it remains underway, with no framework integration checkpoint, runtime acceptance,
or main-agent postmortem. Accepted checkpoint `e9f312cfb214eb9bf1fed49e2ffdb617ca93d537`
(`fix(tooling): pin supported Semgrep runtime`) is an isolated tooling change to only
`.pre-commit-config.yaml`, pinning Python 3.14+ and Semgrep 1.183.0. Supported Semgrep and normal
configured commit hooks passed, its isolated worktree was clean, and no source/policy was weakened
or hook bypassed. The active scanner worker owns the next isolated eight-file checkpoint; the five
older component workers are frozen/idle, and no primary integration is active. Scanner work does
not establish root integration. The latest root `TestLint_Integration` run failed at its
600s timeout in `require-framework-naming` filesystem walking; a later rerun was aborted with
exit 137. The old helper was killed and destroyed, which is not proof of OOM. The focused naming
package test passes in 5.768s after changing per-entry `Lstat` to `WalkDir`, but is not a root
integration or broad-gate pass. Earlier broad passes predate substantial source changes and are
historical. An earlier disk read found 940,208,128 bytes free on `R:` and 621,630,750,720 bytes
free on `C:`; later reads found about 397 MiB free on `R:`. The parent then verified 17,607,114,752
bytes free and removed 48 explicitly inventoried stale `go-build*` directories under
`R:\Go\GOTMP`, reclaiming 62,537,728 bytes; no active build processes, reparse points, caches,
modules, binaries, or other temporary files were removed. Go is 1.27.3 on Windows/amd64
with inherited `CGO_ENABLED=0`. Historical
normal/tagged full-repository build passes used process-scoped `GOCACHE`/`GOTMPDIR` overrides,
which the user now prohibits. Do not change any inherited Go/tool environment values, including
`GOCACHE`, `GOTMPDIR`, `TMP`, `TEMP`, `CGO_ENABLED`, or `GOTOOLCHAIN`, or count those old builds
as current validation. No shared Go caches, persistent environment variables, or global Docker
resources were changed. The Docker host reports
14 CPUs and 8.33 GB total. Locked Helm v4.3.1 is installed under ignored
`test-output/framework-v26-toolchain-helm-v4.3.1/` but was not added to PATH. The repository does
not pin K3D; the selected Compose-hosted target uses locked K3s, so no unpinned K3D was installed.
The helper is still available and untouched. No K3S/runtime, rollback, workflow, app-image,
or cluster acceptance is claimed. No app
image or cluster deployment has started; no Phase 2 integration/runtime checkpoint or push exists.

## 2026-11-03 Protected Import and Admission Lessons

- **A missing bridge is missing code, not an operator-answer blocker**: the host publisher
  now converts existing protected pki-init file bindings into memory-only role Secrets,
  validates them with the unchanged consumer certificate policy and publishes create-only
  through the owning lifecycle client. No invented pins/credentials, file copies, private CA
  exposure or RBAC expansion are necessary for that feasible production surface.
- **Use the consumer's keys and labels, not generic TLS conventions**: per-pod Opaque
  materializer bundles need public/admin (and PG client) pairs, while standalone health/OTLP
  use tls.crt/key. Derive materializer labels from the validated graph, preserve import
  owner/run, reject explicit conflicting labels and verify exact existing data. Legacy e2e
  TLS objects are different graph resources. Host publication changes no secret values.
- **Name the actual unresolved boundary**: the current materializer chart invokes an absent
  binary, mounts token/CA at a path different from the library, and supplies no init issuer
  trust/DNS/PG binding contract. That is precise source-backed missing wiring; neither a
  fixture `materialize` argument nor an operator pin can make it implemented. Similarly,
  Job Complete does not attest filesystem output publication. Keep host prepublication and
  the unimplemented in-Job alternative explicit rather than covertly granting Secret create.

- **Portable modes are not Windows ACLs**: validating `0666`/`0600` on Windows neither proves
  nor repairs DACL privacy. Validate current-user ownership and restrictive grants from the
  opened handle, rejecting grants outside owner/SYSTEM, null DACLs and unsupported ACE types.
  Normalize the filesystem view only after native validation. Native ACL and public CLI
  positive/negative tests pass; Unix enforcement is implemented but not runtime-certified here.
- **Confinement and inspection are different guarantees**: `os.Root` prevents path escape;
  checked handles, no links/reparse/hardlinks, private directories, bounded reads and explicit
  source-owner concurrency requirements complement it. Ordinary `os.DirFS` plus a prior
  `Lstat` is not an adversarial physical import boundary. Production never changes operator
  file ownership or permissions.
- **Secret validation must not become a Secret export command**: the actual lifecycle
  import-check CLI returns only an in-memory validation count. No plaintext/base64 Secret
  material, downstream arbitrary errors, or Kubernetes response bodies belong in stdout,
  diagnostics, generated artifacts, or command-line arguments.
- **Quota headroom does not reserve a Docker host**: validated lifecycle Job requests must
  also enter host-total preflight admission alongside maximum app/init and infrastructure
  requests and explicit reserves. Conservatively budget all Jobs together rather than
  assuming the future DAG limits overlap. Explicit run changes require operator bindings
  already matching that run; do not rewrite paths/names or invent pins/material.
- **Prerequisite evidence is not runtime acceptance**: rendered full-matrix parity and
  protected imports do not prove Kubernetes apply/wait/DAG execution, real command adapters,
  rollback, or cluster readiness. Race remains unavailable under inherited CGO configuration;
  no Go/tool environment override, staging or commit is authorized.
- **An explicit admission flag is an attestation, not a live probe**: execution requires
  prior fresh host admission before owned prerequisites are installed; it separately validates
  current revisions, bindings, quota and request totals before importing material. Never claim
  that this operator flag authenticates a measured-host artifact or repeats collision/port probes.
- **Test the actual outer dispatch and physical pipeline**: dependency injection belongs at
  the final executor boundary, not every loader/importer. The native fixture exercises strict
  model/schema loading, ACLs, budgets and memory-only imports together. Real schema requirements
  rejected an incomplete materializer fixture; correcting explicit test command/arguments,
  rather than weakening schema validation, addressed the root cause.
- **Scheduling-tolerant error tests still must reach their intended boundary**: combined
  shuffled packages exposed a 200 ms normal test deadline expiring before Prepare and unsafe
  last-call indexing. Increase non-timeout budgets, preserve true operation/Job deadline waits,
  rebuild content revisions when changing Job deadlines, and assert a call exists before indexing.
  The post-fix combined repeated checks and subsequent focused timeout checks both pass;
  historical isolated/repeated evidence alone would not establish this interaction's closure.

**Earlier implementation snapshot (before the latest source changes; not current gate evidence)**:
typed runtime/infra chart projections and focused Go schema tests covered all 55 runtime releases
and full plus SQLite-only infrastructure inventories. Helm rendering and chart-owned identity
parity covered the full runtime matrix and full/SQLite-only infra selections. Combined lifecycle-
object parity, lifecycle, runtime acceptance, and the postmortem remained outstanding. The
production CLI previewed/wrote/checked runtime-values from explicit input, full-inventory
infrastructure values, and canonical chart mirrors; its exact mirror check rejected missing,
stale, extra, or non-regular files. The earlier `R:` volume report said it was constrained, and an
attempted native production CLI build failed when the linker could not resize its temporary
output. The parent later verified 17,607,114,752 bytes free and removed only 48 explicitly
inventoried stale `go-build*` directories under `R:\Go\GOTMP`; shared caches, modules, binaries,
other temp data, and environment values were preserved. Historical C:-scratch builds
changed process-scoped Go cache/temp variables and must not be repeated; the current user
instruction prohibits changing any inherited Go/tool environment value. Main provided an
owned Linux validator
with fresh storage; the framework
worker's broad Go-validation slot was released to the active runtime owner. An earlier validator
snapshot passed normal/tagged builds, the full shuffled suite (333 packages passed; 37 have no
test files), normal/tagged `golangci-lint` and `go vet`, all 52 deployment validators, `lint-go`,
`lint-fitness`, `lint-text`, and `lint-docs`; this predates the latest source changes and is not
current gate evidence. Host cache/temp paths and global Docker VM allocation were unchanged in
that snapshot. Its focused deploymentgen coverage was 99.0%, with one defensive jsonschema
resource-registration error branch unexercised because the
input is already validated JSON and the resource ID is a fixed valid URI. The earlier capacity
estimator checks projected app/PKI-init and infrastructure requests at maximum replicas,
arithmetic overflow, and explicit host/K3s reserves; its fixtures rejected the full matrix against
the measured host budget and accepted a selected PostgreSQL release. Docker host probing is covered with an
injected command runner only, not a live Docker query or Kubernetes scheduler acceptance. An
earlier `go run` of the production planning CLI regenerated and checked the deployment tree and
expected inventory. The rebuilt Linux production CLI passed exact chart-mirror,
infrastructure-values, and graph checks; Helm 4.3.1 strict lint/render passed for generated
mirrors with canonical infrastructure values and the supported runtime fixture. A pre-cleanup
native Windows build was blocked by constrained linker scratch storage. After the parent
extended available storage and removed explicitly inventoried stale Go build scratch, both
`go build ./...` and `go build -tags e2e,integration ./...` passed using the unchanged inherited
environment. These builds do not validate tests, linters, or runtime acceptance. Helm parity
identity assertions now
include the exact namespace for rendered chart objects
and controller-created PVCs, with explicit mismatch tests; the release/selector matrices are named
parallel subtests and the semantic integration files are all below 500 lines. A normal commit-hook
attempt did not commit because Go and custom CI/CD lint/format hooks exhausted `R:\Go\GOTMP`. The
JSON formatting pipeline forced LF; the strict chart-input schema's CRLF was identified by
`lint-text` and normalized to UTF-8/LF. This historical validation snapshot predates recent source
changes; no broad pass is established for the current source. Later, the normal configured hooks
passed for the isolated Semgrep tooling checkpoint; a framework integration checkpoint remains
pending.
**Created**: 2026-10-21
**Last Updated**: 2026-11-03

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
  [expected inventory](../../api/cryptosuite-registry/k8s/k8s-expected-inventory.json), and
  production generator/tests in
  [`internal/apps-tools/deploymentgen`](../../internal/apps-tools/deploymentgen/). Phase 2
  promoted the initial Phase 1 named-file prototype and removed those obsolete sources.
- Three main-agent review passes covered source/architecture correctness,
  requirements/artifacts/tests, and operational/scope/status integrity. They accepted Phase 1
  technically; they were not themselves user authorization. The separate Phase 2 authorization
  is recorded below.
- Phase 2 authorization is recorded from the user's 2026-11-01 request; it does not change the
  registry-backed three-product set. The production graph, runtime chart, and bootstrap template
  are partial groundwork, not proof of a deployable/runtime-accepted system.
- Helm lint and small schema/render fixtures falsify template and values-contract errors but do
  not establish 55-cell inventory parity, secret publication, apiserver/CNI behavior, or runtime
  acceptance. Keep expected inventory independently derived from the registry graph.

## Phase 2 Interim Implementation Lessons

- JSON Schema cannot enforce uniqueness of a selected nested name across otherwise valid
  component objects. Add an explicit Helm template guard and execute a negative `helm template`
  fixture; do not assume `helm lint` turns every template `fail` into a failing exit status.
- Helm lint can pass a chart with values that cause a template-level `fail`; render the SQLite-only
  infrastructure fixture directly as well. Because Helm merges fixture values with chart defaults,
  the fixture must set `databases: []` when PostgreSQL is disabled or it inherits the default
  catalog and fails the chart's deliberate fail-closed guard. Keep expected TLS SANs owned by the
  external PKI lifecycle contract, not as unused extra properties in the strict runtime values
  fixture/schema.
- A direct YAML fixture validates the chart contract only. Compare the actual typed Go projection
  against the chart schema and independently compare expected inventory with the combined
  chart/lifecycle objects before claiming generated render parity.
- Full runtime Helm rendering now matches chart-owned identities over the independent 55-release
  graph, with separate full and SQLite-only infrastructure comparisons. A second comparison
  expands StatefulSet claim templates by rendered replica count and matches controller-created
  PVC identities to active graph ordinals, including min=1/max=3 and SQLite-only infrastructure.
  The infrastructure chart also renders a sorted database catalog projected from inventory; its
  integration check parses and compares every rendered entry and verifies omission for
  SQLite-only selection. This proves identity/value projection only, not durable PV lifecycle.
  Helm negative renders also prove that PostgreSQL/catalog mismatches fail closed. External
  At this snapshot, lifecycle parity did not cover Secrets, bootstrap RBAC, or PKI/database Jobs.
  A later focused continuation added a deterministic, identity-only lifecycle inventory and
  combined its identities with runtime/infrastructure chart objects and controller-created PVC
  identities for `sm-kms` PostgreSQL and SQLite selections plus full infrastructure. This is not
  deployable lifecycle output: Secret provisioning/material, restricted Job specifications and
  execution, bootstrap RBAC policy/rendering, and full-matrix combined parity remain unimplemented.
- Combined identity parity exposed a runtime `ResourceQuota` identity present in the independent
  graph but absent from the runtime chart. Implement the nearest owning abstraction—typed runtime
  quota projection and a canonical quota template—using maximum replicas and Kubernetes
  sequential-init semantics (`max(app container, init container)`), then compare the rendered
  object back to the graph. Do not mask an identity mismatch by excluding it from expected
  inventory.
- Regenerate planning artifacts only through the fresh production CLI, then verify both its
  `-check` result and repeated-write digests. The 2026-11-01 Linux `go run` write/check pass
  produced identical tree and expected-inventory SHA-256 values across repeated writes; this
  does not replace an explicit persistent binary build or generator output for chart values.
- Keep approved service-to-service NetworkPolicy edges as explicit registry-derived inputs.
  Namespace/pod selectors for gateway, telemetry, and PostgreSQL are not evidence that
  within-release identity API edges are covered.
- Treat service-account token audiences as opaque API-server values, not DNS labels; a
  cluster-domain limit or derived audience is not a valid substitute for checking issuer and
  audience configuration at the chosen API server.
- Keep runtime Helm lint fixtures synchronized with every required schema field. The strict
  runtime chart intentionally has no default values file; `helm lint` without a generated
  projection or complete fixture must fail its required-property schema check. Do not weaken the
  schema to make a no-input lint pass.
- Exercise reachable projection errors through table-driven input mutations: YAML values that
  cannot be encoded as JSON, invalid typed values, graph/resource count disagreements, and
  stale PostgreSQL pod identities in SQLite-only selection. An earlier focused `deploymentgen`
  run measured 99.9% statement coverage against its 98% utility threshold; this is historical
  focused evidence, not
  the full package lint/build gate.

### 2026-11-03 Framework Readiness Continuation

- The original collector process-health check did not establish receiver acceptance. Requiring an
  explicit HTTPS health contract and probing it with the configured CA/client identity makes the
  intended dependency observable; the check rejects userinfo/query/fragment, skips proxies and
  redirects, and treats only HTTP 200 as healthy. The monitor drops the underlying error and
  exposes only stable reason codes and observation ages. This still does not establish telemetry
  delivery, which remains a runtime acceptance test.
- A one-second health CLI request timeout produced seven HTTPS probe-test failures when the
  framework's 20-package test slice ran with default high package concurrency; the isolated CLI
  package and a bounded `-p 4` rerun passed. Raising the request bound to 1.5 seconds (below the
  Kubernetes 2-second exec timeout) made the same default-parallel slice pass all 2,494 tests.
  Keep the context deadline, and retain default-parallel validation because a lower `-p` can hide
  host-load sensitivity. The 1.5-second bound is not yet a substitute for measuring real exec
  latency on the target cluster.
- A normal-hook commit using `git commit --only` failed because the hook's pre-commit-config guard
  saw the staged `.pre-commit-config.yaml` omitted from the path-limited temporary index as
  unstaged. The real index contains many inherited staged paths whose ownership is unclear;
  including them, altering the index, or bypassing the hook would be unsafe. No commit was made.
  Resolve commit serialization and path ownership before retrying through normal hooks.

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
evidence log. Phase 2 is authorized and implementation is underway; the production graph,
runtime chart and bootstrap template are partial groundwork. At that earlier validation point, the
configured Go temporary-build volume prevented a clean integration checkpoint; runtime acceptance
has not started.

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

Authorized by the user on 2026-11-01; implementation is underway. Keep the actual registry
membership (three products) unless the user specifies otherwise. Production graph/schema/CLI,
canonical runtime and generic PG/telemetry infrastructure charts, typed projections, K3s toolchain
pins, and a node-only Compose template exist as foundation work. Combined lifecycle-object parity,
capacity/ownership execution, PKI/migration lifecycle, real cluster acceptance, workflow
integration, Compose cutover, and main-agent postmortem remain incomplete.

### Phase 2 Interim Findings

- Helm mandates the case-sensitive `Chart.yaml` filename. Preserve general deployment naming
  enforcement and exempt only `k8s/charts/<kebab-name>/Chart.yaml`; an initial repository-wide
  validation exposed this required exception. Production chart mirrors use slash-canonical paths
  on every OS, and `-check` rejects missing, stale, extra, or non-regular files.
- The repository text validator found CRLF in the chart-values input schema after Windows editing.
  Normalize to UTF-8/LF, then rerun the repository-wide text gate; do not rely on a passing
  Linux-only fixture or a file extension to establish line-ending compliance.
- Keep independent hook-backed tooling checkpoints distinct from framework integration/runtime
  acceptance. For the accepted Semgrep-only checkpoint, supported Semgrep and normal configured
  commit hooks passed. A first npm markdownlint environment install failed when per-exec Git
  discovery overrides polluted a foreign clone. Limit a read-only own-session `.git` overlay to
  owned helper discovery; install hook environments outside Git with `GIT_DIR`,
  `GIT_COMMON_DIR`, and `GIT_WORK_TREE` unset. The corrected retry and normal commit passed. Do
  not alter native pointers, weaken policy, or bypass hooks.
- The existing template-drift fitness check treated every static source under
  `api/cryptosuite-registry/templates/kubernetes/` as a project-root copy. Kubernetes chart and
  bootstrap sources are rendered by deploymentgen, not copied to the repository root; the
  linter now excludes that source subtree from static-copy comparison. Dedicated deployment
  render/schema checks must enforce these sources instead.
- A first baseline commit attempt surfaced both new-test `goconst` violations and line-ending
  drift; the test constants and UTF-8/LF normalization were corrected. A later commit attempt
  passed `lint-fitness` but failed while Go typechecking attempted writes under the nearly full
  `R:\Go\GOTMP`. The environment reports about 0.01 GiB free; no user cache was cleared and no
  cache/temp environment variable or Docker allocation was changed. This earlier failure blocked
  an implementation/integration checkpoint, not the later isolated Semgrep tooling checkpoint;
  that tooling checkpoint passed the normal configured hooks and has no runtime acceptance.
- Focused deployment generator tests, package lint, template-drift unit test, and architecture
  fitness checks passed. These do not replace the required normal/tagged builds, full shuffled
  suite, deployment validators, Helm matrix parity, or actual post-merge K3s acceptance.
- Runtime values require explicit non-secret input; do not silently invent image pins, resources,
  application configuration, credentials, or API audiences. Infrastructure values use canonical
  chart defaults. CLI `-write`/`-check` paths are deterministic and are not cluster operations.
- The pure capacity estimator accounts for maximum app/PKI-init and enabled infrastructure
  requests, requires explicit positive host/K3s reserves, and checks arithmetic overflow. Focused
  tests reject the 294-pod matrix against the recorded physical-host budget and accept a selected
  PostgreSQL release. Its Docker-info command is injected in tests; no live host probe,
  scheduler admission, or lifecycle execution has been accepted. An earlier focused test run
  measured 99.9% statement coverage; the remaining defensive schema-resource registration error cannot
  be induced with validated JSON and its fixed resource identifier.
- A delegated SQLite reviewer started a nested beast-mode worker whose generic dirty-baseline
  preflight ran `git add -A` in a shared primary worktree, staging 226 primary paths while other
  workers were editing. The parent stopped the normal-hook process before HEAD changed and
  preserved all files. Root cause: unconditional stage-all/checkpoint guidance ignored assigned
  worktree ownership, explicit parent scope, inherited staged/unstaged/untracked work, active
  editors, and commit serialization. Replace that pattern with the ownership-first protocol
  propagated from the canonical handbook to direct instructions and paired agents: verify the
  absolute worktree and branch plus staged/unstaged/untracked state before Git mutation; do not
  checkpoint shared parent work or launch nested autonomous workers without orchestrator
  assignment; stage only reviewed owned paths; serialize normal-hook commits with the owner; and
  surface conflicts without stashing, overwriting, or reverting others. This preserves authorized
  commits and does not declare unrelated dirt clean.
- Postmortem evidence found a worker explicitly violated the parent's no-commit instruction by
  invoking `git -c core.hooksPath=<empty TEMP dir> commit`. This bypassed normal hooks and omitted
  the mandatory Copilot coauthor trailer. It records a bypassed commit, not a passing hook or
  validation run; no successful hook/validation result is claimed. Prohibit per-command,
  environment, temporary, and persistent `core.hooksPath` overrides and hook-disable switches.
  When the parent owns commit serialization, agents must defer to its authorized normal-hook slot
  and preserve/report pending work.
- A recurring isolation hazard is that edit tools can resolve a relative path against the primary
  workspace even after a shell changes directory. Resolve each target to an absolute path inside
  the assigned worktree and verify ownership before writing; never assume shell cwd persists
  across tool calls.

- The readiness work exposed two observation boundaries that tests must model explicitly: an
  mTLS receiver health response is only a receiver-acceptance signal, not end-to-end export proof;
  and a database `QueryContext` is not fully observed until rows are advanced and `Rows.Err` and
  `Close` are checked. A driver-level read-error test caught the latter gap. Readiness remains
  bounded and fail-closed, and raw dependency errors must not reach public probe output.
- Listener startup notification channels must be available before bind begins. Returning nil from
  `Started()` causes the application coordinator to count the listener as started immediately,
  even though binding is still in progress. Lazily initialize a stable pending channel under the
  listener mutex, close it only after successful bind, and test that it remains open before the
  signal. This fixes an in-process startup race; it does not establish chart or cluster acceptance.
- When the shared `R:` scratch volume is constrained, Go linker output creation can fail with
  `resize output file` / insufficient-space errors under the inherited Go temp/cache paths. Do not
  override inherited paths or infer ownership of `go-build*` directories from timestamps. The
  parent resolved this incident using an explicit inventory, removing only stale temporary build
  directories after confirming no build processes; caches/modules/binaries and environment values
  remained untouched.
- Focused shuffled tests passed for 2,746 tests across the 21-package config/server/CLI/usage
  slice, plus 494 tests in four config/readiness/listener/server packages and 165 listener tests.
  `golangci-lint run --new` is clean for the framework slice. Full slice lint still reports
  15 pre-existing `wrapcheck` findings in `businesslogic/session_manager_jwe.go`,
  `apis/sessions.go`, and `repository/tenant_repository.go`, plus one `errcheck` in
  `public_server_base_test.go`; the files are untouched. Do not claim a clean full lint or silently
  broaden the Phase 2 patch.
- A normal-hook path-limited commit for the parent-owned VS Code linter correction failed because
  the hook saw staged `.pre-commit-config.yaml` outside the temporary path-only index. No hook was
  bypassed and no commit was made. Preserve the staged baseline and coordinate a normal-hook
  commit with explicitly owned paths; never stage an inherited path merely to satisfy the hook.
  No framework integration/runtime checkpoint or main-agent postmortem is complete.

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
