# Phase 5: Documentation & Demo Implementation Guide

**Duration**: Days 12-14 (8-12 hours)
**Prerequisites**: Phase 4 complete (all advanced testing implemented)
**Status**: ❌ Not Started

## Overview

Phase 5 creates demonstration videos showcasing cryptoutil capabilities. Per specs/001-cryptoutil/TASKS.md:

> "Minimal documentation. Products must be intuitive and work without users and developers reading large amounts of docs."

Focus is on **showing** not **telling** - create visual demonstrations that prove the system works end-to-end.

**Task Breakdown**:

- P5.1: JOSE Authority Demo (2h, 5-10min video)
- P5.2: Identity Server Demo (2-3h, 10-15min video)
- P5.3: KMS Demo (2-3h, 10-15min video)
- P5.4: CA Server Demo (2-3h, 10-15min video)
- P5.5: Integration Demo (3-4h, 15-20min video)
- P5.6: Unified Suite Demo (3-4h, 20-30min video)

## Demo Video Requirements

### Technical Setup

**Recording Tools**:

- Windows: OBS Studio (free, open source)
- Screen resolution: 1920x1080
- Frame rate: 30 FPS
- Audio: Clear voice narration
- Format: MP4 (H.264 + AAC)

**Demonstration Environment**:

- Use Docker Compose deployments
- Show real services running (not mocked)
- Use PowerShell terminal for Windows commands
- Show browser UI for Swagger endpoints
- Use Postman/curl for API demonstrations

**Storage**:

- Save videos to `docs/demos/` directory
- Use descriptive filenames: `P5.X-<name>-YYYYMMDD.mp4`
- Create thumbnail images: `P5.X-<name>-thumbnail.png`
- Keep raw recording files for re-editing if needed

### Video Structure Template

Each demo video follows this structure:

1. **Intro** (30s)
   - Title card: "cryptoutil - <Feature Name>"
   - Brief overview of what will be demonstrated
   - Prerequisites and setup requirements

2. **Setup** (1-2min)
   - Show Docker Compose startup
   - Verify services healthy
   - Show Swagger UI endpoints

3. **Core Demonstration** (60-80% of video)
   - Walk through primary use cases
   - Show API calls and responses
   - Demonstrate error handling
   - Show integration with other services

4. **Advanced Features** (10-20% of video)
   - Show configuration options
   - Demonstrate security features
   - Show observability integration

5. **Wrap-up** (30s)
   - Summary of what was shown
   - Links to documentation
   - Next steps for users

## Task Details

---

### P5.1: JOSE Authority Demo

**Priority**: HIGH
**Effort**: 2 hours
**Duration**: 5-10 minutes
**Status**: ❌ Not Started

**Objective**: Demonstrate JOSE (JWK, JWS, JWE, JWT) cryptographic operations through API and CLI.

**Demo Script**:

1. **Start Services** (1min)

   ```bash
   docker compose -f ./deployments/compose/compose.yml up -d jose-sqlite jose-postgres-1
   docker compose ps
   ```

2. **JWK Generation** (1min)
   - Open Swagger UI: `https://localhost:8080/ui/swagger/`
   - Navigate to `/jwk/v1/generate`
   - Generate RSA-2048 key
   - Generate EC P-256 key
   - Generate Ed25519 key
   - Show returned JWK structures

3. **JWS Operations** (2min)
   - Sign payload with RS256
   - Sign payload with ES256
   - Sign payload with EdDSA
   - Verify signatures
   - Show compact JWS format

4. **JWE Operations** (2min)
   - Encrypt with RSA-OAEP + AES-256-GCM
   - Encrypt with ECDH-ES + AES-256-GCM
   - Decrypt ciphertext
   - Show compact JWE format

5. **JWT Operations** (2min)
   - Generate JWT with claims
   - Validate JWT structure
   - Show JWT inspection
   - Demonstrate expiration handling

