# Quizme v3 — Framework v10: Critical Issues from Deep Analysis

**Created**: 2026-04-13
**Context**: Deep analysis of the entire codebase (configs/, deployments/, cicd-lint, framework/,
ENG-HANDBOOK.md, pki-init code, shared-postgres, secrets, compose files) revealed 8 critical
issues that require user decisions before Phase 1 implementation can begin.

**Instructions**: Answer each question with A, B, C, D, or E (custom). Write your answer on
the **Answer:** line.

---

## Question 1: pki-init Runtime Bug — Code/Compose Argument Mismatch

**Question**: The Go implementation of pki-init (`internal/apps/framework/tls/init.go`) expects
exactly 2 positional args: `<tier-id> <target-dir>`. It does `len(args) != 2` with NO flag
parsing (no pflag, no cobra, no flag.FlagSet). But ALL compose files pass
`["init", "--output-dir=/certs"]` — after the router strips "init", only 1 arg `"--output-dir=/certs"`
reaches Init(), causing it to print usage and exit 1. **Containers ALWAYS fail on pki-init.**

How should this be fixed?

**A)** Rewrite `init.go` to use pflag: `--tier-id=<id>` and `--output-dir=<dir>` flags.
Compose becomes `["init", "--tier-id=sm-kms", "--output-dir=/certs"]`. This aligns with the
rest of the framework (which uses pflag everywhere) and is the most extensible.

**B)** Keep init.go positional but fix compose to pass positional args:
`["init", "sm-kms", "/certs"]`. Minimal code change — just fix compose files.
But fragile (positional args are order-dependent and not self-documenting).

**C)** Add a `--domain` flag to init.go (in addition to --output-dir) so compose becomes
`["init", "--domain=sm-kms", "--output-dir=/certs"]`. The --domain flag was already in the
plan (Decision 4) so add it now alongside the fix.

**D)** This is out of scope for v10. Document the bug, create a separate fix task outside
the template registry work, and make the template use whatever format the fixed code will use.

**E)**

**Answer**:

**Rationale**: This is a runtime failure in all containers that attempt pki-init. The template
must know the correct argument format to generate valid compose files. If it's D (out of scope),
the template content for pki-init service command must use a TBD placeholder.

---

## Question 2: Docker Compose exec-form Cannot Shell-Expand Environment Variables

**Question**: Decision 8 (quizme-v2 Q2) specified using `SUITE_ARGS` / `PRODUCT_ARGS` environment
variables to pass additional `--config=` flags at higher deployment tiers. However, Docker Compose
JSON array command format `["server", "--config=...", "$SUITE_ARGS"]` does NOT shell-expand —
it passes the literal string `"$SUITE_ARGS"`. This makes the decided mechanism non-functional.

How should config hierarchy work at deployment time?

**A)** Switch to shell-form command: `command: /bin/sh -c "exec /app/sm-kms server --config=... $SUITE_ARGS"`.
Shell-form supports variable expansion. Downside: loses exec-form signal handling (PID 1 is sh, not
the app). Tini entrypoint mitigates this.

**B)** Create a custom entrypoint wrapper script (`docker-entrypoint.sh`) that reads config
overlay paths from an env var and constructs the full command line. More complex but clean
signal handling.

**C)** Abandon the env var approach entirely. Instead, product/suite compose files use Docker Compose
`command:` override to specify the full command with all `--config=` flags inline. No env var
expansion needed — each tier's compose explicitly lists its configs.

**D)** Use Docker Compose `x-*` extension fields or profile-based config injection. More complex
but fully declarative without shell expansion.

**E)**

**Answer**:

**Rationale**: This directly impacts how product/suite compose templates pass config overlays to
PS-ID services. The template content depends on the chosen mechanism. Option C is the simplest
but means product/suite templates must enumerate all config paths.

---

## Question 3: shared-postgres Documentation/Implementation Gaps — v10 Scope?

**Question**: Deep analysis found 5 significant gaps between ENG-HANDBOOK documentation and the
actual shared-postgres implementation:

1. **DDL/DML user separation**: ENG-HANDBOOK §7 documents separate DDL and DML database users.
   Actual implementation has single `cryptoutil_admin` user with all privileges.
2. **Missing postgresql.conf**: ENG-HANDBOOK references `postgresql.conf` for tuning.
   Actual `deployments/shared-postgres/` has no postgresql.conf file.
3. **Incomplete replication script**: `setup-logical-replication.sh` only configures replication
   for pki-ca (1 of 10 PS-IDs). All other 9 PS-IDs are missing.
4. **Follower database count**: Follower `init-follower.sql` creates 16 databases, not the
   30 (3 tiers × 10) specified by Decision 10.
