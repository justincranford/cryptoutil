// Copyright (c) 2025 Justin Cranford
//
//

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	testify "github.com/stretchr/testify/require"
)

func TestParseConfigFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		params       []string
		defaultValue string
		expected     string
	}{
		{name: "empty params returns default", params: []string{}, defaultValue: "default.yml", expected: "default.yml"},
		{name: "--config space value", params: []string{"--config", "/path/to/config.yml"}, defaultValue: "default.yml", expected: "/path/to/config.yml"},
		{name: "--config=value", params: []string{"--config=/path/to/config.yml"}, defaultValue: "default.yml", expected: "/path/to/config.yml"},
		{name: "-c space value", params: []string{"-c", "/path/to/config.yml"}, defaultValue: "default.yml", expected: "/path/to/config.yml"},
		{name: "-c=value", params: []string{"-c=/path/to/config.yml"}, defaultValue: "default.yml", expected: "/path/to/config.yml"},
		{name: "--config at end without value", params: []string{"--config"}, defaultValue: "default.yml", expected: "default.yml"},
		{name: "no matching flag returns default", params: []string{"--other", "value"}, defaultValue: "default.yml", expected: "default.yml"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := parseConfigFlag(tc.params, tc.defaultValue)
			testify.Equal(t, tc.expected, result)
		})
	}
}

func TestParseDSNFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		params   []string
		expected string
	}{
		{name: "empty params returns empty", params: []string{}, expected: ""},
		{name: "-u space value", params: []string{"-u", "postgres://localhost:5432/db"}, expected: "postgres://localhost:5432/db"},
		{name: "-u=value", params: []string{"-u=postgres://localhost:5432/db"}, expected: "postgres://localhost:5432/db"},
		{name: "--database-url space value", params: []string{"--database-url", "postgres://localhost:5432/db"}, expected: "postgres://localhost:5432/db"},
		{name: "--database-url=value", params: []string{"--database-url=postgres://localhost:5432/db"}, expected: "postgres://localhost:5432/db"},
		{name: "-u at end without value", params: []string{"-u"}, expected: ""},
		{name: "no matching flag returns empty", params: []string{"--other", "value"}, expected: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := parseDSNFlag(tc.params)
			testify.Equal(t, tc.expected, result)
		})
	}
}

func TestResolveDSNValue(t *testing.T) {
	t.Parallel()

	t.Run("plain value returned as-is", func(t *testing.T) {
		t.Parallel()

		result := resolveDSNValue("postgres://localhost:5432/db")
		testify.Equal(t, "postgres://localhost:5432/db", result)
	})

	t.Run("file URL reads from file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dsnFile := filepath.Join(tmpDir, "dsn.txt")
		testify.NoError(t, os.WriteFile(dsnFile, []byte("postgres://filehost:5432/filedb\n"), 0o600))

		result := resolveDSNValue("file://" + dsnFile)
		testify.Equal(t, "postgres://filehost:5432/filedb", result)
	})

	t.Run("file URL with nonexistent file returns empty", func(t *testing.T) {
		t.Parallel()

		result := resolveDSNValue("file:///nonexistent/path/dsn.txt")
		testify.Equal(t, "", result)
	})

	t.Run("empty value returned as-is", func(t *testing.T) {
		t.Parallel()

		result := resolveDSNValue("")
		testify.Equal(t, "", result)
	})
}
