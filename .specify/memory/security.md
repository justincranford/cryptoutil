# Security Implementation Patterns

**Referenced by**: `.github/instructions/03-06.security.instructions.md`

## Vulnerability Monitoring

Check weekly: `govulncheck ./...`. Sources: <https://pkg.go.dev/vuln/list>, <https://github.com/advisories>, <https://www.cvedetails.com/>. Update incrementally with testing.

---

## Multi-Layer Key Hierarchy

### Barrier Architecture

**Pattern: Unseal → Root → Intermediate → Content Keys**

```
┌──────────────┐
│ Unseal Key   │ ← Stored in Docker secrets, derived with HKDF
└──────┬───────┘
       │ Unseals
       ▼
┌──────────────┐
│ Root Key     │ ← Master key for encryption hierarchy
└──────┬───────┘
       │ Derives
       ▼
┌──────────────┐
│ Intermediate │ ← Domain-specific keys (per service, per tenant)
└──────┬───────┘
       │ Encrypts
       ▼
┌──────────────┐
│ Content Keys │ ← Ephemeral keys for data encryption
└──────────────┘
```

**Key Characteristics**:

- **Unseal Key**: Never stored in application, injected via Docker secrets at runtime
- **Root Key**: Encrypted at rest with unseal key, rotated annually
- **Intermediate Keys**: Encrypted with root key, rotated quarterly
- **Content Keys**: Encrypted with intermediate keys, rotated per-operation or hourly

**See**: `02-07.cryptography.instructions.md` for HKDF-based key derivation

---

## Network Security

### IP Allowlisting

**Pattern: IP addresses + CIDR ranges**

```yaml
security:
  ip_allowlist:
    - 127.0.0.1          # Loopback IPv4
    - ::1                # Loopback IPv6
    - 10.0.0.0/8         # Private network (Class A)
    - 172.16.0.0/12      # Private network (Class B)
    - 192.168.0.0/16     # Private network (Class C)
```

**Implementation**:

```go
import "net"

func isIPAllowed(remoteAddr string, allowlist []string) bool {
    ip := net.ParseIP(remoteAddr)
    if ip == nil {
        return false
    }

    for _, entry := range allowlist {
        if strings.Contains(entry, "/") {
            // CIDR range
            _, ipNet, _ := net.ParseCIDR(entry)
            if ipNet.Contains(ip) {
                return true
            }
        } else {
            // Exact IP match
            if ip.Equal(net.ParseIP(entry)) {
                return true
            }
        }
    }
    return false
}
```

### Per-IP Rate Limiting

**Pattern: Token bucket algorithm with per-IP tracking**

```go
import "golang.org/x/time/rate"

type IPRateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit  // Requests per second
    burst    int         // Burst capacity
}

func (rl *IPRateLimiter) Allow(ip string) bool {
    rl.mu.Lock()
    limiter, exists := rl.limiters[ip]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.limiters[ip] = limiter
    }
    rl.mu.Unlock()

    return limiter.Allow()
}
```

**Recommended Limits**:

- Public APIs: 100 requests/min per IP (burst: 20)
- Admin APIs: 10 requests/min per IP (burst: 5)
- Login endpoints: 5 requests/min per IP (burst: 2)

---

## Web Security Headers

### CORS Configuration

**Pattern: Restrict origins, credentials, methods**

```yaml
cors:
  allowed_origins:
    - "http://localhost:8080"
    - "https://app.example.com"
  allowed_methods:
    - GET
    - POST
    - PUT
    - DELETE
  allowed_headers:
    - Authorization
    - Content-Type
  expose_headers:
    - X-Request-ID
  allow_credentials: true
  max_age: 3600  # Preflight cache (1 hour)
```

**Implementation (Fiber)**:

```go
import "github.com/gofiber/fiber/v3/middleware/cors"

app.Use(cors.New(cors.Config{
    AllowOrigins:     config.CORS.AllowedOrigins,
    AllowMethods:     config.CORS.AllowedMethods,
    AllowHeaders:     config.CORS.AllowedHeaders,
    ExposeHeaders:    config.CORS.ExposeHeaders,
    AllowCredentials: config.CORS.AllowCredentials,
    MaxAge:           config.CORS.MaxAge,
}))
```

