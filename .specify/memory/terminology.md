# RFC 2119 Terminology - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/01-01.terminology.instructions.md`

**Source**: RFC 2119 "Key words for use in RFCs to Indicate Requirement Levels"

## Requirement Keywords

**From RFC 2119 and `.specify/memory/constitution.md` Section VIII**:

### Absolute Requirements

**MUST** = **REQUIRED** = **MANDATORY** = **SHALL**

**Definition**: Absolute requirement - NO exceptions allowed

**Usage**:
- Use when non-compliance breaks functionality, security, or correctness
- Indicates implementation is defective if not followed
- Test suite MUST verify compliance

**Examples**:
- "All cryptographic operations MUST use FIPS-approved algorithms"
- "Bind address MUST be configurable, NEVER hardcoded to 0.0.0.0"
- "Task is NOT complete until ALL quality gate checks MUST pass"

### Absolute Prohibitions

**MUST NOT** = **SHALL NOT**

**Definition**: Absolute prohibition - NO exceptions allowed

**Usage**:
- Use when behavior would cause security vulnerability, data loss, or critical failure
- Indicates implementation is defective if not followed
- Test suite MUST verify prohibition is enforced

**Examples**:
- "Services MUST NOT bypass otel-collector sidecar"
- "MUST NOT use ambiguous `auth` abbreviation"
- "MUST NOT cache authorization decisions (violates Zero Trust)"

### Recommendations

**SHOULD** = **RECOMMENDED**

**Definition**: Highly desirable - may ignore with strong justification

**Usage**:
- Use when compliance is best practice but exceptions are valid
- Document reason if recommendation not followed
- Default path unless specific circumstances justify exception

**Examples**:
- "Bind address SHOULD be 127.0.0.1 in tests (prevents Windows Firewall prompts)"
- "Test files SHOULD be under 500 lines"
- "Mutation score SHOULD be ≥98% for infrastructure/utility packages"

**SHOULD NOT** = **NOT RECOMMENDED**

**Definition**: Not advisable - may do with strong justification

**Usage**:
- Use when behavior is discouraged but valid use cases exist
- Requires explicit documentation of justification
- Review required before proceeding

**Examples**:
- "External access to admin endpoint SHOULD NOT be enabled (security risk)"
- "Self-signed TLS client certificates SHOULD NOT be used"
- "Pre-configured user-specific data in YAML SHOULD NOT be used (SQL realm required)"

### Optional

**MAY** = **OPTIONAL**

**Definition**: Truly optional - implementer decides based on needs

**Usage**:
- Use when feature/behavior is neither required nor discouraged
- No justification needed for inclusion or omission
- User discretion without review

**Examples**:
- "Port 9091 MAY be used if port 9090 conflicts with another service"
- "HTTP MAY be used temporarily for clear text debugging (NEVER in production)"
- "Service-specific metadata MAY be included in telemetry spans"

---

## Emphasis Keywords (Instruction Files Only)

**Purpose**: Highlight historically regression-prone areas or add emphasis to requirements

### CRITICAL

**Definition**: Historically regression-prone areas requiring extra attention

**Usage**:
- Use when past incidents (P0 post-mortems) show frequent violations
- Signals higher risk of LLM agent context loss or human error
- May be synonym for MUST if historical emphasis needed

**Examples**:
- "**CRITICAL**: format_go self-modification regression has occurred MULTIPLE times"
- "**CRITICAL**: Windows Firewall exception prevention - ALWAYS bind to 127.0.0.1"
- "**CRITICAL**: Evidence required for task completion claims"

**Context Determination**:
- Check for P0 post-mortem references (historical regression)
- Look for "has occurred multiple times" language (repetition pattern)
- Presence of detailed anti-pattern documentation (lessons learned)
- If yes → Historical emphasis, heightened attention required
- If no → Synonym for MUST with extra emphasis

### ALWAYS

**Definition**: Emphatic MUST (no exceptions)

**Usage**:
- Use for absolute requirements with historical violations
- Strengthens MUST when past non-compliance caused incidents
- Signals zero tolerance for exceptions

**Examples**:
- "**ALWAYS** read complete context before refactoring self-modifying code"
- "**ALWAYS** analyze baseline HTML before writing tests"
- "**ALWAYS** commit changes immediately when work is complete"

**Semantic Equivalence**: ALWAYS = MUST (emphatic form)

### NEVER

**Definition**: Emphatic MUST NOT (no exceptions)

**Usage**:
- Use for absolute prohibitions with historical violations
- Strengthens MUST NOT when past violations caused incidents
- Signals zero tolerance for exceptions

