// Copyright (c) 2025-2026 Justin Cranford.
package intermediate

import (
	"bytes"
	"io"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestGenerateSerialNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		reader  func() io.Reader
		wantErr string
	}{
		{
			name: "nil reader",
			reader: func() io.Reader {
				return nil
			},
			wantErr: "random reader is required",
		},
		{
			name: "reader error",
			reader: func() io.Reader {
				return bytes.NewReader(nil)
			},
			wantErr: "failed to generate random bytes",
		},
		{
			name: "zero bytes are normalized to a positive serial",
			reader: func() io.Reader {
				return bytes.NewReader(make([]byte, cryptoutilSharedMagic.SerialNumberLength))
			},
		},
		{
			name: "high bit is cleared",
			reader: func() io.Reader {
				serialBytes := make([]byte, cryptoutilSharedMagic.SerialNumberLength)
				serialBytes[0] = ^byte(0)

				return bytes.NewReader(serialBytes)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			serialNumber, err := generateSerialNumber(tc.reader())
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, serialNumber)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, serialNumber)
			require.True(t, serialNumber.Sign() > 0)
			require.Less(t, serialNumber.BitLen(), cryptoutilSharedMagic.SerialNumberLength*cryptoutilSharedMagic.BitsPerByte)
		})
	}
}
