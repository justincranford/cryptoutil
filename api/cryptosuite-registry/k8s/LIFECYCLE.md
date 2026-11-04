# Kubernetes lifecycle projection contract

This is the bounded Phase2 lifecycle specification/rendering prerequisite, not
cluster acceptance or proof that every lifecycle command is executable.

## Inputs and ownership

`cicd-deployment lifecycle -input operator.json` validates version-1 typed input
against the independently constructed deployment inventory. Every expected
Secret and Job must appear exactly once; extra or cross-namespace resources fail.
The file contains references and non-secret settings, never private key bytes,
passwords, or token contents.

Each `secrets` entry supplies `namespace`, `name`, `type`, and a `files` map from
Kubernetes data keys to protected relative file paths. Supported types are
`Opaque` and `kubernetes.io/tls`. TLS and client-certificate consumers require
`tls.crt`/`tls.key`; trust consumers require `ca.crt`; PostgreSQL bootstrap
consumers require `username`/`password`; LGTM administration requires `password`.
Other service-specific file names and formats must match the operator's explicit
runtime configuration.

An ordinal-specific Opaque `*-tls` Secret consumed by the materializer instead
contains exactly `public.crt`, `public.key`, `admin.crt`, `admin.key`, plus
`postgres.crt`/`postgres.key` for PostgreSQL pools. Standalone health/OTLP
Secrets contain exactly `tls.crt`/`tls.key`. The legacy `*-e2e-tls` bindings
are separate graph resources, not substitutes for ordinal materializer bundles.

Each `jobs` entry supplies the stable graph name without its revision suffix,
namespace, digest-pinned suite image, `/app/cryptoutil` command, explicit arguments,
CPU/memory requests and limits, deadline (1–3600 seconds), backoff (0–3), and
explicit Secret read/write and ConfigMap references. There are no invented image
pins, implicit commands, or resource/deadline defaults.

Executable PKI Jobs must invoke `pki-verify` with a protected verification
manifest, no Secret writes and no API token. Generation-only `pki-init` remains
renderable for preview compatibility but is rejected before lifecycle transport.
Database Jobs cannot invoke either PKI command, receive a
projected API token, write Secrets, or read signing/application credentials.
Migration Jobs may read only their own member's migration credential Secret.
PKI write permissions cannot overwrite authorities, unseal/pepper, or credentials.
All references must be owned by the expected graph.

## Revision and output

Calculate the required revision before deliberately updating the model:

```text
go run ./cmd/cicd-deployment lifecycle -input operator.json -format revision
```

The revision is a content-derived hash of the canonical non-secret operator
input, including pins, file bindings, commands, resources, and deadlines.
Array order is significant. The model's revision must match; the CLI never
silently changes it.

```text
go run ./cmd/cicd-deployment lifecycle -input operator.json
go run ./cmd/cicd-deployment lifecycle -input operator.json -write
go run ./cmd/cicd-deployment lifecycle -input operator.json -check
go run ./cmd/cicd-deployment lifecycle -input operator.json -format secret-plan
```

`-select tier:scope:mode` limits validation and output to that graph selection.
`secret-plan` previews file bindings only and prohibits write/check. The default
values output is schema-validated and renders the canonical
`cryptoutil-lifecycle` chart. Helm renders restricted Jobs and namespace-scoped
bootstrap ServiceAccounts/Roles/RoleBindings, never Secret material or hooks.

Runtime and infrastructure quotas must reserve Job headroom, not just maximum
application replicas. Use these lifecycle-aware projections instead of the
app-only projections before admitting or installing the combined graph:

```text
go run ./cmd/cicd-deployment lifecycle -input operator.json -format runtime-values -select tier:scope:mode -runtime-input runtime-input.json
go run ./cmd/cicd-deployment lifecycle -input operator.json -format infrastructure-values
```

Each projection replaces its existing quota with a conservative sum of all
same-namespace Jobs plus the original maximum-replica budget. Ownership/revision
must agree. Bootstrap database traffic is permitted only to the owned PostgreSQL
leader/follower on port 5432; SQLite-only graphs omit those policies.

