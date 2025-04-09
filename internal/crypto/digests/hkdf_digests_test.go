package digests

import (
	"bytes"
	"errors"
	"testing"
)

type TestCaseHKDFHappyPath struct {
	name              string
	digestName        string
	secret            []byte
	salt              []byte
	info              []byte
	outputBytesLength int
}

type TestCaseHKDFSadPath struct {
	name              string
	digestName        string
	secret            []byte
	salt              []byte
	info              []byte
	outputBytesLength int
	expectedError     error
}

func TestHKDFHappyPath(t *testing.T) {
	happyPathTests := []TestCaseHKDFHappyPath{
		{"Valid SHA512", "SHA512", []byte("secret"), []byte("salt"), []byte("info"), 64},
		{"Valid SHA384", "SHA384", []byte("secret"), []byte("salt"), []byte("info"), 48},
		{"Valid SHA256", "SHA256", []byte("secret"), []byte("salt"), []byte("info"), 32},
		{"Valid SHA224", "SHA224", []byte("secret"), []byte("salt"), []byte("info"), 28},
	}

	for _, tt := range happyPathTests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := HKDF(tt.digestName, tt.secret, tt.salt, tt.info, tt.outputBytesLength)
			if err != nil {
				t.Errorf("HKDF(%s) unexpected error: %v", tt.digestName, err)
			}
			if len(output) != tt.outputBytesLength {
				t.Errorf("HKDF(%s) output length = %d, want %d", tt.digestName, len(output), tt.outputBytesLength)
			}
		})
	}

	t.Run("Unique Output for Different Salts", func(t *testing.T) {
		output1, _ := HKDF("SHA256", []byte("secret"), []byte("salt1"), []byte("info"), 32)
		output2, _ := HKDF("SHA256", []byte("secret"), []byte("salt2"), []byte("info"), 32)
		if bytes.Equal(output1, output2) {
			t.Errorf("HKDF output should be unique for different salts")
		}
	})
}

func TestHKDFHappyPathDifferentDigest(t *testing.T) {
	output1, _ := HKDF("SHA224", []byte("secret"), []byte("salt"), []byte("info"), 28)
	output2, _ := HKDF("SHA256", []byte("secret"), []byte("salt"), []byte("info"), 28)
	output3, _ := HKDF("SHA384", []byte("secret"), []byte("salt"), []byte("info"), 28)
	output4, _ := HKDF("SHA512", []byte("secret"), []byte("salt"), []byte("info"), 28)
	if bytes.Equal(output1, output2) {
		t.Errorf("HKDF output should be unique for different salts")
	} else if bytes.Equal(output1, output3) {
		t.Errorf("HKDF output should be unique for different salts")
	} else if bytes.Equal(output1, output4) {
		t.Errorf("HKDF output should be unique for different salts")
	} else if bytes.Equal(output2, output3) {
		t.Errorf("HKDF output should be unique for different salts")
	} else if bytes.Equal(output2, output4) {
		t.Errorf("HKDF output should be unique for different salts")
	} else if bytes.Equal(output3, output4) {
		t.Errorf("HKDF output should be unique for different salts")
	}
}

func TestHKDFHappyPathDifferentSecret(t *testing.T) {
	output1, _ := HKDF("SHA256", []byte("secret1"), []byte("salt"), []byte("info"), 32)
	output2, _ := HKDF("SHA256", []byte("secret2"), []byte("salt"), []byte("info"), 32)
	if bytes.Equal(output1, output2) {
		t.Errorf("HKDF output should be unique for different salts")
	}
}

func TestHKDFHappyPathDifferentSalt(t *testing.T) {
	output1, _ := HKDF("SHA256", []byte("secret"), []byte("salt1"), []byte("info"), 32)
	output2, _ := HKDF("SHA256", []byte("secret"), []byte("salt2"), []byte("info"), 32)
	if bytes.Equal(output1, output2) {
		t.Errorf("HKDF output should be unique for different salts")
	}
}

func TestHKDFHappyPathDifferentInfo(t *testing.T) {
	output1, _ := HKDF("SHA256", []byte("secret"), []byte("salt"), []byte("info1"), 28)
	output2, _ := HKDF("SHA256", []byte("secret"), []byte("salt"), []byte("info2"), 28)
	if bytes.Equal(output1, output2) {
		t.Errorf("HKDF output should be unique for different salts")
	}
}

func TestHKDFSadPath(t *testing.T) {
	sadPathTests := []TestCaseHKDFSadPath{
		{"Invalid Digest Name", "InvalidDigest", []byte("secret"), []byte("salt"), []byte("info"), 32, ErrInvalidNilDigestFunction},
		{"Nil Secret", "SHA256", nil, []byte("salt"), []byte("info"), 32, ErrInvalidNilSecret},
		{"Empty Secret", "SHA256", []byte{}, []byte("salt"), []byte("info"), 32, ErrInvalidEmptySecret},
		// {"Nil Salt", "SHA256", []byte("secret"), nil, []byte("info"), 32, ErrInvalidNilSalt},
		// {"Empty Salt", "SHA256", []byte("secret"), []byte{}, []byte("info"), 32, ErrInvalidEmptySalt},
		// {"Nil Info", "SHA256", []byte("secret"), []byte("salt"), nil, 32, ErrInvalidNilInfo},
		// {"Empty Info", "SHA256", []byte("secret"), []byte("salt"), []byte{}, 32, ErrInvalidEmptyInfo},
		{"Negative Output Length", "SHA256", []byte("secret"), []byte("salt"), []byte("info"), -1, ErrInvalidOutputBytesLengthNegative},
		{"Zero Output Length", "SHA256", []byte("secret"), []byte("salt"), []byte("info"), 0, ErrInvalidOutputBytesLengthZero},
		{"Excessive Output Length", "SHA224", []byte("secret"), []byte("salt"), []byte("info"), 224*32 + 1, ErrInvalidOutputBytesLengthTooBig},
		{"Excessive Output Length", "SHA256", []byte("secret"), []byte("salt"), []byte("info"), 255*32 + 1, ErrInvalidOutputBytesLengthTooBig},
		{"Excessive Output Length", "SHA384", []byte("secret"), []byte("salt"), []byte("info"), 384*32 + 1, ErrInvalidOutputBytesLengthTooBig},
		{"Excessive Output Length", "SHA512", []byte("secret"), []byte("salt"), []byte("info"), 512*32 + 1, ErrInvalidOutputBytesLengthTooBig},
	}

	for _, tt := range sadPathTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := HKDF(tt.digestName, tt.secret, tt.salt, tt.info, tt.outputBytesLength)
			if err == nil || !errors.Is(err, tt.expectedError) {
				t.Errorf("HKDF(%s) error = %v, expected %v", tt.digestName, err, tt.expectedError)
			}
		})
	}
}
