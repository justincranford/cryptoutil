# Two HTTPS Endpoints - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/02-03.https-ports.instructions.md`

## Overview

All cryptoutil services MUST implement TWO separate HTTPS servers:

1. **Public HTTPS Endpoint** (Business APIs, browser UIs, external access)
2. **Private HTTPS Endpoint** (Administration, health checks, graceful shutdown)

**Why Two Separate Endpoints**:

- **Security isolation**: Admin APIs separate from public APIs (different threat models)
- **Network isolation**: Admin accessible only from localhost/orchestrator (not external)
- **Health check patterns**: Kubernetes-style livez/readyz probes (separate concerns)
- **Authentication requirements**: Admin optional mTLS, public mandatory authentication

---

## Binding Parameters

### Default Configuration

| Parameter | Public Endpoint | Private Endpoint |
|-----------|----------------|------------------|
| Protocol | `https` | `https` |
| Bind Address | `127.0.0.1` | `127.0.0.1` |
| Port | `8080` | `9090` |

**MANDATORY: HTTPS Endpoint binding MUST be documented as `<configurable_protocol>://<configurable_address>:<configurable_port>`**

### Environment-Specific Overrides

| Environment | Public Bind Address | Private Bind Address | Rationale |
|-------------|--------------------|--------------------|-----------|
| Unit/Integration Tests | `127.0.0.1` | `127.0.0.1` | Prevent Windows Firewall prompts |
| E2E Tests (local) | `127.0.0.1` | `127.0.0.1` | Prevent Windows Firewall prompts |
| Docker Containers | `0.0.0.0` | `127.0.0.1` | Public externally accessible, admin localhost-only |
| Production | Configurable | `127.0.0.1` | Public per deployment, admin localhost-only |

---

## External TLS Subject Alt Names

**For auto-generated TLS certificates** (development, testing):

### Public HTTP Endpoint

**DNS Names**:

```
dnsName: ["localhost"]
```

**IP Addresses**:

```
ipAddress: [
  "127.0.0.1",              # IPv4 loopback
  "::1",                    # IPv6 loopback
  "::ffff:127.0.0.1"        # IPv4-mapped IPv6 loopback
]
```

### Private HTTP Endpoint

**DNS Names**:

```
dnsName: ["localhost"]
```

**IP Addresses**:

```
ipAddress: [
  "127.0.0.1",              # IPv4 loopback
  "::1",                    # IPv6 loopback
  "::ffff:127.0.0.1"        # IPv4-mapped IPv6 loopback
]
```

**Purpose**: If TLS Server certification chains are generated automatically, these values populate the Subject Alt Name extension of TLS Server leaf certificates.

---

## External CORS Origins

### Public HTTP Endpoint (/browser paths only)

**Allowed Origins**:

```
[
  "http://localhost:8080",
  "http://127.0.0.1:8080",
  "http://[::1]:8080",
  "https://localhost:8080",
  "https://127.0.0.1:8080",
  "https://[::1]:8080"
]
```

**Scope**: ONLY `/browser/**` request paths (browser-to-service APIs/UI)

**Why**: Browser-based requests require CORS policies to prevent cross-origin attacks

### Private HTTP Endpoint

**CORS**: NOT APPLICABLE (no browser UI, admin-only access)

---

## ServerConfig Pattern - CRITICAL

```go
type ServerConfig struct {
    // Public Endpoint Binding
    BindPublicProtocol    string   // "https" (default), "http" (tests/dev only)
    BindPublicAddress     string   // "127.0.0.1" (default), "0.0.0.0" (containers)
    BindPublicPort        uint16   // 8080 (default), 0 (tests - dynamic allocation)

    // Private Endpoint Binding
    BindPrivateProtocol   string   // "https" (default), "http" (tests/dev only)
    BindPrivateAddress    string   // "127.0.0.1" (default), rarely changed
    BindPrivatePort       uint16   // 9090 (default), 0 (tests - dynamic allocation)

    // Public TLS Configuration
    TLSPublicDNSNames     []string // []string{"localhost"} (default)
    TLSPublicIPAddresses  []string // []string{"127.0.0.1", "::1", "::ffff:127.0.0.1"} (default)

    // Private TLS Configuration
    TLSPrivateDNSNames    []string // []string{"localhost"} (default)
    TLSPrivateIPAddresses []string // []string{"127.0.0.1", "::1", "::ffff:127.0.0.1"} (default)

    // CORS Configuration (Public /browser paths only)
    CORSAllowedOrigins    []string // []string{"http://localhost:8080", "http://127.0.0.1:8080", ...} (default)
}
```

