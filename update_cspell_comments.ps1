param(
    [int]$MaxWords = 0  # 0 means process all words
)

$mappings = Get-Content "word_dictionaries.json" | ConvertFrom-Json
$content = Get-Content ".vscode/cspell.json" -Raw

# Remove existing comments from words (lines that have // after the word)
$content = $content -replace '(\s*"[^"]*",?)\s*//.*', '$1'

# Get all words to process
$allWords = $mappings.PSObject.Properties.Name

# Limit words if MaxWords parameter is specified and > 0
if ($MaxWords -gt 0 -and $allWords.Count -gt $MaxWords) {
    $allWords = $allWords | Select-Object -First $MaxWords
    Write-Host "Processing first $MaxWords words out of $($mappings.PSObject.Properties.Name.Count) total words"
} else {
    Write-Host "Processing all $($allWords.Count) words"
}

# Add new comments based on dictionary coverage
foreach ($word in $allWords) {
    $dictionaries = $mappings.$word
    if ($dictionaries -and $dictionaries.Trim()) {
        # Word is covered by dictionaries - add lowercase comment
        $comment = "// $dictionaries"
    } else {
        # Word is NOT covered by any dictionary - add UPPERCASE comment
        $comment = "// NOT COVERED BY ANY DICTIONARY"
    }

    # Replace the word line with comment + word
    # Match the exact line with the word (assuming 4 spaces indentation)
    $pattern = '(?m)^(\s*)"(' + [regex]::Escape($word) + ')",?\s*$'
    $replacement = '$1' + $comment + "`n" + '$1"' + $word + '",'
    $content = $content -replace $pattern, $replacement
}

# Write the file with UTF-8 encoding without BOM
$utf8NoBom = New-Object System.Text.UTF8Encoding $false
[System.IO.File]::WriteAllText(".vscode/cspell.json", $content, $utf8NoBom)
