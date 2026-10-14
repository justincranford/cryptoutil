// Copyright (c) 2025-2026 Justin Cranford.
//
//

package config

import (
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// TestGetTLSPEMBytes tests getTLSPEMBytes with various inputs.
func TestGetTLSPEMBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() *viper.Viper
		key     string
		wantNil bool
	}{
		{
			name:    "nil value for non-existent key",
			setup:   func() *viper.Viper { return viper.New() },
			key:     "non-existent-key",
			wantNil: true,
		},
		{
			name: "nil for non-bytes value",
			setup: func() *viper.Viper {
				v := viper.New()
				v.Set("int-value", cryptoutilSharedMagic.AnswerToLifeUniverseEverything)

				return v
			},
			key:     "int-value",
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			v := tc.setup()

			result := getTLSPEMBytes(v, tc.key)
			if tc.wantNil {
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)
			}
		})
	}
}

// TestRegisterAsSettings tests all RegisterAs* setting functions.
func TestRegisterAsSettings(t *testing.T) {
	t.Parallel()

	t.Run("bool", func(t *testing.T) {
		t.Parallel()

		setting := Setting{Value: true}
		require.True(t, RegisterAsBoolSetting(&setting))
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		setting := Setting{Value: "test-value"}
		require.Equal(t, "test-value", RegisterAsStringSetting(&setting))
	})

	t.Run("uint16", func(t *testing.T) {
		t.Parallel()

		setting := Setting{Value: uint16(cryptoutilSharedMagic.TestServerPort)}
		require.Equal(t, uint16(cryptoutilSharedMagic.TestServerPort), RegisterAsUint16Setting(&setting))
	})

	t.Run("string slice", func(t *testing.T) {
		t.Parallel()

		setting := Setting{Value: []string{testItem1, testItem2}}
		require.Equal(t, []string{testItem1, testItem2}, RegisterAsStringSliceSetting(&setting))
	})

	t.Run("string array", func(t *testing.T) {
		t.Parallel()

		setting := Setting{Value: []string{testItem1, testItem2}}
		require.Equal(t, []string{testItem1, testItem2}, RegisterAsStringArraySetting(&setting))
	})

	t.Run("duration", func(t *testing.T) {
		t.Parallel()

		setting := Setting{Value: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Minute}
		require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute, RegisterAsDurationSetting(&setting))
	})

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		setting := Setting{Value: cryptoutilSharedMagic.JoseJAMaxMaterials}
		require.Equal(t, cryptoutilSharedMagic.JoseJAMaxMaterials, RegisterAsIntSetting(&setting))
	})
}