5. **Stale postgres-url.secret values**: Some PS-ID secrets reference removed per-PS-ID
   postgres hostnames (e.g., `sm-kms-postgres:5432` from before v8 consolidation).

Are these v10 scope?

**A)** YES — fix ALL 5 as part of v10 Phase 1 (Task 1.11 shared-postgres templates).
The template MUST reflect the correct architecture, so fixing the actual implementation is
required for templates to validate successfully.

**B)** PARTIAL — fix gaps 3-5 (replication, follower, stale URLs) in v10 since they directly
affect template compliance. Defer gaps 1-2 (DDL/DML separation, postgresql.conf) to a
dedicated database infrastructure task.

**C)** PARTIAL — only fix gap 5 (stale URLs) in v10 since it's trivial. Defer all others
to a separate phase.

**D)** NO — all 5 are out of v10 scope. Template-actual comparison for shared-postgres will
initially FAIL (known drift). Document the gaps and create a follow-up plan.

**E)**

**Answer**:

**Rationale**: v10's core mission is template registry + drift detection. If shared-postgres
templates are created but the actual files have known gaps, the linter will report failures
that are NOT regressions — they're pre-existing gaps now made visible. The question is whether
to fix them now or accept initial noise.

---

## Question 4: Config Classification Edge Cases

**Question**: Decision 7 (quizme-v2 Q1) splits standalone configs into framework vs domain.
Deep analysis found these edge cases that don't clearly belong to either category:

1. **pki-ca `storage.type`**: Controls whether pki-ca uses PostgreSQL, SQLite, or file-based
   storage. Is this framework (since it's analogous to `database-url`) or domain (since it's
   specific to how pki-ca stores certificates)?

2. **identity-rp `authz-server-url`, `client-id`, `redirect-uri`**: These configure the OIDC
   relying party. Are they framework (service connectivity like OTLP endpoint) or domain
   (identity-rp-specific business logic)?

3. **identity-spa `static-files-path`**: Configures where the SPA serves static files from.
   Framework (deployment/serving concern) or domain (SPA-specific)?

How should edge cases be classified?

**A)** **Conservative**: ALL ambiguous settings go to domain config. The framework config
template is minimal and uniform across all PS-IDs. Any setting that appears in only one
PS-ID is domain by definition.

**B)** **Pragmatic**: Classify by analogy — `storage.type` → framework (like `database-url`);
`authz-server-url` → framework (like OTLP endpoint); `static-files-path` → domain (SPA-specific).

**C)** **Simple rule**: If the setting exists in only 1-2 PS-IDs, it's domain. If it exists
in 3+ PS-IDs, it's framework. This gives a mechanical decision rule.

