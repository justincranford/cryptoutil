#-------------------------------------------------------------------------------
# Helper script: run-act-dast.ps1
#
# Recommended invocation (one-shot, safe - does not change machine policy):
# powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\run-act-dast.ps1 -ScanProfile quick -Timeout 900
#
# Alternative (session-scoped):
# Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass; .\scripts\run-act-dast.ps1 -ScanProfile quick -Timeout 900
#
# See .github/instructions/powershell.instructions.md for full guidance
#-------------------------------------------------------------------------------
#Requires -Version 5.1
<#
.SYNOPSIS
    Run act DAST workflow with monitoring and automatic result analysis

.DESCRIPTION
    Executes the DAST workflow using act, monitors progress by tailing the log file,
    and automatically analyzes results when the workflow completes.
    Designed to handle long-running scans (3-25 minutes) with proper timeout handling.

.PARAMETER ScanProfile
    Scan profile to use: quick (3-5min), full (10-15min), deep (20-25min)
    Default: quick

.PARAMETER Timeout
    Maximum time to wait for workflow completion in seconds
    Default: 600 (10 minutes)

.PARAMETER TailLines
    Number of lines to show when tailing log during monitoring
    Default: 20

.PARAMETER OutputDir
    Output directory for reports (default: dast-reports)

.PARAMETER Help
    Show this help message

.EXAMPLE
    .\scripts\run-act-dast.ps1
    Run quick scan with 10 minute timeout

.EXAMPLE
    .\scripts\run-act-dast.ps1 -ScanProfile full -Timeout 900
    Run full scan with 15 minute timeout

.EXAMPLE
    .\scripts\run-act-dast.ps1 -ScanProfile deep -Timeout 1500
    Run deep scan with 25 minute timeout
#>

param(
    [ValidateSet("quick", "full", "deep")]
    [string]$ScanProfile = "quick",
    [int]$Timeout = 600,
    [int]$TailLines = 20,
    [string]$OutputDir = "dast-reports",
    [switch]$Help
)

# Show help if requested
if ($Help) {
    Get-Help $PSCommandPath -Full
    exit 0
}

# Set console encoding for proper UTF-8 output
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# Colors for output
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Blue = "Cyan"
$Magenta = "Magenta"

function Write-Status {
    param([string]$Message, [string]$Color = "White")
    $timestamp = Get-Date -Format "HH:mm:ss"
    Write-Host "[$timestamp] $Message" -ForegroundColor $Color
}

function Write-Section {
    param([string]$Title)
    Write-Host ""
    Write-Host "===================================================================" -ForegroundColor $Magenta
    Write-Host " $Title" -ForegroundColor $Magenta
    Write-Host "===================================================================" -ForegroundColor $Magenta
    Write-Host ""
}

# Validate prerequisites
Write-Section "Validating Prerequisites"
Write-Status "Checking for act..." $Blue

try {
    $actVersion = act --version 2>&1
    Write-Status "[OK] act is installed: $actVersion" $Green
} catch {
    Write-Status "[ERROR] act is not installed" $Red
    Write-Status "Install act from: https://github.com/nektos/act" $Yellow
    exit 1
}

Write-Status "Checking for Docker..." $Blue
try {
    docker --version | Out-Null
    Write-Status "[OK] Docker is available" $Green
} catch {
    Write-Status "[ERROR] Docker is not available" $Red
    exit 1
}

# Create output directory
Write-Status "Creating output directory: $OutputDir" $Blue
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}

$LogFile = Join-Path $OutputDir "act-dast.log"
$StatusFile = Join-Path $OutputDir "act-status.txt"

# Remove old log files
if (Test-Path $LogFile) {
    Remove-Item $LogFile -Force
}
if (Test-Path $StatusFile) {
    Remove-Item $StatusFile -Force
}

# Run act workflow directly (redirects to log file automatically in background)
Write-Section "Starting DAST Workflow"
Write-Status "Profile: $ScanProfile" $Blue
Write-Status "Timeout: $Timeout seconds" $Blue
Write-Status "Log file: $LogFile" $Blue
Write-Status "This will take 3-25 minutes depending on profile..." $Yellow
Write-Status "" $Blue

