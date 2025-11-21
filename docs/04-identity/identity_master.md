# IDENTITY MODULE - MASTER COORDINATION FILE
# OAuth 2.1 + OpenID Connect Identity Provider
# Multi-File Session Management

## 🎯 SESSION OVERVIEW

This identity module implementation is structured across multiple focused files for optimal Copilot Chat sessions.

**Total Tasks**: 15 (8 core + 7 advanced auth tasks)
**Estimated Timeline**: ~4-5 hours total
**Architecture**: OAuth 2.1 AuthZ Server + OIDC IdP Server + Resource Server

**✅ COMPLETED**: All granular authentication task documentation files have been created and are ready for implementation. Advanced authentication methods (Tasks 9-13) are scheduled after integration testing for phased rollout.

## 📋 SESSION WORKFLOW

1. **Start with Task 1** (`01_foundation_setup.md`)
2. **Complete each task** using its dedicated file
3. **Commit after each task** with semantic message
4. **Reference this master file** for overall coordination
5. **Use regrouping points** to plan next steps

**PHASED IMPLEMENTATION**: Core functionality (Tasks 1-10) implemented first, followed by advanced authentication methods (Tasks 11-15) after integration testing validation.

## 📁 FILE STRUCTURE

```text
workflow-reports/identity/
├── identity_master.md              # This coordination file
├── 01_foundation_setup.md          # Task 1: Domain models & config
├── 02_storage_interfaces.md        # Task 2: Database abstractions
├── 03_token_operations.md          # Task 3: JWT issuance/validation
├── 04_authz_server_core.md        # Task 4: OAuth 2.1 + initial client auth
├── 05_client_auth_basic.md         # Task 5: Basic client auth methods
├── 06_client_auth_mtls.md          # Task 6: mTLS client auth
├── 07_oidc_identity_provider.md    # Task 7: OIDC IdP + initial user auth
├── 08_http_servers_apis.md         # Task 8: HTTP servers & APIs
├── 09_spa_relying_party.md         # Task 9: SPA relying party application
├── 10_integration_testing.md       # Task 10: E2E & integration tests
├── 11_client_mfa_chains.md         # Task 11: Client MFA chains
├── 12_user_auth_sms_magic.md       # Task 12: Magic Links & SMS OTP
├── 13_user_auth_adaptive.md        # Task 13: Step-Up & Risk-Based Auth
├── 14_user_auth_biometric.md       # Task 14: Biometric Authentication
├── 15_user_auth_hardware.md        # Task 15: Hardware Security Keys
```

## 🚀 TASK STATUS OVERVIEW

| Task | File | Status | Time | Description |
|------|------|--------|------|-------------|
| 1 | `01_foundation_setup.md` | status:ready | 15 min | Domain models, errors, basic config |
| 2 | `02_storage_interfaces.md` | status:pending | 20 min | Database interfaces & implementations |
| 3 | `03_token_operations.md` | status:pending | 25 min | JWT token operations with cryptoutil |
| 4 | `04_authz_server_core.md` | status:pending | 15 min | OAuth 2.1 server + initial client auth (private_key_jwt, client_secret_jwt, bearer_token) |
| 5 | `05_client_auth_basic.md` | status:ready | 10 min | Basic client auth methods (client_secret_basic, client_secret_post) |
| 6 | `06_client_auth_mtls.md` | status:ready | 15 min | mTLS client auth (tls_client_auth, self_signed_tls_client_auth) |
| 7 | `07_oidc_identity_provider.md` | status:pending | 20 min | OIDC IdP + initial user auth (Passkey, Email+OTP, TOTP, Username/Password) |
| 8 | `08_http_servers_apis.md` | status:pending | 35 min | HTTP servers, CLI clients, admin APIs |
| 9 | `09_spa_relying_party.md` | status:pending | 20 min | SPA relying party application |
| 10 | `10_integration_testing.md` | status:pending | 40 min | Complete spec coverage testing |
| 11 | `11_client_mfa_chains.md` | status:ready | 20 min | Client MFA chains (combining 4/4b/4c methods) |
| 12 | `12_user_auth_sms_magic.md` | status:ready | 15 min | Magic Links and SMS OTP |
| 13 | `13_user_auth_adaptive.md` | status:ready | 20 min | Step-Up and Risk-Based Authentication |
| 14 | `14_user_auth_biometric.md` | status:ready | 15 min | Biometric Authentication |
| 15 | `15_user_auth_hardware.md` | status:ready | 15 min | Hardware Security Keys (FIDO U2F/U2F2) |

