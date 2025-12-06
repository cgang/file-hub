package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cgang/file-hub/pkg/config"
	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/stor"
	"github.com/cgang/file-hub/pkg/users"
	"github.com/cgang/file-hub/pkg/web"
)

func main() {
	log.Default().SetOutput(os.Stdout) // switch to standard output

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to get config: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	db.Init(ctx, cfg.Database.URI)
	stor.Init(ctx, cfg)
	users.Init(ctx, cfg.Realm)

	web.Start(ctx, cfg.Web, cfg.Realm)

	// wait for termination signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs
	log.Printf("Received signal %s, shutting down...", sig)

	web.Stop(ctx)

	cancel()

	// TODO wait for ongoing operations to finish
}
