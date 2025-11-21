# Task 5d: Biometric Authentication

**Status:** status:pending
**Estimated Time:** 35 minutes
**Priority:** High (Biometric and hardware-based authentication)

## 🎯 GOAL

Implement biometric authentication methods for OIDC: Passkey/WebAuthn, TOTP/HOTP, and hardware security keys. These provide strong, phishing-resistant authentication using biometric factors and hardware tokens.

## 📋 TASK OVERVIEW

Add support for modern biometric and hardware-based authentication methods. This includes WebAuthn/FIDO2 passkeys for passwordless authentication, time-based and HMAC-based OTP tokens, and hardware security keys for maximum security.

## 🔧 INPUTS & CONTEXT

**Location:** `/internal/identity/idp/userauth/`

**Dependencies:** Task 5 (OIDC Identity Provider core), WebAuthn/FIDO2 libraries, hardware token support

**Methods to Implement:**

- `passkey_webauthn`: FIDO2/WebAuthn passkey authentication
- `totp_hotp`: Time-based and HMAC-based one-time passwords
- Hardware security keys (YubiKey, etc.)
- Biometric factors (fingerprint, face, etc.)

**Security:** Phishing resistance, hardware-backed keys, biometric verification, replay attack prevention

## 📁 FILES TO MODIFY/CREATE

### 1. Biometric Authentication Framework (`/internal/identity/idp/userauth/`)

```text
userauth/
├── interface.go              # UserAuth interface (extend existing)
├── webauthn_auth.go         # WebAuthn/FIDO2 implementation
├── totp_hotp_auth.go        # TOTP/HOTP implementation
├── hardware_key_auth.go     # Hardware security key support
├── biometric_verifier.go    # Biometric verification utilities
└── token_manager.go         # Hardware token management
```

### 2. Integration Points

**Modify `/internal/identity/idp/handlers.go`:**

- Add WebAuthn registration/challenge endpoints
- Add TOTP/HOTP setup and verification
- Integrate hardware token authentication

**Modify `/internal/identity/idp/user_profiles.go`:**

- Add biometric credential storage
- Support hardware token registration
- Manage TOTP/HOTP secrets securely

## 🔄 IMPLEMENTATION STEPS

### Step 1: WebAuthn Framework

```go
type WebAuthnAuthenticator struct {
    webauthn *webauthn.WebAuthn
    store     CredentialStore
}

func (w *WebAuthnAuthenticator) Method() string {
    return "passkey_webauthn"
}

func (w *WebAuthnAuthenticator) BeginRegistration(user *User) (*protocol.CredentialCreation, error) {
    // Begin WebAuthn credential registration
    // Generate challenge and options
    // Return credential creation data
}

func (w *WebAuthnAuthenticator) FinishRegistration(user *User, response *protocol.ParsedCredentialCreationData) (*webauthn.Credential, error) {
    // Verify registration response
    // Store credential securely
    // Return verified credential
}

func (w *WebAuthnAuthenticator) BeginLogin(user *User) (*protocol.CredentialAssertion, error) {
    // Begin WebAuthn authentication
    // Generate assertion challenge
    // Return credential assertion data
}

func (w *WebAuthnAuthenticator) FinishLogin(user *User, response *protocol.ParsedCredentialAssertionData) (*webauthn.Credential, error) {
    // Verify authentication response
    // Check credential validity
    // Return authenticated credential
}
```

### Step 2: TOTP/HOTP Implementation

