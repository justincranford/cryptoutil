// Copyright (c) 2025 Justin Cranford

package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientSecretHistory_TableName(t *testing.T) {
	t.Parallel()

	history := ClientSecretHistory{}
	tableName := history.TableName()

	require.Equal(t, "client_secret_history", tableName, "TableName should return 'client_secret_history'")
}
