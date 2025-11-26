package main

import (
	"log"

	"github.com/cgang/file-hub/pkg/config"
	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/stor"
	"github.com/cgang/file-hub/pkg/users"
	"github.com/cgang/file-hub/pkg/web"
)

func main() {
	// Load configuration from file or use defaults
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Panicf("Failed to load config file: %s", err)
	}

	// Initialize database connection
	database, err := db.New(cfg.Database.URI)
	if err != nil {
		log.Panicf("Failed to connect to database: %s", err)
	}
	defer database.Close()

	// Initialize database tables
	if err := database.InitDB(); err != nil {
		log.Panicf("Failed to initialize database: %s", err)
	}

	// Initialize user service
	userService := users.NewService(database)

	log.Println("Initializing WebDAV server...")
	storage := stor.NewStorage(userService)

	web.Start(cfg.Web, storage, userService)
}
