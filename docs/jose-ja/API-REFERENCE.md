# JOSE-JA API Reference

This document provides a comprehensive reference for the JOSE-JA (JSON Object Signing and Encryption - JWK Authority) service API.

## Base URLs

| Environment | Service API | Browser API | Admin API |
|-------------|-------------|-------------|-----------|
| Development | `https://127.0.0.1:8080/service/api/v1/jose` | `https://127.0.0.1:8080/browser/api/v1/jose` | `https://127.0.0.1:9090/admin/v1` |
| Docker | `https://jose-ja:8080/service/api/v1/jose` | `https://jose-ja:8080/browser/api/v1/jose` | `https://127.0.0.1:9090/admin/v1` |

## Authentication

### Service API (`/service/**`)

- **Bearer Token**: `Authorization: Bearer <access_token>`
- **mTLS**: Client certificate authentication.
- **API Key**: `X-API-Key: <api_key>` (for machine-to-machine).

### Browser API (`/browser/**`)

- **Session Cookie**: `Cookie: session=<session_token>`
- **CSRF Token**: Required via `X-CSRF-Token` header.
- **CORS**: Restricted to configured origins.

### Admin API (`/admin/**`)

- Localhost-only access (127.0.0.1).
- Optional mTLS for additional security.

## JWK Operations

### Generate JWK

Creates a new Elastic JWK with an initial Material Key.

**Endpoint**: `POST /service/api/v1/jose/jwk/generate`

**Request**:
```json
{
    "kid": "optional-custom-kid",
    "algorithm": "RS256",
    "key_type": "RSA",
    "key_size": 2048,
    "use": "sig",
    "max_material_keys": 10
}
```

**Parameters**:
| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `kid` | string | No | UUIDv7 | Key identifier |
| `algorithm` | string | Yes | - | JWA algorithm (RS256, ES256, etc.) |
| `key_type` | string | Yes | - | Key type (RSA, EC, oct) |
| `key_size` | integer | Conditional | - | Key size in bits (RSA: 2048/3072/4096) |
| `use` | string | Yes | - | Key use (sig or enc) |
| `max_material_keys` | integer | No | 10 | Maximum rotations allowed |

**Response** (201 Created):
```json
{
    "kid": "019bd10b-5d65-7bdd-a717-d1f057c85b8a",
    "algorithm": "RS256",
    "key_type": "RSA",
    "use": "sig",
    "created_at": "2024-01-15T10:30:00Z"
}
```

**Supported Algorithms**:
| Algorithm | Key Type | Key Size | Use |
|-----------|----------|----------|-----|
| RS256, RS384, RS512 | RSA | 2048, 3072, 4096 | sig |
| ES256 | EC | P-256 | sig |
| ES384 | EC | P-384 | sig |
| ES512 | EC | P-521 | sig |
| EdDSA | OKP | Ed25519 | sig |
| RSA-OAEP, RSA-OAEP-256 | RSA | 2048, 3072, 4096 | enc |
| A128KW, A192KW, A256KW | oct | 128, 192, 256 | enc |
| A128GCM, A192GCM, A256GCM | oct | 128, 192, 256 | enc |

---

### List JWKs

Lists all Elastic JWKs for the authenticated tenant.

**Endpoint**: `GET /service/api/v1/jose/jwk/list`

**Query Parameters**:
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number |
| `size` | integer | 20 | Items per page (max: 100) |
| `use` | string | - | Filter by key use (sig/enc) |

**Response** (200 OK):
```json
{
    "keys": [
        {
            "kid": "019bd10b-5d65-7bdd-a717-d1f057c85b8a",
            "algorithm": "RS256",
            "key_type": "RSA",
            "use": "sig",
            "created_at": "2024-01-15T10:30:00Z",
            "material_count": 3
        }
    ],
    "pagination": {
        "page": 1,
        "size": 20,
        "total": 1
    }
}
```

---

### Rotate Material Key

Creates a new Material Key for an existing Elastic JWK.

**Endpoint**: `POST /service/api/v1/jose/jwk/{kid}/rotate`

**Path Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `kid` | string | Elastic JWK key identifier |

**Response** (200 OK):
```json
{
    "kid": "019bd10b-5d65-7bdd-a717-d1f057c85b8a",
    "material_kid": "019bd10b-6789-7abc-b123-456789abcdef",
    "material_count": 4,
    "rotated_at": "2024-01-15T11:00:00Z"
}
```

