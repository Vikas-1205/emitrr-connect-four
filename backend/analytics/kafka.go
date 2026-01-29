package analytics

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaProducer sends analytics events to Kafka
type KafkaProducer struct {
	writer *kafka.Writer
}

// Event represents an analytics event
type Event struct {
	Type      string                 `json:"type"`
	GameID    string                 `json:"gameId,omitempty"`
	Player1   string                 `json:"player1,omitempty"`
	Player2   string                 `json:"player2,omitempty"`
	Winner    string                 `json:"winner,omitempty"`
	Duration  int                    `json:"duration,omitempty"`
	MoveCount int                    `json:"moveCount,omitempty"`
	IsBotGame bool                   `json:"isBotGame,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		Async:        true,
	}

	return &KafkaProducer{writer: writer}
}

// SendEvent sends an analytics event to Kafka
func (p *KafkaProducer) SendEvent(event Event) error {
	if p.writer == nil {
		return nil // Kafka not configured
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(event.GameID),
		Value: data,
		Time:  event.Timestamp,
	}

	return p.writer.WriteMessages(context.Background(), msg)
}

// Close closes the Kafka producer
func (p *KafkaProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// KafkaConsumer consumes analytics events from Kafka
type KafkaConsumer struct {
	reader *kafka.Reader
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(brokers []string, topic, groupID string) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &KafkaConsumer{reader: reader}
}

// StartConsuming starts consuming events and processing them
func (c *KafkaConsumer) StartConsuming(handler func(Event)) {
	for {
		msg, err := c.reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Kafka read error: %v", err)
			continue
		}

		var event Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			continue
		}

		handler(event)
	}
}

// Close closes the Kafka consumer
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

// UserMetrics tracks user-specific analytics
type UserMetrics struct {
	Username     string    `json:"username"`
	TotalGames   int       `json:"totalGames"`
	Wins         int       `json:"wins"`
	Losses       int       `json:"losses"`
	Draws        int       `json:"draws"`
	WinRate      float64   `json:"winRate"`
	BotGames     int       `json:"botGames"`
	PvPGames     int       `json:"pvpGames"`
	TotalMoves   int       `json:"totalMoves"`
	AvgGameTime  float64   `json:"avgGameTime"`
	TotalTime    int       `json:"totalTime"`
	LastPlayed   time.Time `json:"lastPlayed"`
	WinStreak    int       `json:"winStreak"`
	MaxWinStreak int       `json:"maxWinStreak"`
}

// HourlyStats tracks games per hour
type HourlyStats struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

// DailyStats tracks games per day
type DailyStats struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// AnalyticsService processes analytics events - Enhanced version
type AnalyticsService struct {
	mu             sync.RWMutex
	totalGames     int
	completedGames int
	totalDuration  int
	botGames       int
	pvpGames       int
	totalMoves     int

	// Winner tracking
	winnerCounts map[string]int

	// User-specific metrics
	userMetrics map[string]*UserMetrics

	// Time-based metrics
	gamesPerHour map[int]int     // hour (0-23) -> count
	gamesPerDay  map[string]int  // YYYY-MM-DD -> count

	// Event log for audit trail
	eventLog []Event
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{
		winnerCounts: make(map[string]int),
		userMetrics:  make(map[string]*UserMetrics),
		gamesPerHour: make(map[int]int),
		gamesPerDay:  make(map[string]int),
		eventLog:     make([]Event, 0),
	}
}

// ProcessEvent processes an analytics event
func (s *AnalyticsService) ProcessEvent(event Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store in event log (keep last 1000 events)
	s.eventLog = append(s.eventLog, event)
	if len(s.eventLog) > 1000 {
		s.eventLog = s.eventLog[1:]
	}

	switch event.Type {
	case "game_started":
		s.processGameStarted(event)
	case "game_ended":
		s.processGameEnded(event)
	}
}

// processGameStarted handles game start events
func (s *AnalyticsService) processGameStarted(event Event) {
	s.totalGames++

	if event.IsBotGame {
		s.botGames++
	} else {
		s.pvpGames++
	}

	// Track games per hour
	hour := event.Timestamp.Hour()
	s.gamesPerHour[hour]++

	// Track games per day
	day := event.Timestamp.Format("2006-01-02")
	s.gamesPerDay[day]++

	// Initialize user metrics if needed
	s.ensureUserMetrics(event.Player1)
	s.ensureUserMetrics(event.Player2)

	// Update user game counts
	if player := s.userMetrics[event.Player1]; player != nil {
		player.TotalGames++
		player.LastPlayed = event.Timestamp
		if event.IsBotGame {
			player.BotGames++
		} else {
			player.PvPGames++
		}
	}

	if !event.IsBotGame {
		if player := s.userMetrics[event.Player2]; player != nil {
			player.TotalGames++
			player.LastPlayed = event.Timestamp
			player.PvPGames++
		}
	}

	log.Printf("[Analytics] Game started: %s (P1: %s, P2: %s, Bot: %v)",
		event.GameID, event.Player1, event.Player2, event.IsBotGame)
}

// processGameEnded handles game end events
func (s *AnalyticsService) processGameEnded(event Event) {
	s.completedGames++

	if event.Duration > 0 {
		s.totalDuration += event.Duration
	}

	if event.MoveCount > 0 {
		s.totalMoves += event.MoveCount
	}

	// Track winner
	if event.Winner != "" && event.Winner != "draw" {
		s.winnerCounts[event.Winner]++

		// Update winner's user metrics
		if player := s.userMetrics[event.Winner]; player != nil {
			player.Wins++
			player.WinStreak++
			if player.WinStreak > player.MaxWinStreak {
				player.MaxWinStreak = player.WinStreak
			}
			player.TotalTime += event.Duration
			player.TotalMoves += event.MoveCount / 2 // Approximate moves per player
			player.WinRate = float64(player.Wins) / float64(player.TotalGames) * 100
			if player.TotalGames > 0 {
				player.AvgGameTime = float64(player.TotalTime) / float64(player.TotalGames)
			}
		}

		// Update loser's metrics
		loser := event.Player1
		if event.Winner == event.Player1 {
			loser = event.Player2
		}
		if player := s.userMetrics[loser]; player != nil && !event.IsBotGame {
			player.Losses++
			player.WinStreak = 0
			player.TotalTime += event.Duration
			player.TotalMoves += event.MoveCount / 2
			player.WinRate = float64(player.Wins) / float64(player.TotalGames) * 100
			if player.TotalGames > 0 {
				player.AvgGameTime = float64(player.TotalTime) / float64(player.TotalGames)
			}
		}
	} else {
		// Draw
		s.ensureUserMetrics(event.Player1)
		if player := s.userMetrics[event.Player1]; player != nil {
			player.Draws++
			player.WinStreak = 0
			player.TotalTime += event.Duration
		}
		if !event.IsBotGame {
			s.ensureUserMetrics(event.Player2)
			if player := s.userMetrics[event.Player2]; player != nil {
				player.Draws++
				player.WinStreak = 0
				player.TotalTime += event.Duration
			}
		}
	}

	log.Printf("[Analytics] Game ended: %s (Winner: %s, Duration: %ds, Moves: %d)",
		event.GameID, event.Winner, event.Duration, event.MoveCount)
}

// ensureUserMetrics creates user metrics if not exists
func (s *AnalyticsService) ensureUserMetrics(username string) {
	if username == "" || username == "ConnectBot" {
		return
	}
	if s.userMetrics[username] == nil {
		s.userMetrics[username] = &UserMetrics{
			Username: username,
		}
	}
}

// GetStats returns current analytics statistics
func (s *AnalyticsService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	avgDuration := 0.0
	if s.completedGames > 0 {
		avgDuration = float64(s.totalDuration) / float64(s.completedGames)
	}

	avgMoves := 0.0
	if s.completedGames > 0 {
		avgMoves = float64(s.totalMoves) / float64(s.completedGames)
	}

	return map[string]interface{}{
		"totalGames":      s.totalGames,
		"completedGames":  s.completedGames,
		"botGames":        s.botGames,
		"pvpGames":        s.pvpGames,
		"avgDuration":     avgDuration,
		"avgMoves":        avgMoves,
		"topWinners":      s.getTopWinners(10),
		"gamesPerHour":    s.getGamesPerHour(),
		"gamesPerDay":     s.getGamesPerDay(),
	}
}

// GetUserMetrics returns metrics for a specific user
func (s *AnalyticsService) GetUserMetrics(username string) *UserMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.userMetrics[username]
}

// GetAllUserMetrics returns all user metrics
func (s *AnalyticsService) GetAllUserMetrics() map[string]*UserMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	result := make(map[string]*UserMetrics)
	for k, v := range s.userMetrics {
		result[k] = v
	}
	return result
}

// getTopWinners returns top N winners
func (s *AnalyticsService) getTopWinners(n int) []map[string]interface{} {
	type winnerEntry struct {
		username string
		wins     int
	}

	// Convert to slice for sorting
	winners := make([]winnerEntry, 0, len(s.winnerCounts))
	for username, wins := range s.winnerCounts {
		winners = append(winners, winnerEntry{username, wins})
	}

	// Sort by wins (descending)
	for i := 0; i < len(winners)-1; i++ {
		for j := i + 1; j < len(winners); j++ {
			if winners[j].wins > winners[i].wins {
				winners[i], winners[j] = winners[j], winners[i]
			}
		}
	}

	// Take top N
	if len(winners) > n {
		winners = winners[:n]
	}

	// Convert to response format
	result := make([]map[string]interface{}, len(winners))
	for i, w := range winners {
		result[i] = map[string]interface{}{
			"rank":     i + 1,
			"username": w.username,
			"wins":     w.wins,
		}
	}

	return result
}

// getGamesPerHour returns games count per hour
func (s *AnalyticsService) getGamesPerHour() []HourlyStats {
	stats := make([]HourlyStats, 24)
	for i := 0; i < 24; i++ {
		stats[i] = HourlyStats{
			Hour:  i,
			Count: s.gamesPerHour[i],
		}
	}
	return stats
}

// getGamesPerDay returns games count per day (last 7 days)
func (s *AnalyticsService) getGamesPerDay() []DailyStats {
	var stats []DailyStats

	// Get last 7 days
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		stats = append(stats, DailyStats{
			Date:  date,
			Count: s.gamesPerDay[date],
		})
	}

	return stats
}

// GetRecentEvents returns recent events for debugging
func (s *AnalyticsService) GetRecentEvents(n int) []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n > len(s.eventLog) {
		n = len(s.eventLog)
	}

	// Return last N events (most recent first)
	result := make([]Event, n)
	for i := 0; i < n; i++ {
		result[i] = s.eventLog[len(s.eventLog)-1-i]
	}

	return result
}
