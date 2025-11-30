# CLI Backward Compatibility Layer Plan

## Executive Summary

Implement legacy command aliases with deprecation warnings, create migration guide, and define 12-month deprecation timeline.

**Status**: Planning
**Dependencies**: Tasks 13-14 (CLI restructure, help system complete)
**Risk Level**: Low (additive change, preserves existing functionality)

## Current Legacy Commands

From `internal/cmd/cryptoutil/cryptoutil.go`:

```go
switch command {
    case "server":
        // DEPRECATED: Legacy alias for kms server
        deprecationWarning("server", "kms server")
        kms.Execute(append([]string{"server"}, parameters...))
    // ... other commands
}
```

**Legacy command**: `cryptoutil server <start|stop|live|ready|init>`
**New command**: `cryptoutil kms server <start|stop|live|ready|init>`

## Backward Compatibility Strategy

### Goals

1. **Zero Breaking Changes**: All existing commands continue to work
2. **Clear Migration Path**: Users informed of new commands via deprecation warnings
3. **Graceful Transition**: 12-month deprecation period before removal
4. **Documentation**: Migration guide with examples and rationale

### Deprecation Warning System

**Deprecation warning output**:

```
⚠️  DEPRECATED: 'cryptoutil server start' is deprecated.
   Use 'cryptoutil kms server start' instead.
   This alias will be removed in cryptoutil v2.0.0 (scheduled for 2026-06-01).
   Migration guide: https://github.com/justincranford/cryptoutil/docs/MIGRATION.md

[Server starts normally after warning]
```

**Warning characteristics**:

- Printed to stderr (not stdout) so scripts don't break
- Shows new command syntax
- Includes removal date
- Links to migration guide
- Non-blocking (command executes after warning)

## Implementation Phases

### Phase 1: Create Deprecation Warning Infrastructure

**Create `internal/common/cli/deprecation/` package**:

```go
// internal/common/cli/deprecation/deprecation.go

package deprecation

import (
    "fmt"
    "os"
    "time"
)

// Warning represents a deprecation warning.
type Warning struct {
    OldCommand     string    // Deprecated command
    NewCommand     string    // Replacement command
    RemovalVersion string    // Version when removed (e.g., "2.0.0")
    RemovalDate    time.Time // Date when removed
    MigrationGuide string    // URL to migration guide
}

// Print outputs deprecation warning to stderr.
func (w Warning) Print() {
    fmt.Fprintf(os.Stderr, "⚠️  DEPRECATED: '%s' is deprecated.\n", w.OldCommand)
    fmt.Fprintf(os.Stderr, "   Use '%s' instead.\n", w.NewCommand)
    fmt.Fprintf(os.Stderr, "   This alias will be removed in cryptoutil v%s (scheduled for %s).\n",
        w.RemovalVersion, w.RemovalDate.Format("2006-01-02"))
    if w.MigrationGuide != "" {
        fmt.Fprintf(os.Stderr, "   Migration guide: %s\n", w.MigrationGuide)
    }
    fmt.Fprintf(os.Stderr, "\n")
}

// SuppressWarnings checks environment variable to suppress warnings.
func SuppressWarnings() bool {
    return os.Getenv("CRYPTOUTIL_SUPPRESS_DEPRECATION_WARNINGS") == "1"
}
```

### Phase 2: Implement Legacy Command Aliases

**Update `internal/cmd/cryptoutil/cryptoutil.go`**:

```go
// internal/cmd/cryptoutil/cryptoutil.go

package cmd

import (
    "time"

    "cryptoutil/internal/cmd/cryptoutil/ca"
    "cryptoutil/internal/cmd/cryptoutil/identity"
    "cryptoutil/internal/cmd/cryptoutil/kms"
    "cryptoutil/internal/common/cli/deprecation"
    "cryptoutil/internal/common/cli/help"
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
    // New commands (preferred)
    case "kms":
        kms.Execute(parameters)
    case "identity":
        identity.Execute(parameters)
    case "ca":
        ca.Execute(parameters)
    case "version":
        help.PrintVersion()
    case "help":
        printUsage(executable)

    // Legacy commands (deprecated)
    case "server":
        // Warn about deprecation
        if !deprecation.SuppressWarnings() {
            deprecation.Warning{
                OldCommand:     "cryptoutil server",
                NewCommand:     "cryptoutil kms server",
                RemovalVersion: "2.0.0",
                RemovalDate:    time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
                MigrationGuide: "https://github.com/justincranford/cryptoutil/docs/MIGRATION.md",
            }.Print()
        }

        // Route to new command
        kms.Execute(append([]string{"server"}, parameters...))

    default:
        printUsage(executable)
        fmt.Printf("Unknown command: %s %s\n", executable, command)
        os.Exit(1)
    }
}
```

