# Gremlins Mutation Testing - Task List

**Purpose**: Comprehensive gremlins mutation testing coverage for all high-value business logic packages.

**Target**: 98%+ test efficacy for each package

**Execution Pattern**: Run commands sequentially, analyze results, document findings

**Note**: Gremlins uses system TEMP/TMP environment variables by default. Ensure adequate disk space (>8GB free) on system temp drive.

---

## KMS (Key Management Service) - 10 Tasks

### Core Business Logic

```powershell
# Task 1: KMS Business Logic (Core Operations)
gremlins unleash --workers 2 --tags '!integration' ./internal/kms/server/businesslogic 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task01_kms_businesslogic.txt
```

```powershell
# Task 2: KMS Barrier - Content Keys Service
gremlins unleash --workers 1 --tags '!integration' ./internal/shared/barrier/contentkeysservice 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task02_kms_barrier_content.txt
```

```powershell
# Task 3: KMS Barrier - Intermediate Keys Service
gremlins unleash --workers 1 --tags '!integration' ./internal/shared/barrier/intermediatekeysservice 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task03_kms_barrier_intermediate.txt
```

```powershell
# Task 4: KMS Barrier - Root Keys Service
gremlins unleash --workers 1 --tags '!integration' ./internal/shared/barrier/rootkeysservice 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task04_kms_barrier_root.txt
```

```powershell
# Task 5: KMS Barrier - Unseal Service
gremlins unleash --workers 1 --tags '!integration' ./internal/kms/server/barrier/unsealservice 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task05_kms_barrier_unseal.txt
```

### Repository Layers

```powershell
# Task 6: KMS ORM Repository
gremlins unleash --workers 2 --tags '!integration' ./internal/kms/server/repository/orm 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task06_kms_orm_repository.txt
```

```powershell
# Task 7: KMS SQL Repository
gremlins unleash --workers 2 --tags '!integration' ./internal/kms/server/repository/sqlrepository 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task07_kms_sql_repository.txt
```

### Handler and Middleware

```powershell
# Task 8: KMS Handlers (OpenAPI Strict Server)
gremlins unleash --workers 2 --tags '!integration' ./internal/kms/server/handler 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task08_kms_handlers.txt
```

```powershell
# Task 9: KMS Middleware
gremlins unleash --workers 2 --tags '!integration' ./internal/kms/server/middleware 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task09_kms_middleware.txt
```

```powershell
# Task 10: KMS Client
gremlins unleash --workers 2 --tags '!integration' ./internal/kms/client 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task10_kms_client.txt
```

---

## Identity (OAuth 2.1 + OIDC) - 30 Tasks

### Authorization Server (authz)

```powershell
# Task 11: Identity Authorization Server (COMPLETED - 91% efficacy baseline)
# Baseline established 2025-12-17: 91 killed, 9 lived, 188 not covered
# Result file: test-output/gremlins/authz_c_drive_final.txt
```

```powershell
# Task 12: Identity Authorization - Client Authentication
gremlins unleash --workers 1 --tags '!integration' ./internal/identity/authz/clientauth 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task12_identity_authz_clientauth.txt
```

```powershell
# Task 13: Identity Authorization - DPoP
gremlins unleash --workers 1 --tags '!integration' ./internal/identity/authz/dpop 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task13_identity_authz_dpop.txt
```

```powershell
# Task 14: Identity Authorization - PAR (Pushed Authorization Requests)
gremlins unleash --workers 1 --tags '!integration' ./internal/identity/authz/par 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task14_identity_authz_par.txt
```

```powershell
# Task 15: Identity Authorization - Storage
gremlins unleash --workers 1 --tags '!integration' ./internal/identity/authz/storage 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task15_identity_authz_storage.txt
```

### Identity Provider (idp)

```powershell
# Task 16: Identity Provider - Core
gremlins unleash --workers 1 --tags '!integration' ./internal/identity/idp 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task16_identity_idp.txt
```

```powershell
# Task 17: Identity Provider - MFA
gremlins unleash --workers 1 --tags '!integration' ./internal/identity/idp/mfa 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task17_identity_idp_mfa.txt
```

```powershell
# Task 18: Identity Provider - Authentication Methods
gremlins unleash --workers 1 --tags '!integration' ./internal/identity/idp/authmethods 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task18_identity_idp_authmethods.txt
```

