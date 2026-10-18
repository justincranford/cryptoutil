package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientSessionTableName(t *testing.T) {
	t.Parallel()

	require.Equal(t, "identity_rp_sessions", (ClientSession{}).TableName())
}