**D)** Don't split edge cases yet. Put them in domain config with a `# TODO: evaluate
framework vs domain` comment. Revisit after seeing how the split works in practice.

**E)**

**Answer**:

**Rationale**: The classification affects what goes into the framework config template (Task 1.3)
and what remains in per-PS-ID domain config files (Task 1.13). Wrong classification means either
the template has PS-ID-specific settings (breaks uniformity) or domain files have generic settings
(defeats the split purpose).

---

## Question 5: Deployment Config Framework/Domain Split

**Question**: Decision 7 splits standalone `configs/<ps-id>/<ps-id>.yml` into framework + domain.
But deployment configs (`deployments/<ps-id>/config/<ps-id>-app-{common,sqlite-1,sqlite-2,postgresql-1,postgresql-2}.yml`)
also exist. Currently, across all 10 PS-IDs, the deployment configs are nearly identical —
only pki-ca has one extra domain-specific key (`crl-directory` in common config).

Do deployment configs also need a framework/domain split?

**A)** NO — deployment configs stay as-is (no split). They contain only 4 instance-specific
keys (`cors-origins`, `otlp-service`, `otlp-hostname`, `database-url`) that are all framework.
The common config has 12 shared keys that are also all framework. Only pki-ca's `crl-directory`
is domain — handle it as an exception (extra key allowed, not checked by template).

**B)** YES — mirror the standalone config split. Create `<ps-id>-app-framework-common.yml` +
`<ps-id>-app-domain-common.yml` etc. Consistent pattern everywhere.

**C)** HYBRID — only split the common config (where pki-ca's `crl-directory` lives).
Instance configs (sqlite-1/2, postgresql-1/2) are pure framework, no split needed.

**D)** Defer this decision until after standalone config split is complete and working.
Start with deployment configs unchanged and revisit if the template linter shows friction.

**E)**

**Answer**:

**Rationale**: A split adds (up to) 50 more files (10 PS-IDs × 5 variants). If deployment
configs are already 99% framework, the split adds complexity without much benefit. But
inconsistency between standalone and deployment config patterns could confuse developers.

---

## Question 6: pki-init Service Name Collision at Product Level

**Question**: All 10 PS-ID compose files define a service named `pki-init`. When a product
compose (e.g., `deployments/sm/compose.yml`) includes multiple PS-ID compose files via
Docker Compose `include:`, the `pki-init` services from each included file collide — Docker
Compose merges them into a single `pki-init` service.

Currently each PS-ID pki-init uses the same image and similar config, so the merge is
"accidentally harmless." But with Decision 4 adding `--domain=<ps-id>`, each PS-ID needs
its pki-init to run with a DIFFERENT domain arg. A merged single service can't do that.

How should this be handled?

**A)** Rename pki-init services to `<ps-id>-pki-init` (e.g., `sm-kms-pki-init`,
`jose-ja-pki-init`). No collision possible. Template uses `__PS_ID__-pki-init` as service name.

**B)** Keep the service named `pki-init` in PS-ID compose but have the product compose
override with a product-level pki-init that runs once with `--domain=<product>` (covering
all PS-IDs in that product). Remove the per-PS-ID pki-init from product includes.

**C)** Use Docker Compose profiles to conditionally enable/disable pki-init per level.
PS-ID level: `profiles: [service]`, product level defines its own `profiles: [product]`.

**D)** This is fine as-is — the merge is acceptable because pki-init generates certs for
the entire cert directory. One pki-init running with any domain produces certs for all.

**E)**

**Answer**:

**Rationale**: The template for PS-ID compose.yml defines the pki-init service. The service
name in the template directly affects whether product-level compose works correctly. This
must be resolved before writing the template content.

---

## Question 7: Scope Assessment — Template Count Explosion

**Question**: The template count grew from the original 14 files (v10 plan creation) to ~60
physical files after adding secrets templates (14 per tier), shared-postgres templates (4+),
and framework/domain config split. After expansion, this produces ~329 expected files
(vs. the original 99).

This is a 4× increase in template file count and a 3.3× increase in expected files.
The implementation effort for Phase 1 grew from ~4h to estimated ~7h.

Is this scope acceptable for v10, or should some categories be deferred?

**A)** Full scope — implement ALL ~60 templates in v10 Phase 1. The secrets and shared-postgres
templates are essential for complete drift detection. Incomplete template coverage means the
linter silently allows drift in uncovered areas.

**B)** Core first — implement the original ~14 templates (deployment + config) in v10.
Defer secrets templates and shared-postgres templates to v11. The linter's comparison is
one-directional (template → actual); uncovered files are allowed as "extra."

**C)** Prioritized — implement deployment + config + secrets templates (~40 files) in v10.
Defer shared-postgres templates to v11 (they have known documentation gaps per Q3).

**D)** Split v10 into phases: Phase 1A (original 14 deployment/config templates), Phase 1B
(secrets templates), Phase 1C (shared-postgres templates). Each phase is a commit checkpoint.
All in v10 but with clear internal milestones.

**E)**

**Answer**:

**Rationale**: Scope management is critical. A 4× scope increase risks incomplete execution
or quality compromises. However, implementing secrets templates alongside deployment templates
ensures they stay synchronized from day one — adding them later means temporary drift.

---

## Question 8: Stale postgres-url.secret Values

**Question**: Some PS-ID `secrets/postgres-url.secret` files still reference per-PS-ID postgres
hostnames from before v8 consolidation:

- `sm-kms`: `postgres://sm_kms_database_user:...@sm-kms-postgres:5432/sm_kms_database`
- `jose-ja`: `postgres://jose_ja_database_user:...@jose-ja-postgres:5432/jose_ja_database`
- etc.

The `sm-kms-postgres` and `jose-ja-postgres` service names no longer exist (removed in v8
when postgres was consolidated to `shared-postgres`). The correct hostname is
`shared-postgres-leader:5432`.

Should these be fixed as part of v10?

**A)** YES — fix all stale hostnames to `shared-postgres-leader:5432` as part of Task 1.8
(fix actual deployment files). The secrets template (Decision 14) already uses the correct
hostname, so fixing actuals is required for template compliance.

**B)** YES — but make it Task 1.9 or a separate commit from the other Task 1.8 deployment
fixes. Different root cause (v8 migration gap vs. v10 template decisions).

**C)** NO — the postgres-url.secret values are Docker secrets and are only used at deployment
time. If nobody is deploying these services to those old hostnames, the stale values
are harmless. Fix when/if someone reports a deployment failure.

**D)** Partially — fix only the ones that would cause the template linter to fail (if secrets
are checked). If secrets are deferred (Q7 option B), this is moot.

**E)**

**Answer**:

**Rationale**: The template from Decision 14 uses `shared-postgres-leader:5432`.
If the secrets-compliance linter (Task 2.6) compares actual secrets against the template,
stale hostnames will be flagged as drift. Fixing them now avoids false-positive noise.
