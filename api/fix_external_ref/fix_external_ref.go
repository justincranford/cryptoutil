//go:build fixexternalref

package main

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"os"
	"strings"
)

func main() {
	files := []string{
		"model/openapi_gen_model.go",
		"server/openapi_gen_server.go",
		"client/openapi_gen_client.go",
	}

	for _, file := range files {
		if err := fixExternalRef(file); err != nil {
			fmt.Fprintf(os.Stderr, "Error fixing %s: %v\n", file, err)
			os.Exit(1)
		}
	}
}

// fixExternalRef replaces ExternalRef0 with externalRef0 in generated files.
// This fixes an oapi-codegen casing bug where import aliases don't match usage.
func fixExternalRef(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	// Replace ExternalRef0 with externalRef0
	fixed := strings.ReplaceAll(string(content), "ExternalRef0", "externalRef0")

	// Write back to file
	return os.WriteFile(filename, []byte(fixed), cryptoutilSharedMagic.CacheFilePermissions)
}
