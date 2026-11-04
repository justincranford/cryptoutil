# Runtime chart validation fixtures

These are Helm schema/render fixtures, not runtime acceptance. The min/max fixture uses
`/bin/true` and a pinned existing image solely to exercise chart structure; it must never be
installed. It demonstrates one active StatefulSet/PVC ordinal (`minReplicas: 1`), three
predeclared ordinal services and per-ordinal PKI Secret references (`maxReplicas: 3`).

Run from the repository root with Helm 4.3.1:

```powershell
helm lint --strict api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime `
  -f api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime/fixtures/values-min1-max3.yaml
helm template cu-s-sm-kms-disk api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime `
  --namespace cu-review-s-sm-kms-disk `
  -f api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime/fixtures/values-min1-max3.yaml
helm lint --strict api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime `
  -f api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime/fixtures/values-min1-max3.yaml `
  --set-string 'workloads[0].credentialSecretNames[0]=sm-kms-e2e-tls'
```

The final command must fail values-schema validation: E2E client TLS material is not an
application credential. Generated application configuration files contain no secret values.
Explicit credential Secret references are mounted read-only; the init materializer's RBAC
contains only that workload's per-ordinal server, health-client, and OTLP-client Secret names
through `maxReplicas`. The init token is projected only into init and uses the API server's
default audience unless a verified provider audience is explicitly supplied.

For a provider whose API server audience has been independently verified, the explicit override
is rendered as supplied:

```powershell
helm lint --strict api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime `
  -f api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime/fixtures/values-min1-max3.yaml `
  --set-string 'apiServerAudience=https://verified-provider.example/issuer'
helm template cu-s-sm-kms-disk api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime `
  --namespace cu-review-s-sm-kms-disk `
  -f api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime/fixtures/values-min1-max3.yaml `
  --set-string 'apiServerAudience=https://verified-provider.example/issuer'
```

The explicit override validates chart wiring only; it does not prove that the local or any
provider API server accepts that audience. Verify issuer/audience configuration before using
the override.

The following render commands must also fail with the indicated template diagnostic. They
exercise cross-object references that JSON Schema cannot validate:

```powershell
helm template cu-s-sm-kms-disk api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime `
  --namespace cu-review-s-sm-kms-disk `
  -f api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime/fixtures/values-min1-max3.yaml `
  --set-string 'workloads[0].frameworkConfigMapName=missing-framework'
# Expected: references missing framework ConfigMap

helm template cu-s-sm-kms-disk api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime `
  --namespace cu-review-s-sm-kms-disk `
  -f api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime/fixtures/values-min1-max3.yaml `
  --set-string 'workloads[0].frameworkConfigMapName=sm-kms-domain'
# Expected: framework ConfigMap must match its PS-ID and framework role

helm template cu-s-sm-kms-disk api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime `
  --namespace cu-review-s-sm-kms-disk `
  -f api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime/fixtures/values-min1-max3.yaml `
  --set-string 'configurationMaps[1].name=sm-kms-framework'
# Expected: duplicate configuration map name
```
