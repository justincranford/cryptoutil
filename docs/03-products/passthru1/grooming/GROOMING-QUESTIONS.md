# Products Plan Grooming Questions

**Purpose**: Structured questions to refine and prioritize the docs/03-products/ reorganization plan
**Created**: November 29, 2025
**Status**: AWAITING ANSWERS

---

## Instructions

Select your answer by changing `[ ]` to `[x]` for each question. Add comments in the "Notes" field if needed.

---

## Section 1: Vision & Strategy (Q1-5)

### Q1. Timeline Reality Check

The plan shows 12+ weeks of phased work.

**Team size:**
- [x] Solo project
- [ ] 2-3 people
- [ ] 4+ team members

**Weekly time commitment:**
- [ ] <5 hours/week (hobby pace)
- [ ] 5-15 hours/week (side project)
- [x] 15-30 hours/week (part-time focus)
- [ ] 30+ hours/week (full-time)

**Notes**:

---

### Q2. Primary Driver

What's the #1 reason for this reorganization?

- [x] Pain today: Code duplication across packages
- [ ] Pain today: Difficulty adding new features
- [ ] Pain today: Hard to understand codebase structure
- [ ] Pain today: Testing is difficult/slow
- [ ] Future-proofing: Anticipating scale
- [x] Future-proofing: Planning to open source
- [x] Future-proofing: Multiple deployment targets
- [x] Learning: Exploring architecture patterns

**Notes**:

---

### Q3. MVP Definition

If you could only ship ONE outcome from this entire effort, what would it be?

- [ ] Clean JOSE library that others can import
- [ ] Deployable PKI/Certificate product
- [ ] Working Identity server with full OAuth 2.1 + OIDC
- [x] KMS with proper key hierarchy and rotation
- [x] Well-organized codebase (structure over features)
- [ ] Other (specify in notes)

**Notes**:

---

### Q4. External Consumers

Who will use these products/libraries?

- [x] Just me (personal project)
- [ ] Internal team only
- [x] Open source - expecting external contributors
- [ ] API customers - paid/enterprise
- [ ] Mix - some internal, some external

**Notes**:

---

### Q5. Monorepo vs Multi-repo

What's your repo strategy?

- [x] Keep monorepo - shared code benefits outweigh complexity
- [ ] Keep monorepo - too much work to split now
- [ ] Consider splitting - products are independent enough
- [ ] Already planning multi-repo - this is temporary
- [ ] Haven't decided yet

**Notes**:

---

## Section 2: Infrastructure Components (Q6-10)

### Q6. Infrastructure Priorities

Which THREE infrastructure components cause you the most pain TODAY?

**1st Priority:**
- [ ] I1. Configuration - config loading is duplicated/inconsistent
- [ ] I2. Networking - HTTP/TLS setup is messy
- [ ] I3. Testing - test utilities are scattered
- [ ] I4. Performance - no profiling/benchmarking infrastructure
- [ ] I5. Telemetry - logging/tracing inconsistent
- [ ] I6. Crypto - crypto code is duplicated
- [ ] I7. Database - ORM/migrations are painful
- [ ] I8. Containers - Docker configs are scattered
- [ ] I9. Deployment - CI/CD is fragile
- [x] I10. EASY OF DEMONSTRABILITY AND EASY OF USE (Write in vote)

**2nd Priority:**
- [ ] I1. Configuration
- [ ] I2. Networking
- [ ] I3. Testing
- [ ] I4. Performance
- [ ] I5. Telemetry
- [ ] I6. Crypto
- [ ] I7. Database
- [ ] I8. Containers
- [ ] I9. Deployment
- [x] I10. EASY OF DEMONSTRABILITY AND EASY OF USE (Write in vote)

**3rd Priority:**
- [ ] I1. Configuration
- [ ] I2. Networking
- [ ] I3. Testing
- [ ] I4. Performance
- [ ] I5. Telemetry
- [ ] I6. Crypto
- [ ] I7. Database
- [ ] I8. Containers
- [X] I9. Deployment

**Notes**:

---

### Q7. Existing vs Future Infrastructure

The plan mixes "refactor existing" (I1-I9) with "not yet implemented" (I10-I16). I DON'T KNOW!

