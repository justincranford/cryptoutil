# Mutation Analysis - Lived Mutations

**Date**: 2026-01-26
**Gremlins Version**: v0.6.0
**Total Lived Mutations**: 29 (4 JOSE-JA + 25 Template)

---

## Executive Summary

**High-Priority Fixes** (Business Logic - CRITICAL):
- 4 lived mutations in JOSE-JA audit repository (boundary/negation conditions)
- 2 lived mutations in Template business logic (realm service, registration service)

**Medium-Priority Fixes** (Infrastructure - IMPORTANT):
- 6 lived mutations in Template config validation

**Low-Priority Fixes** (Non-Critical Infrastructure):
- 17 lived mutations in Template TLS certificate generator

**ROI Assessment**:
- **High ROI**: 6 mutations (authentication/authorization/audit paths)
- **Medium ROI**: 6 mutations (config validation)
- **Low ROI**: 17 mutations (TLS generator - error paths, rarely executed)

---

## JOSE-JA Lived Mutations (4 total)

### 1. audit_repository.go:112:24 - CONDITIONALS_BOUNDARY

**Location**: `repository/audit_repository.go:112:24`
**Mutation Type**: CONDITIONALS_BOUNDARY
**Original Code**:
```go
//nolint:gosec // Not cryptographic - only for sampling decision.
return rand.Float64() < cryptoutilSharedMagic.JoseJAAuditFallbackSamplingRate, nil
```

**Mutated Code** (likely):
```go
return rand.Float64() <= cryptoutilSharedMagic.JoseJAAuditFallbackSamplingRate, nil
```

**Severity**: MEDIUM
**Risk**: Audit sampling rate boundary condition
- **Impact**: Edge case where rand.Float64() == JoseJAAuditFallbackSamplingRate
- **Probability**: Low (exact float equality rare)
- **Context**: Fallback sampling when config not found

**Test Gap**: Missing boundary condition test for sampling rate edge case

**Recommended Fix**:
```go
// Add test case in audit_repository_test.go
func TestShouldAudit_FallbackSamplingBoundary(t *testing.T) {
    // Test exact boundary: rand.Float64() == fallback sampling rate
    // Verify <= vs < behavior
}
```

**Priority**: MEDIUM (audit reliability)

---

### 2. audit_repository.go:101:26 - CONDITIONALS_NEGATION

**Location**: `repository/audit_repository.go:101:26`
**Mutation Type**: CONDITIONALS_NEGATION
**Original Code**:
```go
if errors.Is(err, gorm.ErrRecordNotFound) {
    //nolint:gosec // Not cryptographic - only for sampling decision.
    return rand.Float64() < cryptoutilSharedMagic.JoseJAAuditFallbackSamplingRate, nil
}
```

**Mutated Code** (likely):
```go
if !errors.Is(err, gorm.ErrRecordNotFound) {
    //nolint:gosec // Not cryptographic - only for sampling decision.
    return rand.Float64() < cryptoutilSharedMagic.JoseJAAuditFallbackSamplingRate, nil
}
```

**Severity**: HIGH
**Risk**: Audit logic inversion
- **Impact**: Wrong behavior when record not found vs other errors
- **Probability**: Medium (database errors common)
- **Context**: Fallback vs error propagation decision

**Test Gap**: Missing test for non-ErrRecordNotFound error cases

**Recommended Fix**:
```go
// Add test case in audit_repository_test.go
func TestShouldAudit_DatabaseErrorPropagation(t *testing.T) {
    // Test non-ErrRecordNotFound errors (e.g., connection failure)
    // Verify error is propagated, not converted to fallback sampling
}
```

**Priority**: HIGH (audit correctness)

---

### 3. audit_repository.go:101:26 - CONDITIONALS_BOUNDARY (duplicate location)