---

## Why This Matters

### Windows Firewall Warning Prompts - CRITICAL

**Problem**: Binding to `0.0.0.0` triggers Windows Firewall exception prompts, blocking CI/CD automation

**Solution**:

- **Container deployments**: Use `0.0.0.0` bind address inside containers (required for external port mapping)
- **Test/dev environments**: Use `127.0.0.1` bind address outside containers (prevents firewall prompts)
- **Configuration-driven**: Bind address MUST be configurable, NEVER hardcoded to `0.0.0.0`

**Impact**: Each `0.0.0.0` binding = 1 Windows Firewall popup = blocked test execution

### IPv4 vs IPv6 Dual-Stack Limitations

**Container Runtime Limitation**: Some runtimes (e.g., Docker Desktop for Windows) have dual-stack routing issues

**Problem**: If HTTP Public Endpoint binds to IPv6 address, container runtime may not route traffic from `external IPv4:port` to `internal IPv6:port`

**Solution**: Use IPv4 `127.0.0.1` or `0.0.0.0` for bind addresses (IPv6 `::1` in TLS SAN for certificate validation, but NOT for binding)

---

## Deployment Environments

### Unit/Integration Tests

**Configuration**:

```yaml
public:
  protocol: https
  address: 127.0.0.1
  port: 0  # Dynamic allocation (prevents port conflicts)

private:
  protocol: https
  address: 127.0.0.1
  port: 0  # Dynamic allocation
```

**Rationale**:

- Port 0 → OS assigns random available port (parallel test safety)
- 127.0.0.1 → Prevents Windows Firewall prompts
- https → Production parity (TLS validation in tests)

### E2E Tests (Local)

**Configuration**:

```yaml
public:
  protocol: https
  address: 127.0.0.1
  port: 8080  # Static port (Docker Compose mapping)

private:
  protocol: https
  address: 127.0.0.1
  port: 9090  # Static port (health check scripts)
```

**Rationale**:

- Static ports → Docker Compose port mapping stability
- 127.0.0.1 → Prevents Windows Firewall prompts
- https → Production parity

### Docker Containers

**Configuration**:

```yaml
public:
  protocol: https
  address: 0.0.0.0  # Bind all interfaces (external access required)
  port: 8080

private:
  protocol: https
  address: 127.0.0.1  # Bind loopback only (localhost access only)
  port: 9090
```

**Rationale**:

- Public 0.0.0.0 → External access from host/other containers
- Private 127.0.0.1 → Admin isolated to localhost (security)
- https → Production parity

### Production

**Configuration**:

```yaml
public:
  protocol: https
  address: ${PUBLIC_BIND_ADDRESS}  # Configurable per deployment
  port: ${PUBLIC_PORT}              # Configurable per deployment

private:
  protocol: https
  address: 127.0.0.1  # ALWAYS localhost (admin security)
  port: 9090          # Standard port (orchestrator health checks)
```

**Rationale**:

- Public configurable → Deployment-specific networking requirements
- Private ALWAYS 127.0.0.1 → Security isolation (admin not exposed externally)
- https → Mandatory TLS encryption

---

## TLS Certificate Configuration

**All services MUST support two sets of HTTPS configurations** (one for Public, one for Private):

### Configuration Requirements

Each configuration set must support:

1. **TLS Server Certificate Chain** (Root CA → Intermediate CA → TLS Server)
2. **TLS Server Private Key**
3. **TLS Client Trusted Certificates** (for mTLS validation)

**Note**: TLS Client Trusted certificates are usually CA certificates; self-signed TLS client certificates are strongly discouraged.

### Configuration Options

#### Production (All Certs Passed to Container)

**Inputs**:

- Static cert chain for Root CA → Issuing CA (provided)
- Static private key for Issuing CA (NOT provided)
- Static cert for TLS Server (provided)
- Static private key for TLS Server (provided via Docker Secret)

**Outcome**: TLS Server cert chain and private key used as-is

#### E2E Dev Tests (Mixed Static & Auto-Generated)

**Inputs**:

- Static cert chain for Root CA → Issuing CA (provided)
- Static private key for Issuing CA (provided via Docker Secret)

**Outcome**: TLS Server TBSCertificate generated and signed with Issuer private key to issue TLS Server cert

