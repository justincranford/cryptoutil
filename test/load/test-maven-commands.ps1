# Test Maven dependency-check commands matching CI workflow
# This script tests the exact commands used in .github/workflows/ci-sast.yml

$ErrorActionPreference = "Stop"

Write-Host "=== Testing OWASP Dependency Check Commands ===" -ForegroundColor Cyan
Write-Host ""

# Change to test/load directory
Set-Location -Path "$PSScriptRoot"

Write-Host "Current directory: $(Get-Location)" -ForegroundColor Yellow
Write-Host ""

# Test 1: Check if database exists from update-only
Write-Host "Test 1: Verify NVD database exists from update-only step" -ForegroundColor Green
if (Test-Path "target\dependency-check\data\dependency-check-db.mv.db") {
    $dbFile = Get-Item "target\dependency-check\data\dependency-check-db.mv.db"
    $sizeMB = [math]::Round($dbFile.Length/1MB, 2)
    Write-Host "  ✅ Database file exists: $sizeMB MB" -ForegroundColor Green
    Write-Host "  Last modified: $($dbFile.LastWriteTime)" -ForegroundColor Gray
} else {
    Write-Host "  ❌ Database file not found - update-only may have failed" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Test 2: Run dependency-check (matches CI workflow check step)
Write-Host "Test 2: Run dependency-check with populated database" -ForegroundColor Green
Write-Host "  Command: mvnw dependency-check:check with autoUpdate=false" -ForegroundColor Gray

$checkArgs = @(
    "org.owasp:dependency-check-maven:check"
    "-Ddependency-check.dataDirectory=target/dependency-check/data"
    "-Ddependency-check.connectionString=jdbc:h2:file:target/dependency-check/data/dependency-check-db;DB_CLOSE_ON_EXIT=FALSE"
    "-Ddependency-check.autoUpdate=false"
    "-Ddependency-check.failOnError=false"
)

& .\mvnw @checkArgs

if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✅ Dependency check completed successfully" -ForegroundColor Green
} else {
    Write-Host "  ❌ Dependency check failed with exit code: $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}
Write-Host ""

# Test 3: Verify report generation
Write-Host "Test 3: Verify dependency-check reports generated" -ForegroundColor Green

$reportFiles = @(
    "target\dependency-check-report.html",
    "target\dependency-check\dependency-check-report.sarif"
)

foreach ($reportFile in $reportFiles) {
    if (Test-Path $reportFile) {
        $size = (Get-Item $reportFile).Length
        Write-Host "  ✅ Report exists: $reportFile ($size bytes)" -ForegroundColor Green
    } else {
        Write-Host "  ⚠️  Report not found: $reportFile" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "=== All Tests Completed ===" -ForegroundColor Cyan
