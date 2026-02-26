package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	web.Start(ctx, cfg)

	// Start gRPC server if configured
	if cfg.Web.GRPCPort > 0 {
		if err := web.StartGRPCServer(cfg.Web.GRPCPort); err != nil {
			log.Printf("Warning: Failed to start gRPC server: %v", err)
		}
	}

	// wait for termination signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs
	log.Printf("Received signal %s, shutting down...", sig)

	web.Stop(ctx)

	// Stop gRPC server
	if cfg.Web.GRPCPort > 0 {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := web.StopGRPCServer(shutdownCtx); err != nil {
			log.Printf("Error stopping gRPC server: %v", err)
		}
		shutdownCancel()
	}

	cancel()

	// TODO wait for ongoing operations to finish
	time.Sleep(1 * time.Second)
}
