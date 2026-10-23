# Kubernetes deployment design

**Status**: Planning draft for human review; no runtime cutover is authorized.
**Created**: 2026-10-22
**Last Updated**: 2026-10-22

## Scope and decisions

Replace application Docker Compose models with generated Helm/Kubernetes artifacts.
Retain Docker Compose **only to host the local K3s cluster**, not PostgreSQL, applications,
PKI, or telemetry outside Kubernetes. Target the same charts at future Kubernetes clusters;
that target is render/schema-checked only, explicitly **not locally runtime-certified**.
No chart, cluster, probe implementation, or existing deployment is changed by this planning work.

The registry contains **7 PS-IDs, 3 products, 1 suite**, not four products:

| Tier | Scopes | Members |
|---|---|---|
| PRODUCT-SERVICE | sm-kms; identity-authz/idp/rs/rp/spa; skeleton-template | One PS-ID each |
| PRODUCT | sm; identity; skeleton | 1; 5; 1 PS-IDs respectively |
| SUITE | cryptoutil | All 7 PS-IDs |

Therefore 7 + 3 + 1 = **11 registered tier scopes**, not the requested 7 + 4 + 1 = 12.
The fourth product has no current registry identity or membership; it is **unresolved**, not
silently renamed, invented, or treated as infrastructure. Older handbook references to a `pki`
product predate PKI consolidation into sm-kms. A fourth group needs an explicit registry/group
decision before implementation readiness. The generator follows actual membership and will
reject `p:pki:pg`. It does not manufacture a fourth deployable product.

### Evidence: effective Compose versus historical prose

| Design intent | Current evidence | Kubernetes disposition |
|---|---|---|
| Registry-driven entities and tier host-port offsets | [registry](../../api/cryptosuite-registry/registry.yaml) | Registry remains canonical; no copied product lists in the generator |
| Recursive composition plus tier overrides | [service entrypoint](../../deployments/sm-kms/compose.yml), [product](../../deployments/identity/compose-apps.yml), [suite overrides](../../deployments/cryptoutil/compose-apps-overrides.yml) | Explicit values composition; one independent release/namespace per selected tier scope and mode |
| Four app identities per member | [sm-kms apps](../../deployments/sm-kms/compose-apps.yml) | PG and SQLite replica groups preserve ordinals/identities, rather than doubling every existing slot |
| In-memory SQLite versus file-backed SQLite | [durable overlay](../../deployments/sm-kms/compose-sqlite-durable.yml) | Separate mem/disk pools; PVC per disk pod; no shared SQLite file |
| Physical PostgreSQL sharing, logical isolation | [shared PostgreSQL](../../deployments/shared-postgres/compose.yml), [database catalog](../../api/cryptosuite-registry/templates/deployments/shared-postgres/init-databases.sql) | Run-owned leader and logical follower; DB/role per release/member; optional external provider later |
| Shared telemetry with collector hop | [shared telemetry](../../deployments/shared-telemetry/compose.yml) | Run-owned collector replicas -> run-owned LGTM; application never bypasses collector |
| Tier-level unseal/pepper sharing | [product override template](../../api/cryptosuite-registry/templates/deployments/__PRODUCT__/compose-apps-overrides.yml), [suite override template](../../api/cryptosuite-registry/templates/deployments/__SUITE__/compose-apps-overrides.yml) | Preserve effective sharing **within a release**; never share across run/mode/tier boundaries |
| Per-instance TLS, DB mTLS, admin loopback | [instance framework config](../../deployments/sm-kms/config/sm-kms-app-framework-postgresql-1.yml), [TLS taxonomy](../deployment-templates.md) | Preserve all trust purposes and specific leaf identities; do not substitute an unauthenticated health port |

Observed templates share the unseal shard set across PS-IDs in a product/suite, whereas some
handbook prose describes per-PS-ID shards. Preserve observed mounts during migration, and reconcile
the prose in the implementation documentation phase; do not change HKDF inputs or silently
re-encrypt persisted data as a side effect. Source references, not secret values, establish this.

