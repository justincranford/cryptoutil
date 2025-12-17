# Gremlins Mutation Testing - Quick Start Guide

**Purpose**: Run gremlins mutation testing on high-value business logic packages to identify test quality gaps.

**Target Packages**: Business logic with complex validation, mappers, and post-OpenAPI custom logic.

## Prerequisites

```powershell
# Install gremlins (if not already installed)
go install github.com/go-gremlins/gremlins/cmd/gremlins@latest

# Verify installation
gremlins --version

# Clean caches to avoid permission issues
go clean -cache -testcache
```

## Quick Start Commands

### Use C: Drive (Avoid R: Drive Permission Issues)

```powershell
# Set temp directories to C: drive
$env:TMPDIR = "C:\Temp"
$env:TEMP = "C:\Temp"
$env:TMP = "C:\Temp"
$env:GOCACHE = "C:\Temp\go-cache"

# Run gremlins with output capture
gremlins unleash --workers 2 --tags '!integration' ./path/to/package 2>&1 | Tee-Object -FilePath ./test-output/gremlins/package_name.txt
```

### Target Efficacy

- **Minimum**: 80% test efficacy (killed / (killed + lived))
- **Good**: 85-90% test efficacy
- **Excellent**: 90%+ test efficacy

## High-Value Packages

### KMS (Key Management Service)

#### 1. Business Logic (Core Operations)

```powershell
# internal/kms/server/businesslogic - Encrypt/Decrypt/Sign/Verify operations
$env:TMPDIR = "C:\Temp"; $env:GOCACHE = "C:\Temp\go-cache"
gremlins unleash --workers 2 --tags '!integration' ./internal/kms/server/businesslogic 2>&1 | Tee-Object -FilePath ./test-output/gremlins/kms_businesslogic.txt
```

**What to look for**:

- Validation logic in mapper functions (toOam*, toOrm*)
- Error handling in PostEncrypt, PostDecrypt, PostSign, PostVerify
- UUID validation, pagination logic, algorithm mapping

#### 2. ORM Mappers

```powershell
# Focus on mapper validation (already unit tested)
gremlins unleash --workers 2 --tags '!integration' ./internal/kms/server/businesslogic 2>&1 | Select-String -Pattern "oam_orm_mapper" | Tee-Object -FilePath ./test-output/gremlins/kms_mappers.txt
```

**Critical paths**:

- `toOptionalOrmUUIDs` - UUID validation
- `toOrmPageNumber`, `toOrmPageSize` - pagination bounds
- `toOrmElasticKeySorts`, `toOrmMaterialKeySorts` - sort validation

#### 3. Barrier Services (Unseal/Encryption)

```powershell
# internal/kms/server/barrier - Key hierarchy operations
$env:TMPDIR = "C:\Temp"; $env:GOCACHE = "C:\Temp\go-cache"
gremlins unleash --workers 1 --tags '!integration' ./internal/kms/server/barrier/contentkeysservice 2>&1 | Tee-Object -FilePath ./test-output/gremlins/kms_barrier_content.txt

gremlins unleash --workers 1 --tags '!integration' ./internal/kms/server/barrier/intermediatekeysservice 2>&1 | Tee-Object -FilePath ./test-output/gremlins/kms_barrier_intermediate.txt
```

**What to look for**:

- Key unwrapping logic
- Nil validation in constructors
- Error propagation from crypto operations

### Identity (OAuth 2.1 + OIDC)

#### 1. Authorization Server (Token Validation)

```powershell
# internal/identity/authz - OAuth token handling
$env:TMPDIR = "C:\Temp"; $env:GOCACHE = "C:\Temp\go-cache"
gremlins unleash --workers 1 --tags '!integration' ./internal/identity/authz 2>&1 | Tee-Object -FilePath ./test-output/gremlins/identity_authz.txt
```

**Results**: 91% efficacy (baseline established 2025-12-17)

**Critical paths**:

- Client authentication (basic, JWT, mTLS)
- Authorization request validation
- Token generation and introspection

#### 2. Identity Provider (MFA Logic)

