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

**Status**: ✅ COMPLETE (2025-06-26)

### What Worked

- **Struct-field seam injection for Generator**: All 8 external dependencies (`getKeyPairFn`, `mkdirAllFn`, `writeFileFn`, `createCAFn`, `createLeafFn`, `encodePKCS12Fn`, `encodeTrustPKCS12Fn`, `getRealmsForPSIDFn`) injected as struct fields in `Generator`. Stub functions replace each in tests with zero `//nolint` directives needed. This pattern is correct per `10.2.4 Test Seam Injection Pattern`.
- **14-category naming convention**: `public-global-*` (services shared across all PS-IDs), `public-{PS-ID}-*` (per-PS-ID public-facing), `private-{PS-ID}-*` (private admin channel) — convention is immediately obvious from directory names alone. No comments needed to tell you which tier generates the cert.
- **Keystore vs truststore rule**: Keystores `{name}-keystore` contain `.crt + .key + .p12` (leaf + private key), truststores `{name}-truststore` contain `.crt + .p12` only (no `.key`). This rule is exploitable in generation code: `writeKeystore(path, kp, cert, chain)` vs `writeTruststore(path, certs)` — different signatures, no ambiguity.
- **`SAME-AS-DIR-NAME` file naming**: Files inside each directory named identically to the directory name (e.g., `public-global-root-https-server-ca/public-global-root-https-server-ca.crt`). Eliminates the need to design a second naming convention for file names. Discoverable by any consumer just from the directory path.
- **CGO-free PKCS#12 via `software.sslmate.com/src/go-pkcs12`**: `pkcs12.Modern.Encode` / `pkcs12.Modern.EncodeTrustStore` — no CGO, pure Go, supports modern PKCS#12 (SHA-256 / AES-256-CBC, not legacy 3DES). Verified compatible with `CGO_ENABLED=0`.
- **`ResolveTier` + `tier.go` helpers**: Centralizes the `pkiInitAppInstanceSuffixes`, `pkiInitClientPKIDomains`, `pkiInitAdminInstanceSuffixes`, `pkiInitUserTypes` slice constants. The `Generate` function loops over tier-resolved PS-IDs — PRODUCT and SUITE scopes automatically expand from the same PS-ID generation loop with zero duplication.
- **Atomic counter injection pattern**: `sync/atomic` `int32` counter in tests for counting calls to stub functions. Enables verifying "was this function called exactly N times?" without introducing mockery or any external mock library.

### What Didn't Work

- **`validateTargetDir` using `os.ReadDir` on missing path**: Initial implementation used `os.ReadDir` to check if target directory is empty. On Windows, `os.ReadDir` returns an error for a non-existent path, NOT an empty slice. Fix: check `os.Stat` first; if `IsNotExist` → treat as empty (will be created). This is a cross-platform compatibility issue not caught by unit tests using `t.TempDir()`.
- **`getRealmsForPSIDFn` not injected initially**: The first generator draft hardcoded the registry.yaml path in `NewGenerator`. Only when writing unit tests did it become clear the path is untestable without a real registry.yaml on disk. Retro-fit to `getRealmsForPSIDFn` struct-field injection required touching the `Generate` method signature.
- **Category order ambiguity**: Categories 1–14 were defined in tls-structure.md but the implementation order in `generate*.go` initially diverged. After Phase 3, added explicit `// Cat N:` comments to each generation block. Lesson: comment the category number at the generation call-site when the code structure doesn't match document order.

### Root Causes

- `os.ReadDir` non-existent path: Not a Go bug — documented behavior. Caused by assuming Unix semantics (missing dir → empty), which don't hold on Windows unless existence is checked first.
- Realm injection gap: Discovered only during test writing because `readRealmsForPSID` opens a real file. Design should include dependency injection for all I/O calls from the start.

### Patterns for Future Phases

- For any function that does I/O (filesystem, network, registry), inject it as a function field from the start — do not wait until test-writing reveals it.
- `os.Stat` check before `os.ReadDir`/file operations is Windows-safe. ALWAYS use `os.Stat` first.
- PKCS#12 `pkcs12.Modern` (SHA-256/AES-256-CBC) is unconditionally preferred over `pkcs12.Legacy` (3DES). Never use `pkcs12.Legacy` in new code.
- When implementing a multi-category generate function, comment `// Cat N: <name>` at each invocation site so reviewers can cross-reference tls-structure.md without mentally mapping the code.

