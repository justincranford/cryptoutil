param(
    [int]$MaxWords = 0  # 0 means process all words
)

$mappings = Get-Content "word_dictionaries.json" | ConvertFrom-Json
$content = Get-Content ".vscode/cspell.json" -Raw

# Get all words to process
$allWords = $mappings.PSObject.Properties.Name

# Filter out words that already have comments in the file
$wordsToProcess = @()
foreach ($word in $allWords) {
    # Check if this word already has a comment on the same line
    $wordPattern = [regex]::Escape('"' + $word + '"') + '\s*,\s*//'
    if ($content -notmatch $wordPattern) {
        # Word doesn't have a comment, so we should process it
        $wordsToProcess += $word
    }
}

# Limit words if MaxWords parameter is specified and > 0
if ($MaxWords -gt 0 -and $wordsToProcess.Count -gt $MaxWords) {
    $wordsToProcess = $wordsToProcess | Select-Object -First $MaxWords
    Write-Host "Processing first $MaxWords words out of $($wordsToProcess.Count) total words"
} else {
    Write-Host "Processing $($wordsToProcess.Count) words"
}

# Add new comments based on dictionary coverage
foreach ($word in $wordsToProcess) {
    $dictionaries = $mappings.$word
    if ($dictionaries -and $dictionaries.Trim()) {
        # Word is covered by dictionaries - add lowercase comment
        $comment = "// $dictionaries"
    } else {
        # Word is NOT covered by any dictionary - add UPPERCASE comment
        $comment = "// NOT COVERED BY ANY DICTIONARY"
    }

    # Replace the word line with word + comment on the same line
    # Match the exact line with the word (assuming 4 spaces indentation)
    $pattern = '(?m)^(\s*)"(' + [regex]::Escape($word) + ')",?\s*$'
    $replacement = '$1"' + $word + '", ' + $comment
    $content = $content -replace $pattern, $replacement
}

# Write the file with UTF-8 encoding without BOM
$utf8NoBom = New-Object System.Text.UTF8Encoding $false
[System.IO.File]::WriteAllText(".vscode/cspell.json", $content, $utf8NoBom)
