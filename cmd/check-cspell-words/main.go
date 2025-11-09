package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	// Read cspell.json
	data, err := os.ReadFile(".vscode/cspell.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading cspell.json: %v\n", err)
		os.Exit(1)
	}

	// Extract words manually
	content := string(data)
	lines := strings.Split(content, "\n")

	var words []string

	inWordsSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"words":`) {
			inWordsSection = true

			continue
		}

		if inWordsSection {
			if strings.HasPrefix(line, "]") {
				break
			}
			// Extract word from line like: "word", // comment
			re := regexp.MustCompile(`"([^"]+)"`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				words = append(words, matches[1])
			}
		}
	}

	// Load all dictionary words
	dictWords := make(map[string]bool)
	dictDir := "docs/dictionaries"

	entries, err := os.ReadDir(dictDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading dictionaries: %v\n", err)
		os.Exit(1)
	}

	fileCount := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}

		path := filepath.Join(dictDir, entry.Name())

		file, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", entry.Name(), err)

			continue
		}

		wordCount := 0

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			word := strings.TrimSpace(scanner.Text())
			if word != "" && !strings.HasPrefix(word, "#") {
				dictWords[strings.ToLower(word)] = true
				wordCount++
			}
		}

		file.Close()

		fileCount++

		fmt.Printf("Loaded %d words from %s\n", wordCount, entry.Name())
	}

	fmt.Printf("\nTotal: %d words from %d dictionaries\n\n", len(dictWords), fileCount)

	// Check each word
	covered := 0
	notCovered := 0

	for _, word := range words {
		lower := strings.ToLower(word)
		if dictWords[lower] {
			fmt.Printf("✓ COVERED: %s\n", word)

			covered++
		} else {
			fmt.Printf("✗ NOT COVERED: %s\n", word)

			notCovered++
		}
	}

	fmt.Printf("\n=== SUMMARY ===\n")

	const percentMultiplier = 100

	fmt.Printf("Total words: %d\n", len(words))
	fmt.Printf("Covered: %d (%.1f%%)\n", covered, float64(covered)/float64(len(words))*percentMultiplier)
	fmt.Printf("Not covered: %d (%.1f%%)\n", notCovered, float64(notCovered)/float64(len(words))*percentMultiplier)

	if notCovered == 0 {
		fmt.Println("\nAll words are covered by dictionaries!")
	}
}