## Protected material import

`ImportLifecycleSecrets` validates the complete contract, reads protected files,
and returns Kubernetes Secret objects in memory. It performs no Kubernetes
mutation, stdout output, or artifact persistence. Import all Secrets before
running Jobs: least-privilege RBAC grants only named `get` or named
`get/update/patch`, never unrestricted `create`, list/watch/delete, wildcards, or
cluster privileges.

The generic `fs.FS` API requires a stable trusted source. The production
`ImportPhysicalLifecycleSecrets` adapter requires an explicit clean absolute
directory and uses `os.Root` confinement and checked opened-handle identities.
It rejects symlink components, reparse points, hard-linked files, unowned or
unprotected directories/files, empty files, and file/aggregate Secret payloads
over 1 MiB. Reads are bounded and close errors propagate.

On Windows, the handle owner must be the current process user; the DACL must be
non-null and grant access only to that owner or SYSTEM. Grants to Everyone,
groups or Administrators, even inherited, are rejected. Unsupported ACE forms
fail closed. Only after native DACL validation does the filesystem view expose
owner-only modes to the generic importer. On Unix, the opened object must have
the current effective UID and owner-only modes without special bits. Production
never repairs permissions or changes ownership. The owner must prevent
concurrent source mutation; root confinement is not a source snapshot.

The actual CLI path reads material into memory but never exports it:

```text
go run ./cmd/cicd-deployment lifecycle -input operator.json -format import-check -secret-root ABSOLUTE_PROTECTED_DIRECTORY
```

`import-check` rejects `-write`/`-check` and emits only the validated Secret count.
Non-import formats reject `-secret-root`. Native Windows and public CLI
positive/negative tests validate this boundary; Unix runtime acceptance remains
unverified in the current Windows environment.

## Guarded execution CLI

The outer `cicd-deployment lifecycle-execute` command wires the protected importer
to the `deploymentrun/lifecycle` execution adapter. It revalidates the current
model revision, exact operator bindings, namespace quota headroom and conservative
host request totals before reading material. All target/material paths below
must be clean absolute paths; context and policy durations have no defaults.

```powershell
go run ./cmd/cicd-deployment lifecycle-execute `
    -input operator.json -runtime-input runtime-input.json -publication-input publication.json -select tier:scope:mode `
    -repository-root ABSOLUTE_CERTIFIED_ROOT -secret-root ABSOLUTE_PROTECTED_DIRECTORY `
    -kubectl ABSOLUTE_CERTIFIED_KUBECTL -helm ABSOLUTE_PINNED_HELM `
    -kubeconfig ABSOLUTE_OWNED_KUBECONFIG -context EXPLICIT_OWNED_CONTEXT `
    -admission-certified -timeout EXPLICIT_DURATION `
    -operation-timeout EXPLICIT_DURATION -poll-interval EXPLICIT_DURATION
```

`-admission-certified` is an explicit operator attestation that fresh read-only
host admission passed for this exact run and inputs **before** installing the
owned cluster prerequisites. It is not a live host measurement or a verified
attestation artifact. Current contract/quota/request validation does not reprobe
Docker, reserves or host ports. Do not set the flag without that prior admission;
existing cluster containers would intentionally collide with a fresh bootstrap.

Before execution, install owned Active namespaces, ConfigMaps, lifecycle-aware
ResourceQuotas and infrastructure StatefulSets using the same certified inputs.
The adapter imports Secrets into memory, renders the canonical lifecycle chart
through Helm stdin, and uses guarded verify-or-create operations without
overwriting, patching or deleting existing resources. It waits through infra PKI,
release PKI, PostgreSQL leader/follower readiness, union database initialization,
all migrations, and replication barriers. Existing objects must match ownership,
run, revision and exact material/spec; replaced UIDs and failed Jobs reject execution.
Job deadlines include creation and waiting; execution has bounded overall and
per-operation deadlines.

