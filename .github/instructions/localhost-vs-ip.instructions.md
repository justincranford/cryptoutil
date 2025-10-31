---
description: "Instructions for localhost vs 127.0.0.1 usage across different runtime environments"
applyTo: "**"
---
# Localhost vs 127.0.0.1 Usage Guidelines

## Environment-Specific Rules

### Rule Summary Table

| Environment | localhost | 127.0.0.1 | Preferred | Rationale |
|-------------|-----------|-----------|-----------|-----------|
| **Local Windows Dev** | ✅ | ✅ | Either | Both resolve correctly in Windows |
| **GitHub Workflows** | ✅ | ✅ | `127.0.0.1` | Explicit IPv4, no DNS lookup, faster |
| **Act Containers** | ✅ | ✅ | `127.0.0.1` | Explicit IPv4, consistent with workflows |
| **Docker Containers (internal)** | ❌ | ✅ | `127.0.0.1` | localhost may resolve to ::1 (IPv6) |
| **Docker Compose (host→container)** | ✅ | ✅ | `localhost` | User-friendly for browser access |
| **Go Code (bind addresses)** | ❌ | ✅ | `127.0.0.1` | Explicit IPv4 binding, avoids IPv6 |
| **Go Code (database DSN)** | ✅ | ✅ | `localhost` | PostgreSQL lib optimizes localhost |
| **Configuration Files** | ⚠️ | ✅ | `127.0.0.1` | Explicit IPv4, predictable behavior |

### Detailed Environment Guidance

#### 1. Local Windows Development
**Context**: Developer machine running Windows with PowerShell

**Rules**:
- ✅ **localhost works perfectly** - Windows DNS resolves to both IPv4 (127.0.0.1) and IPv6 (::1)
- ✅ **127.0.0.1 works perfectly** - Direct IPv4 address
- 🎯 **Preference**: Use `localhost` in documentation for user-friendliness
- 🎯 **Use 127.0.0.1** in automated scripts for predictability

**Examples**:
```powershell
# User-facing: Use localhost
Start-Process https://localhost:8080/ui/swagger

# Scripts: Use 127.0.0.1
Invoke-WebRequest -Uri "https://127.0.0.1:8080/ui/swagger/doc.json"
```

#### 2. GitHub Actions Workflows (Ubuntu Runners)
**Context**: GitHub-hosted Ubuntu 22.04 runners executing CI/CD workflows

**Rules**:
- ✅ **localhost works** - Resolves to 127.0.0.1 on Ubuntu
- ✅ **127.0.0.1 preferred** - Explicit IPv4, no DNS lookup overhead
- 🎯 **ALWAYS use 127.0.0.1** in workflow bash scripts for consistency

**Why 127.0.0.1**:
- No DNS resolution delay
- Explicit IPv4 (avoids IPv6 ambiguity)
- Consistent with TLS certificate SANs (certificates include 127.0.0.1)
- Faster connectivity checks

**Examples**:
```yaml
# ✅ CORRECT: Use 127.0.0.1 in workflows
env:
  APP_PUBLIC_TARGET_URL: https://127.0.0.1:8080
  APP_PRIVATE_TARGET_URL: http://127.0.0.1:9090

# Connectivity checks
check_endpoint "https://127.0.0.1:8080/ui/swagger/doc.json"
```

#### 3. Act Containers (Local Workflow Testing)
**Context**: nektos/act running workflows in Docker containers on Windows

**Rules**:
- ✅ **localhost works** - Act container resolves localhost
- ✅ **127.0.0.1 preferred** - Consistent with GitHub Actions
- 🎯 **Use 127.0.0.1** to maintain workflow portability

**Rationale**: Act should mimic GitHub Actions behavior exactly

#### 4. Docker Containers (Internal Healthchecks)
**Context**: Commands executed inside Docker containers (healthcheck scripts, CMD, ENTRYPOINT)

**Rules**:
- ❌ **localhost UNRELIABLE** - May resolve to ::1 (IPv6) instead of 127.0.0.1 (IPv4)
- ✅ **127.0.0.1 REQUIRED** - Explicit IPv4 binding
- 🎯 **ALWAYS use 127.0.0.1** in Docker healthcheck commands

**Why 127.0.0.1**:
- Cryptoutil binds to 127.0.0.1 (IPv4 only) by default
- If container resolves localhost→::1, healthcheck fails (connection refused)
- Explicit IPv4 ensures correct protocol family

**Examples**:
```yaml
# ✅ CORRECT: Use 127.0.0.1 in Docker healthchecks
healthcheck:
  test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/livez"]

# ❌ WRONG: localhost may resolve to IPv6
healthcheck:
  test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://localhost:9090/livez"]
```

#### 5. Docker Compose (Host Access from Browser)
**Context**: User accessing services from host machine via browser

**Rules**:
- ✅ **localhost preferred** - User-friendly, works on all platforms
- ✅ **127.0.0.1 acceptable** - Also works but less intuitive
- 🎯 **Use localhost in user-facing documentation**

**Examples**:
```markdown
<!-- User documentation -->
Access Swagger UI at: https://localhost:8080/ui/swagger
Access Grafana at: http://localhost:3000
```

#### 6. Go Code - Bind Addresses (Server Listen)
**Context**: Go application bind addresses in server configuration

**Rules**:
- ❌ **localhost NOT ALLOWED** - May bind to IPv6 (::1) only
- ✅ **127.0.0.1 REQUIRED** - Explicit IPv4 binding
- ✅ **0.0.0.0 for all interfaces** - Public services
- 🎯 **Use magic.IPv4Loopback constant** for consistency

