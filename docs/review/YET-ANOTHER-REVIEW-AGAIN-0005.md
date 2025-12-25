# Review 0005: ROOT CAUSE - SpecKit Fundamental Workflow Flaw

**Date**: 2025-12-24
**Severity**: CRITICAL
**Category**: SpecKit Methodology
**Status**: SYSTEMIC ISSUE IDENTIFIED

---

## Core Problem Statement

User reports: **"no amount of backport to copilot instructions + constitution + memory + spec seems to help"**

Regenerating derived documents (plan.md, tasks.md, DETAILED.md, EXECUTIVE.md) after fixing authoritative sources ALWAYS introduces NEW divergences or reintroduces OLD errors.

---

## Evidence of Systemic Failure

### Issue Pattern Timeline

1. **2025-12-24**: User identifies 6+ CRITICAL ERRORS in generated plan.md/tasks.md
   - learn-ps instead of learn-im
   - Per-service admin ports (9090/9091/9092/9093) instead of single 9090
   - Environment-based database choice instead of deployment-based
   - Schema-only multi-tenancy instead of dual-layer
   - Generic CRLDP URLs instead of base64-url-encoded serials
   - Wrong implementation order

2. **User attempts fixes**: Backports to constitution.md, clarify.md, plan.md
   - Result: Some files updated, others missed (spec.md, clarify.md still have errors)

3. **User regenerates**: Creates NEW plan.md, tasks.md, DETAILED.md, EXECUTIVE.md
   - Result: Still has errors (spec.md not fixed, clarify.md partially wrong)

4. **User frustration**: "Why do you keep fucking up these things? They have been clarified a dozen times."

### Files with CONFIRMED Contradictions

| File | learn-im | Admin Ports | Multi-Tenancy | CRLDP Format | Last Updated |
|------|----------|-------------|---------------|--------------|--------------|
| constitution.md | ✅ CORRECT | ✅ 9090 ALL | ✅ Dual-Layer | ✅ base64-url | 2025-12-24 |
| spec.md | ❌ learn-ps | ❌ 9090/9091/9092/9093 | ❌ Schema-Only | ⚠️ Ambiguous | NOT UPDATED |
| clarify.md | ✅ learn-im | ❌ 9090/9091/9092/9093 | ✅ Dual-Layer | ⚠️ Ambiguous | 2025-12-24 |
| plan.md | ✅ learn-im | ✅ 9090 ALL | ✅ Dual-Layer | ✅ base64-url | 2025-12-24 |
| tasks.md | ✅ learn-im | ✅ 9090 ALL | ✅ Dual-Layer | ✅ base64-url | 2025-12-24 |

**Critical Observation**: Authoritative sources (constitution.md, spec.md, clarify.md) CONTRADICT each other on admin ports and multi-tenancy.

---

## Root Cause Analysis

### Flawed Assumption: Single Source of Truth

SpecKit assumes:

1. **constitution.md** (Step 1) defines "project delivery requirements"
2. **spec.md** (Step 2) defines "technical specification"
3. **clarify.md** (Step 3) provides "Q&A clarifications"
4. **plan.md** (Step 4) is DERIVED from Steps 1-3
5. **tasks.md** (Step 5) is DERIVED from Step 4

**Problem**: Steps 1-3 are ALL authoritative, but there's NO mechanism to ensure they're consistent.

### Actual Divergence Pattern

```
constitution.md (Step 1)
    ↓
spec.md (Step 2) ← Generated from constitution, but may diverge
    ↓
clarify.md (Step 3) ← Q&A refines/contradicts spec
    ↓
plan.md (Step 4) ← Regenerated from 1+2+3, but which wins?
    ↓
tasks.md (Step 5) ← Regenerated from 4
```

**When user fixes plan.md**:

- plan.md updated ✅
- constitution.md updated ✅
- spec.md NOT updated ❌ (still has old values)
- clarify.md partially updated ⚠️ (some sections missed)

**Next regeneration reads**:

- constitution.md (9090 for ALL) ✅
- spec.md (9090/9091/9092/9093) ❌
- clarify.md (9090/9091/9092/9093) ❌
- **Result**: LLM sees contradiction, picks wrong value

---

## Why Backports Fail

### Problem 1: Incomplete Backporting

User fixes:

