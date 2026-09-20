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

$lines = @(Get-Content $justfile)
$inRecipe = $false
$recipeCommands = @()

foreach ($line in $lines) {
    # Check if this is the start of our recipe
    if ($line -match "^@$([Regex]::Escape($Recipe)):\s*$") {
        $inRecipe = $true
        continue
    }
    
    # Check if we've hit a new recipe (starts with @)
    if ($inRecipe -and $line -match "^@" -and $line -ne "") {
        break
    }
    
    # If we're in the recipe, collect the command (remove leading whitespace)
    if ($inRecipe) {
        if ($line -match "^\s+(.+)$") {
            $cmd = $matches[1]
            if ($cmd -and -not $cmd.StartsWith('#')) {
                $recipeCommands += $cmd
            }
        }
    }
}

if ($recipeCommands.Count -eq 0) {
    Write-Host "Recipe '$Recipe' not found in Justfile"
    exit 1
}

foreach ($cmd in $recipeCommands) {
    Write-Host "Running: $cmd"
    Invoke-Expression $cmd
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Command failed: $cmd"
        exit $LASTEXITCODE
    }
}