Current mixed-memory topology: 7 standalone members + (1 + 5 + 1) grouped members + 7 suite
members = **21 tier/member occurrences**; each has SQLite-1/2 and PostgreSQL-1/2, yielding
`21 * 4` = **84 legacy app identities**. Shared PostgreSQL has 21 databases on each of its two
nodes, not 84 databases or 21 PostgreSQL containers. Global named volumes, fixed Grafana
host ports, shared network names, and unqualified helper names are collisions to eliminate,
not isolation guarantees to copy.

Durable SQLite overlays currently exist at standalone tier only. Extending that capability
to grouped/suite releases is a **proposed generalization**, not a claim about current support.
Compose app-only fragments become chart composition internals, not an additional deployment tier.
Selecting a named Compose instance maps to an exact pod/instance Service; a diagnostic
single-pod run can override bounds to 1 explicitly. Test-only OTLP host exposure becomes
an explicitly authorized temporary port-forward, not a permanent public collector Service.

## Deployment modes, replicas, and storage semantics

| Mode | Pools per included PS-ID | Default pods per PS-ID | Purpose |
|---|---|---:|---|
| `pg` | `pg` | 2 | Recommended multi-pod shared-state mode |
| `mixed-mem` | `pg`, `mem` | 2 + 2 | Exact legacy four-slot coverage, including disposable SQLite |
| `mixed-disk` | `pg`, `disk` | 2 + 2 | Same coverage with durable independent SQLite |
| `mem` | `mem` | 2 | Lightweight SQLite-only diagnostics |
| `disk` | `disk` | 2 | Durable SQLite-only diagnostics |

Every pool gets `min=2,max=2` from configuration, not Go/Helm literals. A normal `pg`
deployment thus has two pods per service; a mixed mode intentionally has four, not eight.
Legacy `postgresql-1/2` map to `pg-0/1`; `sqlite-1/2` map to `mem-0/1` or `disk-0/1`.
An ordinal beyond 1 is an explicitly new replica identity. A pod restart retains its identity.

Bounds can later be changed globally or with `PSID/pool` overrides in
[k8s-model.json](k8s-model.json), including an intentional minimum of 1; the old plan's
hard-coded two-pod floor is superseded. Validate `1 <= min <= max`. PostgreSQL HPA uses
these bounds (equal bounds are valid); require CPU requests, metrics-server availability,
and tests before enabling variable scaling. Pre-provision identities through `max`.
Do not let Helm's `spec.replicas` fight an active HPA: initialize at min, then let HPA own it.

SQLite replicas are **not a shared-state HA service**. Each has an independent in-memory DB
or its own PVC. No pooled customer-traffic Service or round-robin gateway may hide this.
Expose explicit instance Services/hostnames; headless discovery does not mean pooled routing.
For SQLite, min is the configured initial count and max a capacity bound; scaling is an
explicit operator action within the bounds, never CPU-driven HPA. Preserve ordinal PVCs on
scale-down; new mem pods start empty, new disk ordinals get new databases.

Use StatefulSets for these initial pools because the migration preserves distinct per-instance
configuration/certificate identities even for PG-backed replicas. Generate the pod configuration
from a validated common config plus a Downward API **pod-name file**, not environment-variable
application configuration. A future identity-free PG Deployment is a separate optimization.
Prefer parallel pod management with independently bounded initialization; startup must not
depend on another app replica becoming ready.

Dependency replicas are a different contract: leader=1, logical follower=1, collector=2,
LGTM=1 initially. A logical follower is OLAP replication, **not automatic failover**.
Do not set two primary PostgreSQL pods or two LGTM writers merely to satisfy the app count.
A PostgreSQL HA operator and production observability backend are future provider choices.

## Naming grammar and isolation

Inputs: registry, versioned model, run identity, revision, cluster/local DNS suffixes,
mode definitions, replica bounds/overrides, and later provider/resource configuration.
All names are deterministic functions of inputs. IDs are lowercase DNS labels; run is
1-8 lowercase alphanumeric characters starting with a letter. Reject overlong names,
duplicate identities, unknown members/pools/modes/overrides, empty groups, and invalid bounds.
Never truncate names or append random suffixes to resolve collisions.

