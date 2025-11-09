param(
    [int]$MaxWords = 0  # 0 means process all words
)

$words = Get-Content "words_list.txt"

# Limit words if MaxWords parameter is specified and > 0
if ($MaxWords -gt 0 -and $words.Count -gt $MaxWords) {
    $words = $words | Select-Object -First $MaxWords
    Write-Host "Limiting to first $MaxWords words out of $($words.Count) total words"
}

$results = @{}
$batchSize = 50  # Process 50 words at a time for better throughput
$maxConcurrentJobs = 8  # Run up to 8 concurrent cspell processes

Write-Host "Processing $($words.Count) words in batches of $batchSize with $maxConcurrentJobs concurrent jobs..."

# Create a cache for dictionary lookups to avoid redundant cspell calls
$dictionaryCache = @{}

for ($i = 0; $i -lt $words.Count; $i += $batchSize) {
    $batchEnd = [math]::Min($i + $batchSize - 1, $words.Count - 1)
    $batch = $words[$i..$batchEnd]
    $batchNum = [math]::Floor($i / $batchSize) + 1
    $totalBatches = [math]::Ceiling($words.Count / $batchSize)

    Write-Host "Processing batch $batchNum of $totalBatches (words $($i+1) to $($batchEnd+1))..."

    # Process words concurrently using jobs
    $jobs = @()
    foreach ($word in $batch) {
        if ($dictionaryCache.ContainsKey($word)) {
            # Use cached result
            $results[$word] = $dictionaryCache[$word]
        } else {
            # Start a new job for cspell lookup
            $job = Start-Job -ScriptBlock {
                param($word, $configPath)
                try {
                    $output = & cspell trace --config $configPath $word 2>$null
                    $dictionaries = ($output -split "`n" |
                        Where-Object { $_ -match "^$word\s+\*\s+(.+)\*\s+" -and $_ -notmatch "\[words\]\*" } |
                        ForEach-Object { ($_.Split()[2] -replace "\*$", "") }) -join ", "
                    return @{ Word = $word; Dictionaries = $dictionaries }
                } catch {
                    return @{ Word = $word; Dictionaries = "" }
                }
            } -ArgumentList $word, (Resolve-Path ".vscode/cspell.json")

            $jobs += $job

            # Limit concurrent jobs
            while ($jobs.Count -ge $maxConcurrentJobs) {
                $completedJobs = $jobs | Where-Object { $_.State -eq "Completed" }
                foreach ($job in $completedJobs) {
                    $result = Receive-Job -Job $job
                    $results[$result.Word] = $result.Dictionaries
                    $dictionaryCache[$result.Word] = $result.Dictionaries
                    Remove-Job -Job $job
                }
                $jobs = $jobs | Where-Object { $_.State -ne "Completed" }

                if ($jobs.Count -ge $maxConcurrentJobs) {
                    Start-Sleep -Milliseconds 100
                }
            }
        }
    }

    # Wait for remaining jobs in this batch to complete
    while ($jobs.Count -gt 0) {
        $completedJobs = $jobs | Where-Object { $_.State -eq "Completed" }
        foreach ($job in $completedJobs) {
            $result = Receive-Job -Job $job
            $results[$result.Word] = $result.Dictionaries
            $dictionaryCache[$result.Word] = $result.Dictionaries
            Remove-Job -Job $job
        }
        $jobs = $jobs | Where-Object { $_.State -ne "Completed" }

        if ($jobs.Count -gt 0) {
            Start-Sleep -Milliseconds 100
        }
    }
}

Write-Host "Saving results to word_dictionaries.json..."
$results | ConvertTo-Json | Out-File "word_dictionaries.json"
Write-Host "Done! Processed $($words.Count) words."
