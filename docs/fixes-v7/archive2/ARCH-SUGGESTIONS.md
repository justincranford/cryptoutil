# ARCHITECTURE.md Suggested Changes

**Source**: Lessons learned from fixes-v7 (11 phases, 220 tasks, 28+ commits).
**Purpose**: Identify gaps in docs/ARCHITECTURE.md that caused friction during execution.
**Status**: ALL 8 SUGGESTIONS APPLIED to ARCHITECTURE.md and propagated to instruction/agent files.

## ✅ Suggestion 1: Add Test Seam Injection Pattern (Section 10.2)

**Gap**: No documentation of the test seam injection pattern used to push jose
coverage from 90% to 95.3%.

**Problem**: Without this pattern, developers revert to mocking interfaces or
accepting low coverage for third-party library wrappers.

**Proposed Addition** (new subsection `10.2.4 Test Seam Injection Pattern`):

```markdown
#### 10.2.4 Test Seam Injection Pattern

**Purpose**: Enable error path testing in third-party library wrappers without
interfaces or mocks.

**Pattern**: Package-level function variables that default to real implementations
but can be replaced in tests with error-returning versions.

```go
// Production code (seams file)
var jwkKeySet = func(key any) (jwk.Set, error) { return jwk.Import(key) }

// Test code
func TestErrorPath(t *testing.T) {
    t.Parallel()
    original := jwkKeySet
    t.Cleanup(func() { jwkKeySet = original })
    jwkKeySet = func(any) (jwk.Set, error) { return nil, errors.New("injected") }
    // ... test error handling path
}
```

**When to Use**: Third-party library calls that may fail but rarely do in practice
(Set, Import, Marshal, PublicKey, keygen). NOT for business logic (use interfaces).

**Coverage Impact**: Typically adds 3-8% coverage by testing default/error switch
branches that are structurally unreachable in normal operation.
```

**Propagation Targets**:
- `.github/instructions/03-02.testing.instructions.md` — Add "Test Seam Injection" subsection
- `.github/agents/implementation-execution.agent.md` — Reference in testing guidance

---

## ✅ Suggestion 2: Add Coverage Ceiling Analysis Methodology (Section 10.2.3)

**Gap**: Coverage targets are blanket percentages (≥95%/≥98%) without guidance on
analyzing structural ceilings.

**Problem**: Teams hit a wall at ~90% and either give up or waste time on
unreachable code. The fixes-v7 JWX-COV-CEILING.md analysis proved that systematic
categorization of uncovered lines enables targeted improvements.

**Proposed Addition** (append to Section 10.2.3):

```markdown
**Coverage Ceiling Analysis** (RECOMMENDED before setting per-package targets):

1. Generate `go tool cover -html=coverage.out`
2. Categorize every uncovered line:
   - **Structurally testable**: Error paths reachable via seam injection or input manipulation
   - **Structurally unreachable**: Default switch cases for exhaustive type switches, dead code paths
   - **Third-party boundary**: Library return errors that require internal library state manipulation
3. Calculate ceiling: `ceiling = (total - unreachable) / total`
4. Set target at ceiling - 2% (safety margin)
5. Document exceptions in package README or coverage analysis file
```

**Propagation Targets**:
- `.github/instructions/03-02.testing.instructions.md` — Update coverage targets section
- `.github/instructions/06-01.evidence-based.instructions.md` — Reference ceiling methodology

---

## ✅ Suggestion 3: Document OTel Collector Configuration Requirements (Section 9.4)

**Gap**: No mention of OTel collector processor constraints (Docker socket for
`resourcedetection`).

**Problem**: The `docker` detector in `resourcedetection` processor blocked E2E
tests across THREE plan iterations (v1, v6, v7). Each time it was tagged
"pre-existing" and deferred.

**Proposed Addition** (append to Section 9.4 Telemetry Strategy):

```markdown
**OTel Collector Processor Constraints**:

| Processor | Requirement | Impact if Missing |
|-----------|-------------|-------------------|
| `resourcedetection` (docker) | `/var/run/docker.sock` mounted | Collector fails to start |
| `resourcedetection` (env) | None | Environment variable enrichment |
| `resourcedetection` (system) | None | Hostname/OS enrichment |

**Recommendation**: Use `detectors: [env, system]` for dev/CI environments.
Add `docker` detector only in production compose files where socket access is guaranteed.
```

