// Copyright (c) 2025 Justin Cranford
//

package main

import (
"os"

sm "cryptoutil/internal/apps/sm/kms"
)

func main() {
os.Exit(sm.Main(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
