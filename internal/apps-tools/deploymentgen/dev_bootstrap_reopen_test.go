// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentBootstrapReopen(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, token, extra string
		mismatch           bool
		wantError          bool
	}{
		{name: "exact incomplete bootstrap", token: strings.Repeat("a", 64)},
		{name: "short token", token: "a", wantError: true},
		{name: "invalid token", token: strings.Repeat("z", 64), wantError: true},
		{name: "noncanonical token", token: strings.Repeat("A", 64), wantError: true},
		{name: "changed topology", token: strings.Repeat("a", 64), mismatch: true, wantError: true},
		{name: "previous Kubernetes access", token: strings.Repeat("a", 64), extra: "kubeconfig", wantError: true},
		{name: "unknown recovery output", token: strings.Repeat("a", 64), extra: "unknown-output", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			base := t.TempDir()
			original, err := CreateDevelopmentPrivateDirectory(base)
			require.NoError(t, err)
			require.NoError(t, original.Write(cryptoutilSharedMagic.COMPOSE_YML, []byte("canonical-topology")))
			require.NoError(t, original.Write("secrets/p2sm1103/k3s-join-token", []byte(tc.token)))

			if tc.extra != "" {
				require.NoError(t, original.Write(tc.extra, []byte("private")))
			}

			require.NoError(t, original.Close())

			expected := []byte("canonical-topology")
			if tc.mismatch {
				expected = []byte("other-topology")
			}

			reopened, token, err := OpenDevelopmentBootstrapDirectory(filepath.Clean(base), expected)
			if tc.wantError {
				require.Error(t, err)
				require.Nil(t, reopened)
				require.Nil(t, token)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.token, string(token))
			require.Error(t, reopened.Write(cryptoutilSharedMagic.COMPOSE_YML, []byte("replacement")))
			require.NoError(t, reopened.Close())
		})
	}
}
