package combinations

import (
	"bytes"
	"fmt"
)

type M [][]byte
type value []byte
type combination []value
type combinations []combination

func ComputeCombinations(m M, n int) (combinations, error) {
	if m == nil {
		return nil, fmt.Errorf("m can't be nil")
	} else if len(m) >= 255 {
		return nil, fmt.Errorf("m can't be greater than 255")
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

func (combinations *combinations) Encode() [][]byte {
	var encodings [][]byte
	for _, combination := range *combinations {
		encodings = append(encodings, combination.Encode())
	}
	return encodings
}

func (combination *combination) Encode() []byte {
	var buffer bytes.Buffer
	buffer.WriteByte(uint8(len(*combination))) // encode number of values
	for _, value := range *combination {
		buffer.WriteByte(uint8(len(value))) // encode value length
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
