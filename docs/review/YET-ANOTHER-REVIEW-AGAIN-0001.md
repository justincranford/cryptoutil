# Review 0001: Critical Service Naming Inconsistency (learn-im vs learn-ps)

**Date**: 2025-12-24
**Severity**: CRITICAL
**Category**: Service Naming
**Status**: FOUND - NOT FIXED

---

## Issue Description

`spec.md` uses **learn-ps** (Pet Store) instead of **learn-im** (InstantMessenger), directly contradicting constitution.md, clarify.md, plan.md, and tasks.md.

---

## Evidence

### Files Using CORRECT naming (learn-im)

1. **constitution.md** (Line 31):

   ```markdown
   | learn-im | Learn | InstantMessenger demonstration service | ✅ | ✅ |
   ```

2. **plan.md** (Line 37):

   ```markdown
   | **Demo: Learn** | 1 service (learn-im) | ❌ NOT STARTED - Phase 3 deliverable |
   ```

3. **tasks.md** (Line 53):

   ```markdown
   ## Phase 3: Learn-IM Demonstration Service
   ```

4. **clarify.md**: Uses learn-im consistently

### Files Using INCORRECT naming (learn-ps)

1. **spec.md** (Line 95):

   ```markdown
   | **learn-ps** | Learn | Pet Store | 8888-8889 | 127.0.0.1:9090 | Educational service demonstrating service template usage |
   ```

2. **spec.md** (Line 1966):

   ```markdown
   ### Learn-PS Demonstration Service (Phase 7)

   **Goal**: Create working Pet Store service using service template, validate reusability and completeness.
   ```

3. **spec.md** (Lines 1974-1975):

   ```markdown
   1. **learn-ps FIRST** (Phase 7):
      - CRITICAL: Implement learn-ps using extracted service template
   ```

4. **spec.md** (Line 1992):

   ```markdown
   - **Service**: PS (Pet Store service)
   ```

5. **spec.md** (Line 1994):

   ```markdown
   - **Scope**: Complete CRUD API for pet store (pets, orders, customers)
   ```

6. **spec.md** (Line 2081):

   ```markdown
   Origins: []string{"https://learn-ps.example.com"},
   ```

7. **spec.md** (Line 2109):

   ```markdown
   - **Starting Point**: Copy entire Learn-PS directory, modify for use case
   ```

---

## Root Cause

`spec.md` was NOT updated during the 2025-12-24 service naming correction that fixed constitution.md, clarify.md, plan.md, tasks.md, DETAILED.md, and EXECUTIVE.md.

---

## Impact

- **HIGH**: User requested clarification "The short name should be learn-im for Learn product Instant Messenger service"
- **Confusion**: Developers seeing spec.md will implement wrong service (Pet Store instead of InstantMessenger)
- **API Design**: Pet Store = CRUD API (pets, orders, customers) vs InstantMessenger = Message passing API (PUT/GET/DELETE /tx and /rx)
- **Divergence**: spec.md contradicts authoritative sources (constitution.md, clarify.md)

---

## Fix Required

Replace ALL occurrences of `learn-ps` and `Pet Store` in spec.md with `learn-im` and `InstantMessenger`.

Update service description from Pet Store CRUD API to InstantMessenger encrypted messaging API.

---

## SpecKit Divergence Pattern

**Observation**: User fixed constitution.md, clarify.md, plan.md, tasks.md, but spec.md was missed.

**Root Cause**: spec.md is an authoritative source (Step 2 of SpecKit), but derived documents (plan.md, tasks.md) were regenerated WITHOUT cross-checking spec.md for consistency.

**Fundamental Flaw**: SpecKit assumes authoritative sources are consistent, but there's NO automated validation to ensure spec.md, constitution.md, and clarify.md don't contradict each other.

**Recommendation**: Add pre-generation validation step that cross-checks all authoritative sources (constitution.md, spec.md, clarify.md) for common terms (service names, ports, architectures) before generating plan.md/tasks.md.
