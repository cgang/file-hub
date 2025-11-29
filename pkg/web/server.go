package web

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/cgang/file-hub/pkg/config"
	"github.com/cgang/file-hub/pkg/session"
	"github.com/cgang/file-hub/pkg/web/api"
	"github.com/cgang/file-hub/pkg/web/internal/auth"
	"github.com/cgang/file-hub/pkg/webdav"
	"github.com/cgang/file-hub/web"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Start(ctx context.Context, cfg config.WebConfig) {

	// Initialize session store
	sessionStore := session.NewStore()
	auth.SetSessionStore(sessionStore)

	// Create a sub filesystem from the embedded files
	uiFiles, err := web.StaticFiles()
	if err != nil {
		log.Fatalf("Failed to load static files: %v", err)
	}

	engine := gin.Default()

	if cfg.Metrics {
		// Register Prometheus metrics endpoint
		engine.Handle(http.MethodGet, "/metrics", gin.WrapH(promhttp.Handler()))
	}

	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
		pprof.Register(engine)
	}

	api.Register(engine.Group("/api"))

	// Register WebDAV with authentication middleware
	webdavGroup := engine.Group("/dav")
	webdavGroup.Use(auth.Authenticate)
	webdav.Register(webdavGroup)

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

func Stop(ctx context.Context) {
	// TODO shutdown server gracefully
}
