package cmd

import (
	"fmt"
	"os"
)

func Execute() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	command := os.Args[1]
	switch command {
	case "server":
		server()
	case "init":
		initConfig()
	case "key":
		key()
	case "help":
		printUsage()
	default:
		printUsage()
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func initConfig() {
	fmt.Println("init not implemented yet")
}

func key() {
	fmt.Println("init not implemented yet")
}

func printUsage() {
	fmt.Println("Usage: cryptoutil <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  server")
	fmt.Println("  init")
	fmt.Println("  key")
}
