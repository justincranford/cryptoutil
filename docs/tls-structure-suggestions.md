# tls-structure.md → ENG-HANDBOOK.md Suggestions

## Executive Summary

Analysis of [tls-structure.md](tls-structure.md) against [ENG-HANDBOOK.md §6.11.3](ENG-HANDBOOK.md#6113-pki-init-certificate-structure) reveals that the handbook covers the 14-category summary table and file format convention, but is missing several important operational and structural details. The following additions are suggested.

1. [pki-init CLI Interface](#1-pki-init-cli-interface) — two positional parameters, 16 valid domain values, and idempotency behavior are not documented in the handbook.
2. [Canonical App Instance Name Convention](#2-canonical-app-instance-name-convention) — the four `{PS-ID}-app-{variant}` suffixes and canonical tier IDs are not enumerated.
3. [Truststore Generation Rule](#3-truststore-generation-rule) — truststores are generated only for CA certificates; end-entity leaves never receive a truststore subdirectory. This rule is implied but not stated.
4. [Per-Category Directory Count Table with Formulas](#4-per-category-directory-count-table-with-formulas) — the handbook states totals (90/150/630) without per-category breakdowns or derivation formulas.
5. [Full Logical Layout Pattern](#5-full-logical-layout-pattern) — the path-pattern `text` block from tls-structure.md showing all 14 category path templates is absent from the handbook.
6. [Per-Category Policy Alignment](#6-per-category-policy-alignment) — the policy rationale for each category (which channel, which auth mechanism, which instances participate) is not in the handbook.
7. [Concrete Worked Examples](#7-concrete-worked-examples) — the full unrolled directory listings (skeleton-template: 90 dirs; sm product: 150 dirs) are exclusively in tls-structure.md.
8. [Realm-Count Parameterization](#8-realm-count-parameterization) — category 5 and 9 directory counts depend on the number of realms per PS-ID read from registry.yaml at runtime; the handbook states fixed counts without this dependency.

---

## Details

### 1. pki-init CLI Interface

**Current state in ENG-HANDBOOK.md**: §6.11.3 says "`pki-init` CLI generates the full `/certs` directory tree" but does not document the CLI signature.

**Suggested addition to §6.11.3**:

> **CLI Interface**: `pki-init <PKI-INIT-DOMAIN> <TARGET-DIRECTORY>`
>
> - `<PKI-INIT-DOMAIN>` — one of the 16 valid tier IDs: `cryptoutil` (suite), `sm`/`jose`/`pki`/`identity`/`skeleton` (products), or any of the 10 PS-IDs.
> - `<TARGET-DIRECTORY>` — root output directory (e.g., `/certs`, `/tmp`). All files are written under `<TARGET-DIRECTORY>/<PKI-INIT-DOMAIN>/`.
> - **Idempotency**: if `<TARGET-DIRECTORY>/<PKI-INIT-DOMAIN>/` exists and is non-empty, `pki-init` refuses to generate and exits with an error. If the subdirectory does not exist or is empty, it creates the full tree.

---

### 2. Canonical App Instance Name Convention

**Current state in ENG-HANDBOOK.md**: The four app instance suffixes appear implicitly in directory patterns (e.g., `{PS-ID}-{sqlite,postgres}-{1,2}`) but the canonical instance names and tier ID list are not stated as a reference table.

**Suggested addition to §6.11.3**:

Canonical naming:

| Type | Values |
|------|--------|
| App instance suffixes | `{PS-ID}-app-sqlite-1`, `{PS-ID}-app-sqlite-2`, `{PS-ID}-app-postgres-1`, `{PS-ID}-app-postgres-2` |
| Suite tier ID | `cryptoutil` |
| Product tier IDs | `sm`, `jose`, `pki`, `identity`, `skeleton` |
| PS-ID tier IDs | `sm-kms`, `sm-im`, `jose-ja`, `pki-ca`, `identity-authz`, `identity-idp`, `identity-rs`, `identity-rp`, `identity-spa`, `skeleton-template` |

---

### 3. Truststore Generation Rule

**Current state in ENG-HANDBOOK.md**: The truststore subdirectory is described but the rule restricting truststores to CA certificates only is not stated.

**Suggested addition to §6.11.3**:

> **Truststore rule**: `pki-init` generates a `truststore/` subdirectory **only for CA certificates** (root and issuing). End-entity (leaf) certificates never receive a `truststore/` subdirectory — trust is established via the issuing CA's truststore. `pki-init` never generates self-signed leaf certificates.

---

### 4. Per-Category Directory Count Table with Formulas

**Current state in ENG-HANDBOOK.md**: §6.11.3 provides total counts (90 per PS-ID, 630 suite) but no per-category breakdown or derivation formula.

**Suggested addition to §6.11.3** — expand the existing counts note with a table:

| Category | Description | Per PS-ID | Per PRODUCT (N PS-IDs) | Per SUITE (10 PS-IDs) |
|----------|-------------|-----------|------------------------|----------------------|
| 1 | Global HTTPS Server CAs | 4 | 4 | 4 |
| 2 | Grafana/OTel Server Certs | 2 | 2 | 2 |
| 3 | PS-ID App Server Certs | 4 | 4×N | 40 |
| 4 | PS-ID HTTPS Client CAs | 12 | 12×N | 120 |
| 5 | PS-ID HTTPS Client Certs | 2×\|realms\|×3 | 2×\|realms\|×3×N | 2×\|realms\|×3×10 |
| 6 | Private mTLS CAs (Admin) | 16 | 16×N | 160 |
| 7 | Private mTLS Leaves (Admin) | 4 | 4×N | 40 |
| 8 | Grafana/OTel Client CAs | 8 | 8 | 8 |
| 9 | Grafana/OTel Client Certs | 2×(4+2) | 2×(4N+2) | 2×(40+2) |
| 10 | PostgreSQL Server CAs | 4 | 4 | 4 |
| 11 | PostgreSQL Server Certs | 2 | 2 | 2 |
| 12 | PostgreSQL Client CAs | 4 | 4 | 4 |
| 13 | PostgreSQL Replication Certs | 2 | 2 | 2 |
| 14 | PS-ID PostgreSQL App Clients | 4 | 4×N | 40 |
| **Total** (2 realms) | | **90** | **varies** | **630** |

Categories 1, 2, 8, 10–13 are deployment-scoped constants. Categories 3–7, 9, 14 scale with PS-ID count. Category 5 additionally scales with realm count (read from registry.yaml at pki-init runtime).

---

### 5. Full Logical Layout Pattern

**Current state in ENG-HANDBOOK.md**: §6.11.3 has the 14-row summary table. The path-level structural pattern is not present.

**Suggested addition to §6.11.3** (or a new §6.11.6 "pki-init Logical Layout"):

```text
TARGET-DIRECTORY/{PKI-INIT-DOMAIN}/
  public-https-server-{root,issuing}-ca{/,/truststore/}SAME-AS-DIR-NAME.{p12,crt,key}
  public-https-server-entity-{grafana-otel-lgtm,otel-collector-contrib}/SAME-AS-DIR-NAME.{p12,crt,key}
  public-https-server-entity-{PS-ID}-{sqlite,postgres}-{1,2}/SAME-AS-DIR-NAME.{p12,crt,key}

  public-https-client-{root,issuing}-ca-{PS-ID}-{sqlite-1,sqlite-2,postgres}{/,/truststore/}SAME-AS-DIR-NAME.{p12,crt,key}
  public-https-client-entity-{PS-ID}-{sqlite-1,sqlite-2,postgres}-{browseruser,serviceuser}-{realm}/SAME-AS-DIR-NAME.{p12,crt,key}

  private-https-mutual-{root,issuing}-ca-{PS-ID}-{sqlite,postgres}-{1,2}{/,/truststore/}SAME-AS-DIR-NAME.{p12,crt,key}
  private-https-mutual-entity-{PS-ID}-{sqlite,postgres}-{1,2}/SAME-AS-DIR-NAME.{p12,crt,key}

  {grafana-otel-lgtm,otel-collector-contrib}-https-client-{root,issuing}-ca{/,/truststore/}SAME-AS-DIR-NAME.{p12,crt,key}
  {grafana-otel-lgtm,otel-collector-contrib}-https-client-entity-{PS-ID}-{sqlite,postgres}-{1,2}/SAME-AS-DIR-NAME.{p12,crt,key}
  {grafana-otel-lgtm,otel-collector-contrib}-https-client-entity-{admin,infra}/SAME-AS-DIR-NAME.{p12,crt,key}

  postgres-tls-server-{root,issuing}-ca{/,/truststore/}SAME-AS-DIR-NAME.{p12,crt,key}
  postgres-tls-server-entity-{leader,follower}/SAME-AS-DIR-NAME.{p12,crt,key}

  postgres-tls-client-{root,issuing}-ca{/,/truststore/}SAME-AS-DIR-NAME.{p12,crt,key}
  postgres-tls-client-entity-{leader,follower}-replication/SAME-AS-DIR-NAME.{p12,crt,key}
  postgres-tls-client-entity-{leader,follower}-{PS-ID}-postgres-{1,2}/SAME-AS-DIR-NAME.{p12,crt,key}
```

---

### 6. Per-Category Policy Alignment

**Current state in ENG-HANDBOOK.md**: §6.11.4 covers PostgreSQL mTLS staging and §6.11.5 covers admin mTLS. There is no summary of which policy applies to each of the 14 categories.

**Suggested addition to §6.11.3** or §6.11.6:

- **Private admin channel** (categories 6–7): Mutual TLS required. Both server and client auth use the same combined leaf cert per instance. Each `{sqlite,postgres}-{1,2}` instance has its own dedicated CA chain.
- **Public HTTPS server** (categories 1–3): Server cert issued by the single global server CA. Client TLS authentication handled separately via categories 4–5.
- **Public HTTPS client** (categories 4–5): Client certificates per PKI domain, per API path prefix (`/browser/` → `browseruser`, `/service/` → `serviceuser`), and per realm type. `postgres-1` and `postgres-2` share one PKI domain (`postgres`) — both instances use certs from the same CA, reducing CA count.
- **PostgreSQL connections** (categories 10–14): Mutual TLS and username+password required. Only postgres-1 and postgres-2 app instances connect to PostgreSQL; sqlite-1 and sqlite-2 do not (no Cat 14 certs for SQLite instances).
- **OTel Collector OTLP** (categories 8–9): Server cert for `:4317`/`:4318`; client certs per PS-ID instance plus one `admin` cert for operators and one `infra` cert for OTel→Grafana forwarding.

---

### 7. Concrete Worked Examples

**Current state in ENG-HANDBOOK.md**: Totals are stated; no worked examples are present. The full unrolled listings exist only in tls-structure.md.

**Suggested addition**: Add cross-references in §6.11.3:

> **Worked examples**: See [docs/tls-structure.md — Example: skeleton-template](tls-structure.md#example-skeleton-template-ps-id) for the full 90-directory listing at PS-ID scope, and [Example: sm](tls-structure.md#example-sm-product) for the 150-directory listing at PRODUCT scope (2 PS-IDs, 2 realms assumed).

---

### 8. Realm-Count Parameterization

**Current state in ENG-HANDBOOK.md**: Directory counts for category 5 and category 9 are stated as fixed ("12 directories") without noting they depend on the number of realms defined per PS-ID.

**Suggested addition to §6.11.3** (in the directory counts note):

> **Realm-dependent categories**: Categories 5 and 9 directory counts depend on `|realms|` — the number of realm types defined for a PS-ID in `registry.yaml` (read at `pki-init` runtime). The examples assume 2 realms (`file`, `db`). If a PS-ID has 3 realms, category 5 generates `2 × 3 × 3 = 18` directories per PS-ID instead of 12. Always derive counts from `2 × |realms| × 3` (category 5) rather than using the example totals directly.
