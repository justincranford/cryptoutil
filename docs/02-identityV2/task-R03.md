# Task R03: Harden Client Authentication (Task 07 Remediation - FIPS Compliance)

**Priority**: ⚠️ **HIGH** - Security vulnerabilities in client authentication
**Effort**: 18 hours (2.25 days)
**Owner**: Security/OAuth engineer
**Dependencies**: R01 (authorization code flow)
**Source**: GAP-ANALYSIS-DETAILED.md Priority 3 issues + FIPS-140-3 mandate

---

## Problem Statement

Client authentication has critical security gaps that violate FIPS-140-3 compliance:

1. **Client Secret Storage**: Secrets stored in plaintext in database (security vulnerability)
2. **bcrypt is NOT FIPS-approved**: Current implementation uses bcrypt which violates FIPS-140-3
3. **FIPS-Approved Replacement Required**: Must use PBKDF2-HMAC-SHA256 with configurable iterations
4. **Algorithm Agility Missing**: No support for configurable secret hashing algorithms
5. **CRL/OCSP Validation**: mTLS authenticators don't validate certificate revocation
6. **Migration Path**: Must support existing bcrypt hashes during migration to PBKDF2

**Impact**: Security vulnerability (plaintext secrets), FIPS-140-3 non-compliance, incomplete mTLS validation.

---

## Acceptance Criteria

- [ ] Client secrets hashed with PBKDF2-HMAC-SHA256 (210,000 iterations minimum, SHA-256, 32-byte key)
- [ ] FIPS mode is ALWAYS enabled (no bcrypt option in production)
- [ ] Legacy bcrypt hash verification supported during migration (read-only)
- [ ] Database migration created to hash existing plaintext secrets with PBKDF2
- [ ] `client_secret_basic` authenticator updated to use PBKDF2 verification
- [ ] `client_secret_post` authenticator updated to use PBKDF2 verification
- [ ] Client registration endpoint hashes secrets before storage
- [ ] CRL validation implemented for mTLS client authentication
- [ ] OCSP validation implemented for mTLS client authentication
- [ ] Certificate revocation checks cached (5-minute TTL)
- [ ] All unit tests pass with table-driven test pattern
- [ ] Integration tests validate secret hashing and revocation checking
- [ ] Pre-commit hooks pass (golangci-lint, cspell, formatting)

---

## Implementation Steps

### Step 1: Use FIPS-Approved Secret Hashing Infrastructure (2 hours)

**Files**:
- `internal/crypto/secret.go` (ALREADY CREATED - use this)
- `internal/crypto/registry.go` (ALREADY CREATED - remove bcrypt option)

**Update `internal/crypto/registry.go`**:
```go
package crypto

import (
    "fmt"
)

// HashSecret hashes a secret using PBKDF2-HMAC-SHA256 (FIPS-approved).
// FIPS mode is ALWAYS enabled - bcrypt option removed.
func HashSecret(secret string) (string, error) {
    // FIPS mode ALWAYS uses PBKDF2
    return HashSecretPBKDF2(secret, DefaultPBKDF2Iterations)
}

// VerifySecret verifies a secret against a hash.
// Supports PBKDF2 (FIPS) and bcrypt (legacy migration only).
func VerifySecret(secret, hash string) error {
    return verifySecret(secret, hash)
}
```

**Remove environment variable check** - FIPS is mandatory, not configurable.

**Tests**: Verify HashSecret always uses PBKDF2, VerifySecret handles both formats.

---

### Step 2: Update Client Secret Authenticators (4 hours)

**File**: `internal/identity/authz/clientauth/client_secret_basic.go`

**Replace bcrypt usage**:
```go
import (
    cryptoutilCrypto "cryptoutil/internal/crypto"
)

func (a *ClientSecretBasicAuthenticator) Authenticate(ctx context.Context, r *fiber.Ctx) (*domain.Client, error) {
    // ... extract clientID and clientSecret from Authorization header ...

    // Retrieve client from database
    client, err := a.clientRepo.GetClientByID(ctx, clientID)
    if err != nil {
        a.logger.Warn("Client not found", "client_id", clientID)
        return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid client credentials")
    }

    // Verify client secret using FIPS-approved PBKDF2 (supports bcrypt legacy)
    if err := cryptoutilCrypto.VerifySecret(clientSecret, client.ClientSecretHash); err != nil {
        a.logger.Warn("Client secret verification failed", "client_id", clientID, "error", err)
        return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid client credentials")
    }

    a.logger.Info("Client authenticated", "client_id", clientID, "method", "client_secret_basic")
    return client, nil
}
```

