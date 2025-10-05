---
description: "Instructions for testing GitHub Actions workflows locally with act"
applyTo: "**"
---
# Act Testing Instructions

## CRITICAL: DAST Workflow Testing

**NEVER use short timeouts when running DAST workflows with act**

### Correct Usage

```powershell
# Run in BACKGROUND with proper output redirection (no timeout - let it complete)
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
act workflow_dispatch -W .github/workflows/dast.yml --input scan_profile=quick --artifact-server-path ./dast-reports 2>&1 | Out-File -FilePath .\dast-reports\act-dast.log -Encoding utf8
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

### Monitoring Workflow Progress

**DO NOT check terminal output during execution** - this can interrupt/cancel the workflow

Wait at least 5 minutes before checking progress:
```powershell
# After workflow completes, check the log
Get-Content .\dast-reports\act-dast.log -Tail 100
```

### Common Mistakes to AVOID

❌ **NEVER do this**: Using `-t` timeout flag or checking output too early
❌ **NEVER do this**: `Start-Sleep -Seconds 60` (way too short)
❌ **NEVER do this**: Checking terminal output while workflow is running

✅ **ALWAYS do this**: Run in background, redirect to file, wait 5+ minutes
✅ **ALWAYS do this**: Use proper PowerShell encoding settings
✅ **ALWAYS do this**: Check log file after completion
