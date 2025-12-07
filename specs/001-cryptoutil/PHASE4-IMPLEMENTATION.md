# Phase 4: Advanced Testing Implementation Guide

**Duration**: Days 10-11 (4-6 hours)  
**Prerequisites**: Phase 3 complete (all coverage targets achieved)  
**Status**: ❌ Not Started

## Overview

Phase 4 adds advanced testing methodologies to strengthen code quality and reliability beyond traditional unit testing. Per 01-02.testing.instructions.md:

- **Benchmarking**: MANDATORY for cryptographic operations (track performance)
- **Fuzz Testing**: MANDATORY for parsers, validators, input handlers
- **Property-Based Testing**: RECOMMENDED for mathematical invariants
- **Mutation Testing**: MANDATORY ≥80% gremlins score per package

**Task Breakdown**:

- P4.1: Add Benchmark Tests (2h) - Performance baselines for crypto operations
- P4.2: Add Fuzz Tests (2h) - Security-focused input validation
- P4.3: Add Property-Based Tests (2h) - Mathematical invariant validation
- P4.4: Mutation Testing Baseline (1h) - Code quality measurement

## Task Details

---

### P4.1: Add Benchmark Tests ⭐ CRITICAL

**Priority**: CRITICAL  
**Effort**: 2 hours  
**Status**: ❌ Not Started

**Objective**: Create benchmarks for ALL cryptographic operations to establish performance baselines and detect regressions.

**Current State**:

- No systematic benchmarks for crypto operations
- No performance baseline tracking
- No regression detection in CI/CD

**Implementation Strategy**:

```bash
# Step 1: Identify crypto operations to benchmark
# Target packages:
# - internal/common/crypto/keygen (key generation)
# - internal/jose/* (JOSE operations)
# - internal/ca/crypto/* (CA crypto operations)
# - internal/kms/server/crypto/* (KMS operations)

# Step 2: Create benchmark test files
mkdir -p internal/common/crypto/keygen
touch internal/common/crypto/keygen/rsa_bench_test.go
touch internal/common/crypto/keygen/ecdsa_bench_test.go
touch internal/common/crypto/keygen/eddsa_bench_test.go
touch internal/common/crypto/keygen/aes_bench_test.go
```

**Files to Create**:

- `internal/common/crypto/keygen/rsa_bench_test.go`
- `internal/common/crypto/keygen/ecdsa_bench_test.go`
- `internal/common/crypto/keygen/eddsa_bench_test.go`
- `internal/common/crypto/keygen/aes_bench_test.go`
- `internal/common/crypto/keygen/hmac_bench_test.go`
- `internal/jose/jws_bench_test.go`
- `internal/jose/jwe_bench_test.go`
- `internal/ca/crypto/sign_bench_test.go`

**Benchmark Pattern for Crypto Operations**:

```go
// File: internal/common/crypto/keygen/rsa_bench_test.go
package keygen_test

import (
    "testing"
    "github.com/stretchr/testify/require"
)

// BenchmarkRSAKeyGeneration benchmarks RSA key generation for multiple key sizes
func BenchmarkRSAKeyGeneration(b *testing.B) {
    keySizes := []int{2048, 3072, 4096}

    for _, keySize := range keySizes {
        b.Run(fmt.Sprintf("RSA-%d", keySize), func(b *testing.B) {
            b.ReportAllocs()
            b.ResetTimer()

            for i := 0; i < b.N; i++ {
                _, err := keygen.GenerateRSAKey(keySize)
                if err != nil {
                    b.Fatal(err)
                }
            }
        })
    }
}

// BenchmarkRSASigning benchmarks RSA signing operations
func BenchmarkRSASigning(b *testing.B) {
    // Setup: generate key once
    key, err := keygen.GenerateRSAKey(2048)
    require.NoError(b, err)

    message := []byte("benchmark message for signing")

    b.ReportAllocs()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        _, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, message)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// BenchmarkRSAVerification benchmarks RSA signature verification
func BenchmarkRSAVerification(b *testing.B) {
    // Setup: generate key and signature once
    key, _ := keygen.GenerateRSAKey(2048)
    message := []byte("benchmark message for verification")
    signature, _ := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, message)

    b.ReportAllocs()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        err := rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, message, signature)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

**JOSE Benchmark Example**:

```go
// File: internal/jose/jws_bench_test.go
package jose_test