## 📋 GLOBAL CONSTRAINTS (Reference for all tasks)

### Dependencies

**ONLY use existing go.mod dependencies:**

- `lestrrat-go/jwx/v3` - JWT operations
- `google/uuid` - UUID generation
- `gofiber/fiber/v2` - HTTP server
- `gorm.io/gorm` - Database ORM
- `github.com/stretchr/testify` - Testing

### Code Reuse

**ONLY allowed from:**

- `internal/common/container`
- `internal/common/crypto`
- `internal/common/pool`
- `internal/common/telemetry`
- `internal/common/util`

### Magic Values

**ALL constants** → `/internal/identity/<package>/magic*.go` files (linting configuration updated to allow this)

### Import Aliases

**CRITICAL**: When creating identity packages, add corresponding import aliases to `.golangci.yml`:

```yaml
# Add to .golangci.yml importas section:
- pkg: cryptoutil/internal/identity/authz
  alias: cryptoutilIdentityAuthz
- pkg: cryptoutil/internal/identity/idp
  alias: cryptoutilIdentityIdp
- pkg: cryptoutil/internal/identity/rs
  alias: cryptoutilIdentityRs
- pkg: cryptoutil/internal/identity/common
  alias: cryptoutilIdentityCommon
```

### Architecture

- **AuthZ Server**: OAuth 2.1 authorization server
- **IdP Server**: OpenID Connect identity provider
- **RS Server**: Resource server with token validation
- **All services**: Independently deployable with HTTP/HTTPS

## 📋 IMPLEMENTATION NOTES

### Critical Constraints

- **DEPENDENCY CONSTRAINT**: Use ONLY existing go.mod dependencies - NO new 3rd-party libraries
- **CONFIGURATION**: All authz, idp, rs implementations must be fully parameterized with NO hardcoded defaults
- **TOKEN FORMATS**: Support JWS, JWE, and UUID access tokens using existing jwx and uuid libraries
- **OPENAPI**: Generate separate component-based specs for authz, idp, and rs services
- **SERVERS**: Each service (authz, idp, rs) must support independent startup as GoFiber servers with optional HTTP/HTTPS
- **MAGIC VALUES**: ALL constant magic values MUST go in `/internal/identity/<PACKAGE>/magic*.go` files
- **CLI/AGENT CLIENTS**: Working CLI clients and agent clients required for each of authz, idp, rs
- **ADMIN APIs**: Authz and idp must offer admin APIs for model CRUD
- **SPA RELYING PARTY**: Working SPA relying party is required
- **CLIENT PROFILES**: Dynamic client authorization flow profiles with custom required/optional scopes and configurable MFA chains
- **AUTH FLOWS**: Parameterized authorization code flows with PKCE, configurable consent screens (1 or 2)
- **CLIENT AUTH METHODS**: client_secret_basic, client_secret_post, client_secret_jwt, private_key_jwt, tls_client_auth, self_signed_tls_client_auth, bearer_token
- **IDP PROFILES**: Multiple authentication profiles (username/password, email+2FA, mobile+SMS, passkey) with configurable MFA factors (TOTP/HOTP)

### Authentication Method Implementation Phases

**PHASE 1 - CORE MVP (Implement First):**

- ✅ Username/Password authentication
- ✅ Email/Password + Email OTP (MFA)
- ✅ TOTP/HOTP with QR code setup
- ✅ Passkey/WebAuthn authentication
- ✅ Basic client authentication (client_secret_basic, client_secret_post)

**PHASE 2 - ADVANCED OAUTH (Implement Second):**

- 🔄 client_secret_jwt, private_key_jwt (secure client auth)
- 🔄 tls_client_auth (mTLS for clients)
- 🔄 bearer_token authentication
- 🔄 SMS OTP and Magic Links
- 🔄 Client MFA chains

**PHASE 3 - ENTERPRISE FEATURES (Implement Last):**

