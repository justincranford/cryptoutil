---
name: secret-gen
description: "Generate Docker secrets with correct format, naming, hex/base64 values, and tier prefix for cryptoutil services. Use when creating or updating deployment secrets to prevent HKDF derivation failures from wrong hex values or naming format violations."
argument-hint: "[PS-ID tier]"
disable-model-invocation: true
---

Generate Docker secrets with correct format, naming, hex/base64 values, and tier prefix.

## Purpose

Use when creating or updating Docker secrets for a cryptoutil service deployment.
Wrong hex values in unseal secrets break HKDF derivation silently. This skill
ensures correct format, naming, and unique random values per shard.

## Tier Prefixes

| Tier | Prefix Pattern | Example |
|------|---------------|---------|
| Service | `{PS-ID}-` | `sm-im-unseal-key-1-of-5-...` |
| Product | `{PRODUCT}-` | `sm-unseal-key-1-of-5-...` |
| Suite | `{SUITE}-` | `cryptoutil-unseal-key-1-of-5-...` |

## Secret File Catalog

| Filename | Value Format | Notes |
|----------|-------------|-------|
| `unseal-{N}of5.secret` | `{prefix}unseal-key-N-of-5-{hex-32-bytes}` | N=1-5, each shard UNIQUE random hex |
| `hash-pepper-v3.secret` | `{prefix}hash-pepper-v3-{base64url-32-bytes}` | URL-safe base64, no padding |
| `postgres-database.secret` | `{prefix_underscore}database` | Underscores, not hyphens |
| `postgres-username.secret` | `{prefix_underscore}database_user` | Underscores, not hyphens |
| `postgres-password.secret` | `{prefix_underscore}database_password` | Underscores, not hyphens |
| `postgres-url.secret` | `postgres://{user}:{pass}@{host}:5432/{db}?sslmode=disable` | Full DSN |
| `browser-username.secret` | Free-form credential | Service tier only |
| `browser-password.secret` | Free-form credential | Service tier only |
| `service-username.secret` | Free-form credential | Service tier only |
| `service-password.secret` | Free-form credential | Service tier only |

## Key Rules

- Filenames use hyphens (NEVER underscores) with `.secret` extension
- Each unseal shard (1-5) MUST have a unique random hex value — NEVER copy across shards
- NEVER copy hex values across different services or tiers
- Hex values: exactly 32 bytes = 64 hex characters (lowercase)
- Base64url values: exactly 32 bytes = 43 base64url characters (no padding `=`)
- PostgreSQL identifiers use underscores: `{ps_id}_database` (replace hyphens with underscores)
- Product/Suite tiers use `.secret.never` for browser/service credentials (marker files only)
- File permissions: 440 (r--r-----) on all `.secret` files
- NEVER store secrets in source code, environment variables, or config YAML

## Template: Generate All Secrets

For service `PS-ID` at service tier:

```bash
# Create secrets directory
mkdir -p deployments/PS-ID/secrets

# Generate 5 unique unseal shards (each with unique random hex)
for i in 1 2 3 4 5; do
  HEX=$(openssl rand -hex 32)
  echo -n "PS-ID-unseal-key-${i}-of-5-${HEX}" > deployments/PS-ID/secrets/unseal-${i}of5.secret
done

# Generate hash pepper (base64url, 32 bytes, no padding)
PEPPER=$(openssl rand -base64 32 | tr '+/' '-_' | tr -d '=')
echo -n "PS-ID-hash-pepper-v3-${PEPPER}" > deployments/PS-ID/secrets/hash-pepper-v3.secret

# PostgreSQL secrets (replace hyphens with underscores for identifiers)
PS_ID_UNDERSCORE=$(echo "PS-ID" | tr '-' '_')
echo -n "${PS_ID_UNDERSCORE}_database" > deployments/PS-ID/secrets/postgres-database.secret
echo -n "${PS_ID_UNDERSCORE}_database_user" > deployments/PS-ID/secrets/postgres-username.secret
PG_PASS=$(openssl rand -hex 16)
echo -n "${PS_ID_UNDERSCORE}_database_password_${PG_PASS}" > deployments/PS-ID/secrets/postgres-password.secret

# PostgreSQL URL
echo -n "postgres://${PS_ID_UNDERSCORE}_database_user:${PS_ID_UNDERSCORE}_database_password_${PG_PASS}@PS-ID-postgres:5432/${PS_ID_UNDERSCORE}_database?sslmode=disable" > deployments/PS-ID/secrets/postgres-url.secret

# Browser and service credentials
echo -n "browser-user" > deployments/PS-ID/secrets/browser-username.secret
echo -n "browser-pass" > deployments/PS-ID/secrets/browser-password.secret
echo -n "service-user" > deployments/PS-ID/secrets/service-username.secret
echo -n "service-pass" > deployments/PS-ID/secrets/service-password.secret
```