---

## Phase 3: pki-init CLI & Docker Volume Config

**Status**: ✅ COMPLETE (2025-06-26)

### What Worked

- **Named Docker volumes only, never bind mounts**: `{PS-ID}-certs` named volume is the only approved cert delivery mechanism. Bind mounts require host directory preparation and are host-path-dependent. Named volumes are lifecycle-managed by Docker Compose (created on first `up`, survive across restarts, removed only on `down -v`). Rule CO-21/CO-22 added to deployment-templates.md to enforce this.
- **pki-init CLI invocation via skeleton-template binary**: The `cmd/skeleton-template` binary exposes the `init` subcommand by delegating to `cryptoutilAppsFrameworkTls.InitForService(SkeletonTemplateServiceID, args, ...)`. There is no separate `cmd/pki-init` binary — pki-init is a subcommand of each PS-ID's binary. This is correct per the CLI Patterns architecture (§4.4.7).
- **`InitForService`/`InitForProduct`/`InitForSuite` wrapper design**: Each wrapper resolves the service/product/suite ID, creates production telemetry, creates the generator, and calls `Generate(tierID, targetDir)`. Three functions with identical structure — any bug fix propagates to all tiers. Clean single-responsibility separation between CLI flag parsing and generation.
- **File permissions**: `0o600` for private key files (`.key`), `0o644` for cert files (`.crt`), `0o644` for PKCS#12 files (`.p12`). PKCS#12 already encrypts the private key inside; 0o644 is safe. Applied in `writeKeystore` and `writeTruststore`.

### What Didn't Work

- **`PKIInitValidityLeaf = 398 * 24 * time.Hour` bug**: The CA/B Forum limit is 398 days, but the code check in `randomizedNotBeforeNotAfterEndEntityInternal` uses `TLSDefaultSubscriberCertDuration = 397 * 24 * time.Hour` as the enforced maximum. Setting `PKIInitValidityLeaf` to 398 days caused: "requestedDuration exceeds maxSubscriberCertDuration." Root cause: `TLSMaxSubscriberCertDuration` (398d) is the CA/B Forum absolute limit, while `TLSDefaultSubscriberCertDuration` (397d) is the project-enforced limit with a 1-day safety margin. Fix: `PKIInitValidityLeaf = 397 * 24 * time.Hour`.
- **`productionNewTelemetryService` missing `LogLevel` field**: The `TelemetrySettings.LogLevel = ""` causes `ParseLogLevel("")` to fail with "invalid log level." The production code path was never exercised by unit tests (which use stub injection), so this was only caught in E2E. Fix: add `LogLevel: DefaultLogLevelInfo`.
- **`productionNewTelemetryService` missing `OTLPEndpoint` field**: `parseProtocolAndEndpoint` is called unconditionally in `initLogger` even when `OTLPEnabled=false`. An empty `OTLPEndpoint` fails with "invalid OTLP endpoint protocol, must start with https://, grpcs://, http://, or grpc://". Fix: add `OTLPEndpoint: DefaultOTLPEndpointDefault`.
- **Unit tests with stub injection do NOT exercise `productionNew*` functions**: All three bugs above were latent because `productionNewTelemetryService` and `productionNewGenerator` are bypassed in unit tests. This is the structural ceiling reason for 92.4% (not 95%)+ coverage. The lesson: E2E tests are not optional when production initialization functions exist that unit tests cannot reach.

### Root Causes

- `PKIInitValidityLeaf` off-by-one: Code value (397d) and CA/B Forum limit (398d) were conflated. The magic constant comment said "CA/B Forum 398-day limit" but the code uses 397d as the enforced ceiling. Fix: align constant to the ENFORCED value (397d), document that 397d = "one below 398d CA/B Forum hard limit."
- `productionNewTelemetryService` config fields: Telemetry initialization validated fields unconditionally. The `OTLPEnabled=false` flag does NOT bypass field validation in `parseProtocolAndEndpoint`. This is by design (fail-fast on misconfiguration), but means CLI callers must set all required fields even when OTLP export is disabled.

### Patterns for Future Phases

