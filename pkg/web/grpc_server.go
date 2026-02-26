package web

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/sync"
	"google.golang.org/grpc"
)

var (
	grpcServer     *grpc.Server
	grpcListener   net.Listener
)

// StartGRPCServer initializes and starts the gRPC server
func StartGRPCServer(grpcPort int) error {
	if grpcPort <= 0 {
		// gRPC server disabled
		return nil
	}

	// Create listener
	var err error
	grpcListener, err = net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		return fmt.Errorf("failed to create gRPC listener: %w", err)
	}

	// Get database connection
	database := db.GetDB()
	if database == nil {
		return fmt.Errorf("database connection not available")
	}

	// Create gRPC server with interceptors
	grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(sync.AuthInterceptor()),
		grpc.StreamInterceptor(sync.StreamAuthInterceptor()),
		grpc.MaxRecvMsgSize(100*1024*1024), // 100MB max message size for uploads
		grpc.MaxSendMsgSize(100*1024*1024), // 100MB max message size for downloads
	)

	// Register sync service
	syncService := sync.NewGRPCService(database)
	sync.RegisterSyncServiceServer(grpcServer, syncService)

	// Start server in background
	go func() {
		log.Printf("Starting gRPC server on port %d", grpcPort)
		if err := grpcServer.Serve(grpcListener); err != nil && err != grpc.ErrServerStopped {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	return nil
}

// StopGRPCServer gracefully stops the gRPC server
func StopGRPCServer(ctx context.Context) error {
	if grpcServer == nil {
		return nil
	}

	log.Println("Shutting down gRPC server...")

	// Use graceful shutdown with context
	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Println("gRPC server stopped gracefully")
		return nil
	case <-ctx.Done():
		log.Println("gRPC server shutdown timed out, forcing stop")
		grpcServer.Stop()
		return ctx.Err()
	}
}

// GetGRPCAddress returns the address the gRPC server is listening on
func GetGRPCAddress() string {
	if grpcListener == nil {
		return ""
	}
	return grpcListener.Addr().String()
}
