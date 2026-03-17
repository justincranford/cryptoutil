// Copyright (c) 2025 Justin Cranford

package lint_gotest

import (
	"fmt"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

// Sequential: mutates registeredLinters package-level state.
func TestLint_LinterReturnsError(t *testing.T) {
	original := registeredLinters

	t.Cleanup(func() { registeredLinters = original })

	registeredLinters = []struct {
		name   string
		linter LinterFunc
	}{
		{"test-fail-linter", func(_ *cryptoutilCmdCicdCommon.Logger, _ []string) error {
			return fmt.Errorf("test linter failure")
		}},
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {"example_test.go"},
	}

	err := Lint(logger, filesByExtension)

	require.Error(t, err)
	require.Contains(t, err.Error(), "lint-go-test failed")
}
