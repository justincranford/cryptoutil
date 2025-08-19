package cmd

import (
	"log"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilServerApplication "cryptoutil/internal/server/application"
)

func server(parameters []string) {
	settings, err := cryptoutilConfig.Parse(parameters, true)
	if err != nil {
		log.Fatal("Error parsing config:", err)
	}
	switch settings.SubCommand {
	case "start":
		start, _, err := cryptoutilServerApplication.StartServerApplication(settings)
		if err != nil {
			log.Fatalf("failed to start server application: %v", err)
		}
		start() // blocks until server receives a signal to shutdown
	case "stop":
		err := cryptoutilServerApplication.SendServerShutdownRequest(settings)
		if err != nil {
			log.Fatalf("failed to stop server application: %v", err)
		}
	default:
		log.Fatalf("unknown subcommand: %v", settings.SubCommand)
	}
}
