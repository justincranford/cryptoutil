# Device Authorization Grant (RFC 8628) Implementation Plan

## Overview

**RFC**: [RFC 8628 - OAuth 2.0 Device Authorization Grant](https://datatracker.ietf.org/doc/html/rfc8628)

**Purpose**: Enable OAuth 2.0 authentication for devices with limited input capabilities (smart TVs, IoT devices, CLI tools, hardware tokens) by delegating user authentication to a secondary device (smartphone, computer).

**Business Value**: HIGH - Critical for IoT device authentication, smart TV apps, CLI tools, and hardware token enrollment.

**Estimated LOE**: ~8 hours (OpenAPI spec + domain models + handlers + repository + tests)

## Flow Overview

### 1. Device Requests Authorization (POST /device_authorization)

**Client Request**:

```http
POST /oauth2/v1/device_authorization HTTP/1.1
Host: authz.example.com
Content-Type: application/x-www-form-urlencoded

client_id=client123&scope=openid profile
```

**Server Response**:

```json
{
  "device_code": "GmRhmhcxhwEzkoEqiMEg_DnyEysNkuNhszIySk9eS",
  "user_code": "WDJB-MJHT",
  "verification_uri": "https://authz.example.com/device",
  "verification_uri_complete": "https://authz.example.com/device?user_code=WDJB-MJHT",
  "expires_in": 1800,
  "interval": 5
}
```

### 2. User Authorizes on Secondary Device

- User visits `verification_uri` on smartphone/computer
- User enters `user_code` (e.g., "WDJB-MJHT")
- User authenticates (login + MFA if required)
- User consents to scopes
- Server marks device_code as "authorized" in database

### 3. Device Polls for Token (POST /token with device_code)

**Client Request** (every 5 seconds per `interval`):

```http
POST /oauth2/v1/token HTTP/1.1
Host: authz.example.com
Content-Type: application/x-www-form-urlencoded

grant_type=urn:ietf:params:oauth:grant-type:device_code&device_code=GmRhmhcxhwEzkoEqiMEg_DnyEysNkuNhszIySk9eS&client_id=client123
```

**Server Responses**:

**Pending Authorization**:

```json
HTTP/1.1 400 Bad Request
{
  "error": "authorization_pending",
  "error_description": "User has not yet authorized the device"
}
```

**Slow Down** (client polling too fast):

```json
HTTP/1.1 400 Bad Request
{
  "error": "slow_down",
  "error_description": "Polling interval must be increased by 5 seconds"
}
```

**Denied by User**:

```json
HTTP/1.1 400 Bad Request
{
  "error": "access_denied",
  "error_description": "User denied the authorization request"
}
```

**Expired**:

```json
HTTP/1.1 400 Bad Request
{
  "error": "expired_token",
  "error_description": "Device code has expired"
}
```

**Success** (user authorized):

```json
HTTP/1.1 200 OK
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "tGzv3JOkF0XG5Qx2TlKWIA",
  "scope": "openid profile"
}
```

## Implementation Tasks

### Task 1: Add Magic Constants (~15 minutes)

**File**: `internal/identity/magic/magic_oauth.go`

```go
// Device Authorization Grant (RFC 8628) constants.
const (
 GrantTypeDeviceCode = "urn:ietf:params:oauth:grant-type:device_code" // Device code grant type.

 ParamDeviceCode = "device_code" // Device code parameter.
 ParamUserCode   = "user_code"   // User code parameter.

 ErrorAuthorizationPending = "authorization_pending" // User has not yet authorized.
 ErrorSlowDown            = "slow_down"              // Polling too fast.
 ErrorExpiredToken        = "expired_token"          // Device code expired.

 DefaultDeviceCodeLength   = 32  // Device code length in bytes (base64url = 43 chars).
 DefaultUserCodeLength     = 8   // User code length in characters (e.g., "WDJB-MJHT").
 DefaultDeviceCodeLifetime = 30 * time.Minute // Device code validity (30 minutes).
 DefaultPollingInterval    = 5 * time.Second  // Minimum polling interval.
)
```

### Task 2: Domain Model (~30 minutes)

**File**: `internal/identity/domain/device_authorization.go`

```go
package domain

import (
 "time"
 googleUuid "github.com/google/uuid"
)

// DeviceAuthorization represents a pending device authorization request (RFC 8628).
type DeviceAuthorization struct {
 ID          googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`

 // Client information.
 ClientID    string `gorm:"type:text;not null;index" json:"client_id"`

 // Device codes.
 DeviceCode  string `gorm:"type:text;not null;uniqueIndex" json:"device_code"`
 UserCode    string `gorm:"type:text;not null;uniqueIndex" json:"user_code"`

 // Request parameters.
 Scope       string `gorm:"type:text" json:"scope"`

 // User information (populated after user authorizes on secondary device).
 UserID      NullableUUID `gorm:"type:text;index" json:"user_id"`

 // Authorization status.
 Status      string    `gorm:"type:text;not null;index" json:"status"` // pending, authorized, denied, used

 // Polling control.
 LastPolledAt *time.Time `gorm:"index" json:"last_polled_at,omitempty"`

 // Request metadata.
 CreatedAt   time.Time `gorm:"not null" json:"created_at"`
 ExpiresAt   time.Time `gorm:"not null;index" json:"expires_at"`

 // Token issuance (populated when grant_type=device_code succeeds).
 UsedAt      *time.Time `gorm:"index" json:"used_at,omitempty"`
}