```text
tier        := s | p | u
scope       := registered PS-ID | registered product | registered suite (according to tier)
mode        := configured mode name (initially pg | mixed-mem | mixed-disk | mem | disk)
key         := tier ":" scope ":" mode
namespace   := "cu-" run "-" tier "-" scope "-" mode
infra-ns    := "cu-" run "-infra"
helm-release:= "cu-" tier "-" scope "-" mode
pool        := pg | mem | disk
workload W  := PS-ID "-" pool
pod         := W "-" zero-based-ordinal
instance-svc:= W "-i" one-based-ordinal-padded-to-at-least-two-digits
headless-svc:= W "-pods"
pooled-svc  := W                         # PG only
leaf-secret := pod "-tls"
sqlite-pvc  := "data-" pod               # disk only; claimTemplate name is data
pod-dns     := pod "." headless-svc "." namespace ".svc." clusterDomain
service-dns := service "." namespace ".svc." clusterDomain
local-sni   := instance-svc "." namespace "." localDomain
pooled-sni  := W "." namespace "." localDomain   # PG only
SQL-stem    := underscore(run,tier,scope,mode,PS-ID)
database    := "d_" SQL-stem
runtime-role:= "u_" SQL-stem
publication := "pub_" SQL-stem
subscription:= "sub_" SQL-stem
job         := purpose "-" revision
```

Kubernetes identities are `(namespace, kind, name)`; namespace names are cluster-scoped.
Helm names may repeat across namespaces but must be <=53 characters. All generated
resource names are <=63 characters, all SQL names <=63 ASCII bytes, full DNS names <=253.
Namespace prefixing plus the structured identity avoids ambiguous parsing of hyphenated IDs.
The generator checks the **complete expanded resource inventory**, not only the short names.
Runtime-assigned IPs, UIDs, PV names, and Job pod suffixes are discovered by labels and
ownerReferences, never generated. PVs must be owned/checked by claim UID, not predicted name.

An exact selector such as `p:identity:mixed-disk` projects one complete release from the
full graph; all selectors together can coexist. Different runs coexist in the same cluster
using different namespaces. One run has a deliberate shared-infrastructure namespace.
Adding a release to that run reconciles its database/trust allowlist without deleting other
releases' entries; bootstrap owns the union catalog, applications cannot overwrite it.

Isolation layers:

1. **Config**: typed precedence = canonical defaults < tier/group < mode < PS-ID < explicit
   instance override. Generate app YAML with kebab-case keys, mount read-only. Materialized
   secrets retain precedence over YAML/CLI; never put credentials in values, ConfigMaps, or
   environment variables. Validate merged output, reject unknown/conflicting keys.
2. **Selectors**: immutable run/release/PS-ID/pool labels; per-instance Service additionally
   selects `statefulset.kubernetes.io/pod-name`. No cross-namespace Service selectors.
3. **Network**: deny ingress/egress by default; permit DNS TCP/UDP 53 to cluster DNS, public
   HTTPS from the selected gateway/test runner, approved within-release identity API edges,
   PG 5432 to infrastructure, collector 4317/4318, and explicit bootstrap-only Kubernetes
   API access. Infrastructure has reverse allowlists for the exact release namespaces and
   pod labels. Cross-namespace selector clauses must be ANDed, not separate OR rules.
   Allow intra-infra collector->LGTM and follower->leader only on configured ports.
   Configured enterprise federation/CRL/OCSP external endpoints need explicit egress rules.
4. **Data**: PG DB and least-privilege role per release/member, read-only follower roles for
   analytics, separate migration credentials, no SQL superuser in apps. Schema tenancy
   inside a member remains the application's existing responsibility.
5. **Crypto**: release-scoped unseal/pepper sharing matches effective Compose; never share
   them across modes/runs. Per-member credentials stay per-member. No `.never` marker is
   loaded as a credential. Import old material for migrated data, generate new material only
   for a new identity; a mode change is not an in-place data migration.
6. **RBAC/storage**: app has no mounted API token, no signer keys, no access to other
   namespaces. Provision/migration Jobs use separately scoped service accounts/roles.
   Workload init is restricted to `get` on its explicit per-ordinal leaf Secret names through
   max; no list/watch/write. Only init mounts its projected API token. NetworkPolicy
   applies to the entire pod, not individual containers: its restricted API egress
   allowance is pod-scoped, even though the application has no API credentials.
   Init materializes the selected ordinal into a memory emptyDir mounted read-only by app.
   A shared StatefulSet service account is a **workload**, not per-pod, RBAC trust boundary;
   it can read that workload's leaf Secrets. Do not claim stronger per-pod RBAC isolation.

