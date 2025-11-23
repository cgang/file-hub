package main

import (
	"log"

	"github.com/cgang/file-hub/internal/config"
	"github.com/cgang/file-hub/internal/stor"
	"github.com/cgang/file-hub/web"
)

func main() {
	// Load configuration from file or use defaults
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Panicf("Failed to load config file: %s", err)
	}

	log.Println("Initializing WebDAV server...")
	storage := stor.NewStorage(cfg.Storage.RootDir)

	web.Start(cfg.Web, storage)
}
