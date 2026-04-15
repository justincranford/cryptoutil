# Quizme V2 - Framework V11: PKI-Init Cert Structure

**Created**: 2025-06-26
**Purpose**: Clarify open design questions discovered during docs/tls-structure.md analysis.

---

## Question 1: Client CA PKI Domain Inconsistency

**Question**: In the Required logical layout, line 4 uses `{sqlite,postgres}-{1,2}` (4 PKI domains for client CAs) but line 5 uses `{sqlite-1,sqlite-2,postgres}` (3 PKI domains for client leaves). The CA has per-instance CAs for all 4 instances, but the client leaves group postgres-1 and postgres-2 into a single `postgres` PKI domain. Is this intentional — that postgres app instances share a single client identity while sqlite instances have separate identities?

**A)** Yes, intentional. postgres-1 and postgres-2 share client identity because they connect to the same database; sqlite instances are isolated by design.
**B)** No, client leaves should also use 4 PKI domains (`sqlite-1`, `sqlite-2`, `postgres-1`, `postgres-2`). Update line 5 to `{sqlite,postgres}-{1,2}`.
**C)** The CAs should also use 3 PKI domains to match. Update line 4 to `{sqlite-1,sqlite-2,postgres}`.
**D)** Different grouping: use `{sqlite,postgres}` (2 PKI domains) for both CAs and leaves.
**E)**

**Answer**: A

**Rationale**: This determines the total directory count for categories 4 and 5, and whether postgres app instances can impersonate each other at the TLS layer.

---

## Question 2: Realm Values

**Question**: Category 5 uses `{realm}` as a parameter in client cert directory names (e.g., `public-{PS-ID}-browseruser-{realm}-https-client-*`). What are the concrete values of `{realm}`? The examples use `file` and `db` as placeholders.

**A)** `file`, `db` — matching the two realm types (FILE realm from config, DB realm from database).
**B)** `realm-file`, `realm-db` — prefixed to make the purpose clearer in directory names.
**C)** Dynamic — realm names come from the service configuration and can vary per PS-ID. pki-init reads a config file to determine realm names.
**D)** Fixed list from registry.yaml — add a `realms` field to the registry for each PS-ID.
**E)** E — add `realms` list to registry.yaml. At a high level, each PS-ID has list of realms that are described by 1) location (file, db, federated), 2) type (e.g. user/pass, mTLS, JWT, opaque access token, session cookie, etc), and 3) unique name. All types are implemented in framework and inherited by all PS-IDs, and individual PS-IDs select which ones they use. Any PS-ID can use any combination of realms they wish to use.

**Answer**: E

**Rationale**: This determines the directory naming pattern and whether pki-init needs external configuration input beyond tier-id and target-dir.

---

## Question 3: Admin Client Cert Purpose

**Question**: Category 9 uses `{PS-ID,admin}` to generate client certs for Grafana LGTM and OTel Collector. The `admin` entity appears alongside PS-IDs. What is the `admin` client cert used for?

**A)** A dedicated admin/ops user identity for directly accessing Grafana UI and OTel APIs (separate from any PS-ID service).
**B)** The pki-init process itself, which needs to authenticate to Grafana/OTel during cert provisioning.
**C)** A shared service account used by all PS-ID instances as a fallback when PS-ID-specific certs aren't available.
**D)** Reserved for future use — generate it now but usage will be defined in v12 or later.
**E)**

**Answer**: A

**Rationale**: This clarifies whether `admin` has a defined consumer or is a placeholder, and whether the cert should have different EKU/SAN constraints than PS-ID certs.

---

## Question 4: Cat 4 postgres-1/postgres-2 CAs vs Cat 5 "postgres" Shared Cert — Design Gap

**Question**: Category 4 generates SEPARATE client CA chains for `postgres-1` and `postgres-2` (`{sqlite,postgres}-{1,2}` = 4 per-instance PKI domains, 16 dirs). Category 5 groups postgres-1 and postgres-2 into a single `postgres` PKI domain and issues ONE shared client cert for both. This creates a structural gap: the shared Cat 5 cert must be signed by some issuing CA, but there is no `postgres` (aggregate) CA in Cat 4 — only `postgres-1` and `postgres-2` CAs. Which issuing CA signs the Cat 5 "postgres" shared cert?

**A)** `postgres-1`'s issuing CA signs the shared `postgres` cert. Both postgres-1 and postgres-2 app instances use this cert and must trust `postgres-1`'s CA chain for client verification. (Asymmetric trust: postgres-2 trusts postgres-1's CA.)
**B)** Change Cat 4 to 3 PKI domains (`sqlite-1`, `sqlite-2`, `postgres`) matching Cat 5. The shared `postgres` CA becomes the issuer for Cat 5 certs. Cat 4 drops from 16 to 12 dirs per PS-ID.
**C)** Both `postgres-1` and `postgres-2` issuing CAs independently sign copies of the shared `postgres` cert (two keystores, same CN/SAN identity, different cert chains). Receiving parties configure trust for both CA chains.
**D)** A new dedicated `postgres` (aggregate) CA chain is added as a 5th PKI domain alongside `sqlite-1`, `sqlite-2`, `postgres-1`, `postgres-2`. Cat 4 grows from 16 to 20 dirs per PS-ID.
**E)**

