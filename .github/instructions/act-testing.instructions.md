---
description: "Instructions for testing GitHub Actions workflows locally with act"
applyTo: "**"
---
# Act Testing Instructions

## CRITICAL: DAST Workflow Testing

**NEVER use short timeouts when running DAST workflows with act**

### Recommended Approach: Use cmd/workflow

**ALWAYS use the provided Go utility for running act workflows**

```bash
# Quick DAST scan (3-5 minutes)
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=quick"

# Full DAST scan (10-15 minutes)
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=full"

# Deep DAST scan (20-25 minutes)
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=deep"
```

**Features:**
- Runs act workflow directly with comprehensive monitoring
- Streams output to both console and log files
- Automatic workflow completion detection
- Comprehensive result analysis (workflow + task status)
- Artifact verification and summary reports
- Single command execution - no prompts
- Supports all workflow types (quality, e2e, dast, sast, robust)

### Manual Usage (Advanced Only)

```powershell
# Run in BACKGROUND with proper output redirection (no timeout - let it complete)
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
act workflow_dispatch -W .github/workflows/dast.yml --input scan_profile=quick --artifact-server-path ./dast-reports 2>&1 | Out-File -FilePath .\dast-reports\act-dast.log -Encoding utf8

# Monitor progress manually
Get-Content .\dast-reports\act-dast.log -Tail 100
```

### Timing Expectations

- **Quick profile**: 3-5 minutes (Nuclei scan, ZAP scans, app build/startup)
- **Full profile**: 10-15 minutes (comprehensive scanning)
- **Deep profile**: 20-25 minutes (exhaustive scanning)

### Why Long Timeouts Are Required

1. Application build: ~4-5 seconds
2. Application startup with health checks: ~5 seconds
3. Security header capture: ~1 second
4. ZAP connectivity check: ~5-10 seconds
5. ZAP Full Scan: 2-10 minutes (depending on profile)
6. ZAP API Scan: 1-5 minutes
7. Nuclei scan: 1-3 minutes (quick) to 15-20 minutes (deep)
8. Artifact collection and cleanup: ~1 second

**Total minimum**: 3-5 minutes for quick, 10-15 minutes for full scans

### Automatic Result Analysis

The `cmd/workflow` utility automatically analyzes:

1. **Workflow Status**: Job succeeded/failed detection
2. **Task Status**: Individual scan completion verification
   - Nuclei Scan (log/SARIF generation)
   - ZAP Full Scan (report generation)
   - ZAP API Scan (report generation)
   - Header Capture (security headers file)
3. **Artifact Generation**: Verification of output files
4. **Summary Report**: Saved to workflow analysis markdown files

### Success Criteria

✓ **Workflow Success**: `Job succeeded` in log output
✓ **Task Success**: All scan artifacts generated in `dast-reports/`
✓ **No Errors**: No `Job failed` or critical error messages

### Common Mistakes to AVOID

❌ **NEVER do this**: Using `-t` timeout flag or checking output too early
❌ **NEVER do this**: `Start-Sleep -Seconds 60` (way too short)
❌ **NEVER do this**: Checking terminal output while workflow is running
❌ **NEVER do this**: Running act commands directly without monitoring/analysis
❌ **CRITICAL - NEVER do this**: `Get-Content -Wait` on log file while scan is running - THIS KILLS THE PROCESS
❌ **CRITICAL - NEVER do this**: Any interactive monitoring commands that lock files or interfere with running processes
❌ **CRITICAL - NEVER do this**: Opening/tailing log files in another terminal while scan is running

✅ **ALWAYS do this**: Use `cmd/workflow` for automated monitoring
✅ **ALWAYS do this**: Allow sufficient timeout for scan profile
✅ **ALWAYS do this**: Review generated workflow analysis markdown files
✅ **ALWAYS do this**: Check log file for detailed error messages AFTER tasks complete
✅ **ALWAYS do this**: Let the utility complete fully before checking any outputs
✅ **ALWAYS do this**: If monitoring is needed, open a SEPARATE PowerShell window and use: `Get-Content .\workflow-reports\*.log -Tail 20` (without -Wait flag, run periodically)
