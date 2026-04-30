//go:build e2e

// Copyright (c) 2025-2026 Justin Cranford.
package e2e_test

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
