# cipher-im - Interactive Messaging Service

A demonstration service showcasing cryptoutil's service template patterns,
including dual HTTPS endpoints, health checks, and end-to-end encryption.

## Features

- **Dual HTTPS Servers**: Separate public (8070) and admin (9090) endpoints
- **End-to-End Encryption**: RSA-OAEP message encryption with Ed25519 signing
- **SessionManager Authentication**: JWE/JWS-based secure session management
- **Health Checks**: Kubernetes-style livez/readyz probes
- **Multi-Database**: PostgreSQL or SQLite (in-memory for development)
- **Docker Ready**: Container deployment with health monitoring

## Quick Start

### Local Development

```bash
# Run with in-memory SQLite (default)
go run ./cmd/cipher-im --dev

# Run with PostgreSQL
go run ./cmd/cipher-im \
  --database-url="postgres://user:pass@localhost:5432/cipherdb"

# Run with custom ports
go run ./cmd/cipher-im \
  --public-port=8071 \
  --admin-port=9091
```

### Docker Deployment

```bash
# Build and start container
cd cmd/cipher-im
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f cipher-im

# Health check
curl -k https://localhost:9090/admin/v1/livez

# Stop and clean up
docker compose down
```

### Development with Hot Reload

```bash
# Use development overrides
docker compose -f docker-compose.yml -f docker-compose.dev.yml up
```

## API Endpoints

### Public Server (HTTPS :8070)

**Service APIs** (`/service/api/v1/*` - headless clients):

```bash
# Health check
GET /service/api/v1/health

# User registration
POST /service/api/v1/register
{
  "username": "alice",
  "password": "secure-password"
}

# User login (returns session token)
POST /service/api/v1/users/login
{
  "username": "alice",
  "password": "secure-password"
}

# Send encrypted message
POST /service/api/v1/messages
Authorization: Bearer <session-token>
{
  "receiver_ids": ["user-uuid-1", "user-uuid-2"],
  "content": "Hello, World!"
}

# List received messages
GET /service/api/v1/messages
Authorization: Bearer <session-token>

# Delete message
DELETE /service/api/v1/messages/:id
Authorization: Bearer <session-token>
```

**Browser APIs** (`/browser/api/v1/*` - browser clients):

- Same endpoints as `/service/api/v1/*`
- Additional middleware: CSRF protection, CORS headers, CSP
- Session-based authentication (cookies)

### Admin Server (HTTPS :9090)

```bash
# Liveness probe (lightweight)
GET /admin/v1/livez
# Response: 200 OK (process alive)

# Readiness probe (heavyweight)
GET /admin/v1/readyz
# Response: 200 OK (database connected, dependencies healthy)

# Graceful shutdown
POST /admin/v1/shutdown
# Response: 200 OK (drain connections, close resources)
```

## Configuration

### Command-Line Flags

```bash
--dev                    Development mode (SQLite in-memory, debug logging)
--public-port=8070       Public HTTPS server port
--admin-port=9090        Admin HTTPS server port
--database-url=...       Database connection string
--log-level=info         Logging level (debug, info, warn, error)
```

### Environment Variables

- `DATABASE_URL`: Database connection string
- `LOG_LEVEL`: Logging verbosity

### Database URLs

```bash
# SQLite in-memory (default with --dev)
--database-url="sqlite::memory:"

# SQLite file-based
--database-url="file:/var/lib/cipher-im/data.db"

# PostgreSQL
--database-url="postgres://user:pass@host:5432/dbname?sslmode=disable"
```

## Architecture

### Service Template Pattern

cipher-im demonstrates cryptoutil's reusable service infrastructure:

- **Dual HTTPS Endpoints**: Public (business) + Admin (operations)
- **Health Checks**: Liveness (process alive) vs Readiness (deps healthy)
- **Graceful Shutdown**: Drain connections, close resources, exit cleanly
- **OpenTelemetry**: OTLP export (traces, metrics, logs)
- **Database Abstraction**: PostgreSQL || SQLite with GORM
- **Config Management**: YAML + CLI flags (no environment variables)

### Encryption Flow

1. **User Registration**:
   - Generate Ed25519 keypair (signing)
   - Generate RSA-4096 keypair (encryption)
   - Store private keys server-side (educational demo only)

2. **Send Message**:
   - Encrypt content with receiver's RSA public key (RSA-OAEP)
   - Sign ciphertext with sender's Ed25519 private key
   - Store per-receiver encrypted copies

3. **Receive Message**:
   - Fetch encrypted messages
   - Decrypt with receiver's RSA private key
   - Verify signature with sender's Ed25519 public key

### Database Schema

```sql
-- Users: authentication and cryptographic keys
users (
  id UUID PRIMARY KEY,
  username TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,  -- PBKDF2-HMAC-SHA256
  public_key_rsa TEXT,           -- PEM-encoded RSA public key
  private_key_rsa TEXT,          -- PEM-encoded RSA private key
  public_key_ed25519 TEXT,       -- PEM-encoded Ed25519 public key
  private_key_ed25519 TEXT       -- PEM-encoded Ed25519 private key
)

-- Messages: encrypted message metadata
messages (
  id UUID PRIMARY KEY,
  sender_id UUID REFERENCES users(id),
  created_at TIMESTAMP
)

-- Message Receivers: per-receiver encrypted content
message_receivers (
  id UUID PRIMARY KEY,
  message_id UUID REFERENCES messages(id),
  receiver_id UUID REFERENCES users(id),
  encrypted_content TEXT,  -- Base64-encoded RSA-OAEP ciphertext
  nonce TEXT,              -- Base64-encoded random nonce
  signature TEXT           -- Base64-encoded Ed25519 signature
)
```

