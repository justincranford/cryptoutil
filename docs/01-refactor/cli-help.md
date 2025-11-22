# CLI Help System Overhaul Plan

## Executive Summary

Design comprehensive help system with service group navigation, examples, common workflows, troubleshooting, and man page integration.

**Status**: Planning
**Dependencies**: Task 13 (CLI restructure complete)
**Risk Level**: Low (documentation enhancement, no functional changes)

## Current Help System Analysis

From `internal/cmd/cryptoutil/cryptoutil.go`:

```go
func Execute() {
    // ... command routing ...
    case "help":
        printUsage(executable)
    default:
        printUsage(executable)
        fmt.Printf("Unknown command: %s %s\n", executable, command)
        os.Exit(1)
}

func printUsage(executable string) {
    fmt.Printf("Usage: %s <command> [options]\n", executable)
    fmt.Println("Commands:")
    fmt.Println("  kms          Key Management Service operations")
    fmt.Println("  identity     OAuth 2.1 / OIDC Identity Platform operations")
    fmt.Println("  ca           Certificate Authority operations")
    fmt.Println("  help         Show this help message")
}
```

**Issues**:
- Minimal help text (command list only)
- No subcommand-specific help
- No examples or common workflows
- No flag documentation
- No troubleshooting guidance
- No version information

## Target Help System Structure

### Help Hierarchy

```
cryptoutil help                          # Top-level help
cryptoutil kms help                      # KMS service group help
cryptoutil kms server help               # KMS server subcommand help
cryptoutil kms server start --help       # KMS server start flag help

cryptoutil identity help                 # Identity service group help
cryptoutil identity authz help           # AuthZ service help
cryptoutil identity authz server help    # AuthZ server subcommand help

cryptoutil ca help                       # CA service group help (future)
cryptoutil ca cert help                  # Certificate operations help
cryptoutil ca cert issue --help          # Certificate issuance flag help
```

### Help Content Components

From [CLI Strategy Framework](cli-strategy.md):

1. **Command Synopsis**: One-line command description
2. **Usage Pattern**: `cryptoutil <service-group> <subcommand> [flags]`
3. **Flags Documentation**: Global flags + subcommand-specific flags with defaults
4. **Examples**: Common use cases with real commands
5. **Related Commands**: Cross-references to related operations
6. **Exit Codes**: List of exit codes and meanings
7. **Configuration**: Reference to config file format
8. **Troubleshooting**: Common errors and solutions

## Implementation Phases

### Phase 1: Create Help Content Structure

**Create `internal/common/cli/help/` package**:

```
internal/common/cli/help/
├── help.go               # Help rendering engine
├── templates.go          # Help text templates
├── examples.go           # Command examples repository
├── troubleshooting.go    # Common errors and solutions
└── version.go            # Version information
```

**Help template structure**:

```go
// internal/common/cli/help/templates.go

package help

type CommandHelp struct {
    Name          string            // Command name
    Synopsis      string            // One-line description
    Usage         string            // Usage pattern
    Description   string            # Detailed description
    GlobalFlags   []FlagHelp        // Global flags (--verbose, --output-format)
    Flags         []FlagHelp        // Command-specific flags
    Examples      []Example         // Usage examples
    RelatedCmds   []string          // Related commands
    ExitCodes     []ExitCode        // Exit codes
    ConfigRef     string            // Config file reference
    Troubleshoot  []Troubleshoot    // Common errors and solutions
}

type FlagHelp struct {
    Name        string   // Flag name (--config)
    Shorthand   string   // Short flag (-c)
    Type        string   // Flag type (string, int, bool)
    Default     string   // Default value
    Description string   // Flag description
    Required    bool     // Required flag?
}

type Example struct {
    Description string   // What the example demonstrates
    Command     string   // Full command
    Output      string   // Expected output (optional)
}

type ExitCode struct {
    Code        int      // Exit code number
    Meaning     string   // What it means
    Example     string   // When it occurs (optional)
}

type Troubleshoot struct {
    Error       string   // Error message
    Cause       string   // Root cause
    Solution    string   // How to fix
}
```

