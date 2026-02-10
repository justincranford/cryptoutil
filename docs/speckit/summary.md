# Speckit Archive: Delta Summary

## Purpose

This document summarizes the deltas between `docs/speckit/constitution.md` (archived) and `docs/ARCHITECTURE.md` (authoritative source of truth).

**ARCHITECTURE.md is the single source of truth.** Constitution.md is retained only for historical reference.

## Key Deltas

| # | Topic | Constitution (Archived) | ARCHITECTURE.md (Authoritative) | Impact |
|---|-------|------------------------|--------------------------------|--------|
| 1 | Product count | "four Products" + "Demo: Cipher" | "five cryptographic-based products" (Cipher is a full product) | ARCHITECTURE.md is correct |
| 2 | Mutation testing | ≥85% Phase 4, ≥98% Phase 5+ | ≥95% mandatory minimum, ≥98% ideal | ARCHITECTURE.md is correct |
| 3 | Path structure | `internal/infra/*`, `internal/product/*` | `internal/shared/*`, `internal/apps/*` | ARCHITECTURE.md matches actual code |
| 4 | Speckit workflow | Section VI: 8-step speckit lifecycle (MANDATORY) | No speckit workflow (removed) | Speckit infrastructure removed |
| 5 | Private endpoint bind | "NEVER configurable, NEVER exposed" | Configurable per environment (127.0.0.1 default) | ARCHITECTURE.md is correct |
| 6 | Service federation | "Circuit breakers, fallback modes, retry" | "No circuit breakers, no retry logic" (multi-level failover) | ARCHITECTURE.md is correct |
| 7 | Config via env vars | "Support configuration via environment variables" | "NO environment variables for configuration" | ARCHITECTURE.md is correct |
| 8 | Token stop condition | "Token usage ≥ 990,000" | Not specified (beast-mode in instructions) | Instructions are authoritative |
| 9 | Service count | "9 total services: 8 product services + 1 demo" | "9 services across 5 products" (all equal) | ARCHITECTURE.md is correct |

## Recommendation

No content from constitution.md needs to be migrated to ARCHITECTURE.md. All constitution topics are already covered (and corrected) in ARCHITECTURE.md and `.github/instructions/*.instructions.md`.

See [details.md](details.md) for section-by-section comparison.