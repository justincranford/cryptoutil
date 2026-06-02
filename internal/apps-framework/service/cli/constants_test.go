// Copyright (c) 2025-2026 Justin Cranford.
//
// SPDX-License-Identifier: AGPL-3.0-only
package cli_test

import (
	"bytes"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkCli "cryptoutil/internal/apps-framework/service/cli"
)

func TestIsHelpRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		args   []string
		expect bool
	}{
		{name: "help_word", args: []string{cryptoutilSharedMagic.CLIHelpCommand}, expect: true},
		{name: "help_long_flag", args: []string{cryptoutilSharedMagic.CLIHelpFlag}, expect: true},
		{name: "help_short_flag", args: []string{"-h"}, expect: true},
		{name: "empty_args", args: []string{}, expect: false},
		{name: "non_help_arg", args: []string{"server"}, expect: false},
		{name: "help_not_first", args: []string{"server", cryptoutilSharedMagic.CLIHelpCommand}, expect: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := cryptoutilAppsFrameworkCli.IsHelpRequest(tc.args)
			require.Equal(t, tc.expect, result)
		})
	}
}

func TestIsHelpRequest_WithClientNotImplementedMessage(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	result := cryptoutilAppsFrameworkCli.IsHelpRequest(
		[]string{"client"},
		cryptoutilAppsFrameworkCli.ClientNotImplementedMessageConfig{
			Stderr:    &stderr,
			ServiceID: cryptoutilSharedMagic.SkeletonTemplateServiceID,
		},
	)

	require.False(t, result)
	require.Contains(t, stderr.String(), "Client subcommand not yet implemented")
	require.Contains(t, stderr.String(), "interacting with the "+cryptoutilSharedMagic.SkeletonTemplateServiceID)
}

func TestIsHelpRequest_WithClientNotImplementedMessage_HelpRequest(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	result := cryptoutilAppsFrameworkCli.IsHelpRequest(
		[]string{cryptoutilSharedMagic.CLIHelpFlag},
		cryptoutilAppsFrameworkCli.ClientNotImplementedMessageConfig{
			Stderr:    &stderr,
			ServiceID: cryptoutilSharedMagic.SkeletonTemplateServiceID,
		},
	)

	require.True(t, result)
	require.Empty(t, stderr.String())
}

func TestIsHelpRequest_WithUsageText_HelpRequest(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	result := cryptoutilAppsFrameworkCli.IsHelpRequest(
		[]string{cryptoutilSharedMagic.CLIHelpFlag},
		cryptoutilAppsFrameworkCli.ClientNotImplementedMessageConfig{
			Stderr:    &stderr,
			ServiceID: cryptoutilSharedMagic.SkeletonTemplateServiceID,
			UsageText: "usage: sm-kms client [flags]",
		},
	)

	require.True(t, result)
	require.Contains(t, stderr.String(), "usage: sm-kms client [flags]")
}

func TestIsHelpRequest_WithUsageTextOnly_HelpRequest(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	result := cryptoutilAppsFrameworkCli.IsHelpRequest(
		[]string{cryptoutilSharedMagic.CLIHelpFlag},
		cryptoutilAppsFrameworkCli.ClientNotImplementedMessageConfig{
			Stderr:    &stderr,
			UsageText: "usage: sm-kms init [flags]",
		},
	)

	require.True(t, result)
	require.Contains(t, stderr.String(), "usage: sm-kms init [flags]")
	require.NotContains(t, stderr.String(), "not yet implemented")
}

func TestIsHelpRequest_WithUsageTextOnly_NotHelpRequest(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	result := cryptoutilAppsFrameworkCli.IsHelpRequest(
		[]string{"some-arg"},
		cryptoutilAppsFrameworkCli.ClientNotImplementedMessageConfig{
			Stderr:    &stderr,
			UsageText: "usage: sm-kms init [flags]",
		},
	)

	require.False(t, result)
	require.Empty(t, stderr.String())
}