**Error Responses**:
| Status | Code | Description |
|--------|------|-------------|
| 404 | NOT_FOUND | Elastic JWK not found |
| 409 | MAX_ROTATIONS | Maximum material keys reached |

---

## JWS Operations

### Sign

Creates a JWS (JSON Web Signature) using the active Material Key.

**Endpoint**: `POST /service/api/v1/jose/jws/sign`

**Request**:
```json
{
    "kid": "019bd10b-5d65-7bdd-a717-d1f057c85b8a",
    "payload": "eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIn0"
}
```

**Parameters**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `kid` | string | Yes | Elastic JWK key identifier |
| `payload` | string | Yes | Base64url-encoded payload to sign |

**Response** (200 OK):
```json
{
    "jws": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjAxOWJkMTBiLTVkNjUtN2JkZC..."
}
```

---

### Verify

Verifies a JWS signature using the appropriate Material Key.

**Endpoint**: `POST /service/api/v1/jose/jws/verify`

**Request**:
```json
{
    "kid": "019bd10b-5d65-7bdd-a717-d1f057c85b8a",
    "jws": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjAxOWJkMTBiLTVkNjUtN2JkZC..."
}
```

**Response** (200 OK):
```json
{
    "valid": true,
    "payload": "eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIn0",
    "material_kid": "019bd10b-6789-7abc-b123-456789abcdef"
}
```

**Note**: Verification automatically selects the correct historical Material Key based on the JWS header.

---

## JWE Operations

### Encrypt

Creates a JWE (JSON Web Encryption) using the active Material Key.

**Endpoint**: `POST /service/api/v1/jose/jwe/encrypt`

**Request**:
```json
{
    "kid": "019bd10b-5d65-7bdd-a717-d1f057c85b8a",
    "plaintext": "eyJzZWNyZXQiOiJkYXRhIn0="
}
```

**Parameters**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `kid` | string | Yes | Elastic JWK key identifier |
| `plaintext` | string | Yes | Base64url-encoded data to encrypt |

**Response** (200 OK):
```json
{
    "jwe": "eyJhbGciOiJSU0EtT0FFUCIsImVuYyI6IkEyNTZHQ00iLCJraWQiOi..."
}
```

---

### Decrypt

Decrypts a JWE using the appropriate Material Key.

**Endpoint**: `POST /service/api/v1/jose/jwe/decrypt`

**Request**:
```json
{
    "kid": "019bd10b-5d65-7bdd-a717-d1f057c85b8a",
    "jwe": "eyJhbGciOiJSU0EtT0FFUCIsImVuYyI6IkEyNTZHQ00iLCJraWQiOi..."
}
```

**Response** (200 OK):
```json
{
    "plaintext": "eyJzZWNyZXQiOiJkYXRhIn0=",
    "material_kid": "019bd10b-6789-7abc-b123-456789abcdef"
}
```

---

## JWT Operations

### Sign JWT

Creates a signed JWT using the active Material Key.

**Endpoint**: `POST /service/api/v1/jose/jwt/sign`

**Request**:
```json
{
    "kid": "019bd10b-5d65-7bdd-a717-d1f057c85b8a",
    "claims": {
        "sub": "user123",
        "aud": "my-app",
        "exp": 1705330800,
        "iat": 1705327200
    }
}
```

**Response** (200 OK):
```json
{
    "jwt": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IjAxOWJkMTBiLTVkNjUtN2JkZC..."
}
```

---

### Verify JWT

Verifies a JWT signature and validates claims.

**Endpoint**: `POST /service/api/v1/jose/jwt/verify`

**Request**:
```json
{
    "kid": "019bd10b-5d65-7bdd-a717-d1f057c85b8a",
    "jwt": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IjAxOWJkMTBiLTVkNjUtN2JkZC..."
}
```

**Response** (200 OK):
```json
{
    "valid": true,
    "claims": {
        "sub": "user123",
        "aud": "my-app",
        "exp": 1705330800,
        "iat": 1705327200
    }
}
```

---

## JWKS Endpoint

Returns public keys for an Elastic JWK in JWKS format.

**Endpoint**: `GET /service/api/v1/jose/elastic-jwks/{kid}/.well-known/jwks.json`

**Path Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `kid` | string | Elastic JWK key identifier |

