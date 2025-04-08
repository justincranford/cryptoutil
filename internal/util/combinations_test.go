package util

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type combinationsHappyPath struct {
	name     string
	m        M
	n        int
	expected combinations
}

type combinationSadPath struct {
	name string
	m    M
	n    int
}

type sequenceHappyPath struct {
	name     string
	inputs   []input
	expected sequence
}

func TestCombinations_HappyPath(t *testing.T) {
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
		{"0 of 0", m, 0, combinations{}},
		{"0 of 1", mA, 0, combinations{}},
		{"1 of 1", mA, 1, combinations{combinationA}},
		{"0 of 2", mAB, 0, combinations{}},
		{"1 of 2", mAB, 1, combinations{combinationA, combinationB}},
		{"2 of 2", mAB, 2, combinations{combinationAB}},
		{"0 of 3", mABC, 0, combinations{}},
		{"1 of 3", mABC, 1, combinations{combinationA, combinationB, combinationC}},
		{"2 of 3", mABC, 2, combinations{combinationAB, combinationAC, combinationBC}},
		{"3 of 3", mABC, 3, combinations{combinationABC}},
		{"0 of 4", mABCD, 0, combinations{}},
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

func TestCombinationsSadPath(t *testing.T) {
	valueA := value("A")

	tests := []combinationSadPath{
		{"nil m", nil, 1},
		{"n < 0", M{valueA}, -1},
		{"n > len(m)", M{valueA}, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Combinations(tc.m, tc.n)
			require.Error(t, err)
		})
	}
}

func TestSequenceHappyPath(t *testing.T) {
	valueA := value("A")
	valueB := value("B")
	valueC := value("C")

	mAB := M{valueA, valueB}
	mABC := M{valueA, valueB, valueC}

	combinationA := combination{valueA}
	combinationB := combination{valueB}
	// combinationC := combination{valueC}
	combinationAB := combination{valueA, valueB}
	combinationAC := combination{valueA, valueC}
	combinationBC := combination{valueB, valueC}
	combinationABC := combination{valueA, valueB, valueC}

	testCases := []sequenceHappyPath{
		{
			name:     "1 of 2 AND 2 of 3",
			inputs:   []input{{mAB, 1}, {mABC, 2}},
			expected: sequence{combinations{combinationA, combinationB}, combinations{combinationAB, combinationAC, combinationBC}},
		},
		{
			name:     "2 of 2 AND 3 of 3",
			inputs:   []input{{mAB, 2}, {mABC, 3}},
			expected: sequence{combinations{combinationAB}, combinations{combinationABC}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Sequence(tc.inputs)
			require.NoError(t, err)
			require.Len(t, result, len(tc.inputs))
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Test %s failed. Expected: %v, Got: %v", tc.name, tc.expected, result)
			}
		})
	}
}

func TestSequenceSadPath(t *testing.T) {
	t.Run("nil input slice", func(t *testing.T) {
		_, err := Sequence(nil)
		require.Error(t, err)
	})

	t.Run("Combinations failure", func(t *testing.T) {
		badInputs := []input{
			{nil, 1}, // This will trigger "m can't be nil" in Combinations
		}
		_, err := Sequence(badInputs)
		require.Error(t, err)
	})
}

func TestEncodeMethods(t *testing.T) {
	valueA := value("A")
	valueB := value("B")
	valueC := value("C")

	comb := combination{valueA, valueB}
	combos := combinations{comb, {valueC}}
	seq := sequence{combos}

	encodedComb := comb.Encode()
	encodedCombos := combos.Encode()
	encodedSeq := seq.Encode()

	require.NotEmpty(t, encodedComb)
	require.NotEmpty(t, encodedCombos)
	require.NotEmpty(t, encodedSeq)
}
