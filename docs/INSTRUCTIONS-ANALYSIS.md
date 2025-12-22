# Instructions Analysis - Multi-Choice Questions

**Analysis Date**: 2025-12-21
**Purpose**: Comprehensive review of `.github/instructions/*.instructions.md` files for quality, accuracy, and completeness
**References**: 
- https://code.visualstudio.com/docs/copilot/customization/custom-instructions
- https://github.com/github/spec-kit

---

## Category 1: File Organization and Naming

### Q1.1: File numbering consistency
**Issue**: Are all instruction files numbered correctly and sequentially?
- A) All files correctly numbered 01-01 through 06-03 (no gaps)
- B) Missing file numbers in sequence (gaps exist)
- C) Duplicate file numbers
- D) Non-standard numbering format
- E) Write-in: _______

**Current State**: A (verified: 01-01 through 06-03, sequential within groups)

### Q1.2: Renamed file references
**Issue**: After renaming 02-03.bind-address to 02-03.https-ports, are all references updated?
- A) All references updated in copilot-instructions.md and other files
- B) Broken references remain in copilot-instructions.md
- C) Broken references remain in other instruction files
- D) Broken references in spec/plan/constitution files
- E) Write-in: _______

**Current State**: A (copilot-instructions.md updated in commit 50cd7e62)

---

## Category 2: Content Duplication

### Q2.1: HTTPS endpoint configuration duplication
**Issue**: Is HTTPS endpoint configuration duplicated across files?
- A) No duplication - architecture references https-ports file only
- B) Duplication between architecture and https-ports files
- C) Duplication between architecture and security files
- D) Duplication across 3+ files
- E) Write-in: _______

**Current State**: A (architecture now references https-ports, content moved in commit 50cd7e62)

### Q2.2: PKI/TLS content duplication
**Issue**: Is PKI/TLS content duplicated across files?
- A) No duplication - cryptography, https-ports, and pki properly separated
- B) Duplication between cryptography and pki files
- C) Duplication between https-ports and pki files
- D) Duplication across 3+ files
- E) Write-in: _______

**Current State**: A (PKI in 02-10.pki, TLS config in 02-03.https-ports, FIPS crypto in 02-08.cryptography)

### Q2.3: Testing patterns duplication
**Issue**: Are testing patterns duplicated across testing.instructions.md and other files?
- A) No duplication - testing patterns centralized in 03-02.testing
- B) Duplication in golang.instructions.md
- C) Duplication in coding.instructions.md
- D) Duplication across 3+ files
- E) Write-in: _______

**Current State**: Unknown (needs review)

---

## Category 3: Content in Wrong File

### Q3.1: Security content in architecture file
**Issue**: Does architecture file contain security implementation details?
- A) No - architecture focuses on service structure, references security file
- B) Contains Windows Firewall prevention details (belongs in security)
- C) Contains TLS configuration details (belongs in https-ports or pki)
- D) Contains cryptography details (belongs in cryptography)
- E) Write-in: _______

**Current State**: A (Windows Firewall in 03-06.security, TLS in 02-03.https-ports)

### Q3.2: Database patterns in wrong file
**Issue**: Are database patterns correctly organized?
- A) database.instructions.md has ORM/SQL, sqlite-gorm.instructions.md has SQLite-specific patterns
- B) SQLite patterns duplicated in database and sqlite-gorm files
- C) GORM patterns in golang.instructions.md (should be in database)
- D) PostgreSQL patterns in github.instructions.md (should be in database)
- E) Write-in: _______

**Current State**: A (proper separation between generic database and SQLite-specific patterns)

---

## Category 4: Verbosity

### Q4.1: Example code verbosity in federation patterns
**Issue**: Are federation configuration examples too verbose?
- A) Concise - one representative example per pattern
- B) Verbose - multiple redundant YAML examples for same concept
- C) Verbose - includes unnecessary Go implementation code
- D) Verbose - both YAML and Go examples for same pattern
- E) Write-in: _______

