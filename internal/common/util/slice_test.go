// Copyright (c) 2025 Justin Cranford
//
//

package util_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/common/util"
)

func TestContains_FoundElement(t *testing.T) {
	t.Parallel()

	val1 := "apple"
	val2 := "banana"
	val3 := "cherry"

	slice := []*string{&val1, &val2, &val3}
	searchItem := "banana"

	// Act
	result := util.Contains(slice, &searchItem)

	// Assert
	require.True(t, result, "Should find element in slice")
}

func TestContains_NotFoundElement(t *testing.T) {
	t.Parallel()

	val1 := "apple"
	val2 := "banana"
	val3 := "cherry"

	slice := []*string{&val1, &val2, &val3}
	searchItem := "orange"

	// Act
	result := util.Contains(slice, &searchItem)

	// Assert
	require.False(t, result, "Should not find element not in slice")
}

func TestContains_EmptySlice(t *testing.T) {
	t.Parallel()

	slice := []*string{}
	searchItem := "apple"

	// Act
	result := util.Contains(slice, &searchItem)

	// Assert
	require.False(t, result, "Should return false for empty slice")
}

func TestContains_FirstElement(t *testing.T) {
	t.Parallel()

	val1 := "first"
	val2 := "second"
	val3 := "third"

	slice := []*string{&val1, &val2, &val3}
	searchItem := "first"

	// Act
	result := util.Contains(slice, &searchItem)

	// Assert
	require.True(t, result, "Should find first element")
}

func TestContains_LastElement(t *testing.T) {
	t.Parallel()

	val1 := "first"
	val2 := "second"
	val3 := "last"

	slice := []*string{&val1, &val2, &val3}
	searchItem := "last"

	// Act
	result := util.Contains(slice, &searchItem)

	// Assert
	require.True(t, result, "Should find last element")
}

func TestContains_IntegerSlice(t *testing.T) {
	t.Parallel()

	val1 := 10
	val2 := 20
	val3 := 30

	slice := []*int{&val1, &val2, &val3}
	searchItem := 20

	// Act
	result := util.Contains(slice, &searchItem)

	// Assert
	require.True(t, result, "Should work with integer slices")
}

func TestContains_NotFoundInteger(t *testing.T) {
	t.Parallel()

	val1 := 10
	val2 := 20
	val3 := 30

	slice := []*int{&val1, &val2, &val3}
	searchItem := 40

	// Act
	result := util.Contains(slice, &searchItem)

	// Assert
	require.False(t, result, "Should not find integer not in slice")
}

func TestContains_SingleElement_Found(t *testing.T) {
	t.Parallel()

	val := "only"
	slice := []*string{&val}
	searchItem := "only"

	// Act
	result := util.Contains(slice, &searchItem)

	// Assert
	require.True(t, result, "Should find single element")
}

func TestContains_SingleElement_NotFound(t *testing.T) {
	t.Parallel()

	val := "only"
	slice := []*string{&val}
	searchItem := "different"

	// Act
	result := util.Contains(slice, &searchItem)

	// Assert
	require.False(t, result, "Should not find different element in single-element slice")
}

func TestContains_DuplicateElements(t *testing.T) {
	t.Parallel()

	val1 := "duplicate"
	val2 := "duplicate"
	val3 := "unique"

	slice := []*string{&val1, &val2, &val3}
	searchItem := "duplicate"

	// Act
	result := util.Contains(slice, &searchItem)

	// Assert
	require.True(t, result, "Should find duplicate element")
}