import (
    "testing"
)

func BenchmarkJWSSign(b *testing.B) {
    algorithms := []string{"RS256", "ES256", "EdDSA"}

    for _, alg := range algorithms {
        b.Run(alg, func(b *testing.B) {
            // Setup: generate key for algorithm
            key := generateKeyForAlgorithm(b, alg)
            payload := []byte(`{"sub":"1234567890","name":"John Doe"}`)

            b.ReportAllocs()
            b.ResetTimer()

            for i := 0; i < b.N; i++ {
                _, err := jws.Sign(payload, key, alg)
                if err != nil {
                    b.Fatal(err)
                }
            }
        })
    }
}

func BenchmarkJWSVerify(b *testing.B) {
    algorithms := []string{"RS256", "ES256", "EdDSA"}

    for _, alg := range algorithms {
        b.Run(alg, func(b *testing.B) {
            // Setup: generate key and signature once
            key := generateKeyForAlgorithm(b, alg)
            payload := []byte(`{"sub":"1234567890","name":"John Doe"}`)
            signature, _ := jws.Sign(payload, key, alg)

            b.ReportAllocs()
            b.ResetTimer()

            for i := 0; i < b.N; i++ {
                _, err := jws.Verify(signature, key)
                if err != nil {
                    b.Fatal(err)
                }
            }
        })
    }
}
```

**Crypto Operations to Benchmark**:

| Category | Operations | Expected Performance |
|----------|------------|---------------------|
| RSA Key Generation | 2048, 3072, 4096 bits | 10-100ms per key |
| ECDSA Key Generation | P-256, P-384, P-521 | 1-10ms per key |
| EdDSA Key Generation | Ed25519 | <1ms per key |
| AES Encryption | 128, 192, 256 bits | <1μs per block |
| HMAC | SHA-256, SHA-512 | <1μs per operation |
| JWS Sign | RS256, ES256, EdDSA | 1-10ms per operation |
| JWS Verify | RS256, ES256, EdDSA | 1-5ms per operation |
| JWE Encrypt | Various algorithms | 1-10ms per operation |
| JWE Decrypt | Various algorithms | 1-10ms per operation |

**Acceptance Criteria**:

- ✅ Benchmarks for ALL cryptographic operations (happy and sad paths)
- ✅ File suffix: `_bench_test.go`
- ✅ Use `b.ReportAllocs()` to track memory allocations
- ✅ Use `b.ResetTimer()` after setup
- ✅ Multiple key sizes/algorithms tested
- ✅ Baselines documented in specs/benchmarks/
- ✅ CI/CD tracks performance regressions

**Validation Commands**:

```bash
# Run all benchmarks
go test -bench=. -benchmem ./internal/common/crypto/keygen
go test -bench=. -benchmem ./internal/jose
go test -bench=. -benchmem ./internal/ca/crypto

# Save baseline
go test -bench=. -benchmem ./... > specs/benchmarks/baseline-$(date +%Y%m%d).txt

# Compare against baseline
benchstat specs/benchmarks/baseline-20240101.txt <(go test -bench=. -benchmem ./...)
```

**CI/CD Integration**:

After P4.1 complete, verify ci-benchmark workflow uses these benchmarks (P1.4).

---

### P4.2: Add Fuzz Tests ⭐ CRITICAL

**Priority**: CRITICAL  
**Effort**: 2 hours  
**Status**: ❌ Not Started

**Objective**: Create fuzz tests for parsers, validators, and input handlers to discover edge cases and security vulnerabilities.

**Current State**:

- No systematic fuzz testing
- Input validation not fuzz-tested
- Parsers not tested with malformed input

**Implementation Strategy**:

```bash
# Step 1: Identify fuzz test targets
# Per 01-02.testing.instructions.md:
# - Parsers (JWT, CSR, certificates)
# - Validators (input validation, format validation)
# - Input handlers (API endpoints accepting user input)

