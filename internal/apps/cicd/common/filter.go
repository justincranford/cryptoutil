// Copyright (c) 2025 Justin Cranford

package common

import (
	"regexp"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// FilterFilesForCommand filters files by applying the self-exclusion pattern
// and generated file exclusion patterns for the given command.
// This is used after directory-level exclusions have already been applied by ListAllFiles.
func FilterFilesForCommand(files []string, commandName string) []string {
	if len(files) == 0 {
		return files
	}

	// Get self-exclusion pattern for this command.
	selfExclusionPattern, hasSelfExclusion := cryptoutilSharedMagic.CICDSelfExclusionPatterns[commandName]

	// Compile patterns.
	var selfExclusionRegex *regexp.Regexp
	if hasSelfExclusion {
		selfExclusionRegex = regexp.MustCompile(selfExclusionPattern)
	}

	generatedPatterns := make([]*regexp.Regexp, 0, len(cryptoutilSharedMagic.GeneratedFileExcludePatterns))
	for _, pattern := range cryptoutilSharedMagic.GeneratedFileExcludePatterns {
		generatedPatterns = append(generatedPatterns, regexp.MustCompile(pattern))
	}

	var filtered []string

	for _, file := range files {
		// Check self-exclusion.
		if selfExclusionRegex != nil && selfExclusionRegex.MatchString(file) {
			continue
		}

		// Check generated file patterns.
		excluded := false

		for _, regex := range generatedPatterns {
			if regex.MatchString(file) {
				excluded = true

				break
			}
		}

		if !excluded {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

// FlattenFileMap flattens a map of files by extension into a single slice.
func FlattenFileMap(filesByExtension map[string][]string) []string {
	var allFiles []string

	for _, files := range filesByExtension {
		allFiles = append(allFiles, files...)
	}

	return allFiles
}

// GetGoFiles extracts Go files from the file map and applies command-specific filtering.
func GetGoFiles(filesByExtension map[string][]string, commandName string) []string {
	goFiles := filesByExtension["go"]

	return FilterFilesForCommand(goFiles, commandName)
}
