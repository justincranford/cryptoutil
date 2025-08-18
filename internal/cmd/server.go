package cmd

import (
	"log"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilServerApplication "cryptoutil/internal/server/application"
)

func server(executable string, parameters []string) {
	settings, err := cryptoutilConfig.Parse(parameters, true)
	if err != nil {
		log.Fatal("Error parsing config:", err)
	}

	start, _, err := cryptoutilServerApplication.StartServerApplication(settings)
	if err != nil {
		log.Fatalf("failed to start server application: %v", err)
	}
	start()
}
