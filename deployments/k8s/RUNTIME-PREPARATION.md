# Phase2 disposable runtime preparation and ownership

## Retained prelaunch result

This is preparation evidence, **not Phase2 acceptance, admission certification,
or permission to adopt shared resources**. No application image execution,
K3s startup, cluster mutation, staging, commits, hooks, or shared cleanup occurred.
The user-authorized disposable development smoke is `s:sm-kms:mem`.
The original checked-in model and old deployment material remain unchanged.

The supported read-only executable route is:

```powershell
go run ./cmd/cicd-deployment dev-runtime-plan `
    -run p2sm1103 -select s:sm-kms:mem `
    -image cryptoutil-phase2-suite:p2sm1103 `
    -reserve-cpu-milli 2000 -reserve-memory-bytes 2147483648
```

Current real-image preflight passed with lifecycle revision
`r80b7fa992f6be951`, one release, 19 protected Secret requirements, two PKI
verification Jobs, and the existing minimum/maximum two application replicas.
Requests are 750m CPU and 939,524,096 bytes; reserves are another 2,000m CPU
and 2,147,483,648 bytes. Docker reports 14 CPUs and 8,326,418,432 bytes.
Existing production checks inspected containers, networks, volumes, cached
infrastructure images and run identity, and bound/closed both loopback ports.
No existing target identity was adopted.

This is a point-in-time check, not a resource/port lease or scheduler proof.
Rerun immediately before any authorized launch.

The simultaneous complete matrix was also checked using literal `'-select='`:
it rejects 36,400m CPU / 49,190,797,312 bytes against the post-reserve
12,000m CPU / 6,178,934,784 bytes available. A target needs at least those
requests **plus** reserves. Serial selectors do not establish simultaneous
matrix capacity; do not silently reduce replicas/requests or alter Docker settings.

## Exact proposed ownership manifest

All identities below are proposed; none of these cluster resources was created.

| Surface | Owned identity |
|---|---|
| Run / Compose project | `p2sm1103` / `cu-p2sm1103-cluster` |
| Docker network | `cu-p2sm1103-nodes` |
| Server | `cu-p2sm1103-cluster-server-0` |
| Agents | `cu-p2sm1103-cluster-agent-0`, `cu-p2sm1103-cluster-agent-1` |
| Server volume | `cu-p2sm1103-cluster-server-0-data` |
| Agent volumes | `cu-p2sm1103-cluster-agent-0-data`, `cu-p2sm1103-cluster-agent-1-data` |
| Host endpoints | `127.0.0.1:6445` API, `127.0.0.1:8443` gateway |
| API service / service CIDR | `10.43.0.1` / `10.43.0.0/16` |
| Pod CIDR / DNS domain | `10.42.0.0/16` / `cluster.local` |
| Smoke namespaces | `cu-p2sm1103-infra`, `cu-p2sm1103-s-sm-kms-mem` |
| Retained StorageClass identity | `cu-p2sm1103-retain` |
| Proposed private output root | `C:\Dev\cryptoutil\test-output\phase2-runtime-p2sm1103` |
| Join-token path relative to generated Compose | `./secrets/p2sm1103/k3s-join-token` |

The canonical node-only Compose template assigns `cu.io/run` and exact
`cu.io/node-id` to nodes/volumes, plus node role or network role as applicable.
Compose project identity must also match. Nodes are privileged by that existing
explicit bootstrap design; this does not relax restricted application/Job policy.
Protected output creation must reject an existing/unowned root and enforce real
owner-only permissions, including Windows ACLs; a filename or chmod alone is not proof.
Any eventual cleanup must recheck exact names, identities and labels and preserve
retained Kubernetes application PVCs. Unrelated old networks/volumes are not ours.

## Actual build provenance

The final LF-normalized source-only recipe was built successfully with native
Linux image defaults, `--platform linux/amd64`, and no Go/tool/environment
overrides. The adjacent deny-by-default Docker ignore excludes configs,
deployment secrets, tests/testdata, worktrees and shared state.

Actual inspected suite reference:

```text
cryptoutil-phase2-suite@sha256:0ba3c947e16f0749807b07c77f43a9942c1af154044e5c01e057c537b3e61b23
```

Labels are `cu.io/owner=phase2-runtime`, `cu.io/run=p2sm1103`, and
OCI revision `491cfd7faea8fc4440f8180aab583d64d4d7ff4c`.
That revision is baseline HEAD, **not a clean-source attestation**:
the image intentionally includes current dirty production source, including
inherited work whose final review belongs to its owner.
The current allowed-source aggregate is
`735e9371d8083fb5be417ad0584e2604ce3d362fe0d895fe47c5e45acb509b4f`
over 1,100 allowed files, hashing each sorted relative path, NUL, and file SHA256.

Recipe SHA256:
`c926d95fdf97ee507d8b275de4899ea26081b3d62bcd4825c0a3f244296503de`.
Ignore SHA256:
`5c153d8e032aa8f809fc1c0ae676d78a145f461f59fac3d6fb5ac07aa790bd8a`.

The inspected build base is
`golang:1.27.3@sha256:2920bb124f1a2e12a1d39f0cab756ae0068441ee1da7655dec7f5e374944095b`.
The lock pins K3s `v1.37.1-k3s1`, Helm `v4.3.1`, kubectl `v1.36.1`,
and the actual cached PostgreSQL/collector/LGTM images.
Native Helm and kubectl paths/checksums were verified against that lock.

Forced-static glibc/SQLite was rejected after its dlopen runtime warning.
The final scratch runtime instead includes the native build's exact matching
libc and loader plus the public OS trust bundle and numeric nonroot suite
executable. `ldd` showed only that libc/loader dependency closure.
Earlier image digests and preflight revisions are superseded, not current acceptance.

## Implemented preparation versus remaining executable work

`BuildDevelopmentRuntimePlan` derives independent graph identities,
Secret **file references**, contract-content-bound lifecycle revision, exact supported
suite app/probe/drain/PKI/DB routes, namespace-wide certificate-only PKI mounts,
explicit materializer trust/identity inputs, and graph-bound PostgreSQL client
CN/role mappings. It does not read old secrets, generate material, write files,
launch resources, or certify the target.

Its minimal framework/domain configuration is a budget/schema projection,
**not a complete runnable service configuration**. Do not deploy it as such.
Database Secret projections also remain budgeting skeletons, not complete
administrator/target/subscriber transport bundles: all referenced DSN and
six-file connection mappings must be constructed before protected input
validation can pass. Expanding those inputs changes the lifecycle contract;
recompute its revision and repeat admission rather than reuse this preview.
Actual protected preparation and orchestration still require code:

1. Generate new disposable join token, credentials, pepper/unseal state and
   Kubernetes-aware issuer/leaf material with existing approved production APIs.
   Existing Compose certificates have no Kubernetes service SANs and cannot be
   relabeled. Legacy `pki-init` SAN topology is not a usable Kubernetes bridge.
   Disposable dev trust is authorized; external approval is not falsely required.
2. Complete per-service framework/domain configuration, exact protected Secret
   keysets, host publication and Job certificate-policy/hash manifests, and
   PostgreSQL administrator/target/subscriber six-file transport mappings.
   Provision new owned PostgreSQL roles before `db-lifecycle init`, which
   intentionally verifies preprovisioned roles rather than granting arbitrary ones.
3. Safely serialize an explicit run-owned model/runtime/lifecycle overlay and
   protected Compose token input. The existing execution loader's default model
   must not accidentally execute checked-in review/r1 instead of this overlay.
4. After a fresh preflight and retained prelaunch report, start only these nodes,
   privately capture owned kubeconfig/context, import actual pinned images into
   their containerd stores, and provision namespaces/quotas/policies/storage.
   No kubeconfig or cluster is available yet; K3s can generate these for the
   disposable target without an invented operator credential.
5. Host-prepublish exact valid PKI bytes, then require verify-only PKI Jobs
   **and live read-only postJob attestation** before database barriers and runtime
   startup. Preserve no-token/no-write PKI Jobs and restricted bootstrap RBAC;
   regeneration with unpublished ephemeral output remains unsupported.

These are actionable implementation gaps, not missing user answers.
The hardware mismatch for the simultaneous matrix is an actual external
capacity boundary. Existing-instance migration additionally needs an explicit
owned source-to-target/data map; disposable empty smoke data cannot prove it.

## Remaining acceptance and cutover

Runtime acceptance still includes TLS/functionality/probes/latency, scheduler
capacity, DB creation/migrations/replication, failure/cancellation barriers,
gateway routing, and retained-PVC replacement/upgrade/uninstall/reclaim behavior.
Charts already declare retention; live persistence is not proven.

Keep service/product/suite Compose and old PostgreSQL/telemetry state unchanged
until source mapping, quiescence, backup/restore, traffic switch and rollback
are demonstrated. `ci-e2e.yml`, `ci-load.yml` and `ci-dast.yml` still use Compose
or broad cleanup and require a separately owned K3s/Helm workflow cutover.
Do not remove their validators or perform broad cleanup. No workflow files were edited.

The final Phase2 postmortem must distinguish local validation from live
acceptance and record static-build rejection/native dependency correction,
legacy SAN incompatibility, full-matrix capacity rejection and preserved state.
Parent framework tracking documents remain untouched by this slice.

## Validation and parent commit boundary

Current full normal/tagged Go builds passed on native Go1.27.3 Windows/amd64.
Owning normal tests passed shuffled count3; tagged owning tests passed.
Coverage is graph 2,140/2,183 = 98.030233623%, CLI 276/277 = 99.638989170%.
Normal/tagged scoped lint and vet passed. Selected actual pinned native Helm
checks passed for lifecycle graph parity, PKI verification, materializer
negative guards, runtime full-matrix rendering and retained PostgreSQL paths.
All 52 deployment validators and `-format charts -check` passed.
These do not certify a running cluster or perform full runtime E2E.

Parent-review allowlist for **this preparation slice only**:

- New `deployments/k8s/Dockerfile`, `Dockerfile.dockerignore`, this report.
- New `internal/apps-tools/deploymentgen/dev_runtime_plan.go`,
  `dev_runtime_cli.go`, `dev_runtime_plan_test.go`, `dev_runtime_cli_test.go`,
  `dev_runtime_mutation_test.go`, `dev_runtime_pki_test.go`.
- Only the `net` import and `dev-runtime-plan` dispatch added to existing `cli.go`.
- Only shared username/follower-bootstrap literal-to-constant references in
  preexisting `lifecycle_deploy.go`; do not claim its whole prior contents.

Build recipe and development planning are separate coherent commit groups.
No commit authorization is implied. Parent must review/serialize normal hooks
and explicitly owned hunks after all scopes freeze. Inherited 305 staged paths
are untouched; their cached binary diff SHA256 remains
`964ffa22b8fd011f84f50349cb62df3fc9b375c94d53d36050a382b55af496cb`.
Other dirty/untracked work is preserved and must not be staged or cleaned here.
