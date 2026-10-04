package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Track HTTP request duration (Latency Histogram: p50, p90, p95, p99)
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: []float64{0.001, 0.002, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5},
		},
		[]string{"path", "method"},
	)

	// Total transaction decisions made by Redis Fast-Path
	transactionDecisionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transaction_decisions_total",
			Help: "Total count of transaction decisions evaluated by the fast-path",
		},
		[]string{"status"}, // "approved" or "rejected"
	)

	// Total messages saved to PostgreSQL by Kafka Consumer
	kafkaAuditSavedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_audit_saved_total",
			Help: "Total count of transaction audit records committed to PostgreSQL by Kafka consumer",
		},
	)
)

// MetricsMiddleware measures latency and registers counts for incoming HTTP requests
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(r.URL.Path, r.Method).Observe(duration)
	})
}
