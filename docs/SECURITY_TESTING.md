# DAST Security Testing Guide

**Dynamic Application Security Testing (DAST) for cryptoutil**

## Overview

This guide covers the Dynamic Application Security Testing (DAST) implementation for cryptoutil, including automated security scanning with OWASP ZAP and Nuclei to identify vulnerabilities in the running application.

## What is DAST?

DAST (Dynamic Application Security Testing) performs security testing against a running application to identify vulnerabilities that might not be caught by static analysis. Unlike SAST (Static Application Security Testing), DAST tests the application from the outside, simulating real-world attack scenarios.

## DAST Tools Used

### 1. OWASP ZAP (Zed Attack Proxy)
- **Purpose**: Comprehensive web application security scanner
- **Capabilities**:
  - Full application scan with spider crawling
  - OpenAPI specification-driven API testing
  - OWASP Top 10 vulnerability detection
  - Custom rule configuration

### 2. Nuclei
- **Purpose**: Fast, template-based vulnerability scanner
- **Capabilities**:
  - CVE-based vulnerability detection
  - Security misconfiguration identification
  - Technology fingerprinting
  - High-speed scanning with templates

## Configuration Files

### `.zap/rules.tsv`
Defines custom scanning rules for OWASP ZAP:
- **FAIL**: Critical security issues that must be addressed
- **WARN**: Important findings that should be reviewed
- **IGNORE**: Known false positives or acceptable risks

### Dual API Context Paths
cryptoutil exposes the SAME OpenAPI-defined operations under two distinct context paths with different middleware stacks:

| Context Path | Intended Clients | Middleware Stack (Additive) |
|--------------|------------------|-----------------------------|
| `/browser/api/v1` | Browser / interactive users (Swagger UI, future web apps) | Common core (recover, requestid, logger, telemetry, IP filter, rate limit, cache control) + CORS + CSP/XSS (Helmet) + Additional Security Headers + CSRF |
| `/service/api/v1` | Headless service-to-service clients | Common core only (CORS & CSRF intentionally skipped via `isNonBrowserUserAPIRequestFunc`) |

Implications for DAST:
- Always test BOTH context paths for security header coverage (browser path has more headers set).
- CSRF checks apply only to `/browser/api/v1` requests (service path is exempt).
- CORS preflight and CSP are only visible on the browser path.
- Rate limiting, IP filtering, and core protections apply to both.

Recommended header verification commands:
```powershell
curl -I https://localhost:8080/browser/api/v1/
curl -I https://localhost:8080/service/api/v1/
```

The deprecated `.zap/dast-config.yml` file has been removed—ZAP configuration now lives inline in the GitHub Actions workflow and in `.zap/rules.tsv`.

### Current Browser Security Headers Matrix

`/browser/api/v1` applies enhanced isolation and user-focused hardening headers; `/service/api/v1` deliberately omits browser-only headers to avoid breaking automation and cross-origin service access. All headers below are enforced by middleware (Helmet + custom) unless noted as conditional.

| Header | /browser/api/v1 | /service/api/v1 | Notes / Rationale |
| ------ | ---------------- | ---------------- | ------------------ |
| Strict-Transport-Security | max-age=63072000; includeSubDomains; preload (conditional on HTTPS) | (omitted) | Added only when request over TLS to prevent accidental preload in local HTTP |
| Content-Security-Policy | Present (Helmet managed) | (omitted) | XSS / injection surface reduction; updated with features as needed |
| X-Content-Type-Options | nosniff | nosniff | Prevent MIME sniffing |
| X-Frame-Options | DENY (or via CSP frame-ancestors) | DENY | Clickjacking defense |
| Referrer-Policy | strict-origin-when-cross-origin | strict-origin-when-cross-origin | Limit sensitive referrer leakage |
| Permissions-Policy | Fine-grained empty allowlist (camera=(), geolocation=(), etc.) | (omitted) | Reduce exposed web platform capabilities |
| Cross-Origin-Opener-Policy | same-origin | (omitted) | Enables cross-origin isolation; mitigates popup-based attacks |
| Cross-Origin-Embedder-Policy | require-corp | (omitted) | Paired with COOP to enable powerful isolated contexts |
| Cross-Origin-Resource-Policy | same-origin | (omitted) | Prevent third-party origins embedding private resources |
| X-Permitted-Cross-Domain-Policies | none | none | Blocks legacy Flash/Adobe cross-domain vectors |
| Clear-Site-Data | cache,cookies,storage,executionContexts (POST /logout only) | (not applied) | Forces session state purge on logout |
| Server | Standardized/minimized | Standardized/minimized | Reduce fingerprinting |

