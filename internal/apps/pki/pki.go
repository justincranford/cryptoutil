// Copyright (c) 2025 Justin Cranford
//
//

package pki

import (
"fmt"
"io"

cryptoutilAppsPkiCa "cryptoutil/internal/apps/pki/ca"
)

// Pki is the entry point for the PKI product.
// Routes to the CA (Certificate Authority) service.
func Pki(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
// Default to "ca" service if no args provided
if len(args) == 0 {
args = []string{"ca"}
}

// For PKI product, only CA service is supported currently
if args[0] != "ca" {
_, _ = fmt.Fprintf(stderr, "Unknown PKI service: %s\n", args[0])
_, _ = fmt.Fprintf(stderr, "Usage: pki [ca] [subcommand] [flags]\n")
_, _ = fmt.Fprintln(stderr, "")
_, _ = fmt.Fprintln(stderr, "Available services:")
_, _ = fmt.Fprintln(stderr, "  ca    Certificate Authority service")

return 1
}

// Route to CA service
return cryptoutilAppsPkiCa.Ca(args[1:], stdin, stdout, stderr)
}
