# Copilot Instructions Optimization Analysis

**Date**: December 19, 2025
**Purpose**: Identify improvements to .github/instructions/*.instructions.md for LLM agent effectiveness
**Format**: Multiple choice A-D with E write-in for quick review
**Source**: VS Code Copilot custom instructions best practices + current instruction file analysis

---

## Organization and Structure

### S1: Instruction File Naming Consistency

**Current State**: Numbered prefixes (01-xx, 02-xx, 03-xx) with descriptive suffixes

**Question**: Should we maintain current numbering scheme or reorganize?

**A)** Keep current numbered prefixes (01-08, 02-03, 03-04, 04-01, 05-01) - works well
**B)** Remove numbers, use only descriptive names (testing.instructions.md, architecture.instructions.md)
**C)** Add sub-numbering (01-01-xx for sub-categories within 01)
**D)** Group by product (kms/, identity/, jose/, ca/ subdirectories)
**E)** Other: _______________

**Your Answer**: A

**Recommendation**: A - Current numbering provides clear priority/loading order, makes cross-references easy

---

### S2: Instruction File Size Management

**Current State**: Files range from 89 lines (versions) to 563 lines (testing)

**Question**: Should we split large instruction files for better LLM processing?

**A)** Keep all files as-is (currently manageable)
**B)** Split testing.instructions.md into multiple files (unit-testing, integration-testing, mutation-testing)
**C)** Split all files >300 lines (apply file size limits to instructions too)
**D)** Consolidate small files into larger thematic files
**E)** Other: Optimize testing.instructions.md to be more concise and fit under 500 line limit

**Your Answer**: E

**Recommendation**: C - Apply same file size limits (300/400/500) to instruction files for consistency

---

## Content Clarity and Effectiveness

### C1: Terminology Standardization Across Instructions

**Current State**: Mixed usage of MUST/REQUIRED/MANDATORY/CRITICAL/ALWAYS

**Question**: How should we standardize requirement keywords?

**A)** Adopt RFC 2119 strictly (MUST, MUST NOT, SHOULD, MAY only)
**B)** Map all synonyms explicitly (MUST=REQUIRED=MANDATORY=CRITICAL per user intent)
**C)** Add terminology section to copilot-instructions.md explaining equivalence
**D)** Current usage is fine (context determines meaning)
**E)** Other: _______________

**Your Answer**: ___

**Recommendation**: C - Add to copilot-instructions.md citing constitution.md Section VIII as authority

---

### C2: CRITICAL vs MANDATORY Emphasis Pattern

**Current State**: "CRITICAL:" prefix used for attention-grabbing, not semantic distinction

**Question**: Should we formalize CRITICAL prefix usage?

**A)** Remove CRITICAL prefix (use MUST/MANDATORY only)
**B)** Reserve CRITICAL for regression-prone patterns (format_go self-modification, Windows Firewall)
**C)** Use CRITICAL for all MUST requirements (emphasize importance)
**D)** Create CRITICAL: prefix convention (high-visibility for LLM attention)
**E)** Other: _______________

**Your Answer**: ___

**Recommendation**: B - Reserve for historically problematic areas requiring extra LLM attention

---

### C3: Example Code Density

**Current State**: Mix of pattern descriptions and code examples

**Question**: Should we increase code example density?

**A)** Add more "CORRECT vs WRONG" code comparison examples
**B)** Remove some code examples (too verbose, increases token usage)
**C)** Current balance is good
**D)** Create separate examples/ directory with detailed code samples
**E)** Other: _______________

**Your Answer**:A

**Recommendation**: A - LLM agents learn better from concrete "DO THIS / DON'T DO THIS" examples

---

## Missing Topics and Gaps

### G1: Service Federation Configuration Patterns

**Current State**: Mentioned in architecture.instructions.md but no detailed patterns

**Question**: Should we add federation configuration guidance?

**A)** Add new 01-09.federation.instructions.md file
**B)** Add section to architecture.instructions.md (YAML config patterns)
**C)** Add section to docker.instructions.md (Docker Compose federation)
**D)** Not needed (addressed in spec.md sufficiently)
**E)** Other: _______________

**Your Answer**: B, and docker.instructions.md should reference architecture.instructions.md
**Recommendation**: B - Add to architecture.instructions.md with YAML examples and service discovery patterns

---

### G2: Hash Service Architecture Patterns

**Current State**: New hash versioning architecture (4 registries × 3 versions) not in instructions

**Question**: Should we add cryptography-specific instructions?

**A)** Add new 01-09.cryptography.instructions.md (FIPS algorithms, hash versioning, key management)
**B)** Add section to security.instructions.md
**C)** Add section to coding.instructions.md (algorithm agility patterns)
**D)** Not needed (constitution.md covers cryptography)
**E)** Other: _______________

**Your Answer**: A

**Recommendation**: A - Cryptography patterns warrant dedicated instruction file (FIPS compliance, hash versioning, algorithm selection)

---

### G3: Service Template Extraction Patterns

**Current State**: Phase 6 requirement (service template) not in instructions

**Question**: Should we add template/reusability guidance?

**A)** Add new 01-10.templates.instructions.md (service template, client SDK, reusability)
**B)** Add section to architecture.instructions.md
**C)** Add section to golang.instructions.md (Go patterns for templates)
**D)** Wait until Phase 6 implementation before documenting
**E)** Other: _______________

**Your Answer**: D

**Recommendation**: D - Wait for Phase 6 concrete implementation, then extract patterns to instructions

---

### G4: Speckit Workflow Integration

**Current State**: Speckit methodology in constitution.md but not integrated into copilot instructions

**Question**: Should we add Speckit-specific copilot instructions?

**A)** Add new 06-01.speckit.instructions.md (workflow gates, evidence requirements, feedback loops)
**B)** Add section to evidence-based-completion.instructions.md
**C)** Add section to git.instructions.md (Speckit commit patterns)
**D)** Not needed (constitution.md Section VIII covers Speckit)
**E)** Other: _______________

**Your Answer**: A

**Recommendation**: A - Speckit has specific LLM agent behaviors (clarify before implement, analyze gaps, checklist validation)

---

## Cross-Reference and Navigation

### N1: Cross-File Reference Consistency

**Current State**: Some files reference others ("Details: .github/instructions/01-04.testing.instructions.md")

**Question**: Should we standardize cross-reference format?

**A)** Use relative paths consistently (../../.github/instructions/file.md)
**B)** Use short names consistently (see testing.instructions.md)
**C)** Use section anchors (#section-name) for precise references
**D)** Current mixed format is fine
**E)** Other: _______________

**Your Answer**: C

**Recommendation**: C - Add anchor links for precise cross-references (easier LLM navigation)

---

### N2: Instruction File Index/TOC

**Current State**: copilot-instructions.md has table but no descriptions

**Question**: Should we enhance the instruction file index?

**A)** Add one-sentence description per file in copilot-instructions.md table
**B)** Create separate INDEX.md with detailed descriptions and when-to-use guidance
**C)** Add "When to apply this file" section to each instruction file header
**D)** Current table is sufficient
**E)** Other: _______________

**Your Answer**: ___

**Recommendation**: C - Help LLM agents understand WHEN to apply each instruction file

---

## Anti-Pattern Documentation

### A1: Lessons Learned Integration

**Current State**: Some anti-patterns documented (testing.instructions.md has "Common Testing Anti-Patterns")

**Question**: Should we systematically document anti-patterns across all files?

**A)** Add "Common Anti-Patterns" section to ALL instruction files
**B)** Create dedicated 07-01.anti-patterns.instructions.md file
**C)** Add anti-patterns only when historically problematic
**D)** Current approach is sufficient
**E)** Other: _______________

**Your Answer**: C, and B with short reason similar to you "Add only for historically problematic areas (prevents instruction bloat)"

**Recommendation**: C - Add only for historically problematic areas (prevents instruction bloat)

---

### A2: Post-Mortem Integration

**Current State**: docs/P0.* post-mortems exist but not referenced in instructions

**Question**: Should we link post-mortems to relevant instruction sections?

**A)** Add "See Post-Mortem: docs/P0.X" references to instruction files
**B)** Create post-mortem index in copilot-instructions.md
**C)** Extract lessons from post-mortems into instruction anti-patterns
**D)** Keep post-mortems separate from instructions
**E)** Other: _______________

**Your Answer**: C

**Recommendation**: C - Extract key lessons into instructions (post-mortems are historical context)

---

## Maintenance and Evolution

### M1: Instruction Version Control

**Current State**: No version tracking in instruction files

**Question**: Should we add versioning to instruction files?

**A)** Add version header to each file (Version: X.Y.Z, Last Updated: DATE)
**B)** Add amendment history table (like constitution.md)
**C)** Use git log as source of truth (no explicit versioning)
**D)** Add "Last Reviewed" date only (simpler than versions)
**E)** Other: _______________

**Your Answer**: C - I don't like "Last Reviewed: YYYY-MM-DD" header in instruction files because it bloats LLM agent context

**Recommendation**: D - "Last Reviewed: YYYY-MM-DD" header prevents stale instructions

---

### M2: Instruction Effectiveness Measurement

**Current State**: No mechanism to measure if instructions are followed

**Question**: Should we add instruction compliance tracking?

**A)** Add pre-commit hooks to validate instruction compliance
**B)** Add "Instruction Violations" section to PROGRESS.md
**C)** Periodic manual review of commits vs instructions
**D)** Trust LLM agents to follow instructions (no explicit tracking)
**E)** Other: _______________

**Your Answer**: B - current name is instructions/DETAILED.md of the current speckit directory, not PROGRESS.md
MUST: the file structure in specs\002-cryptoutil the the current best practice, so if copilot instructions reference incorrect structure they need fixing

**Recommendation**: C - Periodic review during Speckit checklist phase

---

### M3: Instruction Consolidation Opportunities

**Current State**: 17 instruction files, some overlap between files

**Question**: Should we consolidate related instructions?

**A)** Merge 03-01 (openapi), 03-02 (cross-platform), 03-03 (git), 03-04 (dast) into 03-00-development.instructions.md
**B)** Merge 02-01 (github), 02-02 (docker), 02-03 (observability) into 02-00-deployment.instructions.md
**C)** Keep current granular structure (easier to update individual topics)
**D)** Consolidate only if files <100 lines
**E)** Other: _______________

**Your Answer**: C, overlap is fine, duplication is not; as long as overlap is additive and not duplicative, then it's good

**Recommendation**: C - Granular files easier to maintain and update independently

---

## Specific Content Updates

### U1: Update Testing Instructions with New Clarifications

**Current State**: testing.instructions.md needs updates from SPECKIT-CONFLICTS-ANALYSIS

**Question**: Which clarifications need immediate integration?

**A)** All 26 clarifications (comprehensive update)
**B)** Only CRITICAL priority (C2, C3, O1, O2, O3) - 5 items
**C)** Only testing-related clarifications (C2, C3, O1, O2, Q1.1, Q1.2, Q2.1, Q2.2) - 8 items
**D)** Already applied in previous commits (no further updates needed)
**E)** Other: Assuming your list of 8 items is testing-only clarifications, the C is sufficient

**Your Answer**: E

**Recommendation**: D - Previous commits applied CRITICAL clarifications to testing.instructions.md

---

### U2: Update Architecture Instructions with Service Patterns

**Current State**: architecture.instructions.md needs CA instance count, admin ports, federation config

**Question**: Which architecture updates are most critical?

**A)** Admin port assignments (9090/9091/9092/9093 per product family)
**B)** CA multi-instance pattern (3 instances matching KMS/JOSE)
**C)** Service federation configuration (YAML static config)
**D)** All of the above
**E)** Other: _______________

**Your Answer**: D

**Recommendation**: D - All three are architectural requirements affecting service deployment

---

### U3: Update Security Instructions with Windows Firewall Prevention

**Current State**: security.instructions.md mentions 127.0.0.1 binding but needs emphasis

**Question**: How should we strengthen Windows Firewall prevention guidance?

**A)** Add "CRITICAL: Windows Firewall Exception Prevention" section at top of file
**B)** Add code examples (CORRECT: 127.0.0.1 vs WRONG: 0.0.0.0)
**C)** Add rationale (CI/CD automation, no user interaction)
**D)** All of the above
**E)** Other: _______________

**Your Answer**: D

**Recommendation**: D - Historical regression (multiple incidents) warrants comprehensive coverage

---

### U4: Add File Size Limits to Coding Instructions

**Current State**: coding.instructions.md mentions limits but needs formal section

**Question**: How should we document file size limits?

**A)** Add "File Size Limits" section matching constitution.md (300/400/500)
**B)** Add refactoring strategies (split by functionality, algorithm, extract helpers)
**C)** Add examples (jwk_util_test.go → multiple focused files)
**D)** All of the above
**E)** Other: coding.instructions.md

- testing.instructions.md has "CRITICAL: Test File Size Limits - MANDATORY"
- scoping to just testing is not correct, it must apply to all main/test code files in Go, Python, and Java

**Your Answer**: E

**Recommendation**: D - File size limits critical for LLM agent token efficiency

---

## Summary

**Total Questions**: 18
**Organization**: 2 questions
**Content Clarity**: 3 questions
**Missing Topics**: 4 questions
**Cross-Reference**: 2 questions
**Anti-Patterns**: 2 questions
**Maintenance**: 3 questions
**Specific Updates**: 4 questions

**Next Steps**:

1. Review answers and prioritize changes
2. Apply critical updates first (CRITICAL clarifications)
3. Add new instruction files if needed (cryptography, speckit, federation)
4. Update existing files with clarifications
5. Commit changes with conventional commits

---

## VS Code Copilot Custom Instructions Best Practices

**Reference**: <https://code.visualstudio.com/docs/copilot/customization/custom-instructions>

### Key Principles from VS Code Docs

1. **Be Specific**: Provide clear, unambiguous rules (avoid vague guidance)
2. **Use Examples**: Code examples > descriptions (LLMs learn from patterns)
3. **Keep Focused**: One topic per file (easier LLM context loading)
4. **Update Regularly**: Review instructions after post-mortems, regressions
5. **Test Instructions**: Verify LLM agent follows new instructions (measure effectiveness)

### Pattern Recognition

**Good Instruction Patterns**:

- ✅ "ALWAYS use X pattern" + code example
- ✅ "NEVER do Y" + anti-pattern example
- ✅ "MANDATORY: Rule with rationale"
- ✅ "CRITICAL: Historical regression prevention"

**Poor Instruction Patterns**:

- ❌ Vague: "Try to write good tests" (what does "good" mean?)
- ❌ Ambiguous: "Use best practices" (which practices?)
- ❌ No examples: "Follow pattern X" without showing X
- ❌ No rationale: "Do X because I said so" (LLM needs context)

### Instruction File Effectiveness Indicators

**High Effectiveness**:

- Code examples with CORRECT vs WRONG comparisons
- Clear rationale for each rule (helps LLM understand intent)
- Historical context (post-mortem references)
- Specific actions ("Run command X", "Check file Y")

**Low Effectiveness**:

- Abstract principles without concrete examples
- Multiple interpretations possible (ambiguous)
- No rationale (LLM can't judge trade-offs)
- Too many exceptions (undermines rule clarity)
