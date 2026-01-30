package game

import (
	"testing"
)

func TestNewBoard(t *testing.T) {
	board := NewBoard()

	// Check dimensions
	if len(board) != Rows {
		t.Errorf("Expected %d rows, got %d", Rows, len(board))
	}

	for i, row := range board {
		if len(row) != Cols {
			t.Errorf("Row %d: expected %d columns, got %d", i, Cols, len(row))
		}
	}

	// Check all cells are empty
	for i, row := range board {
		for j, cell := range row {
			if cell != Empty {
				t.Errorf("Cell [%d][%d] should be Empty (0), got %d", i, j, cell)
			}
		}
	}
}

func TestDropDisc(t *testing.T) {
	board := NewBoard()

	// Drop disc in column 3
	row := board.DropDisc(3, Player1)

	// Should land at bottom (row 5)
	if row != 5 {
		t.Errorf("Expected row 5, got %d", row)
	}

	// Check disc is placed
	if board[5][3] != Player1 {
		t.Errorf("Expected Player1 disc at [5][3], got %d", board[5][3])
	}

	// Drop another disc in same column
	row = board.DropDisc(3, Player2)

	// Should land at row 4
	if row != 4 {
		t.Errorf("Expected row 4, got %d", row)
	}

	// Check disc is placed
	if board[4][3] != Player2 {
		t.Errorf("Expected Player2 disc at [4][3], got %d", board[4][3])
	}
}

func TestDropDiscFullColumn(t *testing.T) {
	board := NewBoard()

	// Fill column 0
	for i := 0; i < Rows; i++ {
		row := board.DropDisc(0, Player1)
		if row < 0 {
			t.Errorf("Unexpected full column on drop %d", i)
		}
	}

	// Try to drop in full column
	row := board.DropDisc(0, Player2)
	if row != -1 {
		t.Errorf("Expected -1 for full column, got %d", row)
	}
}

func TestDropDiscInvalidColumn(t *testing.T) {
	board := NewBoard()

	// Test invalid columns
	row := board.DropDisc(-1, Player1)
	if row != -1 {
		t.Errorf("Expected -1 for column -1, got %d", row)
	}

	row = board.DropDisc(7, Player1)
	if row != -1 {
		t.Errorf("Expected -1 for column 7, got %d", row)
	}
}

func TestIsValidMove(t *testing.T) {
	board := NewBoard()

	// Empty column should be valid
	if !board.IsValidMove(3) {
		t.Error("Empty column 3 should be valid")
	}

	// Fill column 3
	for i := 0; i < Rows; i++ {
		board.DropDisc(3, Player1)
	}

	// Full column should be invalid
	if board.IsValidMove(3) {
		t.Error("Full column 3 should be invalid")
	}

	// Invalid column numbers
	if board.IsValidMove(-1) {
		t.Error("Column -1 should be invalid")
	}
	if board.IsValidMove(7) {
		t.Error("Column 7 should be invalid")
	}
}

func TestGetValidMoves(t *testing.T) {
	board := NewBoard()

	// All columns should be valid initially
	moves := board.GetValidMoves()
	if len(moves) != Cols {
		t.Errorf("Expected %d valid moves, got %d", Cols, len(moves))
	}

	// Fill column 0 and 6
	for i := 0; i < Rows; i++ {
		board.DropDisc(0, Player1)
		board.DropDisc(6, Player2)
	}

	moves = board.GetValidMoves()
	if len(moves) != 5 {
		t.Errorf("Expected 5 valid moves after filling 2 columns, got %d", len(moves))
	}
}

func TestCheckWinHorizontal(t *testing.T) {
	board := NewBoard()

	// Create horizontal win at bottom
	board[5][0] = Player1
	board[5][1] = Player1
	board[5][2] = Player1
	board[5][3] = Player1

	if !board.CheckWin(Player1) {
		t.Error("Expected horizontal win for Player1")
	}

	if board.CheckWin(Player2) {
		t.Error("Should not be a win for Player2")
	}
}

func TestCheckWinVertical(t *testing.T) {
	board := NewBoard()

	// Create vertical win
	board[5][0] = Player1
	board[4][0] = Player1
	board[3][0] = Player1
	board[2][0] = Player1

	if !board.CheckWin(Player1) {
		t.Error("Expected vertical win for Player1")
	}
}

func TestCheckWinDiagonal(t *testing.T) {
	board := NewBoard()

	// Create diagonal win (bottom-left to top-right)
	board[5][0] = Player1
	board[4][1] = Player1
	board[3][2] = Player1
	board[2][3] = Player1

	if !board.CheckWin(Player1) {
		t.Error("Expected diagonal win for Player1")
	}
}

func TestCheckWinAntiDiagonal(t *testing.T) {
	board := NewBoard()

	// Create anti-diagonal win (top-left to bottom-right)
	board[2][0] = Player1
	board[3][1] = Player1
	board[4][2] = Player1
	board[5][3] = Player1

	if !board.CheckWin(Player1) {
		t.Error("Expected anti-diagonal win for Player1")
	}
}

func TestIsFull(t *testing.T) {
	board := NewBoard()

	if board.IsFull() {
		t.Error("Empty board should not be full")
	}

	// Fill the board
	for i := 0; i < Rows; i++ {
		for j := 0; j < Cols; j++ {
			board[i][j] = Player1
		}
	}

	if !board.IsFull() {
		t.Error("Full board should be marked as full")
	}
}

func TestNoWinWithThreeInARow(t *testing.T) {
	board := NewBoard()

	// Only 3 in a row - should not be a win
	board[5][0] = Player1
	board[5][1] = Player1
	board[5][2] = Player1

	if board.CheckWin(Player1) {
		t.Error("Three in a row should not be a win")
	}
}

func TestCopyBoard(t *testing.T) {
	board := NewBoard()
	board[5][3] = Player1
	board[4][3] = Player2

	copyBoard := board.Copy()

	// Verify copy has same values
	if copyBoard[5][3] != Player1 || copyBoard[4][3] != Player2 {
		t.Error("Copy should have same values as original")
	}

	// Modify copy and ensure original is unchanged
	copyBoard[3][3] = Player1
	if board[3][3] != Empty {
		t.Error("Modifying copy should not affect original")
	}
}
