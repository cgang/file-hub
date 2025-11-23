package main

import (
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/cgang/file-hub/internal/config"
	"github.com/cgang/file-hub/internal/stor"
	"github.com/cgang/file-hub/internal/webdav"
	"github.com/cgang/file-hub/web"
	"github.com/gin-gonic/gin"
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

	// Convert the storage root directory to an absolute path for logging
	storageRootAbs, err := filepath.Abs(cfg.Storage.RootDir)
	if err != nil {
		log.Printf("Warning: Could not resolve absolute path for storage directory '%s': %v", cfg.Storage.RootDir, err)
		storageRootAbs = cfg.Storage.RootDir  // Use original path if absolute path resolution fails
	}
	log.Printf("WebDAV root directory (absolute path): %s", storageRootAbs)

	webdavServer := webdav.NewFromConfig(cfg.WebDAV, storage, cfg.Storage.RootDir)

	// Create a sub filesystem from the embedded files
	subFS, err := fs.Sub(web.Assets, "dist")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}

	// Set up a file server for the embedded static assets
	fileServer := http.FileServer(http.FS(subFS))

	// Add a catch-all route to serve the frontend, while preserving API routes
	webdavServer.Engine.NoRoute(func(c *gin.Context) {
		// If it's an API call, don't serve frontend
		// Note: WebDAV routes are handled by the router before reaching NoRoute
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.AbortWithStatusJSON(404, gin.H{"error": "API endpoint not found"})
			return
		}

		// Otherwise, serve the frontend
		filePath := strings.TrimPrefix(c.Request.URL.Path, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		// Try to serve the requested file
		file, err := subFS.Open(filePath)
		if err != nil {
			// If the file doesn't exist, serve index.html for the frontend router
			c.Request.URL.Path = "/"
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		defer file.Close()

		// Serve the file
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	log.Println("Starting WebDAV server with integrated UI...")
	webdavServer.Start()
}
