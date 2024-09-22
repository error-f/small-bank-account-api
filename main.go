package main

import (
	"log"
)

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}

	server.StartServer()
}
