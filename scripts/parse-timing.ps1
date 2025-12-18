# Parse test timing from JSON output
# Usage: Get-Content test-output/baseline-timing-002.txt | .\scripts\parse-timing.ps1

param(
    [Parameter(ValueFromPipeline = $true)]
    [string]$InputLine
)

begin {
    $packages = @{}
    $slowPackages = @()
    $threshold = 12.0  # 12 second threshold
}

process {
    if ($InputLine -match '^\{') {
        try {
            $json = $InputLine | ConvertFrom-Json
            if ($json.Action -eq 'pass' -and $json.Package -and $json.Elapsed) {
                # Only process top-level package results (not individual test results)
                if (-not $json.Test) {
                    $pkg = $json.Package
                    $elapsed = [double]$json.Elapsed

                    $packages[$pkg] = $elapsed

                    if ($elapsed -gt $threshold) {
                        $slowPackages += [PSCustomObject]@{
                            Package = $pkg
                            Elapsed = $elapsed
                        }
                    }
                }
            }
        }
        catch {
            # Ignore non-JSON lines
        }
    }
}

end {
    Write-Host "`n=== Test Timing Baseline Summary ===`n" -ForegroundColor Cyan

    # Total packages
    Write-Host "Total packages tested: $($packages.Count)" -ForegroundColor Green

    # Packages exceeding threshold
    Write-Host "Packages exceeding ${threshold}s: $($slowPackages.Count)" -ForegroundColor $(if ($slowPackages.Count -gt 0) { 'Red' } else { 'Green' })

    if ($slowPackages.Count -gt 0) {
        Write-Host "`n=== Slow Packages (>${threshold}s) ===`n" -ForegroundColor Yellow

        $slowPackages | Sort-Object -Property Elapsed -Descending | ForEach-Object {
            $pct = ($_.Elapsed / $threshold) * 100
            Write-Host ("{0,-80} {1,8:F2}s ({2,5:F1}% of threshold)" -f $_.Package, $_.Elapsed, $pct) -ForegroundColor Red
        }

        Write-Host "`n=== Top 10 Slowest Packages ===`n" -ForegroundColor Yellow
        $top10 = $slowPackages | Sort-Object -Property Elapsed -Descending | Select-Object -First 10
        $top10 | ForEach-Object {
            Write-Host ("{0,-80} {1,8:F2}s" -f $_.Package, $_.Elapsed) -ForegroundColor Yellow
        }
    }

    Write-Host "`n=== All Packages Timing ===`n" -ForegroundColor Cyan
    $packages.GetEnumerator() | Sort-Object -Property Value -Descending | ForEach-Object {
        $color = if ($_.Value -gt $threshold) { 'Red' } elseif ($_.Value -gt 6) { 'Yellow' } else { 'Green' }
        Write-Host ("{0,-80} {1,8:F2}s" -f $_.Key, $_.Value) -ForegroundColor $color
    }

    Write-Host "`n=== Statistics ===`n" -ForegroundColor Cyan
    $totalTime = ($packages.Values | Measure-Object -Sum).Sum
    $avgTime = ($packages.Values | Measure-Object -Average).Average
    $maxTime = ($packages.Values | Measure-Object -Maximum).Maximum
    $minTime = ($packages.Values | Measure-Object -Minimum).Minimum

    Write-Host "Total test time:   $($totalTime.ToString('F2'))s" -ForegroundColor White
    Write-Host "Average per pkg:   $($avgTime.ToString('F2'))s" -ForegroundColor White
    Write-Host "Slowest package:   $($maxTime.ToString('F2'))s" -ForegroundColor White
    Write-Host "Fastest package:   $($minTime.ToString('F2'))s" -ForegroundColor White

    # Optimization targets
    if ($slowPackages.Count -gt 0) {
        Write-Host "`n=== Optimization Targets ===`n" -ForegroundColor Magenta
        Write-Host "Priority order (slowest first):`n" -ForegroundColor White

        $i = 1
        $slowPackages | Sort-Object -Property Elapsed -Descending | ForEach-Object {
            $savings = $_.Elapsed - $threshold
            Write-Host ("P1.{0,2}: {1,-70} (save ~{2:F2}s)" -f ($i + 1), $_.Package, $savings) -ForegroundColor Magenta
            $i++
        }
    }
}
