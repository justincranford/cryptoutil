# Identity V2 – Historical Baseline Assessment

Task 01 deliverable.

## Overview

This assessment reconciles the original identity plan (commits `1974b06` through `2514fef`) with the repository state from anchor commit `15cd829760f6bd6baf147cd953f8a7759e0800f4` through `HEAD` (current commit `c91278f`). The goal is to catalogue what exists, identify partial implementations, and surface actionable gaps for the Identity V2 remediation program.

## Methodology

- Reviewed the commit range `15cd829760f6bd6baf147cd953f8a7759e0800f4..HEAD` with emphasis on identity-related commits (`1974b06`, `0418528`, `d784ca6`, `74dcf14`, `5c04e44`, `dc68619`, `80d4e00`).
- Inspected code under `internal/identity/**`, CLI entries (`cmd/identity/**`), SPA assets, Docker orchestration, and existing tests.
- Compared the observed behaviour against expectations captured in `docs/identity/identity_master.md` and accompanying legacy task files.
- Verified infrastructure additions (mock service orchestration from `5c04e44`, documentation refreshes in `80d4e00`, `a6884d3`, `d91791b`).

## Timeline Highlights

| Seq | Commit | Date | Summary | Components | Notes |
| --- | --- | --- | --- | --- | --- |
| 1 | `1974b06` | 2025-11-06 | Task 2 – storage interfaces with GORM | Repository, domain | CRUD scaffolding exists but lacks end-to-end tests. |
| 2 | `0418528` | 2025-11-07 | Task 4 – OAuth 2.1 authz core | AuthZ server | Core handlers created with PKCE enforcement but contain TODOs for persistence and consent. |
| 3 | `d784ca6` | 2025-11-07 | Task 8 – HTTP servers & CLI (**partial**) | AuthZ/IDP/RS servers | Servers start but endpoints return placeholders; login/consent still stubbed. |
| 4 | `74dcf14` | 2025-11-07 | Task 10 – integration infra (**partial**) | Integration tests | Docker orchestration introduced, but integration tests remain TODO heavy. |
| 5 | `5c04e44` | 2025-11-09 | Mock services added to e2e lifecycle | Test infrastructure | Deterministic orchestration now spins up services, exposing behavioural gaps in identity flows. |
| 6 | `dc68619` | 2025-11-08 | SPA RP + mock service updates | SPA, AuthZ | SPA assets enriched; backend still returns placeholder responses. |
| 7 | `80d4e00` | 2025-11-09 | Documentation – 20 task blueprint | Docs | Highlights need for remediation; superseded by Identity V2 plan. |

## Component Comparison Matrices

### Authorization Server (AuthZ)

| Legacy expectation (commit) | Observed in `HEAD` | Status | Evidence |
| --- | --- | --- | --- |
| Authorization code flow with PKCE, consent, and code persistence (`0418528`) | Parameter validation enforces PKCE, but authorization request storage, consent UX, and code issuance remain TODOs. | Partial | `internal/identity/authz/handlers_authorize.go` (TODO block lines 62–88). |
| Token endpoint validates codes, generates access/refresh tokens (`0418528`) | Placeholder tokens minted without verifying authorization codes or PKCE verifiers; refresh grants rely on persisted token entities. | Partial | `internal/identity/authz/handlers_token.go` (TODOs lines 44–63). |
| Client authentication methods integrated (`ca597cd`, `35fde63`) | Basic and mTLS authenticators exist but are not wired into authorization_code flow yet. | Partial | `internal/identity/authz/client_authentication.go`, TODOs noted in authorization code path. |
| Introspection and revocation endpoints operational (`0418528`) | Implemented against token repository with correct RFC 7662/7009 semantics. | Working | `internal/identity/authz/handlers_introspect_revoke.go`. |

### Identity Provider (IdP)

| Legacy expectation (commit) | Observed in `HEAD` | Status | Evidence |
| --- | --- | --- | --- |
| Interactive login with session + MFA hand-off (`f181869`, `d850fad`) | Login handler validates inputs but returns placeholder JSON; password checks and session creation marked TODO. | Partial | `internal/identity/idp/handlers_login.go`. |
| Consent and profile management flows (`f181869`) | Consent endpoints not implemented; IdP service routes mostly stubs. | Missing | Absence of consent handlers; TODOs in `internal/identity/idp/service.go`. |
| User repository integration (`1974b06`) | Repository abstractions exist; no end-to-end test coverage confirming usage. | Partial | `internal/identity/repository/orm/user_repository.go`; TODOs in login handler. |

