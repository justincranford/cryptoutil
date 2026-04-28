//go:build e2e

// Copyright (c) 2025 Justin Cranford

package e2e

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
