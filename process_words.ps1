param(
    [int]$MaxWords = 0  # 0 means process all words
)

$words = Get-Content "words_list.txt"

# Limit words if MaxWords parameter is specified and > 0
if ($MaxWords -gt 0 -and $words.Count -gt $MaxWords) {
    $words = $words | Select-Object -First $MaxWords
    Write-Host "Limiting to first $MaxWords words out of $(Get-Content "words_list.txt" | Measure-Object | Select-Object -ExpandProperty Count) total words"
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
