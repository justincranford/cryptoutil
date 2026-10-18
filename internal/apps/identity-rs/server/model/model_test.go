package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceRecordTableName(t *testing.T) {
	t.Parallel()

	require.Equal(t, "identity_rs_resources", (ResourceRecord{}).TableName())
}
