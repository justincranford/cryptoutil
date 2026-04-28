# Quizme v1 - Framework V19 Contradiction Resolution

## Question 1: Source of Truth When Task Status and Narrative Conflict

**Question**: If a tasks file shows many `Status: ❌` entries but lessons/executive narrative claims completion, which source should govern V19 execution gating?

**A)** Trust lessons/executive narrative as final truth.
**B)** Average both sources and proceed if most sections look complete.
**C)** Treat task-level status plus code evidence as primary truth, and downgrade narrative claims when inconsistent. (Probably recommended - newer reliability rule)
**D)** Ignore both and rely only on latest chat memory.
**E)**

**Answer**:

**Rationale**: This resolves the v17 contradiction between header/narrative completion and many unresolved task statuses.

## Question 2: Exclusions Policy for Lint Fitness

**Question**: When plan docs say temporary exclusions were removed, but linter code still contains active exclusion maps, what should V19 do?

**A)** Keep exclusions as-is if lint currently passes.
**B)** Hide exclusions from docs and keep code unchanged.
**C)** Re-audit each exclusion, remove stale entries immediately, and document only intentional permanent exceptions. (Probably recommended - newer code-first policy)
**D)** Remove all exclusions immediately, even if it breaks builds, and fix later.
**E)**

**Answer**:

**Rationale**: This addresses v18 claims versus actual exclusion state in fitness linter packages.

## Question 3: Handling Encoding-Corrupted Planning Documents

**Question**: For mojibake-corrupted status symbols/text in planning docs, what is the correct V19 handling?

**A)** Leave as-is if humans can still infer meaning.
**B)** Rewrite from scratch without preserving original semantics.
**C)** Repair encoding to UTF-8 clean text while preserving exact task intent and evidence references. (Probably recommended - newer documentation integrity practice)
**D)** Move corrupted docs to archive and stop using them.
**E)**

**Answer**:

**Rationale**: Corrupted symbols in v18 tasks reduce auditability and can hide incorrect status interpretation.

## Question 4: Quality Gate Completion Under Infrastructure Blockers

**Question**: If race tests are blocked by missing local toolchain (for example gcc on Windows), when can a phase be marked complete?

**A)** Always complete if other checks pass.
**B)** Complete and mention blocker only in lessons.
**C)** Mark blocked until blocker is resolved or approved alternate evidence path is documented in tasks and plan. (Probably recommended - newer strict gate semantics)
**D)** Remove race gate from plan permanently.
**E)**

**Answer**:

**Rationale**: Prevents premature completion claims when mandatory quality gates were not actually executed.

## Question 5: Session Evidence Source for 7-Day Chat Analysis

**Question**: For analyzing work quality from past sessions, which source should V19 trust most?

**A)** `debug-logs/*/main.jsonl` only.
**B)** `models.json` only.
**C)** `transcripts/*.jsonl` as primary, with `debug-logs` as metadata index. (Probably recommended - newer evidence path)
**D)** User recollection only.
**E)**

**Answer**:

**Rationale**: Last-7-day debug logs are metadata-light; transcript JSONL contains substantive conversation and tool execution evidence.
