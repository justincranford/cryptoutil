# Quiz Me v2 - Framework v2

**Purpose**: Follow-up decisions needed after deep analysis of Q1, Q5, Q6 from quizme-v1.
**Evidence**: See  est-output/framework-v2-quizme-analysis/analysis.md for full analysis.
**After answering**: Merge into plan.md/tasks.md, then delete this file.

---

## Question 1: Skeleton-Template Decision (from Q1 analysis)

**Context**: skeleton-template is 21 files/66KB. lint-fitness is 10,500 lines/23 sub-linters. /new-service skill uses skeleton-template as copy source. Analysis found skeleton serves triple duty: human reference, scaffolding source, lint-fitness validation target.

**Question**: Based on the analysis, which approach for skeleton-template?

**A)** Keep skeleton-template as-is: human reference + scaffolding source + lint-fitness target. Invest in keeping it current with latest patterns. Minimal domain logic only.
**B)** Keep skeleton-template but add a CRUD reference endpoint (like sm-im messages) to demonstrate the full domain pattern including repository, service, and handler layers.
**C)** Deprecate skeleton-template. Update /new-service skill to copy from sm-im or jose-ja instead (both are mature references). Redirect lint-fitness to validate those services.
**D)** Keep skeleton-template minimal, but add a cicd new-service Go command that generates a new service programmatically (not copy-paste) using Go templates. skeleton-template becomes the source data for code generation.
**E)**

**Answer**:

**Rationale**: Determines Phase 6 Task 6.3 and ongoing skeleton-template investment.

---

## Question 2: Identity Domain Extraction Approach (from Q5 analysis)

**Context**: authz=133 files/916KB, idp=129 files/862KB, rp/rs/spa=10-18 files each. Shared identity packages=1,390KB. pki-ca has _ca-archived pattern (111 files/880KB). Analysis recommends Extract + Replace + Staged Reintegration.

**Question**: Do you approve the recommended staged extraction approach?

**A)** Yes, approve as-is: Archive shared packages to _archived/, replace all 6 services with fresh skeletons, stage reintegration (rp/rs/spa first, then authz, then idp, then pki-ca).
**B)** Modified: Archive only the 5 identity services. Keep pki-ca on its current incremental path (it's further along at ~85%). Do NOT extract pki-ca domain.
**C)** Modified: Do rp/rs/spa extraction immediately (they're tiny stubs anyway). Defer authz/idp extraction until framework-v2 patterns are proven on the 3 stubs. This reduces risk.
**D)** Reject extraction. Refactor identity services in-place to latest builder pattern. Slower but preserves working domain code without archival hibernation risk.
**E)**

**Answer**:

**Rationale**: Determines Phase 7-8 scope and whether identity gets clean-slate or in-place refactoring.

---

## Question 3: InsecureSkipVerify Phased Removal (from Q6 analysis)

**Context**: 26 files with InsecureSkipVerify: true. Server already auto-generates 3-tier CA hierarchy in Auto mode. Analysis identified 6 use cases with different CA scopes. Recommended phased approach: 2A (integration/contract tests), 2B (E2E/Docker), 2C (mTLS, deferred), 2D (PostgreSQL/infra TLS, deferred).

**Question**: Do you approve the phased approach? What is the scope for framework-v2?

**A)** Phase 2A only (integration + contract tests): Add TLSBundle() to ServiceServer, migrate test HTTP clients. Eliminates ~90% of InsecureSkipVerify. E2E and mTLS deferred.
**B)** Phase 2A + 2B (integration + contract + E2E): Also add TLS CA as Docker secrets for E2E tests. Eliminates 100% of InsecureSkipVerify.
**C)** Phase 2A + 2B + 2C (all except PostgreSQL/infra): Also implement service-to-service mTLS with shared CA at deployment level.
**D)** All phases (2A-2D): Complete TLS everywhere, including PostgreSQL TLS and infrastructure TLS for OTel/Grafana.
**E)**

**Answer**:

**Rationale**: Determines Phase 2 scope and ARCHITECTURE.md TLS documentation updates.

---

## Question 4: ARCHITECTURE.md TLS Gaps (from Q6 analysis)

**Context**: Analysis found 6 gaps in ARCHITECTURE.md related to TLS:
1. TLS Certificate Configuration table (Production/E2E/Unit) exists in instructions but NOT in ARCHITECTURE.md
2. Secrets Coordination Strategy (12.3.3) doesn't include TLS CA/cert/key secrets
3. No TLS test bundle pattern documented for integration tests
4. No ServiceServer.TLSBundle() accessor pattern documented
5. No mTLS deployment architecture documented
6. TLS mode taxonomy (Static/Mixed/Auto) exists in code but not in ARCHITECTURE.md

**Question**: Should these gaps be fixed as part of framework-v2, or deferred?

**A)** Fix ALL 6 gaps as part of Phase 2 (when we do the InsecureSkipVerify removal). Document what we implement.
**B)** Fix gaps 1, 3, 4, 6 now (they describe what already exists in code). Defer gaps 2, 5 (mTLS) to when mTLS is actually implemented.
**C)** Fix ALL gaps now, including the mTLS architecture design (even if implementation is deferred). Design-first approach.
**D)** Defer all gaps. Focus on code changes; update ARCHITECTURE.md in the Knowledge Propagation phase (Phase 9).
**E)**

**Answer**:

**Rationale**: Determines scope of ARCHITECTURE.md updates in framework-v2.

---

## Question 5: Architecture Status Table Accuracy

**Context**: ARCHITECTURE.md Section 3.2 implementation status table shows identity-authz and identity-idp as "Complete 100%" and identity-rs as "Complete 100%", but framework-v2 plan describes them as "Stub" and "half-baked". The table is stale.

**Question**: How should the status table be corrected?

**A)** Update immediately to reflect actual state. authz/idp/rs show domain logic but framework patterns are incomplete. Mark as "Partial" with notes.
**B)** Update after extraction decision (Q2 above). If we extract domain, status drops to 0%. If we refactor in-place, status stays at current level.
**C)** Remove the implementation status table entirely. It's hard to maintain and becomes stale quickly. Replace with a "last verified" date per service.
**D)** Keep the table but add a "Framework Compliance" column separate from "Domain Completeness". A service can be 100% framework-compliant but 0% domain-complete.
**E)**

**Answer**:

**Rationale**: Ensures ARCHITECTURE.md accurately reflects reality.