Namespaces alone do not isolate network traffic. Test the active K3s NetworkPolicy controller;
creating objects is not evidence they are enforced. No hostNetwork, hostPort, privileged app
containers, wildcard namespace allowlists, public admin Service, or global Helm fullname override.
Local bootstrap nodes are privileged Docker infrastructure, confined to trusted developer hosts.
Shared LGTM is a run-level trust boundary, not isolated telemetry tenancy; use a separate run
or separate backend for mutually untrusted deployments. Retain run/release/mode/PS-ID/pod
telemetry attributes to prevent collisions and make shared-backend provenance auditable.

## Helm generation and local cluster boundary

Proposed future layout (not implemented here):

```text
api/cryptosuite-registry/k8s/                 typed inputs/schema + canonical templates
internal/apps-tools/deploymentgen/            pure Go graph, naming, rendering, validation
cmd/cicd-deployment/                          generation/check/render/lifecycle CLI
deployments/k8s/charts/cryptoutil-runtime/    one generic chart; iterate typed member/pool list
deployments/k8s/charts/cryptoutil-infra/      PG/telemetry/provider resources
deployments/k8s/generated/<run>/<key>/        values, chart metadata, expected inventory
deployments/k8s/local/                       generated K3s-only Compose + node configs
```

One runtime chart serves all tiers; do not nest seven duplicated subcharts plus four
handwritten umbrella charts. Use registry-derived typed member lists and common helpers.
No chart `lookup`, random generation, timestamps, environment-dependent inputs, `latest`
images, or secrets in deterministic outputs. Pin image/chart/K3s/Helm versions and digests
in a toolchain lock; validate API compatibility before adopting versions. Dependency locks
and vendored canonical templates must make offline rendering reproducible.

Generate chart/values/schema/K3s bootstrap files from reviewed templates, not raw YAML
string concatenation. Humans edit the model or canonical templates, never generated copies.
Helm provides the final Kubernetes renderer; generated expected inventory independently
checks Helm output. `cicd-lint` remains detection/formatting-only; do **not** add generation
to it. The planning script is deliberately outside production package discovery.
Helm values may use idiomatic camelCase; **application config YAML remains kebab-case**.

Local choice: **K3s directly in Docker Compose**, one server and two agent nodes by default,
named `cu-<run>-server-0` / `cu-<run>-agent-0/1`, project `cu-<run>-cluster`, network
`cu-<run>-nodes`. This satisfies the requested Compose-hosted cluster without nesting k3d
inside a Compose container or mounting the Docker socket into a controller. Native k3d is
an optional alternative bootstrap adapter, not a nested orchestration layer; use its
versioned config schema if adopted. Both would run the same charts. Only the chosen K3s
Compose adapter is an initial runtime acceptance target.

Bind host Kubernetes API and development TLS gateway to **127.0.0.1**, default ports 6445
and 8443 respectively; allocate/override ports per cluster run after collision checking.
Within a shared cluster, namespaces/SNI distinguish releases without more host ports.
Use a locked TLS-passthrough-capable gateway for end-to-end public mTLS; ordinary TLS
termination ingress is not equivalent. Install gateway API/CRDs/controller once as
cluster-owned prerequisites, not in each app release. Its cluster-scoped names and any
StorageClasses are namespaced by run/cluster identity and never deleted by app uninstall.
Generate loopback DNS/hosts mappings (do not assume arbitrary nested `.localhost` DNS works).
Public cert SANs include service, pod, and external SNI names. Preserve legacy numeric
host-port access only through an explicitly generated optional compatibility map.
The inventory currently names application/dependency resources; cluster controllers, gateway
routes, generated bootstrap manifests, DNS hosts mapping, and PVs are future renderer outputs.
Their omission is explicit; this is not a claim of a complete deployable Helm resource list.

