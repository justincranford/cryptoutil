# Quality Gates Summary — fixes-v6

Comprehensive audit of EVERY project file against ARCHITECTURE.md standards.

See [quality-gates-details.md](quality-gates-details.md) for per-finding data.

---

## CRITICAL Priority (3 findings)

Ordered by ease of fix (easiest first) within severity.

| # | ID | Ease | Finding |
|---|-----|------|---------|
| 1 | F-6.1 | Easy | [math/rand in production code (FIPS violation)](#f-61-mathrand-in-production-code-fips-violation) — 1 import + usage change |
| 2 | F-6.2 | Easy | [context.Background() in HTTP handlers](#f-62-contextbackground-in-http-handlers) — 2 line changes |
| 3 | F-6.3 | Medium | [InsecureSkipVerify conditional in production CLI](#f-63-insecureskipverify-conditional-in-production-cli) — add error path |

---

## HIGH Priority (12 findings)

| # | ID | Ease | Finding |
|---|-----|------|---------|
| 4 | F-6.4 | Easy | [ValidateUUIDs wraps wrong error (bug)](#f-64-validateuuids-wraps-wrong-error-bug) — 1 line fix |
| 5 | F-6.5 | Easy | [Copy-paste bug — "sqlite" in PostgreSQL function](#f-65-copy-paste-bug--sqlite-in-postgresql-function) — 1 string fix |
| 6 | F-6.6 | Easy | [Generic error messages leak JWK context](#f-66-generic-error-messages-leak-jwk-context) — 2 string changes |
| 7 | F-6.9 | Easy | [Test helper file invisible — space in filename](#f-69-test-helper-file-invisible--space-in-filename) — 1 rename |
| 8 | F-6.11 | Easy | [PKI CA magic package not in shared/magic](#f-611-pki-ca-magic-package-not-in-sharedmagic) — 1 file, 28 lines |
| 9 | F-6.12 | Easy | [Identity config magic file not in shared/magic](#f-612-identity-config-magic-file-not-in-sharedmagic) — 1 file, 62 lines |
| 10 | F-6.8 | Medium | [Error sentinels typed as string not error](#f-68-error-sentinels-typed-as-string-not-error) — change type + callers |
| 11 | F-6.15 | Medium | [//nolint:wsl violations (5 instances)](#f-615-nolintwsl-violations-5-instances) — restructure 5 sections |
| 12 | F-6.13 | Medium | [Test files exceed 500-line hard limit](#f-613-test-files-exceed-500-line-hard-limit) — split 3 files |
| 13 | F-6.7 | Complex | [Shared imports from apps/template (dependency inversion)](#f-67-shared-packages-import-from-appstemplate-dependency-inversion) — extract interface |
| 14 | F-6.10 | Complex | [Identity magic package not in shared/magic](#f-610-identity-magic-package-not-in-sharedmagic) — 12 files, 754 lines |
| 15 | F-6.14 | Complex | [Identity server & cmd packages have zero tests](#f-614-identity-server--cmd-packages-have-zero-tests) — 11 untested files |

---

## MEDIUM Priority (24 findings)

| # | ID | Ease | Finding |
|---|-----|------|---------|
| 16 | F-6.17 | Easy | [pool.go long if/else chain (12 branches)](#f-617-poolgo-long-ifelse-chain-12-branches) — switch conversion |
| 17 | F-6.24 | Easy | [Empty shared/barrier/ directory](#f-624-empty-sharedbarrier-directory) — delete or add doc.go |
| 18 | F-6.25 | Easy | [TLS chain.go constants outside magic](#f-625-tls-chaingo-constants-outside-magic) — move 3 constants |
| 19 | F-6.26 | Easy | [Algorithm string constants not in magic](#f-626-algorithm-string-constants-in-jose-files-not-in-magic) — move 13 constants |
| 20 | F-6.27 | Easy | [pwdgen.go policy constants outside magic](#f-627-pwdgengo-policy-constants-outside-magic) — move constants |
| 21 | F-6.28 | Easy | [unseal_keys_service else-if chain](#f-628-unseal_keys_service-else-if-chain) — switch conversion |
| 22 | F-6.34 | Easy | [//nolint:wrapcheck,thelper blanket suppression](#f-634-nolintwrapchekthelper-blanket-suppression) — remove + fix |
| 23 | F-6.37 | Easy | [Unused sentinel errors in sharding.go](#f-637-unused-sentinel-errors-in-databaseshardinggo) — remove 6 unused |
| 24 | F-6.29 | Medium | [context.Background() in barrier services](#f-629-contextbackground-in-barrier-services) — propagate ctx |
| 25 | F-6.31 | Medium | [Observability tests — 30 standalone functions](#f-631-observability-tests--30-standalone-functions) — table-driven conversion |
| 26 | F-6.36 | Medium | [Duplicate identity/demo constants](#f-636-duplicate-identitydemo-constants-with-demo-package) — dedup |
| 27 | F-6.38 | Medium | [SQL interpolation in sharding (defense)](#f-638-sql-interpolation-in-sharding-defense-in-depth) — validate input |
| 28 | F-6.18 | Medium | [pool.go at 451 lines](#f-618-poolgo-at-451-lines-approaching-limit) — split before limit |
| 29 | F-6.19 | Medium | [certificates.go at 474 lines](#f-619-certificatesgo-at-474-lines) — split before limit |
| 30 | F-6.20 | Medium | [identity/issuer/jws.go at 494 lines](#f-620-identityissuerjwsgo-at-494-lines) — split before limit |
| 31 | F-6.21 | Medium | [pki/ca/cli/cli.go at 492 lines](#f-621-pkicaclicligo-at-492-lines) — split before limit |
| 32 | F-6.22 | Medium | [workflow_executor.go at 491 lines](#f-622-workflowworkflow_executorgo-at-491-lines) — split before limit |
| 33 | F-6.16 | Complex | [35 test files missing t.Parallel()](#f-616-35-test-files-missing-tparallel) — 35 files to modify |
| 34 | F-6.23 | Complex | [Container package has zero tests](#f-623-container-package-has-zero-tests) — write tests (Docker) |
| 35 | F-6.30 | Complex | [91 httptest.NewServer usages](#f-630-91-httptestnewserver-usages-in-tests) — audit 91 usages |
| 36 | F-6.32 | Complex | [identity/rp/ and identity/spa/ zero tests](#f-632-identityrp-and-identityspa-have-zero-tests) — write tests |
| 37 | F-6.33 | Complex | [pki/ca/domain/ has zero tests](#f-633-pkicadomain-has-zero-tests) — write tests |
| 38 | F-6.35 | Complex | [jose package name mismatch](#f-635-jose-package-name-mismatch) — rename pkg + all imports |
| 39 | F-6.39 | Complex | [fmt.Errorf without %w audit](#f-639-fmterrorf-without-w-audit-needed) — audit 1,089 instances |

---

## LOW Priority (7 findings)

| # | ID | Ease | Finding |
|---|-----|------|---------|
| 40 | F-6.40 | Easy | [CICD import alias convention deviation](#f-640-cicd-import-alias-convention-deviation) — document or standardize |
| 41 | F-6.41 | Easy | [Hardcoded test password in testutil.go](#f-641-hardcoded-test-password-in-testutilgo) — generate dynamically |
| 42 | F-6.42 | Easy | [E2E test constants could be in magic](#f-642-e2e-test-constants-could-be-in-magic) — evaluate |
| 43 | F-6.43 | Easy | [Template CLI constants not in magic](#f-643-template-cli-constants-not-in-magic) — evaluate |
| 44 | F-6.45 | Easy | [Hardcoded boundary UUIDs in tests](#f-645-hardcoded-boundary-uuids-in-tests-documented) — document intent |
| 45 | F-6.46 | Easy | [pool.go //nolint:errcheck on close()](#f-646-poolgo-nolinterrcheck-on-close) — add justification |
| 46 | F-6.44 | Medium | [ValidateUUID takes *string pointer](#f-644-validateuuid-takes-string-pointer-unnecessarily) — change signature + callers |

---

## Overlap with fixes-v5

| fixes-v6 Finding | Overlaps with fixes-v5 Finding | Notes |
|-------------------|-------------------------------|-------|
| F-6.16 (missing t.Parallel) | F-2.9 (PKI subtests) | fixes-v6 is broader scope (35 files vs 1) |
| F-6.10 (identity magic) | F-4.2 (scattered constants) | fixes-v6 identifies specific location |
| F-6.25 (TLS constants) | F-4.2 (scattered constants) | Same category, different file |

All other fixes-v6 findings are NEW and not duplicated from fixes-v5.