**File**: `internal/identity/authz/clientauth/client_secret_post.go`

**Replace bcrypt usage** (same pattern as client_secret_basic).

**Tests**: Verify PBKDF2 hash verification, legacy bcrypt support, authentication failures logged.

---

### Step 3: Update Client Registration Endpoint (3 hours)

**File**: `internal/identity/authz/handlers_client_registration.go`

**Hash secret before storage**:
```go
import (
    cryptoutilCrypto "cryptoutil/internal/crypto"
)

func (s *Service) RegisterClient(c *fiber.Ctx) error {
    var req RegisterClientRequest
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
    }

    // Generate client secret if not provided
    clientSecret := req.ClientSecret
    if clientSecret == "" {
        clientSecret = generateClientSecret() // 32 bytes crypto/rand
    }

    // Hash client secret with PBKDF2-HMAC-SHA256 (FIPS-approved)
    hashedSecret, err := cryptoutilCrypto.HashSecret(clientSecret)
    if err != nil {
        s.logger.Error("Failed to hash client secret", "error", err)
        return fiber.NewError(fiber.StatusInternalServerError, "Failed to register client")
    }

    client := &domain.Client{
        ID:                googleUuid.Must(googleUuid.NewV7()),
        ClientID:          req.ClientID,
        ClientSecretHash:  hashedSecret, // Store PBKDF2 hash, not plaintext
        RedirectURIs:      req.RedirectURIs,
        GrantTypes:        req.GrantTypes,
        ResponseTypes:     req.ResponseTypes,
        Scope:             req.Scope,
        CreatedAt:         time.Now(),
    }

    if err := s.clientRepo.CreateClient(ctx, client); err != nil {
        s.logger.Error("Failed to create client", "client_id", client.ClientID, "error", err)
        return fiber.NewError(fiber.StatusInternalServerError, "Failed to register client")
    }

    // Return client_secret in response (ONLY TIME IT'S SHOWN)
    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "client_id":     client.ClientID,
        "client_secret": clientSecret, // Show once, then never stored in plaintext
    })
}

func generateClientSecret() string {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        panic(fmt.Sprintf("Failed to generate client secret: %v", err))
    }
    return base64.RawURLEncoding.EncodeToString(b)
}
```

**Tests**: Verify secret hashing before storage, client_secret returned only during registration.

---

### Step 4: Database Migration for Existing Secrets (3 hours)

**File**: `internal/identity/storage/migrations/0003_hash_client_secrets.up.sql`

**Migration strategy**:
1. Add new column `client_secret_hash_pbkdf2`
2. For each client with plaintext `client_secret`, hash it with PBKDF2
3. Copy PBKDF2 hash to `client_secret_hash_pbkdf2`
4. Rename `client_secret_hash_pbkdf2` to `client_secret_hash`
5. Drop old `client_secret` column

**SQL Migration** (PostgreSQL):
```sql
-- Add temporary column for PBKDF2 hashes
ALTER TABLE clients ADD COLUMN client_secret_hash_pbkdf2 TEXT;

-- Application code will run update to hash plaintext secrets
-- (Cannot hash in SQL, must use application crypto library)

-- After migration complete, rename column
ALTER TABLE clients RENAME COLUMN client_secret_hash_pbkdf2 TO client_secret_hash;
ALTER TABLE clients DROP COLUMN client_secret IF EXISTS;
```

