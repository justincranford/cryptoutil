# Recovery Codes Implementation Plan

**Status**: 🟡 In Progress
**Priority**: MEDIUM (MANDATORY)
**Estimated LOE**: 4-6 hours
**Business Value**: Account recovery, security fallback

---

## Overview

Recovery codes provide backup authentication when users lose access to their primary MFA factors (TOTP, passkeys, hardware tokens). This is **critical for account recovery** and prevents permanent account lockout scenarios.

### Use Cases

1. **Lost Device**: User loses phone with TOTP/passkey
2. **Hardware Failure**: Security key hardware malfunction
3. **Factor Reset**: Admin-initiated MFA factor reset
4. **Emergency Access**: Critical system access during outages

### Security Requirements

- **Single-use only**: Each code can be used exactly once
- **Short-lived**: Codes should expire after 30-90 days
- **Limited quantity**: 8-12 codes per user (NIST recommendation)
- **High entropy**: 12-16 characters, alphanumeric
- **Hashed storage**: Codes stored as hashes (like passwords)
- **Audit trail**: Track code usage (who, when, from where)

---

## Implementation Tasks

### Task 1: Magic Constants (30 minutes)

**File**: `internal/identity/magic/magic_mfa.go`

Add recovery code constants:

```go
const (
    // Recovery code generation
    DefaultRecoveryCodeLength = 16           // 16 characters per code
    DefaultRecoveryCodeCount  = 10           // 10 codes per batch
    RecoveryCodeCharset       = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Exclude ambiguous chars

    // Recovery code lifecycle
    DefaultRecoveryCodeLifetime = 90 * 24 * time.Hour // 90 days

    // MFA factor type
    MFATypeRecoveryCode = "recovery_code"
)
```

---

### Task 2: Domain Model (1 hour)

**File**: `internal/identity/domain/recovery_code.go`

Create recovery code model:

```go
package domain

import (
    "time"
    googleUuid "github.com/google/uuid"
)

// RecoveryCode represents a single-use backup authentication code
type RecoveryCode struct {
    ID        googleUuid.UUID `gorm:"type:text;primaryKey"`
    UserID    googleUuid.UUID `gorm:"type:text;index;not null"`
    CodeHash  string          `gorm:"type:text;not null"`       // bcrypt hash of code
    Used      bool            `gorm:"not null;default:false;index"`
    UsedAt    *time.Time      `gorm:"index"`
    CreatedAt time.Time       `gorm:"not null"`
    ExpiresAt time.Time       `gorm:"not null;index"`
}

// IsExpired checks if the recovery code has expired
func (r *RecoveryCode) IsExpired() bool {
    return time.Now().UTC().After(r.ExpiresAt)
}

// IsUsed checks if the recovery code has already been used
func (r *RecoveryCode) IsUsed() bool {
    return r.Used
}

// MarkAsUsed marks the recovery code as used
func (r *RecoveryCode) MarkAsUsed() {
    r.Used = true
    now := time.Now().UTC()
    r.UsedAt = &now
}
```

**Tests**: `internal/identity/domain/recovery_code_test.go`

- TestRecoveryCode_IsExpired
- TestRecoveryCode_IsUsed
- TestRecoveryCode_MarkAsUsed

---

### Task 3: Recovery Code Generator (1 hour)

**File**: `internal/identity/mfa/recovery_code_generator.go`

Generate recovery codes:

```go
package mfa

import (
    "crypto/rand"
    "fmt"
    cryptoutilMagic "cryptoutil/internal/identity/magic"
)

// GenerateRecoveryCode generates a cryptographically random recovery code
// Format: XXXX-XXXX-XXXX-XXXX (4 groups of 4 chars)
func GenerateRecoveryCode() (string, error) {
    const groupSize = 4
    const groupCount = 4
    const totalChars = groupSize * groupCount

    randomBytes := make([]byte, totalChars)
    if _, err := rand.Read(randomBytes); err != nil {
        return "", fmt.Errorf("failed to generate random bytes: %w", err)
    }

    charset := cryptoutilMagic.RecoveryCodeCharset
    code := make([]byte, totalChars)

    for i := range totalChars {
        code[i] = charset[int(randomBytes[i])%len(charset)]
    }

    // Format with hyphens: XXXX-XXXX-XXXX-XXXX
    formatted := fmt.Sprintf("%s-%s-%s-%s",
        code[0:4],
        code[4:8],
        code[8:12],
        code[12:16])

    return formatted, nil
}

// GenerateRecoveryCodes generates a batch of recovery codes
func GenerateRecoveryCodes(count int) ([]string, error) {
    codes := make([]string, count)
    seen := make(map[string]bool, count)

    for i := range count {
        for {
            code, err := GenerateRecoveryCode()
            if err != nil {
                return nil, err
            }

            if !seen[code] {
                codes[i] = code
                seen[code] = true
                break
            }
        }
    }

    return codes, nil
}
```

**Tests**: `internal/identity/mfa/recovery_code_generator_test.go`

- TestGenerateRecoveryCode_Format (XXXX-XXXX-XXXX-XXXX pattern)
- TestGenerateRecoveryCode_Length (19 chars with hyphens)
- TestGenerateRecoveryCode_Uniqueness (1000 samples, no collisions)
- TestGenerateRecoveryCodes_Batch (10 codes, all unique)

---

### Task 4: Repository Interface (30 minutes)

**File**: `internal/identity/repository/recovery_code_repository.go`

```go
package repository

import (
    "context"
    googleUuid "github.com/google/uuid"
    cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

type RecoveryCodeRepository interface {
    // Create stores a new recovery code
    Create(ctx context.Context, code *cryptoutilIdentityDomain.RecoveryCode) error

    // CreateBatch stores multiple recovery codes in a transaction
    CreateBatch(ctx context.Context, codes []*cryptoutilIdentityDomain.RecoveryCode) error

    // GetByUserID retrieves all recovery codes for a user
    GetByUserID(ctx context.Context, userID googleUuid.UUID) ([]*cryptoutilIdentityDomain.RecoveryCode, error)

    // GetByID retrieves a recovery code by ID
    GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.RecoveryCode, error)

    // Update modifies an existing recovery code (typically to mark as used)
    Update(ctx context.Context, code *cryptoutilIdentityDomain.RecoveryCode) error

    // DeleteByUserID removes all recovery codes for a user (regeneration scenario)
    DeleteByUserID(ctx context.Context, userID googleUuid.UUID) error

    // DeleteExpired removes all expired recovery codes
    DeleteExpired(ctx context.Context) (int64, error)

    // CountUnused returns count of unused, unexpired codes for a user
    CountUnused(ctx context.Context, userID googleUuid.UUID) (int64, error)
}
```

---

### Task 5: Repository Implementation (1 hour)

**File**: `internal/identity/repository/orm/recovery_code_repository.go`

Implement GORM repository with transaction support for batch operations.

---

### Task 6: Database Migration (30 minutes)

**File**: `internal/identity/repository/orm/migrations/000012_recovery_codes.up.sql`

```sql
CREATE TABLE IF NOT EXISTS recovery_codes (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_recovery_codes_user_id ON recovery_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_recovery_codes_used ON recovery_codes(used);
CREATE INDEX IF NOT EXISTS idx_recovery_codes_expires_at ON recovery_codes(expires_at);
CREATE INDEX IF NOT EXISTS idx_recovery_codes_used_at ON recovery_codes(used_at);
```

---

### Task 7: Recovery Code Service (1.5 hours)

**File**: `internal/identity/mfa/recovery_code_service.go`

Service methods:

```go
type RecoveryCodeService struct {
    repo cryptoutilIdentityRepository.RecoveryCodeRepository
}

// GenerateForUser generates a new batch of recovery codes for a user
// Returns plaintext codes (shown once to user) and stores hashed versions
func (s *RecoveryCodeService) GenerateForUser(ctx context.Context, userID googleUuid.UUID, count int) ([]string, error)

// Verify checks if a recovery code is valid and marks it as used
func (s *RecoveryCodeService) Verify(ctx context.Context, userID googleUuid.UUID, plaintext string) error

// RegenerateForUser deletes old codes and generates new batch
func (s *RecoveryCodeService) RegenerateForUser(ctx context.Context, userID googleUuid.UUID, count int) ([]string, error)

// GetRemainingCount returns count of unused, unexpired codes
func (s *RecoveryCodeService) GetRemainingCount(ctx context.Context, userID googleUuid.UUID) (int64, error)
```

---

### Task 8: API Handlers (1 hour)

**Administrative Endpoints** (requires admin auth):

1. `POST /oidc/v1/mfa/recovery-codes/generate`
   - Request: `{"user_id": "uuid"}`
   - Response: `{"codes": ["XXXX-...", ...], "expires_at": "2025-04-10T..."}`
   - **Security**: Show codes ONCE, user must save them

2. `GET /oidc/v1/mfa/recovery-codes/count`
   - Request: `?user_id=uuid`
   - Response: `{"remaining": 7, "total": 10}`

3. `POST /oidc/v1/mfa/recovery-codes/regenerate`
   - Request: `{"user_id": "uuid"}`
   - Response: `{"codes": [...], "expires_at": "..."}`
   - **Action**: Deletes old codes, generates new batch

**User Verification Endpoint**:

1. `POST /oidc/v1/mfa/verify-recovery-code`
   - Request: `{"code": "XXXX-XXXX-XXXX-XXXX"}`
   - Response: `{"verified": true}` or `{"error": "invalid_code"}`
   - **Side Effect**: Marks code as used on success

---

### Task 9: Integration with Login Flow (30 minutes)

Modify `/oidc/v1/login` to support recovery code verification:

- After username/password validation
- If user has MFA enabled
- Allow recovery code as alternative to TOTP/passkey
- Validate recovery code via RecoveryCodeService

---

### Task 10: Unit Tests (1 hour)

**Domain Model Tests**: 3 tests (IsExpired, IsUsed, MarkAsUsed)
**Generator Tests**: 4 tests (format, length, uniqueness, batch)
**Service Tests**: 6 tests (generate, verify, regenerate, count, expired, used)
**Handler Tests**: 8 tests (generate success, invalid user, verify success, invalid code, regenerate, count)

---

### Task 11: Integration Tests (1 hour)

**E2E Flow Tests**:

1. Generate codes → verify one → count remaining (9 left)
2. Generate codes → use all → verify fails
3. Generate codes → wait for expiration → verify fails
4. Regenerate codes → old codes invalid, new codes work

---

## Security Considerations

### Code Storage ✅

- **Hashed with bcrypt**: Like passwords, never store plaintext
- **Work factor**: bcrypt cost 10 (default)
- **Salt**: Automatically handled by bcrypt

### Single-use Enforcement ✅

- Mark as Used on successful verification
- Check IsUsed() before verification
- Audit trail with UsedAt timestamp

### Expiration ✅

- 90-day default lifetime
- Check IsExpired() before verification
- Automatic cleanup via DeleteExpired()

### Rate Limiting ⚠️

- Implement per-user rate limiting (5 attempts/hour)
- Prevents brute-force attacks on recovery codes
- Track failed attempts in audit log

### User Notification 📧

- Email notification when codes generated
- Email notification when code used
- Warning when only 1-2 codes remaining

---

## Testing Strategy

### Unit Tests (13 tests)

- Domain model methods (3)
- Generator functions (4)
- Service methods (6)

### Handler Tests (8 tests)

- Happy paths + error cases

### Integration Tests (4 E2E flows)

- Full recovery code lifecycle

---

## Success Criteria

- ✅ 13+ unit tests passing
- ✅ 4+ integration tests passing
- ✅ Recovery codes generated with high entropy
- ✅ Single-use enforcement working
- ✅ Expiration validation working
- ✅ Integration with login flow working
- ✅ Pre-commit hooks satisfied

---

*Implementation Plan Version: 1.0.0*
*Author: GitHub Copilot (Agent)*
*Created: 2025-01-08*
