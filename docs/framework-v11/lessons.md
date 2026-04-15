# Lessons - Framework V11: PKI-Init Cert Structure

**Created**: 2025-06-26
**Last Updated**: 2025-06-26

---

## Phase 1: Cert Structure Documentation

**Status**: ✅ COMPLETE (2025-06-26)

### What Worked

- **quizme format was invaluable**: Three design questions (Q1: postgres instance identity sharing, Q2: realm enumeration strategy, Q3: admin cert purpose) each resolved significant gaps that would have caused rework in Phase 2. The A-D option format forced precise articulation of alternatives.
- **Examples with explicit counts**: Adding concrete skeleton-template (86 dirs) and sm (144 dirs) examples immediately exposed count discrepancies between design intent and the actual layout pattern. Starting with examples grounded the discussion.
- **Separating global vs. per-PS-ID dirs**: The Directory Count Summary table with explicit "global dirs + PS-ID-specific dirs" breakdown made scaling behavior obvious. At SUITE scope (608 total vs. old 876 estimate), the savings were concrete.
- **14-category architecture**: Naming the 14 cert categories explicitly (not just listing directories) gave the design a vocabulary. "Cat 5" is now unambiguous shorthand.
- **File Format Convention section**: The explicit rule that truststores NEVER contain `.key` files prevented a latent confusion between keypairs (keystore) and CA chains (truststore). This would have caused implementation bugs in Phase 2.
- **`TARGET-DIRECTORY/{PKI-INIT-DOMAIN}/` positional arg design**: Two positional args (`tier-id`, `target-dir`) is cleaner than `--output-dir` and `--domain` flags. The output always goes in a subdirectory named after the domain, which prevents clobbering when generating multiple tiers.
- **Realm count as `|realms|` not hardcoded**: Cat 5 formula uses `2 × |realms| × 3` where `|realms|` comes from registry.yaml. The examples assume 2 realms but the design is general. Making this explicit prevented a future count discrepancy.

### What Didn't Work

- **Initial truststore-per-cert design**: The original design had all 14 category types with both keystore and truststore per cert. Realizing leaf certs never need truststores (only CA certs do) required removing ~6 categories of truststore directories. This was caught during tls-structure.md review, not during initial design.
- **Initial count estimate (120)**: The first count assumed keystores + truststores for every cert. After removing leaf cert truststores and accounting for the postgres instance identity sharing (Q1=A), the count dropped from 120 to 86. Better to derive counts from the pattern rather than estimating.
- **Q4 (postgres CA signing gap) not caught until quizme-v2**: The Cat 4 vs Cat 5 structural inconsistency (4 per-instance CAs but only 3 leaf PKI domains) was discovered during deep analysis AFTER Phase 1 was marked complete. This should have been part of Q1 in quizme-v1. Lesson: when accepting Q1=A (shared postgres identity), immediately check which CA signs that shared cert.
- **Algorithm and validity periods not specified**: Phase 1 focused on directory structure but left CA key algorithm (ECDSA vs RSA, key sizes) and cert validity periods unspecified. These are now Q5 and Q6 in quizme-v2. Phase 2 (Generator Rewrite) is now blocked pending those answers.

### Root Causes

- Truststore-per-leaf design: Came from over-applying the PKI "every cert has an associated trust anchor" principle. In practice, the trust anchor is the CA cert's truststore, not the leaf's. Fixed by rule: truststores only for CA certs.
- Count discrepancy: Counts were estimated rather than derived. Fixed by the formula in the directory count table.
- Q4 gap: Q1's answer (postgres instances share identity) was accepted without tracing the implication (if they share identity, which CA issues the shared cert?). Fixed by adding Q4 to quizme-v2.

### Patterns for Future Phases

- When accepting a "shared identity" design decision, immediately trace: "Which CA signs this shared cert? How do all recipients configure trust for it?"
- Derive directory counts from patterns (expand `{a,b}×{1,2}` etc.) rather than estimating.
- When a quizme answer changes a directory count, update ALL downstream counts (per-PS-ID, per-PRODUCT, per-SUITE) in the same document edit.
- The `Required logical layout` section is the single source of truth. All category descriptions, counts, and examples MUST be derivable from it.
- Algorithm agility mandate (`02-05.security.instructions.md`) applies to pki-init CA key generation. Do not hardcode algorithm choices — specify via config struct with FIPS defaults.

---

## Phase 2: Generator Rewrite

*(To be filled during Phase 2 execution)*

---

## Phase 3: pki-init CLI & Docker Volume Config

*(To be filled during Phase 3 execution)*

---

## Phase 4: Template & Deployment Updates

*(To be filled during Phase 4 execution)*

---

## Phase 5: Quality Gates & Testing

*(To be filled during Phase 5 execution)*

---

## Phase 6: Knowledge Propagation

*(To be filled during Phase 6 execution)*