- [ ] Remove I10-I16 from this plan entirely
- [ ] Keep I10-I16 for visibility but mark as "future"
- [ ] Move I10-I16 to separate "Future Vision" document
- [ ] Keep only HIGH priority future items (I11 Auditing, I16 Security)

**Notes**:

---

### Q8. I11 Auditing Priority

You marked auditing as HIGH priority. What drives this?

- [ ] SOC2 compliance requirement
- [ ] PCI-DSS compliance requirement
- [ ] HIPAA compliance requirement
- [X] FedRAMP compliance requirement
- [X] Internal security policy
- [ ] Customer requirement
- [X] Best practice (no specific requirement)
- [ ] Speculative - should lower priority

**Notes**:

---

### Q9. Infrastructure Testing Coverage

You target 95% coverage for infrastructure vs 85% for products.

- [X] Keep 95%/85% split - infrastructure is more critical I STRONGLY PREFER REUSABLE CODE FOR ALL PRODUCTS
- [ ] Uniform 90% for everything
- [ ] Uniform 85% for everything
- [ ] Lower targets - 80% infra, 70% products
- [ ] No fixed targets - coverage varies by criticality

**Notes**:

---

### Q10. Infrastructure Dependencies

Should there be a formal dependency graph between infrastructure components?

- [X] Yes - need to understand extraction order
- [ ] No - components are independent enough
- [ ] Partial - only for tightly coupled components
- [ ] Haven't thought about this yet

**Notes**:

---

## Section 3: Products (Q11-15)

### Q11. P1 JOSE Scope

JOSE is listed as a "product" but it's really a crypto library.

- [X] Keep as Product - it has standalone CLI/API value (Example is JWT issuer usable for KMS and Identity products)
- [ ] Move to Infrastructure - it's a building block only
- [ ] Split - core library is infra, CLI tools are product

**Notes**:

---

### Q12. P2 Identity Current State

Identity authz server just started working. What's the right next step?

- [ ] Stabilize first - complete remaining features before reorganizing
- [X] Reorganize first - easier to add features with clean structure
- [ ] Parallel - stabilize AND reorganize incrementally
- [ ] Pause Identity - focus on other products first

**Current Identity status (for reference):**

- [x] Basic OAuth 2.1 token endpoint working
- [ ] Client credentials flow working
- [x] Authorization code flow working
- [x] User authentication working
- [ ] MFA/WebAuthn working
- [ ] Session management working

**Notes**:

---

### Q13. P3 KMS Relationship

How should KMS and Identity interact?

- [ ] Identity → KMS (Identity uses KMS for key storage)
- [x] KMS → Identity (KMS uses Identity for authentication)
- [ ] Bidirectional (both use each other)
- [ ] Independent (no runtime dependency)
- [ ] Shared infra only (both use same crypto/database infra)

**Notes**:

---

### Q14. P4 Certificates Priority

You list Certificates as HIGH priority future product. What drives this?

- [ ] Real customer/business need
- [x] Personal interest/learning PKI
- [x] Completes the security story (KMS + Identity + Certs = full PKI)
- [x] TLS cert automation for other products
- [ ] Speculative - should lower priority
- [ ] Remove from plan entirely

**Notes**:

---

### Q15. Embedded Service Pattern

How should products share functionality?

- [x] Direct imports - products import each other's packages
- [ ] Embedded libraries - thin wrappers for embedding
- [ ] HTTP/gRPC APIs - products call each other as services
- [ ] Shared infrastructure only - no product-to-product dependencies
- [x] Mix - depends on the specific use case

**Notes**:

---

## Section 4: Migration & Risk (Q16-20)

### Q16. Breaking Changes Tolerance

