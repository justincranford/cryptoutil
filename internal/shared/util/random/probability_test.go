// Copyright (c) 2025-2026 Justin Cranford.
package random

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

type errorReader struct {
	err error
}

func (reader errorReader) Read([]byte) (int, error) {
	return 0, reader.err
}

func readerForUint32(value uint32) io.Reader {
	buffer := make([]byte, cryptoutilSharedMagic.BytesPerUint32)
	binary.LittleEndian.PutUint32(buffer, value)

	return bytes.NewReader(buffer)
}

func TestReadRandomUint32(t *testing.T) {
	t.Parallel()

	readErr := errors.New("injected random reader failure")
	zero := uint32(cryptoutilSharedMagic.Float32Zero)
	max := uint32(math.MaxUint32)
	nonSymmetric := uint32(math.MaxUint32 / cryptoutilSharedMagic.BytesPerUint32)
	tests := []struct {
		name             string
		reader           io.Reader
		want             uint32
		wantErr          error
		wantErrSubstring string
	}{
		{name: testCaseNameZero, reader: readerForUint32(zero), want: zero},
		{name: "maximum", reader: readerForUint32(max), want: max},
		{name: "non-symmetric sample", reader: readerForUint32(nonSymmetric), want: nonSymmetric},
		{name: testCaseNameShortRead, reader: bytes.NewReader(nil), wantErr: io.EOF},
		{name: testCaseNameReaderFailure, reader: errorReader{err: readErr}, wantErr: readErr},
		{name: testCaseNameNilReader, wantErrSubstring: testRandomReaderNilError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := readRandomUint32(tc.reader)
			switch {
			case tc.wantErr != nil:
				require.ErrorIs(t, err, tc.wantErr)
			case tc.wantErrSubstring != "":
				require.ErrorContains(t, err, tc.wantErrSubstring)
			default:
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestNormalizedRandomFloat64_Boundaries(t *testing.T) {
	t.Parallel()

	zero := uint32(cryptoutilSharedMagic.Float32Zero)
	max := uint32(math.MaxUint32)
	maxFraction := float64(max) / math.Ldexp(float64(cryptoutilSharedMagic.Float32One), cryptoutilSharedMagic.BytesPerUint32*cryptoutilSharedMagic.BitsToBytes)
	tests := []struct {
		name   string
		reader io.Reader
		want   float64
	}{
		{name: testCaseNameZero, reader: readerForUint32(zero), want: float64(cryptoutilSharedMagic.Float32Zero)},
		{name: "maximum remains below one", reader: readerForUint32(max), want: maxFraction},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizedRandomFloat64(tc.reader)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
			require.GreaterOrEqual(t, got, float64(cryptoutilSharedMagic.Float32Zero))
			require.Less(t, got, float64(cryptoutilSharedMagic.Float32One))
		})
	}
}

func TestSkipByProbability_HappyPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		prob        float32
		randomValue float64
		wantSkip    bool
		production  bool
	}{
		{name: "always run", prob: cryptoutilSharedMagic.Float32One, randomValue: cryptoutilSharedMagic.ConfidenceWeightFactors},
		{name: "zero probability skips even at zero", prob: cryptoutilSharedMagic.Float32Zero, randomValue: float64(cryptoutilSharedMagic.Float32Zero), wantSkip: true},
		{name: "sample below probability runs", prob: cryptoutilSharedMagic.ConfidenceWeightFactors, randomValue: cryptoutilSharedMagic.TestProbQuarter},
		{name: "sample at probability skips", prob: cryptoutilSharedMagic.ConfidenceWeightFactors, randomValue: cryptoutilSharedMagic.ConfidenceWeightFactors, wantSkip: true},
		{name: "sample above probability skips", prob: cryptoutilSharedMagic.ConfidenceWeightFactors, randomValue: math.Nextafter(cryptoutilSharedMagic.ConfidenceWeightFactors, cryptoutilSharedMagic.TestProbAlways), wantSkip: true},
		{name: "production random source", prob: cryptoutilSharedMagic.TestProbAlways, production: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.production {
				SkipByProbability(t, tc.prob)

				return
			}

			didSkip := false

			var gotSkipMessage []any

			skipByProbability(t, tc.prob, func() (float64, error) {
				return tc.randomValue, nil
			}, func(args ...any) {
				didSkip = true
				gotSkipMessage = args
			})

			require.Equal(t, tc.wantSkip, didSkip)

			if tc.wantSkip {
				require.Equal(t, []any{SkipMessage}, gotSkipMessage)
			}
		})
	}
}

