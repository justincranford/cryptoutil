# Task 4c: mTLS Client Auth

**Status:** status:pending
**Estimated Time:** 15 minutes
**Priority:** High (Mutual TLS client authentication)

## 🎯 GOAL

Implement mutual TLS (mTLS) client authentication methods for OAuth 2.1: `tls_client_auth` and `self_signed_tls_client_auth`. These provide strong cryptographic client authentication using X.509 certificates.

## 📋 TASK OVERVIEW

Add support for client authentication using TLS client certificates. This includes both CA-issued certificates (`tls_client_auth`) and self-signed certificates (`self_signed_tls_client_auth`) for development and testing scenarios.

## 🔧 INPUTS & CONTEXT

**Location:** `/internal/identity/authz/clientauth/`

**Dependencies:** Task 4 (OAuth 2.1 server core), certificate validation infrastructure

**Methods to Implement:**

- `tls_client_auth`: Client authentication using CA-issued X.509 certificates
- `self_signed_tls_client_auth`: Client authentication using self-signed certificates (development/testing)

**Security:** Certificate validation, revocation checking, certificate pinning support

## 📁 FILES TO MODIFY/CREATE

### 1. Client Authentication Framework (`/internal/identity/authz/clientauth/`)

```text
clientauth/
├── interface.go             # ClientAuth interface (extend existing)
├── tls_client_auth.go       # tls_client_auth implementation
├── self_signed_auth.go      # self_signed_tls_client_auth implementation
└── certificate_validator.go # Certificate validation utilities
```

### 2. Integration Points

**Modify `/internal/identity/authz/handlers.go`:**

- Add TLS client certificate extraction from request context
- Integrate mTLS auth methods into token endpoint

**Modify `/internal/identity/authz/client_profiles.go`:**

- Add certificate-based client identification
- Support certificate pinning and validation rules

## 🔄 IMPLEMENTATION STEPS

### Step 1: Certificate Validation Framework

```go
type CertificateValidator interface {
    ValidateCertificate(clientCert *x509.Certificate, rawCerts [][]byte) error
    IsRevoked(serialNumber *big.Int) bool
}

type CACertificateValidator struct {
    // CA certificate validation
}

type SelfSignedValidator struct {
    // Self-signed certificate validation (pinned certificates)
}
```

### Step 2: Implement TLS Client Auth

```go
type TLSClientAuthenticator struct {
    validator CertificateValidator
}

func (t *TLSClientAuthenticator) Method() string {
    return "tls_client_auth"
}

func (t *TLSClientAuthenticator) Authenticate(ctx *fiber.Ctx) (*ClientProfile, error) {
    // Extract client certificate from TLS connection
    // Validate certificate chain and revocation
    // Map certificate to client profile
    // Return client profile or error
}
```

### Step 3: Implement Self-Signed Auth

```go
type SelfSignedAuthenticator struct {
    pinnedCertificates map[string]*x509.Certificate
}

func (s *SelfSignedAuthenticator) Method() string {
    return "self_signed_tls_client_auth"
}

func (s *SelfSignedAuthenticator) Authenticate(ctx *fiber.Ctx) (*ClientProfile, error) {
    // Extract client certificate from TLS connection
    // Validate against pinned self-signed certificates
    // Map certificate fingerprint to client profile
    // Return client profile or error
}
```

### Step 4: Register Auth Methods

```go
var authenticators = map[string]ClientAuthenticator{
    "tls_client_auth":             &TLSClientAuthenticator{validator: &CACertificateValidator{}},
    "self_signed_tls_client_auth": &SelfSignedAuthenticator{pinnedCerts: loadPinnedCerts()},
}
```

## ✅ ACCEPTANCE CRITERIA

- ✅ `tls_client_auth` method works with CA-issued certificates
- ✅ `self_signed_tls_client_auth` method works with pinned self-signed certificates
- ✅ Certificate validation includes revocation checking
- ✅ Invalid/revoked certificates properly rejected
- ✅ Certificate-to-client mapping works correctly
- ✅ Integration with Task 4 token endpoint
- ✅ Unit tests with 95%+ coverage
- ✅ Documentation updated

## 🧪 TESTING REQUIREMENTS

### Unit Tests

- Valid CA-issued certificate authentication
- Valid self-signed certificate authentication
- Invalid certificate rejection
- Revoked certificate handling
- Certificate pinning validation
- Client profile mapping

### Integration Tests

- End-to-end token request with TLS client auth
- End-to-end token request with self-signed auth
- Certificate validation error responses

## 📚 REFERENCES

- [RFC 8705](https://tools.ietf.org/html/rfc8705) - OAuth 2.0 Mutual-TLS Client Authentication
- [RFC 5280](https://tools.ietf.org/html/rfc5280) - Internet X.509 Public Key Infrastructure Certificate
- [RFC 6125](https://tools.ietf.org/html/rfc6125) - Representation and Verification of Domain-Based Application Service Identity
