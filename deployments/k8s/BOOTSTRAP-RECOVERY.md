# Bounded disposable bootstrap recovery

This applies only to the explicitly authorized `p2sm1103` / `s:sm-kms:mem`
development smoke. It is not production cluster certification or Phase2 acceptance.

The existing suite tool route is:

```text
go run ./cmd/cicd-deployment dev-smoke --repository-root ABSOLUTE_ROOT --image ACTUAL_DIGEST --kubectl PINNED_BINARY --helm PINNED_BINARY --timeout 20m --resume-bootstrap
```

Add `--preflight-only` for read-only admission, without material generation or
node launch. Do not delete retained state or rename the run to evade admission.

## Recovery boundary

- Explicit recovery requires the original owner-protected directory containing
  exactly canonical `compose.yml` and the original 64-character hexadecimal
  join-token file. Links, extra files, previous kubeconfig, incomplete files,
  changed topology, invalid token and unsafe handles are rejected.
- The original token and topology are never overwritten. The protected directory
  is reopened through `os.Root`; every file and directory handle is checked.
- Every expected local volume is inspected, including creation identity,
  mountpoint, driver, options and ownership labels. The complete snapshot and
  Docker engine identity must match again on each admission.
- Complete, unfiltered container/network/volume inventories reject existing
  compute, missing or duplicate volumes, foreign exact-name collisions and
  renamed resources belonging to this run/project. Fresh admission is unchanged;
  retained resources are not hidden from it.
- Capacity, reserves, replicas and port admission are unchanged. Recovery is not
  a simultaneous-matrix capacity exemption.
- After obtaining Kubernetes access, before image import, Secret publication or
  Helm mutation, a live inventory must contain only baseline system namespaces
  and named baseline K3s workloads, with no cryptoutil run-labelled objects.
  A kubeconfig or application-state collision requires a different explicitly
  owned recovery contract; this bootstrap-only route does not adopt applications.
- Failure observations verify the captured server identity first. Docker state
  and bounded startup stdout/stderr remain memory-only. Only fixed diagnostic
  classifications are emitted; private logs, health output and credentials are
  never printed or persisted. Observed markers are not guessed root causes.
- UID-guarded cleanup removes only captured owned compute/network. Data volumes
  and original protected inputs are retained.

## Coverage ceiling evidence

The native profile `test-output/coverage-analysis/dev-smoke-resume-current.out`
measured deploymentgen at 2,472 / 2,527 statements (97.8235%) before the final
projection-only test additions. This is **not** a strict 98% result.

Six statement-level caller guards have structural JSON barriers:

| Source | Barrier |
| --- | --- |
| `deploymentgen/cli.go`, lifecycle identity serialization | Closed inventory structs contain only JSON-encodable fields; canonical encoder seams are separately tested. |
| `deploymentgen/cli.go`, chart path serialization | A locally constructed `[]string` cannot contain unsupported JSON values or cycles. |
| `deploymentgen/lifecycle_deploy.go`, `LifecycleRevision` | The closed lifecycle input uses strings, integers, slices and string maps, not arbitrary interfaces or cyclic pointers. |
| `deploymentgen/lifecycle_cli.go`, projected runtime decode | The same concrete runtime type is decoded from freshly generated, validated canonical JSON, not caller-supplied bytes. |
| `deploymentgen/lifecycle_cli.go`, projected infrastructure decode | The same concrete infrastructure type is decoded from freshly generated, validated canonical JSON. |
| `deploymentgen/lifecycle_cli.go`, final lifecycle chart serialization | Its exhaustive dispatch binds only the three closed projection structs before canonical encoding. |

The conservative ceiling is `(2527 - 6) / 2527 = 99.76256%`; applying the coverage
skill's two-percentage-point buffer gives 97.76256%. The measured 97.8235% clears
that source-backed target, not the unadjusted 98% target. This analysis does not
exclude files or profiles, remove guards, or waive reachable error paths.
Filesystem, ownership, cancellation, entropy and native-process errors are not
classified as structurally unreachable.

## Evidence currency

Read-only native recovery admission passed against the original retained
protected inputs and all three immutable archive streams. A new current-source
Linux/amd64 suite image was built without Go/environment overrides:

```text
cryptoutil-phase2-suite@sha256:462f2c4b4ebc7e6c2d77dc8efa19950ee0d8b6e9839d4853c66105e4e367f650
```

The 1,122 source-only inputs hashed to
`62efbcb83c22f8352051e9b02e174984e4a09cf22f1a5af76a105dc81c43a534`;
the aggregate was unchanged before and after the build. The recipe and ignore
hashes remain the values recorded in the preparation report.

The first actual bootstrap attempt became unhealthy before kubeconfig creation.
Its exact owned compute/network was cleaned and three data volumes retained.
Recovery preflight and image construction do not establish live health,
authenticated API, admin mTLS, image-import interoperability, persistence,
rollback, workflow cutover or full-matrix acceptance.
