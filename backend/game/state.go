package game

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// GameStatus represents the current state of a game
type GameStatus string

const (
	StatusWaiting    GameStatus = "waiting"
	StatusInProgress GameStatus = "in_progress"
	StatusCompleted  GameStatus = "completed"
	StatusForfeited  GameStatus = "forfeited"
)

// Player represents a game player
type Player struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	IsBot    bool   `json:"isBot"`
}

// Move represents a single game move
type Move struct {
	Player    int       `json:"player"`
	Column    int       `json:"column"`
	Row       int       `json:"row"`
	Timestamp time.Time `json:"timestamp"`
}

// Game represents a single game session
type Game struct {
	ID              string     `json:"id"`
	Player1         *Player    `json:"player1"`
	Player2         *Player    `json:"player2"`
	Board           Board      `json:"board"`
	CurrentPlayer   int        `json:"currentPlayer"` // 1 or 2
	Status          GameStatus `json:"status"`
	Winner          int        `json:"winner"` // 0=none/draw, 1=player1, 2=player2
	Moves           []Move     `json:"moves"`
	CreatedAt       time.Time  `json:"createdAt"`
	CompletedAt     *time.Time `json:"completedAt"`
	IsBotGame       bool       `json:"isBotGame"`
	DisconnectedAt  map[int]*time.Time `json:"-"` // Track disconnection time per player
	mu              sync.RWMutex
}

// NewGame creates a new game session
func NewGame(player1 *Player) *Game {
	return &Game{
		ID:             uuid.New().String(),
		Player1:        player1,
		Board:          NewBoard(),
		CurrentPlayer:  1,
		Status:         StatusWaiting,
		Moves:          make([]Move, 0),
		CreatedAt:      time.Now(),
		DisconnectedAt: make(map[int]*time.Time),
	}
}

// AddPlayer2 adds the second player to the game
func (g *Game) AddPlayer2(player2 *Player) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Player2 = player2
	g.IsBotGame = player2.IsBot
	g.Status = StatusInProgress
}

// MakeMove attempts to make a move for the current player
func (g *Game) MakeMove(col int) (row int, success bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Status != StatusInProgress {
		return -1, false
	}

	row = g.Board.DropDisc(col, g.CurrentPlayer)
	if row == -1 {
		return -1, false
	}

	// Record the move
	g.Moves = append(g.Moves, Move{
		Player:    g.CurrentPlayer,
		Column:    col,
		Row:       row,
		Timestamp: time.Now(),
	})

	// Check for win
	if g.Board.CheckWin(g.CurrentPlayer) {
		g.Winner = g.CurrentPlayer
		g.Status = StatusCompleted
		now := time.Now()
		g.CompletedAt = &now
		return row, true
	}

	// Check for draw
	if g.Board.IsFull() {
		g.Status = StatusCompleted
		now := time.Now()
		g.CompletedAt = &now
		return row, true
	}

	// Switch player
	if g.CurrentPlayer == 1 {
		g.CurrentPlayer = 2
	} else {
		g.CurrentPlayer = 1
	}

	return row, true
}

// GetState returns a thread-safe copy of the game state
func (g *Game) GetState() GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	state := GameState{
		ID:            g.ID,
		Board:         g.Board,
		CurrentPlayer: g.CurrentPlayer,
		Status:        g.Status,
		Winner:        g.Winner,
		IsBotGame:     g.IsBotGame,
		MoveCount:     len(g.Moves),
	}

	if g.Player1 != nil {
		state.Player1Username = g.Player1.Username
	}
	if g.Player2 != nil {
		state.Player2Username = g.Player2.Username
	}

	return state
}

// GameState is a snapshot of game state for transmission
type GameState struct {
	ID              string     `json:"id"`
	Board           Board      `json:"board"`
	CurrentPlayer   int        `json:"currentPlayer"`
	Status          GameStatus `json:"status"`
	Winner          int        `json:"winner"`
	Player1Username string     `json:"player1Username"`
	Player2Username string     `json:"player2Username"`
	IsBotGame       bool       `json:"isBotGame"`
	MoveCount       int        `json:"moveCount"`
}

// SetDisconnected marks a player as disconnected
func (g *Game) SetDisconnected(playerNum int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := time.Now()
	g.DisconnectedAt[playerNum] = &now
}

// SetReconnected clears the disconnection time for a player
func (g *Game) SetReconnected(playerNum int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.DisconnectedAt, playerNum)
}

// CheckDisconnectionTimeout checks if a player has been disconnected too long
func (g *Game) CheckDisconnectionTimeout(timeout time.Duration) (forfeited bool, forfeitedPlayer int) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for playerNum, disconnectTime := range g.DisconnectedAt {
		if disconnectTime != nil && time.Since(*disconnectTime) > timeout {
			return true, playerNum
		}
	}
	return false, 0
}

// Forfeit marks the game as forfeited
func (g *Game) Forfeit(losingPlayer int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.Status = StatusForfeited
	if losingPlayer == 1 {
		g.Winner = 2
	} else {
		g.Winner = 1
	}
	now := time.Now()
	g.CompletedAt = &now
}

// GetDuration returns the game duration in seconds
func (g *Game) GetDuration() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.CompletedAt != nil {
		return int(g.CompletedAt.Sub(g.CreatedAt).Seconds())
	}
	return int(time.Since(g.CreatedAt).Seconds())
}

// GetPlayer returns the player struct for given player number
func (g *Game) GetPlayer(playerNum int) *Player {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if playerNum == 1 {
		return g.Player1
	}
	return g.Player2
}
