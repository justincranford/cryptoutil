---
description: "Instructions for testing GitHub Actions workflows locally with act"
applyTo: "**"
---
# Act Testing Instructions

## CRITICAL: DAST Workflow Testing

**NEVER use short timeouts when running DAST workflows with act**

### Recommended Approach: Use run-act-dast.ps1 Script

**ALWAYS use the provided script for running act DAST workflows**

```powershell
# Quick scan (3-5 minutes) with 10 minute timeout
.\scripts\run-act-dast.ps1

# Full scan (10-15 minutes) with 15 minute timeout
.\scripts\run-act-dast.ps1 -ScanProfile full -Timeout 900

# Deep scan (20-25 minutes) with 25 minute timeout
.\scripts\run-act-dast.ps1 -ScanProfile deep -Timeout 1500
```

**Script features:**
- Runs act workflow directly (no background jobs)
- Streams output to both console and log file
- Automatic workflow completion detection
- Comprehensive result analysis (workflow + task status)
- Artifact verification and summary report
- Single command execution - no prompts

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

The `run-act-dast.ps1` script automatically analyzes:

1. **Workflow Status**: Job succeeded/failed detection
2. **Task Status**: Individual scan completion verification
   - Nuclei Scan (log/SARIF generation)
   - ZAP Full Scan (report generation)
   - ZAP API Scan (report generation)
   - Header Capture (security headers file)
3. **Artifact Generation**: Verification of output files
4. **Summary Report**: Saved to `dast-reports/act-status.txt`

### Success Criteria

✓ **Workflow Success**: `Job succeeded` in log output
✓ **Task Success**: All scan artifacts generated in `dast-reports/`
✓ **No Errors**: No `Job failed` or critical error messages

### Common Mistakes to AVOID

❌ **NEVER do this**: Using `-t` timeout flag or checking output too early
❌ **NEVER do this**: `Start-Sleep -Seconds 60` (way too short)
❌ **NEVER do this**: Checking terminal output while workflow is running
❌ **NEVER do this**: Running act commands directly without monitoring/analysis

✅ **ALWAYS do this**: Use `run-act-dast.ps1` script for automated monitoring
✅ **ALWAYS do this**: Allow sufficient timeout for scan profile
✅ **ALWAYS do this**: Review generated `act-status.txt` summary
✅ **ALWAYS do this**: Check log file for detailed error messages if tasks fail
