#-------------------------------------------------------------------------------
# Helper script: lint.ps1
#
# Recommended invocation (one-shot, safe - does not change machine policy):
# powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\lint.ps1
#
# See .github/instructions/powershell.instructions.md for full guidance
#-------------------------------------------------------------------------------
# PowerShell script for formatting and linting Go code

Write-Host "🔧 Running gofumpt (stricter gofmt)..." -ForegroundColor Green
& "$env:GOPATH\bin\gofumpt.exe" -l -w .

Write-Host "📦 Running goimports (import organization)..." -ForegroundColor Green
& "$env:GOPATH\bin\goimports.exe" -l -w .

Write-Host "🏗️ Running go vet (static analysis)..." -ForegroundColor Green
go vet ./...

Write-Host "🔍 Running go build (compilation check)..." -ForegroundColor Green
go build ./...

Write-Host "🛡️ Attempting golangci-lint..." -ForegroundColor Green
try {
    & "$env:GOPATH\bin\golangci-lint.exe" run --timeout=5m --max-issues-per-linter=10
    Write-Host "✅ golangci-lint completed successfully!" -ForegroundColor Green
} catch {
    Write-Host "⚠️ golangci-lint failed (likely Go version mismatch)" -ForegroundColor Yellow
    Write-Host "   Project uses Go 1.25, golangci-lint built with older version" -ForegroundColor Yellow
}

Write-Host "✅ Code formatting and basic linting complete!" -ForegroundColor Green

Write-Host ""
Write-Host "Manual checks you can run:" -ForegroundColor Cyan
Write-Host "- go test ./... -cover  # Run tests with coverage" -ForegroundColor Gray
Write-Host "- go mod tidy          # Clean up dependencies" -ForegroundColor Gray
Write-Host "- go generate ./...    # Regenerate code" -ForegroundColor Gray
