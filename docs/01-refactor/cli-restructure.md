# CLI Command Restructure Plan

## Executive Summary

Implement service-group-based CLI structure with `kms`, `identity`, and `ca` top-level commands, migrating existing `server` command to `kms server` while maintaining backward compatibility.

**Status**: Planning
**Dependencies**: Tasks 10-12 (identity, KMS, CA extractions complete)
**Risk Level**: Medium (CLI UX change, backward compatibility required)

## Current CLI Structure

From `internal/cmd/cryptoutil/cryptoutil.go`:

```
cryptoutil
├── server              # KMS server operations (start, stop, live, ready, init)
└── identity            # Identity service routing
    ├── authz           # OAuth 2.1 Authorization Server (not implemented)
    ├── idp             # OIDC Identity Provider (not implemented)
    ├── rs              # Resource Server (not implemented)
    └── spa-rp          # SPA Relying Party (not implemented)
```

**Issues**:

- `server` is ambiguous (actually KMS server, not identity/CA)
- No `kms` top-level command (inconsistent with service groups)
- `identity` subcommands not implemented in main CLI
- No `ca` command (future)

## Target CLI Structure

From [CLI Strategy Framework](cli-strategy.md):

```
cryptoutil
├── kms                 # KMS service group (NEW)
│   ├── server          # Server operations (moved from top-level)
│   │   ├── start       # Start KMS server
│   │   ├── stop        # Stop KMS server
│   │   ├── live        # Liveness check
│   │   ├── ready       # Readiness check
│   │   └── init        # Initialize KMS
│   ├── key             # Key operations (future)
│   │   ├── generate    # Generate new key
│   │   ├── list        # List keys
│   │   ├── export      # Export key
│   │   └── delete      # Delete key
│   └── barrier         # Barrier operations (future)
│       ├── unseal      # Unseal barrier
│       └── seal        # Seal barrier
├── identity            # Identity service group (ENHANCED)
│   ├── authz           # OAuth 2.1 Authorization Server
│   │   └── server      # Server operations (NEW)
│   │       ├── start   # Start AuthZ server
│   │       └── stop    # Stop AuthZ server
│   ├── idp             # OIDC Identity Provider
│   │   └── server      # Server operations (NEW)
│   │       ├── start   # Start IdP server
│   │       └── stop    # Stop IdP server
│   ├── rs              # Resource Server
│   │   └── server      # Server operations (NEW)
│   │       ├── start   # Start RS server
│   │       └── stop    # Stop RS server
│   └── spa-rp          # SPA Relying Party
│       └── server      # Server operations (NEW)
│           ├── start   # Start SPA-RP server
│           └── stop    # Stop SPA-RP server
├── ca                  # CA service group (NEW - future)
│   ├── server          # Server operations
│   │   ├── start       # Start CA server
│   │   └── stop        # Stop CA server
│   ├── cert            # Certificate operations
│   │   ├── issue       # Issue certificate
│   │   ├── renew       # Renew certificate
│   │   ├── revoke      # Revoke certificate
│   │   └── list        # List certificates
│   └── crl             # CRL operations
│       ├── generate    # Generate CRL
│       └── publish     # Publish CRL
└── server              # DEPRECATED (legacy alias for kms server)
    ├── start           # → kms server start
    ├── stop            # → kms server stop
    ├── live            # → kms server live
    ├── ready           # → kms server ready
    └── init            # → kms server init
```

## Implementation Phases

### Phase 1: Create Service Group Subdirectories

**Create new command directories**:

```bash
# KMS commands
mkdir -p internal/cmd/cryptoutil/kms/server
mkdir -p internal/cmd/cryptoutil/kms/key
mkdir -p internal/cmd/cryptoutil/kms/barrier

# Identity commands (enhance existing)
mkdir -p internal/cmd/cryptoutil/identity/authz/server
mkdir -p internal/cmd/cryptoutil/identity/idp/server
mkdir -p internal/cmd/cryptoutil/identity/rs/server
mkdir -p internal/cmd/cryptoutil/identity/spa-rp/server

# CA commands (skeleton)
mkdir -p internal/cmd/cryptoutil/ca/server
mkdir -p internal/cmd/cryptoutil/ca/cert
mkdir -p internal/cmd/cryptoutil/ca/crl
```

### Phase 2: Migrate `server` Command to `kms server`

**Move existing server command**:

```bash
# Rename server.go → kms/server/server.go
git mv internal/cmd/cryptoutil/server.go internal/cmd/cryptoutil/kms/server/server.go
```

**Update package declaration**:

