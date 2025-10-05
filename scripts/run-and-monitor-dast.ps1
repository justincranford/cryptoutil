#!/usr/bin/env pwsh
#-------------------------------------------------------------------------------
# Helper script: run-and-monitor-dast.ps1
#
# Recommended invocation (one-shot, safe - does not change machine policy):
# powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\run-and-monitor-dast.ps1 -ScanProfile quick
#
# See .github/instructions/powershell.instructions.md for full guidance
#-------------------------------------------------------------------------------
#Requires -Version 7.0
<#
.SYNOPSIS
    Run act DAST workflow and automatically monitor progress until completion.

.DESCRIPTION
    This script:
    1. Starts the act DAST workflow in the background
    2. Continuously monitors and displays progress
    3. Detects when workflow completes
    4. Automatically analyzes results
    5. Reports success/failure of each step including ZAP connectivity check

.PARAMETER ScanProfile
    Scan profile to use (quick, full, deep). Default: quick

.PARAMETER CheckIntervalSeconds
    Seconds between log checks. Default: 5

.EXAMPLE
    .\scripts\run-and-monitor-dast.ps1

.EXAMPLE
    .\scripts\run-and-monitor-dast.ps1 -ScanProfile full
#>

param(
    [ValidateSet('quick', 'full', 'deep')]
    [string]$ScanProfile = 'quick',

    [int]$CheckIntervalSeconds = 5
)

$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

$LogFile = ".\dast-reports\act-dast.log"
$WorkflowFile = ".\.github\workflows\dast.yml"

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "DAST Workflow Runner & Monitor" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "Profile: $ScanProfile" -ForegroundColor Yellow
Write-Host "Log: $LogFile" -ForegroundColor Yellow
Write-Host "Check interval: $CheckIntervalSeconds seconds" -ForegroundColor Yellow
Write-Host ""

# Clean up old log if exists
if (Test-Path $LogFile) {
    Write-Host "Removing old log file..." -ForegroundColor Gray
    Remove-Item $LogFile -Force
}

# Start workflow in background
Write-Host "Starting act workflow..." -ForegroundColor Green
$actCommand = "act workflow_dispatch -W $WorkflowFile --input scan_profile=$ScanProfile --artifact-server-path ./dast-reports 2>&1 | Out-File -FilePath $LogFile -Encoding utf8"
$job = Start-Job -ScriptBlock {
    param($cmd)
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
    Invoke-Expression $cmd
} -ArgumentList $actCommand

Write-Host "Workflow started (Job ID: $($job.Id))" -ForegroundColor Green
Write-Host "Monitoring progress..." -ForegroundColor Cyan
Write-Host ""

$startTime = Get-Date
$lastSize = 0
$lastLines = @()
$stepsFound = @{}

# Key steps to track
$keySteps = @{
    'Set up job' = $false
    'Checkout code' = $false
    'Build application' = $false
    'Start application' = $false
    'Capture baseline response headers' = $false
    'Verify ZAP can reach target' = $false
    'Run OWASP ZAP DAST Scan' = $false
    'Run OWASP ZAP API Scan' = $false
    'Nuclei - Vulnerability Scan' = $false
    'Collect scan artifacts' = $false
}