**Answer**:

**Rationale**: This determines whether the Cat 4/Cat 5 design is internally consistent. Option B would change the Cat 4 directory count from 16 to 12 per PS-ID (saving 4 dirs across all PS-ID scopes). Option D adds 4 dirs. This must be resolved before Phase 2 (Generator Rewrite) begins.

---

## Question 5: CA Key Algorithm and Default Key Size

**Question**: `02-05.security.instructions.md` MANDATES algorithm agility: "ALL crypto operations MUST support configurable algorithms with FIPS-approved defaults. Use config structs with Algorithm and KeySize fields." pki-init's CA key generation must therefore be configurable. What should the FIPS-approved DEFAULTS be?

**A)** ECDSA P-384 for all cert types (root CAs, issuing CAs, and leaf certs). Consistent algorithm across all tiers.
**B)** ECDSA P-384 for CAs (root + issuing), ECDSA P-256 for leaf certs. Tiered approach: CA longevity favors P-384; leaf certs are short-lived and P-256 is sufficient.
**C)** RSA-4096 for all cert types. Maximum compatibility with legacy clients/servers that may not support ECDSA.
**D)** Three-tier: ECDSA P-521 for root CAs, P-384 for issuing CAs, P-256 for leaf certs. Principle of least privilege — stronger algorithms protect longer-lived material.
**E)**

**Answer**:

**Rationale**: This determines the default cert generation in `pki-init`. Regardless of choice, the code MUST expose `Algorithm` / `KeySize` config per `02-05.security.instructions.md`. The answer here sets the default values only. FIPS 140-3 approves all options (RSA ≥2048, ECDSA P-256+).

---

## Question 6: Dev Cert Validity Periods

**Question**: pki-init generates dev/deployment/test certs. CA/Browser Forum Baseline Requirements cap public leaf cert validity at ≤398 days, intermediate at 5–10 years, root at 20–25 years — but those rules apply to publicly trusted CAs only. These are internal certs. What validity profile should pki-init use?

**A)** Follow CA/B Forum strictly: root 20yr, issuing 5yr, leaf 398 days. Encourages good hygiene even in dev environments; minimizes friction if certs ever tested against strict validators.
**B)** Pragmatic: root 10yr, issuing 5yr, leaf 3yr. Shorter than CA/B Forum maximums but avoids constant leaf rotation in long-lived dev environments.
**C)** Developer-friendly: root 10yr, issuing 5yr, leaf 10yr. Minimizes cert rotation friction in dev; acceptable for internal-only use.
**D)** Configurable via pki-init flags: `--root-validity=25y --issuing-validity=5y --leaf-validity=1y` as defaults (configurable). Gives operators full control.
**E)**

**Answer**:

**Rationale**: Regardless of choice, validity must be exposed as configurable per the algorithm agility mandate. The answer here sets defaults. Short leaf validity (option A/D defaults) catches expired-cert bugs early in dev; long validity (option C) reduces rotation toil.

---

## Question 7: ENG-HANDBOOK.md InsecureSkipVerify Contradiction (Pre-existing)

**Question**: There is a pre-existing contradiction in the ENG-HANDBOOK.md instruction files that must be resolved before framework-v11 Phase 6 (Knowledge Propagation):
- `02-05.security.instructions.md` states: "NEVER InsecureSkipVerify: true."
- `03-02.testing.instructions.md` (§10.3.4 Test HTTP Client Patterns) shows a code example with `InsecureSkipVerify: true` and comment `// test certs only`.

With framework-v11 generating a proper CA hierarchy, test code can use `RootCAs: testCAPool` instead. How should this contradiction be resolved?

**A)** Fix `03-02.testing.instructions.md` §10.3.4 example to use `RootCAs: testCAPool` with a `TLSRootCAPool()` helper loaded from the pki-init test CA. Remove `InsecureSkipVerify: true` entirely. The "test certs only" exception no longer applies when test CAs are available.
**B)** Add a narrow exception note in `02-05.security.instructions.md`: "Exception: test code MAY use `InsecureSkipVerify: true` ONLY in unit tests where no test CA pool is available."
**C)** Keep `03-02.testing.instructions.md` §10.3.4 as-is (it is a test example context, not production); add a clarifying note to `02-05.security.instructions.md` that the NEVER rule applies to production code only.
**D)** Treat as documentation debt; do not resolve in v11 scope. Track as a known issue and address in a future documentation phase when test CA infrastructure is mature.
**E)**

**Answer**:

**Rationale**: This is a documentation quality issue (not a code defect) that falls within v11 Phase 6 (Knowledge Propagation). Resolution requires editing two instruction files, running `go run ./cmd/cicd-lint lint-docs`, and verifying propagation. Option A is the most consistent with the security mandate but requires ensuring test infrastructure provides a `TLSRootCAPool()` helper before the example is updated.
