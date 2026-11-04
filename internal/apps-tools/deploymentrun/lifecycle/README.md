# Bounded lifecycle execution adapter

The executable library uses the existing `deploymentgen` contract and canonical
`cryptoutil-lifecycle` chart. It does not own generator, chart, schema, admission
or CLI changes. It does not create clusters or perform destructive cleanup.

## Production CLI and library boundary

Dispatch is implemented in `internal/apps-tools/cicd_deployment`,
**not** `deploymentgen.Main`: importing the executor there creates a Go cycle.
Keep revision/run/namespace bindings unchanged and validate the coupled current
admission contract before execution:

1. Decode the model and non-secret lifecycle input; build the selected inventory.
2. Run lifecycle-aware preflight with explicit `*LifecycleInput`; require
   revision/namespace agreement and lifecycle-aware runtime/infrastructure quotas.
3. Import protected material using
   `deploymentgen.ImportPhysicalLifecycleSecrets(model, inventory, input, root)`.
   Never serialize or print the imported result.
4. Require `-publication-input` and prepare the complete protected PKI
   verification plan/runtime policy. Reject generation-only Jobs before transport.
5. Construct a new client for this invocation and execute:

```go
client, err := lifecycle.NewKubectlClient(lifecycle.KubectlConfig{
    Kubectl:        absoluteKubectl,
    Helm:           absoluteHelm,
    RepositoryRoot: absoluteTrustedRepository,
    Kubeconfig:     absoluteProtectedKubeconfig,
    Context:        explicitOwnedContext,
    PollInterval:   explicitPollInterval,
}, lifecycle.ExecRunner{})
// Handle err without printing any imported material.
result, err := lifecycle.Execute(ctx, lifecycle.Request{
    Model: model, Inventory: inventory, Input: input, Secrets: imported,
    VerifyPKI: preparedReadOnlyPKIVerifier,
}, client, lifecycle.Policy{
    MaxDuration: explicitOverallDeadline,
    OperationTimeout: explicitOperationDeadline,
})
```

`ctx` must carry parent cancellation (including interrupt handling). All paths,
context and durations are explicit: no cluster/context discovery or defaults.
The constructor is local only. The client renders the canonical chart through
Helm stdin; Kubernetes payloads travel only through stdin and bounded memory.
The built-in runner discards stderr and does not log stdout, arguments or input.

## Ordering and errors

Execution first validates the complete contract, imported identity/type/key/data
bounds and the stage graph. It then requires owned Active namespaces and explicit
ConfigMaps, imports all Secrets, and creates/verifies bootstrap RBAC.

Completion barriers, with deterministic serial siblings:

```text
infra verify-only PKI + postJob attestation -> all release verify-only PKI + postJob attestation
          -> ready PG leader + follower
          -> union db-init -> all member migrations -> replication
```

SQLite-only selections omit every PG stage. No Job is submitted before its
dependencies complete. Job apply and completion wait share its operator-supplied
`activeDeadlineSeconds`; all operations are bounded by the overall deadline,
and non-Job operations additionally by `OperationTimeout`.

Errors identify operation/resource only; arbitrary client/process errors are
redacted. `errors.Is` supports cancellation/deadline and `ErrCollision`,
`ErrPrerequisite`, `ErrJobFailed`, `ErrReplaced`, `ErrTransport`.
`Result` counts only confirmed Secret imports and completed Jobs, including
partial progress on failure. Errors never trigger rollback.
PKI Jobs count only after a mandatory read-only postComplete callback verifies
live Job UID/revision/spec/security and mounted Secret UID/owner/type/full
keys/bytes, then current certificate and runtime policy. The production CLI
supplies that callback; library callers must prepare an equivalent whole-request
policy. Missing callbacks or generating PKI commands reject before Prepare.
Verification Jobs use `/app/cryptoutil pki-verify` with protected
`pki-verification.json`, no API token and no Secret writes. Generation output
attestation/in-Job publication is intentionally unsupported.

## Ownership and retry

This is deliberately conservative apply semantics: **verify-or-create**, not
patch/replace. Atomic create fails if another actor wins the existence race.
Existing run/release/revision labels, explicit manifest fields and complete
Secret data must agree. Jobs also reject added executable/security Pod fields;
known inert Kubernetes defaults are accepted. Failed Jobs are not deleted or
restarted. Completion waits guard UID replacement and fail on Job `Failed`;
PG readiness requires observed generation, updated/ready replicas and matching
controller revisions.

Use a new client for every invocation/retry. Inputs and clients are not shared
with concurrent executions. Changing imported material (including PKI-generated
TLS data) requires explicitly refreshed, matching operator material or a separate
owner-authorized reconciliation mechanism; this adapter never overwrites it.
Admission mutations or noncanonical resource quantities can fail verification
closed and must be certified by the runtime owner, not silently accepted.

## Exact remaining runtime prerequisites

- A real reachable Kubernetes context and protected kubeconfig, plus operator
  RBAC for named get/create and readiness reads. Cluster creation is external.
- Certified Helm/kubectl versions and trusted canonical chart/repository files.
- Owned Active namespaces (`cu.io/run`, `cu.io/release`), referenced ConfigMaps,
  Kubernetes-managed `kube-root-ca.crt`, bootstrap API/DNS connectivity and
  lifecycle-aware quotas/network policies already installed by the parent.
- For PG selections, owned leader/follower StatefulSets/services/storage and
  working database readiness. The adapter waits for them; it does not install,
  clear or manipulate their volumes.
- Real digest-pinned images containing the implemented suite verifier,
  materializer and database lifecycle commands, protected material and
  service-specific configuration; actual operations require runtime certification.
- Explicit protected publication and hash/certificate verification manifests.
  Host publication precedes Jobs; later imports must match exact published bytes.
  CLI/admission/import/execution wiring is implemented, but injected tests do
  not claim live-cluster acceptance.

Validation uses only injected clients/runners and bounded local subprocess
helpers. No cluster, Docker/testcontainers, environment override or global
cleanup is involved.
