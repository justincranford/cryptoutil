package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthorizationDecisionTableName(t *testing.T) {
	t.Parallel()

	require.Equal(t, "authorization_decisions", (AuthorizationDecision{}).TableName())
}
