# Pushed Authorization Requests (RFC 9126) Implementation Plan

**Status**: 🟡 In Progress
**RFC**: [RFC 9126](https://www.rfc-editor.org/rfc/rfc9126.html)
**Priority**: HIGH (MANDATORY)
**Estimated LOE**: 6-8 hours
**Business Value**: Security enhancement for authorization flows

---

## Overview

Pushed Authorization Requests (PAR) allows OAuth clients to push authorization request parameters directly to the authorization server before redirecting the user agent. This provides several security benefits:

1. **Request Integrity**: Authorization parameters cannot be tampered with in transit
2. **Confidentiality**: Sensitive parameters (e.g., `code_challenge`) not exposed in browser URLs
3. **Size Limits**: Removes URL length constraints for authorization requests
4. **Phishing Resistance**: Prevents interception of authorization parameters

### PAR Flow

```
Client                             Authorization Server
  |                                          |
  |---(1) POST /par ----------------------->|
  |    (authorization parameters)           |
  |                                          |
  |<--(2) request_uri + expires_in ---------|
  |                                          |
  |---(3) GET /authorize ------------------>|
  |    (request_uri only)                   |
  |                                          |
  |<--(4) redirect to callback -------------|
  |    (authorization code)                 |
```

**Key Concepts**:
- **request_uri**: Opaque reference to pushed authorization request (`urn:ietf:params:oauth:request_uri:xxx`)
- **Lifetime**: Short-lived (60-600 seconds), single-use
- **Storage**: Ephemeral storage in database with automatic expiration
- **Client Authentication**: Required for confidential clients, optional for public clients

---

## RFC 9126 Requirements

### Section 2.1: Pushed Authorization Request Endpoint

**Endpoint**: POST `/oauth2/v1/par`

**Request Parameters** (all from OAuth 2.1 /authorize):
- `client_id` (REQUIRED for public clients)
- `response_type` (REQUIRED) - "code" for authorization code flow
- `redirect_uri` (REQUIRED)
- `scope` (OPTIONAL)
- `state` (OPTIONAL but RECOMMENDED)
- `code_challenge` (REQUIRED per PKCE)
- `code_challenge_method` (REQUIRED) - "S256"
- All other standard OAuth parameters supported

**Client Authentication**:
- Confidential clients: MUST authenticate (client_secret_basic, client_secret_post, etc.)
- Public clients: MAY authenticate, MUST include client_id in body

**Response** (201 Created):
```json
{
  "request_uri": "urn:ietf:params:oauth:request_uri:6esc_11ACC5bwc014ltc14eY22c",
  "expires_in": 90
}
```

**Error Response** (400 Bad Request):
```json
{
  "error": "invalid_request",
  "error_description": "Missing required parameter: code_challenge"
}
```

### Section 2.2: Using the request_uri

**Authorization Request**: GET `/oauth2/v1/authorize?client_id=xxx&request_uri=urn:ietf:params:oauth:request_uri:xxx`

**Server Behavior**:
1. Validate request_uri format (`urn:ietf:params:oauth:request_uri:` prefix)
2. Retrieve stored authorization request parameters
3. Check expiration (single-use, time-bounded)
4. Validate client_id matches stored request
5. Proceed with normal authorization code flow using stored parameters

**Error Handling**:
- Expired request_uri → `invalid_request_uri`
- Used request_uri → `invalid_request_uri`
- Unknown request_uri → `invalid_request_uri`
- Client mismatch → `invalid_request`

---

## Implementation Tasks

### Task 1: Magic Constants (30 minutes)

**File**: `internal/identity/magic/magic_oauth.go`

Add PAR-specific constants:
```go
const (
    // PAR endpoint
    EndpointPAR = "/oauth2/v1/par"

    // PAR parameters
    ParamRequestURI    = "request_uri"
    ParamExpiresIn     = "expires_in"

    // PAR errors
    ErrorInvalidRequestURI    = "invalid_request_uri"
    ErrorInvalidRequestObject = "invalid_request_object"
)
```

**File**: `internal/identity/magic/magic_timeouts.go`

Add PAR timeout constants:
```go
const (
    // DefaultPARLifetime is the default lifetime for pushed authorization requests (90 seconds)
    DefaultPARLifetime = 90 * time.Second

    // DefaultRequestURILength is the default length for request_uri identifiers (32 bytes = ~43 chars base64url)
    DefaultRequestURILength = 32
)
```

**File**: `internal/identity/magic/magic_uris.go`

Add PAR URI prefix:
```go
const (
    // RequestURIPrefix is the URN prefix for PAR request_uri values
    RequestURIPrefix = "urn:ietf:params:oauth:request_uri:"
)
```

---

### Task 2: Domain Model (1 hour)

**File**: `internal/identity/domain/pushed_authorization_request.go`

Create domain model for PAR:
```go
package domain

import (
    "time"
    googleUuid "github.com/google/uuid"
)

// PushedAuthorizationRequest represents a pushed authorization request (RFC 9126)
type PushedAuthorizationRequest struct {
    ID                 googleUuid.UUID `gorm:"type:text;primaryKey"`
    RequestURI         string          `gorm:"type:text;uniqueIndex;not null"` // urn:ietf:params:oauth:request_uri:xxx
    ClientID           googleUuid.UUID `gorm:"type:text;index;not null"`

    // Stored authorization parameters (JSON serialized)
    ResponseType       string   `gorm:"type:text;not null"`
    RedirectURI        string   `gorm:"type:text;not null"`
    Scope              string   `gorm:"type:text"`
    State              string   `gorm:"type:text"`
    CodeChallenge      string   `gorm:"type:text;not null"`
    CodeChallengeMethod string  `gorm:"type:text;not null"`
    Nonce              string   `gorm:"type:text"`

    // Additional parameters as JSON blob
    AdditionalParams   string   `gorm:"type:text;serializer:json"`

    // Lifecycle tracking
    Used               bool     `gorm:"not null;default:false;index"`
    ExpiresAt          time.Time `gorm:"not null;index"`
    CreatedAt          time.Time `gorm:"not null"`
    UsedAt             *time.Time
}

// IsExpired checks if the request has expired
func (p *PushedAuthorizationRequest) IsExpired() bool {
    return time.Now().UTC().After(p.ExpiresAt)
}

// IsUsed checks if the request has already been used
func (p *PushedAuthorizationRequest) IsUsed() bool {
    return p.Used
}

// MarkAsUsed marks the request as used and records the timestamp
func (p *PushedAuthorizationRequest) MarkAsUsed() {
    p.Used = true
    now := time.Now().UTC()
    p.UsedAt = &now
}
```

**Tests**: `internal/identity/domain/pushed_authorization_request_test.go`

Test coverage:
- IsExpired() - expired vs not expired
- IsUsed() - used vs not used
- MarkAsUsed() - sets flag and timestamp

---

### Task 3: Repository Interface (30 minutes)

**File**: `internal/identity/repository/pushed_authorization_request_repository.go`

Create repository interface:
```go
package repository

import (
    "context"
    googleUuid "github.com/google/uuid"
    cryptoutilIdentityDomain "github.com/soyrochus/cryptoutil/internal/identity/domain"
)

// PushedAuthorizationRequestRepository manages pushed authorization requests (RFC 9126)
type PushedAuthorizationRequestRepository interface {
    Create(ctx context.Context, req *cryptoutilIdentityDomain.PushedAuthorizationRequest) error
    GetByRequestURI(ctx context.Context, requestURI string) (*cryptoutilIdentityDomain.PushedAuthorizationRequest, error)
    GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.PushedAuthorizationRequest, error)
    Update(ctx context.Context, req *cryptoutilIdentityDomain.PushedAuthorizationRequest) error
    DeleteExpired(ctx context.Context) (int64, error)
}
```

---

### Task 4: Repository Implementation (1 hour)

**File**: `internal/identity/repository/orm/pushed_authorization_request_repository.go`

Implement GORM repository:
```go
package orm

import (
    "context"
    "errors"
    "fmt"
    "time"

    googleUuid "github.com/google/uuid"
    "gorm.io/gorm"

    cryptoutilIdentityApperr "github.com/soyrochus/cryptoutil/internal/identity/apperr"
    cryptoutilIdentityDomain "github.com/soyrochus/cryptoutil/internal/identity/domain"
)

type pushedAuthorizationRequestRepository struct {
    db *gorm.DB
}

func NewPushedAuthorizationRequestRepository(db *gorm.DB) *pushedAuthorizationRequestRepository {
    return &pushedAuthorizationRequestRepository{db: db}
}

func (r *pushedAuthorizationRequestRepository) Create(ctx context.Context, req *cryptoutilIdentityDomain.PushedAuthorizationRequest) error {
    if err := r.db.WithContext(ctx).Create(req).Error; err != nil {
        return fmt.Errorf("failed to create pushed authorization request: %w", err)
    }
    return nil
}

func (r *pushedAuthorizationRequestRepository) GetByRequestURI(ctx context.Context, requestURI string) (*cryptoutilIdentityDomain.PushedAuthorizationRequest, error) {
    var req cryptoutilIdentityDomain.PushedAuthorizationRequest
    if err := r.db.WithContext(ctx).Where("request_uri = ?", requestURI).First(&req).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, cryptoutilIdentityApperr.ErrPushedAuthorizationRequestNotFound
        }
        return nil, fmt.Errorf("failed to get pushed authorization request: %w", err)
    }
    return &req, nil
}

func (r *pushedAuthorizationRequestRepository) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.PushedAuthorizationRequest, error) {
    var req cryptoutilIdentityDomain.PushedAuthorizationRequest
    if err := r.db.WithContext(ctx).First(&req, "id = ?", id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, cryptoutilIdentityApperr.ErrPushedAuthorizationRequestNotFound
        }
        return nil, fmt.Errorf("failed to get pushed authorization request: %w", err)
    }
    return &req, nil
}

func (r *pushedAuthorizationRequestRepository) Update(ctx context.Context, req *cryptoutilIdentityDomain.PushedAuthorizationRequest) error {
    if err := r.db.WithContext(ctx).Save(req).Error; err != nil {
        return fmt.Errorf("failed to update pushed authorization request: %w", err)
    }
    return nil
}

func (r *pushedAuthorizationRequestRepository) DeleteExpired(ctx context.Context) (int64, error) {
    result := r.db.WithContext(ctx).Where("expires_at < ?", time.Now().UTC()).Delete(&cryptoutilIdentityDomain.PushedAuthorizationRequest{})
    if result.Error != nil {
        return 0, fmt.Errorf("failed to delete expired pushed authorization requests: %w", result.Error)
    }
    return result.RowsAffected, nil
}
```

**Error Definition**: `internal/identity/apperr/errors.go`

Add error constant:
```go
var ErrPushedAuthorizationRequestNotFound = &AppError{
    Code:    "pushed_authorization_request_not_found",
    Message: "Pushed authorization request not found",
    Status:  http.StatusNotFound,
}
```

---

### Task 5: Database Migration (30 minutes)

**File**: `internal/identity/repository/orm/migrations/000011_pushed_authorization_request.up.sql`

Create migration:
```sql
-- Pushed Authorization Requests table (RFC 9126)
CREATE TABLE IF NOT EXISTS pushed_authorization_requests (
    id TEXT PRIMARY KEY,
    request_uri TEXT NOT NULL UNIQUE,
    client_id TEXT NOT NULL,

    -- Authorization parameters
    response_type TEXT NOT NULL,
    redirect_uri TEXT NOT NULL,
    scope TEXT,
    state TEXT,
    code_challenge TEXT NOT NULL,
    code_challenge_method TEXT NOT NULL,
    nonce TEXT,

    -- Additional parameters (JSON)
    additional_params TEXT,

    -- Lifecycle tracking
    used BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used_at TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_par_client_id ON pushed_authorization_requests(client_id);
CREATE INDEX IF NOT EXISTS idx_par_expires_at ON pushed_authorization_requests(expires_at);
CREATE INDEX IF NOT EXISTS idx_par_used ON pushed_authorization_requests(used);
```

**File**: `internal/identity/repository/orm/migrations/000011_pushed_authorization_request.down.sql`

```sql
DROP TABLE IF EXISTS pushed_authorization_requests;
```

---

### Task 6: Request URI Generator (30 minutes)

**File**: `internal/identity/authz/request_uri_generator.go`

Create request_uri generator:
```go
package authz

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"

    cryptoutilMagic "github.com/soyrochus/cryptoutil/internal/common/magic"
)

// GenerateRequestURI generates a cryptographically random request_uri per RFC 9126
// Format: urn:ietf:params:oauth:request_uri:<base64url-encoded-random-bytes>
func GenerateRequestURI() (string, error) {
    randomBytes := make([]byte, cryptoutilMagic.DefaultRequestURILength)
    if _, err := rand.Read(randomBytes); err != nil {
        return "", fmt.Errorf("failed to generate random bytes for request_uri: %w", err)
    }

    // Base64url encode without padding
    encoded := base64.RawURLEncoding.EncodeToString(randomBytes)

    return cryptoutilMagic.RequestURIPrefix + encoded, nil
}
```

**Tests**: `internal/identity/authz/request_uri_generator_test.go`

Test coverage:
- Uniqueness (1000 samples)
- Format (starts with `urn:ietf:params:oauth:request_uri:`)
- Length (≥43 characters)
- No collisions

---

### Task 7: PAR Handler (2 hours)

**File**: `internal/identity/authz/handlers_par.go`

Implement POST /par endpoint:
```go
package authz

import (
    "fmt"
    "time"

    "github.com/gofiber/fiber/v2"
    googleUuid "github.com/google/uuid"

    cryptoutilIdentityDomain "github.com/soyrochus/cryptoutil/internal/identity/domain"
    cryptoutilMagic "github.com/soyrochus/cryptoutil/internal/common/magic"
)

// PARResponse represents the response from POST /par (RFC 9126 Section 2.1)
type PARResponse struct {
    RequestURI string `json:"request_uri"`
    ExpiresIn  int    `json:"expires_in"`
}

// handlePAR handles POST /oauth2/v1/par - Pushed Authorization Request endpoint (RFC 9126)
func (s *Server) handlePAR(c *fiber.Ctx) error {
    ctx := c.Context()

    // 1. Extract and validate client authentication
    clientID, err := s.extractAndValidateClient(c)
    if err != nil {
        return s.sendOAuthError(c, fiber.StatusBadRequest, cryptoutilMagic.ErrorInvalidClient, "Client authentication failed")
    }

    // 2. Parse authorization request parameters
    responseType := c.FormValue("response_type")
    redirectURI := c.FormValue("redirect_uri")
    scope := c.FormValue("scope")
    state := c.FormValue("state")
    codeChallenge := c.FormValue("code_challenge")
    codeChallengeMethod := c.FormValue("code_challenge_method")
    nonce := c.FormValue("nonce")

    // 3. Validate required parameters
    if responseType == "" {
        return s.sendOAuthError(c, fiber.StatusBadRequest, cryptoutilMagic.ErrorInvalidRequest, "Missing required parameter: response_type")
    }
    if redirectURI == "" {
        return s.sendOAuthError(c, fiber.StatusBadRequest, cryptoutilMagic.ErrorInvalidRequest, "Missing required parameter: redirect_uri")
    }
    if codeChallenge == "" {
        return s.sendOAuthError(c, fiber.StatusBadRequest, cryptoutilMagic.ErrorInvalidRequest, "Missing required parameter: code_challenge")
    }
    if codeChallengeMethod == "" {
        return s.sendOAuthError(c, fiber.StatusBadRequest, cryptoutilMagic.ErrorInvalidRequest, "Missing required parameter: code_challenge_method")
    }

    // 4. Validate response_type (only "code" supported)
    if responseType != "code" {
        return s.sendOAuthError(c, fiber.StatusBadRequest, cryptoutilMagic.ErrorUnsupportedResponseType, "Only response_type=code is supported")
    }

    // 5. Validate code_challenge_method (only S256 supported per PKCE)
    if codeChallengeMethod != "S256" {
        return s.sendOAuthError(c, fiber.StatusBadRequest, cryptoutilMagic.ErrorInvalidRequest, "Only code_challenge_method=S256 is supported")
    }

    // 6. Validate redirect_uri against client configuration
    client, err := s.clientRepo.GetByID(ctx, clientID)
    if err != nil {
        return s.sendOAuthError(c, fiber.StatusBadRequest, cryptoutilMagic.ErrorInvalidClient, "Client not found")
    }

    if !contains(client.RedirectURIs, redirectURI) {
        return s.sendOAuthError(c, fiber.StatusBadRequest, cryptoutilMagic.ErrorInvalidRequest, "redirect_uri not registered for client")
    }

    // 7. Generate request_uri
    requestURI, err := GenerateRequestURI()
    if err != nil {
        s.logger.Error("Failed to generate request_uri", "error", err)
        return s.sendOAuthError(c, fiber.StatusInternalServerError, cryptoutilMagic.ErrorServerError, "Internal server error")
    }

    // 8. Create and store PushedAuthorizationRequest
    now := time.Now().UTC()
    expiresAt := now.Add(cryptoutilMagic.DefaultPARLifetime)

    par := &cryptoutilIdentityDomain.PushedAuthorizationRequest{
        ID:                  googleUuid.New(),
        RequestURI:          requestURI,
        ClientID:            clientID,
        ResponseType:        responseType,
        RedirectURI:         redirectURI,
        Scope:               scope,
        State:               state,
        CodeChallenge:       codeChallenge,
        CodeChallengeMethod: codeChallengeMethod,
        Nonce:               nonce,
        Used:                false,
        ExpiresAt:           expiresAt,
        CreatedAt:           now,
    }

    if err := s.parRepo.Create(ctx, par); err != nil {
        s.logger.Error("Failed to create pushed authorization request", "error", err)
        return s.sendOAuthError(c, fiber.StatusInternalServerError, cryptoutilMagic.ErrorServerError, "Internal server error")
    }

    // 9. Return response (201 Created)
    response := PARResponse{
        RequestURI: requestURI,
        ExpiresIn:  int(cryptoutilMagic.DefaultPARLifetime.Seconds()),
    }

    return c.Status(fiber.StatusCreated).JSON(response)
}

// extractAndValidateClient extracts and validates client credentials from request
// Supports client_secret_basic, client_secret_post, and public clients
func (s *Server) extractAndValidateClient(c *fiber.Ctx) (googleUuid.UUID, error) {
    // Try client_secret_basic (HTTP Basic Auth)
    clientID, clientSecret, ok := c.Request().URI().QueryArgs().GetUfloatOrDefault("client_id", 0) != 0
    if ok {
        // Authenticate with client secret
        // ... (implementation similar to existing client auth)
    }

    // Try client_secret_post (form body)
    clientIDStr := c.FormValue("client_id")
    if clientIDStr != "" {
        clientID, err := googleUuid.Parse(clientIDStr)
        if err != nil {
            return googleUuid.Nil, fmt.Errorf("invalid client_id format")
        }
        return clientID, nil
    }

    return googleUuid.Nil, fmt.Errorf("missing client credentials")
}
```

**Tests**: `internal/identity/authz/handlers_par_test.go`

Test coverage:
- Happy path (valid PAR request returns 201 with request_uri)
- Missing response_type (400 error)
- Missing redirect_uri (400 error)
- Missing code_challenge (400 error)
- Invalid redirect_uri (400 error)
- Invalid client_id (400 error)
- Unsupported response_type (400 error)
- Unsupported code_challenge_method (400 error)

---

### Task 8: Modify /authorize Handler (1.5 hours)

**File**: `internal/identity/authz/handlers_authorize.go`

Modify existing handleAuthorize to support request_uri parameter:

```go
func (s *Server) handleAuthorize(c *fiber.Ctx) error {
    ctx := c.Context()

    // Check for request_uri parameter (PAR flow)
    requestURI := c.Query("request_uri")
    if requestURI != "" {
        return s.handleAuthorizeWithPAR(c, requestURI)
    }

    // Existing authorization code flow logic
    // ...
}

// handleAuthorizeWithPAR processes authorization request using PAR request_uri
func (s *Server) handleAuthorizeWithPAR(c *fiber.Ctx, requestURI string) error {
    ctx := c.Context()

    // 1. Validate request_uri format
    if !strings.HasPrefix(requestURI, cryptoutilMagic.RequestURIPrefix) {
        return s.sendOAuthError(c, fiber.StatusBadRequest, cryptoutilMagic.ErrorInvalidRequestURI, "Invalid request_uri format")
    }

    // 2. Retrieve stored PAR
    par, err := s.parRepo.GetByRequestURI(ctx, requestURI)
    if err != nil {
        return s.sendOAuthError(c, fiber.StatusBadRequest, cryptoutilMagic.ErrorInvalidRequestURI, "request_uri not found or expired")
    }

    // 3. Validate expiration
    if par.IsExpired() {
        return s.sendOAuthError(c, fiber.StatusBadRequest, cryptoutilMagic.ErrorInvalidRequestURI, "request_uri has expired")
    }

    // 4. Validate single-use
    if par.IsUsed() {
        return s.sendOAuthError(c, fiber.StatusBadRequest, cryptoutilMagic.ErrorInvalidRequestURI, "request_uri already used")
    }

    // 5. Validate client_id matches (required in query string)
    clientIDStr := c.Query("client_id")
    if clientIDStr == "" {
        return s.sendOAuthError(c, fiber.StatusBadRequest, cryptoutilMagic.ErrorInvalidRequest, "Missing required parameter: client_id")
    }

    clientID, err := googleUuid.Parse(clientIDStr)
    if err != nil || clientID != par.ClientID {
        return s.sendOAuthError(c, fiber.StatusBadRequest, cryptoutilMagic.ErrorInvalidRequest, "client_id mismatch")
    }

    // 6. Mark PAR as used
    par.MarkAsUsed()
    if err := s.parRepo.Update(ctx, par); err != nil {
        s.logger.Error("Failed to mark PAR as used", "error", err)
        return s.sendOAuthError(c, fiber.StatusInternalServerError, cryptoutilMagic.ErrorServerError, "Internal server error")
    }

    // 7. Proceed with normal authorization flow using PAR parameters
    // (inject PAR parameters into request context for downstream handlers)
    c.Locals("par_response_type", par.ResponseType)
    c.Locals("par_redirect_uri", par.RedirectURI)
    c.Locals("par_scope", par.Scope)
    c.Locals("par_state", par.State)
    c.Locals("par_code_challenge", par.CodeChallenge)
    c.Locals("par_code_challenge_method", par.CodeChallengeMethod)
    c.Locals("par_nonce", par.Nonce)

    // Continue with existing authorize flow logic
    // ...
}
```

**Integration Points**:
- Modify session creation to use PAR parameters from c.Locals()
- Update PKCE validation to use stored code_challenge from PAR
- Ensure redirect_uri from PAR is used for callback

---

### Task 9: Route Registration (15 minutes)

**File**: `internal/identity/authz/routes.go`

Add PAR endpoint:
```go
func (s *Server) RegisterRoutes(app *fiber.App) {
    oauth := app.Group("/oauth2/v1")

    // Existing routes
    oauth.Post("/token", s.handleToken)
    oauth.Post("/introspect", s.handleIntrospect)
    oauth.Post("/revoke", s.handleRevoke)
    oauth.Get("/authorize", s.handleAuthorize)
    oauth.Post("/authorize", s.handleAuthorize)
    oauth.Post("/device_authorization", s.handleDeviceAuthorization)

    // NEW: PAR endpoint
    oauth.Post("/par", s.handlePAR)

    // ... rest of routes
}
```

---

### Task 10: Factory Integration (15 minutes)

**File**: `internal/identity/repository/factory.go`

Add PAR repository to factory:
```go
type RepositoryFactory struct {
    db                     *gorm.DB
    userRepo               UserRepository
    clientRepo             ClientRepository
    authCodeRepo           AuthorizationCodeRepository
    sessionRepo            SessionRepository
    deviceAuthRepo         DeviceAuthorizationRepository
    parRepo                PushedAuthorizationRequestRepository  // NEW
    // ... other repos
}

func NewRepositoryFactory(db *gorm.DB) (*RepositoryFactory, error) {
    return &RepositoryFactory{
        db:             db,
        userRepo:       cryptoutilIdentityORM.NewUserRepository(db),
        clientRepo:     cryptoutilIdentityORM.NewClientRepository(db),
        authCodeRepo:   cryptoutilIdentityORM.NewAuthorizationCodeRepository(db),
        sessionRepo:    cryptoutilIdentityORM.NewSessionRepository(db),
        deviceAuthRepo: cryptoutilIdentityORM.NewDeviceAuthorizationRepository(db),
        parRepo:        cryptoutilIdentityORM.NewPushedAuthorizationRequestRepository(db),  // NEW
        // ... other repos
    }, nil
}

func (f *RepositoryFactory) PushedAuthorizationRequestRepository() PushedAuthorizationRequestRepository {
    return f.parRepo
}
```

---

### Task 11: Unit Tests (1.5 hours)

**File**: `internal/identity/domain/pushed_authorization_request_test.go`

Test domain model methods:
- TestPushedAuthorizationRequest_IsExpired
- TestPushedAuthorizationRequest_IsUsed
- TestPushedAuthorizationRequest_MarkAsUsed

**File**: `internal/identity/authz/request_uri_generator_test.go`

Test request_uri generation:
- TestGenerateRequestURI_Uniqueness (1000 samples)
- TestGenerateRequestURI_Format (URN prefix)
- TestGenerateRequestURI_Length (≥43 chars)

**File**: `internal/identity/authz/handlers_par_test.go`

Test PAR handler:
- TestHandlePAR_HappyPath (201 with request_uri)
- TestHandlePAR_MissingResponseType (400 error)
- TestHandlePAR_MissingRedirectURI (400 error)
- TestHandlePAR_MissingCodeChallenge (400 error)
- TestHandlePAR_InvalidRedirectURI (400 error)
- TestHandlePAR_InvalidClient (400 error)

---

### Task 12: Integration Tests (1.5 hours)

**File**: `internal/identity/authz/handlers_par_flow_integration_test.go`

Test PAR E2E flow:
- TestPARFlow_HappyPath:
  1. POST /par → get request_uri
  2. GET /authorize with request_uri → redirect to login
  3. POST /login → redirect to consent
  4. POST /consent → redirect to callback with code
  5. POST /token with code → get tokens

- TestPARFlow_ExpiredRequestURI:
  1. POST /par → get request_uri
  2. Manually expire PAR in database
  3. GET /authorize with request_uri → 400 invalid_request_uri

- TestPARFlow_UsedRequestURI:
  1. POST /par → get request_uri
  2. GET /authorize with request_uri → success
  3. GET /authorize with same request_uri → 400 invalid_request_uri (used)

- TestPARFlow_ClientMismatch:
  1. Create client A, POST /par → get request_uri
  2. GET /authorize with client B's client_id → 400 client_id mismatch

---

### Task 13: Cleanup Job (Optional, 30 minutes)

**File**: `internal/identity/authz/cleanup_expired_par.go`

Background job to delete expired PAR entries:
```go
package authz

import (
    "context"
    "time"
)

// StartPARCleanupJob starts a background goroutine to periodically delete expired PAR entries
func (s *Server) StartPARCleanupJob(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            deleted, err := s.parRepo.DeleteExpired(ctx)
            if err != nil {
                s.logger.Error("Failed to delete expired PAR entries", "error", err)
            } else if deleted > 0 {
                s.logger.Info("Deleted expired PAR entries", "count", deleted)
            }
        case <-ctx.Done():
            return
        }
    }
}
```

Call from server startup:
```go
// Start PAR cleanup job (runs every 5 minutes)
go s.StartPARCleanupJob(ctx, 5*time.Minute)
```

---

## Security Considerations

### Request Integrity

PAR protects authorization parameters from:
- **URL tampering**: Parameters stored server-side, only opaque request_uri exposed
- **Phishing attacks**: Attackers cannot intercept/modify authorization parameters
- **Parameter injection**: All parameters validated before storage

### Confidentiality

PAR prevents exposure of sensitive data:
- **PKCE code_challenge**: Not visible in browser URL bar or HTTP logs
- **State parameter**: Protected from interception
- **Custom parameters**: Kept confidential on server

### Replay Protection

Single-use request_uri enforcement:
- Mark as used after first /authorize request
- Return error on subsequent attempts
- Automatic cleanup of expired entries

### Client Authentication

Confidential clients MUST authenticate at /par:
- Prevents unauthorized clients from pushing requests
- Links request_uri to authenticated client
- Client_id validated again at /authorize

### Lifetime Management

Short-lived request_uri (90 seconds):
- Reduces window for attacks
- Prevents stale authorization requests
- Automatic expiration cleanup

---

## Testing Strategy

### Unit Tests (6 tests)

**Domain Model**:
- IsExpired(), IsUsed(), MarkAsUsed() methods

**Request URI Generator**:
- Uniqueness, format, length validation

**PAR Handler**:
- Happy path, missing parameters, invalid client

### Integration Tests (4 tests)

**E2E Flow**:
- PAR → Authorize → Login → Consent → Token
- Expired request_uri handling
- Used request_uri handling
- Client_id mismatch handling

### Manual Testing

**Postman/curl**:
```bash
# 1. Push authorization request
curl -X POST https://localhost:8080/oauth2/v1/par \
  -d "client_id=xxx" \
  -d "client_secret=yyy" \
  -d "response_type=code" \
  -d "redirect_uri=https://example.com/callback" \
  -d "scope=openid profile" \
  -d "state=random-state" \
  -d "code_challenge=xxx" \
  -d "code_challenge_method=S256"

# Response: {"request_uri": "urn:ietf:params:oauth:request_uri:xxx", "expires_in": 90}

# 2. Use request_uri in authorization request
curl -X GET "https://localhost:8080/oauth2/v1/authorize?client_id=xxx&request_uri=urn:ietf:params:oauth:request_uri:xxx"
```

---

## Success Criteria

- ✅ All 13 tasks completed
- ✅ 10+ unit tests passing (100% pass rate)
- ✅ 4+ integration tests passing (E2E flow validated)
- ✅ RFC 9126 Sections 2.1-2.2 fully implemented
- ✅ Database migration applied successfully
- ✅ PAR endpoint operational and tested
- ✅ /authorize handler supports request_uri parameter
- ✅ Security requirements met (single-use, expiration, client validation)
- ✅ Pre-commit hooks satisfied (linting, formatting)

---

## Files Changed Summary

**New Files** (~1,200 lines):
- `internal/identity/domain/pushed_authorization_request.go` (90 lines)
- `internal/identity/domain/pushed_authorization_request_test.go` (80 lines)
- `internal/identity/repository/pushed_authorization_request_repository.go` (30 lines)
- `internal/identity/repository/orm/pushed_authorization_request_repository.go` (120 lines)
- `internal/identity/repository/orm/migrations/000011_pushed_authorization_request.up.sql` (30 lines)
- `internal/identity/repository/orm/migrations/000011_pushed_authorization_request.down.sql` (5 lines)
- `internal/identity/authz/request_uri_generator.go` (30 lines)
- `internal/identity/authz/request_uri_generator_test.go` (80 lines)
- `internal/identity/authz/handlers_par.go` (150 lines)
- `internal/identity/authz/handlers_par_test.go` (250 lines)
- `internal/identity/authz/handlers_par_flow_integration_test.go` (350 lines)
- `internal/identity/authz/cleanup_expired_par.go` (40 lines)

**Modified Files**:
- `internal/identity/magic/magic_oauth.go` (+10 lines)
- `internal/identity/magic/magic_timeouts.go` (+10 lines)
- `internal/identity/magic/magic_uris.go` (+5 lines)
- `internal/identity/apperr/errors.go` (+5 lines)
- `internal/identity/repository/factory.go` (+15 lines)
- `internal/identity/authz/routes.go` (+2 lines)
- `internal/identity/authz/handlers_authorize.go` (+100 lines for PAR integration)

**Total**: ~1,400 lines added/modified

---

## Lessons Learned

(To be updated after implementation)

---

*Implementation Plan Version: 1.0.0*
*Author: GitHub Copilot (Agent)*
*Created: 2025-01-08*
