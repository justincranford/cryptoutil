# Passthru2 Grooming Session 7: Final Implementation & Closing Questions

**Purpose**: Final grooming session to address remaining implementation details and closing items before passthru2 completion.
**Created**: 2025-12-01
**Status**: AWAITING ANSWERS

---

## Session 6 Analysis Summary (Grooming 6 Feedback)

### Key Decision from Session 6

| Topic | Decision | Implications |
|-------|----------|--------------|
| **Vault Integration** | NO HashiCorp Vault | 1) Commercial license prohibited, 2) KMS already provides secrets management |

**Clarification**: Any "vault integration" references mean generic secrets abstraction pattern, NOT HashiCorp Vault dependency. The KMS server provides its own secrets management capabilities.

---

## Section 1: Claims & Scopes Implementation (Q1-5)

### Q1. OIDC Claim Extraction Strategy

How comprehensively should we extract OIDC claims from tokens?

- [ ] A. Standard claims only (sub, iss, aud, exp, iat, nbf)
- [ ] B. Standard + common optional (name, email, preferred_username, groups)
- [ ] C. All OIDC standard claims + custom claim passthrough
- [ ] D. Configurable whitelist of claims to extract

Notes:

---

### Q2. Custom Claim Namespace

For custom claims (tenant_id, service_name, etc.), should we:

- [ ] A. Use flat structure (tenant_id at root)
- [ ] B. Use namespaced claims (e.g., `urn:cryptoutil:tenant_id`)
- [ ] C. Support both with configuration
- [ ] D. Follow issuer's convention (Identity uses flat, federated may vary)

Notes:

---

### Q3. Scope Hierarchy Structure

For the hybrid scope model, should scopes be:

- [ ] A. Flat strings only (`kms:read`, `kms:write`, `kms:encrypt`)
- [ ] B. Hierarchical with inheritance (`kms:admin` implies `kms:read`, `kms:write`)
- [ ] C. Resource-based (`kms:pool:read`, `kms:key:encrypt`)
- [ ] D. Configurable scope model per deployment

Notes:

---

### Q4. Scope Validation Mode

When validating scopes, should we:

- [ ] A. Require exact scope match
- [ ] B. Support wildcard scopes (`kms:*`)
- [ ] C. Support scope prefixes (`kms:` matches `kms:read`)
- [ ] D. Configurable matching strategy

Notes:

---

### Q5. Missing Scope Behavior

When a request lacks required scopes:

- [ ] A. Return 403 with scope_insufficient error
- [ ] B. Return 403 with list of missing scopes
- [ ] C. Return 403 with required vs provided comparison
- [ ] D. Configurable detail level (matches ErrorDetailLevel)

Notes:

---

## Section 2: Demo Integration Architecture (Q6-10)

### Q6. Demo Binary Distribution

How should the demo binary be distributed?

- [ ] A. Single binary with subcommands (`cryptoutil-demo kms|identity|all`)
- [ ] B. Separate binaries (`demo-kms`, `demo-identity`)
- [ ] C. Unified binary with flags (`cryptoutil --mode=demo-kms`)
- [ ] D. Scripts calling existing binaries

Notes:

---

### Q7. Demo Compose Architecture

For integration demo, should we:

- [ ] A. Single compose file with all services
- [ ] B. Separate compose files per product + integration overlay
- [ ] C. Base compose + profile-based variants
- [ ] D. Existing compose structure is sufficient

Notes:

---

### Q8. Demo Script Implementation

Demo scripts should be:

- [ ] A. Go CLI with structured output
- [ ] B. Shell scripts (bash/PowerShell)
- [ ] C. Python scripts
- [ ] D. Go CLI preferred (cross-platform)

Notes:

---

### Q9. Demo Data Verification

After demo setup, verification should:

- [ ] A. Query APIs to confirm entities exist
- [ ] B. Perform functional test (encrypt → decrypt)
- [ ] C. Both A and B
- [ ] D. Existence check sufficient, functional test in CI

Notes:

---

### Q10. Demo Cleanup Strategy

For demo reset/cleanup:

- [ ] A. Delete all data, reseed from scratch
- [ ] B. Soft reset (mark deleted, reseed)
- [ ] C. Truncate + reseed (fastest)
- [ ] D. Configurable per deployment

Notes:

---

## Section 3: Phase 4 Realm Authentication (Q11-15)

### Q11. File Realm Format

File realm configuration should use:

- [ ] A. YAML file with users/credentials
- [ ] B. JSON file
- [ ] C. htpasswd-style format
- [ ] D. YAML preferred (consistent with other config)

Notes:

---

### Q12. File Realm Password Storage

Passwords in file realm should be stored as:

- [ ] A. PBKDF2-SHA256 hashes (consistent with Identity)
- [ ] B. bcrypt hashes
- [ ] C. Argon2 hashes
- [ ] D. PBKDF2 (FIPS-approved, per security instructions)

Notes:

---

### Q13. DB Realm Table Naming

For KMS-specific realm tables:

- [ ] A. `kms_users`, `kms_roles`, `kms_permissions`
- [ ] B. `realm_users`, `realm_roles` (generic)
- [ ] C. Prefix with tenant ID if multi-tenant
- [ ] D. Configurable prefix

Notes:

---

### Q14. Realm Priority Resolution

When multiple realms match (file + db + federation):

- [ ] A. First match wins (config order)
- [ ] B. Most specific wins (user in file overrides federation)
- [ ] C. Merge permissions from all matching realms
- [ ] D. Configurable resolution strategy

Notes:

---

### Q15. Federation Trust Model

For federated identity validation:

- [ ] A. Trust issuer + audience validation
- [ ] B. Trust issuer + audience + signature verification
- [ ] C. Full OIDC Discovery validation
- [ ] D. Configurable trust level

Notes:

---

## Section 4: Phase 5 CI & Quality (Q16-20)

### Q16. Coverage Enforcement Strategy

Coverage gates should:

- [ ] A. Fail CI below threshold (strict)
- [ ] B. Warn below threshold, fail on regression
- [ ] C. Report only, no enforcement
- [ ] D. Strict for core packages, warn for others

Notes:

---

### Q17. Benchmark Baseline Management

Benchmark baselines should be:

- [ ] A. Committed to repo (tracked)
- [ ] B. Stored in CI artifacts
- [ ] C. Local-only (untracked)
- [ ] D. Committed + CI comparison

Notes:

---

### Q18. Test Factory Organization

Test factories (`testutil`) should:

- [ ] A. Single package with all factories
- [ ] B. Per-domain factories (`testutil/identity`, `testutil/kms`)
- [ ] C. Co-located with tests (per package)
- [ ] D. Shared base + per-package extensions

Notes:

---

### Q19. Integration Test Timeout

Default integration timeout should be:

- [ ] A. 30 seconds
- [ ] B. 60 seconds (per Session 3 Q25)
- [ ] C. 120 seconds
- [ ] D. Configurable with 60s default

Notes:

---

### Q20. Demo Profile CI Isolation

Demo CI jobs should:

- [ ] A. Run in same workflow as unit tests
- [ ] B. Separate workflow triggered on specific paths
- [ ] C. Manual trigger only
- [ ] D. Nightly scheduled + manual trigger

Notes:

---

## Section 5: Phase 6 Migration & Closing (Q21-25)

### Q21. Package Migration Strategy

For `internal/common` → `internal/infra` migration:

- [ ] A. One package at a time, commit between each
- [ ] B. All packages in single commit
- [ ] C. Create aliases first, then migrate
- [ ] D. One package at a time (safer, per instructions)

Notes:

---

### Q22. Import Update Automation

For updating import paths:

- [ ] A. Manual find/replace
- [ ] B. go-importpath-rewriter tool
- [ ] C. Custom script
- [ ] D. IDE refactoring + verification

Notes:

---

### Q23. TLS Package Scope Verification

TLS package (`internal/infra/tls/`) should include:

- [ ] A. Cert generation only
- [ ] B. Cert generation + mTLS helpers
- [ ] C. Full scope (generation + validation + helpers)
- [ ] D. Per Session 5 Q7 decision (full scope)

Notes:

---

### Q24. Acceptance Criteria Verification

Acceptance criteria should be verified by:

- [ ] A. Manual checklist review
- [ ] B. Automated tests for each criterion
- [ ] C. CI job that validates all criteria
- [ ] D. Both automated tests and CI job

Notes:

---

### Q25. Passthru2 Completion Documentation

What documentation is needed at passthru2 close?

- [ ] A. Updated README only
- [ ] B. README + architecture docs
- [ ] C. README + architecture + migration guide
- [ ] D. All above + demo walkthrough

Notes:

---

**Status**: AWAITING ANSWERS (Change [ ] to [x] as applicable and add notes if needed)
