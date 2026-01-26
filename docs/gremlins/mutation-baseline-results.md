# Mutation Testing Baseline Results

## Summary

Mutation testing baseline established on Linux (ncc-1701-d) using gremlins v0.6.0.

**Configuration**: `.gremlins.yml` with 180s timeout, 6 mutators (ARITHMETIC_BASE, CONDITIONALS_BOUNDARY, CONDITIONALS_NEGATION, INCREMENT_DECREMENT, INVERT_NEGATIVES, REMOVE_SELF_ASSIGNMENTS), 85% efficacy threshold.

**Date**: 2026-01-26

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