**Location**: `repository/audit_repository.go:101:26`
**Mutation Type**: CONDITIONALS_BOUNDARY
**Original Code**: Same as mutation #2 above
**Severity**: HIGH
**Priority**: HIGH (same fix as mutation #2)

---

### 4. server/server.go:130:34 - CONDITIONALS_NEGATION

**Location**: `server/server.go:130:34`
**Original Code**:
```go
if err := s.app.Start(ctx); err != nil {
    return fmt.Errorf("failed to start application: %w", err)
}
```

**Mutated Code** (likely):
```go
if err := s.app.Start(ctx); err == nil {
    return fmt.Errorf("failed to start application: %w", err)
}
```

**Severity**: MEDIUM
**Risk**: Server startup error handling inversion
- **Impact**: Error when app starts successfully, no error when app fails
- **Probability**: Low (startup errors rare in tests)
- **Context**: Application initialization

**Test Gap**: Missing test for app.Start() failure case

**Recommended Fix**:
```go
// Add test case in server_test.go
func TestStart_ApplicationStartupFailure(t *testing.T) {
    // Mock app.Start() to return error
    // Verify server.Start() propagates error correctly
}
```

**Priority**: MEDIUM (startup robustness)

---

## Template Service Lived Mutations (25 total)

### Config Package (6 mutations)

#### 5-10. config/config.go (6 mutations)

**Mutation Types**: CONDITIONALS_NEGATION (2), INCREMENT_DECREMENT (2), CONDITIONALS_BOUNDARY (2)

**Context**: Configuration validation logic

**Severity**: MEDIUM
**Risk**: Invalid config acceptance or valid config rejection

**Test Gap**: Missing edge case validation tests (zero values, boundary conditions)

**Recommended Fix**:
```go
// Add comprehensive validation tests in config_loading_test.go
func TestConfigValidation_EdgeCases(t *testing.T) {
    tests := []struct{
        name string
        config Config
        wantErr bool
    }{
        {"zero browser rate limit", Config{BrowserRateLimit: 0}, true},
        {"zero service rate limit", Config{ServiceRateLimit: 0}, true},
        {"max uint16 port", Config{BindPublicPort: 65535}, false},
        {"min valid port", Config{BindPublicPort: 1}, false},
    }
    // ... test all boundary conditions
}
```

**Priority**: MEDIUM (config robustness)

---

### TLS Generator Package (17 mutations - LOW PRIORITY)

#### 11-27. config/tls_generator/tls_generator.go (17 mutations)

**Mutation Types**:
- CONDITIONALS_BOUNDARY: 9 mutations
- CONDITIONALS_NEGATION: 4 mutations
- ARITHMETIC_BASE: 2 mutations
- INVERT_NEGATIVES: 1 mutation
- INCREMENT_DECREMENT: 1 mutation

**Context**: TLS certificate generation and validation (auto-generated certs for dev/test)

**Severity**: LOW
**Risk**: TLS generation error handling edge cases
- **Impact**: Mostly error path mutations (validityDays <= 0, nil checks, PEM parsing)
- **Probability**: Very low (error paths, fallback logic, rarely executed in production)
- **Context**: Auto-generated certificates used in development/testing only

**Production Usage**: NONE (production uses static TLS certificates, bypasses generator)

**Test Gap**: Missing error path tests for TLS generator

**Recommended Action**: **DEFER** (low ROI)
- TLS generator is infrastructure code for dev/test environments
- Production bypasses generator entirely (static certs)
- Error paths are defensive and rarely executed
- 17 mutations would require ~10-15 hours test development
- **Cost-benefit**: Very low ROI for non-production code

**If Time Permits** (low priority):
```go
// Comprehensive TLS generator error path tests
func TestGenerateTLSMaterial_ErrorPaths(t *testing.T) {
    // Test all nil checks, PEM parsing failures, invalid inputs
    // 17 test cases for 17 mutations
}
```

**Priority**: LOW (defer unless excess capacity)

---

### Realm Service (1 mutation - HIGH PRIORITY)

#### 28. server/service/realm_service.go:435:23 - CONDITIONALS_BOUNDARY

**Location**: `server/service/realm_service.go:435:23`
**Mutation Type**: CONDITIONALS_BOUNDARY
**Original Code**:
```go
if c.MinSecretLength < 1 {
    return fmt.Errorf("min_secret_length must be at least 1")
}
```

**Mutated Code** (likely):
```go
if c.MinSecretLength <= 1 {
    return fmt.Errorf("min_secret_length must be at least 1")
}
```

**Severity**: HIGH
**Risk**: Authentication security boundary
- **Impact**: MinSecretLength=1 rejected instead of accepted
- **Probability**: Medium (common configuration value)
- **Context**: Client secret validation for realm configuration

**Test Gap**: Missing test for MinSecretLength=1 (boundary case)

**Recommended Fix**:
```go
// Add test case in realm_service_test.go
func TestBasicClientIDSecretConfig_ValidateMinSecretLength(t *testing.T) {
    tests := []struct{
        name string
        minLength int
        wantErr bool
    }{
        {"zero length", 0, true},
        {"one character (minimum valid)", 1, false},  // MISSING TEST
        {"normal length", 12, false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            config := &BasicClientIDSecretConfig{MinSecretLength: tt.minLength}
            err := config.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Priority**: HIGH (authentication security)

---

### Registration Service (1 mutation - HIGH PRIORITY)

#### 29. server/service/registration_service.go:232:67 - ARITHMETIC_BASE

**Location**: `server/service/registration_service.go:232:67`
**Mutation Type**: ARITHMETIC_BASE
**Original Code**:
```go
expiresAt := time.Now().UTC().Add(DefaultRegistrationExpiryHours * time.Hour)
```

**Mutated Code** (likely - mutation could be + instead of *, or / instead of*):
```go
expiresAt := time.Now().UTC().Add(DefaultRegistrationExpiryHours + time.Hour)
// OR
expiresAt := time.Now().UTC().Add(DefaultRegistrationExpiryHours / time.Hour)
```

**Severity**: HIGH
**Risk**: Registration expiry calculation
- **Impact**: Wrong expiry time (hours → nanoseconds if +, or divide by 10^9 if /)
- **Probability**: High (all unverified clients affected)
- **Context**: Client registration pending approval

**Test Gap**: Missing test verifying exact expiry duration

**Recommended Fix**:
```go
// Add test case in registration_service_test.go
func TestRegister_UnverifiedClientExpiryDuration(t *testing.T) {
    before := time.Now().UTC()

    result, err := service.Register(ctx, &RegistrationRequest{
        ClientID: "test-client",
        ClientSecret: "test-secret",
        CreateTenant: false,
        ExistingTenantID: &existingTenantID,
    })

    require.NoError(t, err)
    require.NotNil(t, result.ExpiresAt)

    expectedExpiry := before.Add(DefaultRegistrationExpiryHours * time.Hour)
    actualExpiry := *result.ExpiresAt

    // Allow 1-second tolerance for test execution time
    diff := actualExpiry.Sub(expectedExpiry).Abs()
    require.Less(t, diff, 1*time.Second,
        "ExpiresAt should be ~%d hours from now, got %v",
        DefaultRegistrationExpiryHours, actualExpiry)
}
```

**Priority**: HIGH (authorization flow correctness)

---

## Summary by Priority

### HIGH Priority (6 mutations - 2-3h to fix)

**Business Logic - CRITICAL**:
1. ✅ audit_repository.go:101:26 (CONDITIONALS_NEGATION) - Database error handling
2. ✅ audit_repository.go:101:26 (CONDITIONALS_BOUNDARY) - Same location
3. ✅ realm_service.go:435:23 (CONDITIONALS_BOUNDARY) - Secret length validation
4. ✅ registration_service.go:232:67 (ARITHMETIC_BASE) - Expiry calculation

**Server Startup - IMPORTANT**:
5. ✅ server.go:130:34 (CONDITIONALS_NEGATION) - App startup error handling

**Audit Sampling - MEDIUM-HIGH**:
6. ✅ audit_repository.go:112:24 (CONDITIONALS_BOUNDARY) - Sampling boundary

**Estimated Effort**: 2-3 hours (6 test cases)

---

### MEDIUM Priority (6 mutations - 2h to fix)

**Config Validation**:
7-12. ✅ config/config.go (6 mutations) - Validation edge cases

**Estimated Effort**: 2 hours (comprehensive validation test suite)

---

### LOW Priority (17 mutations - DEFER)

**TLS Generator Infrastructure**:
13-29. ⚠️ config/tls_generator/tls_generator.go (17 mutations) - Error paths

**Rationale for Deferral**:
- Non-production code (dev/test only)
- Error handling edge cases
- High effort (10-15h), very low ROI
- Production uses static certs (bypasses generator)

**Estimated Effort**: 10-15 hours (if implemented)

---

## Mutation Killing Strategy

### Phase 1: High-Priority Mutations (Task 6.3.1)

**Target**: Kill 6 high-priority mutations
**Estimated**: 2-3 hours
**Approach**: Targeted test cases for critical business logic

**Files to Modify**:
- `internal/apps/jose/ja/repository/audit_repository_test.go` (3 tests)
- `internal/apps/jose/ja/server/server_test.go` (1 test)
- `internal/apps/template/service/server/service/realm_service_test.go` (1 test)
- `internal/apps/template/service/server/service/registration_service_test.go` (1 test)

**Expected Outcome**:
- JOSE-JA efficacy: 96.15% → ~98% (kill 3 of 4 lived)
- Template efficacy: 91.75% → ~93% (kill 2 of 25 lived)

---

### Phase 2: Medium-Priority Mutations (Task 6.3.2)

**Target**: Kill 6 config validation mutations
**Estimated**: 2 hours
**Approach**: Comprehensive edge case validation suite

**Files to Modify**:
- `internal/apps/template/service/config/config_loading_test.go`

**Expected Outcome**:
- Template efficacy: ~93% → ~95% (kill 6 more lived)

---

### Phase 3: Low-Priority Mutations (Task 6.3.3 - OPTIONAL)

**Target**: Kill 17 TLS generator mutations
**Estimated**: 10-15 hours (if pursued)
**Approach**: Error path testing for TLS certificate generation

**Decision**: **DEFER** (low ROI, non-production code)

**Expected Outcome** (if implemented):
- Template efficacy: ~95% → ~98% (kill 17 more lived)

---

## Success Metrics

### Minimum Viable (Phase 1 + 2)

**Mutation Kills**: 12 of 29 lived mutations (41% reduction)
**Efficacy Targets**:
- JOSE-JA: 96.15% → ~98% ✅ (exceeds 85% baseline)
- Template: 91.75% → ~95% ✅ (exceeds 85% baseline)

**Time Investment**: 4-5 hours
**ROI**: HIGH (critical business logic coverage)

---

### Stretch Goal (Phase 1 + 2 + 3)

**Mutation Kills**: 29 of 29 lived mutations (100% reduction)
**Efficacy Targets**:
- JOSE-JA: 96.15% → ~99% ✅
- Template: 91.75% → ~98% ✅

**Time Investment**: 14-20 hours
**ROI**: MEDIUM (includes low-value infrastructure code)

---

## Next Steps

1. ✅ Document this analysis (current file)
2. ⏳ Commit analysis: "docs(mutation): analyze 29 lived mutations by priority"
3. ⏳ Task 6.3: Implement Phase 1 tests (HIGH priority - 6 mutations)
4. ⏳ Task 6.3: Implement Phase 2 tests (MEDIUM priority - 6 mutations)
5. ⏳ Re-run gremlins, verify efficacy ≥95% for both services
6. ⏳ Task 6.4: Enable CI/CD mutation testing with 85% threshold
7. ⚠️ DEFER Phase 3 (TLS generator) unless excess capacity

---

## References

- Baseline Results: docs/gremlins/mutation-baseline-results.md
- Gremlins Logs: /tmp/gremlins_jose_ja.log, /tmp/gremlins_template.log
- Configuration: .gremlins.yml (180s timeout, 85% threshold)
- Commits: 00399210 (template fix), 3e23ef86 (baseline), 992479f9 (task tracking)
