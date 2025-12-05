# Cross-Reference Analysis: Spec Kit, Copilot Instructions, Feature Template

**Date**: December 5, 2025
**Purpose**: Validate alignment between three sources of development guidance

---

## Executive Summary

### Sources Analyzed

| Source | Location | Purpose |
|--------|----------|---------|
| **Spec Kit (github/spec-kit)** | External spec-driven.md, README.md | Industry-standard spec-driven development methodology |
| **Copilot Instructions** | `.github/instructions/*.md`, `.github/copilot-instructions.md` | Project-specific coding and workflow rules |
| **Feature Template** | `docs/feature-template/*.md` | LLM Agent autonomous execution planning |
| **Constitution** | `.specify/memory/constitution.md` | Immutable project principles |

### Overall Assessment

| Aspect | Status | Details |
|--------|--------|---------|
| **Conflicts** | ⚠️ Minor | 3 conflicts identified |
| **Gaps** | ⚠️ Moderate | 7 gaps identified |
| **Suboptimal** | ⚠️ Present | 9 issues identified |

---

## Source 1: Spec Kit (github/spec-kit)

### 1.1 Conflicts with Other Sources

| Conflict | Spec Kit Says | Other Source Says | Resolution |
|----------|---------------|-------------------|------------|
| **C1: Article I Library-First** | "Every feature MUST begin as standalone library" | Constitution says P1-P4 products, not libraries | **OK**: Products ARE libraries with server wrappers |
| **C2: Article II CLI Mandate** | "All libraries MUST expose CLI interface" | No CLI requirement in Copilot Instructions | **GAP**: Add CLI to constitution for consistency |
| **C3: Branch Workflow** | `/speckit.specify` creates feature branches | Constitution uses `specs/001-cryptoutil/` | **OK**: Using numbered directories instead of branches |

### 1.2 Missing from Other Sources

| Item | In Spec Kit | Missing From | Impact |
|------|-------------|--------------|--------|
| **M1: NEEDS CLARIFICATION markers** | Template requires explicit `[NEEDS CLARIFICATION]` markers | Copilot Instructions, Feature Template | MEDIUM - ambiguities go undocumented |
| **M2: Phase -1 Gates** | Pre-Implementation gates (Simplicity, Anti-Abstraction) | Constitution mentions gates but not "Phase -1" naming | LOW - similar concept exists |
| **M3: Complexity Tracking section** | Justified exceptions documented in implementation plan | Feature Template lacks explicit complexity tracking | MEDIUM - over-engineering risks |
| **M4: Contract-First Testing** | "Contract tests mandatory before implementation" | Testing Instructions only mentions table-driven tests | HIGH - API contracts not validated |

### 1.3 Suboptimal in Spec Kit

| Issue | Description | Recommendation |
|-------|-------------|----------------|
| **S1: Nine Articles too generic** | Articles I-IX designed for generic projects, not crypto/security-focused | Constitution correctly specializes with FIPS 140-3 |
| **S2: No security-specific gates** | No mention of FIPS compliance, key management | Constitution adds security principles ✅ |
| **S3: Single iteration assumption** | Spec-driven.md assumes greenfield; less guidance on iteration 2+ | Need better iteration continuation guidance |

---

## Source 2: Copilot Instructions (.github/instructions/)

### 2.1 Conflicts with Other Sources

| Conflict | Copilot Says | Other Source Says | Resolution |
|----------|--------------|-------------------|------------|
| **C4: File size limits** | 300/400/500 lines soft/medium/hard | Feature Template has no file limits | **ALIGN**: Add limits to Feature Template |
| **C5: Evidence-based completion** | 05-01: "PROJECT-STATUS.md is ONLY authoritative source" | Constitution: "PROGRESS.md is authoritative for speckit" | **CONFLICT**: Clarify which file for which purpose |
| **C6: Test parallelism** | "t.Parallel() is a FEATURE" | Constitution: "go test ./... -p=1" for reliability | **CLARIFY**: Default parallel, -p=1 for known flaky |

### 2.2 Missing from Other Sources

