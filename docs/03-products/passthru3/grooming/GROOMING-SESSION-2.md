# Grooming Session 2: Implementation Patterns & Code Standards

**Purpose**: Define exact implementation patterns BEFORE coding
**Date**: 2025-12-01

---

## Topic 1: Code Structure

### Q1.1: What structure should integration.go follow?

**Decision**: Mirror existing demo patterns (kms.go, identity.go)

**File Structure**:

```go
// Package declaration and imports
package demo

// Constants and magic values
const (
    integrationDemoName = "Integration"
    stepCount           = 7
)

// Main entry point
func RunIntegrationDemo(cfg *config.Config) error {
    ctx, cancel := context.WithTimeout(context.Background(), demoTimeout)
    defer cancel()

    // Setup demo runner
    demo := newDemoRunner(integrationDemoName, stepCount)

    // Execute steps
    if err := demo.runStep(1, "Start Identity server", func() error { ... }); err != nil {
        return err
    }
    // ... remaining steps

    return demo.complete()
}
```

### Q1.2: How should step execution be standardized?

**Decision**: Use demoRunner pattern from existing demos

**Pattern**:

```go
type demoRunner struct {
    name       string
    totalSteps int
    passCount  int
    failCount  int
    output     *strings.Builder
}

func (d *demoRunner) runStep(num int, name string, fn func() error) error {
    fmt.Printf("Step %d/%d: %s... ", num, d.totalSteps, name)
    if err := fn(); err != nil {
        d.failCount++
        fmt.Println("❌ FAIL")
        fmt.Printf("  Error: %v\n", err)
        return err
    }
    d.passCount++
    fmt.Println("✅ PASS")
    return nil
}
```

---

## Topic 2: Server Lifecycle Management

### Q2.1: How to start servers cleanly?

**Decision**: Use goroutines with proper shutdown coordination

**Pattern**:

```go
func startIdentityServer(ctx context.Context, cfg *config.Config) (cleanup func(), err error) {
    // Create server
    server, err := identity.NewServer(cfg)
    if err != nil {
        return nil, fmt.Errorf("create identity server: %w", err)
    }

    // Start in goroutine
    errCh := make(chan error, 1)
    go func() {
        if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            errCh <- err
        }
    }()

    // Wait for ready or error
    select {
    case err := <-errCh:
        return nil, fmt.Errorf("server startup failed: %w", err)
    case <-server.Ready():
        // Server is ready
    case <-time.After(startupTimeout):
        return nil, fmt.Errorf("server startup timeout")
    }

    // Return cleanup function
    return func() {
        ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
        defer cancel()
        server.Shutdown(ctx)
    }, nil
}
```

### Q2.2: How to handle cleanup on failure?

**Decision**: Use defer stack for guaranteed cleanup

**Pattern**:

```go
func RunIntegrationDemo(cfg *config.Config) error {
    var cleanups []func()
    defer func() {
        // Execute cleanups in reverse order
        for i := len(cleanups) - 1; i >= 0; i-- {
            cleanups[i]()
        }
    }()

    // Step 1: Start Identity
    identityCleanup, err := startIdentityServer(ctx, cfg)
    if err != nil {
        return err
    }
    cleanups = append(cleanups, identityCleanup)

    // Step 2: Start KMS (will be cleaned up even if later steps fail)
    kmsCleanup, err := startKMSServer(ctx, cfg)
    if err != nil {
        return err
    }
    cleanups = append(cleanups, kmsCleanup)

    // ... continue with other steps
}
```

---

## Topic 3: HTTP Client Configuration

### Q3.1: How to configure TLS for demo?

**Decision**: Custom transport with InsecureSkipVerify for demo self-signed certs

**Pattern**:

```go
func newDemoHTTPClient() *http.Client {
    return &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true, // Demo only - NEVER in production
            },
        },
        Timeout: requestTimeout,
    }
}
```

**NOTE**: This is acceptable ONLY for demos. Production code must validate certs.