### Phase 3: Create Migration Guide

**Create `docs/MIGRATION.md`**:

```markdown
# Migration Guide: cryptoutil v1.x to v2.0

## Overview

cryptoutil v2.0 introduces a service-group-based CLI structure for better organization and consistency. All existing functionality is preserved, but command names have changed to reflect the new structure.

**Timeline**: Legacy commands deprecated January 2025, removed June 2026 (12-month transition period)

## Breaking Changes

### CLI Command Structure

#### Before (v1.x - Deprecated)

```bash
# Legacy KMS server commands
cryptoutil server start --config configs/production/config.yml
cryptoutil server stop
cryptoutil server live
cryptoutil server ready
cryptoutil server init
```

#### After (v2.0+ - Recommended)

```bash
# New KMS server commands
cryptoutil kms server start --config configs/production/config.yml
cryptoutil kms server stop
cryptoutil kms server live
cryptoutil kms server ready
cryptoutil kms server init
```

### Migration Steps

#### Step 1: Identify Legacy Commands

Search your scripts, CI/CD workflows, and documentation for legacy commands:

```bash
# Search for legacy commands
grep -r "cryptoutil server" .
```

#### Step 2: Update to New Commands

Replace legacy commands with new equivalents:

| Legacy Command | New Command |
|----------------|-------------|
| `cryptoutil server start` | `cryptoutil kms server start` |
| `cryptoutil server stop` | `cryptoutil kms server stop` |
| `cryptoutil server live` | `cryptoutil kms server live` |
| `cryptoutil server ready` | `cryptoutil kms server ready` |
| `cryptoutil server init` | `cryptoutil kms server init` |

#### Step 3: Test New Commands

Verify new commands work in your environment:

```bash
# Test KMS server start
cryptoutil kms server start --dev

# Test KMS server stop
cryptoutil kms server stop
```

#### Step 4: Update Documentation

Update internal documentation, runbooks, and onboarding guides with new command structure.

## Rationale for Change

### Service Group Organization

cryptoutil now supports three service groups:

- **kms**: Key Management Service (hierarchical key management)
- **identity**: OAuth 2.1 / OIDC Identity Platform (authorization, authentication)
- **ca**: Certificate Authority (PKI operations) [future]

Old `cryptoutil server` was ambiguous (which server?). New `cryptoutil kms server` is explicit and consistent.

### Consistency Across Services

All services now follow the same pattern:

```bash
cryptoutil <service-group> <subcommand> [flags]

# Examples
cryptoutil kms server start
cryptoutil identity authz server start
cryptoutil ca server start  # future
```

### Future-Proofing

New structure supports future expansion:

```bash
# Key management operations (future)
cryptoutil kms key generate --type rsa --size 2048
cryptoutil kms key list
cryptoutil kms barrier unseal

# Certificate operations (future)
cryptoutil ca cert issue --profile webserver --cn example.com
cryptoutil ca crl generate
```

## Deprecation Timeline

| Date | Event |
|------|-------|
| 2025-01-01 | Legacy commands deprecated, warnings added |
| 2025-06-01 | 6-month checkpoint, review usage metrics |
| 2025-12-01 | 12-month checkpoint, final migration reminders |
| 2026-06-01 | Legacy commands removed in v2.0.0 |

## Suppressing Deprecation Warnings

For automated scripts where warnings cause issues, suppress warnings via environment variable:

```bash
export CRYPTOUTIL_SUPPRESS_DEPRECATION_WARNINGS=1
cryptoutil server start  # No warning printed
```

**Note**: This is a temporary workaround. Migrate to new commands as soon as possible.

## Getting Help

### CLI Help System

```bash
# Top-level help
cryptoutil help

# Service group help
cryptoutil kms help

# Subcommand help
cryptoutil kms server help

