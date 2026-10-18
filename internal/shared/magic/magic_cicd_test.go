package magic_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestCICDCmdDirConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{name: "CICDCmdDirCicdLint", got: cryptoutilSharedMagic.CICDCmdDirCicdLint, expected: "cicd-lint"},
		{name: "CICDCmdDirWorkflow", got: cryptoutilSharedMagic.CICDCmdDirWorkflow, expected: "cicd-workflow"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, tc.got, "magic constant %s has unexpected value", tc.name)
		})
	}
}

func TestSuiteCountConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      int
		expected int
	}{
		{name: "SuiteProductCount", got: cryptoutilSharedMagic.SuiteProductCount, expected: 3},
		{name: "SuiteServiceCount", got: cryptoutilSharedMagic.SuiteServiceCount, expected: 7},
		{name: "RequiredConfigOverlayCount", got: cryptoutilSharedMagic.RequiredConfigOverlayCount, expected: 10},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, tc.got, "magic constant %s has unexpected value", tc.name)
		})
	}
}

func TestPBKDF2HashFunction_AcceptsAlgorithmAliasesAndFallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		algorithm     string
		expectedBytes int
	}{
		{name: "sha 256", algorithm: cryptoutilSharedMagic.SHA_256, expectedBytes: 32},
		{name: "sha 256 lowercase", algorithm: "sha256", expectedBytes: 32},
		{name: "sha 256 compact uppercase", algorithm: "SHA256", expectedBytes: 32},
		{name: "sha 512", algorithm: cryptoutilSharedMagic.SHA_512, expectedBytes: 64},
		{name: "sha 512 lowercase", algorithm: "sha512", expectedBytes: 64},
		{name: "sha 512 compact uppercase", algorithm: "SHA512", expectedBytes: 64},
		{name: "sha 384", algorithm: cryptoutilSharedMagic.SHA_384, expectedBytes: 48},
		{name: "sha 384 lowercase", algorithm: "sha384", expectedBytes: 48},
		{name: "sha 384 compact uppercase", algorithm: "SHA384", expectedBytes: 48},
		{name: "fallback", algorithm: "unknown", expectedBytes: 32},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expectedBytes, cryptoutilSharedMagic.PBKDF2HashFunction(tc.algorithm)().Size())
		})
	}
}