**Why 127.0.0.1**:
- Go's net package interprets localhost differently across platforms
- Explicit IPv4 ensures consistent binding behavior
- TLS certificates include 127.0.0.1 in SANs

**Examples**:
```go
// ✅ CORRECT: Use magic constants for explicit IPv4
import cryptoutilMagic "cryptoutil/internal/common/magic"

bindAddress := cryptoutilMagic.IPv4Loopback // "127.0.0.1"

// ✅ CORRECT: Direct IPv4 loopback
bindAddress := "127.0.0.1"

// ❌ WRONG: localhost may bind IPv6 only on some systems
bindAddress := "localhost"
```

#### 7. Go Code - Database DSN (Client Connection)
**Context**: Database connection strings in Go code

**Rules**:
- ✅ **localhost PREFERRED** - PostgreSQL client library optimizations
- ✅ **127.0.0.1 acceptable** - Works but may miss optimizations
- 🎯 **Use localhost for PostgreSQL DSN**

**Why localhost for databases**:
- PostgreSQL libpq has special optimizations for "localhost"
- May use Unix domain sockets instead of TCP on Linux (faster)
- Standard practice in database connection strings

**Examples**:
```go
// ✅ CORRECT: Use localhost for PostgreSQL DSN
dsn := "postgres://user:pass@localhost:5432/dbname?sslmode=disable"

// ✅ ACCEPTABLE: 127.0.0.1 works but may be slightly slower
dsn := "postgres://user:pass@127.0.0.1:5432/dbname?sslmode=disable"
```

#### 8. Configuration Files (YAML, ENV)
**Context**: Application configuration files

**Rules**:
- ⚠️ **localhost DISCOURAGED** - Ambiguous resolution
- ✅ **127.0.0.1 PREFERRED** - Explicit IPv4
- 🎯 **Exception**: Database URLs can use localhost

**Examples**:
```yaml
# ✅ CORRECT: Explicit IPv4 for bind addresses
bind_public_address: "127.0.0.1"
bind_private_address: "127.0.0.1"

# ✅ CORRECT: localhost acceptable for database DSN
database_url: "postgres://user:pass@localhost:5432/db"

# ✅ CORRECT: Explicit IPs in allowlist
allowed_ips: ["127.0.0.1", "::1"]
```

## Common Patterns by Use Case

### Use Case 1: HTTP Connectivity Verification
```bash
# ✅ Workflows/Scripts: Use 127.0.0.1
curl -skf https://127.0.0.1:8080/ui/swagger/doc.json

# ✅ Docker healthchecks: Use 127.0.0.1
wget --no-check-certificate -q -O /dev/null https://127.0.0.1:9090/livez
```

### Use Case 2: User Documentation
```markdown
# ✅ User-facing docs: Use localhost
Access the application at: https://localhost:8080/ui/swagger
```

### Use Case 3: Go Server Binding
```go
// ✅ Server bind: Use magic.IPv4Loopback
bindAddress := cryptoutilMagic.IPv4Loopback // "127.0.0.1"

// ✅ Database connection: Use localhost
databaseURL := "postgres://user:pass@localhost:5432/db"
```

### Use Case 4: TLS Certificates
```go
// ✅ Include both in certificate SANs
DNSNames: []string{"localhost"}
IPAddresses: []net.IP{
    net.ParseIP("127.0.0.1"),  // IPv4
    net.ParseIP("::1"),         // IPv6
}
```

## Additional Runtime Environments

### Kubernetes Pods
**Context**: Application running inside Kubernetes pod

**Rules**:
- ✅ **127.0.0.1 for pod-internal communication**
- ✅ **Service names for inter-pod communication**
- ❌ **localhost rarely used in K8s context**

**Example**:
```yaml
# Liveness probe
livenessProbe:
  httpGet:
    path: /livez
    port: 9090
    host: 127.0.0.1  # Pod-internal check
```

### WSL2 (Windows Subsystem for Linux)
**Context**: Developer using WSL2 on Windows

**Rules**:
- ✅ **localhost works** - WSL2 bridges networking to Windows
- ✅ **127.0.0.1 works** - Direct IPv4
- 🎯 **Preference depends on context** (same as Windows dev)

## Migration Guide

### Incorrect → Correct Patterns

```yaml
# ❌ BEFORE: Ambiguous localhost in Docker healthcheck
healthcheck:
  test: ["CMD", "wget", "https://localhost:9090/livez"]

# ✅ AFTER: Explicit 127.0.0.1
healthcheck:
  test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/livez"]
```

```go
// ❌ BEFORE: Ambiguous localhost bind
bindAddress := "localhost"

// ✅ AFTER: Explicit IPv4 using magic constant
bindAddress := cryptoutilMagic.IPv4Loopback // "127.0.0.1"
```

```bash
# ❌ BEFORE: localhost in workflow script
curl -k https://localhost:8080/api/health

# ✅ AFTER: 127.0.0.1 for predictability
curl -skf https://127.0.0.1:8080/api/health
```

## Summary: Decision Tree

```
Are you writing code that binds/listens on an address?
├─ YES → Use 127.0.0.1 (explicit IPv4)
│        Exception: Database DSN can use localhost
│
└─ NO → Are you inside a Docker container?
         ├─ YES → Use 127.0.0.1 (avoids IPv6 issues)
         │
         └─ NO → Are you writing CI/CD workflow scripts?
                  ├─ YES → Use 127.0.0.1 (consistency, speed)
                  │
                  └─ NO → Are you writing user-facing documentation?
                           ├─ YES → Use localhost (user-friendly)
                           │
                           └─ NO → Default to 127.0.0.1 (safest choice)
```
