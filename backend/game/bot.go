package game

import (
	"math"
	"math/rand"
)

const (
	WinScore    = 100000
	ThreeScore  = 100
	TwoScore    = 10
	CenterBonus = 3
)

// Difficulty levels for bot
const (
	DifficultyEasy   = "easy"
	DifficultyMedium = "medium"
	DifficultyHard   = "hard"
)

// Bot represents the AI player
type Bot struct {
	PlayerNum  int    // 1 or 2
	Difficulty string // easy, medium, hard
	MaxDepth   int    // Search depth based on difficulty
}

// NewBot creates a new bot for the given player number with default medium difficulty
func NewBot(playerNum int) *Bot {
	return NewBotWithDifficulty(playerNum, DifficultyMedium)
}

// NewBotWithDifficulty creates a bot with specified difficulty
func NewBotWithDifficulty(playerNum int, difficulty string) *Bot {
	depth := 4 // Default medium
	switch difficulty {
	case DifficultyEasy:
		depth = 2
	case DifficultyMedium:
		depth = 4
	case DifficultyHard:
		depth = 6
	}
	return &Bot{
		PlayerNum:  playerNum,
		Difficulty: difficulty,
		MaxDepth:   depth,
	}
}

// GetMove calculates the best move using Minimax with Alpha-Beta pruning
func (b *Bot) GetMove(board Board) int {
	// First, check for immediate winning move
	for _, col := range board.GetValidMoves() {
		testBoard := board.Copy()
		testBoard.DropDisc(col, b.PlayerNum)
		if testBoard.CheckWin(b.PlayerNum) {
			return col
		}
	}

	// Second, check if we need to block opponent's winning move
	opponent := 3 - b.PlayerNum
	for _, col := range board.GetValidMoves() {
		testBoard := board.Copy()
		testBoard.DropDisc(col, opponent)
		if testBoard.CheckWin(opponent) {
			return col
		}
	}

	// Use minimax for strategic play
	bestScore := math.MinInt32
	bestMoves := []int{}

	for _, col := range board.GetValidMoves() {
		testBoard := board.Copy()
		testBoard.DropDisc(col, b.PlayerNum)
		score := b.minimax(testBoard, b.MaxDepth-1, math.MinInt32, math.MaxInt32, false)
		
		if score > bestScore {
			bestScore = score
			bestMoves = []int{col}
		} else if score == bestScore {
			bestMoves = append(bestMoves, col)
		}
	}

	// Return one of the best moves (randomly if tie)
	if len(bestMoves) > 0 {
		return bestMoves[rand.Intn(len(bestMoves))]
	}

	// Fallback: return any valid move
	validMoves := board.GetValidMoves()
	if len(validMoves) > 0 {
		return validMoves[rand.Intn(len(validMoves))]
	}
	return -1
}

// minimax implements the Minimax algorithm with Alpha-Beta pruning
func (b *Bot) minimax(board Board, depth int, alpha, beta int, isMaximizing bool) int {
	opponent := 3 - b.PlayerNum

	// Terminal conditions
	if board.CheckWin(b.PlayerNum) {
		return WinScore + depth // Prefer faster wins
	}
	if board.CheckWin(opponent) {
		return -WinScore - depth // Delay losses
	}
	if board.IsFull() || depth == 0 {
		return b.evaluateBoard(board)
	}

	validMoves := board.GetValidMoves()
	if len(validMoves) == 0 {
		return 0
	}

	// Prioritize center column for move ordering
	orderedMoves := b.orderMoves(validMoves)

	if isMaximizing {
		maxScore := math.MinInt32
		for _, col := range orderedMoves {
			testBoard := board.Copy()
			testBoard.DropDisc(col, b.PlayerNum)
			score := b.minimax(testBoard, depth-1, alpha, beta, false)
			maxScore = max(maxScore, score)
			alpha = max(alpha, score)
			if beta <= alpha {
				break // Beta cutoff
			}
		}
		return maxScore
	} else {
		minScore := math.MaxInt32
		for _, col := range orderedMoves {
			testBoard := board.Copy()
			testBoard.DropDisc(col, opponent)
			score := b.minimax(testBoard, depth-1, alpha, beta, true)
			minScore = min(minScore, score)
			beta = min(beta, score)
			if beta <= alpha {
				break // Alpha cutoff
			}
		}
		return minScore
	}
}

// orderMoves orders moves to improve alpha-beta pruning efficiency
func (b *Bot) orderMoves(moves []int) []int {
	// Prioritize center columns
	center := Cols / 2
	ordered := make([]int, len(moves))
	copy(ordered, moves)
	
	// Simple bubble sort by distance from center
	for i := 0; i < len(ordered)-1; i++ {
		for j := 0; j < len(ordered)-i-1; j++ {
			dist1 := abs(ordered[j] - center)
			dist2 := abs(ordered[j+1] - center)
			if dist1 > dist2 {
				ordered[j], ordered[j+1] = ordered[j+1], ordered[j]
			}
		}
	}
	return ordered
}

// evaluateBoard scores the current board position
func (b *Bot) evaluateBoard(board Board) int {
	score := 0
	opponent := 3 - b.PlayerNum

	// Score center column (center control is important)
	centerCol := Cols / 2
	centerCount := 0
	for row := 0; row < Rows; row++ {
		if board[row][centerCol] == b.PlayerNum {
			centerCount++
		}
	}
	score += centerCount * CenterBonus

	// Score all windows of 4
	score += b.scoreAllWindows(board, b.PlayerNum) - b.scoreAllWindows(board, opponent)

	return score
}

// scoreAllWindows evaluates all possible 4-cell windows for a player
func (b *Bot) scoreAllWindows(board Board, player int) int {
	score := 0

	// Horizontal windows
	for row := 0; row < Rows; row++ {
		for col := 0; col <= Cols-4; col++ {
			window := []int{board[row][col], board[row][col+1], board[row][col+2], board[row][col+3]}
			score += b.scoreWindow(window, player)
		}
	}

	// Vertical windows
	for row := 0; row <= Rows-4; row++ {
		for col := 0; col < Cols; col++ {
			window := []int{board[row][col], board[row+1][col], board[row+2][col], board[row+3][col]}
			score += b.scoreWindow(window, player)
		}
	}

	// Positive diagonal windows
	for row := 3; row < Rows; row++ {
		for col := 0; col <= Cols-4; col++ {
			window := []int{board[row][col], board[row-1][col+1], board[row-2][col+2], board[row-3][col+3]}
			score += b.scoreWindow(window, player)
		}
	}

	// Negative diagonal windows
	for row := 0; row <= Rows-4; row++ {
		for col := 0; col <= Cols-4; col++ {
			window := []int{board[row][col], board[row+1][col+1], board[row+2][col+2], board[row+3][col+3]}
			score += b.scoreWindow(window, player)
		}
	}

	return score
}

// scoreWindow evaluates a single window of 4 cells
func (b *Bot) scoreWindow(window []int, player int) int {
	_ = 3 - player // opponent not needed for window scoring
	playerCount := 0
	emptyCount := 0

	for _, cell := range window {
		if cell == player {
			playerCount++
		} else if cell == Empty {
			emptyCount++
		}
	}

	// If opponent has pieces in this window, it's blocked
	if playerCount+emptyCount != 4 {
		return 0
	}

	switch playerCount {
	case 4:
		return WinScore
	case 3:
		if emptyCount == 1 {
			return ThreeScore
		}
	case 2:
		if emptyCount == 2 {
			return TwoScore
		}
	}

	return 0
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