try {
    while ($job.State -eq 'Running' -or !(Test-Path $LogFile)) {
        Start-Sleep -Seconds $CheckIntervalSeconds

        if (Test-Path $LogFile) {
            $currentSize = (Get-Item $LogFile).Length

            if ($currentSize -gt $lastSize) {
                # Get new content
                $allContent = Get-Content $LogFile -Raw -Encoding UTF8
                $newContent = $allContent.Substring($lastSize)

                # Display only last 10 lines of new content
                $newLines = $newContent -split "`n" | Where-Object { $_.Trim() -ne '' } | Select-Object -Last 10

                foreach ($line in $newLines) {
                    # Color code based on content
                    if ($line -match '✅|Success') {
                        Write-Host $line -ForegroundColor Green
                    }
                    elseif ($line -match '❌|Failure|error|ERROR|failed') {
                        Write-Host $line -ForegroundColor Red
                    }
                    elseif ($line -match '⭐|Run Main') {
                        Write-Host $line -ForegroundColor Yellow
                    }
                    elseif ($line -match 'Verify ZAP') {
                        Write-Host $line -ForegroundColor Magenta
                    }
                    else {
                        Write-Host $line -ForegroundColor Gray
                    }

                    # Track steps
                    foreach ($step in $keySteps.Keys) {
                        if ($line -match [regex]::Escape($step)) {
                            $stepsFound[$step] = $true
                        }
                    }
                }

                # Show progress indicator
                $elapsed = (Get-Date) - $startTime
                $completedSteps = ($stepsFound.Values | Where-Object { $_ }).Count
                Write-Host "`n[$($elapsed.ToString('mm\:ss'))] Size: $currentSize bytes | Steps: $completedSteps/$($keySteps.Count)" -ForegroundColor Cyan
                Write-Host "----------------------------------------" -ForegroundColor DarkGray

                $lastSize = $currentSize
            }
        }

        # Check if job failed
        if ($job.State -eq 'Failed') {
            Write-Host "`nJob failed!" -ForegroundColor Red
            break
        }
    }

    # Wait for job to complete
    $job | Wait-Job | Out-Null

    Write-Host "`n======================================" -ForegroundColor Cyan
    Write-Host "WORKFLOW COMPLETED" -ForegroundColor Cyan
    Write-Host "======================================" -ForegroundColor Cyan

    $totalElapsed = (Get-Date) - $startTime
    Write-Host "Total time: $($totalElapsed.ToString('mm\:ss'))" -ForegroundColor Yellow
    Write-Host ""

    # Analyze results
    Write-Host "ANALYZING RESULTS..." -ForegroundColor Cyan
    Write-Host ""

    if (Test-Path $LogFile) {
        $logContent = Get-Content $LogFile -Raw

        # Check for key outcomes
        Write-Host "Step Analysis:" -ForegroundColor Yellow
        Write-Host "-------------" -ForegroundColor Yellow

        foreach ($step in $keySteps.Keys) {
            if ($logContent -match "✅.*$([regex]::Escape($step))") {
                Write-Host "✅ $step" -ForegroundColor Green
            }
            elseif ($logContent -match "❌.*$([regex]::Escape($step))") {
                Write-Host "❌ $step" -ForegroundColor Red
            }
            elseif ($logContent -match $([regex]::Escape($step))) {
                Write-Host "⏸️  $step (started but not completed)" -ForegroundColor Yellow
            }
            else {
                Write-Host "⏭️  $step (skipped/not reached)" -ForegroundColor Gray
            }
        }

        Write-Host ""
        Write-Host "ZAP Connectivity Check:" -ForegroundColor Magenta
        Write-Host "---------------------" -ForegroundColor Magenta

        if ($logContent -match "Verify ZAP can reach target") {
            Write-Host "✅ Connectivity check step was executed" -ForegroundColor Green

            if ($logContent -match "Target.*is NOT reachable from Docker container") {
                Write-Host "❌ ZAP connectivity check FAILED (as expected - Docker networking issue)" -ForegroundColor Red
                Write-Host "   This confirms the check is working correctly!" -ForegroundColor Yellow
            }
            elseif ($logContent -match "Target is reachable from Docker container") {
                Write-Host "✅ ZAP connectivity check PASSED (target is reachable)" -ForegroundColor Green
            }
            else {
                Write-Host "⚠️  ZAP connectivity check result unclear" -ForegroundColor Yellow
            }
        }
        else {
            Write-Host "⏭️  ZAP connectivity check was not executed (workflow stopped earlier)" -ForegroundColor Gray
        }

        Write-Host ""
        Write-Host "Overall Status:" -ForegroundColor Cyan
        Write-Host "--------------" -ForegroundColor Cyan

        if ($logContent -match "Job succeeded") {
            Write-Host "✅ Workflow SUCCEEDED" -ForegroundColor Green
        }
        elseif ($logContent -match "Job failed") {
            Write-Host "❌ Workflow FAILED" -ForegroundColor Red
        }
        else {
            Write-Host "⚠️  Workflow status unclear" -ForegroundColor Yellow
        }

        # Check for specific errors
        $errors = $logContent | Select-String -Pattern "ERROR|error|failed to" -AllMatches
        if ($errors.Matches.Count -gt 0) {
            Write-Host "`nErrors found: $($errors.Matches.Count)" -ForegroundColor Red
            Write-Host "See $LogFile for details" -ForegroundColor Yellow
        }
    }
    else {
        Write-Host "❌ Log file not found!" -ForegroundColor Red
    }

    Write-Host ""
    Write-Host "======================================" -ForegroundColor Cyan
    Write-Host "Full log available at: $LogFile" -ForegroundColor Gray

}
finally {
    # Cleanup
    if ($job.State -eq 'Running') {
        Write-Host "Stopping job..." -ForegroundColor Yellow
        Stop-Job -Job $job
    }
    Remove-Job -Job $job -Force -ErrorAction SilentlyContinue
}
