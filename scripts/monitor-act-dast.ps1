#!/usr/bin/env pwsh
#Requires -Version 7.0

<#
.SYNOPSIS
    Monitor act DAST workflow execution progress.

.DESCRIPTION
    Periodically checks the act-dast.log file and displays new content.
    Continues until workflow completes (success or failure) or user stops (Ctrl+C).

.PARAMETER LogFile
    Path to the act log file (default: ./dast-reports/act-dast.log)

.PARAMETER IntervalSeconds
    Seconds between log checks (default: 30)

.EXAMPLE
    .\scripts\monitor-act-dast.ps1

.EXAMPLE
    .\scripts\monitor-act-dast.ps1 -IntervalSeconds 60
#>

param(
    [string]$LogFile = ".\dast-reports\act-dast.log",
    [int]$IntervalSeconds = 30
)

$ErrorActionPreference = 'Stop'

Write-Host "Monitoring DAST workflow: $LogFile" -ForegroundColor Cyan
Write-Host "Checking every $IntervalSeconds seconds. Press Ctrl+C to stop." -ForegroundColor Yellow
Write-Host ""

$lastSize = 0
$startTime = Get-Date

while ($true) {
    if (Test-Path $LogFile) {
        $currentSize = (Get-Item $LogFile).Length

        if ($currentSize -gt $lastSize) {
            $newContent = Get-Content $LogFile -Raw -Encoding UTF8
            $newLines = $newContent.Substring($lastSize)
            Write-Host $newLines -NoNewline

            # Check for completion markers
            if ($newLines -match "Job (succeeded|failed|cancelled)") {
                $elapsed = (Get-Date) - $startTime
                Write-Host ""
                Write-Host "Workflow completed after $($elapsed.ToString('mm\:ss'))" -ForegroundColor Green
                break
            }

            $lastSize = $currentSize
        }
    } else {
        Write-Host "Waiting for log file to be created..." -ForegroundColor Yellow
    }

    Start-Sleep -Seconds $IntervalSeconds
}