### Phase 2: Implement Top-Level Help

**Create `cryptoutil help` command**:

```go
// internal/cmd/cryptoutil/help.go

package cmd

import (
    "fmt"
    
    "cryptoutil/internal/common/cli/help"
)

func printTopLevelHelp() {
    h := help.CommandHelp{
        Name: "cryptoutil",
        Synopsis: "Cryptographic key management, identity platform, and certificate authority",
        Usage: "cryptoutil <command> [options]",
        Description: `
cryptoutil is a comprehensive cryptographic toolkit providing:
- Key Management Service (KMS): Hierarchical key management with barrier encryption
- Identity Platform: OAuth 2.1 Authorization Server, OIDC Identity Provider, Resource Servers
- Certificate Authority: PKI operations (certificate issuance, CRL management) [future]

Use 'cryptoutil <command> help' for detailed command information.
`,
        GlobalFlags: []help.FlagHelp{
            {
                Name: "--verbose",
                Shorthand: "-v",
                Type: "bool",
                Default: "false",
                Description: "Enable verbose output (shows debug information)",
            },
            {
                Name: "--output-format",
                Shorthand: "-o",
                Type: "string",
                Default: "text",
                Description: "Output format (text, json, yaml)",
            },
            {
                Name: "--config",
                Shorthand: "-c",
                Type: "string",
                Default: "",
                Description: "Configuration file path",
            },
        },
        Examples: []help.Example{
            {
                Description: "Start KMS server with production config",
                Command: "cryptoutil kms server start --config configs/kms/production.yml",
            },
            {
                Description: "Start OAuth 2.1 Authorization Server",
                Command: "cryptoutil identity authz server start --config configs/identity/authz.yml",
            },
            {
                Description: "Get detailed KMS help",
                Command: "cryptoutil kms help",
            },
            {
                Description: "Show version information",
                Command: "cryptoutil version",
            },
        },
        RelatedCmds: []string{
            "cryptoutil kms help",
            "cryptoutil identity help",
            "cryptoutil ca help",
            "cryptoutil version",
        },
        ExitCodes: []help.ExitCode{
            {Code: 0, Meaning: "Success"},
            {Code: 1, Meaning: "General error"},
            {Code: 2, Meaning: "Misuse of command (invalid flags, missing arguments)"},
            {Code: 3, Meaning: "Configuration error (invalid config file)"},
            {Code: 4, Meaning: "Authentication/authorization error"},
            {Code: 5, Meaning: "Resource not found"},
            {Code: 6, Meaning: "Network error"},
            {Code: 7, Meaning: "Cryptographic operation failed"},
            {Code: 8, Meaning: "Database error"},
        },
    }

    help.Render(h)
}
```

**Help rendering engine**:

