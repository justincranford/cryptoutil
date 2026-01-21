// Copyright (c) 2025 Justin Cranford
//
//

package cmd

import (
	"log"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilServerApplication "cryptoutil/internal/kms/server/application"
)

// Server handles the KMS server command and subcommands.
func Server(parameters []string) {
	// reuse same Settings for start, ready, live, stop sub-commands, since they need to share private API coordinates
	settings, err := cryptoutilConfig.Parse(parameters, true)
	if err != nil {
		log.Fatal("Error parsing config:", err)
	}

	switch settings.SubCommand {
	case "start":
		startServerListenerApplication, err := cryptoutilServerApplication.StartServerListenerApplication(settings)
		if err != nil {
			log.Fatalf("failed to start server application: %v", err)
		}

		startServerListenerApplication.StartFunction() // blocks until server receives a signal to shutdown
	case "stop":
		err := cryptoutilServerApplication.SendServerListenerShutdownRequest(settings)
		if err != nil {
			log.Fatalf("failed to stop server application: %v", err)
		}
	case "live":
		_, err := cryptoutilServerApplication.SendServerListenerLivenessCheck(settings)
		if err != nil {
			log.Fatalf("failed to check server liveness: %v", err)
		}
	case "ready":
		_, err := cryptoutilServerApplication.SendServerListenerReadinessCheck(settings)
		if err != nil {
			log.Fatalf("failed to check server readiness: %v", err)
		}
	case "init":
		err := cryptoutilServerApplication.ServerInit(settings)
		if err != nil {
			log.Fatalf("failed to init server application: %v", err)
		}
	default:
		log.Fatalf("unknown subcommand: %v", settings.SubCommand)
	}
}