**Go Migration Code**:
```go
func MigrateClientSecretsToPBKDF2(ctx context.Context, db *gorm.DB) error {
    var clients []domain.Client
    if err := db.WithContext(ctx).Find(&clients).Error; err != nil {
        return fmt.Errorf("failed to retrieve clients: %w", err)
    }

    for _, client := range clients {
        // Skip if already hashed (starts with "pbkdf2$" or "$2a$")
        if strings.HasPrefix(client.ClientSecretHash, "pbkdf2$") || strings.HasPrefix(client.ClientSecretHash, "$2a$") {
            continue
        }

        // Hash plaintext secret with PBKDF2
        hashedSecret, err := cryptoutilCrypto.HashSecret(client.ClientSecretHash)
        if err != nil {
            return fmt.Errorf("failed to hash secret for client %s: %w", client.ClientID, err)
        }

        // Update client with PBKDF2 hash
        client.ClientSecretHash = hashedSecret
        if err := db.WithContext(ctx).Save(&client).Error; err != nil {
            return fmt.Errorf("failed to update client %s: %w", client.ClientID, err)
        }
    }

    return nil
}
```

**Tests**: Verify migration hashes all plaintext secrets, skips already-hashed secrets.

---

### Step 5: Implement CRL Validation for mTLS (5 hours)

**File**: `internal/identity/authz/clientauth/tls_client_auth.go`

**Add CRL validation**:
```go
import (
    "crypto/x509"
    "net/http"
    "time"
)

func (a *TLSClientAuthenticator) Authenticate(ctx context.Context, r *fiber.Ctx) (*domain.Client, error) {
    // ... extract client certificate from TLS connection ...

    // Validate certificate revocation (CRL)
    if err := validateCertificateRevocation(cert); err != nil {
        a.logger.Warn("Certificate revoked", "subject", cert.Subject, "error", err)
        return nil, fiber.NewError(fiber.StatusUnauthorized, "Certificate revoked")
    }

    // ... existing authentication logic ...
}

func validateCertificateRevocation(cert *x509.Certificate) error {
    // Check CRL if available
    if len(cert.CRLDistributionPoints) > 0 {
        revoked, err := checkCRL(cert)
        if err != nil {
            return fmt.Errorf("CRL check failed: %w", err)
        }
        if revoked {
            return fmt.Errorf("certificate revoked (CRL)")
        }
    }

    return nil // Not revoked
}

func checkCRL(cert *x509.Certificate) (bool, error) {
    // Download CRL from first distribution point
    crlURL := cert.CRLDistributionPoints[0]
    resp, err := http.Get(crlURL)
    if err != nil {
        return false, fmt.Errorf("failed to download CRL: %w", err)
    }
    defer resp.Body.Close()

    crlBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return false, fmt.Errorf("failed to read CRL: %w", err)
    }

    crl, err := x509.ParseCRL(crlBytes)
    if err != nil {
        return false, fmt.Errorf("failed to parse CRL: %w", err)
    }

    // Check if certificate serial number in revoked list
    for _, revokedCert := range crl.TBSCertList.RevokedCertificates {
        if revokedCert.SerialNumber.Cmp(cert.SerialNumber) == 0 {
            return true, nil // Revoked
        }
    }

    return false, nil // Not revoked
}
```

**Caching**: Implement 5-minute TTL cache for CRL downloads (avoid repeated HTTP requests).

**Tests**: Verify CRL download, parsing, revocation checking, caching.

---

### Step 6: Implement OCSP Validation for mTLS (5 hours)

**File**: `internal/identity/authz/clientauth/tls_client_auth.go`