```powershell
# Task 19: Identity Provider - Session Management
gremlins unleash --workers 1 --tags '!integration' ./internal/identity/idp/session 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task19_identity_idp_session.txt
```

### Domain Models and Mappers

```powershell
# Task 20: Identity Domain Models
gremlins unleash --workers 2 --tags '!integration' ./internal/identity/domain 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task20_identity_domain.txt
```

```powershell
# Task 21: Identity Repository - ORM
gremlins unleash --workers 2 --tags '!integration' ./internal/identity/repository/orm 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task21_identity_orm_repository.txt
```

```powershell
# Task 22: Identity Repository - SQL
gremlins unleash --workers 2 --tags '!integration' ./internal/identity/repository/sqlrepository 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task22_identity_sql_repository.txt
```

### Additional Identity Packages

```powershell
# Task 23-40: Reserved for additional Identity subpackages (handlers, middleware, clients)
```

---

## JOSE (Cryptographic Operations) - 20 Tasks

### Core Crypto

```powershell
# Task 41: JOSE Crypto - Core (LARGE PACKAGE - may fail during coverage gathering)
gremlins unleash --workers 2 --tags '!integration' ./internal/jose/crypto 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task41_jose_crypto.txt
```

```powershell
# Task 42: JOSE Crypto - JWK Utilities
gremlins unleash --workers 2 --tags '!integration' ./internal/jose/crypto/jwk 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task42_jose_jwk.txt
```

```powershell
# Task 43: JOSE Crypto - JWKGen Service
gremlins unleash --workers 2 --tags '!integration' ./internal/jose/crypto/jwkgen 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task43_jose_jwkgen.txt
```

```powershell
# Task 44: JOSE Crypto - Keygen
gremlins unleash --workers 2 --tags '!integration' ./internal/jose/crypto/keygen 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task44_jose_keygen.txt
```

### JOSE Server

```powershell
# Task 45: JOSE Server - Core
gremlins unleash --workers 2 --tags '!integration' ./internal/jose/server 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task45_jose_server.txt
```

```powershell
# Task 46: JOSE Server - Keystore
gremlins unleash --workers 2 --tags '!integration' ./internal/jose/server/keystore 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task46_jose_keystore.txt
```

```powershell
# Task 47: JOSE Server - Middleware
gremlins unleash --workers 2 --tags '!integration' ./internal/jose/server/middleware 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task47_jose_middleware.txt
```

```powershell
# Task 48-60: Reserved for additional JOSE subpackages
```

---

## CA (Certificate Authority) - 20 Tasks

### Certificate Issuance

```powershell
# Task 61: CA Issuer
gremlins unleash --workers 2 --tags '!integration' ./internal/ca/issuer 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task61_ca_issuer.txt
```

```powershell
# Task 62: CA Certificate Profiles
gremlins unleash --workers 2 --tags '!integration' ./internal/ca/profiles 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task62_ca_profiles.txt
```

```powershell
# Task 63: CA Serial Number Generation
gremlins unleash --workers 2 --tags '!integration' ./internal/ca/serial 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task63_ca_serial.txt
```

### CA Protocol Handlers

```powershell
# Task 64: CA CMP Handler
gremlins unleash --workers 2 --tags '!integration' ./internal/ca/handlers/cmp 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task64_ca_cmp.txt
```

```powershell
# Task 65: CA CMPv2 Handler
gremlins unleash --workers 2 --tags '!integration' ./internal/ca/handlers/cmpv2 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task65_ca_cmpv2.txt
```

```powershell
# Task 66: CA SCEP Handler
gremlins unleash --workers 2 --tags '!integration' ./internal/ca/handlers/scep 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task66_ca_scep.txt
```

```powershell
# Task 67: CA EST Handler
gremlins unleash --workers 2 --tags '!integration' ./internal/ca/handlers/est 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task67_ca_est.txt
```

```powershell
# Task 68: CA OCSP Handler
gremlins unleash --workers 2 --tags '!integration' ./internal/ca/handlers/ocsp 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task68_ca_ocsp.txt
```

```powershell
# Task 69: CA CRL Distribution Point
gremlins unleash --workers 2 --tags '!integration' ./internal/ca/handlers/crldp 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task69_ca_crldp.txt
```

```powershell
# Task 70-80: Reserved for additional CA subpackages (RA, timestamping, validation)
```

---

## Shared Utilities (High-Value Only) - 20 Tasks

