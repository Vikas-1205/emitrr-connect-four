package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/vikasgurjar/connect-four/game"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// MessageType defines WebSocket message types
type MessageType string

const (
	MsgJoinQueue     MessageType = "join_queue"
	MsgGameStart     MessageType = "game_start"
	MsgMakeMove      MessageType = "make_move"
	MsgMoveMade      MessageType = "move_made"
	MsgGameOver      MessageType = "game_over"
	MsgError         MessageType = "error"
	MsgWaiting       MessageType = "waiting"
	MsgOpponentLeft  MessageType = "opponent_left"
	MsgReconnect     MessageType = "reconnect"
	MsgReconnected   MessageType = "reconnected"
	MsgBotMove       MessageType = "bot_move"
)

// WSMessage is the WebSocket message structure
type WSMessage struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Client represents a connected WebSocket client
type Client struct {
	ID         string
	Username   string
	Conn       *websocket.Conn
	GameID     string
	PlayerNum  int // 1 or 2
	Difficulty string
	Hub        *Hub
	mu         sync.Mutex
}

// Hub manages all game connections and matchmaking
type Hub struct {
	clients      map[string]*Client           // clientID -> client
	games        map[string]*game.Game        // gameID -> game
	waitingQueue chan *Client                 // Players waiting for match
	userGames    map[string]string            // username -> gameID (for reconnection)
	analytics    chan AnalyticsEvent
	mu           sync.RWMutex
}

// NewHub creates a new connection hub
func NewHub() *Hub {
	hub := &Hub{
		clients:      make(map[string]*Client),
		games:        make(map[string]*game.Game),
		waitingQueue: make(chan *Client, 100),
		userGames:    make(map[string]string),
		analytics:    make(chan AnalyticsEvent, 1000),
	}
	go hub.runMatchmaking()
	go hub.runDisconnectionChecker()
	return hub
}

// GetAnalyticsChannel returns the analytics event channel
func (h *Hub) GetAnalyticsChannel() chan AnalyticsEvent {
	return h.analytics
}

// HandleWebSocket handles new WebSocket connections
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		ID:   uuid.New().String(),
		Conn: conn,
		Hub:  h,
	}

	h.mu.Lock()
	h.clients[client.ID] = client
	h.mu.Unlock()

	go client.readPump()
}

// readPump handles incoming messages from a client
func (c *Client) readPump() {
	defer func() {
		c.Hub.handleDisconnect(c)
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			c.sendError("Invalid message format")
			continue
		}

		c.handleMessage(msg)
	}
}

// handleMessage processes different message types
func (c *Client) handleMessage(msg WSMessage) {
	switch msg.Type {
	case MsgJoinQueue:
		c.handleJoinQueue(msg.Payload)
	case MsgMakeMove:
		c.handleMakeMove(msg.Payload)
	case MsgReconnect:
		c.handleReconnect(msg.Payload)
	default:
		c.sendError("Unknown message type")
	}
}

// JoinQueuePayload is the payload for join_queue message
type JoinQueuePayload struct {
	Username   string `json:"username"`
	Difficulty string `json:"difficulty"`
}

// handleJoinQueue processes a player joining the matchmaking queue
func (c *Client) handleJoinQueue(payload json.RawMessage) {
	var data JoinQueuePayload
	if err := json.Unmarshal(payload, &data); err != nil || data.Username == "" {
		c.sendError("Invalid username")
		return
	}

	c.Username = data.Username
	// Store difficulty preference (default to medium if not specified)
	if data.Difficulty == "" {
		data.Difficulty = "medium"
	}
	c.Difficulty = data.Difficulty

	// Check if player has an active game to reconnect to
	c.Hub.mu.RLock()
	existingGameID, hasGame := c.Hub.userGames[data.Username]
	c.Hub.mu.RUnlock()

	if hasGame {
		c.Hub.mu.RLock()
		g, exists := c.Hub.games[existingGameID]
		c.Hub.mu.RUnlock()

		if exists && g.Status == game.StatusInProgress {
			// Reconnect to existing game
			c.reconnectToGame(g)
			return
		}
	}

	// Add to waiting queue
	c.sendMessage(MsgWaiting, map[string]string{"message": "Waiting for opponent..."})
	c.Hub.waitingQueue <- c
}