Kubeconfig exports use the host-reachable API, with CA validation and private file permissions;
agents use the internal server DNS. Generate K3s node config files; mount join tokens through
Docker secret files, not inline Compose values/environment. Preserve node data on routine Compose down/up. Multi-node
spread protects against a worker loss, **not** loss of the Docker host or single control plane.
Configure topology spread, PDBs, resource requests/limits, quotas with rollout headroom, and
graceful termination; PDB is not a hard guarantee of ready-pod minimum during failure.
All-variations defaults require `21 * (1+2+2+1+1) * 2` = **294 app pods** plus
1 leader + 1 follower + 2 collectors + 1 LGTM = **299 steady-state workload pods**, excluding
Jobs and cluster system/gateway pods. At a provisional 128Mi request per app pod that is
`294 * 128Mi` = **36.75Gi just for applications**; size from measurements, not this estimate.
Preflight must reject insufficient resources or install a requested subset, never silently
lower replica bounds. At default bounds, disk pools need `21 * 2 * 2` = **84 SQLite PVCs**.

Future Kubernetes provider values explicitly select storage class, ingress/gateway, external
DB/telemetry, secret provider, DNS, and cluster domain. Do not assume K3s local-path, Docker
volumes, or localhost there. Render/schema tests demonstrate portability of manifests only;
future cluster runtime support stays untested until a separately authorized acceptance run.

## PKI, initialization, migrations, and persistence

Preserve all 14 [certificate categories](../deployment-templates.md), with grammar-derived
SANs/identities rather than old Compose DNS names:

| Category | Role | Owner/consumer |
|---|---|---|
| 1 | Global public server root/issuing CAs | Restricted provisioner; public trust bundles |
| 2 | Collector/LGTM server leaves | Infra workloads |
| 3 | PS-ID public server leaves | Individual application pod |
| 4 | PS-ID public client root/issuing CAs | Release/member trust domain |
| 5 | PS-ID public client leaves, including realm identity | Authorized clients/E2E, not every app |
| 6 | Private admin mutual root/issuing CAs | Per-pod admin/probe trust |
| 7 | Private admin server/client leaves | Pod admin server + in-container probe client |
| 8 | Collector/LGTM client root/issuing CAs | Infra receivers' trust bundles |
| 9 | Application/infra OTLP client leaves | App->collector and collector->LGTM |
| 10 | PostgreSQL server root/issuing CAs | DB server trust bundles |
| 11 | PostgreSQL leader/follower server leaves | Corresponding DB pod |
| 12 | PostgreSQL client root/issuing CAs | Database client trust |
| 13 | PostgreSQL replication client leaves | Leader/follower replication roles |
| 14 | PS-ID PostgreSQL client leaves | Each app pod's authenticated DB role |

Deterministic **names/configuration** do not imply deterministic private keys. Generate secret
material with approved CSPRNG or import it securely; never commit it or derive it from run names.
Static Secret references + read-only file projections replace Docker secrets. Provisioner has
signing keys; workloads have only needed leaves/trust bundles, mounted with restrictive modes,
non-root ownership and memory-backed scratch space. Kubernetes Secret is not encryption by itself:
configure K3s/K8s encryption at rest, RBAC, backup protection, and avoid Helm release Secret data
containing private material. No giant shared cert-directory mount into every pod.

Initialization DAG (Go lifecycle controller/CLI, not unreliable Helm template ordering):

1. Validate plan/toolchain/resources; create/label owned namespaces, policies, storage prerequisites,
   and restricted bootstrap identities. Import/generate authorities and preserve existing material.
2. Provision infra server/replication leaves and release/member client issuers. Build explicit
   trust bundles. Start leader and telemetry receiver/backend independently; follower server can
   start before subscriptions, but subscription setup is separate.
3. Reconcile logical DBs and least-privilege roles for the selected **union** catalog. Provision
   app/pod leaves and credentials, then run one versioned migration Job per release/member against
   the leader with an advisory migration lock and bounded timeout. Apps verify schema version;
   do not race migrations on two starting replicas. SQLite runs migrations on its own DB.
4. Start apps with init materialization, secret/schema validation, and startup probes.
   A `pki-init-<revision>` or migration Job has a content-addressed revision in implementation;
   `r1` is only the human-review example. Unchanged revision is idempotent; incompatible Job
   changes create a new name, no illegal in-place mutation of immutable Job fields.
