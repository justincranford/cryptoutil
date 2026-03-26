// Copyright (c) 2025 Justin Cranford

// Package cmd provides CLI commands for the CA Server.
package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestNewStartCommand_NotNil(t *testing.T) {
	t.Parallel()

	cmd := NewStartCommand()

	require.NotNil(t, cmd)
	require.Equal(t, "start", cmd.Use)
}

func TestNewStartCommand_HasExpectedFlags(t *testing.T) {
	t.Parallel()

	cmd := NewStartCommand()

	require.NotNil(t, cmd.Flags().Lookup("bind"))
	require.NotNil(t, cmd.Flags().Lookup("port"))
	require.NotNil(t, cmd.Flags().Lookup(cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault))
}

func TestNewHealthCommand_NotNil(t *testing.T) {
	t.Parallel()

	cmd := NewHealthCommand()

	require.NotNil(t, cmd)
	require.Equal(t, "health", cmd.Use)
}

func TestNewHealthCommand_HasURLFlag(t *testing.T) {
	t.Parallel()

	cmd := NewHealthCommand()

	require.NotNil(t, cmd.Flags().Lookup("url"))
}
