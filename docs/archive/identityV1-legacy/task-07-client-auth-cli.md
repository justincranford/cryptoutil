# Task 07: Client Authentication CLI Documentation

## Overview

This document provides CLI commands and examples for testing and configuring client authentication methods in the Identity V2 system.

## Authentication Methods

### 1. client_secret_basic (HTTP Basic Auth)

**Configuration:**

```yaml
# In client registration
token_endpoint_auth_method: client_secret_basic
```

**Testing with curl:**

```bash
# Authorization Code Token Request
curl -X POST https://localhost:8443/oauth2/token \
  -H "Authorization: Basic $(echo -n 'client_id:client_secret' | base64)" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=AUTH_CODE" \
  -d "redirect_uri=https://example.com/callback"

# Client Credentials Grant
curl -X POST https://localhost:8443/oauth2/token \
  -H "Authorization: Basic $(echo -n 'client_id:client_secret' | base64)" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "scope=api:read api:write"
```

**PowerShell Example:**

```powershell
$clientId = "your_client_id"
$clientSecret = "your_client_secret"
$credentials = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("${clientId}:${clientSecret}"))

Invoke-RestMethod -Method Post -Uri "https://localhost:8443/oauth2/token" `
  -Headers @{
    "Authorization" = "Basic $credentials"
    "Content-Type" = "application/x-www-form-urlencoded"
  } `
  -Body @{
    grant_type = "client_credentials"
    scope = "api:read api:write"
  }
```

---

### 2. client_secret_post (Form-Encoded Body)

**Configuration:**

```yaml
# In client registration
token_endpoint_auth_method: client_secret_post
```

**Testing with curl:**

```bash
curl -X POST https://localhost:8443/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=your_client_id" \
  -d "client_secret=your_client_secret" \
  -d "scope=api:read api:write"
```

**PowerShell Example:**

```powershell
Invoke-RestMethod -Method Post -Uri "https://localhost:8443/oauth2/token" `
  -Headers @{ "Content-Type" = "application/x-www-form-urlencoded" } `
  -Body @{
    grant_type = "client_credentials"
    client_id = "your_client_id"
    client_secret = "your_client_secret"
    scope = "api:read api:write"
  }
```

---

### 3. client_secret_jwt (HMAC-signed JWT)

**Configuration:**

```yaml
# In client registration
token_endpoint_auth_method: client_secret_jwt
token_endpoint_auth_signing_alg: HS256  # or HS384, HS512
```

**Generate JWT (Go):**

```go
import (
    "time"
    joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
    joseJws "github.com/lestrrat-go/jwx/v3/jws"
    joseJwt "github.com/lestrrat-go/jwx/v3/jwt"
)

// Create JWT claims
token := joseJwt.New()
token.Set(joseJwt.IssuerKey, "your_client_id")
token.Set(joseJwt.SubjectKey, "your_client_id")
token.Set(joseJwt.AudienceKey, "https://localhost:8443/oauth2/token")
token.Set(joseJwt.ExpirationKey, time.Now().Add(5*time.Minute).Unix())
token.Set(joseJwt.IssuedAtKey, time.Now().Unix())
token.Set(joseJwt.JwtIDKey, uuid.NewString())

// Create HMAC key from client secret
key, _ := joseJwk.FromRaw([]byte("your_client_secret"))
key.Set(joseJwk.KeyIDKey, "client-secret-key")
key.Set(joseJwk.AlgorithmKey, joseJwa.HS256)

// Sign JWT
signed, _ := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.HS256, key))
```

**Testing with curl:**

```bash
curl -X POST https://localhost:8443/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=your_client_id" \
  -d "client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer" \
  -d "client_assertion=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d "scope=api:read"
```

---

### 4. private_key_jwt (RSA/ECDSA-signed JWT)

**Configuration:**

```yaml
# In client registration
token_endpoint_auth_method: private_key_jwt
token_endpoint_auth_signing_alg: RS256  # or RS384, RS512, ES256, ES384, ES512
jwks_uri: https://client.example.com/.well-known/jwks.json
# OR inline jwks:
jwks:
  keys:
    - kty: RSA
      kid: rsa-key-1
      use: sig
      alg: RS256
      n: "..."
      e: "AQAB"
```

