# API Architecture

## Overview

cryptoutil implements a sophisticated dual-context API architecture that separates browser-based clients from service-to-service communication while maintaining a comprehensive security model.

## Context Path Hierarchy

```
cryptoutil Server Applications
│
├── 🌐 Public Fiber App (Port 8080 - HTTPS)
│   │
│   ├── 📋 Swagger UI Routes
│   │   ├── GET /ui/swagger/doc.json              # OpenAPI spec JSON
│   │   └── GET /ui/swagger/*                     # Swagger UI interface
│   │
│   ├── 🔒 CSRF Token Route  
│   │   └── GET /browser/api/v1/csrf-token        # Get CSRF token for browser clients
│   │
│   ├── 🌐 Browser API Context (/browser/api/v1)  # For browser clients with CORS/CSRF
│   │   ├── POST   /browser/api/v1/elastickey           # Create elastic key
│   │   ├── GET    /browser/api/v1/elastickey/{id}      # Get elastic key by ID
│   │   ├── GET    /browser/api/v1/elastickeys          # Find elastic keys (filtered)
│   │   ├── PUT    /browser/api/v1/elastickey/{id}      # Update elastic key
│   │   ├── DELETE /browser/api/v1/elastickey/{id}      # Delete elastic key
│   │   ├── POST   /browser/api/v1/materialkey          # Create material key
│   │   ├── GET    /browser/api/v1/materialkey/{id}     # Get material key by ID
│   │   ├── GET    /browser/api/v1/materialkeys         # Find material keys (filtered)
│   │   ├── PUT    /browser/api/v1/materialkey/{id}     # Update material key
│   │   ├── DELETE /browser/api/v1/materialkey/{id}     # Delete material key
│   │   ├── POST   /browser/api/v1/crypto/encrypt       # Encrypt operation
│   │   ├── POST   /browser/api/v1/crypto/decrypt       # Decrypt operation
│   │   ├── POST   /browser/api/v1/crypto/sign          # Sign operation
│   │   ├── POST   /browser/api/v1/crypto/verify        # Verify operation
│   │   └── POST   /browser/api/v1/crypto/generate      # Generate operation
│   │
│   └── 🔧 Service API Context (/service/api/v1)  # For service clients without browser middleware
│       ├── POST   /service/api/v1/elastickey           # Create elastic key
│       ├── GET    /service/api/v1/elastickey/{id}      # Get elastic key by ID
│       ├── GET    /service/api/v1/elastickeys          # Find elastic keys (filtered)
│       ├── PUT    /service/api/v1/elastickey/{id}      # Update elastic key
│       ├── DELETE /service/api/v1/elastickey/{id}      # Delete elastic key
│       ├── POST   /service/api/v1/materialkey          # Create material key
│       ├── GET    /service/api/v1/materialkey/{id}     # Get material key by ID
│       ├── GET    /service/api/v1/materialkeys         # Find material keys (filtered)
│       ├── PUT    /service/api/v1/materialkey/{id}     # Update material key
│       ├── DELETE /service/api/v1/materialkey/{id}     # Delete material key
│       ├── POST   /service/api/v1/crypto/encrypt       # Encrypt operation
│       ├── POST   /service/api/v1/crypto/decrypt       # Decrypt operation
│       ├── POST   /service/api/v1/crypto/sign          # Sign operation
│       ├── POST   /service/api/v1/crypto/verify        # Verify operation
│       └── POST   /service/api/v1/crypto/generate      # Generate operation
│
└── 🔐 Private Fiber App (Port 9090 - HTTP)
    ├── 🩺 Health Check Routes
    │   ├── GET  /livez                              # Liveness probe (Kubernetes)
    │   └── GET  /readyz                             # Readiness probe (Kubernetes)
    │
    └── 🛑 Management Routes
        └── POST /shutdown                           # Graceful shutdown endpoint
```

## API Context Design

### Browser API Context (`/browser/api/v1/*`)

**Purpose**: Designed for web applications and browser-based clients that need full CORS and CSRF protection.

**Security Features**:
- CORS headers for cross-origin requests
- CSRF token validation for state-changing operations
- Content Security Policy (CSP) headers
- XSS protection headers
- Secure cookie handling

**Usage Example**:
```javascript
// 1. Get CSRF token
const response = await fetch('/browser/api/v1/csrf-token', {
  credentials: 'same-origin'
});
const csrfData = await response.json();

// 2. Use token in subsequent requests
await fetch('/browser/api/v1/elastickey', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-CSRF-Token': getCsrfTokenFromCookie() // Handled by Swagger UI script
  },
  credentials: 'same-origin',
  body: JSON.stringify({
    name: 'my-key',
    algorithm: 'RSA',
    provider: 'CRYPTOUTIL'
  })
});
```

### Service API Context (`/service/api/v1/*`)

**Purpose**: Optimized for service-to-service communication without browser-specific overhead.

**Security Features**:
- IP allowlisting and rate limiting
- OpenAPI request/response validation
- Structured logging and telemetry
- No CORS or CSRF overhead

**Usage Example**:
```bash
# Direct service-to-service communication
curl -X POST http://localhost:8080/service/api/v1/elastickey \
  -H "Content-Type: application/json" \
  -d '{
    "name": "service-key",
    "algorithm": "AES",
    "provider": "CRYPTOUTIL"
  }'
```

### Management Interface (`localhost:9090`)

**Purpose**: Private administrative and monitoring interface.

