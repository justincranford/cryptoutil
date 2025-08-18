package cmd

import (
	"fmt"
	"os"
)

func Execute() {
	executable := os.Args[0]
	if len(os.Args) < 2 {
		printUsage(executable)
		os.Exit(1)
	}
	command := os.Args[1]
	parameters := os.Args[2:]

	switch command {
	case "server":
		server(executable, parameters)
	case "init":
		initConfig()
	case "key":
		key()
	case "help":
		printUsage(executable)
	default:
		printUsage(executable)
		fmt.Printf("Unknown command: %s %s\n", executable, command)
		os.Exit(1)
	}
}

func initConfig() {
	fmt.Println("init not implemented yet")
}

func key() {
	fmt.Println("init not implemented yet")
}

func printUsage(executable string) {
	fmt.Printf("Usage: %s <command> [options]\n", executable)
	fmt.Println("  server")
	fmt.Println("  init")
	fmt.Println("  key")
}
