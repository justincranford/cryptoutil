// Copyright (c) 2025 Justin Cranford
//
//

package demo

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDemo_NoArguments(t *testing.T) {
	t.Parallel()

	exitCode := Demo([]string{}, nil, nil, nil)
	require.Equal(t, ExitFailure, exitCode)
}

func TestDemo_OneArgument(t *testing.T) {
	t.Parallel()

	exitCode := Demo([]string{"demo"}, nil, nil, nil)
	require.Equal(t, ExitFailure, exitCode)
}

func TestDemo_HelpCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "help command", args: []string{"demo", "help"}},
		{name: "help flag short", args: []string{"demo", "-h"}},
		{name: "help flag long", args: []string{"demo", "--help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			exitCode := Demo(tt.args, nil, nil, nil)
			require.Equal(t, ExitSuccess, exitCode)
		})
	}
}

func TestDemo_VersionCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "version command", args: []string{"demo", "version"}},
		{name: "version flag short", args: []string{"demo", "-v"}},
		{name: "version flag long", args: []string{"demo", "--version"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			exitCode := Demo(tt.args, nil, nil, nil)
			require.Equal(t, ExitSuccess, exitCode)
		})
	}
}

func TestDemo_UnknownCommand(t *testing.T) {
	t.Parallel()

	exitCode := Demo([]string{"demo", "nonexistent"}, nil, nil, nil)
	require.Equal(t, ExitFailure, exitCode)
}

func TestDemo_PrintUsage(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		printUsage()
	})
}

func TestDemo_PrintVersion(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		printVersion()
	})
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	require.NotNil(t, config)
	require.Equal(t, OutputHuman, config.OutputFormat)
	require.True(t, config.ContinueOnError)
	require.False(t, config.Verbose)
	require.False(t, config.Quiet)
	require.Greater(t, config.HealthTimeout, time.Duration(0))
	require.Greater(t, config.RetryCount, 0)
	require.Greater(t, config.RetryDelay, time.Duration(0))
}

func TestDetectNoColor(t *testing.T) {
	t.Parallel()

	// detectNoColor checks environment variables; just verify it returns a bool without panicking.
	_ = detectNoColor()
}

func TestParseArgs_EmptyArgs(t *testing.T) {
	t.Parallel()

	config := parseArgs([]string{})
	require.NotNil(t, config)
	require.Equal(t, OutputHuman, config.OutputFormat)
}

func TestParseArgs_AllFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		args   []string
		check  func(t *testing.T, c *Config)
	}{
		{
			name: "output json",
			args: []string{"--output", "json"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.Equal(t, OutputJSON, c.OutputFormat)
			},
		},
		{
			name: "output short json",
			args: []string{"-o", "json"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.Equal(t, OutputJSON, c.OutputFormat)
			},
		},
		{
			name: "output structured",
			args: []string{"--output", "structured"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.Equal(t, OutputStructured, c.OutputFormat)
			},
		},
		{
			name: "output human",
			args: []string{"--output", "human"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.Equal(t, OutputHuman, c.OutputFormat)
			},
		},
		{
			name: "no-color",
			args: []string{"--no-color"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.True(t, c.NoColor)
			},
		},
		{
			name: "verbose",
			args: []string{"--verbose"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.True(t, c.Verbose)
			},
		},
		{
			name: "quiet",
			args: []string{"--quiet"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.True(t, c.Quiet)
			},
		},
		{
			name: "quiet short",
			args: []string{"-q"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.True(t, c.Quiet)
			},
		},
		{
			name: "continue-on-error",
			args: []string{"--continue-on-error"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.True(t, c.ContinueOnError)
			},
		},
		{
			name: "fail-fast",
			args: []string{"--fail-fast"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.False(t, c.ContinueOnError)
			},
		},
		{
			name: "health-timeout",
			args: []string{"--health-timeout", "5s"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second, c.HealthTimeout)
			},
		},
		{
			name: "retry",
			args: []string{"--retry", "5"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, c.RetryCount)
			},
		},
		{
			name: "output missing value",
			args: []string{"--output"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.Equal(t, OutputHuman, c.OutputFormat)
			},
		},
		{
			name: "health-timeout missing value",
			args: []string{"--health-timeout"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.Equal(t, DefaultHealthTimeout, c.HealthTimeout)
			},
		},
		{
			name: "retry missing value",
			args: []string{"--retry"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.Greater(t, c.RetryCount, 0)
			},
		},
		{
			name: "multiple flags",
			args: []string{"--verbose", "--no-color", "-o", "json", "--fail-fast"},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				require.True(t, c.Verbose)
				require.True(t, c.NoColor)
				require.Equal(t, OutputJSON, c.OutputFormat)
				require.False(t, c.ContinueOnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := parseArgs(tt.args)
			require.NotNil(t, config)
			tt.check(t, config)
		})
	}
}

func TestExitCodes(t *testing.T) {
	t.Parallel()

	require.Equal(t, 0, ExitSuccess)
	require.Equal(t, 1, ExitPartialFailure)
	require.Equal(t, 2, ExitFailure)
}

func TestOutputFormats(t *testing.T) {
	t.Parallel()

	require.Equal(t, OutputFormat("human"), OutputHuman)
	require.Equal(t, OutputFormat("json"), OutputJSON)
	require.Equal(t, OutputFormat("structured"), OutputStructured)
}
