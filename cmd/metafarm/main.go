package main

import (
	"log"
	"metafarm/internal/app"
	"metafarm/internal/storage"
	"os"
)

var (
	openaiKey string
)

func init() {
	openaiKey = os.Getenv("METAFARM_OPENAI_KEY")
}

func main() {
	storage, err := storage.NewInMemoryStorage()
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	server := app.NewServer(":8080", storage, openaiKey)
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
