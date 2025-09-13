# Security Guide

## Overview

cryptoutil implements a comprehensive multi-layered security architecture designed to protect cryptographic operations, key material, and API access. The security model follows defense-in-depth principles with multiple independent security barriers.

## Security Architecture

### Multi-Layer Security Model

```
┌─────────────────────────────────────────────────────────────┐
│                    Network Security                         │
│  • IP Allowlisting (Individual IPs + CIDR blocks)          │
│  • Rate Limiting (Per-IP throttling)                       │
│  • DDoS Protection                                         │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│                  Transport Security                         │
│  • TLS 1.3 with auto-generated certificates               │
│  • Certificate validation and management                   │
│  • Secure cipher suites                                   │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│                 Application Security                        │
│  • CORS (Cross-Origin Resource Sharing)                   │
│  • CSRF (Cross-Site Request Forgery) Protection           │
│  • CSP (Content Security Policy)                          │
│  • XSS Protection Headers                                 │
│  • Security Headers (Helmet.js equivalent)                │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│                Cryptographic Security                       │
│  • FIPS 140-3 Approved Algorithms                         │
│  • Hierarchical Key Management (Barrier System)           │
│  • Encrypted Key Storage                                  │
│  • Key Versioning and Rotation                            │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│                 Operational Security                        │
│  • Comprehensive Audit Logging                            │
│  • Secure Failure Modes                                   │
│  • Graceful Degradation                                   │
│  • Secret Management                                      │
└─────────────────────────────────────────────────────────────┘
```

## Network Security

### IP Allowlisting

Restricts access to approved IP addresses and network ranges:

```yaml
# Configuration
allowed_ips:
  - "127.0.0.1"
  - "::1"
  - "192.168.1.100"

allowed_cidrs:
  - "10.0.0.0/8"
  - "192.168.0.0/16"
  - "172.16.0.0/12"
```

**Features**:
- Individual IPv4 and IPv6 address support
- CIDR block notation for network ranges
- Real-time IP validation and blocking
- Detailed logging of access attempts

**Implementation**:
```go
// IP filtering middleware in application_listener.go
func commonIPFilterMiddleware(telemetryService, settings) {
  // Parse and validate allowed IPs/CIDRs
  // Block requests from non-allowed sources
  // Log access attempts and denials
}
```

### Rate Limiting

Prevents DoS attacks through per-IP request throttling:

```yaml
# Configuration
ip_rate_limit: 100  # requests per second per IP
```

**Features**:
- Per-IP address rate limiting
- Configurable threshold (requests per second)
- Automatic blocking of excessive requests
- Rate limit status logging

**Response for Rate Limit Exceeded**:
```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json

{
  "status": 429,
  "error": "Too Many Requests",
  "message": "Rate limit exceeded"
}
```

## Transport Security

### TLS Configuration

Automatic TLS certificate generation and management:

```yaml
# TLS Configuration
bind_public_protocol: "https"
bind_private_protocol: "http"  # Internal management interface

tls_public_dns_names:
  - "localhost"
  - "cryptoutil.example.com"

tls_public_ip_addresses:
  - "127.0.0.1"
  - "::1"
  - "192.168.1.100"
```

**Features**:
- Automatic certificate generation for development
- Support for custom DNS names and IP addresses
- TLS 1.3 with secure cipher suites
- Certificate validation and management

**Certificate Generation**:
```go
// Automatic TLS certificate creation
publicTLSServerCertificate, _, _, err := cryptoutilCertificate.BuildTLSCertificate(publicTLSServerSubject)
publicTLSServerConfig = &tls.Config{
  Certificates: []tls.Certificate{publicTLSServerCertificate},
  ClientAuth:   tls.NoClientCert
}
```

## Application Security

### CORS (Cross-Origin Resource Sharing)

Enables secure browser-based API access:

```yaml
# CORS Configuration
cors_allowed_origins: "http://localhost:3000,https://app.example.com"
cors_allowed_methods: "GET,POST,PUT,DELETE,OPTIONS"
cors_allowed_headers: "Content-Type,Authorization,X-CSRF-Token"
cors_max_age: 86400
```

**Features**:
- Configurable allowed origins, methods, and headers
- Preflight request handling
- Credential support for authenticated requests
- Only applied to browser API context (`/browser/api/v1/*`)

### CSRF (Cross-Site Request Forgery) Protection

Prevents unauthorized state-changing requests:

```yaml
# CSRF Configuration
csrf_token_name: "csrf_token"
csrf_token_same_site: "Strict"
csrf_token_max_age: "1h"
csrf_token_cookie_secure: true
csrf_token_cookie_http_only: true
csrf_token_single_use_token: false
```