# Step 2: Create fuzz test files
touch internal/jose/jwt_parser_fuzz_test.go
touch internal/ca/parser/csr_fuzz_test.go
touch internal/ca/parser/cert_fuzz_test.go
touch internal/identity/auth/password_fuzz_test.go
```

**Files to Create**:

- `internal/jose/jwt_parser_fuzz_test.go`
- `internal/jose/jwk_parser_fuzz_test.go`
- `internal/ca/parser/csr_fuzz_test.go`
- `internal/ca/parser/cert_fuzz_test.go`
- `internal/identity/auth/password_fuzz_test.go`
- `internal/identity/authz/scope_fuzz_test.go`

**Fuzz Test Pattern**:

```go
// File: internal/jose/jwt_parser_fuzz_test.go
package jose_test

import (
    "testing"
)

// FuzzJWTParser fuzzes JWT parsing to find edge cases and crashes
func FuzzJWTParser(f *testing.F) {
    // Seed corpus with valid JWTs
    f.Add([]byte("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature"))
    f.Add([]byte("eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature"))

    // Fuzz with random mutations
    f.Fuzz(func(t *testing.T, jwtBytes []byte) {
        // Parser should never panic, always return error for invalid input
        _, err := jwt.Parse(string(jwtBytes))

        // We expect many errors for malformed input, but no panics
        if err != nil {
            // Validate error is well-formed
            if err.Error() == "" {
                t.Errorf("empty error message for invalid JWT")
            }
        }
    })
}

// FuzzJWTValidation fuzzes JWT validation logic
func FuzzJWTValidation(f *testing.F) {
    // Seed corpus
    f.Add([]byte("header"), []byte("payload"), []byte("signature"))

    f.Fuzz(func(t *testing.T, header, payload, signature []byte) {
        // Construct JWT from parts
        jwtStr := string(header) + "." + string(payload) + "." + string(signature)

        // Validation should never panic
        _, err := jwt.Validate(jwtStr)

        // Expect errors for most random input
        _ = err
    })
}
```

**CSR Parser Fuzz Test**:

```go
// File: internal/ca/parser/csr_fuzz_test.go
package parser_test

import (
    "testing"
    "crypto/x509"
)

// FuzzCSRParser fuzzes CSR parsing to find edge cases
func FuzzCSRParser(f *testing.F) {
    // Seed corpus with valid PEM-encoded CSRs
    validCSR := `-----BEGIN CERTIFICATE REQUEST-----
MIICZjCCAU4CAQAwITEfMB0GA1UEAwwWdGVzdC1jc3ItZnV6ei1zZWVkLTAwMTCC
ASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAL...
-----END CERTIFICATE REQUEST-----`

    f.Add([]byte(validCSR))

    f.Fuzz(func(t *testing.T, csrBytes []byte) {
        // Parser should never panic
        _, err := x509.ParseCertificateRequest(csrBytes)

        // Most random input will be invalid
        _ = err
    })
}
```

**Critical Fuzz Test Requirements**:

Per 01-02.testing.instructions.md:

- **Unique names**: `FuzzHKDFAllVariants` NOT `FuzzHKDF` (avoid substring conflicts)
- **Minimum fuzz time**: 15 seconds per test
- **Run from project root**: `go test -fuzz=FuzzXXX -fuzztime=15s ./path`
- **PowerShell**: Use unquoted names, `;` for chaining

**Acceptance Criteria**:

- ✅ Fuzz tests for ALL parsers and validators
- ✅ File suffix: `_fuzz_test.go`
- ✅ Unique fuzz function names (no substring conflicts)
- ✅ Seed corpus with valid inputs
- ✅ Tests never panic (only return errors)
- ✅ Minimum 15 second fuzz time per test
- ✅ CI/CD runs fuzz tests (ci-fuzz workflow)

**Validation Commands**:

```bash
# Run individual fuzz test (15s minimum)
go test -fuzz=FuzzJWTParser -fuzztime=15s ./internal/jose

# Run all fuzz tests in package
go test -fuzz=. -fuzztime=15s ./internal/jose

# Verify fuzz test names are unique (no substrings)
grep -r "func Fuzz" internal/ | cut -d'(' -f1 | sort | uniq -d
```

**CI/CD Integration**:

After P4.2 complete, verify ci-fuzz workflow runs these tests (P1.6).

---

### P4.3: Add Property-Based Tests

**Priority**: MEDIUM  
**Effort**: 2 hours  
**Status**: ❌ Not Started

**Objective**: Create property-based tests using gopter to validate mathematical invariants and cryptographic properties.

**Current State**:

- No property-based testing infrastructure
- Cryptographic properties not systematically validated
- No roundtrip testing for crypto operations

**Implementation Strategy**:

```bash
# Step 1: Add gopter dependency
# cspell:ignore leanovate
go get github.com/leanovate/gopter

