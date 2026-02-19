//go:build integration || e2e

// Copyright (c) 2025 Justin Cranford
//
// NOTE: These tests require a PostgreSQL database and are skipped in CI without the integration tag.
//

package client

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanupTestCertificates(t *testing.T) {
	// List PEM files in the current package directory
	files, err := os.ReadDir(".")
	if err != nil {
		t.Logf("Warning: Could not read directory for PEM file cleanup: %v", err)

		return
	}

	var pemFiles []string

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".pem") {
			pemFiles = append(pemFiles, file.Name())
		}
	}

	// List the PEM files found
	if len(pemFiles) > 0 {
		t.Logf("Found PEM files in %s directory:", "internal/client")

		for _, pemFile := range pemFiles {
			t.Logf("  - %s", pemFile)
		}

		// Delete the PEM files
		for _, pemFile := range pemFiles {
			err := os.Remove(pemFile)
			require.NoError(t, err, "Failed to delete PEM file %s", pemFile)
			t.Logf("Successfully deleted PEM file: %s", pemFile)
		}
	} else {
		t.Logf("No PEM files found in %s directory", "internal/client")
	}
}
