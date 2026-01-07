// Copyright (c) 2025 Justin Cranford

package main

import (
	"os"

	cryptoutilCACmd "cryptoutil/internal/cmd/cryptoutil/ca"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"start"}
	}

	cryptoutilCACmd.Execute(args)
}
