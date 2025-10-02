#Requires -Version 5.1
<#
.SYNOPSIS
    Run DAST (Dynamic Application Security Testing) scans locally

.DESCRIPTION
    Starts cryptoutil application and runs OWASP ZAP and Nuclei security scans.
    Generates comprehensive security reports for local development testing.

.PARAMETER Config
    Configuration file to use (default: configs/local/config.yml)

.PARAMETER Port
    Port to run the application on (default: 8080)

.PARAMETER TargetUrl
    Target URL for scanning (default: http://localhost:8080)

.PARAMETER SkipZap
    Skip OWASP ZAP scanning

.PARAMETER SkipNuclei
    Skip Nuclei scanning

.PARAMETER OutputDir
    Output directory for reports (default: dast-reports)

.PARAMETER Help
    Show this help message

.EXAMPLE
    .\scripts\dast.ps1
    Run DAST with default settings

.EXAMPLE
    .\scripts\dast.ps1 -Config "configs/test/config.yml" -Port 9090
    Run DAST with custom configuration and port

.EXAMPLE
    .\scripts\dast.ps1 -SkipZap
    Run only Nuclei scan, skip ZAP
#>

param(
    [string]$Config = "configs/local/config.yml",
    [int]$Port = 8080,
    [string]$TargetUrl = "",
    [switch]$SkipZap,
    [switch]$SkipNuclei,
    [string]$OutputDir = "dast-reports",
    [switch]$Help
)

# Show help if requested
if ($Help) {
    Get-Help $PSCommandPath -Full
    exit 0
}

# Set target URL if not provided
if (-not $TargetUrl) {
    $TargetUrl = "http://localhost:$Port"
}

# Colors for output
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Blue = "Cyan"

function Write-Status {
    param([string]$Message, [string]$Color = "White")
    Write-Host "[DAST] $Message" -ForegroundColor $Color
}

function Write-Error-Status {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Red
}

function Write-Success-Status {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Green
}

function Write-Warning-Status {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Yellow
}

# Check prerequisites
Write-Status "Checking prerequisites..." $Blue

# Check if Docker is available
try {
    docker --version | Out-Null
    Write-Status "Docker is available"
} catch {
    Write-Error-Status "Docker is not available. Please install Docker Desktop."
    exit 1
}

# Pull ZAP image if not skipping ZAP
if (-not $SkipZap) {
    Write-Status "Pulling OWASP ZAP Docker image..." $Blue
    docker pull zaproxy/zap-stable:latest
    if ($LASTEXITCODE -ne 0) {
        Write-Error-Status "Failed to pull ZAP Docker image"
        exit 1
    }
}

# Check if Nuclei is available
if (-not $SkipNuclei) {
    try {
        nuclei -version | Out-Null
        Write-Status "Nuclei is available"
    } catch {
        Write-Warning-Status "Nuclei not found. Installing..."
        go install -v github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest
        if ($LASTEXITCODE -ne 0) {
            Write-Error-Status "Failed to install Nuclei"
            exit 1
        }
        # Update nuclei templates
        nuclei -update-templates
    }
}

# Create output directory
Write-Status "Creating output directory: $OutputDir" $Blue
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}

# Build application
Write-Status "Building cryptoutil application..." $Blue
go build -o cryptoutil.exe ./cmd/cryptoutil
if ($LASTEXITCODE -ne 0) {
    Write-Error-Status "Failed to build application"
    exit 1
}

# Start application
Write-Status "Starting cryptoutil on port $Port..." $Blue
$AppProcess = Start-Process -FilePath ".\cryptoutil.exe" -ArgumentList "server", "start", "--dev", "--config", $Config -PassThru -NoNewWindow
$AppPID = $AppProcess.Id

# Wait for application to be ready
Write-Status "Waiting for application to be ready..." $Blue
$MaxAttempts = 30
$Attempt = 0
$Ready = $false

# Health check on HTTP port 9090 (both endpoints)
$HealthUrls = @(
    "http://localhost:9090/readyz",
    "http://localhost:9090/livez"
)

