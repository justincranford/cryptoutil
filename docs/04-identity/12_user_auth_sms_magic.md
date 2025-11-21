# Task 5b: SMS OTP and Magic Links

**Status:** status:pending
**Estimated Time:** 25 minutes
**Priority:** High (Passwordless user authentication methods)

## 🎯 GOAL

Implement passwordless user authentication methods for OIDC: SMS OTP and Magic Links. These provide secure, user-friendly alternatives to traditional password-based authentication.

## 📋 TASK OVERVIEW

Add support for SMS-based one-time passwords and email-based magic links as primary authentication methods. These methods eliminate password management while providing strong security through possession factors.

## 🔧 INPUTS & CONTEXT

**Location:** `/internal/identity/idp/userauth/`

**Dependencies:** Task 5 (OIDC Identity Provider core), SMS/email infrastructure, OTP generation

**Methods to Implement:**

- `sms_otp`: User authentication via SMS-delivered one-time passwords
- `magic_link`: User authentication via email-delivered magic links
- OTP generation and validation
- Rate limiting and security controls

**Security:** OTP expiration, rate limiting, secure delivery channels, replay attack prevention

## 📁 FILES TO MODIFY/CREATE

### 1. Passwordless Authentication Framework (`/internal/identity/idp/userauth/`)

```text
userauth/
├── interface.go              # UserAuth interface (extend existing)
├── sms_otp.go               # SMS OTP implementation
├── magic_link.go            # Magic link implementation
├── otp_generator.go         # OTP generation utilities
├── delivery_service.go      # SMS/Email delivery abstraction
└── rate_limiter.go          # Rate limiting for OTP requests
```

### 2. Integration Points

**Modify `/internal/identity/idp/handlers.go`:**

- Add OTP request endpoints
- Add OTP verification endpoints
- Integrate with OIDC authentication flows

**Modify `/internal/identity/idp/user_profiles.go`:**

- Add phone number and email verification
- Support user preference for auth methods

## 🔄 IMPLEMENTATION STEPS

### Step 1: OTP Generation Framework

```go
type OTPGenerator interface {
    GenerateOTP(length int) (string, error)
    GenerateSecureToken() (string, error)
}

type TOTPGenerator struct {
    // Time-based OTP generation
}

type HOTPGenerator struct {
    // HMAC-based OTP generation
}
```

### Step 2: Delivery Service Abstraction

```go
type DeliveryService interface {
    SendSMS(phoneNumber, message string) error
    SendEmail(email, subject, body string) error
}

type TwilioSMSService struct {
    client *twilio.Client
}

type SendGridEmailService struct {
    client *sendgrid.Client
}
```

### Step 3: Implement SMS OTP Auth

```go
type SMSOTPAuthenticator struct {
    generator OTPGenerator
    delivery  DeliveryService
    store     OTPStore
}

func (s *SMSOTPAuthenticator) Method() string {
    return "sms_otp"
}

func (s *SMSOTPAuthenticator) InitiateAuth(ctx *fiber.Ctx, userID string) (*AuthChallenge, error) {
    // Generate OTP
    // Send via SMS to user's phone
    // Store OTP with expiration
    // Return challenge ID
}

func (s *SMSOTPAuthenticator) VerifyAuth(ctx *fiber.Ctx, challengeID, otp string) (*UserProfile, error) {
    // Validate OTP against stored value
    // Check expiration and rate limits
    // Return user profile on success
}
```

### Step 4: Implement Magic Link Auth

```go
type MagicLinkAuthenticator struct {
    generator OTPGenerator
    delivery  DeliveryService
    store     TokenStore
}

func (m *MagicLinkAuthenticator) Method() string {
    return "magic_link"
}

func (m *MagicLinkAuthenticator) InitiateAuth(ctx *fiber.Ctx, userID string) (*AuthChallenge, error) {
    // Generate secure token
    // Create magic link URL
    // Send via email
    // Store token with expiration
    // Return challenge ID
}

func (m *MagicLinkAuthenticator) VerifyAuth(ctx *fiber.Ctx, token string) (*UserProfile, error) {
    // Validate token from link click
    // Check expiration and single-use
    // Return user profile on success
}
```

### Step 5: Rate Limiting and Security

```go
type RateLimiter struct {
    attempts map[string]*AttemptRecord
}

func (r *RateLimiter) CheckRateLimit(identifier string) error {
    // Implement exponential backoff
    // Track failed attempts
    // Block after threshold
}
```

### Step 6: Register Auth Methods

```go
var authenticators = map[string]UserAuthenticator{
    "sms_otp":    &SMSOTPAuthenticator{generator: &TOTPGenerator{}, delivery: &TwilioSMSService{}},
    "magic_link": &MagicLinkAuthenticator{generator: &SecureTokenGenerator{}, delivery: &SendGridEmailService{}},
}
```

## ✅ ACCEPTANCE CRITERIA

- ✅ SMS OTP method works with phone number delivery
- ✅ Magic link method works with email delivery
- ✅ OTPs expire after configurable time (default 5 minutes)
- ✅ Rate limiting prevents brute force attacks
- ✅ Magic links are single-use and expire
- ✅ Secure token generation for links
- ✅ Integration with OIDC authentication flows
- ✅ User preference for auth methods
- ✅ Unit tests with 95%+ coverage
- ✅ Documentation updated

## 🧪 TESTING REQUIREMENTS

### Unit Tests

- Valid SMS OTP authentication flow
- Valid magic link authentication flow
- OTP expiration handling
- Rate limiting enforcement
- Invalid OTP/link rejection
- Single-use link validation
- User profile mapping

### Integration Tests

- End-to-end SMS OTP authentication
- End-to-end magic link authentication
- Rate limiting behavior
- OTP expiration scenarios
- Multiple delivery service providers

## 📚 REFERENCES

- [RFC 6238](https://tools.ietf.org/html/rfc6238) - TOTP: Time-Based One-Time Password Algorithm
- [RFC 4226](https://tools.ietf.org/html/rfc4226) - HOTP: An HMAC-Based One-Time Password Algorithm
- [RFC 7616](https://tools.ietf.org/html/rfc7616) - HTTP Digest Access Authentication