// TableName returns the database table name.
func (DeviceAuthorization) TableName() string {
 return "device_authorizations"
}

// IsExpired checks if device code has expired.
func (d *DeviceAuthorization) IsExpired() bool {
 return time.Now().After(d.ExpiresAt)
}

// IsPending checks if authorization is pending user action.
func (d *DeviceAuthorization) IsPending() bool {
 return d.Status == "pending"
}

// IsAuthorized checks if user has authorized the device.
func (d *DeviceAuthorization) IsAuthorized() bool {
 return d.Status == "authorized"
}

// IsDenied checks if user denied the authorization.
func (d *DeviceAuthorization) IsDenied() bool {
 return d.Status == "denied"
}

// IsUsed checks if device code has been exchanged for token.
func (d *DeviceAuthorization) IsUsed() bool {
 return d.Status == "used"
}
```

### Task 3: Repository Interface (~30 minutes)

**File**: `internal/identity/repository/device_authorization_repository.go`

```go
package repository

import (
 "context"
 googleUuid "github.com/google/uuid"
 cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

// DeviceAuthorizationRepository manages device authorization requests.
type DeviceAuthorizationRepository interface {
 Create(ctx context.Context, auth *cryptoutilIdentityDomain.DeviceAuthorization) error
 GetByDeviceCode(ctx context.Context, deviceCode string) (*cryptoutilIdentityDomain.DeviceAuthorization, error)
 GetByUserCode(ctx context.Context, userCode string) (*cryptoutilIdentityDomain.DeviceAuthorization, error)
 GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.DeviceAuthorization, error)
 Update(ctx context.Context, auth *cryptoutilIdentityDomain.DeviceAuthorization) error
 DeleteExpired(ctx context.Context) error
}
```

### Task 4: Repository Implementation (~45 minutes)

**File**: `internal/identity/repository/sqlrepository/device_authorization_repository.go`

Standard GORM CRUD implementation with error mapping.

### Task 5: Code Generators (~30 minutes)

**File**: `internal/identity/authz/device_code_generator.go`

```go
package authz