## Security Notes

**EDUCATIONAL USE ONLY**: This service stores private keys server-side for
demonstration purposes. Production systems MUST store private keys client-side
only (user devices) and never transmit them to servers.

**Security Features**:

- ✅ HTTPS-only (TLS 1.3+)
- ✅ JWT authentication (Bearer tokens)
- ✅ Password hashing (PBKDF2-HMAC-SHA256, 600k iterations)
- ✅ Rate limiting (per-IP token bucket)
- ✅ CORS protection (browser APIs)
- ✅ CSRF protection (browser APIs)

**Security Limitations**:

- ❌ Server-side private key storage (educational only)
- ❌ No key rotation mechanism
- ❌ No message forward secrecy (use Signal Protocol for production)

## Testing

```bash
# Unit tests
go test ./internal/apps/cipher/im/server/... -v

# E2E tests
go test ./internal/apps/cipher/im/e2e/... -v

# Coverage report
go test ./internal/apps/cipher/im/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Mutation testing
gremlins unleash ./internal/apps/cipher/im/server

# All quality checks
golangci-lint run ./internal/apps/cipher/im/... ./cmd/cipher-im/...
```

## Development

### Project Structure

```
cmd/cipher-im/
├── main.go                    # Application entrypoint
├── Dockerfile                 # Multi-stage container build
├── docker-compose.yml         # Container orchestration
├── docker-compose.dev.yml     # Development overrides
└── .dockerignore              # Build context optimization

internal/apps/cipher/im/
├── server/                    # HTTP server implementation
│   ├── server.go              # Server config and initialization
│   ├── public.go              # Public HTTPS endpoints
│   ├── admin.go               # Admin HTTPS endpoints
│   ├── middleware.go          # JWT authentication
│   └── handlers.go            # Request handlers
├── repository/                # Database access layer
│   └── repository.go          # User/message CRUD
├── crypto/                    # Cryptographic operations
│   ├── rsa.go                 # RSA-OAEP encryption
│   └── ed25519.go             # Ed25519 signing
└── e2e/                       # End-to-end tests
    └── cipher_im_e2e_test.go   # Integration test suite
```

### Adding Features

1. **New Endpoint**: Add handler to `server/handlers.go`
2. **New Middleware**: Add to `server/middleware.go`
3. **New Repository**: Add method to `repository/repository.go`
4. **New Tests**: Add to `server/*_test.go` or `e2e/`

### Code Quality Standards

- **Coverage**: ≥90% for server, ≥85% for crypto (current: 90.5%)
- **Mutation**: ≥85% gremlins score (current: 98.4%)
- **Linting**: golangci-lint clean
- **Tests**: All passing, no skips without tracking

## Deployment

### Production Checklist

- [ ] Replace default JWT secret with strong random value
- [ ] Configure PostgreSQL database (not SQLite)
- [ ] Enable TLS certificate validation (not self-signed)
- [ ] Configure rate limiting per production traffic
- [ ] Set up monitoring (OTLP → otel-collector → Grafana)
- [ ] Configure backups (database, secrets)
- [ ] Review security limitations (server-side keys)

### Kubernetes Deployment

```yaml
apiVersion: v1
kind: Service
metadata:
  name: cipher-im
spec:
  ports:
    - name: public
      port: 8070
      targetPort: 8070
    - name: admin
      port: 9090
      targetPort: 9090
  selector:
    app: cipher-im

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cipher-im
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cipher-im
  template:
    metadata:
      labels:
        app: cipher-im
    spec:
      containers:
        - name: cipher-im
          image: cipher-im:latest
          ports:
            - containerPort: 8070
            - containerPort: 9090
          livenessProbe:
            httpGet:
              path: /admin/v1/livez
              port: 9090
              scheme: HTTPS
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /admin/v1/readyz
              port: 9090
              scheme: HTTPS
            initialDelaySeconds: 5
            periodSeconds: 5
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: cipher-im-secrets
                  key: database-url
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: cipher-im-secrets
                  key: jwt-secret
```

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker compose logs cipher-im

# Common issues:
# - Database connection failed: verify DATABASE_URL
# - Port conflict: check ports 8070/9090 availability
# - Health check failing: verify /admin/v1/livez responds
```

### Database Errors

```bash
# SQLite: permission denied
# Fix: Ensure volume mount has write permissions

# PostgreSQL: connection refused
# Fix: Verify PostgreSQL running and DATABASE_URL correct

# Migration failed
# Fix: Check migration files in internal/learn/repository/migrations/
```

### Authentication Issues

```bash
# Invalid JWT token
# Fix: Ensure JWT_SECRET matches between instances

# Password hash verification failed
# Fix: Re-register user (passwords not migrated)
```

## License

See [LICENSE](../../LICENSE) at repository root.

## Links

- **Main Repository**: [cryptoutil](https://github.com/justincranford/cryptoutil)
- **Documentation**: [docs/](../../docs/)
- **Service Template**: [internal/template/](../../internal/template/)
- **Related Services**: jose-ja, pki-ca, identity-authz