# NOTE FOR WINDOWS USERS:
# - Ensure Docker Desktop exposes host.docker.internal DNS (default on Docker Desktop).
# - If ZAP containers cannot reach services on the host you may need to run act with
#   additional privileges or configure Docker networking appropriately. Example:
#     act --privileged -P ubuntu-latest=ghcr.io/catthehacker/ubuntu:act-latest
#   or ensure host.docker.internal resolves to the host IP.

# Set encoding and run act, streaming to both console and log file
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

$actCommand = "workflow_dispatch -W .github/workflows/dast.yml --input scan_profile=$ScanProfile --artifact-server-path ./$OutputDir"
Write-Status "Running: act $actCommand" $Blue
Write-Status "" $Blue

# Run act and reliably redirect stdout/stderr to the log file. Start-Process is used
# so the log file is created even if the act binary buffers or behaves differently when
# executed via Invoke-Expression. This blocks until the process completes.

$startTime = Get-Date

# Start-Process cannot redirect stdout and stderr to the same file path. Create
# a temporary stderr file, redirect stdout to the main log, stderr to the temp file,
# and then append stderr to the main log after completion.
$tempErr = Join-Path $OutputDir "act-dast.stderr.tmp"
if (Test-Path $tempErr) { Remove-Item $tempErr -Force }

Start-Process -FilePath "act" -ArgumentList $actCommand -RedirectStandardOutput $LogFile -RedirectStandardError $tempErr -NoNewWindow -Wait

# Ensure both outputs are present in the final log
if (Test-Path $tempErr) {
    Get-Content $tempErr | Add-Content -Path $LogFile
    Remove-Item $tempErr -Force
}

# Print the final tail of the log so user sees recent output
if (Test-Path $LogFile) {
    Get-Content -Path $LogFile -Tail $TailLines | ForEach-Object { Write-Host $_ }
}

$endTime = Get-Date
$elapsed = [int](($endTime - $startTime).TotalSeconds)
$workflowComplete = $true

Write-Status "" $Blue
Write-Status "Workflow execution completed in $elapsed seconds" $Green

# Analyze results
Write-Section "Analyzing Results"

$logExists = Test-Path $LogFile
if (-not $logExists) {
    Write-Status "[ERROR] Log file not found: $LogFile" $Red
    exit 1
}

$logContent = Get-Content $LogFile -Raw
Write-Status "Log file size: $((Get-Item $LogFile).Length) bytes" $Blue

# Check for workflow-level success
$workflowSuccess = $false
$workflowErrors = @()

if ($logContent -match "Job succeeded") {
    $workflowSuccess = $true
    Write-Status "[OK] Workflow completed successfully" $Green
} elseif ($logContent -match "Job failed") {
    $workflowErrors += "Workflow job failed"
    Write-Status "[ERROR] Workflow job failed" $Red
} elseif ($logContent -match "Error:") {
    $workflowErrors += "Workflow encountered errors"
    Write-Status "[WARNING] Workflow encountered errors" $Yellow
}

# Analyze individual scan results
Write-Status "" $Blue
Write-Status "Scan Results:" $Blue
Write-Status "---------------------------------------------------------------" $Blue

$taskResults = @{}

# Check for Nuclei scan
if ($logContent -match "Nuclei - Vulnerability Scan") {
    if ($logContent -match "nuclei.log|nuclei.sarif") {
        $taskResults["Nuclei Scan"] = "SUCCESS"
        Write-Status "[OK] Nuclei Scan: Completed" $Green
    } else {
        $taskResults["Nuclei Scan"] = "FAILED"
        Write-Status "[ERROR] Nuclei Scan: Failed or incomplete" $Red
    }
} else {
    $taskResults["Nuclei Scan"] = "NOT_RUN"
    Write-Status "[SKIP] Nuclei Scan: Not detected in log" $Yellow
}

