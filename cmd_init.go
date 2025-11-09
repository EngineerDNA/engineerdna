package main

import (
	"fmt"
	"log"

	"github.com/engineerdna/engineerdna/internal/config"
	"github.com/engineerdna/engineerdna/internal/db"
)

func cmdInit() {
	fmt.Println("Initializing EngineerDNA...")

	cfg := config.DefaultConfig()
	if err := cfg.EnsureDirectories(); err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}

	// Initialize database
	database, err := config.NewDatabase(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := db.RunMigrations(database.DB); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize encryption key
	_, err = config.NewKeyStore()
	if err != nil {
		log.Fatalf("Failed to initialize key store: %v", err)
	}

	fmt.Println("EngineerDNA initialized successfully")
	fmt.Printf("Database: %s\n", cfg.Database.Path)
	fmt.Printf("Plugins directory: %s\n", cfg.Plugins.Directory)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Run 'engineerdna serve' to start the server")
	fmt.Println("2. Open http://localhost:3847 in your browser")
}
