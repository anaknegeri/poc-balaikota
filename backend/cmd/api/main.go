package main

import (
	"log"

	"people-counting/config"
	"people-counting/internal/app"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	// Load configuration
	cfg := config.Load()

	// Create and initialize server
	server := app.NewServer(cfg)
	if err := server.Initialize(); err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	// Run server
	log.Printf("Server is running on port %s", cfg.Server.Port)
	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