5. After migrations, schema-copy/reconcile each corresponding subscriber DB, then create/check
   its logical subscription. DDL and schema evolution require explicit reconciliation; logical
   replication does not replicate DDL. Retain the existing client_secret->client_secret_hash
   compatibility handling. Follower availability/lag must not gate OLTP app readiness.
6. Wait for all requested workloads, verify endpoints and functional behavior, record graph/input
   digests and non-secret deployment evidence. Polling has deadlines and explicit failures.

Do not attach app readiness to follower subscription readiness while subscription setup waits
for app readiness: that is a startup cycle. Do not put long migrations in a startup handler,
use probe success as migration authorization, or rely on a pre-install hook that waits for
resources Helm has not yet created. Jobs, phase ownership, and failure propagation must be
visible; Helm hooks alone are not an orchestration DAG.

PVC retention: keep on upgrade, scale-down, uninstall and routine namespace lifecycle actions;
app uninstall must not delete namespaces. Provision retained claims under explicit bootstrap
ownership, with StatefulSet retention policies and Helm keep annotations where applicable.
Select a **Retain** reclaim StorageClass for persistent data; K3s default local-path deletion
behavior is not acceptable without an explicit disposable override. Local-path is node-local:
PVC retention is not worker/host fault tolerance; use node-affinity-aware scheduling, backups,
and a replicated CSI provider later for cross-node failover.
Back up PostgreSQL, SQLite, authority keys, unseal shards and pepper together; test restore
with a durable API record before cleanup. Never automatically regenerate keys on reinstall.
Namespace deletion can remove retained PVC objects even with retained PVs, so cleanup must check
run ownership labels **and recorded UIDs**, disposable consent/configuration, external dependency
refs, and required backups. Namespace names alone are not authorization.
Deleting one app release never deletes infra, another release's DB, or shared telemetry.

## Framework probe contract

Current [admin implementation](../../internal/apps-framework/service/server/listener/admin.go)
has livez/readyz but no startupz; readyz reads a mutable `ready` flag, not current dependencies.
[CLI startup](../../internal/apps-framework/service/cli/server_start.go) sets ready before
starting listeners; the older [listener lifecycle](../../internal/apps-framework/service/server/application/application_listener.go)
also sets it early. Therefore chart probes alone cannot satisfy the requested semantics.
Extend the **shared framework**, CLI/router/usage/constants, generated PS-ID surfaces and
canonical templates, builder/lifecycle, test helpers and all seven services consistently.
Do not implement seven independent probe systems. Preserve existing paths and intentional
shutdown compatibility, documenting changed liveness-during-drain semantics.

| Probe | Framework endpoint | Predicate | Failure |
|---|---|---|---|
| startup | `/admin/api/v1/startupz` | One-way per-process initialization latch reached only after config/secrets, TLS/barrier/session setup, expected schema, routes, both listeners and first required dependency observations succeed | Restart only after startup budget expires |
| liveness | `/admin/api/v1/livez` | Admin event loop serves a cheap response; no DB/OTLP/network checks, startup latch, or cluster-wide synchronization | Restart after threshold |
| readiness | `/admin/api/v1/readyz` | Public listener responsive, startup complete, not draining, required DB and collector observations healthy and fresh | Remove this pod from Service endpoints; never restart for dependency outage |

Startup probes disable readiness/liveness until they first succeed; kubelet then stops
calling startup. A monotonic in-memory startup latch is appropriate here. It resets on process
restart, not on a temporary dependency outage or PKI reload. Initialization that cannot recover
should fail explicitly rather than spin forever. Never require every sibling replica, follower,
or suite service to be ready before one process may start.

**"Instantaneous, stateless" needs a precise boundary.** A handler cannot synchronously prove
remote health with zero latency, zero I/O and no remembered observations. Choose fast,
read-only, non-mutating handlers over immutable per-pod health snapshots. A bounded background
monitor checks actual dependencies; the monitor necessarily retains state/timestamps.
No request counters, blocking remote calls, lazy initialization, retries, migration work or
status mutations inside probe handlers. This is explicit bounded-staleness, not a claim of
perfect instantaneous knowledge. Direct timed pings in every handler are the rejected alternative.