**Propagation Targets**:
- `.github/instructions/02-03.observability.instructions.md` — Add OTel processor constraints
- `.github/instructions/04-01.deployment.instructions.md` — Reference Docker socket requirement
- `deployments/shared-telemetry/otel/otel-collector-config.yaml` — Fix config (remove `docker`)

---

## ✅ Suggestion 4: Expand Documentation Propagation Mapping (Section 12.7)

**Gap**: Mapping table has only 5 entries but there are 148 cross-references
across 18 instruction files and 15 across 5 agents. The mapping is severely
incomplete — it covers <4% of actual cross-references.

**Problem**: Without a complete mapping, changes to ARCHITECTURE.md sections
don't trigger updates in dependent instruction files. The mapping exists in
principle but not in practice.

**Proposed Change**: Replace the 5-entry table with a comprehensive mapping or
a reference to a generated mapping. Since manual maintenance of 148+ mappings is
impractical, the mapping should either be:

1. **Auto-generated**: A CI/CD linter (`cicd lint-propagation`) that extracts all
   `See [ARCHITECTURE.md Section X.Y]` references and produces a mapping table.
2. **Section-level**: Map at the ## section level (14 sections → 18 instruction files),
   not at ### subsection level.

**Proposed Section-Level Mapping**:

| ARCHITECTURE.md Section | Primary Instruction File(s) |
|------------------------|----------------------------|
| 1. Executive Summary | (none — context only) |
| 2. Strategic Vision | 01-01.terminology, 01-02.beast-mode, 02-02.versions |
| 3. Product Suite | 02-01.architecture |
| 4. System Architecture | 02-01.architecture, 03-03.golang |
| 5. Service Architecture | 02-01.architecture, 03-04.data-infrastructure |
| 6. Security Architecture | 02-05.security, 02-06.authn |
| 7. Data Architecture | 03-04.data-infrastructure |
| 8. API Architecture | 02-04.openapi |
| 9. Infrastructure Architecture | 02-03.observability, 04-01.deployment, 03-05.linting |
| 10. Testing Architecture | 03-02.testing |
| 11. Quality Architecture | 03-05.linting, 03-01.coding, 06-01.evidence-based |
| 12. Deployment Architecture | 04-01.deployment, 02-05.security |
| 13. Development Practices | 05-02.git, 03-01.coding, 03-03.golang |
| 14. Operational Excellence | (none — no instruction file covers ops) |
| Appendix A-C | (none — reference only) |

**Propagation Targets**:
- `docs/ARCHITECTURE.md` Section 12.7 — Replace 5-entry table with section-level mapping
- `.github/copilot-instructions.md` — Reference the mapping in documentation propagation section
- CI/CD — Consider `cicd lint-propagation` linter for automated verification

---

## ✅ Suggestion 5: Add Plan Lifecycle Management (Section 13)

**Gap**: No guidance on documentation plan lifecycle. fixes-v1 through v7
proliferation (7 iterations) shows the need for a disciplined approach.

**Problem**: Each fixes-vN plan started fresh, duplicated analysis from prior
plans, and created confusion about what was actually incomplete. The v7
"consolidation" itself took a full phase.

**Proposed Addition** (new subsection `13.6 Plan Lifecycle Management`):

```markdown
### 13.6 Plan Lifecycle Management

**Single Living Plan**: Maintain ONE plan per work stream. NEVER create vN+1.
Instead, add new phases to the existing plan.

**Lifecycle**: Active → Archived (move completed plan to `archive/` subdirectory)

**Structure**: `plan.md` (strategy + phases), `tasks.md` (checkboxes only),
`lessons.md` (post-completion retrospective).

**Completion Criteria**: Plan is "complete" when ALL tasks are checked OR
remaining tasks are explicitly descoped with justification.

**Anti-Patterns**:
- Creating fixes-v2 instead of adding Phase N+1 to fixes-v1
- Mixing analysis with checkboxes in tasks.md
- tasks.md exceeding 300 lines (split into per-phase files)
```

**Propagation Targets**:
- `.github/instructions/06-01.evidence-based.instructions.md` — Reference plan lifecycle
- `.github/agents/implementation-planning.agent.md` — Enforce single-plan pattern
- `.github/agents/implementation-execution.agent.md` — Reference archive workflow

