# Architecture Decision Quiz - fixes-v8

**Purpose**: Capture user decisions on key architecture questions before Phase 3 execution.
**Lifecycle**: Ephemeral - delete after decisions recorded in plan.md.

---

## Q1: Next Service Priority

Which service area should receive the most attention in fixes-v8?

- A) **Identity services alignment** - Migrate identity-authz/idp/rs/rp/spa to full template compliance (migration renumbering, E2E decomposition, rp/spa buildout)
- B) **PKI-CA hardening** - Deepen pki-ca compliance testing, certificate lifecycle E2E, CA/Browser Forum validation
- C) **SM/JOSE polish** - Ensure sm-kms, sm-im, jose-ja are perfectly aligned before moving to new services
- D) **Template evolution** - Evolve the service template itself (new builder features, middleware, session patterns)
- E) _(write your answer)_

**Answer**:

---

## Q2: Identity Architecture Direction

The 5 identity services share a common domain model and repository layer. Should they:

- A) **Stay monolithic** - Keep shared domain/repo, align ServerManager with template lifecycle, single E2E suite
- B) **Gradual separation** - Start splitting into per-service domain/repo, migrate one at a time (recommended: authz first)
- C) **Full separation** - Split all 5 into fully standalone services like sm-kms (high effort, high alignment)
- D) **Hybrid** - Keep shared domain but give each service its own migration range and E2E tests
- E) _(write your answer)_

**Answer**:

---

## Q3: Identity Migration Numbering

Identity migrations use 0002-0011 (predates template spec). The template reserves 1001-1999 for shared migrations and mandates 2001+ for domain. Options:

- A) **Renumber now** - Rename 0002-0011 → 2001-2010 across all environments (clean break, risk to existing deployments)
- B) **Keep dual ranges** - Leave existing as 0002-0011, new migrations start at 2001+ (pragmatic, some confusion)
- C) **Template range expansion** - Reserve 0001-0999 as legacy, 1001-1999 template, 2001+ domain (documentation fix only)
- D) **Defer** - Address when identity services hit prod deployment (lower risk, defers debt)
- E) _(write your answer)_

**Answer**:

---

## Q4: identity-rp and identity-spa Scope

These services are minimal (~10 files, ~4 tests each). What is the near-term goal?

- A) **Full buildout** - Match identity-authz/idp maturity level (~100+ files, full test coverage, all endpoints)
- B) **Functional MVP** - Implement core happy-path flows (OAuth callback, SPA token handling) with test coverage
- C) **Placeholder maintenance** - Keep current skeleton, focus effort on authz/idp/rs
- D) **Remove entirely** - These services aren't needed yet, remove code to reduce maintenance burden
- E) _(write your answer)_

**Answer**:

---

## Q5: E2E Test Strategy for Identity

Currently identity has a shared E2E test suite. Should it:

- A) **Stay shared** - Single E2E suite tests all 5 services together (simpler, tests interactions)
- B) **Per-service E2E** - Each identity service gets its own E2E suite (isolates failures, more CI parallelism)
- C) **Both** - Shared integration E2E + per-service targeted E2E (most coverage, most maintenance)
- D) **Flow-based** - Organize by user flows (login flow, consent flow, token refresh flow) regardless of service boundaries
- E) _(write your answer)_

**Answer**:

---

## Instructions

1. Write your answer (A-E) for each question
2. For option E, write your custom answer
3. Save this file
4. Agent will record decisions in plan.md and delete this file
