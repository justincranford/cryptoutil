# Two HTTPS Endpoints - Complete Specifications

**Referenced by**: `.github/instructions/02-03.https-ports.instructions.md`

## Overview

All cryptoutil services MUST implement TWO separate HTTPS servers:

1. **Public HTTPS Endpoint** (Business APIs, browser UIs, external access)
2. **Private HTTPS Endpoint** (Administration, health checks, graceful shutdown)

**Why**: Security isolation (admin APIs separate from public), network isolation (admin localhost-only), health check patterns (Kubernetes-style livez/readyz probes), authentication requirements (admin optional mTLS, public mandatory authentication)

---

## Binding Parameters

| Parameter | Public Default | Private Default |
|-----------|----------------|-----------------|
| Protocol | `https` | `https` |
| Bind Address | `127.0.0.1` | `127.0.0.1` |
| Port | `8080` | `9090` |

**Environment-Specific Overrides**:

| Environment | Public Bind | Private Bind | Rationale |
|-------------|-------------|--------------|-----------|
| Unit/Integration Tests | `127.0.0.1` | `127.0.0.1` | Prevent Windows Firewall prompts |
| E2E Tests (local) | `127.0.0.1` | `127.0.0.1` | Prevent Windows Firewall prompts |
| Docker Containers | `0.0.0.0` | `127.0.0.1` | Public externally accessible, admin localhost-only |
| Production | Configurable | `127.0.0.1` | Public per deployment, admin localhost-only |

**TLS Subject Alt Names** (both endpoints, development/testing): DNS Names: `["localhost"]`, IP Addresses: `["127.0.0.1", "::1", "::ffff:127.0.0.1"]`

**CORS Configuration** (Public HTTP Endpoint `/browser` paths only): Allowed Origins: `http://localhost:8080`, `http://127.0.0.1:8080`, `http://[::1]:8080`, `https://localhost:8080`, `https://127.0.0.1:8080`, `https://[::1]:8080`

**See**: `service-template.md` for complete ServerConfig pattern

---

## Key Takeaways

**Windows Firewall Warning Prompts**: Binding to `0.0.0.0` triggers Windows Firewall exception prompts, blocking CI/CD automation. Use `0.0.0.0` inside containers (required for external port mapping), use `127.0.0.1` outside containers (prevents firewall prompts). Bind address MUST be configurable, NEVER hardcoded to `0.0.0.0`.

**IPv4 vs IPv6 Dual-Stack Limitations**: Some container runtimes (e.g., Docker Desktop for Windows) have dual-stack routing issues. Use IPv4 `127.0.0.1` or `0.0.0.0` for bind addresses (IPv6 `::1` in TLS SAN for certificate validation, but NOT for binding).

---

## Deployment Environments

**Unit/Integration Tests**: Port 0 (dynamic allocation prevents port conflicts), 127.0.0.1 (prevents Windows Firewall prompts), https (production parity)

**E2E Tests (Local)**: Static ports 8080/9090 (Docker Compose port mapping stability), 127.0.0.1 (prevents Windows Firewall prompts), https (production parity)

**Docker Containers**: Public 0.0.0.0:8080 (external access from host/other containers), Private 127.0.0.1:9090 (admin isolated to localhost), https (production parity)

**Production**: Public configurable (deployment-specific networking), Private ALWAYS 127.0.0.1:9090 (security isolation, orchestrator health checks), https (mandatory TLS)

---

## TLS Certificate Configuration

**All services MUST support two sets of HTTPS configurations** (one for Public, one for Private):

**Configuration Requirements**: Each set must support TLS Server Certificate Chain (Root CA → Intermediate CA → TLS Server), TLS Server Private Key, TLS Client Trusted Certificates (for mTLS validation, usually CA certificates)

**Configuration Options**:

| Environment | Static Cert Chain | Static TLS Server | Issuing CA Key | Outcome |
|-------------|-------------------|-------------------|----------------|---------|
| Production | ✅ Provided | ✅ Provided (Docker Secret) | ❌ NOT PROVIDED | TLS Server cert chain used as-is |
| E2E Dev Tests | ✅ Provided | ❌ NOT PROVIDED | ✅ Provided (Docker Secret) | TLS Server TBSCertificate generated, signed with Issuer key |
| Unit/Integration Dev Tests | ❌ NOT PROVIDED | ❌ NOT PROVIDED | ❌ NOT PROVIDED | Auto-create all certs from Root CA → TLS Server |

**Certificate Settings**: Cert chains, TLS Server cert specified in config file (file paths OR embedded PEM). TLS Server Private Key via Docker Secret (NEVER in config/environment). Issuing CA Private Key for E2E Dev+Test via Docker Secret (for TLS Server cert signing).

