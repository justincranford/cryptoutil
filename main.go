package main

import (
	cryptoutilServer "cryptoutil/internal/listener"
)

func main() {
	start, _ := cryptoutilServer.NewListener("localhost:8080", true)
	start()
}
