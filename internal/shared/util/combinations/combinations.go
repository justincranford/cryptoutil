// Copyright (c) 2025 Justin Cranford
//
//

// Package combinations provides utilities for generating combinatorial permutations.
package combinations

import (
	"bytes"
	"fmt"
)

const (
	// maxUint8Value is the maximum value for uint8 (255), used for bounds checking in encoding.
	maxUint8Value = 255
)

type (
	M            [][]byte
	value        []byte
	combination  []value
	combinations []combination
)

func ComputeCombinations(m M, n int) (combinations, error) {
	if m == nil {
		return nil, fmt.Errorf("m can't be nil")
	} else if len(m) >= maxUint8Value {
		return nil, fmt.Errorf("m can't be greater than %d", maxUint8Value)
	} else if n == 0 {
		return combinations{}, nil
	} else if n < 0 {
		return nil, fmt.Errorf("n can't be negative")
	} else if n > len(m) {
		return nil, fmt.Errorf("n can't be greater than m")
	}

	var result combinations

	combination := make(combination, n) // Properly initialize the 'combination' slice

	var helper func(int, int)

	helper = func(start, depth int) {
		if depth == n {
			// Directly create a new 'combination' instance as a slice
			combo := append(combination[:0:0], combination...) // Create a copy of 'combination'
			result = append(result, combo)                     // Add the new combination to the result

			return
		}

		for i := start; i < len(m); i++ {
			combination[depth] = m[i]
			helper(i+1, depth+1)
		}
	}
	helper(0, 0)

	return result, nil
}

// ENCODE

func (c *combinations) Encode() [][]byte {
	encodings := make([][]byte, 0, len(*c))
	for _, combination := range *c {
		encodings = append(encodings, combination.Encode())
	}

	return encodings
}

func (c *combination) Encode() []byte {
	var buffer bytes.Buffer
	// Add bounds checking for uint8 conversion
	combLen := len(*c)
	if combLen > maxUint8Value {
		panic("combination length exceeds uint8 maximum")
	}

	buffer.WriteByte(uint8(combLen)) // encode number of values

	for _, value := range *c {
		valueLen := len(value)
		if valueLen > maxUint8Value {
			panic("value length exceeds uint8 maximum")
		}

		buffer.WriteByte(uint8(valueLen)) // encode value length
		buffer.Write(value)
	}

	return buffer.Bytes()
}

// ToString methods

func (m M) ToString() string {
	var buffer bytes.Buffer

	buffer.WriteString("[")

	for i, v := range m {
		buffer.WriteString(string(v))

		if i < len(m)-1 {
			buffer.WriteString(", ")
		}
	}

	buffer.WriteString("]")

	return buffer.String()
}

func (v value) ToString() string {
	return fmt.Sprintf("%q", []byte(v))
}

func (c combination) ToString() string {
	var buffer bytes.Buffer

	buffer.WriteString("[")

	for i, v := range c {
		buffer.WriteString(v.ToString())

		if i < len(c)-1 {
			buffer.WriteString(", ")
		}
	}

	buffer.WriteString("]")

	return buffer.String()
}

func (c combinations) ToString() string {
	var buffer bytes.Buffer

	buffer.WriteString("[")

	for i, combination := range c {
		buffer.WriteString(combination.ToString())

		if i < len(c)-1 {
			buffer.WriteString(", ")
		}
	}

	buffer.WriteString("]")

	return buffer.String()
}