### CSRF Protection

**Pattern: Double-submit cookie or synchronizer token**

```go
import "github.com/gofiber/fiber/v3/middleware/csrf"

app.Use(csrf.New(csrf.Config{
    KeyLookup:      "header:X-CSRF-Token",
    CookieName:     "csrf_token",
    CookieSameSite: "Strict",
    CookieSecure:   true,
    CookieHTTPOnly: true,
    Expiration:     1 * time.Hour,
}))
```

**Exempt Paths**: `/service/**` endpoints (headless clients, not browsers)

### Content Security Policy (CSP)

**Pattern: Strict CSP headers for browser clients**

```yaml
csp:
  default_src: "'self'"
  script_src: "'self' 'unsafe-inline'"  # Allow inline scripts (minimize usage)
  style_src: "'self' 'unsafe-inline'"   # Allow inline styles
  img_src: "'self' data: https:"
  connect_src: "'self' wss:"            # Allow WebSocket connections
  font_src: "'self'"
  object_src: "'none'"
  base_uri: "'self'"
  form_action: "'self'"
```

**Implementation**:

```go
app.Use(func(c *fiber.Ctx) error {
    c.Set("Content-Security-Policy", config.CSP.ToString())
    return c.Next()
})
```

---

## Secret Management

### Docker Secrets Pattern

**MANDATORY: Use Docker secrets, NEVER environment variables**

**Configuration**:

```yaml
# docker-compose.yml
services:
  cryptoutil:
    secrets:
      - database_url
      - unseal_key
      - tls_cert
      - tls_key
    command:
      - "--database-url=file:///run/secrets/database_url"
      - "--unseal-key=file:///run/secrets/unseal_key"
      - "--tls-cert=file:///run/secrets/tls_cert"
      - "--tls-key=file:///run/secrets/tls_key"

secrets:
  database_url:
    file: ./deployments/compose/secrets/database_url.secret
  unseal_key:
    file: ./deployments/compose/secrets/unseal_key.secret
  tls_cert:
    file: ./deployments/compose/secrets/tls_cert.secret
  tls_key:
    file: ./deployments/compose/secrets/tls_key.secret
```

**Application Code**:

```go
func loadSecret(path string) ([]byte, error) {
    if strings.HasPrefix(path, "file://") {
        secretPath := strings.TrimPrefix(path, "file://")
        return os.ReadFile(secretPath)
    }
    return []byte(path), nil  // Fallback: treat as literal value (dev only)
}

// Usage
dbURL, err := loadSecret(config.DatabaseURL)
unsealKey, err := loadSecret(config.UnsealKey)
```

**Secret File Permissions**: **MANDATORY 440 (r--r----)**

```bash
chmod 440 deployments/compose/*/secrets/*.secret
```

### Kubernetes Secrets

**Pattern: Mount as files or reference directly**

```yaml
# deployment.yaml
apiVersion: v1
kind: Pod
metadata:
  name: cryptoutil
spec:
  containers:
  - name: cryptoutil
    image: cryptoutil:latest
    args:
      - "--database-url=file:///var/secrets/database_url"
    volumeMounts:
    - name: secrets
      mountPath: "/var/secrets"
      readOnly: true
  volumes:
  - name: secrets
    secret:
      secretName: cryptoutil-secrets
      defaultMode: 0440
```

**Create Secret**:

```bash
kubectl create secret generic cryptoutil-secrets \
  --from-file=database_url=./secrets/database_url.secret \
  --from-file=unseal_key=./secrets/unseal_key.secret
```

---

## Windows Firewall Exception Prevention

### CRITICAL: Bind to 127.0.0.1 in Tests

**Problem**: Binding to `0.0.0.0` triggers Windows Firewall prompts, blocking automation

**Violation Impact**:

- ❌ Each `0.0.0.0` binding = 1 Windows Firewall popup
- ❌ Blocks CI/CD automation (requires manual approval)
- ❌ Security risk (accidentally exposing test services to network)

**Correct Patterns**:

```go
// ✅ CORRECT: Bind to loopback only (no firewall prompt)
import cryptoutilMagic "cryptoutil/internal/shared/magic"

addr := fmt.Sprintf("%s:%d", cryptoutilMagic.IPv4Loopback, port)  // "127.0.0.1:port"
listener, err := net.Listen("tcp", addr)

// ❌ WRONG: Bind to all interfaces (triggers firewall prompt)
listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
```

**Configuration Files**:

```yaml
# ✅ CORRECT: Test configs use loopback
server:
  bind_address: 127.0.0.1
  bind_port: 0  # Dynamic allocation

# ❌ WRONG: Production pattern in test configs
server:
  bind_address: 0.0.0.0  # Only for Docker containers, NEVER for local tests
```

**Environment-Specific Binding**:

| Environment | Bind Address | Rationale |
|-------------|--------------|-----------|
| Local tests | 127.0.0.1 | Avoid firewall prompts |
| Docker containers | 0.0.0.0 | Container network isolation |
| Kubernetes pods | 0.0.0.0 | Pod network isolation |
| GitHub Actions | 127.0.0.1 | CI/CD environment |

### Localhost vs 127.0.0.1 Decision Matrix

| Environment | localhost | 127.0.0.1 | Preferred | Rationale |
|-------------|-----------|-----------|-----------|-----------|
| **Local Windows Dev** | ✅ | ✅ | `127.0.0.1` | Avoid firewall prompts |
| **GitHub Workflows** | ✅ | ✅ | `127.0.0.1` | Explicit IPv4, no DNS resolution |
| **Act Containers** | ✅ | ✅ | `127.0.0.1` | Consistent with GitHub |
| **Docker Containers (internal)** | ❌ | ✅ | `127.0.0.1` | Alpine resolves localhost to ::1 (IPv6) |
| **Docker Compose (host→container)** | ✅ | ✅ | `localhost` | Docker DNS handles resolution |
| **Go Code (bind addresses)** | ❌ | ✅ | `127.0.0.1` | Explicit IPv4, no ambiguity |
| **Go Code (database DSN)** | ✅ | ✅ | `localhost` | PostgreSQL driver handles resolution |

**Quick Reference for Go Server Binding**:

```go
// ALWAYS use magic constant
bindAddress := cryptoutilMagic.IPv4Loopback  // "127.0.0.1"
```

---

## Cryptographic Best Practices

### Random Number Generation

**MANDATORY: crypto/rand ALWAYS, NEVER math/rand**

```go
import crand "crypto/rand"

// ✅ CORRECT: Cryptographically secure random
nonce := make([]byte, 32)
if _, err := crand.Read(nonce); err != nil {
    return fmt.Errorf("failed to generate nonce: %w", err)
}

// ❌ WRONG: Predictable random (NOT secure)
import "math/rand"
nonce := make([]byte, 32)
rand.Read(nonce)  // DO NOT USE
```

**See**: `02-07.cryptography.instructions.md` for complete cryptographic requirements

### Certificate Validation

**MANDATORY: Full cert chain validation, TLS 1.3+, NEVER InsecureSkipVerify**

```go
import (
    "crypto/tls"
    "crypto/x509"
)

// ✅ CORRECT: Strict TLS configuration
tlsConfig := &tls.Config{
    MinVersion:         tls.VersionTLS13,
    InsecureSkipVerify: false,  // ALWAYS validate certificates
    RootCAs:            certPool,
    ClientCAs:          certPool,
    ClientAuth:         tls.RequireAndVerifyClientCert,
}

// ❌ WRONG: Insecure TLS (bypasses validation)
tlsConfig := &tls.Config{
    InsecureSkipVerify: true,   // NEVER do this
    MinVersion:         tls.VersionTLS12,  // Too old
}
```

