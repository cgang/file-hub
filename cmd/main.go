package main

import (
	"github.com/cgang/file-hub/internal/config"
	"github.com/cgang/file-hub/internal/stor"
	"github.com/cgang/file-hub/internal/webdav"
	"log"
)

func main() {
	// Load configuration from file or use defaults
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Printf("Failed to load config file: %v, using defaults", err)
		cfg = config.GetDefaultConfig()
	}

	log.Println("Initializing WebDAV server...")
	storage := &stor.OsStorage{}
	server := webdav.NewFromConfig(cfg.WebDAV, storage, cfg.Storage.RootDir)

	log.Println("Starting WebDAV server...")
	server.Start()
}
