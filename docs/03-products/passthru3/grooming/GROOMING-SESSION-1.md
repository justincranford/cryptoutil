# Grooming Session 1: Architecture & Design Decisions

**Purpose**: Lock down all architectural decisions BEFORE implementation
**Date**: 2025-12-01

---

## Topic 1: Integration Demo Architecture

### Q1.1: How should the integration demo manage server lifecycles?

**Decision**: Use embedded servers (not Docker) for demo simplicity

**Rationale**:
- KMS and Identity demos already use embedded approach
- Faster startup for demo purposes
- No Docker dependency for basic demo
- Consistent with existing demo pattern in kms.go and identity.go

**Implementation**:
```go
// Pattern: Start servers as goroutines with cleanup
func startIdentityServer(ctx context.Context) (*http.Server, error)
func startKMSServer(ctx context.Context) (*http.Server, error)
```

### Q1.2: What ports should the integration demo use?

**Decision**: Use dynamic port allocation or predefined demo ports

| Service | Demo Port | Production Port |
|---------|-----------|-----------------|
| Identity AuthZ | 18080 | 8082 |
| KMS | 18081 | 8080 |
| Admin/Health | 19090 | 9090 |

**Rationale**:
- Avoids conflict with any running production services
- Matches existing identity.go pattern (port 18080)

### Q1.3: How should inter-service communication work?

**Decision**: HTTPS with self-signed certs, skip verification for demo

**Rationale**:
- TLS is required for OAuth 2.1
- Demo self-signed certs already exist
- Skip verification acceptable for demo only (never production)

---

## Topic 2: OAuth Token Flow

### Q2.1: Which OAuth grant type for KMS authentication?

**Decision**: client_credentials grant

**Rationale**:
- Service-to-service authentication
- No user interaction required
- Already implemented in Identity demo

**Implementation**:
```go
tokenResp, err := getClientCredentialsToken(tokenEndpoint, clientID, clientSecret, scopes)
```

### Q2.2: How should KMS validate Identity tokens?

**Decision**: JWKS validation using Identity's public keys

**Steps**:
1. KMS fetches JWKS from Identity's `/oauth2/v1/jwks` endpoint
2. KMS validates JWT signature against JWKS
3. KMS validates claims (iss, aud, exp, nbf)
4. KMS checks required scopes in token

**Implementation**:
```go
func validateToken(ctx context.Context, token string, jwksURL string, requiredScopes []string) error
```

### Q2.3: What scopes are required for KMS operations?

**Decision**: Define demo-specific scopes

| Scope | Operations Allowed |
|-------|-------------------|
| `kms:read` | List keys, get key metadata |
| `kms:write` | Create keys, update keys |
| `kms:sign` | Sign operations |
| `kms:encrypt` | Encrypt/decrypt operations |
| `demo:all` | All demo operations (convenience) |

---

## Topic 3: Error Handling

### Q3.1: How should demo report failures?

**Decision**: Clear step-by-step output with failure details

**Pattern**:
```
Step 1/7: Start Identity server... ✅ PASS
Step 2/7: Start KMS server... ✅ PASS
Step 3/7: Service health checks... ❌ FAIL

Error: KMS health check failed after 30s timeout
  URL: https://127.0.0.1:18081/health
  Error: connection refused
  
Suggestion: Check if port 18081 is available
```

### Q3.2: Should demo continue on failure?

**Decision**: Fail-fast - stop on first failure

**Rationale**:
- Each step depends on previous
- No point continuing if Identity server fails
- Clear error message more useful than multiple failures

---

## Topic 4: Configuration

### Q4.1: Where should demo configuration live?

**Decision**: Embedded defaults with optional config file override

**Default Config**:
```yaml
# Embedded in integration.go
identity:
  port: 18080
  admin_port: 19090
kms:
  port: 18081
  admin_port: 19091
demo_client:
  id: demo-client
  secret: demo-secret
  scopes: ["demo:all"]
timeouts:
  startup: 30s
  health_check: 10s
  token_request: 5s
```

### Q4.2: Should demo use existing config files?

**Decision**: Yes, reuse existing demo configs

| Service | Config File |
|---------|-------------|
| Identity | `configs/identity/authz-demo.yaml` |
| KMS | Embedded config |

---

## Sign-Off

**All architectural decisions in this document are LOCKED**

- [ ] Reviewed and approved
- [ ] No open questions remain
- [ ] Ready for implementation

**Date**: ____________
**Approved By**: ____________
