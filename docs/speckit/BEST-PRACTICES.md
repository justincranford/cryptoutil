# Speckit Best Practices

**Purpose**: Guide for using Spec Kit to verify completed work and identify subsequent work
**Location**: Based on `.github/instructions/06-01.speckit.instructions.md`

---

## Quick Reference

### Core Files

| File | Purpose | When to Update |
|------|---------|----------------|
| `.specify/memory/constitution.md` | Immutable principles | NEVER (unless stakeholder-approved) |
| `specs/001-cryptoutil/spec.md` | WHAT the system does | Adding features, changing APIs, updating status |
| `specs/001-cryptoutil/plan.md` | HOW and WHEN to implement | Starting phases, completing milestones, adjusting timelines |
| `docs/02-identityV2/PROJECT-STATUS.md` | Identity single source of truth | After every Identity work session |
| `docs/NOT-FINISHED.md` | Track incomplete work | After completing analysis of incomplete items |

---

## Verifying Completed Work

### 1. Constitution Compliance Check

```bash
# Verify work adheres to constitution principles:
# - FIPS 140-3 compliance (no bcrypt, scrypt, Argon2)
# - Evidence-based completion (build clean, tests pass)
# - Code quality excellence (no //nolint exceptions)
# - KMS hierarchical security (proper key hierarchy)
# - Product architecture clarity (infrastructure vs product separation)
```

### 2. Evidence Collection

**MANDATORY for every completed task:**

| Evidence Type | Command | Pass Criteria |
|---------------|---------|---------------|
| Build | `go build ./...` | Zero errors |
| Lint | `golangci-lint run` | Zero issues |
| Tests | `go test ./...` | All pass |
| Coverage | `go test -cover ./...` | ≥80% production |

### 3. Specification Update

Update `spec.md` status indicators:

```markdown
| Feature | Status |
|---------|--------|
| Complete | ✅ Working |
| Partial | ⚠️ Partial |
| Not Started | ❌ Not Implemented |
```

### 4. Plan Update

Update `plan.md` with:

- Phase status changes
- Success criteria checkboxes
- Completion dates
- Version number bump

### 5. Single Source Update

Update the relevant PROJECT-STATUS.md:

```bash
# Identity work → docs/02-identityV2/PROJECT-STATUS.md
# CA work → docs/05-ca/README.md
# General → docs/NOT-FINISHED.md
```

---

## Identifying Subsequent Work

### Method 1: Spec Review

1. Open `specs/001-cryptoutil/spec.md`
2. Search for `❌` (not implemented)
3. Search for `⚠️` (partial)
4. Review tables for missing status indicators

### Method 2: Plan Review

1. Open `specs/001-cryptoutil/plan.md`
2. Check unchecked items: `- [ ]`
3. Review phase status (not-started, in-progress)
4. Look at success criteria

### Method 3: TODO Scan

```bash
# Scan for TODOs in specific package
grep -r "TODO\|FIXME" internal/ca/

# Count TODOs by severity
grep -rn "TODO" internal/ | wc -l
```

### Method 4: NOT-FINISHED Review

1. Read `docs/NOT-FINISHED.md`
2. Check "Active Work Streams" section
3. Review "Deferred Work" priorities

### Method 5: Grooming Sessions

Use speckit grooming sessions in `docs/speckit/` for structured validation:

1. Answer 50 multiple-choice questions
2. Compare with provided answers
3. Identify gaps in understanding
4. Update specs based on gaps

---

## Grooming Session Process

### Creating a New Session

```markdown
# GROOMING-SESSION-##: [Topic]

## Overview
- **Focus Area**: [Specific area]
- **Related Spec Section**: [Link]
- **Prerequisites**: [Required knowledge]

## Questions

### Q1: [Question text]
A) [Option A]
B) [Option B]
C) [Option C]
D) [Option D]

**Answer**: [Letter]
**Explanation**: [Why correct]
```

### Using Grooming Sessions

1. **Review Questions**: Read all 50 questions
2. **Answer Questions**: Select best answer
3. **Check Answers**: Compare with provided
4. **Identify Gaps**: Note incorrect answers
5. **Update Specs**: Refine based on gaps
6. **Repeat**: Create new session if needed

---

## Work Verification Checklist

### Pre-Commit

- [ ] `go build ./...` passes
- [ ] `golangci-lint run` passes
- [ ] Related tests pass
- [ ] Coverage maintained

### Post-Commit

- [ ] spec.md status updated
- [ ] plan.md status updated
- [ ] PROJECT-STATUS.md updated (if applicable)
- [ ] NOT-FINISHED.md reviewed

### Session End

- [ ] All commits have proper messages
- [ ] Documentation updated
- [ ] Incomplete work documented
- [ ] Next steps identified

---

## Common Patterns

### Marking Feature Complete

```markdown
# In spec.md
| Feature | Status |
|---------|--------|
| My Feature | ✅ Complete |

# In plan.md
### Task X.Y: My Feature ✅
| Deliverable | Status |
|-------------|--------|
| Component A | ✅ Complete |
```

### Marking Phase Complete

```markdown
# In plan.md
## Phase X: Name

**Status**: ✅ COMPLETE
**Duration**: X weeks
**Goal**: [Goal statement]
```

### Documenting Blockers

```markdown
# In NOT-FINISHED.md
## Blockers

| Item | Reason | Owner | ETA |
|------|--------|-------|-----|
| Feature X | Dependency on Y | @team | Q2 |
```

---

## Anti-Patterns to Avoid

| Anti-Pattern | Why Bad | Instead |
|--------------|---------|---------|
| Modifying constitution | Breaks principles | Get stakeholder approval |
| Skipping evidence | No proof of completion | Always collect evidence |
| Outdated spec status | Misleading project state | Update immediately |
| Multiple status trackers | Confusion | Use single source |
| Skipping grooming | Gaps undetected | Complete sessions |

---

## Integration with Git Workflow

### Commit Message Pattern

```bash
# When updating specs
docs(speckit): update spec with [feature] completion

# When marking phase complete
docs(speckit): complete Phase X in plan.md

# When archiving completed work
docs: archive [sprint/feature] documentation
```

### PR Description Pattern

```markdown
## Speckit Updates

- Updated `spec.md` with [changes]
- Updated `plan.md` with [changes]
- Grooming session [##] completed
```

---

*Best Practices Version: 1.0.0*
*Based on: 06-01.speckit.instructions.md*
