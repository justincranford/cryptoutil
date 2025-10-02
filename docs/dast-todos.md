# DAST TODO List - Deep Analysis and Remediation Tasks

**Document Status**: Discovery Phase  
**Created**: 2025-09-30  
**Purpose**: Comprehensive task list for addressing DAST workflow issues and security findings

> Maintenance Guideline: If a file/config/feature is removed or a decision makes a task permanently obsolete, DELETE its tasks and references here immediately. Keep only (1) active remediation work, (2) still-relevant observations, (3) forward-looking backlog items. Historical context belongs in commit messages or durable docs, not this actionable list.

---

## Executive Summary

This document contains a comprehensive analysis of:
1. Nuclei security scan findings requiring remediation
2. GitHub Actions workflow compatibility issues with `act` (local testing)
3. OWASP ZAP configuration analysis and potential conflicts
4. Security header implementation gaps

### Current Sprint Priority Short List (Ordered)

> Source: Latest act run & user directive (2025-10-01). Execute in order; update after each completion.

1. Replace references to non-existing `/browser/api/v1/health` with actual health endpoints: `/readyz` (readiness) and `/livez` (liveness). (Note: Private fiber exposes `/livez`, `/readyz`; there is no `/livenessz` endpoint in code. Public endpoints intentionally do not expose health.)
2. Make Security Summary Dynamic (only list scanners actually executed; currently claims ZAP scans ran while steps are commented out).
3. Port / Service Noise Suppression (exclude infrastructure ports 22,111,5432 from Nuclei; eliminate `rpcbind-portmapper-detect`, `pgsql-detect`, SSH noise).
4. Header Artifact Export (capture canonical curl header sets for Swagger UI, Browser API context, Service API context to correlate with Nuclei false negatives).
5. Dummy Token Documentation (document rationale for dummy `GITHUB_TOKEN` in act runs and guard conditions).
6. Adopt Caching for Go Dependencies (already using setup-go cache; add explicit Go module + Nuclei template caching and planned ZAP image/template caching strategy).

---

**Priority Levels**:
- 🔴 **CRITICAL**: Security vulnerabilities requiring immediate attention
- 🟠 **HIGH**: Important issues affecting security posture or workflow reliability
- 🟡 **MEDIUM**: Configuration improvements and optimizations
- 🟢 **LOW**: Nice-to-have improvements and documentation updates

---

## Section 1: Nuclei Security Findings (from dast-github-action-nuclei.log)

### 1.1 Security Headers - Missing HTTP Security Headers (🟠 HIGH)

**Finding**: Multiple missing HTTP security headers detected by Nuclei scan

**Missing Headers**:
- `Strict-Transport-Security` (HSTS)
- `Content-Security-Policy` (CSP)
- `Permissions-Policy`
- `X-Frame-Options`
- `X-Content-Type-Options`
- `X-Permitted-Cross-Domain-Policies`
- `Referrer-Policy`
- `Cross-Origin-Embedder-Policy` (COEP)
- `Clear-Site-Data`
- `Cross-Origin-Opener-Policy` (COOP)
- `Cross-Origin-Resource-Policy` (CORP)

**Current State**:
- ✅ CSP is implemented in `publicBrowserXSSMiddlewareFunction` via Helmet middleware
- ✅ `X-Frame-Options` is set to "DENY" in Helmet config
- ✅ `X-Content-Type-Options: nosniff` is set in `publicBrowserAdditionalSecurityHeadersMiddleware`
- ✅ `Referrer-Policy: same-origin` is set in Helmet config
- ⚠️ HSTS is conditionally set (only for HTTPS protocol)
- ❌ Several headers are missing or not visible to scanner

**Tasks**:

#### Task 1.1.1: Verify Current Security Header Implementation (🟡 MEDIUM)
- **Description**: Test and verify which security headers are actually being sent by the application
- **Action Items**:
  - Use `curl -I https://localhost:8080/ui/swagger/` to inspect Swagger UI static response headers
  - Use `curl -I https://localhost:8080/browser/api/v1/` to inspect Browser API headers (browser middlewares applied: CORS, CSP/XSS, additional security headers, CSRF)
  - Use `curl -I https://localhost:8080/service/api/v1/` to inspect Service API headers (no browser-only middlewares: CORS, CSRF skipped via `isNonBrowserUserAPIRequestFunc`)
  - Compare actual headers with Nuclei scan results to identify discrepancies
  - Document which headers are working and which are missing
- **Files**: `internal/server/application/application_listener.go` (lines 550-680)
- **Expected Outcome**: Clear understanding of which headers are missing vs. not detected by scanner, across BOTH context paths (`/browser/api/v1` with browser middlewares, `/service/api/v1` without them)

#### Task 1.1.2: Review Middleware Execution Order (🟡 MEDIUM)
- **Description**: Helmet and security middleware may not be applied to all routes
- **Action Items**:
  - Verify middleware execution order in `publicMiddlewares` array
  - Check if security headers are applied to Swagger UI routes (`/ui/swagger/*`)
  - Check if security headers are applied to service API routes (`/service/api/v1/*`)
  - Ensure middleware is not being skipped by `isNonBrowserUserAPIRequestFunc`
- **Files**: `internal/server/application/application_listener.go` (lines 200-220)
- **Root Cause**: Middleware might be selectively applied, causing scanner to miss headers on some endpoints

