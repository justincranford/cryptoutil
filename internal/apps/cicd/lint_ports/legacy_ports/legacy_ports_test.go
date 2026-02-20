// Copyright (c) 2025 Justin Cranford

package legacy_ports

import (
	"testing"
	
	"github.com/stretchr/testify/require"
)

func TestGetServiceForLegacyPort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		port uint16
		want string
	}{
		{name: "cipher-im 8888", port: 8888, want: "cipher-im"},
		{name: "cipher-im 8889", port: 8889, want: "cipher-im"},
		{name: "cipher-im 8890", port: 8890, want: "cipher-im"},
		{name: "jose-ja 9443", port: 9443, want: "jose-ja"},
		{name: "jose-ja 8092", port: 8092, want: "jose-ja"},
		{name: "pki-ca 8443", port: 8443, want: "pki-ca"},
		{name: "unknown port", port: 12345, want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GetServiceForLegacyPort(tt.port)
			require.Equal(t, tt.want, got)
		})
	}
}
