# Task 10.6: Unified Identity CLI and Profile System

## Task Reflection

### What Went Well

- ✅ **Task 10.5 Endpoints**: Core OAuth/OIDC endpoints implemented, integration tests now passing
- ✅ **Separate Service Binaries**: Individual `cmd/identity/{authz,idp,rs}/main.go` work for standalone launches
- ✅ **Docker Compose**: `identity-compose.yml` orchestrates all services successfully

### At Risk Items

- ❌ **No One-Liner Bootstrap**: Goal of `./identity start --profile demo` not achievable with current structure
- ❌ **Hard-Coded Configurations**: All `cmd/identity/*/main.go` have TODOs for YAML loading, use hard-coded configs
- ❌ **Manual Service Launch**: Users must start 3 separate binaries or use Docker (no unified local CLI)
- ❌ **No Profile System**: Different deployment scenarios (demo, authz-only, full-stack) require manual config editing

### Could Be Improved

- **CLI User Experience**: Current commands require Docker knowledge or multiple terminal windows
- **Configuration Management**: No validation, no defaults, no environment-specific profiles
- **Developer Onboarding**: New developers struggle with complex manual setup vs. simple `./identity start`
- **Testing Workflows**: No easy way to start specific service combinations for targeted testing

### Dependencies and Blockers

- **Dependency on Task 10.5**: Requires working health endpoints for readiness checks
- **Enables Task 10.7**: OpenAPI sync needs running services, unified CLI simplifies testing
- **Enables Tasks 11-15**: Feature development requires easy service startup for testing
- **Unblocks Bootstrap Goal**: This task is CRITICAL for achieving "one-liner bootstrap" objective

---

## Objective

Create a **unified `./identity` CLI tool** using Cobra framework that enables one-liner bootstrap of identity services in various configurations. Implement a profile system allowing users to start service combinations with commands like `./identity start --profile demo` or `./identity start authz idp --config custom.yml`.

**Acceptance Criteria**:

- CLI binary `./identity` built with Cobra framework
- Commands: `start`, `stop`, `status`, `health`, `test`, `logs`
- Flags: `--profile`, `--docker`, `--local`, `--config`, `--background`
- Profile system: `configs/identity/profiles/{demo,authz-only,authz-idp,full-stack,ci}.yml`
- One-liner starts services: `./identity start --profile demo`
- Health check integration: Wait for services ready before returning
- YAML config loading: Remove all hard-coded configs from `cmd/identity/*/main.go`

---

## Historical Context

- **Original Setup**: Three separate binaries (`authz`, `idp`, `rs`) launched individually
- **Docker Workaround**: `identity-compose.yml` added for orchestration but requires Docker knowledge
- **Configuration TODOs**: Line ~22-25 in all `cmd/identity/*/main.go` have "// TODO: Load configuration from YAML file"
- **Bootstrap Goal**: User requested "one-liner bootstrap" as primary objective, currently not achievable

---

## Scope

### In-Scope

1. **Unified CLI Binary** (`cmd/identity/main.go`):
   - Cobra command structure with subcommands
   - Configuration loading from YAML files
   - Service lifecycle management (start, stop, status)
   - Health check integration
   - Logging and telemetry setup

2. **Profile System** (`configs/identity/profiles/`):
   - `demo.yml`: All services (AuthZ + IdP + RS) with demo data
   - `authz-only.yml`: Just Authorization Server for testing
   - `authz-idp.yml`: AuthZ + IdP without Resource Server
   - `full-stack.yml`: All services with production-like config
   - `ci.yml`: Minimal config for CI/CD pipelines

3. **Commands**:
   - `./identity start [services...] [flags]`: Start services (defaults to --profile demo)
   - `./identity stop [services...]`: Stop running services gracefully
   - `./identity status`: Show running services and health status
   - `./identity health`: Check health endpoints and report readiness
   - `./identity test --suite [unit|integration|e2e]`: Run test suites
   - `./identity logs [service] [--follow]`: View service logs

4. **Flags**:
   - `--profile <name>`: Load profile from `configs/identity/profiles/<name>.yml`
   - `--docker`: Use Docker Compose orchestration
   - `--local`: Run services as local processes (default)
   - `--config <path>`: Override with custom config file
   - `--background / -d`: Detach services to background
   - `--wait`: Wait for health checks before returning (default true)
   - `--timeout <duration>`: Health check timeout (default 30s)

