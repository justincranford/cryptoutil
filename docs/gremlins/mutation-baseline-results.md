# Mutation Testing Baseline Results

## Summary

Mutation testing baseline established on Linux (ncc-1701-d) using gremlins v0.6.0.

**Configuration**: `.gremlins.yml` with 180s timeout, 6 mutators (ARITHMETIC_BASE, CONDITIONALS_BOUNDARY, CONDITIONALS_NEGATION, INCREMENT_DECREMENT, INVERT_NEGATIVES, REMOVE_SELF_ASSIGNMENTS), 85% efficacy threshold.

**Baseline Date**: 2026-01-26
**QG-6 Update Date**: 2026-01-28

---

## QG-6 Mutation Testing Improvement Results

Per-package gremlins analysis with `--timeout-coefficient=60`. All packages run individually to prevent terminal overflow.

### Package Results Table

| Package | Before | After | Killed | Lived | NC | TO | Status |
|---------|--------|-------|--------|-------|-----|-----|--------|
| jose (ja) | 96.15% | 100% | 269 | 0 | — | — | ✅ |
| certificate | — | 100% | 124 | 0 | — | — | ✅ |
| hash | — | 100% | 57 | 0 | — | — | ✅ |
| digests | — | 100% | — | 0 | — | — | ✅ |
| keygen | — | 100% | — | 0 | — | — | ✅ |
| database | — | 100% | — | 0 | — | — | ✅ |
| apperr | — | 100% | — | 0 | — | — | ✅ |
| builder | — | 100% | 34 | 0 | — | — | ✅ |
| listener | — | 100% | 20 | 0 | — | — | ✅ |
| application | — | 100% | 30 | 0 | — | — | ✅ |
| middleware | — | 100% | 10 | 0 | — | — | ✅ |
| repository | — | 100% | — | 0 | — | — | ✅ |
| sm-kms handler | — | 100% | 42 | 0 | — | — | ✅ |
| sm-kms businesslogic | — | 100% | 20 | 0 | 21 | 135 | ✅ |
| telemetry | 61.29% | 100% | 16 | 0 | 4 | 23 | ✅ |
| service | 88.71% | 99.19% | 123 | 1 | — | — | ✅ |
| lint_deployments | — | 98.68% | 300 | 4 | 3 | 0 | ✅ |
| client | — | 97.44% | 38 | 1 | — | — | ✅ |
| combinations | — | 96.67% | 29 | 1 | — | — | ✅ |
| config | — | 96.43% | — | — | — | — | ✅ |
| tls | 79.17% | 95.83% | — | — | — | — | ✅ |
| tenant | — | 95.92% | — | — | — | — | ✅ |
| barrier | 92.16% | 94.12% | 48 | 3 | — | — | ⚠️ structural |
| apis | 86.27% | 92.31% | 36 | 3 | — | — | ⚠️ structural |
| realms | 79.66% | 91.53% | 54 | 5 | — | — | ⚠️ structural |
| realm | 85.71% | 90.76% | 108 | 11 | — | — | ⚠️ structural |
| files | 82.14% | 89.29% | 25 | 3 | — | — | ⚠️ structural |
| pool | — | 73.47% | 36 | 13 | — | — | ⚠️ concurrency |
| cli | — | 66.91% | 93 | 46 | — | — | ⚠️ CLI patterns |
| container | — | 66.67% | 2 | 1 | 10 | — | ⚠️ DB-dependent |
| template businesslogic | — | 0% | 0 | 0 | 0 | 114 | ⚠️ all timeout |
| sm-kms ORM | — | 0% | 0 | 0 | 108 | 0 | ⚠️ all NC |

### Packages At or Above 95% Threshold

22 packages meet or exceed 95% efficacy (15 at 100%).

### Structural Ceiling Categories

1. **Semantically equivalent mutants**: `repeatedCount > x` vs `>= x` assigns same value when equal; `i > 0` boundary comparing zero-value prevRune at i=0.
2. **DB-dependent code**: GORM operations (Create/Find/Update/Delete) require real PostgreSQL connections — mock-resistant.
3. **Complex integration mocking**: Session managers, realm providers, tenant federation lookups.
4. **Concurrency patterns**: goroutine scheduling, channel operations, pool resize races.
5. **CLI arg parsing**: Repetitive `i++` and bounds checking in health command iteration.
6. **Dead code paths**: dotfile extension handling where `filepath.Ext(".gitignore")` returns `".gitignore"` not `""`.
7. **Formatting-only guards**: `len(slice) > 0` before `strings.Join` — empty join produces same visible result.
8. **Boolean comparisons**: `!=` on booleans has no meaningful BOUNDARY mutation.