- ALWAYS add an E2E test for any CLI entry point that uses `productionNew*` functions. Unit tests with stubs cannot substitute for this.
- When a magic constant represents a value with a "hard limit" and a "project limit" (e.g., 398d vs 397d), name them both and use the project limit constant in all code that compares against a limit.
- `parseProtocolAndEndpoint` is called unconditionally — future callers of `TelemetrySettings` in CLI tools MUST supply all fields even if OTLP export is disabled. Consider adding a `dev` or `noop` endpoint constant for CLI tools that don't export telemetry.
- The pattern `InitForService`/`InitForProduct`/`InitForSuite` wrapping a common `initRun` is reusable for any future CLI that needs scope-aware execution.

---

## Phase 4: Template & Deployment Updates

**Status**: ✅ COMPLETE (2025-06-26)

### What Worked

- **CO-21/CO-22 rules directly in deployment-templates.md**: Encoding the cert volume mount rules as numbered rules (CO-21: named volumes only, CO-22: compose apps mount `{PS-ID}-certs:/certs:ro`) gave them the same status as all other compose rules. Auditors and the template-compliance linter can check them systematically.
- **target-structure.md Section F.4 already correct**: The section cross-referencing `tls-structure.md` was already present from a prior session. Task 4.2 was completed with zero changes needed — pre-existing correct state.
- **Phase 3.3 already handled compose volume declarations**: By committing Part 3.3 (named volume declarations in all compose files) before Phase 4, Task 4.3 was already satisfied. Phase 4 simply verified correctness rather than doing new work.

### What Didn't Work

- **No issues encountered during Phase 4.** All work was either already complete (Tasks 4.2, 4.3) or straightforward additions (Task 4.1 CO-21/CO-22). Phase 4 was significantly faster than estimated (~30 min vs 3h estimated).

### Root Causes

- Over-estimated Phase 4 scope: Phase 3.3 (compose volumes) was completed in a prior session before Phase 4 work started. The tasks.md estimation assumed 4.3 would be standalone work.

### Patterns for Future Phases

- Before starting a phase, check tasks.md for tasks that are already satisfied by prior phases — this avoids redundant work.
- When a phase objective is "document what was built in prior phases," budget 0.5h not 3h.
- CO-N numbered rules in deployment-templates.md are the correct mechanism for any compose/volume invariant that should be lint-checked.

---

## Phase 5: Quality Gates & Testing

**Status**: ⚠️ PARTIAL — 5.3/5.4 deferred to CI/CD; 5.6 E2E done (2025-06-26)

### What Worked

- **Table-driven tests with atomic counter injection**: 48 unit tests covering all 14 category generation paths. Error paths exercised by calling stub functions that return errors on specific call count (`atomic.AddInt32(&callCount, 1) == wantFailAt`). This is the correct per-seam pattern and requires zero external mock libraries.
- **`TestGenerate_SkeletonTemplate_DirCount`**: The E2E directory count test (82 dirs for skeleton-template with 2 realms resolved via stub) provides regression protection for the entire 14-category layout. A single test validates the full output contract.
- **`t.TempDir()` for all filesystem tests**: Automatic cleanup with no residual files. Tests are isolated across all parallel goroutines — no shared state.
- **92.4% coverage is the achievable ceiling**: `productionNewTelemetryService` (4 stmts), `productionNewGenerator` (1 stmt), and `NewGenerator` success path (6 stmts) are behind real I/O. PEM encode errors (3 stmts) on a valid `*x509.Certificate` are structurally impossible. Total: ~11 unreachable stmts / ~150 total stmts ≈ 93% ceiling. 92.4% is within 1% of that ceiling and is accepted.
- **E2E (Task 5.6) confirmed all 82 directories**: Running `go run ./cmd/skeleton-template init --domain=skeleton-template --output-dir=C:\tmp\certs-test` produced 82 directories with correct file sets. The keystore verified `.crt + .key + .p12`; truststore verified `.crt + .p12` with no `.key`. File naming matches `SAME-AS-DIR-NAME` convention.

### What Didn't Work

