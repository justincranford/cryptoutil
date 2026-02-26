// Copyright (c) 2025 Justin Cranford
//
//

package util_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedUtil "cryptoutil/internal/shared/util"
)

func TestContains(t *testing.T) {
	t.Parallel()

	const (
		testApple     = "apple"
		testBanana    = "banana"
		testCherry    = "cherry"
		testFirst     = "first"
		testSecond    = "second"
		testLast      = "last"
		testOnly      = "only"
		testDuplicate = "duplicate"
	)

	tests := []struct {
		name       string
		setupFn    func() (slice any, searchItem any)
		wantResult bool
	}{
		{
			name: "found_element",
			setupFn: func() (any, any) {
				val1 := testApple
				val2 := testBanana
				val3 := testCherry
				slice := []*string{&val1, &val2, &val3}
				searchItem := testBanana

				return slice, &searchItem
			},
			wantResult: true,
		},
		{
			name: "not_found_element",
			setupFn: func() (any, any) {
				val1 := testApple
				val2 := testBanana
				val3 := testCherry
				slice := []*string{&val1, &val2, &val3}
				searchItem := "orange"

				return slice, &searchItem
			},
			wantResult: false,
		},
		{
			name: "empty_slice",
			setupFn: func() (any, any) {
				slice := []*string{}
				searchItem := testApple

				return slice, &searchItem
			},
			wantResult: false,
		},
		{
			name: "first_element",
			setupFn: func() (any, any) {
				val1 := testFirst
				val2 := testSecond
				val3 := "third"
				slice := []*string{&val1, &val2, &val3}
				searchItem := testFirst

				return slice, &searchItem
			},
			wantResult: true,
		},
		{
			name: "last_element",
			setupFn: func() (any, any) {
				val1 := testFirst
				val2 := testSecond
				val3 := testLast
				slice := []*string{&val1, &val2, &val3}
				searchItem := testLast

				return slice, &searchItem
			},
			wantResult: true,
		},
		{
			name: "integer_slice",
			setupFn: func() (any, any) {
				val1 := cryptoutilSharedMagic.JoseJADefaultMaxMaterials
				val2 := cryptoutilSharedMagic.MaxErrorDisplay
				val3 := cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days
				slice := []*int{&val1, &val2, &val3}
				searchItem := cryptoutilSharedMagic.MaxErrorDisplay

				return slice, &searchItem
			},
			wantResult: true,
		},
		{
			name: "not_found_integer",
			setupFn: func() (any, any) {
				val1 := cryptoutilSharedMagic.JoseJADefaultMaxMaterials
				val2 := cryptoutilSharedMagic.MaxErrorDisplay
				val3 := cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days
				slice := []*int{&val1, &val2, &val3}
				searchItem := 40

				return slice, &searchItem
			},
			wantResult: false,
		},
		{
			name: "single_element_found",
			setupFn: func() (any, any) {
				val := testOnly
				slice := []*string{&val}
				searchItem := testOnly

				return slice, &searchItem
			},
			wantResult: true,
		},
		{
			name: "single_element_not_found",
			setupFn: func() (any, any) {
				val := testOnly
				slice := []*string{&val}
				searchItem := "different"

				return slice, &searchItem
			},
			wantResult: false,
		},
		{
			name: "duplicate_elements",
			setupFn: func() (any, any) {
				val1 := testDuplicate
				val2 := testDuplicate
				val3 := "unique"
				slice := []*string{&val1, &val2, &val3}
				searchItem := testDuplicate

				return slice, &searchItem
			},
			wantResult: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			slice, searchItem := tc.setupFn()

			var result bool

			switch s := slice.(type) {
			case []*string:
				searchItemType, ok := searchItem.(*string)
				require.True(t, ok, "Error asserting searchItem to *string")

				//nolint:errcheck // Test cleanup - error irrelevant
				result = cryptoutilSharedUtil.Contains(s, searchItemType)
			case []*int:
				searchItemType, ok := searchItem.(*int)
				require.True(t, ok, "Error asserting searchItem to *int")

				//nolint:errcheck // Test cleanup - error irrelevant
				result = cryptoutilSharedUtil.Contains(s, searchItemType)
			default:
				require.FailNow(t, "Unsupported slice type", "%T", slice)
			}

			require.Equal(t, tc.wantResult, result)
		})
	}
}
