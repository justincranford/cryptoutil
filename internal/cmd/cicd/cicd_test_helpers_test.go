// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"fmt"

	"cryptoutil/internal/cmd/cicd/common"
	goEnforceTestPatternsPkg "cryptoutil/internal/cmd/cicd/go_enforce_test_patterns"
)

// goEnforceTestPatterns is a test helper that delegates to the real Enforce function in the
// go_enforce_test_patterns package. Some tests call this helper by name; providing it ensures
// those tests compile and remain stable.
func goEnforceTestPatterns(logger *common.Logger, allFiles []string) error {
	if err := goEnforceTestPatternsPkg.Enforce(logger, allFiles); err != nil {
		return fmt.Errorf("go-enforce-test-patterns failed: %w", err)
	}

	return nil
}