Private stdin and API bodies are not persisted or logged. The CLI emits only
result counts and safe failure categories, not arbitrary transport diagnostics.
Synthetic pins/material in tests never become execution defaults. A successful
command does not certify application client acceptance or rollback.

## Host PKI publication bridge

`pki-publish` is a production host command, not a Helm hook or an in-Job writer.
It reads existing protected `pki-init` output through the lifecycle file bindings,
validates every reserved runtime pod's role-separated certificates and private
keys with the **unchanged materializer consumer policy**, and precreates absent
Secrets through the lifecycle verify-or-create client. It generates no keys,
rewrites no bytes/permissions, and creates no temporary material files.

The existing suite generation command uses real flags, not positional fixture
arguments:

```text
cryptoutil pki-init --domain=OPERATOR_APPROVED_TIER --output-dir=OPERATOR_OWNED_OUTPUT
```

Map generated `{directory}/{directory}.crt` and `.key` files to the required
Secret keys in the explicit lifecycle input. Private CA keys and `.p12` files
must not be substituted for leaves. The operator must protect the existing
source before import; stock Unix generation modes are directory `0755`,
certificate `0444`, private key `0440`, whereas this importer requires
owner-only directories/files. Publication never repairs those modes or copies
them to another staging directory.

```powershell
go run ./cmd/cicd-deployment pki-publish `
    -input operator.json -runtime-input runtime-input.json -publication-input publication.json `
    -select tier:scope:mode -repository-root ABSOLUTE_CERTIFIED_ROOT `
    -secret-root ABSOLUTE_PROTECTED_DIRECTORY -kubectl ABSOLUTE_CERTIFIED_KUBECTL `
    -helm ABSOLUTE_PINNED_HELM -kubeconfig ABSOLUTE_OWNED_KUBECONFIG `
    -context EXPLICIT_OWNED_CONTEXT -admission-certified `
    -timeout EXPLICIT_DURATION -operation-timeout EXPLICIT_DURATION -poll-interval EXPLICIT_DURATION
```

`publication.json` contains a `pods` array with exactly every selected runtime
ordinal, including reserved maximum replicas. Each entry supplies `namespace`,
`podName`, `expectedAdminDNS`, `expectedPublicDNSNames`, and `trust`. Trust maps
`public-server`, `admin-server`, `health-client`, `otlp-client`, and (for PG)
`postgres-client` to `{ "secret": NAME, "key": KEY }` references in that same
namespace's protected imports. No certificate/key bytes belong in this file.
For PG, `postgresqlClientBindings` maps every configured pod to explicit
`CommonName` and `DatabaseRole` values, matching the existing consumer contract.
Missing/duplicate/unowned pods, unknown properties/trust roles, invalid references,
incomplete key sets, malformed/forged chains, key mismatches, wrong SAN/EKU/CN,
expired certificates, oversized payloads, and forged labels reject publication
before Kubernetes mutation.

The publisher derives `cu.io/workload`, `cu.io/psid`, `cu.io/pool`,
`cu.io/run`, `cu.io/release`, and `cu.io/materializer-role` from the validated
runtime graph while preserving `cu.io/owner=external-secret-import`. Explicit
conflicting labels reject; they are never silently corrected. Existing exact
data/type/labels verify successfully; changed material or ownership collides.
Later lifecycle execution must import the **same** key sets and bytes.

The ordering is explicitly:

```text
fresh admission -> owned cluster prerequisites -> protected host pki-publish
-> lifecycle Secret verification -> infra PKI verification Job Complete + read-only postJob attestation
-> release PKI verification Job Complete + read-only postJob attestation
-> database readiness -> union DB -> migrations -> replication
```

Host publication is a pre-Job prerequisite. PKI completion now additionally
requires verification of those exact prepublished bytes, not regenerated output.
The bridge does not broaden named bootstrap RBAC or add an in-Job
publication alternative. PKI Jobs cannot create absent Secrets under that RBAC,
and regenerated bytes cannot overwrite host-published immutable imports.

## Executable materializer and explicit init contract

