param(
    [int]$MaxWords = 0  # 0 means process all words
)

# Read the cspell.json file
$content = Get-Content ".vscode/cspell.json" -Raw

# Extract all words from the words array that don't have comments
$wordsToProcess = @()
$wordsSection = [regex]::Match($content, '"words":\s*\[([^\]]*)\]').Groups[1].Value
$wordMatches = [regex]::Matches($wordsSection, '"([^"]+)",?')

foreach ($match in $wordMatches) {
    $word = $match.Groups[1].Value
    # Check if this word already has a comment on the same line
    $wordPattern = [regex]::Escape('"' + $word + '"') + '\s*,\s*//'
    if ($content -notmatch $wordPattern) {
        # Word doesn't have a comment, so we should process it
        $wordsToProcess += $word
    }
}

# Load mappings for dictionary information
$mappings = Get-Content "word_dictionaries.json" | ConvertFrom-Json

# Filter to only words that exist in mappings
$wordsToProcess = $wordsToProcess | Where-Object { $mappings.PSObject.Properties.Name -contains $_ }

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
