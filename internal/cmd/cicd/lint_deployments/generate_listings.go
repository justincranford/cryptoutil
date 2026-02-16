package lint_deployments

import (
	json "encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)
const (
	fileTypeSecret  = "secret"
	fileTypeCompose = "compose"
	fileTypeDocker  = "docker"
	fileTypeConfig  = "config"
	fileTypeSQL     = "sql"
	fileTypeDoc     = "doc"
	fileTypeJSON    = "json"
	fileTypeOther   = "other"

	filePermissions = 0o600
)
// FileEntry represents metadata about a single file in the listing.
type FileEntry struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

// DirectoryListing maps relative file paths to their metadata.
type DirectoryListing map[string]FileEntry

// classifyFileType determines the type of a file based on its name and extension.
func classifyFileType(filename string) string {
	lower := strings.ToLower(filename)

	switch {
	case strings.HasSuffix(lower, ".secret") || strings.HasSuffix(lower, ".never"):
		return fileTypeSecret
	case strings.HasPrefix(lower, "compose") && (strings.HasSuffix(lower, ".yml") || strings.HasSuffix(lower, ".yaml")):
		return fileTypeCompose
	case strings.HasPrefix(lower, "dockerfile") || lower == "dockerfile":
		return fileTypeDocker
	case strings.HasSuffix(lower, ".yml") || strings.HasSuffix(lower, ".yaml"):
		return fileTypeConfig
	case strings.HasSuffix(lower, ".sql"):
		return fileTypeSQL
	case strings.HasSuffix(lower, ".md"):
		return fileTypeDoc
	case strings.HasSuffix(lower, ".json"):
		return fileTypeJSON
	default:
		return fileTypeOther
	}
}

// classifyFileStatus determines if a file is required or optional based on its path and deployment type.
func classifyFileStatus(relPath string) string {
	lower := strings.ToLower(relPath)

	switch {
	case strings.Contains(lower, "compose.yml"):
		return RequiredFileStatus
	case strings.HasSuffix(lower, ".secret"):
		return RequiredFileStatus
	case strings.HasSuffix(lower, "dockerfile"):
		return RequiredFileStatus
	case strings.HasSuffix(lower, ".md"):
		return OptionalFileStatus
	case strings.HasSuffix(lower, ".never"):
		return RequiredFileStatus
	default:
		return OptionalFileStatus
	}
}

// GenerateDirectoryListing walks a directory and creates a listing of all files with metadata.
func GenerateDirectoryListing(baseDir string) (DirectoryListing, error) {
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", baseDir)
	}

	listing := make(DirectoryListing)

	err := filepath.WalkDir(baseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories themselves, only list files.
		if d.IsDir() {
			return nil
		}

		// Skip generated listing files to avoid self-reference.
		if strings.HasSuffix(path, "_all_files.json") {
			return nil
		}

		// Get relative path from base directory.
		relPath, relErr := filepath.Rel(baseDir, path)
		if relErr != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, relErr)
		}

		// Normalize path separators to forward slashes.
		relPath = filepath.ToSlash(relPath)

		listing[relPath] = FileEntry{
			Type:   classifyFileType(d.Name()),
			Status: classifyFileStatus(relPath),
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", baseDir, err)
	}

	return listing, nil
}

// GenerateDeploymentsListing generates a JSON listing of all files in the deployments directory.
func GenerateDeploymentsListing(deploymentsDir string) ([]byte, error) {
	listing, err := GenerateDirectoryListing(deploymentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate deployments listing: %w", err)
	}

	return marshalListing(listing)
}

// GenerateConfigsListing generates a JSON listing of all files in the configs directory.
func GenerateConfigsListing(configsDir string) ([]byte, error) {
	listing, err := GenerateDirectoryListing(configsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate configs listing: %w", err)
	}

	return marshalListing(listing)
}

// marshalListing converts a DirectoryListing to sorted, indented JSON bytes.
func marshalListing(listing DirectoryListing) ([]byte, error) {
	// Sort keys for deterministic output.
	keys := make([]string, 0, len(listing))
	for k := range listing {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	// Build ordered map for JSON marshal.
	ordered := make([]struct {
		Key   string
		Value FileEntry
	}, 0, len(keys))

	for _, k := range keys {
		ordered = append(ordered, struct {
			Key   string
			Value FileEntry
		}{Key: k, Value: listing[k]})
	}

	// Use a custom marshal to maintain key order.
	result := "{\n"

	for i, item := range ordered {
		entryJSON, err := json.Marshal(item.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal entry for %s: %w", item.Key, err)
		}

		result += fmt.Sprintf("  %q: %s", item.Key, string(entryJSON))
		if i < len(ordered)-1 {
			result += ","
		}

		result += "\n"
	}

	result += "}\n"

	return []byte(result), nil
}

// WriteListingFile generates a listing and writes it to the specified output path.
func WriteListingFile(baseDir string, outputPath string) error {
	listing, err := GenerateDirectoryListing(baseDir)
	if err != nil {
		return err
	}

	data, marshalErr := marshalListing(listing)
	if marshalErr != nil {
		return marshalErr
	}

	if writeErr := os.WriteFile(outputPath, data, filePermissions); writeErr != nil {
		return fmt.Errorf("failed to write listing file %s: %w", outputPath, writeErr)
	}

	return nil
}
