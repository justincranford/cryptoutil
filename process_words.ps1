param(
    [int]$MaxWords = 0  # 0 means process all words
)

# Read words from cspell.json words array, filtering out those that already have comments
$content = Get-Content ".vscode/cspell.json" -Raw
$words = @()

# Find the words array and extract words that don't have comments after them
if ($content -match '"words":\s*\[([^\]]*)\]') {
    $wordsSection = $matches[1]
    $lines = $wordsSection -split "`n"

    for ($i = 0; $i -lt $lines.Count; $i++) {
        $line = $lines[$i].Trim()
        if ($line -match '"([^"]+)",?') {
            $word = $matches[1]
            # Check if the next line is a comment (starts with //)
            $nextLine = if ($i + 1 -lt $lines.Count) { $lines[$i + 1].Trim() } else { "" }
            if ($nextLine -notmatch '^//') {
                # Word doesn't have a comment after it
                $words += $word
            }
        }
    }
}

# Limit words if MaxWords parameter is specified and > 0
if ($MaxWords -gt 0 -and $words.Count -gt $MaxWords) {
    $words = $words | Select-Object -First $MaxWords
    Write-Host "Limiting to first $MaxWords words out of $($words.Count) total words"
}

$results = @{}
$batchSize = 20  # Process 20 words at a time

Write-Host "Processing $($words.Count) words in batches of $batchSize..."

for ($i = 0; $i -lt $words.Count; $i += $batchSize) {
    $batchEnd = [math]::Min($i + $batchSize - 1, $words.Count - 1)
    $batch = $words[$i..$batchEnd]
    $batchNum = [math]::Floor($i / $batchSize) + 1
    $totalBatches = [math]::Ceiling($words.Count / $batchSize)

    Write-Host "Processing batch $batchNum of $totalBatches (words $($i+1) to $($batchEnd+1))..."

    foreach ($word in $batch) {
        try {
            $output = cspell trace --config .vscode/cspell.json $word 2>$null
            $dictionaries = ($output -split "`n" |
                Where-Object { $_ -match "^$word\s+\*\s+(.+)\*\s+" -and $_ -notmatch "\[words\]\*" } |
                ForEach-Object { ($_.Split()[2] -replace "\*$", "") }) -join ", "
            $results[$word] = $dictionaries
        } catch {
            Write-Host "Error processing word: $word"
            $results[$word] = ""
        }
    }
}

Write-Host "Saving results to word_dictionaries.json..."
$results | ConvertTo-Json | Out-File "word_dictionaries.json"
Write-Host "Done! Processed $($words.Count) words."
