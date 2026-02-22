// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package ca

import (
"bytes"
"testing"

"github.com/stretchr/testify/require"

)

func TestCA_SubcommandHelpFlags(t *testing.T) {
t.Parallel()

tests := []struct {
subcommand string
helpTexts  []string
}{
{subcommand: "client", helpTexts: []string{"pki ca client", "Run client operations"}},
{subcommand: "init", helpTexts: []string{"pki ca init", "Initialize database schema"}},
{subcommand: "health", helpTexts: []string{"pki ca health", "Check service health"}},
{subcommand: "livez", helpTexts: []string{"pki ca livez", "Check service liveness"}},
{subcommand: "readyz", helpTexts: []string{"pki ca readyz", "Check service readiness"}},
{subcommand: "shutdown", helpTexts: []string{"pki ca shutdown", "Trigger graceful shutdown"}},
}

for _, tc := range tests {
t.Run(tc.subcommand, func(t *testing.T) {
t.Parallel()

for _, flag := range []string{"--help", "-h", "help"} {
var stdout, stderr bytes.Buffer

exitCode := Ca([]string{tc.subcommand, flag}, nil, &stdout, &stderr)
require.Equal(t, 0, exitCode, "%s %s should succeed", tc.subcommand, flag)

combined := stdout.String() + stderr.String()
for _, expected := range tc.helpTexts {
require.Contains(t, combined, expected, "%s output should contain: %s", flag, expected)
}
}
})
}
}

func TestCA_MainHelp(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ca([]string{"--help"}, nil, &stdout, &stderr)
require.Equal(t, 0, exitCode)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, "pki ca")
}

func TestCA_Version(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ca([]string{"version"}, nil, &stdout, &stderr)
require.Equal(t, 0, exitCode)
}

func TestCA_SubcommandNotImplemented(t *testing.T) {
t.Parallel()

tests := []struct {
subcommand string
errorText  string
}{
{subcommand: "client", errorText: "not yet implemented"},
{subcommand: "init", errorText: "not yet implemented"},
}

for _, tc := range tests {
t.Run(tc.subcommand, func(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ca([]string{tc.subcommand}, nil, &stdout, &stderr)
require.Equal(t, 1, exitCode, "%s should exit with 1", tc.subcommand)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, tc.errorText)
})
}
}

func TestCA_ServerHelp(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ca([]string{"server", "--help"}, nil, &stdout, &stderr)
require.Equal(t, 0, exitCode)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, "pki ca server")
}

func TestCA_UnknownSubcommand(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ca([]string{"unknown-subcommand"}, nil, &stdout, &stderr)
require.Equal(t, 1, exitCode)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, "Unknown subcommand")
}

// TestCA_ServerParseError verifies the Parse error path in caServerStart.
func TestCA_ServerParseError(t *testing.T) {
var stdout, stderr bytes.Buffer

//nolint:goconst // Test-specific invalid flag, not a magic string.
exitCode := caServerStart([]string{"--this-flag-does-not-exist"}, &stdout, &stderr)
require.Equal(t, 1, exitCode)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, "Failed to parse configuration")
}

// TestCA_ServerCreateError verifies the NewFromConfig error path.
func TestCA_ServerCreateError(t *testing.T) {
var stdout, stderr bytes.Buffer

exitCode := caServerStart([]string{}, &stdout, &stderr)
require.Equal(t, 1, exitCode)

combined := stdout.String() + stderr.String()
require.Contains(t, combined, "Failed to create server")
}

func TestCA_SubcommandLiveServer(t *testing.T) {
tests := []struct {
subcommand       string
url              string
expectedExitCode int
expectedOutputs  []string
}{
{
subcommand:       "health",
url:              publicBaseURL + "/service/api/v1",
expectedExitCode: 0,
expectedOutputs:  []string{"HTTP 200"},
},
{
subcommand:       "livez",
url:              adminBaseURL,
expectedExitCode: 0,
expectedOutputs:  []string{"HTTP 200"},
},
{
subcommand:       "readyz",
url:              adminBaseURL,
expectedExitCode: 0,
expectedOutputs:  []string{"HTTP 200"},
},
}

for _, tc := range tests {
t.Run(tc.subcommand, func(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

exitCode := Ca([]string{tc.subcommand, "--url", tc.url}, nil, &stdout, &stderr)
require.Equal(t, tc.expectedExitCode, exitCode, "%s should succeed", tc.subcommand)

output := stdout.String() + stderr.String()
for _, expected := range tc.expectedOutputs {
require.Contains(t, output, expected, "Output should contain: %s", expected)
}
})
}
}
