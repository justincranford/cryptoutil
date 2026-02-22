package server

import (
"context"
"testing"

"github.com/stretchr/testify/require"
)

func TestNewKMSServer_NilChecks(t *testing.T) {
t.Parallel()

tests := []struct {
name    string
ctx     context.Context
wantErr string
}{
{
name:    "nil context",
ctx:     nil,
wantErr: "context cannot be nil",
},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

server, err := NewKMSServer(tc.ctx, nil) //nolint:staticcheck // SA1012: Intentionally testing nil context handling
require.Error(t, err)
require.Nil(t, server)
require.Contains(t, err.Error(), tc.wantErr)
})
}
}

func TestNewKMSServer_NilSettings(t *testing.T) {
t.Parallel()

ctx := context.Background()

server, err := NewKMSServer(ctx, nil)
require.Error(t, err)
require.Nil(t, server)
require.Contains(t, err.Error(), "settings cannot be nil")
}