```go
// OLD (internal/cmd/cryptoutil/server.go)
package cmd

func server(parameters []string) {
    // ... existing server logic
}

// NEW (internal/cmd/cryptoutil/kms/server/server.go)
package server

// Execute runs KMS server commands.
func Execute(parameters []string) {
    settings, err := cryptoutilConfig.Parse(parameters, true)
    if err != nil {
        log.Fatal("Error parsing config:", err)
    }

    switch settings.SubCommand {
    case "start":
        startServerListenerApplication(settings)
    case "stop":
        sendServerListenerShutdownRequest(settings)
    case "live":
        sendServerListenerLivenessCheck(settings)
    case "ready":
        sendServerListenerReadinessCheck(settings)
    case "init":
        serverInit(settings)
    default:
        log.Fatalf("unknown kms server subcommand: %v", settings.SubCommand)
    }
}
```

**Create `internal/cmd/cryptoutil/kms/kms.go`** (KMS dispatcher):

```go
// Copyright (c) 2025 Justin Cranford
//
//

package kms

import (
    "fmt"
    "os"

    "cryptoutil/internal/cmd/cryptoutil/kms/server"
)

// Execute runs KMS service group commands.
func Execute(parameters []string) {
    if len(parameters) < 1 {
        printUsage()
        os.Exit(1)
    }

    subcommand := parameters[0]
    subcommandParams := parameters[1:]

    switch subcommand {
    case "server":
        server.Execute(subcommandParams)
    case "key":
        keyExecute(subcommandParams) // Future: key operations
    case "barrier":
        barrierExecute(subcommandParams) // Future: barrier operations
    case "help":
        printUsage()
    default:
        printUsage()
        fmt.Printf("Unknown kms subcommand: %s\n", subcommand)
        os.Exit(1)
    }
}

func printUsage() {
    fmt.Println("Usage: cryptoutil kms <subcommand> [options]")
    fmt.Println("Subcommands:")
    fmt.Println("  server     KMS server operations (start, stop, live, ready, init)")
    fmt.Println("  key        Key management operations (generate, list, export, delete) [future]")
    fmt.Println("  barrier    Barrier operations (unseal, seal) [future]")
    fmt.Println("  help       Show this help message")
}

func keyExecute(parameters []string) {
    fmt.Println("Key operations not yet implemented")
    fmt.Println("Future: cryptoutil kms key <generate|list|export|delete> [options]")
    os.Exit(1)
}

func barrierExecute(parameters []string) {
    fmt.Println("Barrier operations not yet implemented")
    fmt.Println("Future: cryptoutil kms barrier <unseal|seal> [options]")
    os.Exit(1)
}
```

### Phase 3: Implement Identity Service Commands

**Create `internal/cmd/cryptoutil/identity/identity.go`** (enhanced dispatcher):

```go
// Copyright (c) 2025 Justin Cranford
//
//

package identity

import (
    "fmt"
    "os"

    authzServer "cryptoutil/internal/cmd/cryptoutil/identity/authz/server"
    idpServer "cryptoutil/internal/cmd/cryptoutil/identity/idp/server"
    rsServer "cryptoutil/internal/cmd/cryptoutil/identity/rs/server"
    spaRpServer "cryptoutil/internal/cmd/cryptoutil/identity/spa-rp/server"
)

// Execute runs identity service group commands.
func Execute(parameters []string) {
    if len(parameters) < 1 {
        printUsage()
        os.Exit(1)
    }

    service := parameters[0]
    serviceParams := parameters[1:]

    switch service {
    case "authz":
        authzServer.Execute(serviceParams)
    case "idp":
        idpServer.Execute(serviceParams)
    case "rs":
        rsServer.Execute(serviceParams)
    case "spa-rp":
        spaRpServer.Execute(serviceParams)
    case "help":
        printUsage()
    default:
        printUsage()
        fmt.Printf("Unknown identity service: %s\n", service)
        os.Exit(1)
    }
}

func printUsage() {
    fmt.Println("Usage: cryptoutil identity <service> <subcommand> [options]")
    fmt.Println("Services:")
    fmt.Println("  authz      OAuth 2.1 Authorization Server")
    fmt.Println("  idp        OIDC Identity Provider")
    fmt.Println("  rs         Resource Server")
    fmt.Println("  spa-rp     SPA Relying Party")
    fmt.Println("")
    fmt.Println("Example:")
    fmt.Println("  cryptoutil identity authz server start --config configs/identity/authz.yml")
}
```

**Create identity server command wrappers**:

