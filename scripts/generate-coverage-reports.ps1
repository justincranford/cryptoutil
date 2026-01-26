# Generate HTML coverage reports for packages below 90%

$packages = @(
    @{name="cipher-im-server-config"; path="./internal/apps/cipher/im/server/config"; coverage="80.4%"},
    @{name="cipher-im-server-apis"; path="./internal/apps/cipher/im/server/apis"; coverage="82.1%"},
    @{name="jose-ja-server-config"; path="./internal/apps/jose/ja/server/config"; coverage="61.9%"},
    @{name="jose-ja-service"; path="./internal/apps/jose/ja/service"; coverage="87.3%"},
    @{name="template-server-repository"; path="./internal/apps/template/service/server/repository"; coverage="84.8%"},
    @{name="template-service-config"; path="./internal/apps/template/service/config"; coverage="81.3%"},
    @{name="template-service-config-tls"; path="./internal/apps/template/service/config/tls_generator"; coverage="80.6%"},
    @{name="template-server-businesslogic"; path="./internal/apps/template/service/server/businesslogic"; coverage="75.2%"},
    @{name="template-server-listener"; path="./internal/apps/template/service/server/listener"; coverage="70.7%"},
    @{name="template-server-barrier"; path="./internal/apps/template/service/server/barrier"; coverage="72.6%"}
)

# Create output directory
$outputDir = "test-output/coverage-html"
New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

Write-Host "`n=== Generating HTML Coverage Reports ===`n" -ForegroundColor Cyan

foreach ($pkg in $packages) {
    $name = $pkg.name
    $path = $pkg.path
    $expected = $pkg.coverage

    $covFile = "$outputDir/$name.cov"
    $htmlFile = "$outputDir/$name.html"

    Write-Host "[$name] Generating coverage..." -ForegroundColor Yellow
    Write-Host "  Path: $path" -ForegroundColor Gray
    Write-Host "  Expected: $expected" -ForegroundColor Gray

    # Run coverage test
    go test -coverprofile=$covFile $path 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ❌ Test failed" -ForegroundColor Red
        continue
    }

    # Generate HTML
    go tool cover -html=$covFile -o $htmlFile
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✅ Generated: $htmlFile" -ForegroundColor Green
    } else {
        Write-Host "  ❌ HTML generation failed" -ForegroundColor Red
    }

    Write-Host ""
}

Write-Host "=== Coverage Reports Complete ===`n" -ForegroundColor Cyan
Write-Host "Reports location: $outputDir" -ForegroundColor Green
