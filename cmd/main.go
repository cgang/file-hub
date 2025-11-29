package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cgang/file-hub/pkg/config"
	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/users"
	"github.com/cgang/file-hub/pkg/web"
)

func main() {
	// Load configuration from file or use defaults
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Panicf("Failed to load config file: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	db.Init(ctx, cfg.Database.URI)
	users.Init(ctx, cfg.Realm)

	web.Start(ctx, cfg.Web)

	// wait for termination signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs
	log.Printf("Received signal %s, shutting down...", sig)

	web.Stop(ctx)

	cancel()

	// TODO wait for ongoing operations to finish
}