6. **Multi-Instance** (1min)
   - Show same operations on port 8081 (PostgreSQL backend)
   - Demonstrate service independence

**Key Takeaways**:

- ✅ JOSE operations work end-to-end
- ✅ Multiple cryptographic algorithms supported
- ✅ Both SQLite and PostgreSQL backends functional
- ✅ API responses are well-formed JSON

**Files to Create**:

- `docs/demos/P5.1-jose-demo.mp4`
- `docs/demos/P5.1-jose-thumbnail.png`
- `docs/demos/P5.1-jose-script.md` (detailed script with timestamps)

---

### P5.2: Identity Server Demo

**Priority**: HIGH
**Effort**: 2-3 hours
**Duration**: 10-15 minutes
**Status**: ❌ Not Started

**Objective**: Demonstrate complete OAuth 2.1 / OpenID Connect flows with MFA and WebAuthn.

**Demo Script**:

1. **Start Services** (1min)

   ```bash
   docker compose -f ./deployments/identity/compose.yml up -d
   docker compose ps
   ```

2. **Client Registration** (2min)
   - Create OAuth client via API
   - Show client credentials
   - Configure allowed scopes
   - Set redirect URIs

3. **User Registration** (2min)
   - Create user account
   - Set password
   - Enroll MFA (TOTP)
   - Show user profile

4. **Authorization Code Flow** (3min)
   - Initiate authorization request
   - User authentication with MFA
   - Authorization consent screen
   - Exchange code for tokens
   - Show access token and refresh token

5. **Client Credentials Flow** (2min)
   - Machine-to-machine authentication
   - Token generation
   - Token introspection
   - Token revocation

6. **MFA Flows** (2min)
   - TOTP authentication
   - Backup codes
   - WebAuthn enrollment (if implemented)

7. **Token Management** (2min)
   - Refresh token usage
   - Token revocation
   - Token introspection
   - Logout

**Key Takeaways**:

- ✅ Complete OAuth 2.1 compliance
- ✅ MFA integration working
- ✅ Multiple client authentication methods
- ✅ Token lifecycle management

**Files to Create**:

- `docs/demos/P5.2-identity-demo.mp4`
- `docs/demos/P5.2-identity-thumbnail.png`
- `docs/demos/P5.2-identity-script.md`

---

### P5.3: KMS Demo

**Priority**: HIGH
**Effort**: 2-3 hours
**Duration**: 10-15 minutes
**Status**: ❌ Not Started

**Objective**: Demonstrate key management, encryption, signing, and hierarchical key security.

**Demo Script**:

1. **Start Services** (1min)

   ```bash
   docker compose -f ./deployments/kms/compose.yml up -d
   ```

2. **Unseal Process** (2min)
   - Show sealed KMS state
   - Provide unseal secrets
   - Show transition to unsealed state
   - Explain hierarchical key architecture

3. **Key Generation** (2min)
   - Generate RSA-2048 key
   - Generate EC P-256 key
   - Generate AES-256 key
   - Generate HMAC-SHA256 key
   - Show key metadata

4. **Encryption Operations** (3min)
   - Encrypt data with AES key
   - Decrypt ciphertext
   - Encrypt with RSA-OAEP
   - Decrypt with RSA private key

5. **Signature Operations** (2min)
   - Sign data with RSA-PSS
   - Verify signature
   - Sign with ECDSA
   - Verify ECDSA signature

6. **Key Lifecycle** (2min)
   - Key rotation
   - Key versioning
   - Key deletion
   - Key recovery

7. **Security Features** (2min)
   - Audit logging
   - Access control
   - Rate limiting
   - IP allowlisting

**Key Takeaways**:

- ✅ Hierarchical key management operational
- ✅ Multiple cryptographic algorithms supported
- ✅ Secure key lifecycle management
- ✅ Audit and compliance features working

**Files to Create**:

