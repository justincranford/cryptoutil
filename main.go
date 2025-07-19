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

	// TODO This is for pre-release testing only, remove before release
	settings.DevMode = true
	settings.Migrations = true

	start, _, err := cryptoutilServerApplication.StartServerApplication(settings)
	if err != nil {
		log.Fatalf("failed to start server application: %v", err)
	}
	start()
}