func TestShouldSkipByProbability_SadPaths(t *testing.T) {
	t.Parallel()

	randomErr := errors.New("injected random sample failure")
	tests := []struct {
		name             string
		prob             float32
		randomFloatFn    func() (float64, error)
		wantErr          error
		wantErrSubstring string
	}{
		{name: "negative probability", prob: -cryptoutilSharedMagic.TestProbTenth, wantErrSubstring: "less than"},
		{name: "probability above one", prob: cryptoutilSharedMagic.Float32One + cryptoutilSharedMagic.TestProbTenth, wantErrSubstring: "greater than"},
		{name: "NaN probability", prob: float32(math.NaN()), wantErrSubstring: "NaN"},
		{name: "nil random function", prob: cryptoutilSharedMagic.Float32One, wantErrSubstring: "random float function is nil"},
		{name: "random source error", prob: cryptoutilSharedMagic.Float32One, randomFloatFn: func() (float64, error) { return 0, randomErr }, wantErr: randomErr},
		{name: "random value outside range", prob: cryptoutilSharedMagic.Float32One, randomFloatFn: func() (float64, error) { return float64(cryptoutilSharedMagic.Float32One), nil }, wantErrSubstring: "outside [0,1)"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := shouldSkipByProbability(tc.prob, tc.randomFloatFn)
			switch {
			case tc.wantErr != nil:
				require.ErrorIs(t, err, tc.wantErr)
			case tc.wantErrSubstring != "":
				require.ErrorContains(t, err, tc.wantErrSubstring)
			default:
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateProbabilityBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		probability float32
		valid       bool
	}{
		{name: "zero is inclusive", probability: cryptoutilSharedMagic.Float32Zero, valid: true},
		{name: "one is inclusive", probability: cryptoutilSharedMagic.Float32One, valid: true},
		{name: "below zero", probability: -cryptoutilSharedMagic.TestProbTenth},
		{name: "above one", probability: cryptoutilSharedMagic.Float32One + cryptoutilSharedMagic.TestProbTenth},
		{name: "not-a-number", probability: float32(math.NaN())},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateProbability(tc.probability)
			if tc.valid {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)
		})
	}
}

func TestSamplingBool(t *testing.T) {
	t.Parallel()

	readErr := errors.New("injected sampling reader failure")
	zero := uint32(cryptoutilSharedMagic.Float32Zero)
	max := uint32(math.MaxUint32)
	tests := []struct {
		name             string
		rate             float64
		reader           io.Reader
		useProduction    bool
		want             bool
		wantErr          error
		wantErrSubstring string
	}{
		{name: "zero rate", rate: cryptoutilSharedMagic.BaselineContributionZero, reader: readerForUint32(zero)},
		{name: "full rate from maximum sample", rate: cryptoutilSharedMagic.TestProbAlways, reader: readerForUint32(max), want: true},
		{name: "sample below rate", rate: cryptoutilSharedMagic.ConfidenceWeightFactors, reader: readerForUint32(zero), want: true},
		{name: "sample above rate", rate: cryptoutilSharedMagic.ConfidenceWeightFactors, reader: readerForUint32(max)},
		{name: "production random source at zero", rate: cryptoutilSharedMagic.BaselineContributionZero, useProduction: true},
		{name: "production random source at full rate", rate: cryptoutilSharedMagic.TestProbAlways, useProduction: true, want: true},
		{name: "negative rate", rate: -cryptoutilSharedMagic.TestProbTenth, reader: readerForUint32(zero), wantErrSubstring: testSamplingRateError},
		{name: "rate above one", rate: cryptoutilSharedMagic.TestProbAlways + cryptoutilSharedMagic.TestProbTenth, reader: readerForUint32(zero), wantErrSubstring: testSamplingRateError},
		{name: "NaN rate", rate: math.NaN(), reader: readerForUint32(zero), wantErrSubstring: testSamplingRateError},
		{name: testCaseNameNilReader, rate: cryptoutilSharedMagic.TestProbAlways, wantErrSubstring: testRandomReaderNilError},
		{name: testCaseNameShortRead, rate: cryptoutilSharedMagic.TestProbAlways, reader: bytes.NewReader(nil), wantErr: io.EOF},
		{name: testCaseNameReaderFailure, rate: cryptoutilSharedMagic.TestProbAlways, reader: errorReader{err: readErr}, wantErr: readErr},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got bool
				err error
			)
			if tc.useProduction {
				got, err = SamplingBool(tc.rate)
			} else {
				got, err = samplingBool(tc.rate, tc.reader)
			}

			switch {
			case tc.wantErr != nil:
				require.ErrorIs(t, err, tc.wantErr)
			case tc.wantErrSubstring != "":
				require.ErrorContains(t, err, tc.wantErrSubstring)
			default:
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestSamplingBool_StatisticalDistribution(t *testing.T) {
	t.Parallel()

	trueCount := 0

	for i := 0; i < cryptoutilSharedMagic.JoseJADefaultListLimit; i++ {
		result, err := SamplingBool(cryptoutilSharedMagic.ConfidenceWeightFactors)
		require.NoError(t, err)

		if result {
			trueCount++
		}
	}

	ratio := float64(trueCount) / cryptoutilSharedMagic.JoseJADefaultListLimit
	require.InDelta(t, cryptoutilSharedMagic.ConfidenceWeightFactors, ratio, 0.2)
}