The supported consumer entry is the existing suite executable, not a standalone
materializer binary:

```text
/app/cryptoutil pki-materialize \
  --instance=/etc/cryptoutil/instance/instance.yaml \
  --input=/etc/cryptoutil/instance/materializer.json
```

`cmd/cryptoutil` routes through `Suite` and `RouteSuite` to the
`pki-materialize` product handler, then the unchanged materializer library.
The canonical schemas and projection require this exact command/argument vector.
The operator must certify a suite image containing that executable; synthetic
fixture image digests are not certification.

Every workload requires explicit non-secret `materializerConfig` JSON. Its
`namespace`, `apiEndpoint`, and `apiServerName` match the selected target,
explicit API service IP, and `kubernetes.default.svc` TLS identity. `trustFiles`
exactly covers public-server, admin-server, health-client, and otlp-client roles;
PG additionally requires postgres-client. Each file must be a named file under
`/run/cryptoutil/trust`, mounted read-only from the workload's external trust
Secret. Coupled lifecycle admission rejects a requested filename absent from
that owned Secret's protected import mapping. Fixture roles explicitly share
`tls.crt`; operators must supply their approved issuer roots, not derive trust
from fetched leaf certificates or tokens.

`pods` exactly covers all reserved ordinals, including those beyond current
replicas. Each entry supplies `podName`, exact `expectedAdminDNS`, and
`expectedPublicDNSNames` including derived pod, ordinal-service, and workload
DNS identities. PG entries additionally provide the explicit full
`postgresqlClientBindings` map with `CommonName` and `DatabaseRole` for each
reserved pod; SQLite rejects those bindings. Identities are never inferred from
Secret content. The consumer rejects unknown/duplicate JSON properties,
unowned/duplicate pods, missing roles, unsafe paths, and invalid identity
bindings before token or Secret retrieval.

The init container receives explicit Downward API `POD_NAME` and
`POD_NAMESPACE`; both must match the supplied contract. Its read-only projected
API volume mounts at `/run/cryptoutil/pki/input`, aligning the shared consumer
paths `token` and `ca.crt`. Its separate issuer trust mount is read-only.
The application container does not receive that API token. The init invokes
the existing bounded consumer, verifies certificate/key/chain/SAN/EKU/PG
identity policy unchanged, and atomically publishes only its current pod's
role files into the existing ephemeral PKI volume.

Validation includes real suite routing, an injected API/filesystem consumer
flow with explicit issuer bytes and verified output, full-matrix native Helm
init/config decoding against the production consumer validator, and negative
render/projection/admission guards. These checks do not attest live image,
cluster, application TLS/client, or rollback acceptance.

## Executable database lifecycle

The existing suite now supports the canonical Job entry:

```text
/app/cryptoutil db-lifecycle --operation=init|migrate|replicate --input=/run/cryptoutil/secrets/OWNED-SECRET/lifecycle.json
```

`db-init` uses `postgres-bootstrap`; replication uses
`postgres-follower-bootstrap`; each migration uses its member's
`PSID-migration-db`. The chart requires the exact operation/manifest arguments,
and host execution validates protected manifests against the independently
selected catalog and every restricted mounted Secret key before transport.
No fixture database command is a supported production substitute.

The adapter uses actual per-service merged embedded migration filesystems and
production drivers. Fresh follower bootstrap applies these within an
owner-controlled transaction, verifies schema, then clears only newly generated
seed/metadata rows so native initial copy cannot collide with their primary keys.
Existing schemas are verify-only and never truncated/re-migrated. Initial copy
requires empty follower tables; exact owned subscriptions and database-scoped
readiness are verified before final schema/clean-cursor parity.

PostgreSQL requires explicit protected DSN, username/password, CA, certificate
and key files; subscriber transport additionally requires explicit server-visible
TLS paths. Roles, privileges, authentication policy, WAL/capacity and actual
image pins remain operator inputs, not invented defaults. Shared Jobs are PG
only; SQLite uses its existing per-pod startup migrations and has no logical
replication. Full manifest/credential/ownership details and the bounded
bootstrap-readiness caveat are in
`internal/apps-tools/deploymentrun/database/README.md`.