| Item | In Copilot Instructions | Missing From | Impact |
|------|-------------------------|--------------|--------|
| **M5: Magic values pattern** | Detailed magic package constants guidance | Spec Kit, Feature Template | LOW - project-specific |
| **M6: SQLite/PostgreSQL compatibility** | Cross-database GORM patterns | Spec Kit | LOW - project-specific |
| **M7: Dynamic port allocation** | Port 0 pattern for test servers | Feature Template | MEDIUM - test isolation |
| **M8: cicd utility pattern** | Self-exclusion in cicd commands | Spec Kit, Feature Template | LOW - project-specific |

### 2.3 Suboptimal in Copilot Instructions

| Issue | Description | Recommendation |
|-------|-------------|----------------|
| **S4: Scattered speckit guidance** | Speckit in 05-01 + 06-01 + copilot-instructions.md | Consolidate into single 06-01.speckit.instructions.md |
| **S5: Evidence file confusion** | PROJECT-STATUS.md vs PROGRESS.md vs EXECUTIVE-SUMMARY.md | Define clear ownership: PROJECT-STATUS=overall, PROGRESS=speckit |
| **S6: Missing test failure protocol** | No guidance on what to do when tests fail | Add: "STOP, document in PROGRESS.md, fix before continuing" |

---

## Source 3: Feature Template (docs/feature-template/)

### 3.1 Conflicts with Other Sources

| Conflict | Feature Template Says | Other Source Says | Resolution |
|----------|----------------------|-------------------|------------|
| **C7: Task naming** | `01-<TASK>.md` through `##-<TASK>.md` | Spec Kit uses `tasks.md` single file | **OK**: Different granularity levels |
| **C8: Quality gates** | Section 8 defines custom gates | Constitution Section VI has specific gates | **ALIGN**: Reference constitution gates |
| **C9: Evidence-based validation** | Version 2.0 added evidence-based validation | 05-01 has more detailed checklist | **ALIGN**: Feature Template should reference 05-01 |

### 3.2 Missing from Other Sources

| Item | In Feature Template | Missing From | Impact |
|------|---------------------|--------------|--------|
| **M9: Stakeholder Analysis** | Detailed stakeholder section | Spec Kit spec.md template | LOW - useful for large teams |
| **M10: Risk Management matrix** | Detailed risk categories | Constitution | MEDIUM - risk not addressed |
| **M11: Post-Mortem section** | Section 7 post-mortem template | Spec Kit | HIGH - learning not captured |

### 3.3 Suboptimal in Feature Template

| Issue | Description | Recommendation |
|-------|-------------|----------------|
| **S7: Too long (1682 lines)** | Violates 500-line hard limit | Split into core template + extension modules |
| **S8: Overlaps with spec.md** | Goals, Architecture duplicate spec.md | Reference spec.md instead of duplicating |
| **S9: No speckit command mapping** | Doesn't show which commands to run | Add "Spec Kit Command Mapping" section |

---

## Source 4: Constitution (.specify/memory/constitution.md)

### 4.1 Validation Against Spec Kit Best Practices

| Spec Kit Requirement | Constitution Status | Notes |
|----------------------|---------------------|-------|
| Immutable principles | ✅ Sections I-V, X | Well-defined principles |
| Amendment process | ⚠️ Missing | No Section 4.2 Amendment Process |
| Workflow steps defined | ✅ Section VI | 8-step workflow documented |
| Pre/Post gates | ✅ Section VI | Comprehensive gates |
| Test-first imperative | ⚠️ Partial | Table-driven tests, but not TDD red-green-refactor |
| Library-first | ❌ Missing | Products defined, not library-first |
| CLI mandate | ❌ Missing | No CLI interface requirement |

### 4.2 Unique Strengths (Not in Spec Kit)

| Strength | Description |
|----------|-------------|
| **FIPS 140-3 mandate** | Cryptographic compliance built into principles |
| **Four Products architecture** | Clear P1-P4 product structure |
| **KMS key hierarchy** | Cryptographic barrier architecture |
| **Evidence-based completion** | Concrete evidence requirements |
| **Code quality excellence** | Linting, coverage targets |