---

## Private HTTPS Endpoint (Admin Server)

**Purpose**: Administration APIs, health checks (livez, readyz), graceful shutdown

**Configuration Guidelines**:

| Setting | Test Environments | Production |
|---------|------------------|-----------|
| Port | 0 (dynamic) | 9090 (standard) |
| Bind Address | 127.0.0.1 | 127.0.0.1 |
| TLS | HTTPS | HTTPS |
| External Access | NOT RECOMMENDED | NOT RECOMMENDED |

**Private HTTP APIs**:

- `/admin/v1/livez` (Liveness): Lightweight check (is process alive?), HTTP 200 OK, failure action: restart container
- `/admin/v1/readyz` (Readiness): Heavyweight check (is service ready?), checks DB connection, dependent services, critical resources, HTTP 200 OK, failure action: remove from load balancer (do NOT restart)
- `/admin/v1/shutdown` (Graceful Shutdown): Trigger shutdown sequence (stop accepting requests, drain active requests up to 30s, close connections, release resources, exit), HTTP 200 OK

**Why Two Separate Health Endpoints** (Kubernetes Standard):

| Scenario | Liveness | Readiness | Action |
|----------|----------|-----------|--------|
| Process alive, dependencies healthy | ✅ Pass | ✅ Pass | Serve traffic |
| Process alive, dependencies down | ✅ Pass | ❌ Fail | Remove from LB, don't restart |
| Process stuck/deadlocked | ❌ Fail | ❌ Fail | Restart container |

**Implementation Source**: KMS Reference uses gofiber middleware providing livez/readyz pattern

**Consumers**: Docker health checks, Kubernetes probes, monitoring systems (Prometheus, Grafana), orchestration tools (Docker Swarm, Nomad)

---

## Public HTTPS Endpoint (Public Server)

**Purpose**: Business APIs (REST, gRPC), browser UIs (HTML, JavaScript, CSS), external client access (service-to-service, user-to-service)

**Configuration Guidelines**:

| Setting | Test Environments | Production |
|---------|------------------|-----------|
| Port | 0 (dynamic) | Service-specific (8080-8089 KMS, 8180-8189 Identity, 8280-8289 JOSE, 8380-8389 CA) |
| Bind Address | 127.0.0.1 | 0.0.0.0 (containers), Configurable (VMs) |
| TLS | HTTPS | HTTPS |
| External Access | YES | YES |

---

## Request Path Prefixes and Middlewares - CRITICAL

**Public HTTP endpoint MUST implement TWO security middleware stacks**:

**Service-to-Service APIs** (`/service/**` prefix):

- Access: Service clients ONLY (headless, non-browser), browser clients BLOCKED
- Middleware: IP allowlist → Rate limiting → Request logging → Authentication (Bearer tokens, mTLS) → Authorization (scope-based, service-to-service permissions)
- Example: `/service/api/v1/keys`, `/service/api/v1/tokens`

**Browser-to-Service APIs/UI** (`/browser/**` prefix):

- Access: Browser clients ONLY (user-facing UIs), service clients BLOCKED
- Middleware: IP allowlist → CSRF protection (SameSite cookies, CSRF tokens) → CORS policies → CSP headers (XSS prevention) → Rate limiting → Request logging → Authentication (session cookies, OAuth tokens) → Authorization (resource-level, user permissions)
- Additional Content: HTML pages, JavaScript, CSS, Images, fonts
- Example: `/browser/api/v1/keys`, `/browser/login`, `/browser/assets/app.js`

**API Consistency**: SAME OpenAPI Specification served at both `/service/api/v1/**` and `/browser/api/v1/**` (API contracts identical, only middleware/authentication differ)

**Middleware Mutual Exclusivity**:

```go
// Headless client authentication
if client.Type == "headless" {
    if !strings.HasPrefix(path, "/service/") {
        return ErrUnauthorized  // Block browser paths
    }
}

// Browser client authentication
if client.Type == "browser" {
    if !strings.HasPrefix(path, "/browser/") {
        return ErrUnauthorized  // Block service paths
    }
}
```

**Rationale**: Security requirements differ dramatically. Browser clients need CORS/CSRF/XSS protection, cookie-based sessions. Headless clients need different authentication (Bearer tokens, mTLS), DDOS prevention, service mesh integration. Cross-client access creates security vulnerabilities.

---

## Cross-References

**Related Documentation**: `service-template.md`, `security.md`, `pki.md`, `docker.md`, `authn-authz-factors.md`
