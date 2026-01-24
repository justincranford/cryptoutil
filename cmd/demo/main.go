// Copyright (c) 2025 Justin Cranford
//
//

// Package main provides a unified demo CLI for cryptoutil products.
// This single binary supports subcommands for KMS, Identity, and full integration demos.
package main

import (
	cryptoutilCmdDemo "cryptoutil/internal/cmd/demo"
)

func main() {
	cryptoutilCmdDemo.Execute()
}