Provisional configurable monitor defaults: period 2s with jitter, check timeout 500ms, stale
after 6s. DB check uses the real application pool (`PingContext` and a minimal schema-aware
read), not a separate credentials-only connection. Collector check verifies its configured
mTLS receiver reachability and receiver health contract; the extension's process-health
endpoint alone **does not prove exporter delivery**. Receiver health/pipeline status must
be configured/tested; report stale/failed receiver checks as unready. Collector->LGTM outages
are visible as degraded/backpressure and trigger readiness failure only when the configured
required collector acceptance policy fails; no global dependency restart cascade.
Actual telemetry delivery remains an E2E acceptance assertion.

Include bounded local-public-loop responsiveness observations (verified TLS, no recursive
readyz calls); otherwise admin responsiveness could mask a dead public listener. During a
monitor stall, stale snapshots fail closed. Never cache an authorization decision in health
state. Expose sanitized reason codes/observation ages, not DSNs, tokens or key material.

Framework handlers target <=10ms p99 under the defined local test load. That bound excludes
exec creation and TLS handshake; measure end-to-end CLI probe latency separately under load.
Initial chart values (tune from startup/load measurements):

| Probe | Period | Timeout | Failure threshold | Success threshold |
|---|---:|---:|---:|---:|
| startup | 5s | 2s | 60 | 1 |
| liveness | 10s | 2s | 3 | 1 |
| readiness | 2s | 2s | 1 | 1 |

Startup budget approximately `5s * 60 = 300s`; do not disguise slow boot with oversized liveness
delays. Worst-case dependency failure-to-unready budget includes monitor period/timeout,
freshness cutoff, probe period and endpoint propagation; test each separately with a fake clock
and live faults. Readiness recovery must not clear startup or restart the container.

**Transport**: use Kubernetes **exec** probes running the existing PS-ID binary's
`startupz`/`livez`/`readyz` CLI with explicit `--cacert`, `--cert`, `--key`, and loopback URL.
Add startupz in the framework router, shared usage, seven service entrypoints/templates and
health helpers; give the HTTP client a bounded context shorter than probe timeout.
Do not use shell/curl/wget, unauthenticated gRPC, or kubelet HTTP probes against admin:
HTTP probes target pod IP, cannot present this admin client key, and HTTPS kubelet probes
do not provide the existing certificate-verification guarantee. No need to expose admin 9090.
Exec overhead for 294 pods is a real capacity measurement, not a reason to weaken mTLS.

Termination: SIGTERM/preStop first sets draining/unready, allows endpoint propagation,
then drains public work, stops monitors, closes DB/telemetry and exits within configurable
terminationGracePeriodSeconds. Liveness remains true while deliberately draining; shutdown
does not mean deadlock. Reuse the authenticated shutdown client without double-draining and
ensure preStop plus process drain fit **one** termination budget.

## Verification and cutover acceptance

Planning checks now: exact matrix/membership counts, default legacy-slot mapping, config
validation, bounds/override behavior, full versus selected graph equivalence, deterministic
bytes, resource/SQL/DNS uniqueness and length checks, other-run isolation, and stale-tree
detection. These prove the proposed grammar, **not** deployability.

Implementation validation must include:

- Generate twice, byte-compare; check no untracked generated drift; lint Helm; render every
  selector and the all-variations union; schema-check against pinned Kubernetes versions.
- Independently assert actual rendered kinds/names/selectors/SANs/secrets/network edges,
  replica bounds/quotas/requests/probes, PVC policies and DB mapping against JSON inventory.
  Do not use the same template output as both expected and actual oracle.
- Build/test the shared probe lifecycle with table-driven fake-clock/failure cases, Fiber
  handler tests, shutdown/startup races and required latency thresholds; preserve parallelism.
- On local K3s, deploy each tier/mode and all concurrently; compare Kubernetes owner UIDs,
  active pod/claim counts and endpoint membership with expected active/min versus max capacity.
- Authenticated functional requests through **each** ready PG pod in all seven services,
  both `/service/` and `/browser/` contracts; verify cross-pod data, PKI issuance/revocation/
  CRL allocation, shared-session behavior, DB mTLS `client_dn`, and OTLP->LGTM delivery.
- SQLite per-instance independence, memory-loss semantics, durable ordinal continuity;
  never pass a shared-state assertion by distributing SQLite traffic randomly.
