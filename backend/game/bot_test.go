package game

import (
	"testing"
)

func TestBotCreation(t *testing.T) {
	bot := NewBot(2)
	if bot.PlayerNum != 2 {
		t.Errorf("Expected player 2, got %d", bot.PlayerNum)
	}
	if bot.MaxDepth != 4 {
		t.Errorf("Default bot should have depth 4, got %d", bot.MaxDepth)
	}
}

func TestBotDifficultyLevels(t *testing.T) {
	// Easy bot
	easyBot := NewBotWithDifficulty(2, DifficultyEasy)
	if easyBot.MaxDepth != 2 {
		t.Errorf("Easy bot should have depth 2, got %d", easyBot.MaxDepth)
	}

	// Medium bot
	mediumBot := NewBotWithDifficulty(2, DifficultyMedium)
	if mediumBot.MaxDepth != 4 {
		t.Errorf("Medium bot should have depth 4, got %d", mediumBot.MaxDepth)
	}

	// Hard bot
	hardBot := NewBotWithDifficulty(2, DifficultyHard)
	if hardBot.MaxDepth != 6 {
		t.Errorf("Hard bot should have depth 6, got %d", hardBot.MaxDepth)
	}
}

func TestBotFindsBestMove(t *testing.T) {
	board := NewBoard()
	bot := NewBot(2)

	// Get a move - should return valid column
	col := bot.GetMove(board)
	if col < 0 || col > 6 {
		t.Errorf("Invalid column returned: %d", col)
	}
}

func TestBotBlocksWin(t *testing.T) {
	board := NewBoard()
	bot := NewBotWithDifficulty(2, DifficultyHard) // Use hard for better blocking

	// Player 1 has 3 in a row horizontally at bottom
	board[5][0] = 1
	board[5][1] = 1
	board[5][2] = 1
	// Column 3 would complete the win

	// Bot (player 2) should block at column 3
	col := bot.GetMove(board)
	if col != 3 {
		t.Errorf("Bot should block at column 3, got %d", col)
	}
}

func TestBotTakesWin(t *testing.T) {
	board := NewBoard()
	bot := NewBot(2)

	// Bot (player 2) has 3 in a row horizontally
	board[5][0] = 2
	board[5][1] = 2
	board[5][2] = 2
	// Column 3 would complete the win for bot

	// Bot should take the winning move
	col := bot.GetMove(board)
	if col != 3 {
		t.Errorf("Bot should win at column 3, got %d", col)
	}
}

func TestBotHandlesNearlyFullBoard(t *testing.T) {
	board := NewBoard()
	bot := NewBot(2)

	// Fill most columns
	for col := 0; col < 6; col++ {
		for row := 0; row < 6; row++ {
			board[row][col] = 1
		}
	}

	// Only column 6 has space
	col := bot.GetMove(board)
	if col != 6 {
		t.Errorf("Bot should choose only available column 6, got %d", col)
	}
}
