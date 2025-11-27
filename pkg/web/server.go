package web

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cgang/file-hub/pkg/config"
	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/session"
	"github.com/cgang/file-hub/pkg/stor"
	"github.com/cgang/file-hub/pkg/users"
	"github.com/cgang/file-hub/pkg/web/api"
	"github.com/cgang/file-hub/pkg/web/internal/auth"
	"github.com/cgang/file-hub/pkg/webdav"
	"github.com/cgang/file-hub/web"
	"github.com/gin-gonic/gin"
)

func Start(cfg config.WebConfig, storage stor.Storage, userService *users.Service, database *db.DB) {
	// Set the user service for authentication
	auth.SetUserService(userService)

	// Initialize session store
	sessionStore := session.NewStore()
	auth.SetSessionStore(sessionStore)

	webdavServer := webdav.New(storage)

	// Create a sub filesystem from the embedded files
	uiFiles, err := web.StaticFiles()
	if err != nil {
		log.Fatalf("Failed to load static files: %v", err)
	}

	engine := gin.Default()

	// Add a route for the setup page
	engine.GET("/setup", func(c *gin.Context) {
		// Check if database is empty to determine if we should show setup page
		if !users.HasAnyUser() {
			c.Redirect(http.StatusFound, "/api/setup")
			return
		}
		c.Redirect(http.StatusFound, "/ui/")
	})

	api.Register(engine.Group("/api"))

	// Register WebDAV with authentication middleware
	webdavGroup := engine.Group("/dav")
	webdavGroup.Use(auth.Authenticate)
	webdavServer.Register(webdavGroup)

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
