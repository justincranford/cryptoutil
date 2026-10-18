// Copyright (c) 2025-2026 Justin Cranford.
package cmdsplit

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const testToolExecutable = "tool"

func TestSplit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cmd  string
		want []string
	}{
		{name: "empty command", cmd: "", want: nil},
		{name: "whitespace command", cmd: "   ", want: nil},
		{name: "unquoted arguments", cmd: "go test ./...", want: []string{"go", "test", "./..."}},
		{name: "quoted argument keeps spaces", cmd: cryptoutilSharedMagic.DevSetupToolTypePython + ` -c "print('hello world')"`, want: []string{cryptoutilSharedMagic.DevSetupToolTypePython, "-c", "print('hello world')"}},
		{name: "empty quoted argument is preserved", cmd: testToolExecutable + ` "" value`, want: []string{testToolExecutable, "", "value"}},
		{name: "unmatched quote keeps remaining text together", cmd: testToolExecutable + ` "value with spaces`, want: []string{testToolExecutable, "value with spaces"}},
		{name: "multiple quoted arguments", cmd: testToolExecutable + ` "first arg" second "third arg"`, want: []string{testToolExecutable, "first arg", "second", "third arg"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, Split(tc.cmd))
		})
	}
}
