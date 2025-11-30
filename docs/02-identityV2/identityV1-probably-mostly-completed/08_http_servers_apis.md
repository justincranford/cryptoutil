# Task 6: HTTP Servers, APIs & Command-Line Applications - Fiber Servers, APIs & CLI Apps

**Status:** status:pending
**Estimated Time:** 35 minutes
**Priority:** High (HTTP server orchestration)

🎯 GOAL

Implement the three independent HTTP servers (AuthZ, IdP, RS) with CLI clients, admin APIs, command-line applications, and proper server lifecycle management. This provides the complete HTTP interface layer including headless client for automated testing and SPA RP application for manual demonstration.

📋 TASK OVERVIEW

Create three independent Fiber HTTP servers for AuthZ, IdP, and RS services. Implement CLI and agent clients for each server, plus admin APIs for management operations. Build command-line applications for all services including headless client for CI/CD testing and SPA RP application for manual testing.

🔧 INPUTS & CONTEXT

**Location:** `/internal/identity/server/`

**Dependencies:** Fiber HTTP server, all previous services

**Servers:** AuthZ (OAuth), IdP (OIDC), RS (Resource Server)

**Constraints:** Independent startup, maximum decoupling

📁 FILES TO CREATE

### 1. Server Orchestration (`/internal/identity/server/`)

``text
├── authz_server.go         # OAuth 2.1 authorization server
├── idp_server.go           # OIDC identity provider server
├── rs_server.go            # Resource server
└── server_manager.go        # Server lifecycle management

``

### 2. CLI Clients (`/internal/identity/client/`)

``text
├── authz_client.go          # AuthZ server CLI client
├── idp_client.go            # IdP server CLI client
├── rs_client.go             # Resource server CLI client
└── client_factory.go        # Client factory with configuration

``

### 3. Admin APIs (`/internal/identity/admin/`)

``text
├── authz_admin.go           # AuthZ admin API endpoints
├── idp_admin.go             # IdP admin API endpoints
├── rs_admin.go              # Resource server admin API
├── client_profile_admin.go  # Client profile management APIs
├── auth_flow_admin.go       # Authorization flow management APIs
├── auth_profile_admin.go    # Authentication profile management APIs
└── admin_handlers.go        # Common admin functionality

``

### 4. Command-Line Applications (`./cmd/identity/`)

``text
├── authz/
│   └── main.go              # OAuth 2.1 authorization server CLI
├── idp/
│   └── main.go              # OIDC identity provider server CLI
├── rs/
│   └── main.go              # Resource server CLI
├── headless-client/
│   └── main.go              # Headless client for testing flows
└── spa-rp/
    └── main.go              # SPA relying party application CLI

``

🎯 IMPLEMENTATION REQUIREMENTS

### HTTP Servers

**AuthZ Server:** OAuth 2.1 endpoints on dedicated port

**IdP Server:** OIDC endpoints on dedicated port

**RS Server:** Protected resource endpoints on dedicated port

**Lifecycle:** Graceful startup/shutdown, health checks

### CLI Clients

**Command Line:** Full CLI interface for each server

**Agent Mode:** Background operation capabilities

**Configuration:** YAML-based client configuration

**Error Handling:** Proper CLI error reporting

### Admin APIs

**CRUD Operations:** Model management for each service

**Monitoring:** Health checks, metrics, diagnostics

**Security:** Protected admin endpoints

**HTTPS:** Admin APIs on separate secure port

**Profile Management:** APIs for creating and managing client profiles, authorization flows, and authentication profiles

### Command-Line Applications

**Server Applications:** Independent CLI apps for each service

**Headless Client:** Automated testing client for CI/CD workflows

**SPA RP Application:** Command-line SPA relying party for manual testing

**Configuration:** YAML-based configuration for all applications

## ✅ COMPLETION CRITERIA

### File Structure Requirements

- ✅ Independent server startup for all 3 services
- ✅ HTTP/HTTPS support with configurable TLS
- ✅ Working CLI clients and agent clients
- ✅ Admin APIs for model CRUD
- ✅ Client profile and authorization flow management APIs
- ✅ Authentication profile management APIs
- ✅ Command-line applications for all services
- ✅ Headless client for automated testing
- ✅ SPA RP application for manual testing
- ✅ OpenAPI specs generated
- ✅ All servers start successfully