```powershell
# internal/identity/idp - Multi-factor authentication
$env:TMPDIR = "C:\Temp"; $env:GOCACHE = "C:\Temp\go-cache"
gremlins unleash --workers 1 --tags '!integration' ./internal/identity/idp 2>&1 | Tee-Object -FilePath ./test-output/gremlins/identity_idp.txt
```

**What to look for**:

- MFA chain validation
- TOTP/HOTP verification
- WebAuthn challenge generation

### JOSE (Cryptographic Operations)

#### 1. Crypto Core

```powershell
# internal/jose/crypto - JWE/JWS operations
$env:TMPDIR = "C:\Temp"; $env:GOCACHE = "C:\Temp\go-cache"
gremlins unleash --workers 2 --tags '!integration' ./internal/jose/crypto 2>&1 | Tee-Object -FilePath ./test-output/gremlins/jose_crypto.txt
```

**Known Issue**: Large packages may fail during coverage gathering. Try smaller subpackages if this occurs.

**What to look for**:

- JWK validation (CreateJWK*, Is*, Extract*)
- Algorithm mapping (ToJWEEncAndAlg, ToJWSAlg)
- Key generation (GenerateJWEJWK, GenerateJWSJWK)

#### 2. JOSE Server Business Logic

```powershell
# internal/jose/server - JOSE service operations
$env:TMPDIR = "C:\Temp"; $env:GOCACHE = "C:\Temp\go-cache"
gremlins unleash --workers 2 --tags '!integration' ./internal/jose/server 2>&1 | Tee-Object -FilePath ./test-output/gremlins/jose_server.txt
```

### CA (Certificate Authority)

#### 1. Certificate Issuance

```powershell
# internal/ca/issuer - Certificate generation logic
$env:TMPDIR = "C:\Temp"; $env:GOCACHE = "C:\Temp\go-cache"
gremlins unleash --workers 2 --tags '!integration' ./internal/ca/issuer 2>&1 | Tee-Object -FilePath ./test-output/gremlins/ca_issuer.txt
```

**What to look for**:

- Serial number generation (must be CSPRNG, >64 bits, <2^159)
- Certificate profile validation
- Extension handling

#### 2. CA Handlers

```powershell
# internal/ca/handlers - CMP/CMPv2/SCEP/EST endpoints
$env:TMPDIR = "C:\Temp"; $env:GOCACHE = "C:\Temp\go-cache"
gremlins unleash --workers 2 --tags '!integration' ./internal/ca/handlers 2>&1 | Tee-Object -FilePath ./test-output/gremlins/ca_handlers.txt
```

## Analyzing Results

### Success Criteria

```powershell
# Count results from output file
$content = Get-Content ./test-output/gremlins/package_name.txt
$killed = ($content | Select-String "^\s+KILLED").Count
$lived = ($content | Select-String "^\s+LIVED").Count
$notCovered = ($content | Select-String "^\s+NOT COVERED").Count

Write-Host "Killed: $killed, Lived: $lived, Not Covered: $notCovered"
Write-Host "Test Efficacy: $([math]::Round($killed / ($killed + $lived) * 100, 2))%"
```

### Interpreting Results

- **KILLED**: Test successfully caught the mutation (good!)
- **LIVED**: Mutation survived (test gap - investigate)
- **NOT COVERED**: Code not executed by tests (add coverage)
- **TIMED OUT**: Mutation caused infinite loop (usually defensive code)

### Common Lived Mutants (Not Always Bad)

1. **Defensive checks**: `if x < 0` → `if x >= 0` when x is always positive
2. **Boundary conditions**: `len(slice) > 0` → `len(slice) >= 0` when empty slice handled elsewhere
3. **Error message variations**: String literal mutations in error messages

## Troubleshooting

### Issue: "Access is denied" on R: drive

**Solution**: Use C: drive temp directories (see commands above)

### Issue: Fails during "Gathering coverage"

**Causes**:

- Package too large (many test files, many variants)
- Test parallelism issues

**Solutions**:

1. Reduce workers: `--workers 1`
2. Split package into smaller subpackages
3. Use `--dry-run` to see mutation generation without execution
4. Run on smaller test files individually

