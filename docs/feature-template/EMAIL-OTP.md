# Email OTP Implementation Plan

**Status**: 🟡 In Progress (30% → 100%)
**Priority**: MEDIUM (MANDATORY)
**Estimated LOE**: 4-6 hours
**Business Value**: Account recovery, passwordless authentication

---

## Overview

Email OTP (One-Time Password) enables users to authenticate using codes sent to their registered email addresses. This provides both primary authentication (passwordless login) and secondary authentication (MFA factor).

### Use Cases

1. **Passwordless Login**: User enters email → receives OTP → logs in without password
2. **MFA Factor**: User logs in with password → receives OTP → completes authentication
3. **Account Recovery**: User loses access to other MFA factors → uses email OTP
4. **Email Verification**: Confirm email ownership during registration

---

## Current Implementation (30%)

Based on spec.md status "⚠️ 30% (missing: email delivery service, rate limiting - MANDATORY)":

### What Exists (30%)
- Domain model: `internal/identity/domain/email_otp.go` (likely partial)
- Repository interface: `internal/identity/repository/email_otp_repository.go` (likely exists)
- Basic OTP generation: Probably uses similar pattern to recovery codes

### What's Missing (70%)
1. **Email Delivery Service** - SMTP integration, template engine
2. **Rate Limiting** - Prevent abuse (5 OTP requests per hour)
3. **Repository Implementation** - GORM implementation if not complete
4. **Database Migration** - Schema for email_otp table
5. **API Handlers** - POST /oidc/v1/mfa/email-otp/send, POST /oidc/v1/mfa/email-otp/verify
6. **Template System** - HTML/plaintext email templates
7. **Comprehensive Tests** - Service layer, handler tests

---

## Implementation Tasks

### Task 1: Discovery - Assess Existing Code (30min)

**Files to Check**:
- `internal/identity/domain/email_otp.go`
- `internal/identity/repository/email_otp_repository.go`
- `internal/identity/repository/orm/email_otp_repository.go`
- `internal/identity/repository/orm/migrations/*_email_otp*.sql`
- `internal/identity/magic/magic_mfa.go` (OTP-related constants)

**Actions**:
1. Read existing domain model
2. Check repository interface methods
3. Verify database migration status
4. Identify missing pieces

---

### Task 2: Magic Constants (if missing) (15min)

**File**: `internal/identity/magic/magic_mfa.go`

Add email OTP constants:
```go
const (
    // Email OTP Configuration
    EmailOTPLength          = 6                       // 6-digit code
    EmailOTPLifetime        = 10 * time.Minute        // 10 minutes
    EmailOTPCharset         = "0123456789"            // Numeric only
    EmailOTPRateLimit       = 5                       // 5 requests per hour
    EmailOTPRateLimitWindow = 1 * time.Hour
)
```

---

### Task 3: Domain Model (if incomplete) (30min)

**File**: `internal/identity/domain/email_otp.go`

Complete domain model:
```go
type EmailOTP struct {
    ID        googleUuid.UUID `gorm:"type:text;primaryKey"`
    UserID    googleUuid.UUID `gorm:"type:text;index;not null"`
    Email     string          `gorm:"type:text;index;not null"`
    CodeHash  string          `gorm:"type:text;not null"` // bcrypt hash
    Used      bool            `gorm:"not null;default:false;index"`
    UsedAt    *time.Time      `gorm:"index"`
    CreatedAt time.Time       `gorm:"not null"`
    ExpiresAt time.Time       `gorm:"not null;index"`
}

func IsExpired() bool
func IsUsed() bool
func MarkAsUsed()
```

---

### Task 4: Email Delivery Service (2h)

**File**: `internal/identity/email/email_service.go` (NEW)

Create email service abstraction:
```go
package email

import (
    "context"
    "fmt"
    "net/smtp"
)

type EmailService interface {
    SendOTP(ctx context.Context, to, code string) error
}

type SMTPEmailService struct {
    host     string
    port     int
    username string
    password string
    from     string
}

func NewSMTPEmailService(host string, port int, username, password, from string) *SMTPEmailService

func (s *SMTPEmailService) SendOTP(ctx context.Context, to, code string) error {
    // Generate email body from template
    subject := "Your One-Time Password"
    body := fmt.Sprintf(emailOTPTemplate, code)
    
    // Connect to SMTP server
    auth := smtp.PlainAuth("", s.username, s.password, s.host)
    addr := fmt.Sprintf("%s:%d", s.host, s.port)
    
    // Send email
    msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", s.from, to, subject, body)
    return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
}

// For testing: MockEmailService stores sent emails in memory
type MockEmailService struct {
    SentEmails []struct{ To, Code string }
}

func (m *MockEmailService) SendOTP(ctx context.Context, to, code string) error {
    m.SentEmails = append(m.SentEmails, struct{ To, Code string }{to, code})
    return nil
}
```

**Email Template**:
```
Your One-Time Password: {{CODE}}

This code expires in 10 minutes.

If you did not request this code, please ignore this email.
```

---

### Task 5: Email OTP Generator (30min)

**File**: `internal/identity/mfa/email_otp_generator.go`

Generate 6-digit numeric OTP:
```go
func GenerateEmailOTP() (string, error) {
    const otpLength = 6
    charset := "0123456789"
    
    randomBytes := make([]byte, otpLength)
    if _, err := rand.Read(randomBytes); err != nil {
        return "", err
    }
    
    code := make([]byte, otpLength)
    for i := range otpLength {
        code[i] = charset[int(randomBytes[i])%len(charset)]
    }
    
    return string(code), nil
}
```

---

### Task 6: Rate Limiting Service (1h)

**File**: `internal/identity/mfa/rate_limiter.go`