Real SQLite/injected orchestration/native Helm checks are local evidence, not
PostgreSQL or live cluster/image/client/rollback acceptance. Regenerating PKI
Job filesystem output is intentionally unsupported by lifecycle execution.

## Executable PKI verification and postJob attestation

The supported suite route is:

```text
/app/cryptoutil pki-verify --input=/run/cryptoutil/secrets/OWNED-SECRET/pki-verification.json
```

The protected version-1 manifest contains `files` (absolute projected paths
mapped to exact SHA256 digests) and `certificates` (explicit certificate, private
key, trust anchor, server/client usage, DNS names and common name). Fields are
strict; explicit empty arrays/strings are supplied where applicable. Every
mounted Secret key except the manifest itself must be hash-bound. Every
certificate/private key must be covered by a certificate/key/trust binding.
Trust and identity are explicit operator policy, never inferred from a token
or untrusted leaf. See `deploymentrun/materializer/VERIFY.md` for the contract.

The host prepares the whole plan before execution, using the same stronger
runtime ordinal/admin/PostgreSQL publication policy. A verification Job reads
only protected projections and applies the existing certificate/key/chain,
validity, signature, SAN/CN and EKU policy. It has neither API token nor Secret
write permissions and produces no material. After owned Job `Complete`, the
host must re-read and verify the Job UID/revision/spec/security and each mounted
Secret UID/type/full key set/full bytes/owner labels, then re-evaluate current
certificate and runtime policy. Only this barrier increments confirmed Jobs
and permits the dependent stage. `Complete` alone is rejected.

This deliberately resolves the supported **prepublished verify-only** contract.
It does not attest generation output:
`internal/apps-framework/tls/init.go:initRun` requires `--domain` and
`--output-dir`, calls `Generate`, and reports filesystem completion. It neither
publishes Kubernetes Secrets nor produces a host-verifiable output attestation.
The lifecycle chart gives the Job only memory-backed `/tmp` scratch and
read-only Secret/ConfigMap projections, with no durable output handoff.
`deploymentrun/lifecycle/kubectl.go:EnsureJob` verifies the desired spec;
`WaitJob` verifies owned UID and Job conditions, not generated certificate/key
bytes.

The current topology imports/publishes exact immutable material **before** these
Jobs. Freshly generated output is a different bundle; named bootstrap RBAC
cannot create absent Secrets, and verify-or-create cannot replace the existing
bundle. Reading a completed container through exec is not an available output
handoff, and logs/temporary host files are not an authorized secret transport.
Therefore Job `Complete` cannot truthfully attest publication of those outputs.

Generation-only Jobs are rejected before Prepare/import; no logs, exec,
temporary host files or broadened create permissions substitute for a protected
output handoff. An in-Job publication/regeneration alternative is not
implemented. It would require a separate owner-approved topology and immutable
material transition, not an operator credential to bypass this intentional gate.

## Remaining runtime prerequisites

The parity tests import actual fixture Secret objects and render lifecycle,
runtime, infrastructure, and controller-generated PVC identities against the
independent graph for the full registry, PG/SQLite selections, and reserved
ordinal bounds. They also verify restricted RBAC/token/resource/deadline fields
and reject forged chart permissions. Fixture digest pins and database arguments
are test data, not runnable operator configuration.

The runtime owner still must supply and certify real digest pins, protected
source material, service-specific commands/configuration, and the owned Kubernetes
target and prerequisites. Rendering all Jobs together does **not** enforce the
DAG; use the guarded execution adapter for completion barriers. Its injected
tests and native Helm interoperability do not certify real image commands,
Kubernetes Secret updates, union/replication behavior, application client
acceptance, or rollback on an actual cluster.

