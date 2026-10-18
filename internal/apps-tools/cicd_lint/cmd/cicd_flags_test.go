// Copyright (c) 2025-2026 Justin Cranford.
package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCicd_OnlyFlagsDelegatesToApplication(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cicd([]string{testCicd, testQ}, strings.NewReader(""), &stdout, &stderr)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "command validation failed")
}
