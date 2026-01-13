# Cipher-IM Testing Guide

## Summary

This guide documents manual testing commands for the cipher-im service after successful User model consolidation and migration fix.

## Test Status

✅ **COMPLETED**:

- User model consolidation (deleted cipher-im User, using template User)
- SQL migrations updated (users, browser_sessions, service_sessions with tenant_id)
- Session manager interface fixed (method signature matching)
- **ALL 10 E2E TESTS PASSING** (verified):
  - TestE2E_RotateRootKey
  - TestE2E_RotateIntermediateKey
  - TestE2E_RotateContentKey
  - TestE2E_GetBarrierKeysStatus
  - TestE2E_FullEncryptionFlow
  - TestE2E_MultiReceiverEncryption
  - TestE2E_MessageDeletion
  - TestE2E_BrowserFullEncryptionFlow
  - TestE2E_BrowserMultiReceiverEncryption
  - TestE2E_BrowserMessageDeletion

✅ **Docker Compose Infrastructure**:

- All 5 containers running and healthy (verified 2026-01-11T04:48)
- Health checks passing for all instances
- Container logs show successful initialization

---

## Docker Compose Testing

### 1. Start the Stack

```powershell
cd cmd\cipher-im
docker compose up -d
```

Expected output:

```
[+] Running 9/9
 ✔ Network cipher-im-network           Created
 ✔ Volume "cipher-im_postgres_data"    Created
 ✔ Volume "cipher-im_grafana_data"     Created
 ✔ Container cipher-im-postgres        Started
 ✔ Container cipher-im-otel-collector  Started
 ✔ Container cipher-im-grafana         Started
 ✔ Container cipher-im-sqlite          Started
 ✔ Container cipher-im-pg-1            Started
 ✔ Container cipher-im-pg-2            Started
```

### 2. Wait for Health Checks (30 seconds minimum)

```powershell
Start-Sleep -Seconds 30
docker compose ps
```

Expected output (all healthy):

```
NAME                        STATUS
cipher-im-grafana          Up (healthy)
cipher-im-otel-collector   Up
cipher-im-pg-1             Up (healthy)
cipher-im-pg-2             Up (healthy)
cipher-im-postgres         Up (healthy)
cipher-im-sqlite           Up (healthy)
```

### 3. View Container Logs

```powershell
# SQLite instance logs
docker compose logs --tail=50 cipher-im-sqlite

# PostgreSQL instance 1 logs
docker compose logs --tail=50 cipher-im-pg-1

# PostgreSQL instance 2 logs
docker compose logs --tail=50 cipher-im-pg-2

# Database logs
docker compose logs --tail=50 cipher-im-postgres

# Follow all logs in real-time
docker compose logs -f
```

Expected log patterns:

```
time=2026-01-11T04:47:31.740Z level=INFO msg="database connection established successfully"
DEBUG initializeFirstRootJWK: Successfully created first root JWK
DEBUG initializeFirstIntermediateJWK: Successfully created first intermediate JWK
🚀 Starting cipher-im service...
   Public Server: https://127.0.0.1:8888
   Admin Server:  https://127.0.0.1:9090
```

### 4. Stop the Stack

```powershell
cd cmd\cipher-im
docker compose down
```

### 5. Stop and Remove Volumes

```powershell
cd cmd\cipher-im
docker compose down -v
```

---

## Service Endpoints

### cipher-im-sqlite (In-Memory SQLite)

- **Public API**: <https://127.0.0.1:8888>
- **Admin API**: <https://127.0.0.1:9090>
- **Database**: SQLite in-memory (`file::memory:?cache=shared`)

### cipher-im-pg-1 (PostgreSQL Instance 1)

- **Public API**: <https://127.0.0.1:8889>
- **Admin API**: <https://127.0.0.1:9091>
- **Database**: Shared PostgreSQL `cipher_im` database

### cipher-im-pg-2 (PostgreSQL Instance 2)

- **Public API**: <https://127.0.0.1:8890>
- **Admin API**: <https://127.0.0.1:9092>
- **Database**: Shared PostgreSQL `cipher_im` database

### Supporting Services

- **PostgreSQL Database**: localhost:5432
- **Grafana OTEL LGTM**: <http://localhost:3000>
- **OpenTelemetry Collector**: Internal only (4317, 4318)

---

## API Testing with curl

**Note**: Use `curl.exe` on Windows (not PowerShell Invoke-WebRequest which requires PowerShell Core 7+ for `-SkipCertificateCheck`)

### Health Check Endpoints

```powershell
# Liveness probe (cipher-im-sqlite)
curl.exe -k https://127.0.0.1:9090/admin/v1/livez

# Readiness probe (cipher-im-sqlite)
curl.exe -k https://127.0.0.1:9090/admin/v1/readyz

# Test all three instances
curl.exe -k https://127.0.0.1:9090/admin/v1/livez  # sqlite
curl.exe -k https://127.0.0.1:9091/admin/v1/livez  # pg-1
curl.exe -k https://127.0.0.1:9092/admin/v1/livez  # pg-2
```

Expected response (HTTP 200):

```json
{"status":"ok"}
```