**Current State**: A (simplified in commit 0d3530af)

### Q4.2: Testing instructions verbosity
**Issue**: Are testing instructions excessively verbose?
- A) Concise - patterns without excessive examples
- B) Verbose - too many code examples for same pattern
- C) Verbose - redundant explanations of same concept
- D) Verbose - excessive anti-pattern examples
- E) Write-in: _______

**Current State**: Unknown (needs review - 03-02.testing.instructions.md is 800+ lines)

### Q4.3: Speckit instructions verbosity
**Issue**: Is speckit.instructions.md appropriately concise?
- A) Concise - 105 lines, essential patterns only
- B) Verbose - excessive examples and explanations
- C) Too concise - missing critical workflow details
- D) Inconsistent - some sections verbose, others too brief
- E) Write-in: _______

**Current State**: A (condensed from 517 to 105 lines in commit c1f7a587)

---

## Category 5: Broken Reference Links

### Q5.1: Internal file references
**Issue**: Are internal references to other instruction files correct?
- A) All references valid (e.g., `02-03.https-ports.instructions.md`)
- B) References to renamed bind-address file exist
- C) References to deleted speckit-detailed file exist
- D) References to non-existent files
- E) Write-in: _______

**Current State**: Needs verification (grep for bind-address and speckit-detailed)

### Q5.2: Project file references
**Issue**: Are references to project files (constitution.md, spec.md, etc.) correct?
- A) All references valid and paths correct
- B) References to moved/renamed files
- C) References to deleted files
- D) Incorrect paths (e.g., wrong directory)
- E) Write-in: _______

**Current State**: Needs verification (grep for constitution.md, spec.md, plan.md, clarify.md)

### Q5.3: External URL references
**Issue**: Are external URLs (NIST, CA/Browser Forum, RFC, etc.) valid?
- A) All external URLs valid and accessible
- B) Broken URLs (404 errors)
- C) Outdated URLs (content moved, redirects)
- D) Missing URLs for cited standards
- E) Write-in: _______

**Current State**: Needs verification (extract and test all https:// URLs)

---

## Category 6: Missing Content

### Q6.1: Service template instructions completeness
**Issue**: Does service-template.instructions.md cover all template requirements?
- A) Complete - dual HTTPS, health checks, middleware, telemetry
- B) Missing middleware pipeline patterns
- C) Missing OpenTelemetry integration details
- D) Missing database abstraction patterns
- E) Write-in: _______

**Current State**: Needs review (02-02.service-template.instructions.md)

### Q6.2: Security instructions completeness
**Issue**: Does security.instructions.md cover all security patterns?
- A) Complete - Windows Firewall, localhost binding, TLS validation
- B) Missing input validation patterns
- C) Missing authentication/authorization patterns (covered elsewhere?)
- D) Missing secret management patterns (Docker/K8s secrets)
- E) Write-in: _______

**Current State**: Needs review (03-06.security.instructions.md)

### Q6.3: CI/CD workflow completeness
**Issue**: Does github.instructions.md cover all workflow patterns?
- A) Complete - all 11 workflows documented with triggers
- B) Missing workflow dependency patterns
- C) Missing artifact upload/download patterns
- D) Missing workflow failure handling patterns
- E) Write-in: _______

**Current State**: Needs review (04-01.github.instructions.md)

---

## Category 7: Ambiguous Content

### Q7.1: Deployment environment clarity
**Issue**: Are deployment environment patterns clearly distinguished?
- A) Clear - production vs test/dev patterns explicitly separated
- B) Ambiguous - mixed guidance for container vs local
- C) Ambiguous - unclear when to use 0.0.0.0 vs 127.0.0.1
- D) Ambiguous - conflicting statements across files
- E) Write-in: _______

**Current State**: A (https-ports file clearly separates production vs test/dev)

