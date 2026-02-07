// Copyright (c) 2025 Justin Cranford

// Package main is the entrypoint for identity-demo demonstration service.
package main

import (
"os"

demo "cryptoutil/internal/cmd/identity/demo"
)

func main() {
os.Exit(demo.Demo(os.Args[1:], os.Stdout, os.Stderr))
}
