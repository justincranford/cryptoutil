package main

import (
	"cryptoutil/internal/workflow"
	"os"
)

func main() {
	os.Exit(workflow.Run(os.Args[1:]))
}
