// Copyright (c) 2025 Justin Cranford
//
//

package util

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidInputs(t *testing.T) {
	testCases := []struct {
		count       int
		bytesLength int
	}{
		{1, 1},       // min count, min length
		{1, 1024},    // min count, high length
		{1024, 32},   // high count, min length
		{1024, 1024}, // high count, high length
		{256, 256},   // intermediate values
	}
	for _, testCase := range testCases {
		t.Run(
			"Count: "+strconv.Itoa(testCase.count)+" Length: "+strconv.Itoa(testCase.bytesLength),
			func(t *testing.T) {
				nBytes, err := GenerateMultipleBytes(testCase.count, testCase.bytesLength)
				require.NoError(t, err)
				require.Len(t, nBytes, testCase.count)

				for _, bytes := range nBytes {
					require.Len(t, bytes, testCase.bytesLength)
				}
			})
	}
}

func TestZeroCount(t *testing.T) {
	_, err := GenerateMultipleBytes(0, 32)
	require.Error(t, err)
	require.Equal(t, "count can't be less than 1", err.Error())
}

func TestNegativeCount(t *testing.T) {
	_, err := GenerateMultipleBytes(-1, 32)
	require.Error(t, err)
	require.Equal(t, "count can't be less than 1", err.Error())
}

func TestZeroLength(t *testing.T) {
	_, err := GenerateMultipleBytes(32, 0)
	require.Error(t, err)
	require.Equal(t, "length can't be less than 1", err.Error())
}

func TestNegativeLength(t *testing.T) {
	_, err := GenerateMultipleBytes(32, -1)
	require.Error(t, err)
	require.Equal(t, "length can't be less than 1", err.Error())
}