# Skip certificate validation for testing (PowerShell 5.1 compatible)
if (-not ([System.Management.Automation.PSTypeName]'ServerCertificateValidationCallback').Type) {
    $certCallback = @"
        using System;
        using System.Net;
        using System.Net.Security;
        using System.Security.Cryptography.X509Certificates;
        public class ServerCertificateValidationCallback
        {
            public static void Ignore()
            {
                if(ServicePointManager.ServerCertificateValidationCallback == null)
                {
                    ServicePointManager.ServerCertificateValidationCallback +=
                        delegate
                        (
                            Object obj,
                            X509Certificate certificate,
                            X509Chain chain,
                            SslPolicyErrors errors
                        )
                        {
                            return true;
                        };
                }
            }
        }
"@
    Add-Type $certCallback
}
[ServerCertificateValidationCallback]::Ignore()

do {
    Start-Sleep -Seconds 2
    $Attempt++
    foreach ($HealthUrl in $HealthUrls) {
        try {
            $Response = Invoke-WebRequest -Uri $HealthUrl -TimeoutSec 5 -UseBasicParsing
            if ($Response.StatusCode -eq 200) {
                $Ready = $true
                Write-Success-Status "Application is ready on $HealthUrl"
                break
            }
        } catch {
            # Continue to next URL
        }
    }
    if (-not $Ready) {
        Write-Status "Attempt $Attempt/$MaxAttempts - waiting for application..."
    }
} while (-not $Ready -and $Attempt -lt $MaxAttempts)

if (-not $Ready) {
    Write-Error-Status "Application failed to start within timeout"
    if ($AppPID) { Stop-Process -Id $AppPID -Force }
    exit 1
}

# Verify OpenAPI spec is available on HTTPS port 8080
$OpenAPIUrls = @(
    "https://localhost:$Port/ui/swagger/doc.json",
    "$TargetUrl/ui/swagger/doc.json"
)

$OpenAPIDownloaded = $false
foreach ($OpenAPIUrl in $OpenAPIUrls) {
    try {
        $OpenAPIResponse = Invoke-WebRequest -Uri $OpenAPIUrl -UseBasicParsing
        $OpenAPIResponse.Content | Out-File -FilePath "$OutputDir/openapi.json" -Encoding UTF8
        Write-Status "OpenAPI specification downloaded from $OpenAPIUrl"
        $OpenAPIDownloaded = $true
        break
    } catch {
        # Continue to next URL
    }
}

if (-not $OpenAPIDownloaded) {
    Write-Warning-Status "OpenAPI specification not available at any expected endpoint"
}

# Run OWASP ZAP scans
if (-not $SkipZap) {
    Write-Status "Running OWASP ZAP Full Scan..." $Blue
    $ZapFullCmd = @(
        "run", "--rm", "-t"
        "-v", "${PWD}/${OutputDir}:/zap/wrk/:rw"
        "zaproxy/zap-stable:latest"
        "zap-full-scan.py"
        "-t", $TargetUrl
        "-r", "zap-full-report.html"
        "-J", "zap-full-report.json"
        "-m", "10"
        "-T", "60"
        "-z", "-config rules.cookie.ignorelist=JSESSIONID,csrftoken"
    )

    & docker $ZapFullCmd
    if ($LASTEXITCODE -eq 0) {
        Write-Success-Status "ZAP Full Scan completed"
    } else {
        Write-Warning-Status "ZAP Full Scan completed with findings (exit code: $LASTEXITCODE)"
    }

    # Run ZAP API scan if OpenAPI spec is available
    if (Test-Path "$OutputDir/openapi.json") {
        Write-Status "Running OWASP ZAP API Scan..." $Blue
        $ZapApiCmd = @(
            "run", "--rm", "-t"
            "-v", "${PWD}/${OutputDir}:/zap/wrk/:rw"
            "zaproxy/zap-stable:latest"
            "zap-api-scan.py"
            "-t", "$TargetUrl/swagger/openapi.json"
            "-f", "openapi"
            "-r", "zap-api-report.html"
            "-J", "zap-api-report.json"
            "-T", "60"
        )

        & docker $ZapApiCmd
        if ($LASTEXITCODE -eq 0) {
            Write-Success-Status "ZAP API Scan completed"
        } else {
            Write-Warning-Status "ZAP API Scan completed with findings (exit code: $LASTEXITCODE)"
        }
    }
}

