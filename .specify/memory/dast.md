# Dynamic Application Security Testing (DAST) - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/05-03.dast.instructions.md`

## CI-DAST Lessons Learned - CRITICAL

### Variable Expansion in Heredocs - MANDATORY Pattern

**ALWAYS use curly braces `${VAR}` syntax for variable expansion in Bash heredocs**:

```yaml
- name: Generate config
  run: |
    cat > ./configs/test/config.yml <<EOF
    database-url: "postgres://${POSTGRES_USER}:${POSTGRES_PASS}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_NAME}?sslmode=disable"
    bind-public-address: "${APP_BIND_PUBLIC_ADDRESS}"
    bind-public-port: ${APP_BIND_PUBLIC_PORT}
    EOF
```

### CRITICAL Rules

**✅ CORRECT Patterns**:
- ALWAYS use `${VAR}` syntax (curly braces) for explicit variable expansion
- ALWAYS verify generated config files have expanded values (not literal $VAR strings)
- ALWAYS test config generation with `cat config.yml` step in workflow
- ALWAYS use quotes around string values with variables

**❌ WRONG Patterns**:
- NEVER use `$VAR` syntax in heredocs (may write literal "$VAR" to file instead of expanded value)
- NEVER rely on implicit variable expansion behavior (shell-dependent, error-prone)
- NEVER skip verification step after config generation

### Historical Mistake: ci-dast PostgreSQL Configuration

**Problem**: Used `$POSTGRES_USER` instead of `${POSTGRES_USER}` in heredoc

**What Happened**:

```yaml
# ❌ WRONG (what was in ci-dast workflow)
- name: Generate config
  run: |
    cat > config.yml <<EOF
    database-url: "postgres://$POSTGRES_USER:$POSTGRES_PASS@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_NAME"
    EOF
```

**Result**:
1. Heredoc wrote **literal string "$POSTGRES_USER"** to config.yml (NOT expanded value)
2. Application read literal "$POSTGRES_USER" as username
3. Application failed to parse, defaulted to 'root'
4. PostgreSQL rejected connection attempts: **"FATAL: role 'root' does not exist"**
5. Workflow logs showed: `database-url: ***$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_NAME` (literal dollar signs)

**Error Pattern**:
- Frequency: 27+ occurrences over 5 minutes
- Interval: Every 10 seconds (health check retry loop)
- Log correlation: PostgreSQL startup succeeded, connections rejected = credentials issue

**Fix**:

```yaml
# ✅ CORRECT (what should have been used)
- name: Generate config
  run: |
    cat > config.yml <<EOF
    database-url: "postgres://${POSTGRES_USER}:${POSTGRES_PASS}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_NAME}?sslmode=disable"
    EOF

    # Verification step
    echo "Generated config:"
    cat config.yml
```

**Variables Affected**: 7 variables (POSTGRES_USER, POSTGRES_PASS, POSTGRES_HOST, POSTGRES_PORT, POSTGRES_NAME, APP_BIND_PUBLIC_ADDRESS, APP_BIND_PUBLIC_PORT)

---

## Debugging Workflow PostgreSQL Issues

### When PostgreSQL Connection Errors Occur

**Step-by-Step Checklist**:

1. **Download workflow artifacts**
   - Container logs (stdout/stderr)
   - PostgreSQL logs (pg_log/*)
   - Generated config files

2. **Check PostgreSQL logs**
   - Look for `FATAL: role '<username>' does not exist` patterns
   - Check for authentication failures
   - Verify database exists

3. **Verify credentials**
   - Compare workflow env vars vs PostgreSQL service container env vars
   - Ensure POSTGRES_USER matches between service and application
   - Check POSTGRES_PASSWORD is set correctly

4. **Check config generation**
   - Search workflow logs for generated config output
   - Look for `cat config.yml` or similar verification steps
   - Identify literal `$VAR` strings in config (NOT expanded)

5. **Verify variable expansion**
   - Ensure config has expanded values: `postgres://cryptoutil_test:...`
   - NOT literal strings: `postgres://$POSTGRES_USER:...`
   - Check all 7 variables expanded correctly

6. **Compare with working workflows**
   - ci-coverage uses correct `${VAR}` syntax
   - ci-mutation uses correct `${VAR}` syntax
   - ci-race uses correct `${VAR}` syntax
   - Use these as reference implementations

### Evidence-Based Root Cause Analysis

**Pattern Recognition**:
- **Frequency of errors**: 27+ occurrences over 5 minutes = persistent misconfiguration
- **Error interval**: Every 10 seconds = health check retry loop
- **Log correlation**: PostgreSQL startup succeeded + connections rejected = credentials issue
- **Config output analysis**: Literal "$POSTGRES_HOST" in logs = variable expansion issue

**Root Cause Identification**:
1. Application unable to connect to database
2. PostgreSQL rejects connections with "role 'root' does not exist"
3. Application not using configured username 'cryptoutil_test'
4. Config file contains literal "$POSTGRES_USER" string
5. Heredoc variable expansion failed
6. Workflow used `$VAR` instead of `${VAR}` syntax

---

## Preventive Checklist for All 9 Services

### Before Deploying Service or Modifying Workflow

**Configuration Verification**:
- [ ] PostgreSQL credentials use `cryptoutil_test/cryptoutil/cryptoutil_test_password`
- [ ] Workflow env vars match PostgreSQL service container env vars
- [ ] Config generation uses `${VAR}` syntax (curly braces) for ALL variables
- [ ] Config generation includes verification step (`cat config.yml` in workflow logs)
- [ ] Workflow tested locally with Docker Compose before pushing

**Runtime Verification**:
- [ ] Workflow logs reviewed after first run to verify config correctness
- [ ] PostgreSQL service health check configured (`pg_isready`, interval 10s)
- [ ] Application startup includes database connection verification
- [ ] Container logs show successful database connection (NOT "role 'root' does not exist")

**Cross-References**:
- PostgreSQL service requirements: See `.specify/memory/github.md`
- Service template health checks: See `.specify/memory/service-template.md`
- Database configuration patterns: See `.specify/memory/database.md`

---

## Nuclei Vulnerability Scanning

### Prerequisites - Start Services First

**CRITICAL: ALWAYS start cryptoutil services before running nuclei scans**

```bash
# Prerequisites - start services first
docker compose -f ./deployments/compose/compose.yml down -v
docker compose -f ./deployments/compose/compose.yml up -d

# Wait for services to be ready (health checks pass)
sleep 30

# Verify services are ready
docker compose -f ./deployments/compose/compose.yml ps
docker compose exec cryptoutil-sqlite wget --no-check-certificate -q -O - https://127.0.0.1:8080/ui/swagger/doc.json
```

**Why**: Nuclei scans running services, not source code. Services must be running and healthy.

### Manual Nuclei Scan Commands

#### Quick Security Scan (Info/Low Severity)

**Use Case**: Fast initial scan for obvious issues

```bash
nuclei -target https://localhost:8080/ -severity info,low
```

**Characteristics**:
- Fast execution (1-2 minutes)
- Low false positive rate
- Good for quick validation

#### Comprehensive Security Scan (Medium/High/Critical)

**Use Case**: Thorough security audit before release

```bash
nuclei -target https://localhost:8080/ -severity medium,high,critical
```

**Characteristics**:
- Longer execution (5-15 minutes)
- Higher false positive rate (requires manual review)
- Covers serious vulnerabilities

#### Targeted Scans by Vulnerability Type

**Use Case**: Focus on specific vulnerability classes

```bash
# CVEs and known vulnerabilities
nuclei -target https://localhost:8080/ -tags cves,vulnerabilities

# Web application vulnerabilities
nuclei -target https://localhost:8080/ -tags xss,sqli,lfi,rfi

# Misconfigurations
nuclei -target https://localhost:8080/ -tags misconfig,exposure

# Authentication/Authorization
nuclei -target https://localhost:8080/ -tags auth,token,jwt
```

#### Performance-Optimized Scans

**Use Case**: Reduce scan time while maintaining coverage

```bash
# Concurrent requests + rate limiting
nuclei -target https://localhost:8080/ -c 25 -rl 100 -severity high,critical

# Exclude low-value checks
nuclei -target https://localhost:8080/ -severity high,critical -exclude-tags info
```

**Options**:
- `-c 25`: 25 concurrent requests (default 25)
- `-rl 100`: Rate limit 100 requests/second (default 150)
- `-exclude-tags info`: Skip informational findings

### Template Management

#### Update Nuclei Templates

**Frequency**: Before each scan session

```bash
# Update templates to latest version
nuclei -update-templates

# Verify update
nuclei -templates-version
```

**Why**: New vulnerabilities discovered daily, templates updated frequently

#### Check Template Version

```bash
# Show current template version
nuclei -templates-version

# Example output:
# Current Version: v9.6.8
# Latest Version: v9.6.8
```

### CI/CD Integration Pattern

**Workflow Example** (`.github/workflows/ci-dast.yml`):

```yaml
name: CI-DAST

on:
  workflow_dispatch:
  schedule:
    - cron: '0 2 * * 0'  # Weekly Sunday 2am UTC

jobs:
  nuclei-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Start services
        run: |
          docker compose -f ./deployments/compose/compose.yml up -d
          sleep 30

      - name: Run Nuclei scan
        run: |
          docker run --network host \
            projectdiscovery/nuclei:latest \
            -target https://localhost:8080/ \
            -severity medium,high,critical \
            -markdown-export nuclei-report.md

      - name: Upload scan results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: nuclei-scan-results
          path: nuclei-report.md
          retention-days: 30
```

**Key Points**:
- `--network host`: Nuclei container can reach localhost services
- `-markdown-export`: Generate readable report
- `if: always()`: Upload results even if scan finds vulnerabilities
- `retention-days: 30`: Keep scan results for 30 days

---

## OWASP ZAP Dynamic Testing

### ZAP Baseline Scan

**Use Case**: Quick scan for common vulnerabilities

```bash
docker run --network host \
  owasp/zap2docker-stable zap-baseline.py \
  -t https://localhost:8080/ \
  -r zap-baseline-report.html
```

**Characteristics**:
- Spider: Crawls application
- Passive scan: No attacks, observes responses
- Fast: 5-10 minutes
- Low risk: No service disruption

### ZAP Full Scan

**Use Case**: Comprehensive active scanning

```bash
docker run --network host \
  owasp/zap2docker-stable zap-full-scan.py \
  -t https://localhost:8080/ \
  -r zap-full-report.html
```

**Characteristics**:
- Active scan: Sends attack payloads
- Slower: 30-60 minutes
- Higher risk: May trigger rate limits, WAF rules

### ZAP API Scan (OpenAPI)

**Use Case**: Test API endpoints defined in OpenAPI spec

```bash
docker run --network host \
  -v $(pwd)/api:/zap/wrk:rw \
  owasp/zap2docker-stable zap-api-scan.py \
  -t https://localhost:8080/ui/swagger/doc.json \
  -f openapi \
  -r zap-api-report.html
```

**Characteristics**:
- Focused on API endpoints
- Uses OpenAPI spec for complete coverage
- Tests authentication/authorization

---

## Cross-References

**Related Documentation**:
- CI/CD workflows: `.specify/memory/github.md`
- Service health checks: `.specify/memory/service-template.md`
- PostgreSQL configuration: `.specify/memory/database.md`
- Docker Compose: `.specify/memory/docker.md`
- Security patterns: `.specify/memory/security.md`

**Tools**:
- Nuclei: https://github.com/projectdiscovery/nuclei
- OWASP ZAP: https://www.zaproxy.org/
- Docker Compose: https://docs.docker.com/compose/