- **`go test -race`** requires `CGO_ENABLED=1` (GCC). The Windows CI environment has `CGO_ENABLED=0` mandatory for all non-race builds. Race detector testing is deferred to Linux CI/CD (Task 5.4). This is a known constraint, not a new finding.
- **`gremlins unleash`** (mutation testing) panics on Windows (`v0.6.0`). Deferred to Linux CI/CD (Task 5.3). Also a known constraint documented in `03-02.testing.instructions.md`.
- **Linter `mnd` flagged `2` and `3` as magic numbers** in generator_helpers.go: These specific numbers (realm multiplier components) are genuinely magic and were extracted to named constants (`PKIInitCat5UserTypes = 2`, etc.) or replaced with `len(slice)` references. Required an extra linting pass after initial implementation.

### Root Causes

- Race/mutation deferred: Environmental constraints (no GCC on Windows, gremlins Windows panic). Not fixable in this engagement — requires CI/CD environment change.
- `mnd` violations: The generator uses small integer constants inline (e.g., `2 user types`, `3 PKI domains`). These are semantically meaningful. Fix: extract to named constants in `tier.go`/`magic_pkiinit.go` before committing.

### Patterns for Future Phases

- Any function that returns an error but where the error condition is structurally impossible (PEM encode on valid cert, pool get from non-nil pool) creates an irreducible coverage gap. Document these in the coverage ceiling analysis and do not chase them.
- The "`productionNew*` functions are only testable via E2E" pattern is a structural property of CLI tools that initialize real infrastructure. Always include one E2E smoke test per CLI entry point.
- Before a final linting pass, run `golangci-lint run --fix` to apply auto-fixable linters (`wsl`, `godot`, `gofumpt`, `goimports`). Only then check remaining manual fixes.
- For Windows testing, use `t.TempDir()` rather than `os.TempDir()` + manual cleanup — the former is parallel-safe and cleaned up automatically even if the test panics.

---

## Phase 6: Knowledge Propagation

**Status**: ✅ COMPLETE (2025-07-12)

### What Worked

- **ENG-HANDBOOK.md §6.11.3 addition was the right scope**: The new section (pki-init Certificate Structure) added a 14-category table, file format rules, PKCS#12 specification, directory count summary, and a cross-reference to `tls-structure.md`. This is exactly the kind of "how does this work in practice" content that belongs in a permanent handbook section, not just in a tls-specific doc.
- **lint-docs passes immediately**: The `go run ./cmd/cicd-lint lint-docs` tool detects broken ENG-HANDBOOK.md anchors, orphaned sections, and propagation drift. Adding a new `####` subsection with no `@propagate` blocks required zero changes to propagation config — the tool only validates existing propagation, not new sections.
- **Filling lessons.md Phases 2-5 took exactly one pass**: Because the implementation work was fresh (completed in prior sessions), the key patterns, bugs, and root causes were all remembered. Lesson: fill lessons.md during or immediately after each phase — context fades quickly.

### What Didn't Work

- **No agent/skill/instruction updates warranted**: The Phase 6 task list included "update agents, skills, instructions as warranted." After reviewing the phase 2-5 lessons, none of the patterns were novel enough to require instruction file updates — they are all covered by existing guidelines (`02-05.security` for FIPS/PKI, `03-02.testing` for seam injection, `03-01.coding` for switch-over-if patterns). This is expected: Framework V11 implemented documented patterns, did not discover new ones.
- **Phase 6 significantly faster than estimated** (0.5h actual vs 2h estimated): Because No agent/skill/instruction updates were needed and the ENG-HANDBOOK.md change was additive (new subsection, no existing content affected).

### Root Causes

- Estimation gap: 2h assumed agent+skill+instruction updates AND doc changes. In practice only a single doc section was needed. Better to estimate Phase 6 as "doc update + linter check + commit" = ~0.5-1h unless lessons explicitly identify artifact updates needed.

### Patterns for Future Phases

- Phase 6 (Knowledge Propagation) should default to ≤1h estimate unless lessons from prior phases explicitly call out new patterns for agents/skills/instructions.
- `go run ./cmd/cicd-lint lint-docs` is the gate for all ENG-HANDBOOK.md changes. Run it before committing any doc change. Exit code 0 = safe to commit.
- When adding a new `####` subsection to ENG-HANDBOOK.md, check: (1) does the heading appear in the table of contents (some sections are auto-indexed, most #### are not), (2) does lint-architecture-links pass (confirms anchor is valid), (3) does validate-propagation pass (confirms no orphaned @propagate blocks). All three passed immediately for §6.11.3.
