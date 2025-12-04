# Identity Services Architecture - Topology Diagram

## System Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Identity Services Stack                             │
│                                                                               │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐              │
│  │ AuthZ Server │      │  IdP Server  │      │  RS Server   │              │
│  │  (Port 8080) │      │ (Port 8081)  │      │ (Port 8082)  │              │
│  │──────────────│      │──────────────│      │──────────────│              │
│  │ OAuth 2.1    │      │ OIDC Provider│      │ Resource     │              │
│  │ /oauth2/v1/* │◄────►│ /oidc/v1/*   │      │ /api/v1/*    │              │
│  │              │      │              │      │              │              │
│  │ Admin: 9090  │      │ Admin: 9090  │      │ Admin: 9090  │              │
│  └──────┬───────┘      └──────┬───────┘      └──────┬───────┘              │
│         │                     │                     │                       │
│         │                     │                     │                       │
│         └──────────┬──────────┴──────────┬──────────┘                       │
│                    │                     │                                  │
│                    ▼                     ▼                                  │
│         ┌──────────────────────────────────────────┐                       │
│         │      PostgreSQL Database (Port 5432)      │                       │
│         │──────────────────────────────────────────│                       │
│         │ • Clients          • Access Tokens       │                       │
│         │ • Users            • Refresh Tokens      │                       │
│         │ • Sessions         • Authorization Codes │                       │
│         │ • JWKs (Rotation)  • Audit Logs          │                       │
│         └──────────────────────────────────────────┘                       │
│                                                                               │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │                    Background Jobs Layer                             │    │
│  │────────────────────────────────────────────────────────────────────│    │
│  │  • Token Cleanup Job (Hourly)                                       │    │
│  │  • Session Cleanup Job (Hourly)                                     │    │
│  │  • Key Rotation Scheduler (Task 08 integration)                     │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                               │
└───────────────────────────────────────┬───────────────────────────────────── ┘
                                        │ Telemetry (OTLP)
                                        ▼
                        ┌──────────────────────────────┐
                        │ OpenTelemetry Collector      │
                        │ (Ports 4317/4318)            │
                        │──────────────────────────────│
                        │ • Traces Processing          │
                        │ • Metrics Aggregation        │
                        │ • Logs Collection            │
                        └──────────┬───────────────────┘
                                   │
                                   ▼
                        ┌──────────────────────────────┐
                        │ Grafana OTEL LGTM (Port 3000)│
                        │──────────────────────────────│
                        │ • Loki (Logs)                │
                        │ • Grafana (Visualization)    │
                        │ • Tempo (Traces)             │
                        │ • Prometheus (Metrics)       │
                        └──────────────────────────────┘
```

## Data Flow Diagrams

### OAuth 2.0 Authorization Code Flow

```
┌─────────┐                ┌──────────────┐               ┌─────────────┐
│ SPA-RP  │                │ AuthZ Server │               │  IdP Server │
│(Client) │                │              │               │             │
└────┬────┘                └──────┬───────┘               └──────┬──────┘
     │                            │                              │
     │ 1. GET /oauth2/v1/authorize│                              │
     │   ?client_id=...           │                              │
     │   &redirect_uri=...        │                              │
     │   &scope=...               │                              │
     │   &state=...               │                              │
     ├───────────────────────────►│                              │
     │                            │                              │
     │                            │ 2. Redirect to login         │
     │                            │    /oidc/v1/auth             │
     │◄───────────────────────────┤                              │
     │                            │                              │
     │ 3. POST /oidc/v1/login     │                              │
     │    (username/password)     │                              │
     ├────────────────────────────┼─────────────────────────────►│
     │                            │                              │
     │                            │ 4. Validate credentials      │
     │                            │    Create session            │
     │                            │◄─────────────────────────────┤
     │                            │                              │
     │ 5. 302 Redirect            │                              │
     │    ?code=AUTH_CODE         │                              │
     │    &state=...              │                              │
     │◄───────────────────────────┤                              │
     │                            │                              │
     │ 6. POST /oauth2/v1/token   │                              │
     │    grant_type=auth_code    │                              │
     │    code=AUTH_CODE          │                              │
     ├───────────────────────────►│                              │
     │                            │                              │
     │                            │ 7. Verify code               │
     │                            │    Issue tokens              │
     │                            │    (JWS/JWE)                 │
     │                            │                              │
     │ 8. 200 OK                  │                              │
     │    {access_token,          │                              │
     │     id_token,              │                              │
     │     refresh_token}         │                              │
     │◄───────────────────────────┤                              │
     │                            │                              │
```

### Resource Access Flow

```
┌─────────┐            ┌──────────────┐              ┌────────────┐
│ SPA-RP  │            │  RS Server   │              │AuthZ Server│
│(Client) │            │              │              │            │
└────┬────┘            └──────┬───────┘              └──────┬─────┘
     │                        │                             │
     │ 1. GET /api/v1/protected/resource                    │
     │    Authorization: Bearer ACCESS_TOKEN                │
     ├───────────────────────►│                             │
     │                        │                             │
     │                        │ 2. Validate token           │
     │                        │    (signature, expiration)  │
     │                        │    OR                       │
     │                        │    POST /oauth2/v1/introspect│
     │                        ├────────────────────────────►│
     │                        │                             │
     │                        │ 3. Token status + scopes    │
     │                        │◄────────────────────────────┤
     │                        │                             │
     │                        │ 4. Check required scopes    │
     │                        │    (read:resource)          │
     │                        │                             │
     │ 5. 200 OK              │                             │
     │    {data: "..."}       │                             │
     │◄───────────────────────┤                             │
     │                        │                             │
```

## Component Dependencies

### Startup Order (Docker Compose)

```
1. PostgreSQL Database
   └─> Healthcheck: pg_isready

2. AuthZ Server (depends_on: postgres healthy)
   └─> Healthcheck: /livez endpoint

3. IdP Server (depends_on: postgres healthy, authz healthy)
   └─> Healthcheck: /livez endpoint

4. RS Server (depends_on: authz healthy)
   └─> Healthcheck: /livez endpoint

5. SPA-RP (depends_on: authz healthy, idp healthy)
   └─> Healthcheck: /livez endpoint

In parallel:
- OTEL Collector (depends_on: none)
- Grafana OTEL LGTM (depends_on: none)
```

### Network Communication Matrix

| From | To | Protocol | Purpose |
|------|-----|----------|---------|
| SPA-RP | AuthZ | HTTPS/HTTP | Authorization requests, token exchange |
| SPA-RP | IdP | HTTPS/HTTP | Login, UserInfo, logout |
| SPA-RP | RS | HTTPS/HTTP | Protected resource access |
| RS | AuthZ | HTTPS/HTTP | Token introspection (optional) |
| All Servers | PostgreSQL | TCP:5432 | Data persistence |
| All Servers | OTEL Collector | gRPC:4317, HTTP:4318 | Telemetry export |
| OTEL Collector | Grafana | gRPC:14317, HTTP:14318 | Telemetry forwarding |

## Health Check Endpoints

All servers expose admin health endpoints on port 9090 (not exposed to host in Docker):

- `GET /livez` - Liveness check (is server running?)
- `GET /readyz` - Readiness check (is server ready to accept traffic?)
- `POST /shutdown` - Graceful shutdown trigger

Public health endpoints:

- AuthZ: `GET /health` (planned)
- IdP: `GET /health` (planned)
- RS: `GET /api/v1/public/health` (implemented)

## Port Allocation

### Host Ports (Exposed)

- 8080: AuthZ Server (OAuth 2.1 endpoints)
- 8081: IdP Server (OIDC endpoints)
- 8082: RS Server (Protected resources)
- 8083: SPA-RP (Relying party application)
- 5433: PostgreSQL (identity services database)
- 3000: Grafana UI
- 4317/4318: OTEL Collector (OTLP gRPC/HTTP)

### Internal Ports (Container-only)

- 9090: Admin APIs (all servers, not exposed to host)
- 5432: PostgreSQL (internal container port)

## Security Boundaries

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Network: identity-network          │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Public APIs (Exposed to host via port mappings)     │   │
│  │  • AuthZ:8080   • IdP:8081   • RS:8082   • SPA:8083 │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Admin APIs (Internal only, 127.0.0.1:9090 binding)  │   │
│  │  • /livez   • /readyz   • /shutdown                  │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Database (Internal only)                             │   │
│  │  • PostgreSQL 5432 (exposed via 5433 for debugging)  │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Telemetry Flow

```
AuthZ Server ────┐
                 │
IdP Server ──────┼──► OTLP (gRPC:4317) ──► OTEL Collector ──► Grafana LGTM
                 │                              │
RS Server ───────┘                              │
                                                 │
                                                 └──► Self-metrics (Prometheus:8888)
```

## Future Enhancements (Planned Tasks)

- **Task 11**: MFA chains with observability
- **Task 12**: OTP/Magic Link services (may add message queue)
- **Task 13**: Adaptive authentication engine
- **Task 14**: WebAuthn/Biometric authentication
- **Task 15**: Hardware credential support (FIDO2/U2F)
- **Task 18**: Enhanced orchestration tooling

## References

- Docker Compose Configuration: `deployments/compose/identity-compose.yml`
- Server Manager: `internal/identity/server/server_manager.go`
- Background Jobs: `internal/identity/jobs/cleanup.go`
- Integration Tests: `internal/identity/integration/integration_test.go`
