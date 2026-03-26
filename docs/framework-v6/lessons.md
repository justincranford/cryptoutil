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

*(To be filled during Phase 10 execution)*

## Phase 11: Knowledge Propagation

*(To be filled during Phase 11 execution)*