# Step 2: Create property test files
touch internal/common/crypto/encrypt_property_test.go
touch internal/common/crypto/sign_property_test.go
touch internal/jose/roundtrip_property_test.go
```

**Files to Create**:

- `internal/common/crypto/encrypt_property_test.go`
- `internal/common/crypto/sign_property_test.go`
- `internal/common/crypto/hash_property_test.go`
- `internal/jose/roundtrip_property_test.go`
- `internal/ca/crypto/sign_property_test.go`

**Property Test Pattern**:

```go
// File: internal/common/crypto/encrypt_property_test.go
package crypto_test

import (
    "testing"
    "bytes"
    "github.com/leanovate/gopter"
    "github.com/leanovate/gopter/gen"
    "github.com/leanovate/gopter/prop"
)

// TestEncryptionRoundTrip validates encrypt(decrypt(x)) == x property
func TestEncryptionRoundTrip(t *testing.T) {
    properties := gopter.NewProperties(nil)

    properties.Property("encrypt then decrypt returns original plaintext", prop.ForAll(
        func(plaintext []byte) bool {
            // Generate random key
            key := make([]byte, 32)
            _, _ = rand.Read(key)

            // Encrypt plaintext
            ciphertext, err := crypto.EncryptAES256GCM(key, plaintext)
            if err != nil {
                t.Logf("encryption failed: %v", err)
                return false
            }

            // Decrypt ciphertext
            decrypted, err := crypto.DecryptAES256GCM(key, ciphertext)
            if err != nil {
                t.Logf("decryption failed: %v", err)
                return false
            }

            // Verify roundtrip property
            return bytes.Equal(plaintext, decrypted)
        },
        gen.SliceOf(gen.UInt8()),  // Generate random byte slices
    ))

    properties.TestingRun(t)
}

// TestSignatureRoundTrip validates sign(verify(x)) == x property
func TestSignatureRoundTrip(t *testing.T) {
    properties := gopter.NewProperties(nil)

    properties.Property("sign then verify succeeds for valid messages", prop.ForAll(
        func(message []byte) bool {
            // Generate key pair
            privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

            // Sign message
            signature, err := crypto.SignRSA(privateKey, message)
            if err != nil {
                return false
            }

            // Verify signature
            valid, err := crypto.VerifyRSA(&privateKey.PublicKey, message, signature)
            if err != nil {
                return false
            }

            return valid
        },
        gen.SliceOfN(32, gen.UInt8()),  // Generate 32-byte messages
    ))

    properties.TestingRun(t)
}
```

**JOSE Roundtrip Property Test**:

```go
// File: internal/jose/roundtrip_property_test.go
package jose_test

import (
    "testing"
    "github.com/leanovate/gopter"
    "github.com/leanovate/gopter/gen"
    "github.com/leanovate/gopter/prop"
)

// TestJWSRoundTrip validates JWS sign/verify roundtrip
func TestJWSRoundTrip(t *testing.T) {
    properties := gopter.NewProperties(nil)

    properties.Property("JWS sign then verify returns original payload", prop.ForAll(
        func(payload string) bool {
            // Generate key
            key, _ := keygen.GenerateRSAKey(2048)

            // Sign
            jws, err := jose.SignJWS([]byte(payload), key, "RS256")
            if err != nil {
                return false
            }

            // Verify and extract payload
            verified, err := jose.VerifyJWS(jws, &key.PublicKey)
            if err != nil {
                return false
            }

            return string(verified) == payload
        },
        gen.AlphaString(),  // Generate random strings
    ))

    properties.TestingRun(t)
}
```

**Cryptographic Properties to Test**:

- Encryption roundtrip: `decrypt(encrypt(x)) == x`
- Signature roundtrip: `verify(sign(x)) == valid`
- Hash determinism: `hash(x) == hash(x)` always
- Key derivation determinism: `derive(password, salt) == derive(password, salt)`
- MAC verification: `verify(mac(message, key), message, key) == true`

**Acceptance Criteria**:

- ✅ Property tests for cryptographic invariants
- ✅ File suffix: `_property_test.go`
- ✅ Use gopter for property generation
- ✅ Test roundtrip properties (encrypt/decrypt, sign/verify)
- ✅ Test determinism properties (hash, KDF)
- ✅ All property tests passing

**Validation Commands**:

```bash
# Run property tests
go test -run=TestEncryptionRoundTrip ./internal/common/crypto
go test -run=TestSignatureRoundTrip ./internal/common/crypto
go test -run=TestJWSRoundTrip ./internal/jose
```

---

### P4.4: Mutation Testing Baseline ⭐ CRITICAL

**Priority**: CRITICAL  
**Effort**: 1 hour  
**Status**: ❌ Not Started

**Objective**: Establish mutation testing baseline using gremlins to measure test suite effectiveness. Target ≥80% mutation score per package.

**Current State**:

- No mutation testing performed
- Test suite quality unknown
- No baseline for tracking improvements

**Implementation Strategy**:

```bash
# Step 1: Install gremlins
go install github.com/go-gremlins/gremlins@latest

