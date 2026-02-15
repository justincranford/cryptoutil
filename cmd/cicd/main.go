// Copyright (c) 2025 Justin Cranford

// Package main provides the cicd command entry point.
package main

import (
"os"

cryptoutilCmdCicd "cryptoutil/internal/cmd/cicd"
)

func main() {
os.Exit(cryptoutilCmdCicd.Cicd(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
