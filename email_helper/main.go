package main

import (
	"emailhelper/server"
	"log"
	"os"
)

func main() {
	port := os.Getenv("EMAIL_LOCAL_PORT")
	if port == "" {
		port = "8090"
	}
	if err := server.Start(port); err != nil {
		log.Fatalf("email helper failed: %v", err)
	}
}