---

## ✅ Suggestion 6: Per-Package Coverage Targets (Section 2.5 / 10.2.3)

**Gap**: Coverage targets are blanket percentages without per-package exceptions.

**Problem**: Packages wrapping third-party libraries (jose, tls) have structural
ceilings below 95%. The blanket mandate creates pressure to either (a) skip the
package or (b) write meaningless tests for unreachable code.

**Proposed Change**: Add exception mechanism to Section 2.5:

```markdown
**Package-Level Exceptions**: Packages MAY have targets below the mandatory
minimum IF a coverage ceiling analysis (see Section 10.2.3) documents the
structural ceiling. Exception format:

| Package | Standard Target | Actual Target | Ceiling | Justification |
|---------|----------------|---------------|---------|---------------|
| internal/shared/crypto/jose | 95% | 95% | ~96% | JWE OKP branches unreachable |
```

**Propagation Targets**:
- `.github/instructions/03-02.testing.instructions.md` — Add exception mechanism to coverage table
- `.github/instructions/06-01.evidence-based.instructions.md` — Reference ceiling analysis

---

## ✅ Suggestion 7: Add Agent ARCHITECTURE.md Requirements (Section 2.1)

**Gap**: `implementation-planning.agent.md` has ZERO ARCHITECTURE.md references.
The agent isolation principle means agents don't inherit instruction files, so
agents that generate plans need their OWN architectural context.

**Problem**: Planning agent creates plans without architectural awareness. Plans
may conflict with ARCHITECTURE.md patterns.

**Proposed Change**: Add to Section 2.1.1:

```markdown
**Agent Self-Containment Requirements**:
- Agents that generate implementation plans MUST reference relevant ARCHITECTURE.md
  sections (testing strategy, quality gates, coding standards)
- Agents that modify code MUST reference coding standards (Section 11, 13)
- Agents that modify deployments MUST reference deployment architecture (Section 12)
```

**Propagation Targets**:
- `.github/agents/implementation-planning.agent.md` — Add ARCHITECTURE.md references
- `.github/instructions/06-02.agent-format.instructions.md` — Add self-containment checklist
- All agent files — Audit for missing ARCHITECTURE.md references

---

## ✅ Suggestion 8: Infrastructure Blocker Escalation (Section 13)

**Gap**: No guidance on handling pre-existing infrastructure blockers.

**Problem**: The OTel Docker socket issue persisted across 3 plan iterations because
each plan treated it as "pre-existing, not our problem." Infrastructure blockers
should have an escalation path.

**Proposed Addition** (append to Section 13):

```markdown
### 13.7 Infrastructure Blocker Escalation

**Rule**: Infrastructure blockers that persist across 2+ plan iterations MUST be
resolved in the NEXT plan's Phase 0 (before feature work begins).

**Escalation Path**:
1. First encounter: Document as blocker, continue with workaround
2. Second encounter: Create dedicated fix task in current plan
3. Third encounter: MANDATORY Phase 0 fix — blocks ALL other work

**Anti-Pattern**: Tagging blockers as "pre-existing" across multiple plans without
resolution.
```

**Propagation Targets**:
- `.github/instructions/01-02.beast-mode.instructions.md` — Reference blocker escalation
- `.github/instructions/06-01.evidence-based.instructions.md` — Add escalation to quality gates

---

## Summary of Propagation Impact

| Suggestion | ARCHITECTURE.md Section | Files Affected |
|-----------|------------------------|----------------|
| 1. Test Seam Pattern | 10.2 (new 10.2.4) | 2 instruction + 1 agent |
| 2. Coverage Ceiling | 10.2.3 (append) | 2 instruction |
| 3. OTel Constraints | 9.4 (append) | 2 instruction + 1 config |
| 4. Propagation Mapping | 12.7 (replace) | 1 architecture + 1 copilot-instructions + CI/CD |
| 5. Plan Lifecycle | 13 (new 13.6) | 1 instruction + 2 agents |
| 6. Per-Package Coverage | 2.5 + 10.2.3 | 2 instruction |
| 7. Agent Requirements | 2.1.1 (append) | 1+ agents + 1 instruction |
| 8. Blocker Escalation | 13 (new 13.7) | 2 instruction |
| **Total** | **6 sections modified, 2 new** | **~12 instruction + ~4 agent + ~2 config** |
