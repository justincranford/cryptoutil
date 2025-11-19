// Copyright (c) 2025 Justin Cranford
//
//

package main

import (
	"cryptoutil/internal/cmd/workflow"
	"os"
)

func main() {
	os.Exit(workflow.Run(os.Args[1:]))
}
