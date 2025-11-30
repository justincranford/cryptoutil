# Products Plan Grooming Analysis

**Purpose**: Analysis of grooming responses and refined recommendations
**Created**: November 29, 2025
**Status**: ANALYSIS COMPLETE

---

## Executive Summary

Based on your responses, here's the refined strategy:

| Aspect | Your Answer | Implication |
|--------|-------------|-------------|
| **Team** | Solo, 15-30 hrs/week | 12-week plan → realistic 16-20 weeks |
| **Primary Driver** | Code duplication + open source + learning | Quality > speed; documentation matters |
| **MVP** | KMS with key hierarchy + well-organized codebase | Focus on KMS first, Identity second |
| **Consumers** | Personal + open source | Public API stability eventually matters |
| **Repo Strategy** | Monorepo | Keep `internal/` structure, shared infra |

---

## Key Insights from Your Answers

### 1. NEW Infrastructure Priority: Demonstrability & Ease of Use

You wrote in "I10. EASY OF DEMONSTRABILITY AND EASY OF USE" as your #1 and #2 priority.

**This is critical insight** - the current plan doesn't address this at all!

**Recommended New Infrastructure Component:**

```
I10. Developer Experience (DX)
├── Demo modes (one-command startup)
├── Sample data/fixtures
├── Interactive documentation (Swagger UI already exists)
├── CLI tooling for common operations
└── Docker Compose profiles (dev, demo, prod)
```

### 2. Product Priorities Clarified

| Product | Your Priority | Reasoning |
|---------|---------------|-----------|
| **P3. KMS** | #1 (MVP) | Core value proposition, key hierarchy |
| **P2. Identity** | #2 | Provides auth for KMS, needs reorganization first |
| **P1. JOSE** | Keep as product | Standalone JWT issuer value |
| **P4. Certificates** | HIGH future | Completes security story |

### 3. Infrastructure to Keep vs Cut

**KEEP (4 items):**

- I11. Auditing - FedRAMP compliance, security policy
- I15. Dev Tools - productivity improvement
- I16. Security - non-negotiable

**CUT (4 items):**

- I10. Messaging - speculative
- I12. Documentation infra - use existing tools
- I13. Internationalization - English only
- I14. Search - not relevant

**ADD (1 new item):**

- I10-NEW. Developer Experience (from your write-in answer)

### 4. Architecture Decision: KMS → Identity

You selected: **KMS → Identity** (KMS uses Identity for authentication)

This means:

- Identity is a dependency of KMS
- Identity must be stable before KMS can use it
- Identity should be extracted/reorganized FIRST

**Revised dependency graph:**

```
Infrastructure (I1-I9) → Identity (P2) → KMS (P3) → Certificates (P4)
                              ↑
                         JOSE (P1) provides token signing
```

### 5. Migration Approach: Big Bang with High Quality Gates

Your answers indicate:

- **Breaking changes OK** - v2 rewrite mentality
- **No rollback plan** - accept risk (bold!)
- **IDE refactoring + big bang** - all at once
- **Strict quality gates**: 100% tests pass, coverage maintained, no lint errors

**Risk mitigation recommendation**: Since you have no rollback plan, the quality gates become critical. Add:

- Commit frequently (every logical unit)
- Tag before major migrations (`pre-infra-extraction`, `pre-product-reorg`)
- Run full test suite after each package move

---

## Refined Phase Plan

### Phase 0: Developer Experience Foundation (NEW - 1-2 weeks)

**Goal**: Make the project demonstrable before reorganization

**Tasks:**

1. Create `make demo` or `go run ./cmd/demo` one-liner
2. Fix Identity authz demo (you just got it working today!)
3. Add sample data/fixtures for quick demos
4. Document demo flow in README
5. Create Docker Compose "demo" profile

**Exit Criteria:**

- [ ] Single command starts working demo
- [ ] README has "Quick Start" section
- [ ] Demo works on clean checkout

### Phase 1: Infrastructure Extraction (4-6 weeks)

**Revised priority order based on your answers:**

| Order | Component | Rationale |
|-------|-----------|-----------|
| 1 | I6. Crypto | Foundation for everything |
| 2 | I7. Database | Used by Identity and KMS |
| 3 | I5. Telemetry | Observability for debugging |
| 4 | I1. Configuration | Standardize config loading |
| 5 | I9. Deployment | CI/CD improvements |
| 6-9 | Others | As needed |

