# Remove words from cspell.json that only appear in docs/dictionaries/ (not in actual code)
# cspell:disable
$wordsToRemove = @(
    'lucene', 'aayushatharva', 'arearange', 'areaspline', 'areasplinerange',
    'boopickle', 'borderopacity', 'circlepin', 'columnrange', 'coordsize',
    'crosshairs', 'debasishg', 'deskcolor', 'Ellipsed', 'errorprone',
    'fasterxml', 'focusinfocus', 'focusoutblur', 'frise', 'geant',
    'Highstock', 'Honsi', 'inkscape', 'jdbc', 'jodah',
    'jqfake', 'labelrank', 'lagarto', 'Lightbend', 'maventest',
    'metarank', 'MGTPE', 'MSPOINTER', 'Nullness', 'onglet',
    'pagecheckerboard', 'pageopacity', 'pebbletemplates', 'redisclient', 'scrollbox',
    'showpageshadow', 'simpleflatmapper', 'sodipodi', 'squarepin', 'strokecolor',
    'strokeweight', 'suzaku', 'taintanalysis', 'tcnative', 'Torstein',
    'typetools', 'unbescape', 'undelegate', 'Unsquish', 'xmlresolver'
)

$cspellPath = "$PSScriptRoot\..\.vscode\cspell.json"
$lines = Get-Content $cspellPath

$newLines = @()
foreach ($line in $lines) {
    $keep = $true
    foreach ($word in $wordsToRemove) {
        if ($line -match "^\s*`"$word`",?\s*$") {
            $keep = $false
            Write-Host "Removing: $word"
            break
        }
    }
    if ($keep) {
        $newLines += $line
    }
}

# Write back with UTF-8 without BOM
$utf8NoBom = New-Object System.Text.UTF8Encoding $false
[System.IO.File]::WriteAllLines($cspellPath, $newLines, $utf8NoBom)

Write-Host "`nRemoved $($wordsToRemove.Count) words from cspell.json"
