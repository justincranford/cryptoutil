# Quiz Me - Framework V21: Open Architectural Decisions

**Created**: 2026-04-30
**Purpose**: Surface genuine unknowns and design decisions for the user before implementation begins.
These questions require human judgment — they are NOT LLM discovery tasks.

---

## Q1: sm-kms businesslogic/ split strategy

**Context**: `internal/apps/sm-kms/server/businesslogic/` contains three categories of code:
- ORM entity mapping (oam_orm_mapper*.go) — should move to `server/model/`
- OAS handler delegation (oas_handlers.go, oam_oas_mapper*.go in handler/) — should move to `server/apis/`
- Pure business logic (businesslogic.go, businesslogic_crypto.go, elastic_key_status_state_machine.go)
  — no canonical home defined by the linter; can stay in `server/businesslogic/` or move to a new
  `server/service/` layer

**Question**: Should the pure business logic in `businesslogic.go` and `businesslogic_crypto.go` be:

a) Left in `server/businesslogic/` as-is (the linter does not require this dir to be removed, only
   that `apis/`, `model/`, and `repository/` exist alongside it), OR

b) Moved to a new `server/service/` subdirectory to complete the layered architecture, OR

c) Another approach?

**Implication**: Option (a) is minimal-change and lower risk. Option (b) is more architecturally
complete but requires additional import-path updates and may reveal additional test dependencies.

---

## Q2: pki-ca APIs layer design

**Context**: pki-ca has a rich domain layer structure:
- `bootstrap/`, `compliance/`, `intermediate/`, `issuer/`, `profile/`, `service/`, `storage/`, etc.
- The existing `server/public_server.go` registers Fiber routes directly using these domain packages.
- The linter requires `server/apis/` to exist but the MANIFEST does not specify what files must be in it.

**Question**: For the pki-ca `server/apis/` directory, should V21:

a) Create a thin `server/apis/` package with a single skeleton file (e.g., `apis.go` with
   `package apis` and a placeholder comment), satisfying the linter structural check without
   moving or wrapping any existing code, OR

b) Create genuine HTTP handler wrappers in `server/apis/` that wrap the existing domain layers,
   following the identity service handler pattern, OR

c) Another approach?

**Implication**: Option (a) passes the structural linter with minimal risk. Option (b) may require
significant refactoring of pki-ca's server/public_server.go route registration logic. The scope
of V21 was intended to be structural (make the dirs exist) not behavioral (rewrite handlers).

---

## Q3: pki-ca repository/ and SQL migrations scope

**Context**: pki-ca uses GORM with domain-specific SQL migrations. The linter requires:
- `server/repository/migrations.go` with `//go:embed migrations` FS
- `server/repository/migrations/` directory

The existing pki-ca SQL migrations likely already exist somewhere. The question is WHERE.

**Question**: Does pki-ca currently have SQL migration files, and if so, where?
Should V21:

a) Move existing pki-ca SQL migrations into `server/repository/migrations/`, OR

b) Create empty `server/repository/migrations/` (for structural compliance) and defer migration
   file reorganization to a later plan, OR

c) Another approach?

**Implication**: Moving existing migration files may affect the golang-migrate embedded FS paths
referenced in production code. A wrong move breaks database schema on startup.

---

## Q4: Scope of lifecycle_test.go and port_conflict_test.go cleanup

**Context**: The `knownServerFileExclusions` map still has all-10-PS-ID exclusions for:
- `__SERVICE___lifecycle_test.go`
- `__SERVICE___port_conflict_test.go`

Current location survey (2026-04-30):
- sm-kms: both at root AND in server/ (root versions appear to be OLD duplicates)
- sm-im: lifecycle + port_conflict only in server/
- jose-ja: both at root (no server/ copies)
- pki-ca: both at root (no server/ copies)
- identity-*: all in server/ only
- skeleton-template: both at root (no server/ copies)

**Question**: Should V21 also:

a) Move jose-ja, pki-ca, skeleton-template lifecycle/port-conflict tests into server/ and clean up
   the sm-kms root duplicates, then narrow the exclusion maps to only the remaining stragglers, OR

b) Leave this for a separate plan — the linter exclusion is already in place and causes no failures,
   so this is cosmetic cleanup that can be deferred, OR

c) Another approach?

**Implication**: Option (a) expands V21 scope. Option (b) keeps V21 focused on the mandatory
structural migration (apis/, model/, repository/) and explicit tooling items.
