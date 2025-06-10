package main

import (
	"log"

	cryptoutilServerApplication "cryptoutil/internal/server/application"
)

func main() {
	start, _, err := cryptoutilServerApplication.StartServerApplication("localhost", 8080, true)
	if err != nil {
		log.Fatalf("failed to start server application: %v", err)
	}
	start()
}
