package main

import (
	"cryptoutil/server"
)

func main() {
	start, _ := server.NewServer("localhost:8080", true)
	start()
}
