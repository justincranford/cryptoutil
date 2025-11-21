# Task 4b: Client Auth Basic Methods

**Status:** status:pending
**Estimated Time:** 10 minutes
**Priority:** High (Basic client authentication)

## 🎯 GOAL

Implement basic client authentication methods for OAuth 2.1: `client_secret_basic` and `client_secret_post`. These are the most commonly used client authentication methods in web and mobile applications.

## 📋 TASK OVERVIEW

Add support for client authentication using HTTP Basic authentication (`client_secret_basic`) and form-encoded client credentials (`client_secret_post`). These methods allow clients to authenticate using their client ID and secret.

## 🔧 INPUTS & CONTEXT

**Location:** `/internal/identity/authz/clientauth/`

**Dependencies:** Task 4 (OAuth 2.1 server core), client profile management

**Methods to Implement:**

- `client_secret_basic`: HTTP Basic authentication with client_id:client_secret
- `client_secret_post`: Client credentials in POST body parameters

**Security:** Client secrets must be validated against stored hashes, rate limiting applied

## 📁 FILES TO MODIFY/CREATE

### 1. Client Authentication Framework (`/internal/identity/authz/clientauth/`)

```text
clientauth/
├── interface.go             # ClientAuth interface (extend existing)
├── basic.go                 # client_secret_basic implementation
├── post.go                  # client_secret_post implementation
└── registry.go              # Authentication method registry
```

### 2. Integration Points

**Modify `/internal/identity/authz/handlers.go`:**

- Add client authentication middleware
- Integrate basic auth methods into token endpoint

**Modify `/internal/identity/authz/client_profiles.go`:**

- Add support for configuring allowed auth methods per client

## 🔄 IMPLEMENTATION STEPS

### Step 1: Extend ClientAuth Interface

```go
type ClientAuthenticator interface {
    Method() string
    Authenticate(ctx *fiber.Ctx) (*ClientProfile, error)
}
```

### Step 2: Implement Basic Authentication

```go
type BasicAuthenticator struct{}

func (b *BasicAuthenticator) Method() string {
    return "client_secret_basic"
}

func (b *BasicAuthenticator) Authenticate(ctx *fiber.Ctx) (*ClientProfile, error) {
    // Parse HTTP Basic auth header
    // Validate client_id:client_secret
    // Return client profile or error
}
```

### Step 3: Implement POST Authentication

```go
type PostAuthenticator struct{}

func (p *PostAuthenticator) Method() string {
    return "client_secret_post"
}

func (p *PostAuthenticator) Authenticate(ctx *fiber.Ctx) (*ClientProfile, error) {
    // Parse client_id and client_secret from form
    // Validate credentials
    // Return client profile or error
}
```

### Step 4: Register Auth Methods

```go
var authenticators = map[string]ClientAuthenticator{
    "client_secret_basic": &BasicAuthenticator{},
    "client_secret_post":  &PostAuthenticator{},
}
```

## ✅ ACCEPTANCE CRITERIA

- ✅ `client_secret_basic` method works with HTTP Basic auth
- ✅ `client_secret_post` method works with form parameters
- ✅ Invalid credentials properly rejected
- ✅ Rate limiting prevents brute force attacks
- ✅ Client profiles correctly returned on successful auth
- ✅ Integration with Task 4 token endpoint
- ✅ Unit tests with 95%+ coverage
- ✅ Documentation updated

## 🧪 TESTING REQUIREMENTS

### Unit Tests

- Valid basic auth header parsing
- Valid POST parameter parsing
- Invalid credentials handling
- Rate limiting behavior
- Client profile retrieval

### Integration Tests

- End-to-end token request with basic auth
- End-to-end token request with POST auth
- Error responses for invalid auth

## 📚 REFERENCES

- [RFC 6749 Section 2.3.1](https://tools.ietf.org/html/rfc6749#section-2.3.1) - Client Secret Basic
- [RFC 6749 Section 2.3.2](https://tools.ietf.org/html/rfc6749#section-2.3.2) - Client Secret POST
- [OAuth 2.1 Security BCP](https://tools.ietf.org/html/draft-ietf-oauth-security-topics-13)