### Code Quality

[ ] All servers and clients compile without errors

[ ] Follow cryptoutil HTTP patterns exactly

[ ] Proper Go documentation comments

[ ] No linting errors (`golangci-lint run`)

### Server Architecture

[ ] Three independent HTTP servers

[ ] Proper middleware and routing

[ ] Health check endpoints

[ ] Graceful shutdown handling

### Client Interfaces

[ ] Full CLI clients for each server

[ ] Agent mode capabilities

[ ] Configuration management

[ ] Error handling and reporting

### Admin Functionality

[ ] CRUD admin APIs for all services

[ ] Secure admin endpoints

[ ] Monitoring and diagnostics

[ ] HTTPS admin interfaces

[ ] OpenAPI specs generated

[ ] Client profile management (create, update, delete profiles with MFA chains)

[ ] Client authentication method management (client_secret_jwt, private_key_jwt, mTLS, bearer tokens)

[ ] Authorization flow configuration management

[ ] Authentication profile management (user auth methods with MFA chains and mTLS domains)

[ ] MFA factor configuration management (TOTP/HOTP setup, factor ordering, QR codes)

[ ] User mTLS domain management (client certificate domains per authentication profile)

### Testing Requirements

- [ ] Parameterized unit tests for all HTTP handlers and middleware
- [ ] Error path testing for invalid requests, authentication failures, and network issues
- [ ] Edge case coverage for CLI argument parsing, agent communication, and API validation
- [ ] Table-driven tests for different server configurations and client modes
- [ ] Command-line application testing for all ./cmd/identity/ applications
- [ ] Headless client integration tests for automated authorization flows
- [ ] SPA RP application testing for manual testing scenarios
- [ ] 95%+ code coverage for HTTP servers, CLI clients, admin APIs, and command-line applications

### OAuth 2.1 & OIDC Compliance Testing

#### AuthZ Server Compliance

- [ ] **PKCE Enforcement**: Mandatory PKCE for all authorization code flows (OAuth 2.1 requirement)
- [ ] **State Parameter Validation**: Required state parameter validation and replay prevention
- [ ] **Redirect URI Strict Matching**: Exact redirect URI matching with no wildcards
- [ ] **Refresh Token Rotation**: Automatic refresh token rotation on use
- [ ] **Confidential Clients Only**: Reject public client requests
- [ ] **JWT Client Authentication**: Support for private_key_jwt client authentication method

#### IdP Server Compliance

- [ ] **OIDC Core 1.0 Compliance**: Complete OpenID Connect Core specification coverage
- [ ] **ID Token Validation**: Proper ID token issuance, signing, and validation
- [ ] **UserInfo Endpoint**: Protected user information endpoint with access token validation
- [ ] **Discovery Document**: `.well-known/openid-configuration` endpoint with complete metadata
- [ ] **JWKS Endpoint**: JSON Web Key Set endpoint for token validation
- [ ] **Session Management**: OIDC session management and logout functionality

#### RS Server Compliance

- [ ] **JWT Access Token Validation**: Support for JWT access tokens with signature verification
- [ ] **Token Introspection**: RFC 7662 token introspection endpoint
- [ ] **Scope Validation**: Proper scope-based access control
- [ ] **Audience Validation**: Token audience claims validation
- [ ] **Token Revocation**: Support for token revocation and blacklisting

#### Command-Line Application Testing

- [ ] **oauth2c Integration**: Automated OAuth 2.0 compliance testing in CI/CD
- [ ] **OIDC Conformance Suite**: Official OIDC compliance certification testing
- [ ] **Security Scanning**: OWASP ZAP OAuth add-on testing
- [ ] **Nuclei OAuth Templates**: Vulnerability scanning with OAuth-specific templates

## 🔗 NEXT STEPS

After completion:

1. **Commit:** `feat: complete Task 6 - HTTP servers, APIs and command-line applications`
2. **Update:** `identity_master.md` status to completed
3. **Begin:** Task 7 - SPA Relying Party

📝 NOTES

Focus on clean server separation

Implement comprehensive CLI interfaces

Design admin APIs for operations

Ensure proper HTTPS configuration

Build command-line applications for testing and demonstration
 
 
 
 
 
 
