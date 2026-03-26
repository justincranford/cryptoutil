// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestGenerateRequestURI_Format(t *testing.T) {
	t.Parallel()

	requestURI, err := GenerateRequestURI()
	require.NoError(t, err, "should generate request_uri without error")
	require.True(t, strings.HasPrefix(requestURI, cryptoutilSharedMagic.RequestURIPrefix), "request_uri should start with URN prefix")
	require.GreaterOrEqual(t, len(requestURI), len(cryptoutilSharedMagic.RequestURIPrefix)+cryptoutilSharedMagic.DefaultCodeChallengeLength, "request_uri should be at least 43 chars long (32 bytes base64url)")
}

func TestGenerateRequestURI_Uniqueness(t *testing.T) {
	t.Parallel()

	const sampleCount = 1000

	seen := make(map[string]bool, sampleCount)

	for range sampleCount {
		requestURI, err := GenerateRequestURI()
		require.NoError(t, err, "should generate request_uri without error")
		require.False(t, seen[requestURI], "request_uri collision detected")
		seen[requestURI] = true
	}
}

func TestGenerateRequestURI_Length(t *testing.T) {
	t.Parallel()

	requestURI, err := GenerateRequestURI()
	require.NoError(t, err, "should generate request_uri without error")

	// 32 bytes base64url encoded = 43 chars + URN prefix
	expectedMinLength := len(cryptoutilSharedMagic.RequestURIPrefix) + cryptoutilSharedMagic.DefaultCodeChallengeLength
	require.GreaterOrEqual(t, len(requestURI), expectedMinLength, "request_uri should be at least 43 chars (32 bytes base64url)")
}

func TestGenerateRequestURI_NoCollisions(t *testing.T) {
	t.Parallel()

	requestURI1, err := GenerateRequestURI()
	require.NoError(t, err, "should generate request_uri without error")

	requestURI2, err := GenerateRequestURI()
	require.NoError(t, err, "should generate request_uri without error")

	require.NotEqual(t, requestURI1, requestURI2, "consecutive request_uri values should be different")
}
