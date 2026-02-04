// Copyright (c) 2025 Justin Cranford
//

package main

import (
"os"

ja "cryptoutil/internal/apps/jose/ja"
)

func main() {
os.Exit(ja.Main(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
