# Unified CLI Guide

The Identity system provides a unified command-line interface for managing all three services (Authorization Server, Identity Provider, and Resource Server) through a single binary.

## Installation

Build the unified CLI and service binaries:

```powershell
# Build unified CLI
go build -o bin/identity.exe ./cmd/identity

# Build service binaries
go build -o bin/authz.exe ./cmd/identity/authz
go build -o bin/idp.exe ./cmd/identity/idp
go build -o bin/rs.exe ./cmd/identity/rs
```

## Quick Start

Start all services with the demo profile:

```powershell
./bin/identity start --profile demo
```

Check service health:

```powershell
./bin/identity health
```

View service status:

```powershell
./bin/identity status
```

Stop all services:

```powershell
./bin/identity stop
```

## Commands

### `start` - Launch Services

Start services using a profile or custom configuration:

```powershell
# Start with demo profile (all services)
./bin/identity start --profile demo

# Start with specific profile
./bin/identity start --profile authz-only

# Start with custom config file
./bin/identity start --config /path/to/custom-profile.yml

# Start in background mode (detached process)
./bin/identity start --profile demo --background

# Start and wait for health checks (default: 30s timeout)
./bin/identity start --profile demo --wait --timeout 60s
```

**Flags:**

- `--profile <name>`: Profile name from `configs/identity/profiles/` (default: demo)
- `--config <path>`: Custom configuration file path
- `--docker`: Use Docker Compose instead of local processes (TODO)
- `--local`: Use local processes (default, explicit flag)
- `--background`: Start services in background (detached mode)
- `--wait`: Wait for services to become healthy before returning
- `--timeout <duration>`: Maximum time to wait for health checks (default: 30s)

**Exit Codes:**

- `0`: All enabled services started successfully and health checks passed
- `1`: Configuration loading failed, profile not found, or service startup failed
- `2`: Services started but health checks failed or timed out

### `stop` - Shutdown Services

Gracefully or forcefully stop running services:

```powershell
# Graceful shutdown of all services (default 10s timeout)
./bin/identity stop

# Force shutdown (immediate termination)
./bin/identity stop --force

# Custom graceful timeout
./bin/identity stop --timeout 30s

# Stop specific services
./bin/identity stop authz idp
```

**Flags:**

- `--force`: Immediate termination without graceful shutdown
- `--timeout <duration>`: Maximum time for graceful shutdown (default: 10s)

**Arguments:**

- `[services...]`: Optional list of service names (authz, idp, rs). Defaults to all services.

**Exit Codes:**

- `0`: All services stopped successfully
- `1`: Failed to stop one or more services

### `status` - Service Status Report

Display running/stopped status of all services:

```powershell
# Table output (default)
./bin/identity status

# JSON output for scripting
./bin/identity status --json
```

**Output Formats:**

**Table (default):**

```
SERVICE  STATUS   PID
authz    running  12345
idp      running  12346
rs       stopped  -
```

**JSON:**

```json
[
  {"name": "authz", "running": true, "pid": 12345},
  {"name": "idp", "running": true, "pid": 12346},
  {"name": "rs", "running": false, "pid": 0}
]
```

**Flags:**

- `--json`: Output in JSON format

**Exit Codes:**

- `0`: Always returns success (status reporting never fails)

### `health` - Health Check Polling

Poll service health endpoints and report status:

```powershell
# Check health of all services
./bin/identity health

# Custom timeout per service
./bin/identity health --timeout 10s
```

**Output Example:**

```
Checking service health...
✅ authz: healthy (database: ok)
✅ idp: healthy (database: ok)
❌ rs: unhealthy (connection refused)
```

**Flags:**

- `--timeout <duration>`: Maximum time to wait per service (default: 5s)

**Exit Codes:**

- `0`: All services healthy
- `1`: One or more services unhealthy or unreachable

**Health Check URLs:**

- **AuthZ**: `https://127.0.0.1:8080/health`
- **IdP**: `https://127.0.0.1:8081/health`
- **RS**: `https://127.0.0.1:8082/health`

### `test` - Run Test Suites

**(Placeholder - Not Yet Implemented)**

Execute test suites for the identity system:

```powershell
# Run all tests
./bin/identity test

# Run specific test suite
./bin/identity test --suite unit
./bin/identity test --suite integration
./bin/identity test --suite e2e

# Run tests for specific package
./bin/identity test --package ./internal/identity/storage
```

