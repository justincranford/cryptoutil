# Infrastructure chart validation fixtures

These are render/schema fixtures only. They do not certify the 55 release matrix, install
Kubernetes resources, provision external Secrets/StorageClasses, initialize PostgreSQL, create
logical subscriptions, or validate runtime mTLS and persistence.

Run from the repository root with the pinned Helm 4.3.1 binary:

```powershell
helm lint --strict api/cryptosuite-registry/templates/kubernetes/cryptoutil-infra
helm template cu-review-infra api/cryptosuite-registry/templates/kubernetes/cryptoutil-infra --namespace cu-review-infra

helm lint --strict api/cryptosuite-registry/templates/kubernetes/cryptoutil-infra `
  -f api/cryptosuite-registry/templates/kubernetes/cryptoutil-infra/fixtures/values-sqlite-only.yaml
helm template cu-review-infra api/cryptosuite-registry/templates/kubernetes/cryptoutil-infra --namespace cu-review-infra `
  -f api/cryptosuite-registry/templates/kubernetes/cryptoutil-infra/fixtures/values-sqlite-only.yaml
```

`invalid-unpinned-image.yaml` must fail Helm's values-schema validation. The sparse overlay
intentionally inherits the other values from the chart's reviewed sample `values.yaml`; generated
deployments must supply the complete typed projection rather than relying on these sample
identities. `invalid-postgres-client-without-postgres.yaml` must fail the chart's consistency
guard because its inherited client list still requests PostgreSQL access.

The duplicate-name overlay must fail `helm template` with
`duplicate infrastructure ConfigMap name: postgres-leader`. It verifies that enabled components
cannot publish colliding ConfigMaps even when individual values pass DNS-label schema checks.

```powershell
helm template cu-review-infra api/cryptosuite-registry/templates/kubernetes/cryptoutil-infra `
  --namespace cu-review-infra `
  -f api/cryptosuite-registry/templates/kubernetes/cryptoutil-infra/fixtures/invalid-duplicate-configmap.yaml
```

The chart owns only run-scoped PostgreSQL and telemetry objects. Cluster gateway/controller,
TLS-passthrough routes, K3s/Kubernetes StorageClass creation, namespace labels, external
Secret/PKI publication, database catalog/role creation, migrations, and logical follower
reconciliation are lifecycle inputs/actions, not Helm-rendered secret values or hooks.