- ✅ constitution.md (authoritative source #1)
- ✅ plan.md (derived document)
- ❌ spec.md (authoritative source #2) - MISSED
- ⚠️ clarify.md (authoritative source #3) - PARTIAL

**Result**: Regeneration reads 2 out of 3 authoritative sources with OLD values.

### Problem 2: No Cross-Validation

SpecKit has NO step to:

- Grep authoritative sources for contradictions
- Flag when constitution.md says "9090" but spec.md says "9091"
- Force user to resolve before generating plan.md

**Result**: LLM silently picks one value (often the wrong one).

### Problem 3: Forward-Only Workflow

SpecKit assumes:

```
constitution → spec → clarify → plan → tasks → implement
(never backward)
```

**Reality**: Implementation insights often refine specifications:

```
implement → lessons learned → update plan → update spec
(bidirectional feedback)
```

**Result**: plan.md has refined CRLDP URL format (base64-url), but spec.md doesn't.

---

## Fundamental Flaws in SpecKit

### Flaw 1: Three Authoritative Sources, Zero Validation

**SpecKit declares**:

- constitution.md: Authoritative for delivery requirements
- spec.md: Authoritative for technical specifications
- clarify.md: Authoritative for implementation decisions

**Problem**: No automated validation that these three sources agree.

**Fix Needed**: Pre-generation validation step that greps all three sources for:

- Service names (learn-im, jose-ja, pki-ca, sm-kms)
- Admin ports (9090, 9091, 9092, 9093)
- Database choices (PostgreSQL, SQLite, environment vs deployment)
- Multi-tenancy patterns (schema-only, dual-layer, RLS)
- CRLDP URLs (format, encoding, batching)

If contradictions found, BLOCK plan.md generation until resolved.

### Flaw 2: No Bidirectional Feedback Loop

**SpecKit assumes**:

- constitution → spec → clarify → plan (forward only)
- Insights from plan/tasks/implement are manual backports

**Problem**: User forgets to backport changes to all sources.

**Fix Needed**: Automated backport validation:

- When plan.md is updated, check if spec.md/clarify.md need updates
- When tasks.md refines a detail, prompt user to update spec.md
- When implement/DETAILED.md discovers a constraint, require constitution.md update

### Flaw 3: LLM Silent Conflict Resolution

**When LLM sees**:

- constitution.md: "Admin port 9090 for ALL services"
- spec.md: "Identity admin port 9091, CA admin port 9092"

**LLM behavior**: Silently picks one (often wrong) without flagging contradiction.

**Fix Needed**: Pre-generation conflict detection that STOPS and asks user which is correct.

---

## Recommended Fixes

### Fix 1: Add Pre-Generation Validation (CRITICAL)

**Before generating plan.md, run automated validation**:

```bash
# Check for service name contradictions
grep -E "learn-ps|learn-im|Pet.*Store|InstantMessenger" constitution.md spec.md clarify.md

# Check for admin port contradictions
grep -E "admin.*port.*9090|9091|9092|9093" constitution.md spec.md clarify.md

# Check for multi-tenancy contradictions
grep -E "schema.*only|dual.*layer|row.*level|tenant_id" constitution.md spec.md clarify.md

# Check for database choice contradictions
grep -E "environment.*database|deployment.*database" constitution.md spec.md clarify.md
```

If any grep returns contradictory values, BLOCK plan.md generation.

Prompt user: "Contradiction detected: constitution.md says X, spec.md says Y. Which is correct?"

### Fix 2: Add Bidirectional Feedback Loop

**When plan.md or tasks.md is updated, check for backport needs**:

```bash
# Compare plan.md service names with spec.md
if plan.md has "learn-im" and spec.md has "learn-ps":
  PROMPT: "spec.md needs update: learn-ps → learn-im. Apply fix?"

# Compare plan.md admin ports with clarify.md
if plan.md has "9090 for ALL" and clarify.md has "9091/9092/9093":
  PROMPT: "clarify.md needs update: Remove per-service admin ports. Apply fix?"
```

### Fix 3: Add Authoritative Source Hierarchy

**Define precedence when contradictions exist**:

1. **constitution.md** (highest precedence) - Delivery requirements
2. **clarify.md** (medium precedence) - User Q&A decisions
3. **spec.md** (lowest precedence) - Initial technical spec

**Rule**: If contradiction detected, constitution.md wins, update spec.md/clarify.md automatically.

### Fix 4: Add Contradiction Dashboard

**Create docs/review/CONTRADICTIONS.md**:

```markdown
# Current Contradictions (Auto-Generated)

## Service Naming
- constitution.md: learn-im
- spec.md: learn-ps ❌ FIX NEEDED
- clarify.md: learn-im ✅

## Admin Ports
- constitution.md: 9090 for ALL
- spec.md: 9090/9091/9092/9093 ❌ FIX NEEDED
- clarify.md: 9090/9091/9092/9093 ❌ FIX NEEDED
```

Run before EVERY plan.md/tasks.md generation.

---

## Impact of Current Flaws

### Time Wasted

User reports trying "a dozen times" to fix these issues:

- 6+ iterations of backporting fixes
- 3+ iterations of regenerating plan.md/tasks.md
- 1+ day of debugging why fixes don't stick

**Estimated time lost**: 8-12 hours across multiple sessions

### User Frustration

User quote: "Why do you keep fucking up these things? They have been clarified a dozen times."

**Erosion of trust**: User questions whether SpecKit is "fundamentally flawed."

### Quality Degradation

- Derived documents (plan.md, tasks.md) have errors
- Implementation will use wrong specifications (learn-ps instead of learn-im, wrong admin ports)
- Potential production bugs from specification errors

---

## Conclusion

**SpecKit is NOT fundamentally flawed, but has serious workflow gaps**:

1. ✅ Spec-driven development is sound
2. ✅ Evidence-based completion is valuable
3. ❌ Multi-source validation is MISSING
4. ❌ Bidirectional feedback is MISSING
5. ❌ Conflict resolution is SILENT (should be EXPLICIT)

**Recommended Action**:

1. Implement pre-generation validation (CRITICAL)
2. Create contradiction dashboard (HIGH)
3. Add bidirectional feedback prompts (MEDIUM)
4. Define authoritative source hierarchy (MEDIUM)

**With these fixes, SpecKit can work reliably.**