#### Unit/Integration Dev Tests (All Auto-Generated)

**Inputs**: None (all certificates auto-generated)

**Outcome**:

- Auto-create all certs from Root CA → TLS Server cert
- Retain TLS Server private key
- Discard all CA private keys (ephemeral, not reused)

### Configuration Settings

**Certificate Chains** (Root CA → Issuing CA):

- Specified in configuration file
- Format: File paths OR embedded PEM-encoded values

**TLS Server Certificate**:

- Specified in configuration file
- Format: File paths OR embedded PEM-encoded values

**TLS Server Private Key**:

- Production: Docker Secret (NEVER in config file or environment)
- E2E: Docker Secret (production parity)
- Unit/Integration: Auto-generated ephemeral key

**Issuing CA Private Key**:

- E2E Dev+Test: Docker Secret (for TLS Server cert signing)
- Production: NOT PROVIDED (certificates pre-signed)

**TLS Client Trusted Certificates**:

- Specified in configuration file
- Format: File paths OR embedded PEM-encoded values
- Usually CA certificates for mTLS validation

---

## Private HTTPS Endpoint (Admin Server)

### Purpose

- Administration APIs
- Health checks (livez, readyz)
- Graceful shutdown trigger

### Configuration Guidelines

| Setting | Test Environments | Production | Rationale |
|---------|------------------|-----------|-----------|
| Port | 0 (dynamic) | 9090 (standard) | Tests: Parallel safety, Prod: Orchestrator consistency |
| Bind Address | 127.0.0.1 | 127.0.0.1 | Security isolation (localhost-only access) |
| TLS | HTTPS | HTTPS | Production parity, encrypted admin traffic |
| External Access | NOT RECOMMENDED | NOT RECOMMENDED | Authentication optional, authorization not supported |

**Exceptional Use Cases** (require justification):

- Port conflict with another service → Use different port (e.g., 9091)
- Clear text debugging → Temporarily use HTTP (NEVER in production)

### Private HTTP APIs

#### `/admin/v1/livez` - Liveness Probe

**Purpose**: Lightweight health check (is service process alive?)

**Check**: Service running, process responsive

**Failure Action**: Restart container (process stuck/crashed)

**Response**:

```
HTTP 200 OK
{"status": "alive"}
```

#### `/admin/v1/readyz` - Readiness Probe

**Purpose**: Heavyweight health check (is service ready for traffic?)

**Checks**:

- Database connection healthy
- Dependent services accessible
- Critical resources available (unseal keys, TLS certs)

**Failure Action**: Remove from load balancer (do NOT restart)

**Response**:

```
HTTP 200 OK
{"status": "ready", "dependencies": {"database": "healthy", "kms": "healthy"}}
```

**Why**: Temporary unavailability (database down, network partition) should NOT trigger restart

#### `/admin/v1/shutdown` - Graceful Shutdown

**Purpose**: Trigger graceful shutdown sequence

**Sequence**:

1. Stop accepting new requests (close listeners)
2. Drain active requests (wait up to 30 seconds)
3. Close database connections
4. Release resources (unseal keys, file handles)
5. Exit process

**Response**:

```
HTTP 200 OK
{"status": "shutting_down"}
```

### Why Two Separate Health Endpoints (Kubernetes Standard)

**Liveness vs Readiness**:

| Scenario | Liveness | Readiness | Action |
|----------|----------|-----------|--------|
| Process alive, dependencies healthy | ✅ Pass | ✅ Pass | Serve traffic |
| Process alive, dependencies down | ✅ Pass | ❌ Fail | Remove from LB, don't restart |
| Process stuck/deadlocked | ❌ Fail | ❌ Fail | Restart container |

**Why Separate**:

- Combined health endpoint can't distinguish these failure modes
- Liveness failure → Restart (drastic action)
- Readiness failure → Wait (graceful degradation)

### Implementation Source

**KMS Reference**: Uses gofiber middleware providing livez/readyz pattern out-of-box

**Consumers**:

- Docker health checks (HEALTHCHECK directive)
- Kubernetes probes (livenessProbe, readinessProbe)
- Monitoring systems (Prometheus, Grafana)
- Orchestration tools (Docker Swarm, Nomad)

---

## Public HTTPS Endpoint (Public Server)

### Purpose

- Business APIs (REST, gRPC)
- Browser UIs (HTML, JavaScript, CSS)
- External client access (service-to-service, user-to-service)

