// Copyright (c) 2025 Justin Cranford

package admin_port_exposure

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckComposeFile_InvalidFile(t *testing.T) {
	t.Parallel()

	violations, err := CheckComposeFile("/nonexistent/compose.yml")
	require.Error(t, err, "should error on invalid file")
	require.Nil(t, violations)
}

// TestCheckComposeFile_FileOpenError tests the error path when compose file cannot be opened.
func TestCheckComposeFile_FileOpenError(t *testing.T) {
	t.Parallel()

	violations, err := CheckComposeFile("/nonexistent/path/to/compose.yml")
	require.Error(t, err, "should return error for non-existent file")
	require.Nil(t, violations, "should return nil violations on error")
	require.Contains(t, err.Error(), "failed to open file", "error should indicate file open failure")
}
