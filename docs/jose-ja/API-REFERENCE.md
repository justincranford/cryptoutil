# JOSE-JA API Reference

## Overview

JOSE-JA (JOSE Authority) is a JSON Object Signing and Encryption service that provides:

- **Elastic JWK Management**: Create, list, get, and delete elastic JWK containers
- **Material Key Rotation**: Automatic key material rotation within elastic JWKs
- **JWS Operations**: Sign and verify payloads
- **JWE Operations**: Encrypt and decrypt payloads
- **JWKS Endpoint**: Public key set for verification

## Base URLs

| Endpoint Type | URL | Description |
|---------------|-----|-------------|
| Public API (Service) | `https://localhost:8060/service/api/v1` | Service-to-service API |
| Public API (Browser) | `https://localhost:8060/browser/api/v1` | Browser-based API |
| Admin API | `https://localhost:9092/admin/api/v1` | Administration and health checks |

## Authentication

### Session-Based Authentication

All API endpoints (except JWKS) require session authentication.

**Step 1: Register and Get Session**

```http
POST /service/api/v1/auth/register
Content-Type: application/json

{
  "username": "admin",
  "password": "secure-password-here"
}
```

**Response:**

```json
{
  "session_token": "eyJ...",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "realm_id": "660e8400-e29b-41d4-a716-446655440001",
  "expires_at": "2025-01-15T12:00:00Z"
}
```

**Step 2: Use Session Token**

```http
Authorization: Bearer eyJ...
```

## API Endpoints

### Elastic JWK Management

#### Create Elastic JWK

Creates a new elastic JWK container that holds multiple material keys.

```http
POST /service/api/v1/elastic-jwks
Authorization: Bearer <session_token>
Content-Type: application/json

{
  "algorithm": "RSA/2048",
  "use": "sig",
  "max_materials": 10
}
```

**Supported Algorithms:**

| Algorithm | Description |
|-----------|-------------|
| `RSA/2048` | RSA 2048-bit key |
| `RSA/3072` | RSA 3072-bit key |
| `RSA/4096` | RSA 4096-bit key |
| `EC/P256` | ECDSA P-256 curve |
| `EC/P384` | ECDSA P-384 curve |
| `EC/P521` | ECDSA P-521 curve |
| `EdDSA/Ed25519` | EdDSA Ed25519 curve |

**Response (201 Created):**

```json
{
  "kid": "elastic-key-123",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "kty": "RSA",
  "alg": "RS256",
  "use": "sig",
  "max_materials": 10,
  "current_material_count": 1,
  "created_at": 1704067200
}
```

#### List Elastic JWKs

```http
GET /service/api/v1/elastic-jwks?offset=0&limit=100
Authorization: Bearer <session_token>
```

**Response (200 OK):**

```json
{
  "items": [
    {
      "kid": "elastic-key-123",
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "kty": "RSA",
      "alg": "RS256",
      "use": "sig",
      "max_materials": 10,
      "current_material_count": 3,
      "created_at": 1704067200
    }
  ],
  "total": 1
}
```

#### Get Elastic JWK

```http
GET /service/api/v1/elastic-jwks/{kid}
Authorization: Bearer <session_token>
```

**Response (200 OK):**

```json
{
  "kid": "elastic-key-123",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "kty": "RSA",
  "alg": "RS256",
  "use": "sig",
  "max_materials": 10,
  "current_material_count": 3,
  "created_at": 1704067200
}
```

#### Delete Elastic JWK

```http
DELETE /service/api/v1/elastic-jwks/{kid}
Authorization: Bearer <session_token>
```

**Response (204 No Content)**

### Material Key Management

#### Create Material JWK

Creates a new material key within an elastic JWK container.

```http
POST /service/api/v1/elastic-jwks/{kid}/materials
Authorization: Bearer <session_token>
```

**Response (201 Created):**

```json
{
  "material_kid": "material-key-456",
  "elastic_jwk_id": "elastic-key-123",
  "public_jwk": { ... },
  "active": true,
  "barrier_version": 1,
  "created_at": 1704067200
}
```

#### List Material JWKs

```http
GET /service/api/v1/elastic-jwks/{kid}/materials?offset=0&limit=100
Authorization: Bearer <session_token>
```

**Response (200 OK):**

```json
{
  "items": [
    {
      "material_kid": "material-key-456",
      "elastic_jwk_id": "elastic-key-123",
      "public_jwk": { ... },
      "active": true,
      "barrier_version": 1,
      "created_at": 1704067200
    }
  ],
  "total": 3
}
```

#### Get Active Material JWK

```http
GET /service/api/v1/elastic-jwks/{kid}/materials/active
Authorization: Bearer <session_token>
```

**Response (200 OK):**

```json
{
  "material_kid": "material-key-456",
  "elastic_jwk_id": "elastic-key-123",
  "public_jwk": { ... },
  "active": true,
  "barrier_version": 1,
  "created_at": 1704067200
}
```

#### Rotate Material JWK

Rotates the active material key, creating a new one and retiring the old.

