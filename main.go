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

	"sush/handlers"
	"sush/store"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://sush:sush@localhost:5432/sush?sslmode=disable"
	}

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	s, err := store.New(connStr, redisAddr)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	h := handlers.New(s)

	mux := http.NewServeMux()

	// register routes
	mux.HandleFunc("GET /health", h.HealthHandler)
	mux.HandleFunc("POST /shorten", h.ShortenHandler)
	mux.HandleFunc("GET /{slug}", h.RedirectHandler)
	mux.HandleFunc("GET /stats/{slug}", h.StatsHandler)
	mux.HandleFunc("GET /metrics", h.MetricsHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// start server in background
	go func() {
		fmt.Println("sush is up at http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server died: %v", err)
		}
	}()

	// wait for ctrl+c
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	fmt.Println("\nshutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}

	fmt.Println("bye")
}
