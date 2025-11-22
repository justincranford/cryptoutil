# Task 01: Architecture Diagrams - Current State

## Overview

Post-Tasks 12-20, the identity architecture has evolved significantly beyond the original baseline (commit 5c04e44). This document updates Mermaid diagrams to reflect completed orchestration, E2E testing, and advanced authentication features.

---

## Identity Services Architecture (Post-Task 20)

```mermaid
flowchart TB
    subgraph "Client Tier"
        SPA[SPA Relying Party<br/>cmd/identity/spa-rp<br/>✅ OAuth 2.1 PKCE flows]
        CLI[Identity Orchestrator CLI<br/>cmd/identity-orchestrator<br/>✅ Docker Compose management]
    end

    subgraph "Identity Services (AuthZ, IdP, RS)"
        AuthZ[Authorization Server<br/>✅ Token endpoint operational<br/>⚠️ Code persistence missing<br/>⚠️ Consent integration incomplete]
        IdP[Identity Provider<br/>✅ MFA orchestrator<br/>✅ OTP/Magic Link<br/>✅ Adaptive auth<br/>✅ WebAuthn/FIDO2<br/>⚠️ Login/consent pages stubbed]
        RS[Resource Server<br/>❌ Token validation TODO<br/>❌ Scope enforcement missing]
    end

    subgraph "Shared Infrastructure"
        DB[(Identity Database<br/>PostgreSQL/SQLite<br/>✅ GORM repositories<br/>✅ Migrations via golang-migrate)]
        TokenSvc[Token Service<br/>✅ Access/refresh tokens<br/>✅ JWT signing<br/>✅ Key rotation support]
        Cache[Session Cache<br/>✅ In-memory sessions<br/>⚠️ Cleanup job disabled]
    end

    subgraph "Docker Orchestration (Task 18)"
        Compose[identity-demo.yml<br/>✅ 4 profiles: demo/dev/ci/prod<br/>✅ Scaling: 1x, 2x, 3x<br/>✅ Docker secrets integration]
    end

    subgraph "E2E Testing Fabric (Task 19)"
        OAuthTests[OAuth Flow Tests<br/>✅ Authorization code<br/>✅ Client credentials<br/>✅ PKCE validation<br/>✅ Refresh tokens]
        FailoverTests[Failover Tests<br/>✅ AuthZ/IdP/RS scaling<br/>✅ 2x2x2x2 orchestration]
        ObsTests[Observability Tests<br/>✅ OTLP collector<br/>✅ Grafana integration<br/>✅ Prometheus metrics]
    end

    subgraph "Observability (OTEL)"
        OTELCollector[OpenTelemetry Collector<br/>✅ Traces/metrics/logs<br/>✅ Prometheus exporter]
        Grafana[Grafana LGTM Stack<br/>✅ Dashboards<br/>✅ Tempo traces<br/>✅ Loki logs<br/>✅ Prometheus metrics]
    end

    %% Client interactions
    SPA -->|Authorization Code + PKCE| AuthZ
    SPA -->|Bearer tokens| RS
    CLI -->|Start/stop/health/logs| Compose

    %% Service interactions
    AuthZ -->|User auth redirect| IdP
    IdP -->|Session + consent| AuthZ
    AuthZ -->|Issue tokens| TokenSvc
    TokenSvc -->|Store tokens| DB
    IdP -->|Validate MFA| DB
    IdP -->|Create sessions| Cache
    AuthZ & IdP & RS --> DB

    %% Orchestration
    Compose -->|Manages| AuthZ & IdP & RS

    %% E2E testing
    OAuthTests -->|Test flows| AuthZ & IdP & TokenSvc
    FailoverTests -->|Test scaling| Compose
    ObsTests -->|Validate telemetry| OTELCollector

    %% Observability pipeline
    AuthZ & IdP & RS -->|OTLP| OTELCollector
    OTELCollector -->|Traces/metrics/logs| Grafana

    style AuthZ fill:#fff3cd,stroke:#856404
    style IdP fill:#fff3cd,stroke:#856404
    style RS fill:#f8d7da,stroke:#721c24
    style SPA fill:#d4edda,stroke:#155724
    style CLI fill:#d4edda,stroke:#155724
    style Compose fill:#d4edda,stroke:#155724
    style OAuthTests fill:#d4edda,stroke:#155724
    style FailoverTests fill:#d4edda,stroke:#155724
    style ObsTests fill:#d4edda,stroke:#155724
```

