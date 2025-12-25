# Review 0004: CRLDP URL Format Missing Base64-URL-Encoding Requirement

**Date**: 2025-12-24
**Severity**: MEDIUM
**Category**: CRLDP URL Format
**Status**: FOUND - NOT FIXED

---

## Issue Description

`spec.md`, `clarify.md`, and memory files specify CRLDP URLs as `https://ca.example.com/crl/serial-12345.crl` without explicitly requiring **base64-url-encoded serial numbers** as specified in plan.md user's critical fixes.

---

## Evidence

### Files Using CORRECT CRLDP URL format (base64-url-encoded)

1. **plan.md** (Lines 79-81):

   ```markdown
   **mTLS Revocation Checking**:

   - BOTH CRLDP and OCSP REQUIRED
   - CRLDP: Immediate sign and publish to HTTPS URL (NOT batched), one serial number per URL
     - URL format: `https://crl.example.com/<base64-url-encoded-serial>.crl`
     - NEVER batch multiple serials into one CRL file
   ```

2. **User's explicit instruction (commit message 9105bf68)**:

   ```
   CRITICAL FIXES:
   - CRLDP: Immediate sign+publish with base64-url-encoded serial, one per URL
   ```

### Files Using AMBIGUOUS CRLDP URL format (missing base64-url-encoding)

1. **spec.md** (Line 2304):

   ```markdown
   - **Distribution**: One serial number per HTTPS URL (e.g., `https://ca.example.com/crl/serial-12345.crl`)
   ```

   (Uses numeric serial `12345` instead of base64-url-encoded)

2. **clarify.md** (Line 753):

   ```markdown
   - Each CRLDP HTTPS URL MUST contain ONLY ONE certificate serial number
   ```

   (No URL format specified)

3. **.specify/memory/pki.md**: No CRLDP URL format specified

---

## Root Cause

User clarified CRLDP URL format in plan.md commit 9105bf68 as using **base64-url-encoded serial numbers**, but spec.md and clarify.md were NOT updated to reflect this requirement.

---

## Impact

- **MEDIUM**: URL format ambiguity could lead to inconsistent implementations
- **Interoperability**: Different encoding schemes (decimal, hex, base64-url) between services
- **RFC Compliance**: Base64-url encoding is URL-safe (RFC 4648 Section 5)
- **Character Safety**: Serial numbers may contain bytes unsafe for URLs without encoding

---

## Fix Required

1. **spec.md**: Update CRLDP URL example to use base64-url-encoded serial
2. **clarify.md**: Add URL format requirement with base64-url encoding
3. **.specify/memory/pki.md**: Add CRLDP URL format pattern

---

## Correct Specification (from plan.md)

**CRLDP URL Format**:

```
https://crl.example.com/<base64-url-encoded-serial>.crl
```

**Encoding Requirements**:

- Use RFC 4648 Section 5 base64-url encoding (URL-safe alphabet)
- Replace `+` with `-`, `/` with `_`, remove `=` padding
- Example: Serial `0x12AB34CD` → base64-url encode → `Eqs0zQ.crl`

**Rationale**:

- URL-safe encoding prevents escaping issues
- Consistent format across all CRL URLs
- RFC 4648 Section 5 standard encoding

---

## SpecKit Divergence Pattern

**Observation**: User added URL format detail to plan.md (Step 4), but spec.md (Step 2) and clarify.md (Step 3) were NOT updated.

**Root Cause**: plan.md refinements are NOT automatically backported to earlier authoritative sources (spec.md, clarify.md).

**Fundamental Flaw**: SpecKit treats documents as forward-only (constitution → spec → clarify → plan → tasks), but implementation insights from later steps (plan, tasks) often refine earlier steps.

**Recommendation**: Add bidirectional feedback loop - when plan.md or tasks.md refines a specification detail, MUST backport to spec.md and clarify.md.