# Run Nuclei scan
if (-not $SkipNuclei) {
    Write-Status "Running Nuclei Vulnerability Scan..." $Blue
    $NucleiCmd = @(
        "-target", $TargetUrl
        "-templates", "cves/,vulnerabilities/,security-misconfiguration/,default-logins/,exposed-panels/,takeovers/,technologies/"
        "-json-export", "$OutputDir/nuclei-report.json"
        "-stats"
        "-silent"
    )

    & nuclei $NucleiCmd
    if ($LASTEXITCODE -eq 0) {
        Write-Success-Status "Nuclei scan completed"
    } else {
        Write-Warning-Status "Nuclei scan completed with findings (exit code: $LASTEXITCODE)"
    }
}

# Generate summary report
Write-Status "Generating summary report..." $Blue
$SummaryFile = "$OutputDir/dast-summary.md"
$Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss UTC"

@"
# DAST Security Scan Results

**Scan Date:** $Timestamp
**Target URL:** $TargetUrl
**Configuration:** $Config

## Scan Coverage

"@ | Out-File -FilePath $SummaryFile -Encoding UTF8

if (-not $SkipZap) {
    @"
- ✅ **OWASP ZAP Full Scan:** Comprehensive web application security testing
- ✅ **OWASP ZAP API Scan:** OpenAPI specification-driven API security testing

"@ | Add-Content -Path $SummaryFile -Encoding UTF8
}

if (-not $SkipNuclei) {
    @"
- ✅ **Nuclei Scan:** CVE and vulnerability template-based testing

"@ | Add-Content -Path $SummaryFile -Encoding UTF8
}

@"
## Reports Generated

"@ | Add-Content -Path $SummaryFile -Encoding UTF8

# List generated files
Get-ChildItem -Path $OutputDir -File | ForEach-Object {
    "- $($_.Name)" | Add-Content -Path $SummaryFile -Encoding UTF8
}

@"

## Next Steps

1. Review scan reports for HIGH or CRITICAL findings
2. Open HTML reports in your browser for detailed analysis
3. Address any security vulnerabilities found
4. Consider adding custom ZAP rules for cryptographic endpoints

## Report Locations

"@ | Add-Content -Path $SummaryFile -Encoding UTF8

if (Test-Path "$OutputDir/zap-full-report.html") {
    "- **ZAP Full Report:** $OutputDir/zap-full-report.html" | Add-Content -Path $SummaryFile -Encoding UTF8
}
if (Test-Path "$OutputDir/zap-api-report.html") {
    "- **ZAP API Report:** $OutputDir/zap-api-report.html" | Add-Content -Path $SummaryFile -Encoding UTF8
}
if (Test-Path "$OutputDir/nuclei-report.json") {
    "- **Nuclei Report:** $OutputDir/nuclei-report.json" | Add-Content -Path $SummaryFile -Encoding UTF8
}

# Cleanup
Write-Status "Cleaning up..." $Blue
if ($AppPID) {
    try {
        Stop-Process -Id $AppPID -Force
        Write-Status "Application stopped"
    } catch {
        Write-Warning-Status "Failed to stop application process"
    }
}

# Show results
Write-Success-Status "DAST scanning completed!"
Write-Status "Summary report: $SummaryFile" $Blue
Write-Status "All reports saved to: $OutputDir" $Blue

if (Test-Path "$OutputDir/zap-full-report.html") {
    Write-Status "Open ZAP Full Report: $OutputDir/zap-full-report.html" $Green
}
if (Test-Path "$OutputDir/zap-api-report.html") {
    Write-Status "Open ZAP API Report: $OutputDir/zap-api-report.html" $Green
}

exit 0
