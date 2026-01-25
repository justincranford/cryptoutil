# verify-docker.ps1
# Verifies Docker Desktop is running before running integration tests
# Usage: .\scripts\verify-docker.ps1

Write-Host "Verifying Docker Desktop is running..." -ForegroundColor Cyan

# Check if docker command exists
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Docker Desktop is not installed" -ForegroundColor Red
    Write-Host "Please install Docker Desktop from https://www.docker.com/products/docker-desktop" -ForegroundColor Yellow
    exit 1
}

# Check if Docker daemon is running
try {
    $null = docker ps 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Docker Desktop is not running" -ForegroundColor Red
        Write-Host "Please start Docker Desktop and wait for it to complete initialization (30-60 seconds)" -ForegroundColor Yellow
        exit 1
    }
} catch {
    Write-Host "❌ Docker Desktop is not running" -ForegroundColor Red
    Write-Host "Please start Docker Desktop and wait for it to complete initialization (30-60 seconds)" -ForegroundColor Yellow
    exit 1
}

Write-Host "✅ Docker Desktop is running" -ForegroundColor Green
exit 0
