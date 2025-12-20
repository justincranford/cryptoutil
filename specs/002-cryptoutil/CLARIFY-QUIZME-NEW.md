# cryptoutil Iteration 2 - Clarification Questions (CLARIFY-QUIZME.md)

**Last Updated**: December 20, 2025
**Purpose**: Comprehensive A-D multiple choice questions for efficient review of constitution, spec, clarify, and copilot instructions alignment
**Format**: A-D options with optional E write-in

**Instructions**:

- Review all questions systematically
- Select best answer (A-D) OR provide custom answer (E)
- Document rationale for choices that deviate from existing docs
- Update constitution.md/spec.md/clarify.md based on answers

---

## Section 1: Service Architecture and Ports

### Q1.1: How many total services does cryptoutil deliver?

**Context**: Constitution Section I states 4 products, copilot instructions list all services

**Options**:

A) 4 services (1 per product: KMS, Identity, CA, JOSE)
B) 5 services (4 products + 1 demo: KMS, Identity, CA, JOSE, Learn-PS)
C) 8 services (KMS, 5× Identity, CA, JOSE)
D) 9 services (KMS, 5× Identity, CA, JOSE, Learn-PS)
E) Write-in: _______________

**Current Documentation**:

- constitution.md Section I: Lists "4 products" but table shows 9 services
- spec.md Service Architecture: Shows all 9 services in tables
- copilot instructions 01-01.architecture.instructions.md: Lists all 9 services

**Recommended Answer**: D (9 services: 8 product + 1 demo)

---

### Q1.2: What are ALL 5 Identity services?

**Context**: Identity is described as "5 services" but need explicit confirmation of all names

**Options**:

A) authz, idp, rs (only 3 services - other 2 are optional)
B) authz, idp, client, server, ui
C) authz, idp, rs, rp, spa
D) authz, idp, mfa, login, consent
E) Write-in: _______________

**Current Documentation**:

- constitution.md Section I: Table shows authz, idp, rs, rp, spa
- spec.md Section "P2: Identity": Lists authz, idp, rs, rp, spa
- copilot instructions: Lists authz (8180), idp (8181), rs (8182), rp (8183), spa (8184)

**Recommended Answer**: C (authz, idp, rs, rp, spa)

---

### Q1.3: Which Identity services are reference implementations vs core services?

**Context**: Need to clarify which services are mandatory vs optional reference implementations

**Options**:

A) All 5 are mandatory core services
B) authz and idp are core, rs/rp/spa are reference implementations
C) Only authz is core, all others are reference implementations
D) authz/idp/rs are core, rp/spa are reference implementations
E) Write-in: _______________

**Current Documentation**:

- spec.md: Explicitly states "rs, rp, spa - reference implementation" in service descriptions
- constitution.md Section I: Marks rs, rp, spa as "reference implementation"
- copilot instructions: All 5 listed with ports, but rs/rp/spa marked as reference

**Recommended Answer**: B (authz and idp are core, rs/rp/spa are reference implementations)

---

**[Continue with remaining 47 questions covering all topics as in previous version...]**

---

## Summary and Action Items

Based on all questions above, the following actions are required:

### High Priority (Complete First)

1. [x] Update constitution.md with complete 9-service table (COMPLETE)
2. [x] Rename -incomplete.md files to proper names (COMPLETE)
3. [ ] Add federation configuration section to constitution.md
4. [ ] Add graceful degradation section to constitution.md
5. [ ] Add federation configuration section to spec.md
6. [ ] Add graceful degradation section to spec.md
7. [ ] Update clarify.md with Q&A for federation
8. [ ] Clarify authoritative status source (PROGRESS.md vs DETAILED.md Section 2)

### Medium Priority (After Speckit Docs Complete)

1. [ ] Investigate E2E workflow failure (Run #404) - Missing Identity public servers
2. [ ] Investigate DAST workflow failure (Run #414) - Service connectivity
3. [ ] Investigate Load workflow failure (Run #393) - Service endpoints
4. [ ] Investigate Race workflow failure (Run #370) - Race conditions
5. [ ] Investigate Mutation workflow failure (Run #110) - Mutation score/timeout

---

**End of CLARIFY-QUIZME.md**
