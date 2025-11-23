package web

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/cgang/file-hub/internal/config"
	"github.com/cgang/file-hub/internal/stor"
	"github.com/cgang/file-hub/internal/webdav"
	"github.com/gin-gonic/gin"
)

//go:embed dist/*
var webFiles embed.FS

func Start(cfg config.WebConfig, storage stor.Storage) {
	webdavServer := webdav.New(storage)

	// Create a sub filesystem from the embedded files
	assets, err := fs.Sub(webFiles, "dist")
	if err != nil {
		log.Fatalf("Failed to create assets filesystem: %v", err)
	}

	engine := gin.Default()
	webdavServer.Register(engine.Group("/webdav"))

	engine.NoRoute(func(c *gin.Context) {
		http.FileServer(http.FS(assets)).ServeHTTP(c.Writer, c.Request)
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Starting Web server at %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
