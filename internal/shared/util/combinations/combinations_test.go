// Copyright (c) 2025 Justin Cranford
//
//

package combinations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type combinationsHappyPath struct {
	name     string
	m        M
	n        int
	expected Combinations
}

type combinationSadPath struct {
	name string
	m    M
	n    int
}

func TestCombinations_HappyPath(t *testing.T) {
	t.Parallel()

	valueA := value("A")
	valueB := value("B")
	valueC := value("C")
	valueD := value("D")

	m := M{}
	mA := M{valueA}
	mAB := M{valueA, valueB}
	mABC := M{valueA, valueB, valueC}
	mABCD := M{valueA, valueB, valueC, valueD}

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

	testCases := []combinationsHappyPath{
		{"0 of 0", m, 0, Combinations{}},
		{"0 of 1", mA, 0, Combinations{}},
		{"1 of 1", mA, 1, Combinations{combinationA}},
		{"0 of 2", mAB, 0, Combinations{}},
		{"1 of 2", mAB, 1, Combinations{combinationA, combinationB}},
		{"2 of 2", mAB, 2, Combinations{combinationAB}},
		{"0 of 3", mABC, 0, Combinations{}},
		{"1 of 3", mABC, 1, Combinations{combinationA, combinationB, combinationC}},
		{"2 of 3", mABC, 2, Combinations{combinationAB, combinationAC, combinationBC}},
		{"3 of 3", mABC, 3, Combinations{combinationABC}},
		{"0 of 4", mABCD, 0, Combinations{}},
		{"1 of 4", mABCD, 1, Combinations{combinationA, combinationB, combinationC, combinationD}},
		{"2 of 4", mABCD, 2, Combinations{combinationAB, combinationAC, combinationAD, combinationBC, combinationBD, combinationCD}},
		{"3 of 4", mABCD, 3, Combinations{combinationABC, combinationABD, combinationACD, combinationBCD}},
		{"4 of 4", mABCD, 4, Combinations{combinationABCD}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ComputeCombinations(tc.m, tc.n)
			for i, combination := range result {
				t.Logf("combination[%d] = %s, 0x%x ", i, combination.ToString(), combination.Encode())
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestCombinationsSadPath(t *testing.T) {
	t.Parallel()

	valueA := value("A")

	tests := []combinationSadPath{
		{"nil m", nil, 1},
		{"n < 0", M{valueA}, -1},
		{"n > len(m)", M{valueA}, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ComputeCombinations(tc.m, tc.n)
			require.Error(t, err)
		})
	}
}

func TestEncode(t *testing.T) {
	t.Parallel()

	valueA := value("A")
	valueB := value("B")
	valueC := value("C")

	expectededCombinationAB := []byte{2, 1, 'A', 1, 'B'}
	expectededCombinationC := []byte{1, 1, 'C'}

	combinationAB := combination{valueA, valueB}
	encodedCombinationAB := combinationAB.Encode()
	require.Equal(t, expectededCombinationAB, encodedCombinationAB)
	t.Logf("combinationAB = %s", combinationAB.ToString())
	t.Logf("encodedCombinationAB = 0x%x", encodedCombinationAB)

	combinationC := combination{valueC}
	encodedCombinationC := combinationC.Encode()
	require.Equal(t, expectededCombinationC, encodedCombinationC)
	t.Logf("combinationC = %s", combinationC.ToString())
	t.Logf("combinationC = 0x%x", encodedCombinationC)

	combos := Combinations{combinationAB, combinationC}

	encodedCombos := combos.Encode()
	require.Equal(t, 2, len(encodedCombos))
	require.Equal(t, expectededCombinationAB, encodedCombos[0])
	require.Equal(t, expectededCombinationC, encodedCombos[1])
}