**Planned Flags:**

- `--suite <type>`: Test suite type (unit, integration, e2e)
- `--package <path>`: Specific package path to test
- `--verbose`: Show verbose test output
- `--coverage`: Generate coverage report

### `logs` - View Service Logs

**(Placeholder - Not Yet Implemented)**

View logs from running services:

```powershell
# View logs for all services
./bin/identity logs

# View logs for specific service
./bin/identity logs authz

# Follow logs in real-time
./bin/identity logs --follow authz

# Show last 50 lines
./bin/identity logs --tail 50 idp
```

**Planned Flags:**

- `--follow`: Continuously stream new log entries (like `tail -f`)
- `--tail <n>`: Show last N lines of logs
- `--since <duration>`: Show logs since duration ago (e.g., 5m, 1h)

**Planned Behavior:**

- **Local processes**: Read from `~/.identity/logs/<service>.log`
- **Docker containers**: Execute `docker compose logs <service>`

## Profiles

Profiles define which services to run and their configurations. They're located in `configs/identity/profiles/`.

### Available Profiles

#### `demo.yml` - Full Stack (All Services)

Starts all three services for local development and testing:

```yaml
services:
  authz:
    enabled: true
    bind_address: "127.0.0.1:8080"
    database_url: "sqlite://file::memory:?cache=shared"
    log_level: "DEBUG"
  idp:
    enabled: true
    bind_address: "127.0.0.1:8081"
    database_url: "sqlite://file::memory:?cache=shared"
    log_level: "DEBUG"
  rs:
    enabled: true
    bind_address: "127.0.0.1:8082"
    database_url: "sqlite://file::memory:?cache=shared"
    log_level: "DEBUG"
```

**Use Case:** Local development, full OAuth 2.1 flow testing

#### `authz-only.yml` - Authorization Server Only

Runs only the Authorization Server for focused development:

```yaml
services:
  authz:
    enabled: true
    bind_address: "127.0.0.1:8080"
    database_url: "sqlite://file::memory:?cache=shared"
    log_level: "DEBUG"
  idp:
    enabled: false
  rs:
    enabled: false
```

**Use Case:** AuthZ server development, token issuance testing

#### `authz-idp.yml` - AuthZ + IdP (No Resource Server)

Runs Authorization Server and Identity Provider for authentication testing:

```yaml
services:
  authz:
    enabled: true
    bind_address: "127.0.0.1:8080"
    database_url: "sqlite://file::memory:?cache=shared"
    log_level: "DEBUG"
  idp:
    enabled: true
    bind_address: "127.0.0.1:8081"
    database_url: "sqlite://file::memory:?cache=shared"
    log_level: "DEBUG"
  rs:
    enabled: false
```

**Use Case:** OAuth 2.1 authorization code flow testing without resource access

#### `full-stack.yml` - Production-Like Configuration

All services with PostgreSQL and INFO logging:

```yaml
services:
  authz:
    enabled: true
    bind_address: "0.0.0.0:8080"
    database_url: "postgresql://user:pass@localhost:5432/authz_db?sslmode=disable"
    log_level: "INFO"
  idp:
    enabled: true
    bind_address: "0.0.0.0:8081"
    database_url: "postgresql://user:pass@localhost:5432/idp_db?sslmode=disable"
    log_level: "INFO"
  rs:
    enabled: true
    bind_address: "0.0.0.0:8082"
    database_url: "postgresql://user:pass@localhost:5432/rs_db?sslmode=disable"
    log_level: "INFO"
```

**Use Case:** Staging environment, production validation

#### `ci.yml` - CI/CD Testing

Minimal configuration for automated testing:

```yaml
services:
  authz:
    enabled: true
    bind_address: "127.0.0.1:18080"
    database_url: "sqlite://file::memory:?cache=shared"
    log_level: "WARN"
  idp:
    enabled: true
    bind_address: "127.0.0.1:18081"
    database_url: "sqlite://file::memory:?cache=shared"
    log_level: "WARN"
  rs:
    enabled: true
    bind_address: "127.0.0.1:18082"
    database_url: "sqlite://file::memory:?cache=shared"
    log_level: "WARN"
```

**Use Case:** GitHub Actions workflows, automated testing

