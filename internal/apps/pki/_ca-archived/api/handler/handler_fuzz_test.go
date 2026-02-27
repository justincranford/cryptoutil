// Copyright (c) 2025 Justin Cranford

package handler

import (
	"testing"
)

// FuzzParseESTCSR fuzzes EST CSR parsing (base64, PEM, DER formats).
func FuzzParseESTCSR(f *testing.F) {
	// Seed corpus with valid base64, PEM, DER data.
	f.Add([]byte("MIICUTCCATkCAQAwDjEMMAoGA1UEAwwDZm9vMIIBIjANBgkqhkiG9w0BAQEFAAOC"))
	f.Add([]byte("-----BEGIN CERTIFICATE REQUEST-----\nMIICUTCCATkCAQAwDjEMMAoGA1UEAwwDZm9v\n-----END CERTIFICATE REQUEST-----"))
	f.Add([]byte{0x30, 0x82, 0x01, 0x51})

	f.Fuzz(func(_ *testing.T, data []byte) {
		h := &Handler{}
		// Should not panic, just return error for invalid data.
		_, _ = h.parseESTCSR(data)
	})
}
