// Copyright (c) 2025 Justin Cranford

package listener

import (
	"context"
	"testing"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestNewHTTPServers_AutoMode_HappyPath(t *testing.T) {
	settings := cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, h)
	require.NotNil(t, h.PublicServer)
	require.NotNil(t, h.AdminServer)
}