5. **YAML Config Loading**:
   - Update `cmd/identity/{authz,idp,rs}/main.go` to load from YAML
   - Remove hard-coded configurations
   - Add config validation on startup
   - Support environment variable overrides

6. **Health Check Integration**:
   - Poll `/health` endpoints after service start
   - Exponential backoff retry (1s, 2s, 4s, 8s, max 30s)
   - Success: Exit 0 with "✅ All services healthy"
   - Failure: Exit 1 with diagnostic output

### Out-of-Scope

- **Service Discovery**: No dynamic service registration/discovery (use fixed ports from profiles)
- **Load Balancing**: No built-in load balancing (defer to Docker/Kubernetes)
- **Secrets Management**: No Vault/secrets integration (use YAML files, environment variables)
- **Hot Reload**: No config hot-reload (require restart for config changes)
- **GUI/TUI**: Command-line only (no terminal UI with curses/bubbletea)

---

## Deliverables

### 1. Unified CLI Binary

**File**: `cmd/identity/main.go` (replaces current structure)

**Structure**:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "github.com/spf13/cobra"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "identity",
        Short: "Unified identity services CLI",
        Long:  "Manage OAuth 2.1 Authorization Server, OIDC Identity Provider, and Resource Server",
    }

    rootCmd.AddCommand(newStartCommand())
    rootCmd.AddCommand(newStopCommand())
    rootCmd.AddCommand(newStatusCommand())
    rootCmd.AddCommand(newHealthCommand())
    rootCmd.AddCommand(newTestCommand())
    rootCmd.AddCommand(newLogsCommand())

    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

**Tests**: `cmd/identity/main_test.go`

- CLI parsing (flags, subcommands)
- Config loading (profile resolution, validation)
- Error handling (missing profile, invalid config)

### 2. Start Command

**File**: `cmd/identity/command_start.go`

**Functionality**:

```bash
# Start all services with demo profile
./identity start --profile demo

# Start specific services
./identity start authz idp --profile ci

# Start with custom config
./identity start --config custom.yml

# Start in Docker
./identity start --profile full-stack --docker

# Start in background (local processes)
./identity start --background
```

**Implementation**:

- Parse flags and service names
- Load profile configuration or custom config
- If `--docker`: Execute `docker compose -f deployments/compose/identity-compose.yml up -d <services>`
- If `--local`: Launch services as child processes with proper environment
- Wait for health checks if `--wait=true`
- Return exit code based on health check results

**Tests**: `cmd/identity/command_start_test.go`

- Start with different profiles
- Start specific service combinations
- Docker vs local mode
- Health check timeout scenarios

### 3. Stop Command

**File**: `cmd/identity/command_stop.go`

**Functionality**:

```bash
# Stop all services
./identity stop

# Stop specific services
./identity stop authz

# Force stop (no graceful shutdown)
./identity stop --force
```

**Implementation**:

- If Docker mode active: `docker compose -f ... down <services>`
- If local processes: Send SIGTERM to PIDs (stored in `~/.identity/pids/*.pid`)
- Wait for graceful shutdown (default 10s timeout)
- If `--force`: Send SIGKILL

**Tests**: `cmd/identity/command_stop_test.go`

- Graceful shutdown
- Force stop
- Stopping non-running services (no error)

### 4. Status Command

**File**: `cmd/identity/command_status.go`

**Functionality**:

```bash
./identity status
# Output:
# SERVICE   STATUS    PID     UPTIME   HEALTH
# authz     running   12345   1h23m    healthy
# idp       running   12346   1h23m    healthy
# rs        running   12347   1h23m    healthy
```

**Implementation**:

- Check PID files or Docker container status
- Query `/health` endpoints
- Format output table (or JSON with `--json` flag)

**Tests**: `cmd/identity/command_status_test.go`

- All services running
- Some services stopped
- No services running

### 5. Health Command

**File**: `cmd/identity/command_health.go`

**Functionality**:

```bash
./identity health
# Polls /health endpoints and reports readiness
# Exit 0 if all healthy, exit 1 if any unhealthy
```

**Implementation**:

- HTTP GET to `https://localhost:{port}/health` for each service
- Parse JSON response: `{"status": "healthy", "database": "ok"}`
- Aggregate results
- Colorized output (green ✅ / red ❌)

**Tests**: `cmd/identity/command_health_test.go`

- All services healthy
- Some services unhealthy
- Services not running

### 6. Test Command

**File**: `cmd/identity/command_test.go`

**Functionality**:

