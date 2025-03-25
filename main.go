package main

import (
	cryptoutilServer "cryptoutil/internal/server"
)

func main() {
	start, _ := cryptoutilServer.NewServer("localhost:8080", true)
	start()
}
