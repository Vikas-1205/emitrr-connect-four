package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/vikasgurjar/connect-four/analytics"
	"github.com/vikasgurjar/connect-four/db"
	"github.com/vikasgurjar/connect-four/server"
)

var (
	hub              *server.Hub
	database         *db.Database
	kafkaProducer    *analytics.KafkaProducer
	localAnalytics   *analytics.AnalyticsService
)

func main() {
	// Initialize hub
	hub = server.NewHub()

	// Initialize database (optional, will work without it)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		var err error
		database, err = db.NewDatabase(dbURL)
		if err != nil {
			log.Printf("Warning: Failed to connect to database: %v", err)
		} else {
			log.Println("Connected to PostgreSQL database")
			go processGameResults()
		}
	} else {
		log.Println("Running without database (DATABASE_URL not set)")
	}

	// Initialize local analytics service (works without Kafka)
	localAnalytics = analytics.NewAnalyticsService()

	// Initialize Kafka (optional)
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers != "" {
		brokers := strings.Split(kafkaBrokers, ",")
		kafkaProducer = analytics.NewKafkaProducer(brokers, "connect-four-events")
		log.Println("Connected to Kafka")
	} else {
		log.Println("Running without Kafka (KAFKA_BROKERS not set)")
	}

	// Start analytics processing (handles both Kafka and local analytics)
	go processAnalyticsEvents()

	// Setup HTTP routes
	mux := http.NewServeMux()

	// WebSocket endpoint
	mux.HandleFunc("/ws", hub.HandleWebSocket)

	// REST API endpoints
	mux.HandleFunc("/api/leaderboard", handleLeaderboard)
	mux.HandleFunc("/api/stats", handleStats)
	mux.HandleFunc("/api/games/history", handleGameHistory)
	mux.HandleFunc("/api/analytics", handleAnalytics)
	mux.HandleFunc("/api/analytics/user/", handleUserAnalytics)
	mux.HandleFunc("/api/health", handleHealth)

	// Serve static files for frontend
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

	// CORS middleware
	handler := corsMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleLeaderboard returns the leaderboard
func handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if database == nil {
		// Return in-memory leaderboard from completed games
		leaderboard := calculateInMemoryLeaderboard()
		json.NewEncoder(w).Encode(leaderboard)
		return
	}

	players, err := database.GetLeaderboard(10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(players)
}

// calculateInMemoryLeaderboard calculates leaderboard from in-memory games
func calculateInMemoryLeaderboard() []map[string]interface{} {
	wins := make(map[string]int)
	losses := make(map[string]int)

	for _, g := range hub.GetGames() {
		state := g.GetState()
		if state.Status != "completed" && state.Status != "forfeited" {
			continue
		}

		p1 := g.GetPlayer(1)
		p2 := g.GetPlayer(2)

		if p1 != nil && !p1.IsBot {
			if state.Winner == 1 {
				wins[p1.Username]++
			} else if state.Winner == 2 {
				losses[p1.Username]++
			}
		}

		if p2 != nil && !p2.IsBot {
			if state.Winner == 2 {
				wins[p2.Username]++
			} else if state.Winner == 1 {
				losses[p2.Username]++
			}
		}
	}

	// Convert to slice and sort
	var leaderboard []map[string]interface{}
	for username, w := range wins {
		leaderboard = append(leaderboard, map[string]interface{}{
			"username": username,
			"wins":     w,
			"losses":   losses[username],
		})
	}

	// Simple sort by wins (descending)
	for i := 0; i < len(leaderboard)-1; i++ {
		for j := i + 1; j < len(leaderboard); j++ {
			if leaderboard[j]["wins"].(int) > leaderboard[i]["wins"].(int) {
				leaderboard[i], leaderboard[j] = leaderboard[j], leaderboard[i]
			}
		}
	}

	if len(leaderboard) > 10 {
		leaderboard = leaderboard[:10]
	}

	return leaderboard
}

// handleStats returns game statistics
func handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if database != nil {
		stats, err := database.GetGameStats()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(stats)
		return
	}

	// Return basic in-memory stats
	games := hub.GetGames()
	completed := 0
	botGames := 0
	for _, g := range games {
		state := g.GetState()
		if state.Status == "completed" || state.Status == "forfeited" {
			completed++
		}
		if state.IsBotGame {
			botGames++
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"totalGames":  len(games),
		"completed":   completed,
		"botGames":    botGames,
		"playerGames": len(games) - botGames,
	})
}