# Flag-level help
cryptoutil kms server start --help
```

### Documentation

- **README**: [README.md](../README.md)
- **CLI Strategy**: [docs/01-refactor/cli-strategy.md](01-refactor/cli-strategy.md)
- **Help System**: [docs/01-refactor/cli-help.md](01-refactor/cli-help.md)

### Support

- **GitHub Issues**: <https://github.com/justincranford/cryptoutil/issues>
- **Migration Questions**: Tag issues with `migration` label

## FAQ

**Q: Why change the CLI structure?**
A: Better organization, consistency across services, and future-proofing for new features.

**Q: Will my existing scripts break?**
A: No. Legacy commands continue to work until v2.0.0 (June 2026). You have 12 months to migrate.

**Q: How do I suppress deprecation warnings?**
A: Set `CRYPTOUTIL_SUPPRESS_DEPRECATION_WARNINGS=1` environment variable.

**Q: What if I find a bug in the new commands?**
A: File a GitHub issue with `bug` and `migration` labels.

**Q: Can I use both old and new commands during migration?**
A: Yes. Legacy commands route to new commands internally, so they're functionally identical.

```

### Phase 4: Update CHANGELOG

**Update `CHANGELOG.md`**:

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Service-group-based CLI structure (`kms`, `identity`, `ca`)
- Comprehensive help system with examples and troubleshooting
- Man pages for all commands
- Deprecation warning system for legacy commands
- Migration guide for CLI changes
- Version information command (`cryptoutil version`)

### Changed
- **BREAKING (v2.0.0)**: `cryptoutil server` → `cryptoutil kms server`
  - Legacy alias supported until v2.0.0 (June 2026)
  - Deprecation warnings added
  - See [MIGRATION.md](docs/MIGRATION.md) for migration guide

### Deprecated
- `cryptoutil server` commands (use `cryptoutil kms server` instead)
  - Removal scheduled for v2.0.0 (June 2026)

## [1.0.0] - 2025-01-01

### Added
- Initial release
- KMS server with hierarchical key management
- PostgreSQL and SQLite backend support
- Multi-layer barrier encryption (unseal modes: 1-of-N, M-of-N)
- OpenAPI-based REST API
- Docker Compose deployment support
- OpenTelemetry observability integration
```

### Phase 5: Testing & Validation

**Test deprecation warnings**:

```bash
# Test legacy command shows warning
cryptoutil server start --dev
# Expected: Warning printed to stderr, then server starts

# Test warning suppression
CRYPTOUTIL_SUPPRESS_DEPRECATION_WARNINGS=1 cryptoutil server start --dev
# Expected: No warning, server starts normally

# Test new command (no warning)
cryptoutil kms server start --dev
# Expected: No warning, server starts normally
```

**Validation checklist**:

- [ ] Legacy `cryptoutil server` commands show deprecation warning
- [ ] Deprecation warning includes new command syntax
- [ ] Deprecation warning includes removal date
- [ ] Deprecation warning links to migration guide
- [ ] Environment variable suppresses warnings
- [ ] New `cryptoutil kms server` commands work without warnings
- [ ] Functionality identical between legacy and new commands

### Phase 6: Update Documentation

**Update README.md**:

```markdown
## CLI Usage

### KMS Server (Recommended)

```bash
# Start KMS server
cryptoutil kms server start --config configs/kms/production.yml

# Stop KMS server
cryptoutil kms server stop

# Check liveness
cryptoutil kms server live

# Check readiness
cryptoutil kms server ready
```

### Legacy Commands (Deprecated)

**⚠️ WARNING: Legacy commands are deprecated and will be removed in v2.0.0**

```bash
# OLD (deprecated, shows warning)
cryptoutil server start --config configs/production/config.yml

# NEW (recommended)
cryptoutil kms server start --config configs/kms/production.yml
```

See [MIGRATION.md](docs/MIGRATION.md) for migration guide.

```

**Update help messages**:

```go
// internal/cmd/cryptoutil/cryptoutil.go