- `docs/demos/P5.3-kms-demo.mp4`
- `docs/demos/P5.3-kms-thumbnail.png`
- `docs/demos/P5.3-kms-script.md`

---

### P5.4: CA Server Demo

**Priority**: HIGH
**Effort**: 2-3 hours
**Duration**: 10-15 minutes
**Status**: ❌ Not Started

**Objective**: Demonstrate certificate authority operations, EST protocol, and PKI workflows.

**Demo Script**:

1. **Start Services** (1min)

   ```bash
   docker compose -f ./deployments/ca/compose.yml up -d
   ```

2. **EST Bootstrap** (2min)
   - Retrieve CA certificates
   - Show trust anchor
   - Explain EST protocol

3. **Certificate Enrollment** (3min)
   - Generate CSR (RSA-2048)
   - EST simpleenroll
   - Receive signed certificate
   - Show certificate chain

4. **Certificate Renewal** (2min)
   - Generate new CSR
   - EST simplereenroll
   - Receive renewed certificate
   - Show certificate expiration handling

5. **CRL Operations** (2min)
   - Generate CRL
   - Revoke certificate
   - Update CRL
   - Show CRL distribution

6. **OCSP Operations** (2min)
   - OCSP request for good certificate
   - OCSP request for revoked certificate
   - Show OCSP response format

7. **TSA Operations** (2min)
   - Timestamp request
   - Receive timestamp token
   - Verify timestamp signature

**Key Takeaways**:

- ✅ EST protocol fully implemented
- ✅ Certificate lifecycle management working
- ✅ CRL and OCSP operational
- ✅ TSA timestamping functional

**Files to Create**:

- `docs/demos/P5.4-ca-demo.mp4`
- `docs/demos/P5.4-ca-thumbnail.png`
- `docs/demos/P5.4-ca-script.md`

---

### P5.5: Integration Demo

**Priority**: HIGH
**Effort**: 3-4 hours
**Duration**: 15-20 minutes
**Status**: ❌ Not Started

**Objective**: Demonstrate integration between Identity, KMS, and CA services for real-world workflows.

**Demo Script**:

1. **Full Stack Startup** (2min)

   ```bash
   docker compose -f ./deployments/compose/compose.yml up -d
   docker compose ps
   ```

2. **Scenario 1: API Client Certificate Authentication** (5min)
   - Register OAuth client with Identity Server
   - Use KMS to generate client certificate key pair
   - Request certificate from CA Server
   - Use certificate for OAuth client authentication
   - Show mTLS connection

3. **Scenario 2: JWT Signing with KMS** (4min)
   - Identity Server requests JWT signing key from KMS
   - KMS generates Ed25519 key
   - Identity Server signs JWT with KMS key
   - Client verifies JWT using KMS public key

4. **Scenario 3: Encrypted User Secrets** (4min)
   - User registers with Identity Server
   - Identity Server uses KMS to encrypt user secrets
   - User authenticates
   - Identity Server decrypts secrets with KMS
   - Show secure secret storage

5. **Scenario 4: Certificate-Based User Authentication** (4min)
   - User enrolls WebAuthn credential
   - CA Server issues user certificate
   - User authenticates with certificate
   - Show certificate validation flow

**Key Takeaways**:

- ✅ Services integrate seamlessly
- ✅ Real-world authentication workflows functional
- ✅ Security architecture validated
- ✅ End-to-end encryption operational

**Files to Create**:

- `docs/demos/P5.5-integration-demo.mp4`
- `docs/demos/P5.5-integration-thumbnail.png`
- `docs/demos/P5.5-integration-script.md`

---

### P5.6: Unified Suite Demo

**Priority**: MEDIUM
**Effort**: 3-4 hours
**Duration**: 20-30 minutes
**Status**: ❌ Not Started

**Objective**: Comprehensive demonstration of all cryptoutil capabilities in a cohesive production-like environment.

**Demo Script**:

1. **Introduction** (2min)
   - Overview of cryptoutil architecture
   - Service component diagram
   - Security architecture highlights

