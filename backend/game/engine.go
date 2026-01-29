package game

const (
	Rows    = 6
	Cols    = 7
	Empty   = 0
	Player1 = 1
	Player2 = 2
)

// Board represents the 7x6 game grid
type Board [Rows][Cols]int

// NewBoard creates an empty game board
func NewBoard() Board {
	return Board{}
}

// DropDisc attempts to drop a disc in the given column
// Returns the row where the disc landed, or -1 if column is full
func (b *Board) DropDisc(col int, player int) int {
	if col < 0 || col >= Cols {
		return -1
	}
	
	// Find the lowest empty row in the column
	for row := Rows - 1; row >= 0; row-- {
		if b[row][col] == Empty {
			b[row][col] = player
			return row
		}
	}
	return -1 // Column is full
}

// IsValidMove checks if a column has space for another disc
func (b *Board) IsValidMove(col int) bool {
	if col < 0 || col >= Cols {
		return false
	}
	return b[0][col] == Empty
}

// GetValidMoves returns all columns that can accept a disc
func (b *Board) GetValidMoves() []int {
	moves := make([]int, 0, Cols)
	for col := 0; col < Cols; col++ {
		if b.IsValidMove(col) {
			moves = append(moves, col)
		}
	}
	return moves
}

// IsFull checks if the board is completely filled
func (b *Board) IsFull() bool {
	for col := 0; col < Cols; col++ {
		if b[0][col] == Empty {
			return false
		}
	}
	return true
}

// CheckWin checks if the given player has won
func (b *Board) CheckWin(player int) bool {
	// Check horizontal
	for row := 0; row < Rows; row++ {
		for col := 0; col <= Cols-4; col++ {
			if b[row][col] == player &&
				b[row][col+1] == player &&
				b[row][col+2] == player &&
				b[row][col+3] == player {
				return true
			}
		}
	}

	// Check vertical
	for row := 0; row <= Rows-4; row++ {
		for col := 0; col < Cols; col++ {
			if b[row][col] == player &&
				b[row+1][col] == player &&
				b[row+2][col] == player &&
				b[row+3][col] == player {
				return true
			}
		}
	}

	// Check diagonal (bottom-left to top-right)
	for row := 3; row < Rows; row++ {
		for col := 0; col <= Cols-4; col++ {
			if b[row][col] == player &&
				b[row-1][col+1] == player &&
				b[row-2][col+2] == player &&
				b[row-3][col+3] == player {
				return true
			}
		}
	}

	// Check diagonal (top-left to bottom-right)
	for row := 0; row <= Rows-4; row++ {
		for col := 0; col <= Cols-4; col++ {
			if b[row][col] == player &&
				b[row+1][col+1] == player &&
				b[row+2][col+2] == player &&
				b[row+3][col+3] == player {
				return true
			}
		}
	}

	return false
}

// Copy creates a deep copy of the board
func (b *Board) Copy() Board {
	var newBoard Board
	for row := 0; row < Rows; row++ {
		for col := 0; col < Cols; col++ {
			newBoard[row][col] = b[row][col]
		}
	}
	return newBoard
}

// GetCell returns the value at a specific cell
func (b *Board) GetCell(row, col int) int {
	if row < 0 || row >= Rows || col < 0 || col >= Cols {
		return -1
	}
	return b[row][col]
}

// WinningCell represents a cell in the winning combination
type WinningCell struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// GetWinningCells returns the 4 cells that form the winning combination
func (b *Board) GetWinningCells(player int) []WinningCell {
	// Check horizontal
	for row := 0; row < Rows; row++ {
		for col := 0; col <= Cols-4; col++ {
			if b[row][col] == player &&
				b[row][col+1] == player &&
				b[row][col+2] == player &&
				b[row][col+3] == player {
				return []WinningCell{
					{Row: row, Col: col},
					{Row: row, Col: col + 1},
					{Row: row, Col: col + 2},
					{Row: row, Col: col + 3},
				}
			}
		}
	}

	// Check vertical
	for row := 0; row <= Rows-4; row++ {
		for col := 0; col < Cols; col++ {
			if b[row][col] == player &&
				b[row+1][col] == player &&
				b[row+2][col] == player &&
				b[row+3][col] == player {
				return []WinningCell{
					{Row: row, Col: col},
					{Row: row + 1, Col: col},
					{Row: row + 2, Col: col},
					{Row: row + 3, Col: col},
				}
			}
		}
	}

	// Check diagonal (bottom-left to top-right)
	for row := 3; row < Rows; row++ {
		for col := 0; col <= Cols-4; col++ {
			if b[row][col] == player &&
				b[row-1][col+1] == player &&
				b[row-2][col+2] == player &&
				b[row-3][col+3] == player {
				return []WinningCell{
					{Row: row, Col: col},
					{Row: row - 1, Col: col + 1},
					{Row: row - 2, Col: col + 2},
					{Row: row - 3, Col: col + 3},
				}
			}
		}
	}

	// Check diagonal (top-left to bottom-right)
	for row := 0; row <= Rows-4; row++ {
		for col := 0; col <= Cols-4; col++ {
			if b[row][col] == player &&
				b[row+1][col+1] == player &&
				b[row+2][col+2] == player &&
				b[row+3][col+3] == player {
				return []WinningCell{
					{Row: row, Col: col},
					{Row: row + 1, Col: col + 1},
					{Row: row + 2, Col: col + 2},
					{Row: row + 3, Col: col + 3},
				}
			}
		}
	}

	return nil
}
