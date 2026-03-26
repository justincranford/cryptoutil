# Lessons - Framework v6: Corrective Standardization

**Created**: 2026-03-25
**Last Updated**: 2026-03-26

## Phase 1: Fix target-structure.md Contradictions

*(To be filled during Phase 1 execution)*

## Phase 2: Create Missing .never Files

*(To be filled during Phase 2 execution)*

## Phase 3: Fix Service-Level Secret Values

*(To be filled during Phase 3 execution)*

## Phase 4: Fix Product-Level and Suite-Level Secret Values

*(To be filled during Phase 4 execution)*

## Phase 5: Restructure Config Directories — Flat Pattern

*(To be filled during Phase 5 execution)*

## Phase 6: Create Missing Config Files

*(To be filled during Phase 6 execution)*

## Phase 7: Clean Up Orphaned/Legacy Files

*(To be filled during Phase 7 execution)*

## Phase 8: Fitness Linter Verification

*(To be filled during Phase 8 execution)*

## Phase 9: Terminology Enforcement

### Findings

- **No violations found**: All config filenames, deployment files, and generated content use approved terminology (`authz`, `authn`, `authorization`, `authentication`).
- **Prior fix confirmed**: `adaptive-auth.yml` was already renamed to `adaptive-authorization.yml` in Phase 5 (Task 5.6).
- **Identity service**: Uses `identity-authz` PS-ID with approved `authz` abbreviation.

### Root Cause

- AI agents generating filenames or content may use the banned standalone `auth` abbreviation instead of `authn`/`authz`/`authentication`/`authorization`.
- Phase 5 already caught and fixed the only occurrence (`adaptive-auth.yml` → `adaptive-authorization.yml`).

### Prevention

- The terminology instruction file (`.github/instructions/01-01.terminology.instructions.md`) explicitly bans standalone `auth` and requires `authn`/`authz`.
- Future fitness linter could enforce filename scanning for banned terms (not implemented — deferring since manual audit found zero violations).
- All generated/renamed filenames should be checked against the banned terms list before commit.

## Phase 10: Migrate internal/apps/ to Flat PS-ID Structure

### Findings

- **Scope**: 10 service directories moved from `internal/apps/{PRODUCT}/{SERVICE}/` to `internal/apps/{PS-ID}/` using `git mv` to preserve history. 319 Go files had import paths updated.
- **Fitness linters updated**:
  - `service_structure`: Replaced `Product`+`Service` field pairs with single `PSID` field; `serviceDir = filepath.Join(appsDir, svc.PSID)`.
  - `cross_service_import_isolation`: Rewrote to use flat `serviceRef{psid, product}` and `collectServices()` scanning for `server/` subdir.
  - `check_skeleton_placeholders`: Added `"internal/apps/skeleton-template/"` to `excludedDirPrefixes`.
- **Incidental fix**: `identity-rs/server/public_server.go` had wrapcheck lint violations (`c.JSON()` / `c.Status().JSON()` return values not wrapped). Fixed by adding error wrapping with `fmt.Errorf`.
- **CRLF**: 4 identity service `server.go` files had CRLF line endings after the `git mv`. Fixed with PowerShell byte-level replacement before commit.
- **Pre-existing sysinfo flakiness**: Under `go test ./...` (full parallel load), several packages fail due to CPU contention causing sysinfo collection to timeout. All pass in isolation. NOT caused by Phase 10.

### Root Cause (CRLF)

The old `internal/apps/identity/{authz,idp,rp,rs}/server/server.go` files were written with CRLF. After `git mv` they carried those endings into the flat structure. Git warned `CRLF will be replaced by LF the next time Git touches it` but does not auto-fix until the next index write.

### Prevention

- Always check CRLF after `git mv` renames, especially for server.go files: `foreach ($f in $files) { $bytes = [System.IO.File]::ReadAllBytes($f); ... }`.
- Use the established PowerShell pattern for CRLF→LF conversion: `[System.IO.File]::WriteAllText($f, $content -replace \`"\`r\`n\`", \`"\`n\`", [System.Text.UTF8Encoding]::new($false))`.

### Key Decision

Flat `internal/apps/{PS-ID}/` aligns `internal/apps/`, `cmd/`, `deployments/`, and `configs/` all with the same PS-ID naming convention. This is the canonical project structure going forward.

## Phase 11: Knowledge Propagation

*(To be filled during Phase 11 execution)*
