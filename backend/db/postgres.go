package db

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// Database wraps the SQL database connection
type Database struct {
	db *sql.DB
}

// PlayerStats represents a player's statistics
type PlayerStats struct {
	Username  string    `json:"username"`
	Wins      int       `json:"wins"`
	Losses    int       `json:"losses"`
	Draws     int       `json:"draws"`
	CreatedAt time.Time `json:"createdAt"`
}

// GameRecord represents a completed game record
type GameRecord struct {
	ID           string    `json:"id"`
	Player1      string    `json:"player1"`
	Player2      string    `json:"player2"`
	Winner       string    `json:"winner"`
	IsBotGame    bool      `json:"isBotGame"`
	Duration     int       `json:"duration"`
	Moves        []byte    `json:"moves"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	CompletedAt  time.Time `json:"completedAt"`
}

// NewDatabase creates a new database connection
func NewDatabase(connStr string) (*Database, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	d := &Database{db: db}
	if err := d.migrate(); err != nil {
		return nil, err
	}

	return d, nil
}

// migrate creates the necessary tables
func (d *Database) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS players (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		wins INT DEFAULT 0,
		losses INT DEFAULT 0,
		draws INT DEFAULT 0,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS games (
		id VARCHAR(36) PRIMARY KEY,
		player1 VARCHAR(50) NOT NULL,
		player2 VARCHAR(50) NOT NULL,
		winner VARCHAR(50),
		is_bot_game BOOLEAN DEFAULT FALSE,
		duration_seconds INT,
		moves JSONB,
		status VARCHAR(20),
		created_at TIMESTAMP DEFAULT NOW(),
		completed_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_players_wins ON players(wins DESC);
	CREATE INDEX IF NOT EXISTS idx_games_created ON games(created_at);
	`

	_, err := d.db.Exec(schema)
	if err != nil {
		log.Printf("Migration warning: %v", err)
	}
	return nil
}

// EnsurePlayer creates a player if they don't exist
func (d *Database) EnsurePlayer(username string) error {
	query := `
		INSERT INTO players (username) VALUES ($1)
		ON CONFLICT (username) DO NOTHING
	`
	_, err := d.db.Exec(query, username)
	return err
}

// UpdatePlayerStats updates a player's win/loss/draw counts
func (d *Database) UpdatePlayerStats(username string, result string) error {
	var query string
	switch result {
	case "win":
		query = `UPDATE players SET wins = wins + 1 WHERE username = $1`
	case "loss":
		query = `UPDATE players SET losses = losses + 1 WHERE username = $1`
	case "draw":
		query = `UPDATE players SET draws = draws + 1 WHERE username = $1`
	default:
		return nil
	}

	_, err := d.db.Exec(query, username)
	return err
}

// GetLeaderboard returns the top players by wins
func (d *Database) GetLeaderboard(limit int) ([]PlayerStats, error) {
	query := `
		SELECT username, wins, losses, draws, created_at
		FROM players
		ORDER BY wins DESC, losses ASC
		LIMIT $1
	`

	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []PlayerStats
	for rows.Next() {
		var p PlayerStats
		if err := rows.Scan(&p.Username, &p.Wins, &p.Losses, &p.Draws, &p.CreatedAt); err != nil {
			continue
		}
		players = append(players, p)
	}

	return players, rows.Err()
}

// SaveGame saves a completed game record
func (d *Database) SaveGame(record GameRecord) error {
	query := `
		INSERT INTO games (id, player1, player2, winner, is_bot_game, duration_seconds, moves, status, created_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	movesJSON, _ := json.Marshal(record.Moves)

	_, err := d.db.Exec(query,
		record.ID,
		record.Player1,
		record.Player2,
		record.Winner,
		record.IsBotGame,
		record.Duration,
		movesJSON,
		record.Status,
		record.CreatedAt,
		record.CompletedAt,
	)

	return err
}

// GetGameStats returns aggregate game statistics
func (d *Database) GetGameStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total games
	var totalGames int
	d.db.QueryRow(`SELECT COUNT(*) FROM games`).Scan(&totalGames)
	stats["totalGames"] = totalGames

	// Average duration
	var avgDuration float64
	d.db.QueryRow(`SELECT COALESCE(AVG(duration_seconds), 0) FROM games WHERE duration_seconds > 0`).Scan(&avgDuration)
	stats["avgDuration"] = avgDuration

	// Bot games vs player games
	var botGames int
	d.db.QueryRow(`SELECT COUNT(*) FROM games WHERE is_bot_game = true`).Scan(&botGames)
	stats["botGames"] = botGames
	stats["playerGames"] = totalGames - botGames

	// Games today
	var gamesToday int
	d.db.QueryRow(`SELECT COUNT(*) FROM games WHERE created_at >= CURRENT_DATE`).Scan(&gamesToday)
	stats["gamesToday"] = gamesToday

	return stats, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}
