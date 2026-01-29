package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/vikasgurjar/connect-four/analytics"
)

// Standalone Kafka Consumer Service for Analytics
// Run this separately from the main game server

var analyticsService *analytics.AnalyticsService

func main() {
	log.Println("Starting Analytics Consumer Service...")

	// Initialize analytics service
	analyticsService = analytics.NewAnalyticsService()

	// Initialize Kafka consumer
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		log.Fatal("KAFKA_BROKERS environment variable is required")
	}

	brokers := strings.Split(kafkaBrokers, ",")
	consumer := analytics.NewKafkaConsumer(brokers, "connect-four-events", "analytics-consumer-group")

	// Start consuming in background
	go func() {
		log.Println("Starting Kafka consumer...")
		consumer.StartConsuming(func(event analytics.Event) {
			analyticsService.ProcessEvent(event)
		})
	}()

	// Setup HTTP server for analytics API
	mux := http.NewServeMux()

	// Analytics endpoints
	mux.HandleFunc("/analytics/stats", handleAnalyticsStats)
	mux.HandleFunc("/analytics/user/", handleUserMetrics)
	mux.HandleFunc("/analytics/events", handleRecentEvents)
	mux.HandleFunc("/health", handleHealthCheck)

	// CORS middleware
	handler := corsMiddleware(mux)

	port := os.Getenv("ANALYTICS_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Analytics service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleAnalyticsStats returns comprehensive analytics stats
func handleAnalyticsStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats := analyticsService.GetStats()
	json.NewEncoder(w).Encode(stats)
}

// handleUserMetrics returns metrics for a specific user
func handleUserMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract username from URL: /analytics/user/{username}
	path := strings.TrimPrefix(r.URL.Path, "/analytics/user/")
	username := strings.TrimSuffix(path, "/")

	if username == "" {
		// Return all users
		allMetrics := analyticsService.GetAllUserMetrics()
		json.NewEncoder(w).Encode(allMetrics)
		return
	}

	metrics := analyticsService.GetUserMetrics(username)
	if metrics == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(metrics)
}

// handleRecentEvents returns recent analytics events
func handleRecentEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	events := analyticsService.GetRecentEvents(50)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"count":  len(events),
	})
}

// handleHealthCheck returns service health status
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "analytics-consumer",
	})
}