Additional notes:
1. COOP + COEP (require-corp) pairing provides stronger isolation; adjust cautiously as it can break cross-origin script/resource loading.
2. Any future addition/removal must be reflected here and validated with regression ZAP scans (pseudo rule IDs 91001–91008) and manual header diff checks.
3. HSTS preload list inclusion should be coordinated with production domain onboarding—avoid premature preload submissions.
4. When expanding CSP, validate against Swagger UI resource loading and future browser client assets.

Validation quick check (PowerShell):
```powershell
curl -I https://localhost:8080/browser/api/v1/ | Select-String -Pattern "strict-transport|content-security|cross-origin|permissions-policy|clear-site|referrer-policy"
curl -I https://localhost:8080/service/api/v1/ | Select-String -Pattern "strict-transport|content-security|cross-origin|permissions-policy|clear-site|referrer-policy"
```

Expected: Browser path shows the extended set; service path shows only the core subset (no CSP, COOP, COEP, CORP, Permissions-Policy, Clear-Site-Data).

## Automated CI/CD Integration

### GitHub Actions Workflow (`.github/workflows/dast.yml`)

**Triggers:**
- Push to main branch
- Pull requests
- Weekly scheduled scans (Sundays at 2 AM UTC)
- Manual dispatch with custom target URL

**Scan Process:**
1. **Application Setup**: Start cryptoutil with test configuration
2. **OWASP ZAP Full Scan**: Comprehensive web application security testing
3. **OWASP ZAP API Scan**: OpenAPI specification-driven API security testing
4. **Nuclei Scan**: Template-based vulnerability scanning
5. **Report Generation**: HTML, JSON, and summary reports

**Test Environment:**
- PostgreSQL database for realistic testing
- Test configuration with simplified authentication
- Health check verification before scanning

## Manual DAST Testing

### Option 1: GitHub Actions Manual Dispatch (Recommended)

**Easiest way to run DAST manually:**

1. **Navigate to GitHub Actions**:
   - Go to your repository: https://github.com/justincranford/cryptoutil
   - Click the **Actions** tab
   - Select **DAST Security Testing** workflow from the left sidebar

2. **Run Workflow Manually**:
   - Click the **Run workflow** dropdown button (top right)
   - Optionally change the target URL (default: `https://localhost:8080`)
   - Click **Run workflow** to start the scan

3. **Monitor Progress**:
   - Watch the workflow execution in real-time
   - Download reports from the **Artifacts** section when complete

### Option 2: Local Development Testing

**For development and testing:**

#### Prerequisites
```powershell
# Install OWASP ZAP
docker pull zaproxy/zap-stable

# Install Nuclei
go install -v github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest
```

### Running Local DAST Scans

**Using the automated scripts (recommended):**

#### PowerShell (Windows)
```powershell
# If execution policy restricts script execution, use:
# powershell -ExecutionPolicy Bypass -File "scripts\dast.ps1" [parameters]

# Run complete DAST suite with default settings
.\scripts\dast.ps1

# Run with custom configuration and port  
.\scripts\dast.ps1 -Config "configs/test/config.yml" -Port 9090

# Run only Nuclei scan, skip ZAP
.\scripts\dast.ps1 -SkipZap

# Save reports to custom directory
.\scripts\dast.ps1 -OutputDir "security-reports"

# Show help
.\scripts\dast.ps1 -Help
```

#### Bash (Linux/macOS)
```bash
# Run complete DAST suite with default settings
./scripts/dast.sh

# Run with custom configuration and port
./scripts/dast.sh --config configs/test/config.yml --port 9090

# Run only Nuclei scan, skip ZAP  
./scripts/dast.sh --skip-zap

# Save reports to custom directory
./scripts/dast.sh --output-dir security-reports

# Show help
./scripts/dast.sh --help
```

**Manual step-by-step process:**

#### 1. Start Application
```powershell
# Build and start cryptoutil
go build -o cryptoutil.exe ./cmd/cryptoutil
./cryptoutil.exe --config configs/local/config.yml &
$APP_PID = $GLOBAL:PID

# Wait for application to be ready
curl http://localhost:9090/readyz
# or
curl http://localhost:9090/healthz
```