```go
// internal/common/cli/help/help.go

package help

import "fmt"

func Render(h CommandHelp) {
    fmt.Printf("%s - %s\n\n", h.Name, h.Synopsis)
    fmt.Printf("Usage:\n  %s\n\n", h.Usage)
    
    if h.Description != "" {
        fmt.Printf("Description:\n%s\n\n", h.Description)
    }
    
    if len(h.GlobalFlags) > 0 {
        fmt.Println("Global Flags:")
        renderFlags(h.GlobalFlags)
        fmt.Println()
    }
    
    if len(h.Flags) > 0 {
        fmt.Println("Flags:")
        renderFlags(h.Flags)
        fmt.Println()
    }
    
    if len(h.Examples) > 0 {
        fmt.Println("Examples:")
        renderExamples(h.Examples)
        fmt.Println()
    }
    
    if len(h.RelatedCmds) > 0 {
        fmt.Println("Related Commands:")
        for _, cmd := range h.RelatedCmds {
            fmt.Printf("  %s\n", cmd)
        }
        fmt.Println()
    }
    
    if len(h.ExitCodes) > 0 {
        fmt.Println("Exit Codes:")
        renderExitCodes(h.ExitCodes)
        fmt.Println()
    }
    
    if len(h.Troubleshoot) > 0 {
        fmt.Println("Troubleshooting:")
        renderTroubleshooting(h.Troubleshoot)
        fmt.Println()
    }
}

func renderFlags(flags []FlagHelp) {
    for _, f := range flags {
        flagStr := fmt.Sprintf("  %s", f.Name)
        if f.Shorthand != "" {
            flagStr += fmt.Sprintf(", %s", f.Shorthand)
        }
        fmt.Printf("%s\n", flagStr)
        fmt.Printf("      %s\n", f.Description)
        if f.Default != "" {
            fmt.Printf("      (default: %s)\n", f.Default)
        }
        if f.Required {
            fmt.Printf("      [REQUIRED]\n")
        }
    }
}

func renderExamples(examples []Example) {
    for i, ex := range examples {
        fmt.Printf("  %d. %s\n", i+1, ex.Description)
        fmt.Printf("     $ %s\n", ex.Command)
        if ex.Output != "" {
            fmt.Printf("     %s\n", ex.Output)
        }
        fmt.Println()
    }
}

func renderExitCodes(codes []ExitCode) {
    for _, code := range codes {
        fmt.Printf("  %d: %s\n", code.Code, code.Meaning)
        if code.Example != "" {
            fmt.Printf("     Example: %s\n", code.Example)
        }
    }
}

func renderTroubleshooting(troubles []Troubleshoot) {
    for i, t := range troubles {
        fmt.Printf("  %d. Error: %s\n", i+1, t.Error)
        fmt.Printf("     Cause: %s\n", t.Cause)
        fmt.Printf("     Solution: %s\n\n", t.Solution)
    }
}
```

### Phase 3: Implement KMS Service Group Help

**Create `cryptoutil kms help`**:

```go
// internal/cmd/cryptoutil/kms/help.go

package kms

import (
    "cryptoutil/internal/common/cli/help"
)

func printKMSHelp() {
    h := help.CommandHelp{
        Name: "cryptoutil kms",
        Synopsis: "Key Management Service operations",
        Usage: "cryptoutil kms <subcommand> [flags]",
        Description: `
KMS provides hierarchical key management with multi-layer barrier encryption:
- Root keys protect intermediate keys
- Intermediate keys protect content encryption keys (CEKs)
- Support for RSA, ECDSA, ECDH, EdDSA, AES, HMAC key types

Subcommands:
  server       Start/stop/manage KMS server
  key          Key generation and management [future]
  barrier      Unseal/seal barrier operations [future]
`,
        Examples: []help.Example{
            {
                Description: "Start KMS server in development mode",
                Command: "cryptoutil kms server start --dev",
            },
            {
                Description: "Start KMS server with production config",
                Command: "cryptoutil kms server start --config configs/kms/production.yml",
            },
            {
                Description: "Stop KMS server gracefully",
                Command: "cryptoutil kms server stop",
            },
            {
                Description: "Check KMS server liveness",
                Command: "cryptoutil kms server live",
            },
            {
                Description: "Check KMS server readiness",
                Command: "cryptoutil kms server ready",
            },
        },
        RelatedCmds: []string{
            "cryptoutil kms server help",
            "cryptoutil kms server start --help",
            "cryptoutil identity help",
            "cryptoutil ca help",
        },
        Troubleshoot: []help.Troubleshoot{
            {
                Error: "Failed to unseal barrier: insufficient unseal secrets",
                Cause: "Less than 3 unseal secrets provided for 3-of-5 mode",
                Solution: "Provide at least 3 unseal secret files via --unseal-files flag or config",
            },
            {
                Error: "Database connection failed: connection refused",
                Cause: "PostgreSQL server not running or wrong host/port",
                Solution: "Verify PostgreSQL is running and DSN in config is correct",
            },
            {
                Error: "TLS certificate not found: /etc/certs/kms.crt",
                Cause: "TLS certificate file missing or incorrect path",
                Solution: "Generate TLS certificate or update tls_cert_file path in config",
            },
        },
    }

    help.Render(h)
}
```

### Phase 4: Implement Identity Service Group Help

**Create `cryptoutil identity help`**:

```go
// internal/cmd/cryptoutil/identity/help.go

package identity

import (
    "cryptoutil/internal/common/cli/help"
)

func printIdentityHelp() {
    h := help.CommandHelp{
        Name: "cryptoutil identity",
        Synopsis: "OAuth 2.1 and OIDC Identity Platform operations",
        Usage: "cryptoutil identity <service> <subcommand> [flags]",
        Description: `
Identity platform provides OAuth 2.1 and OIDC services:
- authz: OAuth 2.1 Authorization Server (token issuance, validation)
- idp: OIDC Identity Provider (authentication, user management)
- rs: Resource Server (API protection, token validation)
- spa-rp: SPA Relying Party (browser-based authentication flow)

Services:
  authz        OAuth 2.1 Authorization Server
  idp          OIDC Identity Provider
  rs           Resource Server
  spa-rp       SPA Relying Party
`,
        Examples: []help.Example{
            {
                Description: "Start OAuth 2.1 Authorization Server",
                Command: "cryptoutil identity authz server start --config configs/identity/authz.yml",
            },
            {
                Description: "Start OIDC Identity Provider",
                Command: "cryptoutil identity idp server start --config configs/identity/idp.yml",
            },
            {
                Description: "Start Resource Server",
                Command: "cryptoutil identity rs server start --config configs/identity/rs.yml",
            },
            {
                Description: "Start SPA Relying Party",
                Command: "cryptoutil identity spa-rp server start --config configs/identity/spa-rp.yml",
            },
        },
        RelatedCmds: []string{
            "cryptoutil identity authz help",
            "cryptoutil identity idp help",
            "cryptoutil identity rs help",
            "cryptoutil identity spa-rp help",
            "cryptoutil kms help",
        },
        Troubleshoot: []help.Troubleshoot{
            {
                Error: "Invalid redirect_uri: http://localhost:3000",
                Cause: "Redirect URI uses http instead of https",
                Solution: "Use https redirect URIs in production (http allowed only for localhost in dev)",
            },
            {
                Error: "Token validation failed: signature mismatch",
                Cause: "JWT signed with different key than used for validation",
                Solution: "Ensure all services use same JWK set from KMS for token signing/validation",
            },
        },
    }

    help.Render(h)
}
```

### Phase 5: Implement Flag-Level Help

**Add `--help` flag support**:

```go
// internal/cmd/cryptoutil/kms/server/server.go

package server

import (
    "flag"
    
    "cryptoutil/internal/common/cli/help"
)

func Execute(parameters []string) {
    flags := flag.NewFlagSet("kms server start", flag.ContinueOnError)
    
    // Define flags
    config := flags.String("config", "", "Configuration file path")
    dev := flags.Bool("dev", false, "Development mode (SQLite in-memory)")
    bindAddress := flags.String("bind-address", "0.0.0.0", "Server bind address")
    port := flags.Int("port", 8080, "Server port")
    
    // Custom usage function
    flags.Usage = func() {
        printServerStartHelp()
    }
    
    // Parse flags
    if err := flags.Parse(parameters); err != nil {
        if err == flag.ErrHelp {
            return // Help already printed
        }
        log.Fatal(err)
    }
    
    // ... rest of server start logic
}

func printServerStartHelp() {
    h := help.CommandHelp{
        Name: "cryptoutil kms server start",
        Synopsis: "Start KMS server",
        Usage: "cryptoutil kms server start [flags]",
        Description: `
Starts the KMS server with hierarchical key management and barrier encryption.
Supports PostgreSQL (production) and SQLite (development) backends.
`,
        Flags: []help.FlagHelp{
            {
                Name: "--config",
                Shorthand: "-c",
                Type: "string",
                Default: "",
                Description: "Configuration file path (YAML format)",
            },
            {
                Name: "--dev",
                Type: "bool",
                Default: "false",
                Description: "Development mode (uses SQLite in-memory database)",
            },
            {
                Name: "--bind-address",
                Type: "string",
                Default: "0.0.0.0",
                Description: "Server bind address (0.0.0.0 = all interfaces)",
            },
            {
                Name: "--port",
                Type: "int",
                Default: "8080",
                Description: "Server HTTP port",
            },
            {
                Name: "--tls-cert-file",
                Type: "string",
                Default: "",
                Description: "TLS certificate file path",
                Required: true,
            },
            {
                Name: "--tls-key-file",
                Type: "string",
                Default: "",
                Description: "TLS private key file path",
                Required: true,
            },
        },
        Examples: []help.Example{
            {
                Description: "Start with config file",
                Command: "cryptoutil kms server start --config configs/kms/production.yml",
            },
            {
                Description: "Start in development mode",
                Command: "cryptoutil kms server start --dev",
            },
            {
                Description: "Start with custom port",
                Command: "cryptoutil kms server start --config configs/kms/dev.yml --port 9090",
            },
        },
        ConfigRef: "See configs/kms/production.yml for full configuration example",
    }

    help.Render(h)
}
```

