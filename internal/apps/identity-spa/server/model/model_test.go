package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBrowserSessionTableName(t *testing.T) {
	t.Parallel()

	require.Equal(t, "identity_spa_sessions", (BrowserSession{}).TableName())
}
