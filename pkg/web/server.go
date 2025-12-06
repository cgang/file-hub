package web

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/cgang/file-hub/pkg/config"
	"github.com/cgang/file-hub/pkg/users"
	"github.com/cgang/file-hub/pkg/web/api"
	"github.com/cgang/file-hub/pkg/web/auth"
	"github.com/cgang/file-hub/pkg/web/dav"
	"github.com/cgang/file-hub/pkg/web/session"
	"github.com/cgang/file-hub/web"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	server *http.Server
)

func defaultRoute(c *gin.Context) {
	ok, err := users.HasAnyUser(c)
	if err != nil {
		c.String(http.StatusInternalServerError, "This site is not properly configured: %s", err)
		return
	}

	if ok {
		c.Redirect(http.StatusFound, "/ui/")
	} else {
		c.Redirect(http.StatusTemporaryRedirect, "/ui/?mode=setup")
	}
}

func Start(ctx context.Context, cfg config.WebConfig, realm string) {
	// Initialize session store
	sessionStore := session.NewStore()
	auth.Init(sessionStore, realm)

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
	webdav := engine.Group("/dav")
	webdav.Use(auth.Authenticate)
	dav.Register(webdav)

	engine.StaticFS("/ui", uiFiles)
	engine.GET("/", defaultRoute)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Starting Web server at %s", addr)
	server = &http.Server{Addr: addr, Handler: engine.Handler()}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Panicf("Failed to start web server: %s", err)
		}
	}()
}

func Stop(ctx context.Context) {
	if server != nil {
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Failed to shutdown web server: %s", err)
		}
		server = nil
	}
}