**CSRF Token Flow**:
1. Client requests CSRF token: `GET /browser/api/v1/csrf-token`
2. Server sets secure cookie with CSRF token
3. Client includes token in request headers: `X-CSRF-Token: <token>`
4. Server validates token matches cookie value

**Swagger UI Integration**:
```javascript
// Automatic CSRF token handling in Swagger UI
function getCSRFToken() {
  const cookies = document.cookie.split(';');
  for (let cookie of cookies) {
    if (cookie.trim().startsWith(csrfTokenName + '=')) {
      return cookie.substring((csrfTokenName + '=').length);
    }
  }
  return null;
}
```

### Content Security Policy (CSP)

Prevents XSS attacks through strict content policies:

```javascript
// Generated CSP header
"default-src 'none'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'self'; object-src 'none';"
```

**CSP Directives**:
- `default-src 'none'` - Deny all by default
- `script-src 'self' 'unsafe-inline' 'unsafe-eval'` - Allow scripts from same origin (required for Swagger UI)
- `style-src 'self' 'unsafe-inline'` - Allow styles from same origin
- `frame-ancestors 'none'` - Prevent clickjacking
- Development mode adds localhost variations

### Security Headers

Comprehensive security headers (Helmet.js equivalent):

```http
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Referrer-Policy: same-origin
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
Permissions-Policy: camera=(), microphone=(), geolocation=(), payment=()
```

**Header Descriptions**:
- **X-Frame-Options**: Prevents clickjacking attacks
- **X-Content-Type-Options**: Prevents MIME sniffing
- **X-XSS-Protection**: Enables browser XSS filtering
- **Strict-Transport-Security**: Enforces HTTPS connections
- **Permissions-Policy**: Restricts browser API access

## Cryptographic Security

### FIPS 140-3 Compliance

Only approved cryptographic algorithms and key sizes:

**Supported Algorithms**:
- **RSA**: 2048, 3072, 4096 bits
- **ECDSA/ECDH**: P-256, P-384, P-521 curves
- **EdDSA**: Ed25519
- **AES**: 128, 192, 256 bits
- **HMAC**: SHA-256, SHA-384, SHA-512

**Key Generation**:
```go
// Example: RSA key generation with FIPS-approved sizes
func GenerateRSAKey(bits int) (*rsa.PrivateKey, error) {
  if bits < 2048 {
    return nil, errors.New("RSA key size must be ≥ 2048 bits for FIPS compliance")
  }
  return rsa.GenerateKey(rand.Reader, bits)
}
```

### Hierarchical Key Management (Barrier System)

Multi-tier key protection inspired by HashiCorp Vault:

```
┌─────────────────┐
│   Unseal Keys   │ ← Master keys for system initialization
└─────────────────┘
         │ (encrypts)
┌─────────────────┐
│   Root Keys     │ ← Primary encryption keys
└─────────────────┘
         │ (encrypts)
┌─────────────────┐
│Intermediate Keys│ ← Secondary encryption layer
└─────────────────┘
         │ (encrypts)
┌─────────────────┐
│ Content Keys    │ ← Actual material encryption keys
└─────────────────┘
```

**Unseal Modes**:

1. **Simple Keys**: Direct unseal key file
```yaml
unseal_mode: "simple"
unseal_files: ["/path/to/unseal.key"]
```

2. **Shared Secrets**: M-of-N secret sharing
```yaml
unseal_mode: "shamir"
unseal_files: 
  - "/path/to/unseal_1of5.secret"
  - "/path/to/unseal_2of5.secret"
  - "/path/to/unseal_3of5.secret"
```

3. **System Fingerprinting**: Hardware-based unsealing
```yaml
unseal_mode: "system"
unseal_files: ["/path/to/system.fingerprint"]
```

### Key Storage Encryption

All sensitive key material is encrypted at rest:

```go
// Key encryption using barrier system
encryptedKey, err := barrierService.Encrypt(rootKey, materialKey)
if err != nil {
  return fmt.Errorf("failed to encrypt key material: %w", err)
}
```

**Encryption Features**:
- AES-GCM encryption for key material
- Unique encryption keys per barrier level
- Key versioning for rotation support
- Secure key derivation functions

## Operational Security

### Audit Logging

Comprehensive logging of security events:

**Logged Events**:
- Authentication attempts and failures
- Rate limiting triggers
- IP access denials
- Key creation, access, and deletion
- Administrative operations
- System startup and shutdown