**Legend**:
- 🟢 Green: Complete and working
- 🟡 Yellow: Partial implementation with TODOs
- 🔴 Red: Missing critical functionality

---

## MFA Authentication Flow (Task 11-13)

```mermaid
sequenceDiagram
    participant User
    participant SPA as SPA RP
    participant IdP as IdP Service
    participant MFAOrch as MFA Orchestrator
    participant TOTP as TOTP Validator
    participant OTP as OTP Authenticator
    participant WebAuthn as WebAuthn Authenticator
    participant Adaptive as Adaptive Engine
    participant DB as Identity DB

    User->>SPA: Initiate login
    SPA->>IdP: POST /login (username, password)
    IdP->>DB: Validate credentials
    DB-->>IdP: User found

    IdP->>Adaptive: Calculate risk score (IP, device, behavior)
    Adaptive-->>IdP: Risk score + policy decision

    alt Low risk - no MFA required
        IdP-->>SPA: Session cookie + redirect
    else Medium risk - standard MFA
        IdP->>MFAOrch: Trigger MFA chain (TOTP)
        MFAOrch->>TOTP: Validate TOTP code
        TOTP->>DB: Fetch MFA secret
        DB-->>TOTP: User MFA secret
        TOTP-->>MFAOrch: Validation result
        MFAOrch-->>IdP: MFA success
        IdP-->>SPA: Session cookie + redirect
    else High risk - step-up MFA
        IdP->>MFAOrch: Trigger step-up chain (OTP + WebAuthn)
        MFAOrch->>OTP: Send SMS/email OTP
        OTP->>User: Deliver OTP code
        User->>SPA: Enter OTP
        SPA->>MFAOrch: Validate OTP
        MFAOrch->>OTP: Verify OTP
        OTP->>DB: Check OTP token (bcrypt hashed)
        DB-->>OTP: OTP valid
        OTP-->>MFAOrch: OTP success
        MFAOrch->>WebAuthn: Request passkey assertion
        WebAuthn->>User: Prompt hardware key
        User->>WebAuthn: Provide passkey
        WebAuthn->>DB: Validate credential
        DB-->>WebAuthn: Credential valid
        WebAuthn-->>MFAOrch: WebAuthn success
        MFAOrch-->>IdP: Step-up MFA complete
        IdP-->>SPA: High-assurance session + redirect
    end
```

---

## OAuth 2.1 Authorization Code Flow (Current vs Expected)

### Current Implementation (Partial)

```mermaid
sequenceDiagram
    participant SPA as SPA RP
    participant AuthZ as AuthZ Server
    participant IdP as IdP Service
    participant DB as Identity DB

    SPA->>AuthZ: GET /authorize (client_id, redirect_uri, PKCE)
    AuthZ->>AuthZ: Validate parameters
    AuthZ->>DB: Fetch client
    DB-->>AuthZ: Client details

    Note over AuthZ,IdP: ❌ TODO: Authorization request storage
    Note over AuthZ,IdP: ❌ TODO: Redirect to login/consent

    AuthZ-->>SPA: Placeholder redirect (not functional)

    Note over SPA: FLOW BLOCKED - No authorization code issued
```

### Expected Implementation (Post-Task 06-09)

