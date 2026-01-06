// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package cipher

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInternalCipher_HelpFlag(t *testing.T) {
	t.Parallel()

	args := []string{"--help"}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := internalCipher(args, nil, stdout, stderr)
	require.Equal(t, 0, exitCode, "help should succeed")
}

func TestInternalCipher_VersionFlag(t *testing.T) {
	t.Parallel()

	args := []string{"--version"}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := internalCipher(args, nil, stdout, stderr)
	require.Equal(t, 0, exitCode, "version should succeed")
}

func TestInternalCipher_UnknownService(t *testing.T) {
	t.Parallel()

	args := []string{"unknown"}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := internalCipher(args, nil, stdout, stderr)
	require.Equal(t, 1, exitCode, "unknown service should fail")
}