### Q7.2: Testing requirement clarity
**Issue**: Are testing requirements (coverage, mutation, timing) clearly specified?
- A) Clear - explicit thresholds (95%/98% coverage, 85%/98% mutation, <15s/<120s timing)
- B) Ambiguous - conflicting thresholds across files
- C) Ambiguous - unclear when to use which threshold
- D) Missing thresholds for some test types
- E) Write-in: _______

**Current State**: Needs review (cross-check testing.instructions.md vs speckit.instructions.md)

---

## Category 8: Content That Should Not Be in Instructions

### Q8.1: Project-specific implementation details
**Issue**: Do instruction files contain project-specific details that belong in specs/constitution?
- A) No - instructions are generic patterns, specs contain project details
- B) Contains specific port numbers that belong in constitution
- C) Contains specific service names that belong in spec
- D) Contains specific file paths that belong in plan
- E) Write-in: _______

**Current State**: Borderline (architecture.instructions.md has service port ranges table - may belong in constitution?)

### Q8.2: Historical/session-specific content
**Issue**: Do instruction files contain historical or session-specific content?
- A) No - instructions are timeless patterns
- B) Contains dated examples ("2025-12-14")
- C) Contains session-specific troubleshooting (belongs in DETAILED.md)
- D) Contains post-mortem references (belongs in EXECUTIVE.md)
- E) Write-in: _______

**Current State**: Needs review (check for dated examples in anti-patterns.instructions.md)

---

## Category 9: Consistency Issues

### Q9.1: Terminology consistency
**Issue**: Is terminology consistent across all instruction files?
- A) Consistent - same terms used for same concepts across files
- B) Inconsistent - "public endpoint" vs "public server" vs "public API"
- C) Inconsistent - "admin endpoint" vs "private endpoint" vs "admin server"
- D) Inconsistent - "bind address" vs "bind IP" vs "listen address"
- E) Write-in: _______

**Current State**: Needs review (grep for terminology variations)

### Q9.2: Format consistency
**Issue**: Is formatting consistent across instruction files?
- A) Consistent - same markdown structure, heading levels, code block formatting
- B) Inconsistent - varying heading depths (some use ###, others ####)
- C) Inconsistent - code block language tags (some use ```yaml, others ```yml)
- D) Inconsistent - bullet point styles (some use -, others use *)
- E) Write-in: _______

**Current State**: Needs review (check markdown formatting patterns)

---

## Priority Actions Required

Based on above analysis, prioritize these actions:

1. **HIGH**: Verify all internal file references (Q5.1) - search for "bind-address", "speckit-detailed"
2. **HIGH**: Test all external URLs (Q5.3) - validate NIST, CA/Browser Forum, RFC links
3. **MEDIUM**: Review testing.instructions.md for verbosity (Q4.2) - 800+ lines may be excessive
4. **MEDIUM**: Verify project file references (Q5.2) - constitution.md, spec.md paths
5. **MEDIUM**: Check terminology consistency (Q9.1) - standardize endpoint naming
6. **LOW**: Review anti-patterns.instructions.md for dated content (Q8.2)
7. **LOW**: Consider moving service port table from architecture to constitution (Q8.1)

---

## Validation Commands

```powershell
# Q5.1: Check for broken internal references
grep -r "bind-address" .github/instructions/
grep -r "speckit-detailed" .github/instructions/

# Q5.2: Check for project file references
grep -r "constitution\.md\|spec\.md\|plan\.md\|clarify\.md" .github/instructions/

# Q9.1: Check terminology variations
grep -ri "public endpoint\|public server\|public api" .github/instructions/ | measure
grep -ri "admin endpoint\|private endpoint\|admin server" .github/instructions/ | measure
```

---

## Notes

- Analysis conducted after commits c1f7a587, 50cd7e62, 0d3530af
- Many questions marked "Needs review" require manual file inspection
- External URL testing requires network connectivity and manual verification
- Some answers may change as project evolves (e.g., new instruction files added)
