# CA & JOSE Server TLS Implementation

## Issue

CA and JOSE servers currently use HTTP instead of HTTPS. They need TLS listener implementation like KMS server.

## Root Cause

Current implementation:

```go
// CA/JOSE servers (HTTP only)
listener, err := lc.Listen(ctx, "tcp", addr)
s.fiberApp.Listener(listener)  // Plain HTTP
```

KMS pattern (HTTPS with TLS):

```go
// KMS server (HTTPS with TLS)
listener, err := lc.Listen(ctx, "tcp", addr)
tlsListener := tls.NewListener(listener, tlsConfig)  // Wrap with TLS
s.fiberApp.Listener(tlsListener)  // HTTPS
```

## Required Changes

### 1. CA Server TLS Support

**File**: `internal/ca/server/server.go`

**Changes**:

1. Add TLS config generation (copy from KMS application_listener.go)
2. Wrap listener with `tls.NewListener(listener, tlsConfig)`
3. Update Start() method to use TLS listener
4. Add TLS certificate generation using server's own issuer
5. Store TLS config in Server struct

**Pattern to follow**:

```go
// Generate TLS config with self-signed cert from CA issuer
tlsConfig, err := generateCATLSConfig(s.issuer, s.settings)
if err != nil {
    return fmt.Errorf("failed to generate TLS config: %w", err)
}

// Wrap listener with TLS
tlsListener := tls.NewListener(listener, tlsConfig)

// Use TLS listener
if err := s.fiberApp.Listener(tlsListener); err != nil {
    return fmt.Errorf("CA server failed: %w", err)
}
```

### 2. JOSE Server TLS Support

**File**: `internal/jose/server/server.go`

**Changes**:

1. Add TLS config generation (use JWKGenService for key material)
2. Wrap listener with `tls.NewListener(listener, tlsConfig)`
3. Update Start() method to use TLS listener
4. Add TLS certificate generation
5. Store TLS config in Server struct

**Pattern to follow**: Same as CA server but use JWKGenService instead of issuer

### 3. Docker Compose Health Checks

**File**: `deployments/ca/compose.yml` and `deployments/jose/compose.yml`

**Changes**:

1. Change health check from HTTP to HTTPS:

   ```yaml
   test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:8443/livez"]
   ```

2. Remove TODO comments about HTTP-only status

## Testing Checklist

- [ ] CA server starts with HTTPS listener
- [ ] CA health check passes with HTTPS
- [ ] CA Swagger UI accessible via HTTPS
- [ ] JOSE server starts with HTTPS listener
- [ ] JOSE health check passes with HTTPS
- [ ] JOSE Swagger UI accessible via HTTPS
- [ ] Docker compose deployments healthy
- [ ] TLS cert validation works
- [ ] No HTTP fallback (enforce HTTPS only)

## Reference Files

- **KMS TLS pattern**: `internal/kms/server/application/application_listener.go` lines 110-170
- **TLS config helper**: `internal/infra/tls/server_config.go`
- **Identity admin TLS**: `internal/identity/idp/server/admin.go` lines 167-195

## Priority

**Medium** - Servers work with HTTP for development, but HTTPS required for production security

## Related Tasks

- Update E2E tests to use HTTPS URLs
- Update DAST scanning to use HTTPS endpoints
- Update documentation showing HTTPS examples