### User Registration

```powershell
# Register user on cipher-im-sqlite
curl.exe -k -X POST https://127.0.0.1:8888/api/v1/register `
  -H "Content-Type: application/json" `
  -d "{\"username\":\"testuser1\",\"password\":\"TestPass123!\"}"

# Register on pg-1
curl.exe -k -X POST https://127.0.0.1:8889/api/v1/register `
  -H "Content-Type: application/json" `
  -d "{\"username\":\"testuser2\",\"password\":\"TestPass123!\"}"

# Register on pg-2
curl.exe -k -X POST https://127.0.0.1:8890/api/v1/register `
  -H "Content-Type: application/json" `
  -d "{\"username\":\"testuser3\",\"password\":\"TestPass123!\"}"
```

Expected response (HTTP 201):

```json
{
  "user_id": "01234567-89ab-cdef-0123-456789abcdef",
  "username": "testuser1"
}
```

### User Login

```powershell
# Login on cipher-im-sqlite
curl.exe -k -X POST https://127.0.0.1:8888/api/v1/login `
  -H "Content-Type: application/json" `
  -d "{\"username\":\"testuser1\",\"password\":\"TestPass123!\"}"
```

Expected response (HTTP 200):

```json
{
  "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-01-11T12:00:00Z"
}
```

**Save the `session_token` for subsequent requests**

### Send Message

```powershell
# Replace <SESSION_TOKEN> and <RECIPIENT_USER_ID> with actual values
curl.exe -k -X POST https://127.0.0.1:8888/api/v1/messages `
  -H "Authorization: Bearer <SESSION_TOKEN>" `
  -H "Content-Type: application/json" `
  -d "{\"recipient_ids\":[\"<RECIPIENT_USER_ID>\"],\"plaintext\":\"Hello World\"}"
```

Expected response (HTTP 201):

```json
{
  "message_id": "01234567-89ab-cdef-0123-456789abcdef",
  "sent_at": "2026-01-11T04:48:00Z"
}
```

### Receive Messages

```powershell
# Get messages for authenticated user
curl.exe -k -X GET https://127.0.0.1:8888/api/v1/messages `
  -H "Authorization: Bearer <SESSION_TOKEN>"
```

Expected response (HTTP 200):

```json
{
  "messages": [
    {
      "message_id": "01234567-89ab-cdef-0123-456789abcdef",
      "sender_id": "sender-user-id",
      "sent_at": "2026-01-11T04:48:00Z",
      "plaintext": "Hello World"
    }
  ]
}
```

### Delete Message

```powershell
# Replace <MESSAGE_ID> with actual message ID
curl.exe -k -X DELETE https://127.0.0.1:8888/api/v1/messages/<MESSAGE_ID> `
  -H "Authorization: Bearer <SESSION_TOKEN>"
```

Expected response (HTTP 204 No Content)

---

## Full Test Flow Example

```powershell
# 1. Start containers
cd cmd\cipher-im
docker compose up -d
Start-Sleep -Seconds 30

# 2. Verify health
curl.exe -k https://127.0.0.1:9090/admin/v1/livez

# 3. Register two users
curl.exe -k -X POST https://127.0.0.1:8888/api/v1/register `
  -H "Content-Type: application/json" `
  -d "{\"username\":\"alice\",\"password\":\"AlicePass123!\"}"

curl.exe -k -X POST https://127.0.0.1:8888/api/v1/register `
  -H "Content-Type: application/json" `
  -d "{\"username\":\"bob\",\"password\":\"BobPass123!\"}"

# 4. Login as Alice (save token)
$response = curl.exe -k -X POST https://127.0.0.1:8888/api/v1/login `
  -H "Content-Type: application/json" `
  -d "{\"username\":\"alice\",\"password\":\"AlicePass123!\"}"
# Extract session_token from $response JSON

# 5. Send message from Alice to Bob
curl.exe -k -X POST https://127.0.0.1:8888/api/v1/messages `
  -H "Authorization: Bearer <ALICE_SESSION_TOKEN>" `
  -H "Content-Type: application/json" `
  -d "{\"recipient_ids\":[\"<BOB_USER_ID>\"],\"plaintext\":\"Hello Bob!\"}"

# 6. Login as Bob
$bobResponse = curl.exe -k -X POST https://127.0.0.1:8888/api/v1/login `
  -H "Content-Type: application/json" `
  -d "{\"username\":\"bob\",\"password\":\"BobPass123!\"}"

# 7. Receive messages as Bob
curl.exe -k -X GET https://127.0.0.1:8888/api/v1/messages `
  -H "Authorization: Bearer <BOB_SESSION_TOKEN>"