### Resource Server (RS)

| Legacy expectation (commit) | Observed in `HEAD` | Status | Evidence |
| --- | --- | --- | --- |
| Protected APIs enforce token validation and scopes (`d784ca6`) | Routes return static JSON with TODOs for token verification and scope checks. | Missing | `internal/identity/server/rs_server.go` (TODO block lines 27–33). |
| Telemetry and shutdown hygiene (`d784ca6`) | Server wiring uses Fiber defaults; telemetry hooks pending. | Partial | `internal/identity/server/rs_server.go`. |

### SPA Relying Party (RP)

| Legacy expectation (commit) | Observed in `HEAD` | Status | Evidence |
| --- | --- | --- | --- |
| Initiate OAuth 2.1 flow with PKCE (`e37c5cc`) | Front-end generates PKCE values, exchange tokens, introspects, and displays UI states. | Working | `cmd/identity/spa-rp/static/oauth.js`. |
| Integrate with backend consent/login screens (`dc68619`) | SPA expects functioning backend endpoints but currently receives placeholder JSON (authorization, login). | Blocked | Backend TODOs in AuthZ/IdP prevent full round-trip. |

## Architecture Snapshot (post `5c04e44`)

```mermaid
flowchart TB
    subgraph Client Tier
        SPA[SPA RP (cmd/identity/spa-rp) – UI ready]
    end
    subgraph Identity Services
        AuthZ[AuthZ Service – handlers with TODO gaps]
        IdP[IdP Service – login/consent stubs]
        RS[Resource Server – token validation TODO]
    end
    subgraph Shared Infrastructure
        DB[(Identity Database via GORM)]
        Tokens[Token Service]
        CLI[CLI & Workflow Orchestrators]
        E2E[E2E Orchestration (5c04e44)]
    end

    SPA -->|Authorization Code + PKCE| AuthZ
    AuthZ -->|User auth redirect| IdP
    IdP -->|Session + consent (planned)| AuthZ
    AuthZ -->|Access/Refresh Tokens| SPA
    SPA -->|Bearer Tokens| RS
    AuthZ & IdP & RS --> DB
    AuthZ --> Tokens
    CLI --> AuthZ & IdP & RS
    E2E --> AuthZ & IdP & RS
```

_Status legend_: “Ready” indicates UI/code exists; “TODO” indicates partial implementation or placeholders.

## Gap Summary

| Gap | Impact | Severity | Owner Task |
| --- | --- | --- | --- |
| Authorization code persistence, consent UX, PKCE verification not implemented. | Blocks compliant OAuth 2.1 flows and SPA integration. | Critical | Task 06 (AuthZ rehab), Task 07 (Client auth policy). |
| IdP login lacks password validation, session issuance, consent screens. | Prevents end-to-end user authentication. | Critical | Task 07, Task 09, Task 13. |
| Resource server skips token validation/scope enforcement. | Undermines API protection, fails spec requirements. | High | Task 10, Task 08. |
| Integration tests limited to scaffolding; critical flows not covered. | Regression risk remains high; gaps undetected. | High | Task 19. |
| Configuration divergence (compose vs CLI vs docs). | Leads to inconsistent startups; manual fixes required. | Medium | Task 03. |
| Documentation mismatches between legacy plan and actual code. | Causes planning confusion and duplicate efforts. | Medium | Task 02, Task 17. |

## Recommendations

1. Treat AuthZ + IdP TODOs as first-class remediation items—failing to implement persistence, consent, and session logic blocks nearly every downstream task.
2. Use the mock-enabled e2e harness (commit `5c04e44`) to capture current failure modes before refactoring; archive logs under `workflow-reports/e2e/` for regression comparison.
3. Establish requirement IDs (Task 02) referencing the gaps above so each remediation PR can link to measurable outcomes.
4. Prioritize configuration normalization (Task 03) before modifying orchestration to avoid multiplying drift.
5. Update legacy documentation references to point at the Identity V2 master plan and this baseline (Task 17 deliverable linkage).