## Template: Product/Suite Tier

For product `PRODUCT` tier (use `.secret.never` for browser/service credentials):

```bash
mkdir -p deployments/PRODUCT/secrets

# Unseal shards (product-level prefix)
for i in 1 2 3 4 5; do
  HEX=$(openssl rand -hex 32)
  echo -n "PRODUCT-unseal-key-${i}-of-5-${HEX}" > deployments/PRODUCT/secrets/unseal-${i}of5.secret
done

# Hash pepper (product-level prefix)
PEPPER=$(openssl rand -base64 32 | tr '+/' '-_' | tr -d '=')
echo -n "PRODUCT-hash-pepper-v3-${PEPPER}" > deployments/PRODUCT/secrets/hash-pepper-v3.secret

# PostgreSQL secrets
echo -n "PRODUCT_database" > deployments/PRODUCT/secrets/postgres-database.secret
echo -n "PRODUCT_database_user" > deployments/PRODUCT/secrets/postgres-username.secret
PG_PASS=$(openssl rand -hex 16)
echo -n "PRODUCT_database_password_${PG_PASS}" > deployments/PRODUCT/secrets/postgres-password.secret

# Marker files for browser/service credentials (NEVER actual secrets)
touch deployments/PRODUCT/secrets/browser-username.secret.never
touch deployments/PRODUCT/secrets/browser-password.secret.never
touch deployments/PRODUCT/secrets/service-username.secret.never
touch deployments/PRODUCT/secrets/service-password.secret.never
```

## Validation Checklist

- [ ] All 5 unseal shards have UNIQUE hex values (diff each pair)
- [ ] Hex values are exactly 64 characters (32 bytes)
- [ ] Base64url pepper has no `+`, `/`, or `=` characters
- [ ] Tier prefix matches deployment level (service/product/suite)
- [ ] PostgreSQL identifiers use underscores (not hyphens)
- [ ] Product/Suite tiers use `.secret.never` for browser/service credentials
- [ ] No secrets committed to git (check `.gitignore` or Docker secrets path)
- [ ] `go run ./cmd/cicd-lint lint-deployments` passes after adding secrets

## References

Read [ARCHITECTURE.md Section 12.3.3 Secrets Coordination Strategy](../../../docs/ARCHITECTURE.md#1233-secrets-coordination-strategy) for deployment-level secrets management — follow the SUITE/PRODUCT/SERVICE coordination rules and secret file conventions.

Read [ARCHITECTURE.md Section 13.3 Secrets Management in Deployments](../../../docs/ARCHITECTURE.md#133-secrets-management-in-deployments) for Docker secrets enforcement — ensure all secret-bearing env vars use Docker secrets paths.

Read [ARCHITECTURE.md Section 6.4.2 Key Hierarchy (Barrier Service)](../../../docs/ARCHITECTURE.md#642-key-hierarchy-barrier-service) for unseal key requirements — understand why each shard must be unique and how HKDF derivation depends on correct hex values.
