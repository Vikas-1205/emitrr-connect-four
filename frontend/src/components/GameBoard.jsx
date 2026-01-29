import { useState, useEffect } from 'react';

const ROWS = 6;
const COLS = 7;

export default function GameBoard({
    board,
    currentPlayer,
    playerNum,
    onColumnClick,
    gameState,
    lastMove
}) {
    const [droppingCell, setDroppingCell] = useState(null);

    const isMyTurn = currentPlayer === playerNum;
    const isGameActive = gameState?.status === 'in_progress';
    const winningCells = gameState?.winningCells || [];

    // Create a set of winning cell keys for quick lookup
    const winningCellSet = new Set(
        winningCells.map(cell => `${cell.row}-${cell.col}`)
    );

    // Handle drop animation
    useEffect(() => {
        if (lastMove) {
            setDroppingCell(`${lastMove.row}-${lastMove.col}`);
            const timeout = setTimeout(() => setDroppingCell(null), 300);
            return () => clearTimeout(timeout);
        }
    }, [lastMove]);

    const handleColumnClick = (col) => {
        if (!isMyTurn || !isGameActive) return;

        // Check if column is full
        if (board[0][col] !== 0) return;

        onColumnClick(col);
    };

    const getCellClass = (row, col) => {
        const cellValue = board[row][col];
        let classes = ['cell'];

        if (cellValue === 1) classes.push('player1');
        else if (cellValue === 2) classes.push('player2');

        if (droppingCell === `${row}-${col}`) {
            classes.push('dropping');
        }

        // Highlight winning cells
        if (winningCellSet.has(`${row}-${col}`)) {
            classes.push('winning');
        }

        return classes.join(' ');
    };

    return (
        <div className="game-board-container">
            <div
                className="turn-indicator"
                style={{
                    marginBottom: '1rem',
                    background: isGameActive
                        ? (isMyTurn ? 'rgba(74, 222, 128, 0.2)' : 'rgba(108, 99, 255, 0.2)')
                        : 'rgba(255, 215, 0, 0.2)'
                }}
            >
                {isGameActive ? (
                    isMyTurn ? '🎯 Your Turn!' : "⏳ Opponent's Turn..."
                ) : gameState?.winner ? (
                    gameState.winner === playerNum ? '🎉 You Won!' : '😔 You Lost'
                ) : (
                    "🤝 It's a Draw!"
                )}
            </div>

            <div className="game-board">
                {Array.from({ length: COLS }).map((_, col) => (
                    <div
                        key={col}
                        className={`column ${!isMyTurn || !isGameActive ? 'disabled' : ''}`}
                        onClick={() => handleColumnClick(col)}
                    >
                        {Array.from({ length: ROWS }).map((_, row) => (
                            <div
                                key={`${row}-${col}`}
                                className={getCellClass(row, col)}
                            />
                        ))}
                    </div>
                ))}
            </div>

            {/* Winning Combination Info */}
            {winningCells.length > 0 && (
                <div style={{
                    marginTop: '1rem',
                    padding: '0.75rem 1rem',
                    background: 'rgba(255, 215, 0, 0.2)',
                    borderRadius: '8px',
                    textAlign: 'center',
                    fontSize: '0.9rem'
                }}>
                    🏆 Winning combination highlighted!
                </div>
            )}
        </div>
    );
}