**Log Format**:
```json
{
  "timestamp": "2025-09-12T10:30:00Z",
  "level": "WARN",
  "message": "Rate limit exceeded",
  "requestId": "req_123456",
  "clientIP": "192.168.1.100",
  "method": "POST",
  "url": "/browser/api/v1/elastickey",
  "userAgent": "Mozilla/5.0...",
  "headers": {...}
}
```

### Secret Management

Secure handling of sensitive configuration:

**Docker Secrets Integration**:
```yaml
# docker-compose.yml
secrets:
  - source: cryptoutil_database_url
    target: /run/secrets/database_url
  - source: cryptoutil_unseal_key
    target: /run/secrets/unseal_key
```

**Environment Variable Protection**:
- No sensitive data in environment variables
- Configuration file encryption support
- Runtime secret injection

### Secure Failure Modes

System designed to fail securely:

**Failure Behaviors**:
- **Unseal Failure**: System remains sealed, no key access
- **Database Failure**: Graceful degradation, health check failures
- **Rate Limit**: Request blocking, no service disruption
- **Certificate Failure**: TLS errors, connection rejection
- **Memory Pressure**: Graceful shutdown, resource cleanup

**Health Check Integration**:
```bash
# Kubernetes liveness probe
curl -f http://localhost:9090/livez || exit 1

# Kubernetes readiness probe  
curl -f http://localhost:9090/readyz || exit 1
```

## Security Configuration

### Development vs Production

**Development Mode** (`--dev` flag):
- Relaxed CSP policies for debugging
- Extended CSRF token details in error responses
- Verbose security logging
- Self-signed certificate acceptance
- In-memory SQLite database

**Production Mode** (default):
- Strict security policies
- Minimal error information disclosure
- Production-ready database connections
- Certificate validation enforcement
- Comprehensive audit logging

### Security Best Practices

**Configuration Recommendations**:

1. **Network Security**:
   ```yaml
   # Restrict to specific IP ranges
   allowed_cidrs: ["10.0.0.0/8"]  # Internal network only
   ip_rate_limit: 10              # Conservative rate limiting
   ```

2. **TLS Configuration**:
   ```yaml
   bind_public_protocol: "https"  # Always use HTTPS in production
   bind_private_protocol: "https" # Secure management interface
   ```

3. **CSRF Protection**:
   ```yaml
   csrf_token_cookie_secure: true      # HTTPS only
   csrf_token_same_site: "Strict"      # Strict same-site policy
   csrf_token_single_use_token: true   # Enhanced security
   ```

4. **Database Security**:
   ```yaml
   database_url: "postgres://user:pass@host:5432/db?sslmode=require"
   ```

### Security Monitoring

**Metrics to Monitor**:
- Rate limiting trigger frequency
- Failed authentication attempts
- Unusual IP access patterns
- Key access frequency and patterns
- System resource utilization

**Alerting Recommendations**:
- Rate limit threshold breaches
- Repeated authentication failures
- Unauthorized IP access attempts
- System health check failures
- Certificate expiration warnings

### Incident Response

**Security Event Response**:

1. **Rate Limit Breach**: Investigate source IP, consider blocking
2. **Authentication Failure**: Check for brute force attempts
3. **Unauthorized Access**: Verify IP allowlist configuration
4. **Key Access Anomaly**: Audit key usage patterns
5. **System Compromise**: Immediate shutdown and investigation

**Recovery Procedures**:
- Graceful shutdown: `curl -X POST http://localhost:9090/shutdown`
- Key rotation: Generate new unseal/root keys
- Certificate renewal: Restart with new TLS configuration
- Database restoration: Restore from encrypted backups

## Security Updates

### Keeping Current

1. **Go Security Updates**: Regularly update Go runtime
2. **Dependency Updates**: Monitor and update dependencies
3. **Algorithm Updates**: Stay current with NIST recommendations
4. **Security Patches**: Apply security fixes promptly

### Security Assessment

**Regular Security Reviews**:
- Code security audits
- Penetration testing
- Vulnerability scanning
- Compliance verification (FIPS 140-3)
- Configuration review

**Security Testing**:
```bash
# Test rate limiting
for i in {1..150}; do curl -s http://localhost:8080/service/api/v1/elastickeys; done

# Test IP filtering
curl -H "X-Forwarded-For: 192.168.255.255" http://localhost:8080/service/api/v1/elastickeys

# Test CSRF protection
curl -X POST http://localhost:8080/browser/api/v1/elastickey \
  -H "Content-Type: application/json" \
  -d '{"name": "test"}'
```

This comprehensive security guide ensures cryptoutil maintains enterprise-grade security across all operational aspects while remaining usable for development and production deployment scenarios.
