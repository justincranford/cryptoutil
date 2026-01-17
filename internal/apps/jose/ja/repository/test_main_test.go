// Copyright (c) 2025 Justin Cranford

package repository_test

import (
	"os"
	"testing"
)

// TestMain initializes shared test fixtures for JOSE-JA repository tests.
func TestMain(m *testing.M) {
	// Run all tests.
	os.Exit(m.Run())
}
