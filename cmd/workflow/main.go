// Copyright (c) 2025 Justin Cranford
//
//

package main

import (
	"os"

	"cryptoutil/internal/cmd/workflow"
)

func main() {
	os.Exit(workflow.Run(os.Args[1:]))
}