// MakeMovePayload is the payload for make_move message
type MakeMovePayload struct {
	Column int `json:"column"`
}

// handleMakeMove processes a player's move
func (c *Client) handleMakeMove(payload json.RawMessage) {
	var data MakeMovePayload
	if err := json.Unmarshal(payload, &data); err != nil {
		c.sendError("Invalid move data")
		return
	}

	if c.GameID == "" {
		c.sendError("Not in a game")
		return
	}

	c.Hub.mu.RLock()
	g, exists := c.Hub.games[c.GameID]
	c.Hub.mu.RUnlock()

	if !exists {
		c.sendError("Game not found")
		return
	}

	// Verify it's the player's turn
	if g.CurrentPlayer != c.PlayerNum {
		c.sendError("Not your turn")
		return
	}

	// Make the move
	row, success := g.MakeMove(data.Column)
	if !success {
		c.sendError("Invalid move")
		return
	}

	// Broadcast move to both players
	moveData := map[string]interface{}{
		"column":        data.Column,
		"row":           row,
		"player":        c.PlayerNum,
		"currentPlayer": g.CurrentPlayer,
	}
	c.Hub.broadcastToGame(c.GameID, MsgMoveMade, moveData)

	// Check if game is over
	state := g.GetState()
	if state.Status == game.StatusCompleted {
		c.Hub.handleGameEnd(g)
		return
	}

	// If it's a bot game and it's the bot's turn, make the bot move
	if g.IsBotGame && g.CurrentPlayer == 2 {
		go c.Hub.makeBotMove(g)
	}
}

// ReconnectPayload is the payload for reconnect message
type ReconnectPayload struct {
	Username string `json:"username"`
	GameID   string `json:"gameId"`
}

// handleReconnect processes a player reconnecting to an existing game
func (c *Client) handleReconnect(payload json.RawMessage) {
	var data ReconnectPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		c.sendError("Invalid reconnect data")
		return
	}

	c.Username = data.Username

	c.Hub.mu.RLock()
	gameID, hasGame := c.Hub.userGames[data.Username]
	c.Hub.mu.RUnlock()

	if !hasGame && data.GameID != "" {
		gameID = data.GameID
	}

	if gameID == "" {
		c.sendError("No active game found")
		return
	}

	c.Hub.mu.RLock()
	g, exists := c.Hub.games[gameID]
	c.Hub.mu.RUnlock()

	if !exists || g.Status != game.StatusInProgress {
		c.sendError("Game no longer available")
		return
	}

	c.reconnectToGame(g)
}

// reconnectToGame reconnects a client to an existing game
func (c *Client) reconnectToGame(g *game.Game) {
	// Determine player number
	if g.Player1 != nil && g.Player1.Username == c.Username {
		c.PlayerNum = 1
	} else if g.Player2 != nil && g.Player2.Username == c.Username {
		c.PlayerNum = 2
	} else {
		c.sendError("You are not a player in this game")
		return
	}

	c.GameID = g.ID
	g.SetReconnected(c.PlayerNum)

	// Send current game state
	c.sendMessage(MsgReconnected, g.GetState())

	// Notify opponent
	c.Hub.notifyOpponent(c, "Opponent reconnected")
}

// runMatchmaking handles the matchmaking queue
func (h *Hub) runMatchmaking() {
	for {
		// Wait for first player
		player1 := <-h.waitingQueue

		// Check if player1 is still connected
		if player1.Conn == nil {
			continue
		}

		// Wait for second player with 10-second timeout
		matched := false
		timeout := time.After(10 * time.Second)

	matchLoop:
		for !matched {
			select {
			case player2 := <-h.waitingQueue:
				// Check if player2 is still connected
				if player2.Conn == nil {
					continue
				}

				// Make sure we're not matching a player with themselves
				if player1.ID == player2.ID || player1.Username == player2.Username {
					// Put player2 back in queue and continue waiting
					go func(p *Client) {
						h.waitingQueue <- p
					}(player2)
					continue
				}

				// Match found! Create a game between two players
				h.createGame(player1, player2, false)
				matched = true
				break matchLoop

			case <-timeout:
				// No opponent found, start game with bot
				h.createBotGame(player1)
				matched = true
				break matchLoop
			}
		}
	}
}