- **References**:
  - [OWASP Secure Headers Project](https://owasp.org/www-project-secure-headers/)
  - [MDN Security Headers](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers#security)

#### Task 1.1.4: Configure Security Headers for Swagger UI (🟡 MEDIUM)
- **Description**: Ensure Swagger UI routes receive appropriate security headers
- **Action Items**:
  - Verify Swagger UI static content handler includes security headers
  - Test if CSP allows Swagger UI to function correctly
  - Adjust CSP if Swagger UI breaks due to restrictive policy
  - Document any CSP exceptions required for Swagger UI
- **Files**: `internal/server/application/application_listener.go` (Swagger UI setup around line 240-280)
- **Note**: Balance between security and Swagger UI functionality

#### Task 1.1.5: Update ZAP Rules for Security Headers (🟢 LOW)
- **Description**: Update ZAP rules.tsv to expect properly configured security headers
- **Action Items**:
  - Change security header rules from IGNORE to WARN or FAIL
  - Add specific rules for verifying HSTS, CSP, and other critical headers
  - Update comments to reflect expected security header configuration
- **Files**: `.zap/rules.tsv` (lines 10-20)
- **Expected Outcome**: ZAP will flag missing security headers as findings

### 1.2 TLS/SSL Configuration Issues (🟡 MEDIUM)

**Finding**: Nuclei detected TLS 1.2 and TLS 1.3, untrusted root certificate (expected for self-signed certs)

**Tasks**:

#### Task 1.2.1: Document TLS Configuration Standards (🟢 LOW)
- **Description**: Create documentation for TLS configuration in production vs. test environments
- **Action Items**:
  - Document that self-signed certificates are expected in test/dev environments
  - Document minimum TLS version requirement (TLS 1.2 minimum, prefer TLS 1.3)
  - Add instructions for using trusted certificates in production
  - Update SECURITY_TESTING.md with TLS configuration guidance
- **Files**: `docs/SECURITY_TESTING.md`
- **Related**: Certificate files (`tls_*.pem`) in workspace root

#### Task 1.2.2: Add ZAP Rule for Untrusted Certificates in Test Environments (🟢 LOW)
- **Description**: Configure ZAP to ignore untrusted certificate warnings in test environments
- **Action Items**:
  - Add specific rule in `.zap/rules.tsv` for untrusted-root-certificate in localhost
  - Set severity to INFO or IGNORE for test environment scans
  - Document that production scans should FAIL on untrusted certificates
- **Files**: `.zap/rules.tsv`

### 1.3 Infrastructure Findings - Non-Application Issues (🟢 LOW)

**Findings**: Nuclei detected infrastructure services not related to cryptoutil application:
- `rpcbind-portmapper-detect` on localhost:111 (RPC service)
- `ssh-*` findings on localhost:22 (SSH service)
- `pgsql-detect` on localhost:5432 (PostgreSQL database)
- `cookies-without-httponly` for `_csrf` cookie

**Tasks**:

#### Task 1.3.1: Scope Nuclei Scans to Application Only (🟡 MEDIUM)
- **Description**: Configure Nuclei to scan only cryptoutil application endpoints, not infrastructure
- **Action Items**:
  - Update `dast.yml` Nuclei flags to exclude infrastructure ports
  - Add `-exclude-ports 22,111,5432` to Nuclei command
  - Focus scan on application ports 8080 (public HTTPS) and 9090 (private HTTP)
  - Document that infrastructure services should be scanned separately
- **Files**: `.github/workflows/dast.yml` (line 149)
- **Current**: `flags: "-c 24 -rl 200 -timeout 5 -stats"`
- **Proposed**: `flags: "-c 24 -rl 200 -timeout 5 -stats -exclude-ports 22,111,5432"`

#### Task 1.3.2: Review CSRF Cookie HttpOnly Configuration (🟡 MEDIUM)
- **Description**: Nuclei flagged `_csrf` cookie without HttpOnly flag
- **Action Items**:
  - Investigate if CSRF cookie needs to be accessible to JavaScript (required for Swagger UI?)
  - If not required by JS, enable HttpOnly flag in CSRF middleware configuration
  - Document rationale for HttpOnly flag decision
  - Test Swagger UI functionality with HttpOnly CSRF cookie
- **Files**: `internal/server/application/application_listener.go` (CSRF middleware config)
- **Related**: Check `publicBrowserCSRFMiddlewareFunction` implementation
- **Note**: CSRF tokens often need JavaScript access for SPA/AJAX requests

#### Task 1.3.3: Document Infrastructure Service Security (🟢 LOW)
- **Description**: Add documentation noting that infrastructure services are outside application scope
- **Action Items**:
  - Update `docs/SECURITY_TESTING.md` to clarify DAST scope boundaries
  - Note that SSH, PostgreSQL, and RPC services should be secured at infrastructure level
  - Reference deployment security documentation for infrastructure hardening
- **Files**: `docs/SECURITY_TESTING.md`

### 1.4 External Service Interaction Finding (🟢 LOW)

**Finding**: `[external-service-interaction]` detected - Host Header Injection test

**Tasks**:

#### Task 1.4.1: Review Host Header Validation (🟢 LOW)
- **Description**: Ensure application validates and restricts Host header to prevent Host Header Injection
- **Action Items**:
  - Review if Fiber framework validates Host header by default
  - Add explicit Host header validation middleware if needed
  - Test with malicious Host headers to verify protection
  - Document Host header security controls
- **Files**: `internal/server/application/application_listener.go`
- **References**: [OWASP Host Header Injection](https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/07-Input_Validation_Testing/17-Testing_for_Host_Header_Injection)

---

## Section 2: GitHub Actions `act` Compatibility Issues (from dast-act-localhost.log)

### 2.1 Artifact Upload Failure with `act` (🟠 HIGH)

**Finding**: Step "GitHub Workflow artifacts" fails with `Unable to get the ACTIONS_RUNTIME_TOKEN env variable`

**Current Conditional**: `if: ${{ !env.ACT }}`

**Root Cause**: The condition `${{ !env.ACT }}` is not working as expected in `act` local execution

**Tasks**:

#### ✅ Task 2.1.1: Fix Conditional Expression for `act` Detection (🟠 HIGH) - COMPLETED
- **Description**: The current conditional is not properly detecting `act` environment
- **Resolution**: Fixed conditional expressions and added graceful fallback
  - Changed `if: ${{ !env.ACT }}` to `if: ${{ env.ACT != 'true' }}`
  - Added `continue-on-error: ${{ env.ACT == 'true' }}` for artifact upload
  - Added debug step to show environment variables for troubleshooting
  - Added local artifact collection (`./dast-reports/`) for act runs
- **Files**: `.github/workflows/dast.yml` (commit 45e7180)
- **Status**: ✅ Completed - Act compatibility issues resolved

#### Task 2.1.2: Alternative - Use GitHub Context Instead of env.ACT (🟠 HIGH)
- **Description**: Use GitHub-native context variables that `act` respects
- **Action Items**:
  - Try `if: github.event_name != 'push' && github.event_name != 'pull_request'` (excludes act)
  - Try `if: runner.environment == 'github-hosted'` (may not work in act)
  - Document which context variables are reliable with `act`
  - Test each approach locally with `act`
- **Files**: `.github/workflows/dast.yml` (line 152)
- **References**: [act documentation on environment detection](https://github.com/nektos/act#skipping-steps)

#### Task 2.1.3: Suppress Artifact Upload Errors in `act` (🟡 MEDIUM)
- **Description**: Use `continue-on-error` for artifact upload step when running in `act`
- **Action Items**:
  - Add `continue-on-error: ${{ env.ACT == 'true' }}` to artifact upload step
  - Test if this allows `act` to continue past artifact upload failure
  - Document that artifact upload is expected to fail in local `act` runs
- **Files**: `.github/workflows/dast.yml` (lines 151-156)
- **Proposed**:
  ```yaml
  - name: GitHub Workflow artifacts
    if: always()
    continue-on-error: ${{ env.ACT == 'true' }}
    uses: actions/upload-artifact@v4
    with:
      name: nuclei.log
      path: nuclei.log
  ```

#### Task 2.1.4: Create Local Artifact Collection Alternative (🟢 LOW)
- **Description**: For `act` runs, copy artifacts to local directory instead of uploading
- **Action Items**:
  - Add conditional step that runs only in `act`: `if: ${{ env.ACT == 'true' }}`
  - Copy `nuclei.log`, `nuclei.sarif` to `./dast-reports/` directory
  - Ensure `./dast-reports/` is git-ignored
  - Document that local artifacts are saved to `./dast-reports/` when using `act`
- **Files**: `.github/workflows/dast.yml` (add new step after line 156)
- **Proposed**:
  ```yaml
  - name: Save artifacts locally (act only)
    if: ${{ env.ACT == 'true' }}
    run: |
      mkdir -p ./dast-reports
      cp nuclei.log ./dast-reports/ || true
      cp nuclei.sarif ./dast-reports/ || true
      echo "Artifacts saved to ./dast-reports/"
  ```

### 2.2 SARIF Upload Conditional Check (🟢 LOW)

**Current**: Step "GitHub Security Dashboard Alerts update" has compound conditional

**Tasks**:

#### Task 2.2.1: Verify SARIF Upload Conditional Works with `act` (🟢 LOW)
- **Description**: Ensure SARIF upload is properly skipped in `act` runs
- **Action Items**:
  - Verify that `${{ !env.ACT }}` works correctly for SARIF upload step
  - Check if `steps.nuclei_scan.outputs.sarif_exists` is accessible in `act`
  - Test both conditions independently
  - Document expected behavior in `act` vs. GitHub Actions
- **Files**: `.github/workflows/dast.yml` (lines 158-163)
- **Current**: `if: ${{ !env.ACT }} && steps.nuclei_scan.outputs.sarif_exists == 'true'`

#### Task 2.2.2: Add Debug Output for Conditional Variables (🟢 LOW)
- **Description**: Add debug step to show values of conditional variables for troubleshooting
- **Action Items**:
  - Add step before artifact upload that prints: `env.ACT`, `steps.nuclei_scan.outputs.sarif_exists`, `github.event_name`
  - Only run in verbose/debug mode or always for troubleshooting
  - Use to verify correct behavior in both `act` and GitHub Actions
- **Files**: `.github/workflows/dast.yml` (add new step before line 151)
- **Proposed**:
  ```yaml
  - name: Debug - Show environment and output variables
    run: |
      echo "ACT environment: ${{ env.ACT }}"
      echo "SARIF exists: ${{ steps.nuclei_scan.outputs.sarif_exists }}"
      echo "Event name: ${{ github.event_name }}"
      echo "Runner environment: ${{ runner.environment }}"
  ```

---

## Section 3: OWASP ZAP Configuration Analysis ✅ **COMPLETED**

### 3.1 ZAP Configuration File Usage

All exploratory and migration tasks related to the deprecated external ZAP config file have been fully removed per maintenance guideline. ZAP configuration is inline in the workflow and governed by active tasks in Sections 3.2 and 4+. No further action required for the deprecated file.

### 3.2 ZAP Rules Configuration (🟡 MEDIUM)

**Tasks**:

#### Task 3.2.1: Review and Update rules.tsv for Current Application (🟡 MEDIUM)
- **Description**: Ensure ZAP rules reflect current application security requirements
- **Action Items**:
  - Review IGNORE rules for health endpoints (lines 7-9) - still valid?
  - Review WARN rules for crypto endpoints (lines 12-13) - endpoints match current API?
  - Update endpoint patterns to match current OpenAPI spec paths
  - Check if Swagger UI path patterns are correct (line 10 is commented)
  - Add rules for any new critical endpoints
- **Files**: `.zap/rules.tsv`
- **Current Endpoints in Code**:
  - Public: `/browser/api/v1/*`, `/service/api/v1/*`, `/ui/swagger/*`
  - Private: `/readyz`, `/healthz`, `/shutdown`

#### Task 3.2.2: Add ZAP Rules for Security Header Findings (🟢 LOW)
- **Description**: Add specific rules to validate security headers in ZAP scans
- **Action Items**:
  - Add WARN/FAIL rules for missing security headers
  - Use rule IDs from `.zap/rules.tsv` comments (10016, 10017, 10020, etc.)
  - Configure expected headers as PASS criteria
  - Test rules with local ZAP scan
- **Files**: `.zap/rules.tsv`

### 3.3 URL and Protocol Mismatches (🟡 MEDIUM)

**Finding**: Configuration files have URL/protocol inconsistencies

**Tasks**:

#### Task 3.3.1: (Removed) Protocol mismatch audit
Removed with deprecated config file; HTTPS target now authoritative in workflow.

#### Task 3.3.2: Verify Target URL Consistency Across All Files (🟢 LOW)
- **Description**: Ensure all DAST-related files use consistent target URLs
- **Action Items**:
  - Check `dast.yml`: `TARGET_URL: ${{ github.event.inputs.target_url || 'https://localhost:8080' }}` ✅
  - ~~Check `dast-config.yml`~~ ✅ **REMOVED** - No longer applicable
  - Check SECURITY_TESTING.md documentation: verify examples use correct URLs
  - Update all references to use `https://localhost:8080` consistently
- **Files**: `.github/workflows/dast.yml`, `docs/SECURITY_TESTING.md`

---

## Section 4: Commented Out OWASP ZAP Steps Analysis

### 4.1 Re-enable OWASP ZAP Scans (🟠 HIGH)

**Current State**: OWASP ZAP Full Scan (lines 125-132) and API Scan (lines 134-141) are commented out

**Tasks**:

#### Task 4.1.1: Test OWASP ZAP Full Scan Locally (🟠 HIGH)
- **Description**: Uncomment and test ZAP Full Scan step locally with `act`
- **Action Items**:
  - Uncomment lines 125-132 in `dast.yml`
  - Run locally: `act --bind -j dast-security-scan`
  - Monitor scan duration and findings
  - Review generated artifacts (zap-report HTML/JSON)
  - Identify any issues with local execution
  - Document findings and adjust configuration if needed
- **Files**: `.github/workflows/dast.yml` (lines 125-132)
- **Expected Duration**: ~10 minutes (per `max_duration: 600` in dast-config.yml)

#### Task 4.1.2: Test OWASP ZAP API Scan Locally (🟠 HIGH)
- **Description**: Uncomment and test ZAP API Scan step locally with `act`
- **Action Items**:
  - Uncomment lines 134-141 in `dast.yml`
  - Verify OpenAPI spec is accessible: `curl -k https://localhost:8080/ui/swagger/doc.json`
  - Run locally: `act --bind -j dast-security-scan`
  - Monitor scan duration and findings
  - Review generated artifacts (zap-api-report HTML/JSON)
  - Identify any issues with OpenAPI-driven scanning
  - Document findings and adjust configuration if needed
- **Files**: `.github/workflows/dast.yml` (lines 134-141)
- **Expected Duration**: ~5 minutes (per `max_duration: 300` in dast-config.yml)

#### Task 4.1.3: Review ZAP Action Parameters and Update (🟡 MEDIUM)
- **Description**: Ensure ZAP action parameters are optimal for cryptoutil
- **Action Items**:
  - Review `cmd_options` for both Full Scan and API Scan
  - Current Full Scan: `-a -j -m 10 -T 60 -z "-config rules.cookie.ignorelist=JSESSIONID,csrftoken,_csrf"`
  - Current API Scan: `-a -j -T 60`
  - Verify `-a` (include Alpha rules) is appropriate for crypto app
  - Verify `-j` (use AJAX spider) is appropriate for API
  - Check if `-m 10` (max scan minutes) should be adjusted
  - Consider adding `-d` (debug mode) for troubleshooting
  - Ensure cookie ignore list includes cryptoutil's CSRF token name from config
- **Files**: `.github/workflows/dast.yml` (lines 128, 137)
- **References**: [ZAP Full Scan Options](https://www.zaproxy.org/docs/docker/full-scan/)

#### Task 4.1.4: Coordinate All Three Scans (ZAP Full, ZAP API, Nuclei) (🟡 MEDIUM)
- **Description**: Optimize execution order and duration when all scans are enabled
- **Action Items**:
  - Current: Only Nuclei is running (~4-5 minutes)
  - With ZAP: Full Scan (10 min) + API Scan (5 min) + Nuclei (5 min) = ~20 minutes total
  - Consider running ZAP scans in parallel if possible (probably not supported)
  - Adjust timeouts if scans are too slow
  - Consider weekly full scans vs. PR quick scans (Nuclei only)
  - Document scan duration expectations in SECURITY_TESTING.md
- **Files**: `.github/workflows/dast.yml`, `docs/SECURITY_TESTING.md`

#### Task 4.1.5: Handle ZAP Artifact Upload (🟡 MEDIUM)
- **Description**: Add artifact upload for ZAP reports (similar to Nuclei)
- **Action Items**:
  - Add step to upload ZAP Full Scan reports (HTML/JSON)
  - Add step to upload ZAP API Scan reports (HTML/JSON)
  - Ensure artifact upload is conditional: `if: ${{ !env.ACT }}`
  - Consider consolidating all DAST reports into single artifact
  - Test artifact upload in GitHub Actions (not `act`)
- **Files**: `.github/workflows/dast.yml` (add after line 156)
- **Proposed**:
  ```yaml
  - name: Upload ZAP Reports
    if: ${{ !env.ACT }}
    uses: actions/upload-artifact@v4
    with:
      name: zap-reports
      path: |
        zap-full-report.html
        zap-full-report.json
        zap-api-report.html
        zap-api-report.json
  ```



## Section 5: Additional Configuration and Documentation Tasks

### 5.0 Newly Identified Empirical Gaps (from latest `act` `nuclei.log` run 2025-10-01)

These tasks were not previously tracked. They arise from concrete evidence in `nuclei.log` (19 matches; 11 header findings; 272 template errors; summary mismatch) and workflow drift.

#### Task 5.0.1: Dynamic Security Summary (🟠 HIGH)
- **Problem**: Summary marks ZAP scans as ✅ though ZAP steps are commented out.
- **Action**:
  - Detect executed scanners via presence of artifacts (`nuclei.log`, ZAP reports) or step outputs.
  - Only print lines for scanners that actually ran; mark skipped ones as ⏭ Skipped (reason).
  - Add guard to prevent stale assertions.
- **Files**: `.github/workflows/dast.yml` (Generate Security Summary step)
- **Success**: Summary accurately reflects executed scans (diff verified in PR).

#### Task 5.0.2: Header False-Negative Investigation (🟠 HIGH)
- **Problem**: Nuclei flags all 11 security headers missing though middleware sets several (CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, etc.).
- **Action**:
  - Add step dumping `curl -vkI` for: `/ui/swagger/`, `${PublicBrowserAPIContextPath}`, `${PublicServiceAPIContextPath}` root.
  - Store as `dast-reports/response-headers.txt` (act) / artifact (GitHub).
  - Compare with nuclei raw request (optional: run nuclei with `-debug -trace-log nuclei-trace.log`).
  - Determine if mismatch due to protocol, hostname, compression, or redirect behavior.
  - Document outcome and adjust scanner or middleware accordingly.
- **Files**: `.github/workflows/dast.yml`, future `docs/SECURITY_TESTING.md` section.
- **Success**: Clear matrix of (Header Present by curl) vs (Reported by Nuclei) with remediation path.

#### Task 5.0.3: Port / Service Noise Suppression (🟡 MEDIUM)
- **Problem**: Non-app ports produce findings (`rpcbind-portmapper-detect`, `pgsql-detect`).
- **Action**: Append `-exclude-ports 22,111,5432` to nuclei flags; optionally restrict template types via `-tags`.
- **Files**: `.github/workflows/dast.yml` (nuclei action flags)
- **Success**: Subsequent scan excludes infrastructure findings; match count reduces accordingly.

#### Task 5.0.4: Template Error Audit (🟡 MEDIUM)
- **Problem**: 272 template errors inflate noise.
- **Action**: Re-run locally with `nuclei -target ... -validate -vv -stats-json -json-export nuclei-errors.json` to categorize (timeout vs syntax vs network).
- **Deliverable**: Error classification table + decisions (ignore / tune timeout / drop templates).
- **Files**: Local process notes + potential flags update.

#### Task 5.0.5: Deterministic Template Version (🟡 MEDIUM)
- **Problem**: Auto-install each run → non-deterministic results.
- **Action**: Cache templates using `actions/cache` keyed on `nuclei-templates-${{ hashFiles('**/nuclei.lock') }}-${{ env.NUCLEI_VERSION }}`; generate `nuclei.lock` file containing template release tag hash.
- **Files**: `.github/workflows/dast.yml`, new `nuclei.lock` (optional manual maintenance until automation added).

#### Task 5.0.6: Header Artifact Export (🟡 MEDIUM)
- **Problem**: Need reproducible header baselines.
- **Action**: Implement a dedicated step: `curl -skI ... >> headers.txt`; include timestamps & endpoint labels.
- **Files**: `.github/workflows/dast.yml` (after app start, before scans)
- **Success**: `headers.txt` attached / saved locally under act.

#### Task 5.0.7: HSTS Conditional Verification (🟢 LOW)
- **Problem**: Nuclei still reports missing `Strict-Transport-Security` though HTTPS endpoint live.
- **Action**: Confirm middleware condition (only when protocol=https); ensure nuclei uses `https://` (not redirect). Possibly add explicit `-H "Host: localhost"` & check if local certificate influences detection.
- **Files**: `application_listener.go`, workflow.

#### Task 5.0.8: Dummy Token Documentation (🟢 LOW)
- **Problem**: Usage rationale not documented; risk of accidental commit misuse.
- **Action**: Add subsection to SECURITY_TESTING.md: "Local act Token Strategy" explaining dummy token scope and safeguards; add echo warning if `GITHUB_TOKEN` already set.
- **Files**: `docs/SECURITY_TESTING.md`, `.github/workflows/dast.yml` (optional echo guard).

#### Task 5.0.9: Go Module Cache Enhancement (🟢 LOW)
- **Problem**: Setup-go cache captures modules, but document explicitly & add fallback metrics.
- **Action**: Add summary note capturing cache hit/miss; adjust docs to reflect cost impact.
- **Files**: `.github/workflows/dast.yml`, `docs/SECURITY_TESTING.md`.

#### Task 5.0.10: Nuclei Template Cache (🟢 LOW)
- **Problem**: Re-install templates each run.
- **Action**: Pre-step using `actions/cache` on `~/.nuclei-templates` or the path used in action; restore before action executes.
- **Files**: `.github/workflows/dast.yml`.

#### Task 5.0.11: Summary Severity Normalization (🟢 LOW)
- **Problem**: All header findings show `[info]`; need internal severity mapping.
- **Action**: Post-process SARIF JSON to categorize into internal severities (Info/Warn/Fail) using config table.
- **Files**: Future script under `scripts/` (PowerShell + bash parity) plus workflow optional step.

#### Task 5.0.12: Trace Logging Option (🟢 LOW)
- **Problem**: Troubleshooting requires ad-hoc manual reruns.
- **Action**: Add input `debug: true` to workflow_dispatch to enable nuclei `-debug -trace-log nuclei-trace.log` and verbose curl.
- **Files**: `.github/workflows/dast.yml`.

#### Task 5.0.13: Document Endpoint Canonical Set (🟢 LOW)
- **Problem**: References to non-existent `/browser/api/v1/health` in planning conversations; canonical endpoints are `/livez`, `/readyz` (private only).
- **Action**: Add table in SECURITY_TESTING.md listing exported vs internal health endpoints; clarify scanner target scope.
- **Files**: `docs/SECURITY_TESTING.md`.

#### Task 5.0.14: Combine Reports Artifact (🟢 LOW)
- **Problem**: Multiple artifact uploads (future ZAP + nuclei) increase clutter.
- **Action**: Single artifact `dast-reports` bundling all outputs (`nuclei.log`, `nuclei.sarif`, headers.txt, zap reports) when not in act.
- **Files**: `.github/workflows/dast.yml`.

#### Task 5.0.15: Scan Profile Matrix (🟢 LOW)
- **Problem**: Need explicit mapping of trigger → scanner set.
- **Action**: Add matrix doc (PR=Quick, Push=Full minus deep templates, Schedule=Full+Deep, Manual=Custom).
- **Files**: `docs/SECURITY_TESTING.md`.

---

### 5.1 Documentation Updates (🟢 LOW)

#### Task 5.1.1: Update SECURITY_TESTING.md with Latest Workflow (🟢 LOW)
- **Description**: Current documentation may be outdated after workflow changes
- **Action Items**:
  - Update documentation to reflect current state of `dast.yml`
  - Document which scans are currently active (Nuclei only vs. all three)
  - Add section on local testing with `act` and known limitations
  - Document `act` artifact upload workaround
  - Add troubleshooting section for common `act` issues
  - Update scan duration estimates
- **Files**: `docs/SECURITY_TESTING.md`

#### Task 5.1.2: Document ZAP Configuration Architecture Decision (🟢 LOW)
- **Description**: Document the decision about dast-config.yml usage (Task 3.1.1 outcome)
- **Action Items**:
  - Add "Configuration Architecture" section to SECURITY_TESTING.md
  - Document why inline config in workflow was chosen (or why config file was chosen)
  - Explain relationship between `dast.yml`, `rules.tsv`, and `dast-config.yml`
  - Provide guidance for future configuration changes
- **Files**: `docs/SECURITY_TESTING.md`

#### Task 5.1.3: Create DAST Troubleshooting Guide (🟢 LOW)
- **Description**: Add comprehensive troubleshooting section for DAST issues
- **Action Items**:
  - Document `act` specific issues and workarounds
  - Document common ZAP scan failures and fixes
  - Document Nuclei template update procedures
  - Add section on interpreting false positives
  - Include curl commands for manual endpoint testing
- **Files**: `docs/SECURITY_TESTING.md` or new `docs/DAST_TROUBLESHOOTING.md`

### 5.2 Workflow Optimizations (🟡 MEDIUM)

#### Task 5.2.1: Add Scan Duration Monitoring (🟢 LOW)
- **Description**: Track and report scan durations for performance monitoring
- **Action Items**:
  - Add timing output at start/end of each scan step
  - Include durations in "Generate Security Summary" step
  - Track duration trends over time (manual for now)
  - Alert if scans exceed expected duration thresholds
- **Files**: `.github/workflows/dast.yml`

#### Task 5.2.2: Implement Differential Scanning Strategy (🟡 MEDIUM)
- **Description**: Run different scan depths based on trigger event
- **Action Items**:
  - **Quick Scan** (PRs): Nuclei only (~5 min)
  - **Full Scan** (main push, scheduled): All three scanners (~20 min)
  - **Deep Scan** (manual dispatch): Extended timeouts, additional templates
  - Implement workflow logic to select scan depth based on trigger
  - Document scan depth strategy in SECURITY_TESTING.md
- **Files**: `.github/workflows/dast.yml`
- **Benefits**: Faster PR feedback, comprehensive scheduled scanning

#### Task 5.2.3: Add Pre-Scan Health Verification (🟡 MEDIUM)
- **Description**: Verify application is fully ready before starting scans
- **Action Items**:
  - Current: Single curl check for Swagger UI
  - Enhancement: Check multiple critical endpoints
  - Verify database connectivity (via health endpoint)
  - Verify application unsealed and ready
  - Fail fast if application not ready (save CI minutes)
  - Add retry logic with exponential backoff
- **Files**: `.github/workflows/dast.yml` (enhance step at line 119-123)

### 5.3 CI/CD Cost Optimization (🟢 LOW)

#### Task 5.3.1: Analyze DAST Workflow CI/CD Minutes Usage (🟢 LOW)
- **Description**: Monitor and optimize GitHub Actions minutes consumption
- **Action Items**:
  - Review historical workflow duration data
  - Identify longest-running steps
  - Calculate monthly CI minutes cost for DAST workflow
  - Compare against GitHub Actions free tier limits
  - Recommend optimizations if approaching limits
- **Files**: GitHub Actions usage reports
- **Reference**: `.github/copilot-instructions.md` (CI/CD Cost Efficiency section)

#### Task 5.3.2: Implement Scan Caching Strategy (🟢 LOW)
- **Description**: Cache Nuclei templates and ZAP data to reduce download time
- **Action Items**:
  - Cache Nuclei templates directory between runs
  - Cache ZAP docker image layers (may not be possible)
  - Document cache invalidation strategy
  - Measure time savings from caching
- **Files**: `.github/workflows/dast.yml`
- **Expected Savings**: 1-2 minutes per run

#### Task 5.3.3: Add Job Filters for Docs-Only Changes (🟢 LOW)
- **Description**: Skip DAST workflow when only documentation files change
- **Action Items**:
  - Add `paths-ignore` to workflow triggers
  - Skip DAST for changes only in: `docs/**`, `*.md`, `.github/copilot-instructions.md`
  - Test that workflow skips correctly
  - Document path filter strategy
- **Files**: `.github/workflows/dast.yml` (lines 3-12, workflow triggers)
- **Expected Savings**: Significant if docs changes are frequent

### 5.4 Security Scanning Coverage (🟡 MEDIUM)

#### Task 5.4.1: Map OWASP Top 10 to DAST Findings (🟢 LOW)
- **Description**: Create mapping of scan findings to OWASP Top 10 categories
- **Action Items**:
  - Review SECURITY_TESTING.md OWASP Top 10 coverage section (exists)
  - Map actual Nuclei findings to OWASP categories
  - Map ZAP findings to OWASP categories (when re-enabled)
  - Document coverage gaps
  - Identify additional tests needed for complete coverage
- **Files**: `docs/SECURITY_TESTING.md`

#### Task 5.4.2: Add Cryptographic-Specific Security Tests (🟡 MEDIUM)
- **Description**: Create custom tests for cryptoutil's unique crypto operations
- **Action Items**:
  - (Optional) Brainstorm cryptographic abuse payloads (do not re-introduce removed config file)
  - Create test cases for key generation endpoints
  - Create test cases for encryption/decryption endpoints
  - Create test cases for signing/verification endpoints
  - Test for timing attack vulnerabilities (constant-time operations)
  - Test for padding oracle vulnerabilities
  - Document custom crypto test cases
- **Files**: New file `tests/security/crypto_dast_tests.md` or similar
- **Tools**: May require custom Nuclei templates or manual testing

#### Task 5.4.3: Test Rate Limiting and DoS Protection (🟡 MEDIUM)
- **Description**: Verify rate limiting middleware is effective
- **Action Items**:
  - Check current rate limiting config in application
  - Create test script to rapidly send requests
  - Verify rate limiting triggers and blocks excessive requests
  - Test per-IP rate limiting
  - Test global rate limiting
  - Document rate limiting thresholds and behavior
- **Files**: `internal/server/application/application_listener.go` (rate limiter middleware)
- **Tools**: Could use Apache Bench, custom script, or Nuclei rate limit tests

---

## Section 6: Testing and Validation Tasks

### 6.1 Local Testing Workflow (🟠 HIGH)

#### Task 6.1.1: Create Comprehensive Local DAST Testing Guide (🟡 MEDIUM)
- **Description**: Step-by-step guide for running DAST locally with `act`
- **Action Items**:
  - Document `act` installation and setup
  - Document required Docker images (ZAP, nuclei-action)
  - Create command reference for common `act` operations
  - Document expected vs. actual behavior differences
  - Create troubleshooting checklist for common issues
  - Add to SECURITY_TESTING.md or separate DAST_LOCAL_TESTING.md
- **Files**: `docs/SECURITY_TESTING.md`

#### Task 6.1.2: Create DAST Pre-Commit Hook (Optional) (🟢 LOW)
- **Description**: Optional pre-commit hook to run quick DAST scan before commit
- **Action Items**:
  - Create script to run Nuclei scan only (fast, ~5 min)
  - Integrate with existing pre-commit hooks
  - Make it optional (not required for all commits)
  - Document how to enable/disable
  - Consider adding to `.github/instructions/` as best practice
- **Files**: New file `.githooks/pre-commit-dast.sh` (optional)
- **Trade-off**: Longer commit time vs. earlier security feedback

### 6.2 GitHub Actions Testing (🟠 HIGH)

#### Task 6.2.1: Test Complete Workflow in GitHub Actions (🟠 HIGH)
- **Description**: Full end-to-end test with all scanners enabled
- **Action Items**:
  - Re-enable ZAP Full Scan and API Scan steps
  - Commit and push to test branch
  - Monitor workflow execution in GitHub Actions
  - Verify all artifacts are uploaded correctly
  - Verify SARIF upload to Security tab works
  - Review all scan findings
  - Fix any issues discovered
  - Merge to main after validation
- **Files**: `.github/workflows/dast.yml`
- **Blocker**: Should be done after fixing `act` compatibility issues

#### Task 6.2.2: Test Manual Workflow Dispatch (🟡 MEDIUM)
- **Description**: Verify manual trigger with custom target URL works
- **Action Items**:
  - Navigate to GitHub Actions → DAST Security Testing workflow
  - Click "Run workflow"
  - Test with default URL: `https://localhost:8080`
  - Test with custom URL (if staging/prod environment available)
  - Verify workflow runs successfully
  - Document manual trigger procedure in SECURITY_TESTING.md
- **Files**: `.github/workflows/dast.yml` (workflow_dispatch input)

#### Task 6.2.3: Test Scheduled Workflow Execution (🟢 LOW)
- **Description**: Verify weekly scheduled scan runs correctly
- **Action Items**:
  - Check workflow schedule: `cron: '0 2 * * 0'` (Sundays at 2 AM UTC)
  - Wait for next scheduled run or temporarily adjust cron for testing
  - Monitor scheduled run
  - Verify notifications are sent for findings
  - Check artifact retention (30 days default)
  - Document scheduled scan in SECURITY_TESTING.md
- **Files**: `.github/workflows/dast.yml` (line 9)

### 6.3 Security Findings Validation (🟡 MEDIUM)

#### Task 6.3.1: Validate Each Nuclei Finding (🟡 MEDIUM)
- **Description**: Manually verify each finding is legitimate or false positive
- **Action Items**:
  - Go through each finding in `nuclei.log`/`nuclei.sarif`
  - For each finding, determine: Real vulnerability? Expected behavior? False positive?
  - Document findings in spreadsheet or findings log
  - Create remediation tickets for real vulnerabilities
  - Update ZAP rules to suppress false positives
  - Document validation process in SECURITY_TESTING.md
- **Files**: `nuclei.log`, `.zap/rules.tsv`

#### Task 6.3.2: Validate ZAP Findings (When Re-enabled) (🟡 MEDIUM)
- **Description**: Review and validate ZAP scan findings
- **Action Items**:
  - Review ZAP Full Scan report (HTML)
  - Review ZAP API Scan report (HTML)
  - Categorize findings: Critical, High, Medium, Low, Info, False Positive
  - Create remediation tickets for legitimate findings
  - Update `.zap/rules.tsv` to suppress known false positives
  - Document validation process
- **Files**: ZAP reports (when generated), `.zap/rules.tsv`

#### Task 6.3.3: Compare DAST Findings with SAST Findings (🟢 LOW)
- **Description**: Cross-reference DAST findings with existing SAST (CodeQL) results
- **Action Items**:
  - Review GitHub Security tab for CodeQL findings
  - Identify overlapping findings (found by both DAST and SAST)
  - Identify unique DAST findings (runtime issues)
  - Identify unique SAST findings (code-level issues)
  - Document comparison in security review report
  - Use to improve both DAST and SAST coverage
- **Files**: GitHub Security Dashboard

---

## Section 7: Long-Term Improvements and Enhancements

### 7.1 Advanced DAST Features (🟢 LOW - Future)

#### Task 7.1.1: Implement Authenticated Scanning (Future) (🟢 LOW)
- **Description**: When authentication is implemented, update DAST to test authenticated endpoints
- **Action Items**:
  - Currently: `authentication: type: "none"` in dast-config.yml
  - When auth implemented: Configure ZAP/Nuclei with test credentials
  - Test authenticated vs. unauthenticated endpoints
  - Verify authorization controls (can user A access user B's data?)
  - Document authenticated scanning configuration
-- **Files**: `.github/workflows/dast.yml`
- **Blocked By**: Authentication implementation in cryptoutil

#### Task 7.1.2: Implement Dynamic Target URL from Deployment (Future) (🟢 LOW)
- **Description**: Scan actual deployed environments (staging, prod) instead of localhost
- **Action Items**:
  - Add deployment environment URL as workflow secret or variable
  - Update workflow to accept deployment URL as input
  - Configure network access from GitHub Actions to deployed environment
  - Implement separate workflows for staging vs. production scans
  - Add production scan approval requirement (manual trigger only)
- **Files**: `.github/workflows/dast.yml`, GitHub Secrets configuration
- **Security Note**: Ensure scans don't disrupt production or expose sensitive data

#### Task 7.1.3: Integrate DAST Results with Security Dashboard (Future) (🟢 LOW)
- **Description**: Centralized security dashboard tracking DAST trends over time
- **Action Items**:
  - Research GitHub Security Dashboard capabilities for custom data
  - Explore third-party security dashboards (Snyk, etc.)
  - Implement trend tracking for findings over time
  - Create visualizations for security posture improvement
  - Set up alerts for regression (new vulnerabilities introduced)
- **Tools**: GitHub Advanced Security, DefectDojo, or custom solution

### 7.2 Testing Infrastructure (🟢 LOW - Future)

#### Task 7.2.1: Create Dedicated DAST Testing Environment (Future) (🟢 LOW)
- **Description**: Separate environment specifically for security testing
- **Action Items**:
  - Deploy cryptoutil in isolated test environment
  - Configure with test data and test secrets
  - Make accessible to DAST scanners (network configuration)
  - Implement automatic deployment on commit
  - Scan deployed environment instead of local startup
- **Benefits**: More realistic testing, no workflow job startup time, persistent target
- **Trade-off**: Additional infrastructure cost and complexity

#### Task 7.2.2: Implement Custom Nuclei Templates for Cryptoutil (Future) (🟢 LOW)
- **Description**: Create cryptoutil-specific vulnerability templates
- **Action Items**:
  - Learn Nuclei template YAML syntax
  - Create templates for crypto-specific vulnerabilities
  - Create templates for elastic key operations
  - Create templates for unseal/seal operations
  - Test templates against cryptoutil
  - Contribute templates to Nuclei community (optional)
- **Files**: New directory `.nuclei-templates/` or similar
- **References**: [Nuclei Template Guide](https://docs.projectdiscovery.io/templates/introduction)

---

## Section 8: Priority and Execution Plan

### Immediate Priority (Complete First) - Sprint 1

**Goal**: Fix blocking issues, enable full DAST workflow

1. ✅ **Task 2.1.1**: Fix `act` artifact upload conditional (🟠 HIGH) ✅ COMPLETED
2. ✅ **Task 2.1.3**: Add continue-on-error for artifact upload in `act` (🟡 MEDIUM)
3. ✅ **Task 1.1.1**: Verify current security header implementation (🟡 MEDIUM)
4. ✅ **Task 1.1.2**: Review middleware execution order (🟡 MEDIUM)
5. ✅ **Task 4.1.1**: Test OWASP ZAP Full Scan locally (🟠 HIGH)
6. ✅ **Task 4.1.2**: Test OWASP ZAP API Scan locally (🟠 HIGH)
7. ✅ **Task 3.1.1**: Decide on ZAP configuration strategy - Option A implemented (🟡 MEDIUM)

**Expected Outcome**: DAST workflow runs successfully with all three scanners in both `act` and GitHub Actions

### High Priority (Complete Second) - Sprint 2

**Goal**: Remediate security findings, optimize workflow

8. ✅ **Task 1.3.1**: Scope Nuclei scans to application only (🟡 MEDIUM)
9. ✅ **Task 3.1.3**: Extract useful config from dast-config.yml (🟡 MEDIUM)
10. ✅ **Task 3.3.1**: Fix protocol mismatch in dast-config.yml - N/A (🟡 MEDIUM)
11. ✅ **Task 4.1.3**: Review and update ZAP action parameters (🟡 MEDIUM)
12. ✅ **Task 6.2.1**: Test complete workflow in GitHub Actions (🟠 HIGH)

**Expected Outcome**: Security posture improved, workflow configuration clean and consistent

### Medium Priority (Complete Third) - Sprint 3

**Goal**: Documentation, validation, optimization

13. ✅ **Task 5.1.1**: Update SECURITY_TESTING.md with latest workflow (🟢 LOW)
14. ✅ **Task 6.3.1**: Validate each Nuclei finding (🟡 MEDIUM)
15. ✅ **Task 6.3.2**: Validate ZAP findings (🟡 MEDIUM)
16. ✅ **Task 1.3.2**: Review CSRF cookie HttpOnly configuration (🟡 MEDIUM)
17. ✅ **Task 3.2.1**: Review and update rules.tsv (🟡 MEDIUM)
18. ✅ **Task 5.2.2**: Implement differential scanning strategy (🟡 MEDIUM)

**Expected Outcome**: All findings validated, false positives suppressed, optimal scanning strategy

### Low Priority (Complete Fourth) - Sprint 4

**Goal**: Polish, monitoring, long-term improvements

19. ✅ **Task 5.1.3**: Create DAST troubleshooting guide (🟢 LOW)
20. ✅ **Task 5.3.3**: Add job filters for docs-only changes (🟢 LOW)
21. ✅ **Task 5.4.1**: Map OWASP Top 10 to DAST findings (🟢 LOW)
22. ✅ **Task 6.1.1**: Create comprehensive local DAST testing guide (🟡 MEDIUM)
23. ✅ All remaining 🟢 LOW priority tasks

**Expected Outcome**: Comprehensive documentation, efficient CI/CD usage, excellent developer experience

### Future Enhancements (Backlog)

- All tasks in Section 7 (Long-Term Improvements)
- Advanced features requiring authentication implementation
- Deployment environment scanning
- Custom Nuclei templates
- Security dashboard integration

---

## Appendix A: Quick Reference

### File Inventory

| File | Purpose | Status |
|------|---------|--------|
| `.github/workflows/dast.yml` | Main DAST workflow | ⚠️ ZAP steps commented, artifact upload broken in `act` |
| `.zap/rules.tsv` | ZAP rule configuration | ✅ Used by workflow |
| `nuclei.log` | Nuclei scan results | ✅ Generated by workflow |
| `nuclei.sarif` | Nuclei SARIF output | ✅ Generated, uploaded to Security tab |
| `dast-github-action-nuclei.log` | GitHub Actions run log | 📊 Analyzed for findings |
| `dast-act-localhost.log` | Local `act` run log | 📊 Analyzed for errors |
| `docs/SECURITY_TESTING.md` | DAST documentation | ⚠️ May be outdated |

### Key Environment Variables

| Variable | Value | Source |
|----------|-------|--------|
| `ACT` | `true` (when running in `act`) | Set by `act` |
| `TARGET_URL` | `https://localhost:8080` | `.github/workflows/dast.yml` |
| `APP_BIND_PUBLIC_PORT` | `8080` | `.github/workflows/dast.yml` |
| `APP_BIND_PRIVATE_PORT` | `9090` | `.github/workflows/dast.yml` |

### Command Reference

```powershell
# Run DAST locally with act
act --bind -j dast-security-scan

# Test application connectivity
curl -f -k https://localhost:8080/ui/swagger/doc.json

# Check response headers
curl -I https://localhost:8080/ui/swagger/

# Run Nuclei scan manually
nuclei -target https://localhost:8080 -c 24 -rl 200 -timeout 5 -stats

# Run ZAP Full Scan manually (Docker)
docker run --rm -v ${PWD}:/zap/wrk/:rw zaproxy/zap-stable zap-full-scan.py -t https://localhost:8080 -r report.html

# Run ZAP API Scan manually (Docker)
docker run --rm -v ${PWD}:/zap/wrk/:rw zaproxy/zap-stable zap-api-scan.py -t https://localhost:8080/ui/swagger/doc.json -f openapi -r report.html
```

---

## Appendix B: Links and References

### Documentation
- [OWASP ZAP Documentation](https://www.zaproxy.org/docs/)
- [Nuclei Documentation](https://docs.projectdiscovery.io/tools/nuclei)
- [GitHub Actions act Documentation](https://github.com/nektos/act)
- [OWASP Secure Headers Project](https://owasp.org/www-project-secure-headers/)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)

### GitHub Actions
- [zaproxy/action-full-scan](https://github.com/zaproxy/action-full-scan)
- [zaproxy/action-api-scan](https://github.com/zaproxy/action-api-scan)
- [projectdiscovery/nuclei-action](https://github.com/projectdiscovery/nuclei-action)
- [actions/upload-artifact](https://github.com/actions/upload-artifact)
- [github/codeql-action/upload-sarif](https://github.com/github/codeql-action)

### Internal Files
- [SECURITY_TESTING.md](./SECURITY_TESTING.md)
- [Copilot Instructions](../.github/copilot-instructions.md)
- [DAST Workflow](../.github/workflows/dast.yml)
- [ZAP Rules](../.zap/rules.tsv)


---

## Document History

| Date | Author | Changes |
|------|--------|---------|
| 2025-09-30 | GitHub Copilot | Initial discovery phase analysis |

---

**End of Document**