### Phase 6: Add Version Information

**Create `cryptoutil version` command**:

```go
// internal/common/cli/help/version.go

package help

import (
    "fmt"
    "runtime"
)

var (
    Version   = "dev"           // Set via -ldflags at build time
    GitCommit = "unknown"       // Set via -ldflags at build time
    BuildDate = "unknown"       // Set via -ldflags at build time
)

func PrintVersion() {
    fmt.Printf("cryptoutil version %s\n", Version)
    fmt.Printf("  Git commit: %s\n", GitCommit)
    fmt.Printf("  Build date: %s\n", BuildDate)
    fmt.Printf("  Go version: %s\n", runtime.Version())
    fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
```

**Update main CLI**:

```go
// internal/cmd/cryptoutil/cryptoutil.go

func Execute() {
    // ... existing command routing ...
    
    case "version":
        help.PrintVersion()
    
    // ... rest of commands ...
}
```

### Phase 7: Create Man Pages

**Generate man pages from help content**:

```bash
# Create man page directories
mkdir -p docs/man/man1

# Generate man pages from Go code
# Use go-md2man or similar tool to convert help text to man format
```

**Man page structure** (example: `cryptoutil.1`):

```
.TH CRYPTOUTIL 1 "January 2025" "cryptoutil 1.0.0" "User Commands"

.SH NAME
cryptoutil \- Cryptographic key management, identity platform, and certificate authority

.SH SYNOPSIS
.B cryptoutil
<command> [options]

.SH DESCRIPTION
cryptoutil is a comprehensive cryptographic toolkit providing:
.IP \(bu 2
Key Management Service (KMS): Hierarchical key management with barrier encryption
.IP \(bu 2
Identity Platform: OAuth 2.1 Authorization Server, OIDC Identity Provider
.IP \(bu 2
Certificate Authority: PKI operations (future)

.SH COMMANDS
.TP
.B kms
Key Management Service operations
.TP
.B identity
OAuth 2.1 and OIDC Identity Platform operations
.TP
.B ca
Certificate Authority operations (future)
.TP
.B version
Show version information
.TP
.B help
Show help message

.SH GLOBAL FLAGS
.TP
.B \-\-verbose, \-v
Enable verbose output
.TP
.B \-\-output\-format, \-o
Output format (text, json, yaml)
.TP
.B \-\-config, \-c
Configuration file path

.SH EXAMPLES
.TP
Start KMS server:
.B cryptoutil kms server start \-\-config configs/kms/production.yml
.TP
Start OAuth 2.1 Authorization Server:
.B cryptoutil identity authz server start \-\-config configs/identity/authz.yml

.SH EXIT CODES
.TP
.B 0
Success
.TP
.B 1
General error
.TP
.B 2
Misuse of command
.TP
.B 3
Configuration error

.SH FILES
.TP
.I configs/kms/production.yml
KMS production configuration
.TP
.I configs/identity/authz.yml
OAuth 2.1 Authorization Server configuration

.SH SEE ALSO
.BR cryptoutil\-kms (1),
.BR cryptoutil\-identity (1),
.BR cryptoutil\-ca (1)

.SH AUTHORS
Justin Cranford <justin@example.com>
```

**Create subcommand man pages**:
- `cryptoutil-kms.1`
- `cryptoutil-identity.1`
- `cryptoutil-ca.1`

### Phase 8: Testing & Validation

