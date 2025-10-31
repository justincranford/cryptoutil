---
description: "Instructions for specialized testing: act workflows and localhost/IP configuration"
applyTo: "**"
---
# Specialized Testing Instructions

## Act Workflow Testing

### CRITICAL: Use cmd/workflow Utility

**ALWAYS use `go run ./cmd/workflow` for running act workflows**

```bash
# Quick DAST scan (3-5 minutes)
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=quick"

# Full DAST scan (10-15 minutes)
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=full"

# Multiple workflows
go run ./cmd/workflow -workflows=e2e,dast

# Available workflows: e2e, dast, sast, robust, quality, load
```

### Timing Expectations
- **Quick profile**: 3-5 minutes (Nuclei + ZAP scans)
- **Full profile**: 10-15 minutes (comprehensive scanning)
- **Deep profile**: 20-25 minutes (exhaustive scanning)

### Common Mistakes to AVOID
❌ **NEVER**: Use `-t` timeout flag or check output too early
❌ **NEVER**: `Start-Sleep -Seconds 60` (way too short)
❌ **NEVER**: `Get-Content -Wait` on log while scan runs (kills process)
❌ **NEVER**: Run act commands directly without monitoring

✅ **ALWAYS**: Use `cmd/workflow` for automated monitoring
✅ **ALWAYS**: Review generated workflow analysis markdown files
✅ **ALWAYS**: Let utility complete before checking outputs

## Localhost vs 127.0.0.1 Usage

### Decision Rules by Environment

| Environment | localhost | 127.0.0.1 | Preferred |
|-------------|-----------|-----------|-----------|
| **Local Windows Dev** | ✅ | ✅ | Either |
| **GitHub Workflows** | ✅ | ✅ | `127.0.0.1` |
| **Act Containers** | ✅ | ✅ | `127.0.0.1` |
| **Docker Containers (internal)** | ❌ | ✅ | `127.0.0.1` |
| **Docker Compose (host→container)** | ✅ | ✅ | `localhost` |
| **Go Code (bind addresses)** | ❌ | ✅ | `127.0.0.1` |
| **Go Code (database DSN)** | ✅ | ✅ | `localhost` |

### Quick Reference

**Docker Healthchecks (ALWAYS use 127.0.0.1):**
```yaml
healthcheck:
  test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/livez"]
```

**Go Server Binding (ALWAYS use 127.0.0.1):**
```go
bindAddress := cryptoutilMagic.IPv4Loopback // "127.0.0.1"
```

**Database DSN (Use localhost):**
```go
dsn := "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
```

**CI/CD Workflows (Prefer 127.0.0.1):**
```bash
curl -skf https://127.0.0.1:8080/ui/swagger/doc.json
```

### Why 127.0.0.1 in Docker Containers
- Cryptoutil binds to `127.0.0.1` (IPv4 only)
- If container resolves `localhost→::1`, healthcheck fails (IPv6 connection refused)
- Explicit IPv4 ensures correct protocol family
