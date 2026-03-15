// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides a stub for the identity demo CLI
// (pending Phase 8 domain reintegration).
package demo

import "io"

// Demo is a stub implementation pending Phase 8 reintegration.
func Demo(_ []string, _ io.Reader, _ io.Writer, stderr io.Writer) int {
_, _ = stderr.Write([]byte("identity demo not yet available (pending Phase 8 reintegration)\n"))

return 1
}