### Issue: Tests hang or timeout

**Cause**: Mutation creates infinite loop

**Solution**: Normal behavior for certain mutations (e.g., loop counters). Gremlins detects this and marks as "TIMED OUT".

## Best Practices

1. **Run after test optimizations**: Ensure tests run fast (<20s) before gremlins
2. **Focus on business logic**: Prioritize validation, mappers, error handling
3. **Ignore infrastructure**: Skip cicd packages, test utilities, main functions
4. **Target 80%+ efficacy**: Below 80% indicates significant test quality gaps
5. **Document baselines**: Save results to track improvements over time
6. **Investigate lived mutants**: Each survivor reveals a test gap or design issue

## Next Steps

After running gremlins:

1. Review lived mutants in output file
2. Add tests for uncovered code paths
3. Strengthen assertions for lived mutants
4. Re-run gremlins to verify improvements
5. Document baseline efficacy for future reference

## Example Workflow

```powershell
# 1. Set environment
$env:TMPDIR = "C:\Temp"; $env:GOCACHE = "C:\Temp\go-cache"

# 2. Run gremlins
gremlins unleash --workers 2 --tags '!integration' ./internal/kms/server/businesslogic 2>&1 | Tee-Object -FilePath ./test-output/gremlins/kms_businesslogic.txt

# 3. Analyze results
$content = Get-Content ./test-output/gremlins/kms_businesslogic.txt
$killed = ($content | Select-String "^\s+KILLED").Count
$lived = ($content | Select-String "^\s+LIVED").Count
Write-Host "Efficacy: $([math]::Round($killed / ($killed + $lived) * 100, 2))%"

# 4. Review lived mutants
Get-Content ./test-output/gremlins/kms_businesslogic.txt | Select-String "LIVED"

# 5. Add missing tests for lived mutants

# 6. Re-run to verify improvement
```

## Reference: Successful Baselines

- **format_go**: 91.67% efficacy (28s, 33 killed, 3 lived)
- **internal/identity/authz**: 91% efficacy (69s, 91 killed, 9 lived)
- **Target for all packages**: ≥80% efficacy
## Known Limitations

### Gremlins Tool Issues

1. **Coverage Gathering Hangs**:
   - Medium/large packages (>10 files, >500 lines) consistently hang during "Gathering coverage..." phase
   - Complex test suites with many dependencies cannot complete coverage analysis
   - Affects: KMS businesslogic, barrier services (root, unseal), JOSE server, Identity domain
   - **Success Rate**: Only 2 of 6 packages completed (33%)

2. **Gremlins Panic Errors**:
   - Internal error: "error, this is temporary" in executor.go:165
   - Affects: KMS barrier intermediate service
   - Root cause: Internal gremlins bug (not test quality issue)

3. **Test Timeouts**:
   - Tests taking >30s cause mutants to TIMED OUT
   - Especially affects crypto operations with real key sizes
   - Solution: Use --timeout=60s or --timeout=120s flag

4. **Disk Space Requirements**:
   - R: drive (Go cache) fills up quickly during parallel mutation testing
   - Requires >8GB free space for medium packages
   - Solution: Clear caches regularly: go clean -cache -modcache -testcache

### When NOT to Use Gremlins

- ❌ Packages with >10 test files
- ❌ Packages with complex dependency graphs
- ❌ Packages with integration tests (even with --tags '!integration')
- ❌ Packages where tests take >20 seconds to run
- ❌ When disk space is limited (<10GB free)

### Alternative: Focus on Test Coverage

**Recommendation**: If gremlins consistently hangs, focus on improving test coverage (95%+ target) instead:

\\\powershell
# Generate coverage report
go test ./internal/package -coverprofile=./test-output/coverage.out

# View HTML report (identify RED uncovered lines)
go tool cover -html=./test-output/coverage.out -o ./test-output/coverage.html

# Add tests for uncovered branches, validate with runTests tool
\\\

High test coverage (95%+) is more reliable indicator of test quality than mutation testing when gremlins tool has limitations.
