# Passthru2 Grooming Session 5: Session 4 Analysis & Implementation Readiness

**Purpose**: Analyze Session 4 decisions, resolve remaining ambiguities, and confirm implementation readiness.
**Created**: 2025-11-30
**Status**: AWAITING ANSWERS

---

## Session 4 Analysis Summary

### Key Decisions Extracted from Session 4

| Topic | Decision | Implications |
|-------|----------|--------------|
| **TLS Deps** | Std lib + x/crypto, minimal now | No ACME in passthru2 |
| **Cert Storage** | PEM + PKCS#12 configurable | Support both formats, PEM default |
| **Root CA** | Custom CA only (99% use case) | Isolated demo environment |
| **TLS Validation** | ALWAYS full validation | No relaxed modes even in demo |
| **TLS Version** | TLS 1.3 only | No TLS 1.2 fallback |
| **UUIDv4 Gen** | Match existing v7 pattern | Use keygen package consistency |
| **Tenant ID Strict** | UUID format only | No hyphens normalization needed |
| **Demo Tenants** | Random per startup | Regenerated each run |
| **Tenant Header** | Authorization linked | Never path/query params |
| **Error Aggregation** | Structured + tree | apperr-style nested errors |
| **Partial Success** | Report + keep running + configurable | Graceful degradation |
| **Retry Strategy** | Configurable | Exponential backoff option |
| **Progress Display** | Counter + spinner | Dual progress indicators |
| **Exit Codes** | sysexits.h or 0/1/2 | Need clarification |
| **Benchmarks** | Compare previous + CI regression | Store baseline locally |
| **Test Fixtures** | testdata/ directories | Standard Go convention |
| **Test Isolation** | Tx rollback + UUIDv7 prefix | Dual isolation strategy |
| **Config Format** | YAML primary | Converter for others if needed |
| **Config Validation** | Load + startup | Fail fast on bad config |
| **Config Location** | Standard paths | /etc, ~/.config search order |
| **Hot Reload** | Defer (maybe C later) | Not for passthru2 |
| **Compose Profiles** | dev, demo, ci + prod | Include production template |

---

## Section 1: Clarifications Needed (Q1-5)

### Q1. Exit Code Strategy Clarification

You mentioned "sysexits.h or D" - sysexits.h is a Unix convention with codes like:

- 64: EX_USAGE (command line usage error)
- 65: EX_DATAERR (data format error)
- 66: EX_NOINPUT (cannot open input)
- 70: EX_SOFTWARE (internal software error)
- 73: EX_CANTCREAT (can't create output file)
- 74: EX_IOERR (input/output error)

Which do you prefer?

- [x] A. Simple (0=success, 1=partial, 2=failure)
- [ ] B. sysexits.h compatible for detailed diagnostics
- [ ] C. Hybrid (0/1/2 for categories, sysexits for specific failures)
- [ ] D. Custom enum in magic package

Notes:

---

### Q2. PKCS#11/YubiKey Future Support

You mentioned PKCS#11 and YubiKey support needed in future. Should we:

- [ ] A. Add interface stubs now for future implementation
- [x] B. Design cert storage API to be extensible
- [ ] C. Just document as future work, no code changes
- [x] D. Create `internal/infra/tls/hsm/` placeholder package

Notes:

---

### Q3. System Trust Store Addition

You mentioned "Maybe in future I would support adding system trust store for HTTPS Server front-end UI and CLI certs". Should we:

- [x] A. Design for this now with feature flag (disabled)
- [ ] B. Defer entirely to future passthru
- [ ] C. Add config option now, implement later
- [ ] D. Document in architecture decisions only

Notes:

---

### Q4. Bearer vs Basic Auth Priority

You specified multiple auth modes. For initial implementation priority:

- [ ] A. Bearer (JWT) first, Basic later
- [ ] B. Basic (realm) first, Bearer (federated) second
- [x] C. Both in parallel (different code paths)
- [ ] D. Depends on demo being showcased

Notes:
KMS server supports service-to-service API and browser-to-service UI/API.
Service-to-service authn could be token (bearer) or clientid/clientsecret (basic).
Browser-to-service UI/API must be session cookie (Cookie header), except for the initial authn: File/DB realms, or Delegate to Authz service.

---

### Q5. TLS Client Custom SAN Extension

You mentioned "TLS Client: Custom SAN extension (uri? other? need to consider options...)". Options:

- [x] A. URI SAN with tenant ID (e.g., `urn:tenant:uuid`)
- [ ] B. Custom OID extension (non-standard)
- [ ] C. DNS SAN with tenant prefix (e.g., `tenant1.kms.local`)
- [ ] D. Defer to Phase 4 realm implementation

Notes:
I have been thinking about it, and not sure if SAN is a good idea.
TLS Client cert Subject is a Client or User, whereas Tenant is an attribute of the Client or User.
Since SAN means Subject Alt Name, then SAN should describe the Client or User, not a Tenant.
I think I want you to suggest different ideas for representing a Tenant in a Subject certificate.
What best practices exist in the real world?
The only one I can think of is maybe LDAP, but I don't want to restrict Distinguished Names to conform to LDAP RDNs/AVAs, DNs should be more free-form.

---

## Section 2: Implementation Priorities (Q6-10)

### Q6. Phase 0 First Task

Which Phase 0 task should be implemented first?

- [2] A. P0.10: Create `internal/infra/tls/` package (foundation for all TLS)
- [3] B. P0.1: Extract telemetry to shared compose
- [1] C. P0.6: Add demo seed data for KMS
- [4] D. P0.5: Create compose profiles

Notes:

---

### Q7. TLS Package Scope

What should `internal/infra/tls/` initially include?

- [ ] A. CA chain generation only
- [ ] B. CA chain + server/client cert generation
- [ ] C. CA chain + certs + TLS config helpers
- [x] D. Full: CA chain + certs + config + mTLS helpers + validation

Notes:

---

### Q8. Config Validation Strictness

For demo mode, should config validation be:

- [x] A. Same strictness as production
- [ ] B. Relaxed (warnings instead of errors for non-critical issues)
- [ ] C. Minimal (only check required fields)
- [ ] D. Configurable per profile

Notes:

---

### Q9. Demo Compose Dependency Order

What should be the service startup order in demo compose?

- [ ] A. Telemetry → DB → Identity → KMS
- [ ] B. DB → Telemetry → Identity → KMS (DB first always)
- [ ] C. Telemetry → DB (parallel) → Identity → KMS
- [ ] D. All parallel with health check dependencies

Notes:
None of the above.
Use the KMS compose.yml order.

---

### Q10. Coverage Baseline

Before starting implementation, should we establish a coverage baseline?

- [ ] A. Yes, run coverage now and document current state
- [ ] B. No, just track from first commit
- [ ] C. Yes, and create coverage trend file
- [x] D. Coverage tracking starts after Phase 0

Notes:

---

## Section 3: Technical Deep Dive (Q11-15)

### Q11. mTLS Certificate Rotation Signal

When mTLS cert rotation is eventually implemented, how should services be notified?

- [3] A. SIGHUP signal (Unix convention)
- [2] B. Admin API endpoint (`POST /admin/reload-certs`)
- [1] C. File system watcher (inotify-style)
- [ ] D. All of the above, defer implementation

Notes:

---

### Q12. PBKDF2 Default Parameters

What should the default PBKDF2 parameters be?

- [x] A. SHA-256, 600,000 iterations, 32-byte salt (OWASP 2024)
- [ ] B. SHA-512, 210,000 iterations, 32-byte salt (OWASP 2024)
- [ ] C. SHA-256, 100,000 iterations, 16-byte salt (faster for demo)
- [ ] D. Configurable with OWASP defaults

Notes:

---

### Q13. Realm User Metadata Schema

For the extensible JSON metadata field in realm users, should we:

- [ ] A. Leave fully unstructured (any JSON)
- [x] B. Define optional schema with validation
- [x] C. Require specific top-level keys (e.g., `custom`, `attributes`)
- [ ] D. Just store raw JSON, no validation

Notes:
Make sure to create a working schema file, which can be used to validate realm config,
for easy detection and reporting of errors back to admins and LLM agents.

---

### Q14. Tenant Isolation Implementation

For database-level tenant isolation, should we:

- [x] A. Separate PostgreSQL schemas per tenant
- [ ] B. Separate PostgreSQL databases per tenant
- [ ] C. Schema-per-tenant with connection pool per tenant
- [ ] D. Defer tenant isolation to Phase 4

Notes:
A is possible for both SQLite and PostgreSQL using schema.table syntax:

- PostgreSQL: CREATE SCHEMA tenant1; CREATE TABLE tenant1.users
- SQLite: ATTACH 'tenant1.db' AS tenant1; CREATE TABLE tenant1.users
- Same SQL interface: SELECT * FROM tenant1.users
- Provides tenant isolation while maintaining code uniformity
- Scales better than separate databases (connection pooling, etc.)

---

### Q15. Demo CLI Color Output

Should demo CLI color output work on Windows?

- [ ] A. Yes, use library that handles Windows ANSI (e.g., fatih/color)
- [ ] B. Yes, but disable colors in CI automatically
- [ ] C. Yes, with `--no-color` flag for CI/logging
- [x] D. All of the above

Notes:

---

## Section 4: Documentation & Testing (Q16-20)

### Q16. Demo Documentation Format

What format should demo documentation use?

- [ ] A. Markdown with code blocks
- [ ] B. Markdown + embedded diagrams (Mermaid)
- [ ] C. Markdown + screenshots
- [x] D. All of the above

Notes:

---

### Q17. API Documentation

Should OpenAPI specs be updated as part of passthru2?

- [ ] A. Yes, ensure all endpoints documented
- [ ] B. Yes, add examples for demo flows
- [ ] C. No, defer to separate docs task
- [x] D. Only update if implementation changes API

Notes:
I am unsure how to answer. It might require observation to determine emergent design while you implement other changes.

---

### Q18. Test Coverage Exclusions

What files should be excluded from coverage requirements?

- [x] A. Only generated code (api/client, api/server)
- [ ] B. + Test utilities and testdata
- [ ] C. + Demo/example binaries
- [ ] D. Explicit list in codecov.yml or similar

Notes:

---

### Q19. Benchmark Storage Format

How should benchmark baselines be stored?

- [x] A. JSON file (machine-readable)
- [x] B. Go test benchmark output format
- [x] C. SQLite database (trend analysis)
- [ ] D. Simple text file with key metrics

Notes:

---

### Q20. Error Message Consistency

Should error messages follow a specific pattern?

- [ ] A. Match existing apperr patterns exactly
- [x] B. Update to RFC 7807 Problem Details style
- [ ] C. Simple descriptive messages (no formal structure)
- [ ] D. Structured internally, human-readable externally

Notes:

---

## Section 5: Final Confirmations (Q21-25)

### Q21. Passthru2 Scope Lock

Is the scope from Sessions 1-4 now locked, or can new features be added?

- [ ] A. Scope locked - only bug fixes and clarifications
- [ ] B. Minor additions OK if they don't delay timeline
- [x] C. Open to additions that improve demo experience
- [ ] D. Strict scope lock - anything else goes to passthru3

Notes:

---

### Q22. Implementation Parallelism

Can multiple phases be worked on in parallel?

- [ ] A. Yes, all phases can progress together
- [ ] B. Phase 0 must complete before others start
- [ ] C. Phase 0 and 1 can parallel, 2+ must wait
- [x] D. Sequential only for clean commits

Notes:
I don't really have a strong opinion here. Sequential is easier for human readability,
but passthru2 is going to be LLM Agent heavy. I only care about final results,
and having milestones or checkpoints where I can do spot checks to validate that
progress is tracking in the direction I want it to go.

---

### Q23. Breaking Change Documentation

If breaking changes occur, how should they be documented?

- [ ] A. In CHANGELOG.md only
- [ ] B. In README.md migration section
- [ ] C. Separate MIGRATION.md file
- [ ] D. All of the above

Notes:
IMPORTANT Breaking changes an completely irrelevant. This project has never been released.
No backwards compatibility or migration is necessary at all.
Make sure this is clear in the 03-products docs, as it might save time from being
wasted on worrying about migration. There is no such thing as migration, this is
an unrelated project.

---

### Q24. Demo Video/Recording

Should passthru2 deliverables include demo recordings?

- [ ] A. Yes, short video for each demo flow
- [ ] B. Yes, GIF animations in docs
- [ ] C. No, documentation sufficient
- [x] D. Only if time permits

Notes:

---

### Q25. Success Metrics

How should passthru2 success be measured?

- [x] A. All acceptance criteria met (from Task List)
- [x] B. Coverage targets achieved (80%+)
- [ ] C. Demo runs in <30 seconds from clone
- [ ] D. All of the above

Notes:

---

**Status**: AWAITING ANSWERS (Change [ ] to [x] as applicable and add notes if needed)
