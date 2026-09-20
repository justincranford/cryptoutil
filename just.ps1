# Simple just recipe runner for Windows
param([string]$Recipe)

if (-not $Recipe) {
    Write-Host "Usage: just <recipe>"
    exit 1
}

$justfile = "c:\Dev\cryptoutil\Justfile"
if (-not (Test-Path $justfile)) {
    Write-Host "Justfile not found"
    exit 1
}

$content = Get-Content $justfile -Raw

# Parse recipe - look for @recipe: pattern
$pattern = "^@$($Recipe):\s*`n([\s\S]*?)(?=^@|\Z)"
if ($content -match $pattern) {
    $commands = $matches[1].Trim()
    # Remove leading indentation (typically 4 spaces)
    $commands = $commands -replace '^\s{4}', '' -split "`n" | ForEach-Object {
        $_.TrimStart()
    } | Where-Object { $_ -and -not $_.StartsWith('#') }
    
    foreach ($cmd in $commands) {
        Write-Host "Running: $cmd"
        Invoke-Expression $cmd
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Command failed: $cmd"
            exit $LASTEXITCODE
        }
    }
} else {
    Write-Host "Recipe '$Recipe' not found in Justfile"
    exit 1
}