### Cryptographic Utilities

```powershell
# Task 81: Shared Crypto - Certificate
gremlins unleash --workers 2 --tags '!integration' ./internal/shared/crypto/certificate 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task81_shared_crypto_certificate.txt
```

```powershell
# Task 82: Shared Crypto - TLS
gremlins unleash --workers 2 --tags '!integration' ./internal/shared/crypto/tls 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task82_shared_crypto_tls.txt
```

```powershell
# Task 83: Shared Crypto - Hashing
gremlins unleash --workers 2 --tags '!integration' ./internal/shared/crypto/hash 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task83_shared_crypto_hash.txt
```

### Validation Utilities

```powershell
# Task 84: Shared Util - Random (UUID validation)
gremlins unleash --workers 2 --tags '!integration' ./internal/shared/util/random 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task84_shared_util_random.txt
```

```powershell
# Task 85: Shared Util - Validation
gremlins unleash --workers 2 --tags '!integration' ./internal/shared/util/validation 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task85_shared_util_validation.txt
```

```powershell
# Task 86: Shared Util - Strings
gremlins unleash --workers 2 --tags '!integration' ./internal/shared/util/strings 2>&1 | Tee-Object -FilePath ./test-output/gremlins/task86_shared_util_strings.txt
```

```powershell
# Task 87-100: Reserved for additional shared utility packages
```

---

## Analysis Commands

### Count Results for Single Task

```powershell
# Replace task01 with actual task number
$content = Get-Content ./test-output/gremlins/task01_kms_businesslogic.txt
$killed = ($content | Select-String "^\s+KILLED").Count
$lived = ($content | Select-String "^\s+LIVED").Count
$notCovered = ($content | Select-String "^\s+NOT COVERED").Count
Write-Host "Task 1 - KMS Business Logic:"
Write-Host "  Killed: $killed, Lived: $lived, Not Covered: $notCovered"
if (($killed + $lived) -gt 0) {
    $efficacy = [math]::Round($killed / ($killed + $lived) * 100, 2)
    Write-Host "  Test Efficacy: $efficacy%"
    if ($efficacy -ge 98) { Write-Host "  ✅ PASSING (≥98%)" -ForegroundColor Green }
    elseif ($efficacy -ge 70) { Write-Host "  ⚠️ WARNING (70-79%)" -ForegroundColor Yellow }
    else { Write-Host "  ❌ FAILING (<70%)" -ForegroundColor Red }
}
```

### Batch Analysis for All Completed Tasks

```powershell
# Analyze all gremlins results in test-output/gremlins/
Get-ChildItem ./test-output/gremlins/task*.txt | ForEach-Object {
    $content = Get-Content $_.FullName
    $killed = ($content | Select-String "^\s+KILLED").Count
    $lived = ($content | Select-String "^\s+LIVED").Count
    if (($killed + $lived) -gt 0) {
        $efficacy = [math]::Round($killed / ($killed + $lived) * 100, 2)
        $status = if ($efficacy -ge 98) { "✅" } elseif ($efficacy -ge 90) { "⚠️" } else { "❌" }
        Write-Host "$status $($_.BaseName): $efficacy% ($killed killed, $lived lived)"
    }
}
```

---

## Execution Strategy

### Sequential Execution (Recommended)

1. Start with Task 1 (KMS businesslogic)
2. Analyze results, document findings
3. If efficacy <98%, investigate lived mutants
4. Continue to Task 2, repeat

### Parallel Execution (Advanced)

Run multiple tasks simultaneously in separate terminals, but limit concurrency to avoid resource contention:

```powershell
# Terminal 1
# Run Task 1

# Terminal 2
# Run Task 2

# Terminal 3
# Run Task 3
```

**Warning**: Running >3 gremlins tasks in parallel may exhaust disk space or memory.

---

## Expected Outcomes

- **80%+ efficacy**: Package has strong test coverage
- **70-79% efficacy**: Acceptable, but investigate lived mutants
- **<70% efficacy**: Significant test quality gaps, requires attention

**Baseline Targets by Package Type**:

- Business logic: ≥95% minimum (98% ideal) (KMS, Identity authz, CA issuer)
- Mappers/validators: ≥90% (oam_orm_mapper, domain models)
- Handlers/middleware: ≥80% (OpenAPI handlers, middleware)
- Utilities: ≥90% (shared crypto, validation)
