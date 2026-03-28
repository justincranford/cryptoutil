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
		{name: "SuiteProductCount", got: cryptoutilSharedMagic.SuiteProductCount, expected: 5},
		{name: "SuiteServiceCount", got: cryptoutilSharedMagic.SuiteServiceCount, expected: 10},
		{name: "RequiredConfigOverlayCount", got: cryptoutilSharedMagic.RequiredConfigOverlayCount, expected: 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, tc.got, "magic constant %s has unexpected value", tc.name)
		})
	}
}