**Features**:
- Health check endpoints for Kubernetes probes
- Graceful shutdown endpoint
- Internal monitoring and debugging
- Separate network interface for security

**Usage Example**:
```bash
# Health checks
curl http://localhost:9090/livez   # Returns 200 if alive
curl http://localhost:9090/readyz  # Returns 200 if ready

# Graceful shutdown
curl -X POST http://localhost:9090/shutdown
```

## Middleware Stack

### Request Flow through Middleware

```
Request Flow through Middleware Stack:
│
├── 🛡️ Common Middlewares (Both Public Contexts)
│   ├── Recover (panic recovery)
│   ├── Request ID generation
│   ├── Basic logging
│   ├── OpenTelemetry tracing
│   ├── Request logger (structured)
│   ├── IP filtering (allowlist)
│   ├── Rate limiting (per IP)
│   └── Cache control headers
│
├── 🌐 Browser Context Additional Middlewares
│   ├── CORS (browser support)
│   ├── XSS protection (Content Security Policy)
│   ├── Security headers (Helmet-style)
│   ├── CSRF protection (browser requests only)
│   └── OpenAPI request validation
│
├── 🔧 Service Context Additional Middlewares
│   └── OpenAPI request validation
│
└── 🔐 Private App Middlewares
    ├── Basic common middlewares
    └── Health check endpoints only
```

## OpenAPI Integration

### Code Generation

The API is driven by OpenAPI 3.0.3 specifications:

- **Components**: `internal/openapi/openapi_spec_components.yaml`
- **Paths**: `internal/openapi/openapi_spec_paths.yaml`
- **Generated Code**:
  - Models: `internal/openapi/model/`
  - Server handlers: `internal/openapi/server/`
  - Go client: `internal/openapi/client/`

### Swagger UI Integration

The Swagger UI includes sophisticated CSRF token handling:

```javascript
// Automatic CSRF token injection for browser API calls
window.fetch = function(url, options) {
  if (url.includes('/browser/api/v1/') && options.method !== 'GET') {
    options.headers = options.headers || {};
    options.headers['X-CSRF-Token'] = getCsrfTokenFromCookie();
  }
  return originalFetch.call(this, url, options);
};
```

## API Resources

### Elastic Keys

Logical key containers with metadata and policies:

- `POST /elastickey` - Create new elastic key
- `GET /elastickey/{id}` - Retrieve elastic key
- `PUT /elastickey/{id}` - Update elastic key
- `DELETE /elastickey/{id}` - Delete elastic key
- `GET /elastickeys` - Query elastic keys with filtering

### Material Keys

Actual cryptographic key material (versioned within elastic keys):

- `POST /materialkey` - Create new material key
- `GET /materialkey/{id}` - Retrieve material key
- `PUT /materialkey/{id}` - Update material key
- `DELETE /materialkey/{id}` - Delete material key
- `GET /materialkeys` - Query material keys with filtering

### Cryptographic Operations

Direct cryptographic operations using keys:

- `POST /crypto/encrypt` - Encrypt data
- `POST /crypto/decrypt` - Decrypt data
- `POST /crypto/sign` - Create digital signature
- `POST /crypto/verify` - Verify digital signature
- `POST /crypto/generate` - Generate key material

## Error Handling

### Standard HTTP Status Codes

All endpoints return consistent error responses:

- `400 Bad Request` - Invalid request format or parameters
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Access denied (IP not allowed, CSRF failure)
- `404 Not Found` - Resource not found
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error
- `502 Bad Gateway` - Upstream service error
- `503 Service Unavailable` - Service temporarily unavailable
- `504 Gateway Timeout` - Request timeout

### Error Response Format

```json
{
  "status": 400,
  "error": "Bad Request",
  "message": "Invalid elastic key algorithm: INVALID_ALG",
  "timestamp": "2025-09-12T10:30:00Z",
  "path": "/browser/api/v1/elastickey"
}
```

## Performance Considerations

### Key Generation Pools

The system uses pre-generated key pools for performance:

- Background key generation
- Configurable pool sizes per algorithm
- Automatic pool replenishment
- Concurrent key generation

### Request Validation

- OpenAPI-based request validation
- Early parameter validation
- Structured error responses
- Request size limits

### Caching Strategy

- No-cache headers for security
- Appropriate cache control for static assets
- Optimized database queries with pagination

## Security Model

### Authentication & Authorization

Currently implemented:
- IP allowlisting (individual IPs and CIDR blocks)
- Rate limiting per IP address
- CSRF protection for browser clients

Future considerations:
- JWT-based authentication
- Role-based access control (RBAC)
- API key management
- OAuth 2.0 integration

### Request Security

1. **Network Layer**: IP filtering, rate limiting
2. **Transport Layer**: TLS encryption, certificate validation
3. **Application Layer**: CORS, CSRF, CSP headers
4. **API Layer**: OpenAPI validation, structured responses
5. **Business Layer**: Key access controls, audit logging

## Monitoring & Observability

### Metrics

- Request counts and latencies per endpoint
- Error rates by status code
- Rate limiting trigger counts
- Key generation pool statistics

### Tracing

- Distributed tracing with OpenTelemetry
- Request correlation across contexts
- Performance bottleneck identification

### Logging

- Structured logging with contextual information
- Request/response logging (excluding sensitive data)
- Security event logging (failed authentications, rate limits)
- Operational event logging (startup, shutdown, health checks)
