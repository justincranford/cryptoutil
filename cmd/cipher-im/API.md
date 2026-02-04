# cipher-im API Reference

Complete API documentation for the cipher-im interactive messaging service.

## Base URLs

- **Public Service APIs**: `https://localhost:8070/service/api/v1`
- **Public Browser APIs**: `https://localhost:8070/browser/api/v1`
- **Admin APIs**: `https://localhost:9090/admin/v1`

## Authentication

### Bearer Token (Service APIs)

```http
Authorization: Bearer <session-token>
```

Obtain session token via `/service/api/v1/users/login` endpoint.
Session tokens use JWE (JSON Web Encryption) or JWS (JSON Web Signature) format.

### Session Cookie (Browser APIs)

Browser APIs use session-based authentication with HTTP-only cookies.
Login via `/browser/api/v1/users/login` sets session cookie automatically.

## Public APIs

### Health Check

**GET** `/service/api/v1/health`

Check service availability.

**Request**: None

**Response** (200 OK):

```json
{
  "status": "ok",
  "service": "cipher-im",
  "version": "1.0.0"
}
```

### User Registration

**POST** `/service/api/v1/register`

Create new user account with cryptographic keypairs.

**Request Body**:

```json
{
  "username": "alice",
  "password": "secure-password-123"
}
```