**Generate RSA Key Pair (OpenSSL):**

```bash
# Generate private key
openssl genrsa -out client_private_key.pem 2048

# Extract public key
openssl rsa -in client_private_key.pem -pubout -out client_public_key.pem

# View private key in PKCS#8 format
openssl pkcs8 -topk8 -inform PEM -outform PEM -in client_private_key.pem -nocrypt
```

**Generate ECDSA Key Pair (OpenSSL):**

```bash
# Generate P-256 private key
openssl ecparam -name prime256v1 -genkey -noout -out client_ec_private_key.pem

# Extract public key
openssl ec -in client_ec_private_key.pem -pubout -out client_ec_public_key.pem
```

**Generate JWT (Go):**

```go
import (
    "crypto/rsa"
    "crypto/x509"
    "encoding/pem"
    joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
    joseJws "github.com/lestrrat-go/jwx/v3/jws"
    joseJwt "github.com/lestrrat-go/jwx/v3/jwt"
)

// Load private key
pemData, _ := os.ReadFile("client_private_key.pem")
block, _ := pem.Decode(pemData)
privateKey, _ := x509.ParsePKCS1PrivateKey(block.Bytes)

// Create JWT claims (same as client_secret_jwt)
token := joseJwt.New()
token.Set(joseJwt.IssuerKey, "your_client_id")
token.Set(joseJwt.SubjectKey, "your_client_id")
token.Set(joseJwt.AudienceKey, "https://localhost:8443/oauth2/token")
token.Set(joseJwt.ExpirationKey, time.Now().Add(5*time.Minute).Unix())
token.Set(joseJwt.IssuedAtKey, time.Now().Unix())
token.Set(joseJwt.JwtIDKey, uuid.NewString())

// Create JWK from RSA private key
key, _ := joseJwk.FromRaw(privateKey)
key.Set(joseJwk.KeyIDKey, "rsa-key-1")
key.Set(joseJwk.AlgorithmKey, joseJwa.RS256)

// Sign JWT
signed, _ := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.RS256, key))
```

**Testing (same curl command as client_secret_jwt)**

---

### 5. tls_client_auth (mTLS with CA-signed Certificates)

**Configuration:**

```yaml
# In client registration
token_endpoint_auth_method: tls_client_auth
tls_client_auth_subject_dn: "CN=client.example.com,O=Example Inc,C=US"
# OR
tls_client_auth_san_dns: client.example.com
# OR
tls_client_auth_san_uri: https://client.example.com
# OR
tls_client_auth_san_ip: 192.0.2.1
```

**Server Configuration:**

```yaml
# Enable mTLS on token endpoint
tls:
  cert_file: /path/to/server_cert.pem
  key_file: /path/to/server_key.pem
  client_ca_file: /path/to/trusted_ca.pem  # CA that signed client certs
  client_auth_type: RequireAndVerifyClientCert
```

**Generate CA-Signed Client Certificate (OpenSSL):**

```bash
# 1. Generate client private key
openssl genrsa -out client_tls_key.pem 2048

# 2. Create CSR
openssl req -new -key client_tls_key.pem -out client_tls.csr \
  -subj "/C=US/O=Example Inc/CN=client.example.com"

# 3. Sign with CA (assuming you have ca.crt and ca.key)
openssl x509 -req -in client_tls.csr \
  -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out client_tls_cert.pem -days 365 \
  -extensions v3_req -extfile <(cat <<EOF
[v3_req]
subjectAltName = DNS:client.example.com
EOF
)
```

**Testing with curl:**

```bash
curl -X POST https://localhost:8443/oauth2/token \
  --cert client_tls_cert.pem \
  --key client_tls_key.pem \
  --cacert ca.crt \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=your_client_id" \
  -d "scope=api:read"
```

**Testing with PowerShell:**

```powershell
# Import client certificate to Windows cert store first, then:
$cert = Get-ChildItem -Path Cert:\CurrentUser\My | Where-Object { $_.Subject -match "client.example.com" }

Invoke-RestMethod -Method Post -Uri "https://localhost:8443/oauth2/token" `
  -Certificate $cert `
  -Headers @{ "Content-Type" = "application/x-www-form-urlencoded" } `
  -Body @{
    grant_type = "client_credentials"
    client_id = "your_client_id"
    scope = "api:read"
  }
