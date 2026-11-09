package main

import (
	"log"

	"chat/internal/server"
)

func main() {
	log.Fatal(server.Run(":8090"))
}