```bash
# Run all tests
./identity test

# Run specific suite
./identity test --suite unit
./identity test --suite integration
./identity test --suite e2e

# Run specific packages
./identity test --package ./internal/identity/authz/...
```

**Implementation**:

- Execute `go test` with appropriate flags
- For e2e tests: Start services if not running, run tests, stop services
- Stream output to stdout
- Return go test exit code

**Tests**: `cmd/identity/command_test_test.go`

- Test execution with different suites
- Service startup for e2e tests

### 7. Logs Command

**File**: `cmd/identity/command_logs.go`

**Functionality**:

```bash
# View logs for all services
./identity logs

# View logs for specific service
./identity logs authz

# Follow logs (tail -f style)
./identity logs --follow
```

**Implementation**:

- If Docker: `docker compose -f ... logs <services>`
- If local: Read log files from `~/.identity/logs/*.log`
- Support `--follow` with `tail -f` behavior

**Tests**: `cmd/identity/command_logs_test.go`

- View logs
- Follow logs
- No logs available (service not started)

### 8. Profile Configuration Files

**Directory**: `configs/identity/profiles/`

**Files**:

- `demo.yml`: Development/demo setup

  ```yaml
  services:
    authz:
      enabled: true
      bind_address: "127.0.0.1:8080"
      database_url: "file:~/.identity/demo.db"
      log_level: "debug"
    idp:
      enabled: true
      bind_address: "127.0.0.1:8081"
      database_url: "file:~/.identity/demo.db"  # Shared with authz
      log_level: "debug"
    rs:
      enabled: true
      bind_address: "127.0.0.1:8082"
      log_level: "debug"
  ```

- `authz-only.yml`: Just Authorization Server

  ```yaml
  services:
    authz:
      enabled: true
      bind_address: "127.0.0.1:8080"
      database_url: "file:~/.identity/authz.db"
      log_level: "info"
    idp:
      enabled: false
    rs:
      enabled: false
  ```

- `authz-idp.yml`: AuthZ + IdP (no Resource Server)
- `full-stack.yml`: All services with production-like settings
- `ci.yml`: Minimal config for CI pipelines (in-memory SQLite)

### 9. Updated Service Main Functions

**Files**: `cmd/identity/{authz,idp,rs}/main.go`

**Changes**:

- Remove hard-coded configuration
- Add YAML config loading via `internal/identity/config` package
- Add config validation on startup
- Support being launched by unified CLI or standalone

**Example** (`cmd/identity/authz/main.go`):

```go
func main() {
    configFile := flag.String("config", "configs/identity/authz.yml", "Configuration file path")
    flag.Parse()

    cfg, err := config.LoadAuthZConfig(*configFile)
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    if err := cfg.Validate(); err != nil {
        log.Fatalf("Invalid configuration: %v", err)
    }

    // ... rest of setup using cfg
}
```

**Tests**: Update existing tests to use config loading

### 10. Configuration Package

**File**: `internal/identity/config/config.go`

**Functionality**:

- Load YAML configuration files
- Merge with environment variable overrides
- Validate configuration (required fields, valid values)
- Provide defaults for optional fields

**Types**:

```go
type ProfileConfig struct {
    Services ServiceConfigs `yaml:"services"`
}

type ServiceConfigs struct {
    AuthZ ServiceConfig `yaml:"authz"`
    IdP   ServiceConfig `yaml:"idp"`
    RS    ServiceConfig `yaml:"rs"`
}

type ServiceConfig struct {
    Enabled      bool   `yaml:"enabled"`
    BindAddress  string `yaml:"bind_address"`
    DatabaseURL  string `yaml:"database_url"`
    LogLevel     string `yaml:"log_level"`
    // ... other settings
}
```

**Tests**: `internal/identity/config/config_test.go`

- Load valid config files
- Environment variable overrides
- Validation errors (missing required fields)
- Default value application

### 11. Documentation

**File**: `docs/identityV2/unified-cli-guide.md`

**Content**:

- CLI command reference with examples
- Profile system explanation
- Configuration file structure
- Local vs Docker mode comparison
- Troubleshooting guide
- Migration guide from old multi-binary setup

**File**: `README.md` (update Getting Started section)

**Changes**:

```markdown
## Quick Start

# Start all services in demo mode
./identity start --profile demo

# Check service status
./identity status

# Run tests
./identity test --suite e2e

# Stop services
./identity stop
```

---

## Validation Criteria

### Automated Tests

