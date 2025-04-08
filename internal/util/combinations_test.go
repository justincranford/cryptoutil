package util

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type testcase struct {
	name     string
	m        M
	n        int
	expected combinations
}

func TestGenerateCombinations(t *testing.T) {
	valueA := value("A")
	valueB := value("B")
	valueC := value("C")
	valueD := value("D")

	m := M{}
	mA := M{valueA}
	mAB := M{valueA, valueB}
	mABC := M{valueA, valueB, valueC}
	mABCD := M{valueA, valueB, valueC, valueD}

	combinationEmpty := combination{}
	combinationA := combination{valueA}
	combinationB := combination{valueB}
	combinationC := combination{valueC}
	combinationD := combination{valueD}
	combinationAB := combination{valueA, valueB}
	combinationAC := combination{valueA, valueC}
	combinationAD := combination{valueA, valueD}
	combinationBC := combination{valueB, valueC}
	combinationBD := combination{valueB, valueD}
	combinationCD := combination{valueC, valueD}
	combinationABC := combination{valueA, valueB, valueC}
	combinationABD := combination{valueA, valueB, valueD}
	combinationACD := combination{valueA, valueC, valueD}
	combinationBCD := combination{valueB, valueC, valueD}
	combinationABCD := combination{valueA, valueB, valueC, valueD}

	testCases := []testcase{
		{"0 of 0", m, 0, combinations{combinationEmpty}},
		{"0 of 1", mA, 0, combinations{combinationEmpty}},
		{"1 of 1", mA, 1, combinations{combinationA}},
		{"0 of 2", mAB, 0, combinations{combinationEmpty}},
		{"1 of 2", mAB, 1, combinations{combinationA, combinationB}},
		{"2 of 2", mAB, 2, combinations{combinationAB}},
		{"0 of 3", mABC, 0, combinations{combinationEmpty}},
		{"1 of 3", mABC, 1, combinations{combinationA, combinationB, combinationC}},
		{"2 of 3", mABC, 2, combinations{combinationAB, combinationAC, combinationBC}},
		{"3 of 3", mABC, 3, combinations{combinationABC}},
		{"0 of 4", mABCD, 0, combinations{combinationEmpty}},
		{"1 of 4", mABCD, 1, combinations{combinationA, combinationB, combinationC, combinationD}},
		{"2 of 4", mABCD, 2, combinations{combinationAB, combinationAC, combinationAD, combinationBC, combinationBD, combinationCD}},
		{"3 of 4", mABCD, 3, combinations{combinationABC, combinationABD, combinationACD, combinationBCD}},
		{"4 of 4", mABCD, 4, combinations{combinationABCD}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Combinations(tc.m, tc.n)
			require.NoError(t, err)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Test %s failed. Expected: %v, Got: %v", tc.name, tc.expected, result)
			}
		})
	}
}