**Examples**:
- "**NEVER** stop working until user explicitly clicks STOP button"
- "**NEVER** use ambiguous `auth` abbreviation - ALWAYS use authn/authz"
- "**NEVER** mark tasks complete without objective evidence"

**Semantic Equivalence**: NEVER = MUST NOT (emphatic form)

---

## Usage Guidelines

### Interpretation Rules

**All keywords are semantically equivalent to their RFC 2119 base**:
- CRITICAL → Check context: Historical emphasis OR synonym for MUST
- ALWAYS → MUST (emphatic)
- NEVER → MUST NOT (emphatic)

**Context Clues for CRITICAL**:
1. **Historical Regression**: References to P0 post-mortems, "has occurred multiple times"
2. **Anti-Pattern Documentation**: Detailed lessons learned sections
3. **Emphasis Only**: No historical references → Synonym for MUST

### Examples with Context

**CRITICAL with Historical Context**:
```markdown
**CRITICAL: format_go self-modification regression has occurred MULTIPLE times**

Historical Incidents:
- Incident 1 (commit b934879b, Nov 17): Added backticks to comments
- Incident 2 (commit 71b0e90d, Nov 20): Added self-exclusion patterns
- Incident 3 (commit b0e4b6ef, Dec 16): Fixed infinite loop
- Incident 4 (commit 8c855a6e, Dec 16): Fixed test data

MANDATORY Prevention Rules:
- NEVER modify comments in enforce_any.go without reading full context
```
**Interpretation**: Historical emphasis (4 documented P0 incidents) + MUST-level requirement

**CRITICAL as MUST Synonym**:
```markdown
**CRITICAL: Evidence required for task completion claims**

NEVER mark tasks complete without objective evidence.
```
**Interpretation**: No historical incidents referenced → Emphatic MUST

---

## Abbreviation Standards - MANDATORY

**Purpose**: Prevent confusion between authentication (identity) and authorization (permissions)

### Authentication & Authorization

**CRITICAL: NEVER use ambiguous `auth` abbreviation - ALWAYS use specific abbreviations**

| Abbreviation | Meaning | Context |
|--------------|---------|---------|
| **authn** | Authentication | Identity verification (who you are) |
| **authz** | Authorization | Permission checking (what you can do) |
| **authnz** | Combined Authentication AND Authorization | When both are involved in a workflow |
| **BANNED: `auth`** | Ambiguous | Could mean authn, authz, or authnz |

### Other Auth* Abbreviations

| Abbreviation | Meaning | Context |
|--------------|---------|---------|
| **authc** | Authentication Context | SAML/OIDC context objects |
| **author** | Author/Authorship | NOT authentication |

### Usage Examples

**✅ CORRECT Usage**:
```
Filenames:
- authn-factors.md (authentication methods)
- authz-policies.md (authorization rules)
- authnz-middleware.go (combined authentication + authorization logic)

Variable Names:
- authnMethod := "bearer_token"
- authzPolicy := loadPolicy("admin")
- authnzMiddleware := newMiddleware()

Documentation:
- "See authn-authz-factors.md for complete authentication/authorization specifications"
```

**❌ WRONG Usage**:
```
Filenames:
- auth-factors.md (ambiguous - authentication or authorization?)
- auth.go (ambiguous - what does this contain?)
- authentication.md (too verbose - use authn-*.md)

Variable Names:
- authMethod (ambiguous - authn or authz?)
- auth := loadAuth() (what does this return?)

Documentation:
- "See auth.md" (which auth? authn or authz?)
```

### Rationale

**Prevents Confusion**:
- `auth` is ambiguous: Could mean authentication, authorization, or both
- Authentication (authn) and authorization (authz) are distinct security concepts
- Mixing these concepts creates security vulnerabilities

**Examples of Confusion**:
- "auth service" → Authentication service or authorization service?
- "auth middleware" → Validates identity or checks permissions?
- "auth token" → Bearer token (authn) or access token (authz)?

**Clear Communication**:
- `authn-service` → Identity verification service (clear)
- `authz-middleware` → Permission checking middleware (clear)
- `authnz-workflow` → Combined identity + permission flow (explicit)

---

## Cross-References

**Related Documentation**:
- Constitution: `.specify/memory/constitution.md` (Section VIII: Terminology)
- Authentication/Authorization: `.specify/memory/authn-authz-factors.md`
- Anti-patterns: `.specify/memory/anti-patterns.md` (CRITICAL keyword usage examples)
- SpecKit workflow: `.specify/memory/speckit.md` (Evidence-based requirements)
