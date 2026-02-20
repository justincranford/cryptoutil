// Copyright (c) 2025 Justin Cranford

// Package common provides shared utilities for lint_compose subpackages.
package common

import (
"path/filepath"
"strings"
)

// LineSeparatorLength defines the length of line separators in output.
const LineSeparatorLength = 60

// FindComposeFiles returns all Docker Compose files from the file map.
func FindComposeFiles(filesByExtension map[string][]string) []string {
var composeFiles []string

// Check yml and yaml files for compose files.
// NOTE: filesByExtension keys are WITHOUT dots (e.g., "yml" not ".yml").
for _, ext := range []string{"yml", "yaml"} {
files, ok := filesByExtension[ext]
if !ok {
continue
}

for _, file := range files {
base := filepath.Base(file)
// Match compose*.yml, docker-compose*.yml patterns.
if strings.HasPrefix(base, "compose") ||
strings.HasPrefix(base, "docker-compose") {
composeFiles = append(composeFiles, file)
}
}
}

return composeFiles
}
