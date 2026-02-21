// Copyright (c) 2025 Justin Cranford

// Package config provides coverage tests for jose-ja configuration parsing.
package config

import (
"testing"

"github.com/spf13/pflag"
"github.com/stretchr/testify/require"
)

// TestParseWithFlagSet_MissingSubcommand tests ParseWithFlagSet when no subcommand is provided.
// Covers the error return after template ParseWithFlagSet fails (missing subcommand).
// NOTE: Must NOT use t.Parallel() - viper global state modified.
func TestParseWithFlagSet_MissingSubcommand(t *testing.T) {
fs := pflag.NewFlagSet("test-missing-subcmd", pflag.ContinueOnError)

_, err := ParseWithFlagSet(fs, []string{}, false)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to parse template settings")
}

// TestParseWithFlagSet_ValidationError tests ParseWithFlagSet when validation fails.
// Covers the error return after validateJoseJASettings fails (max-materials=0 < minimum 1).
// NOTE: Must NOT use t.Parallel() - viper global state modified.
func TestParseWithFlagSet_ValidationError(t *testing.T) {
fs := pflag.NewFlagSet("test-validation-err", pflag.ContinueOnError)

_, err := ParseWithFlagSet(fs, []string{"start", "--dev", "--max-materials=0"}, false)
require.Error(t, err)
require.Contains(t, err.Error(), "jose-ja settings validation failed")
}

// TestParse_GlobalFlagSet tests the Parse function using pflag.CommandLine.
// NOTE: Must NOT use t.Parallel() - modifies global pflag.CommandLine state.
// Must be called at most ONCE per test binary execution.
func TestParse_GlobalFlagSet(t *testing.T) {
settings, err := Parse([]string{"start", "--dev"}, false)
require.NoError(t, err)
require.NotNil(t, settings)
}