```go
// internal/cmd/cryptoutil/identity/authz/server/server.go
package server

import (
    "fmt"
    "os"
)

// Execute runs AuthZ server commands.
func Execute(parameters []string) {
    if len(parameters) < 1 {
        fmt.Println("Usage: cryptoutil identity authz server <start|stop> [options]")
        os.Exit(1)
    }

    subcommand := parameters[0]
    subcommandParams := parameters[1:]

    switch subcommand {
    case "start":
        startAuthZServer(subcommandParams)
    case "stop":
        stopAuthZServer(subcommandParams)
    default:
        fmt.Printf("Unknown authz server subcommand: %s\n", subcommand)
        os.Exit(1)
    }
}

func startAuthZServer(parameters []string) {
    // Call existing cmd/identity/authz/main.go logic
    fmt.Println("Starting OAuth 2.1 Authorization Server...")
    // TODO: Import and call identity authz server start logic
}

func stopAuthZServer(parameters []string) {
    fmt.Println("Stopping OAuth 2.1 Authorization Server...")
    // TODO: Implement stop logic
}
```

**Repeat for `idp/server/`, `rs/server/`, `spa-rp/server/`**

### Phase 4: Create CA Skeleton Commands

**Create `internal/cmd/cryptoutil/ca/ca.go`**:

```go
// Copyright (c) 2025 Justin Cranford
//
//

package ca

import (
    "fmt"
    "os"
)

// Execute runs CA service group commands.
func Execute(parameters []string) {
    if len(parameters) < 1 {
        printUsage()
        os.Exit(1)
    }

    subcommand := parameters[0]
    subcommandParams := parameters[1:]

    switch subcommand {
    case "server":
        serverExecute(subcommandParams)
    case "cert":
        certExecute(subcommandParams)
    case "crl":
        crlExecute(subcommandParams)
    case "help":
        printUsage()
    default:
        printUsage()
        fmt.Printf("Unknown ca subcommand: %s\n", subcommand)
        os.Exit(1)
    }
}

func printUsage() {
    fmt.Println("Usage: cryptoutil ca <subcommand> [options]")
    fmt.Println("Subcommands:")
    fmt.Println("  server     CA server operations (start, stop)")
    fmt.Println("  cert       Certificate operations (issue, renew, revoke, list)")
    fmt.Println("  crl        CRL operations (generate, publish)")
    fmt.Println("  help       Show this help message")
    fmt.Println("")
    fmt.Println("Note: CA service is in skeleton stage. Full implementation pending.")
}

func serverExecute(parameters []string) {
    fmt.Println("CA server not yet implemented")
    os.Exit(1)
}

func certExecute(parameters []string) {
    fmt.Println("Certificate operations not yet implemented")
    os.Exit(1)
}

func crlExecute(parameters []string) {
    fmt.Println("CRL operations not yet implemented")
    os.Exit(1)
}
```

### Phase 5: Update Main CLI Dispatcher

**Update `internal/cmd/cryptoutil/cryptoutil.go`**:

```go
// Copyright (c) 2025 Justin Cranford
//
//

package cmd

import (
    "fmt"
    "os"

    "cryptoutil/internal/cmd/cryptoutil/ca"
    "cryptoutil/internal/cmd/cryptoutil/identity"
    "cryptoutil/internal/cmd/cryptoutil/kms"
)

func Execute() {
    executable := os.Args[0]
    if len(os.Args) < 2 {
        printUsage(executable)
        os.Exit(1)
    }

    command := os.Args[1]
    parameters := os.Args[2:]

    switch command {
    case "kms":
        kms.Execute(parameters)
    case "identity":
        identity.Execute(parameters)
    case "ca":
        ca.Execute(parameters)
    case "server":
        // DEPRECATED: Legacy alias for kms server
        deprecationWarning("server", "kms server")
        kms.Execute(append([]string{"server"}, parameters...))
    case "help":
        printUsage(executable)
    default:
        printUsage(executable)
        fmt.Printf("Unknown command: %s %s\n", executable, command)
        os.Exit(1)
    }
}

func printUsage(executable string) {
    fmt.Printf("Usage: %s <command> [options]\n", executable)
    fmt.Println("Commands:")
    fmt.Println("  kms          Key Management Service operations")
    fmt.Println("  identity     OAuth 2.1 / OIDC Identity Platform operations")
    fmt.Println("  ca           Certificate Authority operations")
    fmt.Println("  help         Show this help message")
    fmt.Println("")
    fmt.Println("Deprecated:")
    fmt.Println("  server       Use 'kms server' instead (legacy alias)")
    fmt.Println("")
    fmt.Println("Examples:")
    fmt.Println("  " + executable + " kms server start --config configs/kms/production.yml")
    fmt.Println("  " + executable + " identity authz server start --config configs/identity/authz.yml")
    fmt.Println("  " + executable + " ca cert issue --profile webserver --cn example.com")
}

func deprecationWarning(oldCommand, newCommand string) {
    fmt.Fprintf(os.Stderr, "⚠️  DEPRECATED: '%s' is deprecated. Use '%s' instead.\n", oldCommand, newCommand)
    fmt.Fprintf(os.Stderr, "   This alias will be removed in cryptoutil v2.0.0\n\n")
}
```

