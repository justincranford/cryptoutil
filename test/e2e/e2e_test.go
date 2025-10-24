//go:build e2e

package test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// TestE2E runs the complete end-to-end test suite
func TestE2E(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