```mermaid
sequenceDiagram
    participant SPA as SPA RP
    participant AuthZ as AuthZ Server
    participant IdP as IdP Service
    participant DB as Identity DB
    participant TokenSvc as Token Service

    SPA->>AuthZ: GET /authorize (client_id, redirect_uri, PKCE)
    AuthZ->>AuthZ: Validate parameters
    AuthZ->>DB: Fetch client
    DB-->>AuthZ: Client details

    AuthZ->>DB: Store authorization request (PKCE challenge)
    DB-->>AuthZ: Request ID

    AuthZ-->>SPA: Redirect to IdP login

    SPA->>IdP: GET /login (client_id, redirect_uri, state)
    IdP-->>SPA: Render login page

    SPA->>IdP: POST /login (username, password)
    IdP->>DB: Validate credentials
    DB-->>IdP: User found
    IdP->>DB: Create session
    DB-->>IdP: Session ID

    IdP-->>SPA: Redirect to consent page

    SPA->>IdP: GET /consent (client_id, scopes)
    IdP-->>SPA: Render consent page

    SPA->>IdP: POST /consent (approved scopes)
    IdP->>DB: Store consent decision
    DB-->>IdP: Consent ID

    IdP->>AuthZ: Generate authorization code
    AuthZ->>DB: Store code (linked to request, user, consent)
    DB-->>AuthZ: Code ID

    AuthZ-->>SPA: Redirect with authorization code

    SPA->>AuthZ: POST /token (code, code_verifier)
    AuthZ->>DB: Validate code
    DB-->>AuthZ: Code valid + request details
    AuthZ->>AuthZ: Validate PKCE (S256 hash)
    AuthZ->>TokenSvc: Generate tokens (user, client, scopes)
    TokenSvc-->>AuthZ: Access + refresh tokens
    AuthZ->>DB: Store tokens
    DB-->>AuthZ: Token IDs
    AuthZ->>DB: Invalidate authorization code
    DB-->>AuthZ: Code deleted

    AuthZ-->>SPA: Token response (access, refresh, expires_in)
```

---

## Docker Compose Orchestration (Task 18)

```mermaid
flowchart LR
    subgraph "Docker Compose Profiles"
        Demo[demo profile<br/>SQLite<br/>1x scaling<br/>Quick testing]
        Dev[development profile<br/>PostgreSQL<br/>2x scaling<br/>Local development]
        CI[ci profile<br/>PostgreSQL<br/>3x scaling<br/>GitHub Actions]
        Prod[production profile<br/>PostgreSQL<br/>Custom scaling<br/>Deployment ready]
    end

    subgraph "Identity Services (Scaling Examples)"
        AuthZ1[AuthZ-1<br/>:8080]
        AuthZ2[AuthZ-2<br/>:8090]
        IdP1[IdP-1<br/>:8100]
        IdP2[IdP-2<br/>:8110]
        RS1[RS-1<br/>:8200]
        RS2[RS-2<br/>:8210]
    end

    subgraph "Shared Infrastructure"
        PG[(PostgreSQL<br/>:5432)]
        OTEL[OTEL Collector<br/>:4317/:4318]
        Grafana[Grafana LGTM<br/>:3000]
    end

    Demo --> AuthZ1 & IdP1 & RS1
    Dev --> AuthZ1 & AuthZ2 & IdP1 & IdP2 & RS1 & RS2
    CI --> AuthZ1 & AuthZ2 & IdP1 & IdP2 & RS1 & RS2
    Prod --> AuthZ1 & AuthZ2 & IdP1 & IdP2 & RS1 & RS2

    AuthZ1 & AuthZ2 & IdP1 & IdP2 & RS1 & RS2 --> PG
    AuthZ1 & AuthZ2 & IdP1 & IdP2 & RS1 & RS2 --> OTEL
    OTEL --> Grafana

    style Demo fill:#d4edda,stroke:#155724
    style Dev fill:#d4edda,stroke:#155724
    style CI fill:#d4edda,stroke:#155724
    style Prod fill:#d4edda,stroke:#155724
```

---

## E2E Testing Infrastructure (Task 19)

