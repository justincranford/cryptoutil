// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestContainerModeDetection tests container mode detection logic based on bind address.
// Container mode is triggered when BindPublicAddress == "0.0.0.0".
func TestContainerModeDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		bindPublicAddress  string
		bindPrivateAddress string
		wantContainerMode  bool
	}{
		{
			name:               "public 0.0.0.0 triggers container mode",
			bindPublicAddress:  cryptoutilSharedMagic.IPv4AnyAddress, // "0.0.0.0"
			bindPrivateAddress: cryptoutilSharedMagic.IPv4Loopback,   // "127.0.0.1"
			wantContainerMode:  true,
		},
		{
			name:               "both 127.0.0.1 is NOT container mode",
			bindPublicAddress:  cryptoutilSharedMagic.IPv4Loopback,
			bindPrivateAddress: cryptoutilSharedMagic.IPv4Loopback,
			wantContainerMode:  false,
		},
		{
			name:               "private 0.0.0.0 does NOT trigger container mode",
			bindPublicAddress:  cryptoutilSharedMagic.IPv4Loopback,
			bindPrivateAddress: cryptoutilSharedMagic.IPv4AnyAddress,
			wantContainerMode:  false,
		},
		{
			name:               "specific IP is NOT container mode",
			bindPublicAddress:  "192.168.1.100",
			bindPrivateAddress: cryptoutilSharedMagic.IPv4Loopback,
			wantContainerMode:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				BindPublicAddress:  tc.bindPublicAddress,
				BindPrivateAddress: tc.bindPrivateAddress,
			}

			isContainerMode := settings.BindPublicAddress == cryptoutilSharedMagic.IPv4AnyAddress
			require.Equal(t, tc.wantContainerMode, isContainerMode)
		})
	}
}
