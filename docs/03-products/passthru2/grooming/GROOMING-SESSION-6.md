# Passthru2 Grooming Session 6: Implementation Validation & Final Decisions

**Purpose**: Validate implementation decisions from Session 5, resolve technical ambiguities for remaining phases.
**Created**: 2025-12-01
**Status**: AWAITING ANSWERS

---

## Session 5 Analysis Summary

### Key Decisions from Session 5

| Topic | Decision | Implications |
|-------|----------|--------------|
| **Exit Codes** | Simple 0/1/2 pattern | 0=success, 1=partial failure, 2=failure |
| **HSM Placeholder** | Design extensible + placeholder pkg | `internal/infra/tls/hsm/` for future PKCS#11 |
| **System Trust Store** | Feature flag (disabled) | Design now, enable later |
| **Auth Modes** | Both in parallel | Bearer + Basic different code paths |
| **TLS Client Tenant** | URI SAN option + alternatives | Need further exploration |
| **Phase 0 Priority** | P0.6 → P0.10 → P0.1 → P0.5 | Demo seed first, then TLS, then telemetry |
| **TLS Package Scope** | Full scope | All helpers + mTLS + validation |
| **Config Strictness** | Same as production | No relaxed demo mode |
| **mTLS Rotation** | FS watcher → Admin API → SIGHUP | Multiple signal options |
| **PBKDF2** | SHA-256, 600K, 32-byte salt | OWASP 2024 defaults |
| **User Metadata** | Schema + required keys | Working schema for validation |
| **Tenant Isolation** | Schema-per-tenant | SQLite + PostgreSQL compatible |
| **CLI Colors** | All options | Windows ANSI + CI disable + flag |
| **Docs Format** | All formats | Markdown + Mermaid + screenshots |
| **API Docs** | Update if API changes | Observe emergent design |
| **Coverage Exclusions** | Generated code only | api/client, api/server |
| **Benchmark Storage** | JSON + Go bench + SQLite | All formats |
| **Error Format** | RFC 7807 Problem Details | Standardized |
| **Scope** | Open for demo improvements | Additions OK |
| **Implementation** | Sequential | Clean commits, checkpoints |
| **Breaking Changes** | Irrelevant | Unreleased project |
| **Success Metrics** | Criteria + 80% coverage | Acceptance criteria met |

---

## Section 1: Tenant Representation in Client Certificates (Q1-5)

From Session 5 Q5, you noted concerns about using SAN for tenant representation.

### Q1. Tenant in Certificate - Distinguished Name Extension

Instead of SAN, should tenant ID be in a custom X.509 extension in the Subject DN?

- [ ] A. Custom OID extension (e.g., `1.3.6.1.4.1.XXXXX.1.1` = tenant ID)
- [ ] B. Use existing standardized extension (serialNumber, UID)
- [ ] C. Embed in Organization Unit (OU=tenant:uuid)
- [ ] D. Defer to runtime header validation (cert auth only proves identity)

Notes:

---

### Q2. Certificate vs Runtime Tenant Binding

Should tenant binding be:

- [ ] A. Statically in certificate (tenant is cryptographic assertion)
- [ ] B. Dynamically at runtime (cert proves who, header/token proves tenant)
- [ ] C. Hybrid (cert can have tenant claim, but runtime can override)
- [ ] D. Configuration-based (operator decides per deployment)

Notes:

---

### Q3. Multi-Tenant Client Certificates

Should a single client certificate support access to multiple tenants?

- [ ] A. No - one cert per tenant (strict binding)
- [ ] B. Yes - cert proves identity, separate authz for tenant access
- [ ] C. Yes - use extension with list of allowed tenant IDs
- [ ] D. Configurable per deployment

Notes:

---

### Q4. Certificate Tenant Claim Validation

If tenant is in certificate, how should it be validated?

- [ ] A. Direct match against request context
- [ ] B. CA-signed assertion (CA attests tenant binding)
- [ ] C. Cross-reference with external directory
- [ ] D. Multiple strategies, configurable

Notes:

---

### Q5. Federation Tenant Resolution

When federated to Identity, how should tenant be determined from JWT?

- [ ] A. Standard `tenant_id` claim
- [ ] B. Part of `aud` claim (e.g., `urn:tenant:uuid`)
- [ ] C. Part of `iss` claim (issuer per tenant)
- [ ] D. Custom claim with configurable name

Notes:

---

## Section 2: Demo CLI Architecture Details (Q6-10)

### Q6. Demo CLI Package Structure

How should the demo CLI be structured?

- [ ] A. `cmd/demo/` with subpackages for each product
- [ ] B. `cmd/demo/` flat structure with separate files
- [ ] C. `internal/demo/` library + `cmd/demo/` thin wrapper
- [ ] D. `cmd/demo-<product>/` separate binaries

Notes:

---

### Q7. Demo CLI Health Check Strategy

For health check waiting, should we:

- [ ] A. Poll endpoints repeatedly until success
- [ ] B. Use Docker health check status via Docker API
- [ ] C. Combination (Docker health + endpoint verification)
- [ ] D. Only endpoint verification (container-agnostic)

Notes:

---

### Q8. Demo CLI Output Streaming

For progress display, should output:

- [ ] A. Buffer until completion then display summary
- [ ] B. Stream in real-time with spinners/progress
- [ ] C. Configurable (quiet/normal/verbose modes)
- [ ] D. Default real-time, `--quiet` for buffered

Notes:

---

### Q9. Demo CLI Error Recovery