import (
 "crypto/rand"
 "encoding/base64"
 "fmt"
 "math/big"
 cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// GenerateDeviceCode generates a cryptographically secure device code.
func GenerateDeviceCode() (string, error) {
 bytes := make([]byte, cryptoutilIdentityMagic.DefaultDeviceCodeLength)
 if _, err := rand.Read(bytes); err != nil {
  return "", fmt.Errorf("failed to generate device code: %w", err)
 }
 return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// GenerateUserCode generates a human-readable user code (e.g., "WDJB-MJHT").
// Format: 4 uppercase letters - 4 uppercase letters.
func GenerateUserCode() (string, error) {
 const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Exclude ambiguous chars (0, O, I, 1).
 const length = 8

 code := make([]byte, length)
 for i := range code {
  num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
  if err != nil {
   return "", fmt.Errorf("failed to generate user code: %w", err)
  }
  code[i] = charset[num.Int64()]
 }

 // Format: WDJB-MJHT
 return fmt.Sprintf("%s-%s", string(code[:4]), string(code[4:])), nil
}
```

### Task 6: Handler - POST /device_authorization (~1 hour)

**File**: `internal/identity/authz/handlers_device_authorization.go`

```go
package authz

import (
 "log/slog"
 "time"
 "github.com/gofiber/fiber/v2"
 googleUuid "github.com/google/uuid"
 cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
 cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// handleDeviceAuthorization handles POST /device_authorization - RFC 8628.
func (s *Service) handleDeviceAuthorization(c *fiber.Ctx) error {
 clientID := c.FormValue(cryptoutilIdentityMagic.ParamClientID)
 scope := c.FormValue(cryptoutilIdentityMagic.ParamScope)

 if clientID == "" {
  return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
   "error": cryptoutilIdentityMagic.ErrorInvalidRequest,
   "error_description": "client_id is required",
  })
 }

 ctx := c.Context()

 // Validate client exists.
 clientRepo := s.repoFactory.ClientRepository()
 client, err := clientRepo.GetByClientID(ctx, clientID)
 if err != nil {
  return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
   "error": cryptoutilIdentityMagic.ErrorInvalidClient,
   "error_description": "Invalid client_id",
  })
 }

 // Generate device code and user code.
 deviceCode, err := GenerateDeviceCode()
 if err != nil {
  slog.ErrorContext(ctx, "Failed to generate device code", "error", err)
  return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
   "error": cryptoutilIdentityMagic.ErrorServerError,
   "error_description": "Failed to generate device code",
  })
 }

 userCode, err := GenerateUserCode()
 if err != nil {
  slog.ErrorContext(ctx, "Failed to generate user code", "error", err)
  return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
   "error": cryptoutilIdentityMagic.ErrorServerError,
   "error_description": "Failed to generate user code",
  })
 }

 // Create device authorization record.
 authID := googleUuid.Must(googleUuid.NewV7())
 deviceAuth := &cryptoutilIdentityDomain.DeviceAuthorization{
  ID:         authID,
  ClientID:   clientID,
  DeviceCode: deviceCode,
  UserCode:   userCode,
  Scope:      scope,
  Status:     "pending",
  CreatedAt:  time.Now(),
  ExpiresAt:  time.Now().Add(cryptoutilIdentityMagic.DefaultDeviceCodeLifetime),
 }

 deviceAuthRepo := s.repoFactory.DeviceAuthorizationRepository()
 if err := deviceAuthRepo.Create(ctx, deviceAuth); err != nil {
  slog.ErrorContext(ctx, "Failed to store device authorization", "error", err)
  return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
   "error": cryptoutilIdentityMagic.ErrorServerError,
   "error_description": "Failed to store device authorization",
  })
 }

 // Construct verification URIs.
 verificationURI := fmt.Sprintf("https://%s/device", c.Hostname())
 verificationURIComplete := fmt.Sprintf("%s?user_code=%s", verificationURI, userCode)

 slog.InfoContext(ctx, "Device authorization request created",
  "device_code", deviceCode[:8]+"...",
  "user_code", userCode,
  "client_id", clientID,
 )

 return c.Status(fiber.StatusOK).JSON(fiber.Map{
  "device_code":                deviceCode,
  "user_code":                  userCode,
  "verification_uri":           verificationURI,
  "verification_uri_complete":  verificationURIComplete,
  "expires_in":                 int(cryptoutilIdentityMagic.DefaultDeviceCodeLifetime.Seconds()),
  "interval":                   int(cryptoutilIdentityMagic.DefaultPollingInterval.Seconds()),
 })
}
```

### Task 7: Handler - POST /token (device_code grant) (~1.5 hours)

**File**: `internal/identity/authz/handlers_token.go` (extend existing)

```go
// handleDeviceCodeGrant handles device_code grant type (RFC 8628).
func (s *Service) handleDeviceCodeGrant(c *fiber.Ctx) error {
 deviceCode := c.FormValue(cryptoutilIdentityMagic.ParamDeviceCode)
 clientID := c.FormValue(cryptoutilIdentityMagic.ParamClientID)

 if deviceCode == "" {
  return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
   "error": cryptoutilIdentityMagic.ErrorInvalidRequest,
   "error_description": "device_code is required",
  })
 }

 if clientID == "" {
  return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
   "error": cryptoutilIdentityMagic.ErrorInvalidRequest,
   "error_description": "client_id is required",
  })
 }

 ctx := c.Context()

 // Retrieve device authorization.
 deviceAuthRepo := s.repoFactory.DeviceAuthorizationRepository()
 deviceAuth, err := deviceAuthRepo.GetByDeviceCode(ctx, deviceCode)
 if err != nil {
  return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
   "error": cryptoutilIdentityMagic.ErrorInvalidGrant,
   "error_description": "Invalid or expired device_code",
  })
 }

 // Validate client_id matches.
 if deviceAuth.ClientID != clientID {
  return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
   "error": cryptoutilIdentityMagic.ErrorInvalidGrant,
   "error_description": "client_id mismatch",
  })
 }

 // Check expiration.
 if deviceAuth.IsExpired() {
  return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
   "error": cryptoutilIdentityMagic.ErrorExpiredToken,
   "error_description": "Device code has expired",
  })
 }

 // Check status.
 switch {
 case deviceAuth.IsDenied():
  return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
   "error": cryptoutilIdentityMagic.ErrorAccessDenied,
   "error_description": "User denied the authorization request",
  })

 case deviceAuth.IsPending():
  // Check polling interval (prevent rapid polling).
  if deviceAuth.LastPolledAt != nil {
   timeSinceLastPoll := time.Since(*deviceAuth.LastPolledAt)
   if timeSinceLastPoll < cryptoutilIdentityMagic.DefaultPollingInterval {
    // Client polling too fast - return slow_down error.
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
     "error": cryptoutilIdentityMagic.ErrorSlowDown,
     "error_description": "Polling interval must be increased by 5 seconds",
    })
   }
  }

  // Update last polled timestamp.
  now := time.Now()
  deviceAuth.LastPolledAt = &now
  if err := deviceAuthRepo.Update(ctx, deviceAuth); err != nil {
   slog.ErrorContext(ctx, "Failed to update polling timestamp", "error", err)
  }

  // Still pending - return authorization_pending error.
  return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
   "error": cryptoutilIdentityMagic.ErrorAuthorizationPending,
   "error_description": "User has not yet authorized the device",
  })

 case deviceAuth.IsAuthorized():
  // User authorized - issue tokens.
  // [Token issuance logic similar to authorization_code grant]
  // ... (implement token generation using tokenSvc)

 default:
  return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
   "error": cryptoutilIdentityMagic.ErrorInvalidGrant,
   "error_description": "Invalid device authorization status",
  })
 }
}
```

### Task 8: OpenAPI Spec Updates (~45 minutes)

**File**: `api/identity/openapi_spec_authz.yaml`

Add endpoints:

- `POST /oauth2/v1/device_authorization`
- Update `POST /oauth2/v1/token` to include `device_code` grant type

### Task 9: Database Migration (~15 minutes)

**File**: `internal/identity/repository/sqlrepository/migrations/000X_device_authorization.up.sql`

```sql
CREATE TABLE IF NOT EXISTS device_authorizations (
    id TEXT PRIMARY KEY,
    client_id TEXT NOT NULL,
    device_code TEXT NOT NULL UNIQUE,
    user_code TEXT NOT NULL UNIQUE,
    scope TEXT,
    user_id TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    last_polled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP
);