Implement per-user rate limiting:
```go
type RateLimiter interface {
    Allow(ctx context.Context, userID googleUuid.UUID) (bool, error)
}

type InMemoryRateLimiter struct {
    requests map[string][]time.Time  // userID -> request timestamps
    limit    int                      // max requests
    window   time.Duration            // time window
    mu       sync.Mutex
}

func (r *InMemoryRateLimiter) Allow(ctx context.Context, userID googleUuid.UUID) (bool, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    key := userID.String()
    now := time.Now().UTC()
    cutoff := now.Add(-r.window)
    
    // Remove old requests
    var recent []time.Time
    for _, t := range r.requests[key] {
        if t.After(cutoff) {
            recent = append(recent, t)
        }
    }
    
    // Check limit
    if len(recent) >= r.limit {
        return false, nil
    }
    
    // Add new request
    recent = append(recent, now)
    r.requests[key] = recent
    
    return true, nil
}
```

---

### Task 7: Email OTP Service (1.5h)

**File**: `internal/identity/mfa/email_otp_service.go`

Service methods:
```go
type EmailOTPService struct {
    repo         EmailOTPRepository
    emailService EmailService
    rateLimiter  RateLimiter
}

func (s *EmailOTPService) SendOTP(ctx context.Context, userID googleUuid.UUID, email string) error {
    // Check rate limit
    allowed, err := s.rateLimiter.Allow(ctx, userID)
    if err != nil {
        return err
    }
    if !allowed {
        return ErrRateLimitExceeded
    }
    
    // Generate OTP
    code, err := GenerateEmailOTP()
    if err != nil {
        return err
    }
    
    // Hash code
    hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    
    // Store in database
    otp := &EmailOTP{
        ID:        googleUuid.New(),
        UserID:    userID,
        Email:     email,
        CodeHash:  string(hash),
        Used:      false,
        CreatedAt: time.Now().UTC(),
        ExpiresAt: time.Now().UTC().Add(EmailOTPLifetime),
    }
    if err := s.repo.Create(ctx, otp); err != nil {
        return err
    }
    
    // Send email
    return s.emailService.SendOTP(ctx, email, code)
}

func (s *EmailOTPService) VerifyOTP(ctx context.Context, userID googleUuid.UUID, code string) error {
    // Get all active OTPs for user
    otps, err := s.repo.GetActiveByUserID(ctx, userID)
    if err != nil {
        return err
    }
    
    // Find matching OTP
    for _, otp := range otps {
        if otp.IsUsed() || otp.IsExpired() {
            continue
        }
        
        if err := bcrypt.CompareHashAndPassword([]byte(otp.CodeHash), []byte(code)); err == nil {
            // Match found - mark as used
            otp.MarkAsUsed()
            return s.repo.Update(ctx, otp)
        }
    }
    
    return ErrInvalidOTP
}
```

---

### Task 8: API Handlers (1h)

**Endpoints**:

1. **POST /oidc/v1/mfa/email-otp/send**
   - Request: `{"user_id": "uuid", "email": "user@example.com"}`
   - Response (200 OK): `{"sent": true, "expires_at": "2025-01-09T..."}`
   - Errors: 400 invalid_request, 404 user_not_found, 429 rate_limit_exceeded, 500 server_error

2. **POST /oidc/v1/mfa/email-otp/verify**
   - Request: `{"code": "123456"}` + `X-User-ID` header
   - Response (200 OK): `{"verified": true}`
   - Errors: 400 invalid_request, 401 invalid_code, 500 server_error

---

### Task 9: Unit Tests (1.5h)

**Test Files**:
- `email_otp_generator_test.go` (4 tests: format, length, uniqueness, numeric-only)
- `email_service_test.go` (3 tests: send success, SMTP error, mock service)
- `rate_limiter_test.go` (5 tests: allow, exceed, window expiry, concurrent, reset)
- `email_otp_service_test.go` (8 tests: send, verify, rate limit, expiration, used, invalid)

---

### Task 10: Handler Tests (1h)

**Test Cases**:
- TestHandleSendEmailOTP_HappyPath
- TestHandleSendEmailOTP_RateLimitExceeded
- TestHandleSendEmailOTP_InvalidEmail
- TestHandleVerifyEmailOTP_HappyPath
- TestHandleVerifyEmailOTP_InvalidCode
- TestHandleVerifyEmailOTP_ExpiredCode

---

## Configuration

**Config File**: `configs/identity/identity-common.yml`

Add email service config:
```yaml
email:
  smtp:
    host: "smtp.example.com"
    port: 587
    username: "notifications@example.com"
    password: "secret"  # Use Docker secrets in production
    from: "CryptoUtil <notifications@example.com>"
  otp:
    lifetime: "10m"
    rate_limit: 5
    rate_limit_window: "1h"
```

---

## Security Considerations

### Code Storage ✅
- Hashed with bcrypt (like recovery codes)
- Plaintext code NEVER stored
- Code shown only in email (transport security via TLS)

### Rate Limiting ✅
- 5 OTP requests per hour per user
- Prevents email bombing attacks
- Returns 429 Too Many Requests

### Expiration ✅
- 10-minute default lifetime
- Shorter than recovery codes (more convenient, less secure)

### Single-Use ✅
- Mark as used after successful verification
- Prevents replay attacks

---

## Next Steps

1. **Discovery**: Check existing code (30% complete)
2. **Implementation**: Complete missing 70% (email service, rate limiting, handlers)
3. **Testing**: Comprehensive test coverage
4. **Documentation**: Update spec.md, create completion summary

---

*Email OTP Implementation Plan Version: 1.0.0*
*Author: GitHub Copilot (Agent)*
*Created: 2025-01-09*