#### 2. Run OWASP ZAP Scan
```powershell
# Full security scan
docker run --rm -t -v "${PWD}/dast-reports:/zap/wrk/:rw" zaproxy/zap-stable zap-full-scan.py -t http://localhost:8080 -r zap-full-report.html -J zap-full-report.json

# API-specific scan using OpenAPI spec
docker run --rm -t -v "${PWD}/dast-reports:/zap/wrk/:rw" zaproxy/zap-stable zap-api-scan.py -t https://localhost:8080/ui/swagger/doc.json -f openapi -r zap-api-report.html -J zap-api-report.json
```

#### 3. Run Nuclei Scan
```powershell
# Comprehensive vulnerability scan
nuclei -target http://localhost:8080 -templates cves/,vulnerabilities/,security-misconfiguration/ -json-export dast-reports/nuclei-report.json -stats
```

#### 4. Cleanup
```powershell
# Stop the application process
Stop-Process -Id $APP_PID
```

### Option 3: API/CLI Triggered Scans

**For automation and integration:**

#### Using GitHub CLI
```powershell
# Install GitHub CLI if not already installed
# winget install GitHub.cli

# Run DAST workflow with default settings
gh workflow run dast.yml --ref main

# Run with custom target URL
gh workflow run dast.yml --ref main -f target_url=https://your-server:8080
```

#### Using GitHub REST API
```powershell
# Set your GitHub token
$TOKEN = "your_github_token_here"

# Trigger workflow with default settings
$Headers = @{
    "Authorization" = "token $TOKEN"
    "Accept" = "application/vnd.github.v3+json"
}

$Body = @{
    ref = "main"
    inputs = @{
        target_url = "https://localhost:8080"
    }
} | ConvertTo-Json

Invoke-RestMethod -Uri "https://api.github.com/repos/justincranford/cryptoutil/actions/workflows/dast.yml/dispatches" -Method POST -Headers $Headers -Body $Body -ContentType "application/json"
```

## Security Testing Coverage

### OWASP Top 10 Coverage
- ✅ **A01: Broken Access Control** - Authentication and authorization testing
- ✅ **A02: Cryptographic Failures** - TLS/SSL and encryption testing
- ✅ **A03: Injection** - SQL, Command, LDAP injection testing
- ✅ **A04: Insecure Design** - Architecture and design flaw testing
- ✅ **A05: Security Misconfiguration** - Server and application config testing
- ✅ **A06: Vulnerable Components** - Dependency vulnerability scanning
- ✅ **A07: Authentication Failures** - Session management and auth testing
- ✅ **A08: Software Integrity Failures** - Code injection and tampering
- ✅ **A09: Logging Failures** - Security event logging verification
- ✅ **A10: Server-Side Request Forgery** - SSRF vulnerability testing

### Cryptographic API Specific Tests
- **Key Generation Endpoints**: Input validation and error handling
- **Encryption/Decryption**: Data integrity and confidentiality
- **Digital Signatures**: Signature verification and non-repudiation
- **Certificate Operations**: X.509 certificate handling and validation
- **JWT Operations**: Token security and claim validation

### Security Headers and Configuration
- **CORS Policy**: Cross-origin resource sharing configuration
- **Security Headers**: HSTS, CSP, X-Frame-Options, etc.
- **Rate Limiting**: DoS protection and abuse prevention
- **Error Handling**: Information disclosure prevention

## Interpreting DAST Results

### Severity Levels
- **CRITICAL**: Immediate security risk requiring urgent remediation
- **HIGH**: Significant security vulnerability requiring prompt attention
- **MEDIUM**: Moderate security issue that should be addressed
- **LOW**: Minor security concern or informational finding
- **INFO**: Informational finding for awareness

### Common Findings and Remediation

#### SQL Injection (CRITICAL)
**Finding**: SQL injection vulnerabilities in database queries
**Remediation**: Use parameterized queries and input validation
**Status**: ✅ Prevented by GORM ORM usage

#### Missing Security Headers (MEDIUM)
**Finding**: Missing HTTP security headers
**Remediation**: Implement security headers in server configuration
**Status**: ⚠️ Review current implementation

#### Information Disclosure (LOW-MEDIUM)
**Finding**: Error messages revealing internal information
**Remediation**: Implement generic error responses for external APIs
**Status**: ✅ Handled by structured error responses