// handleGameHistory returns recent completed games
func handleGameHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var gameHistory []map[string]interface{}

	for _, g := range hub.GetGames() {
		state := g.GetState()
		if state.Status != "completed" && state.Status != "forfeited" {
			continue
		}

		p1 := g.GetPlayer(1)
		p2 := g.GetPlayer(2)

		player1Name := "Unknown"
		player2Name := "Unknown"
		if p1 != nil {
			player1Name = p1.Username
		}
		if p2 != nil {
			player2Name = p2.Username
		}

		gameHistory = append(gameHistory, map[string]interface{}{
			"id":       state.ID,
			"player1":  player1Name,
			"player2":  player2Name,
			"winner":   state.Winner,
			"duration": g.GetDuration(),
			"isBotGame": state.IsBotGame,
		})
	}

	// Sort by most recent first (we'll assume later games have larger IDs)
	// Reverse the slice
	for i, j := 0, len(gameHistory)-1; i < j; i, j = i+1, j-1 {
		gameHistory[i], gameHistory[j] = gameHistory[j], gameHistory[i]
	}

	// Limit to 20 games
	if len(gameHistory) > 20 {
		gameHistory = gameHistory[:20]
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"games": gameHistory,
	})
}

// handleHealth returns server health status
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "healthy",
		"database": database != nil,
		"kafka":    kafkaProducer != nil,
		"time":     time.Now().Format(time.RFC3339),
	})
}

// processGameResults saves completed games to database
func processGameResults() {
	ticker := time.NewTicker(5 * time.Second)
	processedGames := make(map[string]bool)

	for range ticker.C {
		if database == nil {
			continue
		}

		for _, g := range hub.GetGames() {
			state := g.GetState()
			if (state.Status != "completed" && state.Status != "forfeited") || processedGames[state.ID] {
				continue
			}

			processedGames[state.ID] = true

			// Ensure players exist
			if state.Player1Username != "" {
				database.EnsurePlayer(state.Player1Username)
			}
			if state.Player2Username != "" && state.Player2Username != "ConnectBot" {
				database.EnsurePlayer(state.Player2Username)
			}

			// Update stats
			if state.Winner == 1 {
				database.UpdatePlayerStats(state.Player1Username, "win")
				if !state.IsBotGame {
					database.UpdatePlayerStats(state.Player2Username, "loss")
				}
			} else if state.Winner == 2 {
				if !state.IsBotGame {
					database.UpdatePlayerStats(state.Player2Username, "win")
				}
				database.UpdatePlayerStats(state.Player1Username, "loss")
			} else {
				database.UpdatePlayerStats(state.Player1Username, "draw")
				if !state.IsBotGame {
					database.UpdatePlayerStats(state.Player2Username, "draw")
				}
			}

			// Save game record
			winner := ""
			if state.Winner == 1 {
				winner = state.Player1Username
			} else if state.Winner == 2 {
				winner = state.Player2Username
			}

			database.SaveGame(db.GameRecord{
				ID:        state.ID,
				Player1:   state.Player1Username,
				Player2:   state.Player2Username,
				Winner:    winner,
				IsBotGame: state.IsBotGame,
				Duration:  g.GetDuration(),
				Status:    string(state.Status),
				CreatedAt: g.CreatedAt,
			})
		}
	}
}

// processAnalyticsEvents forwards analytics events to Kafka and local analytics
func processAnalyticsEvents() {
	analyticsChannel := hub.GetAnalyticsChannel()

	for event := range analyticsChannel {
		// Create analytics event
		analyticsEvent := analytics.Event{
			Type:      event.Type,
			GameID:    event.GameID,
			Player1:   event.Player1,
			Player2:   event.Player2,
			Winner:    event.Winner,
			Duration:  event.Duration,
			MoveCount: event.MoveCount,
			IsBotGame: event.IsBotGame,
			Timestamp: event.Timestamp,
		}

		// Send to Kafka if configured
		if kafkaProducer != nil {
			kafkaProducer.SendEvent(analyticsEvent)
		}

		// Process in local analytics service (always works)
		if localAnalytics != nil {
			localAnalytics.ProcessEvent(analyticsEvent)
		}

		// Log locally
		log.Printf("[Analytics] %s: game=%s", event.Type, event.GameID)
	}
}

// handleAnalytics returns comprehensive analytics stats
func handleAnalytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats := localAnalytics.GetStats()
	json.NewEncoder(w).Encode(stats)
}

// handleUserAnalytics returns analytics for a specific user
func handleUserAnalytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract username from URL: /api/analytics/user/{username}
	path := strings.TrimPrefix(r.URL.Path, "/api/analytics/user/")
	username := strings.TrimSuffix(path, "/")

	if username == "" {
		// Return all user metrics
		allMetrics := localAnalytics.GetAllUserMetrics()
		json.NewEncoder(w).Encode(allMetrics)
		return
	}

	metrics := localAnalytics.GetUserMetrics(username)
	if metrics == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(metrics)
}