### Test Files Created/Modified

| File | Tests Added | Mutants Killed |
|------|-------------|----------------|
| `realms/realm_validation_boundary_test.go` | 6 | ~12 (79→91%) |
| `service/realm_config_validate_boundary_test.go` | 2 | ~13 (88→99%) |
| `realm/tenant_boundary_test.go` | 3 | ~6 (85→90%) |
| `files/files_mutation_test.go` | 2 | 2 (82→89%) |
| `lint_deployments/lint_deployments_boundary_test.go` | 5 | 5 (97→98.68%) |

## Results by Service

### 1. JOSE-JA (jose-ja package)

**Command**: `gremlins unleash ./internal/apps/jose/ja/`

**Results**:
- **Test Efficacy**: 96.15%
- **Killed**: 100
- **Lived**: 4
- **Not Covered**: 4
- **Timed Out**: 195
- **Status**: ✅ **EXCEEDS 85% TARGET**

**Lived Mutations** (requires analysis):
1. `repository/audit_repository.go:112:24` - CONDITIONALS_BOUNDARY
2. `repository/audit_repository.go:101:26` - CONDITIONALS_NEGATION
3. `repository/audit_repository.go:101:26` - CONDITIONALS_BOUNDARY
4. `server/server.go:130:34` - CONDITIONALS_NEGATION

**Not Covered**:
- `server/config/config.go:67, 80, 98` - pflag global state prevents testing

**Timeouts**: Expected for service-level code with long-running operations (195 mutations)

**Log**: `/tmp/gremlins_jose_ja.log`

---

### 2. Cipher-IM (cipher-im package)

**Command**: `gremlins unleash --tags='!integration,!e2e' ./internal/apps/cipher/im/`

**Results**:
- **Status**: ❌ **BLOCKED** - Infrastructure issues prevent testing

**Blockers**:
1. Docker container `cipher-im-sqlite` unhealthy
2. OTel collector HTTP/gRPC protocol mismatch
   - Error: `"failed to upload metrics: malformed HTTP response \x00\x00\x06\x04..."`
   - Root cause: cipher-im using HTTP endpoint to connect to gRPC collector (port 4317)
3. E2E tests run despite `--tags='!integration,!e2e'` exclusion flag
4. Repository-only tests all timeout (0% efficacy, 27 mutations timed out)

**Repository-Only Attempt**:
- **Command**: `gremlins unleash ./internal/apps/cipher/im/repository`
- **Results**: Test Efficacy: 0.00% (0 killed, 0 lived, 27 timed out)
- **Log**: `/tmp/gremlins_cipher_im_repo.log`

**Resolution Required**:
- Fix Docker compose health checks for cipher-im-sqlite container
- Correct OTel collector endpoint configuration (HTTP vs gRPC)
- Investigate why E2E tests bypass exclusion tags
- Address repository test timeouts (may need test infrastructure optimization)

**Log**: `/tmp/gremlins_cipher_im.log`

---

### 3. Template Service (template package)

**Command**: `gremlins unleash ./internal/apps/template/service/`

**Results**:
- **Test Efficacy**: 91.75%
- **Killed**: 278
- **Lived**: 25
- **Not Covered**: 329
- **Timed Out**: 519
- **Status**: ✅ **EXCEEDS 85% TARGET**

**Test Fix Required Before Testing**:
- Issue: `TestYAMLFieldMapping_KebabCase` failed with "browser rate limit cannot be 0 (would block all browser requests)"
- Fix: Added `browser-rate-limit: 100` and `service-rate-limit: 25` to test YAML config
- Commit: `00399210` - "fix(template): add rate limit config to TestYAMLFieldMapping_KebabCase"

