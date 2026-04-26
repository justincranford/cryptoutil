# Quiz Me — Framework V17 Open Questions

**Purpose**: Genuine architectural questions that need human answers before specific Phase 5 tasks
can proceed. All Phase 1-4 tasks can proceed without answers. Phase 5 tasks that depend on answers
are noted per question.

---

## Q1: Are product service dirs canonical or duplicates? (Blocks Task 5.7)

**Question**: For each of the 5 service-named subdirectories found inside product directories,
is the code there the **canonical implementation** (not yet moved to the PS-ID dir) or a
**duplicate** (PS-ID dir is already the canonical location)?

| Directory | PS-ID Counterpart | Canonical location? |
|-----------|------------------|---------------------|
| `internal/apps/sm/kms/` | `internal/apps/sm-kms/` | Unknown |
| `internal/apps/sm/im/` | `internal/apps/sm-im/` | Unknown |
| `internal/apps/jose/ja/` | `internal/apps/jose-ja/` | Unknown |
| `internal/apps/pki/ca/` | `internal/apps/pki-ca/` | Unknown |
| `internal/apps/skeleton/template/` | `internal/apps/skeleton-template/` | Unknown |

**Why it matters**: If the product dir contains the real implementation and the PS-ID dir is just
a wrapper/router, then Phase 5 Task 5.7 must MOVE code to the PS-ID dir (significant work, risk
of import breakage). If the product dir is already redundant (PS-ID dir is the real impl), then
Task 5.7 is just a delete (low risk).

**Answers**:
- `sm/kms/`: ___
- `sm/im/`: ___
- `jose/ja/`: ___
- `pki/ca/`: ___
- `skeleton/template/`: ___

---

## Q2: Do identity-rp and identity-spa have OpenAPI specs? (Blocks Task 5.5, 5.6)

**Question**: Do `identity-rp` (Relying Party) and `identity-spa` (Single Page App) expose any
HTTP API endpoints that should be documented in an OpenAPI spec? Or are these browser-only
services that serve static assets or redirect flows with no machine-readable API surface?

**Why it matters**: `swagger.go` is required in every PS-ID directory (per MANIFEST.yaml). If
these services do not have an OpenAPI spec, `swagger.go` should be a stub that serves a
placeholder spec or redirects to a combined spec. If they do have an API, a real spec is needed.

**Context**: All 8 other PS-IDs have `swagger.go`. The two missing ones are both browser-facing
services, which might suggest their API surface is through session cookies and redirects rather
than machine-readable endpoints.

**Answers**:
- `identity-rp`: _**(yes/no/partial) — notes:**_
- `identity-spa`: _**(yes/no/partial) — notes:**_

---

## Q3: Which identity services use the shared server builder pattern? (Informs Task 5.2-5.6)

**Question**: When adding `testmain_test.go` to `identity-authz`, `identity-idp`, `identity-rs`,
`identity-rp`, and `identity-spa`, which pattern should TestMain follow?

**Options**:
- A: Standard builder pattern (`testserver.StartAndWait` from `internal/apps-framework/service/testing/testserver/`)
- B: Custom per-service setup (some identity services may have unique initialization needs,
  e.g., identity-authz may need specific realm/tenant seeding)

**Why it matters**: If all 5 identity services use the standard builder, Task 5.2-5.6 can use
a single template for all. If each has custom setup, each requires individual investigation.

**Answers**:
- `identity-authz`: ___ (standard builder / custom)
- `identity-idp`: ___ (standard builder / custom)
- `identity-rs`: ___ (standard builder / custom)
- `identity-rp`: ___ (standard builder / custom)
- `identity-spa`: ___ (standard builder / custom)

---

## Resolution Notes

_Fill answers here before starting Phase 5. All answers will inform the implementation approach
in Tasks 5.1-5.7. Once answered, update the relevant tasks with concrete implementation details._
