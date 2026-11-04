// Copyright (c) 2025-2026 Justin Cranford.
package cicd_golangci

import (
	"bytes"
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInternalMainWithRunner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want []string
		err  error
		code int
	}{
		{name: cryptoutilSharedMagic.EMPTY},
		{name: "separator", args: []string{"--", testMainFixArgument}, want: []string{testMainFixArgument}},
		{name: "flags", args: []string{testMainFixArgument}, want: []string{testMainFixArgument}},
		{name: "failure", err: errors.New("runner failure"), code: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stderr bytes.Buffer

			code := internalMainWithRunner(tt.args, &stderr, func(ctx context.Context, root string, args []string) error {
				require.NotNil(t, ctx)
				require.Equal(t, ".", root)
				require.Equal(t, tt.want, args)

				return tt.err
			})
			require.Equal(t, tt.code, code)

			if tt.err != nil {
				require.Contains(t, stderr.String(), tt.err.Error())
			} else {
				require.Empty(t, stderr.String())
			}
		})
	}
}

func TestInternalMainRejectsCallerPackages(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	require.Equal(t, 1, InternalMain([]string{"./..."}, nil, nil, &stderr))
	require.NotEmpty(t, stderr.String())
}