2. **Infrastructure Setup** (3min)
   - Show Docker Compose configuration
   - Start all services
   - Verify health checks
   - Show observability stack (Grafana)

3. **JOSE Capabilities** (3min)
   - Highlight key operations from P5.1
   - Show performance metrics

4. **Identity Management** (4min)
   - Highlight key operations from P5.2
   - Show OAuth flows
   - Demonstrate MFA

5. **Key Management** (4min)
   - Highlight key operations from P5.3
   - Show unseal process
   - Demonstrate encryption/signing

6. **PKI Operations** (4min)
   - Highlight key operations from P5.4
   - Show certificate lifecycle
   - Demonstrate EST protocol

7. **Integration Scenarios** (5min)
   - Highlight scenarios from P5.5
   - Show service interactions
   - Demonstrate security features

8. **Observability & Operations** (3min)
   - Show Grafana dashboards
   - Demonstrate tracing
   - Show logs aggregation
   - Display metrics

9. **Wrap-up** (2min)
   - Summary of capabilities
   - Deployment options
   - Future roadmap
   - Community resources

**Key Takeaways**:

- ✅ Complete cryptoutil suite demonstrated
- ✅ Production-ready deployment shown
- ✅ All services working together
- ✅ Observability and monitoring operational

**Files to Create**:

- `docs/demos/P5.6-unified-demo.mp4`
- `docs/demos/P5.6-unified-thumbnail.png`
- `docs/demos/P5.6-unified-script.md`

---

## Demo Creation Workflow

### Pre-Recording Checklist

- [ ] Script finalized with timing estimates
- [ ] Test environment verified working
- [ ] Screen recording software configured
- [ ] Audio input tested (clear voice)
- [ ] Browser tabs prepared (Swagger UI)
- [ ] Terminal windows prepared (PowerShell)
- [ ] Demo data prepared (test accounts, keys, etc.)

### Recording Process

1. **Dry Run**: Practice full demo 2-3 times
2. **Record**: Capture screen and audio
3. **Review**: Check for errors, pacing, clarity
4. **Re-record** sections if needed (splice together)
5. **Edit**: Trim mistakes, add title cards, annotations
6. **Export**: MP4 format, 1920x1080, 30fps

### Post-Production

1. **Create thumbnail** (PNG, 1920x1080)
2. **Save script** with actual timestamps
3. **Upload** to docs/demos/
4. **Update README** with demo links
5. **Commit and push**

### Quality Standards

- ✅ Clear audio narration (no background noise)
- ✅ Smooth pacing (not rushed, not too slow)
- ✅ Error-free demonstration (or errors explained)
- ✅ Professional appearance (no distracting elements)
- ✅ Concise (respect stated duration targets)

## Progress Tracking

After completing each demo, update `PROGRESS.md`:

```bash
# Edit PROGRESS.md to mark task complete
# Update executive summary percentages
# Commit and push
git add specs/001-cryptoutil/PROGRESS.md docs/demos/
git commit -m "docs(speckit): add P5.X demo video"
git push
```

## Validation Checklist

Before marking Phase 5 complete, verify:

- [ ] P5.1: JOSE demo video created and uploaded
- [ ] P5.2: Identity demo video created and uploaded
- [ ] P5.3: KMS demo video created and uploaded
- [ ] P5.4: CA demo video created and uploaded
- [ ] P5.5: Integration demo video created and uploaded
- [ ] P5.6: Unified suite demo video created and uploaded
- [ ] All scripts saved with actual timestamps
- [ ] All thumbnails created
- [ ] README.md updated with demo links
- [ ] PROGRESS.md updated with all P5.1-P5.6 marked complete

## Completion

After Phase 5 complete:

- **ALL 42 tasks complete** (100%)
- Execute `/speckit.checklist` to validate
- Update PROGRESS.md executive summary to 100%
- Celebrate! 🎉
