// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package testutil

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafeBuffer_ConcurrentWriteRead(t *testing.T) {
	t.Parallel()

	var sb SafeBuffer

	const (
		goroutines = 10
		iterations = 100
	)

	var wg sync.WaitGroup

	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()

			for range iterations {
				_, err := sb.Write([]byte("x"))
				require.NoError(t, err)

				_ = sb.String()
				_ = sb.Len()
			}
		}()
	}

	wg.Wait()

	require.Equal(t, goroutines*iterations, sb.Len())
}

func TestSafeBuffer_Reset(t *testing.T) {
	t.Parallel()

	var sb SafeBuffer

	_, err := sb.Write([]byte("hello"))
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, sb.Len())
	require.Equal(t, "hello", sb.String())

	sb.Reset()
	require.Equal(t, 0, sb.Len())
	require.Equal(t, "", sb.String())
}
