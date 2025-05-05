package keygen

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidInputs(t *testing.T) {
	testCases := []struct {
		secretBytesCount  int
		secretBytesLength int
	}{
		{1, 32},   // min count, min length
		{1, 64},   // min count, max length
		{254, 32}, // max count, min length
		{254, 64}, // max count, min length
		{128, 48}, // intermediate values
	}
	for _, testCase := range testCases {
		t.Run(
			"Count: "+strconv.Itoa(testCase.secretBytesCount)+" Length: "+strconv.Itoa(testCase.secretBytesLength),
			func(t *testing.T) {
				sharedSecrets, err := GenerateSharedSecrets(testCase.secretBytesCount, testCase.secretBytesLength)
				require.NoError(t, err)
				require.Len(t, sharedSecrets, testCase.secretBytesCount)

				for _, secret := range sharedSecrets {
					require.Len(t, secret, testCase.secretBytesLength)
				}
			})
	}
}

func TestZeroCount(t *testing.T) {
	_, err := GenerateSharedSecrets(0, 32)
	require.Error(t, err)
	require.Equal(t, "secretBytes count can't be zero", err.Error())
}

func TestNegativeCount(t *testing.T) {
	_, err := GenerateSharedSecrets(-1, 32)
	require.Error(t, err)
	require.Equal(t, "secretBytes count can't be negative", err.Error())
}

func TestCountGreaterThan256(t *testing.T) {
	_, err := GenerateSharedSecrets(257, 32)
	require.Error(t, err)
	require.Equal(t, "secretBytes count can't be greater than 256", err.Error())
}

func TestLengthLessThan32(t *testing.T) {
	_, err := GenerateSharedSecrets(2, 31)
	require.Error(t, err)
	require.Equal(t, "secretBytes length can't be greater than 32", err.Error())
}

func TestLengthGreaterThan64(t *testing.T) {
	_, err := GenerateSharedSecrets(2, 65)
	require.Error(t, err)
	require.Equal(t, "secretBytes length can't be greater than 64", err.Error())
}