On demo step failure, should we:

- [ ] A. Skip failed step, continue with rest
- [ ] B. Retry failed step N times before skipping
- [ ] C. Fail fast (stop on first error)
- [ ] D. Configurable per step or globally

Notes:

---

### Q10. Demo Data Consistency

After demo setup, how should we verify consistency?

- [ ] A. Query all created entities and validate
- [ ] B. Run predefined test operations (encrypt/decrypt/sign)
- [ ] C. Check only existence (IDs present in responses)
- [ ] D. Full validation including functional tests

Notes:

---

## Section 3: Phase 1-2 Implementation Details (Q11-15)

### Q11. KMS --demo Flag Behavior

What should `--demo` flag enable?

- [ ] A. Auto-seed data only
- [ ] B. Relaxed security settings for demo
- [ ] C. Auto-seed + verbose logging + relaxed timeouts
- [ ] D. Auto-seed only, no security relaxation (Session 5 Q8)

Notes:

---

### Q12. Identity --demo Flag Behavior

What should Identity's `--demo` flag enable?

- [ ] A. Same as KMS (auto-seed, no security changes)
- [ ] B. Pre-configured OAuth clients with known secrets
- [ ] C. Both A and B
- [ ] D. Additional: disable PKCE requirement for demo flows

Notes:

---

### Q13. Demo User Password Strategy

Should demo passwords be:

- [ ] A. Same for all users (e.g., `demo-password`)
- [ ] B. Predictable pattern (e.g., `{username}-password`)
- [ ] C. Documented in config but varied
- [ ] D. Generated but logged at startup

Notes:

---

### Q14. Demo Client Secret Strategy

Should demo client secrets be:

- [ ] A. Static, documented values
- [ ] B. Generated, stored in Docker secrets
- [ ] C. Static for public clients, generated for confidential
- [ ] D. All generated, all in Docker secrets

Notes:

---

### Q15. Key Pool Auto-Seeding

What key pools should be auto-seeded for demo?

- [ ] A. Minimal: one encryption pool, one signing pool
- [ ] B. Comprehensive: encryption + signing + MAC for each algorithm family
- [ ] C. Match what's documented in existing demos
- [ ] D. Configurable via demo-seed.yml

Notes:

---

## Section 4: Phase 3-4 Integration Details (Q16-20)

### Q16. JWKS Cache Implementation

For JWKS caching, should we:

- [ ] A. Use go-jose/v4 built-in fetching
- [ ] B. Implement custom cache with configurable TTL
- [ ] C. Use external library (e.g., MicahParks/keyfunc)
- [ ] D. Simple in-memory cache, refresh on 401

Notes:

---

### Q17. Token Introspection Batching

Should introspection support batching?

- [ ] A. Yes, batch multiple tokens in single request
- [ ] B. No, individual requests only (simpler)
- [ ] C. Optional batching (single endpoint, array input)
- [ ] D. Depends on Identity implementation

Notes:

---

### Q18. Realm File vs DB Precedence

When both file and DB realms are enabled, what takes precedence?

- [ ] A. File always wins (override)
- [ ] B. DB always wins (authoritative)
- [ ] C. First match wins (file checked first)
- [ ] D. Configurable priority order

Notes:

---

### Q19. Federation Token Trust

For federated Identity validation:

- [ ] A. Trust any token from configured issuers
- [ ] B. Require specific audience claim
- [ ] C. Validate both issuer and audience
- [ ] D. Additional claim requirements configurable

Notes:

---

### Q20. Multi-Tenant Database Schema Naming

For schema-per-tenant isolation, naming convention:

- [ ] A. `tenant_<uuid>` (e.g., `tenant_abc123...`)
- [ ] B. `t_<short_id>` (e.g., `t_abc123`)
- [ ] C. Custom prefix configurable (e.g., `kms_tenant_<id>`)
- [ ] D. Hash-based for privacy (e.g., `t_<hash(uuid)>`)

Notes:

---

## Section 5: Final Implementation Questions (Q21-25)

### Q21. Coverage Reporting Integration

For coverage tracking, should we:

- [ ] A. Native Go coverage only (go test -cover)
- [ ] B. Integrate with Codecov/Coveralls
- [ ] C. HTML reports in test-reports/
- [ ] D. All of the above

Notes:

---

### Q22. Benchmark Regression Threshold

What regression threshold should fail CI?

- [ ] A. Any regression (strict)
- [ ] B. >5% regression (reasonable noise)
- [ ] C. >10% regression (significant only)
- [ ] D. Configurable threshold

Notes:

---

### Q23. Documentation Update Strategy

As implementation progresses, docs should be updated:

- [ ] A. After each phase completion
- [ ] B. Continuously with each commit
- [ ] C. Only at major milestones
- [ ] D. At end of passthru2

Notes:

---

### Q24. Test Isolation Implementation

For test isolation, prefer:

- [ ] A. UUIDv7 prefixes in test data
- [ ] B. Transaction rollback per test
- [ ] C. Both (belt and suspenders)
- [ ] D. Separate test databases per suite

Notes:

---

### Q25. Passthru2 Completion Criteria

Beyond acceptance criteria, what signals completion?

- [ ] A. All tasks checked in TASK-LIST.md
- [ ] B. All CI workflows passing
- [ ] C. Demo runs successfully end-to-end
- [ ] D. All of the above

Notes:

---

**Status**: AWAITING ANSWERS (Change [ ] to [x] as applicable and add notes if needed)
