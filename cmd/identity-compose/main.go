// Copyright (c) 2025 Justin Cranford

// Package main is the entrypoint for identity-compose orchestration service.
package main

import (
"os"

compose "cryptoutil/internal/cmd/identity/compose"
)

func main() {
os.Exit(compose.Compose(os.Args[1:], os.Stdout, os.Stderr))
}