The checked-in model still has the preview revision `r1`; explicitly calculate
and approve its content revision from real operator input. The toolchain lock
pins infrastructure images/tools, not application, suite Job or materializer
images and runtime configuration. No certified runtime/lifecycle/publication
bindings for those inputs and protected material were present in the inspected
deployment input surfaces; unrelated existing protected files are not implicitly
approved sources. The current suite `pki-init`
requires `--domain`/`--output-dir` and writes filesystem certificates; test
positional arguments are not executable configuration or a Kubernetes Secret
publisher. The host bridge now supplies guarded publication of existing protected
outputs. The suite consumer entry/chart contract is implemented and locally
validated. Executable database migration/replication adapters and their protected
Job contracts and prepublished verify-only PKI completion attestation are
implemented with local evidence. In-Job generation/publication remains
intentionally unsupported; actual PostgreSQL union/copy behavior
and live image/cluster/client/rollback acceptance remain uncertified.

The read-only `bootstrap-preflight` admission command is separate. Its host
measurement, maximum application requests, explicit reserves, collision and
port checks are retained. Admission now requires explicit `-lifecycle-input`
and includes all selected Job requests alongside maximum application/init and
infrastructure requests plus explicit reserves. Runtime/infrastructure revisions
and quota headroom must also validate against that lifecycle input. A `-run`
override requires operator bindings already matching the new run; it never
rewrites source paths or creates material.

```powershell
go run ./cmd/cicd-deployment bootstrap-preflight `
    -runtime-input runtime-input.json -lifecycle-input operator.json -select tier:scope:mode `
    -reserve-cpu-milli EXPLICIT_RESERVE -reserve-memory-bytes EXPLICIT_RESERVE
```

No cluster was started for this slice.
For an owner-authorized local runtime, the existing model requests one K3s
server and two agents on `linux/amd64`, API `127.0.0.1:6445`, gateway
`127.0.0.1:8443`, API service IP `10.43.0.1`, service CIDR `10.43.0.0/16`,
pod CIDR `10.42.0.0/16`, and cluster domain `cluster.local`.
`k8s-toolchain.lock.json` pins K3s `v1.37.1-k3s1`, kubectl `v1.36.1`,
Helm `v4.3.1`, PostgreSQL 18, collector 0.162.0 and LGTM 0.36.0, with
image/binary digests. Use those exact lock references, not tags alone.
The canonical local-K3s Compose plan requires an explicit owner-only protected
token file at `./secrets/RUN/k3s-join-token` relative to its Compose file;
no token or target context is invented. Runtime additionally requires:

- Certified reachable image digests for the suite (including `pki-verify`,
  `pki-materialize`, `db-lifecycle`) and all selected applications.
- Exact selected/run-bound runtime/lifecycle/publication manifests, content
  revision, maximum reserved ordinals, resource requests/limits/quotas and
  positive physical-host/K3s reserves; fresh Docker capacity/collision/port
  admission evidence before starting owned cluster resources.
- Complete owner-protected source files with approved certificates/private
  keys/trust/credentials, verification digests/role/identity bindings, and
  materializer trust/Downward API inputs matching those exact sources.
- Explicit protected kubeconfig/context, owned Active namespaces/configuration,
  namespace labels/quotas/network policies and host named-resource permissions.
- For PG: leader/follower storage and readiness, explicit DSNs/CA/client
  certificate/key/role credentials, approved migration/replication privileges,
  server-visible subscriber TLS paths and adequate logical-WAL/copy capacity.

These are operator material/certification and live validation requirements,
not defaults supplied by local fixtures. Actual TLS clients, PostgreSQL initial
copy/reconciliation and failure/rollback behavior still need owned runtime E2E.
This slice's read-only availability snapshot found all four lock-pinned
infrastructure images already cached, a kubectl `v1.36.1` Windows binary whose
SHA256 matches the lock, and the pinned Helm used by native tests. Docker
reported 14 CPUs and 8,326,418,432 memory bytes, with no running containers.
That snapshot is not lifecycle-aware admission: exact selections, requests,
reserves, collision/port checks and operator certification are still required.
Race validation remains unavailable with the current CGO configuration; no
environment override is permitted.
