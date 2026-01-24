// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

// Package cipher provides the command-line interface for the cipher product.
package cipher

import (
	"io"

	cryptoutilAppsCipher "cryptoutil/internal/apps/cipher"
)

// Cipher is the thin entry point that delegates to internalCipher for testability.
// This follows the main/internalMain pattern from copilot instructions.
func Cipher(args []string) int {
	return internalCipher(args, nil, nil, nil)
}

// internalCipher is the testable implementation that accepts injected dependencies.
// Parameters follow the standard pattern: args, stdin, stdout, stderr.
// Currently stdin/stdout/stderr are unused but reserved for future use.
func internalCipher(args []string, _ io.Reader, _ io.Writer, _ io.Writer) int {
	return cryptoutilAppsCipher.Cipher(args)
}