**Add OCSP validation** (prioritize OCSP over CRL):
```go
import (
    "crypto/x509/ocsp"
)

func validateCertificateRevocation(cert *x509.Certificate) error {
    // Check OCSP if available (faster than CRL)
    if len(cert.OCSPServer) > 0 {
        ocspStatus, err := checkOCSP(cert)
        if err != nil {
            return fmt.Errorf("OCSP check failed: %w", err)
        }
        if ocspStatus == ocsp.Revoked {
            return fmt.Errorf("certificate revoked (OCSP)")
        }
        return nil // OCSP check passed
    }

    // Fallback to CRL if OCSP not available
    if len(cert.CRLDistributionPoints) > 0 {
        revoked, err := checkCRL(cert)
        if err != nil {
            return fmt.Errorf("CRL check failed: %w", err)
        }
        if revoked {
            return fmt.Errorf("certificate revoked (CRL)")
        }
    }

    return nil // Not revoked
}

func checkOCSP(cert *x509.Certificate) (int, error) {
    // Build OCSP request
    ocspRequest, err := ocsp.CreateRequest(cert, cert, nil)
    if err != nil {
        return 0, fmt.Errorf("failed to create OCSP request: %w", err)
    }

    // Send OCSP request
    ocspURL := cert.OCSPServer[0]
    resp, err := http.Post(ocspURL, "application/ocsp-request", bytes.NewReader(ocspRequest))
    if err != nil {
        return 0, fmt.Errorf("failed to send OCSP request: %w", err)
    }
    defer resp.Body.Close()

    ocspResp, err := io.ReadAll(resp.Body)
    if err != nil {
        return 0, fmt.Errorf("failed to read OCSP response: %w", err)
    }

    // Parse OCSP response
    parsedResp, err := ocsp.ParseResponse(ocspResp, cert)
    if err != nil {
        return 0, fmt.Errorf("failed to parse OCSP response: %w", err)
    }

    return parsedResp.Status, nil
}
```

**Caching**: Implement 5-minute TTL cache for OCSP responses.

**Tests**: Verify OCSP request creation, response parsing, status checking, caching.

---

## Testing Requirements

### Unit Tests

**File**: `internal/crypto/secret_test.go`
- PBKDF2 hash generation (table-driven test with various iteration counts)
- PBKDF2 hash verification (valid/invalid secrets)
- Legacy bcrypt hash verification (backward compatibility)
- Hash format validation (pbkdf2$ prefix, components)

**File**: `internal/identity/authz/clientauth/client_secret_basic_test.go`
- Client secret verification with PBKDF2 hashes
- Legacy bcrypt hash support during migration
- Authentication failure logging
- Invalid client ID handling

**File**: `internal/identity/authz/clientauth/tls_client_auth_test.go`
- CRL download and parsing (mock HTTP server)
- OCSP request/response handling (mock HTTP server)
- Certificate revocation detection
- Cache TTL enforcement (5 minutes)

### Integration Tests

**File**: `internal/identity/test/e2e/client_authentication_test.go`
- End-to-end client authentication:
  1. Register client with generated secret
  2. Authenticate with client_secret_basic (PBKDF2 hash)
  3. Verify legacy bcrypt hash still works (migration support)
  4. Attempt authentication with wrong secret (fails)
  5. mTLS authentication with valid certificate
  6. mTLS authentication with revoked certificate (fails)

---

## Pre-commit Enforcement

- Run `golangci-lint run --fix` on all modified files
- Fix all linting errors (wsl, godot, mnd, errcheck)
- Ensure test coverage ≥85% for crypto infrastructure code
- Ensure test coverage ≥80% for authenticator code
- Run `go test ./internal/crypto/... ./internal/identity/authz/clientauth/... -cover` before committing
- Commit with conventional commit message: `feat(authz): implement FIPS-compliant client secret hashing with PBKDF2-HMAC-SHA256`

---

## Validation

**Success Criteria**:
- [ ] Client secrets hashed with PBKDF2-HMAC-SHA256 (FIPS-approved)
- [ ] Legacy bcrypt hashes verified during migration
- [ ] Database migration hashes all plaintext secrets
- [ ] Client authenticators use PBKDF2 verification
- [ ] CRL validation implemented for mTLS
- [ ] OCSP validation implemented for mTLS
- [ ] Revocation checks cached (5-minute TTL)
- [ ] All unit tests pass
- [ ] Integration test validates authentication and revocation
- [ ] Pre-commit hooks pass
- [ ] Code coverage ≥85% (crypto) and ≥80% (authenticators)

---

## References

- GAP-ANALYSIS-DETAILED.md: Priority 3 issues (lines 246-327)
- REMEDIATION-MASTER-PLAN-2025.md: R03 section (updated with FIPS requirements)
- FIPS-140-3: https://csrc.nist.gov/publications/detail/fips/140/3/final
- PBKDF2 RFC 2898: https://datatracker.ietf.org/doc/html/rfc2898
- OCSP RFC 6960: https://datatracker.ietf.org/doc/html/rfc6960
- internal/crypto/secret.go: PBKDF2 implementation reference
