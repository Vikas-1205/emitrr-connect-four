# 4 in a Row — Real-Time Multiplayer Game

A real-time, backend-driven Connect Four game with player matchmaking, competitive AI bot, and live leaderboard.

## 🎮 Features

### Core Features
- **Real-time Multiplayer** — Play against other players via WebSocket
- **Smart AI Bot** — Minimax algorithm with alpha-beta pruning (depth 6)
- **Auto-Matchmaking** — 10-second timeout, then bot takes over
- **Reconnection Support** — 30-second window to rejoin if disconnected
- **Leaderboard** — Track wins across all players
- **Kafka Analytics** — Game events streamed for analytics (bonus)

### Enhanced Features
- **Bot Difficulty Levels** — Easy, Medium, Hard (user selectable)
- **Sound Effects** — Disc drop, win/lose, turn notification
- **Move Timer** — 30-second countdown per turn
- **Confetti Celebration** — Win animation
- **Player Avatars** — Auto-generated based on username
- **Dark/Light Theme** — Toggle in header
- **Game History** — View recent games
- **Winning Line Highlight** — Shows the 4 winning discs

## 🛠 Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | **Go (Golang)** |
| Real-time | **WebSocket** (gorilla/websocket) |
| Database | **PostgreSQL** (optional) |
| Analytics | **Kafka** (optional) |
| Frontend | **React + Vite** |

## 📁 Project Structure

```
emitrr/
├── backend/
│   ├── main.go              # Server entry point
│   ├── go.mod               # Go modules
│   ├── game/
│   │   ├── engine.go        # Game logic (board, win detection)
│   │   ├── state.go         # Game state management
│   │   └── bot.go           # Minimax AI bot with difficulty levels
│   ├── server/
│   │   └── websocket.go     # WebSocket hub & matchmaking
│   ├── db/
│   │   └── postgres.go      # Database layer
│   ├── analytics/
│   │   └── kafka.go         # Kafka producer/consumer & analytics service
│   └── cmd/analytics/
│       └── main.go          # Standalone analytics consumer
├── frontend/
│   ├── package.json
│   ├── vite.config.js
│   ├── index.html
│   └── src/
│       ├── App.jsx          # Main application
│       ├── index.css        # Styling (dark/light themes)
│       ├── hooks/
│       │   └── useWebSocket.js
│       ├── components/
│       │   ├── GameBoard.jsx
│       │   ├── LobbyScreen.jsx
│       │   ├── WaitingScreen.jsx
│       │   ├── GameInfo.jsx
│       │   ├── GameOverModal.jsx
│       │   ├── Leaderboard.jsx
│       │   ├── MoveTimer.jsx
│       │   ├── ThemeToggle.jsx
│       │   ├── SoundToggle.jsx
│       │   └── GameHistory.jsx
│       ├── utils/
│       │   ├── sounds.js    # Sound effects manager
│       │   ├── confetti.js  # Win celebration
│       │   └── avatars.js   # Avatar generator
│       └── context/
│           └── ThemeContext.jsx
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** — [Install Go](https://go.dev/doc/install)
- **Node.js 18+** — [Install Node.js](https://nodejs.org/)
- **PostgreSQL** (optional) — For persistent leaderboard
- **Kafka** (optional) — For analytics

### 1. Clone and Setup

```bash
git clone https://github.com/yourusername/emitrr.git
cd emitrr
```

### 2. Start Backend

```bash
cd backend

# Install dependencies
go mod tidy

# Run (works without database)
go run main.go
```

The server will start at `http://localhost:8080`

#### With Optional Services:
```bash
# With PostgreSQL
DATABASE_URL="postgres://user:pass@localhost:5432/connectfour?sslmode=disable" go run main.go

# With Kafka
KAFKA_BROKERS="localhost:9092" go run main.go
```

### 3. Start Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

Frontend will be available at `http://localhost:5173`

### 4. Play!

1. Open `http://localhost:5173` in your browser
2. Enter a username
3. Select bot difficulty (if no opponent found)
4. Wait for opponent (or bot after 10 seconds)
5. Click columns to drop your disc
6. Connect 4 to win!

## 🐳 Docker Deployment

### Using Docker Compose (Recommended)

```bash
docker-compose up --build
```

This starts:
- Backend on port 8080
- Frontend on port 80

### Manual Docker Build

```bash
# Build backend
cd backend
docker build -t connect-four-backend .
docker run -p 8080:8080 connect-four-backend

# Build frontend
cd frontend
npm run build
# Serve the dist folder with any static server
```

## 🎯 Game Rules

- **Board**: 7 columns × 6 rows
- **Objective**: Connect 4 discs in a row (horizontal, vertical, or diagonal)
- **Turns**: Players alternate, dropping one disc per turn
- **Winning**: First to connect 4 wins
- **Draw**: Board fills up with no winner

## 🤖 Bot Strategy

The AI uses **Minimax with Alpha-Beta Pruning**:

| Difficulty | Search Depth | Behavior |
|------------|--------------|----------|
| Easy | 2 moves | Makes occasional mistakes |
| Medium | 4 moves | Balanced challenge |
| Hard | 6 moves | Expert level, hard to beat |

**Priority Order:**
1. **Immediate winning move** — Takes it
2. **Block opponent's win** — Prevents it
3. **Strategic positioning** — Prefers center column

## 🔌 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/ws` | WebSocket | Game connection |
| `/api/leaderboard` | GET | Top 10 players |
| `/api/stats` | GET | Game statistics |
| `/api/games/history` | GET | Recent completed games |
| `/api/analytics` | GET | Comprehensive analytics |
| `/api/analytics/user/{username}` | GET | User-specific metrics |
| `/api/health` | GET | Server health check |

## 📡 WebSocket Messages

### Client → Server

```json
// Join matchmaking queue
{ "type": "join_queue", "payload": { "username": "player1", "difficulty": "medium" } }

// Make a move
{ "type": "make_move", "payload": { "column": 3 } }

// Reconnect to game
{ "type": "reconnect", "payload": { "username": "player1", "gameId": "..." } }
```

### Server → Client

```json
// Game started
{ "type": "game_start", "payload": { "gameId": "...", "playerNum": 1, "state": {...} } }

// Move made
{ "type": "move_made", "payload": { "column": 3, "row": 5, "player": 1 } }

// Game over (includes winning cells)
{ "type": "game_over", "payload": { "winner": 1, "winningCells": [...] } }
```

## 📊 Kafka Analytics (Bonus)

### Events Published
- `game_started` — When a match begins
- `game_ended` — Winner, duration, move count

### Analytics Tracked
- Average game duration
- Games per hour/day
- Most frequent winners
- User-specific metrics (win rate, streak)

## 📝 Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | 8080 | Server port |
| `DATABASE_URL` | No | — | PostgreSQL connection string |
| `KAFKA_BROKERS` | No | — | Comma-separated Kafka brokers |

## 🧪 Testing

```bash
cd backend
go test ./game/...   # Test game logic
go test ./server/... # Test WebSocket handling
```

## 📄 License

MIT License

---

Built with ❤️ for Emitrr Backend Engineering Internship Assignment