**Exit Criteria:**

- [ ] All tests pass (100% green)
- [ ] Coverage ≥95% for infrastructure
- [ ] Zero lint errors
- [ ] Import paths use `internal/infra/*`

### Phase 2: Product Reorganization (4-6 weeks)

**Revised order based on dependency:**

| Order | Product | Rationale |
|-------|---------|-----------|
| 1 | P1. JOSE | No dependencies, foundation |
| 2 | P2. Identity | Required by KMS |
| 3 | P3. KMS | Your MVP, uses Identity |

**Exit Criteria:**

- [ ] All tests pass (100% green)
- [ ] Coverage ≥85% for products
- [ ] Zero lint errors
- [ ] Products in `internal/product/*`

### Phase 3: New Capabilities (4-6 weeks)

**Based on your priorities:**

| Order | Item | Rationale |
|-------|------|-----------|
| 1 | I11. Auditing | FedRAMP compliance |
| 2 | I16. Security scanning | Non-negotiable |
| 3 | I15. Dev Tools | Productivity |
| 4 | P4. Certificates | Complete security story |

**Exit Criteria:**

- [ ] Auditing infrastructure in place
- [ ] Security scanning integrated
- [ ] Certificate product architecture defined

---

## Revised Timeline

| Phase | Original | Revised | Reason |
|-------|----------|---------|--------|
| Phase 0 | N/A | 1-2 weeks | NEW - demonstrability |
| Phase 1 | 4 weeks | 4-6 weeks | Solo pace, quality gates |
| Phase 2 | 4 weeks | 4-6 weeks | Solo pace, quality gates |
| Phase 3 | 4 weeks | 4-6 weeks | New capabilities |
| **Total** | **12 weeks** | **13-20 weeks** | Realistic for solo + quality |

---

## Action Items

### Immediate (This Week)

1. [ ] Create Phase 0 task list for demonstrability
2. [ ] Tag current commit as `pre-reorganization`
3. [ ] Measure current test coverage baseline
4. [ ] Delete empty folders (`passthru1/`, `products/`)

### Short Term (Next 2 Weeks)

1. [ ] Complete Phase 0 (demo mode working)
2. [ ] Create I10-DX infrastructure component doc
3. [ ] Update README with Quick Start
4. [ ] Start I6 Crypto extraction planning

### Medium Term (Next Month)

1. [ ] Complete Phase 1 infrastructure extraction
2. [ ] Begin Phase 2 product reorganization
3. [ ] Maintain 90%+ test coverage throughout

---

## Questions for Grooming Session 2

Based on this analysis, here are follow-up questions:

### Demonstrability Deep Dive

1. What does "easy to demonstrate" look like to you?
   - [ ] Web UI with login flow
   - [ ] CLI commands with JSON output
   - [ ] Swagger UI interactive API
   - [ ] Video/GIF walkthrough
   - [ ] All of the above

2. Who is the demo audience?
   - [ ] Myself (learning/validation)
   - [ ] Potential contributors
   - [ ] Potential employers/clients
   - [ ] Conference talks/presentations

### KMS MVP Scope

1. What's the minimum viable KMS?
   - [ ] Key generation only
   - [ ] Key generation + storage
   - [ ] Key generation + storage + rotation
   - [ ] Full hierarchy (root → intermediate → content keys)
   - [ ] All above + HSM support

2. What key types are must-have for MVP?
   - [ ] RSA (signing, encryption)
   - [ ] ECDSA (signing)
   - [ ] ECDH (key agreement)
   - [ ] AES (symmetric encryption)
   - [ ] All of the above

### Identity Integration

1. When KMS uses Identity for auth, what's the auth model?
   - [ ] API keys (simple)
   - [ ] OAuth 2.1 client credentials
   - [ ] OAuth 2.1 authorization code (user context)
   - [ ] mTLS (certificate-based)
   - [ ] Mix depending on use case

---

## Updated Documentation Structure

Based on your preference for minimal docs + delete empty folders:

```
docs/03-products/
├── README.md                    # Keep - vision overview
├── GROOMING-QUESTIONS.md        # Keep - your answers
├── GROOMING-ANALYSIS.md         # NEW - this file
├── infrastructure/
│   └── I01-configuration.md     # Keep, add others as extracted
└── (delete passthru1/, products/)
```

---

**Status**: ANALYSIS COMPLETE
**Next Step**: Answer Grooming Session 2 questions OR begin Phase 0
