# Quiz Me - Framework v2

**Purpose**: Key decisions needed from the user before implementation.
**After answering**: Merge into plan.md/tasks.md, then delete this file.

---

## Question 1: Skeleton-Template Purpose

**Question**: Skeleton-template exists as both a working service AND a reference implementation. lint-fitness validates architecture patterns across all 10 services. Is skeleton-template redundant with lint-fitness, or does it serve a distinct purpose?

**A)** Skeleton-template is redundant with lint-fitness. Remove it and rely on lint-fitness for pattern enforcement.
**B)** Skeleton-template is a reference implementation for HUMANS (copy-paste starting point). lint-fitness is for MACHINES (automated validation). Both serve distinct purposes.
**C)** Merge skeleton-template INTO lint-fitness as test fixtures. No standalone service needed.
**D)** Skeleton-template should be a minimal smoke test, not a full service. Reduce to essentials.
**E)**

**Answer**:

**Rationale**: Determines Phase 6 scope (Task 6.3) and whether skeleton-template gets full investment or deprecation.

---

## Question 2: lint-fitness Value vs Cost

**Question**: lint-fitness is 10,500 lines across 23 sub-linters. Some sub-linters use real-file scanning (good) and some use synthetic test content (questionable). What is the right investment level?

**A)** Full investment: 98% coverage, 95% mutation, all 23 sub-linters maintained as-is.
**B)** Selective investment: Keep the high-value sub-linters (port validation, naming, kebab-case), prune low-value ones.
**C)** Convert synthetic-content sub-linters to real-file validators (scan actual codebase files, not test fixtures).
**D)** Freeze lint-fitness at current state. Focus effort on semgrep rules instead (more standard, less custom code).
**E)**

**Answer**:

**Rationale**: 10,500 lines is substantial custom tooling. Determines Phase 6 scope and ongoing maintenance burden.

---

## Question 3: Builder Refactoring Approach

**Question**: Currently services call multiple With*() builder methods. The goal is to make product-services as thin as Spring Boot @SpringBootApplication. What builder API shape do you prefer?

**A)** Single domain config struct: `NewServerBuilder(ctx, templateCfg).WithDomain(domainCfg).Build()` - one struct per service.
**B)** Keep fluent builder but make ALL calls optional with sensible defaults. Only WithDomainMigrations and WithPublicRouteRegistration remain.
**C)** Template provides a `RunService(ctx, domainCfg)` one-liner. Services have zero builder code.
**D)** Convention-based: Builder auto-discovers domain config from well-known file paths. Zero explicit calls needed.
**E)**

**Answer**:

**Rationale**: Determines Phase 3 scope. Affects all 10 services and future service creation via `/new-service` skill.

---

## Question 4: Sequential Exemption Priority

**Question**: 171 Sequential exemptions exist. Categories: viper/pflag (58), os.Chdir (37), seam variables (11), pgDriver (11), os.Stderr (5), others (49). Which category to tackle first?

**A)** viper/pflag (58) - largest category, biggest impact. SEAM PATTERN injection for viper globals.
**B)** os.Chdir (37) - second largest. Evaluate if tests can use t.TempDir() + relative paths instead.
**C)** seam variables (11) - these ARE the SEAM PATTERN. Already correct, just needs documentation alignment.
**D)** Start with the smallest categories (os.Stderr=5, pgDriver=11) to build momentum, then tackle larger ones.
**E)**

**Answer**:

**Rationale**: Determines Phase 4 execution order. The viper/pflag category alone could eliminate 34% of all exemptions.

---

## Question 5: Identity Services Migration Depth

**Question**: Identity services (authz, idp, rp, rs, spa) are described as half-baked. PKI-CA also needs domain completion. What migration depth?

**A)** Full domain logic: Each identity service gets complete domain implementation comparable to sm-im and jose-ja.
**B)** Infrastructure only: Migrate to latest builder pattern + contract tests. Domain logic deferred to separate plan.
**C)** One service deep: Fully implement identity-authz as reference, then replicate pattern to other 4.
**D)** Stub services: Builder + contract tests + placeholder domain routes. Functional but minimal business logic.
**E)**

**Answer**:

**Rationale**: Determines Phase 7-8 scope. Full domain logic is months of work. Infrastructure-only is weeks. This is the biggest scope decision.

---

## Question 6: InsecureSkipVerify Removal Strategy

**Question**: Phase 2 removes InsecureSkipVerify from all test HTTP clients. The replacement is a TLS test bundle (self-signed CA chain) in service-template. How should test TLS be configured?

**A)** Shared test CA chain in service-template. All services use the same test root CA. Tests trust this root CA.
**B)** Per-service test certs generated in TestMain. Each service has its own ephemeral test CA.
**C)** Use the existing server's auto-generated certs. Extract the CA from the running server and add to client trust store.
**D)** Keep InsecureSkipVerify for unit tests (no real servers), remove only for integration/E2E tests.
**E)**

**Answer**:

**Rationale**: Determines Phase 2 architecture. Option C leverages existing auto-cert generation with zero new infrastructure.

---

## Question 7: Agent Semantic Commit Enforcement

**Question**: Agents currently make bulk commits mixing multiple semantic categories. The Multi-Category Fix Commit Rule says each root-cause category = separate commit. How to enforce?

**A)** Add commit linting to pre-commit hooks (commitlint or similar). Reject commits that don't match conventional format.
**B)** Add a lint-fitness sub-linter that validates commit message format in CI.
**C)** Agent instruction improvements only. No automated enforcement. Trust the AI to follow instructions.
**D)** Both A and B: pre-commit hook + CI validation. Defense in depth.
**E)**

**Answer**:

**Rationale**: Determines Phase 9 (Task 9.2) scope. Automated enforcement is more reliable but adds tooling overhead.