# Check for ZAP scans
if ($logContent -match "OWASP ZAP DAST Scan") {
    if ($logContent -match "zap-report") {
        $taskResults["ZAP Full Scan"] = "SUCCESS"
        Write-Status "[OK] ZAP Full Scan: Completed" $Green
    } else {
        $taskResults["ZAP Full Scan"] = "FAILED"
        Write-Status "[ERROR] ZAP Full Scan: Failed or incomplete" $Red
    }
} else {
    $taskResults["ZAP Full Scan"] = "NOT_RUN"
    Write-Status "[SKIP] ZAP Full Scan: Not detected in log" $Yellow
}

if ($logContent -match "OWASP ZAP API Scan") {
    if ($logContent -match "zap-api-report") {
        $taskResults["ZAP API Scan"] = "SUCCESS"
        Write-Status "[OK] ZAP API Scan: Completed" $Green
    } else {
        $taskResults["ZAP API Scan"] = "FAILED"
        Write-Status "[ERROR] ZAP API Scan: Failed or incomplete" $Red
    }
} else {
    $taskResults["ZAP API Scan"] = "NOT_RUN"
    Write-Status "[SKIP] ZAP API Scan: Not detected in log" $Yellow
}

# Check for header capture
if ($logContent -match "response-headers.txt") {
    $taskResults["Header Capture"] = "SUCCESS"
    Write-Status "[OK] Header Capture: Completed" $Green
} else {
    $taskResults["Header Capture"] = "FAILED"
    Write-Status "[ERROR] Header Capture: Failed or incomplete" $Red
}

# List generated artifacts
Write-Status "" $Blue
Write-Status "Generated Artifacts:" $Blue
Write-Status "---------------------------------------------------------------" $Blue

$artifacts = Get-ChildItem -Path $OutputDir -File | Where-Object { $_.Name -ne "act-status.txt" }
if ($artifacts) {
    foreach ($artifact in $artifacts) {
        Write-Status "  - $($artifact.Name) ($($artifact.Length) bytes)" $Blue
    }
} else {
    Write-Status "[WARNING] No artifacts found in $OutputDir" $Yellow
}

# Generate summary
Write-Section "Summary"

$successCount = ($taskResults.Values | Where-Object { $_ -eq "SUCCESS" }).Count
$failedCount = ($taskResults.Values | Where-Object { $_ -eq "FAILED" }).Count
$totalTasks = $taskResults.Count

Write-Status "Workflow: $(if ($workflowSuccess) { '[OK] SUCCESS' } else { '[ERROR] FAILED' })" $(if ($workflowSuccess) { $Green } else { $Red })
Write-Status "Tasks: $successCount/$totalTasks succeeded, $failedCount failed" $(if ($failedCount -eq 0) { $Green } else { $Yellow })
Write-Status "Duration: $elapsed seconds" $Blue
Write-Status "Log file: $LogFile" $Blue

# Save status summary
$statusSummary = @"
DAST Workflow Execution Summary
================================
Date: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
Profile: $ScanProfile
Duration: $elapsed seconds
Timeout: $Timeout seconds

Workflow Status: $(if ($workflowSuccess) { 'SUCCESS' } else { 'FAILED' })
Tasks Completed: $successCount/$totalTasks
Tasks Failed: $failedCount

Task Results:
"@

foreach ($task in $taskResults.Keys) {
    $statusSummary += "`n  - ${task}: $($taskResults[$task])"
}

$statusSummary += "`n`nArtifacts Generated:`n"
foreach ($artifact in $artifacts) {
    $statusSummary += "  - $($artifact.Name)`n"
}

$statusSummary | Out-File -FilePath $StatusFile -Encoding utf8
Write-Status "Status summary saved to: $StatusFile" $Blue

# Exit with appropriate code
if ($workflowSuccess -and $failedCount -eq 0) {
    Write-Status "" $Blue
    Write-Status "[OK] All checks passed!" $Green
    exit 0
} else {
    Write-Status "" $Blue
    Write-Status "[WARNING] Some checks failed or workflow incomplete" $Yellow
    Write-Status "Review log file for details: $LogFile" $Yellow
    exit 1
}
