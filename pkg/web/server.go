package web

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cgang/file-hub/pkg/config"
	"github.com/cgang/file-hub/pkg/stor"
	"github.com/cgang/file-hub/pkg/webdav"
	"github.com/cgang/file-hub/web"
	"github.com/gin-gonic/gin"
)

func Start(cfg config.WebConfig, storage stor.Storage) {
	webdavServer := webdav.New(storage)

	// Create a sub filesystem from the embedded files
	uiFiles, err := web.StaticFiles()
	if err != nil {
		log.Fatalf("Failed to load static files: %v", err)
	}

	engine := gin.Default()
	webdavServer.Register(engine.Group("/dav"))

	engine.StaticFS("/ui", uiFiles)
	engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/ui/")
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Starting Web server at %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
