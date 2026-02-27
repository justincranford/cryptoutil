// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

//go:build !integration

// Package config tests error path coverage for ParseWithFlagSet.
package config

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

// TestParseWithFlagSet_MissingSubcommand covers the template parse failure path.
// Triggers "missing subcommand" error from template config.
func TestParseWithFlagSet_MissingSubcommand(t *testing.T) {
	// Not parallel: viper global state shared between tests.
	fs := pflag.NewFlagSet("test-pki-missing-subcmd", pflag.ContinueOnError)

	_, err := ParseWithFlagSet(fs, []string{}, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse template settings")
}

// TestParseWithFlagSet_ValidationError covers the validateCASettings failure path.
// Triggers CA validation error by providing a non-existent CA config file path.
func TestParseWithFlagSet_ValidationError(t *testing.T) {
	// Not parallel: viper global state shared between tests.
	fs := pflag.NewFlagSet("test-pki-ca-validation", pflag.ContinueOnError)

	_, err := ParseWithFlagSet(fs, []string{"start", "--dev", "--ca-config", "/nonexistent/ca-config.yaml"}, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "pki-ca settings validation failed")
}

// TestParse_WrapperDelegates covers the Parse wrapper function (uses pflag.CommandLine).
// Must run exactly once per test binary (not parallel, not count > 1).
func TestParse_WrapperDelegates(t *testing.T) {
	// Not parallel: uses global pflag.CommandLine - can only be registered once.
	settings, err := Parse([]string{"start", "--dev"}, false)
	require.NoError(t, err)
	require.NotNil(t, settings)
}
