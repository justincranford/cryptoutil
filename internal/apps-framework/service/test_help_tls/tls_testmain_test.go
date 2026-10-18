// Copyright (c) 2025-2026 Justin Cranford.

package test_help_tls

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTestTLSSettingsForTestMain(t *testing.T) {
	t.Parallel()

	settings := NewTestTLSSettingsForTestMain()

	require.NotNil(t, settings)
	require.NotEmpty(t, settings.StaticCertPEM)
	require.NotEmpty(t, settings.StaticKeyPEM)
}