### 4.3 Recommended Updates to Constitution

| Update | Rationale | Priority |
|--------|-----------|----------|
| Add Amendment Process (Section XI) | Spec Kit requires documented change process | HIGH |
| Add CLI Interface requirement | Align with Spec Kit Article II | MEDIUM |
| Add NEEDS CLARIFICATION requirement | Explicit ambiguity tracking | HIGH |
| Add Complexity Tracking requirement | Prevent over-engineering | MEDIUM |
| Clarify PROJECT-STATUS vs PROGRESS | Resolve authoritative source confusion | HIGH |

---

## Consolidated Recommendations

### High Priority (Fix Immediately)

| ID | Action | Owner | Files to Update |
|----|--------|-------|-----------------|
| **H1** | Add Amendment Process to Constitution | Constitution | constitution.md |
| **H2** | Clarify authoritative status files | All | constitution.md, 05-01, 06-01 |
| **H3** | Add `[NEEDS CLARIFICATION]` requirement | Constitution | constitution.md, 06-01.speckit |
| **H4** | Add Contract-First Testing to Testing Instructions | Copilot | 01-02.testing |

### Medium Priority (Fix This Iteration)

| ID | Action | Owner | Files to Update |
|----|--------|-------|-----------------|
| **M1** | Consolidate speckit guidance | Copilot | 06-01.speckit.instructions.md |
| **M2** | Add CLI interface requirement | Constitution | constitution.md |
| **M3** | Add Complexity Tracking to Feature Template | Feature Template | feature-template.md |
| **M4** | Add test failure protocol | Copilot | 01-02.testing |

### Low Priority (Future Iteration)

| ID | Action | Owner | Files to Update |
|----|--------|-------|-----------------|
| **L1** | Split feature-template.md (1682 lines) | Feature Template | feature-template.md |
| **L2** | Add Risk Management to Constitution | Constitution | constitution.md |
| **L3** | Add Post-Mortem section to Spec Kit guidance | Copilot | 06-01.speckit |

---

## Iteration Recommendation

### Question: Fix Iteration 1 or Start Iteration 2?

**Recommendation: Complete Iteration 1 FIRST**

**Rationale**:
1. CHECKLIST-ITERATION-1.md claims 44/44 tasks (100%) but has known gaps
2. Tests require `-p=1` for reliability (test parallelism issues unfixed)
3. client_secret_jwt (70%) and private_key_jwt (50%) incomplete
4. ANALYSIS.md identifies 7 gaps not addressed

**Iteration 1 Completion Criteria (Before Starting Iteration 2)**:
- [ ] `go test ./...` passes without `-p=1` flag
- [ ] All auth methods at 100% or explicitly deferred
- [ ] All identified gaps have tasks in tasks.md
- [ ] Constitution updated with H1-H4 recommendations

**Iteration 2 Should Focus On**:
- P1 JOSE Authority standalone service
- P4 CA Server REST API
- Hardware Security Keys (U2F/FIDO)
- Email/SMS OTP delivery services

---

## Status Files Clarification

### Proposed Ownership Model

| File | Purpose | Owner | Update Frequency |
|------|---------|-------|------------------|
| `specs/001-cryptoutil/PROGRESS.md` | Spec Kit iteration tracking | /speckit.* commands | Every workflow step |
| `specs/001-cryptoutil/spec.md` | Product requirements | /speckit.specify | When requirements change |
| `specs/001-cryptoutil/tasks.md` | Task breakdown | /speckit.tasks | When plan changes |
| `specs/001-cryptoutil/CHECKLIST-*.md` | Gate validation | /speckit.checklist | End of iteration |
| `PROJECT-STATUS.md` (root) | Overall project health | Manual | Weekly/major milestones |
| `EXECUTIVE-SUMMARY.md` | Stakeholder overview | Manual | Monthly/releases |

---

*Cross-Reference Version: 1.0.0*
*Generated: December 5, 2025*
