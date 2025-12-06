---
description: "Create or update project governing principles and development guidelines"
---

# /speckit.constitution

Create or update the project constitution - the immutable principles that govern all development.

## User Input

```
$ARGUMENTS
```

You MUST consider the user input before proceeding (if not empty).

## Outline

1. **Read existing constitution**: Check `.specify/memory/constitution.md` for current principles.

2. **Analyze project context**:
   - Review existing codebase patterns and conventions
   - Identify technology stack and constraints
   - Understand compliance requirements (FIPS 140-3, security, etc.)

3. **Constitution structure** must include:
   - Core Principles (immutable)
   - Security Requirements
   - Quality Gates (pre-commit, pre-push, testing)
   - Governance (decision authority, work patterns, documentation)

4. **Update or create** the constitution with:
   - Clear, enforceable principles
   - Version number and ratification date
   - Amendment history

## cryptoutil Constitution Principles

The cryptoutil constitution MUST enforce:

### I. FIPS 140-3 Compliance First

- All cryptographic operations use NIST FIPS 140-3 approved algorithms
- FIPS mode ALWAYS enabled, NEVER disabled
- Approved: RSA ≥2048, AES ≥128, EC NIST curves, EdDSA, PBKDF2-HMAC-SHA256
- BANNED: bcrypt, scrypt, Argon2, MD5, SHA-1

### II. Evidence-Based Task Completion

- No task complete without verifiable evidence
- Code evidence: `go build ./...` clean, `golangci-lint run` clean, coverage ≥90%
- Test evidence: All tests passing, no skips without tracking
- Integration evidence: Core E2E demos work

### III. Code Quality Excellence

- ALL linting/formatting errors MANDATORY to fix - NO EXCEPTIONS
- NEVER use `//nolint:` directives except for documented linter bugs
- File size limits: 300 (soft), 400 (medium), 500 (hard → refactor required)
- Coverage targets: 90%+ production, 95%+ infrastructure, 100% utility

### IV. KMS Hierarchical Key Security

- Multi-layer cryptographic barrier: Unseal → Root → Intermediate → Content keys
- All keys encrypted at rest with proper versioning and rotation
- NEVER use environment variables for secrets in production

### V. Product Architecture Clarity

- Clear separation: Infrastructure (internal/infra/) vs Products (internal/product/)
- Products: JOSE (P1), Identity (P2), KMS (P3), Certificates (P4)

## Output

Write the constitution to `.specify/memory/constitution.md` with:

- Versioned principles
- Ratification date
- Clear enforcement mechanisms