How much are you willing to break existing cmd/* entry points and APIs?

- [ ] Zero tolerance - must maintain backward compatibility
- [ ] Minor breaks OK - document migration path
- [x] Major breaks OK - this is a v2 rewrite
- [ ] Internal only - no external consumers to break

**Notes**:

---

### Q17. Rollback Strategy

If reorganization causes unforeseen issues, what's your rollback plan?

- [ ] Git revert to pre-migration commit
- [ ] Maintain parallel structures temporarily
- [ ] Feature flags to switch old/new
- [ ] Work in feature branch first (merge when stable)
- [x] No rollback plan - accept the risk

**Notes**:

---

### Q18. Current Test Coverage

What's your actual test coverage baseline today?

- [x] >90% overall - very confident (ABSOLUTELY IMPERITIVE, THIS KEEPS LLMS AGENTS HONEST)
- [ ] 80-90% overall - reasonably confident
- [ ] 70-80% overall - some gaps
- [ ] 60-70% overall - significant gaps
- [ ] <60% overall - major risk
- [ ] Don't know - need to measure first

**Notes**:

---

### Q19. CI/CD Impact

Will the reorganization require changes to `.github/workflows/*`?

- [ ] No changes needed - paths are abstracted
- [ ] Minor path updates only
- [ ] Major workflow restructuring needed
- [ ] Need new workflows for new products
- [X] Haven't considered this yet

**Notes**:

---

### Q20. Documentation Debt

You have empty folders and incomplete docs. What's the approach?

- [ ] Complete all docs before any code changes
- [ ] Docs follow implementation (code first)
- [X] Minimal docs per phase, expand later
- [X] Delete empty folders, start fresh
- [X] Docs are optional - focus on code

**Notes**:

---

## Section 5: Practical Execution (Q21-25)

### Q21. Import Path Migration

Moving to `internal/infra/*` changes 100+ import paths. How to handle?

- [ ] Manual find/replace (tedious but safe)
- [ ] Scripted sed/awk replacement
- [ ] Go refactoring tools (gorename, gopls)
- [X] IDE refactoring (VS Code, GoLand)
- [ ] Gradual migration with type aliases
- [X] Big bang - all at once

**Notes**:

---

### Q22. Working Software Priority

Identity just started working. What's the priority?

- [ ] Stabilize Identity first - add features, then reorganize
- [X] Reorganize first - clean structure enables faster features
- [ ] Parallel - stabilize AND reorganize incrementally
- [ ] Abandon Identity reorganization - focus on other products

**Notes**:

---

### Q23. Phase Order

Should Infrastructure extraction come before Product organization?

- [X] Keep current order: Infrastructure → Products
- [ ] Reverse order: Products → Infrastructure
- [ ] Interleaved: Extract infra as needed per product
- [ ] No phases: Do everything incrementally

**Notes**:

---

### Q24. Phase Gate Criteria

What must be true before moving from Phase 1 to Phase 2?

- [X] All tests pass (100% green)
- [X] Coverage maintained (no regression)
- [X] No new lint errors
- [ ] Documentation updated
- [ ] Peer review completed
- [ ] Deployed to staging environment
- [ ] Other (specify in notes)

**Notes**:

---

### Q25. What Can Be Cut?

Which future infrastructure (I10-I16) should be REMOVED from this plan?

**I10. Messaging:**

- [ ] Keep - needed soon
- [X] Cut - speculative

**I11. Auditing:**

- [X] Keep - compliance requirement
- [ ] Cut - nice to have

**I12. Documentation (infra):**

- [ ] Keep - docs are critical
- [X] Cut - use existing tools

**I13. Internationalization:**

- [ ] Keep - global users expected
- [X] Cut - English only for now

**I14. Search:**

- [ ] Keep - search is important
- [X] Cut - not relevant to this project

**I15. Dev Tools:**

- [X] Keep - improves productivity
- [ ] Cut - existing tools are fine

**I16. Security (scanning):**

- [X] Keep - security is non-negotiable
- [ ] Cut - rely on existing SAST/DAST

**Notes**:

---

## Summary Section (Complete After Answering All Questions)

### Top 3 Priorities

After answering, identify your top 3:

- [ ] Priority 1: _______________
- [ ] Priority 2: _______________
- [ ] Priority 3: _______________

### Key Decisions

- [ ] Decision 1: _______________
- [ ] Decision 2: _______________
- [ ] Decision 3: _______________

### Recommended Next Actions

- [ ] Action 1: _______________
- [ ] Action 2: _______________
- [ ] Action 3: _______________

---

**Status**: AWAITING ANSWERS
**Next Step**: Complete all checkboxes, then request plan refactoring