// createGame creates a new game between two players
func (h *Hub) createGame(p1, p2 *Client, isBotGame bool) {
	player1 := &game.Player{
		ID:       p1.ID,
		Username: p1.Username,
		IsBot:    false,
	}

	g := game.NewGame(player1)

	var player2 *game.Player
	if isBotGame {
		player2 = &game.Player{
			ID:       "bot",
			Username: "ConnectBot",
			IsBot:    true,
		}
	} else {
		player2 = &game.Player{
			ID:       p2.ID,
			Username: p2.Username,
			IsBot:    false,
		}
	}

	g.AddPlayer2(player2)

	// Store game
	h.mu.Lock()
	h.games[g.ID] = g
	h.userGames[p1.Username] = g.ID
	if !isBotGame {
		h.userGames[p2.Username] = g.ID
	}
	h.mu.Unlock()

	// Associate clients with game
	p1.GameID = g.ID
	p1.PlayerNum = 1

	if !isBotGame {
		p2.GameID = g.ID
		p2.PlayerNum = 2
	}

	// Send game start to both players
	gameState := g.GetState()
	gameState.Player1Username = p1.Username
	if isBotGame {
		gameState.Player2Username = "ConnectBot"
	} else {
		gameState.Player2Username = p2.Username
	}

	startPayload := map[string]interface{}{
		"gameId":    g.ID,
		"yourTurn":  true,
		"playerNum": 1,
		"state":     gameState,
	}
	p1.sendMessage(MsgGameStart, startPayload)

	if !isBotGame {
		startPayload["yourTurn"] = false
		startPayload["playerNum"] = 2
		p2.sendMessage(MsgGameStart, startPayload)
	}

	// Send analytics event
	h.analytics <- AnalyticsEvent{
		Type:      "game_started",
		GameID:    g.ID,
		Player1:   p1.Username,
		Player2:   gameState.Player2Username,
		IsBotGame: isBotGame,
		Timestamp: time.Now(),
	}
}

// createBotGame creates a game with a bot opponent
func (h *Hub) createBotGame(player *Client) {
	h.createGame(player, nil, true)
}

// makeBotMove executes the bot's move
func (h *Hub) makeBotMove(g *game.Game) {
	h.makeBotMoveWithDifficulty(g, "medium")
}

// makeBotMoveWithDifficulty executes the bot's move with specified difficulty
func (h *Hub) makeBotMoveWithDifficulty(g *game.Game, difficulty string) {
	// Short delay to make it feel more natural
	time.Sleep(500 * time.Millisecond)

	bot := game.NewBotWithDifficulty(2, difficulty)
	col := bot.GetMove(g.Board)

	row, success := g.MakeMove(col)
	if !success {
		log.Printf("Bot failed to make move in game %s", g.ID)
		return
	}

	// Broadcast bot's move
	moveData := map[string]interface{}{
		"column":        col,
		"row":           row,
		"player":        2,
		"currentPlayer": g.CurrentPlayer,
		"isBot":         true,
	}
	h.broadcastToGame(g.ID, MsgMoveMade, moveData)

	// Check if game is over
	if g.Status == game.StatusCompleted {
		h.handleGameEnd(g)
	}
}

