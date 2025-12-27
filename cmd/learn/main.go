// Copyright (c) 2025 Justin Cranford
//
//

package main

import (
	"os"

	cryptoutilLearnCmd "cryptoutil/internal/cmd/learn"
)

func main() {
	exitCode := cryptoutilLearnCmd.Learn(os.Args[1:])
	os.Exit(exitCode)
}
