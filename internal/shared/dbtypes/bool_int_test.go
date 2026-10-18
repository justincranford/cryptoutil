// Copyright (c) 2025-2026 Justin Cranford.
//

package dbtypes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testBoolIntFalseCaseName = "false"
	testBoolIntTrueCaseName  = "true"
)

func TestBoolInt_Value(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   BoolInt
		want int64
	}{
		{name: testBoolIntTrueCaseName, in: true, want: 1},
		{name: testBoolIntFalseCaseName, in: false, want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.in.Value()
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestBoolInt_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      any
		want    BoolInt
		wantErr string
	}{
		{name: testBoolIntTrueCaseName, in: true, want: true},
		{name: testBoolIntFalseCaseName, in: false, want: false},
		{name: "int64 true", in: int64(1), want: true},
		{name: "int64 false", in: int64(0), want: false},
		{name: "int32 true", in: int32(1), want: true},
		{name: "byte string true", in: []byte("1"), want: true},
		{name: "string false", in: "0", want: false},
		{name: "NULL", in: nil, wantErr: "cannot scan NULL"},
		{name: "unsupported type", in: float64(1), wantErr: "cannot scan float64"},
		{name: "invalid integer", in: "not-a-number", wantErr: "cannot parse"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got BoolInt

			err := got.Scan(tc.in)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestBoolInt_ScanNilReceiver(t *testing.T) {
	t.Parallel()

	var value *BoolInt

	require.ErrorContains(t, value.Scan(int64(1)), "nil *BoolInt")
}