// handleGameEnd processes game completion
func (h *Hub) handleGameEnd(g *game.Game) {
	state := g.GetState()

	// Get winning cells if there's a winner
	var winningCells []game.WinningCell
	if state.Winner > 0 {
		winningCells = g.Board.GetWinningCells(state.Winner)
	}

	result := map[string]interface{}{
		"winner":       state.Winner,
		"status":       state.Status,
		"duration":     g.GetDuration(),
		"winningCells": winningCells,
	}

	h.broadcastToGame(g.ID, MsgGameOver, result)

	// Send analytics event
	winnerName := ""
	if state.Winner == 1 && g.Player1 != nil {
		winnerName = g.Player1.Username
	} else if state.Winner == 2 && g.Player2 != nil {
		winnerName = g.Player2.Username
	}

	h.analytics <- AnalyticsEvent{
		Type:      "game_ended",
		GameID:    g.ID,
		Winner:    winnerName,
		Duration:  g.GetDuration(),
		MoveCount: state.MoveCount,
		IsBotGame: state.IsBotGame,
		Timestamp: time.Now(),
	}

	// Immediately clean up user-game mapping so players can start new games
	h.mu.Lock()
	if g.Player1 != nil {
		delete(h.userGames, g.Player1.Username)
	}
	if g.Player2 != nil && !g.Player2.IsBot {
		delete(h.userGames, g.Player2.Username)
	}
	h.mu.Unlock()

	// Also clear the gameID from connected clients
	h.mu.RLock()
	for _, client := range h.clients {
		if client.GameID == g.ID {
			client.GameID = ""
		}
	}
	h.mu.RUnlock()
}

// handleDisconnect handles a client disconnecting
func (h *Hub) handleDisconnect(c *Client) {
	h.mu.Lock()
	delete(h.clients, c.ID)
	h.mu.Unlock()

	if c.GameID == "" {
		return
	}

	h.mu.RLock()
	g, exists := h.games[c.GameID]
	h.mu.RUnlock()

	if !exists || g.Status != game.StatusInProgress {
		return
	}

	// Mark player as disconnected
	g.SetDisconnected(c.PlayerNum)

	// Notify opponent
	h.notifyOpponent(c, "Opponent disconnected. Waiting 30 seconds for reconnection...")
}

// runDisconnectionChecker periodically checks for disconnection timeouts
func (h *Hub) runDisconnectionChecker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.RLock()
		games := make([]*game.Game, 0, len(h.games))
		for _, g := range h.games {
			if g.Status == game.StatusInProgress {
				games = append(games, g)
			}
		}
		h.mu.RUnlock()

		for _, g := range games {
			if forfeited, player := g.CheckDisconnectionTimeout(30 * time.Second); forfeited {
				g.Forfeit(player)
				h.handleGameEnd(g)
			}
		}
	}
}

// broadcastToGame sends a message to all clients in a game
func (h *Hub) broadcastToGame(gameID string, msgType MessageType, payload interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.GameID == gameID {
			client.sendMessage(msgType, payload)
		}
	}
}

// notifyOpponent sends a message to the opponent
func (h *Hub) notifyOpponent(c *Client, message string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.GameID == c.GameID && client.ID != c.ID {
			client.sendMessage(MsgOpponentLeft, map[string]string{"message": message})
		}
	}
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(msgType MessageType, payload interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, _ := json.Marshal(payload)
	msg := WSMessage{
		Type:    msgType,
		Payload: data,
	}

	msgBytes, _ := json.Marshal(msg)
	c.Conn.WriteMessage(websocket.TextMessage, msgBytes)
}

// sendError sends an error message to the client
func (c *Client) sendError(errMsg string) {
	c.sendMessage(MsgError, map[string]string{"error": errMsg})
}

// AnalyticsEvent represents an analytics event
type AnalyticsEvent struct {
	Type      string    `json:"type"`
	GameID    string    `json:"gameId"`
	Player1   string    `json:"player1,omitempty"`
	Player2   string    `json:"player2,omitempty"`
	Winner    string    `json:"winner,omitempty"`
	Duration  int       `json:"duration,omitempty"`
	MoveCount int       `json:"moveCount,omitempty"`
	IsBotGame bool      `json:"isBotGame"`
	Timestamp time.Time `json:"timestamp"`
}

// GetGames returns all games for leaderboard/stats
func (h *Hub) GetGames() []*game.Game {
	h.mu.RLock()
	defer h.mu.RUnlock()

	games := make([]*game.Game, 0, len(h.games))
	for _, g := range h.games {
		games = append(games, g)
	}
	return games
}