```http
POST /service/api/v1/elastic-jwks/{kid}/rotate
Authorization: Bearer <session_token>
```

**Response (200 OK):**

```json
{
  "material_kid": "material-key-789",
  "elastic_jwk_id": "elastic-key-123",
  "public_jwk": { ... },
  "active": true,
  "barrier_version": 1,
  "created_at": 1704067300,
  "previous_material_kid": "material-key-456"
}
```

### JWKS Endpoint

Returns the JSON Web Key Set containing all public keys. This endpoint does NOT require authentication.

```http
GET /service/api/v1/jwks.json
```

**Alternate Endpoints:**

- `GET /browser/api/v1/jwks.json`
- `GET /.well-known/jwks.json` (Standard well-known endpoint)

**Response (200 OK):**

```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "material-key-456",
      "alg": "RS256",
      "n": "0vx7agoebGcQ...",
      "e": "AQAB"
    }
  ]
}
```

### Cryptographic Operations

#### Sign Payload

```http
POST /service/api/v1/sign
Authorization: Bearer <session_token>
Content-Type: application/json

{
  "kid": "elastic-key-123",
  "payload": "base64-encoded-payload"
}
```

**Response (200 OK):**

```json
{
  "jws": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Verify Signature

```http
POST /service/api/v1/verify
Authorization: Bearer <session_token>
Content-Type: application/json

{
  "jws": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK):**

```json
{
  "valid": true,
  "payload": "base64-decoded-payload",
  "kid": "material-key-456"
}
```

#### Encrypt Payload

```http
POST /service/api/v1/encrypt
Authorization: Bearer <session_token>
Content-Type: application/json

{
  "kid": "elastic-key-123",
  "plaintext": "base64-encoded-plaintext"
}
```

**Response (200 OK):**

```json
{
  "jwe": "eyJhbGciOiJSU0EtT0FFUC0yNTYiLCJlbmMiOiJBMjU2R0NNIn0..."
}
```

#### Decrypt Payload

```http
POST /service/api/v1/decrypt
Authorization: Bearer <session_token>
Content-Type: application/json

{
  "jwe": "eyJhbGciOiJSU0EtT0FFUC0yNTYiLCJlbmMiOiJBMjU2R0NNIn0..."
}
```

**Response (200 OK):**

```json
{
  "plaintext": "base64-decoded-plaintext",
  "kid": "material-key-456"
}
```

### Session Management

#### Issue Session

```http
POST /service/api/v1/sessions/issue
Content-Type: application/json

{
  "username": "admin",
  "password": "secure-password"
}
```

**Response (200 OK):**

```json
{
  "session_token": "eyJ...",
  "expires_at": "2025-01-15T12:00:00Z"
}
```

#### Validate Session

```http
POST /service/api/v1/sessions/validate
Content-Type: application/json

{
  "session_token": "eyJ..."
}
```

**Response (200 OK):**

```json
{
  "valid": true,
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "realm_id": "660e8400-e29b-41d4-a716-446655440001",
  "expires_at": "2025-01-15T12:00:00Z"
}
```

## Admin Endpoints

### Health Check Endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /admin/api/v1/livez` | Liveness probe - Is the process alive? |
| `GET /admin/api/v1/readyz` | Readiness probe - Is the service ready for traffic? |
| `POST /admin/api/v1/shutdown` | Graceful shutdown trigger |

**Liveness Response (200 OK):**

```json
{
  "status": "ok"
}
```

**Readiness Response (200 OK):**

```json
{
  "status": "ok",
  "checks": {
    "database": "ok",
    "barrier": "ok"
  }
}
```

## Error Responses

All error responses follow a consistent format:

```json
{
  "error": "Error description",
  "code": "ERROR_CODE",
  "request_id": "uuid"
}
```

### Common Error Codes

| HTTP Status | Code | Description |
|-------------|------|-------------|
| 400 | `BAD_REQUEST` | Invalid request body or parameters |
| 401 | `UNAUTHORIZED` | Missing or invalid authentication |
| 403 | `FORBIDDEN` | Insufficient permissions |
| 404 | `NOT_FOUND` | Resource not found |
| 409 | `CONFLICT` | Resource already exists |
| 500 | `INTERNAL_ERROR` | Server error |

## Multi-Tenancy

JOSE-JA uses tenant-based data isolation:

- **tenant_id**: Scopes all data (elastic JWKs, material keys, audit logs)
- **realm_id**: Authentication scope only, NOT used for data filtering

When registering, a new tenant is created. The `tenant_id` parameter controls tenant membership:

- **Absent**: Creates new tenant
- **Present**: Joins existing tenant (requires invitation)

## Rate Limiting

Default rate limits:

| Endpoint Type | Limit |
|---------------|-------|
| Public APIs | 100 requests/minute per IP |
| Admin APIs | 10 requests/minute per IP |

## TLS Requirements

All endpoints require HTTPS. TLS 1.3 is recommended.

## Cross-References

- [DEPLOYMENT.md](DEPLOYMENT.md) - Deployment and configuration guide
- [Service Template](../service-template/) - Shared infrastructure patterns