```go
type TOTPAuthenticator struct {
    issuer string
    digits int
}

func (t *TOTPAuthenticator) Method() string {
    return "totp"
}

func (t *TOTPAuthenticator) GenerateSecret() (string, error) {
    // Generate random secret
    // Return base32 encoded secret
}

func (t *TOTPAuthenticator) GenerateTOTP(secret string) (string, error) {
    // Generate time-based OTP
    // Return 6-digit code
}

func (t *TOTPAuthenticator) ValidateTOTP(secret, code string) bool {
    // Validate TOTP code with time window
    // Allow for clock skew
    // Return validation result
}

type HOTPAuthenticator struct {
    counter uint64
}

func (h *HOTPAuthenticator) Method() string {
    return "hotp"
}

func (h *HOTPAuthenticator) GenerateHOTP(secret string) (string, error) {
    // Generate HMAC-based OTP
    // Increment counter
    // Return 6-digit code
}

func (h *HOTPAuthenticator) ValidateHOTP(secret, code string) bool {
    // Validate HOTP code
    // Check counter values
    // Prevent replay attacks
    // Return validation result
}
```

### Step 3: Hardware Token Support

```go
type HardwareKeyAuthenticator struct {
    supportedKeys []string // YubiKey, Titan, etc.
}

func (h *HardwareKeyAuthenticator) Method() string {
    return "hardware_key"
}

func (h *HardwareKeyAuthenticator) RegisterKey(userID string, keyInfo *HardwareKeyInfo) error {
    // Register hardware security key
    // Store key metadata
    // Generate attestation certificate
}

func (h *HardwareKeyAuthenticator) AuthenticateWithKey(ctx *fiber.Ctx, challenge string) (*AuthResult, error) {
    // Authenticate using hardware key
    // Verify key signature
    // Check key validity
    // Return authentication result
}
```

### Step 4: Biometric Verification

```go
type BiometricVerifier struct {
    supportedTypes []BiometricType
}

func (b *BiometricVerifier) VerifyBiometric(biometricData *BiometricData, storedTemplate *BiometricTemplate) bool {
    // Verify biometric match
    // Calculate confidence score
    // Apply liveness detection
    // Return verification result
}

func (b *BiometricVerifier) EnrollBiometric(userID string, biometricData *BiometricData) (*BiometricTemplate, error) {
    // Enroll biometric template
    // Generate secure template
    // Store encrypted template
    // Return enrollment result
}
```

### Step 5: Register Auth Methods

```go
var authenticators = map[string]UserAuthenticator{
    "passkey_webauthn": &WebAuthnAuthenticator{webauthn: webAuthnInstance, store: &CredentialStore{}},
    "totp":            &TOTPAuthenticator{issuer: "cryptoutil", digits: 6},
    "hotp":            &HOTPAuthenticator{counter: 0},
    "hardware_key":    &HardwareKeyAuthenticator{supportedKeys: []string{"yubikey", "titan"}},
}
```

## ✅ ACCEPTANCE CRITERIA

- ✅ WebAuthn passkey registration and authentication works
- ✅ TOTP generates and validates 6-digit codes correctly
- ✅ HOTP generates and validates codes with proper counter management
- ✅ Hardware security keys integrate with WebAuthn/FIDO2
- ✅ Biometric verification supports fingerprint and facial recognition
- ✅ Phishing-resistant authentication prevents credential theft
- ✅ Secure credential storage with encryption
- ✅ Integration with OIDC authentication flows
- ✅ Unit tests with 95%+ coverage
- ✅ Documentation updated

## 🧪 TESTING REQUIREMENTS

### Unit Tests

- WebAuthn credential registration and verification
- TOTP code generation and validation
- HOTP code generation and counter management
- Hardware key registration and authentication
- Biometric template enrollment and verification
- Credential storage and retrieval

### Integration Tests

- End-to-end WebAuthn authentication flow
- End-to-end TOTP setup and authentication
- End-to-end HOTP token authentication
- Hardware security key integration
- Biometric authentication scenarios

## 📚 REFERENCES

- [WebAuthn Specification](https://www.w3.org/TR/webauthn-2/) - Web Authentication API
- [FIDO2 Specifications](https://fidoalliance.org/specifications/) - FIDO2 Technical Specifications
- [RFC 6238](https://tools.ietf.org/html/rfc6238) - TOTP: Time-Based One-Time Password Algorithm
- [RFC 4226](https://tools.ietf.org/html/rfc4226) - HOTP: An HMAC-Based One-Time Password Algorithm
