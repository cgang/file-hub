package main

import (
	"log"
	"github.com/cgang/file-hub/internal/webdav"
)

func main() {
	// Initialize WebDAV server with default configuration
	config := webdav.Config{
		RootDir: "./webdav_root",
		Port:    "8080",
	}

	log.Println("Initializing WebDAV server...")
	server := webdav.New(config)

	log.Println("Starting WebDAV server...")
	server.Start()
}