**Help command tests**:

```go
// internal/common/cli/help/help_test.go

func TestTopLevelHelp(t *testing.T) {
    t.Parallel()

    // Capture stdout
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w

    // Print help
    printTopLevelHelp()

    // Restore stdout
    w.Close()
    os.Stdout = old

    // Read captured output
    var buf bytes.Buffer
    io.Copy(&buf, r)
    output := buf.String()

    // Verify help content
    assert.Contains(t, output, "cryptoutil - Cryptographic key management")
    assert.Contains(t, output, "Usage:")
    assert.Contains(t, output, "Global Flags:")
    assert.Contains(t, output, "Examples:")
    assert.Contains(t, output, "Exit Codes:")
}

func TestKMSHelp(t *testing.T) {
    t.Parallel()

    // Similar test for KMS help
    output := captureHelpOutput(printKMSHelp)

    assert.Contains(t, output, "cryptoutil kms")
    assert.Contains(t, output, "Key Management Service")
    assert.Contains(t, output, "Troubleshooting:")
}
```

**Validation checklist**:
- [ ] `cryptoutil help` shows top-level help
- [ ] `cryptoutil kms help` shows KMS service group help
- [ ] `cryptoutil identity help` shows identity service group help
- [ ] `cryptoutil kms server start --help` shows flag-level help
- [ ] `cryptoutil version` shows version information
- [ ] Help text includes examples, troubleshooting, exit codes
- [ ] Man pages generated and readable via `man cryptoutil`

## Documentation Updates

### Update README.md

**Add help system documentation**:

```markdown
## CLI Help System

cryptoutil provides comprehensive help at every level:

```bash
# Top-level help
cryptoutil help

# Service group help
cryptoutil kms help
cryptoutil identity help

# Subcommand help
cryptoutil kms server help

# Flag-level help
cryptoutil kms server start --help

# Version information
cryptoutil version
```

### Man Pages

Install man pages:

```bash
sudo cp docs/man/man1/cryptoutil*.1 /usr/local/share/man/man1/
sudo mandb
```

View man pages:

```bash
man cryptoutil
man cryptoutil-kms
man cryptoutil-identity
```
```

## Risk Assessment

### Low Risks

1. **Help Content Accuracy**
   - Mitigation: Automated tests validate help output
   - No impact on functionality (documentation only)

2. **Man Page Formatting**
   - Mitigation: Use standard man page tools (go-md2man)
   - Fallback: Plain text help always available

3. **Version Information**
   - Mitigation: Build-time ldflags injection
   - Fallback: Show "dev" version if ldflags missing

## Success Metrics

- [ ] `cryptoutil help` shows comprehensive top-level help
- [ ] `cryptoutil <service-group> help` shows service-specific help
- [ ] `cryptoutil <cmd> --help` shows flag-level help with examples
- [ ] `cryptoutil version` shows version, commit, build date
- [ ] Man pages installed and readable via `man cryptoutil`
- [ ] Help text includes examples, troubleshooting, exit codes
- [ ] Automated tests validate help content accuracy

## Timeline

- **Phase 1**: Create help content structure (1 hour)
- **Phase 2**: Implement top-level help (1 hour)
- **Phase 3**: Implement KMS service group help (1 hour)
- **Phase 4**: Implement identity service group help (1 hour)
- **Phase 5**: Implement flag-level help (1 hour)
- **Phase 6**: Add version information (30 minutes)
- **Phase 7**: Create man pages (1 hour)
- **Phase 8**: Testing & validation (1 hour)

**Total**: 7.5 hours (1 day)

## Cross-References

- [CLI Restructure](cli-restructure.md) - Command structure implementation
- [CLI Strategy Framework](cli-strategy.md) - Help content guidelines
- [Service Groups Taxonomy](service-groups.md) - Service descriptions
- [Directory Blueprint](blueprint.md) - Package organization

## Next Steps

After help system overhaul:
1. **Task 15**: Backward compatibility layer refinement
2. **Task 16-18**: Infrastructure updates (workflows, importas, telemetry)
3. **Task 19-20**: Integration testing, documentation finalization
