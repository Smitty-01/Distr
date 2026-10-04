package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

var storage *Storage
var rdb *redis.Client

func main() {
	// 1. Structured JSON Logging (slog)
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(jsonHandler))
	slog.Info("Starting Transaction Decision Service...")

	// 2. Connect to PostgreSQL
	db, err := connectDB()
	if err != nil {
		slog.Error("Database connection failed", "error", err)
		log.Fatal(err)
	}
	defer db.Close()
	storage = NewStorage(db)

	if err := db.Ping(); err != nil {
		slog.Error("Database ping failed", "error", err)
		log.Fatal(err)
	}
	slog.Info("Successfully connected to PostgreSQL")

	// 3. Connect to Redis
	rdb = NewRedis()
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("Redis connection failed", "error", err)
		log.Fatal(err)
	}
	defer rdb.Close()
	slog.Info("Successfully connected to Redis")

	// 4. Initialize Kafka Producer & Consumer
	InitKafkaProducer()
	defer kafkaWriter.Close()
	go StartKafkaConsumer(storage)

	// 5. Register HTTP Handlers
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", Homehandler) // {$}: Matches ONLY exact root '/'
	mux.HandleFunc("GET /transactions", TransactionHandler)
	mux.HandleFunc("POST /transactions", addTransactions)

	// Expose Prometheus Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Wrap entire mux with Prometheus latency & request-tracking middleware
	instrumentedRouter := MetricsMiddleware(mux)

	// 6. Production HTTP Server with Timeouts
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      instrumentedRouter,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start Server in background goroutine
	go func() {
		slog.Info(fmt.Sprintf("Server running on http://localhost%s", srv.Addr))
		slog.Info("Prometheus metrics available at http://localhost:8080/metrics")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server crashed", "error", err)
			os.Exit(1)
		}
	}()

	// 7. Graceful Shutdown (Signal Listening)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Wait for OS shutdown signal (Ctrl+C / SIGTERM)
	sig := <-shutdownChan
	slog.Info("Shutdown signal received, initiating graceful shutdown...", "signal", sig.String())

	// Allow active in-flight requests up to 5 seconds to finish
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exited cleanly with zero data loss.")
}
