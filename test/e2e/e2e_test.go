//go:build e2e

package test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// TestE2EIntegration runs the complete end-to-end test suite
func TestE2EIntegration(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
