// Copyright (c) 2025 Justin Cranford

package outdated_deps

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestSaveDepCache_MarshalError(t *testing.T) {
	t.Parallel()

	stubMarshalFn := func(_ any, _ string, _ string) ([]byte, error) {
		return nil, fmt.Errorf("injected marshal error")
	}

	cacheFile := t.TempDir() + "/test-cache.json"
	cache := cryptoutilSharedMagic.DepCache{}

	err := saveDepCache(cacheFile, cache, stubMarshalFn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to marshal cache JSON")

	// Verify file was not created.
	_, statErr := os.Stat(cacheFile)
	require.True(t, os.IsNotExist(statErr))
}
