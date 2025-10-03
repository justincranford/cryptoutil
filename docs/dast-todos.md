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

### DAST Workflow Performance Optimization (🟠 HIGH - TOP PRIORITY)

**Context**: Current DAST workflow runtime is ~10-15 minutes. Multiple optimization opportunities identified to reduce CI/CD costs and improve developer experience through faster feedback loops.

#### Task O1: Implement Trigger-Based Job Filtering (🟠 HIGH)
#### Task O1: Implement Differential Scanning Strategy (🟠 HIGH)  
- **Description**: Use different scan depths based on trigger type for optimal speed vs thoroughness balance
- **Context**: Current workflow uses same 600s timeout for all triggers. PRs need fast feedback, scheduled scans need thoroughness
- **Action Items**:
  - Add `scan_profile` input to workflow_dispatch with options: quick/full/deep
  - Configure Quick Profile (PRs): 60s timeout, limited templates (~2-3 minutes)
  - Configure Full Profile (main push): Current 600s timeout (~10 minutes)  
  - Configure Deep Profile (scheduled/manual): 1200s timeout, all templates (~20 minutes)
  - Add conditional logic to set Nuclei flags based on trigger type
- **Files**: `.github/workflows/dast.yml` (inputs, Nuclei step flags)
- **Expected Savings**: 70% faster PR feedback (10min → 3min)
- **Implementation**: Conditional Nuclei timeout and template selection

#### Task O2: Fix and Enhance Nuclei Template Caching (🟡 MEDIUM)
- **Description**: Improve template caching effectiveness to reduce download time on each run
- **Context**: Current cache uses non-existent `nuclei.lock` file, making caching ineffective
- **Action Items**:
  - Fix cache key to use `go.sum` hash instead of missing `nuclei.lock`
  - Update cache path configuration for better hit rates
  - Add cache statistics to workflow summary for monitoring
  - Test cache effectiveness across multiple runs
- **Files**: `.github/workflows/dast.yml` (Cache Nuclei Templates step)
- **Expected Savings**: ~1 minute per run when templates cached
- **Implementation**: Update cache key pattern and restore-keys

#### Task O3: Implement Parallel Step Execution (🟡 MEDIUM)
- **Description**: Parallelize setup steps that don't depend on each other
- **Context**: Currently all setup steps run sequentially, but some can run in parallel
- **Action Items**:
  - Run directory creation in background (`mkdir -p configs/test & mkdir -p ./dast-reports &`)
  - Parallelize config file creation with other setup tasks
  - Optimize application startup sequence
  - Combine redundant curl connectivity tests
- **Files**: `.github/workflows/dast.yml` (Start application step)
- **Expected Savings**: ~30 seconds per run
- **Implementation**: Background processes and command chaining

#### Task O4: Remove Redundant and Optimize Steps (🟢 LOW)
- **Description**: Clean up workflow by removing duplicate operations and optimizing step efficiency  
- **Context**: Workflow has duplicate curl tests and can be streamlined
- **Action Items**:
  - Remove duplicate "Test application curl connectivity" step
  - Combine config file creation into single heredoc operation
  - Optimize cleanup logic to be more efficient
  - Reduce verbose logging where not needed for debugging
- **Files**: `.github/workflows/dast.yml` (various steps)
- **Expected Savings**: ~15 seconds per run, cleaner workflow
- **Implementation**: Step consolidation and removal

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

### TOP PRIORITY - Performance Optimization (Sprint 0)
1. **Task O1**: Differential Scanning Strategy (major runtime improvement)
2. **Task O2**: Fix Nuclei Template Caching (consistent time savings)
3. **Task O3**: Parallel Step Execution (moderate improvement)
4. **Task O4**: Remove Redundant Steps (cleanup and polish)

### Immediate (Sprint 1)
6. **Task 1**: Security header analysis (baseline ready)
7. **Task 2**: ZAP Full Scan re-enablement
8. **Task 3**: ZAP API Scan re-enablement

### Next (Sprint 2)  
9. **Task 4**: ZAP rules configuration review
10. **Task 6**: Documentation updates

### Future (Sprint 3)
11. **Task 5**: CI/CD optimization (legacy - covered by O1-O5)

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