### Creating Custom Profiles

Create a new YAML file in `configs/identity/profiles/`:

```yaml
# configs/identity/profiles/my-profile.yml
services:
  authz:
    enabled: true                              # Start this service
    bind_address: "127.0.0.1:8080"             # Host:port binding
    database_url: "sqlite://file::memory:?cache=shared"  # Database connection
    log_level: "DEBUG"                         # Logging level (DEBUG, INFO, WARN, ERROR)
  idp:
    enabled: false                             # Don't start this service
  rs:
    enabled: false
```

**Validation Rules:**

- At least one service must be enabled
- `bind_address` must be valid host:port format
- `database_url` must be valid SQLite or PostgreSQL DSN
- `log_level` must be one of: DEBUG, INFO, WARN, ERROR

## Configuration

### Service-Specific Configs

Each service binary can also be run standalone with its own config file:

```powershell
# Run Authorization Server directly
./bin/authz.exe --config configs/identity/authz.yml

# Run Identity Provider directly
./bin/idp.exe --config configs/identity/idp.yml

# Run Resource Server directly
./bin/rs.exe --config configs/identity/rs.yml
```

**Individual Service Configs:**

- `configs/identity/authz.yml`: AuthZ server (port 8080, admin 9090)
- `configs/identity/idp.yml`: IdP server (port 8081, admin 9091)
- `configs/identity/rs.yml`: RS server (port 8082, admin 9092)

### PID File Management

Process IDs are stored in `~/.identity/pids/` for lifecycle management:

- `~/.identity/pids/authz.pid`
- `~/.identity/pids/idp.pid`
- `~/.identity/pids/rs.pid`

**Manual Cleanup:**

If services crash or are killed externally, remove stale PID files:

```powershell
Remove-Item -Force ~/.identity/pids/*.pid
```

### Log File Management (Planned)

Service logs will be stored in `~/.identity/logs/`:

- `~/.identity/logs/authz.log`
- `~/.identity/logs/idp.log`
- `~/.identity/logs/rs.log`

## Troubleshooting

### Services Won't Start

**Problem:** `identity start` fails with "failed to start service"

**Solutions:**

1. Check if ports are already in use:

   ```powershell
   netstat -ano | Select-String "8080|8081|8082"
   ```

2. Verify service binaries exist:

   ```powershell
   Test-Path bin/authz.exe, bin/idp.exe, bin/rs.exe
   ```

3. Check profile configuration:

   ```powershell
   Get-Content configs/identity/profiles/demo.yml
   ```

4. Remove stale PID files:

   ```powershell
   Remove-Item -Force ~/.identity/pids/*.pid
   ```

### Health Checks Fail

**Problem:** `identity health` shows services as unhealthy

**Solutions:**

1. Verify services are actually running:

   ```powershell
   ./bin/identity status
   ```

2. Check if services are listening on expected ports:

   ```powershell
   netstat -ano | Select-String "8080|8081|8082"
   ```

3. Manually test health endpoint:

   ```powershell
   curl.exe -k https://127.0.0.1:8080/health
   ```

4. Check service logs for startup errors (once logs command implemented)

### Services Won't Stop

**Problem:** `identity stop` doesn't terminate services

**Solutions:**

1. Use force stop:

   ```powershell
   ./bin/identity stop --force
   ```

2. Manually kill processes (if PID files exist):

   ```powershell
   Get-Content ~/.identity/pids/authz.pid | ForEach-Object { Stop-Process -Id $_ -Force }
   ```

3. Kill by port:

   ```powershell
   # Find process on port 8080
   $pid = (netstat -ano | Select-String "8080" | Select-String "LISTENING")[0].ToString().Split()[-1]
   Stop-Process -Id $pid -Force
   ```

### Profile Not Found

**Problem:** `identity start --profile custom` fails with "profile not found"

**Solutions:**

1. Verify profile file exists:

   ```powershell
   Test-Path configs/identity/profiles/custom.yml
   ```

2. List available profiles:

   ```powershell
   Get-ChildItem configs/identity/profiles/*.yml | Select-Object Name
   ```

3. Use absolute path:

   ```powershell
   ./bin/identity start --config C:/Dev/Projects/cryptoutil/configs/identity/profiles/custom.yml
   ```

### Database Connection Errors

**Problem:** Services fail with "failed to connect to database"

