// Copyright (c) 2025 ZREV Enterprises LLC. All rights reserved.
// Use of this source code is governed by the MIT License.

package pbkdf2

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashPasswordWithIterations_CrandReadError(t *testing.T) {
	injectedErr := errors.New("injected crand.Read error")
	orig := pbkdf2CrandReadFn

	pbkdf2CrandReadFn = func(_ []byte) (int, error) { return 0, injectedErr }

	defer func() { pbkdf2CrandReadFn = orig }()

	_, err := HashPasswordWithIterations("password123", Iterations600k)

	require.ErrorIs(t, err, injectedErr)
}

func TestVerifyPassword_FormatErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		stored   string
		wantErr  string
	}{
		{
			name:    "non-empty first part",
			stored:  "invalid$pbkdf2-sha256$600000$aGVsbG8$aGVsbG8",
			wantErr: "invalid hash format: expected empty first part",
		},
		{
			name:    "wrong algorithm",
			stored:  "$wrongalgo$600000$aGVsbG8$aGVsbG8",
			wantErr: "invalid hash algorithm",
		},
		{
			name:    "non-numeric iterations",
			stored:  "$pbkdf2-sha256$notanumber$aGVsbG8$aGVsbG8",
			wantErr: "invalid iterations",
		},
		{
			name:    "invalid salt base64",
			stored:  "$pbkdf2-sha256$600000$!invalidbase64!$aGVsbG8",
			wantErr: "invalid salt encoding",
		},
		{
			name:    "invalid hash base64",
			stored:  "$pbkdf2-sha256$600000$aGVsbG8$!invalidbase64!",
			wantErr: "invalid hash encoding",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := VerifyPassword("password123", tc.stored)

			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
