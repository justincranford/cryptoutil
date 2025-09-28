#Requires -Version 5.1
<#
.SYNOPSIS
    Run comprehensive security scans locally for cryptoutil

.DESCRIPTION
    This script runs the same security scans as CI/CD pipeline locally for development workflow.
    Includes static analysis, vulnerability scanning, container security, and dependency analysis.

.PARAMETER All
    Run all security scans (default)

.PARAMETER StaticOnly
    Run only static analysis tools (staticcheck, golangci-lint)

.PARAMETER VulnOnly
    Run only vulnerability scans (govulncheck, trivy)

.PARAMETER ContainerOnly
    Run only container security scans (trivy image, docker scout)

.PARAMETER OutputDir
    Directory to save security reports (default: security-reports)

.PARAMETER ImageTag
    Docker image tag to scan for container security (default: cryptoutil:latest)

.PARAMETER SkipDocker
    Skip Docker-based scans (trivy, docker scout)

.PARAMETER Help
    Show this help message

.EXAMPLE
    .\scripts\security-scan.ps1
    Run all security scans with default settings

.EXAMPLE
    .\scripts\security-scan.ps1 -StaticOnly -OutputDir "reports"
    Run only static analysis tools, save reports to "reports" directory

.EXAMPLE
    .\scripts\security-scan.ps1 -ContainerOnly -ImageTag "cryptoutil:dev"
    Run only container security scans on specific image
#>

param(
    [switch]$All,
    [switch]$StaticOnly,
    [switch]$VulnOnly,
    [switch]$ContainerOnly,
    [string]$OutputDir = "security-reports",
    [string]$ImageTag = "cryptoutil:latest",
    [switch]$SkipDocker,
    [switch]$Help
)

# Show help and exit
if ($Help) {
    Get-Help $MyInvocation.MyCommand.Path -Detailed
    exit 0
}

# Determine what scans to run
if (!$StaticOnly -and !$VulnOnly -and !$ContainerOnly) {
    $All = $true
}

# Color output functions
function Write-Header($message) {
    Write-Host "`n🛡️  $message" -ForegroundColor Cyan
    Write-Host ("=" * ($message.Length + 4)) -ForegroundColor Cyan
}

function Write-Success($message) {
    Write-Host "✅ $message" -ForegroundColor Green
}

function Write-Warning($message) {
    Write-Host "⚠️  $message" -ForegroundColor Yellow
}

function Write-Error($message) {
    Write-Host "❌ $message" -ForegroundColor Red
}

function Write-Info($message) {
    Write-Host "ℹ️  $message" -ForegroundColor Blue
}

# Check prerequisites
function Test-Prerequisites {
    Write-Header "Checking Prerequisites"
    
    $missing = @()
    
    # Check Go
    if (!(Get-Command go -ErrorAction SilentlyContinue)) {
        $missing += "Go"
    }
    
    # Check Docker (if not skipping Docker scans)
    if (!$SkipDocker -and !(Get-Command docker -ErrorAction SilentlyContinue)) {
        $missing += "Docker"
    }
    
    if ($missing.Count -gt 0) {
        Write-Error "Missing required tools: $($missing -join ', ')"
        exit 1
    }
    
    Write-Success "All prerequisites available"
}

# Create output directory
function Initialize-OutputDirectory {
    Write-Info "Creating output directory: $OutputDir"
    
    if (Test-Path $OutputDir) {
        Remove-Item $OutputDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
    
    Write-Success "Output directory ready: $OutputDir"
}

# Install Go security tools
function Install-GoTools {
    Write-Header "Installing Go Security Tools"
    
    $tools = @(
        @{Name="staticcheck"; Package="honnef.co/go/tools/cmd/staticcheck@latest"}
        @{Name="govulncheck"; Package="golang.org/x/vuln/cmd/govulncheck@latest"}
        @{Name="golangci-lint"; Package="github.com/golangci/golangci-lint/cmd/golangci-lint@latest"}
    )
    
    foreach ($tool in $tools) {
        Write-Info "Installing $($tool.Name)..."
        go install $tool.Package
        if ($LASTEXITCODE -eq 0) {
            Write-Success "$($tool.Name) installed successfully"
        } else {
            Write-Warning "Failed to install $($tool.Name)"
        }
    }
}

# Run static analysis
function Invoke-StaticAnalysis {
    if (!$All -and !$StaticOnly) { return }
    
    Write-Header "Static Code Analysis"
    
    # Staticcheck
    Write-Info "Running Staticcheck..."
    staticcheck -f sarif ./... > "$OutputDir/staticcheck.sarif" 2>$null
    staticcheck ./...
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Staticcheck completed - no issues found"
    } else {
        Write-Warning "Staticcheck found potential issues - check output above"
    }
    
    # golangci-lint
    Write-Info "Running golangci-lint..."
    golangci-lint run --timeout=10m --config=.golangci.yml --out-format=sarif > "$OutputDir/golangci-lint.sarif"
    golangci-lint run --timeout=10m --config=.golangci.yml
    if ($LASTEXITCODE -eq 0) {
        Write-Success "golangci-lint completed - no issues found"
    } else {
        Write-Warning "golangci-lint found potential issues - check output above"
    }
}

