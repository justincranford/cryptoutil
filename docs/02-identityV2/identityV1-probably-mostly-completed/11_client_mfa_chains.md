# Task 4d: Client MFA Chains

**Status:** status:pending
**Estimated Time:** 20 minutes
**Priority:** High (Client MFA chains for enhanced security)

## 🎯 GOAL

Implement client MFA (Multi-Factor Authentication) chains for OAuth 2.1: `private_key_jwt` and `client_secret_jwt`. These provide enhanced security by requiring multiple authentication factors for client applications.

## 📋 TASK OVERVIEW

Add support for chained client authentication methods where clients must provide multiple forms of authentication. This includes JWT-based authentication methods that can be combined with other factors for stronger security.

## 🔧 INPUTS & CONTEXT

**Location:** `/internal/identity/authz/clientauth/`

**Dependencies:** Task 4 (OAuth 2.1 server core), JWT infrastructure, Task 4b (basic client auth)

**Methods to Implement:**

- `private_key_jwt`: Client authentication using JWT signed with private key
- `client_secret_jwt`: Client authentication using JWT signed with client secret
- MFA chains: Support for combining multiple authentication methods

**Security:** JWT validation, signature verification, expiration checking, audience validation

## 📁 FILES TO MODIFY/CREATE

### 1. JWT Authentication Framework (`/internal/identity/authz/clientauth/`)

```text
clientauth/
├── interface.go              # ClientAuth interface (extend existing)
├── private_key_jwt.go        # private_key_jwt implementation
├── client_secret_jwt.go      # client_secret_jwt implementation
├── jwt_validator.go          # JWT validation utilities
└── mfa_chain.go             # MFA chain orchestration
```

### 2. Integration Points

**Modify `/internal/identity/authz/handlers.go`:**

- Add JWT parameter extraction from request
- Integrate JWT auth methods into token endpoint
- Support MFA chain validation

**Modify `/internal/identity/authz/client_profiles.go`:**

- Add JWT-based client identification
- Support MFA chain configuration per client

## 🔄 IMPLEMENTATION STEPS

### Step 1: JWT Validation Framework

```go
type JWTValidator interface {
    ValidateJWT(jwtString string, clientID string) (*jwt.Token, error)
    ExtractClaims(token *jwt.Token) (*ClientClaims, error)
}

type PrivateKeyJWTValidator struct {
    keyStore KeyStore
}

type ClientSecretJWTValidator struct {
    secretStore SecretStore
}
```

### Step 2: Implement Private Key JWT Auth

```go
type PrivateKeyJWTAuthenticator struct {
    validator JWTValidator
}

func (p *PrivateKeyJWTAuthenticator) Method() string {
    return "private_key_jwt"
}

func (p *PrivateKeyJWTAuthenticator) Authenticate(ctx *fiber.Ctx) (*ClientProfile, error) {
    // Extract client_assertion parameter
    // Validate JWT signature with client's public key
    // Verify claims (iss, sub, aud, exp, etc.)
    // Map client ID to profile
    // Return client profile or error
}
```

### Step 3: Implement Client Secret JWT Auth

```go
type ClientSecretJWTAuthenticator struct {
    validator JWTValidator
}

func (c *ClientSecretJWTAuthenticator) Method() string {
    return "client_secret_jwt"
}

func (c *ClientSecretJWTAuthenticator) Authenticate(ctx *fiber.Ctx) (*ClientProfile, error) {
    // Extract client_assertion parameter
    // Validate JWT signature with client's secret
    // Verify claims (iss, sub, aud, exp, etc.)
    // Map client ID to profile
    // Return client profile or error
}
```

### Step 4: Implement MFA Chains

```go
type MFAAuthenticator struct {
    authenticators []ClientAuthenticator
}

func (m *MFAAuthenticator) Authenticate(ctx *fiber.Ctx) (*ClientProfile, error) {
    // Execute all authenticators in chain
    // Require all methods to succeed
    // Aggregate results and return profile
    // Fail fast on first authentication failure
}
```

### Step 5: Register Auth Methods

```go
var authenticators = map[string]ClientAuthenticator{
    "private_key_jwt":    &PrivateKeyJWTAuthenticator{validator: &PrivateKeyJWTValidator{}},
    "client_secret_jwt":  &ClientSecretJWTAuthenticator{validator: &ClientSecretJWTValidator{}},
    "mfa_chain":         &MFAAuthenticator{authenticators: []ClientAuthenticator{...}},
}
```

## ✅ ACCEPTANCE CRITERIA

- ✅ `private_key_jwt` method works with RSA/ECDSA signed JWTs
- ✅ `client_secret_jwt` method works with HMAC signed JWTs
- ✅ JWT validation includes signature, expiration, and audience verification
- ✅ MFA chains support combining multiple authentication methods
- ✅ Invalid JWTs properly rejected with appropriate error responses
- ✅ Client profile mapping works correctly
- ✅ Integration with Task 4 token endpoint
- ✅ Unit tests with 95%+ coverage
- ✅ Documentation updated

## 🧪 TESTING REQUIREMENTS

### Unit Tests

- Valid private key JWT authentication
- Valid client secret JWT authentication
- Invalid JWT rejection (bad signature, expired, wrong audience)
- MFA chain validation (all methods required)
- Partial MFA chain failure handling
- Client profile mapping from JWT claims

### Integration Tests

- End-to-end token request with private key JWT
- End-to-end token request with client secret JWT
- End-to-end token request with MFA chain
- JWT validation error responses

## 📚 REFERENCES

- [RFC 7523](https://tools.ietf.org/html/rfc7523) - JSON Web Token (JWT) Profile for OAuth 2.0 Client Authentication
- [RFC 7521](https://tools.ietf.org/html/rfc7521) - Assertion Framework for OAuth 2.0 Client Authentication
- [RFC 8725](https://tools.ietf.org/html/rfc8725) - JSON Web Token Best Current Practices