### Configuration Guidelines

| Setting | Test Environments | Production | Rationale |
|---------|------------------|-----------|-----------|
| Port | 0 (dynamic) | Service-specific (8080-8089 KMS, 8180-8189 Identity) | Tests: Parallel safety, Prod: Service identification |
| Bind Address | 127.0.0.1 | 0.0.0.0 (containers), Configurable (VMs) | Tests: Prevent firewall, Prod: External access |
| TLS | HTTPS | HTTPS | Production parity, encrypted traffic |
| External Access | YES | YES | Browser clients, external services |

**Service-Specific Port Ranges**:

- KMS: 8080-8089
- Identity: 8180-8189
- JOSE: 8280-8289
- CA: 8380-8389

---

## Request Path Prefixes and Middlewares - CRITICAL

**Public HTTP endpoint MUST implement TWO security middleware stacks**:

### Service-to-Service APIs (`/service/**` prefix)

**Access Control**:

- Service clients ONLY (headless, non-browser)
- Browser clients MUST be blocked by authorization checks

**Middleware Stack**:

1. IP allowlist (restrict to known service CIDR ranges)
2. Rate limiting (per-IP, per-service quotas)
3. Request logging (audit trail, forensics)
4. Authentication (Bearer tokens, mTLS certificates)
5. Authorization (scope-based, service-to-service permissions)

**Example Endpoints**:

- `/service/api/v1/keys` (KMS key management)
- `/service/api/v1/tokens` (Identity token issuance)

### Browser-to-Service APIs/UI (`/browser/**` prefix)

**Access Control**:

- Browser clients ONLY (user-facing UIs)
- Service clients MUST be blocked by authorization checks

**Middleware Stack**:

1. IP allowlist (restrict to user access ranges)
2. CSRF protection (SameSite cookies, CSRF tokens)
3. CORS policies (Allowed-Origin enforcement)
4. CSP headers (Content Security Policy, XSS prevention)
5. Rate limiting (per-IP, per-user quotas)
6. Request logging (audit trail, forensics)
7. Authentication (session cookies, OAuth tokens)
8. Authorization (resource-level, user permissions)

**Additional Content**:

- HTML pages (login, dashboard, admin UI)
- JavaScript (client-side logic)
- CSS (styling)
- Images, fonts (static assets)

**Example Endpoints**:

- `/browser/api/v1/keys` (KMS key management UI API)
- `/browser/login` (Identity login page)
- `/browser/assets/app.js` (JavaScript bundle)

### API Consistency - MANDATORY

**SAME OpenAPI Specification** served at both prefixes:

- `/service/api/v1/**` (service clients)
- `/browser/api/v1/**` (browser clients)

**Why**: API contracts identical, only middleware/authentication differ

### Middleware Mutual Exclusivity - CRITICAL

**Authorization Enforcement**:

```go
// Headless client authentication
if client.Type == "headless" {
    // ONLY authorize /service/** paths
    if !strings.HasPrefix(path, "/service/") {
        return ErrUnauthorized  // Block browser paths
    }
}

// Browser client authentication
if client.Type == "browser" {
    // ONLY authorize /browser/** paths
    if !strings.HasPrefix(path, "/browser/") {
        return ErrUnauthorized  // Block service paths
    }
}
```

**Rationale**: Security requirements differ dramatically:

**Browser Clients**:

- CORS (prevent cross-origin attacks)
- CSRF (prevent forged requests)
- XSS (prevent script injection via CSP)
- Cookie-based sessions (SameSite, HttpOnly, Secure flags)

**Headless Clients**:

- Different authentication (Bearer tokens, mTLS certificates)
- DDOS prevention (bot detection, rate limiting)
- Service mesh integration (mTLS, service discovery)
- NO browser-specific protections (CORS/CSRF/XSS irrelevant)

**Cross-client access patterns prevented** because:

- Browser clients don't need DDOS bot protections (low volume)
- Headless clients can't use CORS/CSRF (no browser context)
- Mixing authentication types creates security vulnerabilities

---

## Cross-References

**Related Documentation**:

- Service template: `.specify/memory/service-template.md`
- Security patterns: `.specify/memory/security.md`
- PKI/TLS configuration: `.specify/memory/pki.md`
- Docker deployment: `.specify/memory/docker.md`
- Authentication/authorization: `.specify/memory/authn-authz-factors.md`