```

---

### 6. self_signed_tls_client_auth (mTLS with Self-Signed Certificates)

**Configuration:**

```yaml
# In client registration
token_endpoint_auth_method: self_signed_tls_client_auth
# MUST provide full public key in JWK format
jwks:
  keys:
    - kty: RSA
      kid: self-signed-tls-key
      use: sig
      x5c: ["MIIDXTCCAkWgAwIBAgIJAK..."]  # Base64-encoded DER certificate
```

**Generate Self-Signed Certificate (OpenSSL):**

```bash
# Generate private key and self-signed certificate in one command
openssl req -x509 -newkey rsa:2048 -keyout client_self_signed_key.pem \
  -out client_self_signed_cert.pem -days 365 -nodes \
  -subj "/C=US/O=Example Inc/CN=client.example.com"

# Convert certificate to DER format for JWK x5c
openssl x509 -in client_self_signed_cert.pem -outform DER -out client_self_signed_cert.der

# Base64-encode DER certificate for x5c field
base64 -w 0 client_self_signed_cert.der
```

**Testing (same curl/PowerShell commands as tls_client_auth)**

---

## Security Policy Profiles

### Default Policy

```yaml
# Allowed methods: all except 'none'
allowed_methods:
  - client_secret_basic
  - client_secret_post
  - client_secret_jwt
  - private_key_jwt
  - tls_client_auth
  - self_signed_tls_client_auth

# Certificate validation
certificate_max_age_days: 365
allow_self_signed_certs: true

# JWT algorithms
allowed_jwt_algorithms:
  - RS256
  - RS384
  - RS512
  - ES256
  - ES384
  - ES512
  - HS256  # Only for development
```

### Strict Policy

```yaml
# Allowed methods: strongest only
allowed_methods:
  - private_key_jwt
  - tls_client_auth

require_mtls: true
require_jwt_signature: true

# Certificate validation
certificate_max_age_days: 90
allow_self_signed_certs: false

# JWT algorithms: RSA/ECDSA only
allowed_jwt_algorithms:
  - RS256
  - RS384
  - RS512
  - ES256
  - ES384
  - ES512
```

### Public Client Policy

```yaml
# Allowed methods: none (PKCE required instead)
allowed_methods:
  - none

require_pkce: true
require_mtls: false
require_jwt_signature: false
```

### Development Policy

```yaml
# Allowed methods: all including 'none'
allowed_methods:
  - client_secret_basic
  - client_secret_post
  - client_secret_jwt
  - private_key_jwt
  - tls_client_auth
  - self_signed_tls_client_auth
  - none

# Relaxed certificate validation
certificate_max_age_days: 3650  # 10 years
allow_self_signed_certs: true

# All JWT algorithms allowed
allowed_jwt_algorithms:
  - RS256
  - RS384
  - RS512
  - ES256
  - ES384
  - ES512
  - HS256
```

---

## Certificate Rotation Best Practices

### 1. Planning Certificate Rotation

**Timeline:**

- **Production (Strict Policy)**: Rotate certificates every 60 days (before 90-day limit)
- **Production (Default Policy)**: Rotate certificates every 300 days (before 365-day limit)
- **Development**: Rotate as needed (10-year limit)

### 2. Rotation Process (Zero-Downtime)

**Step 1: Generate New Certificate**

```bash
# Generate new key pair
openssl genrsa -out client_new_key.pem 2048

# Create CSR
openssl req -new -key client_new_key.pem -out client_new.csr \
  -subj "/C=US/O=Example Inc/CN=client.example.com"

# Sign with CA
openssl x509 -req -in client_new.csr -CA ca.crt -CAkey ca.key \
  -out client_new_cert.pem -days 365
```

**Step 2: Update Client Registration (Add New Certificate)**

```bash
# Add new certificate to client's JWKS (for self_signed_tls_client_auth)
curl -X PATCH https://localhost:8443/admin/clients/your_client_id \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jwks": {
      "keys": [
        {
          "kid": "old-cert-2024",
          "kty": "RSA",
          "x5c": ["MIID...old..."]
        },
        {
          "kid": "new-cert-2025",
          "kty": "RSA",
          "x5c": ["MIID...new..."]
        }
      ]
    }
  }'
