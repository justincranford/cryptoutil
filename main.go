package main

import (
	"log"

	cryptoutilServer "cryptoutil/internal/listener"
)

func main() {
	start, _, err := cryptoutilServer.NewHttpListener("localhost", 8080, true)
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}
	start()
}