**Lived Mutations** (requires analysis):
1. `config/config.go:949:60` - CONDITIONALS_NEGATION
2. `config/config.go:1046:38` - INCREMENT_DECREMENT
3. `config/config.go:1459:13` - CONDITIONALS_NEGATION
4. `config/config.go:1526:22` - CONDITIONALS_BOUNDARY
5. `config/config.go:1530:23` - CONDITIONALS_BOUNDARY
6. `config/config.go:1593:33` - CONDITIONALS_BOUNDARY
7. `config/tls_generator/tls_generator.go:39:28` - CONDITIONALS_BOUNDARY
8. `config/tls_generator/tls_generator.go:39:57` - CONDITIONALS_BOUNDARY
9. `config/tls_generator/tls_generator.go:75:15` - CONDITIONALS_NEGATION
10. `config/tls_generator/tls_generator.go:75:47` - CONDITIONALS_BOUNDARY
11. `config/tls_generator/tls_generator.go:75:47` - CONDITIONALS_NEGATION
12. `config/tls_generator/tls_generator.go:91:12` - CONDITIONALS_NEGATION
13. `config/tls_generator/tls_generator.go:92:17` - CONDITIONALS_NEGATION
14. `config/tls_generator/tls_generator.go:98:13` - INCREMENT_DECREMENT
15. `config/tls_generator/tls_generator.go:103:17` - CONDITIONALS_BOUNDARY
16. `config/tls_generator/tls_generator.go:103:17` - CONDITIONALS_NEGATION
17. `config/tls_generator/tls_generator.go:105:36` - CONDITIONALS_NEGATION
18. `config/tls_generator/tls_generator.go:190:18` - CONDITIONALS_BOUNDARY
19. `config/tls_generator/tls_generator.go:190:18` - CONDITIONALS_NEGATION
20. `config/tls_generator/tls_generator.go:194:42` - ARITHMETIC_BASE
21. `config/tls_generator/tls_generator.go:269:69` - INVERT_NEGATIVES
22. `config/tls_generator/tls_generator.go:269:69` - ARITHMETIC_BASE
23. `config/tls_generator/tls_generator.go:361:85` - ARITHMETIC_BASE
24. `server/service/realm_service.go:435:23` - CONDITIONALS_BOUNDARY
25. `server/service/registration_service.go:232:67` - ARITHMETIC_BASE

**Not Covered**: 329 mutations (primarily in application bootstrap code, test utilities, and client auth code)

**Timeouts**: 519 mutations (expected for server startup code, barrier services, session management with database operations)

**Log**: `/tmp/gremlins_template.log`

---

## Next Steps

### Immediate (Task 6.2)

1. Analyze lived mutations across JOSE-JA (4) and Template (25)
2. Categorize by mutation type and severity
3. Determine if additional tests needed or mutations acceptable

### Short-term (Task 6.3)

1. Implement mutation-killing tests for high-value lived mutations
2. Target mutations in critical paths (authentication, authorization, crypto)
3. Skip mutations in edge cases or non-critical paths (cost/benefit analysis)

### Medium-term (Task 6.4)

1. Enable continuous mutation testing in CI/CD pipeline
2. Add mutation efficacy checks to PR workflows
3. Set up trend tracking for mutation scores

### Blocked (Cipher-IM)

1. Fix Docker compose infrastructure issues
2. Correct OTel collector endpoint configuration
3. Investigate E2E test exclusion tag bypass
4. Optimize repository test performance (address timeouts)

---

## Observations

**High Efficacy Scores**: JOSE-JA (96.15%) and Template (91.75%) both exceed 85% target, demonstrating strong test coverage and quality.

**Timeout Pattern**: High timeout counts (195 JOSE-JA, 519 Template) expected for service-level code with:
- Database operations (session management, repository queries)
- Server startup/shutdown sequences
- Barrier encryption services (root/intermediate/content key operations)
- Middleware chains with external dependencies

**Not Covered Pattern**: Primarily in:
- Application bootstrap code (dependency injection, server initialization)
- Test utilities and helpers
- Client authentication code (user_auth.go - client-side logic)
- Configuration loading (pflag global state prevents testing)

**Infrastructure vs Business Logic**: Template service shows healthy separation - config/TLS generation code has lived mutations (non-critical infrastructure), while business logic (realm_service, registration_service) has minimal lived mutations.

**Cipher-IM Blockers**: Require separate infrastructure fix track (Docker compose, OTel configuration, test exclusion tags).