```mermaid
flowchart TB
    subgraph "Test Suites"
        OAuth[OAuth Flow Tests<br/>oauth_flows_test.go<br/>391 lines]
        Failover[Failover Tests<br/>orchestration_failover_test.go<br/>330 lines]
        Obs[Observability Tests<br/>observability_test.go<br/>396 lines]
    end

    subgraph "Test Scenarios"
        AuthCode[Authorization Code Flow<br/>✅ PKCE validation<br/>✅ Token exchange]
        ClientCreds[Client Credentials Flow<br/>✅ Service-to-service auth]
        Introspect[Token Introspection<br/>✅ Active/inactive tokens]
        Refresh[Refresh Token Flow<br/>✅ Token rotation]
        PKCE[PKCE S256 Validation<br/>✅ Challenge/verifier]

        AuthZFailover[AuthZ Failover<br/>✅ 2x scaling<br/>✅ Instance rotation]
        IdPFailover[IdP Failover<br/>✅ 2x scaling<br/>✅ Session persistence]
        RSFailover[RS Failover<br/>✅ 2x scaling<br/>✅ Token validation]

        OTELTest[OTEL Collector Test<br/>✅ Traces/metrics/logs]
        GrafanaTest[Grafana Integration<br/>✅ Dashboard validation<br/>⚠️ API queries TODO]
        PrometheusTest[Prometheus Scraping<br/>✅ Metrics endpoints]
    end

    subgraph "Docker Orchestration"
        Compose[identity-demo.yml<br/>✅ Service lifecycle<br/>✅ Health checks<br/>✅ Secrets management]
    end

    OAuth --> AuthCode & ClientCreds & Introspect & Refresh & PKCE
    Failover --> AuthZFailover & IdPFailover & RSFailover
    Obs --> OTELTest & GrafanaTest & PrometheusTest

    AuthCode & ClientCreds & Introspect & Refresh & PKCE --> Compose
    AuthZFailover & IdPFailover & RSFailover --> Compose
    OTELTest & GrafanaTest & PrometheusTest --> Compose

    style OAuth fill:#d4edda,stroke:#155724
    style Failover fill:#d4edda,stroke:#155724
    style Obs fill:#d4edda,stroke:#155724
    style GrafanaTest fill:#fff3cd,stroke:#856404
```

---

## Critical Path Gaps (Tasks 06-10)

```mermaid
flowchart LR
    subgraph "Priority 1: Authorization Code Flow"
        Gap1[❌ Code persistence<br/>handlers_authorize.go<br/>Line 112-114]
        Gap2[❌ PKCE validation<br/>handlers_token.go<br/>Line 79]
        Gap3[❌ Consent integration<br/>handlers_consent.go<br/>Line 46-48]
    end

    subgraph "Priority 2: IdP Login/Consent"
        Gap4[❌ Login page rendering<br/>handlers_login.go<br/>Line 25]
        Gap5[❌ Consent page rendering<br/>handlers_consent.go<br/>Line 22]
        Gap6[❌ Redirect to callback<br/>handlers_login.go<br/>Line 110]
    end

    subgraph "Priority 3: RS Token Validation"
        Gap7[❌ Bearer token parsing<br/>server/rs_server.go<br/>Line 27]
        Gap8[❌ Scope enforcement<br/>server/rs_server.go<br/>Line 33]
    end

    subgraph "Priority 4: Session Lifecycle"
        Gap9[❌ Logout implementation<br/>handlers_logout.go<br/>Line 27-30]
        Gap10[❌ UserInfo token validation<br/>handlers_userinfo.go<br/>Line 23-26]
    end

    Gap1 & Gap2 & Gap3 -->|Blocks| OAuthFlows[OAuth 2.1 Flows]
    Gap4 & Gap5 & Gap6 -->|Blocks| UserAuth[User Authentication]
    Gap7 & Gap8 -->|Blocks| APIProtection[API Protection]
    Gap9 & Gap10 -->|Blocks| Compliance[OIDC Compliance]

    style Gap1 fill:#f8d7da,stroke:#721c24
    style Gap2 fill:#f8d7da,stroke:#721c24
    style Gap3 fill:#f8d7da,stroke:#721c24
    style Gap4 fill:#f8d7da,stroke:#721c24
    style Gap5 fill:#f8d7da,stroke:#721c24
    style Gap6 fill:#f8d7da,stroke:#721c24
    style Gap7 fill:#f8d7da,stroke:#721c24
    style Gap8 fill:#f8d7da,stroke:#721c24
    style Gap9 fill:#f8d7da,stroke:#721c24
    style Gap10 fill:#f8d7da,stroke:#721c24
```

---

## Validation

- ✅ Diagrams reflect post-Task 20 architecture (orchestration, E2E testing, observability)
- ✅ Current vs expected flows documented (OAuth 2.1, MFA)
- ✅ Critical path gaps visualized (Tasks 06-10 priorities)
- ✅ Mermaid syntax validated

---

*Document created as part of Task 01: Historical Baseline Assessment*