### Q3.2: How to perform health checks?

**Decision**: Poll health endpoint with exponential backoff

**Pattern**:

```go
func waitForHealth(ctx context.Context, client *http.Client, url string) error {
    backoff := 100 * time.Millisecond
    maxBackoff := 2 * time.Second

    for {
        select {
        case <-ctx.Done():
            return fmt.Errorf("health check timeout: %w", ctx.Err())
        default:
            resp, err := client.Get(url)
            if err == nil && resp.StatusCode == http.StatusOK {
                resp.Body.Close()
                return nil
            }
            if resp != nil {
                resp.Body.Close()
            }

            time.Sleep(backoff)
            backoff = min(backoff*2, maxBackoff)
        }
    }
}
```

---

## Topic 4: OAuth Token Operations

### Q4.1: How to request client_credentials token?

**Decision**: Standard OAuth 2.0 token request

**Pattern**:

```go
func getClientCredentialsToken(
    ctx context.Context,
    client *http.Client,
    tokenEndpoint string,
    clientID string,
    clientSecret string,
    scopes []string,
) (*TokenResponse, error) {
    data := url.Values{
        "grant_type":    {"client_credentials"},
        "client_id":     {clientID},
        "client_secret": {clientSecret},
        "scope":         {strings.Join(scopes, " ")},
    }

    req, err := http.NewRequestWithContext(ctx, "POST", tokenEndpoint, strings.NewReader(data.Encode()))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("token request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("token request failed: %d: %s", resp.StatusCode, string(body))
    }

    var tokenResp TokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }

    return &tokenResp, nil
}

type TokenResponse struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
    ExpiresIn   int    `json:"expires_in"`
    Scope       string `json:"scope,omitempty"`
}
```

### Q4.2: How to validate JWT token?

**Decision**: Use jose library for JWKS validation

**Pattern**:

```go
func validateJWT(
    ctx context.Context,
    client *http.Client,
    token string,
    jwksURL string,
    expectedIssuer string,
    expectedAudience string,
) error {
    // Fetch JWKS
    resp, err := client.Get(jwksURL)
    if err != nil {
        return fmt.Errorf("fetch JWKS: %w", err)
    }
    defer resp.Body.Close()

    var jwks jose.JSONWebKeySet
    if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
        return fmt.Errorf("decode JWKS: %w", err)
    }

    // Parse JWT
    parsedJWT, err := jwt.ParseSigned(token)
    if err != nil {
        return fmt.Errorf("parse JWT: %w", err)
    }

    // Verify signature and claims
    var claims jwt.Claims
    if err := parsedJWT.Claims(jwks, &claims); err != nil {
        return fmt.Errorf("verify JWT: %w", err)
    }

    // Validate standard claims
    if err := claims.Validate(jwt.Expected{
        Issuer:   expectedIssuer,
        Audience: jwt.Audience{expectedAudience},
        Time:     time.Now(),
    }); err != nil {
        return fmt.Errorf("validate claims: %w", err)
    }

    return nil
}
```

---

## Topic 5: Linting Compliance

### Q5.1: What linting rules must integration.go follow?

**Decision**: Full compliance, zero exceptions

| Linter | Requirement |
|--------|-------------|
| wsl | Proper blank lines between statement types |
| godot | All comments end with period |
| errcheck | All errors checked |
| govet | No vet warnings |
| staticcheck | No staticcheck warnings |
| mnd | No magic numbers (use constants) |

### Q5.2: Required package documentation

**Pattern**:

```go
// Package demo provides demonstration commands for cryptoutil functionality.
// This file implements the integration demo which demonstrates KMS and Identity
// server interaction including OAuth 2.1 token flow and authenticated operations.
package demo
```

---

## Sign-Off

**All implementation patterns in this document are LOCKED**

- [ ] Reviewed and approved
- [ ] No open questions remain
- [ ] Ready for implementation

**Date**: ____________
**Approved By**: ____________
