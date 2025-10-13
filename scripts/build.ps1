#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Docker Build Script with Mandatory Args Validation

.DESCRIPTION
    Builds the cryptoutil Docker image with proper versioning and validation.

.PARAMETER AppVersion
    Application version (mandatory)

.EXAMPLE
    .\build.ps1 -AppVersion v1.0.0
#>

param(
    [Parameter(Mandatory = $true)]
    [string]$AppVersion
)

Write-Host "Building cryptoutil with version: $AppVersion" -ForegroundColor Green
Write-Host "VCS_REF will be set to current commit hash" -ForegroundColor Cyan
Write-Host "BUILD_DATE will be set to current timestamp" -ForegroundColor Cyan

# Get current commit hash
$VcsRef = git rev-parse HEAD 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to get git commit hash. Are you in a git repository?"
    exit 1
}

# Get current timestamp
$BuildDate = Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ"

Write-Host "Using VCS_REF: $VcsRef" -ForegroundColor Yellow
Write-Host "Using BUILD_DATE: $BuildDate" -ForegroundColor Yellow

$dockerArgs = @(
    "build",
    "--build-arg", "APP_VERSION=$AppVersion",
    "--build-arg", "VCS_REF=$VcsRef",
    "--build-arg", "BUILD_DATE=$BuildDate",
    "-t", "cryptoutil:$AppVersion",
    "-f", "deployments/Dockerfile",
    "."
)

Write-Host "Running: docker $dockerArgs" -ForegroundColor Gray

& docker @dockerArgs

if ($LASTEXITCODE -eq 0) {
    Write-Host "SUCCESS: Docker image built as cryptoutil:$AppVersion" -ForegroundColor Green
} else {
    Write-Error "ERROR: Docker build failed"
    exit 1
}