### Phase 6: Testing & Validation

**Test commands**:

```bash
# Test new KMS commands
./cryptoutil kms server start --dev
./cryptoutil kms server stop
./cryptoutil kms server live
./cryptoutil kms server ready

# Test legacy alias (should warn)
./cryptoutil server start --dev  # Should show deprecation warning

# Test identity commands (should show "not yet implemented")
./cryptoutil identity authz server start
./cryptoutil identity idp server start

# Test CA commands (should show "not yet implemented")
./cryptoutil ca server start
./cryptoutil ca cert issue --profile webserver

# Test help
./cryptoutil help
./cryptoutil kms help
./cryptoutil identity help
./cryptoutil ca help
```

**Validation checklist**:

- [ ] `cryptoutil kms server start` works (same as old `server start`)
- [ ] `cryptoutil server start` shows deprecation warning but still works
- [ ] `cryptoutil identity authz server start` shows placeholder message
- [ ] `cryptoutil ca server start` shows placeholder message
- [ ] Help messages show new command structure
- [ ] All existing functionality preserved

### Phase 7: Documentation Updates

**Update README.md**:

```markdown
## CLI Usage

### KMS Server

```bash
# Start KMS server
cryptoutil kms server start --config configs/production/config.yml

# Stop KMS server
cryptoutil kms server stop

# Check liveness
cryptoutil kms server live

# Check readiness
cryptoutil kms server ready
```

### Identity Services

```bash
# OAuth 2.1 Authorization Server
cryptoutil identity authz server start --config configs/identity/authz.yml

# OIDC Identity Provider
cryptoutil identity idp server start --config configs/identity/idp.yml

# Resource Server
cryptoutil identity rs server start --config configs/identity/rs.yml
```

### Certificate Authority (Future)

```bash
# Issue certificate
cryptoutil ca cert issue --profile webserver --cn example.com

# Revoke certificate
cryptoutil ca cert revoke --serial 0x1234567890abcdef --reason keyCompromise
```

### Legacy Commands (Deprecated)

```bash
# OLD (deprecated, shows warning)
cryptoutil server start --config configs/production/config.yml

# NEW (recommended)
cryptoutil kms server start --config configs/production/config.yml
```

```

**Update docs/01-refactor/cli-strategy.md** with implementation details

## Risk Assessment

### Medium Risks

1. **User Experience Change**
   - Mitigation: Deprecation warnings + 12-month transition period + legacy aliases
   - Rollback: Keep old command structure, remove new commands

2. **Identity Command Integration**
   - Mitigation: Reuse existing cmd/identity/* entry points
   - Fallback: Keep placeholder messages until identity implementation ready

### Low Risks

1. **CA Skeleton Commands**
   - Mitigation: Clearly marked as "not yet implemented"
   - No impact on existing functionality

2. **Help Message Clarity**
   - Mitigation: Clear usage examples in help text
   - User feedback collection

## Success Metrics

- [ ] `cryptoutil kms server start` works (same functionality as old `server start`)
- [ ] `cryptoutil server start` shows deprecation warning but works
- [ ] `cryptoutil identity <service> server <start|stop>` structure implemented
- [ ] `cryptoutil ca` skeleton commands created
- [ ] Help messages show new structure with examples
- [ ] README.md updated with new CLI patterns
- [ ] Zero functional regressions (all existing commands work)

## Timeline

- **Phase 1**: Create subdirectories (30 minutes)
- **Phase 2**: Migrate `server` → `kms server` (1 hour)
- **Phase 3**: Implement identity commands (2 hours)
- **Phase 4**: Create CA skeleton (30 minutes)
- **Phase 5**: Update main dispatcher (1 hour)
- **Phase 6**: Testing & validation (1 hour)
- **Phase 7**: Documentation updates (1 hour)

**Total**: 7 hours (1 day)

## Cross-References

- [CLI Strategy Framework](cli-strategy.md) - Command patterns and flag conventions
- [Service Groups Taxonomy](service-groups.md) - KMS, Identity, CA definitions
- [KMS Extraction](kms-extraction.md) - Internal package restructuring
- [Identity Extraction](identity-extraction.md) - Identity module separation
- [CA Preparation](ca-preparation.md) - CA skeleton structure

## Next Steps

After CLI restructure:
1. **Task 14**: CLI help system overhaul
2. **Task 15**: Backward compatibility layer refinement
3. **Task 16-18**: Infrastructure updates (workflows, importas, telemetry)