- ✅ CLI command tests passing: `go test ./cmd/identity/...`
- ✅ Config loading tests passing: `go test ./internal/identity/config/...`
- ✅ Integration tests work with unified CLI: `go test ./internal/identity/integration/...`
- ✅ Linting passes: `golangci-lint run`

### Manual Testing

1. **Demo Profile Bootstrap**:

   ```bash
   ./identity start --profile demo
   # Expect: All 3 services start, health checks pass, exit 0

   ./identity status
   # Expect: Table showing all services running and healthy

   ./identity health
   # Expect: ✅ All services healthy

   ./identity stop
   # Expect: Graceful shutdown, exit 0
   ```

2. **Specific Service Combinations**:

   ```bash
   ./identity start authz idp --profile ci
   # Expect: Only authz and idp start, rs not started

   ./identity logs authz --follow
   # Expect: Stream authz logs
   ```

3. **Docker Mode**:

   ```bash
   ./identity start --profile full-stack --docker
   # Expect: Docker compose up -d, health checks pass

   ./identity stop --docker
   # Expect: Docker compose down
   ```

### Success Metrics

- One-liner bootstrap works: `./identity start --profile demo` completes in <30s
- Zero hard-coded configs remain in `cmd/identity/*/main.go`
- All profiles load and validate successfully
- Health checks integrate with service startup (wait for ready before returning)

---

## Dependencies

### Depends On (Must Be Complete)

- ✅ **Task 10.5**: Health endpoints required for readiness checks
- ✅ **Task 05**: Storage layer for database configuration

### Enables (Blocked Until Complete)

- **Task 10.7**: OpenAPI sync (easier to test with unified CLI)
- **Tasks 11-15**: Feature development (developers use `./identity start` for testing)
- **Task 17**: Orchestration suite (builds on unified CLI for Docker profiles)

---

## Known Risks

1. **Process Management Complexity**
   - **Risk**: Managing child processes in local mode (PID tracking, signal handling)
   - **Mitigation**: Use `os/exec` with proper context cancellation; store PIDs in `~/.identity/pids/`

2. **Cross-Platform Compatibility**
   - **Risk**: Signal handling differs between Windows and Unix
   - **Mitigation**: Abstract process management in `internal/identity/process/` with OS-specific implementations

3. **Health Check Flakiness**
   - **Risk**: Services may take variable time to start (database migrations, network binding)
   - **Mitigation**: Exponential backoff retry with configurable timeout (default 30s, max 2 minutes)

4. **Configuration Migration**
   - **Risk**: Users with existing hard-coded configs may break on upgrade
   - **Mitigation**: Detect old-style invocations, print migration guide, provide backwards compatibility flag

---

## Implementation Notes

### Phased Approach

1. **Phase 1**: Cobra CLI structure, basic commands (start, stop, status)
2. **Phase 2**: Profile system and YAML loading
3. **Phase 3**: Health check integration
4. **Phase 4**: Remove hard-coded configs from service binaries
5. **Phase 5**: Docker mode integration
6. **Phase 6**: Testing and documentation

### Code Organization

- **CLI Layer** (`cmd/identity/`): Cobra commands, flag parsing
- **Config Layer** (`internal/identity/config/`): YAML loading, validation
- **Process Management** (`internal/identity/process/`): Child process lifecycle
- **Health Checking** (`internal/identity/healthcheck/`): HTTP polling with retry logic

### Testing Strategy

- **Unit Tests**: CLI flag parsing, config loading, validation
- **Integration Tests**: Full service lifecycle (start, health check, stop)
- **E2E Tests**: Use unified CLI in existing e2e test suite

---

## Exit Criteria

- [ ] Unified CLI binary builds successfully
- [ ] All commands implemented (start, stop, status, health, test, logs)
- [ ] Profile system complete with all 5 profiles
- [ ] YAML config loading removes hard-coded configs
- [ ] One-liner bootstrap works: `./identity start --profile demo`
- [ ] All tests passing (unit, integration, e2e)
- [ ] Documentation complete (unified-cli-guide.md, README updates)
- [ ] Linting passes with zero violations
- [ ] Code review complete
- [ ] Commit with message: `feat(identity): complete task 10.6 - unified cli and profiles`

---

## References

- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Viper Configuration](https://github.com/spf13/viper) (optional, for advanced config merging)
- `cmd/identity/{authz,idp,rs}/main.go` - Current hard-coded configurations
- `deployments/compose/identity-compose.yml` - Docker orchestration reference
- `docs/identityV2/task-10.5-authz-idp-endpoints.md` - Health endpoint implementation