**Response** (200 OK):
```json
{
    "keys": [
        {
            "kty": "RSA",
            "kid": "019bd10b-6789-7abc-b123-456789abcdef",
            "use": "sig",
            "alg": "RS256",
            "n": "0vx7agoebGcQSuuPiLJ...",
            "e": "AQAB"
        }
    ]
}
```

**Caching**: Response includes `Cache-Control: max-age=300` (5 minutes).

**Note**: Returns 404 for symmetric keys (oct type) as they have no public component.

---

## Audit Endpoints

### Get Audit Configuration

**Endpoint**: `GET /admin/v1/audit/config`

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `tenant_id` | uuid | Tenant to query |
| `operation` | string | Optional: Filter by operation |

**Response** (200 OK):
```json
{
    "configs": [
        {
            "tenant_id": "019bd10b-5d65-7bdd-a717-d1f057c85b8a",
            "operation": "jws:sign",
            "enabled": true,
            "sampling_rate": 100
        }
    ]
}
```

---

### Set Audit Configuration

**Endpoint**: `POST /admin/v1/audit/config`

**Request**:
```json
{
    "tenant_id": "019bd10b-5d65-7bdd-a717-d1f057c85b8a",
    "operation": "jws:sign",
    "enabled": true,
    "sampling_rate": 100
}
```

**Operations**:
- `jwk:generate`
- `jwk:rotate`
- `jws:sign`
- `jws:verify`
- `jwe:encrypt`
- `jwe:decrypt`
- `jwt:sign`
- `jwt:verify`

---

## Health Endpoints

### Liveness

**Endpoint**: `GET /admin/v1/livez`

**Response** (200 OK):
```json
{
    "status": "ok"
}
```

### Readiness

**Endpoint**: `GET /admin/v1/readyz`

**Response** (200 OK):
```json
{
    "status": "ready",
    "checks": {
        "database": "ok",
        "barrier": "ok"
    }
}
```

---

## Rate Limiting

Service API endpoints are rate-limited per IP address:

| Endpoint Category | Rate Limit | Burst |
|-------------------|------------|-------|
| JWK Operations | 100 req/min | 20 |
| Sign/Verify | 500 req/min | 50 |
| Encrypt/Decrypt | 500 req/min | 50 |
| JWKS Endpoint | 1000 req/min | 100 |

**Response** (429 Too Many Requests):
```json
{
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Retry after 60 seconds.",
    "retry_after": 60
}
```

---

## Error Responses

All errors follow a consistent format:

```json
{
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {},
    "request_id": "019bd10b-5d65-7bdd-a717-d1f057c85b8a"
}
```

### Common Error Codes

| HTTP Status | Code | Description |
|-------------|------|-------------|
| 400 | INVALID_REQUEST | Malformed request body |
| 401 | UNAUTHORIZED | Missing or invalid authentication |
| 403 | FORBIDDEN | Insufficient permissions |
| 404 | NOT_FOUND | Resource not found |
| 409 | CONFLICT | Resource conflict (e.g., max rotations) |
| 422 | VALIDATION_ERROR | Request validation failed |
| 429 | RATE_LIMIT_EXCEEDED | Rate limit exceeded |
| 500 | INTERNAL_ERROR | Internal server error |

---

## SDK Examples

### Go

```go
import "cryptoutil/api/client"

client := client.New("https://jose-ja:8080/service/api/v1/jose")
client.SetBearerToken(accessToken)

// Generate JWK
jwk, err := client.JWK.Generate(&client.GenerateRequest{
    Algorithm: "RS256",
    KeyType:   "RSA",
    KeySize:   2048,
    Use:       "sig",
})

// Sign data
jws, err := client.JWS.Sign(&client.SignRequest{
    KID:     jwk.KID,
    Payload: base64.URLEncode([]byte(`{"sub":"user123"}`)),
})
```

### curl

```bash
# Generate JWK
curl -X POST https://127.0.0.1:8080/service/api/v1/jose/jwk/generate \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"algorithm":"RS256","key_type":"RSA","key_size":2048,"use":"sig"}'

# Sign data
curl -X POST https://127.0.0.1:8080/service/api/v1/jose/jws/sign \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"kid":"019bd10b-5d65-7bdd-a717-d1f057c85b8a","payload":"eyJzdWIiOiJ1c2VyMTIzIn0"}'
```
