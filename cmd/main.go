package main

import (
	"github.com/cgang/file-hub/internal/stor"
	"github.com/cgang/file-hub/internal/webdav"
	"log"
)

func main() {
	// Initialize WebDAV server with default configuration
	config := webdav.Config{
		RootDir: "./webdav_root",
		Port:    "8080",
	}

	log.Println("Initializing WebDAV server...")
	storage := &stor.OsStorage{}
	server := webdav.New(config, storage)

	log.Println("Starting WebDAV server...")
	server.Start()
}
