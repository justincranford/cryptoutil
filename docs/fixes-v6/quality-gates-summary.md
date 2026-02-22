# Quality Gates Summary — fixes-v6

Comprehensive audit of EVERY project file against ARCHITECTURE.md standards.

See [quality-gates-details.md](quality-gates-details.md) for per-finding data.

---

## CRITICAL Priority (3 findings)

1. **[F-6.1: math/rand in production code (FIPS violation)](#f-61-mathrand-in-production-code-fips-violation)** — `jose/ja/repository/audit_repository.go` uses `math/rand` for audit sampling. Architecture MANDATES `crypto/rand` ALWAYS.

2. **[F-6.2: context.Background() in HTTP handlers](#f-62-contextbackground-in-http-handlers)** — `template/service/server/apis/sessions.go` uses `context.Background()` instead of Fiber request context. Breaks trace correlation and cancellation propagation.

3. **[F-6.3: InsecureSkipVerify conditional in production CLI](#f-63-insecureskipverify-conditional-in-production-cli)** — `template/service/cli/http_client.go` falls back to `InsecureSkipVerify: true` when no CA cert provided. Architecture says NEVER.

---

## HIGH Priority (12 findings)

4. **[F-6.4: ValidateUUIDs wraps wrong error (bug)](#f-64-validateuuids-wraps-wrong-error-bug)** — `shared/util/random/uuid.go` wraps `ErrUUIDsCantBeNil` instead of actual validation error. Silent data corruption.

5. **[F-6.5: Copy-paste bug — "sqlite" in PostgreSQL function](#f-65-copy-paste-bug--sqlite-in-postgresql-function)** — `shared/container/postgres.go` error says "failed to start sqlite container".

6. **[F-6.6: Generic error messages leak JWK context](#f-66-generic-error-messages-leak-jwk-context)** — `shared/apperr/app_errors.go` sentinel errors say "jwks can't be nil/empty" but used generically.

7. **[F-6.7: Shared packages import from apps/template (dependency inversion)](#f-67-shared-packages-import-from-appstemplate-dependency-inversion)** — 4 files in `shared/` import `apps/template/service/telemetry`. Violates dependency direction.

8. **[F-6.8: Error sentinels typed as string not error](#f-68-error-sentinels-typed-as-string-not-error)** — `jws_jwk_util.go`/`jwe_jwk_util.go` use `string` typed sentinels instead of `errors.New()`.

9. **[F-6.9: Test helper file invisible — space in filename](#f-69-test-helper-file-invisible--space-in-filename)** — `usernames_passwords_test util.go` contains a space; Go ignores it.

10. **[F-6.10: Identity magic package not in shared/magic](#f-610-identity-magic-package-not-in-sharedmagic)** — 12 files, 754 lines in `internal/apps/identity/magic/` which should be in `internal/shared/magic/`.

11. **[F-6.11: PKI CA magic package not in shared/magic](#f-611-pki-ca-magic-package-not-in-sharedmagic)** — `internal/apps/pki/ca/magic/magic.go` with 28 lines scattered outside shared.

12. **[F-6.12: Identity config magic file not in shared/magic](#f-612-identity-config-magic-file-not-in-sharedmagic)** — `identity/config/magic.go` with hardcoded ports, timeouts.

13. **[F-6.13: Test files exceed 500-line hard limit](#f-613-test-files-exceed-500-line-hard-limit)** — `businesslogic_crud_test.go` (514), `oam_orm_mapper_test.go` (506), `tls_error_paths_test.go` (504).

14. **[F-6.14: Identity server & cmd packages have zero tests](#f-614-identity-server--cmd-packages-have-zero-tests)** — `identity/server/` (4 files) and `identity/cmd/` (7 files, 432-line CLI) completely untested.

15. **[F-6.15: //nolint:wsl violations (5 instances)](#f-615-nolintwsl-violations-5-instances)** — Architecture says NEVER use `//nolint:wsl`. Found in telemetry and identity unified files.

---

## MEDIUM Priority (24 findings)

16. **[F-6.16: 35 test files missing t.Parallel()](#f-616-35-test-files-missing-tparallel)** — Architecture REQUIRES t.Parallel() on all test functions and subtests.

17. **[F-6.17: pool.go long if/else chain (12 branches)](#f-617-poolgo-long-ifelse-chain-12-branches)** — Should use switch statement.

18. **[F-6.18: pool.go at 451 lines (approaching limit)](#f-618-poolgo-at-451-lines-approaching-limit)** — Approaching 500-line hard limit.

19. **[F-6.19: certificates.go at 474 lines](#f-619-certificatesgo-at-474-lines)** — Approaching 500-line hard limit.

20. **[F-6.20: identity/issuer/jws.go at 494 lines](#f-620-identityissuerjwsgo-at-494-lines)** — Approaching 500-line hard limit.

21. **[F-6.21: pki/ca/cli/cli.go at 492 lines](#f-621-pkicaclicligo-at-492-lines)** — Approaching 500-line hard limit.

22. **[F-6.22: workflow/workflow_executor.go at 491 lines](#f-622-workflowworkflow_executorgo-at-491-lines)** — Approaching 500-line hard limit.

23. **[F-6.23: Container package has zero tests](#f-623-container-package-has-zero-tests)** — Infrastructure code requiring ≥98% coverage.

24. **[F-6.24: Empty shared/barrier/ directory](#f-624-empty-sharedbarrier-directory)** — Referenced in architecture but actual impl in apps/template.

25. **[F-6.25: TLS chain.go constants outside magic](#f-625-tls-chaingo-constants-outside-magic)** — `DefaultCAChainLength`, `DefaultCADuration`, `DefaultEndEntityDuration` not in magic.

26. **[F-6.26: Algorithm string constants in jose files not in magic](#f-626-algorithm-string-constants-in-jose-files-not-in-magic)** — 13 algorithm string constants in `jws_jwk_util.go`.

27. **[F-6.27: pwdgen.go policy constants outside magic](#f-627-pwdgengo-policy-constants-outside-magic)** — `basicPolicyMinLength`, etc. defined locally.

28. **[F-6.28: unseal_keys_service else-if chain](#f-628-unseal_keys_service-else-if-chain)** — 6-branch else-if should be switch.

29. **[F-6.29: context.Background() in barrier services](#f-629-contextbackground-in-barrier-services)** — root_keys_service.go and intermediate_keys_service.go use context.Background() for DB transactions.

30. **[F-6.30: 91 httptest.NewServer usages in tests](#f-630-91-httptestnewserver-usages-in-tests)** — Architecture says ALWAYS use Fiber app.Test() for handler tests.

31. **[F-6.31: Observability tests — 30 standalone functions](#f-631-observability-tests--30-standalone-functions)** — Should be table-driven.

32. **[F-6.32: identity/rp/ and identity/spa/ have zero tests](#f-632-identityrp-and-identityspa-have-zero-tests)** — Service entry points untested.

33. **[F-6.33: pki/ca/domain/ has zero tests](#f-633-pkicadomain-has-zero-tests)** — Domain models untested.

34. **[F-6.34: //nolint:wrapcheck,thelper blanket suppression](#f-634-nolintwrapchekthelper-blanket-suppression)** — `keygen_error_paths_test.go` uses forbidden blanket nolint.

35. **[F-6.35: jose package name mismatch](#f-635-jose-package-name-mismatch)** — Directory `jose/` but `package crypto`. Confusing.

36. **[F-6.36: duplicate identity/demo constants with demo package](#f-636-duplicate-identitydemo-constants-with-demo-package)** — Repeated demo credentials and ports.

37. **[F-6.37: Unused sentinel errors in database/sharding.go](#f-637-unused-sentinel-errors-in-databaseshardinggo)** — 6 of 8 declared but never used.

38. **[F-6.38: SQL interpolation in sharding (defense in depth)](#f-638-sql-interpolation-in-sharding-defense-in-depth)** — `SET search_path TO "%s"` string interpolation.

39. **[F-6.39: Fmt.Errorf without %w audit needed](#f-639-fmterrorf-without-w-audit-needed)** — ~1,089 instances without `%w` vs 2,047 with. Many acceptable for validation errors.

---

## LOW Priority (7 findings)

40. **[F-6.40: CICD import alias convention deviation](#f-640-cicd-import-alias-convention-deviation)** — Uses `lintGo*` instead of `cryptoutil*`.

41. **[F-6.41: Hardcoded test password in testutil.go](#f-641-hardcoded-test-password-in-testutilgo)** — `TestPassword123!` instead of dynamic generation.

42. **[F-6.42: E2E test constants could be in magic](#f-642-e2e-test-constants-could-be-in-magic)** — `identity/test/e2e/constants.go`.

43. **[F-6.43: Template CLI constants not in magic](#f-643-template-cli-constants-not-in-magic)** — `template/service/cli/constants.go`.

44. **[F-6.44: ValidateUUID takes *string pointer unnecessarily](#f-644-validateuuid-takes-string-pointer-unnecessarily)** — All callers use value semantics.

45. **[F-6.45: Hardcoded boundary UUIDs in tests (documented)](#f-645-hardcoded-boundary-uuids-in-tests-documented)** — Acceptable for edge-case testing.

46. **[F-6.46: pool.go //nolint:errcheck on close()](#f-646-poolgo-nolinterrcheck-on-close)** — Channel close cannot fail; acceptable.

---

## Overlap with fixes-v5

| fixes-v6 Finding | Overlaps with fixes-v5 Finding | Notes |
|-------------------|-------------------------------|-------|
| F-6.16 (missing t.Parallel) | F-2.9 (PKI subtests) | fixes-v6 is broader scope (35 files vs 1) |
| F-6.10 (identity magic) | F-4.2 (scattered constants) | fixes-v6 identifies specific location |
| F-6.25 (TLS constants) | F-4.2 (scattered constants) | Same category, different file |

All other fixes-v6 findings are NEW and not duplicated from fixes-v5.
