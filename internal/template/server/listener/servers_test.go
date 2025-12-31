// Copyright (c) 2025 Justin Cranford

package listener

import (
	"context"
	"testing"

	cryptoutilConfig "cryptoutil/internal/shared/config"

	"github.com/stretchr/testify/require"
)

func TestNewHTTPServers_AutoMode_HappyPath(t *testing.T) {
	settings := &cryptoutilConfig.ServerSettings{
		BindPublicPort:        0,
		BindPrivatePort:       0,
		TLSPublicMode:         cryptoutilConfig.TLSModeAuto,
		TLSPrivateMode:        cryptoutilConfig.TLSModeAuto,
		TLSPublicDNSNames:     []string{"localhost"},
		TLSPublicIPAddresses:  []string{"127.0.0.1"},
		TLSPrivateDNSNames:    []string{"localhost"},
		TLSPrivateIPAddresses: []string{"127.0.0.1"},
	}

	ctx := context.Background()
	h, err := NewHTTPServers(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, h)
	require.NotNil(t, h.PublicServer)
	require.NotNil(t, h.AdminServer)
}
