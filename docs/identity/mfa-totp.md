# TOTP Multi-Factor Authentication (MFA)

## Overview

The Identity service provides Time-based One-Time Password (TOTP) multi-factor authentication compliant with RFC 6238. This implementation includes account lockout protection, backup codes for account recovery, and a 30-minute MFA step-up requirement for sensitive operations.

## Features

### Core TOTP Functionality

- **RFC 6238 Compliance**: Implements Time-based One-Time Password Algorithm
- **Standard Configuration**: 6-digit codes, 30-second time window, SHA-1 hash algorithm
- **QR Code Enrollment**: Automatic QR code generation for authenticator apps (Google Authenticator, Authy, Microsoft Authenticator)
- **Backup Codes**: 10 single-use backup codes for account recovery

### Security Features

- **Account Lockout**: 5 consecutive failed verification attempts trigger a 15-minute account lockout
- **Single-Use Backup Codes**: Each backup code can only be used once
- **MFA Step-Up**: Sensitive operations require re-verification within 30 minutes
- **PBKDF2 Hashing**: Backup codes hashed using PBKDF2-HMAC-SHA256 (FIPS 140-3 compliant)

## API Endpoints

All endpoints are available at both `/browser/api/v1/mfa/totp/*` and `/service/api/v1/mfa/totp/*` paths.

### Enroll TOTP MFA

**Endpoint**: `POST /oidc/v1/mfa/totp/enroll`

**Request**:
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response** (200 OK):
```json
{
  "secret": "JBSWY3DPEHPK3PXP",
  "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
  "backup_codes": ["A1B2C3D4", "E5F6G7H8", "I9J0K1L2", "M3N4O5P6", "Q7R8S9T0", "U1V2W3X4", "Y5Z6A7B8", "C9D0E1F2", "G3H4I5J6", "K7L8M9N0"]
}
```

**Error Responses**:
- `400 Bad Request`: User ID missing or user already enrolled
- `500 Internal Server Error`: Database or cryptographic operation failure

### Verify TOTP Code

**Endpoint**: `POST /oidc/v1/mfa/totp/verify`

**Response** (200 OK - Valid):
```json
{
  "verified": true,
  "mfa_verified_at": "2025-01-28T12:34:56Z"
}
```

**Response** (403 Forbidden - Locked):
```json
{
  "error": "totp_account_locked",
  "error_description": "Account locked due to too many failed verification attempts. Try again after 15 minutes."
}
```

### Generate New Backup Codes

**Endpoint**: `POST /oidc/v1/mfa/totp/backup-codes/generate`

**Note**: Generating new backup codes **invalidates all previous backup codes**.

### Verify Backup Code

**Endpoint**: `POST /oidc/v1/mfa/totp/backup-codes/verify`

**Note**: Each backup code can only be used once. Used codes are marked with `used: true`.

## Database Schema

### totp_secrets Table

```sql
CREATE TABLE totp_secrets (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL,
    algorithm TEXT NOT NULL DEFAULT 'SHA1',
    digits INTEGER NOT NULL DEFAULT 6,
    period INTEGER NOT NULL DEFAULT 30,
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### backup_codes Table

```sql
CREATE TABLE backup_codes (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES totp_secrets(user_id) ON DELETE CASCADE
);
```

## Usage Examples

### Browser Client (JavaScript)

```javascript
// Enrollment Flow
const enrollResponse = await fetch('/browser/api/v1/mfa/totp/enroll', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-CSRF-Token': csrfToken,
  },
  credentials: 'include',
  body: JSON.stringify({user_id: userId})
});

const enrollData = await enrollResponse.json();
document.getElementById('qr-code').src = enrollData.qr_code;

// Verification
const verifyResponse = await fetch('/browser/api/v1/mfa/totp/verify', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-CSRF-Token': csrfToken,
  },
  credentials: 'include',
  body: JSON.stringify({user_id: userId, code: userEnteredCode})
});

const verifyData = await verifyResponse.json();
if (verifyData.verified) {
  console.log('MFA verified at:', verifyData.mfa_verified_at);
}
```

### Service Client (Go)

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

func enrollTOTP(userID string) error {
    reqBody := map[string]string{"user_id": userID}
    jsonData, _ := json.Marshal(reqBody)
    
    resp, err := http.Post(
        "https://identity-authz:8180/service/api/v1/mfa/totp/enroll",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    defer resp.Body.Close()
    
    var enrollData struct {
        Secret      string   `json:"secret"`
        BackupCodes []string `json:"backup_codes"`
    }
    json.NewDecoder(resp.Body).Decode(&enrollData)
    return nil
}
```

## Security Considerations

### Account Lockout

- **Threshold**: 5 consecutive failed TOTP verification attempts
- **Lockout Duration**: 15 minutes from the last failed attempt
- **Reset**: Successful verification resets the failure counter to 0
- **Lockout Response**: HTTP 403 Forbidden with error message

### Backup Codes

- **Single-Use**: Each code can only be used once
- **Hashing**: Codes are hashed using PBKDF2-HMAC-SHA256 (FIPS 140-3 compliant)
- **Generation**: New codes invalidate all previous codes
- **Storage**: Codes are never stored in plaintext

### MFA Step-Up Requirement

- **Time Window**: 30 minutes from last successful verification
- **Enforcement**: Sensitive operations check `last_used_at` timestamp
- **Re-verification**: Operations requiring step-up prompt for TOTP code again

### Best Practices

1. **Always Use HTTPS**: TOTP codes transmitted over TLS 1.3+
2. **Store Backup Codes Securely**: Encrypt backup codes at rest
3. **Prompt for Immediate Verification**: Confirm enrollment by verifying a code immediately
4. **Monitor Failed Attempts**: Alert users of unusual failed verification patterns
5. **Regenerate Backup Codes**: Prompt users to generate new codes after using one
6. **Time Synchronization**: Ensure server time is synchronized (NTP) for accurate TOTP validation

## Troubleshooting

### Common Issues

**Problem**: "Invalid TOTP code" even though code is correct
- **Solution**: Check server time synchronization (NTP). TOTP depends on accurate time.

**Problem**: Account locked after 5 failures
- **Solution**: Wait 15 minutes for automatic unlock.
- **Alternative**: Use backup code for immediate access.

**Problem**: Backup code not working
- **Solution**: Check if code was already used (`used: true` in database).
- **Alternative**: Use a different backup code.

**Problem**: QR code not scanning
- **Solution**: Ensure QR code is displayed at sufficient size (minimum 200x200 pixels).
- **Alternative**: Manually enter the secret key in authenticator app.

## Migration Guide

### Migrations

- **Migration 0008**: totp_secrets table and indexes
- **Migration 0009**: backup_codes table with foreign key to totp_secrets

Migrations are applied automatically on service startup using `golang-migrate`.

## References

- **RFC 6238**: TOTP: Time-Based One-Time Password Algorithm
- **OWASP Authentication Cheat Sheet**: Backup Codes for MFA
- **FIPS 140-3**: Cryptographic Module Validation Program