CREATE INDEX idx_device_authorizations_client_id ON device_authorizations(client_id);
CREATE INDEX idx_device_authorizations_status ON device_authorizations(status);
CREATE INDEX idx_device_authorizations_expires_at ON device_authorizations(expires_at);
```

### Task 10: Unit Tests (~2 hours)

**Files**:

- `internal/identity/authz/device_code_generator_test.go`
- `internal/identity/authz/handlers_device_authorization_test.go`
- `internal/identity/authz/handlers_token_device_code_test.go`
- `internal/identity/repository/sqlrepository/device_authorization_repository_test.go`

**Coverage targets**: 95%+ for all files

### Task 11: Integration Tests (~1 hour)

**File**: `internal/identity/test/e2e/device_authorization_flow_test.go`

Test complete flow:

1. Device requests authorization
2. User visits verification URI and authorizes
3. Device polls and receives token

## Error Codes Reference

| Error | HTTP Status | Description |
|-------|-------------|-------------|
| `authorization_pending` | 400 | User has not yet authorized the device |
| `slow_down` | 400 | Polling interval too short |
| `access_denied` | 400 | User denied authorization |
| `expired_token` | 400 | Device code expired |
| `invalid_grant` | 400 | Invalid/unknown device code |
| `invalid_client` | 400 | Invalid client_id |

## Security Considerations

1. **Device code entropy**: 32 bytes (256 bits) → 43 base64url characters
2. **User code format**: 8 alphanumeric chars (no ambiguous: 0, O, I, 1) → ~34 bits entropy
3. **Polling rate limiting**: Minimum 5-second interval between polls
4. **Expiration**: 30-minute default lifetime for device codes
5. **Single-use enforcement**: Mark as "used" after successful token exchange
6. **Client validation**: Verify client_id exists and matches device authorization

## Testing Strategy

1. **Unit tests**: Code generators, handlers, repository
2. **Table-driven tests**: All error scenarios (pending, denied, expired, slow_down)
3. **Property tests**: User code uniqueness, device code uniqueness
4. **Integration tests**: Full E2E flow with real database
5. **Race detection**: Concurrent polling from multiple devices

## Next Steps After Implementation

1. Create user verification UI (`/device` endpoint in IdP service)
2. Add rate limiting per client_id
3. Add telemetry/metrics for device authorization success rate
4. Document in demo guide with CLI tool example