# Step 2: Run mutation testing (excludes integration tests)
gremlins unleash --tags=!integration

# Step 3: Save baseline report
gremlins unleash --tags=!integration > specs/mutation-testing/baseline-$(date +%Y%m%d).txt
```

**Files to Create**:

- `specs/mutation-testing/baseline-YYYYMMDD.txt`
- `specs/mutation-testing/README.md` (track improvements)

**Mutation Testing Configuration**:

Per 01-02.testing.instructions.md:

- Use gremlins for mutation testing
- Target: ≥80% mutation score per package
- Run: `gremlins unleash --tags=!integration`
- Focus on: business logic, parsers, validators, crypto operations

**Baseline Report Structure**:

```bash
# Run gremlins and save detailed report
gremlins unleash --tags=!integration --output=json > specs/mutation-testing/baseline.json

# Generate human-readable summary
gremlins unleash --tags=!integration > specs/mutation-testing/baseline.txt
```

**Acceptance Criteria**:

- ✅ Gremlins installed and configured
- ✅ Baseline mutation test run completed
- ✅ Report saved to specs/mutation-testing/
- ✅ Packages with <80% score identified
- ✅ Improvement plan documented
- ✅ CI/CD tracks mutation score (optional)

**Validation Commands**:

```bash
# Run full mutation testing suite
gremlins unleash --tags=!integration

# Run mutation testing on specific package
gremlins unleash --tags=!integration ./internal/ca/handler

# View mutation score summary
gremlins unleash --tags=!integration | grep "Mutation Score"
```

**Expected Baseline Results**:

| Package | Expected Mutation Score | Action Required |
|---------|-------------------------|-----------------|
| internal/ca/handler | 60-70% | Improve to 80%+ |
| internal/identity/auth | 50-60% | Improve to 80%+ |
| internal/common/crypto | 75-85% | Maintain/improve |
| internal/jose | 70-80% | Improve to 80%+ |

**Post-Baseline Actions**:

1. Identify packages with <80% mutation score
2. Analyze which mutants survived (test gaps)
3. Add targeted tests to kill surviving mutants
4. Re-run gremlins to verify improvements
5. Track scores over time in specs/mutation-testing/

---

## Progress Tracking

After completing each task, update `PROGRESS.md`:

```bash
# Edit PROGRESS.md to mark task complete
# Update executive summary percentages
# Commit and push
git add specs/001-cryptoutil/PROGRESS.md
git commit -m "docs(speckit): mark P4.X complete"
git push
```

## Validation Checklist

Before marking Phase 4 complete, verify:

- [ ] P4.1: Benchmarks created for all crypto operations
- [ ] P4.2: Fuzz tests created for all parsers/validators
- [ ] P4.3: Property-based tests validate crypto invariants
- [ ] P4.4: Mutation testing baseline established
- [ ] PROGRESS.md updated with all P4.1-P4.4 marked complete
- [ ] Baseline reports saved in specs/ directories
- [ ] CI/CD workflows integrate new tests

## Next Phase

After Phase 4 complete:

- Proceed to Phase 5: Documentation & Demo
- Use PHASE5-IMPLEMENTATION.md guide
- Update PROGRESS.md executive summary
