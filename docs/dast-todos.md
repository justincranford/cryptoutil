# DAST TODO List - Active Tasks Only

**Document Status**: Active Remediation Phase  
**Created**: 2025-09-30  
**Updated**: 2025-10-02  
**Purpose**: Actionable task list for remaining DAST workflow improvements

> Maintenance Guideline: If a file/config/feature is removed or a decision makes a task permanently obsolete, DELETE its tasks and references here immediately. Keep only (1) active remediation work, (2) still-relevant observations, (3) forward-looking backlog items. Historical context belongs in commit messages or durable docs, not this actionable list.

---

## Executive Summary

**CURRENT STATUS** (2025-10-02): ✅ **Core DAST Infrastructure Working**

- ✅ **Nuclei security scanning** - Working correctly, 0 vulnerabilities found
- ✅ **GitHub Actions `act` compatibility** - Fully functional
- 🟡 **OWASP ZAP integration** - Ready for re-enablement
- 🟡 **Security header validation** - Baseline captured, needs analysis

**Next Phase**: Enable ZAP scanners and validate security findings

---

## Active Tasks

### Security Header Investigation (🟡 MEDIUM)

#### Task 1: Compare Security Header Baseline with Expected Headers (🟡 MEDIUM)
- **Description**: Analyze captured response headers against security requirements
- **Action Items**:
  - Review `dast-reports/response-headers.txt` baseline
  - Compare with middleware configuration in `application_listener.go`
  - Document which headers are present/missing
  - Identify false negatives in future Nuclei scans
- **Files**: `dast-reports/response-headers.txt`, `internal/server/application/application_listener.go`
- **Success Criteria**: Clear matrix of expected vs actual headers

### OWASP ZAP Re-enablement (🟡 MEDIUM)

#### Task 2: Test OWASP ZAP Full Scan Locally (🟡 MEDIUM)
- **Description**: Uncomment and test ZAP Full Scan step locally with `act`
- **Action Items**:
  - Uncomment lines in `dast.yml` for ZAP Full Scan
  - Run locally: `act --bind -j dast-security-scan`
  - Monitor scan duration and findings
  - Review generated artifacts (zap-report HTML/JSON)
- **Files**: `.github/workflows/dast.yml`
- **Expected Duration**: ~10 minutes

#### Task 3: Test OWASP ZAP API Scan Locally (🟡 MEDIUM)
- **Description**: Uncomment and test ZAP API Scan step locally with `act`
- **Action Items**:
  - Uncomment lines in `dast.yml` for ZAP API Scan
  - Verify OpenAPI spec accessibility
  - Run locally and review findings
- **Files**: `.github/workflows/dast.yml`
- **Expected Duration**: ~5 minutes

### Workflow Optimization (🟢 LOW)

#### Task 4: Review ZAP Rules Configuration (🟢 LOW)
- **Description**: Ensure ZAP rules reflect current application security requirements
- **Action Items**:
  - Review `.zap/rules.tsv` for current endpoints
  - Update endpoint patterns to match current OpenAPI spec
  - Add rules for security header validation
- **Files**: `.zap/rules.tsv`

#### Task 5: Add Job Filters for Docs-Only Changes (🟢 LOW)
- **Description**: Skip DAST workflow when only documentation files change
- **Action Items**:
  - Add `paths-ignore` to workflow triggers
  - Skip DAST for changes only in: `docs/**`, `*.md`
  - Test that workflow skips correctly
- **Files**: `.github/workflows/dast.yml`
- **Expected Savings**: Significant CI minutes for docs changes

### Documentation Updates (🟢 LOW)

#### Task 6: Update SECURITY_TESTING.md (🟢 LOW)
- **Description**: Update documentation to reflect current working state
- **Action Items**:
  - Document successful Nuclei scanning configuration
  - Add troubleshooting section for common issues
  - Update scan duration estimates
  - Document local testing with `act`
- **Files**: `docs/SECURITY_TESTING.md`

---

## Priority Execution Order

### Immediate (Sprint 1)
1. **Task 1**: Security header analysis (baseline ready)
2. **Task 2**: ZAP Full Scan re-enablement
3. **Task 3**: ZAP API Scan re-enablement

### Next (Sprint 2)  
4. **Task 4**: ZAP rules configuration review
5. **Task 6**: Documentation updates

### Future (Sprint 3)
6. **Task 5**: CI/CD optimization

---

## Quick Reference

### Successful Configuration
- **Nuclei flags**: `-c 24 -rl 200 -timeout 600 -stats -ept tcp,javascript`
- **Act compatibility**: `github.actor == 'nektos/act'` detection working
- **Artifact collection**: Local artifacts saved to `./dast-reports/`

### Next Steps
1. Analyze security header baseline
2. Re-enable ZAP scanners  
3. Validate complete DAST workflow

---

**Last Updated**: 2025-10-02  
**Completed tasks removed per maintenance guideline**