#### Weak TLS Configuration (HIGH)
**Finding**: Insecure TLS/SSL configuration
**Remediation**: Enforce TLS 1.2+ with strong cipher suites
**Status**: ✅ Enforced in security instructions

### False Positives
Common false positives specific to cryptoutil:
- **Directory Browsing**: Expected for `/swagger/` endpoints
- **Missing XSS Protection**: Not applicable for API-only service
- **Content Type Issues**: Expected for JWE/JWS operations (text/plain)

## Integration with Existing Security

### Relationship to SAST
- **SAST**: Analyzes source code for security vulnerabilities
- **DAST**: Tests running application for runtime vulnerabilities
- **Combined Coverage**: Comprehensive security testing approach

### Security Scanning Stack
1. **SAST**: CodeQL (GitHub Advanced Security)
2. **DAST**: OWASP ZAP + Nuclei (this implementation)
3. **Dependency Scanning**: GitHub Dependabot
4. **Container Scanning**: Docker Scout
5. **Infrastructure**: Terraform security scanning

## Monitoring and Alerting

### GitHub Actions Integration
- **Workflow Status**: Visible in GitHub Actions tab
- **Security Alerts**: Failed scans trigger notifications
- **Artifact Storage**: Reports stored for 30 days
- **Summary Reports**: Generated in workflow summary

### Scheduled and Manual Scans
- **Scheduled**: Every Sunday at 2 AM UTC (automatic)
- **Manual**: Run workflow manually via GitHub Actions UI anytime
- **API Triggered**: Use GitHub CLI or REST API for automation
- **Target**: Configurable target URL (default: https://localhost:8080)
- **Notifications**: Team notifications for HIGH/CRITICAL findings
- **Trend Analysis**: Historical comparison of security posture

## Best Practices

### Development Workflow
1. **Pre-deployment**: Run DAST scans before production deployment
2. **Regular Scanning**: Weekly automated scans for ongoing monitoring
3. **Rapid Response**: Address CRITICAL/HIGH findings within 24-48 hours
4. **Documentation**: Document all findings and remediation actions

### Configuration Management
1. **Custom Rules**: Maintain ZAP rules for project-specific requirements
2. **Endpoint Coverage**: Ensure all critical endpoints are tested
3. **Payload Customization**: Update custom payloads for new features
4. **False Positive Management**: Regularly review and update ignore rules

### Continuous Improvement
1. **Template Updates**: Keep Nuclei templates updated
2. **Tool Versions**: Regularly update DAST tool versions
3. **Coverage Analysis**: Analyze scan coverage and adjust configuration
4. **Security Training**: Team training on DAST result interpretation

## Troubleshooting

### Common Issues

#### Application Startup Failures
```bash
# Check application logs
./cryptoutil --config configs/test/config.yml --log-level DEBUG

# Verify database connectivity
psql -h localhost -U testuser -d cryptoutil_test -c "SELECT 1;"
```

#### ZAP Scan Timeouts
```bash
# Increase timeout in workflow
cmd_options: '-a -j -T 120'  # Increase to 120 seconds

# Check target accessibility
curl -v http://localhost:9090/readyz
# or
curl -v http://localhost:9090/healthz
```

#### Nuclei Template Issues
```bash
# Update templates
nuclei -update-templates

# Verify template syntax
nuclei -validate -templates /path/to/custom/template.yml
```

### Getting Help
- **OWASP ZAP Documentation**: https://www.zaproxy.org/docs/
- **Nuclei Documentation**: https://docs.projectdiscovery.io/tools/nuclei
- **GitHub Actions Logs**: Check workflow execution logs for detailed error information

## Conclusion

The DAST implementation provides comprehensive dynamic security testing for cryptoutil, covering:
- ✅ **Automated CI/CD Integration**: Seamless security testing in development workflow
- ✅ **Multiple Tool Coverage**: OWASP ZAP and Nuclei for comprehensive scanning
- ✅ **Cryptographic API Focus**: Specialized testing for encryption/decryption endpoints
- ✅ **Custom Configuration**: Project-specific rules and endpoint coverage
- ✅ **Continuous Monitoring**: Weekly automated scans and real-time feedback

This DAST implementation enhances the overall security posture of cryptoutil by identifying runtime vulnerabilities and ensuring ongoing security compliance.