- Deny cross-release/run traffic and SQL access, prove CNI enforcement, reject TLS mismatches
  specifically with TLS errors, verify no admin Service route, and exercise lost collector/DB,
  stale monitor, public-loop failure, pod kill, rollout, scaling, and migration failure.
- Durable-record and key continuity across pod/node/Helm upgrade, backup/restore, uninstall/
  reinstall and explicit disposable cleanup; verify other releases/data remain untouched.
- Enforce planned resources with an allowlist for controller-created system/Job resources;
  an unexpected workload or missing dependency fails inventory parity, not just total counts.

Keep Compose apps as the baseline during implementation comparison, then retire their
service/E2E orchestration after acceptance, updating canonical templates, validators,
registry contracts and handbook propagation together. Keep only local **cluster** Compose.
Do not remove the current working deployment while the replacement is only render-tested.
Use an explicit non-secret source-to-target migration map: each existing tier/member database,
SQLite ordinal and secret reference maps to one selected run/tier/scope/mode identity.
Additional simultaneously installed modes start with new data and fresh material; never
auto-clone an old DB or re-use its unseal set across modes. A data fork with new unseal
material requires a separately verified re-encryption operation. Quiesce writers, take
backups, migrate, verify records and key continuity, switch traffic, and retain a tested
rollback route; Helm rollback alone cannot undo a database schema/data migration.

The pre-existing PKI hot-reload track remains Phases 4-6 of [plan.md](plan.md), distinct from
probe state and deployment graph. No trigger/allowlist decision is implied by this design;
the allowlist/trigger design is itself a future phase task, not missing input for this request.

## Regeneration and review

From repository root, PowerShell syntax (the Go program is cross-platform):

```powershell
$src = @(
    '.\docs\framework-v26\k8s-gen.go',
    '.\docs\framework-v26\k8s-model.go',
    '.\docs\framework-v26\k8s-inventory.go',
    '.\docs\framework-v26\k8s-render.go'
)
go run $src -write
go run $src -check
go run $src -format json -select p:identity:mixed-disk
go run $src -select s:sm-kms:pg -run trial
go test ($src + @('.\docs\framework-v26\k8s-model_test.go', '.\docs\framework-v26\k8s-gen_test.go')) -shuffle=on
```

Without flags, output is a Markdown preview; JSON is machine-readable names/config only.
`-write` changes only [k8s-deploy-tree.md](k8s-deploy-tree.md); `-check` compares exact bytes
and fails on drift. A subset or alternate-run preview cannot overwrite the exhaustive tree.
Files use `//go:build ignore` and explicit named-file execution to avoid adding a planning
package to production builds or the CI package/coverage baseline. Migrate this prototype
into tested production tooling during Phase 2; it does not belong in `cicd-lint`.
The prototype inventory intentionally does not contain private material or live-cluster access.

Review the generated [full tree](k8s-deploy-tree.md) before authorizing implementation.
The current example is exhaustive over **registered scopes x the five configured modes**,
not all possible future input values or a fabricated fourth product. Resource limits, provider
implementations and full chart generation remain implementation work, with no current runtime
success claim.

## Verified upstream references

Retrieved 2026-10-22; pin implementation versions rather than adopting rolling-doc versions:

- [Kubernetes probe semantics](https://kubernetes.io/docs/concepts/workloads/pods/probes/)
- [Probe configuration and transports](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [StatefulSet identity, headless Services, PVC retention](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
- [NetworkPolicy enforcement and selector semantics](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [Persistent volume lifecycle](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
- [K3s requirements](https://docs.k3s.io/installation/requirements)
- [K3s networking](https://docs.k3s.io/networking/basic-network-options)
- [k3d versioned config, if the optional adapter is adopted](https://k3d.io/stable/usage/configfile/)
- [Helm dependency composition](https://helm.sh/docs/chart_best_practices/dependencies/)
- [Helm values conventions](https://helm.sh/docs/chart_best_practices/values/)
- [Handbook deployment architecture](../ENG-HANDBOOK.md#12-deployment-architecture),
  [testing](../ENG-HANDBOOK.md#10-testing-architecture),
  [quality](../ENG-HANDBOOK.md#11-quality-strategy), and
  [coding](../ENG-HANDBOOK.md#14-development-practices)
