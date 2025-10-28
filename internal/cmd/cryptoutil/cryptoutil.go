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

func printUsage(executable string) {
	fmt.Printf("Usage: %s <command> [options]\n", executable)
	fmt.Println("  server")
}
