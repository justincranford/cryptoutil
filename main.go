package main

import (
	cryptoutilServer "cryptoutil/internal/listener"
	"log"
)

func main() {
	start, _, err := cryptoutilServer.NewListener("localhost", 8080, true)
	if err != nil {
		log.Printf("failed to create listener: %v", err)
	}
	start()
}
