#!/usr/bin/env pwsh
#-------------------------------------------------------------------------------
# Helper script: mutation-test.ps1
#
# Recommended invocation (one-shot, safe - does not change machine policy):
# powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\mutation-test.ps1
#
# See .github/instructions/powershell.instructions.md for full guidance
#-------------------------------------------------------------------------------
<#
.SYNOPSIS
    Run mutation testing with Gremlins for the cryptoutil project
.DESCRIPTION
    This script runs mutation testing to validate test suite quality.
    It runs Gremlins mutation testing tool on selected packages with good test coverage.
.PARAMETER Target
    The target package or directory to test (default: high-coverage packages)
.PARAMETER DryRun
    Run in dry-run mode without executing tests
.PARAMETER Workers
    Number of parallel workers (default: 2)
.PARAMETER TimeoutCoeff
    Timeout coefficient multiplier (default: 3)
#>

param(
    [string]$Target = "",
    [switch]$DryRun = $false,
    [int]$Workers = 2,
    [int]$TimeoutCoeff = 3
)

# Exit on any error
$ErrorActionPreference = "Stop"

Write-Host "🧪 Starting Mutation Testing with Gremlins" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green

# Check if gremlins is installed
if (-not (Get-Command "gremlins" -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Gremlins not found. Installing..." -ForegroundColor Yellow
    go install github.com/go-gremlins/gremlins/cmd/gremlins@latest
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Failed to install Gremlins" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Gremlins installed successfully" -ForegroundColor Green
}

# High-coverage packages to test
$highCoveragePackages = @(
    "./internal/common/util/datetime/",
    "./internal/common/util/thread/",
    "./internal/common/util/sysinfo/",
    "./internal/common/util/combinations/",
    "./internal/common/crypto/certificate/",
    "./internal/common/crypto/digests/"
)

# If no target specified, use high-coverage packages
if ($Target -eq "") {
    $targetsToTest = $highCoveragePackages
} else {
    $targetsToTest = @($Target)
}

$totalKilled = 0
$totalLived = 0
$totalNotCovered = 0
$totalTimedOut = 0
$totalNotViable = 0

Write-Host "📊 Target packages:" -ForegroundColor Cyan
$targetsToTest | ForEach-Object { Write-Host "  - $_" -ForegroundColor White }
Write-Host ""

foreach ($package in $targetsToTest) {
    Write-Host "🎯 Testing package: $package" -ForegroundColor Cyan
    Write-Host "----------------------------------------" -ForegroundColor Gray

    # Build the gremlins command
    $cmd = @(
        "unleash"
        $package
        "--workers", $Workers
        "--timeout-coefficient", $TimeoutCoeff
        "--output", "gremlins-$($package -replace '[/\\]', '').json"
    )

    if ($DryRun) {
        $cmd += "--dry-run"
    }

    # Run gremlins and capture output
    try {
        $output = & gremlins @cmd
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Mutation testing completed for $package" -ForegroundColor Green
        } else {
            Write-Host "⚠️  Mutation testing completed with warnings for $package" -ForegroundColor Yellow
        }

        # Parse results from output
        $output | ForEach-Object {
            if ($_ -match "Killed: (\d+), Lived: (\d+), Not covered: (\d+)") {
                $totalKilled += [int]$matches[1]
                $totalLived += [int]$matches[2]
                $totalNotCovered += [int]$matches[3]
            }
            if ($_ -match "Timed out: (\d+), Not viable: (\d+)") {
                $totalTimedOut += [int]$matches[1]
                $totalNotViable += [int]$matches[2]
            }
        }

        Write-Host $output -ForegroundColor White

    } catch {
        Write-Host "❌ Failed to run mutation testing on $package" -ForegroundColor Red
        Write-Host $_.Exception.Message -ForegroundColor Red
    }

    Write-Host ""
}

# Summary
Write-Host "📈 MUTATION TESTING SUMMARY" -ForegroundColor Green
Write-Host "===========================" -ForegroundColor Green
Write-Host "Total Killed: $totalKilled" -ForegroundColor Green
Write-Host "Total Lived: $totalLived" -ForegroundColor Red
Write-Host "Total Not Covered: $totalNotCovered" -ForegroundColor Yellow
Write-Host "Total Timed Out: $totalTimedOut" -ForegroundColor Cyan
Write-Host "Total Not Viable: $totalNotViable" -ForegroundColor Gray

$totalTested = $totalKilled + $totalLived
if ($totalTested -gt 0) {
    $efficacy = ($totalKilled / $totalTested) * 100
    Write-Host "Test Efficacy: $($efficacy.ToString('F2'))%" -ForegroundColor $(if ($efficacy -ge 75) { "Green" } else { "Red" })
} else {
    Write-Host "Test Efficacy: N/A (no mutations tested)" -ForegroundColor Yellow
}

# Set exit code based on results
if ($totalLived -gt 0 -and -not $DryRun) {
    Write-Host "❌ Mutation testing found $totalLived survived mutations" -ForegroundColor Red
    exit 1
} else {
    Write-Host "✅ Mutation testing completed successfully" -ForegroundColor Green
    exit 0
}
