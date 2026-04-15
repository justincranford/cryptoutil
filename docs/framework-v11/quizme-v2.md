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

**Answer**:

**Rationale**: This determines the total directory count for categories 4 and 5, and whether postgres app instances can impersonate each other at the TLS layer.

---

## Question 2: Realm Values

**Question**: Category 5 uses `{realm}` as a parameter in client cert directory names (e.g., `public-{PS-ID}-browseruser-{realm}-https-client-*`). What are the concrete values of `{realm}`? The examples use `file` and `db` as placeholders.

**A)** `file`, `db` — matching the two realm types (FILE realm from config, DB realm from database).
**B)** `realm-file`, `realm-db` — prefixed to make the purpose clearer in directory names.
**C)** Dynamic — realm names come from the service configuration and can vary per PS-ID. pki-init reads a config file to determine realm names.
**D)** Fixed list from registry.yaml — add a `realms` field to the registry for each PS-ID.
**E)**

**Answer**:

**Rationale**: This determines the directory naming pattern and whether pki-init needs external configuration input beyond tier-id and target-dir.

---

## Question 3: Admin Client Cert Purpose

**Question**: Category 9 uses `{PS-ID,admin}` to generate client certs for Grafana LGTM and OTel Collector. The `admin` entity appears alongside PS-IDs. What is the `admin` client cert used for?

**A)** A dedicated admin/ops user identity for directly accessing Grafana UI and OTel APIs (separate from any PS-ID service).
**B)** The pki-init process itself, which needs to authenticate to Grafana/OTel during cert provisioning.
**C)** A shared service account used by all PS-ID instances as a fallback when PS-ID-specific certs aren't available.
**D)** Reserved for future use — generate it now but usage will be defined in v12 or later.
**E)**

**Answer**:

**Rationale**: This clarifies whether `admin` has a defined consumer or is a placeholder, and whether the cert should have different EKU/SAN constraints than PS-ID certs.
