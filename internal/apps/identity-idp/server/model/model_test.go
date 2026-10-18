package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessionRecordTableName(t *testing.T) {
	t.Parallel()

	require.Equal(t, "identity_idp_sessions", (SessionRecord{}).TableName())
}
