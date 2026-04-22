# tls-structure.md → ENG-HANDBOOK.md Suggestions

## Executive Summary

Analysis of [tls-structure.md](tls-structure.md) and [framework-v15/pki-init-order.md](framework-v16/pki-init-order.md) against [ENG-HANDBOOK.md §6.5](ENG-HANDBOOK.md#65-pki-architecture--strategy) reveals several documentation gaps. The handbook covers PKI architecture, TLS configuration, and CA hierarchy at a policy level, but the implementation-specific details in `tls-structure.md` have gaps that affect developers implementing new services and cert categories.

1. [Admin CA Bundle (`issuing-ca.pem`) Undocumented](#1-admin-ca-bundle-issuing-capem-undocumented) — the CA chain file distributed to admin mTLS clients is absent from tls-structure.md.
2. [`tls-config.yml` Dynamic Cert Pattern Undocumented](#2-tls-configyml-dynamic-cert-pattern-undocumented) — the `TLSModeMixed` mechanism for selecting pre-generated vs dynamically generated certs is not described.
3. [Realm Dynamic Binding Undocumented (Decision 8)](#3-realm-dynamic-binding-undocumented-decision-8) — Decision 8 is referenced but the actual binding mechanism is unexplained.
4. [`postgres` vs `postgres-1`/`postgres-2` Naming Ambiguity](#4-postgres-vs-postgres-1postgres-2-naming-ambiguity) — cert categories use inconsistent PostgreSQL naming conventions without explanation.
5. [Directory Count Formula Derivation Missing](#5-directory-count-formula-derivation-missing) — the 90/630 counts are stated without showing the derivation formula for each category.
6. [V12/V13 Dependency Graph Missing from `pki-init-order.md`](#6-v12v13-dependency-graph-missing-from-pki-init-ordermd) — a visual or structured dependency graph for phase ordering is absent.
7. [Cross-Plan Dependency Explanation Needs Clarification](#7-cross-plan-dependency-explanation-needs-clarification) — the V12→V13 parallel execution claim is accurate but confusing without a clear prerequisite chain.

---

## Details

### 1. Admin CA Bundle (`issuing-ca.pem`) Undocumented

**Current state in `tls-structure.md`**: Cat 6 (Private mTLS CAs) documents the `private-https-mutual-{root,issuing}-ca-{PS-ID}-{sqlite,postgres}-{1,2}/` structure with keystore and truststore subdirectories. However, the truststore file that an admin mTLS client must present to verify the admin endpoint is not called out explicitly. When a service's admin port (`127.0.0.1:9090`) requires mutual TLS, the connecting client (e.g., the `livez` CLI command) must trust the server's CA. The relevant file is `private-https-mutual-issuing-ca-{PS-ID}-{instance}/truststore/private-https-mutual-issuing-ca-{PS-ID}-{instance}.crt`.

**Gap**: `tls-structure.md` does not explain that:
1. The `truststore/` file within Cat 6 is what the admin mTLS client must include in its trusted CA bundle.
2. The `livez` CLI subcommand must be configured with the `--ca-cert` flag pointing to the issuing CA's truststore cert to verify the admin endpoint.
3. When `HEALTHCHECK` in Docker uses `livez`, it must mount the CA bundle at container build time.

**Suggested addition** to the "Policy Alignment" section of `tls-structure.md`:

> **Admin mTLS client trust**: When the admin port (`127.0.0.1:9090`) requires mTLS, any client connecting to it (including the `livez` CLI healthcheck and Docker `HEALTHCHECK` command) MUST:
> 1. Present a client cert issued by the relevant `private-https-mutual-issuing-ca-{PS-ID}-{instance}` CA
> 2. Trust the server cert via the `private-https-mutual-issuing-ca-{PS-ID}-{instance}/truststore/` CA bundle
>
> The `livez` CLI uses `--cert`, `--key`, and `--ca-cert` flags for this. A passing Docker `HEALTHCHECK` using `livez` is the canonical proof of admin mTLS end-to-end connectivity (per ENG-HANDBOOK.md §5.5).

---

### 2. `tls-config.yml` Dynamic Cert Pattern Undocumented

**Current state in `tls-structure.md`**: The document describes the expected cert directory layout but does not explain how a service reads that layout at runtime. In particular, the `TLSModeMixed` mode — where a service uses `tls-config.yml` to determine which cert directories to load from the `/certs` volume vs. which to generate dynamically — is not documented anywhere.

**Gap**: Developers implementing a new service do not know:
1. What `tls-config.yml` contains and how it is structured.
2. When to use `TLSModeMixed` vs. `TLSModePreGenerated` vs. `TLSModeAutoGenerate`.
3. How the service reads the cert file paths from the layout (SAME-AS-DIR-NAME convention).

**Suggested addition** as a new section "Runtime TLS Configuration" in `tls-structure.md`:

> **`tls-config.yml` and TLS Mode Selection**
>
> Each PS-ID app instance reads a `tls-config.yml` file at startup that selects one of three TLS modes:
>
> | Mode | Description | When to Use |
> |------|-------------|-------------|
> | `TLSModeAutoGenerate` | Service generates all TLS certs at startup (default for unit/integration tests) | Development / tests without pki-init volume |
> | `TLSModePreGenerated` | Service reads all certs from `/certs` volume; no generation | Production with pki-init volume |
> | `TLSModeMixed` | Service reads pre-generated certs for some roles, generates dynamically for others | Hybrid environments where some cert types are not yet pre-generated |
>
> The `tls-config.yml` also specifies which cert directory paths (from the SAME-AS-DIR-NAME convention) map to which TLS role (public server, private server, public client, private client).
>
> **SAME-AS-DIR-NAME convention**: Within each cert directory, files use the directory name as their basename. For example:
> ```
> /certs/sm-kms/public-https-server-entity-sm-kms-sqlite-1/
>   public-https-server-entity-sm-kms-sqlite-1.crt   ← PEM cert
>   public-https-server-entity-sm-kms-sqlite-1.key   ← PEM private key
>   public-https-server-entity-sm-kms-sqlite-1.p12   ← PKCS#12 bundle
> ```

---

### 3. Realm Dynamic Binding Undocumented (Decision 8)

**Current state in `tls-structure.md`**: Cat 5 (PS-ID HTTPS Client Certs) states: "Realm values are dynamic per PS-ID, read from `registry.yaml` at pki-init runtime (Decision 8)." This refers to an unexplained design decision about how client cert realm names are derived.

**Gap**: Developers implementing a new PS-ID or adding a realm do not know:
1. Where in `registry.yaml` the realm names are defined.
2. How `pki-init` reads them at runtime.
3. What happens if `registry.yaml` changes (new realm added — does pki-init re-run regenerate all certs?).
4. Why the realm name appears in the cert directory path vs. in the cert's Subject Alternative Name or CN.

**Suggested addition** to Cat 5 description in `tls-structure.md`:

> **Decision 8: Realm Dynamic Binding**
>
> `pki-init` reads the PS-ID's realm list from `api/cryptosuite-registry/registry.yaml` at runtime. The realm names become suffixes in the Cat 5 client cert directory names. Realm lists are PS-ID specific:
>
> | PS-ID | Default Realms |
> |-------|----------------|
> | `sm-kms` | `file`, `db` |
> | `sm-im` | `file`, `db` |
> | (all others) | `file`, `db` |
>
> **When a realm is added**: pki-init must be re-run for the affected PS-ID domain to generate client certs for the new realm. Existing certs are not regenerated (pki-init refuses to overwrite a non-empty output directory).
>
> **Why realm in directory name (not SAN/CN)**: The directory name is the GORM-readable identifier for cert selection. The realm is embedded in the directory path rather than the cert's Subject so that the framework can select the correct cert without parsing X.509 extensions at runtime.

---

### 4. `postgres` vs `postgres-1`/`postgres-2` Naming Ambiguity

**Current state in `tls-structure.md`**: Certificate category descriptions and the logical layout pattern use inconsistent PostgreSQL naming:

| Category | Naming Used | Meaning |
|----------|-------------|---------|
| Cat 4 (PS-ID HTTPS Client CAs) | `postgres` (shared domain) | One PKI domain covering both postgres-1 and postgres-2 app instances |
| Cat 5 (PS-ID HTTPS Client Certs) | `postgres` (shared domain) | Client certs for both postgres app instances share one PKI domain |
| Cat 6 (Private mTLS CAs) | `postgres-1`, `postgres-2` (individual) | Separate CA chains per postgres instance |
| Cat 7 (Private mTLS Leaves) | `postgres-1`, `postgres-2` (individual) | Separate leaf certs per postgres instance |
| Cat 14 (PS-ID PostgreSQL App Clients) | `postgres-1`, `postgres-2` (individual) | Separate PG client certs per postgres instance |

**Gap**: The distinction between "shared postgres PKI domain" (Cat 4-5) and "individual postgres-1/postgres-2" (Cat 6-7, 14) is never explained. Developers implementing new cert categories do not know which naming convention to use or why.

**Suggested addition** as a new sub-section "PostgreSQL Naming Conventions" in `tls-structure.md`:

> **PostgreSQL Naming Conventions: Shared Domain vs. Individual Instances**
>
> Two distinct naming patterns appear in cert directories:
>
> | Pattern | Used By | Rationale |
> |---------|---------|-----------|
> | `postgres` (singular, shared domain) | Cat 4 (Public HTTPS Client CAs), Cat 5 (Public HTTPS Client Certs) | A PS-ID app's public HTTPS client cert is the SAME regardless of which postgres instance it connects to. The client identity is the app instance (`{PS-ID}-{postgres-1}` or `{PS-ID}-{postgres-2}`), not the target DB. |
> | `postgres-1`, `postgres-2` (individual) | Cat 6 (Private mTLS CAs), Cat 7 (Private mTLS Leaves), Cat 14 (PG App Client Certs) | Admin mTLS and PostgreSQL mTLS are per-instance: each postgres container has its own server cert, each app instance has a unique identity cert presented to that specific postgres instance. |
>
> **Decision rationale**: Public HTTPS client certs authenticate the app instance to an upstream service (e.g., another PS-ID's `/service/**` endpoint). The upstream doesn't distinguish which DB the app is connected to, so a single PKI domain `postgres` covers both instances. PostgreSQL mTLS certs, however, are presented directly to the DB server — each postgres container validates the client cert against its own CA (`ssl_ca_file`), requiring per-instance identities.

---

### 5. Directory Count Formula Derivation Missing

**Current state in `tls-structure.md`**: The Directory Count Summary table states 90 dirs per PS-ID and 630 dirs per SUITE, with per-category counts. However, the 630 total does not show the derivation formula. Per ENG-HANDBOOK.md §14.1.2:

> "Derive directory/file counts from pattern expansion (MANDATORY): Always show the derivation formula rather than a raw count. Example: `30 global + 60 per-PS-ID × 10 = 630`. Raw counts without formulas are unverifiable during review."

**Gap**: The category-level counts are shown but the aggregation formula is missing. The 90 per-PS-ID and 630 per-SUITE counts should be derived explicitly.

**Suggested addition** at the end of the Directory Count Summary table:

```
Per-PS-ID total breakdown:
  Global (same across all tier levels):  4 + 2 + 8 + 4 + 2 + 4 + 2 = 26 dirs
  Per-PS-ID (scaled by N PS-IDs):        4 + 12 + 12 + 16 + 4 + 12 + 4 = 64 dirs
  Total per PS-ID: 26 global + 64 per-PS-ID = 90 dirs

Per-SUITE total breakdown:
  Global dirs (appear once regardless of PS-ID count): 4 + 2 + 8 + 4 + 2 + 4 + 2 = 26 dirs
  Scaled dirs (× 10 PS-IDs):                           4 + 12 + 12 + 16 + 4 + 12 + 4 = 64 × 10 = 640 dirs (*)

  (*) Cat 9 Grafana/OTel Client Certs breaks the pattern:
      Per-PS-ID contribution: 4 app instances × 2 services = 8 dirs
      Global contribution:    admin + infra × 2 services = 4 dirs
      Total Cat 9: (8 × 10 PS-IDs) + 4 global = 84 dirs (not 120 like a pure per-PS-ID count)

  Corrected SUITE total:
    Cat 1:  4  (global)
    Cat 2:  2  (global)
    Cat 3:  40 (4 × 10)
    Cat 4:  120 (12 × 10)
    Cat 5:  120 (12 × 10, assuming 2 realms)
    Cat 6:  160 (16 × 10)
    Cat 7:  40 (4 × 10)
    Cat 8:  8  (global)
    Cat 9:  84 (8×10 per-PS-ID + 4 global admin/infra)
    Cat 10: 4  (global)
    Cat 11: 2  (global)
    Cat 12: 4  (global)
    Cat 13: 2  (global)
    Cat 14: 40 (4 × 10)
    ────────────────────
    Total: 630
```

**Note**: The "90 per PS-ID" figure includes only the per-PS-ID scope. It excludes Cat 9's global `admin`/`infra` dirs since those are not scoped to a specific PS-ID.

---

### 6. V12/V13 Dependency Graph Missing from `pki-init-order.md`

**Current state in `pki-init-order.md`**: The V12 and V13 phase sequences are described in prose and as ordered lists, but there is no visual or structured dependency graph showing which phases can run in parallel vs. which have strict serial dependencies.

**Gap**: Without a dependency graph, a developer cannot determine:
1. Which V12 phases must complete before any V13 phase can start.
2. Whether V13 Phase 0 truly can be parallelized with V12 Phases 1-9 (and which part of V12 Phase 0 is the prerequisite).
3. Which phases within V12 depend on each other vs. which are independent.

**Suggested addition** as a "Dependency Graph" section in `pki-init-order.md`:

```
V12/V13 Phase Dependency Graph
═══════════════════════════════════════════════════════════════════

SERIAL START → V12 Phase 0 (pki-init: Cat 9 infra + Cat 14)
                     │
          ┌──────────┴────────────────────────────────┐
          │ (V12 Phases 1-9 continue)                  │ (V13 Phase 0 begins in parallel)
          ▼                                            ▼
 V12 Phase 1: PG Server TLS                  V13 Phase 0: pki-init
      │                                       (Cat 2,3,4,8,9 app)
      ▼                                            │
 V12 Phase 2: PG Replication Server TLS           │
      │                                            │
      ▼                                            │
 V12 Phase 3: Verify PG Standalone                │
      │                                            │
      ▼                                            │
 V12 Phase 4: PG Client mTLS                      │
      │                                            │
      ▼                                            │
 V12 Phase 5: PG Replication mTLS                 │
      │                                            │
      ▼                                            │
 V12 Phase 6: Verify PG Full Stack                │
      │                                            │
      ▼                                            │
 V12 Phases 7-9: Templates + Lint + Verify        │
      │                                            │
      ▼                                            │
 V12 Phase 10: Private Admin mTLS Trust            │
      │                                            │
      ▼                                            │
 V12 Phase 11: Knowledge Propagation              │
      │                                            │
      └──────────────────┬────────────────────────┘
                         │ (V12 complete + V13 Phase 0 complete)
                         ▼
                V13 Phase 1: OTel Collector Server TLS
                         │
                         ▼
                V13 Phases 2-11: (serial, per V13 rationale)
```

**Key parallel window**: V12 Phase 0 must complete first (prerequisite: Cat 9 infra cert needed by V13). After V12 Phase 0, V13 Phase 0 (code changes to pki-init generator for Cat 2,3,4,8,9 app) can begin while V12 Phases 1-9 are executing. V13 Phases 1+ cannot start until BOTH V12 (including Cat 9 infra from Phase 0) is complete AND V13 Phase 0 (new cert generation code) is complete.

---

### 7. Cross-Plan Dependency Explanation Needs Clarification

**Current state in `pki-init-order.md`**: The "Cross-Plan Dependency" section states:

> "Execution order: V12 → V13 (V13's Phase 0 can begin in parallel with V12 Phases 1-9, but V13 Phases 1+ require V12 Phase 0 complete for the Cat 9 infra cert used in OTel→Grafana)."

This is accurate but creates a potential misreading: if V12 Phase 0 is a prerequisite of V13 Phase 1+, and V13 Phase 0 "begins in parallel with V12 Phases 1-9", then V12 Phase 0 must already be complete before V13 Phase 0 starts — making the overall prerequisite chain:

```
V12 Phase 0 → {V12 Phases 1-9 ‖ V13 Phase 0} → V13 Phases 1+
```

**Gap**: The current text does not make the implicit prerequisite explicit: V12 Phase 0 must complete before EITHER of the two parallel workstreams (V12 Phases 1-9 and V13 Phase 0) can begin.

**Suggested replacement** for the Cross-Plan Dependency section:

> **Cross-Plan Dependency**
>
> V12 and V13 share the following prerequisite chain:
>
> 1. **V12 Phase 0 must complete first** (generates Cat 9 infra cert needed by V13 Phase 5+ and Cat 14 cert needed by V12 Phase 4).
> 2. **After V12 Phase 0**, two workstreams can run in parallel if two developers or parallel tracks are available:
>    - **Workstream A**: V12 Phases 1-11 (PostgreSQL mTLS + admin mTLS)
>    - **Workstream B**: V13 Phase 0 (pki-init code changes for Cat 2, 3, 4, 8, 9 app certs)
> 3. **V13 Phases 1+ require**:
>    - V12 fully complete (Cat 9 infra cert exists in the `/certs` volume)
>    - V13 Phase 0 complete (pki-init generates all V13 cert categories)
>
> When working sequentially (single developer), the order is: V12 Phases 0-11 → V13 Phases 0-11.