# Run vulnerability scans
function Invoke-VulnerabilityScans {
    if (!$All -and !$VulnOnly) { return }
    
    Write-Header "Vulnerability Scanning"
    
    # govulncheck
    Write-Info "Running Go vulnerability check..."
    govulncheck ./... > "$OutputDir/govulncheck.txt" 2>&1
    govulncheck ./...
    if ($LASTEXITCODE -eq 0) {
        Write-Success "No known Go vulnerabilities found"
    } else {
        Write-Warning "Go vulnerabilities detected - check output above"
    }
    
    # Trivy file system scan (if not skipping Docker)
    if (!$SkipDocker) {
        Write-Info "Running Trivy file system scan..."
        docker run --rm -v "${PWD}:/workspace" aquasec/trivy:latest fs --format sarif --output /workspace/$OutputDir/trivy-fs.sarif /workspace
        docker run --rm -v "${PWD}:/workspace" aquasec/trivy:latest fs /workspace
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Trivy file system scan completed"
        } else {
            Write-Warning "Trivy found potential vulnerabilities - check output above"
        }
    }
}

# Run container security scans
function Invoke-ContainerScans {
    if (!$All -and !$ContainerOnly) { return }
    if ($SkipDocker) {
        Write-Warning "Skipping container scans (Docker disabled)"
        return
    }
    
    Write-Header "Container Security Scanning"
    
    # Check if image exists
    Write-Info "Checking for Docker image: $ImageTag"
    docker images $ImageTag --format "table" | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Warning "Docker image '$ImageTag' not found. Building..."
        docker build -t $ImageTag -f deployments/Dockerfile .
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Failed to build Docker image"
            return
        }
        Write-Success "Docker image built successfully"
    }
    
    # Trivy image scan
    Write-Info "Running Trivy container image scan..."
    docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v "${PWD}/${OutputDir}:/output" aquasec/trivy:latest image --format sarif --output /output/trivy-image.sarif $ImageTag
    docker run --rm -v /var/run/docker.sock:/var/run/docker.sock aquasec/trivy:latest image $ImageTag
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Trivy image scan completed"
    } else {
        Write-Warning "Trivy found container vulnerabilities - check output above"
    }
    
    # Docker Scout (if available)
    Write-Info "Running Docker Scout scans..."
    
    # Quick overview
    docker scout quickview $ImageTag > "$OutputDir/docker-scout-quickview.txt" 2>&1
    docker scout quickview $ImageTag
    
    # CVE analysis
    docker scout cves --format sarif --output "$OutputDir/docker-scout-cves.sarif" $ImageTag 2>$null
    docker scout cves $ImageTag
    
    # Recommendations
    docker scout recommendations $ImageTag > "$OutputDir/docker-scout-recommendations.txt" 2>&1
    docker scout recommendations $ImageTag
    
    Write-Success "Docker Scout scans completed"
}

# Generate summary report
function New-SecurityReport {
    Write-Header "Generating Security Summary Report"
    
    $reportPath = "$OutputDir\security-summary.md"
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss UTC"
    
    $reportContent = @()
    $reportContent += "# Security Scan Results"
    $reportContent += ""
    $reportContent += "**Scan Date:** $timestamp"
    $reportContent += "**Target:** cryptoutil project"
    $reportContent += "**Output Directory:** $OutputDir"
    $reportContent += ""
    $reportContent += "## Scan Coverage"
    $reportContent += ""

    if ($All -or $StaticOnly) {
        $reportContent += "### Static Analysis"
        $reportContent += "* Staticcheck: Go static analysis and lint checking"
        $reportContent += "* golangci-lint: Comprehensive Go linting with multiple analyzers"
        $reportContent += ""
    }

    if ($All -or $VulnOnly) {
        $reportContent += "### Vulnerability Scanning"
        $reportContent += "* govulncheck: Official Go vulnerability database scanning"
        $reportContent += "* Trivy FS: File system and dependency vulnerability scanning"
        $reportContent += ""
    }

    if (($All -or $ContainerOnly) -and !$SkipDocker) {
        $reportContent += "### Container Security"
        $reportContent += "* Trivy Image: Container image vulnerability scanning"
        $reportContent += "* Docker Scout: Advanced container security analysis and recommendations"
        $reportContent += ""
    }

    $reportContent += "## Report Files Generated"
    $reportContent += ""

    # List generated report files
    if (Test-Path $OutputDir) {
        Get-ChildItem $OutputDir -File | ForEach-Object {
            $reportContent += "* $($_.Name)"
        }
    }

    $reportContent += ""
    $reportContent += "## Next Steps"
    $reportContent += ""
    $reportContent += "1. Review detailed reports in the $OutputDir directory"
    $reportContent += "2. Address HIGH and CRITICAL findings immediately"
    $reportContent += "3. Update dependencies for known vulnerabilities"
    $reportContent += "4. Consider security recommendations from Docker Scout"
    $reportContent += "5. Run security scans regularly as part of development workflow"
    $reportContent += ""
    $reportContent += "## Report Locations"
    $reportContent += ""
    $reportContent += "All security reports are saved in: **$OutputDir**"

    $reportContent | Out-File -FilePath $reportPath -Encoding UTF8
    Write-Success "Security summary report generated: $reportPath"
}

# Main execution
function Main {
    Write-Header "Cryptoutil Security Scanner"
    Write-Info "Starting comprehensive security analysis..."
    
    Test-Prerequisites
    Initialize-OutputDirectory
    Install-GoTools
    
    Invoke-StaticAnalysis
    Invoke-VulnerabilityScans  
    Invoke-ContainerScans
    
    New-SecurityReport
    
    Write-Header "Security Scan Complete"
    Write-Success "All security scans completed successfully!"
    Write-Info "Reports saved to: $OutputDir"
    Write-Info "Review the security-summary.md file for an overview of all findings"
}

# Execute main function
Main