**Solutions:**

1. For SQLite (in-memory), ensure correct DSN format:

   ```yaml
   database_url: "sqlite://file::memory:?cache=shared"
   ```

2. For PostgreSQL, verify database is running:

   ```powershell
   docker compose -f deployments/compose/compose.yml ps postgres
   ```

3. Check PostgreSQL connection:

   ```powershell
   psql -h localhost -p 5432 -U user -d authz_db
   ```

## Common Workflows

### Local Development Workflow

```powershell
# 1. Start all services
./bin/identity start --profile demo --wait

# 2. Verify health
./bin/identity health

# 3. Check status
./bin/identity status

# 4. Make code changes...

# 5. Stop services
./bin/identity stop

# 6. Rebuild binaries
go build -o bin/authz.exe ./cmd/identity/authz
go build -o bin/idp.exe ./cmd/identity/idp
go build -o bin/rs.exe ./cmd/identity/rs

# 7. Restart services
./bin/identity start --profile demo --wait
```

### Testing OAuth 2.1 Flow

```powershell
# Start AuthZ and IdP only
./bin/identity start --profile authz-idp --wait

# Verify both services healthy
./bin/identity health

# Test authorization endpoint
curl.exe -k "https://127.0.0.1:8080/oauth/authorize?client_id=test&response_type=code&redirect_uri=http://localhost:3000/callback"

# Test token endpoint
curl.exe -k -X POST "https://127.0.0.1:8080/oauth/token" `
  -H "Content-Type: application/x-www-form-urlencoded" `
  -d "grant_type=authorization_code&code=ABC123&client_id=test&redirect_uri=http://localhost:3000/callback"

# Cleanup
./bin/identity stop
```

### CI/CD Testing Workflow

```powershell
# Start services with CI profile
./bin/identity start --profile ci --background --wait --timeout 60s

# Run tests
go test ./internal/identity/... -v

# Cleanup
./bin/identity stop --force
```

## Advanced Usage

### Running Multiple Environments Simultaneously

Use custom configs with different ports to run multiple environments:

```yaml
# configs/identity/profiles/env1.yml
services:
  authz:
    enabled: true
    bind_address: "127.0.0.1:8080"
    # ...
```

```yaml
# configs/identity/profiles/env2.yml
services:
  authz:
    enabled: true
    bind_address: "127.0.0.1:9080"
    # ...
```

```powershell
# Start environment 1
./bin/identity start --config configs/identity/profiles/env1.yml --background

# Start environment 2 (different ports)
./bin/identity start --config configs/identity/profiles/env2.yml --background
```

**Note:** Current implementation doesn't support multiple simultaneous instances. PID file collision would occur.

### Scripting with JSON Output

```powershell
# Get status as JSON and parse
$status = ./bin/identity status --json | ConvertFrom-Json

# Check if specific service is running
$authzRunning = ($status | Where-Object { $_.name -eq "authz" }).running

if ($authzRunning) {
    Write-Host "AuthZ is running on PID: $(($status | Where-Object { $_.name -eq 'authz' }).pid)"
} else {
    Write-Host "AuthZ is not running"
}
```

## Future Enhancements

### Planned Features

- **Docker Compose Integration**: `--docker` flag to use containers instead of local processes
- **Log Aggregation**: `logs` command to view service logs from `~/.identity/logs/`
- **Test Runner**: `test` command to execute test suites with coverage reports
- **Config Validation**: `validate` command to check profile syntax and settings
- **Service Reload**: `reload` command to refresh configs without full restart
- **Metrics Endpoint**: Expose Prometheus metrics at `http://127.0.0.1:9090/metrics`
- **Multi-Instance Support**: Run multiple profiles simultaneously with namespace isolation

### Roadmap

1. **Phase 1 (Current)**: Core lifecycle management (start/stop/status/health)
2. **Phase 2**: Docker Compose integration and log viewing
3. **Phase 3**: Test runner and coverage reporting
4. **Phase 4**: Advanced features (reload, metrics, multi-instance)

## See Also

- [Task 10.6 Implementation Details](./task-10.6-unified-cli.md)
- [Identity System README](../../README.md)
- [Configuration Reference](./configuration-reference.md) (TODO)
- [OAuth 2.1 Flow Guide](./oauth-flow-guide.md) (TODO)