**See**: `02-09.pki.instructions.md` for complete PKI requirements

---

## Audit Logging

### Security Event Logging

**Pattern: Structured logs with security-relevant fields**

```go
import "go.uber.org/zap"

logger.Info("authentication_attempt",
    zap.String("event", "authn_attempt"),
    zap.String("user_id", userID),
    zap.String("ip_address", remoteIP),
    zap.String("authn_method", "password"),
    zap.Bool("success", true),
    zap.Duration("duration", time.Since(start)),
)

logger.Warn("authorization_denied",
    zap.String("event", "authz_denied"),
    zap.String("user_id", userID),
    zap.String("resource", resourceID),
    zap.String("required_scope", requiredScope),
    zap.Strings("user_scopes", userScopes),
)
```

**Security Events to Log**:

- Authentication attempts (success + failures)
- Authorization denials
- Key generation/rotation events
- Certificate issuance/revocation
- Admin API access (/shutdown, /livez, /readyz)
- Rate limit violations
- IP allowlist violations

**Log Retention**: Minimum 90 days for security events, 1 year for compliance

---

## Secure Failure Modes

### Key Versioning and Rotation

**Pattern: Elastic Key Ring (active key + historical keys)**

```go
type KeyRing struct {
    ActiveKeyID   string
    ActiveKey     []byte
    HistoricalKeys map[string][]byte  // keyID -> key material
}

func (kr *KeyRing) Encrypt(plaintext []byte) ([]byte, error) {
    // ALWAYS use active key for encryption
    ciphertext, err := aesGCMEncrypt(kr.ActiveKey, plaintext)
    if err != nil {
        return nil, err
    }

    // Prepend key ID to ciphertext
    return append([]byte(kr.ActiveKeyID), ciphertext...), nil
}

func (kr *KeyRing) Decrypt(ciphertext []byte) ([]byte, error) {
    // Extract key ID from ciphertext
    keyID := string(ciphertext[:len(kr.ActiveKeyID)])

    // Use matching historical key for decryption
    key, exists := kr.HistoricalKeys[keyID]
    if !exists && keyID != kr.ActiveKeyID {
        return nil, fmt.Errorf("unknown key ID: %s", keyID)
    }
    if keyID == kr.ActiveKeyID {
        key = kr.ActiveKey
    }

    return aesGCMDecrypt(key, ciphertext[len(keyID):])
}
```

**Rotation Strategy**:

1. Generate new key → set as active
2. Keep old active key in historical keys (don't delete immediately)
3. Decrypt uses key ID from ciphertext to find correct historical key
4. Re-encrypt with new active key on next write (lazy migration)

**See**: `02-07.cryptography.instructions.md` for HKDF-based key derivation patterns

### Multiple Unseal Modes

**Supported Modes**:

- **Auto-unseal**: Unseal keys stored in Docker secrets, automatic unsealing on startup
- **Manual-unseal**: Operator provides unseal key via admin API
- **Multi-party-unseal**: Shamir secret sharing (3 of 5 key shares required)

**Configuration**:

```yaml
unseal:
  mode: auto  # auto, manual, multi-party
  auto:
    key_path: file:///run/secrets/unseal_key
  multi_party:
    threshold: 3
    total_shares: 5
```

---

## Key Takeaways

1. **Windows Firewall Prevention**: ALWAYS bind to 127.0.0.1 in tests (NEVER 0.0.0.0)
2. **Secret Management**: Docker secrets or Kubernetes secrets (NEVER environment variables)
3. **Cryptographic Security**: crypto/rand ALWAYS (NEVER math/rand), TLS 1.3+ with full cert validation
4. **Multi-Layer Keys**: Unseal → Root → Intermediate → Content (hierarchical encryption)
5. **Network Security**: IP allowlisting + per-IP rate limiting + CORS + CSRF + CSP headers
6. **Audit Logging**: Log all security events with structured fields, 90-day retention minimum
7. **Elastic Key Rotation**: Active key for encryption, historical keys for decryption (key ID embedded in ciphertext)