- 🔄 Hardware security keys (FIDO U2F/U2F2)
- 🔄 Biometric authentication
- 🔄 Risk-based authentication
- 🔄 Step-up authentication
- 🔄 mTLS for users with certificate domains
- 🔄 self_signed_tls_client_auth

### Implementation Guidance

- **OAUTH2/OIDC COVERAGE**: Complete OAuth 2.1 and OIDC specifications must be covered in design, implementation, and ESPECIALLY integration/E2E tests
- **SERVICE INITIALIZATION**: Study `ServerApplicationBasic` pattern for `TelemetryService` and `JWKGenService` initialization
- **DECOUPLING**: `/internal/identity/` must be maximally decoupled for future extraction into separate repository
- **TERMINOLOGY CLARIFICATION**:
  - **Client Profile**: OAuth client configuration (scopes, flows, redirect URIs)
  - **Authentication Profile**: User authentication methods (password, MFA, passkey)
  - **Authorization Flow**: OAuth grant types and PKCE configurations
  - **User Profile**: OIDC user claims and profile information
- **PHASED IMPLEMENTATION**: Authentication methods prioritized into phases (Core MVP → Advanced OAuth → Enterprise)

## 🎯 ACCEPTANCE CRITERIA SUMMARY

### Functional Requirements

- ✅ Complete OAuth 2.1 authorization code flow with PKCE
- ✅ Client credentials grant type, refresh token flow
- ✅ OIDC ID tokens and userinfo endpoint
- ✅ Token introspection and revocation
- ✅ Secure client registration and management
- ✅ User authentication and profile management
- ✅ Resource server with token validation
- ✅ Independent server startup for authz, idp, rs services
- ✅ Working CLI clients and agent clients for each service
- ✅ Admin APIs for model CRUD operations
- ✅ Working SPA relying party application
- ✅ Complete OAuth 2.1 and OIDC specification coverage
- ✅ **Parameterized client authorization flow profiles with custom scopes**
- ✅ **Authorization code flows with required/optional scopes and configurable consent screens**
- ✅ **Client Authentication Methods**: client_secret_basic, client_secret_post, client_secret_jwt, private_key_jwt, tls_client_auth, self_signed_tls_client_auth, bearer_token
- ✅ **Client MFA Chains**: Optional, configurable multi-factor authentication for client authentication**
- ✅ **PHASE 1 AUTH METHODS**: Username/password, email+OTP, TOTP/HOTP with QR codes, passkeys, basic client auth
- ✅ **Multi-Factor Authentication (MFA)**: TOTP/HOTP with QR code setup, configurable factor ordering (1-N factors per authentication profile)
- 🔄 **PHASE 2 AUTH METHODS**: Client MFA chains, SMS OTP, magic links, adaptive auth, biometrics, hardware keys (after integration testing)

### Phased Implementation Strategy

**WHY PHASED APPROACH:**

- **Risk Management**: Implement core functionality first, add complexity incrementally
- **MVP Focus**: Phase 1 covers 80% of common use cases with proven methods
- **Testing Maturity**: Each phase allows thorough testing before adding complexity
- **Team Scaling**: Later phases can be implemented by different team members
- **Market Validation**: Phase 1 can be deployed while Phase 2/3 are developed

**PHASE 1 - CORE MVP (Tasks 1-10):**

- Complete OAuth 2.1 compliance with PKCE
- Production-ready user authentication (password, email OTP, TOTP, passkeys)
- Basic client authentication for web/mobile apps
- Full test coverage and security validation
- Ready for production deployment

**PHASE 2 - ADVANCED AUTH METHODS (Tasks 11-15):**

- Client MFA chains (combining multiple client auth methods)
- SMS OTP and Magic Links for users
- Adaptive authentication (Step-Up, Risk-Based)
- Biometric authentication
- Hardware security keys (FIDO U2F/U2F2)
- Implemented after integration testing validation

**PHASE TRANSITIONS:**

- **Phase 1 → 2**: When core OAuth/OIDC flows are stable and tested
- **Phase 2 → 3**: When enterprise customers require advanced security features
- Each phase maintains backward compatibility with previous phases

### Security Requirements

- ✅ JWT token signing with RSA keys
- ✅ Secure password hashing
- ✅ Rate limiting and abuse prevention
- ✅ Audit logging for security events
- ✅ Input validation and sanitization
- ✅ Secure session management
- ✅ Magic values properly isolated in dedicated files
- ✅ Maximum decoupling for microservice extraction

