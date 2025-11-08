package cmd

import (
	"fmt"
	"os"
)

func Execute() {
	executable := os.Args[0] // Example executable: ./cryptoutil
	if len(os.Args) < 2 {
		printUsage(executable)
		os.Exit(1)
	}

	command := os.Args[1]     // Example command: server
	parameters := os.Args[2:] // Example parameters: --config-file, --port, --host, etc.

	switch command {
	case "server":
		server(parameters)
	case "identity":
		identity(parameters)
	// case "kv":
	// 	kv(parameters)
	case "help":
		printUsage(executable)
	default:
		printUsage(executable)
		fmt.Printf("Unknown command: %s %s\n", executable, command)
		os.Exit(1)
	}
}

func identity(parameters []string) {
	if len(parameters) < 1 {
		fmt.Println("Usage: cryptoutil identity <service> [options]")
		fmt.Println("Services:")
		fmt.Println("  authz    - OAuth 2.1 Authorization Server")
		fmt.Println("  idp      - OIDC Identity Provider")
		fmt.Println("  rs       - Resource Server")
		fmt.Println("  spa-rp   - SPA Relying Party")
		os.Exit(1)
	}

	service := parameters[0]
	serviceParams := parameters[1:]

	switch service {
	case "authz":
		identityAuthz(serviceParams)
	case "idp":
		identityIdp(serviceParams)
	case "rs":
		identityRs(serviceParams)
	case "spa-rp":
		identitySpaRp(serviceParams)
	default:
		fmt.Printf("Unknown identity service: %s\n", service)
		fmt.Println("Available services: authz, idp, rs, spa-rp")
		os.Exit(1)
	}
}

func identityAuthz(parameters []string) {
	fmt.Println("Starting OAuth 2.1 Authorization Server...")
	// TODO: Implement OAuth 2.1 Authorization Server
	fmt.Println("OAuth 2.1 Authorization Server not yet implemented")
	os.Exit(1)
}

func identityIdp(parameters []string) {
	fmt.Println("Starting OIDC Identity Provider...")
	// TODO: Implement OIDC Identity Provider
	fmt.Println("OIDC Identity Provider not yet implemented")
	os.Exit(1)
}

func identityRs(parameters []string) {
	fmt.Println("Starting Resource Server...")
	// TODO: Implement Resource Server
	fmt.Println("Resource Server not yet implemented")
	os.Exit(1)
}

func identitySpaRp(parameters []string) {
	fmt.Println("Starting SPA Relying Party...")
	// TODO: Implement SPA Relying Party
	fmt.Println("SPA Relying Party not yet implemented")
	os.Exit(1)
}

func printUsage(executable string) {
	fmt.Printf("Usage: %s <command> [options]\n", executable)
	fmt.Println("  server")
	fmt.Println("  identity")
}
