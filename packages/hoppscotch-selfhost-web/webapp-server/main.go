package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hoppscotch-selfhost-web/webapp-server/internal/bundle"
	"hoppscotch-selfhost-web/webapp-server/internal/config"
	"hoppscotch-selfhost-web/webapp-server/internal/crypto"
	"hoppscotch-selfhost-web/webapp-server/internal/server"
)

func main() {
	log.Println("Initializing Hoppscotch Web Static Server")

	cfg := config.Load()

	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	builder, err := bundle.NewBuilder()
	if err != nil {
		log.Fatalf("Failed to create bundle builder: %v", err)
	}

	content, files, err := builder.Build(cfg.FrontendPath)
	if err != nil {
		log.Fatalf("Failed to build bundle: %v", err)
	}

	bundleManager, err := bundle.NewManager(
		content,
		files,
		keyPair.SigningKey,
		keyPair.VerifyingKey,
		bundle.DefaultMaxSize,
	)
	if err != nil {
		log.Fatalf("Failed to create bundle manager: %v", err)
	}

	srv := server.New(bundleManager)
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	addr := fmt.Sprintf(":%d", cfg.Port)
	
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
