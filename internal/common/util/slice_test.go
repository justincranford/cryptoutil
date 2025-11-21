// Copyright (c) 2025 Justin Cranford
//
//

package util_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/common/util"
)

func TestContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFn    func() (slice any, searchItem any)
		wantResult bool
	}{
		{
			name: "found_element",
			setupFn: func() (any, any) {
				val1 := "apple"
				val2 := "banana"
				val3 := "cherry"
				slice := []*string{&val1, &val2, &val3}
				searchItem := "banana"

				return slice, &searchItem
			},
			wantResult: true,
		},
		{
			name: "not_found_element",
			setupFn: func() (any, any) {
				val1 := "apple"
				val2 := "banana"
				val3 := "cherry"
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
				searchItem := "apple"

				return slice, &searchItem
			},
			wantResult: false,
		},
		{
			name: "first_element",
			setupFn: func() (any, any) {
				val1 := "first"
				val2 := "second"
				val3 := "third"
				slice := []*string{&val1, &val2, &val3}
				searchItem := "first"

				return slice, &searchItem
			},
			wantResult: true,
		},
		{
			name: "last_element",
			setupFn: func() (any, any) {
				val1 := "first"
				val2 := "second"
				val3 := "last"
				slice := []*string{&val1, &val2, &val3}
				searchItem := "last"

				return slice, &searchItem
			},
			wantResult: true,
		},
		{
			name: "integer_slice",
			setupFn: func() (any, any) {
				val1 := 10
				val2 := 20
				val3 := 30
				slice := []*int{&val1, &val2, &val3}
				searchItem := 20

				return slice, &searchItem
			},
			wantResult: true,
		},
		{
			name: "not_found_integer",
			setupFn: func() (any, any) {
				val1 := 10
				val2 := 20
				val3 := 30
				slice := []*int{&val1, &val2, &val3}
				searchItem := 40

				return slice, &searchItem
			},
			wantResult: false,
		},
		{
			name: "single_element_found",
			setupFn: func() (any, any) {
				val := "only"
				slice := []*string{&val}
				searchItem := "only"

				return slice, &searchItem
			},
			wantResult: true,
		},
		{
			name: "single_element_not_found",
			setupFn: func() (any, any) {
				val := "only"
				slice := []*string{&val}
				searchItem := "different"

				return slice, &searchItem
			},
			wantResult: false,
		},
		{
			name: "duplicate_elements",
			setupFn: func() (any, any) {
				val1 := "duplicate"
				val2 := "duplicate"
				val3 := "unique"
				slice := []*string{&val1, &val2, &val3}
				searchItem := "duplicate"

				return slice, &searchItem
			},
			wantResult: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			slice, searchItem := tc.setupFn()

			var result bool

			switch s := slice.(type) {
			case []*string:
				result = util.Contains(s, searchItem.(*string))
			case []*int:
				result = util.Contains(s, searchItem.(*int))
			default:
				t.Fatalf("Unsupported slice type: %T", slice)
			}

			require.Equal(t, tc.wantResult, result)
		})
	}
}