# 8. Cleanup
docker compose down -v
```

---

## E2E Test Execution

### Run All E2E Tests

```powershell
cd c:\Dev\Projects\cryptoutil
go test ./internal/apps/cipher/im/e2e -v -count=1
```

Expected output:

```
=== RUN   TestE2E_RotateRootKey
--- PASS: TestE2E_RotateRootKey (0.15s)
=== RUN   TestE2E_RotateIntermediateKey
--- PASS: TestE2E_RotateIntermediateKey (0.12s)
=== RUN   TestE2E_RotateContentKey
--- PASS: TestE2E_RotateContentKey (0.11s)
=== RUN   TestE2E_GetBarrierKeysStatus
--- PASS: TestE2E_GetBarrierKeysStatus (0.08s)
=== RUN   TestE2E_FullEncryptionFlow
--- PASS: TestE2E_FullEncryptionFlow (0.45s)
=== RUN   TestE2E_MultiReceiverEncryption
--- PASS: TestE2E_MultiReceiverEncryption (0.52s)
=== RUN   TestE2E_MessageDeletion
--- PASS: TestE2E_MessageDeletion (0.38s)
=== RUN   TestE2E_BrowserFullEncryptionFlow
--- PASS: TestE2E_BrowserFullEncryptionFlow (0.41s)
=== RUN   TestE2E_BrowserMultiReceiverEncryption
--- PASS: TestE2E_BrowserMultiReceiverEncryption (0.49s)
=== RUN   TestE2E_BrowserMessageDeletion
--- PASS: TestE2E_BrowserMessageDeletion (0.39s)
PASS
ok      cryptoutil/internal/apps/cipher/im/e2e  3.110s
```

### Run Individual Test

```powershell
go test ./internal/apps/cipher/im/e2e -v -run TestE2E_FullEncryptionFlow
```

---

## Troubleshooting

### Container Won't Start

```powershell
# Check logs for specific container
docker compose logs cipher-im-sqlite

# Check all container statuses
docker compose ps

# Restart specific container
docker compose restart cipher-im-sqlite
```

### Health Check Failing

```powershell
# Exec into container to test health endpoint
docker compose exec cipher-im-sqlite wget --no-check-certificate -q -O - https://127.0.0.1:9090/admin/v1/livez

# Check if server is listening
docker compose exec cipher-im-sqlite netstat -tuln | findstr 9090
```

### Database Issues

```powershell
# Check PostgreSQL logs
docker compose logs cipher-im-postgres

# Connect to PostgreSQL
docker compose exec cipher-im-postgres psql -U cipher_user -d cipher_im

# Check tables
\dt

# Check users
SELECT id, username, tenant_id FROM users;
```

### Certificate Issues

If you get SSL/TLS certificate errors:

- Use `curl.exe -k` flag (insecure, for development only)
- Do NOT use PowerShell `Invoke-WebRequest` without PowerShell Core 7+ (lacks `-SkipCertificateCheck`)
- Containers use auto-generated self-signed certificates in dev mode

### PowerShell Version Issues

Windows PowerShell 5.1 does NOT support `-SkipCertificateCheck` parameter.

**Solutions**:

1. Use `curl.exe` with `-k` flag (recommended)
2. Upgrade to PowerShell Core 7+: `winget install Microsoft.PowerShell`
3. Use certificate validation workaround (not recommended)

---

## Code Coverage

### Run Coverage for cipher-im Packages

```powershell
# Main package
go test ./internal/apps/cipher/im -coverprofile=test-output/coverage_im_main.out -covermode=atomic

# Repository package
go test ./internal/apps/cipher/im/repository -coverprofile=test-output/coverage_im_repo.out -covermode=atomic

# Server package
go test ./internal/apps/cipher/im/server -coverprofile=test-output/coverage_im_server.out -covermode=atomic

# View coverage report
go tool cover -func=test-output/coverage_im_main.out
```

**Note**: Running full test suite with coverage may cause memory exhaustion. Use individual package coverage as shown above.

---

## OpenAPI/Swagger Documentation

If available, Swagger UI should be accessible at:

- <https://127.0.0.1:8888/swagger/index.html> (cipher-im-sqlite)
- <https://127.0.0.1:8889/swagger/index.html> (cipher-im-pg-1)
- <https://127.0.0.1:8890/swagger/index.html> (cipher-im-pg-2)

**Note**: UI files not found in workspace - may not be implemented yet.

---

## Summary of Verified Features

✅ **User Management**:

- User registration with username/password
- User login with session token issuance
- Template User model with tenant_id support

✅ **Session Management**:

- Browser session creation (JWT with HS256)
- Service session creation (JWT with HS256)
- Tenant-aware session tables

✅ **Message Encryption**:

- Multi-receiver message encryption (JWE)
- Message decryption with content keys
- Message deletion

✅ **Barrier Key Management**:

- Root JWK initialization
- Intermediate JWK initialization
- Key rotation (root, intermediate, content)
- Key status retrieval

✅ **Infrastructure**:

- SQLite in-memory deployment
- PostgreSQL multi-instance deployment
- Health check endpoints (livez, readyz)
- OpenTelemetry integration
- Grafana OTEL LGTM observability stack

---

## Contact/Support

For issues or questions, see:

- E2E test source: `internal/apps/cipher/im/e2e/`
- Docker Compose config: `cmd/cipher-im/docker-compose.yml`
- Application config: `configs/cipher/cipher-im-config.yml`
- Migration files: `internal/apps/cipher/im/repository/migrations/`

---

**Last Updated**: 2026-01-11
**Verified By**: Automated E2E test suite + manual Docker Compose deployment