**Response** (201 Created):

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "alice",
  "public_key_rsa": "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkq...",
  "public_key_ed25519": "-----BEGIN PUBLIC KEY-----\nMCowBQYDK2Vw..."
}
```

**Errors**:

- `400 Bad Request`: Invalid username/password format
- `409 Conflict`: Username already exists

**Notes**:

- Username: 3-32 chars, alphanumeric + underscore
- Password: 8+ chars, min 1 uppercase, 1 lowercase, 1 digit
- Generates RSA-4096 keypair (encryption)
- Generates Ed25519 keypair (signing)
- Private keys stored server-side (educational demo only)

### User Login

**POST** `/service/api/v1/users/login`

Authenticate user and obtain session token.

**Request Body**:

```json
{
  "username": "alice",
  "password": "secure-password-123"
}
```

**Response** (200 OK):

```json
{
  "token": "<session-token>",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

**Errors**:

- `400 Bad Request`: Missing username or password
- `401 Unauthorized`: Invalid credentials

**Notes**:

- Token valid for 1 hour
- Use token in `Authorization: Bearer <token>` header
- Session token format: JWE (encrypted) or JWS (signed)

### Send Message

**POST** `/service/api/v1/messages`

Send encrypted message to one or more receivers.

**Headers**:

```http
Authorization: Bearer <session-token>
Content-Type: application/json
```

**Request Body**:

```json
{
  "receiver_ids": [
    "550e8400-e29b-41d4-a716-446655440001",
    "550e8400-e29b-41d4-a716-446655440002"
  ],
  "content": "Hello, World! This is a secret message."
}
```

**Response** (201 Created):

```json
{
  "message_id": "650e8400-e29b-41d4-a716-446655440003",
  "sender_id": "550e8400-e29b-41d4-a716-446655440000",
  "receiver_count": 2,
  "created_at": "2025-12-24T10:30:00Z"
}
```

**Errors**:

- `400 Bad Request`: Invalid receiver IDs or empty content
- `401 Unauthorized`: Missing or invalid session token
- `404 Not Found`: One or more receiver IDs do not exist

**Notes**:

- Content encrypted per-receiver using RSA-OAEP
- Each receiver gets separate encrypted copy
- Signature created using sender's Ed25519 private key
- Max content length: 1000 characters (plaintext)

### List Messages

**GET** `/service/api/v1/messages`

List all messages received by authenticated user.

**Headers**:

```http
Authorization: Bearer <session-token>
```

**Query Parameters**:

- `page`: Page number (default: 1)
- `size`: Page size (default: 20, max: 100)

**Response** (200 OK):

```json
{
  "messages": [
    {
      "id": "650e8400-e29b-41d4-a716-446655440003",
      "sender_id": "550e8400-e29b-41d4-a716-446655440000",
      "sender_username": "alice",
      "content": "Hello, World! This is a secret message.",
      "created_at": "2025-12-24T10:30:00Z",
      "signature": "MEQCIHx7...",
      "signature_valid": true
    }
  ],
  "pagination": {
    "page": 1,
    "size": 20,
    "total": 1
  }
}
```

**Errors**:

- `401 Unauthorized`: Missing or invalid session token

**Notes**:

- Content decrypted using receiver's RSA private key
- Signature verified using sender's Ed25519 public key
- Messages sorted by created_at descending (newest first)

### Delete Message

**DELETE** `/service/api/v1/messages/:id`

Delete specific message (receiver copy only).

**Headers**:

```http
Authorization: Bearer <session-token>
```

**Path Parameters**:

- `id`: Message ID (UUID)

**Response** (204 No Content):

No response body.

**Errors**:

- `401 Unauthorized`: Missing or invalid session token
- `403 Forbidden`: User is not the receiver of this message
- `404 Not Found`: Message does not exist

**Notes**:

- Deletes receiver's copy only (soft delete)
- Other receivers' copies unaffected
- Sender can delete only if also a receiver

## Admin APIs

### Liveness Probe

**GET** `/admin/v1/livez`

Lightweight health check (process alive?).

**Response** (200 OK):

```json
{
  "status": "ok",
  "timestamp": "2025-12-24T10:30:00Z"
}
```

**Response** (503 Service Unavailable):

```json
{
  "status": "error",
  "message": "Process not responding"
}
```

**Notes**:

- Kubernetes liveness probe endpoint
- Failure action: Restart container
- Check frequency: Every 10 seconds

### Readiness Probe

**GET** `/admin/v1/readyz`

Heavyweight health check (dependencies healthy?).

**Response** (200 OK):

```json
{
  "status": "ok",
  "checks": {
    "database": "ok",
    "dependencies": "ok"
  },
  "timestamp": "2025-12-24T10:30:00Z"
}
```

**Response** (503 Service Unavailable):

```json
{
  "status": "error",
  "checks": {
    "database": "error",
    "dependencies": "ok"
  },
  "message": "Database connection failed",
  "timestamp": "2025-12-24T10:30:00Z"
}
```

**Notes**:

- Kubernetes readiness probe endpoint
- Checks database connectivity
- Failure action: Remove from load balancer (do NOT restart)
- Check frequency: Every 5 seconds

### Graceful Shutdown

**POST** `/admin/v1/shutdown`

Trigger graceful shutdown sequence.

**Response** (200 OK):

```json
{
  "status": "shutting_down",
  "message": "Graceful shutdown initiated"
}
```

**Notes**:

- Stops accepting new requests
- Drains active requests (max 30 seconds)
- Closes database connections
- Releases resources
- Exits process

### Barrier Key Rotation

**POST** `/admin/api/v1/barrier/rotate/root`

Manually rotate root encryption key.

**Request Body**:

```json
{
  "reason": "Scheduled annual rotation per security policy"
}
```

**Response** (200 OK):

```json
{
  "old_key_uuid": "019b7c3e-3457-7823-a023-ba80e542d714",
  "new_key_uuid": "019b7c3e-4567-7824-b045-ce91f7654321",
  "reason": "Scheduled annual rotation per security policy",
  "rotated_at": 1767316010
}
```

**Errors**:

- `400 Bad Request`: Missing or invalid reason (10-500 chars required)
- `500 Internal Server Error`: Rotation failed

**Notes**:

- Root key encrypts intermediate keys
- Old messages remain decryptable with old key
- New messages encrypted with new key
- Reason field mandatory (audit trail)

**POST** `/admin/api/v1/barrier/rotate/intermediate`

Manually rotate intermediate encryption key.

**Request Body**:

```json
{
  "reason": "Quarterly rotation per compliance requirements"
}
```

**Response** (200 OK):

```json
{
  "old_key_uuid": "019b7c3e-3457-7824-af6c-907af8511c23",
  "new_key_uuid": "019b7c3e-4678-7825-c156-df02e8765432",
  "reason": "Quarterly rotation per compliance requirements",
  "rotated_at": 1767316110
}
```

**POST** `/admin/api/v1/barrier/rotate/content`

Manually rotate content encryption key.

**Request Body**:

```json
{
  "reason": "Hourly rotation for forward secrecy"
}
```

**Response** (200 OK):

```json
{
  "new_key_uuid": "019b7c3e-4789-7826-d267-ef13f9876543",
  "reason": "Hourly rotation for forward secrecy",
  "rotated_at": 1767316210
}
```

**Notes**:

- Content key encrypts actual message content
- Elastic rotation: no old_key_uuid (old keys retained for decryption)
- New messages use new key immediately
- Old messages decrypt with original content key

### Barrier Key Status

**GET** `/admin/api/v1/barrier/keys/status`

Get current barrier encryption keys status.

**Response** (200 OK):

```json
{
  "root_key": {
    "uuid": "019b7c3e-3457-7823-a023-ba80e542d714",
    "created_at": 1767316010301,
    "updated_at": 1767316010301
  },
  "intermediate_key": {
    "uuid": "019b7c3e-3457-7824-af6c-907af8511c23",
    "created_at": 1767316010313,
    "updated_at": 1767316010313
  }
}
```

**Notes**:

- Returns latest root and intermediate keys
- Content keys NOT included (elastic rotation - no "latest" concept)
- Timestamps in Unix milliseconds
- Used for monitoring key age

## Error Responses

All error responses follow consistent format:

```json
{
  "code": "ERR_INVALID_REQUEST",
  "message": "Human-readable error message",
  "details": {
    "field": "username",
    "reason": "Username already exists"
  },
  "request_id": "750e8400-e29b-41d4-a716-446655440004"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `ERR_INVALID_REQUEST` | 400 | Malformed request or validation error |
| `ERR_UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `ERR_FORBIDDEN` | 403 | Valid auth but insufficient permissions |
| `ERR_NOT_FOUND` | 404 | Resource does not exist |
| `ERR_CONFLICT` | 409 | Duplicate resource or state conflict |
| `ERR_INTERNAL` | 500 | Unhandled server error |
| `ERR_UNAVAILABLE` | 503 | Temporary unavailability |

## Rate Limiting

Public APIs enforce rate limiting per IP address:

- **Public APIs**: 100 requests/minute (burst: 20)
- **Admin APIs**: 10 requests/minute (burst: 5)
- **Login endpoint**: 5 requests/minute (burst: 2)

**Rate Limit Headers**:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 2025-12-24T10:31:00Z
```

**Rate Limit Exceeded** (429 Too Many Requests):

```json
{
  "code": "ERR_RATE_LIMIT",
  "message": "Rate limit exceeded",
  "details": {
    "limit": 100,
    "reset_at": "2025-12-24T10:31:00Z"
  }
}
```

## CORS Policy

Browser APIs (`/browser/**`) support CORS with specific origins:

**Allowed Origins**:

- `http://localhost:8070`
- `https://localhost:8070`
- `http://127.0.0.1:8070`
- `https://127.0.0.1:8070`

**CORS Headers**:

```http
Access-Control-Allow-Origin: https://localhost:8070
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 3600
```

**Service APIs** (`/service/**`) do NOT support CORS (headless clients only).

## Examples

### Complete Registration → Login → Send → Receive Flow

```bash
# 1. Register Alice
curl -k -X POST https://localhost:8070/service/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"SecurePass123"}'
# Response: {"id":"<alice-id>","username":"alice",...}

# 2. Register Bob
curl -k -X POST https://localhost:8070/service/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"username":"bob","password":"SecurePass456"}'
# Response: {"id":"<bob-id>","username":"bob",...}

# 3. Alice logs in
curl -k -X POST https://localhost:8070/service/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"SecurePass123"}'
# Response: {"token":"<alice-session-token>","expires_in":3600}

# 4. Alice sends message to Bob
curl -k -X POST https://localhost:8070/service/api/v1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <alice-session-token>" \
  -d '{"receiver_ids":["<bob-id>"],"content":"Hello Bob!"}'
# Response: {"message_id":"<msg-id>","receiver_count":1}

# 5. Bob logs in
curl -k -X POST https://localhost:8070/service/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"username":"bob","password":"SecurePass456"}'
# Response: {"token":"<bob-session-token>","expires_in":3600}

# 6. Bob retrieves messages
curl -k -X GET https://localhost:8070/service/api/v1/messages \
  -H "Authorization: Bearer <bob-session-token>"
# Response: {"messages":[{"content":"Hello Bob!","sender_username":"alice"}]}

# 7. Bob deletes message
curl -k -X DELETE https://localhost:8070/service/api/v1/messages/<msg-id> \
  -H "Authorization: Bearer <bob-session-token>"
# Response: 204 No Content
```

### Health Check Monitoring

```bash
# Liveness check (lightweight)
curl -k https://localhost:9090/admin/v1/livez
# Response: {"status":"ok"}

# Readiness check (heavyweight)
curl -k https://localhost:9090/admin/v1/readyz
# Response: {"status":"ok","checks":{"database":"ok"}}

# Continuous monitoring
watch -n 5 'curl -sk https://localhost:9090/admin/v1/livez'
```

## See Also

- [README.md](README.md): Quick start and deployment
- [ENCRYPTION.md](ENCRYPTION.md): Cryptographic architecture
- [TUTORIAL.md](TUTORIAL.md): Step-by-step user guide
