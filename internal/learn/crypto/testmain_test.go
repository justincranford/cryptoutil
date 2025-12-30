// Copyright (c) 2025 Justin Cranford
//

package crypto

import (
	"os"
	"testing"
)

var testSetupComplete bool

func TestMain(m *testing.M) {
	// Setup: Crypto package tests are lightweight (no heavyweight dependencies).
	// TestMain provides framework for future setup if needed (e.g., benchmark data).
	testSetupComplete = true

	// Run all tests.
	exitCode := m.Run()

	// Cleanup (none currently needed).
	testSetupComplete = false

	os.Exit(exitCode)
}