```

**Step 3: Test New Certificate**

```bash
# Verify new certificate works
curl -X POST https://localhost:8443/oauth2/token \
  --cert client_new_cert.pem \
  --key client_new_key.pem \
  -d "grant_type=client_credentials&client_id=your_client_id"
```

**Step 4: Deploy New Certificate to Client Applications**

- Update client application configurations
- Monitor error rates during rollout
- Keep old certificate available for rollback

**Step 5: Remove Old Certificate (After Grace Period)**

```bash
# After 7-14 days, remove old certificate
curl -X PATCH https://localhost:8443/admin/clients/your_client_id \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jwks": {
      "keys": [
        {
          "kid": "new-cert-2025",
          "kty": "RSA",
          "x5c": ["MIID...new..."]
        }
      ]
    }
  }'
```

### 3. Automated Rotation (Recommended)

**Using cert-manager (Kubernetes):**

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: client-mtls-cert
spec:
  secretName: client-mtls-secret
  duration: 2160h  # 90 days
  renewBefore: 360h  # 15 days before expiry
  issuerRef:
    name: ca-issuer
    kind: Issuer
  commonName: client.example.com
  dnsNames:
    - client.example.com
```

**Using Let's Encrypt (for public clients):**

```bash
# Install certbot
apt-get install certbot

# Obtain certificate
certbot certonly --standalone -d client.example.com

# Auto-renewal (cron job)
0 0 * * * certbot renew --quiet --deploy-hook "systemctl reload oauth2-client"
```

---

## Troubleshooting

### Common Errors

**1. "invalid_client: Client authentication failed"**

- Check client_id and client_secret are correct
- For JWT methods, verify JWT signature and claims (iss, sub, aud, exp)
- For mTLS, verify certificate is valid and matches registration

**2. "invalid_client: Unsupported authentication method"**

- Verify client's `token_endpoint_auth_method` matches request
- Check policy allows the authentication method

**3. "invalid_client: Certificate validation failed"**

- For CA-signed: Verify certificate is signed by trusted CA
- For self-signed: Verify certificate matches JWKS in client registration
- Check certificate is not expired
- Verify certificate age is within policy limits (90 or 365 days)

**4. "invalid_client: JWT validation failed"**

- Check JWT `aud` claim matches token endpoint URL
- Verify JWT `exp` claim is not expired (must be future timestamp)
- Ensure JWT `iat` claim is not in the future
- For client_secret_jwt: Verify client_secret matches
- For private_key_jwt: Verify public key in JWKS matches private key used to sign

**5. "invalid_client: JWKS fetch failed"**

- Verify `jwks_uri` is accessible from authorization server
- Check JWKS JSON format is correct
- Ensure JWKS contains required fields (kty, kid, use, n, e for RSA)

### Debugging Commands

**Test JWT Decoding:**

```bash
# Decode JWT without verification (inspect claims)
echo "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." | \
  cut -d. -f2 | base64 -d | jq .
```

**Test Certificate Chain:**

```bash
# Verify certificate is signed by CA
openssl verify -CAfile ca.crt client_cert.pem

# View certificate details
openssl x509 -in client_cert.pem -text -noout

# Check certificate expiration
openssl x509 -in client_cert.pem -noout -dates
```

**Test mTLS Connection:**

```bash
# Test TLS handshake with client certificate
openssl s_client -connect localhost:8443 \
  -cert client_cert.pem \
  -key client_key.pem \
  -CAfile ca.crt \
  -showcerts
```

---

## References

- [RFC 6749: OAuth 2.0](https://datatracker.ietf.org/doc/html/rfc6749)
- [RFC 7521: OAuth 2.0 Assertion Framework](https://datatracker.ietf.org/doc/html/rfc7521)
- [RFC 7523: OAuth 2.0 JWT Client Authentication](https://datatracker.ietf.org/doc/html/rfc7523)
- [RFC 8705: OAuth 2.0 mTLS Client Authentication](https://datatracker.ietf.org/doc/html/rfc8705)
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)
