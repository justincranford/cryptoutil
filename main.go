package main

import (
	"log"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilServerApplication "cryptoutil/internal/server/application"
)

func main() {
	settings, err := cryptoutilConfig.Parse()
	if err != nil {
		log.Fatal("Error parsing config:", err)
	}

	start, _, err := cryptoutilServerApplication.StartServerApplication(settings, "localhost", 8080, true)
	if err != nil {
		log.Fatalf("failed to start server application: %v", err)
	}
	start()
}