func printUsage(executable string) {
    fmt.Printf("Usage: %s <command> [options]\n", executable)
    fmt.Println("Commands:")
    fmt.Println("  kms          Key Management Service operations")
    fmt.Println("  identity     OAuth 2.1 / OIDC Identity Platform operations")
    fmt.Println("  ca           Certificate Authority operations")
    fmt.Println("  version      Show version information")
    fmt.Println("  help         Show this help message")
    fmt.Println("")
    fmt.Println("Deprecated (use new commands):")
    fmt.Println("  server       ⚠️  DEPRECATED: Use 'kms server' instead")
    fmt.Println("               Scheduled for removal in v2.0.0 (2026-06-01)")
    fmt.Println("")
    fmt.Println("Migration guide: https://github.com/justincranford/cryptoutil/docs/MIGRATION.md")
}
```

### Phase 7: CI/CD Workflow Updates

**Update workflows to use new commands**:

```yaml
# .github/workflows/ci-e2e.yml

steps:
  - name: Start KMS Server
    run: |
      # Use new command syntax (no deprecation warnings in CI)
      ./cryptoutil kms server start --dev &
      sleep 5

  - name: Stop KMS Server
    run: |
      ./cryptoutil kms server stop
```

**Update scripts to use new commands**:

```powershell
# scripts/start-dev.ps1

Write-Host "Starting KMS server in development mode..."
& $PSScriptRoot/../cryptoutil kms server start --dev
```

## Risk Assessment

### Low Risks

1. **Backward Compatibility Preserved**
   - Mitigation: Legacy commands route to new commands internally
   - No functional changes, only command naming

2. **User Confusion During Transition**
   - Mitigation: Clear deprecation warnings with new command syntax
   - Migration guide with examples
   - 12-month transition period

3. **CI/CD Script Updates**
   - Mitigation: Update workflows to use new commands
   - Environment variable to suppress warnings if needed

## Success Metrics

- [ ] Legacy `cryptoutil server` commands show deprecation warning
- [ ] Deprecation warning includes new command, removal date, migration guide link
- [ ] `CRYPTOUTIL_SUPPRESS_DEPRECATION_WARNINGS=1` suppresses warnings
- [ ] New `cryptoutil kms server` commands work without warnings
- [ ] `docs/MIGRATION.md` created with migration examples
- [ ] `CHANGELOG.md` updated with breaking changes and timeline
- [ ] README.md updated to recommend new commands
- [ ] CI/CD workflows updated to use new commands

## Timeline

- **Phase 1**: Create deprecation warning infrastructure (1 hour)
- **Phase 2**: Implement legacy command aliases (30 minutes)
- **Phase 3**: Create migration guide (1.5 hours)
- **Phase 4**: Update CHANGELOG (30 minutes)
- **Phase 5**: Testing & validation (1 hour)
- **Phase 6**: Update documentation (1 hour)
- **Phase 7**: CI/CD workflow updates (30 minutes)

**Total**: 6 hours (1 day)

## Deprecation Timeline (Detailed)

### 2025-01-01: v1.1.0 Release (Deprecation Announced)

- Legacy commands deprecated
- Deprecation warnings added
- Migration guide published
- New commands fully functional

### 2025-03-01: 3-Month Checkpoint

- Review deprecation warning metrics (how many users still using legacy commands)
- Send email to mailing list with migration reminders
- Update documentation with migration examples

### 2025-06-01: 6-Month Checkpoint

- Review usage metrics again
- Identify holdout users and reach out individually
- Blog post: "6 Months Until CLI v2.0"

### 2025-09-01: 9-Month Checkpoint

- Final migration reminders
- Update deprecation warnings with stronger language
- Offer migration assistance to enterprise users

### 2025-12-01: 12-Month Checkpoint

- Last call for migration
- Blog post: "Final Warning: CLI v2.0 in 6 Months"
- Update warnings with "URGENT" prefix

### 2026-06-01: v2.0.0 Release (Legacy Commands Removed)

- Remove all legacy command aliases
- Remove deprecation warning code
- Celebrate clean, consistent CLI structure 🎉

## Cross-References

- [CLI Restructure](cli-restructure.md) - Command structure implementation
- [CLI Help System](cli-help.md) - Help content and navigation
- [CLI Strategy Framework](cli-strategy.md) - Command design patterns
- [Service Groups Taxonomy](service-groups.md) - Service group definitions

## Next Steps

After backward compatibility layer:

1. **Task 16**: Workflow path filter updates
2. **Task 17**: Importas migration
3. **Task 18**: Observability updates
4. **Task 19-20**: Integration testing, documentation finalization
