package combinations

import (
	"bytes"
	"fmt"
)

type M []value
type value []byte
type combination []value
type combinations []combination

func ComputeCombinations(m M, n int) (combinations, error) {
	if m == nil {
		return nil, fmt.Errorf("m can't be nil")
	} else if n == 0 {
		return combinations{}, nil
	} else if n < 0 {
		return nil, fmt.Errorf("n cannot be negative")
	} else if n > len(m) {
		return nil, fmt.Errorf("n cannot be greater than m")
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

// Encoding functions

func (c combinations) Encode() []byte {
	var buffer bytes.Buffer
	for i, combination := range c {
		buffer.WriteByte(uint8(i))         // Write index as a uint8
		buffer.Write(combination.Encode()) // Write encoded combination
	}
	return buffer.Bytes()
}

func (v combination) Encode() []byte {
	var buffer bytes.Buffer
	for i, value := range v {
		buffer.WriteByte(uint8(i)) // Write index as a uint8
		buffer.Write(value)        // Write encoded value
	}
	return buffer.Bytes()
}

// M ToString method
func (m M) ToString() string {
	var buffer bytes.Buffer
	buffer.WriteString("[")
	for i, v := range m {
		buffer.WriteString(v.ToString())
		if i < len(m)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString("]")
	return buffer.String()
}

// value ToString method
func (v value) ToString() string {
	return fmt.Sprintf("%q", []byte(v))
}

// combination ToString method
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

// combinations ToString method
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