### Quality Requirements

- ✅ 95%+ code coverage with comprehensive tests for each task
- ✅ Parameterized unit tests covering happy and sad paths
- ✅ Edge case coverage and table-driven testing
- ✅ OpenAPI 3.0 component-based specifications
- ✅ Integration with existing cryptoutil patterns
- ✅ Structured logging and metrics
- ✅ Performance benchmarks meeting requirements
- ✅ Independent server startup for authz, idp, rs with HTTP/HTTPS support
- ✅ Complete OAuth 2.1 and OIDC specification coverage in integration and E2E tests
- ✅ Fully parameterized configurations (no hardcoded defaults)

### Compliance Testing Requirements

- ✅ **OAuth 2.1 BCP Compliance**: Security best current practices validation
- ✅ **OIDC Core 1.0 Certification**: Official OpenID Connect compliance testing
- ✅ **Multi-Level Testing Strategy**: Pre-commit, CI/CD, integration, and E2E compliance validation
- ✅ **Automated Compliance Tools**: oauth2c, OIDC Conformance Suite, OWASP ZAP, Nuclei integration
- ✅ **Security Scanning**: OAuth/OIDC-specific vulnerability detection and prevention
- ✅ **Specification Coverage**: Complete OAuth 2.1 and OIDC Core specification validation
- ✅ **PKCE Enforcement**: Mandatory PKCE for all authorization code flows
- ✅ **State Parameter Validation**: Required state validation and replay prevention
- ✅ **JWT Security**: Proper token signing, validation, and secure handling
- ✅ **Rate Limiting & Abuse Prevention**: DoS protection and security hardening

### Operational Requirements

- ✅ Health check endpoints
- ✅ Graceful shutdown handling
- ✅ Configuration validation
- ✅ Database migrations
- ✅ Docker containerization ready
- ✅ Fully parameterized configurations with no hardcoded defaults

## 🔄 SESSION MANAGEMENT

### Current Session Focus

**Active Task**: None (Ready to start Task 1)
**Next Action**: Open `01_foundation_setup.md` and begin implementation

### Cross-Task Validation Checkpoints

**Before starting each task (except Task 1), perform these quick validations:**

1. **Dependency Check**: Verify previous task's outputs exist and compile
2. **Integration Test**: Run basic compilation test to ensure interfaces match
3. **Import Validation**: Confirm only allowed `internal/common/*` imports are used

**Example validation commands:**

```bash
# Check previous task outputs exist
go build ./internal/identity/...

# Quick compilation test
go test -c ./internal/identity/... 2>&1 | head -20

# Import validation (should only show allowed internal/common/*)
go list -f '{{.Imports}}' ./internal/identity/... | grep -v "internal/common"
```

### Regrouping Protocol

After each task completion:

1. **Validate** completion criteria are met
2. **Cross-Task Check**: Run validation checkpoint for next task readiness
3. **Commit** with semantic message: `feat: complete Task X - [brief description]`
4. **Update** this master file with new status
5. **Plan** next task and open its file
6. **Reference** constraints section as needed

### Rollback Procedures

**If a task needs significant rework:**

1. **Assess Impact**: Check which downstream tasks are affected
2. **Stash Changes**: `git stash` current work if needed
3. **Reset to Last Good**: `git reset --hard HEAD~X` (where X = number of commits to undo)
4. **Reapply Selectively**: `git cherry-pick` only the commits you want to keep
5. **Update Status**: Mark affected tasks as `status:pending` in this file
6. **Restart**: Begin the reworked task with fresh context

**Emergency rollback (if multiple tasks affected):**

```bash
# Create backup branch
git branch backup-before-rollback

# Reset to before problematic task
git reset --hard <commit-hash-before-issue>

# Update all affected task statuses to pending
# Restart from the earliest affected task
```

### Emergency Reset

If Copilot loses context:

1. Close all files except the active task file
2. Re-read the task file's goal and constraints
3. Restart with focused prompt: "You are working on Task X from `filename`. Focus only on this task."

---

**Ready to begin? Start with Task 1: Open `01_foundation_setup.md`**
 
 
 
 
 
 
 
 
